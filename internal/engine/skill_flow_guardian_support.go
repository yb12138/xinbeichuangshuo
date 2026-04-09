package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildGuardianSupportChoicePrompt(choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "angel_bond_heal_target":
		options := buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【天使羁绊】请选择1名角色获得+1治疗：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "frost_prayer_target":
		options := buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【冰霜祷言】请选择1名角色获得+1治疗：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "god_protection_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			return nil
		}
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-1),
				Label: fmt.Sprintf("消耗%d点水晶，抵御%d点士气下降", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【神之庇护】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	default:
		return nil
	}
}

func (e *GameEngine) handleGuardianSupportChoiceInput(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"angel_bond_heal_target": func(idx int, data map[string]interface{}) error {
			return e.handleAngelBondHealChoice(playerID, idx, data)
		},
		"frost_prayer_target": e.handleFrostPrayerChoice,
		"god_protection_x":    e.handleGodProtectionChoice,
	})
}

func (e *GameEngine) handleGodProtectionChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	userCtx, ok := ctxData["user_ctx"].(*model.Context)
	if !ok || userCtx == nil || userCtx.TriggerCtx == nil || userCtx.TriggerCtx.DamageVal == nil {
		return fmt.Errorf("神之庇护上下文丢失")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	x := selectionIndex + 1
	if x < 1 || x > maxX {
		return fmt.Errorf("无效的X值")
	}
	if current := *userCtx.TriggerCtx.DamageVal; current > 0 && x > current {
		return fmt.Errorf("X值超过当前可抵御的士气下降")
	}
	if !e.ConsumeCrystalCost(user.ID, x) {
		return fmt.Errorf("神之庇护需要%d点水晶（红宝石可替代）", x)
	}
	*userCtx.TriggerCtx.DamageVal -= x
	if *userCtx.TriggerCtx.DamageVal < 0 {
		*userCtx.TriggerCtx.DamageVal = 0
	}
	e.Log(fmt.Sprintf("%s 发动 [神之庇护]，消耗%d点水晶（可由红宝石替代）抵御了%d点士气下降", user.Name, x, x))

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil && e.resumePendingMoraleLoss(userCtx) {
		return nil
	}
	if e.State.PendingInterrupt == nil {
		e.enterResponseWindow()
	}
	return nil
}

func (e *GameEngine) handleAngelBondHealChoice(playerID string, selectionIndex int, ctxData map[string]interface{}) error {
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	e.Heal(target.ID, 1)
	userName := playerID
	if user := e.State.Players[playerID]; user != nil {
		userName = user.Name
	}
	e.Log(fmt.Sprintf("%s 的 [天使羁绊] 生效：%s 获得 +1 治疗", userName, target.Name))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.applyChoiceResumePoint(mustChoiceResumePointFromMap(ctxData, "resume_phase"))
	}
	return nil
}

func (e *GameEngine) handleFrostPrayerChoice(selectionIndex int, ctxData map[string]interface{}) error {
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
	e.Heal(targetID, 1)
	e.Log(fmt.Sprintf("%s 的 [冰霜祷言] 生效：%s +1治疗", user.Name, target.Name))
	e.PopInterrupt()
	return nil
}
