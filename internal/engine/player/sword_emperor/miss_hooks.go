// gameflow: 剑帝攻击未命中 Timing Hook 实现。

package sword_emperor

import (
	"fmt"

	"starcup-engine/internal/engine/player"
)

func attackMissHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	attacker := rt.GetPlayer(ctx.SourceID)
	if attacker == nil || !rt.IsCharacter(attacker, "sword_emperor") {
		return player.TimingHookResult{}
	}
	// 反击不触发剑帝未命中效果
	if ctx.IsCounter {
		return player.TimingHookResult{}
	}

	// 剑魂守护：若未被禁用且剑魂未满，尝试将攻击牌转化为剑魂
	if attacker.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] <= 0 &&
		SwordSoulCount(attacker) < SwordSoulCap &&
		ctx.Card != nil {
		if card, ok := rt.TakeDiscardPileCardByID(ctx.Card.ID); ok && PlaceSwordEmperorSwordSoul(attacker, card) {
			rt.Log(fmt.Sprintf("%s 的 [剑魂守护] 生效：将本次攻击牌转化为1张剑魂（当前%d）", attacker.Name, SwordSoulCount(attacker)))
		}
	}

	// 佯攻：剑气+1
	qi := AddSwordQi(attacker, 1)
	rt.Log(fmt.Sprintf("%s 的 [佯攻] 生效：剑气+1（当前%d）", attacker.Name, qi))

	return player.TimingHookResult{}
}

// angelSoulMissHook is the armed 天使之魂 branch for active-attack miss settlement.
func angelSoulMissHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	attacker := rt.GetPlayer(ctx.SourceID)
	if attacker == nil || !rt.IsCharacter(attacker, "sword_emperor") || ctx.IsCounter {
		return player.TimingHookResult{}
	}
	if attacker.TurnState.UsedSkillCounts["se_angel_soul_armed"] > 0 {
		if gained := rt.AddCampMorale(attacker.Camp, 1); gained > 0 {
			rt.Log(fmt.Sprintf("%s 的 [天使之魂] 未命中分支生效：%s方士气+%d", attacker.Name, attacker.Camp, gained))
		} else {
			rt.Log(fmt.Sprintf("%s 的 [天使之魂] 未命中分支生效：%s方士气已满，未再增加", attacker.Name, attacker.Camp))
		}
	}
	return player.TimingHookResult{}
}

// demonSoulMissHook is the armed 恶魔之魂 branch for active-attack miss settlement.
func demonSoulMissHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	attacker := rt.GetPlayer(ctx.SourceID)
	if attacker == nil || !rt.IsCharacter(attacker, "sword_emperor") || ctx.IsCounter {
		return player.TimingHookResult{}
	}
	if attacker.TurnState.UsedSkillCounts["se_demon_soul_armed"] > 0 {
		qi := AddSwordQi(attacker, 2)
		rt.Log(fmt.Sprintf("%s 的 [恶魔之魂] 未命中分支生效：剑气+2（当前%d）", attacker.Name, qi))
	}
	return player.TimingHookResult{}
}

func attackMissCleanupHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	attacker := rt.GetPlayer(ctx.SourceID)
	if attacker == nil || !rt.IsCharacter(attacker, "sword_emperor") || ctx.IsCounter {
		return player.TimingHookResult{}
	}
	ClearCombatTokens(attacker)
	return player.TimingHookResult{}
}
