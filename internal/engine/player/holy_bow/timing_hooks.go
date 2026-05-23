// gameflow: 圣弓 Timing Hook 实现。

package holy_bow

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// turnStartResetHook 回合开始重置圣弓计数器。
func turnStartResetHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !player.IsCharacter(p, "holy_bow") {
		return player.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["hb_special"] = 0
	p.TurnState.UsedSkillCounts["hb_auto_fill"] = 0
	return player.TimingHookResult{}
}

// turnEndAutoFillHook 回合结束自动填充检查。
func turnEndAutoFillHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !player.IsCharacter(p, "holy_bow") {
		return player.TimingHookResult{}
	}
	if p.TurnState.UsedSkillCounts["hb_auto_fill"] > 0 {
		return player.TimingHookResult{}
	}
	if p.TurnState.UsedSkillCounts["hb_special"] <= 0 {
		var resourceModes []string
		if rt.CanPayCrystalCost(p.ID, 1) {
			resourceModes = append(resourceModes, "crystal")
		}
		if p.Gem > 0 {
			resourceModes = append(resourceModes, "gem")
		}
		if len(resourceModes) > 0 {
			p.TurnState.UsedSkillCounts["hb_auto_fill"] = 1
			rt.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: p.ID,
				Context: map[string]interface{}{
					"choice_type":    "hb_auto_fill_resource",
					"user_id":        p.ID,
					"resource_modes": resourceModes,
				},
			})
			rt.Log(fmt.Sprintf("%s 的 [自动填充] 触发：请选择消耗资源与增益", p.Name))
			return player.TimingHookResult{Interrupted: true}
		}
	}
	p.TurnState.UsedSkillCounts["hb_auto_fill"] = 1
	return player.TimingHookResult{}
}

// damageCalculateHook 圣弓·天之弓被动：非圣命格主动攻击伤害 -1。
func damageCalculateHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "holy_bow") {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack || ctx.CounterInitiator != "" || ctx.Card == nil {
		return player.TimingHookResult{}
	}
	if strings.TrimSpace(ctx.Card.Faction) == "圣" {
		return player.TimingHookResult{}
	}
	rt.Log(fmt.Sprintf("[Passive] %s 的 [天之弓] 生效：非圣命格主动攻击伤害 -1", p.Name))
	return player.TimingHookResult{DamageDelta: -1}
}

func postAttackHitHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !player.IsCharacter(p, "holy_bow") {
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
