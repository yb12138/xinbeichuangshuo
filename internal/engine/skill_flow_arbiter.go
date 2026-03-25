package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildArbiterChoicePrompt(choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "arbiter_forced_doomsday_target":
		options := buildPromptOptionsForPlayerIDs(e.State.Players, parseStringSliceContextValue(data["target_ids"]))
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【末日审判（强制）】请选择目标角色：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "arbiter_balance_mode":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【判决天平】请选择一个分支：",
			Options: []model.PromptOption{
				{ID: "0", Label: "弃掉当前手上的所有手牌"},
				{ID: "1", Label: "将手牌补到上限，并我方战绩区+1红宝石"},
			},
			Min: 1,
			Max: 1,
		}
	default:
		return nil
	}
}

func (e *GameEngine) handleArbiterChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "arbiter_forced_doomsday_target":
		return true, e.handleArbiterForcedDoomsdayChoice(selectionIndex, ctxData)
	case "arbiter_balance_mode":
		return true, e.handleArbiterBalanceChoice(selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func (e *GameEngine) handleArbiterForcedDoomsdayChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	ensurePlayerTokensMap(user)
	judgment := user.Tokens["judgment"]
	e.RemoveFieldCard(user.ID, model.EffectHeroTaunt)
	user.Tokens["judgment"] = 0
	user.Tokens["arbiter_forced_doomsday_done_turn"] = 1
	if judgment > 0 {
		e.InflictDamage(userID, targetID, judgment, "magic")
	}
	e.Log(fmt.Sprintf("%s 触发强制 [末日审判]，对 %s 造成%d点法术伤害", user.Name, target.Name, judgment))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.routePendingDamageOr(model.TurnStageTurnEnd, func() {
			e.enterTurnEndStage()
		})
	}
	return nil
}

func (e *GameEngine) handleArbiterBalanceChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		for _, c := range user.Hand {
			e.State.DiscardPile = append(e.State.DiscardPile, c)
		}
		user.Hand = nil
		e.Log(fmt.Sprintf("%s 选择判决天平分支1：弃掉所有手牌", user.Name))
	case 1:
		maxHand := user.MaxHand
		if len(user.Hand) < maxHand {
			e.DrawCards(user.ID, maxHand-len(user.Hand))
		}
		e.ModifyGem(string(user.Camp), 1)
		e.Log(fmt.Sprintf("%s 选择判决天平分支2：补牌到上限并我方战绩区+1红宝石", user.Name))
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	e.PopInterrupt()
	return nil
}
