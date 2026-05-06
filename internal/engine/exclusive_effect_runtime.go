// gameflow: 场上效果牌按 FieldHook 结算（中毒等）。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// runFieldCardsForHook 按场上牌的结算钩子处理效果牌（如中毒等）。
func (e *GameEngine) runFieldCardsForHook(p *model.Player, hook model.FieldHook, ctx *model.Context) {
	var remain []*model.FieldCard

	for _, fc := range p.Field {
		if fc.Mode != model.FieldEffect || fc.Hook != hook {
			remain = append(remain, fc)
			continue
		}

		switch fc.Effect {
		case model.EffectPoison:
			e.applyPoisonEffect(p, fc.SourceID, ctx)
		case model.EffectWeak:
			remain = append(remain, fc)
			continue
		default:
			remain = append(remain, fc)
			continue
		}

		e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
		e.Log(fmt.Sprintf("[Field] %s 面前的【%s】触发效果并被弃置", p.Name, fc.Card.Name))
	}

	p.Field = remain
}

// applyPoisonEffect 应用中毒效果。
func (e *GameEngine) applyPoisonEffect(p *model.Player, sourceID string, ctx *model.Context) {
	allowCrimsonFaithHeal := sourceID != "" && sourceID == p.ID
	e.AddPendingDamage(model.PendingDamage{
		SourceID:              sourceID,
		TargetID:              p.ID,
		Damage:                1,
		DamageType:            "poison",
		AllowCrimsonFaithHeal: allowCrimsonFaithHeal,
	})
	e.Log(fmt.Sprintf("[Effect] %s 受到中毒伤害", p.Name))
}

// applyShieldEffect 应用圣盾效果。

// applyWeakEffect 应用虚弱效果。

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
		Hook:     model.FieldHookManual,
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
