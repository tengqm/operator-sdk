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

	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/constants"
)

// AuthProxyPatch scaffolds the patch file for enabling prometheus metrics for manager Pod.
// TemplateMixin is the mixin that should be embedded in Template builders.
// ProjectNameMixin provides templates with an injectable project name field.
type AuthProxyPatch struct {
	file.TemplateMixin
	file.ProjectNameMixin

	// The version of the `kube-rbac-proxy` image.
	RbacProxyVersion string
}

var (
	// This ensures AuthProxyPatch implements the KB pkg/model/file.Template interface.
	_ file.Template = &AuthProxyPatch{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface
func (f *AuthProxyPatch) SetTemplateDefaults() error {
	// `Path` is returned by the `GetPath()` method of `PathMixin` which is nested in TemplateMixin.
	if f.Path == "" {
		f.Path = filepath.Join("config", "default", "manager_auth_proxy_patch.yaml")
	}

	// IfExistsAction specifies the behavior when creating a file that already exists.
	f.IfExistsAction = file.Error

	// `TemplateBody` is defined in the TemplateMixin, will be returned by `GetBody()`.
	f.TemplateBody = kustomizeAuthProxyPatchTemplate

	f.RbacProxyVersion = constants.RbacProxyVersion

	return nil
}

const kustomizeAuthProxyPatchTemplate = `# This patch injects a sidecar container which is a HTTP proxy for the
# controller manager, it performs RBAC authorization against the Kubernetes API using SubjectAccessReviews.
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
spec:
  template:
    spec:
      containers:
        - name: kube-rbac-proxy
          image: gcr.io/kubebuilder/kube-rbac-proxy:{{- .RbacProxyVersion -}}
          args:
            - "--secure-listen-address=0.0.0.0:8443"
            - "--upstream=http://127.0.0.1:8080/"
            - "--logtostderr=true"
            - "--v=10"
          ports:
            - containerPort: 8443
              name: https
        - name: manager
          args:
            - "--metrics-addr=127.0.0.1:8080"
            - "--enable-leader-election"
            - "--leader-election-id={{ .ProjectName }}"
`
