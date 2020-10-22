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

// ManagerRoleBinding scaffolds the config/rbac/role_binding.yaml file.
// TemplateMixin is the mixin that is embedded in Template builders.
type ManagerRoleBinding struct {
	file.TemplateMixin
}

var (
	// Ensure that ManagerRoleBinding implements the Template interface.
	_ file.Template = &ManagerRoleBinding{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface.
func (f *ManagerRoleBinding) SetTemplateDefaults() error {
	// Path is returned by the `GetPath()` method of PathMixin nested in TemplateMixin.
	if f.Path == "" {
		f.Path = filepath.Join("config", "rbac", "role_binding.yaml")
	}

	// TemplateBody is defined in the TemplateMixin, will be returned by `GetBody()`.
	f.TemplateBody = roleBindingTemplate

	return nil
}

const roleBindingTemplate = `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: manager-role
subjects:
  - kind: ServiceAccount
    name: default
    namespace: system
`
