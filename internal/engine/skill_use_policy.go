// gameflow: 技能流策略钩子（executeSkillFlow 后置等）。

package engine

import (
	"fmt"
	"starcup-engine/internal/model"
	"strings"
)

type skillUsePolicy struct {
	// prepareSkillUse 阶段覆盖默认 CostDiscards，决定后续弃牌交互/校验需要的张数。
	resolveDiscardCount func(player *model.Player, skillDef *model.SkillDefinition) int
	// validateSkillDiscardSelection 阶段的二次校验，用于检查弃牌组合规则（顺序、元素、联动要求等）。
	validateDiscardedCards func(use *skillUseRequest) error
	// resolveSkillTargets 阶段的声明式目标规则（人数、阵营、自身约束、附加谓词）。
	targetRules targetRuleSet
	// consumeSkillInputs 阶段自定义弃牌入弃牌堆行为（例如只丢部分牌，其余交给目标/转其他区域）。
	appendDiscardPile func(e *GameEngine, use *skillUseRequest)
	// executeSkillFlow 阶段的后置钩子：资源与输入已消耗后触发；返回 handled=true 可跳过默认 handler。
	afterConsume func(e *GameEngine, use *skillUseRequest) (bool, error)
	// finishSkillUse 阶段是否跳过 Action 技能的默认收尾（HasActed/LastActionType 自动写入）。
	skipAutoPhaseEnd bool
	// validateSkillDiscardSelection 阶段对专属卡改为“仅检查不自动消耗”，由技能自定义流程手动处理。
	manualExclusiveCard bool
}

// skillUsePolicies 承接 docs/character_skills_config.md 中当前 SkillDefinition 尚不能直接表达的差异规则。
var skillUsePolicies = map[string]skillUsePolicy{
	"angel_blessing": {
		targetRules: targetRuleSet{
			Count:       targetCountRule{Min: 1, Max: 2, Err: "天使祝福只能指定 1 名或 2 名目标"},
			Distinct:    true,
			DistinctErr: "天使祝福指定 2 名目标时不能重复选择同一角色",
		},
	},
	"angel_cleanse": {
		targetRules: targetRuleSet{
			// 允许不选目标：若场上无基础效果，技能直接结算并跳过“移除基础效果”步骤。
			Count: targetCountRule{Min: 0, Max: 1, Err: "风之洁净最多指定1名目标"},
		},
	},
	"elementalist_freeze": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 2, Max: 2, Err: "冰冻需要指定2名目标"},
		},
	},
	"holy_lancer_punishment": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 1, Max: 1, Err: "惩戒需要指定1名其他角色"},
			Slots: []targetSlotRule{
				{Index: 0, Self: targetSelfOther, Err: "惩戒目标必须是其他角色"},
			},
			Checks: []targetCheckRule{
				{Kind: targetCheckTargetMinHeal, Index: 0, Min: 1, Err: "惩戒目标至少需要有1点治疗"},
			},
		},
	},
	"css_blood_rose": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 2, Max: 2, Err: "血染蔷薇需要恰好指定2名目标"},
			Slots: []targetSlotRule{
				{Index: 1, Camp: targetCampAlly, Err: "血染蔷薇的第2个目标必须是我方角色"},
			},
		},
	},
	// 神官-神圣领域：严格弃 2 张牌，再选择分支目标。
	"priest_divine_domain": {
		resolveDiscardCount: func(player *model.Player, skillDef *model.SkillDefinition) int {
			return skillDef.CostDiscards
		},
	},
	// 神官-水之神力：先弃 1 张水系牌，再把另一张手牌交给队友。
	"priest_water_power": {
		resolveDiscardCount: func(player *model.Player, skillDef *model.SkillDefinition) int {
			return skillDef.CostDiscards
		},
		validateDiscardedCards: func(use *skillUseRequest) error {
			if len(use.discardedCards) != 2 {
				return fmt.Errorf("水之神力需要选择1张水系牌并额外选择1张手牌交给队友")
			}
			if use.discardedCards[0].Element != model.ElementWater {
				return fmt.Errorf("水之神力第一张必须弃置水系牌")
			}
			return nil
		},
		appendDiscardPile: func(e *GameEngine, use *skillUseRequest) {
			if len(use.discardedCards) >= 2 {
				// 第二张弃牌由技能效果交给队友，不进入弃牌堆。
				e.State.DiscardPile = append(e.State.DiscardPile, use.discardedCards[0])
				return
			}
			appendDiscardedCardsToPile(e, use)
		},
	},
	// 阴阳师-式神降临：弃置 2 张命格相同的手牌。
	"onmyoji_shikigami_descend": {
		validateDiscardedCards: func(use *skillUseRequest) error {
			if len(use.discardedCards) != 2 {
				return fmt.Errorf("式神降临需要弃置2张手牌")
			}
			f1 := strings.TrimSpace(use.discardedCards[0].Faction)
			f2 := strings.TrimSpace(use.discardedCards[1].Faction)
			if f1 == "" || f2 == "" || f1 != f2 {
				return fmt.Errorf("式神降临需要弃置2张命格相同的手牌")
			}
			return nil
		},
	},
	"onmyoji_life_barrier": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 1, Err: "生命结界需要且仅能指定1名其他队友"},
			Slots: []targetSlotRule{
				{Index: 0, Camp: targetCampAlly, Self: targetSelfOther, Err: "生命结界目标必须是其他队友"},
			},
		},
	},
	"bw_blazing_codex": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 1, Max: 1, Err: "苍炎法典需要且仅能指定1名其他角色"},
			Slots: []targetSlotRule{
				{Index: 0, Self: targetSelfOther, Err: "苍炎法典不能以自己为目标"},
			},
		},
	},
	"bw_heavenfire_cleave": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 1, Max: 1, Err: "天火断空需要且仅能指定1名其他角色"},
			Slots: []targetSlotRule{
				{Index: 0, Self: targetSelfOther, Err: "天火断空不能以自己为目标"},
			},
		},
	},
	"mb_thunder_scatter": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 1, Err: "雷光散射至多指定1名敌方角色作为额外目标"},
			Slots: []targetSlotRule{
				{Index: 0, Camp: targetCampEnemy, Err: "雷光散射的额外目标必须是敌方角色"},
			},
		},
	},
	"mb_demon_eye": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 1, Err: "魔眼需要且仅能指定1名其他角色"},
			Slots: []targetSlotRule{
				{Index: 0, Self: targetSelfOther, Err: "魔眼不能以自己为目标"},
			},
		},
	},
	"ml_phantom_stardust": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 1, Err: "幻影星尘需要且仅能指定1名敌方角色"},
			Slots: []targetSlotRule{
				{Index: 0, Camp: targetCampEnemy, Err: "幻影星尘目标必须是敌方角色"},
			},
		},
	},
	"ml_fullness": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 1, Err: "充盈至多指定1名其他队友"},
			Slots: []targetSlotRule{
				{Index: 0, Camp: targetCampAlly, Self: targetSelfOther, Err: "充盈的可选目标必须是其他队友"},
			},
		},
	},
	"sage_arcane_codex": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 1, Err: "魔道法典需要且仅能指定1名其他角色"},
			Slots: []targetSlotRule{
				{Index: 0, Self: targetSelfOther, Err: "魔道法典不能以自己为目标"},
			},
		},
	},
	"sage_holy_codex": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 6},
		},
	},
	"bd_dissonance_chord": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 1},
		},
	},
	"bd_hope_fugue": {
		manualExclusiveCard: true,
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 1, Err: "希望赋格曲至多指定1名其他队友"},
			Slots: []targetSlotRule{
				{Index: 0, Camp: targetCampAlly, Self: targetSelfOther, Err: "希望赋格曲的目标必须是其他队友"},
			},
		},
	},
	"ss_soul_link": {
		manualExclusiveCard: true,
	},
	// 封印师-封印破碎：目标面前必须存在基础效果。
	"seal_break": {
		targetRules: targetRuleSet{
			Count: targetCountRule{Min: 0, Max: 1},
			Checks: []targetCheckRule{
				{Kind: targetCheckAnyBasicFieldWhenNone, Err: "场上没有可收回的基础效果"},
				{Kind: targetCheckHasBasicFieldOnTarget, Index: 0, Err: "%s 面前没有可收回的基础效果", WithTargetName: true},
			},
		},
	},
	// 灵符师灵符技能：技能效果通过延迟 followup 串行结算，绕过默认 handler。
	"sc_talisman_thunder": {
		afterConsume: func(e *GameEngine, use *skillUseRequest) (bool, error) {
			if err := e.beginSpiritCasterTalisman(use.player, use.skillID, use.resolvedTargetIDs(), use.discardedCards); err != nil {
				return false, err
			}
			return true, nil
		},
	},
	"sc_talisman_wind": {
		afterConsume: func(e *GameEngine, use *skillUseRequest) (bool, error) {
			if err := e.beginSpiritCasterTalisman(use.player, use.skillID, use.resolvedTargetIDs(), use.discardedCards); err != nil {
				return false, err
			}
			return true, nil
		},
	},
	// 冒险家-欺诈：技能本身会继续驱动后续选择，不在此处自动结束法术行动。
	"adventurer_fraud": {
		skipAutoPhaseEnd: true,
	},
	// 圣女-圣疗：治疗分配与额外行动类型由后续中断流完成。
	"saint_heal": {
		skipAutoPhaseEnd: true,
	},
	// 瘟疫术士-死亡之触：多段选择与延迟伤害结束后，按正常 ActionEnd 再触发阶段结束效果。
	"plague_death_touch": {
		skipAutoPhaseEnd: true,
	},
}

func resolveSkillUsePolicy(skillID string) skillUsePolicy {
	if policy, ok := skillUsePolicies[skillID]; ok {
		return policy
	}
	return skillUsePolicy{}
}

func appendDiscardedCardsToPile(e *GameEngine, use *skillUseRequest) {
	e.State.DiscardPile = append(e.State.DiscardPile, use.discardedCards...)
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
