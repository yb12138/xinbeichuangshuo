// gameflow: 新角色：计数器/牌面互动（如欺诈、暗灭应战可见性）。

package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

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

func getHeroTauntCard(player *model.Player) *model.FieldCard {
	return getFieldEffectCard(player, model.EffectHeroTaunt)
}

// canCastMagicInAction 判断玩家在自己行动阶段能否使用法术牌。
func (e *GameEngine) canCastMagicInAction(player *model.Player) bool {
	if player == nil {
		return false
	}
	// 魔枪黑暗束缚：始终不能使用法术牌。
	if e.isMagicLancer(player) {
		return false
	}
	// 格斗家百式幻龙拳：形态期间不能执行法术行动。
	if e.isFighter(player) {
		if hasFighterHundredDragonForm(player) {
			return false
		}
	}
	// 魔剑士暗影抗拒：行动阶段不能使用法术牌。
	if e.isMagicSwordsman(player) {
		if hasMagicSwordsmanShadowForm(player) {
			return false
		}
	}
	return true
}

// canUseShadowRejectResponseMagic 判断魔剑士【暗影抗拒】是否允许在“非自己行动阶段”用法术响应。
// 规则：必须不是当前回合玩家本人。
func (e *GameEngine) canUseShadowRejectResponseMagic(player *model.Player) bool {
	if e == nil || player == nil {
		return false
	}
	if !e.isMagicSwordsman(player) {
		return false
	}
	if len(e.State.PlayerOrder) == 0 {
		return false
	}
	if e.State.CurrentTurn < 0 || e.State.CurrentTurn >= len(e.State.PlayerOrder) {
		return false
	}
	return e.State.PlayerOrder[e.State.CurrentTurn] != player.ID
}

// reverseOrderTargetIDsFrom 按“逆向”顺序返回角色 ID（从 source 的前一位开始）。
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
	if e == nil || player == nil || player.Tokens == nil {
		return nil
	}
	order := player.Tokens["fighter_hundred_dragon_target_order"]
	if order <= 0 || order > len(e.State.PlayerOrder) {
		return nil
	}
	return e.State.Players[e.State.PlayerOrder[order-1]]
}

func (e *GameEngine) clearFighterHundredDragon(player *model.Player, logLine string) bool {
	if e == nil || player == nil || !e.isFighter(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	beforePoses := e.snapshotPlayerPoses()
	active := hasFighterHundredDragonForm(player) || player.Tokens["fighter_hundred_dragon_target_order"] > 0
	leaveFighterHundredDragonForm(player)
	player.Tokens["fighter_hundred_dragon_target_order"] = 0
	if active && logLine != "" {
		e.Log(logLine)
	}
	e.dispatchOrientationChanges(beforePoses)
	return active
}

func markElfBlessings(player *model.Player, cards []model.Card) {
	if player == nil || len(cards) == 0 {
		return
	}
	exists := map[string]bool{}
	for _, fc := range elfBlessingCoverCards(player) {
		c := fc.Card
		if c.ID != "" {
			exists[c.ID] = true
		}
	}
	for _, c := range cards {
		if c.ID == "" || exists[c.ID] {
			continue
		}
		player.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  player.ID,
			SourceID: player.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectElfBlessing,
			Hook:     model.FieldHookManual,
		})
		exists[c.ID] = true
	}
	syncElfBlessings(player)
}

func syncElfBlessings(player *model.Player) {
	if player == nil {
		return
	}
	blessings := elfBlessingCards(player)
	player.Blessings = blessings
	blessingIDs := map[string]bool{}
	for _, c := range blessings {
		if c.ID != "" {
			blessingIDs[c.ID] = true
		}
	}
	newZone := make([]string, 0, len(player.CharaZone)+len(blessings))
	zoneHas := map[string]bool{}
	for _, z := range player.CharaZone {
		if !strings.HasPrefix(z, elfBlessingPrefix) {
			newZone = append(newZone, z)
			zoneHas[z] = true
			continue
		}
		cardID := strings.TrimPrefix(z, elfBlessingPrefix)
		if blessingIDs[cardID] {
			newZone = append(newZone, z)
			zoneHas[z] = true
		}
	}
	for _, c := range blessings {
		if c.ID == "" {
			continue
		}
		key := elfBlessingPrefix + c.ID
		if zoneHas[key] {
			continue
		}
		newZone = append(newZone, key)
	}
	player.CharaZone = newZone
}

func countElfBlessings(player *model.Player) int {
	if player == nil {
		return 0
	}
	return len(elfBlessingCoverCards(player))
}

func isElfBlessingCard(player *model.Player, cardID string) bool {
	if player == nil || cardID == "" {
		return false
	}
	for _, c := range elfBlessingCards(player) {
		if c.ID == cardID {
			return true
		}
	}
	return false
}

func removeElfBlessingByCardID(player *model.Player, cardID string) bool {
	if player == nil || cardID == "" {
		return false
	}
	removed := false
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectElfBlessing {
			continue
		}
		if !removed && fc.Card.ID == cardID {
			player.RemoveFieldCard(fc)
			removed = true
		}
		if removed {
			break
		}
	}

	target := elfBlessingPrefix + cardID
	newZone := make([]string, 0, len(player.CharaZone))
	removedZone := false
	for _, z := range player.CharaZone {
		if !removedZone && z == target {
			removedZone = true
			continue
		}
		newZone = append(newZone, z)
	}
	player.CharaZone = newZone
	if removed {
		syncElfBlessings(player)
	}
	return removed || removedZone
}

func elfBlessingHandIndices(player *model.Player) []int {
	if player == nil {
		return nil
	}
	var idxs []int
	for i := 0; i < countElfBlessings(player); i++ {
		idxs = append(idxs, i)
	}
	return idxs
}

func playableCardCount(player *model.Player) int {
	if player == nil {
		return 0
	}
	return len(player.Hand) + countElfBlessings(player)
}

func getPlayableCardByIndex(player *model.Player, index int) (card model.Card, fromBlessing bool, blessingIndex int, ok bool) {
	if player == nil || index < 0 {
		return model.Card{}, false, -1, false
	}
	if index < len(player.Hand) {
		return player.Hand[index], false, -1, true
	}
	blessings := elfBlessingCards(player)
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
		removeElfBlessingByCardID(player, card.ID)
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
	for i, c := range elfBlessingCards(player) {
		if c.ID == cardID {
			return base + i
		}
	}
	return -1
}

func getPlayableCardIndicesByType(player *model.Player, cardType model.CardType) []int {
	if player == nil {
		return nil
	}
	var out []int
	for i, c := range player.Hand {
		if c.Type == cardType {
			out = append(out, i)
		}
	}
	base := len(player.Hand)
	for i, c := range elfBlessingCards(player) {
		if c.Type == cardType {
			out = append(out, base+i)
		}
	}
	return out
}

func getPlayableCardIndicesByElement(player *model.Player, element model.Element) []int {
	if player == nil {
		return nil
	}
	var out []int
	for i, c := range player.Hand {
		if c.Element == element {
			out = append(out, i)
		}
	}
	base := len(player.Hand)
	for i, c := range elfBlessingCards(player) {
		if c.Element == element {
			out = append(out, base+i)
		}
	}
	return out
}

func elfBlessingCoverCards(player *model.Player) []*model.FieldCard {
	if player == nil {
		return nil
	}
	return player.GetCoverCardsByEffect(model.EffectElfBlessing)
}

func elfBlessingCards(player *model.Player) []model.Card {
	covers := elfBlessingCoverCards(player)
	if len(covers) == 0 {
		return nil
	}
	out := make([]model.Card, 0, len(covers))
	for _, fc := range covers {
		if fc == nil {
			continue
		}
		out = append(out, fc.Card)
	}
	return out
}

func getCardIndicesByType(player *model.Player, cardType model.CardType) []int {
	if player == nil {
		return nil
	}
	var out []int
	for i, c := range player.Hand {
		if c.Type == cardType {
			out = append(out, i)
		}
	}
	return out
}

func getCardIndicesByElement(player *model.Player, element model.Element) []int {
	if player == nil {
		return nil
	}
	var out []int
	for i, c := range player.Hand {
		if c.Element == element {
			out = append(out, i)
		}
	}
	return out
}

func getSameElementCounts(player *model.Player) map[model.Element]int {
	out := map[model.Element]int{}
	if player == nil {
		return out
	}
	for _, c := range player.Hand {
		if c.Element == "" {
			continue
		}
		out[c.Element]++
	}
	return out
}

func hasPendingActionSource(player *model.Player, source string) bool {
	if player == nil || source == "" {
		return false
	}
	for _, act := range player.TurnState.PendingActions {
		if act.Source == source {
			return true
		}
	}
	return false
}

func clearElfElementalShotCombatState(player *model.Player) {
	if player == nil {
		return
	}
	if player.TurnState.SkillFlowState == nil {
		player.TurnState.SkillFlowState = make(map[string]int)
	}
	player.TurnState.SkillFlowState["elf_elemental_shot_water_pending"] = 0
	player.TurnState.SkillFlowState["elf_elemental_shot_earth_pending"] = 0
	player.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] = 0
}

func (e *GameEngine) queueElfAnimalResponse(source, target *model.Player, pd *model.PendingDamage) bool {
	if e == nil || e.dispatcher == nil || source == nil || target == nil || pd == nil {
		return false
	}
	if !e.isElfArcher(source) || !source.IsActive {
		return false
	}
	if pd.TargetID == "" || pd.TargetID == source.ID {
		return false
	}
	if !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) || pd.Card == nil || pd.IsCounter {
		return false
	}

	damageVal := pd.Damage
	ctx := e.buildContext(source, target, model.TimingOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  pd.SourceID,
		TargetID:  pd.TargetID,
		DamageVal: &damageVal,
		Card:      pd.Card,
		AttackInfo: &model.AttackEventInfo{
			ActionType: "Attack",
			IsHit:      true,
		},
	})

	skillIDs := make([]string, 0, 2)
	for _, skillID := range []string{"elf_animal_companion", "elf_pet_empower"} {
		if e.dispatcher.isSkillStillUsable(skillID, source, ctx) {
			skillIDs = append(skillIDs, skillID)
		}
	}
	if len(skillIDs) == 0 {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: source.ID,
		SkillIDs: skillIDs,
		Context:  ctx,
	})
	e.Log(fmt.Sprintf("%s 的 [动物伙伴] 响应窗口开启", source.Name))
	return true
}

func (e *GameEngine) canPayOnmyojiBindingCost(camp model.Camp) bool {
	gems := e.GetCampGems(string(camp))
	crystals := e.GetCampCrystals(string(camp))
	// 需求：严格消耗 1 红宝石 + 1 蓝水晶（不允许替代）。
	return gems >= 1 && crystals >= 1
}

func (e *GameEngine) payOnmyojiBindingCost(camp model.Camp) bool {
	if !e.canPayOnmyojiBindingCost(camp) {
		return false
	}
	// 严格扣除 1 红宝石 + 1 蓝水晶。
	e.ModifyGem(string(camp), -1)
	e.ModifyCrystal(string(camp), -1)
	return true
}

func onmyojiCanUseFactionCounter(incoming *model.Card) bool {
	if incoming == nil {
		return false
	}
	// 欺诈视为攻击但无命格，不可触发阴阳转换。
	if incoming.Name == "欺诈" {
		return false
	}
	return incoming.Faction != ""
}

func collectOnmyojiCounterOptions(player *model.Player, incoming *model.Card) []map[string]interface{} {
	if player == nil || incoming == nil {
		return nil
	}
	var options []map[string]interface{}
	for i, c := range player.Hand {
		if c.Type != model.CardTypeAttack {
			continue
		}
		useFaction := false
		canCounter := false
		if c.Element == incoming.Element || c.Element == model.ElementDark {
			canCounter = true
		}
		if !canCounter && onmyojiCanUseFactionCounter(incoming) && c.Faction != "" && c.Faction == incoming.Faction {
			canCounter = true
			useFaction = true
		}
		if !canCounter {
			continue
		}
		label := fmt.Sprintf("%d: %s", i+1, formatCardInfo(c))
		if useFaction {
			label += "（阴阳转换）"
		}
		options = append(options, map[string]interface{}{
			"card_id":     c.ID,
			"card_index":  i,
			"use_faction": useFaction,
			"label":       label,
		})
	}
	return options
}

func bloodLimit(player *model.Player) int {
	if player == nil {
		return 3
	}
	ensurePlayerTokensMap(player)
	if v := player.Tokens["css_blood_cap"]; v > 0 {
		return v
	}
	return 3
}

func addBlood(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "css_blood", delta, bloodLimit(player))
}

func (e *GameEngine) isRoseCourtyardActive() bool {
	for _, p := range e.State.Players {
		if p == nil || !e.isCrimsonSwordSpirit(p) {
			continue
		}
		for _, fc := range p.Field {
			if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectRoseCourtyard {
				return true
			}
		}
	}
	return false
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
