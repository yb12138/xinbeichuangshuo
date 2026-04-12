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

	regSpecExtractChoice(reg, (*GameEngine).buildSystemChoicePrompt, (*GameEngine).handleExtractChoiceSelections, (*GameEngine).cancelExtractChoice)

	regSpecSingleAndCancel(reg, "hom_dual_echo_target", (*GameEngine).buildTargetChoicePrompt, (*GameEngine).handleWarHomunculusChoiceInput, (*GameEngine).cancelHomDualEchoChoice)

	regSpecSingleAndMulti(reg, "bt_reverse_branch2_pick", (*GameEngine).buildButterflyChoicePrompt, (*GameEngine).handleButterflyChoiceInput, (*GameEngine).handleButterflyReverseBranch2PickSelections)

	regSpecSingleAndMulti(reg, "bt_cocoon_overflow_discard", (*GameEngine).buildButterflyChoicePrompt, (*GameEngine).handleButterflyChoiceInput, (*GameEngine).handleButterflyCocoonOverflowSelections)

	regSpecSingleAndMulti(reg, "bp_curse_discard", (*GameEngine).buildBloodPriestessChoicePrompt, (*GameEngine).handleBloodPriestessChoiceInput, (*GameEngine).handleBloodCurseDiscardSelections)

	regSpecSingleAndMulti(reg, "ss_recall_pick", (*GameEngine).buildSoulSorcererChoicePrompt, (*GameEngine).handleSoulSorcererChoiceInput, (*GameEngine).handleSoulRecallSelections)

	e.bootstrapChoiceSpecsFromCatalog(reg)
}
