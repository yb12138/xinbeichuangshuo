// gameflow: 月女神（Moon Goddess）亵渎队列、美杜莎之眼、月之轮回辅助。
package engine

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
)

func (e *GameEngine) tryQueueMoonGoddessBlasphemy(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 || !runtimeutil.IsMagicDamageType(pd.DamageType) {
		return false
	}
	source := e.State.Players[pd.SourceID]
	if source == nil || !e.isMoonGoddess(source) {
		return false
	}
	currentTurnSource := false
	if e.State != nil && e.State.CurrentTurn >= 0 && e.State.CurrentTurn < len(e.State.PlayerOrder) {
		currentTurnSource = e.State.PlayerOrder[e.State.CurrentTurn] == source.ID
	}
	if !source.IsActive && !currentTurnSource {
		return false
	}
	target := e.State.Players[pd.TargetID]
	if target == nil || target.Camp == source.Camp {
		return false
	}
	ensurePlayerTokensMap(source)
	if source.TurnState.UsedSkillCounts["mg_blasphemy"] > 0 {
		return false
	}
	if source.TurnState.SkillFlowState["mg_blasphemy_pending"] > 0 {
		return false
	}
	if source.Heal <= 0 {
		return false
	}
	e.PushInterrupt(&model.Interrupt{
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
	e.Log(fmt.Sprintf("%s 的 [月渎] 可触发：请选择目标（或跳过）", source.Name))
	return true
}

func (e *GameEngine) maybeMoonGoddessMedusa(attacker *model.Player, target *model.Player, sourceSkill string, attackCard *model.Card, userCtx *model.Context) bool {
	if attacker == nil || target == nil || attackCard == nil {
		return false
	}
	// 美杜莎之眼仅允许在"攻击开始"时机触发，避免攻击结算后误触发。
	if userCtx == nil || !userCtx.AttackDeclaredPhase() {
		return false
	}
	// 欺诈/圣屑飓暴属于"转化攻击"，不触发美杜莎之眼。
	if sourceSkill == "adventurer_fraud" || sourceSkill == "hb_holy_shard_storm" {
		return false
	}
	if attackCard.Element == "" {
		return false
	}
	// 只有攻击方的对立阵营（被攻击方阵营）中的月之女神可触发。
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp == attacker.Camp || !e.isMoonGoddess(p) {
			continue
		}
		if !e.moonGoddessHasElementDarkMoon(p, attackCard.Element) {
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
		e.PushInterrupt(&model.Interrupt{
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
		e.Log(fmt.Sprintf("%s 的 [美杜莎之眼] 可触发：请选择要展示并移除的%s系暗月", p.Name, attackCard.Element))
		return true
	}
	return false
}

func (e *GameEngine) maybeMoonGoddessMoonCycleAtTurnEnd(player *model.Player) bool {
	if player == nil || !e.isMoonGoddess(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.TurnState.UsedSkillCounts == nil {
		player.TurnState.UsedSkillCounts = map[string]int{}
	}
	if player.TurnState.UsedSkillCounts["mg_moon_cycle"] > 0 {
		return false
	}
	canBranch1 := moonGoddessDarkMoonCount(player) > 0
	canBranch2 := player.Heal > 0
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
	player.TurnState.UsedSkillCounts["mg_moon_cycle"] = 1
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "mg_moon_cycle_mode",
			"user_id":     player.ID,
			"modes":       modes,
			"target_ids":  append([]string{}, e.State.PlayerOrder...),
		},
	})
	e.Log(fmt.Sprintf("%s 的 [月之轮回] 触发：请选择发动分支", player.Name))
	return true
}

func moonGoddessNewMoon(player *model.Player) int {
	return tokenValueBounded(player, "mg_new_moon", moonGoddessNewMoonCapEngine)
}

func addMoonGoddessNewMoon(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "mg_new_moon", delta, moonGoddessNewMoonCapEngine)
}

func moonGoddessPetrify(player *model.Player) int {
	return tokenValueBounded(player, "mg_petrify", moonGoddessPetrifyCapEngine)
}

func addMoonGoddessPetrify(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "mg_petrify", delta, moonGoddessPetrifyCapEngine)
}

func moonGoddessDarkMoonCovers(player *model.Player) []*model.FieldCard {
	return coverCardsByEffect(player, model.EffectMoonDarkMoon)
}

func moonGoddessDarkMoonCount(player *model.Player) int {
	count := len(moonGoddessDarkMoonCovers(player))
	if player != nil && count <= 0 {
		leaveMoonGoddessDarkMoonForm(player)
	}
	return count
}

func addMoonGoddessDarkMoonCards(player *model.Player, cards []model.Card) int {
	if player == nil || len(cards) == 0 {
		return 0
	}
	added := 0
	for _, c := range cards {
		player.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  player.ID,
			SourceID: player.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectMoonDarkMoon,
			Hook:     model.FieldHookManual,
		})
		added++
	}
	if added > 0 {
		enterMoonGoddessDarkMoonForm(player)
	}
	moonGoddessDarkMoonCount(player)
	return added
}

func (e *GameEngine) applyMoonGoddessDarkMoonCurse(player *model.Player, removed int) {
	if player == nil || removed <= 0 {
		return
	}
	beforePoses := e.snapshotPlayerPoses()
	actual := e.applyCampMoraleLoss(player.Camp, removed)
	e.Log(fmt.Sprintf("%s 的 [暗月诅咒] 触发：移除%d个暗月，我方士气-%d", player.Name, removed, actual))
	moonGoddessDarkMoonCount(player)
	e.dispatchOrientationChanges(beforePoses)
	e.checkGameEnd()
}

func (e *GameEngine) removeMoonGoddessDarkMoonByFieldIndex(player *model.Player, fieldIdx int) (model.Card, bool) {
	if player == nil || fieldIdx < 0 || fieldIdx >= len(player.Field) {
		return model.Card{}, false
	}
	fc := player.Field[fieldIdx]
	if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
		return model.Card{}, false
	}
	card := fc.Card
	player.RemoveFieldCard(fc)
	e.applyMoonGoddessDarkMoonCurse(player, 1)
	return card, true
}

func (e *GameEngine) removeMoonGoddessDarkMoonAny(player *model.Player, n int) []model.Card {
	if player == nil || n <= 0 {
		return nil
	}
	var removed []model.Card
	for _, fc := range append([]*model.FieldCard{}, player.Field...) {
		if len(removed) >= n {
			break
		}
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
			continue
		}
		removed = append(removed, fc.Card)
		player.RemoveFieldCard(fc)
	}
	if len(removed) > 0 {
		e.applyMoonGoddessDarkMoonCurse(player, len(removed))
	}
	return removed
}

func (e *GameEngine) moonGoddessEnemyIDs(user *model.Player) []string {
	if e == nil || user == nil {
		return nil
	}
	return e.campEnemyIDs(user.Camp)
}

func (e *GameEngine) moonGoddessHasElementDarkMoon(user *model.Player, ele model.Element) bool {
	if user == nil || ele == "" {
		return false
	}
	for _, fc := range moonGoddessDarkMoonCovers(user) {
		if fc.Card.Element == ele {
			return true
		}
	}
	return false
}
