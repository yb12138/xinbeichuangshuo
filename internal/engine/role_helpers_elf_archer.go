package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

func markElfBlessings(player *model.Player, cards []model.Card) {
	if player == nil || len(cards) == 0 {
		return
	}
	exists := map[string]bool{}
	for _, fc := range elfBlessingCoverCards(player) {
		c := fc.Card
		if c.ID != "" {
			exists[c.ID] = true
		}
	}
	for _, c := range cards {
		if c.ID == "" || exists[c.ID] {
			continue
		}
		player.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  player.ID,
			SourceID: player.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectElfBlessing,
			Hook:     model.FieldHookManual,
		})
		exists[c.ID] = true
	}
	syncElfBlessings(player)
}

func syncElfBlessings(player *model.Player) {
	if player == nil {
		return
	}
	blessings := elfBlessingCards(player)
	player.Blessings = blessings
	blessingIDs := map[string]bool{}
	for _, c := range blessings {
		if c.ID != "" {
			blessingIDs[c.ID] = true
		}
	}
	newZone := make([]string, 0, len(player.CharaZone)+len(blessings))
	zoneHas := map[string]bool{}
	for _, z := range player.CharaZone {
		if !strings.HasPrefix(z, elfBlessingPrefix) {
			newZone = append(newZone, z)
			zoneHas[z] = true
			continue
		}
		cardID := strings.TrimPrefix(z, elfBlessingPrefix)
		if blessingIDs[cardID] {
			newZone = append(newZone, z)
			zoneHas[z] = true
		}
	}
	for _, c := range blessings {
		if c.ID == "" {
			continue
		}
		key := elfBlessingPrefix + c.ID
		if zoneHas[key] {
			continue
		}
		newZone = append(newZone, key)
	}
	player.CharaZone = newZone
}

func countElfBlessings(player *model.Player) int {
	if player == nil {
		return 0
	}
	return len(elfBlessingCoverCards(player))
}

func isElfBlessingCard(player *model.Player, cardID string) bool {
	if player == nil || cardID == "" {
		return false
	}
	for _, c := range elfBlessingCards(player) {
		if c.ID == cardID {
			return true
		}
	}
	return false
}

func removeElfBlessingByCardID(player *model.Player, cardID string) bool {
	if player == nil || cardID == "" {
		return false
	}
	removed := false
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectElfBlessing {
			continue
		}
		if !removed && fc.Card.ID == cardID {
			player.RemoveFieldCard(fc)
			removed = true
		}
		if removed {
			break
		}
	}

	target := elfBlessingPrefix + cardID
	newZone := make([]string, 0, len(player.CharaZone))
	removedZone := false
	for _, z := range player.CharaZone {
		if !removedZone && z == target {
			removedZone = true
			continue
		}
		newZone = append(newZone, z)
	}
	player.CharaZone = newZone
	if removed {
		syncElfBlessings(player)
	}
	return removed || removedZone
}

func elfBlessingHandIndices(player *model.Player) []int {
	if player == nil {
		return nil
	}
	var idxs []int
	for i := 0; i < countElfBlessings(player); i++ {
		idxs = append(idxs, i)
	}
	return idxs
}

func elfBlessingCoverCards(player *model.Player) []*model.FieldCard {
	if player == nil {
		return nil
	}
	return player.GetCoverCardsByEffect(model.EffectElfBlessing)
}

func elfBlessingCards(player *model.Player) []model.Card {
	covers := elfBlessingCoverCards(player)
	if len(covers) == 0 {
		return nil
	}
	out := make([]model.Card, 0, len(covers))
	for _, fc := range covers {
		if fc == nil {
			continue
		}
		out = append(out, fc.Card)
	}
	return out
}

func clearElfElementalShotCombatState(player *model.Player) {
	if player == nil {
		return
	}
	if player.TurnState.SkillFlowState == nil {
		player.TurnState.SkillFlowState = make(map[string]int)
	}
	player.TurnState.SkillFlowState["elf_elemental_shot_water_pending"] = 0
	player.TurnState.SkillFlowState["elf_elemental_shot_earth_pending"] = 0
	player.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] = 0
}

func (e *GameEngine) queueElfAnimalResponse(source, target *model.Player, pd *model.PendingDamage) bool {
	if e == nil || e.dispatcher == nil || source == nil || target == nil || pd == nil {
		return false
	}
	if !e.isElfArcher(source) || !source.IsActive {
		return false
	}
	if pd.TargetID == "" || pd.TargetID == source.ID {
		return false
	}
	if !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) || pd.Card == nil || pd.IsCounter {
		return false
	}

	damageVal := pd.Damage
	ctx := e.buildContext(source, target, model.TimingOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  pd.SourceID,
		TargetID:  pd.TargetID,
		DamageVal: &damageVal,
		Card:      pd.Card,
		AttackInfo: &model.AttackEventInfo{
			ActionType: "Attack",
			IsHit:      true,
		},
	})

	skillIDs := make([]string, 0, 2)
	for _, skillID := range []string{"elf_animal_companion", "elf_pet_empower"} {
		if e.dispatcher.isSkillStillUsable(skillID, source, ctx) {
			skillIDs = append(skillIDs, skillID)
		}
	}
	if len(skillIDs) == 0 {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: source.ID,
		SkillIDs: skillIDs,
		Context:  ctx,
	})
	e.Log(fmt.Sprintf("%s 的 [动物伙伴] 响应窗口开启", source.Name))
	return true
}
