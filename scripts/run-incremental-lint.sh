#!/bin/bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

golangci_lint_bin="${1:?missing golangci-lint binary path}"
timeout="${2:?missing timeout}"
shift 2

changed_dirs=""

while IFS= read -r changed_file; do
  dir="$(dirname "$changed_file")"
  if [[ "$dir" == "." ]]; then
    target="."
  else
    target="./$dir"
  fi

  case " $changed_dirs " in
    *" $target "*) ;;
    *) changed_dirs="$changed_dirs $target" ;;
  esac
done < <((git diff --name-only --diff-filter=ACMR HEAD -- '*.go'; git ls-files --others --exclude-standard -- '*.go') | awk 'NF')

if [[ -z "${changed_dirs## }" ]]; then
  echo "No changed Go files detected; skipping incremental golangci-lint."
  exit 0
fi

for target in $changed_dirs; do
  echo "--> Linting $target"
  "$golangci_lint_bin" run "$@" --timeout="$timeout" "$target"
done
