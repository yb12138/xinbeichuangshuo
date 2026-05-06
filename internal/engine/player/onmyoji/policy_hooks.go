// gameflow: 阴阳师策略 Hook 声明式注册。

package onmyoji

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// factionElementHook 命格元素匹配策略。
func factionElementHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	req := ctx.CombatRequest
	player := ctx.Player
	counterCard := ctx.CounterCard

	if req == nil || req.Card == nil || player == nil {
		return engineplayer.TimingHookResult{}
	}

	// 检查是否是阴阳师
	if !engineplayer.IsCharacter(player, "onmyoji") {
		return engineplayer.TimingHookResult{}
	}

	// 检查是否可以使用命格应战
	if !CanUseFactionCounter(req.Card) {
		return engineplayer.TimingHookResult{}
	}

	// 检查反击牌命格是否匹配
	if counterCard == nil || counterCard.Faction == "" || counterCard.Faction != req.Card.Faction {
		return engineplayer.TimingHookResult{}
	}

	return engineplayer.TimingHookResult{Handled: true, UseFaction: true}
}

// factionResolveHook 命格结算策略。
func factionResolveHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := ctx.Player
	counterCardPtr := ctx.CounterCardPtr
	useFaction := ctx.UseFaction

	if !useFaction || player == nil || counterCardPtr == nil {
		return engineplayer.TimingHookResult{}
	}

	ApplyFactionCounterBonuses(rt.AsChoiceRuntime(), player, counterCardPtr)
	return engineplayer.TimingHookResult{Handled: true}
}
