// gameflow: 战斗栈：发起攻击/法术战斗、阶段切换、与应战/承伤衔接。

package engine

import (
	"errors"
	"fmt"
	"starcup-engine/internal/model"
	"strings"
)

// initCombat 初始化战斗，将 CombatRequest 推入栈并进入战斗交互阶段
func (e *GameEngine) initCombat(attackerID, targetID string, card *model.Card, isForcedHit, canBeResponded, ignoreShield bool, interceptTags map[model.CombatInterceptTag]bool, isCounter ...bool) {
	e.setCombatStage(model.CombatStageDeclare)
	e.clearSubflow()
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
		IgnoreShield:   ignoreShield,
		CanBeResponded: canBeResponded,
		IsCounter:      len(isCounter) > 0 && isCounter[0],
		InterceptTags:  model.CloneCombatInterceptTags(interceptTags),
	}
	if combatReq.IsForcedHit {
		combatReq.SetInterceptTag(model.CombatInterceptForceHit)
	}
	if combatReq.IgnoreShield {
		combatReq.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
	}
	if !combatReq.CanBeResponded {
		combatReq.SetInterceptTag(model.CombatInterceptUnrespondable)
	}

	// 推入战斗栈
	e.State.CombatStack = append(e.State.CombatStack, combatReq)
}

// ResolveDamage 结算伤害（Step 7 & 8）
func (e *GameEngine) ResolveDamage(attackerID, victimID string, card *model.Card, damageType model.DamageType) error {
	e.setCombatStage(model.CombatStageCalcDamage)
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
	if damageType == model.AttackDamage {
		action := model.Action{
			SourceID: attackerID,
			TargetID: victimID,
			Type:     model.ActionAttack,
			Card:     card,
		}
		damage = e.applyAttackDamageModifiers(attacker, victim, damage, action)
	}

	// 3. 触发 TimingOnDamageTaken 检查减伤技能
	damageVal := damage
	damageEventCtx := &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  attackerID,
		TargetID:  victimID,
		DamageVal: &damageVal, // 允许技能修改伤害值
		Card:      card,
	}
	damageSkillCtx := e.buildContext(victim, attacker, model.TimingOnDamageTaken, damageEventCtx)
	damageSkillCtx.Flags["IsMagicDamage"] = !strings.EqualFold(string(damageType), string(model.AttackDamage))
	if damageSkillCtx.Selections == nil {
		damageSkillCtx.Selections = map[string]any{}
	}
	damageSkillCtx.Selections["damage_type"] = damageType
	e.dispatcher.OnTiming(model.TimingOnDamageTaken, damageSkillCtx)

	// 检查是否有中断（如减伤技能需要确认）
	if e.State.PendingInterrupt != nil {
		e.Log("等待减伤技能响应...")
		return nil // 暂停执行，等待中断处理
	}

	// 4. 使用修改后的伤害值
	finalDamage := damageVal
	if finalDamage < 0 {
		finalDamage = 0
	}

	// 6. 应用伤害（扣除生命值/摸牌）
	e.setCombatStage(model.CombatStageApply)
	e.applyDamage(victim, finalDamage, damageType)

	return nil
}

func (e *GameEngine) applyAttackDamageModifiers(attacker, target *model.Player, baseDamage int, action model.Action) int {
	damage := baseDamage
	if attacker != nil && target != nil {
		modifyDamageCtx := &model.EventContext{
			Type:      model.EventAttack,
			SourceID:  action.SourceID,
			TargetID:  action.TargetID,
			DamageVal: &damage,
			Card:      action.Card,
			AttackInfo: &model.AttackEventInfo{
				ActionType:       string(action.Type),
				IsHit:            true,
				CanBeResponded:   true,
				CounterInitiator: action.CounterInitiator,
			},
		}
		modifyCtx := e.buildContext(attacker, target, model.TimingOnDamageCalculated, modifyDamageCtx)
		e.dispatcher.OnTiming(modifyCtx.Timing, modifyCtx)
	}
	return e.applyPassiveAttackEffects(attacker, target, damage, action)
}

// resolveCombatDamage 结算战斗伤害（从 CombatStack 栈顶）

// 使用新的 ResolveDamage 函数

// clearCombatStack 清空战斗栈
func (e *GameEngine) clearCombatStack() {
	e.State.CombatStack = []model.CombatRequest{}
	e.clearCombatStage()
}

// finishTakeHit 完成受到伤害后的流程 (扣血、事件、回合结束)

// 4. 执行扣血

// 5. 触发伤害承受事件

// 受伤响应可能产生中断（例如减伤/弃牌等），等待用户处理后继续

// 6. 触发攻击行动结束事件

// 攻击后响应（如神圣追击）出现中断时，暂停，避免提前切回合

// 7. 检查圣剑第3次攻击的摸X弃X效果

// 等待中断处理完成后再继续

// 8. 回到额外行动阶段，交由状态机统一处理 PendingActions/回合结束
// 这里已手动触发过一次 OnPhaseEnd，清空 LastActionType 防止重复触发

// holySwordDrawInterruptIfNeeded 在满足条件时推送圣剑摸X弃X中断
func (e *GameEngine) holySwordDrawInterruptIfNeeded(attacker *model.Player) bool {
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
	e.setReturnPoint(model.TurnStageExtraAction)

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

func (e *GameEngine) maybeHolySwordDrawFromPhaseEndCtx(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil || ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if !e.holySwordDrawInterruptIfNeeded(ctx.User) {
		return false
	}
	// 圣剑中断先打断当前 ActionEnd，处理完后回到同一个 ActionEnd 继续派发风怒/剑影等响应技能。
	ctx.User.TurnState.MarkActionEndNeedsInterruptHookSkipOnce()
	e.setReturnPoint(model.TurnStageActionEnd)
	return true
}

// addCampResource 添加阵营资源 (水晶或宝石)，战绩区总上限为 5。
// 返回 true 表示本次成功增加资源。
func (e *GameEngine) addCampResource(camp model.Camp, resourceType string) bool {
	const maxTotalResources = 5

	if camp == model.RedCamp {
		if resourceType == "gem" {
			currentTotal := e.State.RedCrystals + e.State.RedGems
			if currentTotal < maxTotalResources {
				e.State.RedGems++
				fmt.Printf("[Combat] 攻击命中！红方阵营获得 1 宝石\n")
				return true
			}
			fmt.Printf("[Combat] 攻击命中！红方阵营资源已满，无法获得宝石\n")
		} else if resourceType == "crystal" {
			currentTotal := e.State.RedCrystals + e.State.RedGems
			if currentTotal < maxTotalResources {
				e.State.RedCrystals++
				fmt.Printf("[Combat] 红方阵营获得 1 水晶\n")
				return true
			}
			fmt.Printf("[Combat] 红方阵营资源已满，获得的水晶被丢弃\n")
		}
	} else {
		// Blue Camp
		if resourceType == "gem" {
			currentTotal := e.State.BlueCrystals + e.State.BlueGems
			if currentTotal < maxTotalResources {
				e.State.BlueGems++
				fmt.Printf("[Combat] 攻击命中！蓝方阵营获得 1 宝石\n")
				return true
			}
			fmt.Printf("[Combat] 攻击命中！蓝方阵营资源已满，无法获得宝石\n")
		} else if resourceType == "crystal" {
			currentTotal := e.State.BlueCrystals + e.State.BlueGems
			if currentTotal < maxTotalResources {
				e.State.BlueCrystals++
				fmt.Printf("[Combat] 蓝方阵营获得 1 水晶\n")
				return true
			}
			fmt.Printf("[Combat] 蓝方阵营资源已满，获得的水晶被丢弃\n")
		}
	}
	return false
}

// rollbackCampResource 回滚一次命中后发放的战绩资源，返回 true 表示回滚成功。

// containsString 检查字符串切片是否包含指定字符串

// applyPassiveAttackEffects 应用攻击者的被动技能效果
func (e *GameEngine) applyPassiveAttackEffects(attacker, target *model.Player, baseDamage int, action model.Action) int {
	return e.dispatchTimingOnDamageCalculated(timingOnDamageCalculatedContext{
		Op:       timingOnDamageCalculatedAttackPassive,
		Attacker: attacker,
		Target:   target,
		Action:   action,
		Damage:   baseDamage,
	}).Damage
}

// applyDamageWithOptions 应用伤害逻辑 (治疗抵消 + 摸牌)
func (e *GameEngine) applyDamageWithOptions(target *model.Player, damage int, damageType model.DamageType, capToHandLimit bool, sourceID string, sourceSkillID string, overflowMoraleLossFixed int) {

	// 5. 造成伤害 (摸牌)
	if damage > 0 {
		e.setCombatStage(model.CombatStageDraw)
		e.Log(fmt.Sprintf("[Damage] %s 受到 %d 点伤害 (摸牌)\n", target.Name, damage))

		// 触发摸牌前事件 (允许水影等技能干预)
		drawCtx := e.newDrawContext(target, damage, "damage_draw")
		if drawCtx == nil {
			return
		}
		drawCtx.Flags["IsMagicDamage"] = !strings.EqualFold(string(damageType), string(model.AttackDamage))
		drawCtx.Flags["FromDamageDraw"] = true
		if capToHandLimit {
			drawCtx.Flags["capToHandLimit"] = true
		}
		if drawCtx.Selections == nil {
			drawCtx.Selections = map[string]any{}
		}
		if sourceID != "" {
			drawCtx.Selections["damage_source_id"] = sourceID
		}
		if sourceSkillID != "" {
			drawCtx.Selections["damage_source_skill_id"] = sourceSkillID
		}
		if overflowMoraleLossFixed > 0 {
			drawCtx.Selections["overflow_morale_loss_fixed"] = overflowMoraleLossFixed
		}
		if strings.EqualFold(string(damageType), "magic_no_morale") {
			drawCtx.Flags["NoMoraleLoss"] = true
		}

		// 若当前处于回合内基础结算或伤害结算链路，爆牌后应继续当前主流程。
		if e.State.TurnStage == model.TurnStageBeforeAction || e.isDamageResolutionActive() {
			drawCtx.Flags["StayInTurn"] = true
		}
		e.startDraw(drawCtx)
	} else {
		e.Log("[Damage] 伤害被完全抵消")
	}
}

// applyDamage 应用伤害逻辑 (治疗抵消 + 摸牌)
func (e *GameEngine) applyDamage(target *model.Player, damage int, damageType model.DamageType) {
	e.applyDamageWithOptions(target, damage, damageType, false, "", "", 0)
}

// resolveCounterAttack resolves a counter attack by popping the current combat
// from the stack and creating a new reflected combat request.
func (e *GameEngine) resolveCounterAttack(counterPlayerID, counterTargetID string, counterCard model.Card) {
	if e.State == nil || len(e.State.CombatStack) == 0 {
		return
	}
	// Pop the original combat
	e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
	// Create the reflected combat
	e.initCombat(counterPlayerID, counterTargetID, &counterCard, false, true, false, nil, true)
}
