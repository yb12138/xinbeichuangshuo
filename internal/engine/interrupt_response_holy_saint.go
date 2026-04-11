// gameflow: 圣系响应（如圣击等）分支。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// handleHolySwordDrawResponse 处理圣剑摸X弃X响应。
func (e *GameEngine) handleHolySwordDrawResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	x := 0
	if len(act.Selections) > 0 {
		x = act.Selections[0]
	}
	if x < 0 || x > 3 {
		x = 0
	}

	e.PopInterrupt()
	if x == 0 {
		e.Log(fmt.Sprintf("[Skill] %s 选择不摸不弃", player.Name))
		e.resumeHolySwordAftermath()
		return nil
	}

	e.DrawCards(player.ID, x)
	e.Log(fmt.Sprintf("[Skill] %s 摸了 %d 张牌", player.Name, x))
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptDiscard,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"discard_count":        x,
			"is_holy_sword":        true,
			"stay_in_turn":         true,
			"is_damage_resolution": false,
		},
	})
	e.Log(fmt.Sprintf("[Skill] %s 需要弃 %d 张牌", player.Name, x))
	return nil
}

func (e *GameEngine) resumeHolySwordAftermath() {
	if e.State.PendingInterrupt != nil {
		return
	}
	if e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction) {
		return
	}
	if e.restoreReturnPoint() {
		return
	}
	e.enterExtraActionStage()
}

// handleSaintHealResponse 处理圣疗分配治疗响应。
func (e *GameEngine) handleSaintHealResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}
	if act.PlayerID != interrupt.PlayerID {
		return fmt.Errorf("不是你的响应回合")
	}

	data, targetIDs, stage, err := parseSaintHealInterruptContext(interrupt)
	if err != nil {
		return err
	}

	// 阶段1（双目标专用）：先确定 3 点治疗在两名目标之间的分配。
	if stage == saintHealStageAllocateHeal {
		return e.resolveSaintHealAllocationStage(act, data, targetIDs)
	}

	// 阶段2：确认额外行动类型后，再按分配结果统一结算治疗并收束流程。
	return e.resolveSaintHealExtraActionStage(act, data, targetIDs)
}

func parseSaintHealInterruptContext(interrupt *model.Interrupt) (map[string]interface{}, []string, string, error) {
	// 圣疗流程上下文：
	// targets = 受治疗目标，stage = 当前交互阶段，allocations = 已确定的治疗分配。
	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return nil, nil, "", fmt.Errorf("中断上下文格式错误")
	}
	targetIDs := saintHealTargetIDsFromContext(data)
	if len(targetIDs) == 0 {
		return nil, nil, "", fmt.Errorf("圣疗缺少目标")
	}
	stage, err := saintHealStageFromContext(data, targetIDs)
	if err != nil {
		return nil, nil, "", err
	}
	return data, targetIDs, stage, nil
}

func (e *GameEngine) resolveSaintHealAllocationStage(
	act model.PlayerAction,
	data map[string]interface{},
	targetIDs []string,
) error {
	// 双目标时先做 2/1 与 1/2 的分配确认，随后进入“额外行动类型选择”阶段。
	if len(targetIDs) != 2 {
		return fmt.Errorf("圣疗双目标分配配置无效")
	}
	if act.Type != model.CmdSelect || len(act.Selections) != 1 {
		return fmt.Errorf("请选择一种治疗分配方式")
	}

	choice := act.Selections[0]
	if choice != 0 && choice != 1 {
		return fmt.Errorf("无效的圣疗分配选项: %d", choice)
	}

	allocations := map[string]int{}
	if choice == 0 {
		allocations[targetIDs[0]] = 2
		allocations[targetIDs[1]] = 1
	} else {
		allocations[targetIDs[0]] = 1
		allocations[targetIDs[1]] = 2
	}

	data["allocations"] = allocations
	data["stage"] = saintHealStageChooseExtraAction
	e.State.PendingInterrupt.Context = data
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) resolveSaintHealExtraActionStage(
	act model.PlayerAction,
	data map[string]interface{},
	targetIDs []string,
) error {
	// 额外行动类型选择完成后，再一次性结算治疗并收尾，避免流程前后穿插。
	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}
	if act.Type != model.CmdSelect || len(act.Selections) != 1 {
		return fmt.Errorf("请选择额外行动类型")
	}

	extraActionType, extraActionLabel, err := parseSaintHealExtraActionSelection(act.Selections[0])
	if err != nil {
		return err
	}
	allocations, err := saintHealAllocationsFromContext(data, targetIDs)
	if err != nil {
		return err
	}

	e.applySaintHealAllocations(targetIDs, allocations)

	e.PopInterrupt()
	model.AppendExtraAction(player, "圣疗", extraActionType)
	e.Log(fmt.Sprintf("[Skill] %s 发动 [圣疗]，获得额外%s行动", player.Name, extraActionLabel))
	player.TurnState.HasActed = true
	player.TurnState.LastActionType = string(model.ActionMagic)
	player.TurnState.LastActionCard = nil
	if !e.routePendingDamageWithReturn(model.TurnStageActionEnd) {
		e.enterActionEndStage()
	}
	return nil
}

func parseSaintHealExtraActionSelection(selection int) (string, string, error) {
	switch selection {
	case 0:
		return "Attack", "攻击", nil
	case 1:
		return "Magic", "法术", nil
	default:
		return "", "", fmt.Errorf("无效的额外行动类型选项: %d", selection)
	}
}

func (e *GameEngine) applySaintHealAllocations(targetIDs []string, allocations map[string]int) {
	for _, targetID := range targetIDs {
		healAmount := allocations[targetID]
		if healAmount <= 0 {
			continue
		}
		e.Heal(targetID, healAmount)
		if target := e.State.Players[targetID]; target != nil {
			e.Log(fmt.Sprintf("[Skill] %s 获得 %d 点治疗", target.Name, healAmount))
		}
	}
}
