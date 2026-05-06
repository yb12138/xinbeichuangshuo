// gameflow: 魔剑士策略 Hook 声明式注册。

package magic_swordsman

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// darkElementResponseHook 暗元素响应策略。
// 非魔剑士玩家面对暗元素攻击时，攻击标记为不可响应。
func darkElementResponseHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	req := ctx.CombatRequest
	if req == nil || req.Card == nil || req.Card.Element != model.ElementDark {
		return engineplayer.TimingHookResult{}
	}

	// 检查目标是否是魔剑士且可以使用暗影抗拒
	target := rt.GetPlayer(req.TargetID)
	currentTurnPlayerID := rt.CurrentTurnPlayerID()

	// 如果目标是魔剑士且可以使用暗影抗拒，则不拦截
	if target != nil && CanUseShadowRejectResponse(target, currentTurnPlayerID) {
		return engineplayer.TimingHookResult{}
	}

	// 标记攻击为不可响应
	req.SetInterceptTag(model.CombatInterceptUnrespondable)
	return engineplayer.TimingHookResult{Handled: true}
}

// shadowRejectMagicBulletHook 暗影抗拒魔弹反击策略。
// 非自己行动阶段使用魔弹应战时触发暗影抗拒。
func shadowRejectMagicBulletHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	req := ctx.CombatRequest
	player := ctx.Player
	card := ctx.CounterCard

	if req == nil || player == nil {
		return engineplayer.TimingHookResult{}
	}

	// 检查是否可以使用暗影抗拒
	currentTurnPlayerID := rt.CurrentTurnPlayerID()

	if !CanUseShadowRejectResponse(player, currentTurnPlayerID) {
		return engineplayer.TimingHookResult{}
	}

	// 检查是否是魔弹
	if card == nil || card.Type != model.CardTypeMagic || card.Name != "魔弹" {
		return engineplayer.TimingHookResult{}
	}

	// 触发暗影抗拒
	rt.Log("[Combat] " + player.Name + " 触发[暗影抗拒]：非自己行动阶段使用【魔弹】应战")
	return engineplayer.TimingHookResult{Handled: true, Card: *card}
}
