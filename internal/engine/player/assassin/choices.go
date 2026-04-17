// gameflow: 暗杀者角色选择流。

package assassin

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

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "assassin_stealth_draw":
		return true, handleAssassinStealthDrawChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleAssassinStealthDrawChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	switch selectionIndex {
	case 0:
		drawCtx := rt.NewDrawContext(user, 1, "assassin_stealth_draw")
		if drawCtx == nil {
			return fmt.Errorf("潜行摸牌上下文创建失败")
		}
		drawCtx.Selections["draw_followup"] = model.DeferredFollowup{
			Type:    "assassin_stealth_apply",
			UserID:  user.ID,
			SkillID: "stealth",
		}
		rt.StartDraw(drawCtx)
		rt.Log(fmt.Sprintf("%s 的 [潜行]：选择先摸1张牌，再进入潜行状态", user.Name))

		rt.PopInterrupt()
		if !rt.HasPendingInterrupt() {
			rt.RestorePhaseAfterInterruptedDraw(drawCtx)
		}
		return nil
	case 1:
		rt.ApplyStealthEffect(user)
		rt.Log(fmt.Sprintf("%s 的 [潜行]：选择不摸牌，直接进入潜行状态", user.Name))

		rt.PopInterrupt()
		if !rt.HasPendingInterrupt() {
			// 规则：潜行选择结束后要回到触发前的等待阶段，不允许隐式回落到任意默认阶段。
			rt.ApplyChoiceResumePoint(mustChoiceResumePointFromMap(ctxData, "waiting_phase"))
		}
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

func mustChoiceResumePointFromMap(data map[string]interface{}, key string) interface{} {
	if data == nil {
		panic(fmt.Sprintf("missing resume point map for key %q", key))
	}
	raw, ok := data[key]
	if !ok {
		panic(fmt.Sprintf("missing resume point key %q", key))
	}
	return mustChoiceResumePoint(raw, key)
}

func mustChoiceResumePoint(raw interface{}, key string) interface{} {
	if raw == nil {
		panic(fmt.Sprintf("nil resume point for key %q", key))
	}
	return raw
}
