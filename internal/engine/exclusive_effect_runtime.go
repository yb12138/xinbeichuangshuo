package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) findSourceEffectCard(source *model.Player, effect model.EffectType) (*model.Player, *model.FieldCard) {
	if source == nil {
		return nil, nil
	}
	return e.FindFieldEffectBySource(effect, source.ID)
}

func (e *GameEngine) attachSourceEffectCard(source *model.Player, target *model.Player, effect model.EffectType, card model.Card) error {
	if source == nil || target == nil {
		return fmt.Errorf("放置来源场上牌时角色不存在")
	}
	if card.ID == "" || card.Name == "" {
		return fmt.Errorf("放置来源场上牌时卡牌无效")
	}
	target.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  target.ID,
		SourceID: source.ID,
		Mode:     model.FieldEffect,
		Effect:   effect,
		Trigger:  model.EffectTriggerManual,
	})
	return nil
}

func (e *GameEngine) detachSourceEffectCard(source *model.Player, effect model.EffectType) (*model.Player, model.Card, bool) {
	holder, fc := e.findSourceEffectCard(source, effect)
	if holder == nil || fc == nil {
		return nil, model.Card{}, false
	}
	holder.RemoveFieldCard(fc)
	return holder, fc.Card, true
}

func (e *GameEngine) findExclusiveEffectCard(source *model.Player, effect model.EffectType) (*model.Player, *model.FieldCard) {
	return e.findSourceEffectCard(source, effect)
}

func (e *GameEngine) attachExclusiveEffectCard(source *model.Player, target *model.Player, effect model.EffectType, card model.Card) error {
	return e.attachSourceEffectCard(source, target, effect, card)
}

func (e *GameEngine) detachExclusiveEffectCard(source *model.Player, effect model.EffectType) (*model.Player, model.Card, bool) {
	return e.detachSourceEffectCard(source, effect)
}

func (e *GameEngine) removeExclusiveEffectCard(source *model.Player, effect model.EffectType, restoreCard bool) bool {
	_, card, ok := e.detachSourceEffectCard(source, effect)
	if !ok {
		return false
	}
	if restoreCard && source != nil {
		source.RestoreExclusiveCard(card)
	} else {
		e.State.DiscardPile = append(e.State.DiscardPile, card)
	}
	return true
}
