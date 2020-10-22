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

package configuration

import (
	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
)

// Main scaffolds a manifest for main Terraform template.
// TemplateMixin is the mixin that is embedded in Template builders.
// ResourceMixin provides templates with a injectable resource field.
type Main struct {
	file.TemplateMixin
	file.ResourceMixin

	// TODO: Remove this
	ProviderName string
	// ProviderFullName is the full name of the terraform provider to use
	// TODO: Add support to more than one provider.
	ProviderFullName string

	// A list of provider options to add.
	// TODO: Revise this into a map.
	ProviderOptions []string
}

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface
func (f *Main) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "%[kind].tf"
		// TODO: the kind must be lower
		f.Path = f.Resource.Replacer().Replace(f.Path)
	}
	if f.Path == "" {
		f.Path = "main.tf"
	}

	// TemplateBody is defined in the TemplateMixin, will be returned by `GetBody()`.
	f.TemplateBody = mainTemplate

	return nil
}

const mainTemplate = `
{{ if .ProviderName }}
terraform {
  required_providers {
    {{ .ProviderName }} = {
	  source = "{{ .ProviderFullName }}"
	}
  }
}

# Provider options
provider "{{ .ProviderName }}" {
  {{ range .ProviderOptions }}
  {{ . }} = var.{{ . }}
  {{ end }}
}
{{ end }}
`
