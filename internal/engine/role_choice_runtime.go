// gameflow: player ChoiceRuntime 的统一 engine 适配与桥接。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type roleChoiceRuntime struct {
	*GameEngine
}

func (r roleChoiceRuntime) LookupPlayer(playerID string) *model.Player {
	if r.GameEngine == nil || r.State == nil || playerID == "" {
		return nil
	}
	return r.State.Players[playerID]
}

func (r roleChoiceRuntime) AllPlayers() map[string]*model.Player {
	if r.GameEngine == nil || r.State == nil {
		return map[string]*model.Player{}
	}
	return r.State.Players
}

func (r roleChoiceRuntime) PlayerOrder() []string {
	if r.GameEngine == nil || r.State == nil || len(r.State.PlayerOrder) == 0 {
		return nil
	}
	return append([]string(nil), r.State.PlayerOrder...)
}

func (r roleChoiceRuntime) HasPendingInterrupt() bool {
	return r.GameEngine != nil && r.State != nil && r.State.PendingInterrupt != nil
}

func (r roleChoiceRuntime) ReplacePendingInterruptContext(data map[string]interface{}) error {
	if r.GameEngine == nil || r.State == nil || r.State.PendingInterrupt == nil {
		return fmt.Errorf("当前没有待处理的选择流程")
	}
	r.State.PendingInterrupt.Context = data
	return nil
}

func (r roleChoiceRuntime) ResumePendingAttackMiss(ctx *model.Context) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.resumePendingAttackMiss(ctx)
}

func (r roleChoiceRuntime) ResumePendingAttackHit(ctxData map[string]interface{}) {
	if r.GameEngine == nil {
		return
	}
	r.resumePendingAttackHit(ctxData)
}

func (r roleChoiceRuntime) ApplyChoiceResumePoint(raw interface{}) {
	if r.GameEngine == nil {
		return
	}
	r.applyChoiceResumePoint(raw)
}

func (r roleChoiceRuntime) RoutePendingDamageOr(defaultReturn interface{}, onNoPending func()) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.routePendingDamageOr(defaultReturn, onNoPending)
}

func (r roleChoiceRuntime) EnterExtraActionStage() {
	if r.GameEngine == nil {
		return
	}
	r.enterExtraActionStage()
}

func (r roleChoiceRuntime) EnterTurnEndStage() {
	if r.GameEngine == nil {
		return
	}
	r.enterTurnEndStage()
}

func (r roleChoiceRuntime) EnterDamageResolution(returnTo interface{}) {
	if r.GameEngine == nil {
		return
	}
	r.enterDamageResolution(returnTo)
}

func (r roleChoiceRuntime) EnterActionExecutionStage() {
	if r.GameEngine == nil {
		return
	}
	r.enterActionExecutionStage()
}

func (r roleChoiceRuntime) RoutePendingDamageWithReturn(returnTo interface{}) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.routePendingDamageWithReturn(returnTo)
}

func (r roleChoiceRuntime) AllOtherPlayerIDs(userID string) []string {
	if r.GameEngine == nil {
		return nil
	}
	return r.allOtherPlayerIDs(userID)
}

func (r roleChoiceRuntime) PlayerOrderPosition(playerID string) int {
	if r.GameEngine == nil {
		return 0
	}
	return r.playerOrderPosition(playerID)
}

func (r roleChoiceRuntime) StartDraw(ctx *model.Context) {
	if r.GameEngine == nil {
		return
	}
	r.startDraw(ctx)
}

func (r roleChoiceRuntime) NewDrawContext(player *model.Player, amount int, reason string) *model.Context {
	if r.GameEngine == nil {
		return nil
	}
	return r.newDrawContext(player, amount, reason)
}

func (r roleChoiceRuntime) RestorePhaseAfterInterruptedDraw(ctx *model.Context) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.restorePhaseAfterInterruptedDraw(ctx)
}

func (r roleChoiceRuntime) PendingDamageQueueLen() int {
	if r.GameEngine == nil || r.State == nil {
		return 0
	}
	return len(r.State.PendingDamageQueue)
}

func (r roleChoiceRuntime) ActionQueueLen() int {
	if r.GameEngine == nil || r.State == nil {
		return 0
	}
	return len(r.State.ActionQueue)
}

func (r roleChoiceRuntime) AttachExclusiveEffectCard(sourceID, targetID string, effect model.EffectType, card model.Card) error {
	if r.GameEngine == nil {
		return fmt.Errorf("engine not available")
	}
	source := r.State.Players[sourceID]
	target := r.State.Players[targetID]
	if source == nil || target == nil {
		return fmt.Errorf("source or target player not found")
	}
	return r.attachExclusiveEffectCard(source, target, effect, card)
}

func (r roleChoiceRuntime) ResumePendingMoraleLoss(ctx *model.Context) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.resumePendingMoraleLoss(ctx)
}

func (r roleChoiceRuntime) EnterResponseWindow() {
	if r.GameEngine == nil {
		return
	}
	r.enterResponseWindow()
}

func (r roleChoiceRuntime) ApplyStealthEffect(player *model.Player) {
	if r.GameEngine == nil {
		return
	}
	r.applyAssassinStealthEffect(player)
}

var _ engineplayer.ChoiceRuntime = roleChoiceRuntime{}

func newRoleChoiceRuntime(e *GameEngine) engineplayer.ChoiceRuntime {
	return roleChoiceRuntime{GameEngine: e}
}

func (e *GameEngine) buildRoleChoicePrompt(roleID, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	entry := roleRegistry.Entry(roleID)
	if entry.ID == "" {
		return nil
	}
	return entry.BuildChoicePrompt(newRoleChoiceRuntime(e), choiceType, playerID, player, data)
}

func (e *GameEngine) handleRoleChoiceInput(roleID, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	entry := roleRegistry.Entry(roleID)
	if entry.ID == "" {
		return false, nil
	}
	return entry.HandleChoice(newRoleChoiceRuntime(e), playerID, selectionIndex, ctxData)
}

func (e *GameEngine) handleRoleChoiceCancel(roleID, playerID string, ctxData map[string]interface{}) (bool, error) {
	entry := roleRegistry.Entry(roleID)
	if entry.ID == "" {
		return false, nil
	}
	if ctxData == nil && e != nil && e.State != nil && e.State.PendingInterrupt != nil {
		ctxData, _ = e.State.PendingInterrupt.Context.(map[string]interface{})
	}
	return entry.HandleChoiceCancel(newRoleChoiceRuntime(e), playerID, ctxData)
}
