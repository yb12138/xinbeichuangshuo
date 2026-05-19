// gameflow: 跨角色共享辅助函数（token、盖牌、士气、身份判定、形态基础设施等）。
package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ---- 士气辅助 ----

// capMoraleLoss 计算士气损失上限（不实际扣除），用于弃牌结算的前置判断。
func (e *GameEngine) capMoraleLoss(camp model.Camp, wantLoss int, extra ...engineplayer.MoraleLossModifierExtra) int {
	if wantLoss <= 0 {
		return 0
	}
	current := e.campMorale(camp)
	loss := wantLoss
	var ex engineplayer.MoraleLossModifierExtra
	if len(extra) > 0 {
		ex = extra[0]
	}
	for _, entry := range roleRegistry.Entries() {
		if entry.MoraleLossModifier != nil {
			loss = entry.MoraleLossModifier(e, camp, current, loss, ex)
		}
	}
	if loss < 0 {
		loss = 0
	}
	if current-loss < 0 {
		loss = current
	}
	if loss <= 0 {
		return 0
	}
	return loss
}

// ApplyCampMoraleLoss 应用士气损失（实际扣除），委托 capMoraleLoss 计算实际扣除量。
func (e *GameEngine) ApplyCampMoraleLoss(camp model.Camp, wantLoss int, extra ...engineplayer.MoraleLossModifierExtra) int {
	loss := e.capMoraleLoss(camp, wantLoss, extra...)
	if loss <= 0 {
		return 0
	}
	if camp == model.RedCamp {
		e.State.RedMorale -= loss
	} else {
		e.State.BlueMorale -= loss
	}
	return loss
}

func (e *GameEngine) addCampMorale(camp model.Camp, amount int) int {
	if amount <= 0 {
		return 0
	}
	current := e.campMorale(camp)
	if current >= standardCampMoraleCapEngine {
		return 0
	}
	actual := amount
	room := standardCampMoraleCapEngine - current
	if actual > room {
		actual = room
	}
	if actual <= 0 {
		return 0
	}
	if camp == model.RedCamp {
		e.State.RedMorale += actual
	} else {
		e.State.BlueMorale += actual
	}
	return actual
}

func (e *GameEngine) campMorale(camp model.Camp) int {
	if camp == model.RedCamp {
		return e.State.RedMorale
	}
	return e.State.BlueMorale
}

func (e *GameEngine) pendingDiscardVictimID() string {
	if e.State.PendingInterrupt == nil || !IsDiscardSelectionInterrupt(e.State.PendingInterrupt) {
		return ""
	}
	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return ""
	}
	victimID, _ := data["victim_id"].(string)
	return victimID
}

// ---- Token cap / 士气上限常量 ----

const standardCampMoraleCapEngine = 15

// ---- 形态基础设施（委托到 player 包） ----

func effectivePlayerOrientation(p *model.Player) model.CharacterOrientation {
	return engineplayer.EffectiveOrientation(p)
}

func effectivePlayerForm(p *model.Player) string {
	return engineplayer.EffectiveForm(p)
}

// ---- 引擎级形态基础设施 ----

type poseSnapshot = engineplayer.PoseSnapshot

func (e *GameEngine) SnapshotPlayerPoses() map[string]poseSnapshot {
	snapshots := make(map[string]poseSnapshot, len(e.State.Players))
	for id, p := range e.State.Players {
		snapshots[id] = poseSnapshot{
			Orientation: effectivePlayerOrientation(p),
			Form:        effectivePlayerForm(p),
		}
	}
	return snapshots
}

func (e *GameEngine) DispatchOrientationChanges(before map[string]poseSnapshot) {
	if e == nil || len(before) == 0 {
		return
	}
	orderedIDs := append([]string{}, e.State.PlayerOrder...)
	seen := make(map[string]bool, len(orderedIDs))
	for _, id := range orderedIDs {
		seen[id] = true
	}
	for id := range e.State.Players {
		if !seen[id] {
			orderedIDs = append(orderedIDs, id)
		}
	}
	for _, playerID := range orderedIDs {
		p := e.State.Players[playerID]
		if p == nil {
			continue
		}
		prev := before[playerID]
		current := poseSnapshot{
			Orientation: effectivePlayerOrientation(p),
			Form:        effectivePlayerForm(p),
		}
		if prev == current {
			continue
		}
		eventCtx := &model.EventContext{
			Type:            model.EventNone,
			SourceID:        playerID,
			TargetID:        playerID,
			OperatorID:      playerID,
			PrevOrientation: prev.Orientation,
			NewOrientation:  current.Orientation,
			PrevForm:        prev.Form,
			NewForm:         current.Form,
		}
		ctx := e.BuildContext(p, p, model.TimingOnOrientationChanged, eventCtx)
		e.dispatcher.OnTiming(ctx.Timing, ctx)
	}
}

// ---- 行动类型限制 ----

// isActionTypeBlocked 判断玩家的行动类型是否被角色能力限制。
func (e *GameEngine) isActionTypeBlocked(p *model.Player, actionType model.ActionType) bool {
	for _, entry := range roleRegistry.Entries() {
		if entry.BlocksActionType != nil && entry.BlocksActionType(p, actionType) {
			return true
		}
	}
	return false
}

// canCastMagicInAction 判断玩家在自己行动阶段能否使用法术牌。
func (e *GameEngine) canCastMagicInAction(p *model.Player) bool {
	if p == nil {
		return false
	}
	return !e.isActionTypeBlocked(p, model.ActionMagic)
}

// ---- 可打牌（手牌 + 可打盖牌） ----

// collectPlayableCoverEffects 收集所有角色声明的可打盖牌效果类型。
func (e *GameEngine) collectPlayableCoverEffects() []model.EffectType {
	var effects []model.EffectType
	for _, entry := range roleRegistry.Entries() {
		effects = append(effects, entry.PlayableCoverEffects...)
	}
	return effects
}

func (e *GameEngine) playableCardCount(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := len(p.Hand)
	for _, effect := range e.collectPlayableCoverEffects() {
		count += engineplayer.CoverCountByEffect(p, effect)
	}
	return count
}

func (e *GameEngine) getPlayableCardByIndex(p *model.Player, index int) (card model.Card, fromCover bool, coverEffect model.EffectType, ok bool) {
	if p == nil || index < 0 {
		return model.Card{}, false, "", false
	}
	if index < len(p.Hand) {
		return p.Hand[index], false, "", true
	}
	offset := index - len(p.Hand)
	for _, effect := range e.collectPlayableCoverEffects() {
		covers := engineplayer.CoverCardsByEffect(p, effect)
		if offset < len(covers) {
			return covers[offset].Card, true, effect, true
		}
		offset -= len(covers)
	}
	return model.Card{}, false, "", false
}

func (e *GameEngine) getPlayableCardByID(p *model.Player, cardID string) (card model.Card, fromCover bool, coverEffect model.EffectType, ok bool) {
	if p == nil || cardID == "" {
		return model.Card{}, false, "", false
	}
	for _, card := range p.Hand {
		if card.ID == cardID {
			return card, false, "", true
		}
	}
	for _, effect := range e.collectPlayableCoverEffects() {
		for _, fc := range engineplayer.CoverCardsByEffect(p, effect) {
			if fc != nil && fc.Card.ID == cardID {
				return fc.Card, true, effect, true
			}
		}
	}
	return model.Card{}, false, "", false
}

func (e *GameEngine) consumePlayableCardByID(p *model.Player, cardID string) (model.Card, error) {
	card, fromCover, coverEffect, ok := e.getPlayableCardByID(p, cardID)
	if !ok {
		return model.Card{}, fmt.Errorf("未找到卡牌: %s", cardID)
	}
	if fromCover {
		engineplayer.RemoveCoverCardByEffectAndID(p, coverEffect, card.ID)
		return card, nil
	}
	for i, handCard := range p.Hand {
		if handCard.ID != cardID {
			continue
		}
		p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
		return handCard, nil
	}
	return model.Card{}, fmt.Errorf("未找到卡牌: %s", cardID)
}

func (e *GameEngine) consumeQueuedActionCard(p *model.Player, qa *model.QueuedAction) (model.Card, error) {
	cardID := queuedActionCardID(qa)
	if cardID == "" {
		return model.Card{}, fmt.Errorf("队列行动缺少卡牌ID")
	}
	return e.consumePlayableCardByID(p, cardID)
}

func queuedActionCardID(qa *model.QueuedAction) string {
	if qa == nil {
		return ""
	}
	if qa.CardID != "" {
		return qa.CardID
	}
	if qa.Card != nil {
		return qa.Card.ID
	}
	return ""
}

// ---- 其他共享辅助 ----

func (e *GameEngine) canUseHealToResist(target *model.Player, sourceID string, damageType model.DamageType, ignoreHeal bool, allowCrimsonFaithHeal bool) bool {
	if target == nil || target.Heal <= 0 {
		return false
	}
	if ignoreHeal {
		return false
	}
	_ = sourceID
	_ = damageType
	_ = allowCrimsonFaithHeal
	return true
}
