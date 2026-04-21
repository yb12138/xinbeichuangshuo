// gameflow: 英雄攻击未命中 Timing Hook 实现。

package hero

import (
	"fmt"

	"starcup-engine/internal/engine/player"
)

func attackMissHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "hero") {
		return player.TimingHookResult{}
	}
	if !ctx.ForceHeroRoarMiss && p.TurnState.UsedSkillCounts["hero_roar_active"] <= 0 {
		return player.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["hero_roar_active"] = 0
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	wisdom := p.Tokens["hero_wisdom"] + 1
	if wisdom > 3 {
		wisdom = 3
	}
	p.Tokens["hero_wisdom"] = wisdom
	rt.Log(fmt.Sprintf("%s 的 [怒吼] 未命中分支生效：智慧+1（当前%d）", p.Name, wisdom))
	return player.TimingHookResult{}
}
