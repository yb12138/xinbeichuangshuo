package server

import (
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/stateview"
)

// buildAvailableActionSkills 返回当前玩家可发动的主动技能列表（用于前端按钮可用态）。
func (r *Room) buildAvailableActionSkills(playerID string) []AvailableSkill {
	p := r.Engine.State.Players[playerID]
	if p == nil || p.Character == nil {
		return nil
	}

	forcedDoomsdayOnly := p.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0
	var list []AvailableSkill
	for _, sd := range p.Character.Skills {
		if sd.Type != model.SkillTypeAction {
			continue
		}
		if forcedDoomsdayOnly && sd.ID != "arbiter_doomsday" {
			continue
		}
		if sd.ID == "adventurer_fraud" {
			// 欺诈：至少满足其一
			// 1) 有2张同系可用于弃牌
			// 2) 有3张同系可作为暗灭攻击
			elemCount := map[model.Element]int{}
			for _, c := range p.Hand {
				elemCount[c.Element]++
			}
			canUseFraud := false
			for ele, n := range elemCount {
				if ele != "" && n >= 2 {
					canUseFraud = true
					break
				}
				if n >= 3 {
					canUseFraud = true
					break
				}
			}
			if !canUseFraud {
				continue
			}
		}
		if sd.ID == "onmyoji_shikigami_descend" {
			factionCount := map[string]int{}
			hasSameFactionPair := false
			for _, c := range p.Hand {
				if c.Faction == "" {
					continue
				}
				factionCount[c.Faction]++
				if factionCount[c.Faction] >= 2 {
					hasSameFactionPair = true
					break
				}
			}
			if !hasSameFactionPair {
				continue
			}
		}
		if sd.ID == "mb_thunder_scatter" {
			if p.TurnState.UsedSkillCounts["mb_charge_lock_turn"] > 0 {
				continue
			}
			if stateview.CountMagicBowChargesByElement(p, model.ElementThunder) <= 0 {
				continue
			}
		}
		if sd.ID == "bd_dissonance_chord" {
			inspiration := 0
			if p.Tokens != nil {
				inspiration = p.Tokens["bd_inspiration"]
			}
			if inspiration <= 1 {
				continue
			}
		}
		if sd.ID == "elementalist_ignite" {
			element := 0
			if p.Tokens != nil {
				element = p.Tokens["element"]
			}
			if element < 3 {
				continue
			}
		}
		// 回合限定：本回合已用过则不再展示。
		if model.ContainsSkillTag(sd.Tags, model.TagTurnLimit) {
			if p.TurnState.UsedSkillCounts[sd.ID] > 0 {
				continue
			}
		}
		// 必杀技资源过滤规则：
		// - 宝石消耗必须由宝石支付
		// - 水晶消耗可由“剩余宝石”替代
		if sd.CostGem > 0 || sd.CostCrystal > 0 {
			if p.Gem < sd.CostGem {
				continue
			}
			usableCrystal := p.Crystal + (p.Gem - sd.CostGem)
			if usableCrystal < sd.CostCrystal {
				continue
			}
		}
		// 独有技：必须拥有对应独有牌（手牌或专属卡区）才能使用。
		if sd.RequireExclusive {
			if !p.HasExclusiveCard(p.Character.ID, sd.Title) {
				continue
			}
		}
		// 通用可用性兜底：复用技能 Handler 的 CanUse，提前过滤“指示物不足/形态不符”等条件。
		// 这样前端会直接把技能置灰（或不展示），避免点击后才报“技能发动失败”。
		if !r.canUseActionSkillNow(p, sd) {
			continue
		}
		list = append(list, AvailableSkill{
			ID:               sd.ID,
			Title:            sd.Title,
			Description:      sd.Description,
			MinTargets:       sd.MinTargets,
			MaxTargets:       sd.MaxTargets,
			TargetType:       int(sd.TargetType),
			CostGem:          sd.CostGem,
			CostCrystal:      sd.CostCrystal,
			CostDiscards:     sd.CostDiscards,
			DiscardType:      string(sd.DiscardType),
			DiscardElement:   string(sd.DiscardElement),
			RequireExclusive: sd.RequireExclusive,
			PlaceCard:        sd.PlaceCard,
			PlaceEffect:      string(sd.PlaceEffect),
		})
	}
	return list
}

func (r *Room) probeTargetForActionSkill(user *model.Player, targetType model.TargetType) *model.Player {
	if r == nil || r.Engine == nil || user == nil {
		return nil
	}
	switch targetType {
	case model.TargetSelf:
		return user
	case model.TargetEnemy:
		for _, p := range r.Engine.State.Players {
			if p != nil && p.Camp != user.Camp {
				return p
			}
		}
	case model.TargetAlly:
		for _, p := range r.Engine.State.Players {
			if p != nil && p.Camp == user.Camp && p.ID != user.ID {
				return p
			}
		}
	case model.TargetAllySelf:
		return user
	case model.TargetAny, model.TargetSpecific:
		return user
	}
	return nil
}

func (r *Room) canUseActionSkillNow(user *model.Player, sd model.SkillDefinition) bool {
	if r == nil || r.Engine == nil || user == nil {
		return false
	}
	if sd.LogicHandler == "" {
		return true
	}
	handler := skills.GetHandler(sd.LogicHandler)
	if handler == nil {
		return true
	}
	probeTarget := r.probeTargetForActionSkill(user, sd.TargetType)
	targetID := user.ID
	if probeTarget != nil {
		targetID = probeTarget.ID
	}
	ctx := &model.Context{
		Game:    r.Engine,
		User:    user,
		Target:  probeTarget,
		Trigger: model.TriggerNone,
		TriggerCtx: &model.EventContext{
			Type:     model.EventNone,
			SourceID: user.ID,
			TargetID: targetID,
		},
		Selections: map[string]any{},
		Flags:      map[string]bool{},
	}
	if probeTarget != nil {
		ctx.Targets = []*model.Player{probeTarget}
	}
	return handler.CanUse(ctx)
}
