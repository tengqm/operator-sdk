// Copyright 2020 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package templates

import (
	"bytes"
	"fmt"
	"text/template"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"

	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/constants"
)

// Watches scaffolds the watches.yaml file
type Watches struct {
	file.TemplateMixin
}

// SetTemplateDefaults sets the default values for templates.
// This method implements input.Template interface.
func (f *Watches) SetTemplateDefaults() error {
	// Path is returned by the `GetPath()` method of PathMixin nested in TemplateMixin.
	if f.Path == "" {
		f.Path = constants.DefaultWatchesFile
	}

	// TemplateBody is defined in the mixin, will be returned by `GetBody()`.
	f.TemplateBody = fmt.Sprintf(watchesTemplate,
		file.NewMarkerFor(f.Path, constants.WatchMarker),
	)
	return nil
}

var (
	// Ensure that Watches implements the Template interface.
	_ file.Template = &Watches{}
	// Ensure WatchesUpdater implements the Inserter interface.
	_ file.Inserter = &WatchesUpdater{}
)

// WatchesUpdater struct for updating the watches file
// `TemplateMixin` is the mixin that should be embedded in Template builders
// `ResourceMixin` provides templates with a injectable resource field.
type WatchesUpdater struct {
	file.TemplateMixin
	file.ResourceMixin

	// TODO: investigate whether to remove this
	UseTemplate string
}

// GetPath returns the path to the file location
// Required by KB pkg/model/file.Builder
func (*WatchesUpdater) GetPath() string {
	return constants.DefaultWatchesFile
}

// GetIfExistsAction returns the behavior when creating a file that already exists.
// Required by KB pkg/model/file.Builder
func (*WatchesUpdater) GetIfExistsAction() file.IfExistsAction {
	return file.Overwrite
}

// GetMarkers returns the different markers where code fragments will be inserted.
// Required by KB pkg/model/file.Inserter interface
func (f *WatchesUpdater) GetMarkers() []file.Marker {
	return []file.Marker{
		file.NewMarkerFor(constants.DefaultWatchesFile, constants.WatchMarker),
	}
}

// GetCodeFragments returns a map that binds markers to code fragments
// Required by KB pkg/model/file.Inserter interface
func (f *WatchesUpdater) GetCodeFragments() file.CodeFragmentsMap {
	fragments := make(file.CodeFragmentsMap, 1)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// Generate watch fragments
	watches := make([]string, 0)
	buf := &bytes.Buffer{}

	// TODO(asmacdo) Move template execution into a function, executed by the apiScaffolder.scaffold()
	// DefaultFuncMap used provide the function "lower", used in the watch fragment.
	tmpl := template.Must(template.New("rules").Funcs(file.DefaultFuncMap()).Parse(watchFragment))
	if err := tmpl.Execute(buf, f); err != nil {
		panic(err)
	}
	watches = append(watches, buf.String())

	if len(watches) != 0 {
		fragments[file.NewMarkerFor(constants.DefaultWatchesFile, constants.WatchMarker)] = watches
	}
	return fragments
}

const watchesTemplate = `---
# Use the 'create api' subcommand to add watches to this file.
%s
`

const watchFragment = `- version: {{ .Resource.Version }}
  group: {{ .Resource.Domain }}
  kind: {{ .Resource.Kind }}
`
