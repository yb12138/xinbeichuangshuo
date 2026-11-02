package bard

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const eternalMovementName = "永恒乐章"

func eternalMovementCard(bard *model.Player) model.Card {
	id := "bd_eternal_movement"
	if bard != nil && bard.ID != "" {
		id = "bd_eternal_movement_" + bard.ID
	}
	return model.Card{
		ID:          id,
		Name:        eternalMovementName,
		Type:        model.CardTypeMagic,
		Element:     model.ElementDark,
		Description: "吟游诗人的永恒乐章指示牌",
	}
}

func FindEternalMovement(rt engineplayer.ChoiceRuntime, bard *model.Player) (*model.Player, *model.FieldCard) {
	return rt.FindEffectCard(bard, model.EffectBardEternalMovement)
}

func EternalHolderID(rt engineplayer.ChoiceRuntime, bard *model.Player) string {
	holder, _ := FindEternalMovement(rt, bard)
	if holder == nil {
		return ""
	}
	return holder.ID
}

func RemoveEternalMovement(rt engineplayer.ChoiceRuntime, bard *model.Player) bool {
	holder, card, ok := rt.DetachEffectCard(bard, model.EffectBardEternalMovement)
	if !ok {
		return false
	}
	bard.RestoreExclusiveCard(card)
	rt.EmitBuffRemovedDispatch(bard.ID, holder.ID, model.EffectBardEternalMovement)
	return true
}

// PlaceEternalMovement 把吟游诗人唯一的那张永恒乐章（专属卡区里的开局专属牌）放到目标面前。
// 永恒乐章只有一张：放置时从专属卡区移出，移除时放回（RemoveEternalMovement），
// 不能另造一张，否则专属卡区里那张不动、移除时再放回一张，就变成两张。
func PlaceEternalMovement(rt engineplayer.ChoiceRuntime, bard, target *model.Player) error {
	if bard == nil || target == nil {
		return fmt.Errorf("放置永恒乐章时角色不存在")
	}
	if target.Camp != bard.Camp {
		return fmt.Errorf("永恒乐章只能放置在我方角色面前")
	}
	// 已在场上（例如放到另一名队友面前）：先收回专属卡区再放出，保持同一张牌
	RemoveEternalMovement(rt, bard)
	card, ok := takeEternalMovementCard(bard)
	if !ok {
		card = eternalMovementCard(bard)
	}
	return rt.AttachEffectCard(bard, target, model.EffectBardEternalMovement, card)
}

// takeEternalMovementCard 从专属卡区取出永恒乐章（按牌名匹配，开局专属牌与兼容 ID 都认）。
func takeEternalMovementCard(bard *model.Player) (model.Card, bool) {
	if bard == nil {
		return model.Card{}, false
	}
	for i, c := range bard.ExclusiveCards {
		if c.Name != eternalMovementName {
			continue
		}
		bard.ExclusiveCards = append(bard.ExclusiveCards[:i:i], bard.ExclusiveCards[i+1:]...)
		return c, true
	}
	return model.Card{}, false
}
