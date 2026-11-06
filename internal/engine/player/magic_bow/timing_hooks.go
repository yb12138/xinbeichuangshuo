// gameflow: 魔弓 Timing Hook 实现。

package magic_bow

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const magicPierceDamageModifierID = "mb_magic_pierce_attack_bonus"

// damageCalculateHook 魔贯冲击：本次主动攻击伤害+1（记在规则修正上，不改实体攻击牌）。
func damageCalculateHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !player.IsCharacter(p, "magic_bow") {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack || ctx.CounterInitiator != "" {
		return player.TimingHookResult{}
	}
	action := model.Action{Type: ctx.ActionType, Card: ctx.Card}
	delta := rt.ConsumeAttackDamageRuleBonus(p, magicPierceDamageModifierID, action)
	if delta == 0 {
		return player.TimingHookResult{}
	}
	rt.Log(fmt.Sprintf("[Passive] %s 的 [魔贯冲击] 生效，本次主动攻击伤害 +%d", p.Name, delta))
	return player.TimingHookResult{DamageDelta: delta}
}

// discardMagicPierceDamageBonus 丢弃未被消费的魔贯冲击加伤（未命中、被圣盾抵挡等），
// 避免残留到本回合后续的攻击上。
func discardMagicPierceDamageBonus(rt player.HookRuntime, p *model.Player) {
	if p == nil {
		return
	}
	rt.ConsumeAttackDamageRuleBonus(p, magicPierceDamageModifierID, model.Action{Type: model.ActionAttack})
}

// attackStateResetHook 新一次攻击宣言时清理上一次攻击遗留的魔贯冲击加伤。
func attackStateResetHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !player.IsCharacter(p, "magic_bow") {
		return player.TimingHookResult{}
	}
	discardMagicPierceDamageBonus(rt, p)
	return player.TimingHookResult{}
}

// attackTargetCtxHook records the target player order index when an attack is declared.
func attackTargetCtxHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || p.TurnState.UsedSkillCounts == nil {
		return player.TimingHookResult{}
	}
	if !player.IsCharacter(p, "magic_bow") {
		return player.TimingHookResult{}
	}
	for i, pid := range rt.GetPlayerOrder() {
		if pid == ctx.TargetID {
			p.TurnState.UsedSkillCounts["mb_last_attack_target_order"] = i + 1
			return player.TimingHookResult{}
		}
	}
	return player.TimingHookResult{}
}

func postAttackHitHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if p.TurnState.SkillFlowState["mb_magic_pierce_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	if player.CoverCountByEffectAndElement(p, model.EffectMagicBowCharge, model.ElementFire) <= 0 {
		p.TurnState.SkillFlowState["mb_magic_pierce_pending"] = 0
		return player.TimingHookResult{}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_magic_pierce_hit_bonus",
			"user_id":     p.ID,
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [魔贯冲击] 命中，可额外移除1个火系充能使伤害+1", p.Name))
	return player.TimingHookResult{Interrupted: true}
}
