// gameflow: Prompt 目标列表生成。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

type targetChoicePromptSpec struct {
	message     func(data map[string]interface{}) string
	allowCancel bool
}

// IsKnownTargetPromptChoiceType 表示 buildTargetChoicePrompt 内含该类型的文案与选项模板。
func IsKnownTargetPromptChoiceType(choiceType string) bool {
	_, ok := targetChoicePromptSpecs[choiceType]
	return ok
}

var targetChoicePromptSpecs = map[string]targetChoicePromptSpec{
	"elf_elemental_shot_water_target": {
		message: func(map[string]interface{}) string { return "【水之矢】请选择+1治疗目标：" },
	},
	"elf_elemental_shot_earth_target": {
		message: func(map[string]interface{}) string { return "【地之矢】请选择1点法术伤害目标：" },
	},
	"elf_pet_empower_target": {
		message: func(map[string]interface{}) string { return "【宠物强化】请选择摸1弃1目标：" },
	},
	"elf_ritual_release_target": {
		message: func(map[string]interface{}) string {
			return "【精灵密仪】你已无祝福，转正并请选择1名敌方角色承受2点法术伤害："
		},
	},
	"priest_divine_domain_damage_target": {
		message: func(map[string]interface{}) string {
			return "【神圣领域·分支①】请选择2点法术伤害目标："
		},
	},
	"priest_divine_domain_heal_target": {
		message: func(map[string]interface{}) string {
			return "【神圣领域·分支②】请选择+1治疗的队友："
		},
	},
	"onmyoji_dark_ritual_target": {
		message: func(map[string]interface{}) string { return "【黑暗祭礼】请选择2点法术伤害目标：" },
	},
	"onmyoji_life_barrier_support_target": {
		message: func(map[string]interface{}) string {
			return "【生命结界·分支①】请选择获得+1宝石/+1治疗的队友："
		},
	},
	"onmyoji_life_barrier_release_target": {
		message: func(map[string]interface{}) string {
			return "【生命结界·分支②】请选择弃1张手牌的队友："
		},
	},
	"hom_dual_echo_target": {
		message: func(data map[string]interface{}) string {
			return fmt.Sprintf("【双重回响】请选择额外造成%d点法术伤害的目标：", runtimeutil.ToIntContextValue(data["damage"]))
		},
		allowCancel: true,
	},
	"mb_thunder_scatter_target": {
		message: func(data map[string]interface{}) string {
			return fmt.Sprintf("【雷光散射】请选择额外受到%d点法术伤害的目标：", runtimeutil.ToIntContextValue(data["extra_x"]))
		},
	},
	"mb_multi_shot_target": {
		message: func(map[string]interface{}) string { return "【多重射击】请选择暗系追加攻击目标：" },
	},
	"mb_demon_eye_target": {
		message: func(map[string]interface{}) string { return "【魔眼】请选择弃1张牌的目标角色：" },
	},
	"sc_hundred_night_target": {
		message: func(map[string]interface{}) string { return "【百鬼夜行】请选择1点法术伤害目标：" },
	},
	"bd_descent_target": {
		message: func(map[string]interface{}) string { return "【沉沦协奏曲】请选择1点法术伤害目标：" },
	},
	"bd_dissonance_target": {
		message: func(map[string]interface{}) string { return "【不谐和弦】请选择目标角色：" },
	},
	"bd_hope_place_target": {
		message: func(map[string]interface{}) string {
			return "【希望赋格曲】请选择放置永恒乐章的目标队友："
		},
	},
	"bd_hope_transfer_target": {
		message: func(map[string]interface{}) string {
			return "【希望赋格曲】请选择转移永恒乐章的目标队友："
		},
	},
	"ml_stardust_target": {
		message: func(map[string]interface{}) string { return "【幻影星尘】请选择2点法术伤害目标：" },
	},
	"se_sword_qi_slash_target": {
		message: func(data map[string]interface{}) string {
			return fmt.Sprintf("【剑气斩】请选择承受%d点法术伤害的目标：", runtimeutil.ToIntContextValue(data["x_value"]))
		},
	},
	"fighter_psi_bullet_target": {
		message: func(map[string]interface{}) string { return "【念弹】请选择1名目标对手：" },
	},
	"fighter_hundred_dragon_target": {
		message: func(map[string]interface{}) string {
			return "【百式幻龙拳】请选择本行动阶段锁定的目标角色："
		},
	},
}

func (e *GameEngine) buildTargetChoicePrompt(choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	spec, ok := targetChoicePromptSpecs[choiceType]
	if !ok {
		return nil
	}

	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	options := make([]model.PromptOption, 0, len(targetIDs)+1)
	for _, targetID := range targetIDs {
		if target := e.State.Players[targetID]; target != nil {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: target.Name})
		}
	}
	if spec.allowCancel {
		options = append(options, model.PromptOption{ID: "cancel", Label: "取消"})
	}

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  spec.message(data),
		Options:  options,
		Min:      1,
		Max:      1,
	}
}
