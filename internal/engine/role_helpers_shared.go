// gameflow: 跨角色共享辅助函数（token、盖牌、士气、身份判定、形态基础设施等）。
package engine

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	butterflydancer "starcup-engine/internal/engine/player/butterfly_dancer"
	elfarcher "starcup-engine/internal/engine/player/elf_archer"
	fighter "starcup-engine/internal/engine/player/fighter"
	magiclancer "starcup-engine/internal/engine/player/magic_lancer"
	magicswordsman "starcup-engine/internal/engine/player/magic_swordsman"
	"starcup-engine/internal/model"
)

// ---- 士气辅助 ----

func (e *GameEngine) moraleFloorForCamp(camp model.Camp) int {
	floor := 0
	for _, p := range e.State.Players {
		if p == nil || p.Camp == camp {
			continue
		}
		if butterflydancer.WitherActive(p) {
			if floor < 1 {
				floor = 1
			}
		}
	}
	return floor
}

func (e *GameEngine) applyCampMoraleLoss(camp model.Camp, wantLoss int) int {
	if wantLoss <= 0 {
		return 0
	}
	current := e.campMorale(camp)
	floor := e.moraleFloorForCamp(camp)
	maxLoss := current - floor
	if maxLoss < 0 {
		maxLoss = 0
	}
	actual := wantLoss
	if actual > maxLoss {
		actual = maxLoss
	}
	if actual <= 0 {
		return 0
	}
	if camp == model.RedCamp {
		e.State.RedMorale -= actual
	} else {
		e.State.BlueMorale -= actual
	}
	return actual
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
	return player.EffectiveOrientation(p)
}

func effectivePlayerForm(p *model.Player) string {
	return player.EffectiveForm(p)
}

// ---- 引擎级形态基础设施 ----

type poseSnapshot struct {
	Orientation model.CharacterOrientation
	Form        string
}

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

// canCastMagicInAction 判断玩家在自己行动阶段能否使用法术牌。
func (e *GameEngine) canCastMagicInAction(player *model.Player) bool {
	if player == nil {
		return false
	}
	if magiclancer.BlocksMagicCasting(player) {
		return false
	}
	if fighter.BlocksMagicCasting(player) {
		return false
	}
	if magicswordsman.BlocksMagicCasting(player) {
		return false
	}
	return true
}

// canUseShadowRejectResponseMagic 判断魔剑士【暗影抗拒】是否允许在"非自己行动阶段"用法术响应。
// 规则：必须不是当前回合玩家本人。
func (e *GameEngine) canUseShadowRejectResponseMagic(player *model.Player) bool {
	if e == nil || player == nil {
		return false
	}
	currentTurnPlayerID := ""
	if len(e.State.PlayerOrder) > 0 && e.State.CurrentTurn >= 0 && e.State.CurrentTurn < len(e.State.PlayerOrder) {
		currentTurnPlayerID = e.State.PlayerOrder[e.State.CurrentTurn]
	}
	return magicswordsman.CanUseShadowRejectResponse(player, currentTurnPlayerID)
}

// reverseOrderTargetIDsFrom 按"逆向"顺序返回角色 ID（从 source 的前一位开始）。
func (e *GameEngine) reverseOrderTargetIDsFrom(sourceID string, includeSelf bool) []string {
	if len(e.State.PlayerOrder) == 0 {
		return nil
	}
	start := -1
	for i, pid := range e.State.PlayerOrder {
		if pid == sourceID {
			start = i
			break
		}
	}
	if start < 0 {
		return nil
	}
	n := len(e.State.PlayerOrder)
	var ids []string
	stepStart := 1
	stepEnd := n
	if includeSelf {
		stepStart = 0
	}
	for step := stepStart; step < stepEnd; step++ {
		idx := (start - step + n) % n
		ids = append(ids, e.State.PlayerOrder[idx])
	}
	return ids
}

func (e *GameEngine) playerOrderPosition(playerID string) int {
	if e == nil || playerID == "" {
		return 0
	}
	for i, pid := range e.State.PlayerOrder {
		if pid == playerID {
			return i + 1
		}
	}
	return 0
}

func (e *GameEngine) fighterLockedTarget(player *model.Player) *model.Player {
	return fighter.LockedTarget(newRoleChoiceRuntime(e), player)
}

func (e *GameEngine) clearFighterHundredDragon(player *model.Player, logLine string) bool {
	return fighter.ClearHundredDragon(newRoleChoiceRuntime(e), player, logLine)
}

func playableCardCount(player *model.Player) int {
	if player == nil {
		return 0
	}
	return len(player.Hand) + elfarcher.CountBlessings(player)
}

func getPlayableCardByIndex(player *model.Player, index int) (card model.Card, fromBlessing bool, blessingIndex int, ok bool) {
	if player == nil || index < 0 {
		return model.Card{}, false, -1, false
	}
	if index < len(player.Hand) {
		return player.Hand[index], false, -1, true
	}
	blessings := elfarcher.BlessingCards(player)
	bidx := index - len(player.Hand)
	if bidx < 0 || bidx >= len(blessings) {
		return model.Card{}, false, -1, false
	}
	return blessings[bidx], true, bidx, true
}

func consumePlayableCardByIndex(player *model.Player, index int) (model.Card, error) {
	card, fromBlessing, _, ok := getPlayableCardByIndex(player, index)
	if !ok {
		return model.Card{}, fmt.Errorf("无效的卡牌索引")
	}
	if fromBlessing {
		elfarcher.RemoveBlessingByCardID(player, card.ID)
		return card, nil
	}
	player.Hand = append(player.Hand[:index], player.Hand[index+1:]...)
	return card, nil
}

func findPlayableCardIndexByID(player *model.Player, cardID string) int {
	if player == nil || cardID == "" {
		return -1
	}
	for i, c := range player.Hand {
		if c.ID == cardID {
			return i
		}
	}
	base := len(player.Hand)
	for i, c := range elfarcher.BlessingCards(player) {
		if c.ID == cardID {
			return base + i
		}
	}
	return -1
}

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
