// gameflow: 士气下降与「伤害导致摸牌」等连锁。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// applyMoraleLossAfterTimingWindow 在 TimingBeforeMoraleLoss 窗口处理完毕后应用士气损失与联动效果。
func (e *GameEngine) applyMoraleLossAfterTimingWindow(victim *model.Player, moraleLoss int, isMagic bool, fromDamageDraw bool, overflowMoraleLossFixed int, discardedCards []model.Card, lossCtx *model.Context) int {
	if victim == nil {
		if len(discardedCards) > 0 {
			e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
		}
		return 0
	}

	finalLoss := moraleLoss
	if lossCtx != nil && lossCtx.EventCtx != nil && lossCtx.EventCtx.DamageVal != nil {
		finalLoss = *lossCtx.EventCtx.DamageVal
	}
	if finalLoss < 0 {
		finalLoss = 0
	}
	if overflowMoraleLossFixed > 0 && finalLoss > 0 {
		finalLoss = overflowMoraleLossFixed
	}

	finalLoss = e.applyCampMoraleLoss(victim.Camp, finalLoss)
	for _, entry := range roleRegistry.Entries() {
		if entry.AfterMoraleLossHook != nil {
			entry.AfterMoraleLossHook(e, victim, finalLoss, fromDamageDraw)
		}
	}
	e.dispatchRoleTimingHook(engineplayer.TimingOnMoraleLossApplied, engineplayer.TimingHookContext{
		TargetID:       victim.ID,
		IsMagicDamage:  isMagic,
		FromDamageDraw: fromDamageDraw,
		MoraleLoss:     finalLoss,
	})
	e.trackPlagueOutbreakMoraleDrop(lossCtx)

	if moraleLoss != finalLoss {
		e.Log(fmt.Sprintf("[System] 士气损失被抵御！原损失: %d, 实际损失: %d", moraleLoss, finalLoss))
	}

	e.finalizeMoraleLossDiscard(victim, discardedCards, lossCtx)
	return finalLoss
}

// trackPlagueOutbreakMoraleDrop 记录瘟疫爆发导致的士气下降次数（非角色特定逻辑）。
func (e *GameEngine) trackPlagueOutbreakMoraleDrop(lossCtx *model.Context) {
	if lossCtx == nil || lossCtx.Selections == nil {
		return
	}
	sourceSkillID, _ := lossCtx.Selections["damage_source_skill"].(string)
	sourceID, _ := lossCtx.Selections["damage_source_id"].(string)
	if sourceSkillID != "plague_outbreak" || sourceID == "" {
		return
	}
	if source := e.State.Players[sourceID]; source != nil {
		source.TurnState.UsedSkillCounts["plague_outbreak_morale_drop"] = 1
	}
}

func (e *GameEngine) finalizeMoraleLossDiscard(victim *model.Player, discardedCards []model.Card, lossCtx *model.Context) {
	absorbByMoonID := ""
	if lossCtx != nil && lossCtx.Selections != nil {
		absorbByMoonID, _ = lossCtx.Selections["mg_new_moon_absorb_by"].(string)
	}
	if absorbByMoonID == "" {
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
		return
	}
	e.Log(fmt.Sprintf("[Skill] %s 的爆牌被 [新月庇护] 吸收为暗月（未进入弃牌堆）", victim.Name))
}
