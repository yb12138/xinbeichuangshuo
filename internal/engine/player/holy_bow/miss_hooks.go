// gameflow: 圣弓攻击未命中 Timing Hook 实现。

package holy_bow

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func attackMissHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "holy_bow") {
		return player.TimingHookResult{}
	}
	if p.TurnState.SkillFlowState["hb_shard_miss_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	p.TurnState.SkillFlowState["hb_shard_miss_pending"] = 0
	maxX := p.Heal
	if maxX > 2 {
		maxX = 2
	}
	if maxX <= 0 {
		rt.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中，但治疗不足，未触发后续效果", p.Name))
		return player.TimingHookResult{}
	}
	validX := ShardMissValidXValues(rt.GetAllPlayersMap(), rt.PlayerOrder(), p, maxX)
	if len(validX) == 0 {
		rt.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中，但没有能弃满牌数的队友可供选择", p.Name))
		return player.TimingHookResult{}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type": "hb_holy_shard_miss_confirm",
			"user_id":     p.ID,
			"target_id":   ctx.TargetID,
			"max_x":       maxX,
			"valid_x":     validX,
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中：可移除治疗并令队友弃牌", p.Name))
	return player.TimingHookResult{Interrupted: true}
}
