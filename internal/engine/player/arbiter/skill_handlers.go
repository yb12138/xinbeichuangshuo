// gameflow: 仲裁者技能处理器。

package arbiter

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// --- helper functions ---

// --- Arbiter handlers ---

type ArbiterLawHandler struct{ engineplayer.BaseHandler }

type ArbiterJudgmentTideHandler struct{ engineplayer.BaseHandler }

type ArbiterRitualHandler struct{ engineplayer.BaseHandler }

type ArbiterRitualBreakHandler struct{ engineplayer.BaseHandler }

type ArbiterDoomsdayHandler struct{ engineplayer.BaseHandler }

type ArbiterBalanceHandler struct{ engineplayer.BaseHandler }

func (h *ArbiterLawHandler) Execute(ctx *model.Context) error {
	ctx.User.Crystal += 2
	ctx.Game.Log(fmt.Sprintf("%s 的 [仲裁法则] 生效，获得2个蓝水晶", ctx.User.Name))
	return nil
}

func (h *ArbiterJudgmentTideHandler) Execute(ctx *model.Context) error {
	if ctx.EventCtx != nil && ctx.EventCtx.DamageVal != nil && *ctx.EventCtx.DamageVal <= 0 {
		return nil
	}
	v := engineplayer.AddToken(ctx.User, "judgment", 1, 4)
	ctx.Game.Log(fmt.Sprintf("%s 的 [审判浪潮] 触发，审判=%d", ctx.User.Name, v))
	return nil
}

func (h *ArbiterRitualHandler) CanUse(ctx *model.Context) bool {
	return !engineplayer.HasForm(ctx.User, model.FormArbiterJudgment) && ctx.User.Gem > 0
}

func (h *ArbiterRitualHandler) Execute(ctx *model.Context) error {
	if ctx.User.Gem <= 0 {
		return nil
	}
	ctx.User.Gem--
	engineplayer.SetForm(ctx.User, model.FormArbiterJudgment)
	ctx.User.MaxHand = 5
	ctx.Game.Log(fmt.Sprintf("%s 发动 [仲裁仪式]，进入审判形态，手牌上限恒定为5", ctx.User.Name))
	return nil
}

func (h *ArbiterRitualBreakHandler) CanUse(ctx *model.Context) bool {
	return engineplayer.HasForm(ctx.User, model.FormArbiterJudgment)
}

func (h *ArbiterRitualBreakHandler) Execute(ctx *model.Context) error {
	engineplayer.ClearForm(ctx.User, model.FormArbiterJudgment)
	if ctx.User.Character != nil && ctx.User.Character.MaxHand > 0 {
		ctx.User.MaxHand = ctx.User.Character.MaxHand
	} else {
		ctx.User.MaxHand = 6
	}
	ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [仪式中断]，转正脱离审判形态并为阵营+1宝石", ctx.User.Name))
	return nil
}

func (h *ArbiterDoomsdayHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return engineplayer.GetToken(ctx.User, "judgment") > 0
}

func (h *ArbiterDoomsdayHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("末日审判需要目标")
	}
	dmg := engineplayer.GetToken(ctx.User, "judgment")
	engineplayer.SetToken(ctx.User, "judgment", 0)
	if dmg > 0 {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, dmg, model.MagicAttack)
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [末日审判]，对 %s 造成%d点法术伤害", ctx.User.Name, ctx.Target.Name, dmg))
	return nil
}

func (h *ArbiterBalanceHandler) CanUse(ctx *model.Context) bool {
	return engineplayer.CanPayCrystalLike(ctx, 1)
}

func (h *ArbiterBalanceHandler) Execute(ctx *model.Context) error {
	// 资源扣除由 UseSkill 统一处理，这里不重复扣费。
	v := engineplayer.AddToken(ctx.User, "judgment", 1, 4)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "arbiter_balance_mode",
			"user_id":     ctx.User.ID,
			"judgment":    v,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [判决天平]，审判=%d，等待选择分支", ctx.User.Name, v))
	return nil
}
