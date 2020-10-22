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
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/kubebuilder/v2/pkg/model/config"
	"sigs.k8s.io/kubebuilder/v2/pkg/plugin"

	"github.com/operator-framework/operator-sdk/internal/kubebuilder/cmdutil"
	"github.com/operator-framework/operator-sdk/internal/plugins/manifests"
	"github.com/operator-framework/operator-sdk/internal/plugins/scorecard"
	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/scaffolds"
)

type initSubcommand struct {
	// Unmarshalled representation of the configuration file.
	config *config.Config

	// The struct used for the `create api` plugin.
	apiPlugin createAPISubcommand

	// If true, run the `create api` plugin.
	doCreateAPI bool

	// Global Provider (and version) to use for the project
	ProviderName string

	// For help text.
	// TODO: Do we need this at all?
	commandName string
}

var (
	// The initSubcommand struct must implement the KB pkg/plugin/Init interface
	_ plugin.InitSubcommand = &initSubcommand{}

	// The struct must implement the KB pkg/plugin/RunOptions interface
	_ cmdutil.RunOptions = &initSubcommand{}
)

// UpdateContext updates a Context with command-specific help text, like description and examples.
// Required by KB pkg/plugin.GenericSubcommand.
func (p *initSubcommand) UpdateContext(ctx *plugin.Context) {
	ctx.Description = `
Initializes a new Terraform-based operator project.

Writes the following files

 - a PROJECT file with the domain and project layout configuration
 - a Makefile that provides an interface for building and managing the operator
 - Kubernetes manifests and kustomize configuration
 - a watches.yaml file that defines the mapping between APIs and Terraform configurations 

Optionally creates a new API, using the same flags as "create api"
`
	ctx.Examples = fmt.Sprintf(`
  # Scaffold a project with no API
  $ %s init --plugins=%s \
      --domain=my.domain \
	  --provider=ibm-cloud/ibm

  # Invokes "create api"
  $ %s init --plugins=%s \
      --domain=my.domain \
      --group=apps --version=v1alpha1 --kind=AppService

  $ %s init --plugins=%s \
      --domain=my.domain \
      --group=apps --version=v1alpha1 --kind=AppService \
      --terraform-config=myrepo/app

  $ %s init --plugins=%s \
      --domain=my.domain \
      --group=apps --version=v1alpha1 --kind=AppService \
      --provider=ibm-cloud/ibm
`,
		ctx.CommandName, pluginKey,
		ctx.CommandName, pluginKey,
		ctx.CommandName, pluginKey,
		ctx.CommandName, pluginKey,
	)
	p.commandName = ctx.CommandName
}

// BindFlags binds the plugin's flags to the CLI. This allows each plugin to define its own
// command line flags for the subcommand.
// Required by KB pkg/plugin.GenericSubcommand.
func (p *initSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.SortFlags = false

	fs.StringVar(&p.config.Domain, "domain", "my.domain", "domain for groups")
	fs.StringVar(&p.config.ProjectName, "project-name", "", "name of this project, defaults to the current directory name")
	fs.StringVar(&p.ProviderName, "provider", "", "name of terraform provider to use")

	// Bind `create api` flags
	p.apiPlugin.BindFlags(fs)
}

// InjectConfig passes a config to the plugin. The plugin may modify the config.
// Initializing, loading, and saving the config is managed by the cli package.
// Required by KB pkg/plugin.GenericSubcommand.
func (p *initSubcommand) InjectConfig(c *config.Config) {
	// The Layout field is used to capture the plugin name
	c.Layout = pluginKey

	p.config = c
	// p.config.Plugins = make(map[string]interface{})
	p.config.EncodePluginConfig("provider", p.ProviderName)

	// Replicate the config to the `create api` plugin
	p.apiPlugin.config = c
}

// Run runs the subcommand.
// Required by KB pkg/plugin.GenericSubcommand.
func (p *initSubcommand) Run() error {
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
func (p *initSubcommand) runPhase2() error {
	if err := manifests.RunInit(p.config); err != nil {
		return err
	}
	if err := scorecard.RunInit(p.config); err != nil {
		return err
	}

	if p.doCreateAPI {
		if err := p.apiPlugin.runPhase2(); err != nil {
			return err
		}
	}

	return nil
}

// Validate verifies that the command can be run. It is the step 1 as per required
// by the internal/kubebuilder/cmdutil.RunOptions interface.
func (p *initSubcommand) Validate() error {
	// Default project name to directory name if not specified
	if p.config.ProjectName == "" {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %v", err)
		}
		p.config.ProjectName = strings.ToLower(filepath.Base(dir))
	}
	// Check if the project name is a valid k8s namespace (DNS 1123 label).
	if err := validation.IsDNS1123Label(p.config.ProjectName); err != nil {
		return fmt.Errorf("project name '%s' is invalid: %v", p.config.ProjectName, err)
	}

	// Invoke Validate on the `create api` plugin if GVK is specified or
	// generateTemplate is true
	// TODO: fix this logic
	defaultOpts := scaffolds.CreateOptions{}
	if !p.apiPlugin.createOptions.GVK.Empty() || p.apiPlugin.createOptions != defaultOpts {
		p.doCreateAPI = true
		return p.apiPlugin.Validate()
	}
	return nil
}

// GetScaffolder creates the Scaffolder instance. It is the step 2 as per required
// by the internal/kubebuilder/cmdutil.RunOptions interface.
func (p *initSubcommand) GetScaffolder() (cmdutil.Scaffolder, error) {
	var apiScaffolder cmdutil.Scaffolder
	var err error
	// Get Scaffolder for `create api` plugin first if needed
	if p.doCreateAPI {
		apiScaffolder, err = p.apiPlugin.GetScaffolder()
		if err != nil {
			return nil, err
		}
	}
	return scaffolds.NewInitScaffolder(p.config, apiScaffolder), nil
}

// PostScaffold finishes the Scaffold command. It is the step 4 as per required
// by the internal/kubebuilder/cmdutil.RunOptions interface.
// Print a message if `create api` plugin is not invoked.
func (p *initSubcommand) PostScaffold() error {
	if !p.doCreateAPI {
		fmt.Printf("Next: define a resource with:\n$ %s create api\n", p.commandName)
	}

	return nil
}
