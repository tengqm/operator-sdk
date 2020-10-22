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

package rbac

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
)

// CRDViewerRole scaffolds the `config/rbac/<kind>_viewer_role.yaml`.
// TemplateMixin is the mixin that is embedded in Template builders.
// ProjectNameMixin provides templates with a injectable ProjectName field.
type CRDViewerRole struct {
	file.TemplateMixin
	file.ResourceMixin
}

var (
	// This ensure the CRDViewerRole implements the KB pkg/model/file.Template interface.
	_ file.Template = &CRDViewerRole{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface.
func (f *CRDViewerRole) SetTemplateDefaults() error {
	// `Path` is returned by the `GetPath()` method of `PathMixin` nested in `TemplateMixin`.
	if f.Path == "" {
		f.Path = filepath.Join("config", "rbac", "%[kind]_viewer_role.yaml")
	}
	f.Path = f.Resource.Replacer().Replace(f.Path)

	// `TemplateBody` is defined in the TemplateMixin, will be returned by `GetBody`
	f.TemplateBody = crdRoleViewerTemplate

	return nil
}

const crdRoleViewerTemplate = `# permissions for end users to view {{ .Resource.Plural }}.
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ lower .Resource.Kind }}-viewer-role
rules:
  - apiGroups:
      - {{ .Resource.Domain }}
    resources:
      - {{ .Resource.Plural }}
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - {{ .Resource.Domain }}
    resources:
      - {{ .Resource.Plural }}/status
    verbs:
      - get
`
