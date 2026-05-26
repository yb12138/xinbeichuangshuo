// gameflow: 技能执行：扣费、建 Context、调 handler。

package engine

import (
	"fmt"
	"sort"
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// 技能发动流程（执行阶段）：
// 在校验通过后，按“消耗输入 -> 扣资源 -> 执行 handler -> 收尾阶段推进”顺序处理。
func (e *GameEngine) consumeSkillInputs(use *skillUseRequest) error {
	e.NotifyCardRevealed(use.player.ID, use.discardedCards, "discard")
	sort.Sort(sort.Reverse(sort.IntSlice(use.discardIndices)))
	for _, idx := range use.discardIndices {
		use.player.Hand = append(use.player.Hand[:idx], use.player.Hand[idx+1:]...)
	}

	if use.policy.ResolveDiscardPile != nil {
		e.State.DiscardPile = append(e.State.DiscardPile, use.policy.ResolveDiscardPile(use.policyContext())...)
	} else {
		e.State.DiscardPile = append(e.State.DiscardPile, use.discardedCards...)
	}

	if err := e.consumeExclusiveSkillCard(use); err != nil {
		return err
	}
	if use.skillDef.PlaceCard {
		if err := e.placeSkillFieldCard(use); err != nil {
			return err
		}
	}
	if use.consumedExclusiveCard != nil && !use.skillDef.PlaceCard {
		e.State.DiscardPile = append(e.State.DiscardPile, *use.consumedExclusiveCard)
	}
	if use.player != nil {
		e.addActionDiscard(use.player.ID, len(use.discardedCards))
		if use.consumedExclusiveCard != nil && !use.skillDef.PlaceCard {
			e.addActionDiscard(use.player.ID, 1)
		}
	}
	return nil
}

func (e *GameEngine) consumeExclusiveSkillCard(use *skillUseRequest) error {
	if use == nil || use.skillDef == nil || use.player == nil || use.player.Character == nil {
		return nil
	}
	if !use.skillDef.RequireExclusive || use.skillDef.CostDiscards > 0 || len(use.discardedCards) > 0 {
		return nil
	}
	if use.policy.ManualExclusiveCard || use.consumedExclusiveCard != nil {
		return nil
	}
	card, ok := use.player.ConsumeExclusiveCard(use.player.Character.ID, use.skillDef.Title)
	if !ok {
		return fmt.Errorf("未找到技能 [%s] 对应的专属技能卡", use.skillDef.Title)
	}
	use.consumedExclusiveCard = &card
	return nil
}

// placeSkillFieldCard 将卡牌放置到目标玩家场上作为场牌
// 用于五系封印、同生共死等需要放置场牌的技能
func (e *GameEngine) placeSkillFieldCard(use *skillUseRequest) error {
	var placedCard model.Card
	placedCardReady := false
	if len(use.discardedCards) > 0 {
		placedCard = use.discardedCards[0]
		placedCardReady = true
	} else if use.consumedExclusiveCard != nil {
		placedCard = *use.consumedExclusiveCard
		placedCardReady = true
	}
	if !placedCardReady {
		return fmt.Errorf("需要专属技能卡才能放置场上牌")
	}
	if len(use.actualTargets) == 0 {
		return fmt.Errorf("放置场上牌需要指定目标")
	}

	fieldTarget := use.actualTargets[0]
	fc := &model.FieldCard{
		Card:     placedCard,
		OwnerID:  fieldTarget.ID, // 场牌在谁面前
		SourceID: use.player.ID,  // 谁放置的这个场牌（用于伤害来源）
		Mode:     use.skillDef.PlaceMode,
		Effect:   use.skillDef.PlaceEffect,
		Hook:     use.skillDef.PlaceHook,
		Meta:     buildFieldCardMeta(use, placedCard), // 【关键】封印会在这里记录绑定元素
	}
	fieldTarget.AddFieldCard(fc)
	e.Log(fmt.Sprintf(
		"[Skill] %s 在 %s 面前放置了场上牌: %s (效果: %s, 触发: %s)",
		use.player.Name, fieldTarget.Name, placedCard.Name, fc.Effect, fc.Hook,
	))
	return nil
}

// buildFieldCardMeta 构建场牌的Meta信息
// 对于五系封印，会在Meta中记录绑定的元素（FieldMetaBoundElement）
func buildFieldCardMeta(use *skillUseRequest, placedCard model.Card) map[string]string {
	if use == nil {
		return nil
	}
	meta := make(map[string]string)
	// 如果是五系封印，记录绑定的元素
	if model.IsElementalSealEffect(use.skillDef.PlaceEffect) {
		// 绑定元素优先使用被放置卡牌的元素，没有则使用技能要求的弃牌元素
		boundElement := placedCard.Element
		if boundElement == "" {
			boundElement = use.skillDef.DiscardElement
		}
		meta[model.FieldMetaBoundElement] = string(boundElement)
	}
	if len(meta) == 0 {
		return nil
	}
	return meta
}

func (e *GameEngine) consumeSkillEnergyCost(use *skillUseRequest) error {
	if consumeSkillEnergyCost(use.player, use.skillDef.CostGem, use.skillDef.CostCrystal) {
		e.recordActionResourceDelta()
		return nil
	}
	return fmt.Errorf(
		"资源扣除失败: 需要 宝石%d/水晶%d，当前 宝石%d/水晶%d",
		use.skillDef.CostGem, use.skillDef.CostCrystal, use.player.Gem, use.player.Crystal,
	)
}

func (e *GameEngine) executeSkillFlow(use *skillUseRequest) error {
	if use.policy.AfterConsume != nil {
		handled, err := use.policy.AfterConsume(enginePolicyHost{e: e}, use.policyContext())
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	}

	handler := skills.GetHandler(use.skillID)
	if handler == nil {
		return fmt.Errorf("skill handler not found for %s", use.skillID)
	}

	ctx := e.BuildContext(use.player, use.target, model.TimingActionDuring, nil)
	ctx.Targets = use.actualTargets
	if ctx.Selections == nil {
		ctx.Selections = map[string]interface{}{}
	}
	ctx.Selections["discardedCards"] = use.discardedCards

	beforePoses := e.SnapshotPlayerPoses()
	beforePendingActions := snapshotPendingActions(use.player)
	if err := handler.Execute(ctx); err != nil {
		return fmt.Errorf("skill execution failed: %v", err)
	}
	e.recordActionResourceDelta()
	e.DispatchOrientationChanges(beforePoses)
	e.publishPendingActionDiff(use.player, beforePendingActions, use.skillDef.Title)
	if use.policy.AfterExecute != nil {
		if err := use.policy.AfterExecute(enginePolicyHost{e: e}, use.policyContext()); err != nil {
			return err
		}
	}
	return nil
}

func (e *GameEngine) finishSkillUse(use *skillUseRequest) error {
	if use.skillDef.PlaceCard && use.skillDef.PlaceMode == model.FieldEffect && len(use.actualTargets) > 0 {
		e.emitBuffAddedDispatch(use.player.ID, use.actualTargets[0].ID, use.skillDef.PlaceEffect)
	}
	e.runTimingActionEndSkillPost(use)
	e.recordSkillUsage(use.player.ID, use.skillDef.Title, use.skillDef.Type)
	e.NotifySkillActivated(use.player.ID, use.skillDef.ID, use.skillDef.Title, use.skillDef.Description, use.resolvedTargetIDs())
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: %s (%s)", use.player.Name, use.skillDef.Title, use.skillDef.Description))

	if use.skillDef.Type == model.SkillTypeAction && !use.policy.SkipAutoPhaseEnd {
		use.player.TurnState.HasActed = true
		use.player.TurnState.LastActionType = string(model.ActionMagic)
		use.player.TurnState.LastActionCard = nil
	}
	return nil
}
