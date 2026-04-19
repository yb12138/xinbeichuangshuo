// gameflow: 勇者 Timing Hook 实现。

package hero

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// postActionEndHook 攻击行动结束后：明镜止水水晶+1。
func postActionEndHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack {
		return player.TimingHookResult{}
	}
	if p.Tokens == nil || p.Tokens["hero_calm_end_crystal_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	p.Tokens["hero_calm_end_crystal_pending"]--
	if p.Tokens["hero_calm_end_crystal_pending"] < 0 {
		p.Tokens["hero_calm_end_crystal_pending"] = 0
	}
	capV := rt.GetPlayerEnergyCap(p)
	if p.Gem+p.Crystal < capV {
		p.Crystal++
		rt.Log(fmt.Sprintf("%s 的 [明镜止水] 结算：水晶+1", p.Name))
	} else {
		rt.Log(fmt.Sprintf("%s 的 [明镜止水] 结算：能量已满，水晶未增加", p.Name))
	}
	return player.TimingHookResult{}
}
