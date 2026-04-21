// gameflow: 剑帝运行时：攻击未命中/命中后结算逻辑。

package sword_emperor

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// PlaceSwordEmperorSwordSoul 尝试将 card 作为剑魂盖牌放置到玩家场上。
func PlaceSwordEmperorSwordSoul(player *model.Player, card model.Card) bool {
	if player == nil || SwordSoulCount(player) >= SwordSoulCap {
		return false
	}
	player.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  player.ID,
		SourceID: player.ID,
		Mode:     model.FieldCover,
		Effect:   model.EffectSwordSoul,
	})
	SyncSwordSoulToken(player)
	return true
}

// ResolveAttackMiss 处理剑帝攻击未命中后的完整结算。
func ResolveAttackMiss(rt engineplayer.ChoiceRuntime, attackerID string, attackCard *model.Card, isCounter bool) {
	attacker := rt.LookupPlayer(attackerID)
	if attacker == nil || !engineplayer.IsCharacter(attacker, "sword_emperor") || isCounter {
		return
	}
	if attacker.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] <= 0 &&
		SwordSoulCount(attacker) < SwordSoulCap &&
		attackCard != nil {
		if card, ok := rt.TakeDiscardPileCardByID(attackCard.ID); ok && PlaceSwordEmperorSwordSoul(attacker, card) {
			rt.Log(fmt.Sprintf("%s 的 [剑魂守护] 生效：将本次攻击牌转化为1张剑魂（当前%d）", attacker.Name, SwordSoulCount(attacker)))
		}
	}

	qi := AddSwordQi(attacker, 1)
	rt.Log(fmt.Sprintf("%s 的 [佯攻] 生效：剑气+1（当前%d）", attacker.Name, qi))

	if attacker.TurnState.UsedSkillCounts["se_angel_soul_armed"] > 0 {
		if gained := rt.AddCampMorale(attacker.Camp, 1); gained > 0 {
			rt.Log(fmt.Sprintf("%s 的 [天使之魂] 未命中分支生效：%s方士气+%d", attacker.Name, attacker.Camp, gained))
		} else {
			rt.Log(fmt.Sprintf("%s 的 [天使之魂] 未命中分支生效：%s方士气已满，未再增加", attacker.Name, attacker.Camp))
		}
	}
	if attacker.TurnState.UsedSkillCounts["se_demon_soul_armed"] > 0 {
		qi = AddSwordQi(attacker, 2)
		rt.Log(fmt.Sprintf("%s 的 [恶魔之魂] 未命中分支生效：剑气+2（当前%d）", attacker.Name, qi))
	}

	ClearCombatTokens(attacker)
}

// ResolveAttackHitAftermath 处理剑帝攻击命中后的后续效果。
func ResolveAttackHitAftermath(rt engineplayer.ChoiceRuntime, pd *model.PendingDamage) {
	if pd == nil || pd.IsCounter || pd.HasCheck(model.PendingDamageCheckAttackMissResolved) {
		return
	}
	attacker := rt.LookupPlayer(pd.SourceID)
	if attacker == nil || !engineplayer.IsCharacter(attacker, "sword_emperor") {
		return
	}
	if attacker.TurnState.UsedSkillCounts["se_angel_soul_armed"] > 0 {
		rt.Heal(attacker.ID, 2)
		rt.Log(fmt.Sprintf("%s 的 [天使之魂] 命中分支生效：治疗+2", attacker.Name))
	}
	ClearCombatTokens(attacker)
}
