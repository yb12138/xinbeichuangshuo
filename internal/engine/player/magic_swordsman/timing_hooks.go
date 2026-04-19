// gameflow: 魔剑士 Timing Hook 实现。

package magic_swordsman

import (
	"starcup-engine/internal/engine/player"
)

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
