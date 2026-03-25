package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// applyMoraleLossAfterTrigger 在 TriggerBeforeMoraleLoss 后应用士气损失与联动效果。
func (e *GameEngine) applyMoraleLossAfterTrigger(victim *model.Player, moraleLoss int, isMagic bool, fromDamageDraw bool, overflowMoraleLossFixed int, discardedCards []model.Card, lossCtx *model.Context) int {
	if victim == nil {
		if len(discardedCards) > 0 {
			e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
		}
		return 0
	}

	finalLoss := moraleLoss
	if lossCtx != nil && lossCtx.TriggerCtx != nil && lossCtx.TriggerCtx.DamageVal != nil {
		finalLoss = *lossCtx.TriggerCtx.DamageVal
	}
	if finalLoss < 0 {
		finalLoss = 0
	}
	if overflowMoraleLossFixed > 0 && finalLoss > 0 {
		finalLoss = overflowMoraleLossFixed
	}

	finalLoss = e.applyCampMoraleLoss(victim.Camp, finalLoss)
	e.applySoulSorcererSoulDevour(victim, finalLoss, fromDamageDraw)
	e.applyDamageDrivenMoraleLossRoleEffects(victim, finalLoss, isMagic, fromDamageDraw, lossCtx)

	if moraleLoss != finalLoss {
		e.Log(fmt.Sprintf("[System] 士气损失被抵御！原损失: %d, 实际损失: %d", moraleLoss, finalLoss))
	}

	e.finalizeMoraleLossDiscard(victim, discardedCards, lossCtx)
	return finalLoss
}

func (e *GameEngine) applyDamageDrivenMoraleLossRoleEffects(victim *model.Player, finalLoss int, isMagic bool, fromDamageDraw bool, lossCtx *model.Context) {
	if victim == nil || !fromDamageDraw || finalLoss <= 0 {
		return
	}

	if e.isBloodPriestess(victim) {
		ensurePlayerTokensMap(victim)
		if e.enterBloodPriestessBleedingForm(victim, "因承受伤害导致我方士气下降") {
			e.Heal(victim.ID, 1)
			e.Log(fmt.Sprintf("%s 的 [流血] 触发：获得1点治疗", victim.Name))
		}
	}

	if isMagic && e.isBlazeWitch(victim) {
		ensurePlayerTokensMap(victim)
		before := victim.Tokens["bw_rebirth"]
		victim.Tokens["bw_rebirth"]++
		if victim.Tokens["bw_rebirth"] > 4 {
			victim.Tokens["bw_rebirth"] = 4
		}
		if victim.Tokens["bw_rebirth"] != before {
			e.Log(fmt.Sprintf("%s 的 [永生银时计] 触发，重生+1（当前%d）", victim.Name, victim.Tokens["bw_rebirth"]))
		}
	}

	// 红莲骑士：仅当“伤害导致且实际发生士气下降”时，强制进入热血沸腾形态。
	if e.isCrimsonKnight(victim) && !hasCrimsonKnightHotBloodedForm(victim) {
		ensurePlayerTokensMap(victim)
		beforePoses := e.snapshotPlayerPoses()
		enterCrimsonKnightHotBloodedForm(victim)
		e.Log(fmt.Sprintf("%s 的 [热血沸腾] 触发，进入热血沸腾形态", victim.Name))
		e.dispatchOrientationChanges(beforePoses)
	}

	if lossCtx == nil || lossCtx.Selections == nil {
		return
	}
	sourceSkillID, _ := lossCtx.Selections["damage_source_skill"].(string)
	sourceID, _ := lossCtx.Selections["damage_source_id"].(string)
	if sourceSkillID != "plague_outbreak" || sourceID == "" {
		return
	}
	if source := e.State.Players[sourceID]; source != nil {
		ensurePlayerTokensMap(source)
		source.Tokens["plague_outbreak_morale_drop_turn"] = 1
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
