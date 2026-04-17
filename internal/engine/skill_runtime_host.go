// gameflow: runtime/skill.Host 的 GameEngine 侧实现。

package engine

import (
	"fmt"

	skillrt "starcup-engine/internal/engine/runtime/skill"
	"starcup-engine/internal/model"
)

type gameSkillHost struct {
	sd *SkillDispatcher
}

func (sd *SkillDispatcher) skillHost() skillrt.Host {
	if sd == nil {
		return nil
	}
	return &gameSkillHost{sd: sd}
}

func (h *gameSkillHost) Log(msg string) {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	h.sd.engine.Log(msg)
}

func (h *gameSkillHost) GameState() *model.GameState {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return nil
	}
	return h.sd.engine.State
}

func (h *gameSkillHost) SnapshotPlayerPoses() any {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return nil
	}
	return h.sd.engine.snapshotPlayerPoses()
}

func (h *gameSkillHost) DispatchOrientationChanges(before any) {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	if m, ok := before.(map[string]poseSnapshot); ok {
		h.sd.engine.dispatchOrientationChanges(m)
	}
}

func (h *gameSkillHost) SyncPendingDamageFromContext(ctx *model.Context) {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	h.sd.engine.syncPendingDamageRuntimeFromContext(ctx)
}

func (h *gameSkillHost) RecordSkillUsage(playerID, title string, skillType model.SkillType) {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	h.sd.engine.recordSkillUsage(playerID, title, skillType)
}

func (h *gameSkillHost) ApplyHitCheckAugment(skillIDs []string, ctx *model.Context) []string {
	if h == nil || h.sd == nil {
		return skillIDs
	}
	return h.sd.applyTimingOnHitCheckResponseSkillAugment(skillIDs, ctx)
}

func (h *gameSkillHost) ApplyHitCheckNormalize(skillIDs []string, ctx *model.Context) []string {
	if h == nil || h.sd == nil {
		return skillIDs
	}
	return h.sd.applyTimingOnHitCheckResponseSkillNormalize(skillIDs, ctx)
}

func (h *gameSkillHost) PublishStartupInterrupt(playerID string, skillIDs []string, sharedCtx *model.Context) {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	e := h.sd.engine
	e.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptStartupSkill,
		PlayerID: playerID,
		SkillIDs: skillIDs,
		Context:  sharedCtx,
	}
}

func (h *gameSkillHost) PublishResponseInterrupt(player *model.Player, skillIDs []string, sharedCtx *model.Context) {
	if h == nil || h.sd == nil || h.sd.engine == nil || player == nil {
		return
	}
	h.sd.engine.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: player.ID,
		SkillIDs: skillIDs,
		Context:  sharedCtx,
	})
}

func (h *gameSkillHost) OnStartupInterruptPublished() {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	e := h.sd.engine
	e.setTurnStage(model.TurnStageActionStart)
	e.clearCombatStage()
	e.clearSubflow()
}

func (h *gameSkillHost) GetMaxHand(player *model.Player) int {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return 0
	}
	return h.sd.engine.GetMaxHand(player)
}

func (h *gameSkillHost) DropQueuedOverflowDiscardForPlayer(playerID string) {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	dropQueuedOverflowDiscardForPlayer(h.sd.engine, playerID)
}

func (h *gameSkillHost) PopInterrupt() {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	h.sd.engine.PopInterrupt()
}

func (h *gameSkillHost) SetPendingInterrupt(intr *model.Interrupt) {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	h.sd.engine.State.PendingInterrupt = intr
}

func (h *gameSkillHost) PendingInterrupt() *model.Interrupt {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return nil
	}
	return h.sd.engine.State.PendingInterrupt
}

func (h *gameSkillHost) EnterDiscardSelection() {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	h.sd.engine.enterDiscardSelection()
}

func (h *gameSkillHost) NotifyInterruptPrompt() {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	h.sd.engine.notifyInterruptPrompt()
}

func (h *gameSkillHost) CaptureResponseResumeStateOnConfirm(skillID string, ctx *model.Context) any {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return responseResumeState{}
	}
	return h.sd.engine.captureResponseResumeStateFromContext(responseCompletionConfirm, skillID, ctx)
}

func (h *gameSkillHost) PrepareConfirmedResponseResume(state any) {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	if s, ok := state.(responseResumeState); ok {
		h.sd.engine.prepareConfirmedResponseResume(s)
	}
}

func (h *gameSkillHost) RestoreConfirmedResponseAfterPop(state any) {
	if h == nil || h.sd == nil || h.sd.engine == nil {
		return
	}
	if s, ok := state.(responseResumeState); ok {
		h.sd.engine.restoreConfirmedResponseAfterPop(s)
	}
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
	filtered := make([]*model.Interrupt, 0, len(e.State.InterruptQueue))
	for _, intr := range e.State.InterruptQueue {
		if intr == nil || intr.PlayerID != playerID || !isDiscardSelectionInterrupt(intr) {
			filtered = append(filtered, intr)
			continue
		}
		data, ok := intr.Context.(map[string]interface{})
		if !ok {
			filtered = append(filtered, intr)
			continue
		}
		victimID, _ := data["victim_id"].(string)
		if victimID == playerID {
			e.Log(fmt.Sprintf("[System] 清理过期中断: %s 的爆牌弃牌请求", player.Name))
			continue
		}
		filtered = append(filtered, intr)
	}
	e.State.InterruptQueue = filtered
}
