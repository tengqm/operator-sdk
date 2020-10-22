/*
Copyright 2019 The Kubernetes Authors.
Modifications copyright 2020 The Operator-SDK Authors

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
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kubebuilder/v2/pkg/model"
	"sigs.k8s.io/kubebuilder/v2/pkg/model/config"
	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
	"sigs.k8s.io/kubebuilder/v2/pkg/model/resource"

	"github.com/operator-framework/operator-sdk/internal/kubebuilder/cmdutil"
	"github.com/operator-framework/operator-sdk/internal/kubebuilder/machinery"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates/config/crd"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates/config/rbac"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates/config/samples"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds/templates/configuration"
)

type CreateOptions struct {
	GVK schema.GroupVersionKind

	// Provider (and version) to use for the API resource
	ResourceProvider string

	// TODO: Fetch template from given filepath or URL
	UseTemplate string
}

// TODO: This is identical to the internal/plugins/terraform/v1.createAPIPlugin
// Should we remove it?
type apiScaffolder struct {
	config *config.Config
	opts   CreateOptions
}

var (
	// The apiScaffolder must implement the KB pkg/plugins.Scaffolder interface
	_ cmdutil.Scaffolder = &apiScaffolder{}
)

// NewCreateAPIScaffolder returns a new Scaffolder for project initialization operations
func NewCreateAPIScaffolder(config *config.Config, opts CreateOptions) cmdutil.Scaffolder {
	return &apiScaffolder{
		config: config,
		opts:   opts,
	}
}

func (s *apiScaffolder) newUniverse(r *resource.Resource) *model.Universe {
	return model.NewUniverse(
		model.WithConfig(s.config),
		model.WithResource(r),
	)
}

// Scaffold performs the scaffolding.
// Required by kubebuilder/pkg/plugin/cmdutil.Scaffolder interface.
func (s *apiScaffolder) Scaffold() error {
	return s.scaffold()
}

// TODO: This is unnecessary. Can be unfolded into the function above.
func (s *apiScaffolder) scaffold() error {

	// TODO: add Plurals
	resourceOptions := resource.Options{
		Group:   s.opts.GVK.Group,
		Version: s.opts.GVK.Version,
		Kind:    s.opts.GVK.Kind,
	}

	if s.config.HasResource(resourceOptions.GVK()) {
		return errors.New("the API resource already exists")
	}

	// Check that the provided group can be added to the project
	if !s.config.MultiGroup && len(s.config.Resources) != 0 && !s.config.HasGroup(resourceOptions.Group) {
		return errors.New("multiple groups are not allowed by default, to enable multi-group set 'multigroup: true' in your PROJECT file")
	}

	resource := resourceOptions.NewResource(s.config, true)
	s.config.UpdateResources(resource.GVK())

	var builders []file.Builder
	builders = append(builders,
		&rbac.CRDViewerRole{},
		&rbac.CRDEditorRole{},
		&rbac.ManagerRoleUpdater{},

		&crd.CRD{},
		&crd.Kustomization{},

		// TODO: Allow spec to be scaffolded.
		&samples.CustomResource{},
		&templates.WatchesUpdater{
			UseTemplate: s.opts.UseTemplate,
		},
	)

	// generate Configuration
	var providerName string
	resourceProvider := s.opts.ResourceProvider
	if resourceProvider != "" {
		parts := strings.Split(resourceProvider, "/")
		if len(parts) > 1 {
			providerName = parts[1]
		} else {
			providerName = parts[0]
		}
	} else {
		providerName = ""
	}

	// TODO: How to get provider options
	providerOptions := []string{}
	builders = append(builders,
		&configuration.Main{
			ProviderName:     providerName,
			ProviderFullName: resourceProvider,
			ProviderOptions:  providerOptions,
		},
	)
	return machinery.NewScaffold().Execute(
		s.newUniverse(resource),
		builders...,
	)
}
