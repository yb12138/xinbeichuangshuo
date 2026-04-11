// gameflow: 元素师技能流。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildElementalistChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "elementalist_bonus_confirm":
		skillName, _ := data["skill_display_name"].(string)
		eleZh := elementNameForPrompt(fmt.Sprint(data["bonus_element"]))
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【%s】是否额外弃1张%s系牌，使本次法术伤害+1？", skillName, eleZh),
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}

	case "elementalist_bonus_card":
		if player == nil {
			return nil
		}
		matching := parseIntSliceContextValue(data["matching_indices"])
		options := make([]model.PromptOption, 0, len(matching))
		for _, idx := range matching {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "请选择额外弃置的同系牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}

	return nil
}

func (e *GameEngine) handleElementalistChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"elementalist_bonus_confirm": e.handleElementalistBonusConfirmChoice,
		"elementalist_bonus_card":    e.handleElementalistBonusCardChoice,
	})
}

func (e *GameEngine) handleElementalistBonusConfirmChoice(selectionIndex int, ctxData map[string]interface{}) error {
	if selectionIndex == 1 {
		return e.resolveElementalistBonus(ctxData, false, -1)
	}
	if selectionIndex != 0 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	bonusElement, _ := ctxData["bonus_element"].(string)
	matching := make([]int, 0)
	for i, card := range user.Hand {
		if string(card.Element) == bonusElement {
			matching = append(matching, i)
		}
	}
	if len(matching) == 0 {
		return e.resolveElementalistBonus(ctxData, false, -1)
	}
	ctxData["choice_type"] = "elementalist_bonus_card"
	ctxData["matching_indices"] = matching
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handleElementalistBonusCardChoice(selectionIndex int, ctxData map[string]interface{}) error {
	matching := parseIntSliceContextValue(ctxData["matching_indices"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, matching)
	if !ok {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	return e.resolveElementalistBonus(ctxData, true, cardIdx)
}

func (e *GameEngine) resolveElementalistBonus(ctxData map[string]interface{}, bonus bool, discardIdx int) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetID, _ := ctxData["damage_target_id"].(string)
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}

	baseDamage := 0
	if value, ok := ctxData["base_damage"].(int); ok {
		baseDamage = value
	} else if value, ok := ctxData["base_damage"].(float64); ok {
		baseDamage = int(value)
	}
	skillName, _ := ctxData["skill_display_name"].(string)
	bonusElement, _ := ctxData["bonus_element"].(string)

	damage := baseDamage
	if bonus {
		if discardIdx < 0 || discardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的弃牌索引")
		}
		card := user.Hand[discardIdx]
		if string(card.Element) != bonusElement {
			return fmt.Errorf("弃牌元素不匹配")
		}
		e.NotifyCardRevealed(userID, []model.Card{card}, "discard")
		user.Hand = append(user.Hand[:discardIdx], user.Hand[discardIdx+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		damage++
	}

	e.InflictDamage(userID, targetID, damage, model.MagicAttack)

	if healTargetID, ok := ctxData["heal_target_id"].(string); ok && healTargetID != "" {
		if healTarget := e.State.Players[healTargetID]; healTarget != nil {
			e.Heal(healTargetID, 1)
			e.Log(fmt.Sprintf("[元素师] %s 为 %s 提供了1点治疗", skillName, healTarget.Name))
		}
	}

	campGemBonus := 0
	if value, ok := ctxData["camp_gem_bonus"].(int); ok {
		campGemBonus = value
	} else if value, ok := ctxData["camp_gem_bonus"].(float64); ok {
		campGemBonus = int(value)
	}
	if campGemBonus > 0 {
		e.ModifyGem(string(user.Camp), campGemBonus)
	}

	grantAttack := false
	if value, ok := ctxData["grant_attack"].(bool); ok {
		grantAttack = value
	}
	grantMagic := false
	if value, ok := ctxData["grant_magic"].(bool); ok {
		grantMagic = value
	}
	if grantAttack {
		model.AppendAttackAction(user, skillName)
	}
	if grantMagic {
		model.AppendMagicAction(user, skillName)
	}

	e.Log(fmt.Sprintf("%s 发动 [%s]，对 %s 造成%d点法术伤害", user.Name, skillName, target.Name, damage))
	e.PopInterrupt()
	return nil
}
