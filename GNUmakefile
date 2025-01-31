NAME=image-builder
ROOT_DIR:=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR=build
PLUGIN_DIR=${BUILD_DIR}/plugins
VERSION=$(shell git describe --tags --abbrev=0 | cut -b 2-)
BINARY=packer-plugin-${NAME}_v${VERSION}_x5.0_linux_amd64
# https://github.com/hashicorp/packer-plugin-sdk/issues/187
HASHICORP_PACKER_PLUGIN_SDK_VERSION?="v0.5.2"
PLUGIN_FQN=$(shell grep -E '^module' <go.mod | sed -E 's/module \s*//')
PLUGIN_PATH=./cmd/plugin

.PHONY: build
build: ## Build the plugin
	@mkdir -p ${PLUGIN_DIR}
	@go build -ldflags="-X 'main.Version=${VERSION}'" -o ${PLUGIN_DIR}/${BINARY} ${PLUGIN_PATH}
	@sha256sum < ${PLUGIN_DIR}/${BINARY} > ${PLUGIN_DIR}/${BINARY}_SHA256SUM

.PHONY: clean
clean: ## Clean the build directory
	@rm -rf ${BUILD_DIR}

.PHONY: packer-build
packer-build: build ## Execute packer with the plugin
	PACKER_PLUGIN_PATH=${ROOT_DIR}${BUILD_DIR} packer build ${HCL}

.PHONY: install
install: build ## Install the plugin
	packer plugins install --path ${BINARY} "$(shell echo "${PLUGIN_FQN}" | sed 's/packer-plugin-//')"

.PHONY: test
test: ## Build the plugin and run all tests
	@yamllint -s internal/integration/*.yaml
	@go test -race ./...

.PHONY: install-packer-sdc
install-packer-sdc: ## Install packer-sdc
	@go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@${HASHICORP_PACKER_PLUGIN_SDK_VERSION}

.PHONY: plugin-check
plugin-check: install-packer-sdc build ## Check the plugin
	$(shell cd ${PLUGIN_DIR}; packer-sdc plugin-check ${BINARY})
	@${PLUGIN_DIR}/${BINARY} describe

.PHONY: generate
generate: install-packer-sdc
# https://github.com/hashicorp/packer-plugin-sdk/issues/187
	@go mod edit -replace "github.com/zclconf/go-cty=github.com/nywilken/go-cty@v1.13.3"
	@go mod tidy
	@go generate ${PLUGIN_PATH}
