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

package config

import (
	"fmt"

	"helm.sh/helm/v3/pkg/strvals"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	crmanager "sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/operator-framework/operator-sdk/internal/terraform/internal/types"
)

// ManagerFactory creates Managers that are specific to custom resources. It is
// used by the TerraformReconciler during resource reconciliation, and it
// improves decoupling between reconciliation logic and the Terraform backend
// components used to manage deployments.
type ManagerFactory interface {
	NewManager(r *unstructured.Unstructured, overrideValues map[string]string) (Manager, error)
}

type managerFactory struct {
	mgr         crmanager.Manager
	templateDir string
}

// NewManagerFactory returns a new Terraform manager factory capable of
// creating and deleting configuration managers.
func NewManagerFactory(mgr crmanager.Manager, templateDir string) ManagerFactory {
	return &managerFactory{mgr, templateDir}
}

func (f managerFactory) NewManager(cr *unstructured.Unstructured, overrideValues map[string]string) (Manager, error) {
	deployName, err := getDeployName(cr)
	if err != nil {
		return nil, fmt.Errorf("failed to get Terraform deployment name: %w", err)
	}

	crValues, ok := cr.Object["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get spec: expected map")
	}

	expOverrides, err := parseOverrides(overrideValues)
	if err != nil {
		return nil, fmt.Errorf("failed to parse override values: %w", err)
	}
	values := mergeMaps(crValues, expOverrides)

	return &manager{
		deployName: deployName,
		namespace:  cr.GetNamespace(),
		values:     values,
		status:     types.StatusFor(cr),
	}, nil
}

// getDeployName returns the deployment name.
// TODO(Qiming): when support to Terraform backend is added which means we can
// have a history of states, this function may need to be changed.
func getDeployName(cr *unstructured.Unstructured) (string, error) {
	// return the CR name.
	deploymentName := cr.GetName()
	return deploymentName, nil
}

func parseOverrides(in map[string]string) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	for k, v := range in {
		val := fmt.Sprintf("%s=%s", k, v)
		if err := strvals.ParseIntoString(val, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
