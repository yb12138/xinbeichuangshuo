// gameflow: 魔剑士角色选择流。

package magic_swordsman

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
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
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "ms_shadow_meteor_release_confirm":
		return true, handleMagicSwordsmanShadowMeteorReleaseConfirmChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleMagicSwordsmanShadowMeteorReleaseConfirmChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex == 0 {
		camp, _ := ctxData["camp"].(string)
		need := 2
		useCrystal := min(need, rt.GetCampCrystals(camp))
		if useCrystal > 0 {
			rt.ModifyCrystal(camp, -useCrystal)
		}
		remain := need - useCrystal
		if remain > 0 {
			rt.ModifyGem(camp, -remain)
		}
		user.Form = ""
		gainedGem := engineplayer.AddPlayerGemWithCap(rt, user, 1)
		rt.Log(fmt.Sprintf("%s 通过[暗影流星]额外效果转正并获得%d红宝石", user.Name, gainedGem))
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.RoutePendingDamageOr(model.TurnStageExtraAction, func() {
			rt.EnterExtraActionStage()
		})
	}
	return nil
}
