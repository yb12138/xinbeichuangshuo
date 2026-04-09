package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildSwordEmperorChoicePrompt(choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "se_sword_qi_slash_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			maxX = 1
		}
		options := make([]model.PromptOption, 0, maxX)
		for xValue := 1; xValue <= maxX; xValue++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", xValue), Label: fmt.Sprintf("移除%d点剑气，对另一名角色造成%d点法术伤害", xValue, xValue)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【剑气斩】请选择X值：", Options: options, Min: 1, Max: 1}
	}

	return nil
}

func (e *GameEngine) handleSwordEmperorChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"se_sword_qi_slash_x":      e.handleSwordEmperorSwordQiSlashXChoice,
		"se_sword_qi_slash_target": e.handleSwordEmperorSwordQiSlashTargetChoice,
	})
}

func (e *GameEngine) handleSwordEmperorSwordQiSlashXChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	if maxX <= 0 {
		return fmt.Errorf("剑气斩没有可选X值")
	}
	if selectionIndex < 0 || selectionIndex >= maxX {
		return fmt.Errorf("无效的X值选项: %d", selectionIndex)
	}
	xValue := selectionIndex + 1
	currentQi := swordEmperorSwordQi(user)
	if xValue > currentQi {
		return fmt.Errorf("剑气不足，当前只有%d点", currentQi)
	}
	if xValue > 3 {
		return fmt.Errorf("剑气斩的X不能超过3")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if len(targetIDs) == 0 {
		return fmt.Errorf("没有可选的剑气斩目标")
	}
	ctxData["x_value"] = xValue
	ctxData["choice_type"] = "se_sword_qi_slash_target"
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handleSwordEmperorSwordQiSlashTargetChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	if xValue <= 0 {
		return fmt.Errorf("剑气斩的X值无效")
	}
	if xValue > swordEmperorSwordQi(user) {
		return fmt.Errorf("剑气不足，无法移除%d点", xValue)
	}
	if rawCtx, _ := ctxData["user_ctx"].(*model.Context); rawCtx != nil && rawCtx.TriggerCtx != nil && rawCtx.TriggerCtx.TargetID == targetID {
		return fmt.Errorf("剑气斩不能选择当前攻击目标")
	}
	nowQi := addSwordEmperorSwordQi(user, -xValue)
	e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: xValue, DamageType: model.MagicAttack})
	e.Log(fmt.Sprintf("%s 发动 [剑气斩]：移除%d点剑气（当前%d），对 %s 造成%d点法术伤害", user.Name, xValue, nowQi, target.Name, xValue))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.resumePendingAttackHit(ctxData)
	}
	return nil
}
