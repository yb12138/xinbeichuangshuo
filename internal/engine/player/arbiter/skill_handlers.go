// gameflow: 仲裁者技能处理器。

package arbiter

import (
	"fmt"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// --- helper functions ---

func getToken(p *model.Player, key string) int {
	if p == nil {
		return 0
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	return p.Tokens[key]
}

func setToken(p *model.Player, key string, v int) {
	if p == nil {
		return
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens[key] = v
}

func addToken(p *model.Player, key string, delta int, minV int, maxV int) int {
	cur := getToken(p, key)
	cur += delta
	if cur < minV {
		cur = minV
	}
	if maxV >= minV && cur > maxV {
		cur = maxV
	}
	setToken(p, key, cur)
	return cur
}

func hasForm(p *model.Player, form string) bool {
	return p != nil && p.Form == form
}

func enterForm(p *model.Player, form string) {
	if p == nil {
		return
	}
	p.Orientation = model.OrientationTapped
	p.Form = form
}

func leaveForm(p *model.Player, form string) {
	if p == nil {
		return
	}
	if form != "" && p.Form != form {
		return
	}
	p.Orientation = model.OrientationNormal
	p.Form = ""
}

// --- Arbiter handlers ---

type ArbiterLawHandler struct{ skills.BaseHandler }

type ArbiterJudgmentTideHandler struct{ skills.BaseHandler }

type ArbiterRitualHandler struct{ skills.BaseHandler }

type ArbiterRitualBreakHandler struct{ skills.BaseHandler }

type ArbiterDoomsdayHandler struct{ skills.BaseHandler }

type ArbiterBalanceHandler struct{ skills.BaseHandler }

func (h *ArbiterLawHandler) Execute(ctx *model.Context) error {
	ctx.User.Crystal += 2
	ctx.Game.Log(fmt.Sprintf("%s 的 [仲裁法则] 生效，获得2个蓝水晶", ctx.User.Name))
	return nil
}

func (h *ArbiterJudgmentTideHandler) Execute(ctx *model.Context) error {
	if ctx.EventCtx != nil && ctx.EventCtx.DamageVal != nil && *ctx.EventCtx.DamageVal <= 0 {
		return nil
	}
	v := addToken(ctx.User, "judgment", 1, 0, 4)
	ctx.Game.Log(fmt.Sprintf("%s 的 [审判浪潮] 触发，审判=%d", ctx.User.Name, v))
	return nil
}

func (h *ArbiterRitualHandler) CanUse(ctx *model.Context) bool {
	return !hasForm(ctx.User, model.FormArbiterJudgment) && ctx.User.Gem > 0
}

func (h *ArbiterRitualHandler) Execute(ctx *model.Context) error {
	if ctx.User.Gem <= 0 {
		return nil
	}
	ctx.User.Gem--
	enterForm(ctx.User, model.FormArbiterJudgment)
	ctx.User.MaxHand = 5
	ctx.Game.Log(fmt.Sprintf("%s 发动 [仲裁仪式]，进入审判形态，手牌上限恒定为5", ctx.User.Name))
	return nil
}

func (h *ArbiterRitualBreakHandler) CanUse(ctx *model.Context) bool {
	return hasForm(ctx.User, model.FormArbiterJudgment)
}

func (h *ArbiterRitualBreakHandler) Execute(ctx *model.Context) error {
	leaveForm(ctx.User, model.FormArbiterJudgment)
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
	return getToken(ctx.User, "judgment") > 0
}

func (h *ArbiterDoomsdayHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("末日审判需要目标")
	}
	dmg := getToken(ctx.User, "judgment")
	setToken(ctx.User, "judgment", 0)
	if dmg > 0 {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, dmg, model.MagicAttack)
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [末日审判]，对 %s 造成%d点法术伤害", ctx.User.Name, ctx.Target.Name, dmg))
	return nil
}

func (h *ArbiterBalanceHandler) CanUse(ctx *model.Context) bool {
	return skills.CanPayCrystalLike(ctx, 1)
}

func (h *ArbiterBalanceHandler) Execute(ctx *model.Context) error {
	// 资源扣除由 UseSkill 统一处理，这里不重复扣费。
	v := addToken(ctx.User, "judgment", 1, 0, 4)
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
