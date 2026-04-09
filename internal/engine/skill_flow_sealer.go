package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func (e *GameEngine) buildSealerChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	if choiceType != "five_elements_bind" {
		return nil
	}
	drawCount := runtimeutil.ToIntContextValue(data["draw_count"])
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  fmt.Sprintf("【五系束缚】%s，你需要做出选择：", player.Name),
		Options: []model.PromptOption{
			{ID: "0", Label: fmt.Sprintf("摸 %d 张牌取消束缚", drawCount)},
			{ID: "1", Label: "跳过行动阶段"},
		},
		Min: 1,
		Max: 1,
	}
}

func (e *GameEngine) handleSealerChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"five_elements_bind": e.handleFiveElementsBindChoice,
	})
}

func (e *GameEngine) handleFiveElementsBindChoice(selectionIndex int, ctxData map[string]interface{}) error {
	drawCount := runtimeutil.ToIntContextValue(ctxData["draw_count"])
	targetPlayerID, _ := ctxData["player_id"].(string)
	player := e.State.Players[targetPlayerID]
	if player == nil {
		e.PopInterrupt()
		return fmt.Errorf("五系束缚目标玩家不存在")
	}

	e.RemoveFieldCard(player.ID, model.EffectFiveElementsBind)

	switch selectionIndex {
	case 0:
		e.Log(fmt.Sprintf("[FiveElementsBind] %s 选择摸 %d 张牌", player.Name, drawCount))
		cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, drawCount)
		e.State.Deck = newDeck
		e.State.DiscardPile = newDiscard
		player.Hand = append(player.Hand, cards...)
		e.NotifyDrawCards(player.ID, drawCount, "five_elements_bind")

		checkCtx := e.buildContext(player, nil, model.TriggerNone, nil)
		checkCtx.Flags["StayInTurn"] = true
		e.checkHandLimit(player, checkCtx)

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setTurnStage(model.TurnStageActionStart)
			e.clearCombatStage()
			e.clearSubflow()
		}
		return nil
	case 1:
		e.Log(fmt.Sprintf("[FiveElementsBind] %s 选择放弃行动", player.Name))
		player.TurnState.UsedSkillCounts["arbiter_skip_forced_doomsday"] = 1

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterTurnEndStage()
		}
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}
