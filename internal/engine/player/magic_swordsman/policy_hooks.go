// gameflow: 魔剑士策略 Hook 声明式注册。

package magic_swordsman

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// PolicySpecs 导出魔剑士策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 暗元素响应策略（阻止非魔剑士响应暗元素攻击）
		{Type: engineplayer.PolicyCombatInteraction, Priority: 400, Hook: darkElementResponseHook},
		// 反击魔弹策略
		{Type: engineplayer.PolicyCombatCounterCard, Priority: 100, Hook: shadowRejectMagicBulletHook},
	}
}

// darkElementResponseHook 暗元素响应策略。
// 非魔剑士玩家面对暗元素攻击时，攻击标记为不可响应。
func darkElementResponseHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	req := ctx.CombatRequest
	if req == nil || req.Card == nil || req.Card.Element != model.ElementDark {
		return engineplayer.PolicyHookResult{}
	}

	// 检查目标是否是魔剑士且可以使用暗影抗拒
	target := host.LookupPlayer(req.TargetID)
	currentTurnPlayerID := ""
	order := host.PlayerOrder()
	turn := host.CurrentTurn()
	if len(order) > 0 && turn >= 0 && turn < len(order) {
		currentTurnPlayerID = order[turn]
	}

	// 如果目标是魔剑士且可以使用暗影抗拒，则不拦截
	if target != nil && host.CanUseShadowRejectResponse(target, currentTurnPlayerID) {
		return engineplayer.PolicyHookResult{}
	}

	// 标记攻击为不可响应
	req.SetInterceptTag(model.CombatInterceptUnrespondable)
	return engineplayer.PolicyHookResult{Handled: true}
}

// shadowRejectMagicBulletHook 暗影抗拒魔弹反击策略。
// 非自己行动阶段使用魔弹应战时触发暗影抗拒。
func shadowRejectMagicBulletHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	req := ctx.CombatRequest
	player := ctx.Player
	card := ctx.CounterCard

	if req == nil || player == nil {
		return engineplayer.PolicyHookResult{}
	}

	// 检查是否可以使用暗影抗拒
	currentTurnPlayerID := ""
	order := host.PlayerOrder()
	turn := host.CurrentTurn()
	if len(order) > 0 && turn >= 0 && turn < len(order) {
		currentTurnPlayerID = order[turn]
	}

	if !host.CanUseShadowRejectResponse(player, currentTurnPlayerID) {
		return engineplayer.PolicyHookResult{}
	}

	// 检查是否是魔弹
	if card.Type != model.CardTypeMagic || card.Name != "魔弹" {
		return engineplayer.PolicyHookResult{}
	}

	// 触发暗影抗拒
	host.Log("[Combat] " + player.Name + " 触发[暗影抗拒]：非自己行动阶段使用【魔弹】应战")
	return engineplayer.PolicyHookResult{Handled: true, Card: card}
}
