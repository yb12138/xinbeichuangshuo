// gameflow: 剑皇：剑气/剑魂资源与战斗结算辅助。
package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func swordEmperorSwordQi(player *model.Player) int {
	return tokenValueBounded(player, "se_sword_qi", swordEmperorSwordQiCapEngine)
}

func addSwordEmperorSwordQi(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "se_sword_qi", delta, swordEmperorSwordQiCapEngine)
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
