package server

import (
	"fmt"

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
		// 第一个不满足的条件决定置灰原因；其余检查照常执行但不覆盖原因。
		reason := ""
		disable := func(text string) {
			if reason == "" {
				reason = text
			}
		}
		if forcedDoomsdayOnly && sd.ID != "arbiter_doomsday" {
			disable("本回合必须发动【末日审判】。")
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
				disable("需要2张同系手牌，或3张同系牌作暗灭攻击。")
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
				disable("需要弃置2张命格相同的手牌才能发动。")
			}
		}
		if sd.ID == "ss_soul_recall" {
			hasMagic := false
			for _, c := range p.Hand {
				if c.Type == model.CardTypeMagic {
					hasMagic = true
					break
				}
			}
			if !hasMagic {
				disable("手牌中没有法术牌。")
			}
		}
		if sd.ID == "mb_thunder_scatter" {
			if p.TurnState.UsedSkillCounts["mb_charge_lock_turn"] > 0 {
				disable("本回合已移除过充能，不能再发动。")
			}
			if stateview.CountMagicBowChargesByElement(p, model.ElementThunder) <= 0 {
				disable("没有雷系充能。")
			}
		}
		if sd.ID == "bd_dissonance_chord" {
			inspiration := 0
			if p.Tokens != nil {
				inspiration = p.Tokens["bd_inspiration"]
			}
			if inspiration <= 1 {
				disable("灵感不足（至少需要2）。")
			}
		}
		if sd.ID == "elementalist_ignite" {
			element := 0
			if p.Tokens != nil {
				element = p.Tokens["element"]
			}
			if element < 3 {
				disable("元素不足3点。")
			}
		}
		// 回合限定：本回合已用过则不再展示。
		if model.ContainsSkillTag(sd.Tags, model.TagTurnLimit) {
			if p.TurnState.UsedSkillCounts[sd.ID] > 0 {
				disable("本回合已发动过。")
			}
		}
		// 必杀技资源过滤规则：
		// - 宝石消耗必须由宝石支付
		// - 水晶消耗可由“剩余宝石”替代
		if sd.CostGem > 0 || sd.CostCrystal > 0 {
			if p.Gem < sd.CostGem {
				disable(energyShortText(sd))
			} else if usableCrystal := p.Crystal + (p.Gem - sd.CostGem); usableCrystal < sd.CostCrystal {
				disable(energyShortText(sd))
			}
		}
		// 独有技：必须拥有对应独有牌（手牌或专属卡区）才能使用。
		if sd.RequireExclusive {
			if !p.HasExclusiveCard(p.Character.ID, sd.Title) {
				disable("缺少「" + sd.Title + "」独有牌。")
			}
		}
		// 弃牌成本可达成性：检查手牌中是否有满足元素/类型/独有技要求的牌。
		// CostDiscards=0 的专属卡区技能仍由上方 HasExclusiveCard 保证可用性。
		if !r.Engine.CanSatisfyActionSkillDiscardRequirement(p, sd) {
			disable(discardShortText(sd))
		}
		// 通用可用性兜底：复用技能 Handler 的 CanUse，提前过滤”指示物不足/形态不符”等条件。
		// 这样前端会直接把技能置灰（或不展示），避免点击后才报“技能发动失败”。
		if reason == "" {
			reason = r.actionSkillBlockedReason(p, sd)
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
			Disabled:         reason != "",
			DisabledReason:   reason,
		})
	}
	return list
}

// buildUsableActionSkills 只返回当前可发动的主动技（bot 决策用）。
func (r *Room) buildUsableActionSkills(playerID string) []AvailableSkill {
	all := r.buildAvailableActionSkills(playerID)
	usable := make([]AvailableSkill, 0, len(all))
	for _, sk := range all {
		if !sk.Disabled {
			usable = append(usable, sk)
		}
	}
	return usable
}

func energyShortText(sd model.SkillDefinition) string {
	return fmt.Sprintf("能量不足（需要 %d 宝石、%d 水晶）。", sd.CostGem, sd.CostCrystal)
}

// discardShortText 弃牌代价无法满足时的说明：优先用技能描述括号里的代价，没有再按配置拼。
func discardShortText(sd model.SkillDefinition) string {
	need := sd.CostDiscards
	if need <= 0 {
		need = 1
	}
	kind := "手牌"
	switch {
	case sd.DiscardElement != "" && sd.DiscardType != "":
		kind = elementName(sd.DiscardElement) + "系" + cardTypeName(sd.DiscardType)
	case sd.DiscardElement != "":
		kind = elementName(sd.DiscardElement) + "系牌"
	case sd.DiscardType != "":
		kind = cardTypeName(sd.DiscardType)
	}
	return fmt.Sprintf("可弃置的%s不足（需要 %d 张）。", kind, need)
}

func elementName(e model.Element) string {
	switch e {
	case model.ElementWater:
		return "水"
	case model.ElementFire:
		return "火"
	case model.ElementEarth:
		return "地"
	case model.ElementWind:
		return "风"
	case model.ElementThunder:
		return "雷"
	case model.ElementLight:
		return "光"
	case model.ElementDark:
		return "暗"
	}
	return string(e)
}

func cardTypeName(t model.CardType) string {
	switch t {
	case model.CardTypeAttack:
		return "攻击牌"
	case model.CardTypeMagic:
		return "法术牌"
	}
	return "牌"
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

// actionSkillBlockedReason 复用技能 Handler 的 CanUse 判定当前能否发动；不能时给出原因。
// Handler 拒绝时优先用 Handler 的说明（UsabilityReasoner），否则用通用文案。
func (r *Room) actionSkillBlockedReason(user *model.Player, sd model.SkillDefinition) string {
	if r.canUseActionSkillNow(user, sd) {
		return ""
	}
	if reasoner, ok := skills.GetHandler(sd.LogicHandler).(interface {
		BlockedReason(ctx *model.Context) string
	}); ok {
		if text := reasoner.BlockedReason(r.probeContext(user, sd)); text != "" {
			return text
		}
	}
	return "当前条件不满足，无法发动。"
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
	// 否则会被 TimingActionDuring 探测误判为 false，导致前端按钮错误置灰。
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
	return handler.CanUse(r.probeContext(user, sd))
}

// probeContext 构造“此刻发动该技能”的探测上下文（目标取第一个合法候选）。
func (r *Room) probeContext(user *model.Player, sd model.SkillDefinition) *model.Context {
	probeTarget := r.probeTargetForActionSkill(user, sd.TargetType)
	targetID := user.ID
	if probeTarget != nil {
		targetID = probeTarget.ID
	}
	ctx := &model.Context{
		Game:   r.Engine,
		User:   user,
		Target: probeTarget,
		Timing: model.TimingActionDuring,
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
	return ctx
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
