package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildMagicSwordsmanChoicePrompt(choiceType, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	if choiceType != "ms_shadow_meteor_release_confirm" {
		return nil
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【暗影流星】是否额外移除我方战绩区2个星石，转正并+1红宝石？",
		Options: []model.PromptOption{
			{ID: "0", Label: "是"},
			{ID: "1", Label: "否"},
		},
		Min: 1,
		Max: 1,
	}
}

func (e *GameEngine) handleMagicSwordsmanChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	if choiceType, _ := ctxData["choice_type"].(string); choiceType != "ms_shadow_meteor_release_confirm" {
		return false, nil
	}

	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	if selectionIndex == 0 {
		camp, _ := ctxData["camp"].(string)
		need := 2
		useCrystal := need
		if useCrystal > e.GetCampCrystals(camp) {
			useCrystal = e.GetCampCrystals(camp)
		}
		if useCrystal > 0 {
			e.ModifyCrystal(camp, -useCrystal)
		}
		remain := need - useCrystal
		if remain > 0 {
			e.ModifyGem(camp, -remain)
		}
		leaveMagicSwordsmanShadowForm(user)
		user.Gem++
		e.Log(fmt.Sprintf("%s 通过[暗影流星]额外效果转正并获得1红宝石", user.Name))
	}
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.routePendingDamageOr(model.TurnStageExtraAction, func() {
			e.enterExtraActionStage()
		})
	}
	return true, nil
}
