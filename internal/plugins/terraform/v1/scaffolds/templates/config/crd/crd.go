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

package crd

import (
	"fmt"
	"path/filepath"

	"github.com/kr/text"
	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
)

// CRD scaffolds a manifest for CRD sample.
// TemplateMixin is the mixin that is embedded in Template builders.
// ResourceMixin provides templates with a injectable resource field.
type CRD struct {
	file.TemplateMixin
	file.ResourceMixin
}

var (
	// This enusre the CRD implements the KB pkg/model/file.Template interface.
	_ file.Template = &CRD{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface
func (f *CRD) SetTemplateDefaults() error {
	// Path is returned by the `GetPath()` method of PathMixin nested in TemplateMixin.
	if f.Path == "" {
		f.Path = filepath.Join("config", "crd", "bases", fmt.Sprintf("%s_%%[plural].yaml", f.Resource.Domain))
	}
	f.Path = f.Resource.Replacer().Replace(f.Path)

	// IfExistsAction specifies the behavior when creating a file that already exists.
	f.IfExistsAction = file.Error

	// TemplateBody is defined in the TemplateMixin, will be returned by `GetBody()`.
	// The indentations are to make sure the schema body fixs in the template.
	f.TemplateBody = fmt.Sprintf(crdTemplate,
		text.Indent(openAPIV3SchemaTemplate, "        "),
	)
	return nil
}

const crdTemplate = `---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: {{ .Resource.Plural }}.{{ .Resource.Domain }}
spec:
  group: {{ .Resource.Domain }}
  names:
    kind: {{ .Resource.Kind }}
    listKind: {{ .Resource.Kind }}List
    plural: {{ .Resource.Plural }}
    singular: {{ .Resource.Kind | lower }}
  scope: Namespaced
  versions:
    - name: {{ .Resource.Version }}
      schema:
%s
      served: true
      storage: true
      subresources:
        status: {}
`

// This field may get customized in the future.
const openAPIV3SchemaTemplate = `openAPIV3Schema:
  description: {{ .Resource.Kind }} is the Schema for the {{ .Resource.Plural }} API
  properties:
    apiVersion:
      description: The version of the schema for the object representation.
      type: string
    kind:
      description: A string value representing the REST resource this object represents.
      type: string
    metadata:
      type: object
    spec:
      description: Spec defines the desired state of {{ .Resource.Kind }}
      type: object
      x-kubernetes-preserve-unknown-fields: true
    status:
      description: Status defines the observed state of {{ .Resource.Kind }}
      type: object
      x-kubernetes-preserve-unknown-fields: true
  type: object
`
