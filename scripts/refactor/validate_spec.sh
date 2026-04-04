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

required_sections=(
  "## 背景/问题"
  "## 范围"
  "## 目标"
  "## 非目标"
  "## 约束"
  "## 验收"
  "## 输出物"
)

missing=0
for section in "${required_sections[@]}"; do
  if ! rg -q "^${section}$" "$spec_file"; then
    echo "[MISSING] ${section}"
    missing=1
  fi
done

if ! rg -q "go test|harness|测试" "$spec_file"; then
  echo "[MISSING] 验收中未检测到测试命令或关键字(go test/harness/测试)"
  missing=1
fi

if [ "$missing" -ne 0 ]; then
  echo "Spec validation failed: ${spec_file}"
  exit 1
fi

echo "Spec validation passed: ${spec_file}"
