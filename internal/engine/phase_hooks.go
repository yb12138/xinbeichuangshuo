// gameflow: 阶段边界上的通用钩子入口。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	playerpkg "starcup-engine/internal/engine/player"
	bardpkg "starcup-engine/internal/engine/player/bard"
	magicswordsman "starcup-engine/internal/engine/player/magic_swordsman"
	moonpkg "starcup-engine/internal/engine/player/moon"
	"starcup-engine/internal/model"
)

type turnTimingHook func(e *GameEngine, player *model.Player) bool

func ensurePlayerTokensMap(player *model.Player) {
	if player != nil && player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
}

type timingOnTurnStartStage int

const (
	timingOnTurnStartBeforeStart timingOnTurnStartStage = iota
	timingOnTurnStartMain
)

type timingOnTurnEndStage int

const (
	timingOnTurnEndPreExtra timingOnTurnEndStage = iota
	timingOnTurnEndFinal
)

// runTimingOnTurnStartStageHooks 统一处理 TimingOnTurnStart 阶段规则。
func (e *GameEngine) runTimingOnTurnStartStageHooks(player *model.Player, stage timingOnTurnStartStage) bool {
	switch stage {
	case timingOnTurnStartBeforeStart:
		for _, hook := range e.turnStartBeforeStartHooks {
			if hook(e, player) {
				return true
			}
		}
		return false
	case timingOnTurnStartMain:
		for _, hook := range e.turnStartMainHooks {
			if hook(e, player) {
				return true
			}
		}
		// TimingHookSpec dispatch for turn start
		if player != nil {
			result := e.dispatchAllRoleTimingHooks(engineplayer.TimingOnTurnStart, engineplayer.TimingHookContext{
				SourceID:     player.ID,
				TurnPlayerID: player.ID,
			})
			if result.Interrupted {
				return true
			}
		}
		return false
	default:
		panic(fmt.Sprintf("unregistered TimingOnTurnStart stage: %d", stage))
	}
}

// runTimingOnTurnEndStageHooks 统一处理 TimingOnTurnEnd 阶段规则。
func (e *GameEngine) runTimingOnTurnEndStageHooks(player *model.Player, stage timingOnTurnEndStage) bool {
	switch stage {
	case timingOnTurnEndPreExtra:
		for _, hook := range e.turnEndPreExtraHooks {
			if hook(e, player) {
				return true
			}
		}
		// TimingHookSpec dispatch for turn end (pre-extra)
		if player != nil {
			result := e.dispatchAllRoleTimingHooks(engineplayer.TimingOnTurnEnd, engineplayer.TimingHookContext{
				SourceID:     player.ID,
				TurnPlayerID: player.ID,
			})
			if result.Interrupted {
				return true
			}
		}
		return false
	case timingOnTurnEndFinal:
		for _, hook := range e.turnEndFinalHooks {
			if hook(e, player) {
				return true
			}
		}
		// TimingHookSpec dispatch for turn end final
		if player != nil {
			result := e.dispatchAllRoleTimingHooks(engineplayer.TimingOnTurnEndFinal, engineplayer.TimingHookContext{
				SourceID:     player.ID,
				TurnPlayerID: player.ID,
			})
			if result.Interrupted {
				return true
			}
		}
		return false
	default:
		panic(fmt.Sprintf("unregistered TimingOnTurnEnd stage: %d", stage))
	}
}

// runTimingOnTurnStartBeforeStartHooks 回合开始前（TurnBeforeStart）固定结算点。
func (e *GameEngine) runTimingOnTurnStartBeforeStartHooks(player *model.Player) bool {
	return e.runTimingOnTurnStartStageHooks(player, timingOnTurnStartBeforeStart)
}

// runTimingOnTurnStartHooks 回合开始（TurnStart）固定结算点。
func (e *GameEngine) runTimingOnTurnStartHooks(player *model.Player) bool {
	return e.runTimingOnTurnStartStageHooks(player, timingOnTurnStartMain)
}

// runTimingBeforeActionExecuteHooks 行动开始（ActionStart）固定结算点。
func (e *GameEngine) runTimingBeforeActionExecuteHooks(player *model.Player) bool {
	for _, hook := range e.beforeActionExecuteHooks {
		if hook(e, player) {
			return true
		}
	}
	// TimingHookSpec dispatch for before action
	if player != nil {
		result := e.dispatchAllRoleTimingHooks(engineplayer.TimingBeforeAction, engineplayer.TimingHookContext{
			SourceID:     player.ID,
			TurnPlayerID: player.ID,
		})
		if result.Interrupted {
			return true
		}
	}
	return false
}

// runTimingOnTurnEndPreExtraHooks 回合结束前置结算（额外行动判定前）。
func (e *GameEngine) runTimingOnTurnEndPreExtraHooks(player *model.Player) bool {
	return e.runTimingOnTurnEndStageHooks(player, timingOnTurnEndPreExtra)
}

// runTimingOnTurnEndFinalHooks 回合结束最终结算点。
func (e *GameEngine) runTimingOnTurnEndFinalHooks(player *model.Player) bool {
	return e.runTimingOnTurnEndStageHooks(player, timingOnTurnEndFinal)
}

func startupBlazeWitchFlameReleaseHook(e *GameEngine, player *model.Player) bool {
	if !isCharacter(player, "blaze_witch") || player.TurnState.HasUsedActionSkill || !playerpkg.HasForm(player, model.FormBlazeWitchFlame) {
		return false
	}
	playerpkg.EnsurePlayerSkillFlowState(player)
	if player.TurnState.SkillFlowState["bw_flame_release_pending"] <= 0 {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	playerpkg.ClearForm(player, model.FormBlazeWitchFlame)
	player.TurnState.SkillFlowState["bw_flame_release_pending"] = 0
	e.Log(fmt.Sprintf("%s 脱离烈焰形态并转正", player.Name))
	e.dispatchOrientationChanges(beforePoses)
	return false
}

func startupAssassinStealthReleaseHook(e *GameEngine, player *model.Player) bool {
	if !isCharacter(player, "assassin") || player.TurnState.HasUsedActionSkill || !playerpkg.HasForm(player, model.FormAssassinStealth) {
		return false
	}
	e.releaseAssassinStealthEffect(player)
	return false
}

func startupMagicSwordsmanShadowReleaseHook(e *GameEngine, player *model.Player) bool {
	magicswordsman.MaybeReleaseShadowAtActionStart(e, player)
	return false
}

func startupArbiterTurnResetHook(_ *GameEngine, player *model.Player) bool {
	player.TurnState.UsedSkillCounts["arbiter_skip_forced_doomsday"] = 0
	player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_done_turn"] = 0
	return false
}

func startupHolyBowTurnResetHook(_ *GameEngine, player *model.Player) bool {
	player.TurnState.UsedSkillCounts["hb_special"] = 0
	player.TurnState.UsedSkillCounts["hb_auto_fill"] = 0
	return false
}

func startupBardRousingHook(e *GameEngine, player *model.Player) bool {
	return bardpkg.MaybeRousingAtTurnStart(newRoleChoiceRuntime(e), player)
}

func turnStartArbiterJudgmentUpkeepHook(e *GameEngine, player *model.Player) bool {
	if !playerpkg.HasForm(player, model.FormArbiterJudgment) || player.TurnState.HasUsedActionSkill {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["judgment"] >= 4 {
		return false
	}
	player.Tokens["judgment"]++
	e.Log(fmt.Sprintf("%s 处于审判形态，回合开始审判+1（当前%d）", player.Name, player.Tokens["judgment"]))
	return false
}

func turnEndMoonGoddessHook(e *GameEngine, player *model.Player) bool {
	return moonpkg.MaybeMoonCycleAtTurnEnd(newRoleChoiceRuntime(e), player)
}

func turnEndBardHook(e *GameEngine, player *model.Player) bool {
	return bardpkg.MaybeVictoryAtTurnEnd(newRoleChoiceRuntime(e), player)
}

func turnEndOnmyojiHook(e *GameEngine, player *model.Player) bool {
	return e.maybeOnmyojiDarkRitual(player)
}

// resolveCrimsonKnightHotFormTurnEnd 统一处理红莲骑士"热血沸腾"在回合结束时的退形态逻辑。
// 返回 true 表示本次确实触发了退形态与治疗。
func (e *GameEngine) resolveCrimsonKnightHotFormTurnEnd(player *model.Player) bool {
	if player == nil || !isCharacter(player, "crimson_knight") {
		return false
	}
	if !playerpkg.HasForm(player, model.FormCrimsonKnightHotBlooded) {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	playerpkg.ClearForm(player, model.FormCrimsonKnightHotBlooded)
	e.Heal(player.ID, 2)
	e.Log(fmt.Sprintf("%s 回合结束脱离 [热血沸腾形态]，获得2点治疗", player.Name))
	e.dispatchOrientationChanges(beforePoses)
	return true
}
