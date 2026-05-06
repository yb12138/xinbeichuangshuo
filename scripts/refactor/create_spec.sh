#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 1 ]; then
  echo "Usage: $0 <spec-name>"
  echo "Example: $0 mainflow-skill-decouple"
  exit 1
fi

name="$1"
out="specs/refactor/${name}.md"
template="specs/refactor/spec.template.md"

if [ -f "$out" ]; then
  echo "Spec already exists: $out"
  exit 1
fi

if [ ! -f "$template" ]; then
  echo "Template not found: $template"
  exit 1
fi

cp "$template" "$out"
echo "Created spec: $out"
