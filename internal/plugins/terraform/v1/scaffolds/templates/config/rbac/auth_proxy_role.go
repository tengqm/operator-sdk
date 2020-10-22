/*
Copyright 2018 The Kubernetes Authors.

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

// AuthProxyRole scaffolds the config/rbac/auth_proxy_role.yaml file.
// TemplateMixin is the mixin that is embedded in Template builders.
type AuthProxyRole struct {
	file.TemplateMixin
}

var (
	// This ensure the AuthProxyRole implements the KB pkg/model/file.Template interface.
	_ file.Template = &AuthProxyRole{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface
func (f *AuthProxyRole) SetTemplateDefaults() error {
	// `Path` is returned by the `GetPath()` method of `PathMixin` nested in `TemplateMixin`.
	if f.Path == "" {
		f.Path = filepath.Join("config", "rbac", "auth_proxy_role.yaml")
	}

	// `TemplateBody` is defined in the TemplateMixin, will be returned by `GetBody`
	f.TemplateBody = proxyRoleTemplate

	return nil
}

const proxyRoleTemplate = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: proxy-role
rules:
- apiGroups: ["authentication.k8s.io"]
  resources:
  - tokenreviews
  verbs: ["create"]
- apiGroups: ["authorization.k8s.io"]
  resources:
  - subjectaccessreviews
  verbs: ["create"]
`
