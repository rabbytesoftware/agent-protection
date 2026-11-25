#!/usr/bin/env bash
set -euo pipefail

MINIMUM="${1:-95}"
PROFILE=$(mktemp)
FILTERED=$(mktemp)

go test -coverprofile="$PROFILE" ./...
grep -vE '(^|/)[^/:]*(mock|fake)[^/:]*\.go:' "$PROFILE" > "$FILTERED"
TOTAL=$(go tool cover -func="$FILTERED" | awk '/^total:/ { sub("%", "", $NF); print $NF }')

echo "Total statement coverage: ${TOTAL}% (minimum ${MINIMUM}%)"
if awk -v total="$TOTAL" -v minimum="$MINIMUM" 'BEGIN { exit !(total < minimum) }'; then
  echo "::error::Statement coverage ${TOTAL}% is below the ${MINIMUM}% minimum"
  exit 1
fi
