// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/operator-framework/operator-sdk/internal/terraform/config"
	"github.com/operator-framework/operator-sdk/internal/terraform/internal/types"
)

// blank assignment to verify that TerraformReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &TerraformReconciler{}

// TerraformReconciler reconciles custom resources as Terraform configurations.
type TerraformReconciler struct {
	Client          client.Client
	EventRecorder   record.EventRecorder
	GVK             schema.GroupVersionKind
	ManagerFactory  config.ManagerFactory
	ReconcilePeriod time.Duration
	OverrideValues  map[string]string
}

const (
	finalizer = "destroy-terraform-config"
)

// Reconcile reconciles the requested resource by creating, updating, or
// destroying a Terraform configuration based on the resource's current state.
// If configuration changes are necessary, Reconcile will create or patch
// the underlying resources to match the expected configuration manifest.
func (r TerraformReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) { //nolint:gocyclo
	o := &unstructured.Unstructured{}
	o.SetGroupVersionKind(r.GVK)
	o.SetNamespace(request.Namespace)
	o.SetName(request.Name)
	log := log.WithValues(
		"namespace", o.GetNamespace(),
		"name", o.GetName(),
		"apiVersion", o.GetAPIVersion(),
		"kind", o.GetKind(),
	)
	log.V(1).Info("Reconciling")

	// Retrieve the custom resource
	err := r.Client.Get(ctx, request.NamespacedName, o)
	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	if err != nil {
		log.Error(err, "Failed to lookup custom resource")
		return reconcile.Result{}, err
	}

	manager, err := r.ManagerFactory.NewManager(o, r.OverrideValues)
	if err != nil {
		log.Error(err, "Failed to get configuration manager")
		return reconcile.Result{}, err
	}

	log = log.WithValues("configuration", manager.DeploymentName())

	// Custom resource is tagged for deletion
	if o.GetDeletionTimestamp() != nil {
		return r.deleteHandler(o, manager)
	}

	// The rest is for creation and update
	status := types.StatusFor(o)
	status.SetCondition(types.TFConfigCondition{
		Type:   types.ConditionInitialized,
		Status: types.StatusTrue,
	})

	if err := manager.Refresh(context.TODO()); err != nil {
		log.Error(err, "Failed to refresh configuration")
		status.SetCondition(types.TFConfigCondition{
			Type:    types.ConditionIrreconcilable,
			Status:  types.StatusTrue,
			Reason:  types.ReasonReconcileError,
			Message: err.Error(),
		})
		_ = r.updateResourceStatus(o, status)
		return reconcile.Result{}, err
	}
	status.RemoveCondition(types.ConditionIrreconcilable)

	// handler for CREATE operation
	if !manager.Exists() {
		return r.createHandler(o, manager)
	}

	// Add finalizer if it is not in the list
	if !contains(o.GetFinalizers(), finalizer) {
		log.V(1).Info("Adding finalizer", "finalizer", finalizer)
		controllerutil.AddFinalizer(o, finalizer)
		if err := r.updateResource(o); err != nil {
			log.Info("Failed to add CR delete finalizer")
			return reconcile.Result{}, err
		}
	}

	// handler for UPDATE operation
	if manager.IsUpdateRequired() {
		return r.updateHandler(o, manager)
	}

	// Now, no update is required ...

	// If a change is made to the CR spec that causes an apply failure, a
	// ConditionConfigFailed is added to the status conditions. If that change
	// is then reverted to its previous state, the operator will stop
	// attempting the apply and will resume reconciling. In this case, we
	// need to remove the ConditionConfigFailed because the failing apply is
	// no longer being attempted.
	status.RemoveCondition(types.ConditionConfigFailed)

	err = manager.Reconcile(context.TODO())
	if err != nil {
		log.Error(err, "Failed to reconcile configuration")
		status.SetCondition(types.TFConfigCondition{
			Type:    types.ConditionIrreconcilable,
			Status:  types.StatusTrue,
			Reason:  types.ReasonReconcileError,
			Message: err.Error(),
		})
		_ = r.updateResourceStatus(o, status)
		return reconcile.Result{}, err
	}
	status.RemoveCondition(types.ConditionIrreconcilable)

	log.Info("Reconciled configuration")
	reason := types.ReasonUpdateSuccessful
	message := ""
	status.SetCondition(types.TFConfigCondition{
		Type:    types.ConditionDeployed,
		Status:  types.StatusTrue,
		Reason:  reason,
		Message: message,
	})
	err = r.updateResourceStatus(o, status)
	return reconcile.Result{RequeueAfter: r.ReconcilePeriod}, err
}

func (r TerraformReconciler) createHandler(o *unstructured.Unstructured, manager config.Manager) (reconcile.Result, error) { //nolint:gocyclo
	// Apply value from OverrideValues map
	for k, v := range r.OverrideValues {
		r.EventRecorder.Eventf(o, "Warning", "OverrideValuesInUse",
			"Template value %q overridden to %q by operator's watches.yaml", k, v)
	}

	status := types.StatusFor(o)
	err := manager.Create(context.TODO())
	if err != nil {
		log.Error(err, "Failed in creating deployment")
		status.SetCondition(types.TFConfigCondition{
			Type:    types.ConditionConfigFailed,
			Status:  types.StatusTrue,
			Reason:  types.ReasonCreateError,
			Message: err.Error(),
		})
		_ = r.updateResourceStatus(o, status)
		return reconcile.Result{}, err
	}
	status.RemoveCondition(types.ConditionConfigFailed)

	// Add finalizer
	log.V(1).Info("Adding finalizer", "finalizer", finalizer)
	controllerutil.AddFinalizer(o, finalizer)
	if err := r.updateResource(o); err != nil {
		log.Info("Failed to add CR delete finalizer")
		return reconcile.Result{}, err
	}

	message := ""
	status.SetCondition(types.TFConfigCondition{
		Type:    types.ConditionDeployed,
		Status:  types.StatusTrue,
		Reason:  types.ReasonCreateSuccessful,
		Message: message,
	})
	// TODO(Qiming): Determine if and where to store deployed config
	err = r.updateResourceStatus(o, status)
	return reconcile.Result{RequeueAfter: r.ReconcilePeriod}, err
}

func (r TerraformReconciler) updateHandler(o *unstructured.Unstructured, manager config.Manager) (reconcile.Result, error) { //nolint:gocyclo
	for k, v := range r.OverrideValues {
		r.EventRecorder.Eventf(o, "Warning", "OverrideValuesInUse",
			"Template value %q overridden to %q by operator's watches.yaml", k, v)
	}

	// Update the deployment
	status := types.StatusFor(o)
	err := manager.Update(context.TODO())
	if err != nil {
		log.Error(err, "Failed to update deployment")
		status.SetCondition(types.TFConfigCondition{
			Type:    types.ConditionConfigFailed,
			Status:  types.StatusTrue,
			Reason:  types.ReasonUpdateError,
			Message: err.Error(),
		})
		_ = r.updateResourceStatus(o, status)
		return reconcile.Result{}, err
	}
	status.RemoveCondition(types.ConditionConfigFailed)

	// TODO(Qiming): Maybe output a diff between configs?
	status.SetCondition(types.TFConfigCondition{
		Type:    types.ConditionDeployed,
		Status:  types.StatusTrue,
		Reason:  types.ReasonUpdateSuccessful,
		Message: "Configuration change updated",
	})
	err = r.updateResourceStatus(o, status)
	return reconcile.Result{RequeueAfter: r.ReconcilePeriod}, err

}

func (r TerraformReconciler) deleteHandler(o *unstructured.Unstructured, manager config.Manager) (reconcile.Result, error) { //nolint:gocyclo
	status := types.StatusFor(o)

	// finalizers list is empty, done
	if !contains(o.GetFinalizers(), finalizer) {
		log.Info("Resource is deleted, skipping reconciliation")
		return reconcile.Result{}, nil
	}

	err := manager.Delete(context.TODO())

	// Failed to delete: update status and return
	if err != nil {
		log.Error(err, "Failed to destroy deployment")
		status.SetCondition(types.TFConfigCondition{
			Type:    types.ConditionConfigFailed,
			Status:  types.StatusTrue,
			Reason:  types.ReasonDeleteError,
			Message: err.Error(),
		})
		_ = r.updateResourceStatus(o, status)
		return reconcile.Result{}, err
	}
	// TODO(Qiming): This looks redundant
	status.RemoveCondition(types.ConditionConfigFailed)

	// Deletion succeeded: update status and then remove finalizer
	log.Info("Deployment destroyed")
	status.SetCondition(types.TFConfigCondition{
		Type:   types.ConditionDeployed,
		Status: types.StatusFalse,
		Reason: types.ReasonDeleteSuccessful,
	})
	status.DeployedConfig = nil
	if err := r.updateResourceStatus(o, status); err != nil {
		log.Info("Failed to update CR status")
		return reconcile.Result{}, err
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(o, finalizer)
	if err := r.updateResource(o); err != nil {
		log.Info("Failed to remove CR finalizer")
		return reconcile.Result{}, err
	}

	// Since the client is hitting a cache, waiting for the deletion here
	// will guarantee that the next reconciliation will see that the CR
	// has been deleted and that there's nothing left to do.
	if err := r.waitForDeletion(o); err != nil {
		log.Info("Failed waiting for CR deletion")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// returns the boolean representation of the annotation string
// will return false if annotation is not set
func hasForceUpdateAnnotation(o *unstructured.Unstructured) bool {
	const updateForceAnnotation = "terraform.sdk.operatorframework.io/force-update"
	force := o.GetAnnotations()[updateForceAnnotation]
	if force == "" {
		return false
	}
	value := false
	if i, err := strconv.ParseBool(force); err != nil {
		log.Info("Could not parse annotation as a boolean",
			"annotation", updateForceAnnotation, "value informed", force)
	} else {
		value = i
	}
	return value
}

func (r TerraformReconciler) updateResource(o client.Object) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.Client.Update(context.TODO(), o)
	})
}

func (r TerraformReconciler) updateResourceStatus(o *unstructured.Unstructured, status *types.TFConfigStatus) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		o.Object["status"] = status
		return r.Client.Status().Update(context.TODO(), o)
	})
}

// waitForDeletion waits for the specified object to disappear
func (r TerraformReconciler) waitForDeletion(o client.Object) error {
	key := client.ObjectKeyFromObject(o)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return wait.PollImmediateUntil(time.Millisecond*20, func() (bool, error) {
		err := r.Client.Get(ctx, key, o)
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, ctx.Done())
}

func contains(l []string, s string) bool {
	for _, elem := range l {
		if elem == s {
			return true
		}
	}
	return false
}
