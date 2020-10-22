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

package kdefault

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
)

// Kustomization scaffolds the Kustomization file for the default patch file.
// TemplateMixin is the mixin that should be embedded in Template builders.
// ProjectNameMixin provides templates with an injectable project name field.
type Kustomization struct {
	file.TemplateMixin
	file.ProjectNameMixin
}

var (
	// Ensure that Kustomization  implements the Template interface.
	_ file.Template = &Kustomization{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface
func (f *Kustomization) SetTemplateDefaults() error {
	// `Path` is returned by the `GetPath()` method of `PathMixin` which is nested in TemplateMixin.
	if f.Path == "" {
		f.Path = filepath.Join("config", "default", "kustomization.yaml")
	}

	// IfExistsAction specifies the behavior when creating a file that already exists.
	f.IfExistsAction = file.Error

	// `TemplateBody` is defined in the TemplateMixin, will be returned by `GetBody()`.
	f.TemplateBody = kustomizeTemplate
	return nil
}

const kustomizeTemplate = `# Adds namespace to all resources.
namespace: {{ .ProjectName }}-system

# Value of this field is prepended to the names of all resources, e.g. a
# Deployment named "wordpress" becomes "alices-wordpress". Note that it should
# also match with the prefix (text before '-') of the namespace field above.
namePrefix: {{ .ProjectName }}-

# Labels to add to all resources and selectors.
#commonLabels:
#  someName: someValue

bases:
  - ../crd
  - ../rbac
  - ../manager
  # [PROMETHEUS] To enable prometheus monitor, uncomment all sections with 'PROMETHEUS'.
  #- ../prometheus

patchesStrategicMerge:
  # Protect the "/metrics" endpoint by putting it behind "auth".
  # If you want the controller-manager to expose the "/metrics" endpoint without
  # any authentication or authorization, please comment the following line.
  - manager_auth_proxy_patch.yaml
`
