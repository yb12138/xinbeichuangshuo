// gameflow: 士气下降与「伤害导致摸牌」等连锁。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// runAfterCampMoraleLossHooks 为不经过爆牌流程的士气下降（技能直接扣士气）补发后置钩子。
// victim 取该阵营按座次的第一名玩家，钩子只依赖其阵营。
func (e *GameEngine) runAfterCampMoraleLossHooks(camp model.Camp, loss int) {
	if e == nil || loss <= 0 {
		return
	}
	var victim *model.Player
	for _, pid := range e.State.PlayerOrder {
		if p := e.State.Players[pid]; p != nil && p.Camp == camp {
			victim = p
			break
		}
	}
	if victim == nil {
		return
	}
	for _, entry := range roleRegistry.Entries() {
		if entry.AfterMoraleLossHook != nil {
			entry.AfterMoraleLossHook(e, victim, loss, false)
		}
	}
}

// ApplyMoraleLossAfterTimingWindow 在 TimingMoraleLossCheck 窗口处理完毕后应用士气损失与联动效果。
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
	if fromDamageDraw && finalLoss > 0 {
		e.addActionMoraleLoss(victim.Camp, finalLoss)
	}
	if finalLoss > 0 {
		e.publishMoraleLossMarker(victim, finalLoss, len(discardedCards), lossCtx)
	}
	e.dispatchSettlementRulebookTiming(model.TimingMoraleLossApplied, victim, nil, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  moraleLossSourceID(lossCtx),
		TargetID:  victim.ID,
		DamageVal: &finalLoss,
	})
	for _, entry := range roleRegistry.Entries() {
		if entry.AfterMoraleLossHook != nil {
			entry.AfterMoraleLossHook(e, victim, finalLoss, fromDamageDraw)
		}
	}
	e.dispatchRoleTimingHook(engineplayer.TimingMoraleLossApplied, engineplayer.TimingHookContext{
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

// publishMoraleLossMarker 公开一次士气下降的原因：谁爆牌（PlayerID）、伤害来源（TargetIDs[1]）、来源技能。
// 前端据此在战斗链里写"红方士气 15 → 14（吟游诗人爆牌 · 凋零）"，具体数值取自同批的士气差值。
func (e *GameEngine) publishMoraleLossMarker(victim *model.Player, loss, discarded int, lossCtx *model.Context) {
	sourceID := moraleLossSourceID(lossCtx)
	skillID := moraleLossSourceSkillID(lossCtx)
	skillName := e.skillTitleFor(sourceID, skillID)
	// 多数技能伤害不显式填来源技能；仅用于展示时退回当前技能因果（凋零、激昂狂想曲……）。
	if skillName == "" {
		if cause := e.CurrentCause(); cause != nil && cause.SkillID != "" {
			skillID = cause.SkillID
			skillName = firstNonEmptyString(cause.SkillName, e.skillTitleFor(cause.OwnerID, cause.SkillID))
		}
	}
	targets := []string{victim.ID}
	if sourceID != "" && sourceID != victim.ID {
		targets = append(targets, sourceID)
	}
	e.publishTimelineMarker(model.TimelineMarkerPayload{
		PlayerID:   victim.ID,
		PlayerName: victim.Name,
		SkillID:    skillID,
		SkillName:  skillName,
		Summary:    fmt.Sprintf("%s 爆牌弃 %d 张，%s士气 -%d", victim.Name, discarded, campDisplayName(victim.Camp), loss),
		TargetIDs:  targets,
		// 不用 effect_token：那会让流图再建一个"效果"节点；前端直接按 narrative_kind 读取原因
		NarrativeKind: "morale_loss",
		Timing:        "morale.loss",
	})
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
