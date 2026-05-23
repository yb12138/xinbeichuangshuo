// gameflow: 神箭手 Timing Hook 实现。

package archer

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const preciseShotDamageModifierID = "archer_precise_shot_damage_delta"

// damageCalculateHook 精准射击：确认发动后，本次主动攻击伤害 -1。
func damageCalculateHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "archer") {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack || ctx.CounterInitiator != "" {
		return player.TimingHookResult{}
	}
	action := model.Action{Type: ctx.ActionType, Card: ctx.Card}
	delta := rt.ConsumeAttackDamageRuleBonus(p, preciseShotDamageModifierID, action)
	if delta == 0 {
		return player.TimingHookResult{}
	}
	rt.Log(fmt.Sprintf("[Passive] %s 的 [精准射击] 生效，本次主动攻击伤害 %d", p.Name, delta))
	return player.TimingHookResult{DamageDelta: delta}
}
