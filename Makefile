# Copyright 2020 The bigshot Authors
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

GOOS ?= $(shell go env GOOS)
GOARCH ?= amd64
BUILD_DIR ?= ./out
COMMAND_PKG ?= cmd
ORG = github.com/DevopsArtFactory
PROJECT = bigshot

REPOPATH ?= $(ORG)/$(PROJECT)
RELEASE_BUCKET ?= devopsartfactory
S3_RELEASE_PATH ?= s3://$(RELEASE_BUCKET)/$(PROJECT)/releases/$(VERSION)
S3_RELEASE_LATEST ?= s3://$(RELEASE_BUCKET)/$(PROJECT)/releases/latest
S3_BLEEDING_EDGE_LATEST ?= s3://$(RELEASE_BUCKET)/edge/latest
S3_WORKER_PATH ?= s3://$(RELEASE_BUCKET)/$(PROJECT)/code

WORKER = lambda
HANDLER = handler
WORKER_ZIP = $(WORKER).zip
WORKER_CODE_PKG ?= code/$(WORKER)
WORKER_BUILD_PACKAGE = $(WORKER_CODE_PKG)/main.go

GCP_ONLY ?= false
GCP_PROJECT ?= bigshot

SUPPORTED_PLATFORMS = linux-amd64 darwin-amd64 windows-amd64.exe linux-arm64
BUILD_PACKAGE = $(REPOPATH)/$(COMMAND_PKG)/$(PROJECT)

bigshot_TEST_PACKAGES = ./pkg/... ./cmd/... ./hack/...
GO_FILES = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./pkg/diag/*")

VERSION_PACKAGE = $(REPOPATH)/pkg/version
COMMIT = $(shell git rev-parse HEAD)
TEST_PACKAGES = ./pkg/... ./hack/...

ifeq "$(strip $(VERSION))" ""
 override VERSION = $(shell git describe --always --tags --dirty)
endif

LDFLAGS_linux = -static
LDFLAGS_darwin =
LDFLAGS_windows =

GO_BUILD_TAGS_linux = "osusergo netgo static_build release"
GO_BUILD_TAGS_darwin = "release"
GO_BUILD_TAGS_windows = "release"

GO_LDFLAGS = -X $(VERSION_PACKAGE).version=$(VERSION)
GO_LDFLAGS += -X $(VERSION_PACKAGE).buildDate=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS += -X $(VERSION_PACKAGE).gitCommit=$(COMMIT)
GO_LDFLAGS += -X $(VERSION_PACKAGE).gitTreeState=$(if $(shell git status --porcelain),dirty,clean)
GO_LDFLAGS += -s -w

GO_LDFLAGS_windows =" $(GO_LDFLAGS)  -extldflags \"$(LDFLAGS_windows)\""
GO_LDFLAGS_darwin =" $(GO_LDFLAGS)  -extldflags \"$(LDFLAGS_darwin)\""
GO_LDFLAGS_linux =" $(GO_LDFLAGS)  -extldflags \"$(LDFLAGS_linux)\""

# Build for local development.
$(BUILD_DIR)/$(PROJECT): $(GO_FILES) $(BUILD_DIR)
	@echo
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 go build -tags $(GO_BUILD_TAGS_$(GOOS)) -ldflags $(GO_LDFLAGS_$(GOOS)) -o $@ $(BUILD_PACKAGE)

.PHONY: install
install: $(BUILD_DIR)/$(PROJECT)
	cp $(BUILD_DIR)/$(PROJECT) $(GOPATH)/bin/$(PROJECT)

.PRECIOUS: $(foreach platform, $(SUPPORTED_PLATFORMS), $(BUILD_DIR)/$(PROJECT)-$(platform))

.PHONY: cross
cross: $(foreach platform, $(SUPPORTED_PLATFORMS), $(BUILD_DIR)/$(PROJECT)-$(platform))

$(BUILD_DIR)/$(PROJECT)-%: $(STATIK_FILES) $(GO_FILES) $(BUILD_DIR) deploy/cross/Dockerfile
	$(eval os = $(firstword $(subst -, ,$*)))
	$(eval arch = $(lastword $(subst -, ,$(subst .exe,,$*))))
	$(eval ldflags = $(GO_LDFLAGS_$(os)))
	$(eval tags = $(GO_BUILD_TAGS_$(os)))

	docker build \
		--build-arg GOOS=$(os) \
		--build-arg GOARCH=$(arch) \
		--build-arg TAGS=$(tags) \
		--build-arg LDFLAGS=$(ldflags) \
		-f deploy/cross/Dockerfile \
		-t bigshot/cross \
		.

	docker run --rm bigshot/cross cat /build/bigshot > $@
	shasum -a 256 $@ | tee $@.sha256
	file $@ || true

.PHONY: $(BUILD_DIR)/VERSION
$(BUILD_DIR)/VERSION: $(BUILD_DIR)
	@ echo $(VERSION) > $@

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

.PHONY: update-edge
update-edge: format cross $(BUILD_DIR)/VERSION upload-edge-only

.PHONY: release
release: clean format linters test cross $(BUILD_DIR)/VERSION upload-only

.PHONY: build
build: format cross $(BUILD_DIR)/VERSION

.PHONY: upload-only
upload-only: version
	@ cp $(BUILD_DIR)/$(PROJECT)-darwin-amd64 $(BUILD_DIR)/$(PROJECT)
	@ aws s3 cp $(BUILD_DIR)/ $(S3_RELEASE_PATH)/ --recursive --include "$(PROJECT)-*" --acl public-read
	@ aws s3 cp $(S3_RELEASE_PATH)/ $(S3_RELEASE_LATEST)/ --recursive --acl public-read

.PHONY: upload-edge-only
upload-edge-only: version
	aws s3 cp $(BUILD_DIR)/ $(S3_BLEEDING_EDGE_LATEST)/ --recursive --include "$(PROJECT)-*" --acl public-read

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: version
version:
	@echo "Current version is ${VERSION}"

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	@ go test -count=1 -v -race -short -timeout=90s $(TEST_PACKAGES)

.PHONY: coverage
coverage: $(BUILD_DIR)
	@ go test -count=1 -race -cover -short -timeout=90s -coverprofile=out/coverage.txt -coverpkg="./pkg/...,./hack..." $(TEST_PACKAGES)
	@- curl -s https://codecov.io/bash > $(BUILD_DIR)/upload_coverage && bash $(BUILD_DIR)/upload_coverage

.PHONY: linters
linters: $(BUILD_DIR)
	@ ./hack/linters.sh

# utilities for bigshot site - not used anywhere else
.PHONY: preview-docs-draft
preview-docs-draft:
	./deploy/docs/preview-docs.sh hugo server -D --bind=0.0.0.0 --ignoreCache

.PHONY: preview-docs
preview-docs:
	./deploy/docs/preview-docs.sh hugo server --bind=0.0.0.0 --ignoreCache

.PHONY: generate-schema
generate-schema:
	go run ./hack/schemas/main.go

.PHONY: worker-build
worker-build:
	GOOS=linux GOARCH=$(GOARCH) go build -tags $(GO_BUILD_TAGS_$(GOOS)) -o $(BUILD_DIR)/$(WORKER_CODE_PKG)/$(HANDLER) $(WORKER_BUILD_PACKAGE)
	@ zip -9 $(BUILD_DIR)/$(WORKER_ZIP) $(BUILD_DIR)/$(WORKER_CODE_PKG)/$(HANDLER)

.PHONY: worker-release
worker-release: worker-build
	@ aws s3 cp $(BUILD_DIR)/$(WORKER_ZIP) $(S3_WORKER_PATH)/$(WORKER_ZIP) --acl public-read --profile art-id