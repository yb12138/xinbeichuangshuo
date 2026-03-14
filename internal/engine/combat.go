package engine

import (
	"errors"
	"fmt"
	"starcup-engine/internal/model"
	"strings"
)

// initCombat 初始化战斗，将 CombatRequest 推入栈并进入战斗交互阶段
func (e *GameEngine) initCombat(attackerID, targetID string, card *model.Card, isForcedHit, canBeResponded bool, isCounter ...bool) {
	attacker := e.State.Players[attackerID]
	target := e.State.Players[targetID]
	if attacker != nil && target != nil && card != nil {
		e.NotifyActionStep(fmt.Sprintf("%s出%s攻击%s", model.GetPlayerDisplayName(attacker), card.Name, model.GetPlayerDisplayName(target)))
		e.NotifyCombatCue(attackerID, targetID, "attack")
	}
	combatReq := model.CombatRequest{
		AttackerID:     attackerID,
		TargetID:       targetID,
		Card:           card,
		IsForcedHit:    isForcedHit,
		CanBeResponded: canBeResponded,
		IsCounter:      len(isCounter) > 0 && isCounter[0],
	}

	// 推入战斗栈
	e.State.CombatStack = append(e.State.CombatStack, combatReq)

	// 设置阶段为战斗交互
	e.State.Phase = model.PhaseCombatInteraction

}

// ResolveDamage 结算伤害（Step 7 & 8）
func (e *GameEngine) ResolveDamage(attackerID, victimID string, card *model.Card, damageType string) error {
	attacker := e.State.Players[attackerID]
	victim := e.State.Players[victimID]

	if attacker == nil || victim == nil {
		return errors.New("攻击者或受害者不存在")
	}

	if card == nil {
		return errors.New("卡牌不存在")
	}

	// 1. 计算基础伤害
	damage := card.Damage

	// 2. 应用攻击者的被动技能效果（仅对攻击伤害）
	if damageType == "Attack" {
		action := model.Action{
			SourceID: attackerID,
			TargetID: victimID,
			Type:     model.ActionAttack,
			Card:     card,
		}
		damage = e.applyPassiveAttackEffects(attacker, victim, damage, action)
	}

	// 3. 触发 TriggerOnDamageTaken 检查减伤技能
	damageVal := damage
	damageEventCtx := &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  attackerID,
		TargetID:  victimID,
		DamageVal: &damageVal, // 允许技能修改伤害值
		Card:      card,
	}
	damageSkillCtx := e.buildContext(victim, attacker, model.TriggerNone, damageEventCtx)
	damageSkillCtx.Flags["IsMagicDamage"] = (damageType != "Attack" && damageType != "attack")
	if strings.Contains(strings.ToLower(damageType), "no_absorb") {
		damageSkillCtx.Flags["NoElementAbsorb"] = true
	}
	e.dispatcher.OnTrigger(model.TriggerOnDamageTaken, damageSkillCtx)

	// 检查是否有中断（如减伤技能需要确认）
	if e.State.PendingInterrupt != nil {
		e.Log("等待减伤技能响应...")
		e.State.Phase = model.PhaseDamageResolution // 标记当前处于伤害结算中
		return nil                                  // 暂停执行，等待中断处理
	}

	// 4. 使用修改后的伤害值
	finalDamage := damageVal
	if finalDamage < 0 {
		finalDamage = 0
	}

	// 6. 应用伤害（扣除生命值/摸牌）
	e.applyDamage(victim, finalDamage, damageType)

	return nil
}

// resolveCombatDamage 结算战斗伤害（从 CombatStack 栈顶）
func (e *GameEngine) resolveCombatDamage(combatReq model.CombatRequest) error {

	attacker := e.State.Players[combatReq.AttackerID]
	target := e.State.Players[combatReq.TargetID]

	if attacker == nil || target == nil {
		return errors.New("攻击者或目标不存在")
	}

	// 使用新的 ResolveDamage 函数
	return e.ResolveDamage(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, "Attack")
}

// clearCombatStack 清空战斗栈
func (e *GameEngine) clearCombatStack() {
	e.State.CombatStack = []model.CombatRequest{}
}

// finishTakeHit 完成受到伤害后的流程 (扣血、事件、回合结束)
func (e *GameEngine) finishTakeHit(target *model.Player, damage int, attackAction model.Action) {
	attacker := e.State.Players[attackAction.SourceID]
	if attacker == nil || target == nil {
		return
	}

	// 4. 执行扣血
	e.applyDamage(target, damage, "Attack")

	// 5. 触发伤害承受事件
	if damage > 0 {
		damageEventCtx := &model.EventContext{
			Type:      model.EventDamage,
			SourceID:  attacker.ID,
			TargetID:  target.ID,
			DamageVal: &damage,
		}
		damageSkillCtx := e.buildContext(target, attacker, model.TriggerNone, damageEventCtx)
		e.dispatcher.OnTrigger(model.TriggerOnDamageTaken, damageSkillCtx)
		// 受伤响应可能产生中断（例如减伤/弃牌等），等待用户处理后继续
		if e.State.PendingInterrupt != nil {
			return
		}
	}

	// 重置临时技能状态
	attacker.TurnState.GaleSlashActive = false
	attacker.TurnState.PreciseShotActive = false

	eventCtx := &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   attacker.ID,
		Card:       attackAction.Card,
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: attackAction.CounterInitiator,
		},
	}
	// 6. 触发攻击行动结束事件
	phaseCtx := e.buildContext(attacker, nil, model.TriggerOnPhaseEnd, eventCtx)
	e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, phaseCtx)
	// 攻击后响应（如神圣追击）出现中断时，暂停，避免提前切回合
	if e.State.PendingInterrupt != nil {
		return
	}

	// 7. 检查圣剑第3次攻击的摸X弃X效果
	if e.triggerHolySwordDrawIfNeeded(attacker) {
		return // 等待中断处理完成后再继续
	}

	// 8. 回到额外行动阶段，交由状态机统一处理 PendingActions/回合结束
	// 这里已手动触发过一次 OnPhaseEnd，清空 LastActionType 防止重复触发
	attacker.TurnState.LastActionType = ""
	if len(e.State.PendingDamageQueue) > 0 {
		e.State.Phase = model.PhasePendingDamageResolution
		e.State.ReturnPhase = model.PhaseExtraAction
	} else {
		e.State.Phase = model.PhaseExtraAction
	}
}

// triggerHolySwordDrawIfNeeded 在满足条件时推送圣剑摸X弃X中断
func (e *GameEngine) triggerHolySwordDrawIfNeeded(attacker *model.Player) bool {
	if attacker == nil || attacker.Character == nil || attacker.TurnState.AttackCount != 3 {
		return false
	}
	hasHolySword := false
	for _, skill := range attacker.Character.Skills {
		if skill.ID == "holy_sword" {
			hasHolySword = true
			break
		}
	}
	if !hasHolySword {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptHolySwordDraw,
		PlayerID: attacker.ID,
		Context: map[string]interface{}{
			"choice_type": "holy_sword_draw",
			"player_id":   attacker.ID,
		},
	})
	e.Log(fmt.Sprintf("[Skill] %s 的 [圣剑] 第3次攻击结束，需选择摸X弃X (X=0-3)", attacker.Name))
	return true
}

// addCampResource 添加阵营资源 (水晶或宝石)，战绩区总上限为 5
func (e *GameEngine) addCampResource(camp model.Camp, resourceType string) {
	const maxTotalResources = 5

	if camp == model.RedCamp {
		if resourceType == "gem" {
			currentTotal := e.State.RedCrystals + e.State.RedGems
			if currentTotal < maxTotalResources {
				e.State.RedGems++
				fmt.Printf("[Combat] 攻击命中！红方阵营获得 1 宝石\n")
			} else {
				fmt.Printf("[Combat] 攻击命中！红方阵营资源已满，无法获得宝石\n")
			}
		} else if resourceType == "crystal" {
			currentTotal := e.State.RedCrystals + e.State.RedGems
			if currentTotal < maxTotalResources {
				e.State.RedCrystals++
				fmt.Printf("[Combat] 红方阵营获得 1 水晶\n")
			} else {
				fmt.Printf("[Combat] 红方阵营资源已满，获得的水晶被丢弃\n")
			}
		}
	} else {
		// Blue Camp
		if resourceType == "gem" {
			currentTotal := e.State.BlueCrystals + e.State.BlueGems
			if currentTotal < maxTotalResources {
				e.State.BlueGems++
				fmt.Printf("[Combat] 攻击命中！蓝方阵营获得 1 宝石\n")
			} else {
				fmt.Printf("[Combat] 攻击命中！蓝方阵营资源已满，无法获得宝石\n")
			}
		} else if resourceType == "crystal" {
			currentTotal := e.State.BlueCrystals + e.State.BlueGems
			if currentTotal < maxTotalResources {
				e.State.BlueCrystals++
				fmt.Printf("[Combat] 蓝方阵营获得 1 水晶\n")
			} else {
				fmt.Printf("[Combat] 蓝方阵营资源已满，获得的水晶被丢弃\n")
			}
		}
	}
}

// containsString 检查字符串切片是否包含指定字符串
func (e *GameEngine) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func cardMatchesExclusiveSkill(player *model.Player, card *model.Card, skillTitle string) bool {
	if player == nil || player.Character == nil || card == nil || skillTitle == "" {
		return false
	}
	return card.MatchExclusive(player.Character.Name, skillTitle)
}

// applyPassiveAttackEffects 应用攻击者的被动技能效果
func (e *GameEngine) applyPassiveAttackEffects(attacker, target *model.Player, baseDamage int, action model.Action) int {
	damage := baseDamage

	// 检查攻击者的主动技能效果
	if attacker.TurnState.PreciseShotActive {
		if cardMatchesExclusiveSkill(attacker, action.Card, "精准射击") {
			damage -= 1 // 精准射击：伤害-1
			if damage < 0 {
				damage = 0
			}
			fmt.Printf("[Skill] %s 的 [精准射击] 发动！伤害 -1\n", attacker.Name)
		} else {
			// 容错兜底：若标记异常残留但当前牌不匹配，立即失效。
			attacker.TurnState.PreciseShotActive = false
		}
	}

	// 检查攻击者的被动技能
	if attacker.Character != nil {
		for _, skill := range attacker.Character.Skills {
			if skill.Type == model.SkillTypePassive {
				// 应用狂化技能
				if skill.ID == "berserker_frenzy" {
					damage += 1 // 基础+1
					if len(attacker.Hand) > 3 {
						damage += 1 // 手牌>=3时额外+1（规则：攻击命中时手牌数大于等于3）
					}
					bonus := damage - baseDamage
					e.NotifyActionStep(fmt.Sprintf("攻击命中，%s发动被动技狂化，当前其手牌数%d，伤害额外+%d", model.GetPlayerDisplayName(attacker), len(attacker.Hand), bonus))
					e.Log(fmt.Sprintf("[Passive] %s 的狂化发动！伤害 %+d (手牌: %d)", attacker.Name, damage-baseDamage, len(attacker.Hand)))
				}
				// 圣剑：第三次攻击强制命中，无法抵挡
				if skill.ID == "holy_sword" {
					currentAttackNumber := attacker.TurnState.AttackCount // 此时AttackCount已经+1了
					if currentAttackNumber == 3 {
						e.Log(fmt.Sprintf("[Passive] %s 的 [圣剑] 发动！本回合第3次攻击强制命中，对方无法抵挡", attacker.Name))
					}
				}
				// 血影狂刀：根据对手手牌数增加伤害
				if skill.ID == "blood_blade" {
					// 规则：必须作为主动攻击打出 (非应战反弹)，且必须使用独有牌
					if action.Type == model.ActionAttack && action.CounterInitiator == "" && action.Card != nil && action.Card.MatchExclusive(attacker.Character.Name, "血影狂刀") {
						targetHandSize := len(target.Hand)
						if targetHandSize == 2 {
							damage += 2
							e.NotifyActionStep(fmt.Sprintf("攻击命中，%s发动被动技血影狂刀，对手手牌为2，伤害+2", model.GetPlayerDisplayName(attacker)))
							e.Log(fmt.Sprintf("[Passive] %s 的 [血影狂刀] 发动！对手手牌为2，伤害 +2", attacker.Name))
						} else if targetHandSize == 3 {
							damage += 1
							e.NotifyActionStep(fmt.Sprintf("攻击命中，%s发动被动技血影狂刀，对手手牌为3，伤害+1", model.GetPlayerDisplayName(attacker)))
							e.Log(fmt.Sprintf("[Passive] %s 的 [血影狂刀] 发动！对手手牌为3，伤害 +1", attacker.Name))
						}
					}
				}
				// 这里可以添加其他角色的被动技能
			}
		}
	}
	if attacker.Tokens == nil {
		attacker.Tokens = map[string]int{}
	}
	// 精灵射手：火之矢在本次攻击结算时额外+1伤害。
	if attacker.Tokens["elf_elemental_shot_fire_pending"] > 0 {
		damage += 1
		attacker.Tokens["elf_elemental_shot_fire_pending"] = 0
		e.Log(fmt.Sprintf("[Passive] %s 的 [火之矢] 生效，伤害 +1", attacker.Name))
	}
	// 魔剑士：暗影形态下攻击伤害+1（含应战攻击）。
	if isCharacter(attacker, "magic_swordsman") && attacker.Tokens["ms_shadow_form"] > 0 {
		damage += 1
		e.Log(fmt.Sprintf("[Passive] %s 的 [暗影之力] 生效，伤害 +1", attacker.Name))
	}
	// 魔枪：暗之解放/充盈的“本回合下一次主动攻击伤害加成”。
	if isCharacter(attacker, "magic_lancer") &&
		action.Type == model.ActionAttack &&
		action.CounterInitiator == "" {
		if attacker.TurnState.UsedSkillCounts == nil {
			attacker.TurnState.UsedSkillCounts = map[string]int{}
		}
		if attacker.TurnState.UsedSkillCounts["ml_dark_release_next_attack_bonus"] > 0 {
			damage += 1
			attacker.TurnState.UsedSkillCounts["ml_dark_release_next_attack_bonus"] = 0
			e.Log(fmt.Sprintf("[Passive] %s 的 [暗之解放] 生效，本次主动攻击伤害 +1", attacker.Name))
		}
		if bonus := attacker.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"]; bonus > 0 {
			damage += bonus
			attacker.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"] = 0
			e.Log(fmt.Sprintf("[Passive] %s 的 [充盈] 生效，本次主动攻击伤害 +%d", attacker.Name, bonus))
		}
	}
	// 格斗家：蓄力一击与百式幻龙拳伤害修正。
	if isCharacter(attacker, "fighter") {
		if action.Type == model.ActionAttack &&
			action.CounterInitiator == "" &&
			attacker.Tokens["fighter_charge_damage_pending"] > 0 {
			damage += 1
			attacker.Tokens["fighter_charge_damage_pending"] = 0
			attacker.Tokens["fighter_charge_pending"] = 0
			e.Log(fmt.Sprintf("[Passive] %s 的 [蓄力一击] 生效，本次主动攻击伤害 +1", attacker.Name))
		}
		if attacker.Tokens["fighter_hundred_dragon_form"] > 0 {
			if action.Type == model.ActionAttack && action.CounterInitiator == "" {
				damage += 2
				e.Log(fmt.Sprintf("[Passive] %s 的 [百式幻龙拳] 生效，本次主动攻击伤害 +2", attacker.Name))
			} else if action.Type == model.ActionAttack && action.CounterInitiator != "" {
				damage += 1
				e.Log(fmt.Sprintf("[Passive] %s 的 [百式幻龙拳] 生效，本次应战攻击伤害 +1", attacker.Name))
			}
		}
	}
	// 勇者：怒吼生效时，本次主动攻击伤害额外+2（命中分支）。
	if isCharacter(attacker, "hero") &&
		action.Type == model.ActionAttack &&
		action.CounterInitiator == "" &&
		attacker.Tokens["hero_roar_damage_pending"] > 0 {
		damage += 2
		attacker.Tokens["hero_roar_damage_pending"] = 0
		attacker.Tokens["hero_roar_active"] = 0
		e.Log(fmt.Sprintf("[Passive] %s 的 [怒吼] 生效，本次主动攻击伤害 +2", attacker.Name))
	}
	// 暗杀者：潜行状态下，主动攻击伤害额外+X（X=当前剩余能量=宝石+水晶）。
	if isCharacter(attacker, "assassin") &&
		action.Type == model.ActionAttack &&
		action.CounterInitiator == "" {
		extra := 0
		if attacker.Tokens != nil {
			extra = attacker.Tokens["assassin_stealth_attack_bonus"]
			// 单次攻击增伤，命中流程只应生效一次。
			attacker.Tokens["assassin_stealth_attack_bonus"] = 0
		}
		// 兜底：若旧流程未写入 token，且当前仍在潜行，则按实时能量计算。
		if extra <= 0 && attacker.HasFieldEffect(model.EffectStealth) {
			extra = attacker.Gem + attacker.Crystal
		}
		if extra > 0 {
			damage += extra
			e.Log(fmt.Sprintf("[Passive] %s 处于[潜行]，本次主动攻击伤害额外+%d（剩余能量）", attacker.Name, extra))
		}
	}
	// 圣弓：主动攻击若非圣命格，本次攻击伤害-1。
	if isCharacter(attacker, "holy_bow") &&
		action.Type == model.ActionAttack &&
		action.CounterInitiator == "" &&
		action.Card != nil {
		if strings.TrimSpace(action.Card.Faction) != "圣" {
			damage--
			if damage < 0 {
				damage = 0
			}
			e.Log(fmt.Sprintf("[Passive] %s 的 [天之弓] 生效：非圣命格主动攻击伤害 -1", attacker.Name))
		}
	}

	return damage
}

// applyDamageWithOptions 应用伤害逻辑 (治疗抵消 + 摸牌)
func (e *GameEngine) applyDamageWithOptions(target *model.Player, damage int, damageType string, capToHandLimit bool) {
	// 说明：
	// OnDamaged 的【圣盾】改为在“承受/防御选择”分支中结算（例如 take 时优先抵挡），
	// 这里不再做通用自动触发，避免出现“伤害已生效却又把圣盾误消耗”的重复结算。

	// 5. 造成伤害 (摸牌)
	if damage > 0 {
		e.Log(fmt.Sprintf("[Damage] %s 受到 %d 点伤害 (摸牌)\n", target.Name, damage))

		// 触发摸牌前事件 (允许水影等技能干预)
		drawEventCtx := &model.EventContext{
			Type:      model.EventBeforeDraw,
			SourceID:  target.ID,
			TargetID:  target.ID,
			DrawCount: &damage,
		}
		drawCtx := e.buildContext(target, target, model.TriggerBeforeDraw, drawEventCtx)
		drawCtx.Flags["IsMagicDamage"] = (damageType != "Attack" && damageType != "attack")
		drawCtx.Flags["FromDamageDraw"] = true
		if capToHandLimit {
			drawCtx.Flags["capToHandLimit"] = true
		}
		if strings.EqualFold(damageType, "magic_no_morale") {
			drawCtx.Flags["NoMoraleLoss"] = true
		}

		// 如果在 BuffResolve 或 PendingDamageResolution 阶段（如中毒伤害），弃牌后应继续当前回合
		if e.State.Phase == model.PhaseBuffResolve || e.State.Phase == model.PhasePendingDamageResolution {
			drawCtx.Flags["StayInTurn"] = true
		}

		e.dispatcher.OnTrigger(model.TriggerBeforeDraw, drawCtx)

		// 检查是否有中断 (如水影技能触发)
		if e.State.PendingInterrupt != nil {
			e.Log("[System] 等待响应前暂停扣卡...")
			return // 暂停执行，等待中断处理完成后恢复
		}

		// 没有中断，继续执行扣卡
		e.resumePendingDraw(drawCtx)
	} else {
		e.Log("[Damage] 伤害被完全抵消")
	}
}

// applyDamage 应用伤害逻辑 (治疗抵消 + 摸牌)
func (e *GameEngine) applyDamage(target *model.Player, damage int, damageType string) {
	e.applyDamageWithOptions(target, damage, damageType, false)
}
