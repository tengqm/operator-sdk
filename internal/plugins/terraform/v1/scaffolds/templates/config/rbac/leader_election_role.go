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

// LeaderElectionRole scaffolds the config/rbac/leader_election_role.yaml file.
// ProjectNameMixin provides templates with a injectable ProjectName field.
type LeaderElectionRole struct {
	file.TemplateMixin
}

var (
	// This ensure the LeaderElectionRole implements the KB pkg/model/file.Template interface.
	_ file.Template = &LeaderElectionRole{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface.
func (f *LeaderElectionRole) SetTemplateDefaults() error {
	// To be returned by the `GetPath()` method of PathMixin nested in TemplateMixin.
	if f.Path == "" {
		f.Path = filepath.Join("config", "rbac", "leader_election_role.yaml")
	}

	// TemplateBody is defined in the TemplateMixin, will be returned by `GetBody()`.
	f.TemplateBody = leaderElectionRoleTemplate

	return nil
}

const leaderElectionRoleTemplate = `# permissions to do leader election.
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: leader-election-role
rules:
  - apiGroups:
      - ""
    resources:
      - configmaps
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - create
      - patch
`
