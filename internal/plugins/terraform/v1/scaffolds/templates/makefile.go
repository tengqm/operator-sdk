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
	"errors"

	"sigs.k8s.io/kubebuilder/v2/pkg/model/file"

	"github.com/operator-framework/operator-sdk/internal/plugins/terraform/v1/constants"
)

// Makefile scaffolds the Makefile
// The TemplateMixin is a mixin required for all template builders. The
// `GetPath`, `GetIfExistsAction`, `GetBody()` methods is implemented in the
// default mixin; this struct only needs to implement `SetTemplateDefaults`.
type Makefile struct {
	file.TemplateMixin

	// ImageName is image name for the controller manager.
	ImageName string

	// Kustomize version to use in the project.
	KustomizeVersion string

	// OperatorVersion is the version of the terraform-operator binary downloaded by the Makefile.
	OperatorVersion string
}

var (
	// Ensure that Dockerfile implements the Template interface.
	_ file.Template = &Makefile{}
)

// SetTemplateDefaults sets the default values for templates.
// This method implements input.Template interface.
func (f *Makefile) SetTemplateDefaults() error {
	// Path is returned by the `GetPath()` method of PathMixin nested in TemplateMixin.
	if f.Path == "" {
		f.Path = "Makefile"
	}

	// TemplateBody is defined in the mixin, will be returned by `GetBody()`.
	f.TemplateBody = makefileTemplate

	// Error if the file already exists!
	f.IfExistsAction = file.Error

	f.ImageName = constants.ImageName
	f.KustomizeVersion = constants.KustomizeVersion

	if f.OperatorVersion == "" {
		return errors.New("terraform-operator version is required in scaffold")
	}

	return nil
}

const makefileTemplate = `
# Image URL to use all building/pushing image targets
IMG ?= {{ .ImageName }}

all: docker-build

# Run against the Kubernetes cluster in ~/.kube/config
run: terraform-operator
	$(TERRAFORM_OPERATOR) run

# Install CRDs into a cluster
install: kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the Kubernetes cluster in ~/.kube/config
deploy: kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Undeploy controller in the Kubernetes cluster in ~/.kube/config
undeploy: kustomize
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

PATH  := $(PATH):$(PWD)/bin
SHELL := env PATH=$(PATH) /bin/sh
OS    = $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH  = $(shell uname -m | sed 's/x86_64/amd64/')
OSOPER   = $(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/darwin/apple-darwin/' | sed 's/linux/linux-gnu/')
ARCHOPER = $(shell uname -m )

kustomize:
ifeq (, $(shell which kustomize 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p bin ;\
	curl -sSLo - https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/{{ .KustomizeVersion }}/kustomize_{{ .KustomizeVersion }}_$(OS)_$(ARCH).tar.gz | tar xzf - -C bin/ ;\
	}
KUSTOMIZE=$(realpath ./bin/kustomize)
else
KUSTOMIZE=$(shell which kustomize)
endif

terraform-operator:
ifeq (, $(shell which terraform-operator 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p bin ;\
	curl -LO https://github.com/operator-framework/operator-sdk/releases/download/{{ .OperatorVersion }}/terraform-operator-{{ .OperatorVersion }}-$(ARCHOPER)-$(OSOPER) ;\
	mv terraform-operator-{{ .OperatorVersion }}-$(ARCHOPER)-$(OSOPER) ./bin/terraform-operator ;\
	chmod +x ./bin/terraform-operator ;\
	}
TERRAFORM_OPERATOR=$(realpath ./bin/terraform-operator)
else
TERRAFORM_OPERATOR=$(shell which terraform-operator)
endif
`
