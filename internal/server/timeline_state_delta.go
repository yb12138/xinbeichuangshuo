package server

import (
	"encoding/json"
	"fmt"
	"sort"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/protocol"
	"starcup-engine/internal/server/stateview"
	"starcup-engine/internal/server/timeline"
)

type publicTimelineSnapshot struct {
	RedMorale    int
	BlueMorale   int
	RedCups      int
	BlueCups     int
	RedGems      int
	BlueGems     int
	RedCrystals  int
	BlueCrystals int
	DeckCount    int
	DiscardCount int
	Players      map[string]publicTimelinePlayerSnapshot
}

type publicTimelinePlayerSnapshot struct {
	ID                 string
	Name               string
	Camp               string
	Heal               int
	Gem                int
	Crystal            int
	HandCount          int
	ExclusiveCardCount int
	Form               string
	Orientation        string
	Tokens             map[string]int
	Status             map[string]int
	Indicators         map[string]int
	Field              []publicTimelineFieldSnapshot
}

type publicTimelineFieldSnapshot struct {
	Key       string
	Text      string
	FieldCard *model.FieldCard
}

func (r *Room) resetPublicTimelineSnapshot() {
	r.publicTimelineSnapshot = r.capturePublicTimelineSnapshot()
}

func (r *Room) broadcastPublicStateDelta(reason string) {
	next := r.capturePublicTimelineSnapshot()
	if next == nil {
		r.publicTimelineSnapshot = nil
		return
	}
	prev := r.publicTimelineSnapshot
	if prev == nil {
		r.publicTimelineSnapshot = next
		return
	}

	deltas := diffPublicTimelineSnapshots(prev, next, reason)
	r.publicTimelineSnapshot = next
	if len(deltas) == 0 {
		return
	}
	r.broadcastTimeline(timeline.Payload{
		Type:   "state_delta",
		Reason: reason,
		Deltas: deltas,
	})
}

func (r *Room) capturePublicTimelineSnapshot() *publicTimelineSnapshot {
	if r == nil || r.Engine == nil || r.Engine.State == nil {
		return nil
	}
	state := r.Engine.State
	snapshot := &publicTimelineSnapshot{
		RedMorale:    state.RedMorale,
		BlueMorale:   state.BlueMorale,
		RedCups:      state.RedCups,
		BlueCups:     state.BlueCups,
		RedGems:      state.RedGems,
		BlueGems:     state.BlueGems,
		RedCrystals:  state.RedCrystals,
		BlueCrystals: state.BlueCrystals,
		DeckCount:    len(state.Deck),
		DiscardCount: len(state.DiscardPile),
		Players:      map[string]publicTimelinePlayerSnapshot{},
	}
	for _, pid := range sortedPlayerIDs(state.Players) {
		p := state.Players[pid]
		if p == nil {
			continue
		}
		playerSnapshot := publicTimelinePlayerSnapshot{
			ID:                 p.ID,
			Name:               p.Name,
			Camp:               string(p.Camp),
			Heal:               p.Heal,
			Gem:                p.Gem,
			Crystal:            p.Crystal,
			HandCount:          len(p.Hand),
			ExclusiveCardCount: len(p.ExclusiveCards),
			Form:               r.Engine.GetPlayerForm(pid),
			Orientation:        string(r.Engine.GetPlayerOrientation(pid)),
			Tokens:             cloneIntMap(p.Tokens),
			Status:             cloneIntMap(p.Status),
			Indicators:         publicTimelineIndicators(state, p),
			Field:              publicTimelineField(p),
		}
		removePrivateTimelineTokens(playerSnapshot.Tokens)
		snapshot.Players[pid] = playerSnapshot
	}
	return snapshot
}

func diffPublicTimelineSnapshots(prev, next *publicTimelineSnapshot, reason string) []protocol.TimelineDelta {
	if prev == nil || next == nil {
		return nil
	}
	var deltas []protocol.TimelineDelta
	addCampDelta := func(deltaType, camp, field string, before, after int) {
		if before == after {
			return
		}
		deltas = append(deltas, protocol.TimelineDelta{
			Type:   deltaType,
			Scope:  "team",
			Camp:   camp,
			Field:  field,
			Before: before,
			After:  after,
			Value:  after - before,
			Reason: reason,
		})
	}
	addGlobalDelta := func(deltaType, field string, before, after int) {
		if before == after {
			return
		}
		deltas = append(deltas, protocol.TimelineDelta{
			Type:   deltaType,
			Scope:  "global",
			Field:  field,
			Before: before,
			After:  after,
			Value:  after - before,
			Reason: reason,
		})
	}

	addCampDelta("morale", "Red", "morale", prev.RedMorale, next.RedMorale)
	addCampDelta("morale", "Blue", "morale", prev.BlueMorale, next.BlueMorale)
	addCampDelta("team_cup", "Red", "cup", prev.RedCups, next.RedCups)
	addCampDelta("team_cup", "Blue", "cup", prev.BlueCups, next.BlueCups)
	addCampDelta("team_gem", "Red", "gem", prev.RedGems, next.RedGems)
	addCampDelta("team_gem", "Blue", "gem", prev.BlueGems, next.BlueGems)
	addCampDelta("team_crystal", "Red", "crystal", prev.RedCrystals, next.RedCrystals)
	addCampDelta("team_crystal", "Blue", "crystal", prev.BlueCrystals, next.BlueCrystals)
	addGlobalDelta("deck_count", "deck_count", prev.DeckCount, next.DeckCount)
	addGlobalDelta("discard_count", "discard_count", prev.DiscardCount, next.DiscardCount)

	for _, pid := range sortedSnapshotPlayerIDs(prev.Players, next.Players) {
		before, hadBefore := prev.Players[pid]
		after, hasAfter := next.Players[pid]
		if !hadBefore || !hasAfter {
			continue
		}
		deltas = append(deltas, diffPublicTimelinePlayer(before, after, reason)...)
	}
	return deltas
}

func diffPublicTimelinePlayer(prev, next publicTimelinePlayerSnapshot, reason string) []protocol.TimelineDelta {
	var deltas []protocol.TimelineDelta
	addPlayerDelta := func(deltaType, field string, before, after int) {
		if before == after {
			return
		}
		deltas = append(deltas, protocol.TimelineDelta{
			Type:         deltaType,
			Scope:        "player",
			TargetUserID: next.ID,
			Camp:         next.Camp,
			Field:        field,
			Before:       before,
			After:        after,
			Value:        after - before,
			Reason:       reason,
		})
	}
	addPlayerTextDelta := func(deltaType, field, before, after string) {
		if before == after {
			return
		}
		deltas = append(deltas, protocol.TimelineDelta{
			Type:         deltaType,
			Scope:        "player",
			TargetUserID: next.ID,
			Camp:         next.Camp,
			Field:        field,
			BeforeText:   before,
			AfterText:    after,
			Reason:       reason,
		})
	}

	addPlayerDelta("player_gem", "gem", prev.Gem, next.Gem)
	addPlayerDelta("player_crystal", "crystal", prev.Crystal, next.Crystal)
	addPlayerDelta("heal", "heal", prev.Heal, next.Heal)
	addPlayerDelta("hand_count", "hand_count", prev.HandCount, next.HandCount)
	addPlayerDelta("exclusive_count", "exclusive_count", prev.ExclusiveCardCount, next.ExclusiveCardCount)
	addPlayerTextDelta("form", "form", prev.Form, next.Form)
	addPlayerTextDelta("orientation", "orientation", prev.Orientation, next.Orientation)
	deltas = append(deltas, diffIntMapDeltas("token", "token", next.ID, next.Camp, prev.Tokens, next.Tokens, reason)...)
	deltas = append(deltas, diffIntMapDeltas("status", "status", next.ID, next.Camp, prev.Status, next.Status, reason)...)
	deltas = append(deltas, diffIntMapDeltas("token", "indicator", next.ID, next.Camp, prev.Indicators, next.Indicators, reason)...)
	deltas = append(deltas, diffFieldCardDeltas(next.ID, next.Camp, prev.Field, next.Field, reason)...)
	return deltas
}

func diffIntMapDeltas(deltaType, scope, playerID, camp string, prev, next map[string]int, reason string) []protocol.TimelineDelta {
	keys := map[string]struct{}{}
	for key := range prev {
		keys[key] = struct{}{}
	}
	for key := range next {
		keys[key] = struct{}{}
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)

	deltas := make([]protocol.TimelineDelta, 0)
	for _, key := range ordered {
		before := prev[key]
		after := next[key]
		if before == after {
			continue
		}
		deltas = append(deltas, protocol.TimelineDelta{
			Type:         deltaType,
			Scope:        "player",
			TargetUserID: playerID,
			Camp:         camp,
			Field:        fmt.Sprintf("%s:%s", scope, key),
			Before:       before,
			After:        after,
			Value:        after - before,
			Reason:       reason,
		})
	}
	return deltas
}

func diffFieldCardDeltas(playerID, camp string, prev, next []publicTimelineFieldSnapshot, reason string) []protocol.TimelineDelta {
	prevCounts := map[string]int{}
	nextCounts := map[string]int{}
	prevCards := map[string]publicTimelineFieldSnapshot{}
	nextCards := map[string]publicTimelineFieldSnapshot{}
	for _, item := range prev {
		key := item.Key
		prevCounts[key]++
		if _, ok := prevCards[key]; !ok {
			prevCards[key] = item
		}
	}
	for _, item := range next {
		key := item.Key
		nextCounts[key]++
		if _, ok := nextCards[key]; !ok {
			nextCards[key] = item
		}
	}
	keys := map[string]struct{}{}
	for key := range prevCounts {
		keys[key] = struct{}{}
	}
	for key := range nextCounts {
		keys[key] = struct{}{}
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)

	var deltas []protocol.TimelineDelta
	for _, key := range ordered {
		beforeCount := prevCounts[key]
		afterCount := nextCounts[key]
		if beforeCount == afterCount {
			continue
		}
		if afterCount > beforeCount {
			card := nextCards[key]
			for i := 0; i < afterCount-beforeCount; i++ {
				deltas = append(deltas, protocol.TimelineDelta{
					Type:         "field_card_added",
					Scope:        "player",
					TargetUserID: playerID,
					Camp:         camp,
					Field:        "field",
					Value:        1,
					Reason:       reason,
					AfterText:    card.Text,
					FieldCard:    cloneFieldCardPtr(card.FieldCard),
				})
			}
		} else {
			card := prevCards[key]
			for i := 0; i < beforeCount-afterCount; i++ {
				deltas = append(deltas, protocol.TimelineDelta{
					Type:         "field_card_removed",
					Scope:        "player",
					TargetUserID: playerID,
					Camp:         camp,
					Field:        "field",
					Value:        -1,
					Reason:       reason,
					BeforeText:   card.Text,
					FieldCard:    cloneFieldCardPtr(card.FieldCard),
				})
			}
		}
	}
	return deltas
}

func publicTimelineIndicators(state *model.GameState, p *model.Player) map[string]int {
	indicators := map[string]int{}
	set := func(key string, value int) {
		if value > 0 {
			indicators[key] = value
		}
	}
	set("elf_blessing_count", stateview.CountElfBlessings(p))
	set("mb_charge_count", stateview.CountMagicBowCharges(p))
	set("sc_power_count", stateview.CountSpiritCasterPowers(p))
	set("mg_dark_moon_count", stateview.CountMoonDarkMoons(p))
	set("bt_cocoon_count", stateview.CountButterflyCocoons(p))
	set("bp_shared_life_active", stateview.CountBloodSharedLifeAsSource(state, p.ID))
	set("bp_shared_life_bound", stateview.CountBloodSharedLifeAsHolder(p))
	set("se_sword_soul_count", stateview.CountSwordEmperorSwordSouls(p))
	set("ml_dark_release_next_attack_bonus", stateview.CombatPolicyAttackBonusByModifierID(p, "ml_dark_release_next_attack_bonus"))
	if stateview.HasRuleModifierWithModifierID(p, "ml_dark_release_lock_turn") {
		indicators["ml_dark_release_lock_turn"] = 1
	}
	return indicators
}

func publicTimelineField(p *model.Player) []publicTimelineFieldSnapshot {
	if p == nil || len(p.Field) == 0 {
		return nil
	}
	out := make([]publicTimelineFieldSnapshot, 0, len(p.Field))
	for index, fc := range p.Field {
		if fc == nil {
			continue
		}
		card := maskPublicTimelineFieldCard(fc)
		text := publicTimelineFieldText(card)
		out = append(out, publicTimelineFieldSnapshot{
			Key:       fmt.Sprintf("%s|%02d|%s", p.ID, index, publicTimelineFieldKey(card)),
			Text:      text,
			FieldCard: card,
		})
	}
	return out
}

func maskPublicTimelineFieldCard(fc *model.FieldCard) *model.FieldCard {
	clone := cloneFieldCardPtr(fc)
	if clone == nil || clone.Mode != model.FieldCover {
		return clone
	}
	clone.Card = model.Card{
		Name:        publicTimelineCoverName(clone.Effect),
		Description: "盖牌（内容对他人不可见）",
	}
	return clone
}

func publicTimelineCoverName(effect model.EffectType) string {
	switch effect {
	case model.EffectMagicBowCharge:
		return "充能"
	case model.EffectSpiritCasterPower:
		return "妖力"
	case model.EffectMoonDarkMoon:
		return "暗月"
	case model.EffectButterflyCocoon:
		return "茧"
	case model.EffectElfBlessing:
		return "祝福"
	case model.EffectSwordSoul:
		return "剑魂"
	default:
		return "盖牌"
	}
}

func publicTimelineFieldText(fc *model.FieldCard) string {
	if fc == nil {
		return "场上牌"
	}
	if fc.Card.Name != "" {
		return fc.Card.Name
	}
	if fc.Effect != "" {
		return string(fc.Effect)
	}
	if fc.Mode != "" {
		return string(fc.Mode)
	}
	return "场上牌"
}

func publicTimelineFieldKey(fc *model.FieldCard) string {
	if fc == nil {
		return ""
	}
	raw, err := json.Marshal(fc)
	if err != nil {
		return publicTimelineFieldText(fc)
	}
	return string(raw)
}

func cloneFieldCardPtr(fc *model.FieldCard) *model.FieldCard {
	if fc == nil {
		return nil
	}
	clone := *fc
	if fc.Meta != nil {
		clone.Meta = make(map[string]string, len(fc.Meta))
		for k, v := range fc.Meta {
			clone.Meta[k] = v
		}
	}
	return &clone
}

func cloneIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func removePrivateTimelineTokens(tokens map[string]int) {
	for _, key := range []string{
		"elf_blessing_count",
		"mb_charge_count",
		"sc_power_count",
		"mg_dark_moon_count",
		"bt_cocoon_count",
		"bp_shared_life_active",
		"bp_shared_life_bound",
		"css_blood_cap",
		"ml_dark_release_next_attack_bonus",
		"ml_dark_release_lock_turn",
		"se_sword_soul_count",
	} {
		delete(tokens, key)
	}
}

func sortedPlayerIDs(players map[string]*model.Player) []string {
	ids := make([]string, 0, len(players))
	for id := range players {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func sortedSnapshotPlayerIDs(maps ...map[string]publicTimelinePlayerSnapshot) []string {
	seen := map[string]struct{}{}
	for _, m := range maps {
		for id := range m {
			seen[id] = struct{}{}
		}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
