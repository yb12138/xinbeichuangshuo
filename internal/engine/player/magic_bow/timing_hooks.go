// gameflow: 魔弓 Timing Hook 实现。

package magic_bow

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

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
	if _, ok := RemoveChargeByElement(p, model.ElementFire); ok {
		queue := rt.GetPendingDamageQueue()
		applied := false
		for i := range queue {
			queued := &queue[i]
			if !strings.EqualFold(string(queued.DamageType), string(model.AttackDamage)) {
				continue
			}
			queued.Damage++
			applied = true
			break
		}
		rt.SetPendingDamageQueue(queue)
		rt.Log(fmt.Sprintf("%s 的 [魔贯冲击] 命中追加生效：额外移除1个火系充能，本次攻击伤害+1", p.Name))
		if !applied {
			rt.Log("[Warn] 魔弓冲击命中追加未找到对应伤害条目，未能叠加伤害")
		}
	}
	p.TurnState.SkillFlowState["mb_magic_pierce_pending"] = 0
	return player.TimingHookResult{}
}
