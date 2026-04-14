// gameflow: 元素师技能流。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildElementalistChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "elementalist_bonus_card":
		if player == nil {
			return nil
		}
		skillName, _ := data["skill_display_name"].(string)
		eleZh := elementNameForPrompt(fmt.Sprint(data["bonus_element"]))
		matching := parseIntSliceContextValue(data["matching_indices"])
		options := make([]model.PromptOption, 0, len(matching)+1)
		for _, idx := range matching {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		options = append(options, model.PromptOption{ID: "cancel", Label: "放弃额外效果"})
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【%s】可额外弃1张%s系牌使本次法术伤害+1（或点击取消放弃本次额外效果）：", skillName, eleZh),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}

	return nil
}

func (e *GameEngine) handleElementalistChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "elementalist_bonus_card":
		return true, e.handleElementalistBonusCardChoice(selectionIndex, ctxData)
	default:
		return false, nil
	}
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

func (e *GameEngine) cancelElementalistBonusCardChoice(_ string) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("没有待处理的元素师额外弃牌选择")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "elementalist_bonus_card" {
		return fmt.Errorf("当前步骤不支持取消")
	}
	return e.resolveElementalistBonus(ctxData, false, -1)
}
