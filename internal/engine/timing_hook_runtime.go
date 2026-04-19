// gameflow: player.HookRuntime 的 engine 适配器。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type hookRuntime struct {
	*GameEngine
}

var _ engineplayer.HookRuntime = hookRuntime{}

func newHookRuntime(e *GameEngine) engineplayer.HookRuntime {
	return hookRuntime{GameEngine: e}
}

func (r hookRuntime) GetPlayer(playerID string) *model.Player {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.Players[playerID]
}

func (r hookRuntime) PushInterrupt(intr *model.Interrupt) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.PushInterrupt(intr)
}

func (r hookRuntime) PushDiscardChoiceInterrupt(playerID string, data map[string]interface{}) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.PushInterrupt(newDiscardChoiceInterrupt(playerID, data))
}

func (r hookRuntime) Heal(targetID string, amount int) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.Heal(targetID, amount)
}

func (r hookRuntime) AddPendingDamage(pd model.PendingDamage) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.AddPendingDamage(pd)
}

func (r hookRuntime) GetMaxHand(player *model.Player) int {
	if r.GameEngine == nil {
		return 0
	}
	return r.GameEngine.GetMaxHand(player)
}

func (r hookRuntime) GetPlayerEnergyCap(player *model.Player) int {
	if r.GameEngine == nil {
		return 0
	}
	return r.GameEngine.getPlayerEnergyCap(player)
}

func (r hookRuntime) DrawCards(playerID string, amount int) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.DrawCards(playerID, amount)
}

func (r hookRuntime) GetPendingDamageQueue() []model.PendingDamage {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.PendingDamageQueue
}

func (r hookRuntime) SetPendingDamageQueue(queue []model.PendingDamage) {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	r.State.PendingDamageQueue = queue
}

func (r hookRuntime) SnapshotPlayerPoses() map[string]engineplayer.PoseSnapshot {
	if r.GameEngine == nil {
		return nil
	}
	internal := r.GameEngine.snapshotPlayerPoses()
	out := make(map[string]engineplayer.PoseSnapshot, len(internal))
	for id, ps := range internal {
		out[id] = engineplayer.PoseSnapshot{
			Orientation: ps.Orientation,
			Form:        ps.Form,
		}
	}
	return out
}

func (r hookRuntime) DispatchOrientationChanges(before map[string]engineplayer.PoseSnapshot) {
	if r.GameEngine == nil {
		return
	}
	internal := make(map[string]poseSnapshot, len(before))
	for id, ps := range before {
		internal[id] = poseSnapshot{
			Orientation: ps.Orientation,
			Form:        ps.Form,
		}
	}
	r.GameEngine.dispatchOrientationChanges(internal)
}

func (r hookRuntime) CampMorale(camp model.Camp) int {
	if r.GameEngine == nil {
		return 0
	}
	return r.GameEngine.campMorale(camp)
}

func (r hookRuntime) HasPendingDiscardFor(playerID string) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.GameEngine.pendingDiscardVictimID() == playerID
}

func (r hookRuntime) PlayerOrder() []string {
	if r.GameEngine == nil || r.GameEngine.State == nil {
		return nil
	}
	return r.GameEngine.State.PlayerOrder
}
