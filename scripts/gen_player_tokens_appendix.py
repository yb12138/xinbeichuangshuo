#!/usr/bin/env python3
"""Generate docs/player_tokens_keys_appendix.md — Player.Tokens string key inventory."""

from __future__ import annotations

import re
import subprocess
from collections import defaultdict
from datetime import date
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]


def run_grep(args: list[str]) -> str:
    r = subprocess.run(
        ["grep", "-rn", *args, "internal", "--include=*.go"],
        cwd=REPO,
        capture_output=True,
        text=True,
    )
    return r.stdout


def sort_key_entry(entry: str) -> tuple:
    parts = entry.split(":", 2)
    if len(parts) >= 2:
        try:
            return (parts[0], int(parts[1]))
        except ValueError:
            pass
    return (entry, 0)


def extract_key(line: str) -> str | None:
    m = re.search(r'Tokens\["([^"]+)"\]', line)
    if m:
        return m.group(1)
    m = re.search(r'delete\([^,]*Tokens,\s*"([^"]+)"', line)
    if m:
        return m.group(1)
    m = re.search(r'(?:getToken|setToken|addToken)\([^,]+,\s*"([^"]+)"', line)
    if m:
        return m.group(1)
    return None


def main() -> None:
    lines: list[str] = []
    lines.extend(run_grep(['Tokens\\["[^"]*"\\]']).splitlines())
    lines.extend(run_grep(['delete([^)]*Tokens,\\s*"[^"]*"']).splitlines())
    lines.extend(run_grep(['-E', '(getToken|setToken|addToken)\\([^,]+,\\s*"[^"]+"']).splitlines())

    seen: set[str] = set()
    uniq: list[str] = []
    for ln in lines:
        if not ln.strip():
            continue
        if ln not in seen:
            seen.add(ln)
            uniq.append(ln)

    by_key: dict[str, list[str]] = defaultdict(list)
    bad: list[str] = []
    for ln in uniq:
        k = extract_key(ln)
        if k:
            by_key[k].append(ln)
        else:
            bad.append(ln)

    for k in by_key:
        by_key[k].sort(key=sort_key_entry)

    out: list[str] = []
    out.append("# 附录：`Player.Tokens` 代码 key 索引（自动生成）\n")
    out.append(
        "\n本文件由 `scripts/gen_player_tokens_appendix.py` 生成，用于盘点 `map[string]int` 使用的字符串 key。"
        " **动态 key**（如 `turnScopedResetKeys`）见「动态下标」一节。\n"
    )
    out.append(
        "\n**迁出 Tokens 的优化方案**（流程/回合/锁、UI 派生与可见性等）："
        "见 [player_tokens_runtime_refactor_plan.md](player_tokens_runtime_refactor_plan.md)。\n"
    )
    out.append(f"\n- **生成日期**：{date.today().isoformat()}\n")
    out.append("- **扫描范围**：`internal/**/*.go`\n")
    out.append(
        "- **匹配模式**：`Tokens[\"...\"]`、`delete(...Tokens, \"...\")`、"
        "`getToken`/`setToken`/`addToken(..., \"...\")`\n"
    )
    out.append(f"- **去重后条目数**：{len(uniq)} 处引用；**不同 key 数**：{len(by_key)}\n")
    out.append("\n**重新生成**：在仓库根目录执行 `python3 scripts/gen_player_tokens_appendix.py`\n")

    out.append("\n## 1. 按 key 字母序（含 `路径:行号`）\n")
    for k in sorted(by_key.keys(), key=str.lower):
        out.append(f"\n### `{k}`\n")
        for entry in by_key[k]:
            parts = entry.split(":", 2)
            if len(parts) >= 3:
                filepath, lineno, code = parts[0], parts[1], parts[2]
                c = code.strip().replace("`", "'")
                if len(c) > 140:
                    c = c[:137] + "..."
                out.append(f"- `{filepath}:{lineno}` — `{c}`\n")
            else:
                out.append(f"- `{entry.strip()}`\n")

    out.append("\n## 2. 动态下标 `player.Tokens[key]`\n\n")
    out.append(
        "以下位置通过变量写入 `Tokens`："
        "`applyRoleDefaults` 中 `cfg.tokens` 的键，或 `resetTurnScopedPlayerTokens` 中 `turnScopedResetKeys` 的键。\n\n"
    )
    dyn = sorted(
        {x for x in run_grep(["player\\.Tokens\\[key\\]"]).splitlines() if x.strip()},
        key=sort_key_entry,
    )
    for d in dyn:
        parts = d.split(":", 2)
        if len(parts) >= 3:
            c = parts[2].strip().replace("`", "'")
            out.append(f"- `{parts[0]}:{parts[1]}` — `{c}`\n")
        else:
            out.append(f"- `{d}`\n")

    rd = REPO / "internal/engine/role_defaults.go"
    out.append("\n## 3. `role_defaults.go` 默认 token 初始化（map 字面量）\n\n")
    out.append(
        "下列行在 `roleDefaultConfigs` 的 `tokens: map[string]int{...}` 中出现；"
        "多数 key 亦在 §1 中有引用，此处标出**初始化默认值**来源。\n\n"
    )
    if rd.is_file():
        text = rd.read_text(encoding="utf-8")
        for i, line in enumerate(text.splitlines(), 1):
            if re.match(r'^\s+"[a-z0-9_]+":\s*[0-9]', line):
                c = line.strip().replace("`", "'")
                out.append(f"- `{rd.relative_to(REPO)}:{i}` — `{c}`\n")

    if bad:
        out.append("\n## 4. 未能解析 key 的行（需人工核对）\n\n")
        for b in bad[:80]:
            out.append(f"- `{b}`\n")
        if len(bad) > 80:
            out.append(f"\n… 共 {len(bad)} 行\n")

    out_path = REPO / "docs/player_tokens_keys_appendix.md"
    out_path.write_text("".join(out), encoding="utf-8")
    print(f"Wrote {out_path} ({out_path.stat().st_size} bytes)")
    print(f"keys={len(by_key)} lines={len(uniq)} bad={len(bad)}")


if __name__ == "__main__":
    main()
