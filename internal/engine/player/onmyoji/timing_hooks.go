// gameflow: 阴阳师 Timing Hook 实现。

package onmyoji

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// turnEndDarkRitualHook 回合结束时黑暗祭礼触发检查。
func turnEndDarkRitualHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := rt.GetPlayer(ctx.SourceID)
	if player == nil || !engineplayer.IsCharacter(player, "onmyoji") || player.Tokens == nil || player.Tokens["onmyoji_ghost_fire"] < 3 {
		return engineplayer.TimingHookResult{}
	}
	targetIDs := rt.CampEnemyIDs(player.Camp)
	if len(targetIDs) == 0 {
		return engineplayer.TimingHookResult{}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "onmyoji_dark_ritual_target",
			"user_id":     player.ID,
			"target_ids":  targetIDs,
			"ghost_fire":  player.Tokens["onmyoji_ghost_fire"],
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [黑暗祭礼] 触发，等待选择2点法术伤害目标", player.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}

// combatInteractionTimingHook 战斗交互阶段 TimingHook。
// 统一处理阴阳师式神咒束（代应战）和阴阳转换（同命格应战）中断检查。
func combatInteractionTimingHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	req := ctx.CombatRequest
	if req == nil {
		return engineplayer.TimingHookResult{}
	}
	if req.IsCounter || req.IsForcedHit || !req.CanBeResponded || req.Card == nil {
		return engineplayer.TimingHookResult{}
	}
	if req.Card.Element == model.ElementDark {
		return engineplayer.TimingHookResult{}
	}

	target := rt.GetPlayer(req.TargetID)
	attacker := rt.GetPlayer(req.AttackerID)
	if target == nil || attacker == nil {
		return engineplayer.TimingHookResult{}
	}
	if attacker.Camp == target.Camp {
		return engineplayer.TimingHookResult{}
	}

	// 阴阳转换：目标是阴阳师 → 自己用同命格应战
	if engineplayer.IsCharacter(target, "onmyoji") {
		return tryYinYangInterrupt(rt, req, target, attacker)
	}

	// 式神咒束：目标不是阴阳师 → 同阵营阴阳师代应战
	return tryBindingInterrupt(rt, req, target, attacker)
}

// tryYinYangInterrupt 阴阳转换中断：目标是阴阳师时，检查是否可使用同命格攻击牌应战。
func tryYinYangInterrupt(rt engineplayer.HookRuntime, req *model.CombatRequest, target, attacker *model.Player) engineplayer.TimingHookResult {
	if req.OnmyojiYinYangChecked {
		return engineplayer.TimingHookResult{}
	}
	req.OnmyojiYinYangChecked = true

	if !CanUseFactionCounter(req.Card) {
		return engineplayer.TimingHookResult{}
	}

	allOptions := CollectCounterOptions(target, req.Card)
	var factionOptions []map[string]interface{}
	for _, option := range allOptions {
		useFaction, _ := option["use_faction"].(bool)
		if useFaction {
			factionOptions = append(factionOptions, option)
		}
	}
	if len(factionOptions) == 0 {
		return engineplayer.TimingHookResult{}
	}

	counterTargetIDs := buildCounterTargetIDs(rt, attacker)
	if len(counterTargetIDs) == 0 {
		return engineplayer.TimingHookResult{}
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: target.ID,
		Context: map[string]interface{}{
			"choice_type":        "onmyoji_yinyang_confirm",
			"actor_id":           target.ID,
			"attacker_id":        req.AttackerID,
			"target_id":          req.TargetID,
			"card_options":       factionOptions,
			"counter_target_ids": counterTargetIDs,
		},
	})
	rt.Log(fmt.Sprintf("%s 可发动 [阴阳转换]，等待其确认", target.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}

// tryBindingInterrupt 式神咒束中断：同阵营阴阳师代替队友应战。
func tryBindingInterrupt(rt engineplayer.HookRuntime, req *model.CombatRequest, target, attacker *model.Player) engineplayer.TimingHookResult {
	if req.OnmyojiBindingChecked {
		return engineplayer.TimingHookResult{}
	}

	crt := rt.AsChoiceRuntime()
	if crt == nil {
		return engineplayer.TimingHookResult{}
	}

	counterTargetIDs := buildCounterTargetIDs(rt, attacker)
	if len(counterTargetIDs) == 0 {
		return engineplayer.TimingHookResult{}
	}

	for _, pid := range rt.GetPlayerOrder() {
		actor := rt.GetPlayer(pid)
		if actor == nil || actor.ID == target.ID {
			continue
		}
		if !engineplayer.IsCharacter(actor, "onmyoji") || actor.Camp != target.Camp {
			continue
		}
		if !engineplayer.HasForm(actor, model.FormOnmyojiShikigami) {
			continue
		}
		if !CanPayBindingCost(crt, actor.Camp) {
			continue
		}
		cardOptions := CollectCounterOptions(actor, req.Card)
		if len(cardOptions) == 0 {
			continue
		}
		req.OnmyojiBindingChecked = true
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: actor.ID,
			Context: map[string]interface{}{
				"choice_type":        "onmyoji_binding_confirm",
				"actor_id":           actor.ID,
				"attacker_id":        req.AttackerID,
				"target_id":          req.TargetID,
				"card_options":       cardOptions,
				"counter_target_ids": counterTargetIDs,
			},
		})
		rt.Log(fmt.Sprintf("%s 可发动 [式神咒束] 代应战，等待其确认", actor.Name))
		return engineplayer.TimingHookResult{Interrupted: true}
	}
	return engineplayer.TimingHookResult{}
}

// buildCounterTargetIDs 构建反击目标列表（攻击者阵营中非攻击者本人）。
func buildCounterTargetIDs(rt engineplayer.HookRuntime, attacker *model.Player) []string {
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		if pid == attacker.ID {
			continue
		}
		p := rt.GetPlayer(pid)
		if p == nil || p.Camp != attacker.Camp {
			continue
		}
		ids = append(ids, pid)
	}
	return ids
}
