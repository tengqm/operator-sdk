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

// ClientClusterRole scaffolds the config/rbac/client_clusterrole.yaml file.
// TemplateMixin is the mixin that is embedded in Template builders.
type ClientClusterRole struct {
	file.TemplateMixin
}

var (
	// This ensure the ClientClusterRole implements the KB pkg/model/file.Template interface.
	_ file.Template = &ClientClusterRole{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface
func (f *ClientClusterRole) SetTemplateDefaults() error {
	// `Path` is returned by the `GetPath()` method of `PathMixin` nested in `TemplateMixin`.
	if f.Path == "" {
		f.Path = filepath.Join("config", "rbac", "auth_proxy_client_clusterrole.yaml")
	}

	// `TemplateBody` is defined in the TemplateMixin, will be returned by `GetBody`
	f.TemplateBody = clientClusterRoleTemplate

	return nil
}

const clientClusterRoleTemplate = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: metrics-reader
rules:
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
`
