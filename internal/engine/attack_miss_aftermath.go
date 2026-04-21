// gameflow: 攻击未命中后的后续（通过 TimingOnAttackMiss 分发到各角色 TimingHookSpec）。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
	beastsamurai "starcup-engine/internal/engine/player/beast_samurai"
	"starcup-engine/internal/model"
)

func (e *GameEngine) resolveMagicBowPierceMiss(attackerID, targetID string, attackCard *model.Card, isCounter bool) {
	e.resolveMagicBowPierceMissWithOverride(attackerID, targetID, attackCard, false, false, isCounter)
}

func (e *GameEngine) resolveMagicBowPierceMissWithOverride(attackerID, targetID string, attackCard *model.Card, forceHeroRoarMiss, forceFighterChargeMiss, isCounter bool) {
	e.dispatchAllRoleTimingHooks(engineplayer.TimingOnAttackMiss, engineplayer.TimingHookContext{
		SourceID:               attackerID,
		TargetID:               targetID,
		Card:                   attackCard,
		ForceHeroRoarMiss:      forceHeroRoarMiss,
		ForceFighterChargeMiss: forceFighterChargeMiss,
		IsCounter:              isCounter,
	})
	// 系统级清理：兽武者攻击令牌清除
	beastsamurai.ClearAttackTokens(e.State.Players[attackerID])
}
