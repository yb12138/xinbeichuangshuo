// gameflow: 圣女中断 prompt 与 action 处理。

package saintess

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const (
	saintHealStageAllocateHeal = "allocate_heal"
)

// --- SaintHeal helpers ---

func saintHealTargetIDsFromContext(data map[string]interface{}) []string {
	if data == nil {
		return nil
	}
	ids, ok := data["targets"].([]string)
	if !ok || len(ids) == 0 {
		return nil
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			return nil
		}
		out = append(out, id)
	}
	return out
}

func saintHealDefaultAllocations(targetIDs []string) map[string]int {
	allocations := map[string]int{}
	switch len(targetIDs) {
	case 1:
		allocations[targetIDs[0]] = 3
	case 2:
		allocations[targetIDs[0]] = 2
		allocations[targetIDs[1]] = 1
	case 3:
		for _, targetID := range targetIDs {
			allocations[targetID] = 1
		}
	}
	return allocations
}

func saintHealAllocationsFromContext(data map[string]interface{}, targetIDs []string) (map[string]int, error) {
	if len(targetIDs) != 2 {
		return saintHealDefaultAllocations(targetIDs), nil
	}
	if data == nil {
		return nil, fmt.Errorf("圣疗双目标缺少治疗分配")
	}
	raw, ok := data["allocations"].(map[string]int)
	if !ok || len(raw) == 0 {
		return nil, fmt.Errorf("圣疗双目标缺少治疗分配")
	}

	out := map[string]int{
		targetIDs[0]: raw[targetIDs[0]],
		targetIDs[1]: raw[targetIDs[1]],
	}
	a := out[targetIDs[0]]
	b := out[targetIDs[1]]
	if a < 0 || b < 0 || a+b > 3 {
		return nil, fmt.Errorf("圣疗双目标治疗分配无效")
	}
	return out, nil
}

func saintHealStageFromContext(data map[string]interface{}, targetIDs []string) (string, error) {
	stage, _ := data["stage"].(string)
	if stage == "" {
		if len(targetIDs) == 2 {
			return saintHealStageAllocateHeal, nil
		}
		return "", nil
	}
	switch stage {
	case saintHealStageAllocateHeal:
		return stage, nil
	default:
		return "", fmt.Errorf("无效的圣疗阶段: %s", stage)
	}
}

// --- SaintHeal prompt ---

func buildSaintHealPrompt(rt player.ChoiceRuntime) *model.Prompt {
	interrupt := rt.GetPendingInterrupt()
	if interrupt == nil {
		return nil
	}
	data, _ := interrupt.Context.(map[string]interface{})
	targetIDs := saintHealTargetIDsFromContext(data)
	if len(targetIDs) == 0 {
		return nil
	}
	stage, err := saintHealStageFromContext(data, targetIDs)
	if err != nil {
		return nil
	}

	if stage == saintHealStageAllocateHeal {
		if len(targetIDs) != 2 {
			return nil
		}
		first := rt.GetPlayers()[targetIDs[0]]
		second := rt.GetPlayers()[targetIDs[1]]
		if first == nil || second == nil {
			return nil
		}
		// 选项顺序对应 selections：selections[0]=第一目标治疗点数，selections[1]=第二目标治疗点数。
		// 前端据 ChoiceType 识别为分配模式，渲染每个目标独立的 0-3 数字选择器并约束总和<=3。
		// 显示角色名和当前治疗量，便于玩家判断分配策略。
		firstLabel := fmt.Sprintf("%s（治疗:%d）", first.Character.Name, first.Heal)
		secondLabel := fmt.Sprintf("%s（治疗:%d）", second.Character.Name, second.Heal)
		return &model.Prompt{
			Type:       model.PromptConfirm,
			ChoiceType: "saint_heal_allocate",
			PlayerID:   interrupt.PlayerID,
			Message:    "【圣疗】请分配治疗（两名角色之和不超过 3，单项可为 0）：",
			Options: []model.PromptOption{
				{ID: targetIDs[0], Label: firstLabel, Hint: "max:3"},
				{ID: targetIDs[1], Label: secondLabel, Hint: "max:3"},
			},
			Min:          2,
			Max:          2,
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, Layout: "heal_allocate", NumericBase: 0},
		}
	}

	return nil
}

// --- SaintHeal action ---

func handleSaintHealAction(rt player.ChoiceRuntime, act model.PlayerAction) (player.InterruptActionResult, error) {
	interrupt := rt.GetPendingInterrupt()
	if interrupt == nil {
		return player.InterruptActionResult{}, fmt.Errorf("没有待处理的中断")
	}
	if act.PlayerID != interrupt.PlayerID {
		return player.InterruptActionResult{}, fmt.Errorf("不是你的响应回合")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return player.InterruptActionResult{}, fmt.Errorf("中断上下文格式错误")
	}
	targetIDs := saintHealTargetIDsFromContext(data)
	if len(targetIDs) == 0 {
		return player.InterruptActionResult{}, fmt.Errorf("圣疗缺少目标")
	}
	stage, err := saintHealStageFromContext(data, targetIDs)
	if err != nil {
		return player.InterruptActionResult{}, err
	}

	if stage == saintHealStageAllocateHeal {
		return resolveSaintHealAllocationStage(rt, act, data, targetIDs)
	}
	return resolveSaintHeal(rt, act.PlayerID, data, targetIDs)
}

func resolveSaintHealAllocationStage(
	rt player.ChoiceRuntime,
	act model.PlayerAction,
	data map[string]interface{},
	targetIDs []string,
) (player.InterruptActionResult, error) {
	if len(targetIDs) != 2 {
		return player.InterruptActionResult{}, fmt.Errorf("圣疗双目标分配配置无效")
	}
	if act.Type != model.CmdSelect || len(act.Selections) != 2 {
		return player.InterruptActionResult{}, fmt.Errorf("请为两名角色分别选择治疗点数")
	}

	first := act.Selections[0]
	second := act.Selections[1]
	if first < 0 || second < 0 || first > 3 || second > 3 || first+second != 3 {
		return player.InterruptActionResult{}, fmt.Errorf("圣疗分配必须满足两项之和=3 且单项在 0..3 之间，当前：%d/%d", first, second)
	}

	allocations := map[string]int{
		targetIDs[0]: first,
		targetIDs[1]: second,
	}

	data["allocations"] = allocations
	delete(data, "stage")
	return resolveSaintHeal(rt, act.PlayerID, data, targetIDs)
}

func resolveSaintHeal(
	rt player.ChoiceRuntime,
	playerID string,
	data map[string]interface{},
	targetIDs []string,
) (player.InterruptActionResult, error) {
	p := rt.GetPlayers()[playerID]
	if p == nil {
		return player.InterruptActionResult{}, fmt.Errorf("玩家不存在")
	}
	allocations, err := saintHealAllocationsFromContext(data, targetIDs)
	if err != nil {
		return player.InterruptActionResult{}, err
	}

	for _, targetID := range targetIDs {
		healAmount := allocations[targetID]
		if healAmount <= 0 {
			continue
		}
		rt.Heal(targetID, healAmount)
		if target := rt.GetPlayers()[targetID]; target != nil {
			rt.Log(fmt.Sprintf("[Skill] %s 获得 %d 点治疗", target.Name, healAmount))
		}
	}

	model.AppendExtraAction(p, "圣疗", "")
	rt.Log(fmt.Sprintf("[Skill] %s 发动 [圣疗]，获得额外行动（可选择攻击或法术）", p.Name))
	p.TurnState.HasActed = true
	p.TurnState.LastActionType = string(model.ActionMagic)
	p.TurnState.LastActionCard = nil
	return player.InterruptActionResult{Consumed: true, AfterConsume: func(rt player.ChoiceRuntime) {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageActionEnd) {
			rt.EnterActionEndStage()
		}
	}}, nil
}
