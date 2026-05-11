#!/bin/bash
set -euo pipefail

mode="${1:?missing mode}"
staticcheck_bin="${2:?missing staticcheck binary path}"

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

go_toolchain="$(awk '/^toolchain / { print $2; exit }' go.mod)"
if [[ -z "$go_toolchain" ]]; then
  echo "Unable to determine Go toolchain from go.mod" >&2
  exit 1
fi
export GOTOOLCHAIN="$go_toolchain"
export GIT_TERMINAL_PROMPT=0

if [[ ! -x "$staticcheck_bin" ]]; then
  echo "staticcheck binary not found at $staticcheck_bin; run 'make install-staticcheck' first." >&2
  exit 1
fi

case "$mode" in
  local)
    change_list_cmd="{ git diff --name-only --diff-filter=ACMR HEAD; git ls-files --others --exclude-standard; }"
    ;;
  staged)
    change_list_cmd="git diff --cached --name-only --diff-filter=ACMR"
    ;;
  *)
    echo "unknown mode: $mode" >&2
    exit 1
    ;;
esac

module_changed="$(
  eval "$change_list_cmd" | grep -E '(^|/)(go\.mod|go\.sum)$' || true
)"

if [[ -n "$module_changed" ]]; then
  echo "Detected go.mod/go.sum changes; running go mod tidy with ${go_toolchain}..."
  go mod tidy

  if ! git diff --quiet -- go.mod go.sum; then
    echo "go.mod or go.sum changed after go mod tidy; please review and re-stage them." >&2
    exit 1
  fi
else
  echo "No go.mod/go.sum changes detected; skipping go mod tidy."
fi

changed_dirs="$(
  eval "$change_list_cmd" | grep '\.go$' | xargs -n1 dirname 2>/dev/null | sed 's#^\.$#./#' | sort -u || true
)"

if [[ -z "${changed_dirs}" ]]; then
  if [[ -n "$module_changed" ]]; then
    echo "No changed Go files detected; running staticcheck on entire repository because module files changed."
    "$staticcheck_bin" ./...
  else
    echo "No changed Go files detected; skipping staticcheck."
  fi
  exit 0
fi

if [[ -n "$module_changed" ]]; then
  echo "Running staticcheck on entire repository..."
  "$staticcheck_bin" ./...
  exit 0
fi

echo "Running staticcheck on changed Go packages..."
while IFS= read -r dir; do
  [[ -z "$dir" ]] && continue
  echo "--> staticcheck $dir"
  "$staticcheck_bin" "$dir"
done <<< "$changed_dirs"
