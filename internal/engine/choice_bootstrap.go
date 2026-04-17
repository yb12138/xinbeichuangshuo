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

	regSpecSingleAndMulti(reg, "bs_alert_source_discard", (*GameEngine).buildBeastSamuraiChoicePrompt, func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
		return true, ge.handleBeastSamuraiDiscardSelections(pid, []int{idx}, choiceCtxAsInterfaceMap(ctx))
	}, func(ge *GameEngine, pid string, sels []int) error {
		return ge.handleBeastSamuraiDiscardSelections(pid, sels, nil)
	})

	regSpecSingleAndMulti(reg, "bs_beast_return_self_discard", (*GameEngine).buildBeastSamuraiChoicePrompt, func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
		return true, ge.handleBeastSamuraiDiscardSelections(pid, []int{idx}, choiceCtxAsInterfaceMap(ctx))
	}, func(ge *GameEngine, pid string, sels []int) error {
		return ge.handleBeastSamuraiDiscardSelections(pid, sels, nil)
	})

	regSpecSingleAndMulti(reg, "bs_beast_return_source_discard", (*GameEngine).buildBeastSamuraiChoicePrompt, func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
		return true, ge.handleBeastSamuraiDiscardSelections(pid, []int{idx}, choiceCtxAsInterfaceMap(ctx))
	}, func(ge *GameEngine, pid string, sels []int) error {
		return ge.handleBeastSamuraiDiscardSelections(pid, sels, nil)
	})

	regSpecSingleAndMulti(reg, "bs_iaijutsu_style_discard", (*GameEngine).buildBeastSamuraiChoicePrompt, func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
		return true, ge.handleBeastSamuraiDiscardSelections(pid, []int{idx}, choiceCtxAsInterfaceMap(ctx))
	}, func(ge *GameEngine, pid string, sels []int) error {
		return ge.handleBeastSamuraiDiscardSelections(pid, sels, nil)
	})

	regSpecSingleAndMulti(reg, "bs_reversal_target_discard", (*GameEngine).buildBeastSamuraiChoicePrompt, func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
		return true, ge.handleBeastSamuraiDiscardSelections(pid, []int{idx}, choiceCtxAsInterfaceMap(ctx))
	}, func(ge *GameEngine, pid string, sels []int) error {
		return ge.handleBeastSamuraiDiscardSelections(pid, sels, nil)
	})

	regSpecExtractChoice(reg, (*GameEngine).buildSystemChoicePrompt, (*GameEngine).handleExtractChoiceSelections, (*GameEngine).cancelExtractChoice)

	regSpecSingleAndCancel(reg, "hom_dual_echo_target", (*GameEngine).buildTargetChoicePrompt, (*GameEngine).handleWarHomunculusChoiceInput, (*GameEngine).cancelHomDualEchoChoice)

	regSpecSingleAndMulti(reg, "bt_reverse_branch2_pick", (*GameEngine).buildButterflyChoicePrompt, (*GameEngine).handleButterflyChoiceInput, (*GameEngine).handleButterflyReverseBranch2PickSelections)

	regSpecSingleAndMulti(reg, "bt_cocoon_overflow_discard", (*GameEngine).buildButterflyChoicePrompt, (*GameEngine).handleButterflyChoiceInput, (*GameEngine).handleButterflyCocoonOverflowSelections)

	regSpecSingleAndMulti(reg, "bp_curse_discard", (*GameEngine).buildBloodPriestessChoicePrompt, (*GameEngine).handleBloodPriestessChoiceInput, (*GameEngine).handleBloodCurseDiscardSelections)

	regSpecSingleAndMulti(reg, "ss_recall_pick", (*GameEngine).buildSoulSorcererChoicePrompt, (*GameEngine).handleSoulSorcererChoiceInput, (*GameEngine).handleSoulRecallSelections)

	regSpecSingleAndMulti(reg, "adventurer_fraud_pick", (*GameEngine).buildAdventurerChoicePrompt, (*GameEngine).handleAdventurerChoiceInput, (*GameEngine).handleAdventurerFraudPickSelections)
	regSpecSingleAndCancel(reg, "valkyrie_heroic_discard_card", roleBuildPrompt("valkyrie"), roleSelect("valkyrie"), roleCancel("valkyrie"))
	regSpecSingleAndCancel(reg, "elf_elemental_shot_cost", roleBuildPrompt("elf_archer"), roleSelect("elf_archer"), roleCancel("elf_archer"))
	regSpecSingleAndCancel(reg, "plague_death_touch_element", roleBuildPrompt("plague_mage"), roleSelect("plague_mage"), roleCancel("plague_mage"))
	regSpecSingleAndCancel(reg, "plague_death_touch_x", roleBuildPrompt("plague_mage"), roleSelect("plague_mage"), roleCancel("plague_mage"))
	regSpecSingleAndCancel(reg, "plague_death_touch_y", roleBuildPrompt("plague_mage"), roleSelect("plague_mage"), roleCancel("plague_mage"))
	regSpecSingleAndMultiAndCancel(reg, "plague_death_touch_cards", roleBuildPrompt("plague_mage"), roleSelect("plague_mage"), func(ge *GameEngine, pid string, sels []int) error {
		return ge.runSequentialChoiceSelections(pid, "plague_death_touch_cards", sels)
	}, roleCancel("plague_mage"))
	regSpecSingleAndCancel(reg, "plague_death_touch_target", roleBuildPrompt("plague_mage"), roleSelect("plague_mage"), roleCancel("plague_mage"))

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
