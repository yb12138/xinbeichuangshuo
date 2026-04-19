// gameflow: 蝶舞者：茧与蝶蛹相关辅助。
package engine

import (
	"fmt"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func butterflyPupa(player *model.Player) int {
	return tokenValueBounded(player, "bt_pupa", -1)
}

func addButterflyPupa(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "bt_pupa", delta, -1)
}

func butterflyCocoonCovers(player *model.Player) []*model.FieldCard {
	return coverCardsByEffect(player, model.EffectButterflyCocoon)
}

func butterflyCocoonCount(player *model.Player) int {
	return len(butterflyCocoonCovers(player))
}

func syncButterflyCocoonToken(player *model.Player) {
	_ = butterflyCocoonCount(player)
}

func addButterflyCocoonCards(player *model.Player, cards []model.Card) int {
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
			Effect:   model.EffectButterflyCocoon,
			Hook:     model.FieldHookManual,
		})
		added++
	}
	syncButterflyCocoonToken(player)
	return added
}

func butterflyCocoonFieldIndices(player *model.Player) []int {
	if player == nil {
		return nil
	}
	var out []int
	for i, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		out = append(out, i)
	}
	return out
}

func removeButterflyCocoonByFieldIndex(player *model.Player, fieldIdx int) (model.Card, bool) {
	if player == nil || fieldIdx < 0 || fieldIdx >= len(player.Field) {
		return model.Card{}, false
	}
	fc := player.Field[fieldIdx]
	if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
		return model.Card{}, false
	}
	card := fc.Card
	player.RemoveFieldCard(fc)
	syncButterflyCocoonToken(player)
	return card, true
}

func removeButterflyCocoonByFieldIndices(player *model.Player, indices []int) ([]model.Card, error) {
	if player == nil {
		return nil, fmt.Errorf("玩家不存在")
	}
	if len(indices) == 0 {
		return nil, nil
	}
	seen := map[int]bool{}
	for _, idx := range indices {
		if idx < 0 || idx >= len(player.Field) {
			return nil, fmt.Errorf("无效的茧索引: %d", idx)
		}
		if seen[idx] {
			return nil, fmt.Errorf("不能重复选择同一个茧")
		}
		seen[idx] = true
		fc := player.Field[idx]
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			return nil, fmt.Errorf("选择的索引不是茧: %d", idx)
		}
	}
	// 从大到小删除，避免索引偏移。
	for i := 0; i < len(indices); i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[i] < indices[j] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
	var removed []model.Card
	for _, idx := range indices {
		fc := player.Field[idx]
		removed = append(removed, fc.Card)
		player.RemoveFieldCard(fc)
	}
	syncButterflyCocoonToken(player)
	return removed, nil
}

func butterflyMirrorPairDefs(player *model.Player) ([]string, []string) {
	if player == nil {
		return nil, nil
	}
	elemToFieldIdx := map[model.Element][]int{}
	for i, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		elemToFieldIdx[fc.Card.Element] = append(elemToFieldIdx[fc.Card.Element], i)
	}
	elements := []model.Element{
		model.ElementFire, model.ElementWater, model.ElementWind, model.ElementThunder,
		model.ElementEarth, model.ElementLight, model.ElementDark,
	}
	var defs []string
	var labels []string
	for _, ele := range elements {
		idxs := elemToFieldIdx[ele]
		if len(idxs) < 2 {
			continue
		}
		for i := 0; i < len(idxs); i++ {
			for j := i + 1; j < len(idxs); j++ {
				left := idxs[i]
				right := idxs[j]
				defs = append(defs, fmt.Sprintf("%d,%d", left, right))
				lc := player.Field[left].Card
				rc := player.Field[right].Card
				labels = append(labels, fmt.Sprintf("%s系茧：%s + %s", elementNameForPrompt(string(ele)), formatCardInfo(lc), formatCardInfo(rc)))
			}
		}
	}
	return defs, labels
}

func (e *GameEngine) queueButterflyWitherFollowup(user *model.Player) {
	if user == nil || !e.isButterflyDancer(user) {
		return
	}
	if user.TurnState.SkillFlowState == nil {
		user.TurnState.SkillFlowState = make(map[string]int)
	}
	user.TurnState.SkillFlowState["bt_wither_pending"]++
	if user.TurnState.SkillFlowState["bt_wither_pending"] > 1 {
		// 已有待处理的凋零询问，累计即可。
		return
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_wither_confirm",
			"user_id":     user.ID,
			"target_ids":  e.butterflyActionTargetIDs(),
		},
	})
	e.Log(fmt.Sprintf("%s 可发动 [凋零]：请选择是否发动", user.Name))
}

// gameflow: 蝶舞者：蛹、毒粉、伤害传递等。

func (e *GameEngine) butterflyActionTargetIDs() []string {
	if e == nil || e.State == nil {
		return nil
	}
	targetIDs := make([]string, 0, len(e.State.PlayerOrder))
	for _, pid := range e.State.PlayerOrder {
		if e.State.Players[pid] != nil {
			targetIDs = append(targetIDs, pid)
		}
	}
	return targetIDs
}

func (e *GameEngine) ResolveButterflyChrysalis(userID string) error {
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	now := addButterflyPupa(user, 1)
	cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 4)
	e.State.Deck = newDeck
	e.State.DiscardPile = newDiscard
	added := addButterflyCocoonCards(user, cards)
	e.Log(fmt.Sprintf("%s 发动 [蛹化]：蛹+1（当前%d），获得%d个茧", user.Name, now, added))
	e.checkHandLimit(user, nil)
	overflow := butterflyCocoonCount(user) - butterflyCocoonCapEngine
	if overflow > 0 {
		e.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: user.ID, Context: map[string]interface{}{"choice_type": "bt_cocoon_overflow_discard", "user_id": user.ID, "discard_count": overflow}})
	}
	return nil
}

func (e *GameEngine) StartButterflyReverse(userID string) error {
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_reverse_mode",
			"user_id":     user.ID,
			"can_branch2": butterflyPupa(user) > 0,
			"target_ids":  e.butterflyActionTargetIDs(),
		},
	})
	e.Log(fmt.Sprintf("%s 发动 [倒逆之蝶]：已弃2张牌，请选择发动分支", user.Name))
	return nil
}

// gameflow: 蝶舞者伤害传递/蛹相关辅助。

func (e *GameEngine) maybeButterflyDamageResponses(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 {
		return false
	}
	// 朝圣：承伤者若为蝶舞者，可在每次伤害中询问一次。
	if !pd.HasCheck(model.PendingDamageCheckBeforeApplyDefend) {
		pd.SetCheck(model.PendingDamageCheckBeforeApplyDefend, true)
		target := e.State.Players[pd.TargetID]
		if target != nil && e.isButterflyDancer(target) && butterflyCocoonCount(target) > 0 {
			indices := butterflyCocoonFieldIndices(target)
			if len(indices) > 0 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: target.ID,
					Context: map[string]interface{}{
						"choice_type":    "bt_pilgrimage_pick",
						"user_id":        target.ID,
						"source_id":      pd.SourceID,
						"target_id":      pd.TargetID,
						"damage_index":   0,
						"cocoon_indices": indices,
					},
				})
				e.Log(fmt.Sprintf("%s 的 [朝圣] 可触发：是否移除1个茧抵御1点伤害", target.Name))
				return true
			}
		}
	}
	// 毒粉/镜花水月仅作用于法术伤害。
	if pd.DamageType != model.MagicDamage {
		return false
	}

	// 毒粉/镜花水月：按"实际法术伤害"值检查，仅询问一次。
	if pd.HasCheck(model.PendingDamageCheckBeforeApplyResponse) {
		return false
	}
	pd.SetCheck(model.PendingDamageCheckBeforeApplyResponse, true)

	if pd.Damage == 1 {
		for _, pid := range e.State.PlayerOrder {
			user := e.State.Players[pid]
			if user == nil || !e.isButterflyDancer(user) || butterflyCocoonCount(user) <= 0 {
				continue
			}
			indices := butterflyCocoonFieldIndices(user)
			if len(indices) == 0 {
				continue
			}
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":    "bt_poison_pick",
					"user_id":        user.ID,
					"source_id":      pd.SourceID,
					"target_id":      pd.TargetID,
					"damage_index":   0,
					"cocoon_indices": indices,
				},
			})
			e.Log(fmt.Sprintf("%s 的 [毒粉] 可触发：是否移除1个茧令该次法术伤害+1", user.Name))
			return true
		}
		return false
	}

	if pd.Damage == 2 {
		for _, pid := range e.State.PlayerOrder {
			user := e.State.Players[pid]
			if user == nil || !e.isButterflyDancer(user) || butterflyCocoonCount(user) < 2 {
				continue
			}
			defs, labels := butterflyMirrorPairDefs(user)
			if len(defs) == 0 {
				continue
			}
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":  "bt_mirror_pair",
					"user_id":      user.ID,
					"source_id":    pd.SourceID,
					"target_id":    pd.TargetID,
					"damage_index": 0,
					"pair_defs":    defs,
					"pair_labels":  labels,
				},
			})
			e.Log(fmt.Sprintf("%s 的 [镜花水月] 可触发：是否移除2张同系茧改写本次伤害来源", user.Name))
			return true
		}
	}
	return false
}
