#!/bin/bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

mode="${1:-incremental}"
fix_mode="${2:-check}"

go_toolchain="$(awk '/^toolchain / { print $2; exit }' go.mod)"
if [[ -z "$go_toolchain" ]]; then
  echo "Unable to determine Go toolchain from go.mod" >&2
  exit 1
fi
export GOTOOLCHAIN="$go_toolchain"

golangci_version="${GOLANGCI_LINT_VERSION:-v1.64.8}"
gobin="$(go env GOBIN)"
if [[ -n "$gobin" ]]; then
  golangci_lint_bin="$gobin/golangci-lint"
else
  gopath="$(go env GOPATH)"
  golangci_lint_bin="$gopath/bin/golangci-lint"
fi

go install github.com/golangci/golangci-lint/cmd/golangci-lint@"$golangci_version"

golangci_args=()
if [[ "$fix_mode" == "fix" ]]; then
  golangci_args+=(--fix --out-format=tab --issues-exit-code=0)
fi

run_full() {
  if [[ "${#golangci_args[@]}" -gt 0 ]]; then
    "$golangci_lint_bin" run --timeout=15m "${golangci_args[@]}" ./...
  else
    "$golangci_lint_bin" run --timeout=15m ./...
  fi
}

run_incremental() {
  local base_ref=""
  local merge_base=""
  local -a changed_go_files=()
  local -a changed_targets=()
  local changed_file=""
  local target_dir=""
  local seen_targets=""

  if [[ -n "${LINT_BASE_REF:-}" ]]; then
    base_ref="${LINT_BASE_REF}"
  elif [[ -n "${GITHUB_BASE_REF:-}" ]]; then
    base_ref="origin/${GITHUB_BASE_REF}"
  elif git rev-parse --verify origin/main >/dev/null 2>&1; then
    base_ref="origin/main"
  elif git rev-parse --verify HEAD~1 >/dev/null 2>&1; then
    base_ref="HEAD~1"
  fi

  if [[ -z "$base_ref" ]]; then
    echo "Unable to determine lint base reference; running full lint."
    run_full
    return
  fi

  if [[ "$base_ref" == "HEAD~1" ]]; then
    merge_base="HEAD~1"
  else
    merge_base="$(git merge-base HEAD "$base_ref")"
  fi

  while IFS= read -r changed_file; do
    changed_go_files+=("$changed_file")
  done < <(git diff --name-only --diff-filter=ACMR "${merge_base}...HEAD" -- '*.go')
  if [[ "${#changed_go_files[@]}" -eq 0 ]]; then
    echo "No changed Go files detected; skipping incremental golangci-lint."
    return
  fi

  for changed_file in "${changed_go_files[@]}"; do
    target_dir="./$(dirname "$changed_file")"
    if [[ "$target_dir" == "./." ]]; then
      target_dir="."
    fi
    case " ${seen_targets} " in
      *" ${target_dir} "*) ;;
      *)
        changed_targets+=("$target_dir")
        seen_targets="${seen_targets} ${target_dir}"
        ;;
    esac
  done

  if [[ "${#golangci_args[@]}" -gt 0 ]]; then
    "$golangci_lint_bin" run --timeout=10m "${golangci_args[@]}" "${changed_targets[@]}"
  else
    "$golangci_lint_bin" run --timeout=10m "${changed_targets[@]}"
  fi
}

case "$mode" in
  incremental)
    run_incremental
    ;;
  full)
    run_full
    ;;
  *)
    echo "Unsupported lint mode: $mode" >&2
    echo "Usage: scripts/lint.sh [incremental|full] [check|fix]" >&2
    exit 1
    ;;
esac
