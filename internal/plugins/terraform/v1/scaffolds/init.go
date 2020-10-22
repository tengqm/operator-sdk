/*
Copyright 2020 The Operator-SDK Authors

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

package scaffolds

import (
	"strings"

	"sigs.k8s.io/kubebuilder/v2/pkg/model"
	"sigs.k8s.io/kubebuilder/v2/pkg/model/config"

	"github.com/operator-framework/operator-sdk/internal/kubebuilder/cmdutil"
	"github.com/operator-framework/operator-sdk/internal/kubebuilder/machinery"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/constants"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates/config/kdefault"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates/config/manager"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates/config/prometheus"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates/config/rbac"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates/configuration"
	"github.com/operator-framework/operator-sdk/internal/version"
)

// operatorVersion is set to the version of terraform-operator at compile-time.
// var operatorVersion = version.ImageVersion

type initScaffolder struct {
	config *config.Config

	// If this field is set, the scaffolder will invoke the `create api` scaffolder
	// when the interafce function Scaffold is invoked.
	apiScaffolder cmdutil.Scaffolder
}

// Ensure initScaffolder implements the interface for creating files.
var _ cmdutil.Scaffolder = &initScaffolder{}

// NewInitScaffolder returns a new Scaffolder for project initialization operations
func NewInitScaffolder(config *config.Config, apiScaffolder cmdutil.Scaffolder) cmdutil.Scaffolder {
	return &initScaffolder{
		config:        config,
		apiScaffolder: apiScaffolder,
	}
}

func (s *initScaffolder) newUniverse() *model.Universe {
	return model.NewUniverse(
		model.WithConfig(s.config),
	)
}

// Scaffold implements Scaffolder
// This method scaffolds the various files and then performs `create api` operation
// if requested.
func (s *initScaffolder) Scaffold() error {
	if err := s.scaffold(); err != nil {
		return err
	}

	if s.apiScaffolder != nil {
		return s.apiScaffolder.Scaffold()
	}
	return nil
}

func (s *initScaffolder) scaffold() error {
	var providerName string
	err := s.config.DecodePluginConfig("provider", providerName)
	if err != nil {
		// TODO: should return error?
		providerName = ""
	} else {
		parts := strings.Split(providerName, "/")
		if len(parts) > 1 {
			providerName = parts[1]
		} else {
			providerName = parts[0]
		}
	}

	// invoke the interface function Execute() to write files to disk
	return machinery.NewScaffold().Execute(
		s.newUniverse(),
		&templates.Dockerfile{
			OperatorVersion: version.ImageVersion,
		},
		&templates.Makefile{
			OperatorVersion: version.ImageVersion,
		},
		&templates.GitIgnore{},
		&templates.Watches{},

		&rbac.AuthProxyRole{},
		&rbac.AuthProxyRoleBinding{},
		&rbac.AuthProxyService{},
		&rbac.ClientClusterRole{},
		&rbac.Kustomization{},
		&rbac.LeaderElectionRole{},
		&rbac.LeaderElectionRoleBinding{},
		&rbac.ManagerRole{},
		&rbac.ManagerRoleBinding{},

		&manager.Manager{
			Image:        constants.ImageName,
			ProviderName: providerName,
		},
		&manager.Kustomization{},

		&prometheus.Kustomization{},
		&prometheus.ServiceMonitor{},

		&kdefault.Kustomization{},
		&kdefault.AuthProxyPatch{},

		&configuration.Placeholder{},
	)
}
