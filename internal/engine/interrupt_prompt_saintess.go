package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

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
	if a <= 0 || b <= 0 || a+b != 3 {
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
		return saintHealStageChooseExtraAction, nil
	}
	switch stage {
	case saintHealStageAllocateHeal, saintHealStageChooseExtraAction:
		return stage, nil
	default:
		return "", fmt.Errorf("无效的圣疗阶段: %s", stage)
	}
}

func (e *GameEngine) saintHealAllocationSummary(targetIDs []string, allocations map[string]int) string {
	parts := make([]string, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		target := e.State.Players[targetID]
		if target == nil {
			continue
		}
		amount := allocations[targetID]
		if amount <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s +%d治疗", target.Name, amount))
	}
	return strings.Join(parts, "，")
}

func (e *GameEngine) buildSaintHealPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
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
		first := e.State.Players[targetIDs[0]]
		second := e.State.Players[targetIDs[1]]
		if first == nil || second == nil {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: interrupt.PlayerID,
			Message:  "【圣疗】请选择3点治疗的分配方式：",
			Options: []model.PromptOption{
				{ID: "0", Label: fmt.Sprintf("%s +2，%s +1", first.Name, second.Name)},
				{ID: "1", Label: fmt.Sprintf("%s +1，%s +2", first.Name, second.Name)},
			},
			Min: 1,
			Max: 1,
		}
	}

	allocations, err := saintHealAllocationsFromContext(data, targetIDs)
	if err != nil {
		return nil
	}
	summary := e.saintHealAllocationSummary(targetIDs, allocations)
	if summary == "" {
		summary = "已选择治疗目标"
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: interrupt.PlayerID,
		Message:  fmt.Sprintf("【圣疗】%s。请选择额外行动类型：", summary),
		Options: []model.PromptOption{
			{ID: "0", Label: "额外攻击行动"},
			{ID: "1", Label: "额外法术行动"},
		},
		Min: 1,
		Max: 1,
	}
}
