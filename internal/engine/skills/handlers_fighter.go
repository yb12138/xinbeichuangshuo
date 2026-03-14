package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

const fighterQiCap = 8

type FighterPsiFieldHandler struct{ BaseHandler }

type FighterChargeStrikeHandler struct{ BaseHandler }

type FighterPsiBulletHandler struct{ BaseHandler }

type FighterHundredDragonHandler struct{ BaseHandler }

type FighterBurstCrashHandler struct{ BaseHandler }

type FighterWarGodDriveHandler struct{ BaseHandler }

type FighterWarGodDriveFollowupHandler struct{ BaseHandler }

func (h *FighterPsiFieldHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil {
		return false
	}
	return ctx.Trigger == model.TriggerOnDamageTaken && *ctx.TriggerCtx.DamageVal > 4
}

func (h *FighterPsiFieldHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil || ctx.User == nil || ctx.Game == nil {
		return nil
	}
	if *ctx.TriggerCtx.DamageVal > 4 {
		*ctx.TriggerCtx.DamageVal = 4
		ctx.Game.Log(fmt.Sprintf("%s 的 [念气力场] 生效：本次伤害被限制为4", ctx.User.Name))
	}
	return nil
}

func (h *FighterChargeStrikeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if getToken(ctx.User, "fighter_hundred_dragon_form") > 0 {
		return false
	}
	if getToken(ctx.User, "fighter_attack_start_skill_lock") > 0 {
		return false
	}
	return getToken(ctx.User, "fighter_qi") < fighterQiCap
}

func (h *FighterChargeStrikeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("蓄力一击上下文无效")
	}
	if getToken(ctx.User, "fighter_qi") >= fighterQiCap {
		return fmt.Errorf("斗气已达上限，不能发动蓄力一击")
	}
	qi := addToken(ctx.User, "fighter_qi", 1, 0, fighterQiCap)
	setToken(ctx.User, "fighter_attack_start_skill_lock", 1)
	setToken(ctx.User, "fighter_charge_pending", 1)
	setToken(ctx.User, "fighter_charge_damage_pending", 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [蓄力一击]：斗气+1（当前%d），本次攻击伤害额外+1；若未命中将按斗气自伤", ctx.User.Name, qi))
	return nil
}

func (h *FighterPsiBulletHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnPhaseEnd {
		return false
	}
	if ctx.TriggerCtx.ActionType != model.ActionMagic {
		return false
	}
	if getToken(ctx.User, "fighter_qi") >= fighterQiCap {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp != ctx.User.Camp {
			return true
		}
	}
	return false
}

func (h *FighterPsiBulletHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("念弹上下文无效")
	}
	if getToken(ctx.User, "fighter_qi") >= fighterQiCap {
		return fmt.Errorf("斗气已达上限，不能发动念弹")
	}
	targetIDs := make([]string, 0)
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil || p.Camp == ctx.User.Camp {
			continue
		}
		targetIDs = append(targetIDs, p.ID)
	}
	if len(targetIDs) == 0 {
		return nil
	}
	qi := addToken(ctx.User, "fighter_qi", 1, 0, fighterQiCap)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "fighter_psi_bullet_target",
			"user_id":     ctx.User.ID,
			"target_ids":  targetIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [念弹]：斗气+1（当前%d），请选择目标对手", ctx.User.Name, qi))
	return nil
}

func (h *FighterHundredDragonHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if getToken(ctx.User, "fighter_hundred_dragon_form") > 0 {
		return false
	}
	return getToken(ctx.User, "fighter_qi") >= 3
}

func (h *FighterHundredDragonHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("百式幻龙拳上下文无效")
	}
	if getToken(ctx.User, "fighter_qi") < 3 {
		return fmt.Errorf("斗气不足3，无法发动百式幻龙拳")
	}
	qi := addToken(ctx.User, "fighter_qi", -3, 0, fighterQiCap)
	setToken(ctx.User, "fighter_hundred_dragon_form", 1)
	setToken(ctx.User, "fighter_hundred_dragon_target_order", 0)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [百式幻龙拳]：移除3斗气（剩余%d），进入持续形态（主动攻击+2，应战攻击+1）", ctx.User.Name, qi))
	return nil
}

func (h *FighterBurstCrashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if getToken(ctx.User, "fighter_attack_start_skill_lock") > 0 {
		return false
	}
	return getToken(ctx.User, "fighter_qi") > 0
}

func (h *FighterBurstCrashHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("气绝崩击上下文无效")
	}
	if getToken(ctx.User, "fighter_qi") <= 0 {
		return fmt.Errorf("斗气不足，无法发动气绝崩击")
	}
	qi := addToken(ctx.User, "fighter_qi", -1, 0, fighterQiCap)
	setToken(ctx.User, "fighter_attack_start_skill_lock", 2)
	setToken(ctx.User, "fighter_qiburst_force_no_counter", 1)
	if ctx.TriggerCtx != nil && ctx.TriggerCtx.AttackInfo != nil {
		ctx.TriggerCtx.AttackInfo.CanBeResponded = false
	}
	if qi > 0 {
		ctx.Game.AddPendingDamage(model.PendingDamage{
			SourceID:   ctx.User.ID,
			TargetID:   ctx.User.ID,
			Damage:     qi,
			DamageType: "magic",
			Stage:      0,
		})
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [气绝崩击]：移除1斗气（剩余%d），本次攻击不可应战，并对自己造成%d点法术伤害", ctx.User.Name, qi, qi))
	return nil
}

func (h *FighterWarGodDriveHandler) CanUse(ctx *model.Context) bool {
	return canPayCrystalLike(ctx, 1)
}

func (h *FighterWarGodDriveHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("斗神天驱上下文无效")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("斗神天驱需要1点蓝水晶（红宝石可替代）")
	}
	discardCount := len(ctx.User.Hand) - 3
	if discardCount > 0 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptDiscard,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"skill_id": "fighter_war_god_drive_followup",
				"user_ctx": ctx,
				"min":      discardCount,
				"max":      discardCount,
				"prompt":   "【斗神天驱】请选择需要弃置的手牌：",
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [斗神天驱]：请先弃置%d张牌（弃到3张）", ctx.User.Name, discardCount))
		return nil
	}
	ctx.Game.Heal(ctx.User.ID, 2)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [斗神天驱]：手牌无需弃置，获得2点治疗", ctx.User.Name))
	return nil
}

func (h *FighterWarGodDriveFollowupHandler) CanUse(ctx *model.Context) bool { return false }

func (h *FighterWarGodDriveFollowupHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("斗神天驱后续结算上下文无效")
	}
	raw, _ := ctx.Selections["discard_indices"]
	indices := make([]int, 0)
	switch v := raw.(type) {
	case []int:
		indices = append(indices, v...)
	case []interface{}:
		for _, item := range v {
			if f, ok := item.(float64); ok {
				indices = append(indices, int(f))
			}
		}
	}
	discarded := removeHandByIndices(ctx.User, indices)
	if len(discarded) > 0 {
		ctx.Game.NotifyCardRevealed(ctx.User.ID, discarded, "discard")
		ctx.Selections["discardedCards"] = discarded
	}
	ctx.Game.Heal(ctx.User.ID, 2)
	ctx.Game.Log(fmt.Sprintf("%s 的 [斗神天驱] 后续结算完成：弃置%d张牌并获得2点治疗", ctx.User.Name, len(discarded)))
	return nil
}
