# Copyright 2021 Ciena Corporation.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Capture version informaton from version file
VERSION=$(shell head -1 ../VERSION)

# Image URL to use all building/pushing image targets
DOCKER_REGISTRY ?= dockerhub.com
DOCKER_TAG ?= $(VERSION)
DOCKER_MANAGER_REPOSITORY ?= rule-provider-mock

ifeq ($(DOCKER_REGISTRY),)
IMG ?= $(DOCKER_MANAGER_REPOSITORY):$(DOCKER_TAG)
else
IMG ?= $(DOCKER_REGISTRY)/$(DOCKER_MANAGER_REPOSITORY):$(DOCKER_TAG)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.DEFAULT_GOAL:=help

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

venv: ## Create python virtual environment and install requirements
	virtualenv -p python3.7 venv

requirements: ## Install python requirements
	pip install -r requirements.txt

.PHONY: protos ## Generate python bindings to gRPC protos
protos:  venv ../apis/ruleprovider.proto
	python -m grpc_tools.protoc -I .. --python_out=. --grpc_python_out=. ../apis/ruleprovider.proto

##@ Build

.PHONY: run
run: ## Run the mock controller locally
	kopf run --verbose --debug --all-namespaces controller.py

.PHONY: docker-build
docker-build: ## Build docker image for the mock controller
	docker build -t ${IMG} -f Dockerfile .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: clean
clean: ## Delete build and/or temporary artifacts
	rm -rf ./bin ./*.out ./*.log $$(find . -name __pycache__)

##@ Deployment

.PHONY: install
install: ## Install the rule provider mock CRD into the K8s cluster specified in ~/.kube/config.
	kubectl apply --wait -f ./apis/ruleprovider_crd.yaml -f ./apis/costprovider_crd.yaml

.PHONY: uninstall
uninstall:  ## Uninstall the rule provider mock CRD from the K8s cluster specified in ~/.kube/config. 
	kubectl delete --ignore-not-found --wait -f ./apis/ruleprovider_crd.yaml -f ./apis/costprovider_crd.yaml

.PHONY: deploy
deploy: manifests kustomize  ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd manifests/deploy && $(KUSTOMIZE)  edit set image mock=${IMG}
	$(KUSTOMIZE)  build manifests/deploy | kubectl apply --wait -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from th K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE)  build manifests/deploy | kubectl delete --ignore-not-found --wait -f -

.PHONY: tail-log
tail-log: ## tail the log of the controller
	kubectl logs -f deploy/mock-controller

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

