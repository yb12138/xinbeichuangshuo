// gameflow: 新角色扩展：field resource morale helpers。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func magicBowChargeCount(player *model.Player, element model.Element) int {
	return coverCountByEffectAndElement(player, model.EffectMagicBowCharge, element)
}

func syncMagicBowChargeToken(player *model.Player) {
	// no-op: mb_charge_count 在服务端 buildStateForPlayer 中按场上盖牌派生写入 PlayerView.tokens
}

func addMagicBowChargeCards(player *model.Player, cards []model.Card) int {
	if player == nil || len(cards) == 0 {
		return 0
	}
	room := magicBowChargeCapEngine - magicBowChargeCount(player, "")
	if room <= 0 {
		return 0
	}
	added := 0
	for _, c := range cards {
		if added >= room {
			break
		}
		player.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  player.ID,
			SourceID: player.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectMagicBowCharge,
		})
		added++
	}
	syncMagicBowChargeToken(player)
	return added
}

func removeMagicBowChargeByElement(player *model.Player, element model.Element) (model.Card, bool) {
	card, ok := removeFirstCoverByEffectAndElement(player, model.EffectMagicBowCharge, element)
	if !ok {
		return model.Card{}, false
	}
	syncMagicBowChargeToken(player)
	return card, true
}

func spiritCasterPowerCovers(player *model.Player) []*model.FieldCard {
	return coverCardsByEffect(player, model.EffectSpiritCasterPower)
}

func spiritCasterPowerCount(player *model.Player, element model.Element) int {
	return coverCountByEffectAndElement(player, model.EffectSpiritCasterPower, element)
}

func syncSpiritCasterPowerToken(player *model.Player) {
	// no-op: sc_power_count 在服务端 buildStateForPlayer 中按场上盖牌派生写入 PlayerView.tokens
}

func addSpiritCasterPowerCard(player *model.Player, card model.Card) bool {
	if player == nil {
		return false
	}
	player.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  player.ID,
		SourceID: player.ID,
		Mode:     model.FieldCover,
		Effect:   model.EffectSpiritCasterPower,
	})
	syncSpiritCasterPowerToken(player)
	return true
}

func removeSpiritCasterPowerByCardID(player *model.Player, cardID string) (model.Card, bool) {
	if player == nil || cardID == "" {
		return model.Card{}, false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSpiritCasterPower {
			continue
		}
		if fc.Card.ID != cardID {
			continue
		}
		card := fc.Card
		player.RemoveFieldCard(fc)
		syncSpiritCasterPowerToken(player)
		return card, true
	}
	return model.Card{}, false
}

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
			Hook: model.FieldHookManual,
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

func (e *GameEngine) moraleFloorForCamp(camp model.Camp) int {
	floor := 0
	for _, p := range e.State.Players {
		if p == nil || !e.isButterflyDancer(p) || p.Tokens == nil {
			continue
		}
		if p.Camp == camp {
			continue
		}
		if p.Tokens["bt_wither_active"] > 0 {
			if floor < 1 {
				floor = 1
			}
		}
	}
	return floor
}

func (e *GameEngine) applyCampMoraleLoss(camp model.Camp, wantLoss int) int {
	if wantLoss <= 0 {
		return 0
	}
	current := e.campMorale(camp)
	floor := e.moraleFloorForCamp(camp)
	maxLoss := current - floor
	if maxLoss < 0 {
		maxLoss = 0
	}
	actual := wantLoss
	if actual > maxLoss {
		actual = maxLoss
	}
	if actual <= 0 {
		return 0
	}
	if camp == model.RedCamp {
		e.State.RedMorale -= actual
	} else {
		e.State.BlueMorale -= actual
	}
	return actual
}

func (e *GameEngine) addCampMorale(camp model.Camp, amount int) int {
	if amount <= 0 {
		return 0
	}
	current := e.campMorale(camp)
	if current >= standardCampMoraleCapEngine {
		return 0
	}
	actual := amount
	room := standardCampMoraleCapEngine - current
	if actual > room {
		actual = room
	}
	if actual <= 0 {
		return 0
	}
	if camp == model.RedCamp {
		e.State.RedMorale += actual
	} else {
		e.State.BlueMorale += actual
	}
	return actual
}
