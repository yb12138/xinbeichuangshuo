// gameflow: 狂战士技能处理器。

package berserker

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// BaseHandler provides empty defaults for CanUse (always true) and Execute (no-op).
type BaseHandler = engineplayer.BaseHandler

// --- Berserker Skill Handlers ---

type BerserkerFrenzyHandler struct{ BaseHandler }

func (h *BerserkerFrenzyHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	if info == nil || info.ActionType != "Attack" {
		return false
	}
	return ctx.DamageSourceDealPhase() || ctx.AttackHitPhase()
}

func (h *BerserkerFrenzyHandler) Execute(ctx *model.Context) error {
	bonus := 0
	sourceDealPhase := ctx.DamageSourceDealPhase()
	switch {
	case sourceDealPhase:
		bonus = 1
	case ctx.AttackHitPhase():
		if len(ctx.User.Hand) > 3 {
			bonus = 1
		}
	default:
		return nil
	}
	if bonus <= 0 {
		return nil
	}
	*ctx.EventCtx.DamageVal += bonus
	if sourceDealPhase {
		ctx.Game.NotifyActionStep(fmt.Sprintf("%s 的被动技【狂化】生效：本次攻击伤害+1", model.GetPlayerDisplayName(ctx.User)))
		ctx.Game.Log(fmt.Sprintf("[Passive] %s 的【狂化】基础效果生效：伤害 +1", ctx.User.Name))
	} else {
		ctx.Game.NotifyActionStep(fmt.Sprintf("攻击命中，%s发动被动技狂化，当前其手牌数%d，伤害额外+1", model.GetPlayerDisplayName(ctx.User), len(ctx.User.Hand)))
		ctx.Game.Log(fmt.Sprintf("[Passive] %s 的【狂化】命中分支生效：手牌 %d，伤害再 +1", ctx.User.Name, len(ctx.User.Hand)))
	}
	return nil
}

type BerserkerTearHandler struct{ BaseHandler }

func (h *BerserkerTearHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || !ctx.AttackHitPhase() || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	// 2. [新增] 资源检查：必须至少有 1 颗宝石
	if ctx.User.Gem < 1 {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	return info.ActionType == "Attack"
}

func (h *BerserkerTearHandler) Execute(ctx *model.Context) error {
	// 撕裂：攻击命中时发动，覆盖主动攻击与应战攻击。
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		info := ctx.EventCtx.AttackInfo
		if info.ActionType == "Attack" {
			if ctx.EventCtx.DamageVal != nil {
				ctx.User.Gem -= 1
				*ctx.EventCtx.DamageVal += 2
				ctx.Game.NotifyActionStep(fmt.Sprintf("%s花费宝石发动撕裂，此次伤害再额外+2点", model.GetPlayerDisplayName(ctx.User)))
				ctx.Game.Log(fmt.Sprintf("%s 发动 [撕裂]，伤害 +2", ctx.User.Name))
			}
		}
	}
	return nil
}

type BloodRoarHandler struct{ BaseHandler }

func (h *BloodRoarHandler) Execute(ctx *model.Context) error {
	// 血腥咆哮：作为主动攻击打出时发动，若攻击的目标拥有的［治疗］为2，则本次攻击强制命中
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		info := ctx.EventCtx.AttackInfo
		// 规则：必须作为主动攻击打出 (非应战反弹)
		if info.ActionType == "Attack" && info.CounterInitiator == "" {
			target := ctx.Target
			if target != nil && target.Heal == 2 {
				info.SetInterceptTag(model.CombatInterceptForceHit)
				info.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
				ctx.Game.Log(fmt.Sprintf("%s 发动 [血腥咆哮]！目标治疗剂为2，强制命中且无视圣盾", ctx.User.Name))
			}
		}
	}
	return nil
}

type BloodBladeHandler struct{ BaseHandler }

func (h *BloodBladeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || !ctx.AttackHitPhase() || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	if info == nil || info.ActionType != "Attack" || info.CounterInitiator != "" || ctx.EventCtx.Card == nil {
		return false
	}
	if ctx.User.Character == nil {
		return false
	}
	if !ctx.EventCtx.Card.MatchExclusive(ctx.User.Character.ID, "血影狂刀") {
		return false
	}
	handCount := len(ctx.Target.Hand)
	return handCount == 2 || handCount == 3
}

func (h *BloodBladeHandler) Execute(ctx *model.Context) error {
	// 血影狂刀：作为主动攻击打出时发动，根据对手手牌数额外伤害
	if ctx.EventCtx != nil && ctx.EventCtx.DamageVal != nil {
		target := ctx.Target
		if target != nil {
			extraDamage := 0
			handCount := len(target.Hand)

			if handCount == 2 {
				extraDamage = 2
			} else if handCount == 3 {
				extraDamage = 1
			}

			if extraDamage > 0 {
				*ctx.EventCtx.DamageVal += extraDamage
				ctx.Game.Log(fmt.Sprintf("%s 发动 [血影狂刀]！对手手牌%d张，伤害 +%d", ctx.User.Name, handCount, extraDamage))
			}
		}
	}
	return nil
}
