// gameflow: 阴阳师策略 Hook 声明式注册。

package onmyoji

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// PolicySpecs 导出阴阳师策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 反击元素策略
		{Type: engineplayer.PolicyCombatCounterElement, Priority: 100, Hook: factionElementHook},
		// 反击结算策略
		{Type: engineplayer.PolicyCombatCounterResolve, Priority: 100, Hook: factionResolveHook},
	}
}

// factionElementHook 命格元素匹配策略。
func factionElementHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	req := ctx.CombatRequest
	player := ctx.Player
	counterCard := ctx.CounterCard

	if req == nil || req.Card == nil || player == nil {
		return engineplayer.PolicyHookResult{}
	}

	// 检查是否是阴阳师
	if !host.IsCharacter(player, "onmyoji") {
		return engineplayer.PolicyHookResult{}
	}

	// 检查是否可以使用命格应战
	if !host.CanUseFactionCounter(req.Card) {
		return engineplayer.PolicyHookResult{}
	}

	// 检查反击牌命格是否匹配
	if counterCard.Faction == "" || counterCard.Faction != req.Card.Faction {
		return engineplayer.PolicyHookResult{}
	}

	return engineplayer.PolicyHookResult{Handled: true, UseFaction: true}
}

// factionResolveHook 命格结算策略。
func factionResolveHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	player := ctx.Player
	counterCardPtr := ctx.CounterCardPtr
	useFaction := ctx.UseFaction

	if !useFaction || player == nil || counterCardPtr == nil {
		return engineplayer.PolicyHookResult{}
	}

	host.ApplyFactionCounterBonuses(player, counterCardPtr)
	return engineplayer.PolicyHookResult{Handled: true}
}
