// gameflow: 血色剑灵 Timing Hook 实现。

package crimson_sword_spirit

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// afterApplyHook 伤害应用后重置血色屏障锁定。
func afterApplyHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	target := rt.GetPlayers()[ctx.TargetID]
	if target == nil {
		return player.TimingHookResult{}
	}
	player.EnsurePlayerSkillFlowState(target)
	target.TurnState.SkillFlowState["css_blood_barrier_lock"] = 0
	return player.TimingHookResult{}
}

// healResistHook 血蔷薇庭院：场上效果在场期间，所有伤害均不可被治疗抵伤。
func healResistHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	if ctx.PendingDamage == nil || ctx.PendingDamage.IgnoreHeal {
		return player.TimingHookResult{}
	}
	for _, p := range rt.GetPlayers() {
		if p == nil || !player.IsCharacter(p, "crimson_sword_spirit") {
			continue
		}
		for _, fc := range p.Field {
			if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectRoseCourtyard {
				ctx.PendingDamage.IgnoreHeal = true
				return player.TimingHookResult{}
			}
		}
	}
	return player.TimingHookResult{}
}

// turnEndHook 回合结束：血蔷薇庭院移回专属卡区。
func turnEndHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "crimson_sword_spirit") {
		return player.TimingHookResult{}
	}
	hasCourtyard := false
	for _, fc := range p.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectRoseCourtyard {
			hasCourtyard = true
			break
		}
	}
	if !hasCourtyard {
		return player.TimingHookResult{}
	}
	p.Tokens["css_blood_cap"] = 3
	if p.Tokens["css_blood"] > 3 {
		p.Tokens["css_blood"] = 3
	}
	if rt.RemoveExclusiveEffectCard(p, model.EffectRoseCourtyard, true) {
		rt.Log(fmt.Sprintf("%s 的 [血蔷薇庭院] 回合结束移回专属卡区", p.Name))
	} else {
		rt.Log(fmt.Sprintf("%s 的 [血蔷薇庭院] 回合结束失效", p.Name))
	}
	return player.TimingHookResult{}
}
