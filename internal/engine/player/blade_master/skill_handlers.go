// gameflow: 风之剑圣技能处理器。

package blade_master

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type BaseHandler = engineplayer.BaseHandler

// --- Blade Master Skill Handlers ---

type WindFuryHandler struct{ BaseHandler }

func (h *WindFuryHandler) CanUse(ctx *model.Context) bool {
	if ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnActionEnd {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["wind_fury"] > 0 {
		return false
	}
	return true
}

func (h *WindFuryHandler) Execute(ctx *model.Context) error {
	model.AppendAttackAction(ctx.User, "风怒追击", model.ElementWind)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [风怒追击]，获得一次额外的[风系]攻击行动机会", ctx.User.Name))
	return nil
}

type HolySwordHandler struct{ BaseHandler }

func (h *HolySwordHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Timing != model.TimingOnAttackDeclared || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.ActionType != string(model.ActionAttack) || ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return ctx.User.TurnState.AttackCount+1 == 3
}

func (h *HolySwordHandler) Execute(ctx *model.Context) error {
	ctx.Game.Log(fmt.Sprintf("%s 的 [圣剑] 发动，本回合第3次攻击强制命中，对方无法抵挡", ctx.User.Name))
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		ctx.EventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptForceHit)
	}
	return nil
}

type SwordShadowHandler struct{ BaseHandler }

func (h *SwordShadowHandler) CanUse(ctx *model.Context) bool {
	if ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if ctx.Timing != model.TimingOnActionEnd {
		return false
	}
	if !engineplayer.CanPayCrystalLike(ctx, 1) {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["sword_shadow"] > 0 {
		return false
	}
	return true
}

func (h *SwordShadowHandler) Execute(ctx *model.Context) error {
	if !engineplayer.SpendCrystalLike(ctx, 1) {
		return fmt.Errorf("发动剑影失败：水晶不足（红宝石可替代）")
	}
	ctx.Game.Log(fmt.Sprintf("%s 消耗1蓝水晶（可由红宝石替代）发动 [剑影]", ctx.User.Name))
	model.AppendAttackAction(ctx.User, "剑影")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [剑影]，获得一次额外的攻击行动机会", ctx.User.Name))
	return nil
}

type GaleSkillHandler struct{ BaseHandler }

func (h *GaleSkillHandler) Execute(ctx *model.Context) error {
	model.AppendAttackAction(ctx.User, "疾风技")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [疾风技]，额外获得1次攻击行动", ctx.User.Name))
	return nil
}

type GaleSlashHandler struct{ BaseHandler }

func (h *GaleSlashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.Timing != model.TimingOnAttackDeclared || ctx.Target == nil || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.ActionType != string(model.ActionAttack) || ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	hasShield := false
	for _, fc := range ctx.Target.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			hasShield = true
			break
		}
	}
	return hasShield
}

func (h *GaleSlashHandler) Execute(ctx *model.Context) error {
	ctx.Game.Log(fmt.Sprintf("%s 发动 [列风技]，目标拥有圣盾，无视圣盾效果且目标无法应战", ctx.User.Name))
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		ctx.EventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
		ctx.EventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
	}
	return nil
}
