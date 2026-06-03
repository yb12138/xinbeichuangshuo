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

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

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

func TestBuildTimelineNotify_UsesNarrativeWindowAsChainID(t *testing.T) {
	room := NewRoom("TIMELINE_CHAIN")
	room.Engine = engine.NewGameEngine(room)

	payload := room.buildTimelineNotify(timeline.Payload{
		Type:     "timeline_marker",
		PlayerID: "p1",
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: "nw-t1-p1",
			ActionID:          "nw-t1-p1-a1-attack",
			NarrativeKind:     "action_started",
			VisualKind:        "action_marker",
		},
	})

	if len(payload.Events) != 1 {
		t.Fatalf("expected 1 timeline event, got %d", len(payload.Events))
	}
	event := payload.Events[0]
	if event.ChainID != "nw-t1-p1" {
		t.Fatalf("expected narrative window chain id, got %+v", event)
	}
	if event.NarrativeWindowID != "nw-t1-p1" || event.ActionID != "nw-t1-p1-a1-attack" {
		t.Fatalf("expected structured narrative ids, got %+v", event)
	}
}

func TestBuildTimelineNotify_FieldDeltaCarriesNarrativeActorAndTarget(t *testing.T) {
	room := NewRoom("TIMELINE_FIELD")
	room.Engine = engine.NewGameEngine(room)

	fieldCard := &model.FieldCard{
		Card:     model.Card{ID: "seal-card", Name: "水之封印", Type: model.CardTypeMagic, Element: model.ElementWater},
		OwnerID:  "p2",
		SourceID: "p1",
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealWater,
	}
	payload := room.buildTimelineNotify(timeline.Payload{
		Type: "state_delta",
		Deltas: []TimelineDelta{{
			Type:         "field_card_added",
			Scope:        "player",
			TargetUserID: "p2",
			FieldCard:    fieldCard,
		}},
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: "nw-t1-p1",
			ActionID:          "nw-t1-p1-a1-skill",
		},
	})

	if len(payload.Events) != 1 {
		t.Fatalf("expected 1 timeline event, got %d", len(payload.Events))
	}
	event := payload.Events[0]
	if event.NarrativeKind != "field_effect_applied" || event.VisualKind != "effect_token" {
		t.Fatalf("expected field effect narrative token, got %+v", event)
	}
	if event.ActorUserID != "p1" {
		t.Fatalf("expected source actor p1, got %+v", event)
	}
	if len(event.TargetUserIDs) != 1 || event.TargetUserIDs[0] != "p2" {
		t.Fatalf("expected target p2, got %+v", event)
	}
	if event.FieldCard == nil || event.FieldCard.Card.Name != "水之封印" {
		t.Fatalf("expected field card payload, got %+v", event.FieldCard)
	}
	if event.EffectType != string(model.EffectSealWater) {
		t.Fatalf("expected effect type %s, got %q", model.EffectSealWater, event.EffectType)
	}
}

func TestBuildTimelineNotify_ActionFlowAttackCounterAndMiss(t *testing.T) {
	room := NewRoom("TIMELINE_FLOW_ATTACK")
	room.Engine = engine.NewGameEngine(room)
	_ = room.Engine.AddPlayer("p1", "Alice", "wind_sword_saint", model.BlueCamp)
	_ = room.Engine.AddPlayer("p2", "Bob", "berserker", model.RedCamp)
	_ = room.Engine.AddPlayer("p3", "Cara", "assassin", model.BlueCamp)

	trace := func(kind, visual, role, combatID string) *model.NarrativeTracePayload {
		return &model.NarrativeTracePayload{
			NarrativeWindowID: "nw-t1-p1",
			ActionID:          "nw-t1-p1-a1-attack",
			CombatID:          combatID,
			NarrativeKind:     kind,
			VisualKind:        visual,
			CardRole:          role,
		}
	}
	record := func(payload timeline.Payload) TimelineNotifyPayload {
		notify := room.buildTimelineNotify(payload)
		room.recordTimelineHistory(notify.Events)
		return notify
	}

	record(timeline.Payload{
		Type:       "timeline_marker",
		PlayerID:   "p1",
		TargetIDs:  []string{"p2"},
		ActionType: "attack",
		Trace:      trace("action_started", "action_marker", "", ""),
	})
	record(timeline.Payload{
		Type:       "combat_cue",
		AttackerID: "p1",
		TargetID:   "p2",
		Phase:      "attack",
		Trace:      trace("combat_declared", "none", "attack", "nw-t1-p1-c1-combat"),
	})
	record(timeline.Payload{
		Type:       "card_revealed",
		PlayerID:   "p1",
		ActionType: "attack",
		Cards:      []model.Card{{ID: "attack-card", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2}},
		Trace:      trace("card_played", "card", "attack", "nw-t1-p1-c1-combat"),
	})
	record(timeline.Payload{
		Type:       "combat_cue",
		AttackerID: "p2",
		TargetID:   "p3",
		Phase:      "counter",
		Trace:      trace("combat_response", "none", "counter", "nw-t1-p1-c2-counter"),
	})
	record(timeline.Payload{
		Type:       "card_revealed",
		PlayerID:   "p2",
		ActionType: "counter",
		Cards:      []model.Card{{ID: "counter-card", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 3}},
		Trace:      trace("card_played", "card", "counter", "nw-t1-p1-c2-counter"),
	})
	missTrace := trace("field_effect_applied", "effect_token", "", "nw-t1-p1-c1-combat")
	missTrace.EffectType = "attack_miss"
	notify := record(timeline.Payload{
		Type:      "timeline_marker",
		PlayerID:  "p1",
		TargetIDs: []string{"p2"},
		Summary:   "未命中",
		Trace:     missTrace,
	})

	if len(notify.ActionFlows) != 1 {
		t.Fatalf("expected one action flow, got %+v", notify.ActionFlows)
	}
	flow := notify.ActionFlows[0]
	if flow.FlowID != "nw-t1-p1-a1-attack" || flow.ActionType != "attack" {
		t.Fatalf("unexpected flow identity %+v", flow)
	}
	if len(flow.Actors) != 3 || flow.Actors[0].PlayerID != "p1" || flow.Actors[1].PlayerID != "p2" || flow.Actors[2].PlayerID != "p3" {
		t.Fatalf("expected stable actor order p1,p2,p3, got %+v", flow.Actors)
	}
	if len(flow.Edges) != 2 {
		t.Fatalf("expected attack and counter edges, got %+v", flow.Edges)
	}
	if flow.Edges[0].Phase != "attack" || len(flow.Edges[0].Cards) != 1 || flow.Edges[0].Cards[0].ID != "attack-card" {
		t.Fatalf("expected attack card on first edge, got %+v", flow.Edges[0])
	}
	if flow.Edges[0].Outcome != "miss" || flow.Edges[0].Label != "未命中" {
		t.Fatalf("expected miss attached to original attack edge, got %+v", flow.Edges[0])
	}
	if flow.Edges[1].Phase != "counter" || len(flow.Edges[1].Cards) != 1 || flow.Edges[1].Cards[0].ID != "counter-card" {
		t.Fatalf("expected counter card on second edge, got %+v", flow.Edges[1])
	}
	if len(flow.Logs) == 0 || !strings.Contains(flow.Logs[0].Text, "未命中") {
		t.Fatalf("expected backend miss log, got %+v", flow.Logs)
	}
	for _, node := range flow.Nodes {
		if node.Kind == "resolution" {
			t.Fatalf("miss should stay on edge and not create resolution node: %+v in flow %+v", node, flow)
		}
	}
}

func TestBuildTimelineNotify_ActionFlowSimpleHitDoesNotCreateRedundantNodes(t *testing.T) {
	room := NewRoom("TIMELINE_FLOW_HIT")
	room.Engine = engine.NewGameEngine(room)
	_ = room.Engine.AddPlayer("p1", "Alice", "wind_sword_saint", model.BlueCamp)
	_ = room.Engine.AddPlayer("p2", "Bob", "berserker", model.RedCamp)

	trace := func(kind, visual, role string) *model.NarrativeTracePayload {
		return &model.NarrativeTracePayload{
			NarrativeWindowID: "nw-t1-p1",
			ActionID:          "nw-t1-p1-a1-attack",
			CombatID:          "nw-t1-p1-c1-combat",
			NarrativeKind:     kind,
			VisualKind:        visual,
			CardRole:          role,
		}
	}
	record := func(payload timeline.Payload) TimelineNotifyPayload {
		notify := room.buildTimelineNotify(payload)
		room.recordTimelineHistory(notify.Events)
		return notify
	}

	record(timeline.Payload{
		Type:       "combat_cue",
		AttackerID: "p1",
		TargetID:   "p2",
		Phase:      "attack",
		Trace:      trace("combat_declared", "none", "attack"),
	})
	record(timeline.Payload{
		Type:       "card_revealed",
		PlayerID:   "p1",
		ActionType: "attack",
		TargetIDs:  []string{"p2"},
		Cards:      []model.Card{{ID: "attack-card", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2}},
		Trace:      trace("card_played", "card", "attack"),
	})
	notify := record(timeline.Payload{
		Type:       "damage_dealt",
		SourceID:   "p1",
		SourceName: "Alice",
		TargetID:   "p2",
		TargetName: "Bob",
		Damage:     2,
		DamageType: "Attack",
		Trace:      trace("damage_dealt", "damage", ""),
	})

	if len(notify.ActionFlows) != 1 {
		t.Fatalf("expected one action flow, got %+v", notify.ActionFlows)
	}
	flow := notify.ActionFlows[0]
	if len(flow.Edges) != 1 {
		t.Fatalf("expected one attack edge, got %+v", flow.Edges)
	}
	edge := flow.Edges[0]
	if edge.Outcome != "hit" || edge.Damage != 2 || edge.DamageType != "Attack" {
		t.Fatalf("expected hit damage to stay on edge, got %+v", edge)
	}
	for _, node := range flow.Nodes {
		if node.Kind == "damage" || node.Kind == "effect" || node.Kind == "resolution" {
			t.Fatalf("simple hit should not create redundant %s node: %+v in flow %+v", node.Kind, node, flow)
		}
	}
}

func TestBuildTimelineNotify_ActionFlowAttackHitSkillAnchorsToAttackEdge(t *testing.T) {
	room := NewRoom("TIMELINE_FLOW_HIT_SKILL")
	room.Engine = engine.NewGameEngine(room)
	_ = room.Engine.AddPlayer("p1", "Alice", "berserker", model.RedCamp)
	_ = room.Engine.AddPlayer("p2", "Bob", "wind_sword_saint", model.BlueCamp)

	trace := func(kind, visual, role string) *model.NarrativeTracePayload {
		return &model.NarrativeTracePayload{
			NarrativeWindowID: "nw-t1-p1",
			ActionID:          "nw-t1-p1-a1-attack",
			CombatID:          "nw-t1-p1-c1-combat",
			NarrativeKind:     kind,
			VisualKind:        visual,
			CardRole:          role,
		}
	}
	record := func(payload timeline.Payload) TimelineNotifyPayload {
		notify := room.buildTimelineNotify(payload)
		room.recordTimelineHistory(notify.Events)
		return notify
	}

	record(timeline.Payload{
		Type:       "combat_cue",
		AttackerID: "p1",
		TargetID:   "p2",
		Phase:      "attack",
		Trace:      trace("combat_declared", "none", "attack"),
	})
	record(timeline.Payload{
		Type:       "card_revealed",
		PlayerID:   "p1",
		ActionType: "attack",
		TargetIDs:  []string{"p2"},
		Cards:      []model.Card{{ID: "attack-card", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2}},
		Trace:      trace("card_played", "card", "attack"),
	})
	record(timeline.Payload{
		Type:       "damage_dealt",
		SourceID:   "p1",
		SourceName: "Alice",
		TargetID:   "p2",
		TargetName: "Bob",
		Damage:     4,
		DamageType: "Attack",
		Trace:      trace("damage_dealt", "damage", ""),
	})
	notify := record(timeline.Payload{
		Type:       "skill_activated",
		PlayerID:   "p1",
		PlayerName: "Alice",
		SkillID:    "berserker_frenzy",
		SkillName:  "狂化",
		EffectText: "攻击命中后伤害增加",
		TargetIDs:  []string{"p2"},
		Trace:      trace("skill_triggered", "skill_token", ""),
	})

	if len(notify.ActionFlows) != 1 {
		t.Fatalf("expected one action flow, got %+v", notify.ActionFlows)
	}
	flow := notify.ActionFlows[0]
	if len(flow.Edges) != 1 || flow.Edges[0].Phase != "attack" {
		t.Fatalf("expected only triggering attack edge, got %+v", flow.Edges)
	}
	attackEdgeID := flow.Edges[0].ID
	foundFrenzy := false
	frenzyNodeID := ""
	for _, node := range flow.Nodes {
		if node.Kind != "skill" || node.SkillName != "狂化" {
			continue
		}
		foundFrenzy = true
		frenzyNodeID = node.ID
		if node.AnchorEdgeID != attackEdgeID {
			t.Fatalf("expected frenzy anchored to triggering attack edge %s, got %+v", attackEdgeID, node)
		}
	}
	if !foundFrenzy {
		t.Fatalf("expected frenzy skill node, got %+v", flow.Nodes)
	}
	if len(flow.Edges[0].NodeIDs) == 0 || !stringSliceContains(flow.Edges[0].NodeIDs, frenzyNodeID) {
		t.Fatalf("expected frenzy node attached to attack edge node_ids, got %+v", flow.Edges[0])
	}
}

func TestBuildTimelineNotify_ActionFlowMagicAndInsertedSkill(t *testing.T) {
	room := NewRoom("TIMELINE_FLOW_MAGIC")
	room.Engine = engine.NewGameEngine(room)
	_ = room.Engine.AddPlayer("p1", "Alice", "sealer", model.BlueCamp)
	_ = room.Engine.AddPlayer("p2", "Bob", "berserker", model.RedCamp)
	_ = room.Engine.AddPlayer("p3", "Cara", "angel", model.BlueCamp)

	record := func(payload timeline.Payload) TimelineNotifyPayload {
		notify := room.buildTimelineNotify(payload)
		room.recordTimelineHistory(notify.Events)
		return notify
	}
	baseTrace := &model.NarrativeTracePayload{
		NarrativeWindowID: "nw-t1-p1",
		ActionID:          "nw-t1-p1-a1-magic",
	}
	record(timeline.Payload{
		Type:       "timeline_marker",
		PlayerID:   "p1",
		TargetIDs:  []string{"p2"},
		ActionType: "magic",
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: baseTrace.NarrativeWindowID,
			ActionID:          baseTrace.ActionID,
			NarrativeKind:     "action_started",
			VisualKind:        "action_marker",
		},
	})
	record(timeline.Payload{
		Type:       "card_revealed",
		PlayerID:   "p1",
		ActionType: "magic",
		Cards:      []model.Card{{ID: "magic-card", Name: "水之封印", Type: model.CardTypeMagic, Element: model.ElementWater}},
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: baseTrace.NarrativeWindowID,
			ActionID:          baseTrace.ActionID,
			NarrativeKind:     "card_played",
			VisualKind:        "card",
			CardRole:          "magic",
		},
	})
	record(timeline.Payload{
		Type:       "skill_activated",
		PlayerID:   "p3",
		PlayerName: "Cara",
		SkillID:    "angel_guard",
		SkillName:  "神圣庇护",
		EffectText: "响应法术结算",
		TargetIDs:  []string{"p1"},
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: baseTrace.NarrativeWindowID,
			ActionID:          baseTrace.ActionID,
			NarrativeKind:     "skill_triggered",
			VisualKind:        "skill_token",
			SkillPhase:        "triggered",
		},
	})
	notify := record(timeline.Payload{
		Type:       "skill_activated",
		PlayerID:   "p3",
		PlayerName: "Cara",
		SkillID:    "angel_guard",
		SkillName:  "神圣庇护",
		EffectText: "响应法术结算完成",
		TargetIDs:  []string{"p1"},
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: baseTrace.NarrativeWindowID,
			ActionID:          baseTrace.ActionID,
			NarrativeKind:     "skill_resolved",
			VisualKind:        "skill_token",
			SkillPhase:        "resolved",
		},
	})

	if len(notify.ActionFlows) != 1 {
		t.Fatalf("expected one action flow, got %+v", notify.ActionFlows)
	}
	flow := notify.ActionFlows[0]
	if flow.ActionType != "magic" {
		t.Fatalf("expected magic action flow, got %+v", flow)
	}
	if len(flow.Edges) != 1 || flow.Edges[0].Phase != "magic" || flow.Edges[0].Cards[0].ID != "magic-card" {
		t.Fatalf("expected magic card on magic edge, got %+v", flow.Edges)
	}
	if len(flow.Nodes) < 2 {
		t.Fatalf("expected card and skill nodes, got %+v", flow.Nodes)
	}
	foundSkill := false
	skillCount := 0
	for _, node := range flow.Nodes {
		if node.Kind == "skill" && node.SkillName == "神圣庇护" {
			skillCount++
			foundSkill = true
			if node.AnchorEdgeID != flow.Edges[0].ID {
				t.Fatalf("expected inserted skill anchored to magic edge, got %+v edge=%s", node, flow.Edges[0].ID)
			}
		}
	}
	if skillCount != 1 {
		t.Fatalf("expected duplicate skill traces to collapse to one node, got %d in %+v", skillCount, flow.Nodes)
	}
	if !foundSkill {
		t.Fatalf("expected inserted skill node, got %+v", flow.Nodes)
	}
}

func TestBuildTimelineNotify_ActionFlowAnchorsEarlySkillAndMergesExtraAction(t *testing.T) {
	room := NewRoom("TIMELINE_FLOW_GALE")
	room.Engine = engine.NewGameEngine(room)
	_ = room.Engine.AddPlayer("p1", "Alice", "blade_master", model.BlueCamp)
	_ = room.Engine.AddPlayer("p2", "Bob", "berserker", model.RedCamp)

	record := func(payload timeline.Payload) TimelineNotifyPayload {
		notify := room.buildTimelineNotify(payload)
		room.recordTimelineHistory(notify.Events)
		return notify
	}
	baseTrace := &model.NarrativeTracePayload{
		NarrativeWindowID: "nw-t1-p1",
		ActionID:          "nw-t1-p1-a1-attack",
	}
	record(timeline.Payload{
		Type:      "skill_activated",
		PlayerID:  "p1",
		SkillID:   "gale_skill",
		SkillName: "疾风技",
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: baseTrace.NarrativeWindowID,
			ActionID:          baseTrace.ActionID,
			NarrativeKind:     "skill_triggered",
			VisualKind:        "skill_token",
			SkillPhase:        "triggered",
		},
	})
	record(timeline.Payload{
		Type:       "timeline_marker",
		PlayerID:   "p1",
		ActionType: "Attack",
		Summary:    "疾风技",
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID:  baseTrace.NarrativeWindowID,
			ActionID:           baseTrace.ActionID,
			NarrativeKind:      "extra_action_granted",
			VisualKind:         "action_marker",
			ExtraActionType:    "Attack",
			ExtraActionElement: "",
		},
	})
	notify := record(timeline.Payload{
		Type:       "card_revealed",
		PlayerID:   "p1",
		ActionType: "attack",
		TargetIDs:  []string{"p2"},
		Cards:      []model.Card{{ID: "gale-card", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2}},
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: baseTrace.NarrativeWindowID,
			ActionID:          baseTrace.ActionID,
			NarrativeKind:     "card_played",
			VisualKind:        "card",
			CardRole:          "attack",
		},
	})

	if len(notify.ActionFlows) != 1 {
		t.Fatalf("expected one action flow, got %+v", notify.ActionFlows)
	}
	flow := notify.ActionFlows[0]
	if len(flow.Edges) != 1 || flow.Edges[0].Phase != "attack" {
		t.Fatalf("expected one attack edge, got %+v", flow.Edges)
	}
	attackEdgeID := flow.Edges[0].ID
	skillCount := 0
	extraActionCount := 0
	skillNodeID := ""
	for _, node := range flow.Nodes {
		switch node.Kind {
		case "skill":
			if node.SkillName != "疾风技" {
				continue
			}
			skillCount++
			skillNodeID = node.ID
			if node.AnchorEdgeID != attackEdgeID {
				t.Fatalf("expected gale skill anchored to attack edge %s, got %+v", attackEdgeID, node)
			}
			if !stringSliceContains(node.TargetUserIDs, "p2") {
				t.Fatalf("expected gale skill target inferred from attack edge, got %+v", node)
			}
			if !strings.Contains(node.EffectText, "额外+1攻击行动") {
				t.Fatalf("expected extra action detail merged into skill node, got %+v", node)
			}
		case "extra_action":
			extraActionCount++
		}
	}
	if skillCount != 1 {
		t.Fatalf("expected one gale skill node, got %d in %+v", skillCount, flow.Nodes)
	}
	if extraActionCount != 0 {
		t.Fatalf("expected extra action marker merged into skill node, got %+v", flow.Nodes)
	}
	if skillNodeID == "" || !stringSliceContains(flow.Edges[0].NodeIDs, skillNodeID) {
		t.Fatalf("expected skill node attached to attack edge node_ids, edge=%+v skill=%s", flow.Edges[0], skillNodeID)
	}
}

func TestBuildTimelineReplayNotify_ActionFlowsGroupedByActionID(t *testing.T) {
	room := NewRoom("TIMELINE_FLOW_REPLAY")
	room.Engine = engine.NewGameEngine(room)
	_ = room.Engine.AddPlayer("p1", "Alice", "sealer", model.BlueCamp)
	_ = room.Engine.AddPlayer("p2", "Bob", "berserker", model.RedCamp)

	record := func(payload timeline.Payload) {
		notify := room.buildTimelineNotify(payload)
		room.recordTimelineHistory(notify.Events)
	}
	record(timeline.Payload{
		Type:      "timeline_marker",
		PlayerID:  "p1",
		Summary:   "法术激荡",
		TargetIDs: []string{"p1"},
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: "nw-t1-p1",
			ActionID:          "nw-t1-p1-a1-skill",
			NarrativeKind:     "extra_action_granted",
			VisualKind:        "action_marker",
			ExtraActionType:   "Attack",
		},
	})
	record(timeline.Payload{
		Type:       "card_revealed",
		PlayerID:   "p1",
		ActionType: "attack",
		TargetIDs:  []string{"p2"},
		Cards:      []model.Card{{ID: "extra-attack", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire}},
		Trace: &model.NarrativeTracePayload{
			NarrativeWindowID: "nw-t1-p1",
			ActionID:          "nw-t1-p1-a2-attack",
			NarrativeKind:     "card_played",
			VisualKind:        "card",
			CardRole:          "attack",
		},
	})

	replay := room.buildTimelineReplayNotify()
	if replay == nil {
		t.Fatalf("expected replay payload")
	}
	if len(replay.ActionFlows) != 2 {
		t.Fatalf("expected skill flow and extra attack flow, got %+v", replay.ActionFlows)
	}
	if replay.ActionFlows[0].ActionID != "nw-t1-p1-a1-skill" || replay.ActionFlows[1].ActionID != "nw-t1-p1-a2-attack" {
		t.Fatalf("expected replay flows sorted by action id occurrence, got %+v", replay.ActionFlows)
	}
	if len(replay.ActionFlows[0].Nodes) != 1 || replay.ActionFlows[0].Nodes[0].Kind != "extra_action" {
		t.Fatalf("expected extra action marker to stay in skill flow, got %+v", replay.ActionFlows[0])
	}
	if len(replay.ActionFlows[1].Edges) != 1 || replay.ActionFlows[1].Edges[0].Cards[0].ID != "extra-attack" {
		t.Fatalf("expected extra action execution to create new attack flow, got %+v", replay.ActionFlows[1])
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
