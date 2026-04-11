// gameflow: 玩家主动发动技能的对外入口（与校验、消耗衔接）。

package engine

import (
	"fmt"
	"starcup-engine/internal/model"
)

type skillUseRequest struct {
	engine                *GameEngine
	player                *model.Player
	skillDef              *model.SkillDefinition
	policy                skillUsePolicy
	skillID               string
	targetIDs             []string
	discardIndices        []int
	requiredDiscards      int
	discardedCards        []model.Card
	consumedExclusiveCard *model.Card
	target                *model.Player
	actualTargets         []*model.Player
}

func (use *skillUseRequest) resolvedTargetIDs() []string {
	ids := make([]string, 0, len(use.actualTargets))
	for _, target := range use.actualTargets {
		if target != nil {
			ids = append(ids, target.ID)
		}
	}
	return ids
}

// UseSkill 使用技能
func (e *GameEngine) UseSkill(playerID, skillID string, targetIDs []string, discardIndices []int) error {
	use, err := e.prepareSkillUse(playerID, skillID, targetIDs, discardIndices)
	if err != nil {
		return err
	}
	if pending, err := e.maybeRequestSkillDiscardSelection(use); pending || err != nil {
		return err
	}
	if err := e.validateSkillDiscardSelection(use); err != nil {
		return err
	}
	if err := e.validateSkillActivation(use); err != nil {
		return err
	}
	if err := e.consumeSkillCoverCost(use); err != nil {
		return err
	}
	if err := e.ensureSkillEnergyCost(use); err != nil {
		return err
	}
	if err := e.resolveSkillTargets(use); err != nil {
		return err
	}
	if err := e.validateSkillFieldPlacement(use); err != nil {
		return err
	}
	if err := e.consumeSkillInputs(use); err != nil {
		return err
	}
	if err := e.consumeSkillEnergyCost(use); err != nil {
		return err
	}

	use.player.TurnState.UsedSkillCounts[skillID]++

	if err := e.executeSkillFlow(use); err != nil {
		return err
	}
	return e.finishSkillUse(use)
}

func (e *GameEngine) prepareSkillUse(playerID, skillID string, targetIDs []string, discardIndices []int) (*skillUseRequest, error) {
	player := e.State.Players[playerID]
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}
	if !player.IsActive {
		return nil, fmt.Errorf("not your turn")
	}
	if player.Character == nil {
		return nil, fmt.Errorf("no character assigned")
	}

	skillDef := findCharacterSkill(player.Character, skillID)
	if skillDef == nil {
		return nil, fmt.Errorf("skill %s not found for character %s", skillID, player.Character.ID)
	}

	policy := resolveSkillUsePolicy(skillID)
	requiredDiscards := skillDef.CostDiscards
	if policy.resolveDiscardCount != nil {
		requiredDiscards = policy.resolveDiscardCount(player, skillDef)
	}

	return &skillUseRequest{
		engine:           e,
		player:           player,
		skillDef:         skillDef,
		policy:           policy,
		skillID:          skillID,
		targetIDs:        append([]string{}, targetIDs...),
		discardIndices:   append([]int{}, discardIndices...),
		requiredDiscards: requiredDiscards,
	}, nil
}

func findCharacterSkill(character *model.Character, skillID string) *model.SkillDefinition {
	if character == nil {
		return nil
	}
	for i := range character.Skills {
		if character.Skills[i].ID == skillID {
			return &character.Skills[i]
		}
	}
	return nil
}
