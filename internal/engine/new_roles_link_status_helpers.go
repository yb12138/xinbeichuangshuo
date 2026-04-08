package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func bardEternalMovementCard(bard *model.Player) model.Card {
	id := "bd_eternal_movement"
	if bard != nil && bard.ID != "" {
		id = "bd_eternal_movement_" + bard.ID
	}
	return model.Card{
		ID:          id,
		Name:        "永恒乐章",
		Type:        model.CardTypeMagic,
		Element:     model.ElementDark,
		Description: "吟游诗人的永恒乐章指示牌",
	}
}

func (e *GameEngine) findBardEternalMovement(bard *model.Player) (*model.Player, *model.FieldCard) {
	return e.findSourceEffectCard(bard, model.EffectBardEternalMovement)
}

func (e *GameEngine) bardEternalHolderID(bard *model.Player) string {
	holder, _ := e.findBardEternalMovement(bard)
	if holder == nil {
		return ""
	}
	return holder.ID
}

func (e *GameEngine) removeBardEternalMovement(bard *model.Player) bool {
	holder, card, ok := e.detachSourceEffectCard(bard, model.EffectBardEternalMovement)
	if !ok {
		return false
	}
	e.State.DiscardPile = append(e.State.DiscardPile, card)
	e.emitBuffRemovedTrigger(bard.ID, holder.ID, model.EffectBardEternalMovement)
	return true
}

func (e *GameEngine) placeBardEternalMovement(bard *model.Player, target *model.Player) error {
	return e.placeBardEternalMovementWithCard(bard, target, bardEternalMovementCard(bard))
}

func (e *GameEngine) placeBardEternalMovementWithCard(bard *model.Player, target *model.Player, card model.Card) error {
	if bard == nil || target == nil {
		return fmt.Errorf("放置永恒乐章时角色不存在")
	}
	if target.Camp != bard.Camp {
		return fmt.Errorf("永恒乐章只能放置在我方角色面前")
	}
	e.removeBardEternalMovement(bard)
	return e.attachSourceEffectCard(bard, target, model.EffectBardEternalMovement, card)
}

func (e *GameEngine) transferBardEternalMovement(bard *model.Player, target *model.Player) error {
	if bard == nil || target == nil {
		return fmt.Errorf("转移永恒乐章时角色不存在")
	}
	if target.Camp != bard.Camp {
		return fmt.Errorf("永恒乐章只能转移给我方角色")
	}
	holder, _ := e.findBardEternalMovement(bard)
	if holder == nil {
		return fmt.Errorf("当前没有永恒乐章可转移")
	}
	if holder.ID == target.ID {
		return fmt.Errorf("永恒乐章已在该角色面前")
	}
	holder, card, ok := e.detachSourceEffectCard(bard, model.EffectBardEternalMovement)
	if !ok || holder == nil {
		return fmt.Errorf("当前没有永恒乐章可转移")
	}
	return e.attachSourceEffectCard(bard, target, model.EffectBardEternalMovement, card)
}

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

func (e *GameEngine) findBloodPriestessSharedLife(priestess *model.Player) (*model.Player, *model.FieldCard) {
	return e.findExclusiveEffectCard(priestess, model.EffectBloodSharedLife)
}

func (e *GameEngine) detachBloodPriestessSharedLife(priestess *model.Player) (*model.Player, model.Card, bool) {
	return e.detachExclusiveEffectCard(priestess, model.EffectBloodSharedLife)
}

func (e *GameEngine) removeBloodPriestessSharedLife(priestess *model.Player, restoreCard bool) bool {
	return e.removeExclusiveEffectCard(priestess, model.EffectBloodSharedLife, restoreCard)
}

func (e *GameEngine) placeBloodPriestessSharedLife(priestess *model.Player, target *model.Player, card model.Card) error {
	if priestess == nil || target == nil {
		return fmt.Errorf("放置同生共死时角色不存在")
	}

	return e.attachExclusiveEffectCard(priestess, target, model.EffectBloodSharedLife, card)
}

func (e *GameEngine) hasFixedMaxHandCap(player *model.Player) bool {
	if player == nil {
		return false
	}
	if e.isPrayerMaster(player) && hasPrayerMasterPrayerForm(player) {
		return true
	}
	if e.isMagicLancer(player) && hasMagicLancerPhantomForm(player) {
		return true
	}
	if e.isHero(player) && hasHeroExhaustionForm(player) {
		return true
	}
	for _, fc := range player.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectMercy {
			return true
		}
	}
	return false
}

func (e *GameEngine) bloodPriestessSharedLifeDeltaFor(player *model.Player) int {
	if player == nil {
		return 0
	}
	delta := 0
	for _, pid := range e.State.PlayerOrder {
		holder := e.State.Players[pid]
		if holder == nil {
			continue
		}
		for _, fc := range holder.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
				continue
			}
			source := e.State.Players[fc.SourceID]
			if source == nil || !e.isBloodPriestess(source) {
				continue
			}
			change := -2
			if hasBloodPriestessBleedingForm(source) {
				change = 1
			}
			if source.ID == player.ID {
				delta += change
				continue
			}
			if fc.OwnerID == player.ID && !e.hasFixedMaxHandCap(player) {
				delta += change
			}
		}
	}
	return delta
}

func (e *GameEngine) enterBloodPriestessBleedingForm(player *model.Player, reason string) bool {
	if player == nil || !e.isBloodPriestess(player) {
		return false
	}
	if hasBloodPriestessBleedingForm(player) {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	enterBloodPriestessBleedingFormState(player)
	if reason == "" {
		reason = "因承受伤害导致我方士气下降"
	}
	e.Log(fmt.Sprintf("%s 的 [流血] 触发：%s，进入流血形态", player.Name, reason))
	e.dispatchOrientationChanges(beforePoses)
	return true
}

func (e *GameEngine) leaveBloodPriestessBleedingForm(player *model.Player, reason string) bool {
	if player == nil || !e.isBloodPriestess(player) {
		return false
	}
	if !hasBloodPriestessBleedingForm(player) {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	leaveBloodPriestessBleedingFormState(player)
	if reason == "" {
		reason = "行动结束时手牌少于3"
	}
	e.Log(fmt.Sprintf("%s 的 [流血·手牌不足脱离] 生效：%s，脱离流血形态", player.Name, reason))
	e.dispatchOrientationChanges(beforePoses)
	return true
}

func (e *GameEngine) resolveBloodPriestessBleedExitOnActionEnd() bool {
	if e == nil || e.State == nil {
		return false
	}
	released := false
	for _, pid := range e.State.PlayerOrder {
		player := e.State.Players[pid]
		if player == nil || !e.isBloodPriestess(player) {
			continue
		}
		if len(player.Hand) >= 3 {
			continue
		}
		if e.leaveBloodPriestessBleedingForm(player, "行动结束时手牌<3") {
			released = true
		}
	}
	return released
}

// maybeTriggerSoulLinkTransfer 在承受伤害前检查灵魂链接转伤流程。
// 返回 true 表示已产生中断，状态机应暂停等待玩家选择。
func (e *GameEngine) maybeTriggerSoulLinkTransfer(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 || pd.FromSoulLink || pd.SoulLinkChecked {
		return false
	}
	pd.SoulLinkChecked = true

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
