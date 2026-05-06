// gameflow: 全部 InterruptChoice 类型以 runtime/choice.ChoiceSpec 单点注册（bootstrapChoiceSpecs 在 NewGameEngine 中调用）。

package engine

import (
	"fmt"

	choicert "starcup-engine/internal/engine/runtime/choice"
	"starcup-engine/internal/model"
)

func engFromHost(h choicert.Host) *GameEngine {
	b, ok := h.(*choiceHostBridge)
	if !ok || b == nil || b.e == nil {
		return nil
	}
	return b.e
}

func registerChoiceSpec(reg *choicert.SpecRegistry, typ string, p catalogSpecPlan) {
	reg.Register(&choicert.ChoiceSpec{
		Type:              typ,
		AutoConsume:       p.autoConsume,
		ConsumesInterrupt: p.consumes,
		BuildPrompt:       wrapChoiceBuild(p.build),
		OnSelect:          wrapChoiceSelect(p.sel),
		OnMultiSelect:     wrapChoiceMulti(p.multi),
		OnCancel:          wrapChoiceCancel(p.cancel),
		AfterConsume:      wrapChoiceAfter(p.after),
	})
}

func wrapChoiceBuild(build func(*GameEngine, string, string, *model.Player, map[string]any) *model.Prompt) func(choicert.Host, string, string, *model.Player, map[string]any) *model.Prompt {
	if build == nil {
		return nil
	}
	return func(h choicert.Host, ct, pid string, pl *model.Player, data map[string]any) *model.Prompt {
		ge := engFromHost(h)
		if ge == nil {
			return nil
		}
		return build(ge, ct, pid, pl, data)
	}
}

func wrapChoiceSelect(sel func(*GameEngine, string, int, map[string]any) (bool, error)) func(choicert.Host, string, int, map[string]any) (bool, error) {
	if sel == nil {
		return nil
	}
	return func(h choicert.Host, pid string, idx int, ctx map[string]any) (bool, error) {
		ge := engFromHost(h)
		if ge == nil {
			return false, fmt.Errorf("choice: engine bridge unavailable")
		}
		return sel(ge, pid, idx, ctx)
	}
}

func wrapChoiceMulti(multi func(*GameEngine, string, []int) error) func(choicert.Host, string, []int, map[string]any) (bool, error) {
	if multi == nil {
		return nil
	}
	return func(h choicert.Host, pid string, sel []int, _ map[string]any) (bool, error) {
		ge := engFromHost(h)
		if ge == nil {
			return false, fmt.Errorf("choice: engine bridge unavailable")
		}
		return true, multi(ge, pid, sel)
	}
}

func wrapChoiceCancel(cancel func(*GameEngine, string, map[string]any) (bool, error)) func(choicert.Host, string, map[string]any) (bool, error) {
	if cancel == nil {
		return nil
	}
	return func(h choicert.Host, pid string, ctx map[string]any) (bool, error) {
		ge := engFromHost(h)
		if ge == nil {
			return false, fmt.Errorf("choice: engine bridge unavailable")
		}
		return cancel(ge, pid, ctx)
	}
}

func wrapChoiceAfter(after func(*GameEngine, map[string]any)) func(choicert.Host, map[string]any) {
	if after == nil {
		return nil
	}
	return func(h choicert.Host, ctx map[string]any) {
		if ge := engFromHost(h); ge != nil {
			after(ge, ctx)
		}
	}
}

func bootstrapChoiceSpecs(e *GameEngine) {
	if e == nil || e.choiceEngine == nil {
		return
	}
	reg := e.choiceEngine.Registry()

	registerChoiceSpec(reg, choiceTypeSystemDiscardCards, catalogSpecPlan{
		build: (*GameEngine).buildSystemChoicePrompt,
		sel: func(ge *GameEngine, pid string, idx int, _ map[string]any) (bool, error) {
			return true, ge.handleSystemDiscardChoiceSelections(pid, []int{idx})
		},
		multi: (*GameEngine).handleSystemDiscardChoiceSelections,
		cancel: func(ge *GameEngine, pid string, _ map[string]any) (bool, error) {
			if ge.State == nil || ge.State.PendingInterrupt == nil {
				return false, fmt.Errorf("当前没有待处理的弃牌操作")
			}
			ctxData, _ := ge.State.PendingInterrupt.Context.(map[string]interface{})
			return true, ge.cancelSystemDiscardChoice(pid, ctxData)
		},
		consumes: func(ctx map[string]any) bool {
			if ctx == nil {
				return false
			}
			skillID, _ := ctx["skill_id"].(string)
			return skillID == ""
		},
		after: (*GameEngine).afterSystemDiscardChoice,
	})

	registerChoiceSpec(reg, "extract", catalogSpecPlan{
		autoConsume: true,
		build:       (*GameEngine).buildSystemChoicePrompt,
		sel: func(ge *GameEngine, pid string, idx int, _ map[string]any) (bool, error) {
			return true, ge.handleExtractChoiceSelections(pid, []int{idx})
		},
		multi: func(ge *GameEngine, pid string, sel []int) error {
			return ge.handleExtractChoiceSelections(pid, sel)
		},
		cancel: func(ge *GameEngine, pid string, _ map[string]any) (bool, error) {
			return true, ge.cancelExtractChoice(pid)
		},
		after: func(ge *GameEngine, _ map[string]any) { ge.enterTurnEndStage() },
	})

	e.bootstrapChoiceSpecsFromCatalog(reg)
}

// roleBuildPrompt creates a build-prompt callback that delegates to the player/<role>/choices.go via RoleEntry registry.
func roleBuildPrompt(roleID string) func(*GameEngine, string, string, *model.Player, map[string]any) *model.Prompt {
	return func(ge *GameEngine, ct, pid string, pl *model.Player, data map[string]any) *model.Prompt {
		return ge.buildRoleChoicePrompt(roleID, ct, pid, pl, choiceCtxAsInterfaceMap(data))
	}
}

// roleSelect creates a select callback that delegates to the player/<role>/choices.go via RoleEntry registry.
func roleSelect(roleID string) func(*GameEngine, string, int, map[string]any) (bool, error) {
	return func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
		return ge.handleRoleChoiceInput(roleID, pid, idx, choiceCtxAsInterfaceMap(ctx))
	}
}

// roleCancel creates a cancel callback that delegates to the player/<role>/choices.go via RoleEntry registry.
func roleCancel(roleID string) func(*GameEngine, string) error {
	return func(ge *GameEngine, pid string) error {
		_, err := roleCancelWithContext(roleID)(ge, pid, nil)
		return err
	}
}

func roleCancelWithContext(roleID string) func(*GameEngine, string, map[string]any) (bool, error) {
	return func(ge *GameEngine, pid string, ctx map[string]any) (bool, error) {
		return ge.handleRoleChoiceCancel(roleID, pid, choiceCtxAsInterfaceMap(ctx))
	}
}

// roleMultiSequential creates a multi-selection callback that processes selections sequentially via role router.
func roleMultiSequential(roleID string) func(*GameEngine, string, []int) error {
	return func(ge *GameEngine, pid string, sels []int) error {
		return ge.runSequentialChoiceSelections(pid, roleID, sels)
	}
}
