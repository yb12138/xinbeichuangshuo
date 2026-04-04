#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "Usage: $0 <spec-file> [agent] [profile]"
  echo "Example: $0 specs/refactor/mainflow-skill-decouple.md codex smoke"
  exit 1
fi

spec_file="$1"
agent="${2:-codex}"
profile="${3:-smoke}"

if [ ! -f "$spec_file" ]; then
  echo "Spec file not found: $spec_file"
  exit 1
fi

new_output="$(bash scripts/refactor/new_run.sh "$spec_file")"
echo "$new_output"
run_dir="$(echo "$new_output" | awk -F': ' '/Created run directory:/ {print $2}' | tail -n1)"

if [ -z "$run_dir" ]; then
  echo "Failed to parse run directory from new_run output"
  exit 1
fi

bash scripts/refactor/validate_spec.sh "${run_dir}/spec.md"
bash scripts/refactor/build_codex_prompt.sh "${run_dir}/spec.md"
bash scripts/refactor/agent_run.sh "${run_dir}" "${agent}" "$(pwd)"
bash scripts/refactor/harness.sh "${profile}" "${run_dir}"
bash scripts/refactor/report.sh "${run_dir}"

echo ""
echo "Workflow completed."
echo "Run directory: ${run_dir}"
echo "Report: ${run_dir}/report.md"
