/*
Copyright 2024 Keiailab.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mongodbv1alpha1 "github.com/keiailab/mongodb-operator/api/v1alpha1"
	"github.com/keiailab/mongodb-operator/internal/resources"
)

const (
	mongodbShardedFinalizer = "mongodbsharded.keiailab.com/finalizer"
)

// MongoDBShardedReconciler reconciles a MongoDBSharded object
type MongoDBShardedReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mongodb.keiailab.com,resources=mongodbshardeds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.keiailab.com,resources=mongodbshardeds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.keiailab.com,resources=mongodbshardeds/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create

func (r *MongoDBShardedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling MongoDBSharded", "namespace", req.Namespace, "name", req.Name)

	// Fetch MongoDBSharded instance
	mdbsh := &mongodbv1alpha1.MongoDBSharded{}
	if err := r.Get(ctx, req.NamespacedName, mdbsh); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("MongoDBSharded resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get MongoDBSharded")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !mdbsh.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, mdbsh)
	}

	// Add finalizer if needed
	if !controllerutil.ContainsFinalizer(mdbsh, mongodbShardedFinalizer) {
		controllerutil.AddFinalizer(mdbsh, mongodbShardedFinalizer)
		if err := r.Update(ctx, mdbsh); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Update status phase to Initializing if pending
	if mdbsh.Status.Phase == "" || mdbsh.Status.Phase == "Pending" {
		mdbsh.Status.Phase = "Initializing"
		if err := r.Status().Update(ctx, mdbsh); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile resources in order

	// 1. Keyfile Secret
	if err := r.reconcileKeyfileSecret(ctx, mdbsh); err != nil {
		return r.updateStatusError(ctx, mdbsh, "KeyfileSecret", err)
	}

	// 2. Config Server
	if err := r.reconcileConfigServer(ctx, mdbsh); err != nil {
		return r.updateStatusError(ctx, mdbsh, "ConfigServer", err)
	}

	// 3. Wait for Config Server to be ready
	if !r.isConfigServerReady(ctx, mdbsh) {
		logger.Info("Waiting for config server to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// 4. Shards
	for i := int32(0); i < mdbsh.Spec.Shards.Count; i++ {
		if err := r.reconcileShard(ctx, mdbsh, i); err != nil {
			return r.updateStatusError(ctx, mdbsh, fmt.Sprintf("Shard-%d", i), err)
		}
	}

	// 5. Wait for Shards to be ready
	if !r.areShardsReady(ctx, mdbsh) {
		logger.Info("Waiting for shards to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// 6. Mongos
	if err := r.reconcileMongos(ctx, mdbsh); err != nil {
		return r.updateStatusError(ctx, mdbsh, "Mongos", err)
	}

	// 7. Update status
	if err := r.updateStatus(ctx, mdbsh); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled MongoDBSharded")
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *MongoDBShardedReconciler) handleDeletion(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling MongoDBSharded deletion")

	if controllerutil.ContainsFinalizer(mdbsh, mongodbShardedFinalizer) {
		// Perform cleanup logic here if needed

		// Remove finalizer
		controllerutil.RemoveFinalizer(mdbsh, mongodbShardedFinalizer)
		if err := r.Update(ctx, mdbsh); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *MongoDBShardedReconciler) reconcileKeyfileSecret(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded) error {
	secret := resources.BuildShardedKeyfileSecret(mdbsh)
	return r.createOrUpdate(ctx, mdbsh, secret)
}

func (r *MongoDBShardedReconciler) reconcileConfigServer(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded) error {
	// Headless service
	svc := resources.BuildConfigServerService(mdbsh)
	if err := r.createOrUpdate(ctx, mdbsh, svc); err != nil {
		return err
	}

	// StatefulSet
	sts := resources.BuildConfigServerStatefulSet(mdbsh)
	return r.createOrUpdate(ctx, mdbsh, sts)
}

func (r *MongoDBShardedReconciler) isConfigServerReady(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded) bool {
	sts := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: mdbsh.Name + "-cfg", Namespace: mdbsh.Namespace}, sts); err != nil {
		return false
	}
	return sts.Status.ReadyReplicas == mdbsh.Spec.ConfigServer.Members
}

func (r *MongoDBShardedReconciler) reconcileShard(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded, shardIndex int32) error {
	// Headless service
	svc := resources.BuildShardService(mdbsh, shardIndex)
	if err := r.createOrUpdate(ctx, mdbsh, svc); err != nil {
		return err
	}

	// StatefulSet
	sts := resources.BuildShardStatefulSet(mdbsh, shardIndex)
	return r.createOrUpdate(ctx, mdbsh, sts)
}

func (r *MongoDBShardedReconciler) areShardsReady(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded) bool {
	for i := int32(0); i < mdbsh.Spec.Shards.Count; i++ {
		sts := &appsv1.StatefulSet{}
		stsName := fmt.Sprintf("%s-shard-%d", mdbsh.Name, i)
		if err := r.Get(ctx, types.NamespacedName{Name: stsName, Namespace: mdbsh.Namespace}, sts); err != nil {
			return false
		}
		if sts.Status.ReadyReplicas != mdbsh.Spec.Shards.MembersPerShard {
			return false
		}
	}
	return true
}

func (r *MongoDBShardedReconciler) reconcileMongos(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded) error {
	// ConfigMap
	cm := resources.BuildMongosConfigMap(mdbsh)
	if err := r.createOrUpdate(ctx, mdbsh, cm); err != nil {
		return err
	}

	// Service
	svc := resources.BuildMongosService(mdbsh)
	if err := r.createOrUpdate(ctx, mdbsh, svc); err != nil {
		return err
	}

	// Deployment
	deploy := resources.BuildMongosDeployment(mdbsh)
	return r.createOrUpdate(ctx, mdbsh, deploy)
}

func (r *MongoDBShardedReconciler) createOrUpdate(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded, obj client.Object) error {
	// Set owner reference
	if err := controllerutil.SetControllerReference(mdbsh, obj, r.Scheme); err != nil {
		return err
	}

	// Check if object exists
	existing := obj.DeepCopyObject().(client.Object)
	err := r.Get(ctx, types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, existing)

	if err != nil {
		if errors.IsNotFound(err) {
			// Create the object
			return r.Create(ctx, obj)
		}
		return err
	}

	// Update the object
	obj.SetResourceVersion(existing.GetResourceVersion())
	return r.Update(ctx, obj)
}

func (r *MongoDBShardedReconciler) updateStatus(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded) error {
	// Update ConfigServer status
	cfgSts := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: mdbsh.Name + "-cfg", Namespace: mdbsh.Namespace}, cfgSts); err == nil {
		mdbsh.Status.ConfigServer = mongodbv1alpha1.ComponentStatus{
			Ready: cfgSts.Status.ReadyReplicas,
			Total: mdbsh.Spec.ConfigServer.Members,
			Phase: r.getComponentPhase(cfgSts.Status.ReadyReplicas, mdbsh.Spec.ConfigServer.Members),
		}
	}

	// Update Shards status
	mdbsh.Status.Shards = []mongodbv1alpha1.ShardStatus{}
	for i := int32(0); i < mdbsh.Spec.Shards.Count; i++ {
		shardSts := &appsv1.StatefulSet{}
		stsName := fmt.Sprintf("%s-shard-%d", mdbsh.Name, i)
		if err := r.Get(ctx, types.NamespacedName{Name: stsName, Namespace: mdbsh.Namespace}, shardSts); err == nil {
			mdbsh.Status.Shards = append(mdbsh.Status.Shards, mongodbv1alpha1.ShardStatus{
				Name:  stsName,
				Ready: shardSts.Status.ReadyReplicas,
				Total: mdbsh.Spec.Shards.MembersPerShard,
				Phase: r.getComponentPhase(shardSts.Status.ReadyReplicas, mdbsh.Spec.Shards.MembersPerShard),
			})
		}
	}

	// Update Mongos status
	mongosDeploy := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: mdbsh.Name + "-mongos", Namespace: mdbsh.Namespace}, mongosDeploy); err == nil {
		mdbsh.Status.Mongos = mongodbv1alpha1.ComponentStatus{
			Ready: mongosDeploy.Status.ReadyReplicas,
			Total: mdbsh.Spec.Mongos.Replicas,
			Phase: r.getComponentPhase(mongosDeploy.Status.ReadyReplicas, mdbsh.Spec.Mongos.Replicas),
		}
	}

	// Update overall phase
	if r.isClusterReady(mdbsh) {
		mdbsh.Status.Phase = "Running"
	} else {
		mdbsh.Status.Phase = "Initializing"
	}

	// Set connection string
	mdbsh.Status.ConnectionString = fmt.Sprintf("mongodb://%s-mongos.%s.svc.cluster.local:27017",
		mdbsh.Name, mdbsh.Namespace)

	mdbsh.Status.ObservedGeneration = mdbsh.Generation

	return r.Status().Update(ctx, mdbsh)
}

func (r *MongoDBShardedReconciler) getComponentPhase(ready, total int32) string {
	if ready == total {
		return "Running"
	}
	if ready > 0 {
		return "Initializing"
	}
	return "Pending"
}

func (r *MongoDBShardedReconciler) isClusterReady(mdbsh *mongodbv1alpha1.MongoDBSharded) bool {
	if mdbsh.Status.ConfigServer.Ready != mdbsh.Spec.ConfigServer.Members {
		return false
	}
	if mdbsh.Status.Mongos.Ready != mdbsh.Spec.Mongos.Replicas {
		return false
	}
	for _, shard := range mdbsh.Status.Shards {
		if shard.Ready != mdbsh.Spec.Shards.MembersPerShard {
			return false
		}
	}
	return true
}

func (r *MongoDBShardedReconciler) updateStatusError(ctx context.Context, mdbsh *mongodbv1alpha1.MongoDBSharded, component string, err error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Error(err, "Failed to reconcile component", "component", component)

	mdbsh.Status.Phase = "Failed"
	mdbsh.Status.Conditions = append(mdbsh.Status.Conditions, metav1.Condition{
		Type:               "ReconcileError",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "ReconcileFailed",
		Message:            fmt.Sprintf("Failed to reconcile %s: %v", component, err),
	})

	if statusErr := r.Status().Update(ctx, mdbsh); statusErr != nil {
		logger.Error(statusErr, "Failed to update status")
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBShardedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbv1alpha1.MongoDBSharded{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
