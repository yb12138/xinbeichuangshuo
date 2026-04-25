package bard

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// MaybeRousingAtTurnStart 检查吟游诗人在回合开始是否满足激昂狂想曲的发动条件。
func MaybeRousingAtTurnStart(rt engineplayer.ChoiceRuntime, current *model.Player) bool {
	if current == nil || !engineplayer.IsCharacter(current, "bard") {
		return false
	}
	if current.TurnState.UsedSkillCounts["bd_rousing_prompted"] > 0 {
		return false
	}
	current.TurnState.UsedSkillCounts["bd_rousing_prompted"] = 1
	if EternalHolderID(rt, current) == "" {
		return false
	}

	ctx := responseContext(rt, current, "turn_start", model.TurnStageActionStart)
	handler := skills.GetHandler("bd_rousing_rhapsody")
	if handler == nil || !handler.CanUse(ctx) {
		return false
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: current.ID,
		SkillIDs: []string{"bd_rousing_rhapsody"},
		Context:  ctx,
	})
	rt.Log(fmt.Sprintf("%s 在回合开始时满足 [激昂狂想曲] 的发动条件", current.Name))
	return true
}

// MaybeVictoryAtTurnEnd 检查吟游诗人在回合结束是否满足胜利交响诗的发动条件。
func MaybeVictoryAtTurnEnd(rt engineplayer.ChoiceRuntime, current *model.Player) bool {
	if current == nil || !engineplayer.IsCharacter(current, "bard") {
		return false
	}
	if current.TurnState.UsedSkillCounts["bd_victory_prompted"] > 0 {
		return false
	}
	current.TurnState.UsedSkillCounts["bd_victory_prompted"] = 1
	if EternalHolderID(rt, current) == "" {
		return false
	}

	ctx := responseContext(rt, current, "turn_end", model.TurnStageTurnEnd)
	handler := skills.GetHandler("bd_victory_symphony")
	if handler == nil || !handler.CanUse(ctx) {
		return false
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: current.ID,
		SkillIDs: []string{"bd_victory_symphony"},
		Context:  ctx,
	})
	rt.Log(fmt.Sprintf("%s 在回合结束时满足 [胜利交响诗] 的发动条件", current.Name))
	return true
}

// TryDescentAfterMagicDamage 在吟游诗人造成法术伤害后，检查是否满足沉沦协奏曲的触发条件。
func TryDescentAfterMagicDamage(rt engineplayer.ChoiceRuntime, pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 {
		return false
	}
	source := rt.GetPlayers()[pd.SourceID]
	target := rt.GetPlayers()[pd.TargetID]
	if source == nil || target == nil || source.Camp == target.Camp {
		return false
	}
	if !engineplayer.IsCharacter(source, "bard") || !source.IsActive {
		return false
	}

	rt.RecordMagicDamageTarget(source.ID, target.ID)
	if rt.MagicDamageTargetCount(source.ID) < 2 {
		return false
	}
	if InEternalPrisonerForm(source) || source.TurnState.UsedSkillCounts["bd_descent"] > 0 {
		return false
	}
	if maxSameElementCount(source) < 2 {
		return false
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: source.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_descent_element",
			"user_id":     source.ID,
		},
	})
	rt.Log(fmt.Sprintf("%s 满足 [沉沦协奏曲] 触发条件，强制进入弃2张同系牌流程", source.Name))
	return true
}

func responseContext(rt engineplayer.ChoiceRuntime, user *model.Player, stage string, resumePoint interface{}) *model.Context {
	ctx := rt.BuildContext(user, nil, model.TimingActive, &model.EventContext{
		Type:     model.EventNone,
		SourceID: user.ID,
	})
	ctx.Selections["bd_song_stage"] = stage
	ctx.Selections["response_resume_phase"] = resumePoint
	return ctx
}

func maxSameElementCount(player *model.Player) int {
	counts := map[model.Element]int{}
	for _, c := range player.Hand {
		if c.Element != "" {
			counts[c.Element]++
		}
	}
	maxCount := 0
	for _, cnt := range counts {
		if cnt > maxCount {
			maxCount = cnt
		}
	}
	return maxCount
}
