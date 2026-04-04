package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildHeroAssassinChoicePrompt(choiceType, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "hero_roar_draw":
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			Message:    "【怒吼】请选择摸牌数量：",
			ChoiceType: choiceType,
			Options: []model.PromptOption{
				{ID: "0", Label: "摸0张"},
				{ID: "1", Label: "摸1张"},
			},
			Min: 1,
			Max: 1,
		}
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

func (e *GameEngine) handleHeroAssassinChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "hero_roar_draw":
		return true, e.handleHeroRoarDrawChoice(selectionIndex, ctxData)
	case "assassin_stealth_draw":
		return true, e.handleAssassinStealthDrawChoice(selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func (e *GameEngine) handleHeroRoarDrawChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	drawCount := 0
	switch selectionIndex {
	case 0:
		drawCount = 0
	case 1:
		drawCount = 1
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if drawCount > 0 {
		e.DrawCards(user.ID, drawCount)
		markRoarOverflowStayInTurn := func(intr *model.Interrupt) {
			if intr == nil || intr.Type != model.InterruptDiscard || intr.PlayerID != user.ID {
				return
			}
			data, ok := intr.Context.(map[string]interface{})
			if !ok || data == nil {
				return
			}
			if _, hasSkillDiscard := data["skill_id"]; hasSkillDiscard {
				return
			}
			victimID, _ := data["victim_id"].(string)
			if victimID != user.ID {
				return
			}
			if fromDamage, _ := data["from_damage_draw"].(bool); fromDamage {
				return
			}
			data["stay_in_turn"] = true
			intr.Context = data
		}
		markRoarOverflowStayInTurn(e.State.PendingInterrupt)
		for _, intr := range e.State.InterruptQueue {
			markRoarOverflowStayInTurn(intr)
		}
	}
	e.Log(fmt.Sprintf("%s 的 [怒吼] 结算：摸%d张牌", user.Name, drawCount))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.enterActionExecutionStage()
	}
	return nil
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
