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

// findHundredDragonLockHolder 在 model.IGameEngine 上查找锁定卡所在的角色（技能 Execute 侧使用）。
func findHundredDragonLockHolder(game model.IGameEngine, fighter *model.Player) *model.Player {
	if game == nil || fighter == nil {
		return nil
	}
	for _, p := range game.GetAllPlayers() {
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectFighterHundredDragonLock && fc.SourceID == fighter.ID {
				return p
			}
		}
	}
	return nil
}

func clearHundredDragonLock(rt engineplayer.ChoiceRuntime, fighter *model.Player) {
	if rt == nil || fighter == nil {
		return
	}
	rt.RemoveEffectCard(fighter, model.EffectFighterHundredDragonLock, false)
}

// LockedTarget 返回百式幻龙拳锁定的目标：以目标面前的锁定专属效果卡为准。
// 百式状态跨回合持续，而 TurnState 每回合重置，所以不能把锁定目标存在 SkillFlowState 里。
func LockedTarget(rt engineplayer.ChoiceRuntime, p *model.Player) *model.Player {
	if rt == nil || p == nil {
		return nil
	}
	holder, _ := rt.FindEffectCard(p, model.EffectFighterHundredDragonLock)
	return holder
}

func ClearHundredDragon(rt engineplayer.ChoiceRuntime, p *model.Player, logLine string) bool {
	if p == nil || !engineplayer.IsCharacter(p, "fighter") {
		return false
	}
	defer rt.PoseChangeGuard()
	active := engineplayer.HasForm(p, model.FormFighterHundredDragon) || LockedTarget(rt, p) != nil
	engineplayer.ClearForm(p, model.FormFighterHundredDragon)
	clearHundredDragonLock(rt, p)
	rt.ClearRuleModifiersByModifierID(p.ID, hundredDragonDisableChargeModifierID)
	if active && logLine != "" {
		rt.Log(logLine)
	}
	return active
}
