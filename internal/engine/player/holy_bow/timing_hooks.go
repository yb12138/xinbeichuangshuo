// gameflow: 圣弓 Timing Hook 实现。

package holy_bow

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/player"
)

func postAttackHitHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if p.TurnState.SkillFlowState["hb_shard_miss_pending"] > 0 {
		p.TurnState.SkillFlowState["hb_shard_miss_pending"] = 0
	}
	if ctx.IsCounter || ctx.Card == nil {
		return player.TimingHookResult{}
	}
	if strings.TrimSpace(ctx.Card.Faction) == "圣" {
		before := Faith(p)
		after := AddFaith(p, 1)
		if after > before {
			rt.Log(fmt.Sprintf("%s 的 [天之弓] 触发：信仰+1（当前%d）", p.Name, after))
		}
	}
	return player.TimingHookResult{}
}
