#!/bin/bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

go_toolchain="$(awk '/^toolchain / { print $2; exit }' go.mod)"
if [[ -z "$go_toolchain" ]]; then
  echo "Unable to determine Go toolchain from go.mod" >&2
  exit 1
fi
export GOTOOLCHAIN="$go_toolchain"

staticcheck_version="v0.6.1"

gobin="$(go env GOBIN)"
if [[ -n "$gobin" ]]; then
  staticcheck_bin="$gobin/staticcheck"
else
  gopath="$(go env GOPATH)"
  staticcheck_bin="$gopath/bin/staticcheck"
fi

echo "Running go mod tidy with ${go_toolchain}..."
go mod tidy

if ! git diff --quiet -- go.mod go.sum; then
  echo "go.mod or go.sum changed after go mod tidy; please review and re-stage them." >&2
  exit 1
fi

echo "Installing staticcheck ${staticcheck_version}..."
go install honnef.co/go/tools/cmd/staticcheck@"$staticcheck_version"

echo "Running staticcheck..."
"$staticcheck_bin" ./...
