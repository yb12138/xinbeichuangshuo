// gameflow: 响应恢复断点标记（命中/未命中/承伤等）。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func choiceResumePointValue(raw interface{}) (interface{}, bool) {
	// 规则：恢复点必须是“可继续执行”的流程节点，None/空值不是合法恢复目标。
	switch value := raw.(type) {
	case model.TurnStage:
		if value != "" && model.IsKnownTurnStage(value) {
			return value, true
		}
	case model.CombatStage:
		if value != model.CombatStageNone && model.IsKnownCombatStage(value) {
			return value, true
		}
	case model.Subflow:
		if value != model.SubflowNone && model.IsKnownSubflow(value) {
			return value, true
		}
	}
	return nil, false
}

func hasChoiceResumePoint(raw interface{}) bool {
	_, ok := choiceResumePointValue(raw)
	return ok
}

func isChoiceResumeTurnStage(raw interface{}, stage model.TurnStage) bool {
	point, ok := choiceResumePointValue(raw)
	if !ok {
		return false
	}
	value, ok := point.(model.TurnStage)
	return ok && value == stage
}

func mustChoiceResumePoint(raw interface{}, label string) interface{} {
	// 规则：可中断的流程必须明确声明恢复点；缺失时应立即失败，而不是猜测默认阶段。
	point, ok := choiceResumePointValue(raw)
	if !ok {
		panic(fmt.Sprintf("invalid %s resume point: %#v", label, raw))
	}
	return point
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

func parseChoiceResumeTurnStage(raw interface{}) model.TurnStage {
	return model.ParseResumePointTurnStage(raw)
}

func parseChoiceResumeCombatStage(raw interface{}) model.CombatStage {
	return model.ParseResumePointCombatStage(raw)
}

func parseChoiceResumeSubflow(raw interface{}) model.Subflow {
	return model.ParseResumePointSubflow(raw)
}

func (e *GameEngine) currentChoiceResumePoint() interface{} {
	if e == nil || e.State == nil {
		panic("currentChoiceResumePoint: engine/state is nil")
	}
	if e.State.Subflow != model.SubflowNone {
		if !model.IsKnownSubflow(e.State.Subflow) {
			panic(fmt.Sprintf("currentChoiceResumePoint: unknown subflow %q", e.State.Subflow))
		}
		return e.State.Subflow
	}
	if e.State.CombatStage != model.CombatStageNone {
		if !model.IsKnownCombatStage(e.State.CombatStage) {
			panic(fmt.Sprintf("currentChoiceResumePoint: unknown combat stage %q", e.State.CombatStage))
		}
		return e.State.CombatStage
	}
	if e.State.TurnStage == "" {
		return nil
	}
	if !model.IsKnownTurnStage(e.State.TurnStage) {
		panic(fmt.Sprintf("currentChoiceResumePoint: unknown turn stage %q", e.State.TurnStage))
	}
	return e.State.TurnStage
}

func (e *GameEngine) applyChoiceResumePoint(raw interface{}) bool {
	if e == nil || e.State == nil {
		return false
	}
	if raw == nil {
		return false
	}
	stage, combat, subflow := model.ParseResumePoint(raw)

	// 规则：一次恢复只落到一个确定节点；其余维度必须回到 None，避免“半恢复”状态污染后续结算。
	e.State.ReturnTurnStage = ""
	e.State.ReturnCombatStage = model.CombatStageNone
	e.State.ReturnSubflow = model.SubflowNone
	e.State.Subflow = subflow
	e.State.CombatStage = combat
	e.State.TurnStage = stage
	return true
}

func (e *GameEngine) setReturnPoint(raw interface{}) bool {
	if e == nil || e.State == nil {
		return false
	}
	if raw == nil {
		return false
	}
	stage, combat, subflow := model.ParseResumePoint(raw)
	e.State.ReturnTurnStage = stage
	e.State.ReturnCombatStage = combat
	e.State.ReturnSubflow = subflow
	return true
}

func (e *GameEngine) restoreReturnPoint() bool {
	if e == nil || e.State == nil {
		return false
	}
	if e.State.ReturnTurnStage == "" && e.State.ReturnCombatStage == model.CombatStageNone && e.State.ReturnSubflow == model.SubflowNone {
		return false
	}
	e.State.TurnStage = e.State.ReturnTurnStage
	e.State.CombatStage = e.State.ReturnCombatStage
	e.State.Subflow = e.State.ReturnSubflow
	e.State.ReturnTurnStage = ""
	e.State.ReturnCombatStage = model.CombatStageNone
	e.State.ReturnSubflow = model.SubflowNone
	return true
}
