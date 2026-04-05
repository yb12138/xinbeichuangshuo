package engine

import (
	"errors"
	"fmt"
	"strings"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// handleActionSelection 处理行动选择阶段的行动
func (e *GameEngine) handleActionSelection(act model.PlayerAction) error {
	currentPid := e.State.PlayerOrder[e.State.CurrentTurn]
	player := e.State.Players[currentPid]

	if act.PlayerID != currentPid {
		return fmt.Errorf("不是你的回合")
	}

	if err := e.validateExtraActionConstraint(player, act); err != nil {
		return err
	}

	validationResult, err := e.validateActionSelectionPolicies(player, act)
	if err != nil {
		return err
	}
	if validationResult.handled {
		return nil
	}

	switch act.Type {
	case model.CmdBuy, model.CmdSynthesize, model.CmdExtract, model.CmdSkill:
		if player.TurnState.HasStartupSkillOrSpecialActionsLocked() &&
			(act.Type == model.CmdBuy || act.Type == model.CmdSynthesize || act.Type == model.CmdExtract) {
			return fmt.Errorf("你本回合已执行启动技能，不能执行特殊行动")
		}

		var actionType model.ActionType
		switch act.Type {
		case model.CmdBuy:
			actionType = model.ActionBuy
		case model.CmdSynthesize:
			actionType = model.ActionSynthesize
		case model.CmdExtract:
			actionType = model.ActionExtract
		case model.CmdSkill:
			if act.SkillID == "" {
				return fmt.Errorf("未指定技能ID")
			}

			targetIDs := append([]string{}, act.TargetIDs...)
			if len(targetIDs) == 0 && act.TargetID != "" {
				targetIDs = append(targetIDs, act.TargetID)
			}
			if err := e.UseSkill(act.PlayerID, act.SkillID, targetIDs, act.Selections); err != nil {
				return fmt.Errorf("技能发动失败: %v", err)
			}
			if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 && act.SkillID == "arbiter_doomsday" {
				player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
				player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_done_turn"] = 1
				consumeHeroTauntRestriction(e, player)
			}
			skillTitle := act.SkillID
			if player.Character != nil {
				for _, s := range player.Character.Skills {
					if s.ID == act.SkillID {
						skillTitle = s.Title
						break
					}
				}
			}
			targets := []string{}
			if act.TargetID != "" {
				targets = append(targets, act.TargetID)
			}
			if len(act.TargetIDs) > 0 {
				targets = append(targets, act.TargetIDs...)
			}
			e.beginActionSummary("skill", player.ID, skillTitle, targets)
			if e.State.PendingInterrupt != nil {
				return nil
			}
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
			return nil
		}

		specialName := ""
		switch actionType {
		case model.ActionBuy:
			specialName = "购买"
		case model.ActionSynthesize:
			specialName = "合成"
		case model.ActionExtract:
			specialName = "提炼"
		}
		if specialName != "" {
			e.beginActionSummary("special", player.ID, specialName, nil)
		}
		if err := e.executeSpecialActionWithRuntime(player, actionType); err != nil {
			return err
		}
		player.TurnState.UsedSkillCounts["hb_special"] = 1
		e.runPostSpecialActionRuntime(player, actionType)
		player.TurnState.LastActionType = string(actionType)

		phaseEventCtx := &model.EventContext{
			Type:       model.EventPhaseEnd,
			SourceID:   player.ID,
			ActionType: actionType,
		}
		phaseCtx := e.buildContext(player, nil, model.TriggerOnPhaseEnd, phaseEventCtx)
		e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, phaseCtx)
		// 特殊行动的 OnPhaseEnd 已在此处派发，不再通过 ActionEnd 阶段重复派发。
		player.TurnState.LastActionType = ""

		if e.State.PendingInterrupt == nil {
			e.enterExtraActionStage()
		}
		return nil

	case model.CmdAttack, model.CmdMagic:
		if act.CardIndex < 0 {
			return fmt.Errorf("需要指定卡牌索引")
		}

		card, _, _, ok := getPlayableCardByIndex(player, act.CardIndex)
		if !ok {
			return fmt.Errorf("无效的卡牌索引")
		}

		if act.Type == model.CmdAttack && card.Type != model.CardTypeAttack {
			return fmt.Errorf("只能使用攻击牌进行攻击")
		}
		if act.Type == model.CmdMagic && card.Type != model.CardTypeMagic {
			return fmt.Errorf("只能使用法术牌进行法术")
		}
		if act.Type == model.CmdMagic && !e.canCastMagicInAction(player) {
			return fmt.Errorf("当前形态不能在行动阶段使用法术牌")
		}

		needTarget := act.Type == model.CmdAttack || (act.Type == model.CmdMagic && card.Name != "魔弹")
		if needTarget && act.TargetID == "" && len(act.TargetIDs) == 0 {
			if act.Type == model.CmdAttack {
				return fmt.Errorf("攻击需要指定目标")
			}
			return fmt.Errorf("该法术需要指定目标")
		}

		if act.TargetID != "" {
			if e.State.Players[act.TargetID] == nil {
				return fmt.Errorf("目标玩家 [%s] 不存在，请检查 ID", act.TargetID)
			}
		}
		if len(act.TargetIDs) > 0 {
			for _, tid := range act.TargetIDs {
				if e.State.Players[tid] == nil {
					return fmt.Errorf("目标玩家 [%s] 不存在，请检查 ID", tid)
				}
			}
		}
		if act.Type == model.CmdAttack {
			attackTargetID := act.TargetID
			if attackTargetID == "" && len(act.TargetIDs) > 0 {
				attackTargetID = act.TargetIDs[0]
			}
			if attackTargetID != "" {
				target := e.State.Players[attackTargetID]
				if target == nil {
					return fmt.Errorf("目标玩家 [%s] 不存在，请检查 ID", attackTargetID)
				}
				if target.Camp == player.Camp {
					return fmt.Errorf("攻击目标必须是敌方角色")
				}
				if hasAssassinStealthForm(target) {
					return fmt.Errorf("目标处于潜行状态，不能成为主动攻击目标")
				}
			}
		}

		var actionType model.ActionType
		if act.Type == model.CmdAttack {
			actionType = model.ActionAttack
		} else if act.Type == model.CmdMagic {
			actionType = model.ActionMagic
		} else {
			return fmt.Errorf("无效的行动类型")
		}

		queuedAction := model.QueuedAction{
			SourceID:    currentPid,
			TargetID:    act.TargetID,
			TargetIDs:   act.TargetIDs,
			Type:        actionType,
			Element:     card.Element,
			Card:        &card,
			CardIndex:   act.CardIndex,
			SourceSkill: "",
		}
		if actionType == model.ActionAttack {
			card = e.transformAttackCard(player, card)
			queuedAction.Element = card.Element
			queuedAction.Card = &card
		}
		targets := []string{}
		if act.TargetID != "" {
			targets = append(targets, act.TargetID)
		}
		if len(act.TargetIDs) > 0 {
			targets = append(targets, act.TargetIDs...)
		}
		if actionType == model.ActionAttack {
			e.beginActionSummary("attack", player.ID, card.Name, targets)
		} else {
			e.beginActionSummary("magic", player.ID, card.Name, targets)
		}

		e.State.ActionQueue = append(e.State.ActionQueue, queuedAction)
		if validationResult.consumeHeroTauntOnAttack {
			consumeHeroTauntRestriction(e, player)
		}

		e.enterActionExecutionStage()
		return nil

	case model.CmdCannotAct:
		if player.TurnState.CurrentExtraAction != "" {
			if e.checkExtraActionCards(player, player.TurnState.CurrentExtraAction, player.TurnState.CurrentExtraElement) {
				return errors.New("当前额外行动仍有可执行动作，不能跳过")
			}
			constraintInfo := e.buildConstraintInfo(player.TurnState.CurrentExtraAction, player.TurnState.CurrentExtraElement)
			e.beginActionSummary("cannot_act", player.ID, "跳过额外行动", nil)
			e.Log(fmt.Sprintf("[Turn] %s 宣告【无法行动】，跳过本次额外行动%s", player.Name, constraintInfo))
			player.TurnState.CurrentExtraAction = ""
			player.TurnState.CurrentExtraElement = nil
			e.enterTurnEndStage()
			return nil
		}

		e.beginActionSummary("cannot_act", player.ID, "无法行动", nil)
		handCount := len(player.Hand)
		if handCount == 0 {
			e.Log(fmt.Sprintf("[Action] %s 宣告【无法行动】（无手牌），结束本回合行动阶段", player.Name))
			player.TurnState.LockSpecialActionsForRemainderOfTurn()
			e.enterTurnEndStage()
			return nil
		}
		canUseMagic := e.canCastMagicInAction(player)
		for idx := 0; idx < playableCardCount(player); idx++ {
			c, _, _, ok := getPlayableCardByIndex(player, idx)
			if !ok {
				continue
			}
			if c.Type == model.CardTypeAttack || (c.Type == model.CardTypeMagic && canUseMagic) {
				return errors.New("你还有可用的攻击/法术牌，无法宣告无法行动")
			}
		}
		e.Log(fmt.Sprintf("[Action] %s 宣告【无法行动】，展示并弃掉全部手牌(%d张)", player.Name, handCount))
		e.NotifyCardRevealed(player.ID, append([]model.Card{}, player.Hand...), "discard")
		for _, c := range player.Hand {
			e.State.DiscardPile = append(e.State.DiscardPile, c)
		}
		player.Hand = player.Hand[:0]
		cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, handCount)
		e.State.Deck = newDeck
		e.State.DiscardPile = newDiscard
		player.Hand = append(player.Hand, cards...)
		e.NotifyDrawCards(player.ID, handCount, "cannot_act_redraw")
		if e.isMagicSwordsman(player) {
			for len(player.Hand) > 0 {
				hasAttack := false
				allMagic := true
				for _, c := range player.Hand {
					if c.Type == model.CardTypeAttack {
						hasAttack = true
						break
					}
					if c.Type != model.CardTypeMagic {
						allMagic = false
					}
				}
				if hasAttack || !allMagic {
					break
				}
				redrawCount := len(player.Hand)
				e.NotifyCardRevealed(player.ID, append([]model.Card{}, player.Hand...), "discard")
				e.State.DiscardPile = append(e.State.DiscardPile, player.Hand...)
				player.Hand = player.Hand[:0]
				nextCards, deck2, discard2 := rules.DrawCards(e.State.Deck, e.State.DiscardPile, redrawCount)
				e.State.Deck = deck2
				e.State.DiscardPile = discard2
				player.Hand = append(player.Hand, nextCards...)
				e.NotifyDrawCards(player.ID, redrawCount, "magic_swordsman_redraw")
				e.Log(fmt.Sprintf("[Action] %s 触发魔剑士重摸：全法术手牌已弃置并重摸%d张", player.Name, redrawCount))
			}
		}
		e.Log(fmt.Sprintf("[Action] %s 重新摸了%d张牌，且本回合不可执行特殊行动", player.Name, handCount))
		player.TurnState.LockSpecialActionsForRemainderOfTurn()
		e.enterActionExecutionStage()
		return nil

	default:
		return fmt.Errorf("无效的行动类型: %s", act.Type)
	}
}

func (e *GameEngine) validateExtraActionConstraint(p *model.Player, act model.PlayerAction) error {
	if p.TurnState.CurrentExtraAction != "" {
		requiredType := p.TurnState.CurrentExtraAction

		if act.Type == model.CmdCannotAct {
			if e.checkExtraActionCards(p, requiredType, p.TurnState.CurrentExtraElement) {
				return fmt.Errorf("当前额外行动仍有可执行动作，不能跳过")
			}
			return nil
		}

		isMatch := false
		if requiredType == "Attack" {
			if act.Type == model.CmdAttack {
				isMatch = true
			}
		} else if requiredType == "Magic" {
			if act.Type == model.CmdMagic || act.Type == model.CmdSkill {
				isMatch = true
			}
		}

		if !isMatch {
			if requiredType == "Attack" && act.Type == model.CmdSkill {
				return fmt.Errorf("当前额外行动必须是 [Attack]，不能使用技能")
			}
			return fmt.Errorf("当前额外行动必须是 [%s]", requiredType)
		}
	}

	if len(p.TurnState.CurrentExtraElement) > 0 && (act.Type == model.CmdAttack || act.Type == model.CmdMagic) {
		if card, _, _, ok := getPlayableCardByIndex(p, act.CardIndex); ok {
			if act.Type == model.CmdAttack {
				card = e.transformAttackCard(p, card)
			}
			isAllowed := false
			for _, allowed := range p.TurnState.CurrentExtraElement {
				if card.Element == allowed {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
				allowed := make([]string, 0, len(p.TurnState.CurrentExtraElement))
				for _, ele := range p.TurnState.CurrentExtraElement {
					if ele == "" {
						continue
					}
					allowed = append(allowed, fmt.Sprintf("%s系", elementNameForPrompt(string(ele))))
				}
				chosen := fmt.Sprintf("%s系", elementNameForPrompt(string(card.Element)))
				if len(allowed) == 0 {
					return fmt.Errorf("当前行动限制元素，你选择了 %s", chosen)
				}
				return fmt.Errorf("当前行动限制元素为 %s，你选择了 %s", strings.Join(allowed, " / "), chosen)
			}
		}
	}

	return nil
}

// checkExtraActionCards 检查玩家是否有符合额外行动约束的牌
func (e *GameEngine) checkExtraActionCards(p *model.Player, mustType string, mustElement []model.Element) bool {
	total := playableCardCount(p)
	for idx := 0; idx < total; idx++ {
		card, _, _, ok := getPlayableCardByIndex(p, idx)
		if !ok {
			continue
		}
		if mustType == "Attack" && card.Type != model.CardTypeAttack {
			continue
		}
		if mustType == "Magic" && card.Type != model.CardTypeMagic {
			continue
		}
		if mustType == "Magic" && !e.canCastMagicInAction(p) {
			continue
		}
		if mustType == "Attack" {
			card = e.transformAttackCard(p, card)
		}

		if len(mustElement) > 0 {
			elementMatch := false
			for _, elem := range mustElement {
				if card.Element == elem {
					elementMatch = true
					break
				}
			}
			if !elementMatch {
				continue
			}
		}
		return true
	}
	if mustType == "Magic" && e.hasUsableActionSkillForExtraMagic(p) {
		return true
	}
	return false
}

func (e *GameEngine) countCoverCardsByEffectAndElement(p *model.Player, effect model.EffectType, element model.Element) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != effect {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

func (e *GameEngine) repairQueuedActionCard(player *model.Player, qa *model.QueuedAction) bool {
	if player == nil || qa == nil {
		return false
	}

	requiredType := model.CardType("")
	switch qa.Type {
	case model.ActionAttack:
		requiredType = model.CardTypeAttack
	case model.ActionMagic:
		requiredType = model.CardTypeMagic
	default:
		return false
	}

	if qa.Card != nil {
		if idx := findPlayableCardIndexByID(player, qa.Card.ID); idx >= 0 {
			if card, _, _, ok := getPlayableCardByIndex(player, idx); ok && card.Type == requiredType {
				if requiredType == model.CardTypeAttack {
					card = e.transformAttackCard(player, card)
				}
				qa.CardIndex = idx
				qa.Element = card.Element
				cardCopy := card
				qa.Card = &cardCopy
				return true
			}
		}
		// 规则约束：队列中的行动卡必须与玩家最初选择的实体卡一致，不允许同类自动替代。
		return false
	}

	return false
}

// buildConstraintInfo 构建约束信息字符串
func (e *GameEngine) buildConstraintInfo(mustType string, mustElement []model.Element) string {
	constraintInfo := ""
	if len(mustElement) > 0 {
		labels := make([]string, 0, len(mustElement))
		for _, ele := range mustElement {
			if ele == "" {
				continue
			}
			labels = append(labels, fmt.Sprintf("%s系", elementNameForPrompt(string(ele))))
		}
		if len(labels) > 0 {
			constraintInfo += fmt.Sprintf("[%s]", strings.Join(labels, "/"))
		}
	}
	if mustType != "" {
		constraintInfo += fmt.Sprintf("[%s行动]", mustType)
	}
	return constraintInfo
}
