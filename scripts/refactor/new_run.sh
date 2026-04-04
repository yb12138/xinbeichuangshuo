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

slug="$(basename "$spec_file" .md | tr ' ' '-' | tr '[:upper:]' '[:lower:]')"
ts="$(date '+%Y%m%d-%H%M%S')"
run_id="${ts}-${slug}"
run_dir="artifacts/refactor_runs/${run_id}"
mkdir -p "${run_dir}/logs"

cp "$spec_file" "${run_dir}/spec.md"

cat > "${run_dir}/README.md" <<EOF
# Refactor Run: ${run_id}

## Files
- spec: spec.md
- codex prompt: codex_prompt.md
- harness logs: logs/
- report: report.md

## Steps
1. Validate spec:
   \`\`\`bash
   make refactor-validate SPEC=${run_dir}/spec.md
   \`\`\`
2. Build Codex prompt:
   \`\`\`bash
   make refactor-prompt SPEC=${run_dir}/spec.md
   \`\`\`
3. Paste \`codex_prompt.md\` into Codex and let Codex implement changes.
4. Run acceptance harness:
   \`\`\`bash
   make harness PROFILE=smoke RUN=${run_dir}
   # optional deeper checks
   make harness PROFILE=full RUN=${run_dir}
   \`\`\`
5. Generate final report:
   \`\`\`bash
   make report RUN=${run_dir}
   \`\`\`
EOF

echo "Created run directory: ${run_dir}"
echo "Next step: make refactor-validate SPEC=${run_dir}/spec.md"
