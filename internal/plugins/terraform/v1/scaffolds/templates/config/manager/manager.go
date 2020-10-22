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

package manager

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"

	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/constants"
)

// Manager scaffolds the manifiests for the manager Deployment.
// TemplateMixin is the mixin that is embedded in Template builders.
// ProjectNameMixin provides templates with a injectable ProjectName field.
type Manager struct {
	file.TemplateMixin
	file.ProjectNameMixin

	// Image is controller manager image name
	Image string

	// ProviderName is the terraform provider to use
	// TODO: make this a list
	ProviderName string
}

var (
	// This ensure the Manager implements the KB pkg/model/file.Template interface.
	_ file.Template = &Manager{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface.
func (f *Manager) SetTemplateDefaults() error {
	// Path is returned by the `GetPath()` method of PathMixin nested in TemplateMixin.
	if f.Path == "" {
		f.Path = filepath.Join("config", "manager", "manager.yaml")
	}

	// TemplateBody is defined in the TemplateMixin, will be returned by `GetBody()`.
	f.TemplateBody = configTemplate

	f.Image = constants.ImageName

	return nil
}

const configTemplate = `apiVersion: v1
kind: Namespace
metadata:
  labels:
    control-plane: controller-manager
  name: system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
  labels:
    control-plane: controller-manager
spec:
  selector:
    matchLabels:
      control-plane: controller-manager
  replicas: 1
  template:
    metadata:
      labels:
        control-plane: controller-manager
    spec:
      containers:
        - name: manager
          args:
            - "--enable-leader-election"
            - "--leader-election-id={{ .ProjectName }}"
          image: {{ .Image }}
		  # Secret named after the provider name, each provider (if there is
		  # more than one) has its own set of options.
		  {{ if .ProviderName }}
		  envFrom:
		    - secretRef:
			    name: {{ .ProviderName }}-options
		  {{ end }}
      terminationGracePeriodSeconds: 10
`
