# Local builds will use local docker for buildkit server. This is set on
BUILDX_BUILDER_NAME	?= default
CONNECTOR_MAIN_GO   ?= main.go
CONNECTOR_NAME      ?= digicert-ca-connector
LD_FLAGS            ?= "-w -s"
TAG                 ?= latest

# Use source makefile directory as image name
IMAGE_NAME          ?= tls-protect-${CONNECTOR_NAME}

# Docker for mac and Linux needs specific arguments to mount ssh agent sock
ifeq ($(OS),Windows_NT)
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        DOCKER_EXTRA_ARGS=-v ${SSH_AUTH_SOCK}:${SSH_AUTH_SOCK} -e SSH_AUTH_SOCK=${SSH_AUTH_SOCK}
    endif
    ifeq ($(UNAME_S),Darwin)
        DOCKER_EXTRA_ARGS=-v /run/host-services/ssh-auth.sock:/run/host-services/ssh-auth.sock -e SSH_AUTH_SOCK=/run/host-services/ssh-auth.sock
    endif
endif

.DEFAULT: clean build

.PHONY: help
help:
	@echo "make                       same effect as: clean build"
	@echo "make build                 build binary for usage in a vSatellite"
	@echo "make build-local           build binary for local usage"
	@echo "make clean                 clean output"
	@echo "make format                cleanup import lines and format the code in the same style as gofmt"
	@echo "make help                  displays this help"
	@echo "\nTests"
	@echo "make test                  run tests"
	@echo "make lint                  run linter"
	@echo "\nImages"
	@echo "make image                 generate a container image"
	@echo "make push                  generate and push a container image to the registry"
	@echo "\nOther"
	@echo "make generate              generate schema"
	@echo "make manifests             generate a create and an update manifest from the existing manifest.json"
	@echo ""

### Init rules

.PHONY: init
init:
	mkdir -p output/reports

.PHONY: init-vars
init-vars:
	$(eval MODULE_FQDN=$(shell GOWORK='off' go list -m))

### Build rules

.PHONY: clean
clean:
	@rm -rf output/
	@rm -f buildx-digest.json
	@rm -f manifest_with_image.json
	@rm -f manifest.create.json
	@rm -f manifest.update.json

.PHONY: build
build: generate
	mkdir -p output/bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o output/bin/${CONNECTOR_NAME} ./cmd/$(notdir $(CURDIR))/${CONNECTOR_MAIN_GO}

.PHONY: build-local
build-local: generate
	go install ./cmd/$(notdir $(CURDIR))/${CONNECTOR_MAIN_GO}

.PHONY: format
format:
	goimports -w .

.PHONY: generate
generate: init-vars
	go mod download
	go generate $(MODULE_FQDN)/...

### Image rules

ifndef CONTAINER_REGISTRY
image:
	$(error CONTAINER_REGISTRY is not set)
push:
	$(error CONTAINER_REGISTRY is not set)
manifests:
	$(error CONTAINER_REGISTRY is not set)
else
.PHONY: image
image: BUILDX_OUTPUT=--output type=image,name=${CONTAINER_REGISTRY}/${IMAGE_NAME}:${TAG} --metadata-file=buildx-digest.json
image: BUILDX_TARGET:=--target image
image: buildx

.PHONY: push
push: BUILDX_OUTPUT=--output type=image,name=${CONTAINER_REGISTRY}/${IMAGE_NAME}:${TAG},push=true --metadata-file=buildx-digest.json
push: BUILDX_TARGET:=--target image
push: buildx

.PHONY: manifests
manifests:
	@echo "Generate manifests"
	@echo "     ImagePath: ${CONTAINER_REGISTRY}/${IMAGE_NAME}:${TAG}"
	@echo "     ImageName: ${IMAGE_NAME}"
	$(eval CONNECTOR_BUILD_DIGEST=$(shell cat buildx-digest.json | jq -r '."containerimage.digest"'))
	@echo "  Build Digest: ${CONNECTOR_BUILD_DIGEST}"
	$(eval CONNECTOR_PLUGIN_TYPE=$(shell cat manifest.json | jq -r '."pluginType"'))
	@echo "   Plugin Type: ${CONNECTOR_PLUGIN_TYPE}"
	$(eval CONNECTOR_WORK_TYPES=$(shell cat manifest.json | jq -r '."workTypes"'))
	@echo "    Work Types: ${CONNECTOR_WORK_TYPES}"
	@test -s "manifest.json" || { echo "No manifest.json file found"; exit 1; }
	@jq '.deployment.image = "${CONTAINER_REGISTRY}/${IMAGE_NAME}:${TAG}" | .deployment.executionTarget = "vsat"' manifest.json > manifest_with_image.json
	@jq '{manifest: .}' manifest_with_image.json > manifest.update.json
	@jq '.pluginType = "${CONNECTOR_PLUGIN_TYPE}"' manifest.update.json > manifest.create.json
endif

.PHONY: buildx
buildx:
	docker --context=default buildx build ${BUILDX_OUTPUT} \
		${BUILDX_TARGET} \
		--file build/Dockerfile \
		${BUILDX_EXTRA_ARGS} \
		--platform=linux/amd64 \
		--builder ${BUILDX_BUILDER_NAME} .

### test rules

.PHONY: lint
lint:
	golangci-lint run --config build/golangci.yaml --out-format colored-line-number --issues-exit-code 1 ./...

.PHONY: test
test:
	go test -cover ./...