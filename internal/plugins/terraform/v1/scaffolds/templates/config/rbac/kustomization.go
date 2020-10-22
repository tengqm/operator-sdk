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

// Kustomization scaffolds the Kustomization file in rbac folder.
// TemplateMixin is the mixin that is embedded in Template builders.
type Kustomization struct {
	file.TemplateMixin
}

var (
	// Ensure that Kustomization implements the Template interface.
	_ file.Template = &Kustomization{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface.
func (f *Kustomization) SetTemplateDefaults() error {
	// `Path` is returned by the `GetPath()` method of `PathMixin` nested in `TemplateMixin`.
	if f.Path == "" {
		f.Path = filepath.Join("config", "rbac", "kustomization.yaml")
	}

	// `TemplateBody` is defined in the TemplateMixin, will be returned by `GetBody`
	f.TemplateBody = kustomizeRBACTemplate

	f.IfExistsAction = file.Error

	return nil
}

const kustomizeRBACTemplate = `resources:
  - role.yaml
  - role_binding.yaml
  - leader_election_role.yaml
  - leader_election_role_binding.yaml
  # Comment the following 4 lines if you want to disable the auth proxy
  # (https://github.com/brancz/kube-rbac-proxy) which protects your /metrics endpoint.
  - auth_proxy_service.yaml
  - auth_proxy_role.yaml
  - auth_proxy_role_binding.yaml
  - auth_proxy_client_clusterrole.yaml
`
