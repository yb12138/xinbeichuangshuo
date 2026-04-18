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

func regSpecSingle(reg *choicert.SpecRegistry, typ string,
	build func(*GameEngine, string, string, *model.Player, map[string]any) *model.Prompt,
	sel func(*GameEngine, string, int, map[string]any) (bool, error),
) {
	reg.Register(&choicert.ChoiceSpec{
		Type: typ,
		BuildPrompt: func(h choicert.Host, ct, pid string, pl *model.Player, data map[string]any) *model.Prompt {
			ge := engFromHost(h)
			if ge == nil {
				return nil
			}
			return build(ge, ct, pid, pl, data)
		},
		OnSelect: func(h choicert.Host, pid string, idx int, ctx map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return sel(ge, pid, idx, ctx)
		},
	})
}

func regSpecSingleAndMulti(reg *choicert.SpecRegistry, typ string,
	build func(*GameEngine, string, string, *model.Player, map[string]any) *model.Prompt,
	sel func(*GameEngine, string, int, map[string]any) (bool, error),
	multi func(*GameEngine, string, []int) error,
) {
	reg.Register(&choicert.ChoiceSpec{
		Type: typ,
		BuildPrompt: func(h choicert.Host, ct, pid string, pl *model.Player, data map[string]any) *model.Prompt {
			ge := engFromHost(h)
			if ge == nil {
				return nil
			}
			return build(ge, ct, pid, pl, data)
		},
		OnSelect: func(h choicert.Host, pid string, idx int, ctx map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return sel(ge, pid, idx, ctx)
		},
		OnMultiSelect: func(h choicert.Host, pid string, sel []int, _ map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return true, multi(ge, pid, sel)
		},
	})
}

func regSpecSingleAndMultiAndCancel(reg *choicert.SpecRegistry, typ string,
	build func(*GameEngine, string, string, *model.Player, map[string]any) *model.Prompt,
	sel func(*GameEngine, string, int, map[string]any) (bool, error),
	multi func(*GameEngine, string, []int) error,
	cancel func(*GameEngine, string) error,
) {
	reg.Register(&choicert.ChoiceSpec{
		Type: typ,
		BuildPrompt: func(h choicert.Host, ct, pid string, pl *model.Player, data map[string]any) *model.Prompt {
			ge := engFromHost(h)
			if ge == nil {
				return nil
			}
			return build(ge, ct, pid, pl, data)
		},
		OnSelect: func(h choicert.Host, pid string, idx int, ctx map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return sel(ge, pid, idx, ctx)
		},
		OnMultiSelect: func(h choicert.Host, pid string, sel []int, _ map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return true, multi(ge, pid, sel)
		},
		OnCancel: func(h choicert.Host, pid string, _ map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return true, cancel(ge, pid)
		},
	})
}

func regSpecExtractChoice(reg *choicert.SpecRegistry,
	build func(*GameEngine, string, string, *model.Player, map[string]any) *model.Prompt,
	multi func(*GameEngine, string, []int) error,
	cancel func(*GameEngine, string) error,
) {
	reg.Register(&choicert.ChoiceSpec{
		Type: "extract",
		BuildPrompt: func(h choicert.Host, ct, pid string, pl *model.Player, data map[string]any) *model.Prompt {
			ge := engFromHost(h)
			if ge == nil {
				return nil
			}
			return build(ge, ct, pid, pl, data)
		},
		OnSelect: func(h choicert.Host, pid string, idx int, _ map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return true, multi(ge, pid, []int{idx})
		},
		OnMultiSelect: func(h choicert.Host, pid string, sel []int, _ map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return true, multi(ge, pid, sel)
		},
		OnCancel: func(h choicert.Host, pid string, _ map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return true, cancel(ge, pid)
		},
	})
}

func regSpecSingleAndCancel(reg *choicert.SpecRegistry, typ string,
	build func(*GameEngine, string, string, *model.Player, map[string]any) *model.Prompt,
	sel func(*GameEngine, string, int, map[string]any) (bool, error),
	cancel func(*GameEngine, string) error,
) {
	reg.Register(&choicert.ChoiceSpec{
		Type: typ,
		BuildPrompt: func(h choicert.Host, ct, pid string, pl *model.Player, data map[string]any) *model.Prompt {
			ge := engFromHost(h)
			if ge == nil {
				return nil
			}
			return build(ge, ct, pid, pl, data)
		},
		OnSelect: func(h choicert.Host, pid string, idx int, ctx map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return sel(ge, pid, idx, ctx)
		},
		OnCancel: func(h choicert.Host, pid string, _ map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return true, cancel(ge, pid)
		},
	})
}

func bootstrapChoiceSpecs(e *GameEngine) {
	if e == nil || e.choiceEngine == nil {
		return
	}
	reg := e.choiceEngine.Registry()

	regSpecSingleAndMultiAndCancel(reg, choiceTypeSystemDiscardCards, (*GameEngine).buildSystemChoicePrompt, func(ge *GameEngine, pid string, idx int, _ map[string]any) (bool, error) {
		return true, ge.handleSystemDiscardChoiceSelections(pid, []int{idx})
	}, (*GameEngine).handleSystemDiscardChoiceSelections, func(ge *GameEngine, pid string) error {
		if ge.State == nil || ge.State.PendingInterrupt == nil {
			return fmt.Errorf("当前没有待处理的弃牌操作")
		}
		ctxData, _ := ge.State.PendingInterrupt.Context.(map[string]interface{})
		return ge.cancelSystemDiscardChoice(pid, ctxData)
	})

	regSpecExtractChoice(reg, (*GameEngine).buildSystemChoicePrompt, (*GameEngine).handleExtractChoiceSelections, (*GameEngine).cancelExtractChoice)

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
