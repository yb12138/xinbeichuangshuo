package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type actionSelectionOptionState struct {
	currentPID               string
	currentExtraAction       string
	isRestrictedExtraAction  bool
	canMagicAction           bool
	canMagicSkillAction      bool
	hasRestrictedExtraAction bool
	hasArbiterForcedDoomsday bool
	hasHeroTaunt             bool
	tauntSourceID            string
	tauntSourceName          string
	tauntRequiresSkip        bool
	hasFighterHundredDragon  bool
	promptChoiceType         string
	promptSkillID            string
	validOptions             []model.PromptOption
	specialOptions           []model.PromptOption
	promptMessage            string
}

type actionSelectionValidationResult struct {
	consumeHeroTauntOnAttack bool
	handled                  bool
}

func (e *GameEngine) buildActionSelectionOptions(currentPID string, player *model.Player) actionSelectionOptionState {
	state := actionSelectionOptionState{
		currentPID: currentPID,
	}
	if player == nil {
		return state
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}

	state.currentExtraAction = player.TurnState.CurrentExtraAction
	state.isRestrictedExtraAction = state.currentExtraAction == "Attack" || state.currentExtraAction == "Magic"
	state.canMagicAction = e.canCastMagicInAction(player)
	state.canMagicSkillAction = e.hasUsableActionSkillForExtraMagic(player)
	state.hasRestrictedExtraAction = true
	if state.isRestrictedExtraAction {
		state.hasRestrictedExtraAction = e.checkExtraActionCards(player, state.currentExtraAction, player.TurnState.CurrentExtraElement)
	}

	e.applyActionSelectionOptionPolicies(player, &state)
	e.appendBaseActionSelectionOptions(player, &state)
	e.finalizeActionSelectionPromptState(player, &state)
	return state
}

func (e *GameEngine) applyActionSelectionOptionPolicies(player *model.Player, state *actionSelectionOptionState) {
	for _, policy := range actionSelectionOptionPolicies {
		if policy != nil {
			policy(e, player, state)
		}
	}
}

func (e *GameEngine) appendBaseActionSelectionOptions(player *model.Player, state *actionSelectionOptionState) {
	if player == nil || state == nil {
		return
	}

	switch state.currentExtraAction {
	case "Attack":
		if state.hasRestrictedExtraAction {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "attack", Label: "攻击"})
		}
	case "Magic":
		if state.hasRestrictedExtraAction {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "magic", Label: "法术"})
		}
	default:
		if state.hasArbiterForcedDoomsday {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "magic", Label: "法术（仅限末日审判）"})
		} else if state.hasFighterHundredDragon {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "attack", Label: "攻击（百式幻龙拳）"})
		} else if state.hasHeroTaunt {
			if !state.tauntRequiresSkip {
				state.validOptions = append(state.validOptions, model.PromptOption{ID: "attack", Label: "攻击（受挑衅约束）"})
			}
		} else {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "attack", Label: "攻击"})
			if state.canMagicAction || state.canMagicSkillAction {
				state.validOptions = append(state.validOptions, model.PromptOption{ID: "magic", Label: "法术"})
			}
		}
	}

	if !state.hasHeroTaunt && !state.hasArbiterForcedDoomsday && !state.hasFighterHundredDragon && !state.isRestrictedExtraAction && !e.hasPerformedStartupThisTurn(player) {
		maxHand := e.GetMaxHand(player)
		canBuyOrSynth := len(player.Hand)+3 <= maxHand

		if canBuyOrSynth {
			state.specialOptions = append(state.specialOptions, model.PromptOption{ID: "buy", Label: "购买"})
		}

		var totalStones int
		if player.Camp == model.RedCamp {
			totalStones = e.State.RedGems + e.State.RedCrystals
		} else {
			totalStones = e.State.BlueGems + e.State.BlueCrystals
		}
		if canBuyOrSynth && totalStones >= 3 {
			state.specialOptions = append(state.specialOptions, model.PromptOption{ID: "synthesize", Label: "合成"})
		}

		currentEnergy := player.Gem + player.Crystal
		if totalStones > 0 && currentEnergy < 3 {
			state.specialOptions = append(state.specialOptions, model.PromptOption{ID: "extract", Label: "提炼"})
		}

		if len(state.specialOptions) > 0 {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "special", Label: "特殊"})
		}
	}

	if !state.isRestrictedExtraAction {
		hasAttackCard := false
		hasMagicCard := false
		for idx := 0; idx < playableCardCount(player); idx++ {
			card, _, _, ok := getPlayableCardByIndex(player, idx)
			if !ok {
				continue
			}
			if card.Type == model.CardTypeAttack {
				hasAttackCard = true
			}
			if card.Type == model.CardTypeMagic && state.canMagicAction {
				hasMagicCard = true
			}
		}
		canNormalAction := hasAttackCard || hasMagicCard || state.canMagicSkillAction
		if state.hasArbiterForcedDoomsday {
			canNormalAction = state.canMagicSkillAction
		} else if state.hasFighterHundredDragon {
			canNormalAction = hasAttackCard
		}
		if state.hasHeroTaunt {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "cannot_act", Label: "跳过行动（移除挑衅）"})
			return
		}
		if !canNormalAction {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "cannot_act", Label: "无法行动（展示手牌）"})
		}
	} else if !state.hasRestrictedExtraAction {
		state.validOptions = append(state.validOptions, model.PromptOption{ID: "cannot_act", Label: "跳过额外行动"})
	}
}

func (e *GameEngine) finalizeActionSelectionPromptState(player *model.Player, state *actionSelectionOptionState) {
	if player == nil || state == nil {
		return
	}
	state.promptMessage = "请选择行动类型"
	if state.currentExtraAction == "Attack" {
		state.promptMessage = "当前为额外攻击行动，仅可执行攻击。请选择行动类型"
	} else if state.currentExtraAction == "Magic" {
		state.promptMessage = "当前为额外法术行动，仅可执行法术。请选择行动类型"
	} else if state.hasArbiterForcedDoomsday {
		state.promptMessage = "你的审判已达上限：本行动阶段必须发动【末日审判】。"
	} else if state.hasFighterHundredDragon {
		if locked := e.fighterLockedTarget(player); locked != nil {
			state.promptMessage = fmt.Sprintf("你处于【百式幻龙拳】状态：本行动阶段只能主动攻击 %s；若本行动阶段结束仍处于该形态，则自动转正。", model.GetPlayerDisplayName(locked))
		} else {
			state.promptMessage = "你处于【百式幻龙拳】状态：本行动阶段只能主动攻击已锁定目标；若本行动阶段结束仍处于该形态，则自动转正。"
		}
	} else if state.hasHeroTaunt {
		if state.tauntRequiresSkip {
			state.promptMessage = fmt.Sprintf("你受到【挑衅】影响：本次行动阶段必须主动攻击 %s，但你没有攻击牌，只能选择跳过行动并移除此牌。", state.tauntSourceName)
		} else {
			state.promptMessage = fmt.Sprintf("你受到【挑衅】影响：本次行动阶段必须且只能主动攻击 %s，或选择跳过行动并移除此牌。", state.tauntSourceName)
		}
	}
	if state.isRestrictedExtraAction && !state.hasRestrictedExtraAction {
		state.promptMessage = "当前为额外行动阶段，但你没有满足约束的可执行动作。可选择跳过本次额外行动。"
	}
}

func (e *GameEngine) validateActionSelectionPolicies(player *model.Player, act model.PlayerAction) (actionSelectionValidationResult, error) {
	var result actionSelectionValidationResult
	for _, policy := range actionSelectionValidationPolicies {
		if policy == nil {
			continue
		}
		if err := policy(e, player, act, &result); err != nil {
			return result, err
		}
	}
	return result, nil
}
