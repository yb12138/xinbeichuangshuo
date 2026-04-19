// gameflow: 灵魂术士：灵魂链接与转伤辅助。
package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) findSoulLink(sorcerer *model.Player) (*model.Player, *model.FieldCard) {
	return e.findExclusiveEffectCard(sorcerer, model.EffectSoulLink)
}

func (e *GameEngine) placeSoulLink(sorcerer *model.Player, target *model.Player, card model.Card) error {
	if sorcerer == nil || target == nil {
		return fmt.Errorf("放置灵魂链接时角色不存在")
	}
	if target.Camp != sorcerer.Camp || target.ID == sorcerer.ID {
		return fmt.Errorf("灵魂链接只能放置于队友")
	}
	if holder, _ := e.findSoulLink(sorcerer); holder != nil {
		return fmt.Errorf("灵魂链接已绑定，不能再次放置或移除")
	}
	return e.attachExclusiveEffectCard(sorcerer, target, model.EffectSoulLink, card)
}

func soulSorcererBlue(player *model.Player) int {
	return tokenValueBounded(player, "ss_blue_soul", soulSorcererBlueCapEngine)
}

func addSoulSorcererBlue(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "ss_blue_soul", delta, soulSorcererBlueCapEngine)
}

func soulSorcererYellow(player *model.Player) int {
	return tokenValueBounded(player, "ss_yellow_soul", soulSorcererYellowCapEngine)
}

func addSoulSorcererYellow(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "ss_yellow_soul", delta, soulSorcererYellowCapEngine)
}

func (e *GameEngine) applySoulSorcererSoulDevour(victim *model.Player, finalLoss int, fromDamageDraw bool) {
	if e == nil || victim == nil || finalLoss <= 0 || !fromDamageDraw {
		return
	}
	for _, player := range e.GetAllPlayers() {
		if player == nil || player.Camp != victim.Camp || !e.isSoulSorcerer(player) {
			continue
		}
		before := soulSorcererYellow(player)
		after := addSoulSorcererYellow(player, finalLoss)
		e.Log(fmt.Sprintf("%s 的 [灵魂吞噬] 触发：黄色灵魂 +%d（%d→%d）", player.Name, finalLoss, before, after))
	}
}

// maybeSoulLinkTransfer 在承受伤害前检查灵魂链接转伤流程。
// 返回 true 表示已产生中断，状态机应暂停等待玩家选择。
func (e *GameEngine) maybeSoulLinkTransfer(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 ||
		pd.HasCheck(model.PendingDamageCheckFromSoulLink) ||
		pd.HasCheck(model.PendingDamageCheckSoulLinkChecked) {
		return false
	}
	pd.SetCheck(model.PendingDamageCheckSoulLinkChecked, true)

	target := e.State.Players[pd.TargetID]
	if target == nil {
		return false
	}

	var sorcerer *model.Player
	var counterpart *model.Player
	// 场景1：灵魂术士本人受伤，另一方是其链接队友。
	if e.isSoulSorcerer(target) {
		holder, _ := e.findSoulLink(target)
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
			p := e.State.Players[fc.SourceID]
			if p == nil || !e.isSoulSorcerer(p) {
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

	blue := soulSorcererBlue(sorcerer)
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

	e.PushInterrupt(&model.Interrupt{
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
	e.Log(fmt.Sprintf("%s 的 [灵魂链接] 可触发：是否移除蓝色灵魂转移伤害（最多%d）", sorcerer.Name, maxX))
	return true
}
