#!/usr/bin/env bash
# Auto-format a file after Claude edits it. Receives hook JSON on stdin.
set -euo pipefail
INPUT="$(cat)"
FILE="$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // .tool_input.path // empty')"
[ -z "$FILE" ] && exit 0

# Normalize to absolute path (tool may pass relative or absolute paths).
if [[ "$FILE" != /* ]]; then
  FILE="$(cd "$(dirname "$FILE")" && pwd)/$(basename "$FILE")"
fi
[ -f "$FILE" ] || exit 0

REPO_ROOT="$(git -C "$(dirname "$FILE")" rev-parse --show-toplevel 2>/dev/null || pwd)"

case "$FILE" in
  *.go)
    command -v goimports >/dev/null 2>&1 && goimports -w "$FILE" || gofmt -w "$FILE"
    ;;
  *.ts|*.tsx|*.js|*.jsx|*.json|*.css)
    case "$FILE" in
      "$REPO_ROOT"/apps/web/*)
        REL="${FILE#"$REPO_ROOT"/apps/web/}"
        if [ -f "$REPO_ROOT/apps/web/node_modules/.bin/prettier" ] && command -v pnpm >/dev/null 2>&1; then
          (cd "$REPO_ROOT/apps/web" && pnpm exec prettier --write "$REL") >/dev/null 2>&1 || true
        elif command -v docker >/dev/null 2>&1; then
          docker run --rm -v "$REPO_ROOT/apps/web:/app" -w /app node:22-alpine \
            sh -c "corepack enable && pnpm exec prettier --write '$REL'" >/dev/null 2>&1 || true
        fi
        ;;
    esac
    ;;
esac
exit 0
