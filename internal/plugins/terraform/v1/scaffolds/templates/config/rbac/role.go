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
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
)

const (
	// scope is this package
	rulesMarker = "rules"
)

// ManagerRole scaffolds the role.yaml file
// TemplateMixin is the mixin that is embedded in Template builders.
type ManagerRole struct {
	file.TemplateMixin
}

var (
	// This ensure the ManagerRole implements the KB pkg/model/file.Template interface.
	_ file.Template = &ManagerRole{}

	defaultRoleFile = filepath.Join("config", "rbac", "role.yaml")
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface.
func (f *ManagerRole) SetTemplateDefaults() error {
	// Path is returned by the `GetPath()` method of PathMixin nested in TemplateMixin.
	if f.Path == "" {
		f.Path = defaultRoleFile
	}

	// TemplateBody is defined in the TemplateMixin, will be returned by `GetBody`
	f.TemplateBody = fmt.Sprintf(roleTemplate,
		file.NewMarkerFor(f.Path, rulesMarker),
	)
	return nil
}

// ManagerRoleUpdater implements an updater for the role template.
// TemplateMixin is the mixin that is embedded in Template builders.
// ResourceMixin provides templates with a injectable resource field.
type ManagerRoleUpdater struct {
	file.TemplateMixin
	file.ResourceMixin

	SkipDefaultRules bool
}

var (
	// This ensures the ManagerRoleUpdater implements the KB pkg/model/file.Inserter interface.
	_ file.Inserter = &ManagerRoleUpdater{}
)

// GetPath is required by the KB pkg/model/file.PathMixin interface.
func (*ManagerRoleUpdater) GetPath() string {
	return defaultRoleFile
}

// GetIfExistsAction is required by the KB pkg/model/file.IfExistsActionMixin interface.
func (*ManagerRoleUpdater) GetIfExistsAction() file.IfExistsAction {
	return file.Overwrite
}

// GetMarkers returns the different markers where code fragments will be inserted.
// Required by the KB pkg/model/file.Inserter interface.
func (f *ManagerRoleUpdater) GetMarkers() []file.Marker {
	return []file.Marker{
		file.NewMarkerFor(defaultRoleFile, rulesMarker),
	}
}

// GetCodeFragments returns a map that binds markers to code fragments.
// Required by the KB pkg/model/file.Inserter interface.
func (f *ManagerRoleUpdater) GetCodeFragments() file.CodeFragmentsMap {
	fragments := make(file.CodeFragmentsMap, 1)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("rules").Parse(rulesFragment))
	if err := tmpl.Execute(buf, f); err != nil {
		panic(err)
	}

	// Generate rule fragment
	rules := []string{buf.String()}
	if len(rules) != 0 {
		fragments[file.NewMarkerFor(defaultRoleFile, rulesMarker)] = rules
	}
	return fragments
}

const roleTemplate = `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: manager-role
rules:
  ##
  ## Base operator rules
  ##
  - apiGroups:
      - ""
    resources:
      - secrets
      - pods
      - pods/exec
      - pods/log
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - apps
    resources:
      - deployments
      - daemonsets
      - replicasets
      - statefulsets
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
%s
`

const rulesFragment = `  ##
  ## Rules for {{- .Resource.Domain -}}/{{- .Resource.Version -}}, Kind: {{- .Resource.Kind -}}
  ##
  - apiGroups:
      - {{- .Resource.Domain -}}
    resources:
      - {{- .Resource.Plural -}}
      - {{- .Resource.Plural -}}/status
      - {{- .Resource.Plural -}}/finalizers
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
`
