// gameflow: 通灵师 handler。

package spirit_caster

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func spiritCasterPowerCount(user *model.Player) int {
	if user == nil {
		return 0
	}
	count := 0
	for _, fc := range user.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSpiritCasterPower {
			continue
		}
		count++
	}
	return count
}

type SpiritCasterTalismanThunderHandler struct{ engineplayer.BaseHandler }

type SpiritCasterTalismanWindHandler struct{ engineplayer.BaseHandler }

type SpiritCasterIncantationHandler struct{ engineplayer.BaseHandler }

type SpiritCasterHundredNightHandler struct{ engineplayer.BaseHandler }

type SpiritCasterSpiritualCollapseHandler struct{ engineplayer.BaseHandler }

func (h *SpiritCasterTalismanThunderHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil
}

func (h *SpiritCasterTalismanThunderHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("灵符-雷鸣上下文无效")
	}
	user := ctx.User
	targetIDs := resolvedTargetIDsFromContext(ctx)
	if len(targetIDs) != 2 {
		return fmt.Errorf("灵符-雷鸣需要且仅需指定2名角色")
	}
	if len(user.Hand) > 0 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"choice_type": "sc_incant_confirm",
				"user_id":     user.ID,
				"skill_id":    "sc_talisman_thunder",
				"target_ids":  targetIDs,
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [灵符-雷鸣]，等待选择是否念咒", user.Name))
	} else {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"choice_type": "sc_spiritual_collapse_confirm",
				"user_id":     user.ID,
				"mode":        "sc_talisman_thunder",
				"target_ids":  targetIDs,
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [灵符-雷鸣]，无手牌可念咒，直接进入灵力崩解选择", user.Name))
	}
	return nil
}

func (h *SpiritCasterTalismanWindHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil
}

func (h *SpiritCasterTalismanWindHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("灵符-风行上下文无效")
	}
	user := ctx.User
	targetIDs := resolvedTargetIDsFromContext(ctx)
	if len(targetIDs) != 2 {
		return fmt.Errorf("灵符-风行需要且仅需指定2名角色")
	}
	if len(user.Hand) > 0 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"choice_type": "sc_incant_confirm",
				"user_id":     user.ID,
				"skill_id":    "sc_talisman_wind",
				"target_ids":  targetIDs,
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [灵符-风行]，等待选择是否念咒", user.Name))
	} else {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"choice_type": "sc_incant_confirm_no_hand",
				"user_id":     user.ID,
				"skill_id":    "sc_talisman_wind",
				"target_ids":  targetIDs,
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [灵符-风行]，无手牌可念咒，直接进入风行弃牌", user.Name))
	}
	return nil
}

// resolvedTargetIDsFromContext 从 Context 中提取目标 ID 列表。
func resolvedTargetIDsFromContext(ctx *model.Context) []string {
	if ctx == nil {
		return nil
	}
	ids := make([]string, 0, len(ctx.Targets))
	for _, t := range ctx.Targets {
		if t != nil {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 && ctx.Target != nil {
		ids = append(ids, ctx.Target.ID)
	}
	return ids
}

func (h *SpiritCasterIncantationHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SpiritCasterIncantationHandler) Execute(ctx *model.Context) error { return nil }

func (h *SpiritCasterHundredNightHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnHitCheck {
		return false
	}
	if ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.ActionType != string(model.ActionAttack) {
		return false
	}
	if !ctx.EventCtx.AttackInfo.IsHit {
		return false
	}
	// 仅主动攻击命中后可发动。
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return spiritCasterPowerCount(ctx.User) > 0
}

func (h *SpiritCasterHundredNightHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("百鬼夜行上下文无效")
	}
	if spiritCasterPowerCount(ctx.User) <= 0 {
		return fmt.Errorf("妖力不足，无法发动百鬼夜行")
	}
	targetIDs := make([]string, 0)
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		targetIDs = append(targetIDs, p.ID)
	}
	if len(targetIDs) == 0 {
		return fmt.Errorf("无可选目标")
	}

	defaultTargetID := ""
	if ctx.Target != nil {
		defaultTargetID = ctx.Target.ID
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":       "sc_hundred_night_power",
			"user_id":           ctx.User.ID,
			"target_ids":        targetIDs,
			"default_target_id": defaultTargetID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 可发动 [百鬼夜行]，请选择要移除的妖力", ctx.User.Name))
	return nil
}

func (h *SpiritCasterSpiritualCollapseHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SpiritCasterSpiritualCollapseHandler) Execute(ctx *model.Context) error { return nil }
