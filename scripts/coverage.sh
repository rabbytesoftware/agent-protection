#!/usr/bin/env bash
# coverage.sh [minimum-percent]
# Runs every Go test with coverage and fails when total statement coverage is
# below the minimum (95 by default, Rabbyte's standard). Struct-only and
# declaration-only files have no statements, so they never count; mock and
# fake files are excluded by name.
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
