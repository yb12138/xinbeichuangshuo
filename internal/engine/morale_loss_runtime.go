// gameflow: 士气下降与「伤害导致摸牌」等连锁。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ApplyMoraleLossAfterTimingWindow 在 TimingBeforeMoraleLoss 窗口处理完毕后应用士气损失与联动效果。
func (e *GameEngine) ApplyMoraleLossAfterTimingWindow(victim *model.Player, moraleLoss int, isMagic bool, fromDamageDraw bool, overflowMoraleLossFixed int, discardedCards []model.Card, lossCtx *model.Context) int {
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

	finalLoss = e.ApplyCampMoraleLoss(victim.Camp, finalLoss)
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
		SourceID:       moraleLossSourceID(lossCtx),
		SourceSkillID:  moraleLossSourceSkillID(lossCtx),
	})

	if moraleLoss != finalLoss {
		e.Log(fmt.Sprintf("[System] 士气损失被抵御！原损失: %d, 实际损失: %d", moraleLoss, finalLoss))
	}

	e.finalizeMoraleLossDiscard(victim, discardedCards, lossCtx)
	return finalLoss
}

func (e *GameEngine) finalizeMoraleLossDiscard(victim *model.Player, discardedCards []model.Card, lossCtx *model.Context) {
	if moraleLossDiscardDestination(lossCtx) != "absorbed" {
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
		return
	}
	e.Log(fmt.Sprintf("[System] %s 的爆牌已被技能效果吸收（未进入弃牌堆）", victim.Name))
}

func moraleLossSourceID(lossCtx *model.Context) string {
	if lossCtx == nil || lossCtx.Selections == nil {
		return ""
	}
	sourceID, _ := lossCtx.Selections["damage_source_id"].(string)
	return sourceID
}

func moraleLossSourceSkillID(lossCtx *model.Context) string {
	if lossCtx == nil || lossCtx.Selections == nil {
		return ""
	}
	sourceSkillID, _ := lossCtx.Selections["damage_source_skill"].(string)
	return sourceSkillID
}

func moraleLossDiscardDestination(lossCtx *model.Context) string {
	if lossCtx == nil || lossCtx.Selections == nil {
		return ""
	}
	destination, _ := lossCtx.Selections["discard_destination"].(string)
	return destination
}
