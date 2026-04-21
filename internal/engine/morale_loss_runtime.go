// gameflow: 士气下降与「伤害导致摸牌」等连锁。

package engine

import (
	"fmt"

	playerpkg "starcup-engine/internal/engine/player"
	bloodpriestesspkg "starcup-engine/internal/engine/player/blood_priestess"
	soulsorcererpkg "starcup-engine/internal/engine/player/soul_sorcerer"
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
	soulsorcererpkg.ApplySoulDevour(e, victim, finalLoss, fromDamageDraw)
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

	if isCharacter(victim, "blood_priestess") {
		ensurePlayerTokensMap(victim)
		if bloodpriestesspkg.EnterBleedingFormWithLog(newRoleChoiceRuntime(e), victim, "因承受伤害导致我方士气下降") {
			e.Heal(victim.ID, 1)
			e.Log(fmt.Sprintf("%s 的 [流血] 触发：获得1点治疗", victim.Name))
		}
	}

	if isMagic && isCharacter(victim, "blaze_witch") {
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
	if isCharacter(victim, "crimson_knight") && !playerpkg.HasForm(victim, model.FormCrimsonKnightHotBlooded) {
		ensurePlayerTokensMap(victim)
		beforePoses := e.snapshotPlayerPoses()
		playerpkg.SetForm(victim, model.FormCrimsonKnightHotBlooded)
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
