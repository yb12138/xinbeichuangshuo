package engine

import (
	"fmt"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

func (e *GameEngine) maybeTriggerBardRousingAtTurnStart(current *model.Player) bool {
	if current == nil || !e.isBard(current) {
		return false
	}
	ensurePlayerTokensMap(current)
	if current.Tokens["bd_rousing_prompted_turn"] > 0 {
		return false
	}
	current.Tokens["bd_rousing_prompted_turn"] = 1
	if e.bardEternalHolderID(current) == "" {
		return false
	}

	ctx := e.bardResponseContext(current, "turn_start", model.TurnStageActionStart)
	handler := skills.GetHandler("bd_rousing_rhapsody")
	if handler == nil || !handler.CanUse(ctx) {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: current.ID,
		SkillIDs: []string{"bd_rousing_rhapsody"},
		Context:  ctx,
	})
	e.Log(fmt.Sprintf("%s 在回合开始时满足 [激昂狂想曲] 的发动条件", current.Name))
	return true
}

func (e *GameEngine) maybeTriggerBardVictoryAtTurnEnd(current *model.Player) bool {
	if current == nil || !e.isBard(current) {
		return false
	}
	ensurePlayerTokensMap(current)
	if current.Tokens["bd_victory_prompted_turn"] > 0 {
		return false
	}
	current.Tokens["bd_victory_prompted_turn"] = 1
	if e.bardEternalHolderID(current) == "" {
		return false
	}

	ctx := e.bardResponseContext(current, "turn_end", model.TurnStageTurnEnd)
	handler := skills.GetHandler("bd_victory_symphony")
	if handler == nil || !handler.CanUse(ctx) {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: current.ID,
		SkillIDs: []string{"bd_victory_symphony"},
		Context:  ctx,
	})
	e.Log(fmt.Sprintf("%s 在回合结束时满足 [胜利交响诗] 的发动条件", current.Name))
	return true
}

func (e *GameEngine) resolveBardForbiddenVerseAfterSong(bard *model.Player, songName string) {
	if bard == nil || !e.isBard(bard) {
		return
	}
	ensurePlayerTokensMap(bard)
	if bardInspiration(bard) < bardInspirationCapEngine {
		now := addBardInspiration(bard, 1)
		removed := e.removeBardEternalMovement(bard)
		if removed {
			e.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感+1（当前%d），并移除永恒乐章", bard.Name, now))
		} else {
			e.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感+1（当前%d）", bard.Name, now))
		}
		return
	}

	if !hasBardEternalPrisonerForm(bard) {
		beforePoses := e.snapshotPlayerPoses()
		enterBardEternalPrisonerForm(bard)
		e.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：转为永恒囚徒形态", bard.Name))
		e.dispatchOrientationChanges(beforePoses)
	}
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   bard.ID,
		TargetID:   bard.ID,
		Damage:     3,
		DamageType: "magic",
	})
	e.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感已满，对自己造成3点法术伤害（来源：%s）", bard.Name, songName))
}
