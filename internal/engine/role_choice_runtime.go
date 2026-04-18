// gameflow: player ChoiceRuntime 的统一 engine 适配与桥接。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
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

func (r roleChoiceRuntime) GetPendingDamage(index int) (*model.PendingDamage, bool) {
	if r.GameEngine == nil || r.State == nil || index < 0 || index >= len(r.State.PendingDamageQueue) {
		return nil, false
	}
	return &r.State.PendingDamageQueue[index], true
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

func (r roleChoiceRuntime) EnqueueVirtualAttack(sourceID, targetID string, card model.Card, sourceSkill string) {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	r.State.ActionQueue = append(r.State.ActionQueue, model.QueuedAction{
		SourceID:        sourceID,
		TargetID:        targetID,
		Type:            model.ActionAttack,
		Element:         card.Element,
		Card:            &card,
		CardIndex:       -1,
		SourceSkill:     sourceSkill,
		UsesVirtualCard: true,
	})
}

func (r roleChoiceRuntime) ReplacePendingInterruptPlayerID(playerID string) {
	if r.GameEngine == nil || r.State == nil || r.State.PendingInterrupt == nil {
		return
	}
	r.State.PendingInterrupt.PlayerID = playerID
}

func (r roleChoiceRuntime) ApplyCampMoraleLoss(camp model.Camp, wantLoss int) int {
	if r.GameEngine == nil {
		return 0
	}
	return r.applyCampMoraleLoss(camp, wantLoss)
}

func (r roleChoiceRuntime) ResolveCounterAttack(counterPlayerID, counterTargetID string, counterCard model.Card) {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	r.resolveCounterAttack(counterPlayerID, counterTargetID, counterCard)
}

func (r roleChoiceRuntime) NotifyCombatCue(attackerID, targetID, cueType string) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.NotifyCombatCue(attackerID, targetID, cueType)
}

func (r roleChoiceRuntime) ConsumePlayableCardByCardID(playerID, cardID string) (model.Card, bool) {
	if r.GameEngine == nil || r.State == nil {
		return model.Card{}, false
	}
	player := r.State.Players[playerID]
	if player == nil {
		return model.Card{}, false
	}
	cardIdx := findPlayableCardIndexByID(player, cardID)
	card, _, _, ok := getPlayableCardByIndex(player, cardIdx)
	if !ok {
		return model.Card{}, false
	}
	if _, err := consumePlayableCardByIndex(player, cardIdx); err != nil {
		return model.Card{}, false
	}
	return card, true
}

func (r roleChoiceRuntime) TopCombatRequest() *model.CombatRequest {
	if r.GameEngine == nil || r.State == nil || len(r.State.CombatStack) == 0 {
		return nil
	}
	return &r.State.CombatStack[len(r.State.CombatStack)-1]
}

func (r roleChoiceRuntime) PopCombatRequest() {
	if r.GameEngine == nil || r.State == nil || len(r.State.CombatStack) == 0 {
		return
	}
	r.State.CombatStack = r.State.CombatStack[:len(r.State.CombatStack)-1]
}

func (r roleChoiceRuntime) EnsureCombatInteractionWindow() {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	if len(r.State.CombatStack) > 0 && r.State.CombatStage == model.CombatStageNone {
		r.State.CombatStage = model.CombatStageHitCheck
	}
}

func (r roleChoiceRuntime) DrawCardsDirect(playerID string, amount int, reason string) {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	p := r.State.Players[playerID]
	if p == nil {
		return
	}
	cards, newDeck, newDiscard := rules.DrawCards(r.State.Deck, r.State.DiscardPile, amount)
	r.State.Deck = newDeck
	r.State.DiscardPile = newDiscard
	p.Hand = append(p.Hand, cards...)
	r.NotifyDrawCards(playerID, amount, reason)
}

func (r roleChoiceRuntime) PendingInterrupt() *model.Interrupt {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.PendingInterrupt
}

func (r roleChoiceRuntime) RoutePendingDamageWithDefaultReturn(defaultReturn interface{}) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.routePendingDamageWithDefaultReturn(defaultReturn)
}

func (r roleChoiceRuntime) RestoreReturnPoint() bool {
	if r.GameEngine == nil {
		return false
	}
	return r.restoreReturnPoint()
}

func (r roleChoiceRuntime) PushDiscardChoiceInterrupt(playerID string, data map[string]interface{}) {
	if r.GameEngine == nil {
		return
	}
	r.PushInterrupt(newDiscardChoiceInterrupt(playerID, data))
}

func (r roleChoiceRuntime) EnterActionEndStage() {
	if r.GameEngine == nil {
		return
	}
	r.enterActionEndStage()
}

func (r roleChoiceRuntime) MagicBulletChain() *model.MagicBulletChain {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.MagicBulletChain
}

func (r roleChoiceRuntime) SetMagicBulletChain(chain *model.MagicBulletChain) {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	r.State.MagicBulletChain = chain
}

func (r roleChoiceRuntime) SetReturnPoint(returnTo interface{}) {
	if r.GameEngine == nil {
		return
	}
	r.setReturnPoint(returnTo)
}

func (r roleChoiceRuntime) GetPlayableCardByIndex(player *model.Player, idx int) (model.Card, bool) {
	card, _, _, ok := getPlayableCardByIndex(player, idx)
	return card, ok
}

func (r roleChoiceRuntime) ConsumePlayableCardByIndex(player *model.Player, idx int) (model.Card, error) {
	return consumePlayableCardByIndex(player, idx)
}

func (r roleChoiceRuntime) PerformMagic(playerID, targetID string, cardIdx int, isFusion bool) error {
	if r.GameEngine == nil {
		return fmt.Errorf("engine not available")
	}
	return r.performMagic(playerID, targetID, cardIdx, isFusion)
}

func (r roleChoiceRuntime) ExecuteMagicBullet(player *model.Player, reverse, isFusion bool, fusionCard *model.Card) error {
	if r.GameEngine == nil {
		return fmt.Errorf("engine not available")
	}
	return r.executeMagicBullet(player, reverse, isFusion, fusionCard)
}

func (r roleChoiceRuntime) FindNextMagicBulletTarget(playerID string) string {
	if r.GameEngine == nil {
		return ""
	}
	return r.findNextMagicBulletTarget(playerID)
}

func (r roleChoiceRuntime) DispatchHitCheckMagicMissileCounter(player *model.Player, chain *model.MagicBulletChain, card *model.Card) error {
	if r.GameEngine == nil {
		return nil
	}
	res := r.dispatchTimingOnHitCheck(timingOnHitCheckContext{
		Op:         timingOnHitCheckMagicMissileCounter,
		Player:     player,
		MagicChain: chain,
		Card:       *card,
	})
	return res.Err
}

func (r roleChoiceRuntime) DispatchHitCheckMagicMissileDefend(player *model.Player, chain *model.MagicBulletChain) error {
	if r.GameEngine == nil {
		return nil
	}
	res := r.dispatchTimingOnHitCheck(timingOnHitCheckContext{
		Op:         timingOnHitCheckMagicMissileDefend,
		Player:     player,
		MagicChain: chain,
	})
	return res.Err
}

func (r roleChoiceRuntime) AddToDiscardPile(cards ...model.Card) {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	r.State.DiscardPile = append(r.State.DiscardPile, cards...)
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
