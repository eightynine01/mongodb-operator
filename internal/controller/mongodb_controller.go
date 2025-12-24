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
	mongodbFinalizer = "mongodb.keiailab.com/finalizer"
	requeueAfter     = 30 * time.Second
)

// MongoDBReconciler reconciles a MongoDB object
type MongoDBReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mongodb.keiailab.com,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.keiailab.com,resources=mongodbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.keiailab.com,resources=mongodbs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete

func (r *MongoDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling MongoDB", "namespace", req.Namespace, "name", req.Name)

	// Fetch MongoDB instance
	mdb := &mongodbv1alpha1.MongoDB{}
	if err := r.Get(ctx, req.NamespacedName, mdb); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("MongoDB resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get MongoDB")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !mdb.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, mdb)
	}

	// Add finalizer if needed
	if !controllerutil.ContainsFinalizer(mdb, mongodbFinalizer) {
		controllerutil.AddFinalizer(mdb, mongodbFinalizer)
		if err := r.Update(ctx, mdb); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Update status phase to Initializing if pending
	if mdb.Status.Phase == "" || mdb.Status.Phase == "Pending" {
		mdb.Status.Phase = "Initializing"
		if err := r.Status().Update(ctx, mdb); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile resources in order

	// 1. Keyfile Secret
	if err := r.reconcileKeyfileSecret(ctx, mdb); err != nil {
		return r.updateStatusError(ctx, mdb, "KeyfileSecret", err)
	}

	// 2. ConfigMap
	if err := r.reconcileConfigMap(ctx, mdb); err != nil {
		return r.updateStatusError(ctx, mdb, "ConfigMap", err)
	}

	// 3. Headless Service
	if err := r.reconcileHeadlessService(ctx, mdb); err != nil {
		return r.updateStatusError(ctx, mdb, "HeadlessService", err)
	}

	// 4. Client Service
	if err := r.reconcileClientService(ctx, mdb); err != nil {
		return r.updateStatusError(ctx, mdb, "ClientService", err)
	}

	// 5. StatefulSet
	if err := r.reconcileStatefulSet(ctx, mdb); err != nil {
		return r.updateStatusError(ctx, mdb, "StatefulSet", err)
	}

	// 6. Update status
	if err := r.updateStatus(ctx, mdb); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled MongoDB")
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *MongoDBReconciler) handleDeletion(ctx context.Context, mdb *mongodbv1alpha1.MongoDB) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling MongoDB deletion")

	if controllerutil.ContainsFinalizer(mdb, mongodbFinalizer) {
		// Perform cleanup logic here if needed

		// Remove finalizer
		controllerutil.RemoveFinalizer(mdb, mongodbFinalizer)
		if err := r.Update(ctx, mdb); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *MongoDBReconciler) reconcileKeyfileSecret(ctx context.Context, mdb *mongodbv1alpha1.MongoDB) error {
	secret := resources.BuildKeyfileSecret(mdb)
	return r.createOrUpdate(ctx, mdb, secret)
}

func (r *MongoDBReconciler) reconcileConfigMap(ctx context.Context, mdb *mongodbv1alpha1.MongoDB) error {
	cm := resources.BuildMongoDBConfigMap(mdb)
	return r.createOrUpdate(ctx, mdb, cm)
}

func (r *MongoDBReconciler) reconcileHeadlessService(ctx context.Context, mdb *mongodbv1alpha1.MongoDB) error {
	svc := resources.BuildHeadlessService(mdb)
	return r.createOrUpdate(ctx, mdb, svc)
}

func (r *MongoDBReconciler) reconcileClientService(ctx context.Context, mdb *mongodbv1alpha1.MongoDB) error {
	svc := resources.BuildClientService(mdb)
	return r.createOrUpdate(ctx, mdb, svc)
}

func (r *MongoDBReconciler) reconcileStatefulSet(ctx context.Context, mdb *mongodbv1alpha1.MongoDB) error {
	sts := resources.BuildReplicaSetStatefulSet(mdb)
	return r.createOrUpdate(ctx, mdb, sts)
}

func (r *MongoDBReconciler) createOrUpdate(ctx context.Context, mdb *mongodbv1alpha1.MongoDB, obj client.Object) error {
	// Set owner reference
	if err := controllerutil.SetControllerReference(mdb, obj, r.Scheme); err != nil {
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

func (r *MongoDBReconciler) updateStatus(ctx context.Context, mdb *mongodbv1alpha1.MongoDB) error {
	// Get StatefulSet status
	sts := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: mdb.Name, Namespace: mdb.Namespace}, sts); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		mdb.Status.ReadyMembers = 0
	} else {
		mdb.Status.ReadyMembers = sts.Status.ReadyReplicas
	}

	// Update phase based on ready members
	if mdb.Status.ReadyMembers == mdb.Spec.Members {
		mdb.Status.Phase = "Running"
	} else if mdb.Status.ReadyMembers > 0 {
		mdb.Status.Phase = "Initializing"
	}

	// Set connection string
	mdb.Status.ConnectionString = fmt.Sprintf("mongodb://%s-headless.%s.svc.cluster.local:27017/?replicaSet=%s",
		mdb.Name, mdb.Namespace, mdb.Spec.ReplicaSetName)

	mdb.Status.Version = mdb.Spec.Version.Version
	mdb.Status.ObservedGeneration = mdb.Generation

	// Update conditions
	mdb.Status.Conditions = r.buildConditions(mdb)

	return r.Status().Update(ctx, mdb)
}

func (r *MongoDBReconciler) buildConditions(mdb *mongodbv1alpha1.MongoDB) []metav1.Condition {
	conditions := []metav1.Condition{}

	// Ready condition
	readyStatus := metav1.ConditionFalse
	readyReason := "NotReady"
	readyMessage := fmt.Sprintf("%d/%d members ready", mdb.Status.ReadyMembers, mdb.Spec.Members)

	if mdb.Status.ReadyMembers == mdb.Spec.Members {
		readyStatus = metav1.ConditionTrue
		readyReason = "Ready"
		readyMessage = "All members are ready"
	}

	conditions = append(conditions, metav1.Condition{
		Type:               "Ready",
		Status:             readyStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             readyReason,
		Message:            readyMessage,
	})

	return conditions
}

func (r *MongoDBReconciler) updateStatusError(ctx context.Context, mdb *mongodbv1alpha1.MongoDB, component string, err error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Error(err, "Failed to reconcile component", "component", component)

	mdb.Status.Phase = "Failed"
	mdb.Status.Conditions = append(mdb.Status.Conditions, metav1.Condition{
		Type:               "ReconcileError",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "ReconcileFailed",
		Message:            fmt.Sprintf("Failed to reconcile %s: %v", component, err),
	})

	if statusErr := r.Status().Update(ctx, mdb); statusErr != nil {
		logger.Error(statusErr, "Failed to update status")
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbv1alpha1.MongoDB{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
