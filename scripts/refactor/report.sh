#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "Usage: $0 <run-dir>"
  exit 1
fi

run_dir="$1"
if [ ! -d "$run_dir" ]; then
  echo "Run directory not found: $run_dir"
  exit 1
fi

report_file="${run_dir}/report.md"
log_dir="${run_dir}/logs"

status_for() {
  local log_file="$1"
  if [ ! -f "$log_file" ]; then
    echo "NOT RUN"
    return
  fi
  if rg -q "^FAIL|\\[FAIL\\]|build failed|^--- FAIL:" "$log_file"; then
    echo "FAIL"
  else
    echo "PASS"
  fi
}

{
  echo "# Refactor Acceptance Report"
  echo ""
  echo "Generated at: $(date '+%Y-%m-%d %H:%M:%S')"
  echo ""
  echo "## Inputs"
  echo "- Spec: ${run_dir}/spec.md"
  echo "- Prompt: ${run_dir}/codex_prompt.md"
  echo ""
  echo "## Git Summary"
  echo '```bash'
  git status --short
  echo '```'
  echo ""
  echo "## Harness Results"
  for name in compile tests engine server rules; do
    echo "- ${name}: $(status_for "${log_dir}/${name}.log")"
  done
  echo ""
  echo "## Key Failures (first 80 lines each)"
  for name in compile tests engine server rules; do
    file="${log_dir}/${name}.log"
    if [ -f "$file" ] && rg -q "^FAIL|\\[FAIL\\]|build failed|^--- FAIL:" "$file"; then
      echo ""
      echo "### ${name}"
      echo '```text'
      sed -n '1,80p' "$file"
      echo '```'
    fi
  done
} > "$report_file"

echo "Report generated: ${report_file}"
