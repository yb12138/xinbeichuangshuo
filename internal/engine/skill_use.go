package engine

import (
	"fmt"
	"sort"
	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

type skillUseRequest struct {
	engine                *GameEngine
	player                *model.Player
	skillDef              *model.SkillDefinition
	policy                skillUsePolicy
	skillID               string
	targetIDs             []string
	discardIndices        []int
	requiredDiscards      int
	discardedCards        []model.Card
	consumedExclusiveCard *model.Card
	target                *model.Player
	actualTargets         []*model.Player
}

func (use *skillUseRequest) resolvedTargetIDs() []string {
	ids := make([]string, 0, len(use.actualTargets))
	for _, target := range use.actualTargets {
		if target != nil {
			ids = append(ids, target.ID)
		}
	}
	return ids
}

// UseSkill 使用技能
func (e *GameEngine) UseSkill(playerID, skillID string, targetIDs []string, discardIndices []int) error {
	use, err := e.prepareSkillUse(playerID, skillID, targetIDs, discardIndices)
	if err != nil {
		return err
	}
	if pending, err := e.maybeRequestSkillDiscardSelection(use); pending || err != nil {
		return err
	}
	if err := e.validateSkillDiscardSelection(use); err != nil {
		return err
	}
	if err := e.validateSkillActivation(use); err != nil {
		return err
	}
	if err := e.consumeSkillCoverCost(use); err != nil {
		return err
	}
	if err := e.ensureSkillEnergyCost(use); err != nil {
		return err
	}
	if err := e.resolveSkillTargets(use); err != nil {
		return err
	}
	if err := e.validateSkillFieldPlacement(use); err != nil {
		return err
	}
	if err := e.consumeSkillInputs(use); err != nil {
		return err
	}
	if err := e.consumeSkillEnergyCost(use); err != nil {
		return err
	}

	use.player.TurnState.UsedSkillCounts[skillID]++

	if err := e.executeSkillFlow(use); err != nil {
		return err
	}
	return e.finishSkillUse(use)
}

func (e *GameEngine) prepareSkillUse(playerID, skillID string, targetIDs []string, discardIndices []int) (*skillUseRequest, error) {
	player := e.State.Players[playerID]
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}
	if !player.IsActive {
		return nil, fmt.Errorf("not your turn")
	}
	if player.Character == nil {
		return nil, fmt.Errorf("no character assigned")
	}

	skillDef := findCharacterSkill(player.Character, skillID)
	if skillDef == nil {
		return nil, fmt.Errorf("skill %s not found for character %s", skillID, player.Character.ID)
	}

	policy := resolveSkillUsePolicy(skillID)
	requiredDiscards := skillDef.CostDiscards
	if policy.resolveDiscardCount != nil {
		requiredDiscards = policy.resolveDiscardCount(player, skillDef)
	}

	return &skillUseRequest{
		engine:           e,
		player:           player,
		skillDef:         skillDef,
		policy:           policy,
		skillID:          skillID,
		targetIDs:        append([]string{}, targetIDs...),
		discardIndices:   append([]int{}, discardIndices...),
		requiredDiscards: requiredDiscards,
	}, nil
}

func findCharacterSkill(character *model.Character, skillID string) *model.SkillDefinition {
	if character == nil {
		return nil
	}
	for i := range character.Skills {
		if character.Skills[i].ID == skillID {
			return &character.Skills[i]
		}
	}
	return nil
}

func (e *GameEngine) maybeRequestSkillDiscardSelection(use *skillUseRequest) (bool, error) {
	if use.requiredDiscards <= 0 || len(use.discardIndices) > 0 {
		return false, nil
	}
	if len(use.player.Hand) < use.requiredDiscards {
		return false, fmt.Errorf("手牌不足：发动 [%s] 需要弃置 %d 张牌", use.skillDef.Title, use.requiredDiscards)
	}

	e.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptDiscard,
		PlayerID: use.player.ID,
		SkillIDs: []string{use.skillID},
		Context: map[string]interface{}{
			"discard_count": use.requiredDiscards,
			"skill_id":      use.skillID,
			"target_ids":    use.targetIDs,
			"resume_phase":  e.currentChoiceResumePoint(),
		},
	}
	e.enterDiscardSelection()
	e.Log(fmt.Sprintf("%s 请选择用于发动 [%s] 的卡牌", use.player.Name, use.skillDef.Title))
	return true, nil
}

func (e *GameEngine) validateSkillDiscardSelection(use *skillUseRequest) error {
	seen := map[int]bool{}
	for _, idx := range use.discardIndices {
		if seen[idx] {
			return fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}

	if use.requiredDiscards > 0 && len(use.discardIndices) != use.requiredDiscards {
		return fmt.Errorf("技能需要弃 %d 张牌，你选择了 %d 张", use.requiredDiscards, len(use.discardIndices))
	}

	discardedCards := make([]model.Card, 0, len(use.discardIndices))
	for _, idx := range use.discardIndices {
		if idx < 0 || idx >= len(use.player.Hand) {
			return fmt.Errorf("弃牌索引越界: %d", idx)
		}

		card := use.player.Hand[idx]
		effectiveElement := card.Element
		if use.skillDef.DiscardElement != "" {
			effectiveElement = e.blazeWitchAttackElement(use.player, card)
		}
		if use.skillDef.DiscardElement != "" && effectiveElement != use.skillDef.DiscardElement {
			return fmt.Errorf("弃牌 %s 不符合元素要求", card.Name)
		}
		if use.skillDef.DiscardType != "" && card.Type != use.skillDef.DiscardType {
			return fmt.Errorf("弃牌 %s 不符合卡牌类型要求", card.Name)
		}
		if use.skillDef.DiscardFate != "" && card.Faction != use.skillDef.DiscardFate {
			return fmt.Errorf("弃牌 %s 不符合命格要求", card.Name)
		}
		if use.skillDef.RequireExclusive && !card.MatchExclusive(use.player.Character.ID, use.skillDef.Title) {
			return fmt.Errorf("弃牌 %s 不是该技能对应的独有牌", card.Name)
		}
		discardedCards = append(discardedCards, card)
	}
	use.discardedCards = discardedCards

	if use.policy.validateDiscardedCards != nil {
		if err := use.policy.validateDiscardedCards(use); err != nil {
			return err
		}
	}

	if use.skillDef.RequireExclusive && use.skillDef.CostDiscards <= 0 && len(use.discardedCards) == 0 {
		if use.player.Character == nil || use.player.Character.ID == "" {
			return fmt.Errorf("角色信息缺失，无法校验独有牌")
		}
		if use.policy.manualExclusiveCard {
			if !use.player.HasExclusiveCard(use.player.Character.ID, use.skillDef.Title) {
				return fmt.Errorf("未找到技能 [%s] 对应的专属技能卡", use.skillDef.Title)
			}
		} else {
			card, ok := use.player.ConsumeExclusiveCard(use.player.Character.ID, use.skillDef.Title)
			if !ok {
				return fmt.Errorf("未找到技能 [%s] 对应的专属技能卡", use.skillDef.Title)
			}
			use.consumedExclusiveCard = &card
		}
	}

	return nil
}

func (e *GameEngine) validateSkillActivation(use *skillUseRequest) error {
	switch use.skillDef.Type {
	case model.SkillTypeStartup:
		if !e.isStartupWindow() {
			return fmt.Errorf("startup skills can only be used during trigger phase")
		}
	case model.SkillTypeAction:
		if !e.isActionSelectionWindow() {
			return fmt.Errorf("action skills can only be used during action phase")
		}
	case model.SkillTypeResponse:
		return fmt.Errorf("response skills are triggered automatically")
	case model.SkillTypePassive:
		return fmt.Errorf("passive skills are triggered automatically")
	}

	if use.player.TurnState.CurrentExtraAction == "Attack" {
		return fmt.Errorf("当前是额外攻击行动，不能使用技能，只能发起攻击")
	}
	if model.ContainsSkillTag(use.skillDef.Tags, model.TagTurnLimit) && use.player.TurnState.UsedSkillCounts[use.skillID] > 0 {
		return fmt.Errorf("skill %s can only be used once per turn", use.skillID)
	}
	return nil
}

func (e *GameEngine) consumeSkillCoverCost(use *skillUseRequest) error {
	if use.skillDef.CostCoverCards <= 0 {
		return nil
	}
	coverCards, err := use.player.ConsumeCoverCards(use.skillDef.CostCoverCards)
	if err != nil {
		return fmt.Errorf("盖牌消耗失败: %v", err)
	}
	e.State.DiscardPile = append(e.State.DiscardPile, coverCards...)
	e.Log(fmt.Sprintf("%s 消耗了 %d 张盖牌作为技能消耗", use.player.Name, use.skillDef.CostCoverCards))
	return nil
}

func (e *GameEngine) ensureSkillEnergyCost(use *skillUseRequest) error {
	if canPaySkillEnergyCost(use.player, use.skillDef.CostGem, use.skillDef.CostCrystal) {
		return nil
	}
	return fmt.Errorf(
		"资源不足: 需要 宝石%d/水晶%d，当前 宝石%d/水晶%d（红宝石可替代水晶）",
		use.skillDef.CostGem, use.skillDef.CostCrystal, use.player.Gem, use.player.Crystal,
	)
}

func (e *GameEngine) resolveSkillTargets(use *skillUseRequest) error {
	if use.skillDef.TargetType == model.TargetNone {
		if use.policy.validateTargets != nil {
			return use.policy.validateTargets(use)
		}
		return nil
	}
	if len(use.targetIDs) == 0 {
		if use.policy.allowZeroTargets {
			if use.policy.validateTargets != nil {
				return use.policy.validateTargets(use)
			}
			return nil
		}
		return fmt.Errorf("skill requires target(s)")
	}

	actualTargets := make([]*model.Player, 0, len(use.targetIDs))
	for _, id := range use.targetIDs {
		target := e.State.Players[id]
		if target == nil {
			return fmt.Errorf("target player %s not found", id)
		}
		actualTargets = append(actualTargets, target)
	}
	if len(actualTargets) == 0 {
		return fmt.Errorf("no valid targets found")
	}
	if use.skillDef.MaxTargets > 0 && len(actualTargets) > use.skillDef.MaxTargets {
		return fmt.Errorf("技能最多只能指定 %d 个目标，你指定了 %d 个", use.skillDef.MaxTargets, len(actualTargets))
	}
	if use.skillDef.MinTargets > 0 && len(actualTargets) < use.skillDef.MinTargets {
		return fmt.Errorf("技能最少需要指定 %d 个目标，你指定了 %d 个", use.skillDef.MinTargets, len(actualTargets))
	}

	for _, target := range actualTargets {
		switch use.skillDef.TargetType {
		case model.TargetSelf:
			if target.ID != use.player.ID {
				return fmt.Errorf("skill can only target self")
			}
		case model.TargetEnemy:
			if target.Camp == use.player.Camp {
				return fmt.Errorf("skill can only target enemies")
			}
		case model.TargetAlly:
			if target.Camp != use.player.Camp {
				return fmt.Errorf("skill can only target allies")
			}
		case model.TargetAllySelf:
			if target.Camp != use.player.Camp {
				return fmt.Errorf("skill can only target allies or self")
			}
		case model.TargetAny:
			// no-op
		default:
			// TargetSpecific 等由 skill policy / handler 继续校验。
		}
	}

	use.actualTargets = actualTargets
	if len(actualTargets) == 1 {
		use.target = actualTargets[0]
	}
	if use.policy.validateTargets != nil {
		return use.policy.validateTargets(use)
	}
	return nil
}

func (e *GameEngine) validateSkillFieldPlacement(use *skillUseRequest) error {
	if !use.skillDef.PlaceCard || len(use.actualTargets) == 0 {
		return nil
	}
	fieldTarget := use.actualTargets[0]
	if use.skillDef.PlaceMode == model.FieldEffect && model.IsBasicEffect(string(use.skillDef.PlaceEffect)) {
		for _, fc := range fieldTarget.Field {
			if fc == nil || fc.Mode != model.FieldEffect {
				continue
			}
			if fc.Effect == use.skillDef.PlaceEffect {
				return fmt.Errorf("%s 面前已有同种基础效果，不可重复放置", fieldTarget.Name)
			}
		}
	}
	return nil
}

func (e *GameEngine) consumeSkillInputs(use *skillUseRequest) error {
	e.NotifyCardRevealed(use.player.ID, use.discardedCards, "discard")
	sort.Sort(sort.Reverse(sort.IntSlice(use.discardIndices)))
	for _, idx := range use.discardIndices {
		use.player.Hand = append(use.player.Hand[:idx], use.player.Hand[idx+1:]...)
	}

	if use.policy.appendDiscardPile != nil {
		use.policy.appendDiscardPile(e, use)
	} else {
		appendDiscardedCardsToPile(e, use)
	}

	if use.skillDef.PlaceCard {
		if err := e.placeSkillFieldCard(use); err != nil {
			return err
		}
	}
	if use.consumedExclusiveCard != nil && !use.skillDef.PlaceCard {
		e.State.DiscardPile = append(e.State.DiscardPile, *use.consumedExclusiveCard)
	}
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
		Trigger:  use.skillDef.PlaceTrigger,
		Meta:     buildFieldCardMeta(use, placedCard), // 【关键】封印会在这里记录绑定元素
	}
	fieldTarget.AddFieldCard(fc)
	e.Log(fmt.Sprintf(
		"[Skill] %s 在 %s 面前放置了场上牌: %s (效果: %s, 触发: %s)",
		use.player.Name, fieldTarget.Name, placedCard.Name, fc.Effect, fc.Trigger,
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
		return nil
	}
	return fmt.Errorf(
		"资源扣除失败: 需要 宝石%d/水晶%d，当前 宝石%d/水晶%d",
		use.skillDef.CostGem, use.skillDef.CostCrystal, use.player.Gem, use.player.Crystal,
	)
}

func (e *GameEngine) executeSkillFlow(use *skillUseRequest) error {
	if use.policy.afterConsume != nil {
		handled, err := use.policy.afterConsume(e, use)
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

	ctx := e.buildContext(use.player, use.target, model.TriggerNone, nil)
	ctx.Targets = use.actualTargets
	if ctx.Selections == nil {
		ctx.Selections = map[string]interface{}{}
	}
	ctx.Selections["discardedCards"] = use.discardedCards

	beforePoses := e.snapshotPlayerPoses()
	if err := handler.Execute(ctx); err != nil {
		return fmt.Errorf("skill execution failed: %v", err)
	}
	e.dispatchOrientationChanges(beforePoses)
	return nil
}

func (e *GameEngine) finishSkillUse(use *skillUseRequest) error {
	if use.skillDef.PlaceCard && use.skillDef.PlaceMode == model.FieldEffect && len(use.actualTargets) > 0 {
		e.emitBuffAddedTrigger(use.player.ID, use.actualTargets[0].ID, use.skillDef.PlaceEffect)
	}
	e.recordSkillUsage(use.player.ID, use.skillDef.Title, use.skillDef.Type)
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: %s (%s)", use.player.Name, use.skillDef.Title, use.skillDef.Description))

	if use.skillDef.Type == model.SkillTypeAction && !use.policy.skipAutoPhaseEnd {
		use.player.TurnState.HasActed = true
		phaseEventCtx := &model.EventContext{
			Type:       model.EventPhaseEnd,
			SourceID:   use.player.ID,
			ActionType: model.ActionMagic,
		}
		phaseCtx := e.buildContext(use.player, nil, model.TriggerOnPhaseEnd, phaseEventCtx)
		e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, phaseCtx)
	}
	return nil
}
