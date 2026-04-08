package engine

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed config/timing_op_bindings.json
var timingOpBindingsJSON []byte

type timingOpPresenceRule struct {
	requireAnyCharacters []string
}

func (r timingOpPresenceRule) enabled(present map[string]bool) bool {
	if len(r.requireAnyCharacters) == 0 {
		return true
	}
	for _, roleID := range r.requireAnyCharacters {
		if present[roleID] {
			return true
		}
	}
	return false
}

type attackDeclaredOpBinding struct {
	op       timingOnAttackDeclaredOp
	priority int
	presence timingOpPresenceRule
	handler  timingOnAttackDeclaredHandler
}

type hitCheckOpBinding struct {
	op       timingOnHitCheckOp
	priority int
	presence timingOpPresenceRule
	handler  timingOnHitCheckHandler
}

type damageCalculatedOpBinding struct {
	op       timingOnDamageCalculatedOp
	priority int
	presence timingOpPresenceRule
	handler  timingOnDamageCalculatedHandler
}

type hitCheckSkillOpBinding struct {
	op       timingOnHitCheckSkillOp
	priority int
	presence timingOpPresenceRule
	handler  timingOnHitCheckSkillHandler
}

type timingOpBindingRegistry struct {
	attackDeclared []attackDeclaredOpBinding
	hitCheck       []hitCheckOpBinding
	damageCalc     []damageCalculatedOpBinding
	hitCheckSkill  []hitCheckSkillOpBinding
}

type timingOpBindingFile struct {
	Bindings []timingOpBindingRow `json:"bindings"`
}

type timingOpBindingRow struct {
	Stage                string   `json:"stage"`
	Op                   string   `json:"op"`
	Handler              string   `json:"handler"`
	Priority             int      `json:"priority"`
	RequireAnyCharacters []string `json:"require_any_characters"`
}

var timingOpBindings = mustLoadTimingOpBindingRegistry()

func mustLoadTimingOpBindingRegistry() timingOpBindingRegistry {
	// 配置加载只发生一次：把外部 JSON 解析成“阶段-op-handler”注册源，
	// 后续每次角色变化只做 presence 过滤与优先级选择，不重复解析文件。
	var file timingOpBindingFile
	if err := json.Unmarshal(timingOpBindingsJSON, &file); err != nil {
		panic(fmt.Sprintf("failed to parse timing op binding config: %v", err))
	}

	registry := timingOpBindingRegistry{}
	for idx, row := range file.Bindings {
		presence := timingOpPresenceRule{requireAnyCharacters: row.RequireAnyCharacters}
		switch row.Stage {
		case "attack_declared":
			op, ok := parseAttackDeclaredOp(row.Op)
			if !ok {
				panic(fmt.Sprintf("invalid attack_declared op at row %d: %s", idx, row.Op))
			}
			handler, ok := attackDeclaredHandlerResolver[row.Handler]
			if !ok {
				panic(fmt.Sprintf("invalid attack_declared handler at row %d: %s", idx, row.Handler))
			}
			registry.attackDeclared = append(registry.attackDeclared, attackDeclaredOpBinding{
				op:       op,
				priority: row.Priority,
				presence: presence,
				handler:  handler,
			})
		case "hit_check":
			op, ok := parseHitCheckOp(row.Op)
			if !ok {
				panic(fmt.Sprintf("invalid hit_check op at row %d: %s", idx, row.Op))
			}
			handler, ok := hitCheckHandlerResolver[row.Handler]
			if !ok {
				panic(fmt.Sprintf("invalid hit_check handler at row %d: %s", idx, row.Handler))
			}
			registry.hitCheck = append(registry.hitCheck, hitCheckOpBinding{
				op:       op,
				priority: row.Priority,
				presence: presence,
				handler:  handler,
			})
		case "damage_calculated":
			op, ok := parseDamageCalculatedOp(row.Op)
			if !ok {
				panic(fmt.Sprintf("invalid damage_calculated op at row %d: %s", idx, row.Op))
			}
			handler, ok := damageCalculatedHandlerResolver[row.Handler]
			if !ok {
				panic(fmt.Sprintf("invalid damage_calculated handler at row %d: %s", idx, row.Handler))
			}
			registry.damageCalc = append(registry.damageCalc, damageCalculatedOpBinding{
				op:       op,
				priority: row.Priority,
				presence: presence,
				handler:  handler,
			})
		case "hit_check_skill":
			op, ok := parseHitCheckSkillOp(row.Op)
			if !ok {
				panic(fmt.Sprintf("invalid hit_check_skill op at row %d: %s", idx, row.Op))
			}
			handler, ok := hitCheckSkillHandlerResolver[row.Handler]
			if !ok {
				panic(fmt.Sprintf("invalid hit_check_skill handler at row %d: %s", idx, row.Handler))
			}
			registry.hitCheckSkill = append(registry.hitCheckSkill, hitCheckSkillOpBinding{
				op:       op,
				priority: row.Priority,
				presence: presence,
				handler:  handler,
			})
		default:
			panic(fmt.Sprintf("invalid stage at row %d: %s", idx, row.Stage))
		}
	}
	return registry
}

func parseAttackDeclaredOp(raw string) (timingOnAttackDeclaredOp, bool) {
	op := timingOnAttackDeclaredOp(raw)
	switch op {
	case timingOnAttackDeclaredCardTransform,
		timingOnAttackDeclaredTargetContext,
		timingOnAttackDeclaredStateReset,
		timingOnAttackDeclaredPreCombat,
		timingOnAttackDeclaredPendingDamageInit,
		timingOnAttackDeclaredInterrupt:
		return op, true
	default:
		return "", false
	}
}

func parseHitCheckOp(raw string) (timingOnHitCheckOp, bool) {
	op := timingOnHitCheckOp(raw)
	switch op {
	case timingOnHitCheckPendingDamageAttackHit,
		timingOnHitCheckCombatInteraction,
		timingOnHitCheckCombatDefendValidation,
		timingOnHitCheckCombatCounterCard,
		timingOnHitCheckCombatCounterElement,
		timingOnHitCheckCombatCounterResolve,
		timingOnHitCheckMagicMissileDefend,
		timingOnHitCheckMagicMissileCounter,
		timingOnHitCheckResponseSkip:
		return op, true
	default:
		return "", false
	}
}

func parseDamageCalculatedOp(raw string) (timingOnDamageCalculatedOp, bool) {
	op := timingOnDamageCalculatedOp(raw)
	switch op {
	case timingOnDamageCalculatedAttackPassive,
		timingOnDamageCalculatedBeforeTaken,
		timingOnDamageCalculatedHealCap,
		timingOnDamageCalculatedHealResist:
		return op, true
	default:
		return "", false
	}
}

func parseHitCheckSkillOp(raw string) (timingOnHitCheckSkillOp, bool) {
	op := timingOnHitCheckSkillOp(raw)
	switch op {
	case timingOnHitCheckSkillAugment, timingOnHitCheckSkillNormalize:
		return op, true
	default:
		return "", false
	}
}

var attackDeclaredHandlerResolver = map[string]timingOnAttackDeclaredHandler{
	"default.attack_declared.card_transform":      timingOpAttackDeclaredCardTransform,
	"default.attack_declared.target_context":      timingOpAttackDeclaredTargetContext,
	"default.attack_declared.state_reset":         timingOpAttackDeclaredStateReset,
	"default.attack_declared.pre_combat":          timingOpAttackDeclaredPreCombat,
	"default.attack_declared.pending_damage_init": timingOpAttackDeclaredPendingDamageInit,
	"default.attack_declared.interrupt":           timingOpAttackDeclaredInterrupt,
	"overlay.moon.pre_combat":                     timingOpAttackDeclaredPreCombatMoonOverlay,
}

var hitCheckHandlerResolver = map[string]timingOnHitCheckHandler{
	"default.hit_check.pending_damage_attack_hit": timingOpHitCheckPendingDamageAttackHit,
	"default.hit_check.combat_interaction":        timingOpHitCheckCombatInteraction,
	"default.hit_check.combat_defend_validation":  timingOpHitCheckCombatDefendValidation,
	"default.hit_check.combat_counter_card":       timingOpHitCheckCombatCounterCard,
	"default.hit_check.combat_counter_element":    timingOpHitCheckCombatCounterElement,
	"default.hit_check.combat_counter_resolve":    timingOpHitCheckCombatCounterResolve,
	"default.hit_check.magic_missile_defend":      timingOpHitCheckMagicMissileDefend,
	"default.hit_check.magic_missile_counter":     timingOpHitCheckMagicMissileCounter,
	"default.hit_check.response_skip":             timingOpHitCheckResponseSkip,
	"overlay.onmyoji.combat_counter_element":      timingOpHitCheckCombatCounterElementOnmyojiOverlay,
}

var damageCalculatedHandlerResolver = map[string]timingOnDamageCalculatedHandler{
	"default.damage_calculated.attack_passive": timingOpDamageCalculatedAttackPassive,
	"default.damage_calculated.before_taken":   timingOpDamageCalculatedBeforeTaken,
	"default.damage_calculated.heal_cap":       timingOpDamageCalculatedHealCap,
	"default.damage_calculated.heal_resist":    timingOpDamageCalculatedHealResist,
	"overlay.crimson_or_plague.heal_resist":    timingOpDamageCalculatedHealResistOverlay,
}

var hitCheckSkillHandlerResolver = map[string]timingOnHitCheckSkillHandler{
	"default.hit_check_skill.augment":   timingOpHitCheckSkillAugment,
	"default.hit_check_skill.normalize": timingOpHitCheckSkillNormalize,
}
