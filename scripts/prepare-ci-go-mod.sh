#!/bin/bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

cache_root="$repo_root/.cache/ci-go"
mkdir -p "$cache_root/go-build" "$cache_root/mod"
export GOCACHE="$cache_root/go-build"
export GOMODCACHE="$cache_root/mod"
export GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}"

if ! grep -qE '=> \.\./' go.mod; then
  echo "go.mod already uses CI-compatible dependency replacements."
  exit 0
fi

echo "Rewriting local sibling replacements for CI..."
go mod edit \
  -replace=github.com/mocachain/moca/v2=github.com/mocachain/moca/v12@v12.2.0-rc4.0.20260320060615-9f8f08384ec3 \
  -replace=github.com/mocachain/moca-common/go=github.com/mocachain/moca-common/go@v1.2.0-rc1.0.20260320043131-de7b6add70a3 \
  -replace=github.com/mocachain/moca-go-sdk=github.com/mocachain/moca-go-sdk@v1.2.0-rc1.0.20260320043142-d578d73e2599

go mod tidy
go mod download
