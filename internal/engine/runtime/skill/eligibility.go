// gameflow: 技能候选收集与可用性判定（与旧 collectSkillsForTiming / isSkillStillUsable 对齐）。

package skill

import (
	skillhandlers "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// Eligibility 负责技能候选收集与可用性检查。
type Eligibility struct {
	cat *Catalog
}

// NewEligibility 创建检查器。
func NewEligibility(cat *Catalog) *Eligibility {
	return &Eligibility{cat: cat}
}

// CollectCandidates 收集指定玩家在指定时机下可触发的技能（含场上 Effect 伪技能）。
func (e *Eligibility) CollectCandidates(
	player *model.Player,
	timing model.FlowTiming,
	ctx *model.Context,
	currentRole model.SkillRole,
) []model.SkillDefinition {
	if player == nil || player.Character == nil {
		return nil
	}
	var skillBatch []model.SkillDefinition

	for _, sk := range player.Character.Skills {
		if ctx != nil && ctx.Timing == model.TimingStartup {
			if sk.Type != model.SkillTypeStartup {
				continue
			}
		} else if sk.Type == model.SkillTypeStartup {
			continue
		}

		if !skillMatchesTiming(sk, timing) {
			continue
		}

		if sk.RequiredRole != model.RoleAny && currentRole != model.RoleAny {
			if sk.RequiredRole != currentRole {
				continue
			}
		}

		if sk.Type == model.SkillTypeAction {
			continue
		}

		if !CanPaySkillEnergyCost(player, sk.CostGem, sk.CostCrystal) {
			continue
		}
		if sk.CostCoverCards > 0 {
			if len(player.GetCoverCards()) < sk.CostCoverCards {
				continue
			}
		}

		if model.ContainsSkillTag(sk.Tags, model.TagTurnLimit) {
			if count, exists := player.TurnState.UsedSkillCounts[sk.ID]; exists && count > 0 {
				continue
			}
		}

		if !e.uniqueSkillCardMatches(player, sk, ctx) {
			continue
		}

		handler := skillhandlers.GetHandler(ResolveHandlerID(sk))
		if handler == nil {
			continue
		}

		if !handler.CanUse(ctx) {
			continue
		}

		skillBatch = append(skillBatch, sk)
	}

	if ctx != nil && ctx.Timing == model.TimingStartup {
		return skillBatch
	}

	for _, fc := range player.Field {
		if fc.Mode != model.FieldEffect || fc.Locked {
			continue
		}

		handlerID := model.GetHandlerIDByEffect(fc.Effect)
		if handlerID == "" {
			continue
		}

		handler := skillhandlers.GetHandler(handlerID)
		if handler == nil {
			continue
		}

		if handler.CanUse(ctx) {
			fieldSkill := model.SkillDefinition{
				ID:           handlerID,
				Title:        fc.Card.Name,
				Type:         model.SkillTypePassive,
				ResponseType: model.ResponseSilent,
				LogicHandler: handlerID,
				Timings:      []model.FlowTiming{timing},
			}
			skillBatch = append(skillBatch, fieldSkill)
		}
	}

	return skillBatch
}

func skillMatchesTiming(skill model.SkillDefinition, timing model.FlowTiming) bool {
	if timing == model.TimingStartup && skill.Type == model.SkillTypeStartup {
		return true
	}
	return skill.HasTiming(timing)
}

func (e *Eligibility) uniqueSkillCardMatches(player *model.Player, skill model.SkillDefinition, ctx *model.Context) bool {
	if !model.ContainsSkillTag(skill.Tags, model.TagUnique) {
		return true
	}
	if player == nil || player.Character == nil || ctx == nil || ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	return ctx.EventCtx.Card.MatchExclusive(player.Character.ID, skill.Title)
}

// IsStillUsable 检查技能是否仍然可用（用于响应链剩余技能过滤）。
func (e *Eligibility) IsStillUsable(skillID string, user *model.Player, ctx *model.Context) bool {
	if e.cat == nil {
		return false
	}
	skillDef := e.cat.FindCharacterSkillOnPlayer(user, skillID)
	if skillDef == nil {
		return false
	}

	if !CanPaySkillEnergyCost(user, skillDef.CostGem, skillDef.CostCrystal) {
		return false
	}
	if !e.uniqueSkillCardMatches(user, *skillDef, ctx) {
		return false
	}

	handler := skillhandlers.GetHandler(ResolveHandlerID(*skillDef))
	if handler == nil {
		return false
	}

	return handler.CanUse(ctx)
}

// FilterRemainingUsable 除去当前技能后，中断列表中仍可用且不互斥的技能 ID。
func (e *Eligibility) FilterRemainingUsable(
	currentSkillID string,
	player *model.Player,
	ctx *model.Context,
	interruptSkillIDs []string,
) []string {
	var remaining []string
	for _, sid := range interruptSkillIDs {
		if sid == currentSkillID {
			continue
		}
		if MutuallyExclusiveResponseSkill(currentSkillID, sid) {
			continue
		}
		if e.IsStillUsable(sid, player, ctx) {
			remaining = append(remaining, sid)
		}
	}
	return remaining
}
