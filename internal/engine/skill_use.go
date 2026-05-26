// gameflow: 玩家主动发动技能的对外入口（与校验、消耗衔接）。

package engine

import (
	"fmt"

	skillrt "starcup-engine/internal/engine/runtime/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

type skillUseRequest struct {
	engine                *GameEngine
	player                *model.Player
	skillDef              *model.SkillDefinition
	policy                SkillPolicy
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

func (use *skillUseRequest) policyContext() types.PolicyContext {
	if use == nil {
		return types.PolicyContext{}
	}
	ctx := types.PolicyContext{
		SkillID:          use.skillID,
		RequiredDiscards: use.requiredDiscards,
		DiscardedCards:   append([]model.Card{}, use.discardedCards...),
		TargetIDs:        append([]string{}, use.targetIDs...),
	}
	if use.player != nil {
		ctx.PlayerID = use.player.ID
	}
	if use.skillDef != nil {
		ctx.SkillDef = *use.skillDef
	}
	for _, target := range use.actualTargets {
		if target != nil {
			ctx.ActualTargetIDs = append(ctx.ActualTargetIDs, target.ID)
		}
	}
	return ctx
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
	e.publishSkillDeclared(use)
	if err := e.consumeSkillInputs(use); err != nil {
		return err
	}
	if err := e.consumeSkillEnergyCost(use); err != nil {
		return err
	}

	use.player.TurnState.UsedSkillCounts[skillID]++

	// 弃牌后如果产生了 PendingDamage（如封印触发），
	// 记录待恢复技能状态，让 Drive 先结算伤害再恢复 handler.Execute()。
	if len(e.State.PendingDamageQueue) > 0 {
		discardedCards := make([]model.Card, len(use.discardedCards))
		copy(discardedCards, use.discardedCards)
		e.skillResume = &skillResumeState{
			playerID:       playerID,
			skillID:        skillID,
			targetIDs:      use.resolvedTargetIDs(),
			discardedCards: discardedCards,
		}
		e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction)
		return nil
	}

	if err := e.executeSkillFlow(use); err != nil {
		return err
	}
	if err := e.finishSkillUse(use); err != nil {
		return err
	}

	// 如果 Execute() 推入了中断，立即通知前端
	// 因为 Drive() 在检测到 PendingInterrupt 时会直接返回而不通知
	if e.State.PendingInterrupt != nil {
		e.notifyInterruptPrompt()
	}
	return nil
}

func (e *GameEngine) publishSkillDeclared(use *skillUseRequest) {
	if e == nil || use == nil || use.player == nil || use.skillDef == nil {
		return
	}
	if e.narrativeTrace == nil || e.narrativeTrace.actionID == "" || e.narrativeTrace.actionActor != use.player.ID {
		e.beginNarrativeAction("skill", use.player.ID)
	}
	e.publishTimelineMarker(model.TimelineMarkerPayload{
		PlayerID:      use.player.ID,
		PlayerName:    use.player.Name,
		ActionType:    "skill",
		SkillID:       use.skillDef.ID,
		SkillName:     use.skillDef.Title,
		EffectText:    use.skillDef.Description,
		TargetIDs:     use.resolvedTargetIDs(),
		NarrativeKind: "skill_declared",
		VisualKind:    "skill_token",
		SkillPhase:    "declared",
		Timing:        "skill.declared",
	})
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

	skillDef := skillrt.FindCharacterSkill(player.Character, skillID)
	if skillDef == nil {
		return nil, fmt.Errorf("skill %s not found for character %s", skillID, player.Character.ID)
	}

	policy := resolveSkillUsePolicy(skillID)
	requiredDiscards := skillDef.CostDiscards
	if policy.ResolveDiscardCount != nil {
		requiredDiscards = policy.ResolveDiscardCount(types.PolicyContext{
			SkillID:          skillID,
			PlayerID:         player.ID,
			SkillDef:         *skillDef,
			RequiredDiscards: requiredDiscards,
			TargetIDs:        append([]string{}, targetIDs...),
		})
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
