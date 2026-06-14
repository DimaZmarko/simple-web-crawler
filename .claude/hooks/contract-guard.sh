#!/usr/bin/env bash
# Warn (never block) when the API contract is edited, so the generated TS client
# and Go types don't silently drift. Receives PostToolUse hook JSON on stdin and,
# when the edited file is the contract, feeds a reminder back to Claude via the
# PostToolUse additionalContext channel. The edit has already happened — this only
# nudges; it does not undo or block anything.
set -euo pipefail

INPUT="$(cat)"
FILE="$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // .tool_input.path // empty')"
[ -z "$FILE" ] && exit 0

case "$FILE" in
  */packages/api-contract/openapi.yaml | packages/api-contract/openapi.yaml)
    MSG="openapi.yaml changed. Before relying on the client/types, run 'make gen-contract' to regenerate the TS client + Go types (golden rule #2: never hand-edit generated code). Contract changes are normally owned by the api-contract-keeper agent (golden rule #1)."
    jq -n --arg msg "$MSG" \
      '{hookSpecificOutput: {hookEventName: "PostToolUse", additionalContext: $msg}}'
    ;;
esac
exit 0
