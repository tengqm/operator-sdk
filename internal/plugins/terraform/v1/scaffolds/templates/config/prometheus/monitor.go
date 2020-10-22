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

package prometheus

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
)

// ServiceMonitor scaffolds an issuer CR and a certificate CR.
// TemplateMixin is the mixin that is embedded in Template builders.
type ServiceMonitor struct {
	file.TemplateMixin
}

var (
	// This enusre the ServiceMonitor implements the KB pkg/model/file.Template interface.
	_ file.Template = &ServiceMonitor{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface
func (f *ServiceMonitor) SetTemplateDefaults() error {
	// `Path` is returned by the `GetPath()` method of `PathMixin` nested in `TemplateMixin`.
	if f.Path == "" {
		f.Path = filepath.Join("config", "prometheus", "monitor.yaml")
	}

	// `TemplateBody` is defined in the TemplateMixin, will be returned by `GetBody`
	f.TemplateBody = serviceMonitorTemplate

	return nil
}

const serviceMonitorTemplate = `---
# Prometheus Monitor Service (Metrics)
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    control-plane: controller-manager
  name: controller-manager-metrics-monitor
  namespace: system
spec:
  endpoints:
    - path: /metrics
      port: https
  selector:
    matchLabels:
      control-plane: controller-manager
`
