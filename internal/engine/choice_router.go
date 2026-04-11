// gameflow: 技能/系统选项输入按 choice_type 分派。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type choicePromptBuilder func(e *GameEngine, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt

type choiceSingleInputHandler func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error)

type choiceMultiInputHandler func(e *GameEngine, playerID string, selections []int) error

type choiceCancelHandler func(e *GameEngine, playerID string) error

var registeredChoicePromptBuilders = []choicePromptBuilder{
	(*GameEngine).buildSystemChoicePrompt,
	(*GameEngine).buildOnmyojiChoicePrompt,
	(*GameEngine).buildBeastSamuraiChoicePrompt,
	(*GameEngine).buildSageChoicePrompt,
	(*GameEngine).buildAdventurerChoicePrompt,
	(*GameEngine).buildPriestChoicePrompt,
	(*GameEngine).buildPrayerMasterChoicePrompt,
	(*GameEngine).buildCrimsonKnightChoicePrompt,
	(*GameEngine).buildWarHomunculusChoicePrompt,
	(*GameEngine).buildValkyrieChoicePrompt,
	(*GameEngine).buildElementalistChoicePrompt,
	(*GameEngine).buildElfArcherChoicePrompt,
	(*GameEngine).buildMagicBowChoicePrompt,
	(*GameEngine).buildSwordEmperorChoicePrompt,
	(*GameEngine).buildMagicLancerChoicePrompt,
	(*GameEngine).buildSoulSorcererChoicePrompt,
	(*GameEngine).buildMoonGoddessChoicePrompt,
	(*GameEngine).buildBloodPriestessChoicePrompt,
	(*GameEngine).buildButterflyChoicePrompt,
	(*GameEngine).buildSpiritCasterChoicePrompt,
	(*GameEngine).buildBardChoicePrompt,
	(*GameEngine).buildHolyBowChoicePrompt,
	(*GameEngine).buildHeroAssassinChoicePrompt,
	(*GameEngine).buildArbiterChoicePrompt,
	(*GameEngine).buildGuardianSupportChoicePrompt,
	(*GameEngine).buildHolyLancerChoicePrompt,
	(*GameEngine).buildSealerChoicePrompt,
	(*GameEngine).buildPlagueMageChoicePrompt,
	(*GameEngine).buildMagicSwordsmanChoicePrompt,
	(*GameEngine).buildCrimsonSwordSpiritChoicePrompt,
	(*GameEngine).buildBlazeWitchChoicePrompt,
	(*GameEngine).buildTargetChoicePrompt,
}

var registeredChoiceSingleInputHandlers = []choiceSingleInputHandler{
	(*GameEngine).handleSystemChoiceInput,
	func(e *GameEngine, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
		return e.handleOnmyojiChoiceInput(selectionIndex, ctxData)
	},
	func(e *GameEngine, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
		return e.handleBeastSamuraiChoiceInput(selectionIndex, ctxData)
	},
	(*GameEngine).handleSageChoiceInput,
	(*GameEngine).handleAdventurerChoiceInput,
	(*GameEngine).handlePriestChoiceInput,
	(*GameEngine).handlePrayerMasterChoiceInput,
	(*GameEngine).handleCrimsonKnightChoiceInput,
	(*GameEngine).handleWarHomunculusChoiceInput,
	(*GameEngine).handleValkyrieChoiceInput,
	(*GameEngine).handleElementalistChoiceInput,
	(*GameEngine).handleElfArcherChoiceInput,
	(*GameEngine).handleMagicBowChoiceInput,
	(*GameEngine).handleSwordEmperorChoiceInput,
	(*GameEngine).handleMagicLancerChoiceInput,
	(*GameEngine).handleFighterChoiceInput,
	(*GameEngine).handleSoulSorcererChoiceInput,
	(*GameEngine).handleMoonGoddessChoiceInput,
	(*GameEngine).handleBloodPriestessChoiceInput,
	(*GameEngine).handleButterflyChoiceInput,
	(*GameEngine).handleSpiritCasterChoiceInput,
	(*GameEngine).handleBardChoiceInput,
	(*GameEngine).handleHolyBowChoiceInput,
	(*GameEngine).handleHeroAssassinChoiceInput,
	(*GameEngine).handleArbiterChoiceInput,
	(*GameEngine).handleGuardianSupportChoiceInput,
	(*GameEngine).handleHolyLancerChoiceInput,
	(*GameEngine).handleSealerChoiceInput,
	(*GameEngine).handlePlagueMageChoiceInput,
	(*GameEngine).handleMagicSwordsmanChoiceInput,
	(*GameEngine).handleCrimsonSwordSpiritChoiceInput,
	(*GameEngine).handleBlazeWitchChoiceInput,
}

var registeredChoiceMultiInputHandlers = map[string]choiceMultiInputHandler{
	"extract":                    (*GameEngine).handleExtractChoiceSelections,
	"bp_curse_discard":           (*GameEngine).handleBloodCurseDiscardSelections,
	"ss_recall_pick":             (*GameEngine).handleSoulRecallSelections,
	"bt_cocoon_overflow_discard": (*GameEngine).handleButterflyCocoonOverflowSelections,
	"bt_reverse_branch2_pick":    (*GameEngine).handleButterflyReverseBranch2PickSelections,
}

var registeredChoiceCancelHandlers = map[string]choiceCancelHandler{
	"extract":              (*GameEngine).cancelExtractChoice,
	"hom_dual_echo_target": (*GameEngine).cancelHomDualEchoChoice,
}

func (e *GameEngine) buildRegisteredChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	for _, builder := range registeredChoicePromptBuilders {
		if prompt := builder(e, choiceType, playerID, player, data); prompt != nil {
			return prompt
		}
	}
	return nil
}

func (e *GameEngine) handleRegisteredChoiceInput(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	for _, handler := range registeredChoiceSingleInputHandlers {
		if handled, err := handler(e, playerID, selectionIndex, ctxData); handled || err != nil {
			return handled, err
		}
	}
	return false, nil
}

func (e *GameEngine) handleRegisteredChoiceMultiSelect(playerID, choiceType string, selections []int) (bool, error) {
	handler, ok := registeredChoiceMultiInputHandlers[choiceType]
	if !ok {
		return false, nil
	}
	return true, handler(e, playerID, selections)
}

func (e *GameEngine) handleRegisteredChoiceCancel(playerID, choiceType string) (bool, error) {
	handler, ok := registeredChoiceCancelHandlers[choiceType]
	if !ok {
		return false, nil
	}
	return true, handler(e, playerID)
}

func (e *GameEngine) handleLegacySequentialCardSelections(playerID string, selections []int) error {
	if len(selections) == 0 {
		return fmt.Errorf("请先选择手牌后再提交")
	}
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的选牌中断")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("选牌上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	needCount, supported := registeredSequentialCardChoiceRemainingCount(choiceType, ctxData)
	if !supported {
		return fmt.Errorf("当前选择类型不支持多选提交流程")
	}
	if needCount < 1 {
		needCount = 1
	}
	if len(selections) == 1 && needCount != 1 {
		return e.handleWeakChoiceInput(playerID, selections[0])
	}
	if len(selections) != needCount {
		return fmt.Errorf("需要选择 %d 张牌", needCount)
	}
	for _, idx := range selections {
		if err := e.handleWeakChoiceInput(playerID, idx); err != nil {
			return err
		}
	}
	return nil
}

func (e *GameEngine) handleWeakChoiceInput(playerID string, selectionIndex int) error {
	if e.State.PendingInterrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	choiceType, _ := ctxData["choice_type"].(string)

	if handled, err := e.handleRegisteredChoiceInput(playerID, selectionIndex, ctxData); handled || err != nil {
		return err
	}

	return fmt.Errorf("未知的选择类型: %s", choiceType)
}

func (e *GameEngine) handleChoiceSelectionInput(playerID string, selectionIndex int) error {
	return e.handleWeakChoiceInput(playerID, selectionIndex)
}

func (e *GameEngine) handleInterruptChoiceAction(act model.PlayerAction) error {
	if act.Type == model.CmdCancel {
		if data, ok := e.State.PendingInterrupt.Context.(map[string]interface{}); ok {
			if ct, _ := data["choice_type"].(string); ct != "" {
				if handled, err := e.handleRegisteredChoiceCancel(act.PlayerID, ct); handled || err != nil {
					return err
				}
			}
		}
	}
	if act.Type == model.CmdSelect {
		if data, ok := e.State.PendingInterrupt.Context.(map[string]interface{}); ok {
			if ct, _ := data["choice_type"].(string); ct != "" {
				if handled, err := e.handleRegisteredChoiceMultiSelect(act.PlayerID, ct, act.Selections); handled || err != nil {
					return err
				}
				if _, isLegacyCardMulti := registeredSequentialCardChoiceRemainingCount(ct, data); isLegacyCardMulti {
					return e.handleLegacySequentialCardSelections(act.PlayerID, act.Selections)
				}
			}
		}
		if len(act.Selections) != 1 {
			return fmt.Errorf("请选择一个选项")
		}
		return e.handleWeakChoiceInput(act.PlayerID, act.Selections[0])
	}
	return fmt.Errorf("当前中断类型不支持该指令")
}

func (e *GameEngine) cancelExtractChoice(playerID string) error {
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.enterActionExecutionStage()
	}
	if p := e.State.Players[playerID]; p != nil {
		e.Log("[System] " + p.Name + " 取消了提炼操作")
	} else {
		e.Log("[System] " + playerID + " 取消了提炼操作")
	}
	return nil
}

func (e *GameEngine) cancelHomDualEchoChoice(playerID string) error {
	e.PopInterrupt()
	if p := e.State.Players[playerID]; p != nil {
		e.Log("[System] " + p.Name + " 取消了 [双重回响] 的目标选择")
	} else {
		e.Log("[System] " + playerID + " 取消了 [双重回响] 的目标选择")
	}
	if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
		e.enterDamageResolution(nil)
	}
	return nil
}
