#!/bin/bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

cache_root="$repo_root/.cache/pre-commit"
mkdir -p "$cache_root/go-build" "$cache_root/mod"
export GOCACHE="$cache_root/go-build"
export GOMODCACHE="$cache_root/mod"
export GOPROXY="${PRE_COMMIT_GOPROXY:-https://proxy.golang.org,direct}"

go_toolchain="$(awk '/^toolchain / { print $2; exit }' go.mod)"
if [[ -z "$go_toolchain" ]]; then
  echo "Unable to determine Go toolchain from go.mod" >&2
  exit 1
fi
export GOTOOLCHAIN="$go_toolchain"

if ! go version >/dev/null 2>&1; then
  echo "Unable to activate Go toolchain ${go_toolchain}; falling back to local toolchain." >&2
  export GOTOOLCHAIN=local
fi

staticcheck_version="v0.6.1"

echo "Running go mod tidy with ${go_toolchain}..."
if grep -qE '=> \.\./' go.mod; then
  echo "Skipping go mod tidy because go.mod uses local sibling replacements."
else
  go mod tidy

  if ! git diff --quiet -- go.mod go.sum; then
    echo "go.mod or go.sum changed after go mod tidy; please review and re-stage them." >&2
    exit 1
  fi
fi

echo "Installing staticcheck ${staticcheck_version}..."
if ! go install honnef.co/go/tools/cmd/staticcheck@"$staticcheck_version"; then
  echo "Unable to install staticcheck; skipping because dependencies are not reachable." >&2
  exit 0
fi

if grep -qE '=> \.\./' go.mod; then
  echo "Skipping staticcheck because go.mod uses local sibling replacements during multi-repo migration."
  exit 0
fi

gobin="$(go env GOBIN)"
if [[ -n "$gobin" ]]; then
  staticcheck_bin="$gobin/staticcheck"
else
  gopath="$(go env GOPATH)"
  staticcheck_bin="$gopath/bin/staticcheck"
fi

echo "Running staticcheck..."
"$staticcheck_bin" ./...
