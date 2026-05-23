// gameflow: 行动提交后的校验与写入队列（与 TurnStage 衔接）。

package engine

import (
	"errors"
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

// HandleActionSelection 处理行动选择阶段的行动
func (e *GameEngine) HandleActionSelection(act model.PlayerAction) error {
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

	handlers := map[model.PlayerActionType]func() error{
		model.CmdBuy: func() error {
			return e.HandleActionSelectionSpecialOrSkill(act, currentPid, player)
		},
		model.CmdSynthesize: func() error {
			return e.HandleActionSelectionSpecialOrSkill(act, currentPid, player)
		},
		model.CmdExtract: func() error {
			return e.HandleActionSelectionSpecialOrSkill(act, currentPid, player)
		},
		model.CmdSkill: func() error {
			return e.HandleActionSelectionSpecialOrSkill(act, currentPid, player)
		},
		model.CmdAttack: func() error {
			return e.HandleActionSelectionAttackOrMagic(act, currentPid, player, validationResult)
		},
		model.CmdMagic: func() error {
			return e.HandleActionSelectionAttackOrMagic(act, currentPid, player, validationResult)
		},
		model.CmdCannotAct: func() error {
			return e.HandleActionSelectionCannotAct(player)
		},
	}
	handler, ok := handlers[act.Type]
	if !ok {
		return fmt.Errorf("无效的行动类型: %s", act.Type)
	}
	return handler()
}

func (e *GameEngine) HandleActionSelectionSpecialOrSkill(act model.PlayerAction, currentPid string, player *model.Player) error {
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
		e.BeginActionSummary("skill", player.ID, skillTitle, targets)
		if e.State.PendingInterrupt != nil {
			return nil
		}
		e.enterActionEndStage()
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
		e.BeginActionSummary("special", player.ID, specialName, nil)
	}
	if err := e.executeSpecialActionWithRuntime(player, actionType); err != nil {
		return err
	}
	e.runPostSpecialActionRuntime(player, actionType)
	player.TurnState.LastActionType = string(actionType)
	player.TurnState.LastActionCard = nil
	e.enterActionEndStage()
	return nil
}

func (e *GameEngine) HandleActionSelectionAttackOrMagic(act model.PlayerAction, currentPid string, player *model.Player, validationResult actionSelectionValidationResult) error {
	card, ok := e.cardForPlayerAction(player, act)
	if !ok {
		return fmt.Errorf("无效的卡牌ID")
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

	if act.TargetID != "" && e.State.Players[act.TargetID] == nil {
		return fmt.Errorf("目标玩家 [%s] 不存在，请检查 ID", act.TargetID)
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
			if target.Character != nil {
				if entry := roleRegistry.Entry(target.Character.ID); entry.TargetFilter != nil && entry.TargetFilter.CannotBeActiveAttackTarget(target) {
					return fmt.Errorf("目标处于潜行状态，不能成为主动攻击目标")
				}
			}
		}
	}

	actionType := model.ActionMagic
	if act.Type == model.CmdAttack {
		actionType = model.ActionAttack
	}
	queuedAction := model.QueuedAction{
		SourceID:    currentPid,
		TargetID:    act.TargetID,
		TargetIDs:   act.TargetIDs,
		Type:        actionType,
		Element:     card.Element,
		Card:        &card,
		CardID:      card.ID,
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
		e.BeginActionSummary("attack", player.ID, card.Name, targets)
	} else {
		e.BeginActionSummary("magic", player.ID, card.Name, targets)
	}

	if actionType == model.ActionAttack && validationResult.afterAttackAccepted != nil {
		if err := validationResult.afterAttackAccepted(e, player, act); err != nil {
			return err
		}
	}

	e.State.ActionQueue = append(e.State.ActionQueue, queuedAction)
	e.enterActionExecutionStage()
	return nil
}

func (e *GameEngine) HandleActionSelectionCannotAct(player *model.Player) error {
	// === 额外行动阶段：直接跳过 ===
	if player.TurnState.CurrentExtraAction != "" {
		if e.checkExtraActionCards(player, player.TurnState.CurrentExtraAction, player.TurnState.CurrentExtraElement) {
			return errors.New("当前额外行动仍有可执行动作，不能跳过")
		}
		e.skipExtraAction(player)
		return nil
	}

	// === 主行动回合：无法行动判断 ===
	canCannotAct, _ := e.checkPlayerCannotAct(player)
	if !canCannotAct {
		return errors.New("你还有可用的攻击/法术牌或技能，无法宣告无法行动")
	}

	// === 执行无法行动流程 ===
	e.executeCannotActFlow(player)
	return nil
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
		if requiredType == model.ExtraActionAny {
			if act.Type == model.CmdAttack || act.Type == model.CmdMagic || act.Type == model.CmdSkill {
				isMatch = true
			}
		} else if requiredType == "Attack" {
			if act.Type == model.CmdAttack {
				isMatch = true
			}
		} else if requiredType == "Magic" {
			if act.Type == model.CmdMagic || act.Type == model.CmdSkill {
				isMatch = true
			}
		}

		if !isMatch {
			if requiredType == model.ExtraActionAny {
				return fmt.Errorf("当前额外行动只能选择攻击或法术")
			}
			if requiredType == "Attack" && act.Type == model.CmdSkill {
				return fmt.Errorf("当前额外行动必须是 [Attack]，不能使用技能")
			}
			return fmt.Errorf("当前额外行动必须是 [%s]", requiredType)
		}
	}

	if len(p.TurnState.CurrentExtraElement) > 0 && (act.Type == model.CmdAttack || act.Type == model.CmdMagic) {
		if card, ok := e.cardForPlayerAction(p, act); ok {
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

func (e *GameEngine) cardForPlayerAction(p *model.Player, act model.PlayerAction) (model.Card, bool) {
	if act.CardID == "" {
		return model.Card{}, false
	}
	card, _, _, ok := e.getPlayableCardByID(p, act.CardID)
	return card, ok
}

// checkExtraActionCards 检查玩家是否有符合额外行动约束的牌
func (e *GameEngine) checkExtraActionCards(p *model.Player, mustType string, mustElement []model.Element) bool {
	if mustType == model.ExtraActionAny {
		return e.checkExtraActionCards(p, string(model.ActionAttack), mustElement) ||
			e.checkExtraActionCards(p, string(model.ActionMagic), mustElement)
	}
	total := e.playableCardCount(p)
	for idx := 0; idx < total; idx++ {
		card, _, _, ok := e.getPlayableCardByIndex(p, idx)
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
		cardID := qa.CardID
		if cardID == "" {
			cardID = qa.Card.ID
		}
		if card, _, _, ok := e.getPlayableCardByID(player, cardID); ok && card.Type == requiredType {
			if requiredType == model.CardTypeAttack {
				card = e.transformAttackCard(player, card)
			}
			qa.CardID = card.ID
			qa.Element = card.Element
			cardCopy := card
			qa.Card = &cardCopy
			return true
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
		if mustType == model.ExtraActionAny {
			constraintInfo += "[攻击/法术行动]"
		} else {
			constraintInfo += fmt.Sprintf("[%s行动]", mustType)
		}
	}
	return constraintInfo
}
