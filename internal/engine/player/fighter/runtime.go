package fighter

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const (
	hundredDragonDisableChargeModifierID = "rm_fighter_hundred_dragon_disable_charged_strike"
	hundredDragonLockCardID              = "fighter_hundred_dragon_lock"
)

func clearChargeStrikeState(game model.IGameEngine, player *model.Player) {
	if game == nil || player == nil {
		return
	}
	engineplayer.SetSkillFlowState(player, "fighter_charge_pending", 0)
	game.ClearRuleModifiersByModifierID(player.ID, "fighter_charge_attack_bonus")
}

func hundredDragonLockCard() model.Card {
	return model.Card{
		ID:      hundredDragonLockCardID,
		Name:    "百式幻龙拳锁定",
		Type:    model.CardTypeMagic,
		Element: model.ElementLight,
	}
}

func placeHundredDragonLock(rt engineplayer.ChoiceRuntime, fighter, target *model.Player) error {
	if rt == nil || fighter == nil || target == nil {
		return fmt.Errorf("百式幻龙拳锁定目标无效")
	}
	rt.RemoveEffectCard(fighter, model.EffectFighterHundredDragonLock, false)
	return rt.AttachEffectCard(fighter, target, model.EffectFighterHundredDragonLock, hundredDragonLockCard())
}

func clearHundredDragonLock(rt engineplayer.ChoiceRuntime, fighter *model.Player) {
	if rt == nil || fighter == nil {
		return
	}
	rt.RemoveEffectCard(fighter, model.EffectFighterHundredDragonLock, false)
}

func LockedTarget(rt engineplayer.ChoiceRuntime, p *model.Player) *model.Player {
	if p == nil {
		return nil
	}
	order := engineplayer.GetSkillFlowState(p, "fighter_hundred_dragon_target_order")
	if order <= 0 {
		return nil
	}
	orderIDs := rt.GetPlayerOrder()
	if order > len(orderIDs) {
		return nil
	}
	return rt.GetPlayers()[orderIDs[order-1]]
}

func ClearHundredDragon(rt engineplayer.ChoiceRuntime, p *model.Player, logLine string) bool {
	if p == nil || !engineplayer.IsCharacter(p, "fighter") {
		return false
	}
	defer rt.PoseChangeGuard()
	active := engineplayer.HasForm(p, model.FormFighterHundredDragon) || engineplayer.GetSkillFlowState(p, "fighter_hundred_dragon_target_order") > 0
	engineplayer.ClearForm(p, model.FormFighterHundredDragon)
	engineplayer.SetSkillFlowState(p, "fighter_hundred_dragon_target_order", 0)
	clearHundredDragonLock(rt, p)
	rt.ClearRuleModifiersByModifierID(p.ID, hundredDragonDisableChargeModifierID)
	if active && logLine != "" {
		rt.Log(logLine)
	}
	return active
}
