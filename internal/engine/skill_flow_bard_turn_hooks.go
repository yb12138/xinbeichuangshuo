package engine

import (
	"fmt"

	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

func (e *GameEngine) maybeTriggerBardRousingAtTurnStart(current *model.Player) bool {
	if current == nil || !e.isBard(current) {
		return false
	}
	if current.TurnState.UsedSkillCounts["bd_rousing_prompted"] > 0 {
		return false
	}
	current.TurnState.UsedSkillCounts["bd_rousing_prompted"] = 1
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
	if current.TurnState.UsedSkillCounts["bd_victory_prompted"] > 0 {
		return false
	}
	current.TurnState.UsedSkillCounts["bd_victory_prompted"] = 1
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

func (e *GameEngine) resetTurnMagicDamageTracker() {
	e.turnMagicDamageTargets = map[string]map[string]bool{}
}

// 吟游诗人：记录“当前回合吟游诗人自己已对哪些敌方角色造成过法术伤害”，并在满足条件时触发沉沦协奏曲。
func (e *GameEngine) tryTriggerBardDescentAfterMagicDamage(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 {
		return false
	}
	source := e.State.Players[pd.SourceID]
	target := e.State.Players[pd.TargetID]
	if source == nil || target == nil || source.Camp == target.Camp {
		return false
	}
	if !e.isBard(source) || !source.IsActive {
		return false
	}

	if e.turnMagicDamageTargets == nil {
		e.resetTurnMagicDamageTracker()
	}
	if _, ok := e.turnMagicDamageTargets[source.ID]; !ok {
		e.turnMagicDamageTargets[source.ID] = map[string]bool{}
	}
	e.turnMagicDamageTargets[source.ID][target.ID] = true
	if len(e.turnMagicDamageTargets[source.ID]) < 2 {
		return false
	}
	if hasBardEternalPrisonerForm(source) || source.TurnState.UsedSkillCounts["bd_descent"] > 0 {
		return false
	}
	if maxSameElementCount(source) < 2 {
		return false
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: source.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_descent_element",
			"user_id":     source.ID,
		},
	})
	e.Log(fmt.Sprintf("%s 满足 [沉沦协奏曲] 触发条件，强制进入弃2张同系牌流程", source.Name))
	return true
}

func (e *GameEngine) bardResponseContext(user *model.Player, stage string, resumePoint interface{}) *model.Context {
	ctx := e.buildContext(user, nil, model.TriggerNone, &model.EventContext{
		Type:     model.EventNone,
		SourceID: user.ID,
	})
	ctx.Selections["bd_song_stage"] = stage
	ctx.Selections["response_resume_phase"] = resumePoint
	return ctx
}

func (e *GameEngine) bardAlliesExcluding(camp model.Camp, excludeID string) []string {
	var ids []string
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp != camp || p.ID == excludeID {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
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
		DamageType: model.MagicAttack,
	})
	e.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感已满，对自己造成3点法术伤害（来源：%s）", bard.Name, songName))
}
