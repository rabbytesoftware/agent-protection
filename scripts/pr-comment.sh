#!/usr/bin/env bash
# pr-comment.sh <ap-binary> <ap-report.json> <ap-exit-code> <component>
# Creates or updates the Agent Protection comment for one component folder on
# the current pull request, so a monorepo gets one comment per component. Needs GH_TOKEN with pull-requests: write, plus the standard
# GITHUB_* variables Actions provides.
set -euo pipefail

AP="${1:?usage: pr-comment.sh <ap-binary> <ap-report.json> <ap-exit-code> <component>}"
REPORT="${2:?usage: pr-comment.sh <ap-binary> <ap-report.json> <ap-exit-code> <component>}"
EXIT_CODE="${3:?usage: pr-comment.sh <ap-binary> <ap-report.json> <ap-exit-code> <component>}"
COMPONENT="${4:?usage: pr-comment.sh <ap-binary> <ap-report.json> <ap-exit-code> <component>}"

MARKER="<!-- agent-protection:${COMPONENT} -->"
PR=$(jq -r '.pull_request.number // empty' "$GITHUB_EVENT_PATH")
if [ -z "$PR" ]; then
  echo "Not a pull request event; no comment posted."
  exit 0
fi

RUN_URL="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"
BODY=$(mktemp)
{
  echo "$MARKER"
  if [ "$COMPONENT" != "." ]; then
    echo "**Component:** \`${COMPONENT}\`"
    echo
  fi
  if [ -f "$REPORT" ] && "$AP" markdown "$REPORT"; then
    echo
    echo "[Workflow run and full ap-report.json artifact]($RUN_URL)"
  else
    echo "### Agent Protection could not run (exit $EXIT_CODE)"
    echo
    echo "See the [job log]($RUN_URL)."
  fi
} > "$BODY"

EXISTING=$(gh api --paginate "repos/${GITHUB_REPOSITORY}/issues/${PR}/comments" \
  --jq ".[] | select(.body | startswith(\"$MARKER\")) | .id" | head -1)
if [ -n "$EXISTING" ]; then
  gh api -X PATCH "repos/${GITHUB_REPOSITORY}/issues/comments/${EXISTING}" -F "body=@${BODY}" > /dev/null
  echo "Updated Agent Protection comment ${EXISTING} on PR #${PR}."
  exit 0
fi
gh api -X POST "repos/${GITHUB_REPOSITORY}/issues/${PR}/comments" -F "body=@${BODY}" > /dev/null
echo "Posted Agent Protection comment on PR #${PR}."
