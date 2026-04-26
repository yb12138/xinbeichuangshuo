// gameflow: player.HookRuntime 的 engine 适配器。

package engine

import (
	"starcup-engine/internal/engine/core/runtimeutil"
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

type combatPolicyRuntime struct {
	hookRuntime
	combatRequest *model.CombatRequest
	chain         *model.MagicBulletChain
}

func (r combatPolicyRuntime) GetCombatRequest() *model.CombatRequest {
	return r.combatRequest
}

func (r combatPolicyRuntime) GetMagicBulletChain() *model.MagicBulletChain {
	return r.chain
}

var _ engineplayer.CombatPolicyRuntime = combatPolicyRuntime{}

func newCombatPolicyRuntime(e *GameEngine, req *model.CombatRequest, chain *model.MagicBulletChain) engineplayer.CombatPolicyRuntime {
	return combatPolicyRuntime{hookRuntime: hookRuntime{GameEngine: e}, combatRequest: req, chain: chain}
}

func (r combatPolicyRuntime) AsChoiceRuntime() engineplayer.ChoiceRuntime {
	if r.GameEngine == nil {
		return nil
	}
	return newRoleChoiceRuntime(r.GameEngine)
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

func (r hookRuntime) SetPendingDamageQueue(queue []model.PendingDamage) {
	if r.GameEngine == nil || r.State == nil {
		return
	}
	r.State.PendingDamageQueue = queue
}

func (r hookRuntime) PoseChangeGuard() func() {
	if r.GameEngine == nil {
		return func() {}
	}
	before := r.GameEngine.snapshotPlayerPoses()
	return func() {
		r.GameEngine.dispatchOrientationChanges(before)
	}
}

func (r hookRuntime) HasPendingDiscardFor(playerID string) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.GameEngine.pendingDiscardVictimID() == playerID
}

// StateReader implementation

func (r hookRuntime) GetPlayers() map[string]*model.Player {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.Players
}

func (r hookRuntime) GetPlayerOrder() []string {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.PlayerOrder
}

func (r hookRuntime) GetCurrentTurnIndex() int {
	if r.GameEngine == nil || r.State == nil {
		return -1
	}
	return r.State.CurrentTurn
}

func (r hookRuntime) GetRedMorale() int {
	if r.GameEngine == nil || r.State == nil {
		return 0
	}
	return r.State.RedMorale
}

func (r hookRuntime) GetBlueMorale() int {
	if r.GameEngine == nil || r.State == nil {
		return 0
	}
	return r.State.BlueMorale
}

func (r hookRuntime) GetPendingInterrupt() *model.Interrupt {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.PendingInterrupt
}

func (r hookRuntime) GetPendingDamageQueue() []model.PendingDamage {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.PendingDamageQueue
}

func (r hookRuntime) GetCombatStack() []model.CombatRequest {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.CombatStack
}

func (r hookRuntime) GetActionQueue() []model.QueuedAction {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.ActionQueue
}

func (r hookRuntime) GetDiscardPile() []model.Card {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.DiscardPile
}

func (r hookRuntime) GetDeck() []model.Card {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.Deck
}

func (r hookRuntime) GetCombatStage() model.CombatStage {
	if r.GameEngine == nil || r.State == nil {
		return ""
	}
	return r.State.CombatStage
}

func (r hookRuntime) GetSubflow() model.Subflow {
	if r.GameEngine == nil || r.State == nil {
		return ""
	}
	return r.State.Subflow
}

func (r hookRuntime) GetMagicBulletChain() *model.MagicBulletChain {
	if r.GameEngine == nil || r.State == nil {
		return nil
	}
	return r.State.MagicBulletChain
}

func (r hookRuntime) RecordMagicDamageTarget(sourceID, targetID string) {
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

func (r hookRuntime) MagicDamageTargetCount(sourceID string) int {
	if r.GameEngine == nil || r.GameEngine.turnMagicDamageTargets == nil {
		return 0
	}
	return len(r.GameEngine.turnMagicDamageTargets[sourceID])
}

func (r hookRuntime) BuildContext(user, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context {
	if r.GameEngine == nil {
		return nil
	}
	return r.GameEngine.buildContext(user, target, timing, eventCtx)
}

func (r hookRuntime) IsSkillStillUsable(skillID string, user *model.Player, ctx *model.Context) bool {
	if r.GameEngine == nil || r.GameEngine.dispatcher == nil {
		return false
	}
	return r.GameEngine.dispatcher.isSkillStillUsable(skillID, user, ctx)
}

func (r hookRuntime) IsMagicDamageType(damageType model.DamageType) bool {
	return runtimeutil.IsMagicDamageType(damageType)
}

// 新增 - 角色身份检查
func (r hookRuntime) IsCharacter(player *model.Player, roleID string) bool {
	if player == nil || player.Character == nil {
		return false
	}
	return player.Character.ID == roleID
}

// 新增 - 形态操作（使用 player 包的辅助函数）
func (r hookRuntime) HasForm(player *model.Player, form string) bool {
	return engineplayer.HasForm(player, form)
}

func (r hookRuntime) SetForm(player *model.Player, form string) bool {
	return engineplayer.SetForm(player, form)
}

func (r hookRuntime) ClearForm(player *model.Player, form string) bool {
	return engineplayer.ClearForm(player, form)
}

// 新增 - 指示物操作
func (r hookRuntime) GetToken(player *model.Player, key string) int {
	if player == nil || player.Tokens == nil {
		return 0
	}
	return player.Tokens[key]
}

func (r hookRuntime) SetToken(player *model.Player, key string, value int) {
	if player == nil {
		return
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	player.Tokens[key] = value
}

// 新增 - 阵营士气（扩展）
func (r hookRuntime) AddCampMorale(camp model.Camp, amount int) int {
	if r.GameEngine == nil {
		return 0
	}
	return r.GameEngine.addCampMorale(camp, amount)
}

func (r hookRuntime) GetCampCups(camp string) int {
	if r.GameEngine == nil {
		return 0
	}
	return r.GameEngine.GetCampCups(camp)
}

// 新增 - 战斗上下文
func (r hookRuntime) GetPendingDamage() *model.PendingDamage {
	if r.GameEngine == nil || r.GameEngine.State == nil || len(r.GameEngine.State.PendingDamageQueue) == 0 {
		return nil
	}
	return &r.GameEngine.State.PendingDamageQueue[0]
}

func (r hookRuntime) GetPendingDamageByIndex(index int) (*model.PendingDamage, bool) {
	if r.GameEngine == nil || r.GameEngine.State == nil || index < 0 || index >= len(r.GameEngine.State.PendingDamageQueue) {
		return nil, false
	}
	return &r.GameEngine.State.PendingDamageQueue[index], true
}

func (r hookRuntime) SetPendingDamageDamage(pd *model.PendingDamage, damage int) {
	if pd == nil {
		return
	}
	pd.Damage = damage
}

func (r hookRuntime) GetCurrentCombat() *model.CombatRequest {
	if r.GameEngine == nil || r.GameEngine.State == nil || len(r.GameEngine.State.CombatStack) == 0 {
		return nil
	}
	return &r.GameEngine.State.CombatStack[len(r.GameEngine.State.CombatStack)-1]
}

func (r hookRuntime) PopCombatRequest() {
	if r.GameEngine == nil || r.GameEngine.State == nil || len(r.GameEngine.State.CombatStack) == 0 {
		return
	}
	r.GameEngine.State.CombatStack = r.GameEngine.State.CombatStack[:len(r.GameEngine.State.CombatStack)-1]
}

// 新增 - 卡牌操作
func (r hookRuntime) GetCardByIndex(player *model.Player, idx int) (model.Card, bool) {
	if player == nil || idx < 0 || idx >= len(player.Hand) {
		return model.Card{}, false
	}
	return player.Hand[idx], true
}

func (r hookRuntime) ConsumeCardByIndex(player *model.Player, idx int) (model.Card, error) {
	return r.GameEngine.consumePlayableCardByIndex(player, idx)
}

func (r hookRuntime) AddToDiscardPile(cards ...model.Card) {
	if r.GameEngine == nil || r.GameEngine.State == nil {
		return
	}
	r.GameEngine.State.DiscardPile = append(r.GameEngine.State.DiscardPile, cards...)
}

func (r hookRuntime) TakeDiscardPileCardByID(cardID string) (model.Card, bool) {
	if r.GameEngine == nil || r.GameEngine.State == nil {
		return model.Card{}, false
	}
	for i, c := range r.GameEngine.State.DiscardPile {
		if c.ID == cardID {
			r.GameEngine.State.DiscardPile = append(r.GameEngine.State.DiscardPile[:i], r.GameEngine.State.DiscardPile[i+1:]...)
			return c, true
		}
	}
	return model.Card{}, false
}

// 新增 - 中断推送（扩展）
func (r hookRuntime) PushInterruptForPlayer(playerID string, intr *model.Interrupt) {
	if r.GameEngine == nil || intr == nil {
		return
	}
	intr.PlayerID = playerID
	r.GameEngine.PushInterrupt(intr)
}

// 新增 - 回合状态
func (r hookRuntime) GetTurnStage() model.TurnStage {
	if r.GameEngine == nil || r.GameEngine.State == nil {
		return ""
	}
	return r.GameEngine.State.TurnStage
}

func (r hookRuntime) SetTurnStage(stage model.TurnStage) {
	if r.GameEngine == nil || r.GameEngine.State == nil {
		return
	}
	r.GameEngine.State.TurnStage = stage
}

func (r hookRuntime) IsPlayerActive(player *model.Player) bool {
	if player == nil {
		return false
	}
	return player.IsActive
}

// 新增 - 状态机控制
func (r hookRuntime) EnterResponseWindow() {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.enterResponseWindow()
}

func (r hookRuntime) EnterActionExecutionStage() {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.enterActionExecutionStage()
}

func (r hookRuntime) EnterDamageResolution(returnTo interface{}) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.enterDamageResolution(returnTo)
}

// Phase C - 回合/攻击 hooks 扩展实现

func (r hookRuntime) CampEnemyIDs(camp model.Camp) []string {
	if r.GameEngine == nil {
		return nil
	}
	return r.GameEngine.campEnemyIDs(camp)
}

func (r hookRuntime) RemoveExclusiveEffectCard(source *model.Player, effect model.EffectType, restoreCard bool) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.GameEngine.removeExclusiveEffectCard(source, effect, restoreCard)
}

func (r hookRuntime) CheckHandLimit(player *model.Player) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.checkHandLimit(player, nil)
}

func (r hookRuntime) HasPendingInterrupt() bool {
	if r.GameEngine == nil || r.GameEngine.State == nil {
		return false
	}
	return r.GameEngine.State.PendingInterrupt != nil
}

func (r hookRuntime) CanPayCrystalCost(playerID string, amount int) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.GameEngine.CanPayCrystalCost(playerID, amount)
}

func (r hookRuntime) DrawCardsWithOptions(playerID string, count int, opts model.DrawOptions) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.DrawCardsWithOptions(playerID, count, opts)
}

func (r hookRuntime) NotifyCardRevealed(playerID string, cards []model.Card, actionType model.DamageType) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.NotifyCardRevealed(playerID, cards, actionType)
}

func (r hookRuntime) GetAllPlayers() []*model.Player {
	if r.GameEngine == nil {
		return nil
	}
	return r.GameEngine.GetAllPlayers()
}

func (r hookRuntime) GetAllPlayersMap() map[string]*model.Player {
	if r.GameEngine == nil {
		return nil
	}
	return r.GameEngine.State.Players
}

func (r hookRuntime) HasUsedActionSkill(player *model.Player) bool {
	if player == nil {
		return false
	}
	return player.TurnState.HasUsedActionSkill
}

func (r hookRuntime) ConsumeAttackDamageRuleBonus(player *model.Player, modifierID string, action model.Action) int {
	if r.GameEngine == nil {
		return 0
	}
	return consumeAttackDamageRuleBonus(player, modifierID, action)
}

func (r hookRuntime) GetPlayerOrientation(player *model.Player) model.CharacterOrientation {
	if player == nil {
		return model.OrientationNormal
	}
	return engineplayer.EffectiveOrientation(player)
}

// 新增 - 场上效果卡查找（用于伤害钩子）
func (r hookRuntime) FindExclusiveEffectCard(source *model.Player, effect model.EffectType) (*model.Player, *model.FieldCard) {
	if r.GameEngine == nil {
		return nil, nil
	}
	return r.GameEngine.findExclusiveEffectCard(source, effect)
}

func (r hookRuntime) FindSourceEffectCard(source *model.Player, effect model.EffectType) (*model.Player, *model.FieldCard) {
	if r.GameEngine == nil {
		return nil, nil
	}
	return r.GameEngine.findSourceEffectCard(source, effect)
}

func (r hookRuntime) AttachExclusiveEffectCard(source, target *model.Player, effect model.EffectType, card model.Card) error {
	if r.GameEngine == nil {
		return nil
	}
	return r.GameEngine.attachExclusiveEffectCard(source, target, effect, card)
}

// 新增 - 伤害应用钩子所需
func (r hookRuntime) RemoveFieldCard(targetID string, effect model.EffectType) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.GameEngine.RemoveFieldCard(targetID, effect)
}

func (r hookRuntime) AsChoiceRuntime() engineplayer.ChoiceRuntime {
	if r.GameEngine == nil {
		return nil
	}
	return newRoleChoiceRuntime(r.GameEngine)
}

// 新增 - 原 PolicySpec 需要的基础设施

func (r hookRuntime) AllPlayers() map[string]*model.Player {
	if r.GameEngine == nil || r.GameEngine.State == nil {
		return nil
	}
	return r.GameEngine.State.Players
}

func (r hookRuntime) NotifyCombatCue(attackerID, targetID, cueType string) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.NotifyCombatCue(attackerID, targetID, cueType)
}

func (r hookRuntime) InitCombat(attackerID, targetID string, card *model.Card, isForcedHit, canBeResponded, ignoreShield bool, interceptTags map[model.CombatInterceptTag]bool, isCounter ...bool) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.initCombat(attackerID, targetID, card, isForcedHit, canBeResponded, ignoreShield, interceptTags, isCounter...)
}

func (r hookRuntime) DispatchOnTiming(ctx *model.Context) {
	if r.GameEngine == nil || r.GameEngine.dispatcher == nil || ctx == nil {
		return
	}
	r.GameEngine.dispatcher.OnTiming(ctx.Timing, ctx)
}

func (r hookRuntime) PendingInterrupt() *model.Interrupt {
	if r.GameEngine == nil || r.GameEngine.State == nil {
		return nil
	}
	return r.GameEngine.State.PendingInterrupt
}

func (r hookRuntime) CurrentTurnPlayerID() string {
	if r.GameEngine == nil || r.GameEngine.State == nil {
		return ""
	}
	idx := r.GameEngine.State.CurrentTurn
	if idx < 0 || idx >= len(r.GameEngine.State.PlayerOrder) {
		return ""
	}
	return r.GameEngine.State.PlayerOrder[idx]
}

func (r hookRuntime) SnapshotPlayerPoses() map[string]engineplayer.PoseSnapshot {
	if r.GameEngine == nil {
		return nil
	}
	return r.GameEngine.snapshotPlayerPoses()
}

func (r hookRuntime) DispatchOrientationChanges(before map[string]engineplayer.PoseSnapshot) {
	if r.GameEngine == nil {
		return
	}
	r.GameEngine.dispatchOrientationChanges(before)
}

func (r hookRuntime) HasPlayableAttackCard(player *model.Player) bool {
	if r.GameEngine == nil {
		return false
	}
	return r.GameEngine.hasPlayableAttackCard(player)
}
