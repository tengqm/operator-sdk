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

package templates

import (
	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
)

var (
	// Ensure that Dockerfile implements the Template interface.
	_ file.Template = &GitIgnore{}
)

// GitIgnore scaffolds the .gitignore file.
// The TemplateMixin is a mixin required for all template builders. The
// `GetPath`, `GetIfExistsAction`, `GetBody()` methods is implemented in the
// default mixin; this struct only needs to implement `SetTemplateDefaults`.
type GitIgnore struct {
	file.TemplateMixin
}

// SetTemplateDefaults implements input.Template
// This method sets the default values for templates.
func (f *GitIgnore) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = ".gitignore"
	}

	// `TemplateBody` is defined in the mixin, will be returned by `GetBody`
	f.TemplateBody = gitignoreTemplate

	return nil
}

const gitignoreTemplate = `
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin

# editor and IDE paraphernalia
.idea
*.swp
*.swo
*~
`
