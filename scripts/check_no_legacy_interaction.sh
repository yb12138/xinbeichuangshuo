#!/usr/bin/env bash
# Guard interaction-boundary cleanup from regressing.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v rg >/dev/null 2>&1; then
  echo "FAIL: ripgrep (rg) is required for legacy interaction checks" >&2
  exit 1
fi

fail=0

check_absent() {
  local label="$1"
  local pattern="$2"
  shift 2
  local matches
  matches="$(rg -n "$pattern" "$@" 2>/dev/null || true)"
  if [[ -n "$matches" ]]; then
    echo "FAIL: $label" >&2
    printf '%s\n' "$matches" >&2
    fail=1
  fi
}

check_absent \
  "prompt cancellation must use cancel_policy, not cancelable" \
  'Prompt\.Cancelable|["'\'']cancelable["'\'']|ctx\["cancelable"\]' \
  internal web/src web/e2e

check_absent \
  "target picker must not resolve targets from labels or player names" \
  'labelMatchesMarkers|playerPromptMarkers|prompt_option_resolve_by_label|resolveOptionPlayerId.*label|target.*label.*fallback' \
  web/src

check_absent \
  "components/composables must not construct raw WebSocket wire envelopes" \
  'Cmd\s*:|["'\'']SubmitAction["'\'']|["'\'']RoomAction["'\'']|option_indexes|extra_args' \
  --glob '!**/__tests__/**' \
  web/src/components web/src/composables

check_absent \
  "prompt flow state must be created/moved through runtime helpers outside the model package" \
  'NewPromptFlowState|flow\.Advance\(' \
  --glob '!**/*_test.go' \
  internal/engine internal/server

check_absent \
  "runtime-backed flow advancement must not pass duplicate choice_type strings" \
  'AdvancePromptFlowRuntimeChoice\([^)]*,\s*"[^"]+"\)' \
  --glob '!**/*_test.go' \
  internal/engine

check_absent \
  "old-service/client compatibility wording should not remain in runtime or e2e fixtures" \
  '兼容旧服务端|旧服务端|旧客户端|向后兼容|legacy choice_type' \
  internal web/src web/e2e

allowed_choice_type_writes="$(cat <<'EOF'
1 internal/engine/player/adventurer/choices.go:adventurer_fraud_attack_element
1 internal/engine/player/bard/choices.go:bd_hope_mode
1 internal/engine/player/bard/choices.go:bd_hope_place_target
1 internal/engine/player/bard/choices.go:bd_hope_transfer_discard
1 internal/engine/player/bard/choices.go:bd_hope_transfer_target
1 internal/engine/player/bard/choices.go:bd_rousing_targets
1 internal/engine/player/bard/choices.go:bd_victory_extract_stone
1 internal/engine/player/beast_samurai/choices.go:system_discard_cards
1 internal/engine/player/blood_priestess/choices.go:bp_blood_sorrow_target
1 internal/engine/player/butterfly_dancer/choices.go:bt_dance_discard
1 internal/engine/player/butterfly_dancer/choices.go:bt_reverse_branch2_cost
1 internal/engine/player/butterfly_dancer/choices.go:bt_reverse_branch2_pick
1 internal/engine/player/butterfly_dancer/choices.go:bt_reverse_target
2 internal/engine/player/butterfly_dancer/choices.go:bt_wither_confirm
1 internal/engine/player/butterfly_dancer/choices.go:bt_wither_target
1 internal/engine/player/crimson_knight/choices.go:crk_bloody_prayer_target
1 internal/engine/player/crimson_sword_spirit/choices.go:css_blood_rose_gain_heal_target
1 internal/engine/player/elementalist/choices.go:elementalist_bonus_card
1 internal/engine/player/elementalist/choices.go:elementalist_freeze_heal_target
1 internal/engine/player/holy_bow/choices.go:hb_auto_fill_gain
1 internal/engine/player/holy_bow/choices.go:hb_meteor_bullet_target
2 internal/engine/player/magic_bow/choices.go:mb_charge_place_cards
2 internal/engine/player/magic_bow/choices.go:mb_demon_eye_charge_card
1 internal/engine/player/magic_bow/choices.go:mb_demon_eye_target
1 internal/engine/player/magic_bow/choices.go:mb_magic_pierce_hit_charge
1 internal/engine/player/magic_bow/choices.go:mb_multi_shot_target
1 internal/engine/player/magic_bow/choices.go:mb_thunder_scatter_extra
1 internal/engine/player/magic_bow/choices.go:mb_thunder_scatter_target
1 internal/engine/player/magic_lancer/choices.go:ml_fullness_discard_step
1 internal/engine/player/moon_goddess/choices.go:mg_medusa_magic_discard
1 internal/engine/player/moon_goddess/choices.go:mg_moon_cycle_heal_target
1 internal/engine/player/moon_goddess/choices.go:mg_pale_moon_discard
1 internal/engine/player/moon_goddess/choices.go:mg_pale_moon_target
1 internal/engine/player/moon_goddess/choices.go:mg_pale_moon_x
1 internal/engine/player/onmyoji/choices.go:onmyoji_binding_card
1 internal/engine/player/onmyoji/choices.go:onmyoji_binding_counter_target
1 internal/engine/player/onmyoji/choices.go:onmyoji_life_barrier_release_combo
1 internal/engine/player/onmyoji/choices.go:onmyoji_life_barrier_release_target
1 internal/engine/player/onmyoji/choices.go:onmyoji_life_barrier_support_target
1 internal/engine/player/onmyoji/choices.go:onmyoji_yinyang_card
1 internal/engine/player/onmyoji/choices.go:onmyoji_yinyang_counter_target
1 internal/engine/player/priest/choices.go:priest_divine_domain_damage_target
1 internal/engine/player/priest/choices.go:priest_divine_domain_heal_target
1 internal/engine/player/spirit_caster/choices.go:sc_hundred_night_exclude_pick
1 internal/engine/player/spirit_caster/choices.go:sc_hundred_night_fire_reveal
2 internal/engine/player/spirit_caster/choices.go:sc_hundred_night_target
1 internal/engine/player/spirit_caster/choices.go:sc_incant_card
2 internal/engine/player/spirit_caster/choices.go:sc_spiritual_collapse_confirm
1 internal/engine/player/sword_emperor/choices.go:se_sword_rain_discard
EOF
)"

current_choice_type_writes="$(
  rg 'ctxData\["choice_type"\]\s*=\s*"[^"]+"' internal/engine/player \
    --glob '!shared_helpers.go' \
    -n 2>/dev/null |
    awk -F: '{
      file=$1
      sub(/^.*ctxData\["choice_type"\][[:space:]]*=[[:space:]]*"/, "", $0)
      sub(/".*$/, "", $0)
      print file ":" $0
    }' |
    sort |
    uniq -c |
    sed 's/^ *//'
)"

if [[ "$current_choice_type_writes" != "$allowed_choice_type_writes" ]]; then
  echo "FAIL: unexpected naked ctxData[\"choice_type\"] writes in player skills" >&2
  echo "Allowed/current diff:" >&2
  diff -u <(printf '%s\n' "$allowed_choice_type_writes") <(printf '%s\n' "$current_choice_type_writes") >&2 || true
  fail=1
fi

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

echo OK
