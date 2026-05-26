// gameflow: 战斗栈：发起攻击/法术战斗、阶段切换、与应战/承伤衔接。

package engine

import (
	"fmt"
	"starcup-engine/internal/model"
	"strings"
)

// initCombat 初始化战斗，将 CombatRequest 推入栈并进入战斗交互阶段
func (e *GameEngine) initCombat(attackerID, targetID string, card *model.Card, isForcedHit, canBeResponded, ignoreShield bool, interceptTags map[model.CombatInterceptTag]bool, elementOverride string, isCounter ...bool) {
	e.setCombatStage(model.CombatStageDeclare)
	e.clearSubflow()
	attacker := e.State.Players[attackerID]
	target := e.State.Players[targetID]
	if attacker != nil && target != nil && card != nil {
		e.beginNarrativeCombat(attackerID, len(isCounter) > 0 && isCounter[0])
		e.NotifyActionStep(fmt.Sprintf("%s出%s攻击%s", model.GetPlayerDisplayName(attacker), card.Name, model.GetPlayerDisplayName(target)))
		e.NotifyCombatCue(attackerID, targetID, "attack")
	}
	combatReq := model.CombatRequest{
		AttackerID:      attackerID,
		TargetID:        targetID,
		Card:            card,
		IsForcedHit:     isForcedHit,
		IgnoreShield:    ignoreShield,
		CanBeResponded:  canBeResponded,
		IsCounter:       len(isCounter) > 0 && isCounter[0],
		InterceptTags:   model.CloneCombatInterceptTags(interceptTags),
		ElementOverride: elementOverride,
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

func (e *GameEngine) ApplyAttackDamageModifiers(attacker, target *model.Player, baseDamage int, action model.Action) int {
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
		e.dispatchRuleTiming(ruleTimingDispatchInput{
			Timing:   model.TimingDamageSourceDeal,
			User:     attacker,
			Target:   target,
			EventCtx: modifyDamageCtx,
			Markers: map[string]any{
				"damage_timeline": true,
			},
		})
	}
	return e.ApplyPassiveAttackEffects(attacker, target, damage, action)
}

// clearCombatStack 清空战斗栈
func (e *GameEngine) clearCombatStack() {
	e.State.CombatStack = []model.CombatRequest{}
	e.clearCombatStage()
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
				e.recordActionResourceDelta()
				fmt.Printf("[Combat] 攻击命中！红方阵营获得 1 宝石\n")
				return true
			}
			fmt.Printf("[Combat] 攻击命中！红方阵营资源已满，无法获得宝石\n")
		} else if resourceType == "crystal" {
			currentTotal := e.State.RedCrystals + e.State.RedGems
			if currentTotal < maxTotalResources {
				e.State.RedCrystals++
				e.recordActionResourceDelta()
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
				e.recordActionResourceDelta()
				fmt.Printf("[Combat] 攻击命中！蓝方阵营获得 1 宝石\n")
				return true
			}
			fmt.Printf("[Combat] 攻击命中！蓝方阵营资源已满，无法获得宝石\n")
		} else if resourceType == "crystal" {
			currentTotal := e.State.BlueCrystals + e.State.BlueGems
			if currentTotal < maxTotalResources {
				e.State.BlueCrystals++
				e.recordActionResourceDelta()
				fmt.Printf("[Combat] 蓝方阵营获得 1 水晶\n")
				return true
			}
			fmt.Printf("[Combat] 蓝方阵营资源已满，获得的水晶被丢弃\n")
		}
	}
	return false
}

// ApplyPassiveAttackEffects 应用攻击者的被动技能效果
func (e *GameEngine) ApplyPassiveAttackEffects(attacker, target *model.Player, baseDamage int, action model.Action) int {
	return e.applyDamageSourceDealAttackModifiers(attacker, target, action, baseDamage)
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
	e.resolveCounterAttackAfterAttackMissTiming(counterPlayerID, counterTargetID, counterCard)
}
