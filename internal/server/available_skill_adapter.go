package server

import (
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/bot"
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
		// 弃牌成本可达成性（非独有技）：检查手牌中是否有满足元素/类型要求的牌。
		// 独有技的可用性已由上方 HasExclusiveCard 保证，其 DiscardElement 仅为元数据，不作为发动门控。
		if !sd.RequireExclusive && !r.Engine.CanSatisfyActionSkillDiscardRequirement(p, sd) {
			continue
		}
		// 通用可用性兜底：复用技能 Handler 的 CanUse，提前过滤”指示物不足/形态不符”等条件。
		// 这样前端会直接把技能置灰（或不展示），避免点击后才报“技能发动失败”。
		if !r.canUseActionSkillNow(p, sd) {
			continue
		}
		targetType := int(sd.TargetType)
		minTargets := sd.MinTargets
		maxTargets := sd.MaxTargets
		// 风之洁净改为“无目标发动”：弃牌后直接进入移除基础效果流程（或在无效果时直接完成）。
		if sd.ID == "angel_cleanse" {
			targetType = int(model.TargetNone)
			minTargets = 0
			maxTargets = 0
		}
		list = append(list, AvailableSkill{
			ID:               sd.ID,
			Title:            sd.Title,
			Description:      sd.Description,
			MinTargets:       minTargets,
			MaxTargets:       maxTargets,
			TargetType:       targetType,
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
	// PlaceCard 主动技里有一类“放置现在、触发在未来”的场牌（如五系封印）。
	// 这些 handler 的 CanUse 语义是“未来触发时机是否成立”，不适合作为“当前可施放性”过滤，
	// 否则会被 TimingActive 探测误判为 false，导致前端按钮错误置灰。
	if sd.Type == model.SkillTypeAction &&
		sd.PlaceCard &&
		sd.PlaceMode == model.FieldEffect &&
		sd.PlaceHook != model.FieldHookManual {
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
		Game:   r.Engine,
		User:   user,
		Target: probeTarget,
		Timing: model.TimingActive,
		EventCtx: &model.EventContext{
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

// buildBotAvailableSkills 构建 bot 决策所需的可用技能列表
func buildBotAvailableSkills(skills []AvailableSkill) []bot.AvailableSkill {
	result := make([]bot.AvailableSkill, 0, len(skills))
	for _, sk := range skills {
		result = append(result, bot.AvailableSkill{
			ID:               sk.ID,
			Title:            sk.Title,
			Description:      sk.Description,
			MinTargets:       sk.MinTargets,
			MaxTargets:       sk.MaxTargets,
			TargetType:       sk.TargetType,
			CostGem:          sk.CostGem,
			CostCrystal:      sk.CostCrystal,
			CostDiscards:     sk.CostDiscards,
			DiscardType:      sk.DiscardType,
			DiscardElement:   sk.DiscardElement,
			RequireExclusive: sk.RequireExclusive,
			PlaceCard:        sk.PlaceCard,
			PlaceEffect:      sk.PlaceEffect,
		})
	}
	return result
}
