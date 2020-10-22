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
	"context"
	"fmt"
	"os/exec"

	"github.com/operator-framework/operator-sdk/internal/terraform/internal/types"
)

// Manager manages a TF Deployment. It can create, update, reconcile,
// and delete a TF Deployment.
type Manager interface {
	DeploymentName() string
	Exists() bool
	IsUpdateRequired() bool
	Refresh(context.Context) error

	Create(context.Context) error
	Update(context.Context) error
	Reconcile(context.Context) error
	Delete(context.Context) error
}

type manager struct {
	deployName string
	namespace  string

	values map[string]interface{}
	status *types.TFConfigStatus

	exists           bool
	isUpdateRequired bool
}

// DeploymentName returns the name of the release.
func (m manager) DeploymentName() string {
	return m.deployName
}

func (m manager) Exists() bool {
	return m.exists
}

func (m manager) IsUpdateRequired() bool {
	return m.isUpdateRequired
}

// Refresh ensures the Terraform storage backend is in sync with the status of
// the custom resource.
func (m *manager) Refresh(ctx context.Context) error {
	// TODO(Qiming): Get history for this deployment when supports to backends
	// are added.
	// TODO(Qiming): Call terraform code rather than run command line.
	// TODO(Qiming): run terraform plan -detailed-exitcode to check if update
	// is required.
	err := exec.Command("terraform", "state", "list").Run()
	if err != nil {
		m.exists = false
	} else {
		m.exists = true
	}

	cmd := exec.Command("terraform", "plan", "--detailed-exitcode")
	err = cmd.Run()
	if err != nil {
		code := cmd.ProcessState.ExitCode()
		if code == 2 {
			// terraform identied diffs to apply
			m.isUpdateRequired = true
		} else {
			// code == -1: can be returned from go lib
			// code == 1: means the terrform command failed to check the status
			return fmt.Errorf("failed to check deployment status: %w", err)
		}
	} else {
		m.isUpdateRequired = false
	}

	return nil
}

// Create creates a Terraform deployment.
func (m manager) Create(ctx context.Context) error {
	err := exec.Command("terraform", "apply", "-auto-approve").Run()
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}
	// TODO(Qiming): return a deployment ID when backend is supported
	return nil
}

// Update performs a Terraform configuration "update".
func (m manager) Update(ctx context.Context) error {
	err := exec.Command("terraform", "apply", "-auto-approve").Run()
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}
	// TODO(Qiming): Check if we need different logic for update
	return nil
}

// Reconcile creates or patches resources as necessary to match the
// deployed release's manifest.
func (m manager) Reconcile(ctx context.Context) error {
	// TODO(Qiming): This function updates the custom resource so that it
	// matches the current configuration
	return nil
}

// Delete performs a Terraform destroy operation.
func (m manager) Delete(ctx context.Context) error {
	err := exec.Command("terraform", "destroy", "-auto-approve").Run()
	// TODO(Qiming): There is a possibility that the configuration is not yet
	// deployed.
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	return nil
}
