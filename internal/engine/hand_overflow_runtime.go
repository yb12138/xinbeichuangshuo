// gameflow: 爆牌相关运行时辅助。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"

	"starcup-engine/internal/model"
)

type handOverflowContext struct {
	isMagic                 bool
	fromDamageDraw          bool
	noMoraleLoss            bool
	stayInTurn              bool
	isDamageResolution      bool
	overflowMoraleLossFixed int
	damageSourceID          any
	damageSourceSkillID     any
	drawResumePoint         interface{}
}

func (e *GameEngine) CheckHandLimitCtx(player *model.Player, ctx *model.Context) {
	if player == nil {
		return
	}
	if ctx != nil && ctx.Flags["preventOverflow"] {
		e.Log(fmt.Sprintf("[System] %s 的本次摸牌忽略手牌上限检查", player.Name))
		return
	}

	overflowCtx := e.buildHandOverflowContext(ctx)
	over := len(player.Hand) - e.GetMaxHand(player)
	if over > 0 {
		e.pushHandOverflowDiscardInterrupt(player, over, overflowCtx)
		e.Log(fmt.Sprintf("[System] %s 手牌超出上限 %d 张！需要选择 %d 张牌丢弃", player.Name, len(player.Hand), over))
		return
	}
}

func (e *GameEngine) buildHandOverflowContext(ctx *model.Context) handOverflowContext {
	result := handOverflowContext{
		isDamageResolution: e.isDamageResolutionActive(),
	}
	if ctx == nil {
		return result
	}

	result.isMagic = ctx.Flags["IsMagicDamage"]
	result.fromDamageDraw = ctx.Flags["FromDamageDraw"]
	result.noMoraleLoss = ctx.Flags["NoMoraleLoss"]
	result.stayInTurn = ctx.Flags["StayInTurn"]

	if ctx.Selections != nil {
		result.overflowMoraleLossFixed = runtimeutil.ToIntContextValue(ctx.Selections["overflow_morale_loss_fixed"])
		result.damageSourceID = ctx.Selections["damage_source_id"]
		result.damageSourceSkillID = ctx.Selections["damage_source_skill_id"]
		if point, ok := choiceResumePointValue(ctx.Selections["draw_resume_phase"]); ok {
			result.drawResumePoint = point
		}
	}
	return result
}

func (e *GameEngine) pushHandOverflowDiscardInterrupt(player *model.Player, discardCount int, overflowCtx handOverflowContext) {
	e.PushInterrupt(newDiscardChoiceInterrupt(player.ID, map[string]interface{}{
		"discard_count":              discardCount,
		"is_magic":                   overflowCtx.isMagic,
		"from_damage_draw":           overflowCtx.fromDamageDraw,
		"no_morale_loss":             overflowCtx.noMoraleLoss,
		"victim_id":                  player.ID,
		"stay_in_turn":               overflowCtx.stayInTurn,
		"is_damage_resolution":       overflowCtx.isDamageResolution,
		"overflow_morale_loss_fixed": overflowCtx.overflowMoraleLossFixed,
		"damage_source_id":           overflowCtx.damageSourceID,
		"damage_source_skill":        overflowCtx.damageSourceSkillID,
		"draw_resume_phase":          overflowCtx.drawResumePoint,
		"remaining_indices":          e.handOverflowSelectableIndices(player),
	}))
}

func (e *GameEngine) handOverflowSelectableIndices(player *model.Player) []int {
	if player == nil {
		return nil
	}
	indices := engineplayer.AllHandIndices(player)
	if len(e.State.ActionQueue) == 0 {
		return indices
	}
	current := e.State.ActionQueue[0]
	if current.SourceID != player.ID || current.UsesVirtualCard {
		return indices
	}
	lockedCardID := queuedActionCardID(&current)
	if lockedCardID == "" {
		return indices
	}
	filtered := make([]int, 0, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < len(player.Hand) && player.Hand[idx].ID == lockedCardID {
			continue
		}
		filtered = append(filtered, idx)
	}
	return filtered
}
