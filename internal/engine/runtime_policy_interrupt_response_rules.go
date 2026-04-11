// gameflow: 中断内响应技能列表规则。

package engine

import (
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

func attackStartMoonGoddessMedusaInterruptHook(e *GameEngine, attacker *model.Player, target *model.Player, currentAction *model.QueuedAction, userCtx *model.Context) bool {
	if e == nil || attacker == nil || target == nil || currentAction == nil {
		return false
	}
	return e.maybeMoonGoddessMedusa(attacker, target, currentAction.SourceSkill, currentAction.Card, userCtx)
}

func actionEndHolySwordInterruptHook(e *GameEngine, ctx *model.Context) bool {
	if e == nil {
		return false
	}
	return e.maybeHolySwordDrawFromPhaseEndCtx(ctx)
}

func augmentBeastSamuraiResponseSkillIDs(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil || ctx == nil || ctx.Timing != model.TimingOnActionEnd || ctx.EventCtx == nil || ctx.EventCtx.ActionType != model.ActionAttack || ctx.User == nil {
		return skillIDs
	}
	if !sd.engine.isBeastSamurai(ctx.User) || containsSkillID(skillIDs, "bs_one_strike_no_thought") || sd.engine.beastSamuraiZanshin(ctx.User) < beastSamuraiZanshinCapEngine {
		return skillIDs
	}
	skillDef := findCharacterSkill(ctx.User.Character, "bs_one_strike_no_thought")
	if skillDef == nil {
		return skillIDs
	}
	handler := skills.GetHandler(skillDef.LogicHandler)
	if handler == nil || !handler.CanUse(ctx) {
		return skillIDs
	}
	return append(skillIDs, skillDef.ID)
}

func normalizeFighterResponseSkillIDs(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil || len(skillIDs) <= 1 || ctx == nil || ctx.User == nil {
		return skillIDs
	}
	if ctx.Timing != model.TimingOnAttackDeclared || !sd.engine.isFighter(ctx.User) || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil || ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return skillIDs
	}
	hasCharge := false
	hasBurst := false
	for _, sid := range skillIDs {
		if sid == "fighter_charge_strike" {
			hasCharge = true
		} else if sid == "fighter_burst_crash" {
			hasBurst = true
		}
	}
	if hasCharge && hasBurst {
		return []string{"fighter_charge_strike"}
	}
	return skillIDs
}
