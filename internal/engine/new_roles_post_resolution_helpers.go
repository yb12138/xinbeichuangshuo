// gameflow: 新角色批次：延迟伤害后/攻击后等公共解析（含祈祷师赐福确认等）。

package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
)

func (e *GameEngine) resolveMagicLancerStardustAfterSelf(user *model.Player) bool {
	if user == nil || !e.isMagicLancer(user) {
		return false
	}
	if user.TurnState.SkillFlowState == nil || user.TurnState.SkillFlowState["ml_stardust_pending"] <= 0 {
		return false
	}

	// 若还在等待本次自伤导致的爆牌弃牌，则延后到 ConfirmDiscard 再判定。
	if e.pendingDiscardVictimID() == user.ID {
		user.TurnState.SkillFlowState["ml_stardust_wait_discard"] = 1
		return false
	}

	before := user.TurnState.SkillFlowState["ml_stardust_morale_before"]
	current := e.campMorale(user.Camp)
	user.TurnState.SkillFlowState["ml_stardust_pending"] = 0
	user.TurnState.SkillFlowState["ml_stardust_wait_discard"] = 0
	user.TurnState.SkillFlowState["ml_stardust_morale_before"] = 0

	if hasMagicLancerPhantomForm(user) {
		beforePoses := e.snapshotPlayerPoses()
		leaveMagicLancerPhantomForm(user)
		e.Log(fmt.Sprintf("%s 的 [幻影星尘] 结算完成，脱离幻影形态并转正", user.Name))
		e.dispatchOrientationChanges(beforePoses)
	}

	if before > 0 && current < before {
		e.Log(fmt.Sprintf("%s 的 [幻影星尘] 未触发后续伤害：本次自伤导致己方士气下降", user.Name))
		return false
	}

	targetIDs := make([]string, 0, len(e.State.PlayerOrder))
	for _, pid := range e.State.PlayerOrder {
		if p := e.State.Players[pid]; p != nil && p.Camp != user.Camp {
			targetIDs = append(targetIDs, pid)
		}
	}
	lockedOrder := user.TurnState.SkillFlowState["ml_stardust_locked_target_order"]
	user.TurnState.SkillFlowState["ml_stardust_locked_target_order"] = 0
	if len(targetIDs) == 0 {
		return false
	}
	if lockedOrder > 0 && lockedOrder <= len(e.State.PlayerOrder) {
		lockedID := e.State.PlayerOrder[lockedOrder-1]
		for _, tid := range targetIDs {
			if tid != lockedID {
				continue
			}
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   lockedID,
				Damage:     2,
				DamageType: model.MagicDamage,
			})
			if target := e.State.Players[lockedID]; target != nil {
				e.Log(fmt.Sprintf("%s 的 [幻影星尘] 生效：对 %s 造成2点法术伤害", user.Name, target.Name))
			}
			return false
		}
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
		if attacker.TurnState.SkillFlowState["hb_shard_miss_pending"] > 0 {
			attacker.TurnState.SkillFlowState["hb_shard_miss_pending"] = 0
		}
		if !pd.IsCounter && pd.Card != nil && strings.TrimSpace(pd.Card.Faction) == "圣" {
			before := holyBowFaith(attacker)
			after := addHolyBowFaith(attacker, 1)
			if after > before {
				e.Log(fmt.Sprintf("%s 的 [天之弓] 触发：信仰+1（当前%d）", attacker.Name, after))
			}
		}
	}
	// 兽灵武士：普通形态下，主动攻击命中时兽魂+1。
	if e.isBeastSamurai(attacker) &&
		!pd.IsCounter &&
		!e.beastSamuraiInIaijutsuForm(attacker) {
		before := e.beastSamuraiBeastSoul(attacker)
		after := e.addBeastSamuraiBeastSoul(attacker, 1, false)
		if after > before {
			e.Log(fmt.Sprintf("%s 的 [兽魂意念] 生效：普通形态主动攻击命中，兽魂+1（当前%d）", attacker.Name, after))
		}
	}

	// 祈祷师威力赐福：命中后由UI严格确认是否移除并令本次攻击伤害+2。
	if getFieldEffectCard(attacker, model.EffectPowerBlessing) != nil {
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: attacker.ID,
			Context: map[string]interface{}{
				"choice_type": "prayer_power_blessing_followup",
				"user_id":     attacker.ID,
				"source_id":   pd.SourceID,
				"target_id":   pd.TargetID,
			},
		})
		return true
	}

	// 精灵射手：水之矢
	if attacker.TurnState.SkillFlowState["elf_elemental_shot_water_pending"] > 0 {
		attacker.TurnState.SkillFlowState["elf_elemental_shot_water_pending"] = 0
		e.Heal(pd.TargetID, 1)
		e.Log(fmt.Sprintf("%s 的 [元素射击·水之矢] 生效：%s +1治疗", attacker.Name, model.GetPlayerDisplayName(e.State.Players[pd.TargetID])))
	}

	// 精灵射手：地之矢
	if attacker.TurnState.SkillFlowState["elf_elemental_shot_earth_pending"] > 0 {
		attacker.TurnState.SkillFlowState["elf_elemental_shot_earth_pending"] = 0
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   attacker.ID,
			TargetID:   pd.TargetID,
			Damage:     1,
			DamageType: model.MagicDamage,
		})
		e.Log(fmt.Sprintf("%s 的 [元素射击·地之矢] 生效：对 %s 追加1点法术伤害", attacker.Name, model.GetPlayerDisplayName(e.State.Players[pd.TargetID])))
	}

	// 魔剑士：黄泉震颤命中后，补至上限并弃2。
	if attacker.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] > 0 {
		attacker.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] = 0
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

	// 魔弓：魔贯冲击命中后，若仍有火系充能则强制再移除1个并使本次伤害+1。
	if attacker.TurnState.SkillFlowState["mb_magic_pierce_pending"] > 0 {
		if magicBowChargeCount(attacker, model.ElementFire) <= 0 {
			attacker.TurnState.SkillFlowState["mb_magic_pierce_pending"] = 0
		} else {
			if _, ok := removeMagicBowChargeByElement(attacker, model.ElementFire); ok {
				applied := false
				for i := range e.State.PendingDamageQueue {
					queued := &e.State.PendingDamageQueue[i]
					if !strings.EqualFold(string(queued.DamageType), string(model.AttackDamage)) {
						continue
					}
					queued.Damage++
					applied = true
					break
				}
				e.Log(fmt.Sprintf("%s 的 [魔贯冲击] 命中追加生效：额外移除1个火系充能，本次攻击伤害+1", attacker.Name))
				if !applied {
					e.Log("[Warn] 魔贯冲击命中追加未找到对应伤害条目，未能叠加伤害")
				}
			}
			attacker.TurnState.SkillFlowState["mb_magic_pierce_pending"] = 0
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
	if actionType == model.ActionAttack && e.isElfArcher(player) {
		if player.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] > 0 {
			model.AppendAttackAction(player, "风之矢")
			e.Log(fmt.Sprintf("%s 的 [元素射击·风之矢] 结算：额外获得1次攻击行动", player.Name))
		}
		clearElfElementalShotCombatState(player)
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
				"choice_type": "prayer_swift_blessing_followup",
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
	if e.isBeastSamurai(source) && strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
		clearBeastSamuraiAttackTokens(source)
	}
	if e.isBeastSamurai(source) && pd.Damage > 0 && e.beastSamuraiInIaijutsuForm(source) {
		beforePoses := e.snapshotPlayerPoses()
		if e.leaveBeastSamuraiIaijutsuForm(source) {
			e.Log(fmt.Sprintf("%s 的 [御魂流居合形态·造成伤害退场] 生效：转正并脱离御魂流居合形态", source.Name))
		}
		e.dispatchOrientationChanges(beforePoses)
	}
	if e.isBlazeWitch(source) && source.TurnState.SkillFlowState != nil && source.TurnState.SkillFlowState["bw_pain_link_pending_discard"] > 0 {
		if source.TurnState.SkillFlowState["bw_pain_link_pending_hits"] > 0 {
			source.TurnState.SkillFlowState["bw_pain_link_pending_hits"]--
		}
		if source.TurnState.SkillFlowState["bw_pain_link_pending_hits"] <= 0 {
			source.TurnState.SkillFlowState["bw_pain_link_pending_hits"] = 0
			source.TurnState.SkillFlowState["bw_pain_link_pending_discard"] = 0
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
		source.TurnState.SkillFlowState != nil &&
		source.TurnState.SkillFlowState["ml_stardust_pending"] > 0 &&
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
	if target != nil && e.isSage(target) && runtimeutil.IsMagicDamageType(pd.DamageType) {
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
					Type:     model.InterruptDiscard,
					PlayerID: target.ID,
					Context: map[string]interface{}{
						"discard_count":        1,
						"stay_in_turn":         true,
						"is_damage_resolution": true,
						"prompt":               "【智慧法典】请选择弃置1张手牌：",
					},
				})
				_ = e.tryQueueMoonGoddessBlasphemy(pd)
				return true
			}
		}
		if pd.Damage == 1 {
			sameCount := maxSameElementCount(target)
			if sameCount >= 2 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: target.ID,
					Context: map[string]interface{}{
						"choice_type": "sage_magic_rebound_confirm",
						"user_id":     target.ID,
					},
				})
				e.Log(fmt.Sprintf("%s 的 [法术反弹] 可触发：承受1点法术伤害，最大同系手牌=%d", target.Name, sameCount))
				_ = e.tryQueueMoonGoddessBlasphemy(pd)
				return true
			}
			e.Log(fmt.Sprintf("%s 的 [法术反弹] 未触发：承受1点法术伤害但同系手牌不足2（当前最大同系=%d）", target.Name, sameCount))
		}
	}
	if pd.Damage > 0 && runtimeutil.IsMagicDamageType(pd.DamageType) {
		if e.tryBardDescentAfterMagicDamage(pd) {
			_ = e.tryQueueMoonGoddessBlasphemy(pd)
			return true
		}
	}
	if e.queueElfAnimalResponse(source, target, pd) {
		_ = e.tryQueueMoonGoddessBlasphemy(pd)
		return true
	}
	if e.tryQueueMoonGoddessBlasphemy(pd) {
		return true
	}
	return false
}
