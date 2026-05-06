// gameflow: 元素师角色选择流。

package elementalist

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(_ engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "elementalist_bonus_card":
		if player == nil {
			return nil
		}
		skillName, _ := data["skill_display_name"].(string)
		if skillName == "" {
			skillName = "元素附加效果"
		}
		eleLabel := promptfmt.ElementName(fmt.Sprint(data["bonus_element"]))
		matching := runtimeutil.ParseChoiceIntSlice(data["matching_indices"])
		options := make([]model.PromptOption, 0, len(matching)+1)
		for _, idx := range matching {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
			})
		}
		options = append(options, model.PromptOption{ID: "cancel", Label: "放弃额外效果"})
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【%s】可额外弃1张%s系牌使本次法术伤害+1（或点击取消放弃本次额外效果）：", skillName, eleLabel),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	default:
		return nil
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "elementalist_bonus_card":
		return true, handleElementalistBonusCardChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, _ string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "elementalist_bonus_card" {
		return false, nil
	}
	return true, resolveElementalistBonus(rt, ctxData, false, -1)
}

func handleElementalistBonusCardChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	matching := runtimeutil.ParseChoiceIntSlice(ctxData["matching_indices"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, matching)
	if !ok {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	return resolveElementalistBonus(rt, ctxData, true, cardIdx)
}

func resolveElementalistBonus(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, bonus bool, discardIdx int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetID, _ := ctxData["damage_target_id"].(string)
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}

	baseDamage := runtimeutil.ToIntContextValue(ctxData["base_damage"])
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
		rt.NotifyCardRevealed(userID, []model.Card{card}, "discard")
		user.Hand = append(user.Hand[:discardIdx], user.Hand[discardIdx+1:]...)
		rt.AppendToDiscard([]model.Card{card})
		damage++
	}

	rt.InflictDamage(userID, targetID, damage, model.MagicAttack)

	if healTargetID, ok := ctxData["heal_target_id"].(string); ok && healTargetID != "" {
		if healTarget := rt.GetPlayers()[healTargetID]; healTarget != nil {
			rt.Heal(healTargetID, 1)
			rt.Log(fmt.Sprintf("[元素师] %s 为 %s 提供了1点治疗", skillName, healTarget.Name))
		}
	}

	campGemBonus := runtimeutil.ToIntContextValue(ctxData["camp_gem_bonus"])
	if campGemBonus > 0 {
		rt.ModifyGem(string(user.Camp), campGemBonus)
	}

	if runtimeutil.ToBoolContextValue(ctxData["grant_attack"]) {
		model.AppendAttackAction(user, skillName)
	}
	if runtimeutil.ToBoolContextValue(ctxData["grant_magic"]) {
		model.AppendMagicAction(user, skillName)
	}

	rt.Log(fmt.Sprintf("%s 发动 [%s]，对 %s 造成%d点法术伤害", user.Name, skillName, target.Name, damage))
	rt.PopInterrupt()
	return nil
}

var _ engineplayer.CancelChoiceHandler = choiceHandler{}
