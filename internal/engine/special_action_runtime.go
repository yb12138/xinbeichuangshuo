package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type specialActionOverride func(e *GameEngine, player *model.Player, actionType model.ActionType) (bool, error)
type specialActionPostHook func(e *GameEngine, player *model.Player, actionType model.ActionType)

var specialActionOverrides = []specialActionOverride{
	specialActionAdventurerUndergroundLawOverride,
}

var specialActionPostHooks = []specialActionPostHook{
	specialActionHolyBowHolyGloryExitHook,
}

func (e *GameEngine) executeSpecialActionWithRuntime(player *model.Player, actionType model.ActionType) error {
	for _, override := range specialActionOverrides {
		if override == nil {
			continue
		}
		handled, err := override(e, player, actionType)
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	}
	return e.executeSpecialAction(player, actionType)
}

func (e *GameEngine) runPostSpecialActionRuntime(player *model.Player, actionType model.ActionType) {
	for _, hook := range specialActionPostHooks {
		if hook != nil {
			hook(e, player, actionType)
		}
	}
}

func specialActionAdventurerUndergroundLawOverride(e *GameEngine, player *model.Player, actionType model.ActionType) (bool, error) {
	if e == nil || player == nil || actionType != model.ActionBuy || !e.playerHasSkill(player, "adventurer_underground_law") {
		return false, nil
	}
	e.resolveAdventurerUndergroundLaw(player)
	return true, nil
}

func specialActionHolyBowHolyGloryExitHook(e *GameEngine, player *model.Player, _ model.ActionType) {
	if e == nil || player == nil || !e.isHolyBow(player) || !hasHolyBowHolyGloryForm(player) {
		return
	}
	beforePoses := e.snapshotPlayerPoses()
	leaveHolyBowHolyGloryForm(player)
	e.Heal(player.ID, 1)
	e.Log(fmt.Sprintf("%s 在圣煌形态下执行特殊行动，脱离圣煌形态并获得1点治疗", player.Name))
	e.dispatchOrientationChanges(beforePoses)
}
