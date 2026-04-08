package engine

import (
	"fmt"

	"starcup-engine/internal/engine/runtimeutil"
	"starcup-engine/internal/model"
)

func (e *GameEngine) maybeTriggerMoonGoddessMedusa(attacker *model.Player, target *model.Player, sourceSkill string, attackCard *model.Card, userCtx *model.Context) bool {
	if attacker == nil || target == nil || attackCard == nil {
		return false
	}
	// 美杜莎之眼仅允许在“攻击开始”时机触发，避免攻击结算后误触发。
	if userCtx == nil || userCtx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	// 欺诈/圣屑飓暴属于“转化攻击”，不触发美杜莎之眼。
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

func (e *GameEngine) maybeTriggerMoonGoddessMoonCycleAtTurnEnd(player *model.Player) bool {
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
			"choice_type": "mg_blasphemy_target",
			"user_id":     source.ID,
			"target_ids":  []string{target.ID},
			"source_id":   pd.SourceID,
			"trigger_pd":  pd,
		},
	})
	source.TurnState.SkillFlowState["mg_blasphemy_pending"] = 1
	e.Log(fmt.Sprintf("%s 的 [月渎] 可触发：请选择目标（或跳过）", source.Name))
	return true
}

func (e *GameEngine) campMorale(camp model.Camp) int {
	if camp == model.RedCamp {
		return e.State.RedMorale
	}
	return e.State.BlueMorale
}

func (e *GameEngine) pendingDiscardVictimID() string {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptDiscard {
		return ""
	}
	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return ""
	}
	victimID, _ := data["victim_id"].(string)
	return victimID
}
