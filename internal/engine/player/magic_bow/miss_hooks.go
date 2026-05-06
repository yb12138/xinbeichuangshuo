// gameflow: 魔弓攻击未命中 Timing Hook 实现。

package magic_bow

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func attackMissHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if p.TurnState.SkillFlowState["mb_magic_pierce_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	p.TurnState.SkillFlowState["mb_magic_pierce_pending"] = 0
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.SourceID,
		TargetID:   ctx.TargetID,
		Damage:     3,
		DamageType: model.MagicAttack,
	})
	rt.Log(fmt.Sprintf("%s 的 [魔贯冲击] 未命中：对 %s 造成3点法术伤害", p.Name, ctx.TargetID))
	return player.TimingHookResult{}
}
