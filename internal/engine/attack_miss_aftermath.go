package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

func (e *GameEngine) resolveHeroRoarMissWithOverride(attackerID string, force bool) {
	attacker := e.State.Players[attackerID]
	if attacker == nil || !e.isHero(attacker) {
		return
	}
	if !force && attacker.TurnState.UsedSkillCounts["hero_roar_active"] <= 0 {
		return
	}
	attacker.TurnState.UsedSkillCounts["hero_roar_active"] = 0
	wisdom := attacker.Tokens["hero_wisdom"] + 1
	if wisdom > 3 {
		wisdom = 3
	}
	attacker.Tokens["hero_wisdom"] = wisdom
	e.Log(fmt.Sprintf("%s 的 [怒吼] 未命中分支生效：智慧+1（当前%d）", attacker.Name, wisdom))
}

func (e *GameEngine) resolveFighterChargeMissWithOverride(attackerID string, force bool) {
	attacker := e.State.Players[attackerID]
	if attacker == nil || !e.isFighter(attacker) {
		return
	}
	if !force && attacker.TurnState.SkillFlowState["fighter_charge_pending"] <= 0 {
		return
	}
	attacker.TurnState.SkillFlowState["fighter_charge_pending"] = 0
	damage := attacker.Tokens["fighter_qi"]
	if damage < 1 {
		damage = 1
	}
	if damage > 0 {
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   attacker.ID,
			TargetID:   attacker.ID,
			Damage:     damage,
			DamageType: model.MagicDamage,
		})
	}
	e.Log(fmt.Sprintf("%s 的 [蓄力一击] 未命中分支生效：对自己造成%d点法术伤害", attacker.Name, damage))
}

func (e *GameEngine) resolveMagicBowPierceMiss(attackerID, targetID string, attackCard *model.Card, isCounter bool) {
	e.resolveMagicBowPierceMissWithOverride(attackerID, targetID, attackCard, false, false, isCounter)
}

func (e *GameEngine) resolveMagicBowPierceMissWithOverride(attackerID, targetID string, attackCard *model.Card, forceHeroRoarMiss, forceFighterChargeMiss, isCounter bool) {
	e.resolveHeroRoarMissWithOverride(attackerID, forceHeroRoarMiss)
	e.resolveFighterChargeMissWithOverride(attackerID, forceFighterChargeMiss)
	e.resolveSwordEmperorAttackMiss(attackerID, attackCard, isCounter)
	e.resolveHolyBowShardMiss(attackerID, targetID)

	attacker := e.State.Players[attackerID]
	target := e.State.Players[targetID]
	clearBeastSamuraiAttackTokens(attacker)
	if attacker == nil || target == nil {
		return
	}
	if attacker.TurnState.SkillFlowState["mb_magic_pierce_pending"] <= 0 {
		return
	}
	attacker.TurnState.SkillFlowState["mb_magic_pierce_pending"] = 0
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   attackerID,
		TargetID:   targetID,
		Damage:     3,
		DamageType: model.MagicAttack,
	})
	e.Log(fmt.Sprintf("%s 的 [魔贯冲击] 未命中：对 %s 造成3点法术伤害", attacker.Name, target.Name))
}

func (e *GameEngine) resolveHolyBowShardMiss(attackerID, targetID string) {
	attacker := e.State.Players[attackerID]
	target := e.State.Players[targetID]
	if attacker == nil || target == nil || !e.isHolyBow(attacker) {
		return
	}
	if attacker.TurnState.SkillFlowState["hb_shard_miss_pending"] <= 0 {
		return
	}
	attacker.TurnState.SkillFlowState["hb_shard_miss_pending"] = 0
	maxX := attacker.Heal
	if maxX > 2 {
		maxX = 2
	}
	if maxX <= 0 {
		e.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中，但治疗不足，未触发后续效果", attacker.Name))
		return
	}
	validX := e.holyBowShardMissValidXValues(attacker, maxX)
	if len(validX) == 0 {
		e.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中，但没有能弃满牌数的队友可供选择", attacker.Name))
		return
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: attacker.ID,
		Context: map[string]interface{}{
			"choice_type": "hb_holy_shard_miss_confirm",
			"user_id":     attacker.ID,
			"target_id":   targetID,
			"max_x":       maxX,
			"valid_x":     validX,
		},
	})
	e.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中：可移除治疗并令队友弃牌", attacker.Name))
}

func (e *GameEngine) resolveShieldBlockedAttackAsMiss(pd *model.PendingDamage) {
	if pd == nil || pd.HasCheck(model.PendingDamageCheckAttackMissResolved) || !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
		return
	}
	attacker := e.State.Players[pd.SourceID]
	target := e.State.Players[pd.TargetID]
	if attacker == nil {
		return
	}

	targetName := pd.TargetID
	if target != nil {
		targetName = target.Name
	}
	e.NotifyCombatCue(pd.SourceID, pd.TargetID, "shield")
	e.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】抵消了本次攻击，判定为未命中", targetName))
	e.Log(fmt.Sprintf("[Combat] %s 的攻击被【圣盾】完全抵消，按未命中处理", attacker.Name))

	e.resolveMagicBowPierceMissWithOverride(
		pd.SourceID,
		pd.TargetID,
		pd.Card,
		pd.HasCheck(model.PendingDamageCheckHeroRoarMissArmed),
		pd.HasCheck(model.PendingDamageCheckFighterChargeMissArmed),
		pd.IsCounter,
	)
	pd.SetCheck(model.PendingDamageCheckAttackMissResolved, true)
}
