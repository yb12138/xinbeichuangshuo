// gameflow: 格斗家攻击未命中 Timing Hook 实现。

package fighter

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func attackMissHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "fighter") {
		return player.TimingHookResult{}
	}
	if !ctx.ForceFighterChargeMiss && p.TurnState.SkillFlowState["fighter_charge_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	p.TurnState.SkillFlowState["fighter_charge_pending"] = 0
	damage := p.Tokens["fighter_qi"]
	if damage < 1 {
		damage = 1
	}
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   p.ID,
		TargetID:   p.ID,
		Damage:     damage,
		DamageType: model.MagicDamage,
	})
	rt.Log(fmt.Sprintf("%s 的 [蓄力一击] 未命中分支生效：对自己造成%d点法术伤害", p.Name, damage))
	return player.TimingHookResult{}
}
