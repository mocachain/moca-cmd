SHELL := /bin/bash

# Load .env file if it exists
-include .env
export

# Keep local Go commands on the repository toolchain even when a newer
# system-wide Go release is installed.
GO_TOOLCHAIN ?= go1.23.12
GO := env GOTOOLCHAIN=$(GO_TOOLCHAIN) go
GOPROXY ?= https://proxy.golang.org,direct
GONOSUMDB ?= github.com/mocachain

# Configure git to use HTTPS+Token for private repositories if GITHUB_TOKEN is set
ifdef GITHUB_TOKEN
  $(shell git config --global url."https://$(GITHUB_TOKEN):@github.com/".insteadOf "https://github.com/" 2>/dev/null)
endif

GITCONFIG_MOUNT :=
ifneq ($(wildcard $(HOME)/.gitconfig),)
GITCONFIG_MOUNT := -v $(HOME)/.gitconfig:/root/.gitconfig:ro
endif

LOCAL_REPLACE_MOUNTS :=
ifneq ($(wildcard ../moca/go.mod),)
LOCAL_REPLACE_MOUNTS += -v $(abspath ../moca):/go/src/github.com/mocachain/moca
endif
ifneq ($(wildcard ../moca-go-sdk/go.mod),)
LOCAL_REPLACE_MOUNTS += -v $(abspath ../moca-go-sdk):/go/src/github.com/mocachain/moca-go-sdk
endif
ifneq ($(wildcard ../moca-common/go/go.mod),)
LOCAL_REPLACE_MOUNTS += -v $(abspath ../moca-common):/go/src/github.com/mocachain/moca-common
endif

.PHONY: all build install-deps lint lint-fix lint-all lint-fix-all hooks pre-commit-staged

build:
	$(GO) build -o ./build/moca-cmd cmd/*.go

golangci_version=v1.64.8
staticcheck_version=v0.6.1
LEFTHOOK_VERSION=v1.11.3
INCREMENTAL_LINT_SCRIPT=./scripts/run-incremental-lint.sh
GO_GOBIN := $(shell $(GO) env GOBIN)
GO_GOPATH := $(shell $(GO) env GOPATH)

ifeq ($(GO_GOBIN),)
golangci_lint_cmd=$(GO_GOPATH)/bin/golangci-lint
lefthook_cmd=$(GO_GOPATH)/bin/lefthook
staticcheck_cmd=$(GO_GOPATH)/bin/staticcheck
else
golangci_lint_cmd=$(GO_GOBIN)/golangci-lint
lefthook_cmd=$(GO_GOBIN)/lefthook
staticcheck_cmd=$(GO_GOBIN)/staticcheck
endif

install-lint:
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)

install-staticcheck:
	@$(GO) install honnef.co/go/tools/cmd/staticcheck@$(staticcheck_version)

install-devtools: install-lint install-staticcheck

install-deps: install-devtools hooks

check-go-env:
	@echo "--> Using Go toolchain: $(GO_TOOLCHAIN)"
	@$(GO) version

check-lint:
	@if [ ! -x "$(golangci_lint_cmd)" ]; then \
		echo "golangci-lint not found at $(golangci_lint_cmd)"; \
		echo "Run 'make install-lint' first."; \
		exit 1; \
	fi
	@echo "--> Using golangci-lint binary: $(golangci_lint_cmd)"
	@$(golangci_lint_cmd) version

check-staticcheck:
	@if [ ! -x "$(staticcheck_cmd)" ]; then \
		echo "staticcheck not found at $(staticcheck_cmd)"; \
		echo "Run 'make install-staticcheck' first."; \
		exit 1; \
	fi
	@echo "--> Using staticcheck binary: $(staticcheck_cmd)"
	@$(staticcheck_cmd) -version

hooks:
	@if [ ! -x "$(lefthook_cmd)" ]; then \
		echo "--> Installing lefthook $(LEFTHOOK_VERSION) into $$(dirname "$(lefthook_cmd)")"; \
		$(GO) install github.com/evilmartians/lefthook@$(LEFTHOOK_VERSION); \
	else \
		echo "--> Using lefthook binary: $(lefthook_cmd)"; \
	fi
	@$(lefthook_cmd) install

lint: check-go-env check-lint
	@echo "--> Running incremental linter"
	@$(INCREMENTAL_LINT_SCRIPT) "$(golangci_lint_cmd)" 10m

lint-fix: check-go-env check-lint
	@echo "--> Running incremental linter with fixes"
	@$(INCREMENTAL_LINT_SCRIPT) "$(golangci_lint_cmd)" 10m --fix --out-format=tab --issues-exit-code=0

lint-all: check-go-env check-lint
	@echo "--> Running full linter"
	@$(golangci_lint_cmd) run --timeout=15m ./...

lint-fix-all: check-go-env check-lint
	@echo "--> Running full linter with fixes"
	@$(golangci_lint_cmd) run --fix --timeout=15m --out-format=tab --issues-exit-code=0 ./...

pre-commit-staged:
	@./scripts/pre-commit.sh

###############################################################################
###                        Docker                                           ###
###############################################################################
DOCKER := $(shell which docker)
DOCKER_IMAGE := mocachain/moca-cmd
COMMIT_HASH := $(shell git rev-parse --short=7 HEAD)
DOCKER_TAG := $(COMMIT_HASH)

build-docker:
	@if [ -n "$(GITHUB_TOKEN)" ]; then \
		echo "Building with GITHUB_TOKEN for private repositories..."; \
		$(DOCKER) build --progress=plain --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN) -t ${DOCKER_IMAGE}:${DOCKER_TAG} .; \
	else \
		echo "Building without GITHUB_TOKEN (public repositories only)..."; \
		$(DOCKER) build --progress=plain -t ${DOCKER_IMAGE}:${DOCKER_TAG} .; \
	fi
	$(DOCKER) tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
	$(DOCKER) tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:${COMMIT_HASH}

.PHONY: build-docker

###############################################################################
###                        Docker Compose                                   ###
###############################################################################
start-dc:
	docker compose up -d && docker attach moca-cmd
stop-dc:
	docker compose down --volumes

.PHONY: build-dcf start-dc stop-dc


###############################################################################
###                                Releasing                                ###
###############################################################################

PACKAGE_NAME:=github.com/mocachain/moca-cmd
GOLANG_CROSS_VERSION  = v1.23
GOPATH ?= $(HOME)/go
release-dry-run:
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-e GOPROXY=$(GOPROXY) \
		-e GOPRIVATE=$(GOPRIVATE) \
		-e GONOSUMDB=$(GONOSUMDB) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v ${GOPATH}/pkg:/go/pkg \
		$(GITCONFIG_MOUNT) \
		$(LOCAL_REPLACE_MOUNTS) \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		--clean --skip validate --skip publish --snapshot

release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-e GOPROXY=$(GOPROXY) \
		-e GOPRIVATE=$(GOPRIVATE) \
		-e GONOSUMDB=$(GONOSUMDB) \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		$(GITCONFIG_MOUNT) \
		$(LOCAL_REPLACE_MOUNTS) \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --clean --skip validate

.PHONY: release-dry-run release
