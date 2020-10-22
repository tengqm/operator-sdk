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

package crd

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"
)

const (
	// scope is this package
	resourceMarker = "crdkustomizeresource"
)

// Kustomization scaffolds the kustomization file in manager folder.
// TemplateMixin is the mixin that is embedded in Template builders.
// ResourceMixin provides templates with a injectable resource field.
type Kustomization struct {
	file.TemplateMixin
	file.ResourceMixin
}

var (
	// Ensure that Kustomization implements the Template interface.
	_ file.Template = &Kustomization{}

	// Ensure that Kustomization implements the Inserter interface.
	_ file.Inserter = &Kustomization{}
)

// SetTemplateDefaults sets the default values for templates.
// Required by the KB pkg/model/file.Template interface.
func (f *Kustomization) SetTemplateDefaults() error {
	// Path is returned by the `GetPath()` method of PathMixin nested in TemplateMixin.
	if f.Path == "" {
		f.Path = filepath.Join("config", "crd", "kustomization.yaml")
	}
	f.Path = f.Resource.Replacer().Replace(f.Path)

	// TemplateBody is defined in the TemplateMixin, will be returned by `GetBody()`.
	// The indentations are to make sure the schema body fixs in the template.
	f.TemplateBody = fmt.Sprintf(kustomizationTemplate,
		file.NewMarkerFor(f.Path, resourceMarker),
	)

	return nil
}

// GetMarkers returns the different markers where code fragments will be inserted.
// Required by KB pkg/model/file.Inserter interface
func (f *Kustomization) GetMarkers() []file.Marker {
	return []file.Marker{
		file.NewMarkerFor(f.Path, resourceMarker),
	}
}

const (
	resourceCodeFragment = `- bases/%s_%s.yaml
`
)

// GetCodeFragments returns a map that binds markers to code fragments
// Required by KB pkg/model/file.Inserter interface
func (f *Kustomization) GetCodeFragments() file.CodeFragmentsMap {
	fragments := make(file.CodeFragmentsMap, 3)

	// Generate resource code fragments
	res := make([]string, 0)
	res = append(res, fmt.Sprintf(resourceCodeFragment, f.Resource.Domain, f.Resource.Plural))

	// Only store code fragments in the map if the slices are non-empty
	if len(res) != 0 {
		fragments[file.NewMarkerFor(f.Path, resourceMarker)] = res
	}

	return fragments
}

var kustomizationTemplate = `# This kustomization.yaml is not intended to be run by itself,
# since it depends on service name and namespace that are out of this kustomize package.
# It should be run by config/default
resources:
%s
`
