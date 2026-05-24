package server

import (
	"encoding/json"
	"strings"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/prompting"
	"starcup-engine/internal/server/timeline"
)

func TestTranslateClientAction_AttackUsesCardIDAndTargets(t *testing.T) {
	room := NewRoom("PROTO")
	room.Engine = engine.NewGameEngine(room)

	if err := room.Engine.AddPlayer("p1", "Alice", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Bob", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	room.Engine.State.Players["p1"].Hand = []model.Card{
		{ID: "card-001", Name: "测试火攻", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}

	req := ClientActionRequest{
		ActionType: model.CmdAttack,
		CardID:     "card-001",
		Targets: []TargetNode{
			{TargetUserID: "p2"},
		},
	}

	got, err := room.translateClientAction("p1", req)
	if err != nil {
		t.Fatalf("translateClientAction error: %v", err)
	}
	if got.Type != model.CmdAttack {
		t.Fatalf("expected action type %s, got %s", model.CmdAttack, got.Type)
	}
	if got.CardID != "card-001" {
		t.Fatalf("expected card id card-001, got %q", got.CardID)
	}
	if got.TargetID != "p2" {
		t.Fatalf("expected target p2, got %q", got.TargetID)
	}
	if len(got.TargetIDs) != 1 || got.TargetIDs[0] != "p2" {
		t.Fatalf("expected target IDs [p2], got %+v", got.TargetIDs)
	}
}

func TestTranslateClientAction_SelectCardIDsStayAsCardIDs(t *testing.T) {
	room := NewRoom("PROTO_SELECT")
	room.Engine = engine.NewGameEngine(room)

	if err := room.Engine.AddPlayer("p1", "Alice", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	room.Engine.State.Players["p1"].Hand = []model.Card{
		{ID: "card-a", Name: "A", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "card-b", Name: "B", Type: model.CardTypeMagic, Element: model.ElementWater},
		{ID: "card-c", Name: "C", Type: model.CardTypeAttack, Element: model.ElementWind},
	}

	got, err := room.translateClientAction("p1", ClientActionRequest{
		ActionType: model.CmdSelect,
		CardIDs:    []string{"card-c", "card-a"},
	})
	if err != nil {
		t.Fatalf("translateClientAction error: %v", err)
	}
	if got.CardID != "card-c" {
		t.Fatalf("expected first card id card-c, got %q", got.CardID)
	}
	if len(got.CardIDs) != 2 || got.CardIDs[0] != "card-c" || got.CardIDs[1] != "card-a" {
		t.Fatalf("expected card ids preserved, got %+v", got.CardIDs)
	}
	if len(got.Selections) != 0 {
		t.Fatalf("expected protocol adapter not to translate card ids to selections, got %+v", got.Selections)
	}
}

func TestHandleAction_NotStartedUsesNotifyTimelineEnvelope(t *testing.T) {
	room := NewRoom("PROTO_ERR")
	client := &Client{
		Room:     room,
		Send:     make(chan []byte, 1),
		PlayerID: "p1",
		Name:     "Alice",
	}

	room.handleAction(client, mustMarshal(ClientActionRequest{ActionType: model.CmdPass}))

	raw := <-client.Send
	var msg WSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal ws message: %v", err)
	}
	if msg.Cmd != CmdNotifyTimeline {
		t.Fatalf("expected cmd %s, got %s", CmdNotifyTimeline, msg.Cmd)
	}

	var payload TimelineNotifyPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if len(payload.Events) != 1 {
		t.Fatalf("expected 1 timeline event, got %+v", payload.Events)
	}
	event := payload.Events[0]
	if event.GameplayType != "error" {
		t.Fatalf("expected gameplay_type=error, got %+v", event)
	}
	if event.Message != "游戏尚未开始" {
		t.Fatalf("expected message 游戏尚未开始, got %q", event.Message)
	}
	if event.Damage != 0 {
		t.Fatalf("expected no damage on error event, got %+v", event)
	}
}

func TestBuildTimelineNotify_DamageEvent(t *testing.T) {
	room := NewRoom("TIMELINE")
	room.Engine = engine.NewGameEngine(room)
	room.Engine.State.CurrentTurn = 2
	room.Engine.State.CombatStage = model.CombatStageApply
	room.Engine.State.CombatStack = []model.CombatRequest{{AttackerID: "p1", TargetID: "p2"}}

	payload := room.buildTimelineNotify(timeline.Payload{
		Type:       "damage_dealt",
		SourceID:   "p1",
		SourceName: "Alice",
		TargetID:   "p2",
		TargetName: "Bob",
		Damage:     3,
		DamageType: "Attack",
		ActionType: "attack",
		Cards: []model.Card{
			{ID: "card-001", Name: "烈焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 3},
		},
		Message: "造成3点伤害",
	})

	if payload.RoomID != "TIMELINE" {
		t.Fatalf("expected room id TIMELINE, got %q", payload.RoomID)
	}
	if payload.SeqStart != 1 || payload.SeqEnd != 1 {
		t.Fatalf("expected seq 1..1, got %d..%d", payload.SeqStart, payload.SeqEnd)
	}
	if len(payload.Events) != 1 {
		t.Fatalf("expected 1 timeline event, got %d", len(payload.Events))
	}

	event := payload.Events[0]
	if event.TurnID != 3 {
		t.Fatalf("expected turn id 3, got %d", event.TurnID)
	}
	if event.CombatStage != string(model.CombatStageApply) {
		t.Fatalf("expected combat stage %s, got %s", model.CombatStageApply, event.CombatStage)
	}
	if event.Type != "TimelineCombatResolved" {
		t.Fatalf("expected type TimelineCombatResolved, got %s", event.Type)
	}
	if event.ActorUserID != "p1" || event.ActorName != "Alice" {
		t.Fatalf("expected actor p1/Alice, got %+v", event)
	}
	if len(event.TargetUserIDs) != 1 || event.TargetUserIDs[0] != "p2" || event.TargetName != "Bob" {
		t.Fatalf("expected target p2/Bob, got %+v", event)
	}
	if len(event.CardIDs) != 1 || event.CardIDs[0] != "card-001" {
		t.Fatalf("expected card ids [card-001], got %+v", event.CardIDs)
	}
	if len(event.Cards) != 1 || event.Cards[0].Name != "烈焰斩" {
		t.Fatalf("expected cards payload, got %+v", event.Cards)
	}
	if event.Damage != 3 || event.DamageType != "Attack" {
		t.Fatalf("expected damage payload, got %+v", event)
	}
	if len(event.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %+v", event.Deltas)
	}
	if event.Deltas[0].Type != "TimelineDeltaDamage" || event.Deltas[0].TargetUserID != "p2" || event.Deltas[0].Value != 3 {
		t.Fatalf("unexpected delta %+v", event.Deltas[0])
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal timeline payload: %v", err)
	}
	text := string(raw)
	if strings.Contains(text, "legacy_payload") {
		t.Fatalf("unexpected legacy_payload field in timeline json: %s", text)
	}
}

func TestBuildTimelineNotify_SelfDamageKeepsTarget(t *testing.T) {
	room := NewRoom("TIMELINE_SELF_DAMAGE")
	room.Engine = engine.NewGameEngine(room)

	payload := room.buildTimelineNotify(timeline.Payload{
		Type:       "damage_dealt",
		SourceID:   "p1",
		SourceName: "Alice",
		TargetID:   "p1",
		TargetName: "Alice",
		Damage:     3,
		DamageType: "magic",
	})

	if len(payload.Events) != 1 {
		t.Fatalf("expected 1 timeline event, got %d", len(payload.Events))
	}
	event := payload.Events[0]
	if event.ActorUserID != "p1" {
		t.Fatalf("expected actor p1, got %+v", event)
	}
	if len(event.TargetUserIDs) != 1 || event.TargetUserIDs[0] != "p1" {
		t.Fatalf("expected self-damage target p1 to be preserved, got %+v", event)
	}
	if len(event.Deltas) != 1 || event.Deltas[0].TargetUserID != "p1" || event.Deltas[0].Value != 3 {
		t.Fatalf("expected self-damage delta on p1, got %+v", event.Deltas)
	}
}

func TestBuildTimelineNotify_StructuredGameplayEvents(t *testing.T) {
	room := NewRoom("TIMELINE_STRUCTURED")
	room.Engine = engine.NewGameEngine(room)

	skillPayload := room.buildTimelineNotify(timeline.Payload{
		Type:       "skill_activated",
		PlayerID:   "p1",
		PlayerName: "Alice",
		SkillID:    "sage_arcane_codex",
		SkillName:  "苍炎法典",
		EffectText: "造成法术伤害",
		TargetIDs:  []string{"p2"},
	})
	skillEvent := skillPayload.Events[0]
	if skillEvent.GameplayType != "skill_activated" || skillEvent.SkillName != "苍炎法典" || skillEvent.EffectText != "造成法术伤害" {
		t.Fatalf("unexpected skill event %+v", skillEvent)
	}
	if len(skillEvent.TargetUserIDs) != 1 || skillEvent.TargetUserIDs[0] != "p2" {
		t.Fatalf("expected skill target p2, got %+v", skillEvent.TargetUserIDs)
	}

	specialPayload := room.buildTimelineNotify(timeline.Payload{
		Type:       "special_action",
		PlayerID:   "p1",
		PlayerName: "Alice",
		ActionType: "Buy",
		Summary:    "Alice 执行特殊行动【购买】",
	})
	specialEvent := specialPayload.Events[0]
	if specialEvent.GameplayType != "special_action" || specialEvent.ActionType != "Buy" || specialEvent.Summary == "" {
		t.Fatalf("unexpected special action event %+v", specialEvent)
	}

	deltaPayload := room.buildTimelineNotify(timeline.Payload{
		Type: "state_delta",
		Deltas: []TimelineDelta{{
			Type:   "morale",
			Scope:  "team",
			Camp:   "Red",
			Field:  "morale",
			Before: 15,
			After:  14,
			Value:  -1,
			Reason: "test",
		}},
	})
	deltaEvent := deltaPayload.Events[0]
	if deltaEvent.GameplayType != "state_delta" || len(deltaEvent.Deltas) != 1 {
		t.Fatalf("unexpected state delta event %+v", deltaEvent)
	}
	if deltaEvent.Deltas[0].Type != "morale" || deltaEvent.Deltas[0].Camp != "Red" || deltaEvent.Deltas[0].Value != -1 {
		t.Fatalf("unexpected delta %+v", deltaEvent.Deltas[0])
	}
}

func TestBuildRequireActionPayload_UsesStructuredPromptField(t *testing.T) {
	prompt := &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   "p1",
		Message:    "请选择应对方式",
		ChoiceType: "assassin_stealth_draw",
		SkillID:    "water_shadow",
		Options: []model.PromptOption{
			{ID: "defend", Label: "防御", ButtonLabel: "防御", Hint: "打出圣光"},
			{ID: "counter", Label: "应战", ButtonLabel: "应战", Hint: "打出同系攻击牌"},
		},
		SpecialOptions: []model.PromptOption{
			{ID: "cancel", Label: "取消", ButtonLabel: "取消"},
		},
		EffectHints:      []string{"命中后附加 1 点伤害"},
		Min:              1,
		Max:              1,
		AttackerID:       "p2",
		CounterTargetIDs: []string{"p3"},
		AttackElement:    string(model.ElementFire),
		Presentation:     &model.PromptPresentation{Kind: model.PresentationResponse, Layout: "inline"},
	}

	payload := prompting.BuildRequireActionPayload(prompt)
	if payload.Prompt == nil {
		t.Fatalf("expected structured prompt payload")
	}
	if payload.Prompt.PlayerID != "p1" || payload.Prompt.AttackerID != "p2" {
		t.Fatalf("unexpected prompt payload %+v", payload.Prompt)
	}
	if payload.Prompt.ChoiceType != "assassin_stealth_draw" || payload.Prompt.SkillID != "water_shadow" {
		t.Fatalf("unexpected structured prompt metadata %+v", payload.Prompt)
	}
	if len(payload.Prompt.Options) != 2 {
		t.Fatalf("expected 2 prompt options, got %+v", payload.Prompt.Options)
	}
	if len(payload.Prompt.SpecialOptions) != 1 {
		t.Fatalf("expected 1 special option, got %+v", payload.Prompt.SpecialOptions)
	}
	if len(payload.Prompt.EffectHints) != 1 || payload.Prompt.EffectHints[0] != "命中后附加 1 点伤害" {
		t.Fatalf("unexpected effect hints %+v", payload.Prompt.EffectHints)
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal require action payload: %v", err)
	}
	text := string(raw)
	if !strings.Contains(text, `"prompt":`) {
		t.Fatalf("expected prompt field in payload json, got %s", text)
	}
	if strings.Contains(text, "legacy_prompt") {
		t.Fatalf("unexpected legacy_prompt field in payload json: %s", text)
	}
}

func TestBuildSyncStatePayload_UsesStructuredFields(t *testing.T) {
	room := NewRoom("SYNC")
	room.Engine = engine.NewGameEngine(room)
	room.Started = true

	if err := room.Engine.AddPlayer("p1", "Alice", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Bob", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	room.Engine.State.Deck = []model.Card{{ID: "deck-1"}, {ID: "deck-2"}}
	room.Engine.State.DiscardPile = []model.Card{{ID: "discard-1"}}
	room.Engine.State.Players["p1"].Hand = []model.Card{
		{ID: "hand-1", Name: "烈焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}

	payload := room.buildSyncStatePayload("p1")
	if payload.RoomState != "Playing" {
		t.Fatalf("expected room_state Playing, got %s", payload.RoomState)
	}
	if payload.DeckCount != 2 || payload.DiscardCount != 1 {
		t.Fatalf("unexpected deck/discard counts: %+v", payload)
	}
	if len(payload.Players) != 2 {
		t.Fatalf("expected 2 players, got %+v", payload.Players)
	}
	if payload.Players[0].ID != "p1" || payload.Players[0].Name != "Alice" {
		t.Fatalf("unexpected first player payload %+v", payload.Players[0])
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal sync payload: %v", err)
	}
	text := string(raw)
	if !strings.Contains(text, `"deck_count":2`) {
		t.Fatalf("expected deck_count in payload json, got %s", text)
	}
	if strings.Contains(text, "legacy_state") {
		t.Fatalf("unexpected legacy_state field in payload json: %s", text)
	}
}

func TestBuildSyncStatePayload_DerivesTurnPlayerFromCurrentTurn(t *testing.T) {
	room := NewRoom("SYNC_TURN")
	room.Engine = engine.NewGameEngine(room)
	room.Started = true

	if err := room.Engine.AddPlayer("p1", "Alice", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Bob", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	room.Engine.State.CurrentTurn = 1
	room.Engine.State.CurrentPlayer = "p1" // Stale legacy field; CurrentTurn/PlayerOrder is authoritative.

	payload := room.buildSyncStatePayload("p1")
	if payload.TurnPlayerID != "p2" {
		t.Fatalf("expected turn player from CurrentTurn p2, got %q", payload.TurnPlayerID)
	}
}
