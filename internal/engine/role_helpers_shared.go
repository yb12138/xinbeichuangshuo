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

// applyCampMoraleLoss 应用士气损失（实际扣除），先经过 MoraleLossModifier 链调整。
func (e *GameEngine) applyCampMoraleLoss(camp model.Camp, wantLoss int, extra ...engineplayer.MoraleLossModifierExtra) int {
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
	if e.State.PendingInterrupt == nil || !isDiscardSelectionInterrupt(e.State.PendingInterrupt) {
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

func (e *GameEngine) snapshotPlayerPoses() map[string]poseSnapshot {
	snapshots := make(map[string]poseSnapshot, len(e.State.Players))
	for id, p := range e.State.Players {
		snapshots[id] = poseSnapshot{
			Orientation: effectivePlayerOrientation(p),
			Form:        effectivePlayerForm(p),
		}
	}
	return snapshots
}

func (e *GameEngine) dispatchOrientationChanges(before map[string]poseSnapshot) {
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
		ctx := e.buildContext(p, p, model.TimingOnOrientationChanged, eventCtx)
		e.dispatcher.OnTiming(ctx.Timing, ctx)
	}
}

// ---- Shared card counter helpers ----

func getFieldEffectCard(player *model.Player, effect model.EffectType) *model.FieldCard {
	if player == nil {
		return nil
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != effect {
			continue
		}
		return fc
	}
	return nil
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

func (e *GameEngine) consumePlayableCardByIndex(p *model.Player, index int) (model.Card, error) {
	card, fromCover, coverEffect, ok := e.getPlayableCardByIndex(p, index)
	if !ok {
		return model.Card{}, fmt.Errorf("无效的卡牌索引")
	}
	if fromCover {
		engineplayer.RemoveCoverCardByEffectAndID(p, coverEffect, card.ID)
		return card, nil
	}
	p.Hand = append(p.Hand[:index], p.Hand[index+1:]...)
	return card, nil
}

func (e *GameEngine) findPlayableCardIndexByID(p *model.Player, cardID string) int {
	if p == nil || cardID == "" {
		return -1
	}
	for i, c := range p.Hand {
		if c.ID == cardID {
			return i
		}
	}
	offset := len(p.Hand)
	for _, effect := range e.collectPlayableCoverEffects() {
		covers := engineplayer.CoverCardsByEffect(p, effect)
		for i, fc := range covers {
			if fc.Card.ID == cardID {
				return offset + i
			}
		}
		offset += len(covers)
	}
	return -1
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

func removeCardsByIndicesFromHand(player *model.Player, indices []int) ([]model.Card, error) {
	if player == nil {
		return nil, fmt.Errorf("玩家不存在")
	}
	for _, idx := range indices {
		if idx < 0 || idx >= len(player.Hand) {
			return nil, fmt.Errorf("无效的手牌索引: %d", idx)
		}
	}
	seen := map[int]bool{}
	for _, idx := range indices {
		if seen[idx] {
			return nil, fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}
	// 从大到小删除，避免索引位移。
	for i := 0; i < len(indices); i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[i] < indices[j] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
	var removed []model.Card
	for _, idx := range indices {
		removed = append(removed, player.Hand[idx])
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}
	return removed, nil
}
