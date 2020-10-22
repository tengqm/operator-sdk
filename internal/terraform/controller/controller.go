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
	"fmt"
	"strings"
	"time"

	"github.com/operator-framework/operator-lib/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/operator-framework/operator-sdk/internal/terraform/config"
)

var log = logf.Log.WithName("terraform-controller")

// Options contains the necessary values to create a new controller that
// manages Terraform configurations in a particular namespace based on a GVK watch.
type Options struct {
	Namespace               string
	GVK                     schema.GroupVersionKind
	ManagerFactory          config.ManagerFactory
	ReconcilePeriod         time.Duration
	WatchDependentResources bool
	OverrideValues          map[string]string
	MaxConcurrentReconciles int
}

// Add creates a new Terraform operator controller and adds it to the manager
func Add(mgr manager.Manager, options Options) error {
	controllerName := fmt.Sprintf("%v-controller", strings.ToLower(options.GVK.Kind))

	r := &TerraformReconciler{
		Client:          mgr.GetClient(),
		GVK:             options.GVK,
		EventRecorder:   mgr.GetEventRecorderFor(controllerName),
		ReconcilePeriod: options.ReconcilePeriod,
		ManagerFactory:  options.ManagerFactory,
		OverrideValues:  options.OverrideValues,
	}

	// Register the GVK with the scheme
	mgr.GetScheme().AddKnownTypeWithName(options.GVK, &unstructured.Unstructured{})
	metav1.AddToGroupVersion(mgr.GetScheme(), options.GVK.GroupVersion())

	c, err := controller.New(controllerName, mgr, controller.Options{
		Reconciler:              r,
		MaxConcurrentReconciles: options.MaxConcurrentReconciles,
	})
	if err != nil {
		return err
	}

	o := &unstructured.Unstructured{}
	o.SetGroupVersionKind(options.GVK)
	if err := c.Watch(&source.Kind{Type: o}, &handler.InstrumentedEnqueueRequestForObject{}); err != nil {
		return err
	}

	log.Info("Watching resource", "apiVersion", options.GVK.GroupVersion(),
		"kind", options.GVK.Kind, "namespace", options.Namespace,
		"reconcilePeriod", options.ReconcilePeriod.String())
	return nil
}
