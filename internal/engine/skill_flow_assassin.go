// gameflow: 暗杀者技能流。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildAssassinChoicePrompt(choiceType, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "assassin_stealth_draw":
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			Message:    "【潜行】是否摸1张牌后进入潜行状态？",
			ChoiceType: choiceType,
			Options: []model.PromptOption{
				{ID: "0", Label: "摸1张牌"},
				{ID: "1", Label: "不摸牌"},
			},
			Min: 1,
			Max: 1,
		}
	default:
		return nil
	}
}

func (e *GameEngine) handleAssassinChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "assassin_stealth_draw":
		return true, e.handleAssassinStealthDrawChoice(selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func (e *GameEngine) handleAssassinStealthDrawChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	switch selectionIndex {
	case 0:
		drawCtx := e.newDrawContext(user, 1, "assassin_stealth_draw")
		if drawCtx == nil {
			return fmt.Errorf("潜行摸牌上下文创建失败")
		}
		drawCtx.Selections["draw_followup"] = model.DeferredFollowup{
			Type:    "assassin_stealth_apply",
			UserID:  user.ID,
			SkillID: "stealth",
		}
		e.startDraw(drawCtx)
		e.Log(fmt.Sprintf("%s 的 [潜行]：选择先摸1张牌，再进入潜行状态", user.Name))

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.restorePhaseAfterInterruptedDraw(drawCtx)
		}
		return nil
	case 1:
		e.applyAssassinStealthEffect(user)
		e.Log(fmt.Sprintf("%s 的 [潜行]：选择不摸牌，直接进入潜行状态", user.Name))

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			// 规则：潜行选择结束后要回到触发前的等待阶段，不允许隐式回落到任意默认阶段。
			e.applyChoiceResumePoint(mustChoiceResumePointFromMap(ctxData, "waiting_phase"))
		}
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}
