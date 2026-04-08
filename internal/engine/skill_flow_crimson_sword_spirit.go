package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildCrimsonSwordSpiritChoicePrompt(choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	if choiceType != "css_dance_mode" {
		return nil
	}
	canCrystal, _ := data["can_crystal"].(bool)
	canGem, _ := data["can_gem"].(bool)
	options := make([]model.PromptOption, 0, 2)
	if canCrystal {
		options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "消耗1蓝水晶（可用红宝石替代）：放置庭院并+2鲜血"})
	}
	if canGem {
		options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "消耗1红宝石：放置庭院并+2鲜血（上限4）且弃牌至4"})
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【散华轮舞】请选择发动分支：",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func (e *GameEngine) handleCrimsonSwordSpiritChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"css_dance_mode": e.handleCrimsonSwordSpiritDanceModeChoice,
	})
}

func (e *GameEngine) handleCrimsonSwordSpiritDanceModeChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	canCrystal, _ := ctxData["can_crystal"].(bool)
	canGem, _ := ctxData["can_gem"].(bool)
	modeList := make([]int, 0, 2)
	if canCrystal {
		modeList = append(modeList, 0)
	}
	if canGem {
		modeList = append(modeList, 1)
	}
	if selectionIndex < 0 || selectionIndex >= len(modeList) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	mode := modeList[selectionIndex]
	if user.Character == nil || user.Character.ID == "" {
		return fmt.Errorf("角色信息缺失")
	}

	courtyardCard, ok := user.ConsumeExclusiveCard(user.Character.ID, "血蔷薇庭院")
	if !ok {
		return fmt.Errorf("未找到【血蔷薇庭院】专属技能卡")
	}
	if err := e.attachExclusiveEffectCard(user, user, model.EffectRoseCourtyard, courtyardCard); err != nil {
		user.RestoreExclusiveCard(courtyardCard)
		return err
	}

	if mode == 0 {
		if !e.ConsumeCrystalCost(user.ID, 1) {
			return fmt.Errorf("蓝水晶不足（红宝石可替代）")
		}
		user.Tokens["css_blood_cap"] = 3
		addBlood(user, 2)
	} else {
		if user.Gem <= 0 {
			return fmt.Errorf("红宝石不足")
		}
		user.Gem--
		user.Tokens["css_blood_cap"] = 4
		addBlood(user, 2)
		overflow := len(user.Hand) - 4
		if overflow > 0 {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptDiscard,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"discard_count": overflow,
					"stay_in_turn":  true,
					"prompt":        fmt.Sprintf("【散华轮舞】请弃置 %d 张手牌至4张：", overflow),
				},
			})
		}
	}

	e.PopInterrupt()
	return nil
}
