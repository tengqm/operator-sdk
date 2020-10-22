// Copyright 2020 The Operator-SDK Authors
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

package terraform

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/v2/pkg/model/config"
	"sigs.k8s.io/kubebuilder/v2/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v2/pkg/plugin"

	"github.com/operator-framework/operator-sdk/internal/kubebuilder/cmdutil"
	"github.com/operator-framework/operator-sdk/internal/plugins/manifests"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds"
)

const (
	groupFlag      = "group"
	versionFlag    = "version"
	kindFlag       = "kind"
	crdVersionFlag = "crd-version"
	providerFlag   = "resource-provider"
	configFlag     = "terraform-config"

	crdVersionV1 = "v1"
	// We skip the v1beta1 version of CRD
	// crdVersionV1beta1 = "v1beta1"
)

type createAPISubcommand struct {
	// Unmarshalled representation of the configuration file.
	config *config.Config

	// Terraform specific options for create.
	createOptions scaffolds.CreateOptions
}

var (
	// The struct must implement the KB pkg/plugin.CreateAPISubcommand interface.
	_ plugin.CreateAPISubcommand = &createAPISubcommand{}
	// The struct must implement the KB RunOptions interface
	_ cmdutil.RunOptions = &createAPISubcommand{}
)

// UpdateContext updates a Context with command-specific help text, like description and examples.
// Required by KB pkg/plugin.GenericSubcommand.
func (p *createAPISubcommand) UpdateContext(ctx *plugin.Context) {
	ctx.Description = `Scaffold a Kubernetes API that is backed by a Terraform configuration.
`
	ctx.Examples = fmt.Sprintf(` $ %s create api \
      --group=apps --version=v1alpha1 \
      --kind=AppService

  $ %s create api \
      --group=apps --version=v1alpha1 \
      --kind=AppService \
      --terraform-config=myrepo/app

  $ %s create api \
      --group=apps --version=v1alpha1 \
      --kind=AppService
      --resource-provider=ibm-cloud/ibm
`,
		ctx.CommandName,
		ctx.CommandName,
		ctx.CommandName,
	)
}

// BindFlags binds the plugin's flags to the CLI. This allows each plugin to define its own
// command line flags for the subcommand.
// Required by KB pkg/plugin.GenericSubcommand.
func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.SortFlags = false

	fs.StringVar(&p.createOptions.GVK.Group, groupFlag, "", "resource group")
	fs.StringVar(&p.createOptions.GVK.Version, versionFlag, "", "resource version")
	fs.StringVar(&p.createOptions.GVK.Kind, kindFlag, "", "resource kind")
	fs.StringVar(&p.createOptions.ResourceProvider, providerFlag, "", "terraform provider to use for this resource")
	fs.StringVar(&p.createOptions.UseTemplate, configFlag, "", "location of terraform config")
}

// InjectConfig passes a config to the plugin. The plugin may modify the config.
// Initializing, loading, and saving the config is managed by the cli package.
// Required by KB pkg/plugin.GenericSubcommand.
func (p *createAPISubcommand) InjectConfig(c *config.Config) {
	p.config = c
}

// Run runs the subcommand.
// Required by KB pkg/plugin.GenericSubcommand.
func (p *createAPISubcommand) Run() error {
	if err := cmdutil.Run(p); err != nil {
		return err
	}

	// Run SDK phase 2 plugins.
	if err := p.runPhase2(); err != nil {
		return err
	}

	return nil
}

// SDK phase 2 plugins.
func (p *createAPISubcommand) runPhase2() error {
	gvk := p.createOptions.GVK
	// TODO: This conversions of GVK is stupid
	return manifests.RunCreateAPI(p.config, config.GVK{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})
}

// Validate verifies that the command can be run. It is the step 1 as per required
// by the internal/kubebuilder/cmdutil.RunOptions interface.
func (p *createAPISubcommand) Validate() error {
	// if p.createOptions.CRDVersion != crdVersionV1 {
	// 	return fmt.Errorf("value of --%s must be %q", crdVersionFlag, crdVersionV1)
	// }
	// TODO: Combine the followings into a single error
	group := strings.TrimSpace(p.createOptions.GVK.Group)
	version := strings.TrimSpace(p.createOptions.GVK.Version)
	kind := strings.TrimSpace(p.createOptions.GVK.Kind)
	if len(group) == 0 {
		return fmt.Errorf("value of --%s must not be empty", groupFlag)
	}
	if len(version) == 0 {
		return fmt.Errorf("value of --%s must not be empty", versionFlag)
	}
	if len(kind) == 0 {
		return fmt.Errorf("value of --%s must not be empty", kindFlag)
	}

	// Validate the resource.
	// TODO: assign Plurals to the Options struct
	r := resource.Options{
		Namespaced: true,
		Group:      group,
		Version:    version,
		Kind:       kind,
	}
	if err := r.Validate(); err != nil {
		return err
	}

	return nil
}

// GetScaffolder creates the Scaffolder instance. It is the step 2 as per required
// by the internal/kubebuilder/cmdutil.RunOptions interface.
func (p *createAPISubcommand) GetScaffolder() (cmdutil.Scaffolder, error) {
	return scaffolds.NewCreateAPIScaffolder(p.config, p.createOptions), nil
}

// PostScaffold finishes the Scaffold command. It is the step 4 as per required
// by the internal/kubebuilder/cmdutil.RunOptions interface.
// For Terraform, we do nothing here.
func (p *createAPISubcommand) PostScaffold() error {
	return nil
}
