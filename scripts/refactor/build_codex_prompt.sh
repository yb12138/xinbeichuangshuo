#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "Usage: $0 <spec-file>"
  exit 1
fi

spec_file="$1"
if [ ! -f "$spec_file" ]; then
  echo "Spec file not found: $spec_file"
  exit 1
fi

out_dir="$(dirname "$spec_file")"
out_file="${out_dir}/codex_prompt.md"
template_file="specs/refactor/codex_prompt_template.md"

if [ ! -f "$template_file" ]; then
  echo "Template not found: ${template_file}"
  exit 1
fi

{
  cat "$template_file"
  echo ""
  echo "## Refactor Spec"
  echo ""
  cat "$spec_file"
} > "$out_file"

echo "Generated: ${out_file}"
echo "Use this file content as your Codex task input."
