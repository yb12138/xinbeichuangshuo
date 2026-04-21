package bard

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func eternalMovementCard(bard *model.Player) model.Card {
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

func FindEternalMovement(rt engineplayer.ChoiceRuntime, bard *model.Player) (*model.Player, *model.FieldCard) {
	return rt.FindSourceEffectCard(bard, model.EffectBardEternalMovement)
}

func EternalHolderID(rt engineplayer.ChoiceRuntime, bard *model.Player) string {
	holder, _ := FindEternalMovement(rt, bard)
	if holder == nil {
		return ""
	}
	return holder.ID
}

func RemoveEternalMovement(rt engineplayer.ChoiceRuntime, bard *model.Player) bool {
	holder, card, ok := rt.DetachSourceEffectCard(bard, model.EffectBardEternalMovement)
	if !ok {
		return false
	}
	rt.AddToDiscardPile(card)
	rt.EmitBuffRemovedDispatch(bard.ID, holder.ID, model.EffectBardEternalMovement)
	return true
}

func PlaceEternalMovement(rt engineplayer.ChoiceRuntime, bard, target *model.Player) error {
	return PlaceEternalMovementWithCard(rt, bard, target, eternalMovementCard(bard))
}

func PlaceEternalMovementWithCard(rt engineplayer.ChoiceRuntime, bard, target *model.Player, card model.Card) error {
	if bard == nil || target == nil {
		return fmt.Errorf("放置永恒乐章时角色不存在")
	}
	if target.Camp != bard.Camp {
		return fmt.Errorf("永恒乐章只能放置在我方角色面前")
	}
	RemoveEternalMovement(rt, bard)
	return rt.AttachSourceEffectCard(bard, target, model.EffectBardEternalMovement, card)
}
