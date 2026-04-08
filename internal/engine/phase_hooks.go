package engine

import (
	"fmt"

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
		return false
	case timingOnTurnEndFinal:
		for _, hook := range e.turnEndFinalHooks {
			if hook(e, player) {
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
	if !e.isBlazeWitch(player) || player.TurnState.HasUsedActionSkill || !hasBlazeWitchFlameForm(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["bw_flame_release_pending"] <= 0 {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	leaveBlazeWitchFlameForm(player)
	player.Tokens["bw_flame_release_pending"] = 0
	e.Log(fmt.Sprintf("%s 脱离烈焰形态并转正", player.Name))
	e.dispatchOrientationChanges(beforePoses)
	return false
}

func startupAssassinStealthReleaseHook(e *GameEngine, player *model.Player) bool {
	if !isCharacter(player, "assassin") || player.TurnState.HasUsedActionSkill || !hasAssassinStealthForm(player) {
		return false
	}
	e.releaseAssassinStealthEffect(player)
	return false
}

func startupMagicSwordsmanShadowReleaseHook(e *GameEngine, player *model.Player) bool {
	e.maybeReleaseMagicSwordsmanShadowAtActionStart(player)
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
	return e.maybeTriggerBardRousingAtTurnStart(player)
}

func turnStartArbiterJudgmentUpkeepHook(e *GameEngine, player *model.Player) bool {
	if !hasArbiterJudgmentForm(player) || player.TurnState.HasUsedActionSkill {
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

func turnEndBeastSamuraiHook(e *GameEngine, player *model.Player) bool {
	if !e.isBeastSamurai(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	if e.beastSamuraiInIaijutsuForm(player) && e.beastSamuraiBeastSoul(player) >= 1 {
		consumed := e.consumeBeastSamuraiBeastSoul(player, 1)
		if consumed > 0 {
			e.Log(fmt.Sprintf("%s 的 [御魂流居合形态·回合结束扣魂] 生效：兽魂-1，残心+1", player.Name))
		}
	}
	if e.beastSamuraiInIaijutsuForm(player) && e.beastSamuraiBeastSoul(player) == 0 {
		beforePoses := e.snapshotPlayerPoses()
		if e.leaveBeastSamuraiIaijutsuForm(player) {
			e.Log(fmt.Sprintf("%s 的 [御魂流居合形态·兽魂归零退场] 生效：转正并脱离御魂流居合形态", player.Name))
		}
		e.dispatchOrientationChanges(beforePoses)
	}
	player.TurnState.UsedSkillCounts["bs_one_strike_armed"] = 0
	clearBeastSamuraiAttackTokens(player)
	return false
}

func turnEndFighterHook(e *GameEngine, player *model.Player) bool {
	if e.isFighter(player) && hasFighterHundredDragonForm(player) {
		e.clearFighterHundredDragon(player, fmt.Sprintf("%s 的 [百式幻龙拳] 在本行动阶段结束时退场并转正", player.Name))
	}
	return false
}

func turnEndElfArcherHook(e *GameEngine, player *model.Player) bool {
	if !e.isElfArcher(player) || !hasElfArcherRitualForm(player) {
		return false
	}
	syncElfBlessings(player)
	if countElfBlessings(player) != 0 || player.Tokens["elf_ritual_release_waiting"] != 0 {
		return false
	}
	targetIDs := e.campEnemyIDs(player.Camp)
	if len(targetIDs) == 0 {
		leaveElfArcherRitualForm(player)
		player.Tokens["elf_ritual_release_waiting"] = 0
		e.Log(fmt.Sprintf("%s 的 [精灵密仪] 结束：无敌方目标，直接转正脱离精灵祝福形态", player.Name))
		return true
	}
	player.Tokens["elf_ritual_release_waiting"] = 1
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "elf_ritual_release_target",
			"user_id":     player.ID,
			"target_ids":  targetIDs,
		},
	})
	return true
}

func turnEndPlagueMageHook(e *GameEngine, player *model.Player) bool {
	if !e.isPlagueMage(player) || player.TurnState.UsedSkillCounts["plague_outbreak_morale_drop"] <= 0 {
		return false
	}
	player.TurnState.UsedSkillCounts["plague_outbreak_morale_drop"] = 0
	e.Heal(player.ID, 1)
	e.Log(fmt.Sprintf("%s 的 [瘟疫] 回合结束奖励生效：+1治疗", player.Name))
	return false
}

func turnEndMoonGoddessHook(e *GameEngine, player *model.Player) bool {
	return e.maybeTriggerMoonGoddessMoonCycleAtTurnEnd(player)
}

func turnEndBardHook(e *GameEngine, player *model.Player) bool {
	return e.maybeTriggerBardVictoryAtTurnEnd(player)
}

func turnEndCrimsonSwordSpiritHook(e *GameEngine, player *model.Player) bool {
	if !e.isCrimsonSwordSpirit(player) {
		return false
	}
	hasCourtyard := false
	for _, fc := range player.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectRoseCourtyard {
			hasCourtyard = true
			break
		}
	}
	if !hasCourtyard {
		return false
	}
	player.Tokens["css_blood_cap"] = 3
	if player.Tokens["css_blood"] > 3 {
		player.Tokens["css_blood"] = 3
	}
	if e.removeExclusiveEffectCard(player, model.EffectRoseCourtyard, true) {
		e.Log(fmt.Sprintf("%s 的 [血蔷薇庭院] 回合结束移回专属卡区", player.Name))
	} else {
		e.Log(fmt.Sprintf("%s 的 [血蔷薇庭院] 回合结束失效", player.Name))
	}
	return false
}

func turnEndCrimsonKnightHook(e *GameEngine, player *model.Player) bool {
	e.resolveCrimsonKnightHotFormTurnEnd(player)
	return false
}

func turnEndWarHomunculusHook(e *GameEngine, player *model.Player) bool {
	if !e.isWarHomunculus(player) || !hasWarHomunculusBurstForm(player) {
		return false
	}
	leaveWarHomunculusBurstForm(player)
	e.Log(fmt.Sprintf("%s 的 [符文改造] 效果结束，脱离蓄势迸发形态", player.Name))
	e.checkHandLimit(player, nil)
	return e.State.PendingInterrupt != nil
}

func turnEndOnmyojiHook(e *GameEngine, player *model.Player) bool {
	return e.maybeTriggerOnmyojiDarkRitual(player)
}

func turnEndHolyBowHook(e *GameEngine, player *model.Player) bool {
	if !e.isHolyBow(player) {
		return false
	}
	if player.TurnState.UsedSkillCounts["hb_auto_fill"] > 0 {
		return false
	}
	if player.TurnState.UsedSkillCounts["hb_special"] <= 0 {
		var resourceModes []string
		if e.CanPayCrystalCost(player.ID, 1) {
			resourceModes = append(resourceModes, "crystal")
		}
		if player.Gem > 0 {
			resourceModes = append(resourceModes, "gem")
		}
		if len(resourceModes) > 0 {
			player.TurnState.UsedSkillCounts["hb_auto_fill"] = 1
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: player.ID,
				Context: map[string]interface{}{
					"choice_type":    "hb_auto_fill_resource",
					"user_id":        player.ID,
					"resource_modes": resourceModes,
				},
			})
			e.Log(fmt.Sprintf("%s 的 [自动填充] 触发：请选择消耗资源与增益", player.Name))
			return true
		}
	}
	player.TurnState.UsedSkillCounts["hb_auto_fill"] = 1
	return false
}

func turnEndHolyLancerHook(_ *GameEngine, player *model.Player) bool {
	if player != nil {
		player.TurnState.UsedSkillCounts["holy_lancer_prayer"] = 0
	}
	return false
}

func turnFallbackCrimsonKnightHook(e *GameEngine, player *model.Player) bool {
	e.resolveCrimsonKnightHotFormTurnEnd(player)
	return false
}

func turnFallbackFighterHook(e *GameEngine, player *model.Player) bool {
	e.clearFighterHundredDragon(player, "")
	return false
}

// resolveCrimsonKnightHotFormTurnEnd 统一处理红莲骑士“热血沸腾”在回合结束时的退形态逻辑。
// 返回 true 表示本次确实触发了退形态与治疗。
func (e *GameEngine) resolveCrimsonKnightHotFormTurnEnd(player *model.Player) bool {
	if player == nil || !e.isCrimsonKnight(player) {
		return false
	}
	if !hasCrimsonKnightHotBloodedForm(player) {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	leaveCrimsonKnightHotBloodedForm(player)
	e.Heal(player.ID, 2)
	e.Log(fmt.Sprintf("%s 回合结束脱离 [热血沸腾形态]，获得2点治疗", player.Name))
	e.dispatchOrientationChanges(beforePoses)
	return true
}
