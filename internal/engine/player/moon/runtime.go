package moon

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func ensureTokens(p *model.Player) {
	if p != nil && p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
}

func hasElementDarkMoon(user *model.Player, ele model.Element) bool {
	if user == nil || ele == "" {
		return false
	}
	for _, fc := range DarkMoonCovers(user) {
		if fc.Card.Element == ele {
			return true
		}
	}
	return false
}

func applyDarkMoonCurse(rt engineplayer.ChoiceRuntime, p *model.Player, removed int) {
	if p == nil || removed <= 0 {
		return
	}
	defer rt.PoseChangeGuard()
	actual := rt.ApplyCampMoraleLoss(p.Camp, removed)
	rt.Log(fmt.Sprintf("%s 的 [暗月诅咒] 触发：移除%d个暗月，我方士气-%d", p.Name, removed, actual))
	rt.CheckGameEnd()
}

// RemoveDarkMoonByFieldIndex removes a dark moon by field index and applies curse.
func RemoveDarkMoonByFieldIndex(rt engineplayer.ChoiceRuntime, p *model.Player, fieldIdx int) (model.Card, bool) {
	if p == nil || fieldIdx < 0 || fieldIdx >= len(p.Field) {
		return model.Card{}, false
	}
	fc := p.Field[fieldIdx]
	if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
		return model.Card{}, false
	}
	card := fc.Card
	p.RemoveFieldCard(fc)
	applyDarkMoonCurse(rt, p, 1)
	return card, true
}

// RemoveDarkMoonAny removes up to n dark moons and applies curse.
func RemoveDarkMoonAny(rt engineplayer.ChoiceRuntime, p *model.Player, n int) []model.Card {
	if p == nil || n <= 0 {
		return nil
	}
	var removed []model.Card
	for _, fc := range append([]*model.FieldCard{}, p.Field...) {
		if len(removed) >= n {
			break
		}
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
			continue
		}
		removed = append(removed, fc.Card)
		p.RemoveFieldCard(fc)
	}
	if len(removed) > 0 {
		applyDarkMoonCurse(rt, p, len(removed))
	}
	return removed
}

// EnemyIDs returns enemy player IDs for the given user's camp.
func EnemyIDs(rt engineplayer.ChoiceRuntime, user *model.Player) []string {
	if rt == nil || user == nil {
		return nil
	}
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		p := rt.GetPlayers()[pid]
		if p == nil || p.Camp == user.Camp {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

// TryQueueBlasphemy queues the blasphemy (月渎) interrupt.
func TryQueueBlasphemy(rt engineplayer.ChoiceRuntime, pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 {
		return false
	}
	if !runtimeutil.IsMagicDamageType(pd.DamageType) {
		return false
	}
	source := rt.GetPlayers()[pd.SourceID]
	if source == nil || !engineplayer.IsCharacter(source, "moon_goddess") {
		return false
	}
	order := rt.GetPlayerOrder()
	idx := rt.GetCurrentTurnIndex()
	currentTurnSource := idx >= 0 && idx < len(order) && order[idx] == source.ID
	if !source.IsActive && !currentTurnSource {
		return false
	}
	target := rt.GetPlayers()[pd.TargetID]
	if target == nil || target.Camp == source.Camp {
		return false
	}
	ensureTokens(source)
	if source.TurnState.UsedSkillCounts["mg_blasphemy"] > 0 {
		return false
	}
	if source.TurnState.SkillFlowState["mg_blasphemy_pending"] > 0 {
		return false
	}
	if source.Heal <= 0 {
		return false
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: source.ID,
		Context: map[string]interface{}{
			"choice_type":            "mg_blasphemy_target",
			"user_id":                source.ID,
			"target_ids":             []string{target.ID},
			"source_id":              pd.SourceID,
			"context_pending_damage": pd,
		},
	})
	source.TurnState.SkillFlowState["mg_blasphemy_pending"] = 1
	rt.Log(fmt.Sprintf("%s 的 [月渎] 可触发：请选择目标（或跳过）", source.Name))
	return true
}

// MaybeMedusa checks and queues the Medusa's Eye (美杜莎之眼) interrupt.
func MaybeMedusa(rt engineplayer.ChoiceRuntime, attacker, target *model.Player, sourceSkill string, attackCard *model.Card, userCtx *model.Context) bool {
	if attacker == nil || target == nil || attackCard == nil {
		return false
	}
	if userCtx == nil || !userCtx.AttackDeclaredPhase() {
		return false
	}
	if sourceSkill == "adventurer_fraud" || sourceSkill == "hb_holy_shard_storm" {
		return false
	}
	if attackCard.Element == "" {
		return false
	}
	for _, pid := range rt.GetPlayerOrder() {
		p := rt.GetPlayers()[pid]
		if p == nil || p.Camp == attacker.Camp || !engineplayer.IsCharacter(p, "moon_goddess") {
			continue
		}
		if !hasElementDarkMoon(p, attackCard.Element) {
			continue
		}
		var selectable []int
		for i, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
				continue
			}
			if fc.Card.Element != attackCard.Element {
				continue
			}
			selectable = append(selectable, i)
		}
		if len(selectable) == 0 {
			continue
		}
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: p.ID,
			Context: map[string]interface{}{
				"choice_type":      "mg_medusa_darkmoon_pick",
				"user_id":          p.ID,
				"attacker_id":      attacker.ID,
				"attack_element":   string(attackCard.Element),
				"darkmoon_indices": selectable,
				"user_ctx":         userCtx,
				"source_skill":     sourceSkill,
			},
		})
		rt.Log(fmt.Sprintf("%s 的 [美杜莎之眼] 可触发：请选择要展示并移除的%s系暗月", p.Name, attackCard.Element))
		return true
	}
	return false
}

// MaybeMoonCycleAtTurnEnd checks and queues the Moon cycle (月之轮回) interrupt.
func MaybeMoonCycleAtTurnEnd(rt engineplayer.ChoiceRuntime, p *model.Player) bool {
	if p == nil || !engineplayer.IsCharacter(p, "moon_goddess") {
		return false
	}
	ensureTokens(p)
	if p.TurnState.UsedSkillCounts == nil {
		p.TurnState.UsedSkillCounts = map[string]int{}
	}
	if p.TurnState.UsedSkillCounts["mg_moon_cycle"] > 0 {
		return false
	}
	canBranch1 := DarkMoonCount(p) > 0
	canBranch2 := p.Heal > 0
	if !canBranch1 && !canBranch2 {
		return false
	}
	var modes []string
	if canBranch1 {
		modes = append(modes, "branch1")
	}
	if canBranch2 {
		modes = append(modes, "branch2")
	}
	p.TurnState.UsedSkillCounts["mg_moon_cycle"] = 1
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type": "mg_moon_cycle_mode",
			"user_id":     p.ID,
			"modes":       modes,
			"target_ids":  rt.GetPlayerOrder(),
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [月之轮回] 触发：请选择发动分支", p.Name))
	return true
}
