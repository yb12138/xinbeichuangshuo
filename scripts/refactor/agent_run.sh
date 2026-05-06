#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "Usage: $0 <run-dir> [agent] [cwd]"
  echo "Example: $0 artifacts/refactor_runs/20260404-120000-demo codex /Users/yb/xinbeichuangshuo"
  exit 1
fi

run_dir="$1"
agent="${2:-codex}"
cwd="${3:-$(pwd)}"

prompt_file="${run_dir}/codex_prompt.md"
log_dir="${run_dir}/logs"
mkdir -p "${log_dir}"

if [ ! -f "$prompt_file" ]; then
  echo "Prompt file not found: ${prompt_file}"
  echo "Run: make refactor-prompt SPEC=${run_dir}/spec.md"
  exit 1
fi

echo "Running agent '${agent}' in cwd: ${cwd}"
echo "Prompt: ${prompt_file}"

case "$agent" in
  codex)
    codex exec --full-auto -C "${cwd}" -o "${run_dir}/agent_last_message.md" - \
      < "${prompt_file}" | tee "${log_dir}/agent.codex.log"
    ;;
  claude)
    cat "${prompt_file}" \
      | claude -p --permission-mode acceptEdits --output-format text \
      | tee "${log_dir}/agent.claude.log"
    cp "${log_dir}/agent.claude.log" "${run_dir}/agent_last_message.md"
    ;;
  *)
    echo "Unsupported agent: ${agent}"
    echo "Available agents: codex | claude"
    exit 1
    ;;
esac

echo "Agent run completed."
echo "Last message: ${run_dir}/agent_last_message.md"
