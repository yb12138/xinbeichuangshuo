// gameflow: 魔剑士 Timing Hook 实现。

package magic_swordsman

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// damageCalculateHook 魔剑士·暗影之力被动增伤：暗影形态下主动攻击伤害 +1。
func damageCalculateHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "magic_swordsman") {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack || ctx.CounterInitiator != "" {
		return player.TimingHookResult{}
	}
	if !rt.HasForm(p, model.FormMagicSwordsmanShadow) {
		return player.TimingHookResult{}
	}
	rt.Log(fmt.Sprintf("[Passive] %s 的 [暗影之力] 生效，伤害 +1", p.Name))
	return player.TimingHookResult{DamageDelta: 1}
}

// attackStateResetHook resets magic swordsman attack-related state when a new attack is declared.
func attackStateResetHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] = 0
	return player.TimingHookResult{}
}

// attackGatingHook applies magic swordsman yellow spring pending no-counter gating.
func attackGatingHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || p.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	if ctx.AttackInfo == nil {
		return player.TimingHookResult{}
	}
	ctx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	return player.TimingHookResult{}
}

func postAttackHitHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if p.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] = 0
	maxHand := rt.GetMaxHand(p)
	if len(p.Hand) < maxHand {
		rt.DrawCards(p.ID, maxHand-len(p.Hand))
	}
	if len(p.Hand) >= 2 {
		rt.PushDiscardChoiceInterrupt(p.ID, map[string]interface{}{
			"discard_count": 2,
			"stay_in_turn":  true,
			"prompt":        "【黄泉震颤】攻击命中后，请弃置2张牌：",
		})
		return player.TimingHookResult{Interrupted: true}
	}
	return player.TimingHookResult{}
}

// beforeActionShadowReleaseHook 行动开始时检查是否脱离暗影形态。
func beforeActionShadowReleaseHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "magic_swordsman") {
		return player.TimingHookResult{}
	}
	if p.Tokens == nil || p.TurnState.HasUsedActionSkill {
		return player.TimingHookResult{}
	}
	if !rt.HasForm(p, model.FormMagicSwordsmanShadow) {
		return player.TimingHookResult{}
	}
	defer rt.PoseChangeGuard()
	rt.ClearForm(p, model.FormMagicSwordsmanShadow)
	rt.Log(fmt.Sprintf("%s 脱离暗影形态并转正", p.Name))
	return player.TimingHookResult{}
}

// cannotActFollowupHook 魔剑士无法行动后续处理：全法术手牌时继续重摸。
func cannotActFollowupHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := ctx.Player
	if p == nil || !rt.IsCharacter(p, "magic_swordsman") {
		return player.TimingHookResult{}
	}
	for len(p.Hand) > 0 {
		hasAttack := false
		allMagic := true
		for _, c := range p.Hand {
			if c.Type == model.CardTypeAttack {
				hasAttack = true
				break
			}
			if c.Type != model.CardTypeMagic {
				allMagic = false
			}
		}
		if hasAttack || !allMagic {
			break
		}
		redrawCount := len(p.Hand)
		rt.NotifyCardRevealed(p.ID, append([]model.Card{}, p.Hand...), "discard")
		rt.AddToDiscardPile(p.Hand...)
		p.Hand = p.Hand[:0]
		drawn := rt.DrawCardsRaw(p.ID, redrawCount)
		p.Hand = append(p.Hand, drawn...)
		rt.NotifyDrawCards(p.ID, redrawCount, "magic_swordsman_redraw")
		rt.Log(fmt.Sprintf("[Action] %s 触发魔剑士重摸：全法术手牌已弃置并重摸%d张", p.Name, redrawCount))
	}
	return player.TimingHookResult{}
}
