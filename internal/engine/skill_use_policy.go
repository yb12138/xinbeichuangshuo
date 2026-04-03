package engine

import (
	"fmt"
	"starcup-engine/internal/model"
	"strings"
)

type skillUsePolicy struct {
	resolveDiscardCount    func(player *model.Player, skillDef *model.SkillDefinition) int
	validateDiscardedCards func(use *skillUseRequest) error
	validateTargets        func(use *skillUseRequest) error
	appendDiscardPile      func(e *GameEngine, use *skillUseRequest)
	afterConsume           func(e *GameEngine, use *skillUseRequest) (bool, error)
	skipAutoPhaseEnd       bool
	allowZeroTargets       bool
	manualExclusiveCard    bool
}

// skillUsePolicies 承接 docs/character_skills_config.md 中当前 SkillDefinition 尚不能直接表达的差异规则。
var skillUsePolicies = map[string]skillUsePolicy{
	"angel_blessing": {
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) < 1 || len(use.actualTargets) > 2 {
				return fmt.Errorf("天使祝福只能指定 1 名或 2 名目标")
			}
			if len(use.actualTargets) == 2 && use.actualTargets[0] != nil && use.actualTargets[1] != nil &&
				use.actualTargets[0].ID == use.actualTargets[1].ID {
				return fmt.Errorf("天使祝福指定 2 名目标时不能重复选择同一角色")
			}
			return nil
		},
	},
	"angel_cleanse": {
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) == 0 || use.actualTargets[0] == nil {
				return fmt.Errorf("风之洁净需要指定目标")
			}
			if !hasBasicFieldEffect(use.actualTargets[0]) {
				return fmt.Errorf("%s 面前没有可移除的基础效果", use.actualTargets[0].Name)
			}
			return nil
		},
	},
	"elementalist_freeze": {
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) != 2 {
				return fmt.Errorf("冰冻需要指定2名目标")
			}
			if use.actualTargets[0] == nil {
				return fmt.Errorf("冰冻缺少法术伤害目标")
			}
			if use.actualTargets[0].Camp == use.player.Camp {
				return fmt.Errorf("冰冻的第1个目标必须是敌方角色")
			}
			return nil
		},
	},
	"holy_lancer_punishment": {
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("惩戒需要指定1名其他角色")
			}
			target := use.actualTargets[0]
			if use.player != nil && target.ID == use.player.ID {
				return fmt.Errorf("惩戒目标必须是其他角色")
			}
			if target.Heal <= 0 {
				return fmt.Errorf("惩戒目标至少需要有1点治疗")
			}
			return nil
		},
	},
	"css_blood_rose": {
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) != 2 {
				return fmt.Errorf("血染蔷薇需要恰好指定2名目标")
			}
			if use.actualTargets[0] == nil || use.actualTargets[1] == nil {
				return fmt.Errorf("血染蔷薇目标无效")
			}
			if use.actualTargets[0].ID == use.actualTargets[1].ID {
				return fmt.Errorf("血染蔷薇不能重复选择同一角色")
			}
			allyCount := 0
			enemyCount := 0
			for _, target := range use.actualTargets {
				if target.Camp == use.player.Camp {
					allyCount++
				} else {
					enemyCount++
				}
			}
			if allyCount != 1 || enemyCount != 1 {
				return fmt.Errorf("血染蔷薇需要恰好指定1名敌方和1名我方角色")
			}
			return nil
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
		allowZeroTargets: true,
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) == 0 {
				return nil
			}
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("生命结界需要且仅能指定1名其他队友")
			}
			target := use.actualTargets[0]
			if target.Camp != use.player.Camp || target.ID == use.player.ID {
				return fmt.Errorf("生命结界目标必须是其他队友")
			}
			return nil
		},
	},
	"bw_blazing_codex": {
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("苍炎法典需要且仅能指定1名其他角色")
			}
			if use.actualTargets[0].ID == use.player.ID {
				return fmt.Errorf("苍炎法典不能以自己为目标")
			}
			return nil
		},
	},
	"bw_heavenfire_cleave": {
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("天火断空需要且仅能指定1名其他角色")
			}
			if use.actualTargets[0].ID == use.player.ID {
				return fmt.Errorf("天火断空不能以自己为目标")
			}
			return nil
		},
	},
	"mb_thunder_scatter": {
		allowZeroTargets: true,
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) == 0 {
				return nil
			}
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("雷光散射至多指定1名敌方角色作为额外目标")
			}
			if use.actualTargets[0].Camp == use.player.Camp {
				return fmt.Errorf("雷光散射的额外目标必须是敌方角色")
			}
			return nil
		},
	},
	"mb_demon_eye": {
		allowZeroTargets: true,
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) == 0 {
				return nil
			}
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("魔眼需要且仅能指定1名其他角色")
			}
			if use.actualTargets[0].ID == use.player.ID {
				return fmt.Errorf("魔眼不能以自己为目标")
			}
			return nil
		},
	},
	"ml_phantom_stardust": {
		allowZeroTargets: true,
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) == 0 {
				return nil
			}
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("幻影星尘需要且仅能指定1名敌方角色")
			}
			if use.actualTargets[0].Camp == use.player.Camp {
				return fmt.Errorf("幻影星尘目标必须是敌方角色")
			}
			return nil
		},
	},
	"ml_fullness": {
		allowZeroTargets: true,
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) == 0 {
				return nil
			}
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("充盈至多指定1名其他队友")
			}
			target := use.actualTargets[0]
			if target.Camp != use.player.Camp || target.ID == use.player.ID {
				return fmt.Errorf("充盈的可选目标必须是其他队友")
			}
			return nil
		},
	},
	"sage_arcane_codex": {
		allowZeroTargets: true,
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) == 0 {
				return nil
			}
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("魔道法典需要且仅能指定1名其他角色")
			}
			if use.actualTargets[0].ID == use.player.ID {
				return fmt.Errorf("魔道法典不能以自己为目标")
			}
			return nil
		},
	},
	"sage_holy_codex": {
		allowZeroTargets: true,
	},
	"bd_dissonance_chord": {
		allowZeroTargets: true,
	},
	"bd_hope_fugue": {
		allowZeroTargets:    true,
		manualExclusiveCard: true,
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) == 0 {
				return nil
			}
			if len(use.actualTargets) != 1 || use.actualTargets[0] == nil {
				return fmt.Errorf("希望赋格曲至多指定1名其他队友")
			}
			target := use.actualTargets[0]
			if target.Camp != use.player.Camp || target.ID == use.player.ID {
				return fmt.Errorf("希望赋格曲的目标必须是其他队友")
			}
			return nil
		},
	},
	"ss_soul_link": {
		manualExclusiveCard: true,
	},
	// 封印师-封印破碎：目标面前必须存在基础效果。
	"seal_break": {
		allowZeroTargets: true,
		validateTargets: func(use *skillUseRequest) error {
			if len(use.actualTargets) == 0 {
				if use == nil || use.engine == nil {
					return fmt.Errorf("封印破碎缺少引擎上下文")
				}
				for _, player := range use.engine.GetAllPlayers() {
					if hasBasicFieldEffect(player) {
						return nil
					}
				}
				return fmt.Errorf("场上没有可收回的基础效果")
			}
			if use.actualTargets[0] == nil {
				return fmt.Errorf("封印破碎需要指定有效目标")
			}
			if !hasBasicFieldEffect(use.actualTargets[0]) {
				return fmt.Errorf("%s 面前没有可收回的基础效果", use.actualTargets[0].Name)
			}
			return nil
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

func clampSkillDiscardCount(required, handSize int) int {
	if required > handSize {
		return handSize
	}
	if required < 0 {
		return 0
	}
	return required
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
