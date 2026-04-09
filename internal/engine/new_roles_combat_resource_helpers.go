package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func bardInspiration(player *model.Player) int {
	return tokenValueBounded(player, "bd_inspiration", bardInspirationCapEngine)
}

func addBardInspiration(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "bd_inspiration", delta, bardInspirationCapEngine)
}

func holyBowFaith(player *model.Player) int {
	return tokenValueBounded(player, "hb_faith", holyBowFaithCapEngine)
}

func addHolyBowFaith(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "hb_faith", delta, holyBowFaithCapEngine)
}

func holyBowCannon(player *model.Player) int {
	return tokenValueBounded(player, "hb_cannon", holyBowCannonCapEngine)
}

func swordEmperorSwordQi(player *model.Player) int {
	return tokenValueBounded(player, "se_sword_qi", swordEmperorSwordQiCapEngine)
}

func addSwordEmperorSwordQi(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "se_sword_qi", delta, swordEmperorSwordQiCapEngine)
}

func swordEmperorEnergyCount(player *model.Player) int {
	if player == nil {
		return 0
	}
	return player.Gem + player.Crystal
}

func swordEmperorSwordSoulCards(player *model.Player) []*model.FieldCard {
	return coverCardsByEffect(player, model.EffectSwordSoul)
}

func swordEmperorSwordSoulCount(player *model.Player) int {
	return len(swordEmperorSwordSoulCards(player))
}

func (e *GameEngine) syncSwordEmperorSwordSoulToken(player *model.Player) {
	if player == nil {
		return
	}
	ensurePlayerTokensMap(player)
	player.Tokens["se_sword_soul_count"] = swordEmperorSwordSoulCount(player)
}

func (e *GameEngine) placeSwordEmperorSwordSoul(player *model.Player, card model.Card) bool {
	if player == nil || swordEmperorSwordSoulCount(player) >= swordEmperorSwordSoulCapEngine {
		return false
	}
	player.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  player.ID,
		SourceID: player.ID,
		Mode:     model.FieldCover,
		Effect:   model.EffectSwordSoul,
	})
	e.syncSwordEmperorSwordSoulToken(player)
	return true
}

func (e *GameEngine) removeSwordEmperorSwordSoul(player *model.Player) (model.Card, bool) {
	if player == nil {
		return model.Card{}, false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSwordSoul {
			continue
		}
		card := fc.Card
		player.RemoveFieldCard(fc)
		e.syncSwordEmperorSwordSoulToken(player)
		return card, true
	}
	return model.Card{}, false
}

func (e *GameEngine) takeDiscardPileCardByID(cardID string) (model.Card, bool) {
	if e == nil || cardID == "" {
		return model.Card{}, false
	}
	for i := len(e.State.DiscardPile) - 1; i >= 0; i-- {
		if e.State.DiscardPile[i].ID != cardID {
			continue
		}
		card := e.State.DiscardPile[i]
		e.State.DiscardPile = append(e.State.DiscardPile[:i], e.State.DiscardPile[i+1:]...)
		return card, true
	}
	return model.Card{}, false
}

func clearSwordEmperorCombatTokens(player *model.Player) {
	if player == nil {
		return
	}
	player.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] = 0
	player.TurnState.UsedSkillCounts["se_angel_soul_armed"] = 0
	player.TurnState.UsedSkillCounts["se_demon_soul_armed"] = 0
}

func (e *GameEngine) resolveSwordEmperorAttackMiss(attackerID string, attackCard *model.Card, isCounter bool) {
	attacker := e.State.Players[attackerID]
	if attacker == nil || !e.isSwordEmperor(attacker) || isCounter {
		return
	}
	if attacker.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] <= 0 &&
		swordEmperorSwordSoulCount(attacker) < swordEmperorSwordSoulCapEngine &&
		attackCard != nil {
		if card, ok := e.takeDiscardPileCardByID(attackCard.ID); ok && e.placeSwordEmperorSwordSoul(attacker, card) {
			e.Log(fmt.Sprintf("%s 的 [剑魂守护] 生效：将本次攻击牌转化为1张剑魂（当前%d）", attacker.Name, swordEmperorSwordSoulCount(attacker)))
		}
	}

	qi := addSwordEmperorSwordQi(attacker, 1)
	e.Log(fmt.Sprintf("%s 的 [佯攻] 生效：剑气+1（当前%d）", attacker.Name, qi))

	if attacker.TurnState.UsedSkillCounts["se_angel_soul_armed"] > 0 {
		if gained := e.addCampMorale(attacker.Camp, 1); gained > 0 {
			e.Log(fmt.Sprintf("%s 的 [天使之魂] 未命中分支生效：%s方士气+%d", attacker.Name, attacker.Camp, gained))
		} else {
			e.Log(fmt.Sprintf("%s 的 [天使之魂] 未命中分支生效：%s方士气已满，未再增加", attacker.Name, attacker.Camp))
		}
	}
	if attacker.TurnState.UsedSkillCounts["se_demon_soul_armed"] > 0 {
		qi = addSwordEmperorSwordQi(attacker, 2)
		e.Log(fmt.Sprintf("%s 的 [恶魔之魂] 未命中分支生效：剑气+2（当前%d）", attacker.Name, qi))
	}

	clearSwordEmperorCombatTokens(attacker)
}

func (e *GameEngine) resolveSwordEmperorAttackHitAftermath(pd *model.PendingDamage) {
	if pd == nil || pd.IsCounter || pd.HasCheck(model.PendingDamageCheckAttackMissResolved) {
		return
	}
	attacker := e.State.Players[pd.SourceID]
	if attacker == nil || !e.isSwordEmperor(attacker) {
		return
	}
	if attacker.TurnState.UsedSkillCounts["se_angel_soul_armed"] > 0 {
		e.Heal(attacker.ID, 2)
		e.Log(fmt.Sprintf("%s 的 [天使之魂] 命中分支生效：治疗+2", attacker.Name))
	}
	clearSwordEmperorCombatTokens(attacker)
}

func (e *GameEngine) beastSamuraiZanshin(player *model.Player) int {
	return tokenValueBounded(player, "bs_zanshin", beastSamuraiZanshinCapEngine)
}

func (e *GameEngine) addBeastSamuraiZanshin(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "bs_zanshin", delta, beastSamuraiZanshinCapEngine)
}

func (e *GameEngine) beastSamuraiBeastSoul(player *model.Player) int {
	return tokenValueBounded(player, "bs_beast_soul", beastSamuraiBeastSoulCapEngine)
}

func (e *GameEngine) addBeastSamuraiBeastSoul(player *model.Player, delta int, ignoreCap bool) int {
	return addTokenValueBoundedWithIgnoreCap(player, "bs_beast_soul", delta, beastSamuraiBeastSoulCapEngine, ignoreCap)
}

func (e *GameEngine) consumeBeastSamuraiBeastSoul(player *model.Player, amount int) int {
	if player == nil || amount <= 0 {
		return 0
	}
	current := e.beastSamuraiBeastSoul(player)
	if amount > current {
		amount = current
	}
	if amount <= 0 {
		return 0
	}
	e.addBeastSamuraiBeastSoul(player, -amount, true)
	e.addBeastSamuraiZanshin(player, amount)
	return amount
}

func (e *GameEngine) beastSamuraiInIaijutsuForm(player *model.Player) bool {
	return effectivePlayerForm(player) == model.FormBeastSamuraiIaijutsu
}

func (e *GameEngine) enterBeastSamuraiIaijutsuForm(player *model.Player) bool {
	if player == nil {
		return false
	}
	changed := effectivePlayerOrientation(player) != model.OrientationTapped || effectivePlayerForm(player) != model.FormBeastSamuraiIaijutsu
	player.Orientation = model.OrientationTapped
	player.Form = model.FormBeastSamuraiIaijutsu
	return changed
}

func (e *GameEngine) leaveBeastSamuraiIaijutsuForm(player *model.Player) bool {
	if player == nil {
		return false
	}
	changed := effectivePlayerOrientation(player) != model.OrientationNormal || effectivePlayerForm(player) != ""
	player.Orientation = model.OrientationNormal
	player.Form = ""
	return changed
}

func clearBeastSamuraiAttackTokens(player *model.Player) {
	if player == nil {
		return
	}
	ensurePlayerTokensMap(player)
	player.Tokens["bs_reversal_pending_x"] = 0
}

func (e *GameEngine) holyBowShardMissEligibleAllies(user *model.Player, x int) []string {
	if e == nil || user == nil || x <= 0 {
		return nil
	}
	allyIDs := make([]string, 0)
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp != user.Camp || p.ID == user.ID {
			continue
		}
		if len(p.Hand) < x {
			continue
		}
		allyIDs = append(allyIDs, p.ID)
	}
	return allyIDs
}

func (e *GameEngine) holyBowShardMissValidXValues(user *model.Player, maxX int) []int {
	if e == nil || user == nil || maxX <= 0 {
		return nil
	}
	valid := make([]int, 0, maxX)
	for x := 1; x <= maxX; x++ {
		if len(e.holyBowShardMissEligibleAllies(user, x)) > 0 {
			valid = append(valid, x)
		}
	}
	return valid
}

func soulSorcererBlue(player *model.Player) int {
	return tokenValueBounded(player, "ss_blue_soul", soulSorcererBlueCapEngine)
}

func addSoulSorcererBlue(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "ss_blue_soul", delta, soulSorcererBlueCapEngine)
}

func soulSorcererYellow(player *model.Player) int {
	return tokenValueBounded(player, "ss_yellow_soul", soulSorcererYellowCapEngine)
}

func addSoulSorcererYellow(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "ss_yellow_soul", delta, soulSorcererYellowCapEngine)
}

func (e *GameEngine) applySoulSorcererSoulDevour(victim *model.Player, finalLoss int, fromDamageDraw bool) {
	if e == nil || victim == nil || finalLoss <= 0 || !fromDamageDraw {
		return
	}
	for _, player := range e.GetAllPlayers() {
		if player == nil || player.Camp != victim.Camp || !e.isSoulSorcerer(player) {
			continue
		}
		before := soulSorcererYellow(player)
		after := addSoulSorcererYellow(player, finalLoss)
		e.Log(fmt.Sprintf("%s 的 [灵魂吞噬] 触发：黄色灵魂 +%d（%d→%d）", player.Name, finalLoss, before, after))
	}
}

func moonGoddessNewMoon(player *model.Player) int {
	return tokenValueBounded(player, "mg_new_moon", moonGoddessNewMoonCapEngine)
}

func addMoonGoddessNewMoon(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "mg_new_moon", delta, moonGoddessNewMoonCapEngine)
}

func moonGoddessPetrify(player *model.Player) int {
	return tokenValueBounded(player, "mg_petrify", moonGoddessPetrifyCapEngine)
}

func addMoonGoddessPetrify(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "mg_petrify", delta, moonGoddessPetrifyCapEngine)
}

func moonGoddessDarkMoonCovers(player *model.Player) []*model.FieldCard {
	return coverCardsByEffect(player, model.EffectMoonDarkMoon)
}

func moonGoddessDarkMoonCount(player *model.Player) int {
	count := len(moonGoddessDarkMoonCovers(player))
	if player != nil && count <= 0 {
		leaveMoonGoddessDarkMoonForm(player)
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
	if added > 0 {
		enterMoonGoddessDarkMoonForm(player)
	}
	moonGoddessDarkMoonCount(player)
	return added
}

func (e *GameEngine) applyMoonGoddessDarkMoonCurse(player *model.Player, removed int) {
	if player == nil || removed <= 0 {
		return
	}
	beforePoses := e.snapshotPlayerPoses()
	actual := e.applyCampMoraleLoss(player.Camp, removed)
	e.Log(fmt.Sprintf("%s 的 [暗月诅咒] 触发：移除%d个暗月，我方士气-%d", player.Name, removed, actual))
	moonGoddessDarkMoonCount(player)
	e.dispatchOrientationChanges(beforePoses)
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
	if e == nil || user == nil {
		return nil
	}
	return e.campEnemyIDs(user.Camp)
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
