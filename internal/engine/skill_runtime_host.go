// gameflow: SkillDispatcher 实现 skillrt.Host，提供技能执行与中断所需的引擎能力。

package engine

import (
	"fmt"

	skillrt "starcup-engine/internal/engine/runtime/skill"
	"starcup-engine/internal/model"
)

var _ skillrt.Host = (*SkillDispatcher)(nil)

func (sd *SkillDispatcher) skillHost() skillrt.Host {
	if sd == nil {
		return nil
	}
	return sd
}

func (sd *SkillDispatcher) Log(msg string) {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.Log(msg)
}

func (sd *SkillDispatcher) GameState() *model.GameState {
	if sd == nil || sd.engine == nil {
		return nil
	}
	return sd.engine.State
}

func (sd *SkillDispatcher) SnapshotPlayerPoses() any {
	if sd == nil || sd.engine == nil {
		return nil
	}
	return sd.engine.SnapshotPlayerPoses()
}

func (sd *SkillDispatcher) DispatchOrientationChanges(before any) {
	if sd == nil || sd.engine == nil {
		return
	}
	if m, ok := before.(map[string]poseSnapshot); ok {
		sd.engine.DispatchOrientationChanges(m)
	}
}

func (sd *SkillDispatcher) SyncPendingDamageFromContext(ctx *model.Context) {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.syncPendingDamageRuntimeFromContext(ctx)
}

func (sd *SkillDispatcher) RecordSkillUsage(playerID, title string, skillType model.SkillType) {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.recordSkillUsage(playerID, title, skillType)
}

func (sd *SkillDispatcher) NotifySkillActivated(playerID, skillID, skillName, effectText string, targetIDs []string) {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.NotifySkillActivated(playerID, skillID, skillName, effectText, targetIDs)
}

func (sd *SkillDispatcher) ApplyHitCheckAugment(skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil {
		return skillIDs
	}
	return sd.applyAttackResponseSkillAugment(skillIDs, ctx)
}

func (sd *SkillDispatcher) ApplyHitCheckNormalize(skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil {
		return skillIDs
	}
	return sd.applyAttackResponseSkillNormalize(skillIDs, ctx)
}

func (sd *SkillDispatcher) PublishStartupInterrupt(playerID string, skillIDs []string, sharedCtx *model.Context) {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptStartupSkill,
		PlayerID: playerID,
		SkillIDs: skillIDs,
		Context:  sharedCtx,
	})
}

func (sd *SkillDispatcher) PublishResponseInterrupt(player *model.Player, skillIDs []string, sharedCtx *model.Context) {
	if sd == nil || sd.engine == nil || player == nil {
		return
	}
	sd.engine.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: player.ID,
		SkillIDs: skillIDs,
		Context:  sharedCtx,
	})
}

func (sd *SkillDispatcher) OnStartupInterruptPublished() {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.setTurnStage(model.TurnStageActionStart)
	sd.engine.clearCombatStage()
	sd.engine.clearSubflow()
}

func (sd *SkillDispatcher) GetMaxHand(player *model.Player) int {
	if sd == nil || sd.engine == nil {
		return 0
	}
	return sd.engine.GetMaxHand(player)
}

func (sd *SkillDispatcher) DropQueuedOverflowDiscardForPlayer(playerID string) {
	if sd == nil || sd.engine == nil {
		return
	}
	dropQueuedOverflowDiscardForPlayer(sd.engine, playerID)
}

func (sd *SkillDispatcher) PopInterrupt() {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.PopInterrupt()
}

func (sd *SkillDispatcher) SetPendingInterrupt(intr *model.Interrupt) {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.State.SetPendingInterrupt(intr)
}

func (sd *SkillDispatcher) PendingInterrupt() *model.Interrupt {
	if sd == nil || sd.engine == nil {
		return nil
	}
	return sd.engine.State.PendingInterrupt
}

func (sd *SkillDispatcher) EnterDiscardSelection() {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.EnterDiscardSelection()
}

func (sd *SkillDispatcher) NotifyInterruptPrompt() {
	if sd == nil || sd.engine == nil {
		return
	}
	sd.engine.notifyInterruptPrompt()
}

func (sd *SkillDispatcher) CaptureResponseResumeStateOnConfirm(skillID string, ctx *model.Context) any {
	if sd == nil || sd.engine == nil {
		return responseResumeState{}
	}
	return sd.engine.captureResponseResumeStateFromContext(responseCompletionConfirm, skillID, ctx)
}

func (sd *SkillDispatcher) PrepareConfirmedResponseResume(state any) {
	if sd == nil || sd.engine == nil {
		return
	}
	if s, ok := state.(responseResumeState); ok {
		sd.engine.prepareConfirmedResponseResume(s)
	}
}

func (sd *SkillDispatcher) RestoreConfirmedResponseAfterPop(state any) {
	if sd == nil || sd.engine == nil {
		return
	}
	if s, ok := state.(responseResumeState); ok {
		sd.engine.restoreConfirmedResponseAfterPop(s)
	}
}

func (sd *SkillDispatcher) ConsumeSkillEnergyCost(playerID string, gemCost, crystalCost int) bool {
	if sd == nil || sd.engine == nil {
		return false
	}
	p := sd.engine.State.Players[playerID]
	if p == nil {
		return false
	}
	if !consumeSkillEnergyCost(p, gemCost, crystalCost) {
		return false
	}
	sd.engine.recordActionResourceDelta()
	return true
}

// dropQueuedOverflowDiscardForPlayer 清理精灵密仪等确认后残留的爆牌弃牌中断。
func dropQueuedOverflowDiscardForPlayer(e *GameEngine, playerID string) {
	if e == nil {
		return
	}
	player := e.State.Players[playerID]
	if player == nil {
		return
	}
	if len(player.Hand) > e.GetMaxHand(player) {
		return
	}
	e.RemoveQueuedInterruptByPredicate(func(intr *model.Interrupt) bool {
		if intr == nil || intr.PlayerID != playerID || !IsDiscardSelectionInterrupt(intr) {
			return false
		}
		data, ok := intr.Context.(map[string]interface{})
		if !ok {
			return false
		}
		victimID, _ := data["victim_id"].(string)
		if victimID == playerID {
			e.Log(fmt.Sprintf("[System] 清理过期中断: %s 的爆牌弃牌请求", player.Name))
			return true
		}
		return false
	})
}
