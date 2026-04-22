// gameflow: 风之剑圣中断 prompt 与 action 处理。

package blade_master

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// MaybeHolySwordDrawInterrupt 检查圣剑第3次攻击结束后的摸牌弃牌中断。
// 作为 TimingOnActionEnd 的 TimingHookSpec 使用。
func MaybeHolySwordDrawInterrupt(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return engineplayer.TimingHookResult{}
	}
	// 只在第3次攻击时触发
	if p.TurnState.AttackCount != 3 {
		return engineplayer.TimingHookResult{}
	}
	// 检查是否有圣剑技能
	hasHolySword := false
	if p.Character != nil {
		for _, skill := range p.Character.Skills {
			if skill.ID == "holy_sword" {
				hasHolySword = true
				break
			}
		}
	}
	if !hasHolySword {
		return engineplayer.TimingHookResult{}
	}
	// 忽略反击发起者的攻击
	attackInfo := ctx.AttackInfo
	if attackInfo != nil && attackInfo.CounterInitiator != "" {
		return engineplayer.TimingHookResult{}
	}

	// 推送圣剑摸X弃X中断
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptHolySwordDraw,
		PlayerID: p.ID,
		Context: map[string]any{
			"choice_type": "holy_sword_draw",
			"player_id":   p.ID,
		},
	})
	rt.Log(fmt.Sprintf("[Skill] %s 的 [圣剑] 第3次攻击结束，需选择摸X弃X (X=0-3)", p.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}

// buildHolySwordDrawPrompt 构建圣剑第3次攻击结束后的摸牌弃牌提示。
func buildHolySwordDrawPrompt(rt engineplayer.ChoiceRuntime) *model.Prompt {
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
func handleHolySwordDrawAction(rt engineplayer.ChoiceRuntime, act model.PlayerAction) error {
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
	rt.PushDiscardChoiceInterrupt(p.ID, map[string]any{
		"discard_count":        x,
		"is_holy_sword":        true,
		"stay_in_turn":         true,
		"is_damage_resolution": false,
	})
	rt.Log(fmt.Sprintf("[Skill] %s 需要弃 %d 张牌", p.Name, x))
	return nil
}

// resumeHolySwordAftermath 圣剑后续恢复流程。
func resumeHolySwordAftermath(rt engineplayer.ChoiceRuntime) {
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

// DispatchHolySwordInterruptForTest 测试辅助函数，直接触发圣剑摸X弃X中断。
// 在生产环境中通过 TimingOnActionEnd 的 TimingHookSpec 自动触发。
func DispatchHolySwordInterruptForTest(engine interface {
	PushInterrupt(intr *model.Interrupt)
	Log(msg string)
}, attacker *model.Player) bool {
	if attacker == nil || attacker.Character == nil || attacker.TurnState.AttackCount != 3 {
		return false
	}
	hasHolySword := false
	for _, skill := range attacker.Character.Skills {
		if skill.ID == "holy_sword" {
			hasHolySword = true
			break
		}
	}
	if !hasHolySword {
		return false
	}
	engine.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptHolySwordDraw,
		PlayerID: attacker.ID,
		Context: map[string]any{
			"choice_type": "holy_sword_draw",
			"player_id":   attacker.ID,
		},
	})
	engine.Log(fmt.Sprintf("[Skill] %s 的 [圣剑] 第3次攻击结束，需选择摸X弃X (X=0-3)", attacker.Name))
	return true
}
