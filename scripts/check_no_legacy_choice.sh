#!/usr/bin/env bash
# 禁止已废弃的 Choice / Interrupt 路由符号回流。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
PAT='handleWeakChoiceInput|dispatchChoiceInputByType|dispatchChoiceRouteByType|LegacyHandleChoice|interruptActionRules|interruptPromptBuilders|registeredSequentialCardChoiceRemainingCount|runChoiceTypeStep|runChoiceRouteStep'
LEGACY_UI_PAT='Prompt\.Cancelable|ctx\["cancelable"\]|labelMatchesMarkers|playerPromptMarkers|prompt_option_resolve_by_label'
if command -v rg >/dev/null 2>&1; then
  if rg -n "$PAT" --glob '*.go' internal/engine internal/engine/runtime; then
    echo "FAIL: forbidden legacy symbols matched above" >&2
    exit 1
  fi
  if rg -n "$LEGACY_UI_PAT" internal web/src web/e2e; then
    echo "FAIL: forbidden legacy UI/protocol symbols matched above" >&2
    exit 1
  fi
else
  matches="$(grep -R -n -E "$PAT" internal/engine internal/engine/runtime --include='*.go' 2>/dev/null || true)"
  if [ -n "$matches" ]; then
    printf '%s\n' "$matches" >&2
    echo "FAIL: forbidden legacy symbols matched above" >&2
    exit 1
  fi
  legacy_matches="$(grep -R -n -E "$LEGACY_UI_PAT" internal web/src web/e2e 2>/dev/null || true)"
  if [ -n "$legacy_matches" ]; then
    printf '%s\n' "$legacy_matches" >&2
    echo "FAIL: forbidden legacy UI/protocol symbols matched above" >&2
    exit 1
  fi
fi
echo OK
