SHELL := /bin/bash

# Load .env file if it exists
-include .env
export

# Keep local Go commands on the repository toolchain even when a newer
# system-wide Go release is installed.
GO_TOOLCHAIN ?= go1.23.12
GO := env GOTOOLCHAIN=$(GO_TOOLCHAIN) go

# Configure git to use HTTPS+Token for private repositories if GITHUB_TOKEN is set
ifdef GITHUB_TOKEN
  $(shell git config --global url."https://$(GITHUB_TOKEN):@github.com/".insteadOf "https://github.com/" 2>/dev/null)
endif

.PHONY: all build

build:
	$(GO) build -o ./build/moca-cmd cmd/*.go

golangci_version=v1.64.8
LEFTHOOK_VERSION=v1.11.3
GO_GOBIN := $(shell $(GO) env GOBIN)
GO_GOPATH := $(shell $(GO) env GOPATH)

ifeq ($(GO_GOBIN),)
golangci_lint_cmd=$(GO_GOPATH)/bin/golangci-lint
lefthook_cmd=$(GO_GOPATH)/bin/lefthook
else
golangci_lint_cmd=$(GO_GOBIN)/golangci-lint
lefthook_cmd=$(GO_GOBIN)/lefthook
endif

define lint_changed_dirs
changed_dirs=""; \
while IFS= read -r changed_file; do \
	dir="$$(dirname "$$changed_file")"; \
	if [ "$$dir" = "." ]; then \
		target="."; \
	else \
		target="./$$dir"; \
	fi; \
	case " $$changed_dirs " in \
		*" $$target "*) ;; \
		*) changed_dirs="$$changed_dirs $$target" ;; \
	esac; \
	done < <((git diff --name-only --diff-filter=ACMR HEAD -- '*.go'; git ls-files --others --exclude-standard -- '*.go') | awk 'NF'); \
if [ -z "$${changed_dirs## }" ]; then \
	echo "No changed Go files detected; skipping incremental golangci-lint."; \
	exit 0; \
fi; \
for target in $$changed_dirs; do \
	echo "--> Linting $$target"; \
	$(golangci_lint_cmd) run $(1) --timeout=10m "$$target" || exit $$?; \
done
endef

lint:
	@echo "--> Running incremental linter"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@bash -lc '$(call lint_changed_dirs,)'

lint-fix:
	@echo "--> Running incremental linter with fixes"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@bash -lc '$(call lint_changed_dirs,--fix --out-format=tab --issues-exit-code=0)'

lint-all:
	@echo "--> Running full linter"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=15m ./...

lint-fix-all:
	@echo "--> Running full linter with fixes"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --fix --timeout=15m --out-format=tab --issues-exit-code=0 ./...

hooks:
	@$(GO) install github.com/evilmartians/lefthook@$(LEFTHOOK_VERSION)
	@$(lefthook_cmd) install

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

PACKAGE_NAME:=github.com/evmos/evmos
GOLANG_CROSS_VERSION  = v1.23
GOPATH ?= $(HOME)/go
release-dry-run:
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v ${GOPATH}/pkg:/go/pkg \
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
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --clean --skip validate

.PHONY: release-dry-run release