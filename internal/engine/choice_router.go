// gameflow: ChoiceEngine 与 InterruptChoice 动作入口（无弱选择 / 无顺序补丁）。

package engine

import (
	"fmt"
	"strconv"

	runtimeutil "starcup-engine/internal/engine/core/runtimeutil"
	choicert "starcup-engine/internal/engine/runtime/choice"
	intr "starcup-engine/internal/engine/runtime/interrupt"
	"starcup-engine/internal/model"
)

// choiceHostBridge 将 *GameEngine 注入 runtime/choice（无 Legacy 回退）。
type choiceHostBridge struct {
	e *GameEngine
}

var _ choicert.Host = (*choiceHostBridge)(nil)

func (*choiceHostBridge) ChoiceEngineHost() {}

// choiceCtxAsAnyMap 将中断 Context 转为 map[string]any。
func choiceCtxAsAnyMap(raw interface{}) (map[string]any, bool) {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, false
	}
	return m, true
}

func choiceCtxAsInterfaceMap(m map[string]any) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (e *GameEngine) applyInterruptChoiceSelect(playerID string, selectionIndex int, ctxData map[string]any) (choicert.HandleResult, error) {
	if e.choiceEngine == nil {
		return choicert.HandleResult{}, fmt.Errorf("选择引擎未初始化")
	}
	return e.choiceEngine.HandleSelectResult(playerID, selectionIndex, ctxData)
}

func (e *GameEngine) handleInterruptChoiceAction(act model.PlayerAction) (intr.ActionResult, error) {
	var (
		res intr.ActionResult
		err error
	)
	skillID, ownerID := e.choiceSkillCause(act.PlayerID)
	e.WithSkillCause(skillID, ownerID, func() {
		res, err = e.handleInterruptChoiceActionInCause(act)
	})
	return res, err
}

// choiceSkillCause 返回当前选择声明的技能与其持有者（未声明时为空）。
func (e *GameEngine) choiceSkillCause(actorID string) (skillID, ownerID string) {
	if e.State.PendingInterrupt == nil {
		return "", ""
	}
	data, ok := choiceCtxAsAnyMap(e.State.PendingInterrupt.Context)
	if !ok {
		return "", ""
	}
	ct, _ := data["choice_type"].(string)
	spec, ok := choiceRouteSkillSpec(ct)
	if !ok {
		return "", ""
	}
	isOwner := func(pid string) bool {
		p := e.State.Players[pid]
		return p != nil && p.Character != nil && (spec.Role == "" || p.Character.ID == spec.Role)
	}
	candidates := []string{actorID}
	for _, key := range []string{"user_id", "actor_id", "player_id", "source_id"} {
		if v, ok := data[key].(string); ok {
			candidates = append(candidates, v)
		}
	}
	for _, pid := range candidates {
		if pid != "" && isOwner(pid) {
			return spec.SkillID, pid
		}
	}
	return "", ""
}

// choiceRouteSkillSpec 在角色注册表中查找声明了技能归属的选择路由。
func choiceRouteSkillSpec(choiceType string) (ChoiceRouteSpec, bool) {
	if choiceType == "" || roleRegistry == nil {
		return ChoiceRouteSpec{}, false
	}
	for _, entry := range roleRegistry.Entries() {
		if route, ok := entry.ChoiceRoutes()[choiceType]; ok && route.SkillID != "" {
			if route.Role == "" {
				route.Role = entry.ID
			}
			return route, true
		}
	}
	return ChoiceRouteSpec{}, false
}

func (e *GameEngine) handleInterruptChoiceActionInCause(act model.PlayerAction) (intr.ActionResult, error) {
	if e.State.PendingInterrupt == nil {
		return intr.ActionResult{}, fmt.Errorf("没有待处理的中断")
	}
	if e.choiceEngine == nil {
		return intr.ActionResult{}, fmt.Errorf("选择引擎未初始化")
	}
	data, ok := choiceCtxAsAnyMap(e.State.PendingInterrupt.Context)
	if !ok {
		return intr.ActionResult{}, fmt.Errorf("中断上下文格式错误")
	}
	var result choicert.HandleResult
	var err error
	if act.Type == model.CmdCancel {
		result, err = e.choiceEngine.HandleCancelResult(act.PlayerID, data)
	} else if act.Type == model.CmdSelect {
		ct, _ := data["choice_type"].(string)
		if ct == "" {
			return intr.ActionResult{}, fmt.Errorf("中断上下文缺少 choice_type")
		}
		// 当 Selections 为空但 TargetID 非空时，从 target_ids 列表解析索引
		selections := act.Selections
		if len(selections) == 0 && len(act.CardIDs) > 0 {
			selections, err = e.choiceSelectionsFromCardIDs(act.CardIDs)
			if err != nil {
				return intr.ActionResult{}, err
			}
		}
		if len(selections) == 0 && act.TargetID != "" {
			ids := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
			for i, id := range ids {
				if id == act.TargetID {
					selections = []int{i}
					break
				}
			}
		}
		if len(selections) == 0 && act.TargetID != "" {
			return intr.ActionResult{}, fmt.Errorf("目标 %q 不在可选列表中", act.TargetID)
		}
		if e.choiceTypeRequiresMultiSelect(ct) {
			result, err = e.choiceEngine.HandleMultiSelectResult(act.PlayerID, ct, selections, data)
		} else if len(selections) > 1 {
			result, err = e.choiceEngine.HandleMultiSelectResult(act.PlayerID, ct, selections, data)
		} else {
			if len(selections) != 1 {
				return intr.ActionResult{}, fmt.Errorf("请选择一个选项")
			}
			result, err = e.applyInterruptChoiceSelect(act.PlayerID, selections[0], data)
		}
	} else {
		return intr.ActionResult{}, fmt.Errorf("当前中断类型不支持该指令")
	}
	if err != nil {
		return intr.ActionResult{}, err
	}
	// AfterPop 可能在之后别的中断弹出时才执行：带上此刻的因果。
	cause := model.CloneCause(e.CurrentCause())
	return intr.ActionResult{
		Consumed: result.ConsumedInterrupt,
		AfterPop: func(_ intr.EngineInterface) {
			if result.AfterConsume != nil {
				e.withCause(cause, func() {
					result.AfterConsume(&choiceHostBridge{e: e})
				})
			}
		},
	}, nil
}

func (e *GameEngine) choiceTypeRequiresMultiSelect(choiceType string) bool {
	switch choiceType {
	case "adventurer_fraud_pick", "bs_reversal_target_discard":
		return true
	default:
		return false
	}
}

func (e *GameEngine) choiceSelectionsFromCardIDs(cardIDs []string) ([]int, error) {
	if len(cardIDs) == 0 {
		return nil, nil
	}
	prompt := e.BuildChoicePrompt()
	if prompt == nil {
		return nil, fmt.Errorf("当前选择缺少可解析的卡牌选项")
	}
	selections := make([]int, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		matched := false
		for optionIndex, option := range prompt.Options {
			if option.CardID != cardID {
				continue
			}
			selection, err := strconv.Atoi(option.ID)
			if err != nil {
				selection = optionIndex
			}
			selections = append(selections, selection)
			matched = true
			break
		}
		if !matched {
			return nil, fmt.Errorf("卡牌 %q 不在当前可选列表中", cardID)
		}
	}
	return selections, nil
}
