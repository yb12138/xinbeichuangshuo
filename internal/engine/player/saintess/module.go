// gameflow: 圣女模块入口声明。

package saintess

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

type baseHandler struct{}

func (baseHandler) CanUse(_ *model.Context) bool { return true }

type FrostPrayerHandler struct{ baseHandler }

func (h *FrostPrayerHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if ctx.Timing != model.TimingOnCardPlayedOrRevealed {
		return false
	}
	if ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	card := ctx.EventCtx.Card
	return card.Element == model.ElementWater || card.Name == "圣光"
}

func (h *FrostPrayerHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.Game == nil || ctx.User == nil {
		return fmt.Errorf("冰霜祷言上下文无效")
	}
	options := make([]model.PromptOption, 0, len(ctx.Game.GetAllPlayers()))
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    p.ID,
			Label: p.Name,
		})
	}
	if len(options) == 0 {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "frost_prayer_target",
			"user_id":     ctx.User.ID,
			"target_ids": func() []string {
				ids := make([]string, 0, len(options))
				for _, opt := range options {
					ids = append(ids, opt.ID)
				}
				return ids
			}(),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 的 [冰霜祷言] 触发，等待选择治疗目标", ctx.User.Name))
	return nil
}

type HealingLightHandler struct{ baseHandler }

func (h *HealingLightHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.Game == nil || ctx.User == nil {
		return fmt.Errorf("治愈之光上下文无效")
	}
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) == 0 {
		return fmt.Errorf("需要指定目标")
	}
	for _, t := range targets {
		if t == nil {
			continue
		}
		ctx.Game.Heal(t.ID, 1)
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [治愈之光]，%d 名角色各 +1 治疗", ctx.User.Name, len(targets)))
	return nil
}

type HealHandler struct{ baseHandler }

func (h *HealHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.Game == nil || ctx.User == nil {
		return fmt.Errorf("治疗术上下文无效")
	}
	if ctx.Target == nil {
		return fmt.Errorf("需要指定目标")
	}
	ctx.Game.Heal(ctx.Target.ID, 2)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [治疗术]，%s 获得 +2 治疗", ctx.User.Name, ctx.Target.Name))
	return nil
}

type SaintHealHandler struct{ baseHandler }

func (h *SaintHealHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.Game == nil || ctx.User == nil {
		return fmt.Errorf("圣疗上下文无效")
	}
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) == 0 || len(targets) > 3 {
		return fmt.Errorf("圣疗需要指定1-3名目标")
	}
	targetIDs := make([]string, 0, len(targets))
	for _, target := range targets {
		if target == nil {
			continue
		}
		targetIDs = append(targetIDs, target.ID)
	}
	if len(targetIDs) == 0 {
		return fmt.Errorf("圣疗缺少有效目标")
	}
	data := map[string]interface{}{
		"targets": targetIDs,
	}
	if len(targetIDs) == 2 {
		data["stage"] = "allocate_heal"
	} else {
		data["stage"] = "choose_extra_action"
		allocations := map[string]int{}
		if len(targetIDs) == 1 {
			allocations[targetIDs[0]] = 3
		} else {
			for _, targetID := range targetIDs {
				allocations[targetID] = 1
			}
		}
		data["allocations"] = allocations
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptSaintHeal,
		PlayerID: ctx.User.ID,
		Context:  data,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣疗]，等待分配治疗并选择额外行动类型", ctx.User.Name))
	return nil
}

type MercyHandler struct{ baseHandler }

func (h *MercyHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if ctx.Timing != model.TimingOnTurnStart && ctx.Timing != model.TimingStartup {
		return false
	}
	if ctx.User.Gem < 1 {
		return false
	}
	return !player.HasForm(ctx.User, model.FormSaintessMercy)
}

func (h *MercyHandler) Execute(ctx *model.Context) error {
	user := ctx.User
	game := ctx.Game
	if user == nil || game == nil {
		return fmt.Errorf("怜悯上下文无效")
	}
	if user.Gem < 1 {
		return fmt.Errorf("宝石不足，无法发动怜悯")
	}
	if player.HasForm(user, model.FormSaintessMercy) {
		return fmt.Errorf("已处于怜悯状态")
	}
	user.Gem -= 1
	user.Crystal += 1
	player.SetForm(user, model.FormSaintessMercy)
	game.Log(fmt.Sprintf("%s 发动 [怜悯]：横置并获得1水晶，手牌上限恒定为7", user.Name))
	return nil
}

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID: "saintess",
		HandLimit: player.HandLimitRuleFuncs{
			Hard: func(p *model.Player) (int, bool) {
				if player.HasForm(p, model.FormSaintessMercy) {
					return 7, true
				}
				return 0, false
			},
		},
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		InterruptSpecs: []player.InterruptSpec{
			{
				Type:                 model.InterruptSaintHeal,
				PhaseSync:            player.InterruptPhaseSyncCombatHeal,
				BuildPrompt:          buildSaintHealPrompt,
				HandleActionResult:   handleSaintHealAction,
				AllowedActionTypes:   []model.PlayerActionType{model.CmdSelect},
				InvalidActionMessage: "当前为【圣疗】选择阶段，请提交选择",
			},
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "frost_prayer", Handler: &FrostPrayerHandler{}},
		{ID: "healing_light", Handler: &HealingLightHandler{}},
		{ID: "heal", Handler: &HealHandler{}},
		{
			ID:      "saint_heal",
			Handler: &SaintHealHandler{},
			Policy: types.SkillPolicy{
				SkipAutoPhaseEnd: true,
			},
		},
		{ID: "mercy", Handler: &MercyHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"frost_prayer_target": types.ChoiceRouteRole("saintess"),
	}
}
