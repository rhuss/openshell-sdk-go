#!/usr/bin/env bash
# Triage state management for PR review comment processing.
# Usage:
#   spex-triage-state.sh init <pr>
#   spex-triage-state.sh get <pr> <comment_id>
#   spex-triage-state.sh set <pr> <comment_id> <action> [reply_id]

set -euo pipefail

STATE_DIR=".specify/triage"
ACTION="$1"
PR="$2"
STATE_FILE="$STATE_DIR/pr-${PR}.json"

case "$ACTION" in
  init)
    mkdir -p "$STATE_DIR"
    if [ ! -f "$STATE_FILE" ]; then
      echo '{"pr":'"$PR"',"comments":{}}' > "$STATE_FILE"
      echo "Initialized triage state for PR #$PR"
    else
      echo "Triage state already exists for PR #$PR"
    fi
    ;;
  get)
    COMMENT_ID="$3"
    if [ ! -f "$STATE_FILE" ]; then
      echo ""
      exit 0
    fi
    jq -r --arg id "$COMMENT_ID" '.comments[$id] // empty' "$STATE_FILE"
    ;;
  set)
    COMMENT_ID="$3"
    COMMENT_ACTION="$4"
    REPLY_ID="${5:-}"
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    jq --arg id "$COMMENT_ID" \
       --arg action "$COMMENT_ACTION" \
       --arg reply "$REPLY_ID" \
       --arg ts "$TIMESTAMP" \
       '.comments[$id] = {"action": $action, "replyId": $reply, "handledAt": $ts}' \
       "$STATE_FILE" > "${STATE_FILE}.tmp" && mv "${STATE_FILE}.tmp" "$STATE_FILE"
    ;;
  *)
    echo "Unknown action: $ACTION" >&2
    exit 1
    ;;
esac
