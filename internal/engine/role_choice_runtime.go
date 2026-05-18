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

// ---- StateReader 实现 ----

func (r roleChoiceRuntime) GetPlayers() map[string]*model.Player {
	if r.GameEngine == nil || r.State == nil {
		return map[string]*model.Player{}
	}
	return r.State.Players
}

func (r roleChoiceRuntime) GetPlayerOrder() []string {
	if r.GameEngine == nil || r.State == nil || len(r.State.PlayerOrder) == 0 {
		return nil
	}
	return append([]string(nil), r.State.PlayerOrder...)
}

func (r roleChoiceRuntime) GetCurrentTurnIndex() int {
	if r.GameEngine == nil || r.State == nil {
		return -1
	}
	return r.State.CurrentTurn
}

func (r roleChoiceRuntime) GetRedMorale() int {
	if r.GameEngine == nil || r.State == nil {
		return 0
	}
	return r.State.RedMorale
}

func (r roleChoiceRuntime) GetBlueMorale() int {
	if r.GameEngine == nil || r.State == nil {
		return 0
	}
	return r.State.BlueMorale
}

func (r roleChoiceRuntime) GetPendingInterrupt() *model.Interrupt {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.PendingInterrupt
}

func (r roleChoiceRuntime) GetPendingDamageQueue() []model.PendingDamage {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.PendingDamageQueue
}

func (r roleChoiceRuntime) GetPendingDamage() *model.PendingDamage {
	if r.GameEngine == nil || r.State == nil || len(r.State.PendingDamageQueue) == 0 {
		return nil
	}
	return &r.State.PendingDamageQueue[0]
}

func (r roleChoiceRuntime) GetPendingDamageByIndex(index int) (*model.PendingDamage, bool) {
	if r.GameEngine == nil || r.State == nil || index < 0 || index >= len(r.State.PendingDamageQueue) {
		return nil, false
	}
	return &r.State.PendingDamageQueue[index], true
}

func (r roleChoiceRuntime) GetCombatStack() []model.CombatRequest {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return append([]model.CombatRequest(nil), r.State.CombatStack...)
}

func (r roleChoiceRuntime) GetActionQueue() []model.QueuedAction {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return append([]model.QueuedAction(nil), r.State.ActionQueue...)
}

func (r roleChoiceRuntime) GetDiscardPile() []model.Card {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return append([]model.Card(nil), r.State.DiscardPile...)
}

func (r roleChoiceRuntime) GetDeck() []model.Card {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return append([]model.Card(nil), r.State.Deck...)
}

func (r roleChoiceRuntime) GetTurnStage() model.TurnStage {
	if r.GameEngine == nil || r.State == nil {
		return ""
	}
	return r.State.TurnStage
}

func (r roleChoiceRuntime) GetCombatStage() model.CombatStage {
	if r.GameEngine == nil || r.State == nil {
		return model.CombatStageNone
	}
	return r.State.CombatStage
}

func (r roleChoiceRuntime) GetSubflow() model.Subflow {
	if r.GameEngine == nil || r.State == nil {
		return model.SubflowNone
	}
	return r.State.Subflow
}

func (r roleChoiceRuntime) GetMagicBulletChain() *model.MagicBulletChain {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.MagicBulletChain
}

// ---- EffectCardOps 实现（统一 API）----

func (r roleChoiceRuntime) FindEffectCard(source *model.Player, effect model.EffectType) (*model.Player, *model.FieldCard) {
	if r.GameEngine == nil {
		return nil, nil
	}
	return r.findSourceEffectCard(source, effect)
}

func (r roleChoiceRuntime) AttachEffectCard(source, target *model.Player, effect model.EffectType, card model.Card) error {
	if r.GameEngine == nil {
		return fmt.Errorf("engine not available")
	}
	return r.attachSourceEffectCard(source, target, effect, card)
}

func (r roleChoiceRuntime) DetachEffectCard(source *model.Player, effect model.EffectType) (*model.Player, model.Card, bool) {
	if r.GameEngine == nil {
		return nil, model.Card{}, false
	}
	return r.detachSourceEffectCard(source, effect)
}

func (r roleChoiceRuntime) RemoveEffectCard(source *model.Player, effect model.EffectType, restoreCard bool) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.removeExclusiveEffectCard(source, effect, restoreCard)
}

func (r roleChoiceRuntime) EmitBuffRemovedDispatch(sourceID, targetID string, effect model.EffectType) {
	if r.GameEngine == nil {
		return
	}
	r.emitBuffRemovedDispatch(sourceID, targetID, effect)
}

// ---- InterruptOps 实现 ----

func (r roleChoiceRuntime) PopInterrupt() {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.PopInterrupt()
}

func (r roleChoiceRuntime) NotifyInterruptPrompt() {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.notifyInterruptPrompt()
}

func (r roleChoiceRuntime) PushInterrupt(intr *model.Interrupt) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.PushInterrupt(intr)
}

func (r roleChoiceRuntime) PushDiscardChoiceInterrupt(playerID string, data map[string]interface{}) {
	if r.GameEngine == nil {
		return
	}
	r.PushInterrupt(newDiscardChoiceInterrupt(playerID, data))
}

// ---- StageOps 实现 ----

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

func (r roleChoiceRuntime) EnterActionEndStage() {
	if r.GameEngine == nil {
		return
	}
	r.enterActionEndStage()
}

func (r roleChoiceRuntime) EnterResponseWindow() {
	if r.GameEngine == nil {
		return
	}
	r.enterResponseWindow()
}

func (r roleChoiceRuntime) ApplyChoiceResumePoint(raw interface{}) {
	if r.GameEngine == nil {
		return
	}
	r.applyChoiceResumePoint(raw)
}

// ---- DamageOps 实现 ----

func (r roleChoiceRuntime) RoutePendingDamageOr(defaultReturn interface{}, onNoPending func()) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.routePendingDamageOr(defaultReturn, onNoPending)
}

func (r roleChoiceRuntime) RoutePendingDamageWithReturn(returnTo interface{}) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.routePendingDamageWithReturn(returnTo)
}

func (r roleChoiceRuntime) ResumePendingMoraleLoss(ctx *model.Context) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.resumePendingMoraleLoss(ctx)
}

// ---- CombatOps 实现 ----

func (r roleChoiceRuntime) EnsureCombatInteractionWindow() {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	if len(r.State.CombatStack) > 0 && r.State.CombatStage == model.CombatStageNone {
		r.State.CombatStage = model.CombatStageHitCheck
	}
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
	cardIdx := r.GameEngine.findPlayableCardIndexByID(player, cardID)
	card, _, _, ok := r.GameEngine.getPlayableCardByIndex(player, cardIdx)
	if !ok {
		return model.Card{}, false
	}
	if _, err := r.GameEngine.consumePlayableCardByIndex(player, cardIdx); err != nil {
		return model.Card{}, false
	}
	return card, true
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
		CardID:          card.ID,
		SourceSkill:     sourceSkill,
		UsesVirtualCard: true,
	})
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

// ---- DrawOps 实现 ----

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

func (r roleChoiceRuntime) DrawRawCards(amount int) ([]model.Card, bool) {
	if r.GameEngine == nil || r.State == nil {
		return nil, false
	}
	cards, newDeck, newDiscard := rules.DrawCards(r.State.Deck, r.State.DiscardPile, amount)
	r.State.Deck = newDeck
	r.State.DiscardPile = newDiscard
	return cards, true
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

// ---- HandOps 实现 ----

func (r roleChoiceRuntime) RoleFixedMaxHandCapValue(player *model.Player) (int, bool) {
	if r.GameEngine == nil {
		return 0, false
	}
	return r.roleFixedMaxHandCapValue(player)
}

func (r roleChoiceRuntime) TakeDiscardPileCardByID(cardID string) (model.Card, bool) {
	if r.GameEngine == nil || r.State == nil || cardID == "" {
		return model.Card{}, false
	}
	for i := len(r.State.DiscardPile) - 1; i >= 0; i-- {
		if r.State.DiscardPile[i].ID != cardID {
			continue
		}
		card := r.State.DiscardPile[i]
		r.State.DiscardPile = append(r.State.DiscardPile[:i], r.State.DiscardPile[i+1:]...)
		return card, true
	}
	return model.Card{}, false
}

// ---- PoseOps 实现 ----

func (r roleChoiceRuntime) PoseChangeGuard() func() {
	if r.GameEngine == nil {
		return func() {}
	}
	before := r.SnapshotPlayerPoses()
	return func() {
		r.DispatchOrientationChanges(before)
	}
}

// ---- MagicBulletOps 实现 ----

func (r roleChoiceRuntime) SetMagicBulletChain(chain *model.MagicBulletChain) {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	r.State.MagicBulletChain = chain
}

func (r roleChoiceRuntime) GetPlayableCardByIndex(player *model.Player, idx int) (model.Card, bool) {
	card, _, _, ok := r.GameEngine.getPlayableCardByIndex(player, idx)
	return card, ok
}

func (r roleChoiceRuntime) GetPlayableCardByCardID(player *model.Player, cardID string) (model.Card, bool) {
	card, _, _, ok := r.GameEngine.getPlayableCardByID(player, cardID)
	return card, ok
}

func (r roleChoiceRuntime) ConsumePlayableCardByIndex(player *model.Player, idx int) (model.Card, error) {
	return r.GameEngine.consumePlayableCardByIndex(player, idx)
}

func (r roleChoiceRuntime) PerformMagic(playerID, targetID string, cardIdx int) error {
	if r.GameEngine == nil {
		return fmt.Errorf("engine not available")
	}
	return r.GameEngine.PerformMagic(playerID, targetID, cardIdx)
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

func (r roleChoiceRuntime) OfferMagicMissileResponseSkills() {
	if r.GameEngine == nil {
		return
	}
	r.offerMagicMissileResponseSkills()
}

func (r roleChoiceRuntime) DispatchHitCheckMagicMissileCounter(player *model.Player, chain *model.MagicBulletChain, card *model.Card) error {
	if r.GameEngine == nil {
		return nil
	}
	return r.GameEngine.applyTimingOnHitCheckMagicMissileCounterValidation(player, chain, *card)
}

func (r roleChoiceRuntime) DispatchHitCheckMagicMissileDefend(player *model.Player, chain *model.MagicBulletChain) error {
	if r.GameEngine == nil {
		return nil
	}
	return r.GameEngine.applyTimingOnHitCheckMagicMissileDefendValidation(player, chain)
}

// ---- SkillOps 实现 ----

func (r roleChoiceRuntime) IsSkillStillUsable(skillID string, user *model.Player, ctx *model.Context) bool {
	if r.GameEngine == nil || r.dispatcher == nil {
		return false
	}
	return r.dispatcher.IsSkillStillUsable(skillID, user, ctx)
}

func (r roleChoiceRuntime) RecordSkillUsage(playerID, title string, skillType model.SkillType) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.recordSkillUsage(playerID, title, skillType)
}

func (r roleChoiceRuntime) IsActionSkillUsableForExtraMagic(player *model.Player, skillDef model.SkillDefinition) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.GameEngine.isActionSkillUsableForExtraMagic(player, skillDef)
}

func (r roleChoiceRuntime) RecordMagicDamageTarget(sourceID, targetID string) {
	if r.GameEngine == nil {
		return
	}
	if r.GameEngine.turnMagicDamageTargets == nil {
		r.GameEngine.turnMagicDamageTargets = map[string]map[string]bool{}
	}
	if _, ok := r.GameEngine.turnMagicDamageTargets[sourceID]; !ok {
		r.GameEngine.turnMagicDamageTargets[sourceID] = map[string]bool{}
	}
	r.GameEngine.turnMagicDamageTargets[sourceID][targetID] = true
}

func (r roleChoiceRuntime) MagicDamageTargetCount(sourceID string) int {
	if r.GameEngine == nil || r.GameEngine.turnMagicDamageTargets == nil {
		return 0
	}
	return len(r.GameEngine.turnMagicDamageTargets[sourceID])
}

// ---- MoraleOps 实现 ----

func (r roleChoiceRuntime) AddCampMorale(camp model.Camp, amount int) int {
	if r.GameEngine == nil {
		return 0
	}
	return r.addCampMorale(camp, amount)
}

func (r roleChoiceRuntime) ApplyCampMoraleLoss(camp model.Camp, wantLoss int) int {
	if r.GameEngine == nil {
		return 0
	}
	return r.GameEngine.ApplyCampMoraleLoss(camp, wantLoss)
}

// ---- GameOps 实现 ----

func (r roleChoiceRuntime) CheckGameEnd() {
	if r.GameEngine == nil {
		return
	}
	r.checkGameEnd()
}

func (r roleChoiceRuntime) RefreshAllPlayerDerivedStates() {
	if r.GameEngine == nil {
		return
	}
	r.RefreshAllPlayerDerivedStates()
}

func (r roleChoiceRuntime) BuildContext(user, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context {
	if r.GameEngine == nil {
		return nil
	}
	return r.GameEngine.BuildContext(user, target, timing, eventCtx)
}

var _ engineplayer.ChoiceRuntime = roleChoiceRuntime{}

// Helper methods for common patterns (not in interface, but useful for role packages)

func (r roleChoiceRuntime) CampEnemyIDs(camp model.Camp) []string {
	if r.GameEngine == nil {
		return nil
	}
	return r.campEnemyIDs(camp)
}

func (r roleChoiceRuntime) CurrentTurnPlayerID() string {
	if r.GameEngine == nil || r.State == nil {
		return ""
	}
	if r.State.CurrentTurn < 0 || r.State.CurrentTurn >= len(r.State.PlayerOrder) {
		return ""
	}
	return r.State.PlayerOrder[r.State.CurrentTurn]
}

func (r roleChoiceRuntime) PendingDiscardVictimID() string {
	if r.GameEngine == nil {
		return ""
	}
	return r.pendingDiscardVictimID()
}

func (r roleChoiceRuntime) TopCombatRequest() *model.CombatRequest {
	if r.GameEngine == nil || r.State == nil || len(r.State.CombatStack) == 0 {
		return nil
	}
	return &r.State.CombatStack[len(r.State.CombatStack)-1]
}

func (r roleChoiceRuntime) MagicBulletChain() *model.MagicBulletChain {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.MagicBulletChain
}

func (r roleChoiceRuntime) PendingInterrupt() *model.Interrupt {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.PendingInterrupt
}

func (r roleChoiceRuntime) AddToDiscardPile(cards ...model.Card) {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	r.State.DiscardPile = append(r.State.DiscardPile, cards...)
}

func (r roleChoiceRuntime) ReplacePendingInterruptPlayerID(newPlayerID string) {
	if r.GameEngine == nil || r.State == nil || r.State.PendingInterrupt == nil {
		return
	}
	r.State.PendingInterrupt.PlayerID = newPlayerID
}

func (r roleChoiceRuntime) PlayerOrder() []string {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.PlayerOrder
}

func (r roleChoiceRuntime) LookupPlayer(playerID string) *model.Player {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.Players[playerID]
}

func (r roleChoiceRuntime) HasPendingInterrupt() bool {
	return r.GetPendingInterrupt() != nil
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

func (r roleChoiceRuntime) AllPlayers() []*model.Player {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	out := make([]*model.Player, 0, len(r.State.Players))
	for _, pid := range r.State.PlayerOrder {
		if p := r.State.Players[pid]; p != nil {
			out = append(out, p)
		}
	}
	return out
}

func (r roleChoiceRuntime) ReplacePendingInterruptContext(data map[string]interface{}) error {
	if r.GameEngine == nil || r.State == nil || r.State.PendingInterrupt == nil {
		return fmt.Errorf("no pending interrupt")
	}
	r.State.PendingInterrupt.Context = data
	return nil
}

func (r roleChoiceRuntime) AllOtherPlayerIDs(userID string) []string {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	out := make([]string, 0, len(r.State.PlayerOrder))
	for _, pid := range r.State.PlayerOrder {
		if pid != userID {
			out = append(out, pid)
		}
	}
	return out
}

func NewRoleChoiceRuntime(e *GameEngine) engineplayer.ChoiceRuntime {
	return roleChoiceRuntime{GameEngine: e}
}

func (e *GameEngine) buildRoleChoicePrompt(roleID, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	entry := roleRegistry.Entry(roleID)
	if entry.ID == "" {
		return nil
	}
	return entry.BuildChoicePrompt(NewRoleChoiceRuntime(e), choiceType, playerID, player, data)
}

func (e *GameEngine) handleRoleChoiceInput(roleID, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	entry := roleRegistry.Entry(roleID)
	if entry.ID == "" {
		return false, nil
	}
	return entry.HandleChoice(NewRoleChoiceRuntime(e), playerID, selectionIndex, ctxData)
}

func (e *GameEngine) handleRoleChoiceCancel(roleID, playerID string, ctxData map[string]interface{}) (bool, error) {
	entry := roleRegistry.Entry(roleID)
	if entry.ID == "" {
		return false, nil
	}
	if ctxData == nil && e != nil && e.State != nil && e.State.PendingInterrupt != nil {
		ctxData, _ = e.State.PendingInterrupt.Context.(map[string]interface{})
	}
	return entry.HandleChoiceCancel(NewRoleChoiceRuntime(e), playerID, ctxData)
}
