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
	if ids, ok := data["targets"].([]string); ok {
		return append([]string{}, ids...)
	}
	raw, _ := data["targets"].([]interface{})
	if len(raw) == 0 {
		return nil
	}
	ids := make([]string, 0, len(raw))
	for _, v := range raw {
		if id, ok := v.(string); ok && id != "" {
			ids = append(ids, id)
		}
	}
	return ids
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

func saintHealAllocationsFromContext(data map[string]interface{}, targetIDs []string) map[string]int {
	if data == nil {
		return saintHealDefaultAllocations(targetIDs)
	}
	if allocations, ok := data["allocations"].(map[string]int); ok && len(allocations) > 0 {
		out := make(map[string]int, len(allocations))
		for targetID, amount := range allocations {
			out[targetID] = amount
		}
		return out
	}
	if raw, ok := data["allocations"].(map[string]interface{}); ok && len(raw) > 0 {
		out := make(map[string]int, len(raw))
		for targetID, value := range raw {
			switch v := value.(type) {
			case int:
				out[targetID] = v
			case float64:
				out[targetID] = int(v)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return saintHealDefaultAllocations(targetIDs)
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
	stage, _ := data["stage"].(string)
	if stage == "" {
		if len(targetIDs) == 2 {
			stage = "allocate_heal"
		} else {
			stage = "choose_extra_action"
		}
	}

	if stage == "allocate_heal" {
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

	allocations := saintHealAllocationsFromContext(data, targetIDs)
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
