// gameflow: 技能流策略钩子（executeSkillFlow 后置等）。

package engine

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// 类型别名，保持 engine 包内兼容性。
type SkillPolicy = types.SkillPolicy
type PolicyContext = types.PolicyContext
type PolicyHost = types.PolicyHost
type BeginSkillFollowupReq = types.BeginSkillFollowupReq

// skillUsePolicies 承接 docs/character_skills_config.md 中当前 SkillDefinition 尚不能直接表达的差异规则。
// 所有角色策略已迁移到 player/<role>/module.go:
// - 神官（priest_divine_domain, priest_water_power）
// - 魔弓（mb_thunder_scatter, mb_demon_eye）
// - 魔枪（ml_phantom_stardust, ml_fullness）
// - 贤者（sage_arcane_codex, sage_holy_codex）
// - 吟游诗人（bd_dissonance_chord, bd_hope_fugue）
// - 灵魂术士（ss_soul_link）
// - 鬼术师（onmyoji_shikigami_descend, onmyoji_life_barrier）
// - 苍炎魔女（bw_blazing_codex, bw_heavenfire_cleave）
// - 血色剑灵（css_blood_rose）
// - 瘟疫术士（plague_death_touch）

var skillUsePolicies = map[string]SkillPolicy{}

func init() {
	mountPlayerSkillPolicySpecs(skillUsePolicies)
}

func resolveSkillUsePolicy(skillID string) SkillPolicy {
	if policy, ok := skillUsePolicies[skillID]; ok {
		return policy
	}
	return SkillPolicy{}
}

func hasBasicFieldEffect(player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect {
			continue
		}
		if model.IsBasicEffect(string(fc.Effect)) {
			return true
		}
	}
	return false
}
