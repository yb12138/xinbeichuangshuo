package skills

import (
	"fmt"
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

type SpiritCasterTalismanThunderHandler struct{ BaseHandler }

type SpiritCasterTalismanWindHandler struct{ BaseHandler }

type SpiritCasterIncantationHandler struct{ BaseHandler }

type SpiritCasterHundredNightHandler struct{ BaseHandler }

type SpiritCasterSpiritualCollapseHandler struct{ BaseHandler }

func (h *SpiritCasterTalismanThunderHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil
}

func (h *SpiritCasterTalismanThunderHandler) Execute(ctx *model.Context) error {
	// 灵符技能主流程在 engine.UseSkill 中处理（需要先处理封印/念咒/后续串行）。
	return nil
}

func (h *SpiritCasterTalismanWindHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil
}

func (h *SpiritCasterTalismanWindHandler) Execute(ctx *model.Context) error {
	// 灵符技能主流程在 engine.UseSkill 中处理（需要先处理封印/念咒/后续串行）。
	return nil
}

func (h *SpiritCasterIncantationHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SpiritCasterIncantationHandler) Execute(ctx *model.Context) error { return nil }

func (h *SpiritCasterHundredNightHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackHit {
		return false
	}
	// 仅主动攻击命中后可发动。
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
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
