// gameflow: 风之剑圣中断 prompt 与 action 处理。

package blade_master

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// buildHolySwordDrawPrompt 构建圣剑第3次攻击结束后的摸牌弃牌提示。
func buildHolySwordDrawPrompt(rt player.ChoiceRuntime) *model.Prompt {
	interrupt := rt.PendingInterrupt()
	if interrupt == nil {
		return nil
	}
	playerID := interrupt.PlayerID

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【圣剑】第3次攻击结束！选择摸X张牌然后弃X张牌 (X=0-3)：",
		Options: []model.PromptOption{
			{ID: "0", Label: "X=0"},
			{ID: "1", Label: "X=1"},
			{ID: "2", Label: "X=2"},
			{ID: "3", Label: "X=3"},
		},
		Min: 1,
		Max: 1,
	}
}

// handleHolySwordDrawAction 处理圣剑摸X弃X响应。
func handleHolySwordDrawAction(rt player.ChoiceRuntime, act model.PlayerAction) error {
	interrupt := rt.PendingInterrupt()
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	p := rt.LookupPlayer(act.PlayerID)
	if p == nil {
		return fmt.Errorf("玩家不存在")
	}

	x := 0
	if len(act.Selections) > 0 {
		x = act.Selections[0]
	}
	if x < 0 || x > 3 {
		x = 0
	}

	rt.PopInterrupt()
	if x == 0 {
		rt.Log(fmt.Sprintf("[Skill] %s 选择不摸不弃", p.Name))
		resumeHolySwordAftermath(rt)
		return nil
	}

	rt.DrawCards(p.ID, x)
	rt.Log(fmt.Sprintf("[Skill] %s 摸了 %d 张牌", p.Name, x))
	rt.PushDiscardChoiceInterrupt(p.ID, map[string]interface{}{
		"discard_count":        x,
		"is_holy_sword":        true,
		"stay_in_turn":         true,
		"is_damage_resolution": false,
	})
	rt.Log(fmt.Sprintf("[Skill] %s 需要弃 %d 张牌", p.Name, x))
	return nil
}

// resumeHolySwordAftermath 圣剑后续恢复流程。
func resumeHolySwordAftermath(rt player.ChoiceRuntime) {
	if rt.HasPendingInterrupt() {
		return
	}
	if rt.RoutePendingDamageWithDefaultReturn(model.TurnStageExtraAction) {
		return
	}
	if rt.RestoreReturnPoint() {
		return
	}
	rt.EnterExtraActionStage()
}
