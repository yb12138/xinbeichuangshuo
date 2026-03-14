package engine

import (
	"fmt"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// Ensure GameEngine implements IGameEngine
var _ model.IGameEngine = (*GameEngine)(nil)

func (e *GameEngine) ModifyGem(camp string, amount int) {
	if amount > 0 {
		for i := 0; i < amount; i++ {
			e.addCampResource(model.Camp(camp), "gem")
		}
	} else if amount < 0 {
		// 扣除宝石逻辑 (暂时简单扣除，不涉及复杂的优先级因为通常是消耗)
		// 如果需要处理 "优先扣除水晶" 这种逻辑，需要明确
		// 暂时直接操作 State
		c := model.Camp(camp)
		if c == model.RedCamp {
			e.State.RedGems += amount // amount is negative
			if e.State.RedGems < 0 {
				e.State.RedGems = 0
			}
		} else {
			e.State.BlueGems += amount
			if e.State.BlueGems < 0 {
				e.State.BlueGems = 0
			}
		}
	}
}

func (e *GameEngine) ModifyCrystal(camp string, amount int) {
	if amount > 0 {
		for i := 0; i < amount; i++ {
			e.addCampResource(model.Camp(camp), "crystal")
		}
	} else if amount < 0 {
		c := model.Camp(camp)
		if c == model.RedCamp {
			e.State.RedCrystals += amount
			if e.State.RedCrystals < 0 {
				e.State.RedCrystals = 0
			}
		} else {
			e.State.BlueCrystals += amount
			if e.State.BlueCrystals < 0 {
				e.State.BlueCrystals = 0
			}
		}
	}
}

// GetUsableCrystal 返回“可用于支付蓝水晶消耗”的总量：
// 蓝水晶 + 可替代的红宝石。
func (e *GameEngine) GetUsableCrystal(playerID string) int {
	p := e.State.Players[playerID]
	if p == nil {
		return 0
	}
	return p.Crystal + p.Gem
}

func (e *GameEngine) CanPayCrystalCost(playerID string, amount int) bool {
	if amount <= 0 {
		return true
	}
	return e.GetUsableCrystal(playerID) >= amount
}

// ConsumeCrystalCost 结算“蓝水晶消耗，可由红宝石替代”。
// 扣除顺序：优先蓝水晶，再扣红宝石。
func (e *GameEngine) ConsumeCrystalCost(playerID string, amount int) bool {
	if amount <= 0 {
		return true
	}
	p := e.State.Players[playerID]
	if p == nil {
		return false
	}
	if p.Crystal+p.Gem < amount {
		return false
	}
	useCrystal := amount
	if useCrystal > p.Crystal {
		useCrystal = p.Crystal
	}
	p.Crystal -= useCrystal
	remain := amount - useCrystal
	if remain > 0 {
		p.Gem -= remain
	}
	return true
}

// canPaySkillEnergyCost 规则：
// 1) 宝石消耗必须由宝石支付（不可用水晶替代）；
// 2) 水晶消耗可由“剩余宝石”替代。
func canPaySkillEnergyCost(p *model.Player, gemCost, crystalCost int) bool {
	if p == nil {
		return false
	}
	if gemCost < 0 {
		gemCost = 0
	}
	if crystalCost < 0 {
		crystalCost = 0
	}
	if p.Gem < gemCost {
		return false
	}
	remainingGem := p.Gem - gemCost
	return p.Crystal+remainingGem >= crystalCost
}

func consumeSkillEnergyCost(p *model.Player, gemCost, crystalCost int) bool {
	if !canPaySkillEnergyCost(p, gemCost, crystalCost) {
		return false
	}
	if gemCost < 0 {
		gemCost = 0
	}
	if crystalCost < 0 {
		crystalCost = 0
	}
	p.Gem -= gemCost
	if p.Crystal >= crystalCost {
		p.Crystal -= crystalCost
		return true
	}
	needGemAsCrystal := crystalCost - p.Crystal
	p.Crystal = 0
	p.Gem -= needGemAsCrystal
	return true
}

func (e *GameEngine) DrawCards(playerID string, amount int) {
	p := e.State.Players[playerID]
	if p == nil {
		return
	}

	// 摸牌前触发（需携带 DrawCount 供水影等技能在中断后恢复摸牌）
	drawCount := amount
	drawEventCtx := &model.EventContext{
		Type:      model.EventBeforeDraw,
		SourceID:  playerID,
		TargetID:  playerID,
		DrawCount: &drawCount,
	}
	ctx := e.buildContext(p, nil, model.TriggerBeforeDraw, drawEventCtx)
	if p.Tokens != nil && p.Tokens["elf_ritual_suppress_overflow"] > 0 {
		ctx.Flags["preventOverflow"] = true
	}

	e.dispatcher.OnTrigger(model.TriggerBeforeDraw, ctx)

	if ctx.Flags["cancelDraw"] {
		e.Log(fmt.Sprintf("%s 的摸牌被取消", p.Name))
		return
	}

	// 真正摸牌
	cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, amount)
	e.State.Deck = newDeck
	e.State.DiscardPile = newDiscard
	p.Hand = append(p.Hand, cards...)
	e.NotifyDrawCards(playerID, amount, "draw")

	// 摸牌后触发
	ctx.Trigger = model.TriggerAfterDraw
	e.dispatcher.OnTrigger(model.TriggerAfterDraw, ctx)

	e.checkHandLimit(p, ctx)
	e.Log(fmt.Sprintf("%s 摸了 %d 张牌", p.Name, amount))
}

func (e *GameEngine) AppendToDiscard(cards []model.Card) {
	if len(cards) == 0 {
		return
	}
	e.State.DiscardPile = append(e.State.DiscardPile, cards...)
}

func (e *GameEngine) DiscardCard(card *model.FieldCard) error {
	e.State.DiscardPile = append(e.State.DiscardPile, card.Card)
	e.Log(fmt.Sprintf("%s 丢弃卡牌 %s (%s)", card.OwnerID, card.Card.Name, card.Card.Element))
	return nil
}

func (e *GameEngine) Heal(playerID string, amount int) {
	p := e.State.Players[playerID]
	if p == nil {
		return
	}
	// 保留“已存在的超上限治疗”（如圣光祈愈带来的超额治疗），
	// 避免常规治疗结算把当前治疗值错误下压。
	oldHeal := p.Heal
	p.Heal += amount
	healCap := p.MaxHeal
	if oldHeal > healCap {
		healCap = oldHeal
	}
	if p.Heal > healCap {
		p.Heal = healCap
	}
	e.addActionHeal(playerID, amount)
	e.Log(fmt.Sprintf("%s 获得了 %d 点治疗，当前治疗: %d", p.Name, amount, p.Heal))
}

func (e *GameEngine) InflictDamage(sourceID, targetID string, amount int, damageType string) {
	// 将伤害推入延迟伤害队列，以便支持中断和触发器
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   sourceID,
		TargetID:   targetID,
		Damage:     amount,
		DamageType: damageType,
		Stage:      0,
		Card: &model.Card{
			Name:        "直接伤害",
			Type:        model.CardTypeMagic, // 默认为法术类型，如果是Attack通常走Combat流程
			Damage:      amount,
			Description: fmt.Sprintf("来自 %s 的伤害", damageType),
		},
	})

	// 如果在 HandleAction 流程中，Phase 会在下一次 Drive 时切换
	// 如果是立即执行的上下文，可能需要手动设 Phase?
	// 通常 InflictDamage 由 Handler 调用，Handler 返回后 Drive 会处理 PendingDamageResolution
}

func (e *GameEngine) emitBuffRemovedTrigger(sourceID, targetID string, effect model.EffectType) {
	target := e.State.Players[targetID]
	if target == nil {
		return
	}
	if sourceID == "" {
		sourceID = targetID
	}
	eventCtx := &model.EventContext{
		Type:     model.EventBuffRemoved,
		SourceID: sourceID, // 谁移除了基础效果
		TargetID: targetID, // 哪个目标身上的基础效果被移除
		BuffID:   string(effect),
	}
	ctx := e.buildContext(target, nil, model.TriggerOnBuffRemoved, eventCtx)
	e.dispatcher.OnTrigger(model.TriggerOnBuffRemoved, ctx)
}

func (e *GameEngine) RemoveFieldCard(targetID string, effect model.EffectType) bool {
	return e.RemoveFieldCardBy(targetID, effect, targetID)
}

func (e *GameEngine) RemoveFieldCardBy(targetID string, effect model.EffectType, sourceID string) bool {
	target := e.State.Players[targetID]
	if target == nil {
		return false
	}
	originalLen := len(target.Field)
	newField := make([]*model.FieldCard, 0)
	var removedCard *model.Card

	for _, fc := range target.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == effect {
			removedCard = &fc.Card
			continue
		}
		newField = append(newField, fc)
	}

	target.Field = newField
	removed := len(newField) < originalLen
	if removed && removedCard != nil {
		e.State.DiscardPile = append(e.State.DiscardPile, *removedCard)
		e.Log(fmt.Sprintf("%s 移除了场上效果牌: %s", target.Name, effect))
		e.emitBuffRemovedTrigger(sourceID, targetID, effect)
	}
	return removed
}

func (e *GameEngine) TakeFieldCard(targetID string, fieldIndex int, sourceID string) (model.Card, error) {
	target := e.State.Players[targetID]
	if target == nil {
		return model.Card{}, fmt.Errorf("目标不存在")
	}
	if fieldIndex < 0 || fieldIndex >= len(target.Field) {
		return model.Card{}, fmt.Errorf("无效的场上牌索引")
	}
	fc := target.Field[fieldIndex]
	if fc == nil {
		return model.Card{}, fmt.Errorf("场上牌不存在")
	}

	target.Field = append(target.Field[:fieldIndex], target.Field[fieldIndex+1:]...)
	e.Log(fmt.Sprintf("%s 的场上牌被收回: %s", target.Name, fc.Effect))
	if fc.Mode == model.FieldEffect {
		e.emitBuffRemovedTrigger(sourceID, targetID, fc.Effect)
	}
	return fc.Card, nil
}

func (e *GameEngine) GetCampCups(camp string) int {
	if model.Camp(camp) == model.RedCamp {
		return e.State.RedCups
	}
	return e.State.BlueCups
}

func (e *GameEngine) GetCampMorale(camp string) int {
	if model.Camp(camp) == model.RedCamp {
		return e.State.RedMorale
	}
	return e.State.BlueMorale
}

func (e *GameEngine) GetCampGems(camp string) int {
	if model.Camp(camp) == model.RedCamp {
		return e.State.RedGems
	}
	return e.State.BlueGems
}

func (e *GameEngine) GetCampCrystals(camp string) int {
	if model.Camp(camp) == model.RedCamp {
		return e.State.RedCrystals
	}
	return e.State.BlueCrystals
}
