# Architecture Overview

## Overview

MongoDB Operator is built using the Kubernetes Operator SDK and controller-runtime framework. This guide explains the operator's architecture, design patterns, and key components.

## Controller Design

### Controller Pattern

The operator follows the standard Kubernetes controller pattern:

```go
// controllers/mongodb_controller.go
package controllers

import (
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"
)

type MongoDBReconciler struct {
    client.Client
    Scheme *runtime.Scheme
    // Additional dependencies
    Executor  *mongodb.Executor
    Builder   *resource.Builder
    ConfigMgr *config.Manager
}

func (r *MongoDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // 1. Fetch MongoDB CRD
    mongodb := &mongodbv1alpha1.MongoDB{}
    if err := r.Get(ctx, req.NamespacedName, mongodb); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. Check deletion and handle finalizers
    if !mongodb.DeletionTimestamp.IsZero() {
        return r.handleDeletion(ctx, mongodb)
    }

    // 3. Add finalizer if not present
    if !containsString(mongodb.Finalizers, mongodbFinalizer) {
        mongodb.Finalizers = append(mongodb.Finalizers, mongodbFinalizer)
        return ctrl.Result{Requeue: true}, r.Update(ctx, mongodb)
    }

    // 4. Build Kubernetes resources
    resources, err := r.Builder.BuildMongoDBResources(ctx, mongodb)
    if err != nil {
        return ctrl.Result{}, err
    }

    // 5. Apply resources to cluster
    for _, res := range resources {
        if err := r.CreateOrUpdate(ctx, res); err != nil {
            return ctrl.Result{}, err
        }
    }

    // 6. Initialize MongoDB cluster
    if !mongodb.Status.Ready {
        if err := r.initializeCluster(ctx, mongodb); err != nil {
            return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
        }
    }

    // 7. Update status
    mongodb.Status.Ready = true
    return ctrl.Result{}, r.Status().Update(ctx, mongodb)
}
```

### Reconciliation Loop

The reconciliation loop handles the following states:

1. **Creation**: Build and create all Kubernetes resources
2. **Initialization**: Execute MongoDB initialization commands
3. **Update**: Detect spec changes and update resources
4. **Scale**: Handle scaling operations (members, resources)
5. **Deletion**: Clean up resources and finalizers

**Requeue Logic:**
```go
// Requeue immediately
return ctrl.Result{Requeue: true}, nil

// Requeue after delay
return ctrl.Result{RequeueAfter: 30 * time.Second}, nil

// Don't requeue (operation complete)
return ctrl.Result{}, nil

// Requeue on error (default behavior)
return ctrl.Result{}, err
```

### Watch Pattern

Controllers watch multiple Kubernetes resources:

```go
func (r *MongoDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&mongodbv1alpha1.MongoDB{}).
        Owns(&appsv1.StatefulSet{}).
        Owns(&corev1.Service{}).
        Owns(&corev1.Secret{}).
        Owns(&corev1.ConfigMap{}).
        Complete(r)
}
```

**Owned Resources:**
- `StatefulSet`: MongoDB shard/member pods
- `Deployment`: Mongos routers
- `Service`: Headless and client services
- `Secret`: Keyfiles and credentials
- `ConfigMap`: MongoDB configuration
- `Job`: Backup operations

## Resource Builders

### Builder Pattern

The Resource Builder follows the builder pattern for constructing Kubernetes resources:

```go
// internal/resource/builder.go
package resource

import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

type Builder struct {
    client.Client
    Scheme *runtime.Scheme
}

func (b *Builder) BuildMongoDBResources(ctx context.Context, mongodb *mongodbv1alpha1.MongoDB) ([]client.Object, error) {
    var resources []client.Object

    // Build StatefulSet
    sts, err := b.BuildStatefulSet(mongodb)
    if err != nil {
        return nil, err
    }
    resources = append(resources, sts)

    // Build Services
    svc, err := b.BuildServices(mongodb)
    if err != nil {
        return nil, err
    }
    resources = append(resources, svc...)

    // Build ConfigMap
    cm, err := b.BuildConfigMap(mongodb)
    if err != nil {
        return nil, err
    }
    resources = append(resources, cm)

    // Build Secret
    secret, err := b.BuildSecret(mongodb)
    if err != nil {
        return nil, err
    }
    resources = append(resources, secret)

    return resources, nil
}
```

### StatefulSet Builder

```go
// internal/resource/statefulset.go
func (b *Builder) BuildStatefulSet(mongodb *mongodbv1alpha1.MongoDB) (*appsv1.StatefulSet, error) {
    sts := &appsv1.StatefulSet{
        ObjectMeta: metav1.ObjectMeta{
            Name:      mongodb.Name,
            Namespace: mongodb.Namespace,
            Labels:    b.buildLabels(mongodb),
        },
        Spec: appsv1.StatefulSetSpec{
            Replicas: &mongodb.Spec.Members,
            Selector: &metav1.LabelSelector{
                MatchLabels: b.buildLabels(mongodb),
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: b.buildLabels(mongodb),
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "mongod",
                            Image: fmt.Sprintf("mongo:%s", mongodb.Spec.Version.Version),
                            Command: []string{"mongod"},
                            Args:    b.buildMongoDBArgs(mongodb),
                            Ports: []corev1.ContainerPort{
                                {
                                    Name:          "mongodb",
                                    ContainerPort: 27017,
                                },
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      "data",
                                    MountPath: "/data/db",
                                },
                                {
                                    Name:      "config",
                                    MountPath: "/etc/mongod.conf",
                                    SubPath:   "mongod.conf",
                                },
                            },
                            Env: b.buildMongoDBEnv(mongodb),
                            LivenessProbe: &corev1.Probe{
                                ProbeHandler: corev1.ProbeHandler{
                                    Exec: &corev1.ExecAction{
                                        Command: []string{"mongosh", "--eval", "db.adminCommand('ping')"},
                                    },
                                },
                                InitialDelaySeconds: 30,
                                TimeoutSeconds:      5,
                            },
                            ReadinessProbe: &corev1.Probe{
                                ProbeHandler: corev1.ProbeHandler{
                                    Exec: &corev1.ExecAction{
                                        Command: []string{"mongosh", "--eval", "db.adminCommand('ping')"},
                                    },
                                },
                                InitialDelaySeconds: 10,
                                TimeoutSeconds:      5,
                            },
                        },
                    },
                    Volumes: []corev1.Volume{
                        {
                            Name: "config",
                            VolumeSource: corev1.VolumeSource{
                                ConfigMap: &corev1.ConfigMapVolumeSource{
                                    LocalObjectReference: corev1.LocalObjectReference{
                                        Name: mongodb.Name + "-config",
                                    },
                                },
                            },
                        },
                    },
                },
            },
            VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
                {
                    ObjectMeta: metav1.ObjectMeta{
                        Name: "data",
                    },
                    Spec: corev1.PersistentVolumeClaimSpec{
                        AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
                        Resources: corev1.VolumeResourceRequirements{
                            Requests: corev1.ResourceList{
                                corev1.ResourceStorage: resource.MustParse(mongodb.Spec.Storage.Size),
                            },
                        },
                        StorageClassName: &mongodb.Spec.Storage.StorageClassName,
                    },
                },
            },
        },
    }

    return sts, nil
}
```

## MongoDB Package (internal/mongodb)

The MongoDB package handles all MongoDB-specific operations:

### Executor

```go
// internal/mongodb/executor.go
package mongodb

import (
    "context"
    "fmt"
)

type Command struct {
    Command string
    Args    []string
}

type Executor struct {
    client Client
    logger logr.Logger
}

type Client interface {
    Execute(ctx context.Context, command Command) (string, error)
}

func (e *Executor) ExecuteCommand(ctx context.Context, command Command) (string, error) {
    e.logger.Info("Executing MongoDB command", "command", command.Command)

    result, err := e.client.Execute(ctx, command)
    if err != nil {
        e.logger.Error(err, "Failed to execute command")
        return "", err
    }

    return result, nil
}
```

### ReplicaSet Operations

```go
// internal/mongodb/replicaset.go
func (e *Executor) InitiateReplicaSet(ctx context.Context, config ReplicaSetConfig) error {
    cmd := Command{
        Command: "mongosh",
        Args: []string{
            "--eval",
            fmt.Sprintf("rs.initiate({_id: %q, members: %s})",
                config.Name,
                config.MembersJSON),
        },
    }

    _, err := e.ExecuteCommand(ctx, cmd)
    return err
}

func (e *Executor) AddMember(ctx context.Context, replicaSetName, memberHost string) error {
    cmd := Command{
        Command: "mongosh",
        Args: []string{
            "--eval",
            fmt.Sprintf("rs.add(%q)", memberHost),
        },
    }

    _, err := e.ExecuteCommand(ctx, cmd)
    return err
}

func (e *Executor) GetStatus(ctx context.Context) (*ReplicaSetStatus, error) {
    cmd := Command{
        Command: "mongosh",
        Args: []string{
            "--eval",
            "JSON.stringify(rs.status())",
        },
    }

    result, err := e.ExecuteCommand(ctx, cmd)
    if err != nil {
        return nil, err
    }

    var status ReplicaSetStatus
    if err := json.Unmarshal([]byte(result), &status); err != nil {
        return nil, err
    }

    return &status, nil
}
```

### Authentication

```go
// internal/mongodb/auth.go
func (e *Executor) CreateAdminUser(ctx context.Context, config AuthConfig) error {
    cmd := Command{
        Command: "mongosh",
        Args: []string{
            "--eval",
            fmt.Sprintf(
                `db.getSiblingDB("admin").createUser({
                    user: %q,
                    pwd: %q,
                    roles: [{role: "root", db: "admin"}]
                })`,
                config.Username,
                config.Password,
            ),
        },
    }

    _, err := e.ExecuteCommand(ctx, cmd)
    return err
}

func (e *Executor) CreateUser(ctx context.Context, database, username, password string, roles []string) error {
    rolesJSON, _ := json.Marshal(roles)
    cmd := Command{
        Command: "mongosh",
        Args: []string{
            "--eval",
            fmt.Sprintf(
                `db.getSiblingDB(%q).createUser({
                    user: %q,
                    pwd: %q,
                    roles: %s
                })`,
                database,
                username,
                password,
                string(rolesJSON),
            ),
        },
    }

    _, err := e.ExecuteCommand(ctx, cmd)
    return err
}
```

### Sharding

```go
// internal/mongodb/sharding.go
func (e *Executor) AddShard(ctx context.Context, shardConnString string) error {
    cmd := Command{
        Command: "mongosh",
        Args: []string{
            "--eval",
            fmt.Sprintf("sh.addShard(%q)", shardConnString),
        },
    }

    _, err := e.ExecuteCommand(ctx, cmd)
    return err
}

func (e *Executor) EnableSharding(ctx context.Context, database string) error {
    cmd := Command{
        Command: "mongosh",
        Args: []string{
            "--eval",
            fmt.Sprintf("sh.enableSharding(%q)", database),
        },
    }

    _, err := e.ExecuteCommand(ctx, cmd)
    return err
}

func (e *Executor) ShardCollection(ctx context.Context, namespace, shardKey string) error {
    cmd := Command{
        Command: "mongosh",
        Args: []string{
            "--eval",
            fmt.Sprintf("sh.shardCollection(%q, {%s})", namespace, shardKey),
        },
    }

    _, err := e.ExecuteCommand(ctx, cmd)
    return err
}
```

## Finalizer Patterns

Finalizers ensure clean resource deletion:

```go
// controllers/mongodb_controller.go
const (
    mongodbFinalizer = "mongodb.keiailab.com/finalizer"
)

func (r *MongoDBReconciler) handleDeletion(ctx context.Context, mongodb *mongodbv1alpha1.MongoDB) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    if !containsString(mongodb.Finalizers, mongodbFinalizer) {
        return ctrl.Result{}, nil
    }

    // 1. Remove admin user if needed
    if err := r.cleanupAdminUser(ctx, mongodb); err != nil {
        log.Error(err, "Failed to cleanup admin user")
        return ctrl.Result{}, err
    }

    // 2. Remove shard if part of sharded cluster
    if mongodb.Spec.ClusterRef != nil {
        if err := r.removeShard(ctx, mongodb); err != nil {
            log.Error(err, "Failed to remove shard")
            return ctrl.Result{}, err
        }
    }

    // 3. Remove finalizer
    mongodb.Finalizers = removeString(mongodb.Finalizers, mongodbFinalizer)
    if err := r.Update(ctx, mongodb); err != nil {
        return ctrl.Result{}, err
    }

    log.Info("MongoDB deleted successfully")
    return ctrl.Result{}, nil
}
```

## Status Management

Status updates track cluster state:

```go
// api/v1alpha1/mongodb_types.go
type MongoDBStatus struct {
    // Ready indicates the MongoDB cluster is operational
    Ready bool `json:"ready"`

    // MembersReady indicates all members are healthy
    MembersReady bool `json:"membersReady"`

    // CurrentPrimary is the hostname of the primary
    CurrentPrimary string `json:"currentPrimary,omitempty"`

    // Conditions provide details about cluster state
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // LastUpdateTime is the last time status was updated
    LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}

// In controller
func (r *MongoDBReconciler) updateStatus(ctx context.Context, mongodb *mongodbv1alpha1.MongoDB, status string) error {
    now := metav1.Now()
    mongodb.Status.LastUpdateTime = &now

    // Update conditions
    condition := metav1.Condition{
        Type:               "Ready",
        Status:             metav1.ConditionTrue,
        LastTransitionTime: now,
        Reason:             "Operational",
        Message:            status,
    }

    mongodb.Status.Conditions = upsertCondition(mongodb.Status.Conditions, condition)

    return r.Status().Update(ctx, mongodb)
}
```

## Error Handling

### Retry Logic

```go
// Retry with exponential backoff
func (r *MongoDBReconciler) reconcileWithRetry(ctx context.Context, mongodb *mongodbv1alpha1.MongoDB) (ctrl.Result, error) {
    err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
        // Fetch latest version
        if err := r.Get(ctx, client.ObjectKeyFromObject(mongodb), mongodb); err != nil {
            return err
        }

        // Apply changes
        return r.applyChanges(ctx, mongodb)
    })

    if err != nil {
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}
```

### Error Events

```go
// Record events for errors
import (
    "k8s.io/client-go/tools/record"
)

type MongoDBReconciler struct {
    client.Client
    Scheme   *runtime.Scheme
    Recorder record.EventRecorder
}

func (r *MongoDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ...

    if err != nil {
        r.Recorder.Eventf(mongodb, corev1.EventTypeWarning, "ReconciliationFailed", "Failed to reconcile: %v", err)
        return ctrl.Result{}, err
    }

    r.Recorder.Eventf(mongodb, corev1.EventTypeNormal, "Reconciled", "Successfully reconciled MongoDB")
    // ...
}
```

## Data Flow

1. **Watch**: Controller watches MongoDB CRD changes
2. **Reconcile**: Triggered on change, executes reconciliation logic
3. **Build**: Resource Builder constructs Kubernetes resources
4. **Apply**: Resources created/updated in cluster
5. **Initialize**: Executor runs MongoDB initialization commands
6. **Status**: Controller updates CRD status
7. **Finalize**: On deletion, finalizers clean up resources
