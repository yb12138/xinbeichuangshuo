package engine

import (
	"fmt"
	"starcup-engine/internal/model"
	"strings"
)

const elfBlessingPrefix = "elf_blessing:"

func isCharacter(player *model.Player, charID string) bool {
	return player != nil && player.Character != nil && player.Character.ID == charID
}

func (e *GameEngine) isElfArcher(player *model.Player) bool {
	return isCharacter(player, "elf_archer")
}

func (e *GameEngine) isPlagueMage(player *model.Player) bool {
	return isCharacter(player, "plague_mage")
}

func (e *GameEngine) isMagicSwordsman(player *model.Player) bool {
	return isCharacter(player, "magic_swordsman")
}

func (e *GameEngine) isCrimsonSwordSpirit(player *model.Player) bool {
	return isCharacter(player, "crimson_sword_spirit")
}

func (e *GameEngine) isPrayerMaster(player *model.Player) bool {
	return isCharacter(player, "prayer_master")
}

func (e *GameEngine) isCrimsonKnight(player *model.Player) bool {
	return isCharacter(player, "crimson_knight")
}

func (e *GameEngine) isWarHomunculus(player *model.Player) bool {
	return isCharacter(player, "war_homunculus")
}

func (e *GameEngine) isPriest(player *model.Player) bool {
	return isCharacter(player, "priest")
}

func (e *GameEngine) isOnmyoji(player *model.Player) bool {
	return isCharacter(player, "onmyoji")
}

func (e *GameEngine) isBlazeWitch(player *model.Player) bool {
	return isCharacter(player, "blaze_witch")
}

func (e *GameEngine) isSage(player *model.Player) bool {
	return isCharacter(player, "sage")
}

func (e *GameEngine) isMagicBow(player *model.Player) bool {
	return isCharacter(player, "magic_bow")
}

func (e *GameEngine) isMagicLancer(player *model.Player) bool {
	return isCharacter(player, "magic_lancer")
}

func (e *GameEngine) isSpiritCaster(player *model.Player) bool {
	return isCharacter(player, "spirit_caster")
}

func (e *GameEngine) isBard(player *model.Player) bool {
	return isCharacter(player, "bard")
}

func (e *GameEngine) isHero(player *model.Player) bool {
	return isCharacter(player, "hero")
}

func (e *GameEngine) isFighter(player *model.Player) bool {
	return isCharacter(player, "fighter")
}

func (e *GameEngine) isHolyBow(player *model.Player) bool {
	return isCharacter(player, "holy_bow")
}

func (e *GameEngine) isHolyLancer(player *model.Player) bool {
	return isCharacter(player, "holy_lancer")
}

func (e *GameEngine) isSoulSorcerer(player *model.Player) bool {
	return isCharacter(player, "soul_sorcerer")
}

func (e *GameEngine) isMoonGoddess(player *model.Player) bool {
	return isCharacter(player, "moon_goddess")
}

func (e *GameEngine) isBloodPriestess(player *model.Player) bool {
	return isCharacter(player, "blood_priestess")
}

func (e *GameEngine) isButterflyDancer(player *model.Player) bool {
	return isCharacter(player, "butterfly_dancer")
}

const magicBowChargeCapEngine = 8
const spiritCasterPowerCapEngine = 2
const bardInspirationCapEngine = 3
const heroTokenCapEngine = 4
const holyBowFaithCapEngine = 10
const holyBowCannonCapEngine = 1
const soulSorcererBlueCapEngine = 6
const soulSorcererYellowCapEngine = 6
const moonGoddessNewMoonCapEngine = 2
const moonGoddessPetrifyCapEngine = 3
const butterflyCocoonCapEngine = 8

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
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		if player.Tokens["fighter_hundred_dragon_form"] > 0 {
			return false
		}
	}
	// 魔剑士暗影抗拒：行动阶段不能使用法术牌。
	if e.isMagicSwordsman(player) {
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		if player.Tokens["ms_shadow_form"] > 0 {
			return false
		}
	}
	return true
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

func markElfBlessings(player *model.Player, cards []model.Card) {
	if player == nil || len(cards) == 0 {
		return
	}
	if player.Blessings == nil {
		player.Blessings = make([]model.Card, 0)
	}
	exists := map[string]bool{}
	for _, c := range player.Blessings {
		if c.ID != "" {
			exists[c.ID] = true
		}
	}
	for _, c := range cards {
		if c.ID == "" || exists[c.ID] {
			continue
		}
		player.Blessings = append(player.Blessings, c)
		exists[c.ID] = true
	}
	syncElfBlessings(player)
}

func syncElfBlessings(player *model.Player) {
	if player == nil {
		return
	}
	blessingIDs := map[string]bool{}
	for _, c := range player.Blessings {
		if c.ID != "" {
			blessingIDs[c.ID] = true
		}
	}
	newZone := make([]string, 0, len(player.CharaZone)+len(player.Blessings))
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
	for _, c := range player.Blessings {
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
	return len(player.Blessings)
}

func isElfBlessingCard(player *model.Player, cardID string) bool {
	if player == nil || cardID == "" {
		return false
	}
	for _, c := range player.Blessings {
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
	newBlessings := make([]model.Card, 0, len(player.Blessings))
	for _, c := range player.Blessings {
		if !removed && c.ID == cardID {
			removed = true
			continue
		}
		newBlessings = append(newBlessings, c)
	}
	player.Blessings = newBlessings

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
	return removed || removedZone
}

func elfBlessingHandIndices(player *model.Player) []int {
	if player == nil {
		return nil
	}
	var idxs []int
	for i := range player.Blessings {
		idxs = append(idxs, i)
	}
	return idxs
}

func playableCardCount(player *model.Player) int {
	if player == nil {
		return 0
	}
	return len(player.Hand) + len(player.Blessings)
}

func getPlayableCardByIndex(player *model.Player, index int) (card model.Card, fromBlessing bool, blessingIndex int, ok bool) {
	if player == nil || index < 0 {
		return model.Card{}, false, -1, false
	}
	if index < len(player.Hand) {
		return player.Hand[index], false, -1, true
	}
	bidx := index - len(player.Hand)
	if bidx < 0 || bidx >= len(player.Blessings) {
		return model.Card{}, false, -1, false
	}
	return player.Blessings[bidx], true, bidx, true
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
	for i, c := range player.Blessings {
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
	for i, c := range player.Blessings {
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
	for i, c := range player.Blessings {
		if c.Element == element {
			out = append(out, base+i)
		}
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

func (e *GameEngine) buildCombatEffectHints(combatReq model.CombatRequest, attacker *model.Player) []string {
	hints := make([]string, 0, 6)
	appendHint := func(s string) {
		if s == "" {
			return
		}
		for _, existing := range hints {
			if existing == s {
				return
			}
		}
		hints = append(hints, s)
	}

	if combatReq.Card != nil && combatReq.Card.Element == model.ElementDark {
		appendHint("暗系攻击：无法应战，只能防御或承受伤害。")
	}

	if !combatReq.CanBeResponded && (combatReq.Card == nil || combatReq.Card.Element != model.ElementDark) {
		if attacker != nil && e.isElfArcher(attacker) && attacker.Tokens["elf_elemental_shot_thunder_pending"] > 0 {
			appendHint("精灵射手发动了[元素射击·雷之矢]：此攻击无法应战。")
		} else {
			appendHint("技能效果：此攻击无法应战。")
		}
	}

	if attacker != nil && attacker.Tokens != nil && attacker.Tokens["mb_magic_pierce_pending"] > 0 {
		appendHint("魔弓发动了[魔贯冲击]：本次攻击伤害额外+1。")
	}
	if attacker != nil && e.isMagicLancer(attacker) {
		if attacker.TurnState.UsedSkillCounts["ml_dark_release_next_attack_bonus"] > 0 {
			appendHint("魔枪发动了[暗之解放]：本回合下一次主动攻击伤害额外+1。")
		}
		if bonus := attacker.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"]; bonus > 0 {
			appendHint(fmt.Sprintf("魔枪[充盈]增幅：本回合下一次主动攻击伤害额外+%d。", bonus))
		}
	}

	if attacker == nil || !e.isElfArcher(attacker) {
		return hints
	}
	if attacker.Tokens == nil {
		attacker.Tokens = map[string]int{}
	}
	if attacker.Tokens["elf_elemental_shot_fire_pending"] > 0 {
		appendHint("元素射击·火之矢：本次攻击伤害额外+1。")
	}
	if attacker.Tokens["elf_elemental_shot_water_pending"] > 0 {
		appendHint("元素射击·水之矢：若本次主动攻击命中，目标+1治疗。")
	}
	if attacker.Tokens["elf_elemental_shot_earth_pending"] > 0 {
		appendHint("元素射击·地之矢：若本次主动攻击命中，目标额外受到1点法术伤害。")
	}
	if hasPendingActionSource(attacker, "风之矢") {
		appendHint("元素射击·风之矢：本次攻击行动结束后，精灵射手额外获得1次攻击行动。")
	}
	return hints
}

func bloodLimit(player *model.Player) int {
	if player == nil {
		return 3
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	if v := player.Tokens["css_blood_cap"]; v > 0 {
		return v
	}
	return 3
}

func addBlood(player *model.Player, delta int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	cur := player.Tokens["css_blood"]
	cur += delta
	if cur < 0 {
		cur = 0
	}
	maxV := bloodLimit(player)
	if cur > maxV {
		cur = maxV
	}
	player.Tokens["css_blood"] = cur
	return cur
}

func (e *GameEngine) isRoseCourtyardActive() bool {
	for _, p := range e.State.Players {
		if p == nil || !e.isCrimsonSwordSpirit(p) {
			continue
		}
		if p.Tokens != nil && p.Tokens["css_rose_courtyard_active"] > 0 {
			return true
		}
		for _, fc := range p.Field {
			if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectRoseCourtyard {
				return true
			}
		}
	}
	return false
}

func (e *GameEngine) canUseHealToResist(target *model.Player, sourceID string, damageType string, ignoreHeal bool, allowCrimsonFaithHeal bool) bool {
	if target == nil || target.Heal <= 0 {
		return false
	}
	if ignoreHeal {
		return false
	}
	if e.isRoseCourtyardActive() {
		return false
	}
	// 红莲骑士：仅允许“腥红信仰白名单”中的自伤使用治疗抵御。
	if e.isCrimsonKnight(target) {
		if target.ID != sourceID {
			return false
		}
		if !allowCrimsonFaithHeal {
			return false
		}
	}
	// 瘟疫法师圣渎：攻击伤害不可用治疗抵挡，法术伤害可以。
	if e.isPlagueMage(target) {
		if strings.EqualFold(damageType, "Attack") {
			return false
		}
	}
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

func magicBowChargeCount(player *model.Player, element model.Element) int {
	if player == nil {
		return 0
	}
	count := 0
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMagicBowCharge {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

func syncMagicBowChargeToken(player *model.Player) {
	if player == nil {
		return
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	player.Tokens["mb_charge_count"] = magicBowChargeCount(player, "")
}

func addMagicBowChargeCards(player *model.Player, cards []model.Card) int {
	if player == nil || len(cards) == 0 {
		return 0
	}
	room := magicBowChargeCapEngine - magicBowChargeCount(player, "")
	if room <= 0 {
		return 0
	}
	added := 0
	for _, c := range cards {
		if added >= room {
			break
		}
		player.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  player.ID,
			SourceID: player.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectMagicBowCharge,
		})
		added++
	}
	syncMagicBowChargeToken(player)
	return added
}

func removeMagicBowChargeByElement(player *model.Player, element model.Element) (model.Card, bool) {
	if player == nil {
		return model.Card{}, false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMagicBowCharge {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		card := fc.Card
		player.RemoveFieldCard(fc)
		syncMagicBowChargeToken(player)
		return card, true
	}
	return model.Card{}, false
}

func spiritCasterPowerCovers(player *model.Player) []*model.FieldCard {
	if player == nil {
		return nil
	}
	var out []*model.FieldCard
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSpiritCasterPower {
			continue
		}
		out = append(out, fc)
	}
	return out
}

func spiritCasterPowerCount(player *model.Player, element model.Element) int {
	count := 0
	for _, fc := range spiritCasterPowerCovers(player) {
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

func syncSpiritCasterPowerToken(player *model.Player) {
	if player == nil {
		return
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	player.Tokens["sc_power_count"] = spiritCasterPowerCount(player, "")
}

func addSpiritCasterPowerCard(player *model.Player, card model.Card) bool {
	if player == nil {
		return false
	}
	if spiritCasterPowerCount(player, "") >= spiritCasterPowerCapEngine {
		return false
	}
	player.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  player.ID,
		SourceID: player.ID,
		Mode:     model.FieldCover,
		Effect:   model.EffectSpiritCasterPower,
	})
	syncSpiritCasterPowerToken(player)
	return true
}

func removeSpiritCasterPowerByCardID(player *model.Player, cardID string) (model.Card, bool) {
	if player == nil || cardID == "" {
		return model.Card{}, false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSpiritCasterPower {
			continue
		}
		if fc.Card.ID != cardID {
			continue
		}
		card := fc.Card
		player.RemoveFieldCard(fc)
		syncSpiritCasterPowerToken(player)
		return card, true
	}
	return model.Card{}, false
}

func butterflyPupa(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := player.Tokens["bt_pupa"]
	if v < 0 {
		v = 0
	}
	player.Tokens["bt_pupa"] = v
	return v
}

func addButterflyPupa(player *model.Player, delta int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := butterflyPupa(player) + delta
	if v < 0 {
		v = 0
	}
	player.Tokens["bt_pupa"] = v
	return v
}

func butterflyCocoonCovers(player *model.Player) []*model.FieldCard {
	if player == nil {
		return nil
	}
	var out []*model.FieldCard
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		out = append(out, fc)
	}
	return out
}

func butterflyCocoonCount(player *model.Player) int {
	count := len(butterflyCocoonCovers(player))
	if player != nil {
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		player.Tokens["bt_cocoon_count"] = count
	}
	return count
}

func syncButterflyCocoonToken(player *model.Player) {
	_ = butterflyCocoonCount(player)
}

func addButterflyCocoonCards(player *model.Player, cards []model.Card) int {
	if player == nil || len(cards) == 0 {
		return 0
	}
	added := 0
	for _, c := range cards {
		player.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  player.ID,
			SourceID: player.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectButterflyCocoon,
			Trigger:  model.EffectTriggerManual,
		})
		added++
	}
	syncButterflyCocoonToken(player)
	return added
}

func butterflyCocoonFieldIndices(player *model.Player) []int {
	if player == nil {
		return nil
	}
	var out []int
	for i, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		out = append(out, i)
	}
	return out
}

func removeButterflyCocoonByFieldIndex(player *model.Player, fieldIdx int) (model.Card, bool) {
	if player == nil || fieldIdx < 0 || fieldIdx >= len(player.Field) {
		return model.Card{}, false
	}
	fc := player.Field[fieldIdx]
	if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
		return model.Card{}, false
	}
	card := fc.Card
	player.RemoveFieldCard(fc)
	syncButterflyCocoonToken(player)
	return card, true
}

func removeButterflyCocoonByFieldIndices(player *model.Player, indices []int) ([]model.Card, error) {
	if player == nil {
		return nil, fmt.Errorf("玩家不存在")
	}
	if len(indices) == 0 {
		return nil, nil
	}
	seen := map[int]bool{}
	for _, idx := range indices {
		if idx < 0 || idx >= len(player.Field) {
			return nil, fmt.Errorf("无效的茧索引: %d", idx)
		}
		if seen[idx] {
			return nil, fmt.Errorf("不能重复选择同一个茧")
		}
		seen[idx] = true
		fc := player.Field[idx]
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			return nil, fmt.Errorf("选择的索引不是茧: %d", idx)
		}
	}
	// 从大到小删除，避免索引偏移。
	for i := 0; i < len(indices); i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[i] < indices[j] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
	var removed []model.Card
	for _, idx := range indices {
		fc := player.Field[idx]
		removed = append(removed, fc.Card)
		player.RemoveFieldCard(fc)
	}
	syncButterflyCocoonToken(player)
	return removed, nil
}

func butterflyMirrorPairDefs(player *model.Player) ([]string, []string) {
	if player == nil {
		return nil, nil
	}
	elemToFieldIdx := map[model.Element][]int{}
	for i, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		elemToFieldIdx[fc.Card.Element] = append(elemToFieldIdx[fc.Card.Element], i)
	}
	elements := []model.Element{
		model.ElementFire, model.ElementWater, model.ElementWind, model.ElementThunder,
		model.ElementEarth, model.ElementLight, model.ElementDark,
	}
	var defs []string
	var labels []string
	for _, ele := range elements {
		idxs := elemToFieldIdx[ele]
		if len(idxs) < 2 {
			continue
		}
		for i := 0; i < len(idxs); i++ {
			for j := i + 1; j < len(idxs); j++ {
				left := idxs[i]
				right := idxs[j]
				defs = append(defs, fmt.Sprintf("%d,%d", left, right))
				lc := player.Field[left].Card
				rc := player.Field[right].Card
				labels = append(labels, fmt.Sprintf("%s系茧：%s + %s", elementNameForPrompt(string(ele)), formatCardInfo(lc), formatCardInfo(rc)))
			}
		}
	}
	return defs, labels
}

func (e *GameEngine) butterflyEnemyIDs(user *model.Player) []string {
	if user == nil {
		return nil
	}
	var out []string
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp == user.Camp {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

func (e *GameEngine) queueButterflyWitherTrigger(user *model.Player) {
	if user == nil || !e.isButterflyDancer(user) {
		return
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	user.Tokens["bt_wither_pending"]++
	if user.Tokens["bt_wither_pending"] > 1 {
		// 已有待处理的凋零询问，累计即可。
		return
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_wither_confirm",
			"user_id":     user.ID,
			"target_ids":  e.butterflyEnemyIDs(user),
		},
	})
	e.Log(fmt.Sprintf("%s 可发动 [凋零]：请选择是否发动", user.Name))
}

func (e *GameEngine) moraleFloorForCamp(camp model.Camp) int {
	floor := 0
	for _, p := range e.State.Players {
		if p == nil || !e.isButterflyDancer(p) || p.Tokens == nil {
			continue
		}
		if p.Camp == camp {
			continue
		}
		if p.Tokens["bt_wither_active"] > 0 {
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

// maybeTriggerButterflyDamageResponses 在伤害正式应用前处理蝶舞者的时点响应（朝圣/毒粉/镜花水月）。
// 返回 true 表示已产生中断，状态机应暂停等待玩家输入。
func (e *GameEngine) maybeTriggerButterflyDamageResponses(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 {
		return false
	}
	// 朝圣：承伤者若为蝶舞者，可在每次伤害中询问一次。
	if !pd.ButterflyPilgrimageChecked {
		pd.ButterflyPilgrimageChecked = true
		target := e.State.Players[pd.TargetID]
		if target != nil && e.isButterflyDancer(target) && butterflyCocoonCount(target) > 0 {
			indices := butterflyCocoonFieldIndices(target)
			if len(indices) > 0 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: target.ID,
					Context: map[string]interface{}{
						"choice_type":    "bt_pilgrimage_pick",
						"user_id":        target.ID,
						"source_id":      pd.SourceID,
						"target_id":      pd.TargetID,
						"damage_index":   0,
						"cocoon_indices": indices,
					},
				})
				e.Log(fmt.Sprintf("%s 的 [朝圣] 可触发：是否移除1个茧抵御1点伤害", target.Name))
				return true
			}
		}
	}
	// 毒粉/镜花水月仅作用于法术伤害。
	if !isMagicLikeDamageType(pd.DamageType) {
		return false
	}

	// 毒粉/镜花水月：按“实际法术伤害”值检查，仅询问一次。
	if pd.ButterflyStage5Checked {
		return false
	}
	pd.ButterflyStage5Checked = true

	if pd.Damage == 1 {
		for _, pid := range e.State.PlayerOrder {
			user := e.State.Players[pid]
			if user == nil || !e.isButterflyDancer(user) || butterflyCocoonCount(user) <= 0 {
				continue
			}
			indices := butterflyCocoonFieldIndices(user)
			if len(indices) == 0 {
				continue
			}
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":    "bt_poison_pick",
					"user_id":        user.ID,
					"source_id":      pd.SourceID,
					"target_id":      pd.TargetID,
					"damage_index":   0,
					"cocoon_indices": indices,
				},
			})
			e.Log(fmt.Sprintf("%s 的 [毒粉] 可触发：是否移除1个茧令该次法术伤害+1", user.Name))
			return true
		}
		return false
	}

	if pd.Damage == 2 {
		for _, pid := range e.State.PlayerOrder {
			user := e.State.Players[pid]
			if user == nil || !e.isButterflyDancer(user) || butterflyCocoonCount(user) < 2 {
				continue
			}
			defs, labels := butterflyMirrorPairDefs(user)
			if len(defs) == 0 {
				continue
			}
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":  "bt_mirror_pair",
					"user_id":      user.ID,
					"source_id":    pd.SourceID,
					"target_id":    pd.TargetID,
					"damage_index": 0,
					"pair_defs":    defs,
					"pair_labels":  labels,
				},
			})
			e.Log(fmt.Sprintf("%s 的 [镜花水月] 可触发：是否移除2张同系茧改写本次伤害来源", user.Name))
			return true
		}
	}
	return false
}

func bardInspiration(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := player.Tokens["bd_inspiration"]
	if v < 0 {
		v = 0
	}
	if v > bardInspirationCapEngine {
		v = bardInspirationCapEngine
	}
	player.Tokens["bd_inspiration"] = v
	return v
}

func addBardInspiration(player *model.Player, delta int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := bardInspiration(player) + delta
	if v < 0 {
		v = 0
	}
	if v > bardInspirationCapEngine {
		v = bardInspirationCapEngine
	}
	player.Tokens["bd_inspiration"] = v
	return v
}

func holyBowFaith(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := player.Tokens["hb_faith"]
	if v < 0 {
		v = 0
	}
	if v > holyBowFaithCapEngine {
		v = holyBowFaithCapEngine
	}
	player.Tokens["hb_faith"] = v
	return v
}

func addHolyBowFaith(player *model.Player, delta int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := holyBowFaith(player) + delta
	if v < 0 {
		v = 0
	}
	if v > holyBowFaithCapEngine {
		v = holyBowFaithCapEngine
	}
	player.Tokens["hb_faith"] = v
	return v
}

func holyBowCannon(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := player.Tokens["hb_cannon"]
	if v < 0 {
		v = 0
	}
	if v > holyBowCannonCapEngine {
		v = holyBowCannonCapEngine
	}
	player.Tokens["hb_cannon"] = v
	return v
}

func soulSorcererBlue(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := player.Tokens["ss_blue_soul"]
	if v < 0 {
		v = 0
	}
	if v > soulSorcererBlueCapEngine {
		v = soulSorcererBlueCapEngine
	}
	player.Tokens["ss_blue_soul"] = v
	return v
}

func addSoulSorcererBlue(player *model.Player, delta int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := soulSorcererBlue(player) + delta
	if v < 0 {
		v = 0
	}
	if v > soulSorcererBlueCapEngine {
		v = soulSorcererBlueCapEngine
	}
	player.Tokens["ss_blue_soul"] = v
	return v
}

func soulSorcererYellow(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := player.Tokens["ss_yellow_soul"]
	if v < 0 {
		v = 0
	}
	if v > soulSorcererYellowCapEngine {
		v = soulSorcererYellowCapEngine
	}
	player.Tokens["ss_yellow_soul"] = v
	return v
}

func addSoulSorcererYellow(player *model.Player, delta int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := soulSorcererYellow(player) + delta
	if v < 0 {
		v = 0
	}
	if v > soulSorcererYellowCapEngine {
		v = soulSorcererYellowCapEngine
	}
	player.Tokens["ss_yellow_soul"] = v
	return v
}

func moonGoddessNewMoon(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := player.Tokens["mg_new_moon"]
	if v < 0 {
		v = 0
	}
	if v > moonGoddessNewMoonCapEngine {
		v = moonGoddessNewMoonCapEngine
	}
	player.Tokens["mg_new_moon"] = v
	return v
}

func addMoonGoddessNewMoon(player *model.Player, delta int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := moonGoddessNewMoon(player) + delta
	if v < 0 {
		v = 0
	}
	if v > moonGoddessNewMoonCapEngine {
		v = moonGoddessNewMoonCapEngine
	}
	player.Tokens["mg_new_moon"] = v
	return v
}

func moonGoddessPetrify(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := player.Tokens["mg_petrify"]
	if v < 0 {
		v = 0
	}
	if v > moonGoddessPetrifyCapEngine {
		v = moonGoddessPetrifyCapEngine
	}
	player.Tokens["mg_petrify"] = v
	return v
}

func addMoonGoddessPetrify(player *model.Player, delta int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	v := moonGoddessPetrify(player) + delta
	if v < 0 {
		v = 0
	}
	if v > moonGoddessPetrifyCapEngine {
		v = moonGoddessPetrifyCapEngine
	}
	player.Tokens["mg_petrify"] = v
	return v
}

func moonGoddessDarkMoonCovers(player *model.Player) []*model.FieldCard {
	if player == nil {
		return nil
	}
	var out []*model.FieldCard
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
			continue
		}
		out = append(out, fc)
	}
	return out
}

func moonGoddessDarkMoonCount(player *model.Player) int {
	count := len(moonGoddessDarkMoonCovers(player))
	if player != nil {
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		player.Tokens["mg_dark_moon_count"] = count
		if count <= 0 {
			player.Tokens["mg_dark_form"] = 0
		}
	}
	return count
}

func addMoonGoddessDarkMoonCards(player *model.Player, cards []model.Card) int {
	if player == nil || len(cards) == 0 {
		return 0
	}
	added := 0
	for _, c := range cards {
		player.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  player.ID,
			SourceID: player.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectMoonDarkMoon,
			Trigger:  model.EffectTriggerManual,
		})
		added++
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	if added > 0 {
		player.Tokens["mg_dark_form"] = 1
	}
	moonGoddessDarkMoonCount(player)
	return added
}

func (e *GameEngine) applyMoonGoddessDarkMoonCurse(player *model.Player, removed int) {
	if player == nil || removed <= 0 {
		return
	}
	actual := e.applyCampMoraleLoss(player.Camp, removed)
	e.Log(fmt.Sprintf("%s 的 [暗月诅咒] 触发：移除%d个暗月，我方士气-%d", player.Name, removed, actual))
	moonGoddessDarkMoonCount(player)
	e.checkGameEnd()
}

func (e *GameEngine) removeMoonGoddessDarkMoonByFieldIndex(player *model.Player, fieldIdx int) (model.Card, bool) {
	if player == nil || fieldIdx < 0 || fieldIdx >= len(player.Field) {
		return model.Card{}, false
	}
	fc := player.Field[fieldIdx]
	if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
		return model.Card{}, false
	}
	card := fc.Card
	player.RemoveFieldCard(fc)
	e.applyMoonGoddessDarkMoonCurse(player, 1)
	return card, true
}

func (e *GameEngine) removeMoonGoddessDarkMoonAny(player *model.Player, n int) []model.Card {
	if player == nil || n <= 0 {
		return nil
	}
	var removed []model.Card
	for _, fc := range append([]*model.FieldCard{}, player.Field...) {
		if len(removed) >= n {
			break
		}
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
			continue
		}
		removed = append(removed, fc.Card)
		player.RemoveFieldCard(fc)
	}
	if len(removed) > 0 {
		e.applyMoonGoddessDarkMoonCurse(player, len(removed))
	}
	return removed
}

func (e *GameEngine) moonGoddessEnemyIDs(user *model.Player) []string {
	if user == nil {
		return nil
	}
	var ids []string
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp == user.Camp {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

func (e *GameEngine) moonGoddessHasElementDarkMoon(user *model.Player, ele model.Element) bool {
	if user == nil || ele == "" {
		return false
	}
	for _, fc := range moonGoddessDarkMoonCovers(user) {
		if fc.Card.Element == ele {
			return true
		}
	}
	return false
}

func bardEternalMovementCard(bard *model.Player) model.Card {
	id := "bd_eternal_movement"
	if bard != nil && bard.ID != "" {
		id = "bd_eternal_movement_" + bard.ID
	}
	return model.Card{
		ID:          id,
		Name:        "永恒乐章",
		Type:        model.CardTypeMagic,
		Element:     model.ElementDark,
		Description: "吟游诗人的永恒乐章指示牌",
	}
}

func (e *GameEngine) findBardEternalMovement(bard *model.Player) (*model.Player, *model.FieldCard) {
	if bard == nil {
		return nil, nil
	}
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBardEternalMovement {
				continue
			}
			if fc.SourceID != bard.ID {
				continue
			}
			return p, fc
		}
	}
	return nil, nil
}

func (e *GameEngine) bardEternalHolderID(bard *model.Player) string {
	holder, _ := e.findBardEternalMovement(bard)
	if holder == nil {
		return ""
	}
	return holder.ID
}

func (e *GameEngine) removeBardEternalMovement(bard *model.Player) bool {
	holder, fc := e.findBardEternalMovement(bard)
	if holder == nil || fc == nil {
		return false
	}
	holder.RemoveFieldCard(fc)
	return true
}

func (e *GameEngine) placeBardEternalMovement(bard *model.Player, target *model.Player) error {
	if bard == nil || target == nil {
		return fmt.Errorf("放置永恒乐章时角色不存在")
	}
	if target.Camp != bard.Camp {
		return fmt.Errorf("永恒乐章只能放置在我方角色面前")
	}
	e.removeBardEternalMovement(bard)
	target.AddFieldCard(&model.FieldCard{
		Card:     bardEternalMovementCard(bard),
		OwnerID:  target.ID,
		SourceID: bard.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectBardEternalMovement,
		Trigger:  model.EffectTriggerManual,
	})
	return nil
}

func soulLinkCard(sorcerer *model.Player) model.Card {
	id := "ss_soul_link"
	if sorcerer != nil && sorcerer.ID != "" {
		id = "ss_soul_link_" + sorcerer.ID
	}
	charName := "灵魂术士"
	if sorcerer != nil && sorcerer.Character != nil && sorcerer.Character.Name != "" {
		charName = sorcerer.Character.Name
	}
	return model.Card{
		ID:              id,
		Name:            "灵魂链接",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Description:     "灵魂术士的专属灵魂链接",
		ExclusiveChar1:  charName,
		ExclusiveSkill1: "灵魂链接",
	}
}

func (e *GameEngine) findSoulLink(sorcerer *model.Player) (*model.Player, *model.FieldCard) {
	if sorcerer == nil {
		return nil, nil
	}
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectSoulLink {
				continue
			}
			if fc.SourceID != sorcerer.ID {
				continue
			}
			return p, fc
		}
	}
	return nil, nil
}

func (e *GameEngine) removeSoulLink(sorcerer *model.Player, restoreCard bool) bool {
	holder, fc := e.findSoulLink(sorcerer)
	if holder == nil || fc == nil {
		if sorcerer != nil && sorcerer.Tokens != nil {
			sorcerer.Tokens["ss_link_active"] = 0
		}
		return false
	}
	holder.RemoveFieldCard(fc)
	if restoreCard && sorcerer != nil {
		sorcerer.RestoreExclusiveCard(fc.Card)
	}
	if sorcerer != nil {
		if sorcerer.Tokens == nil {
			sorcerer.Tokens = map[string]int{}
		}
		sorcerer.Tokens["ss_link_active"] = 0
	}
	return true
}

func (e *GameEngine) placeSoulLink(sorcerer *model.Player, target *model.Player, card model.Card) error {
	if sorcerer == nil || target == nil {
		return fmt.Errorf("放置灵魂链接时角色不存在")
	}
	if target.Camp != sorcerer.Camp || target.ID == sorcerer.ID {
		return fmt.Errorf("灵魂链接只能放置于队友")
	}
	e.removeSoulLink(sorcerer, false)
	target.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  target.ID,
		SourceID: sorcerer.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectSoulLink,
		Trigger:  model.EffectTriggerManual,
	})
	if sorcerer.Tokens == nil {
		sorcerer.Tokens = map[string]int{}
	}
	sorcerer.Tokens["ss_link_active"] = 1
	return nil
}

func bloodPriestessSharedLifeCard(priestess *model.Player) model.Card {
	id := "bp_shared_life"
	if priestess != nil && priestess.ID != "" {
		id = "bp_shared_life_" + priestess.ID
	}
	charName := "血之巫女"
	if priestess != nil && priestess.Character != nil && priestess.Character.Name != "" {
		charName = priestess.Character.Name
	}
	return model.Card{
		ID:              id,
		Name:            "同生共死",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Description:     "血之巫女的专属同生共死",
		ExclusiveChar1:  charName,
		ExclusiveSkill1: "同生共死",
	}
}

func (e *GameEngine) findBloodPriestessSharedLife(priestess *model.Player) (*model.Player, *model.FieldCard) {
	if priestess == nil {
		return nil, nil
	}
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
				continue
			}
			if fc.SourceID != priestess.ID {
				continue
			}
			return p, fc
		}
	}
	return nil, nil
}

func (e *GameEngine) removeBloodPriestessSharedLife(priestess *model.Player, restoreCard bool) bool {
	holder, fc := e.findBloodPriestessSharedLife(priestess)
	if holder == nil || fc == nil {
		if priestess != nil && priestess.Tokens != nil {
			priestess.Tokens["bp_shared_life_active"] = 0
		}
		return false
	}
	holder.RemoveFieldCard(fc)
	if restoreCard && priestess != nil {
		priestess.RestoreExclusiveCard(fc.Card)
	} else {
		e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
	}
	if priestess != nil {
		if priestess.Tokens == nil {
			priestess.Tokens = map[string]int{}
		}
		priestess.Tokens["bp_shared_life_active"] = 0
	}
	return true
}

func (e *GameEngine) placeBloodPriestessSharedLife(priestess *model.Player, target *model.Player, card model.Card) error {
	if priestess == nil || target == nil {
		return fmt.Errorf("放置同生共死时角色不存在")
	}
	e.removeBloodPriestessSharedLife(priestess, false)
	target.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  target.ID,
		SourceID: priestess.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectBloodSharedLife,
		Trigger:  model.EffectTriggerManual,
	})
	if priestess.Tokens == nil {
		priestess.Tokens = map[string]int{}
	}
	priestess.Tokens["bp_shared_life_active"] = 1
	return nil
}

func (e *GameEngine) hasFixedMaxHandCap(player *model.Player) bool {
	if player == nil {
		return false
	}
	if e.isMagicLancer(player) && player.Tokens != nil && player.Tokens["ml_phantom_form"] > 0 {
		return true
	}
	if e.isHero(player) && player.Tokens != nil && player.Tokens["hero_exhaustion_form"] > 0 {
		return true
	}
	for _, fc := range player.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectMercy {
			return true
		}
	}
	return false
}

func (e *GameEngine) bloodPriestessSharedLifeDeltaFor(player *model.Player) int {
	if player == nil {
		return 0
	}
	delta := 0
	for _, pid := range e.State.PlayerOrder {
		holder := e.State.Players[pid]
		if holder == nil {
			continue
		}
		for _, fc := range holder.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
				continue
			}
			source := e.State.Players[fc.SourceID]
			if source == nil || !e.isBloodPriestess(source) {
				continue
			}
			change := -2
			if source.Tokens != nil && source.Tokens["bp_bleed_form"] > 0 {
				change = 1
			}
			if source.ID == player.ID {
				delta += change
				continue
			}
			if fc.OwnerID == player.ID && !e.hasFixedMaxHandCap(player) {
				delta += change
			}
		}
	}
	return delta
}

func (e *GameEngine) maybeAutoReleaseBloodPriestessByHand(player *model.Player, reason string) bool {
	if player == nil || !e.isBloodPriestess(player) {
		return false
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	if player.Tokens["bp_bleed_form"] <= 0 || len(player.Hand) >= 3 {
		return false
	}
	player.Tokens["bp_bleed_form"] = 0
	if reason == "" {
		reason = "手牌少于3"
	}
	e.Log(fmt.Sprintf("%s 的 [流血] 强制重置：%s，脱离流血形态", player.Name, reason))
	return true
}

// maybeTriggerSoulLinkTransfer 在承受伤害前检查灵魂链接转伤流程。
// 返回 true 表示已产生中断，状态机应暂停等待玩家选择。
func (e *GameEngine) maybeTriggerSoulLinkTransfer(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 || pd.FromSoulLink || pd.SoulLinkChecked {
		return false
	}
	pd.SoulLinkChecked = true

	target := e.State.Players[pd.TargetID]
	if target == nil {
		return false
	}

	var sorcerer *model.Player
	var counterpart *model.Player
	// 场景1：灵魂术士本人受伤，另一方是其链接队友。
	if e.isSoulSorcerer(target) {
		holder, _ := e.findSoulLink(target)
		if holder != nil {
			sorcerer = target
			counterpart = holder
		}
	} else {
		// 场景2：链接队友受伤，寻找来源为该灵魂术士的链接牌。
		for _, fc := range target.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectSoulLink {
				continue
			}
			p := e.State.Players[fc.SourceID]
			if p == nil || !e.isSoulSorcerer(p) {
				continue
			}
			sorcerer = p
			counterpart = p
			break
		}
	}
	if sorcerer == nil || counterpart == nil {
		return false
	}

	blue := soulSorcererBlue(sorcerer)
	if blue <= 0 {
		return false
	}
	maxX := pd.Damage
	if blue < maxX {
		maxX = blue
	}
	if maxX <= 0 {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: sorcerer.ID,
		Context: map[string]interface{}{
			"choice_type":     "ss_link_transfer_x",
			"sorcerer_id":     sorcerer.ID,
			"damage_index":    0,
			"source_id":       pd.SourceID,
			"target_id":       pd.TargetID,
			"counterpart_id":  counterpart.ID,
			"max_x":           maxX,
			"original_damage": pd.Damage,
		},
	})
	e.Log(fmt.Sprintf("%s 的 [灵魂链接] 可触发：是否移除蓝色灵魂转移伤害（最多%d）", sorcerer.Name, maxX))
	return true
}

func (e *GameEngine) maybeTriggerMoonGoddessMedusa(attacker *model.Player, target *model.Player, sourceSkill string, attackCard *model.Card, userCtx *model.Context) bool {
	if attacker == nil || target == nil || attackCard == nil {
		return false
	}
	// 欺诈/圣屑飓暴属于“转化攻击”，不触发美杜莎之眼。
	if sourceSkill == "adventurer_fraud" || sourceSkill == "hb_holy_shard_storm" {
		return false
	}
	if attackCard.Element == "" {
		return false
	}
	// 只有攻击方的对立阵营（被攻击方阵营）中的月之女神可触发。
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp == attacker.Camp || !e.isMoonGoddess(p) {
			continue
		}
		if !e.moonGoddessHasElementDarkMoon(p, attackCard.Element) {
			continue
		}
		var selectable []int
		for i, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
				continue
			}
			if fc.Card.Element != attackCard.Element {
				continue
			}
			selectable = append(selectable, i)
		}
		if len(selectable) == 0 {
			continue
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: p.ID,
			Context: map[string]interface{}{
				"choice_type":        "mg_medusa_darkmoon_pick",
				"user_id":            p.ID,
				"attacker_id":        attacker.ID,
				"attack_element":     string(attackCard.Element),
				"darkmoon_indices":   selectable,
				"user_ctx":           userCtx,
				"source_skill":       sourceSkill,
				"medusa_target_hint": attacker.ID,
			},
		})
		e.Log(fmt.Sprintf("%s 的 [美杜莎之眼] 可触发：请选择要展示并移除的%s系暗月", p.Name, attackCard.Element))
		return true
	}
	return false
}

func (e *GameEngine) maybeTriggerMoonGoddessMoonCycleAtTurnEnd(player *model.Player) bool {
	if player == nil || !e.isMoonGoddess(player) {
		return false
	}
	canBranch1 := moonGoddessDarkMoonCount(player) > 0
	canBranch2 := player.Heal > 0
	if !canBranch1 && !canBranch2 {
		return false
	}
	var modes []string
	if canBranch1 {
		modes = append(modes, "branch1")
	}
	if canBranch2 {
		modes = append(modes, "branch2")
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "mg_moon_cycle_mode",
			"user_id":     player.ID,
			"modes":       modes,
			"target_ids":  append([]string{}, e.State.PlayerOrder...),
		},
	})
	e.Log(fmt.Sprintf("%s 的 [月之轮回] 触发：请选择发动分支", player.Name))
	return true
}

func (e *GameEngine) tryQueueMoonGoddessBlasphemy(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 || !isMagicLikeDamageType(pd.DamageType) {
		return false
	}
	source := e.State.Players[pd.SourceID]
	if source == nil || !e.isMoonGoddess(source) {
		return false
	}
	if source.Tokens == nil {
		source.Tokens = map[string]int{}
	}
	if source.Tokens["mg_blasphemy_used_turn"] > 0 {
		return false
	}
	if source.Tokens["mg_blasphemy_pending"] > 0 {
		return false
	}
	if source.Heal <= 0 {
		return false
	}
	targetIDs := e.moonGoddessEnemyIDs(source)
	if len(targetIDs) == 0 {
		return false
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: source.ID,
		Context: map[string]interface{}{
			"choice_type": "mg_blasphemy_target",
			"user_id":     source.ID,
			"target_ids":  targetIDs,
			"source_id":   pd.SourceID,
			"trigger_pd":  pd,
		},
	})
	source.Tokens["mg_blasphemy_pending"] = 1
	e.Log(fmt.Sprintf("%s 的 [月渎] 可触发：请选择目标（或跳过）", source.Name))
	return true
}

func bardMaxSameElementCount(player *model.Player) int {
	if player == nil {
		return 0
	}
	count := map[model.Element]int{}
	maxCount := 0
	for _, c := range player.Hand {
		if c.Element == "" {
			continue
		}
		count[c.Element]++
		if count[c.Element] > maxCount {
			maxCount = count[c.Element]
		}
	}
	return maxCount
}

func (e *GameEngine) campMorale(camp model.Camp) int {
	if camp == model.RedCamp {
		return e.State.RedMorale
	}
	return e.State.BlueMorale
}

func (e *GameEngine) pendingDiscardVictimID() string {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptDiscard {
		return ""
	}
	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return ""
	}
	victimID, _ := data["victim_id"].(string)
	return victimID
}

// resolveMagicLancerStardustAfterSelf 在“幻影星尘”自伤完全结算后执行后续：
// 1) 脱离幻影形态；
// 2) 若未造成己方士气下降，则弹出目标选择并造成2点法术伤害。
// 返回 true 表示产生了新的中断。
func (e *GameEngine) resolveMagicLancerStardustAfterSelf(user *model.Player) bool {
	if user == nil || !e.isMagicLancer(user) {
		return false
	}
	if user.Tokens == nil || user.Tokens["ml_stardust_pending"] <= 0 {
		return false
	}

	// 若还在等待本次自伤导致的爆牌弃牌，则延后到 ConfirmDiscard 再判定。
	if e.pendingDiscardVictimID() == user.ID {
		user.Tokens["ml_stardust_wait_discard"] = 1
		return false
	}

	before := user.Tokens["ml_stardust_morale_before"]
	current := e.campMorale(user.Camp)
	user.Tokens["ml_stardust_pending"] = 0
	user.Tokens["ml_stardust_wait_discard"] = 0
	user.Tokens["ml_stardust_morale_before"] = 0

	if user.Tokens["ml_phantom_form"] > 0 {
		user.Tokens["ml_phantom_form"] = 0
		e.Log(fmt.Sprintf("%s 的 [幻影星尘] 结算完成，脱离幻影形态并转正", user.Name))
	}

	if before > 0 && current < before {
		e.Log(fmt.Sprintf("%s 的 [幻影星尘] 未触发后续伤害：本次自伤导致己方士气下降", user.Name))
		return false
	}

	targetIDs := make([]string, 0, len(e.State.PlayerOrder))
	for _, pid := range e.State.PlayerOrder {
		if p := e.State.Players[pid]; p != nil {
			targetIDs = append(targetIDs, pid)
		}
	}
	if len(targetIDs) == 0 {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "ml_stardust_target",
			"user_id":     user.ID,
			"target_ids":  targetIDs,
		},
	})
	return true
}

// handlePostAttackHitEffects 处理“攻击命中后”的角色附加效果。
// 返回 true 表示产生了中断，状态机应暂停。
func (e *GameEngine) handlePostAttackHitEffects(pd *model.PendingDamage) bool {
	if pd == nil {
		return false
	}
	attacker := e.State.Players[pd.SourceID]
	if attacker == nil {
		return false
	}
	if attacker.Tokens == nil {
		attacker.Tokens = map[string]int{}
	}
	// 圣弓：主动攻击命中且本次攻击为圣命格时，信仰+1（上限10）。
	if e.isHolyBow(attacker) {
		if attacker.Tokens["hb_shard_miss_pending"] > 0 {
			attacker.Tokens["hb_shard_miss_pending"] = 0
		}
		if !pd.IsCounter && pd.Card != nil && strings.TrimSpace(pd.Card.Faction) == "圣" {
			before := holyBowFaith(attacker)
			after := addHolyBowFaith(attacker, 1)
			if after > before {
				e.Log(fmt.Sprintf("%s 的 [天之弓] 触发：信仰+1（当前%d）", attacker.Name, after))
			}
		}
	}

	// 祈祷师威力赐福：命中后由UI严格确认是否移除并令本次攻击伤害+2。
	if getFieldEffectCard(attacker, model.EffectPowerBlessing) != nil {
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: attacker.ID,
			Context: map[string]interface{}{
				"choice_type": "prayer_power_blessing_trigger",
				"user_id":     attacker.ID,
				"source_id":   pd.SourceID,
				"target_id":   pd.TargetID,
			},
		})
		return true
	}

	// 精灵射手：水之矢
	if attacker.Tokens["elf_elemental_shot_water_pending"] > 0 {
		attacker.Tokens["elf_elemental_shot_water_pending"] = 0
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: attacker.ID,
			Context: map[string]interface{}{
				"choice_type": "elf_elemental_shot_water_target",
				"user_id":     attacker.ID,
				"target_ids":  append([]string{}, e.State.PlayerOrder...),
			},
		})
		return true
	}

	// 精灵射手：地之矢
	if attacker.Tokens["elf_elemental_shot_earth_pending"] > 0 {
		attacker.Tokens["elf_elemental_shot_earth_pending"] = 0
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: attacker.ID,
			Context: map[string]interface{}{
				"choice_type": "elf_elemental_shot_earth_target",
				"user_id":     attacker.ID,
				"target_ids":  append([]string{}, e.State.PlayerOrder...),
			},
		})
		return true
	}

	// 魔剑士：黄泉震颤命中后，补至上限并弃2。
	if attacker.Tokens["ms_yellow_spring_pending"] > 0 {
		attacker.Tokens["ms_yellow_spring_pending"] = 0
		maxHand := e.GetMaxHand(attacker)
		if len(attacker.Hand) < maxHand {
			e.DrawCards(attacker.ID, maxHand-len(attacker.Hand))
		}
		if len(attacker.Hand) >= 2 {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptDiscard,
				PlayerID: attacker.ID,
				Context: map[string]interface{}{
					"discard_count": 2,
					"stay_in_turn":  true,
					"prompt":        "【黄泉震颤】攻击命中后，请弃置2张牌：",
				},
			})
			return true
		}
	}

	// 魔弓：魔贯冲击命中后可额外移除1个火系充能使伤害再+1（至多一次）。
	if attacker.Tokens["mb_magic_pierce_pending"] > 0 {
		if magicBowChargeCount(attacker, model.ElementFire) <= 0 {
			attacker.Tokens["mb_magic_pierce_pending"] = 0
		} else {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: attacker.ID,
				Context: map[string]interface{}{
					"choice_type": "mb_magic_pierce_hit_confirm",
					"user_id":     attacker.ID,
					"source_id":   pd.SourceID,
					"target_id":   pd.TargetID,
				},
			})
			return true
		}
	}

	return false
}

// handlePostActionEndEffects 处理行动结束后的场上效果追加结算。
// 返回 true 表示产生了中断，状态机应暂停。
func (e *GameEngine) handlePostActionEndEffects(player *model.Player, actionType model.ActionType) bool {
	if player == nil {
		return false
	}
	if actionType != model.ActionAttack && actionType != model.ActionMagic {
		return false
	}
	// 勇者：明镜止水在“本次攻击结束时”获得1点水晶（红宝石不替代，受能量上限限制）。
	if e.isHero(player) && player.Tokens != nil && player.Tokens["hero_calm_end_crystal_pending"] > 0 && actionType == model.ActionAttack {
		player.Tokens["hero_calm_end_crystal_pending"]--
		if player.Tokens["hero_calm_end_crystal_pending"] < 0 {
			player.Tokens["hero_calm_end_crystal_pending"] = 0
		}
		capV := e.getPlayerEnergyCap(player)
		if player.Gem+player.Crystal < capV {
			player.Crystal++
			e.Log(fmt.Sprintf("%s 的 [明镜止水] 结算：水晶+1", player.Name))
		} else {
			e.Log(fmt.Sprintf("%s 的 [明镜止水] 结算：能量已满，水晶未增加", player.Name))
		}
	}
	// 祈祷师迅捷赐福：攻击/法术行动结束后可移除并获得额外攻击行动。
	if getFieldEffectCard(player, model.EffectSwiftBlessing) != nil {
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: player.ID,
			Context: map[string]interface{}{
				"choice_type": "prayer_swift_blessing_trigger",
				"user_id":     player.ID,
				"action_type": string(actionType),
			},
		})
		return true
	}
	return false
}

// handlePostDamageResolved 处理“伤害结算完成后”的附加效果。
// 返回 true 表示产生了中断，状态机应暂停。
func (e *GameEngine) handlePostDamageResolved(pd *model.PendingDamage) bool {
	if pd == nil {
		return false
	}
	source := e.State.Players[pd.SourceID]
	if source == nil {
		return false
	}
	if e.isBlazeWitch(source) && source.Tokens != nil && source.Tokens["bw_pain_link_pending_discard"] > 0 {
		if source.Tokens["bw_pain_link_pending_hits"] > 0 {
			source.Tokens["bw_pain_link_pending_hits"]--
		}
		if source.Tokens["bw_pain_link_pending_hits"] <= 0 {
			source.Tokens["bw_pain_link_pending_hits"] = 0
			source.Tokens["bw_pain_link_pending_discard"] = 0
			if len(source.Hand) > 3 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptDiscard,
					PlayerID: source.ID,
					Context: map[string]interface{}{
						"discard_count": len(source.Hand) - 3,
						"stay_in_turn":  true,
						"prompt":        "【痛苦链接】请弃牌至3张手牌：",
					},
				})
				_ = e.tryQueueMoonGoddessBlasphemy(pd)
				return true
			}
		}
	}
	if e.isMagicLancer(source) &&
		source.Tokens != nil &&
		source.Tokens["ml_stardust_pending"] > 0 &&
		pd.SourceID == source.ID &&
		pd.TargetID == source.ID {
		if e.resolveMagicLancerStardustAfterSelf(source) {
			_ = e.tryQueueMoonGoddessBlasphemy(pd)
			return true
		}
	}
	if pd.Damage <= 0 {
		return false
	}
	target := e.State.Players[pd.TargetID]
	if target != nil && e.isSage(target) && isMagicLikeDamageType(pd.DamageType) {
		if pd.Damage > 3 {
			maxEnergy := e.getPlayerEnergyCap(target)
			if target.Gem+target.Crystal < maxEnergy {
				room := maxEnergy - (target.Gem + target.Crystal)
				gain := 2
				if gain > room {
					gain = room
				}
				target.Gem += gain
				if gain > 0 {
					e.Log(fmt.Sprintf("%s 的 [智慧法典] 触发：获得%d点红宝石", target.Name, gain))
				}
			} else {
				e.Log(fmt.Sprintf("%s 的 [智慧法典] 触发：能量已满，红宝石未增加", target.Name))
			}
			if len(target.Hand) > 0 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: target.ID,
					Context: map[string]interface{}{
						"choice_type": "sage_wisdom_codex_discard_confirm",
						"user_id":     target.ID,
					},
				})
				_ = e.tryQueueMoonGoddessBlasphemy(pd)
				return true
			}
		}
		if pd.Damage == 1 {
			if maxSameElementCount(target) >= 2 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: target.ID,
					Context: map[string]interface{}{
						"choice_type": "sage_magic_rebound_confirm",
						"user_id":     target.ID,
					},
				})
				_ = e.tryQueueMoonGoddessBlasphemy(pd)
				return true
			}
		}
	}
	if pd.Damage > 0 && isMagicLikeDamageType(pd.DamageType) {
		if e.tryTriggerBardDescentAfterMagicDamage(pd) {
			_ = e.tryQueueMoonGoddessBlasphemy(pd)
			return true
		}
	}
	// 动物伙伴：仅精灵射手自己的回合，且该伤害由其造成。
	if e.isElfArcher(source) && source.IsActive {
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: source.ID,
			Context: map[string]interface{}{
				"choice_type": "elf_animal_companion_confirm",
				"user_id":     source.ID,
			},
		})
		_ = e.tryQueueMoonGoddessBlasphemy(pd)
		return true
	}
	if e.tryQueueMoonGoddessBlasphemy(pd) {
		return true
	}
	return false
}
