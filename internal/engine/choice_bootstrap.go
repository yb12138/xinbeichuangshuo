// gameflow: 全部 InterruptChoice 类型以 runtime/choice.ChoiceSpec 单点注册（bootstrapChoiceSpecs 在 NewGameEngine 中调用）。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
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
		Type:                typ,
		AutoConsume:         p.autoConsume,
		ConsumesInterrupt:   p.consumes,
		SequentialRemaining: nil,
		BuildPrompt:         wrapChoiceBuild(p.build),
		OnSelect:            wrapChoiceSelect(p.sel),
		OnMultiSelect:       wrapChoiceMulti(p.multi),
		OnCancel:            wrapChoiceCancel(p.cancel),
		AfterConsume:        wrapChoiceAfter(p.after),
	})
}

func registerRoleChoiceSpec(reg *choicert.SpecRegistry, roleID string, spec engineplayer.ChoiceSpec) {
	if reg == nil || roleID == "" || spec.ChoiceType == "" {
		return
	}
	var multi func(choicert.Host, string, []int, map[string]any) (bool, error)
	if spec.HandleMultiSelect != nil {
		multi = func(h choicert.Host, playerID string, selections []int, ctx map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return spec.HandleMultiSelect(NewRoleChoiceRuntime(ge), playerID, selections, choiceCtxAsInterfaceMap(ctx))
		}
	} else if spec.SequentialRemaining != nil {
		multi = func(h choicert.Host, playerID string, selections []int, _ map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return true, ge.runSequentialChoiceSelections(playerID, spec.ChoiceType, selections)
		}
	}
	reg.Register(&choicert.ChoiceSpec{
		Type: spec.ChoiceType,
		BuildPrompt: func(h choicert.Host, choiceType, playerID string, player *model.Player, data map[string]any) *model.Prompt {
			ge := engFromHost(h)
			if ge == nil {
				return nil
			}
			return ge.buildRoleChoicePrompt(roleID, choiceType, playerID, player, choiceCtxAsInterfaceMap(data))
		},
		OnSelect: func(h choicert.Host, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			return ge.handleRoleChoiceInput(roleID, playerID, selectionIndex, choiceCtxAsInterfaceMap(ctxData))
		},
		OnMultiSelect: multi,
		OnCancel: func(h choicert.Host, playerID string, ctxData map[string]any) (bool, error) {
			ge := engFromHost(h)
			if ge == nil {
				return false, fmt.Errorf("choice: engine bridge unavailable")
			}
			handled, err := ge.handleRoleChoiceCancel(roleID, playerID, choiceCtxAsInterfaceMap(ctxData))
			if err != nil || handled {
				return handled, err
			}
			return ge.cancelPromptFlowChoice(playerID, ctxData)
		},
		SequentialRemaining: func(ctxData map[string]any) (int, bool) {
			if spec.SequentialRemaining == nil {
				return 0, false
			}
			return spec.SequentialRemaining(choiceCtxAsInterfaceMap(ctxData))
		},
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

	registerChoiceSpec(reg, "extract", catalogSpecPlan{
		autoConsume: true,
		build:       (*GameEngine).buildSystemChoicePrompt,
		sel: func(ge *GameEngine, pid string, idx int, _ map[string]any) (bool, error) {
			return true, ge.HandleExtractChoiceSelections(pid, []int{idx})
		},
		multi: func(ge *GameEngine, pid string, sel []int) error {
			return ge.HandleExtractChoiceSelections(pid, sel)
		},
		cancel: func(ge *GameEngine, pid string, _ map[string]any) (bool, error) {
			return true, ge.cancelExtractChoice(pid)
		},
		after: func(ge *GameEngine, _ map[string]any) { ge.enterTurnEndStage() },
	})

	e.bootstrapRoleChoiceSpecs(reg)
	e.bootstrapChoiceSpecsFromCatalog(reg)
}

func (e *GameEngine) bootstrapRoleChoiceSpecs(reg *choicert.SpecRegistry) {
	if reg == nil || roleRegistry == nil {
		return
	}
	for _, entry := range roleRegistry.Entries() {
		for _, spec := range entry.ChoiceSpecs {
			if spec.ChoiceType != "" && !reg.Has(spec.ChoiceType) {
				registerRoleChoiceSpec(reg, entry.ID, spec)
			}
		}
		for choiceType, route := range entry.ChoiceRoutes() {
			if choiceType == "" || reg.Has(choiceType) || route.Kind != ChoiceRouteKindRole {
				continue
			}
			roleID := route.Role
			if roleID == "" {
				roleID = entry.ID
			}
			if spec, ok := entry.ChoiceSpecFor(choiceType); ok {
				registerRoleChoiceSpec(reg, roleID, spec)
				continue
			}
			registerRoleChoiceSpec(reg, roleID, engineplayer.ChoiceSpec{ChoiceType: choiceType})
		}
	}
}
