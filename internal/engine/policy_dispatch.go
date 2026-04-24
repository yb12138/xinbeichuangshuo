// gameflow: Policy Hook 注册与分派基础设施。

package engine

import (
	"fmt"
	"sort"

	engineplayer "starcup-engine/internal/engine/player"
	blazewitchplayer "starcup-engine/internal/engine/player/blaze_witch"
	magicswordsman "starcup-engine/internal/engine/player/magic_swordsman"
	onmyojiplayer "starcup-engine/internal/engine/player/onmyoji"
	"starcup-engine/internal/model"
)

type policyHookEntry struct {
	RoleID   string
	Priority int
	Hook     engineplayer.PolicyHookFunc
}

// mountRolePolicyHooks 从 RoleRegistry 收集 PolicySpecs，按策略类型分组并排序。
func mountRolePolicyHooks() map[engineplayer.PolicyType][]policyHookEntry {
	hooks := make(map[engineplayer.PolicyType][]policyHookEntry)
	for _, entry := range roleRegistry.Entries() {
		for _, spec := range entry.PolicySpecs {
			s := spec
			hooks[s.Type] = append(hooks[s.Type], policyHookEntry{
				RoleID:   entry.ID,
				Priority: s.Priority,
				Hook:     s.Hook,
			})
		}
	}
	for pt := range hooks {
		sort.Slice(hooks[pt], func(i, j int) bool {
			return hooks[pt][i].Priority < hooks[pt][j].Priority
		})
	}
	return hooks
}

// dispatchPolicyHook 按策略类型分派 Hook，首个返回 Stop/Err 即停止（短路）。
func (e *GameEngine) dispatchPolicyHook(pt engineplayer.PolicyType, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	hooks := e.policyHooks[pt]
	host := newPolicyHost(e)
	for _, entry := range hooks {
		result := entry.Hook(host, ctx)
		if result.Stop || result.Err != nil {
			return result
		}
	}
	return engineplayer.PolicyHookResult{}
}

// dispatchAllPolicyHooks 按策略类型分派 Hook，全部执行（不短路），聚合结果。
func (e *GameEngine) dispatchAllPolicyHooks(pt engineplayer.PolicyType, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	hooks := e.policyHooks[pt]
	host := newPolicyHost(e)
	aggregated := engineplayer.PolicyHookResult{}
	for _, entry := range hooks {
		result := entry.Hook(host, ctx)
		if result.Handled {
			aggregated.Handled = true
		}
		if result.Stop {
			aggregated.Stop = true
		}
		if result.Err != nil && aggregated.Err == nil {
			aggregated.Err = result.Err
		}
		if result.UseFaction {
			aggregated.UseFaction = true
		}
		if result.MaxHeal != 0 {
			aggregated.MaxHeal += result.MaxHeal
		}
		if len(result.SkillIDs) > 0 {
			aggregated.SkillIDs = append(aggregated.SkillIDs, result.SkillIDs...)
		}
	}
	return aggregated
}

// ---------- PolicyHost 适配器 ----------

type policyHostBridge struct {
	e *GameEngine
}

func newPolicyHost(e *GameEngine) engineplayer.PolicyHost {
	return &policyHostBridge{e: e}
}

func (h *policyHostBridge) Log(message string) {
	if h.e != nil {
		h.e.Log(message)
	}
}

func (h *policyHostBridge) LookupPlayer(playerID string) *model.Player {
	if h.e == nil || h.e.State == nil {
		return nil
	}
	return h.e.State.Players[playerID]
}

func (h *policyHostBridge) AllPlayers() map[string]*model.Player {
	if h.e == nil || h.e.State == nil {
		return nil
	}
	return h.e.State.Players
}

func (h *policyHostBridge) State() *model.GameState {
	if h.e == nil {
		return nil
	}
	return h.e.State
}

func (h *policyHostBridge) PlayerOrder() []string {
	if h.e == nil || h.e.State == nil {
		return nil
	}
	return h.e.State.PlayerOrder
}

func (h *policyHostBridge) CurrentTurn() int {
	if h.e == nil || h.e.State == nil {
		return 0
	}
	return h.e.State.CurrentTurn
}

func (h *policyHostBridge) PushInterrupt(intr *model.Interrupt) {
	if h.e != nil {
		h.e.PushInterrupt(intr)
	}
}

func (h *policyHostBridge) PopInterrupt() {
	if h.e != nil {
		h.e.PopInterrupt()
	}
}

func (h *policyHostBridge) NotifyCardRevealed(playerID string, cards []model.Card, actionType model.DamageType) {
	if h.e != nil {
		h.e.NotifyCardRevealed(playerID, cards, actionType)
	}
}

func (h *policyHostBridge) NotifyCombatCue(attackerID, targetID, cueType string) {
	if h.e != nil {
		h.e.NotifyCombatCue(attackerID, targetID, cueType)
	}
}

func (h *policyHostBridge) ConsumeCardByIndex(player *model.Player, idx int) (model.Card, error) {
	if h.e == nil {
		return model.Card{}, nil
	}
	return consumePlayableCardByIndex(player, idx)
}

func (h *policyHostBridge) AddToDiscardPile(cards ...model.Card) {
	if h.e != nil {
		h.e.State.DiscardPile = append(h.e.State.DiscardPile, cards...)
	}
}

func (h *policyHostBridge) InitCombat(attackerID, targetID string, card *model.Card, isForcedHit, canBeResponded, ignoreShield bool, interceptTags map[model.CombatInterceptTag]bool, isCounter ...bool) {
	if h.e != nil {
		h.e.initCombat(attackerID, targetID, card, isForcedHit, canBeResponded, ignoreShield, interceptTags, isCounter...)
	}
}

func (h *policyHostBridge) ResolveMagicBowPierceMiss(attackerID, targetID string, attackCard *model.Card, isCounter bool) {
	if h.e != nil {
		h.e.resolveMagicBowPierceMiss(attackerID, targetID, attackCard, isCounter)
	}
}

func (h *policyHostBridge) TopCombatRequest() *model.CombatRequest {
	if h.e == nil || len(h.e.State.CombatStack) == 0 {
		return nil
	}
	return &h.e.State.CombatStack[len(h.e.State.CombatStack)-1]
}

func (h *policyHostBridge) PopCombatRequest() {
	if h.e != nil && len(h.e.State.CombatStack) > 0 {
		h.e.State.CombatStack = h.e.State.CombatStack[:len(h.e.State.CombatStack)-1]
	}
}

func (h *policyHostBridge) BuildContext(user, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context {
	if h.e == nil {
		return nil
	}
	return h.e.buildContext(user, target, timing, eventCtx)
}

func (h *policyHostBridge) DispatchOnTiming(ctx *model.Context) {
	if h.e != nil && h.e.dispatcher != nil {
		h.e.dispatcher.OnTiming(ctx.Timing, ctx)
	}
}

func (h *policyHostBridge) PendingInterrupt() *model.Interrupt {
	if h.e == nil {
		return nil
	}
	return h.e.State.PendingInterrupt
}

func (h *policyHostBridge) IsCharacter(player *model.Player, roleID string) bool {
	return engineplayer.IsCharacter(player, roleID)
}

func (h *policyHostBridge) HasForm(player *model.Player, form string) bool {
	return engineplayer.HasForm(player, form)
}

func (h *policyHostBridge) GetToken(player *model.Player, key string) int {
	if player == nil || player.Tokens == nil {
		return 0
	}
	return player.Tokens[key]
}

// ---------- 阴阳师委托方法 ----------

func (h *policyHostBridge) ApplyFactionCounterBonuses(actor *model.Player, card *model.Card) {
	if h.e != nil {
		h.e.applyOnmyojiFactionCounterBonuses(actor, card)
	}
}

func (h *policyHostBridge) CanUseFactionCounter(incoming *model.Card) bool {
	return onmyojiplayer.CanUseFactionCounter(incoming)
}

// ---------- 魔剑士委托方法 ----------

func (h *policyHostBridge) CanUseShadowRejectResponse(player *model.Player, currentTurnPlayerID string) bool {
	return magicswordsman.CanUseShadowRejectResponse(player, currentTurnPlayerID)
}

// ---------- 其他角色策略委托方法 ----------

func (h *policyHostBridge) ApplyMoonMedusaInterrupt(attacker, target *model.Player, action *model.QueuedAction, ctx *model.Context) bool {
	if h.e == nil {
		return false
	}
	return attackStartMoonGoddessMedusaInterruptHook(h.e, attacker, target, action, ctx)
}

func (h *policyHostBridge) ApplyBeastSamuraiResponseSkillAugment(skillIDs []string, ctx *model.Context) []string {
	if h.e == nil {
		return skillIDs
	}
	return augmentBeastSamuraiResponseSkillIDs(h.e.dispatcher, skillIDs, ctx)
}

func (h *policyHostBridge) ApplyFighterResponseSkillNormalize(skillIDs []string, ctx *model.Context) []string {
	if h.e == nil {
		return skillIDs
	}
	return normalizeFighterResponseSkillIDs(h.e.dispatcher, skillIDs, ctx)
}

func (h *policyHostBridge) ApplyArbiterSkillPostCleanup(ctx *model.Context) {
	// 暂不实现：skillPostArbiterForcedDoomsdayCleanupHook 需要 skillUseRequest，暂用原有 buildPresenceHooks
}

func (h *policyHostBridge) ApplyAdventurerUndergroundLawOverride(player *model.Player, action model.PlayerAction) (model.PlayerAction, bool) {
	// 暂不实现：需要 model.ActionType 转换
	return action, false
}

func (h *policyHostBridge) ApplyHolyBowHolyGloryExitHook(player *model.Player, _ model.ActionType) {
	if h.e == nil || player == nil || !engineplayer.IsCharacter(player, "holy_bow") || !engineplayer.HasForm(player, model.FormHolyBowHolyGlory) {
		return
	}
	beforePoses := h.e.snapshotPlayerPoses()
	engineplayer.ClearForm(player, model.FormHolyBowHolyGlory)
	h.e.Heal(player.ID, 1)
	h.e.Log(fmt.Sprintf("%s 在圣煌形态下执行特殊行动，脱离圣煌形态并获得1点治疗", player.Name))
	h.e.dispatchOrientationChanges(beforePoses)
}

func (h *policyHostBridge) ApplyBlazeWitchAttackCardTransform(player *model.Player, card model.Card) model.Card {
	return blazewitchplayer.ApplyAttackCardTransform(player, card)
}

func (h *policyHostBridge) HandlePostDamageResolved(pd *model.PendingDamage) bool {
	if h.e == nil || pd == nil {
		return false
	}
	result := h.e.dispatchAllRoleTimingHooks(engineplayer.TimingPostDamageResolved, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		DamageType:    pd.DamageType,
		Damage:        pd.Damage,
		PendingDamage: pd,
	})
	return result.Interrupted
}
