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

package samples

import (
	"path/filepath"
	"strings"
	"text/template"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
)

// CustomResource scaffolds a sample manifest for a CustomResource (CR).
// TemplateMixin is the mixin that is embedded in Template builders.
// ResourceMixin provides templates with a injectable resource field.
type CustomResource struct {
	file.TemplateMixin
	file.ResourceMixin

	// Spec is the serialized spec for the custom resource. It can be specified
	// when scaffolding this sample.
	Spec string
}

var (
	// This ensure the CustomResource implements the KB pkg/model/file.Template interface.
	_ file.Template = &CustomResource{}

	// This ensure the CustomResource implements the KB pkg/model/file.UseCustomFuncMap interface
	// which allows a template to use a custom FuncMap instead of the default one.
	_ file.UseCustomFuncMap = &CustomResource{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface.
func (f *CustomResource) SetTemplateDefaults() error {
	// Path is returned by the `GetPath()` method of PathMixin nested in TemplateMixin.
	if f.Path == "" {
		f.Path = filepath.Join("config", "samples", "%[group]_%[version]_%[kind].yaml")
	}
	f.Path = f.Resource.Replacer().Replace(f.Path)

	f.IfExistsAction = file.Error

	// Set the default if it is empty
	if len(f.Spec) == 0 {
		f.Spec = defaultSpecTemplate
	}

	// TemplateBody is defined in the TemplateMixin, will be returned by `GetBody`
	f.TemplateBody = crSampleTemplate
	return nil
}

// GetFuncMap implements file.UseCustomFuncMap
func (f *CustomResource) GetFuncMap() template.FuncMap {
	fm := file.DefaultFuncMap()
	fm["indent"] = func(spaces int, v string) string {
		padding := strings.Repeat(" ", spaces)
		return padding + strings.Replace(v, "\n", "\n"+padding, -1)
	}
	return fm
}

const defaultSpecTemplate = `foo: bar`

const crSampleTemplate = `apiVersion: {{ .Resource.Domain }}/{{ .Resource.Version }}
kind: {{ .Resource.Kind }}
metadata:
  name: {{ lower .Resource.Kind }}-sample
spec:
{{ .Spec | indent 2 }}
`
