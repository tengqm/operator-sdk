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

package templates

import (
	"errors"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"

	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/constants"
)

var (
	// Ensure that Dockerfile implements the Template interface.
	_ file.Template = &Dockerfile{}
)

// Dockerfile scaffolds a Dockerfile for building a manager container.
// The TemplateMixin is a mixin required for all template builders. The
// `GetPath`, `GetIfExistsAction`, `GetBody()` methods is implemented in the
// default mixin; this struct only needs to implement `SetTemplateDefaults`.
type Dockerfile struct {
	file.TemplateMixin

	// OperatorVersion is the version of the Dockerfile's base image.
	OperatorVersion string

	// TODO: Determine if this is necessary
	TemplateDir string
}

// SetTemplateDefaults implements input.Template
// This method sets the default values for templates.
func (f *Dockerfile) SetTemplateDefaults() error {
	// `Path` is returned by the `GetPath()` method of `PathMixin` which is
	// nested in TemplateMixin.
	if f.Path == "" {
		f.Path = "Dockerfile"
	}

	// `TemplateBody` is defined in the mixin, will be returned by `GetBody`
	f.TemplateBody = dockerfileTemplate

	if f.OperatorVersion == "" {
		return errors.New("terraform-operator version is required in scaffold")
	}

	f.TemplateDir = constants.TemplateDir

	return nil
}

// TODO: Add RUN command in the container to install terraform providers
const dockerfileTemplate = `FROM quay.io/tengqm/terraform-operator:{{ .OperatorVersion }}

COPY watches.yaml ${HOME}/watches.yaml
`
