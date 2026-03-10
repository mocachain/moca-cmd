#!/usr/bin/env bash
# Run from moca-cmd repo root. Uses locally built ./build/moca-cmd and devnet config.
set -e
CMD="$(cd "$(dirname "$0")/../.." && pwd)/build/moca-cmd"
HOME_DIR="$(cd "$(dirname "$0")" && pwd)"
PWFILE="$HOME_DIR/testkey/password.txt"

echo "=== devnet bucket create ==="
"$CMD" --home "$HOME_DIR" --passwordfile "$PWFILE" bucket create \
  --tags='[{"key":"key1","value":"value1"},{"key":"key2","value":"value2"}]' \
  bucket1

echo "=== devnet object put ==="
"$CMD" --home "$HOME_DIR" --passwordfile "$PWFILE" object put \
  --tags='[{"key":"key1","value":"value1"},{"key":"key2","value":"value2"}]' \
  --contentType "application/octet-stream" \
  "$(cd "$(dirname "$0")/../.." && pwd)/go.mod" \
  bucket1/go.mod

echo "=== done ==="
