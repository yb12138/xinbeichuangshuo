// gameflow: 灵魂术士：灵魂链接与转伤辅助。

package soul_sorcerer

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// FindSoulLink 查找灵魂术士已放置的灵魂链接效果。
func FindSoulLink(rt engineplayer.ChoiceRuntime, sorcerer *model.Player) (*model.Player, *model.FieldCard) {
	return rt.FindEffectCard(sorcerer, model.EffectSoulLink)
}

// PlaceSoulLink 将灵魂链接放置于指定队友面前。
func PlaceSoulLink(rt engineplayer.ChoiceRuntime, sorcerer, target *model.Player, card model.Card) error {
	if sorcerer == nil || target == nil {
		return fmt.Errorf("放置灵魂链接时角色不存在")
	}
	if target.Camp != sorcerer.Camp || target.ID == sorcerer.ID {
		return fmt.Errorf("灵魂链接只能放置于队友")
	}
	if holder, _ := FindSoulLink(rt, sorcerer); holder != nil {
		return fmt.Errorf("灵魂链接已绑定，不能再次放置或移除")
	}
	return rt.AttachEffectCard(sorcerer, target, model.EffectSoulLink, card)
}

// ApplySoulDevour 灵魂吞噬：当队友因伤害导致士气下降时，灵魂术士获得黄色灵魂。
func ApplySoulDevour(rt model.IGameEngine, victim *model.Player, finalLoss int, fromDamageDraw bool) {
	if rt == nil || victim == nil || finalLoss <= 0 || !fromDamageDraw {
		return
	}
	for _, player := range rt.GetAllPlayers() {
		if player == nil || player.Camp != victim.Camp || !engineplayer.IsCharacter(player, "soul_sorcerer") {
			continue
		}
		before := YellowSoul(player)
		after := AddYellowSoul(player, finalLoss)
		rt.Log(fmt.Sprintf("%s 的 [灵魂吞噬] 触发：黄色灵魂 +%d（%d→%d）", player.Name, finalLoss, before, after))
	}
}

// MaybeSoulLinkTransfer 在承受伤害前检查灵魂链接转伤流程。
// 返回 true 表示已产生中断，状态机应暂停等待玩家选择。
func MaybeSoulLinkTransfer(rt engineplayer.ChoiceRuntime, pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 ||
		pd.HasCheck(model.PendingDamageCheckFromSoulLink) ||
		pd.HasCheck(model.PendingDamageCheckSoulLinkChecked) {
		return false
	}
	pd.SetCheck(model.PendingDamageCheckSoulLinkChecked, true)

	target := rt.GetPlayers()[pd.TargetID]
	if target == nil {
		return false
	}

	var sorcerer *model.Player
	var counterpart *model.Player
	// 场景1：灵魂术士本人受伤，另一方是其链接队友。
	if engineplayer.IsCharacter(target, "soul_sorcerer") {
		holder, _ := FindSoulLink(rt, target)
		if holder != nil {
			sorcerer = target
			counterpart = holder
		}
	} else {
		// 场景2：链接队友受伤，寻找来源为该灵魂术士的链接牌。
		for _, fc := range target.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectSoulLink {
				continue
			}
			p := rt.GetPlayers()[fc.SourceID]
			if p == nil || !engineplayer.IsCharacter(p, "soul_sorcerer") {
				continue
			}
			sorcerer = p
			counterpart = p
			break
		}
	}
	if sorcerer == nil || counterpart == nil {
		return false
	}

	blue := BlueSoul(sorcerer)
	if blue <= 0 {
		return false
	}
	maxX := pd.Damage
	if blue < maxX {
		maxX = blue
	}
	if maxX <= 0 {
		return false
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: sorcerer.ID,
		Context: map[string]interface{}{
			"choice_type":     "ss_link_transfer_x",
			"sorcerer_id":     sorcerer.ID,
			"damage_index":    0,
			"source_id":       pd.SourceID,
			"target_id":       pd.TargetID,
			"counterpart_id":  counterpart.ID,
			"max_x":           maxX,
			"original_damage": pd.Damage,
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [灵魂链接] 可触发：是否移除蓝色灵魂转移伤害（最多%d）", sorcerer.Name, maxX))
	return true
}
