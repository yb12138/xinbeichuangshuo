package engine_test

import (
	"testing"

	"starcup-engine/internal/engine"
	skillhandlers "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

const testNarrativeExtraActionSkillID = "test_narrative_extra_action"

type narrativeExtraActionHandler struct{}

func (narrativeExtraActionHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil
}

func (narrativeExtraActionHandler) Execute(ctx *model.Context) error {
	model.AppendAttackAction(ctx.User, "测试额外行动")
	return nil
}

func init() {
	skillhandlers.Register(testNarrativeExtraActionSkillID, narrativeExtraActionHandler{})
}

func newNarrativeTraceGame(observer model.GameObserver) *engine.GameEngine {
	game := engine.NewGameEngine(observer)
	_ = game.AddPlayer("p1", "Alice", "berserker", model.RedCamp)
	_ = game.AddPlayer("p2", "Bob", "berserker", model.BlueCamp)
	_ = game.AddPlayer("p3", "Carol", "berserker", model.RedCamp)
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Players["p1"].IsActive = true
	game.State.Players["p1"].TurnState = model.NewPlayerTurnState()
	return game
}

func narrativeEvents(obs *testutils.CaptureObserver, eventType model.GameEventType) []model.GameEvent {
	var out []model.GameEvent
	for _, event := range obs.Events {
		if event.Type == eventType {
			out = append(out, event)
		}
	}
	return out
}

func TestNarrativeTrace_AttackFlowCarriesWindowActionAndCombatIDs(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newNarrativeTraceGame(obs)
	game.State.Players["p1"].Hand = []model.Card{
		{ID: "atk-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("attack action failed: %v", err)
	}

	markers := narrativeEvents(obs, model.EventTimelineMarker)
	if len(markers) == 0 || markers[0].Narrative == nil {
		t.Fatalf("expected action_started timeline marker with narrative trace, got %+v", markers)
	}
	startTrace := markers[0].Narrative
	if startTrace.NarrativeKind != "action_started" || startTrace.NarrativeWindowID == "" || startTrace.ActionID == "" {
		t.Fatalf("unexpected action_started trace %+v", startTrace)
	}

	cardEvents := narrativeEvents(obs, model.EventCardRevealed)
	if len(cardEvents) == 0 || cardEvents[0].Narrative == nil {
		t.Fatalf("expected card_revealed narrative trace, got %+v", cardEvents)
	}
	cardTrace := cardEvents[0].Narrative
	if cardTrace.NarrativeWindowID != startTrace.NarrativeWindowID || cardTrace.ActionID != startTrace.ActionID {
		t.Fatalf("expected card trace to inherit action ids, got card=%+v action=%+v", cardTrace, startTrace)
	}
	if cardTrace.NarrativeKind != "card_played" || cardTrace.VisualKind != "card" || cardTrace.CardRole != "attack" {
		t.Fatalf("unexpected card trace %+v", cardTrace)
	}

	cues := narrativeEvents(obs, model.EventCombatCue)
	if len(cues) == 0 || cues[0].Narrative == nil {
		t.Fatalf("expected combat cue narrative trace, got %+v", cues)
	}
	cue := cues[0]
	if cue.CombatCue.AttackerID != "p1" || cue.CombatCue.TargetID != "p2" || cue.CombatCue.Phase != "attack" {
		t.Fatalf("unexpected combat cue payload %+v", cue.CombatCue)
	}
	if cue.Narrative.NarrativeWindowID != startTrace.NarrativeWindowID || cue.Narrative.ActionID != startTrace.ActionID || cue.Narrative.CombatID == "" {
		t.Fatalf("expected combat cue to carry window/action/combat ids, got %+v", cue.Narrative)
	}
	if cue.Narrative.NarrativeKind != "combat_declared" {
		t.Fatalf("expected combat_declared narrative kind, got %+v", cue.Narrative)
	}
}

func TestNarrativeTrace_CounterResponseUsesResponderAndBounceTarget(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newNarrativeTraceGame(obs)
	p2 := game.State.Players["p2"]
	p2.Hand = []model.Card{
		{ID: "counter-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	game.State.Players["p1"].Hand = []model.Card{
		{ID: "atk-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("attack action failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		TargetID:  "p3",
		CardID:    testutils.PlayableCardID(t, game, "p2", 0),
		ExtraArgs: []string{"counter"},
	}); err != nil {
		t.Fatalf("counter response failed: %v", err)
	}

	cardEvents := narrativeEvents(obs, model.EventCardRevealed)
	if len(cardEvents) < 2 || cardEvents[1].Narrative == nil {
		t.Fatalf("expected counter card narrative trace, got %+v", cardEvents)
	}
	if cardEvents[1].CardRevealed.PlayerID != "p2" || cardEvents[1].Narrative.CardRole != "counter" {
		t.Fatalf("expected counter card from responder p2, got event=%+v trace=%+v", cardEvents[1].CardRevealed, cardEvents[1].Narrative)
	}

	var counterCue *model.GameEvent
	for i := range obs.Events {
		event := &obs.Events[i]
		if event.Type == model.EventCombatCue && event.CombatCue != nil && event.CombatCue.Phase == "counter" {
			counterCue = event
			break
		}
	}
	if counterCue == nil || counterCue.Narrative == nil {
		t.Fatalf("expected counter combat cue with narrative trace")
	}
	if counterCue.CombatCue.AttackerID != "p2" || counterCue.CombatCue.TargetID != "p3" {
		t.Fatalf("expected responder-to-bounce-target cue, got %+v", counterCue.CombatCue)
	}
	if counterCue.Narrative.NarrativeKind != "combat_response" || counterCue.Narrative.CardRole != "counter" {
		t.Fatalf("unexpected counter cue trace %+v", counterCue.Narrative)
	}
}

func TestNarrativeTrace_ExtraActionGrantedAfterSkillResolution(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newNarrativeTraceGame(obs)
	p1 := game.State.Players["p1"]
	p1.Character.Skills = append(p1.Character.Skills, model.SkillDefinition{
		ID:           testNarrativeExtraActionSkillID,
		Title:        "测试额外行动",
		Type:         model.SkillTypeAction,
		Description:  "获得额外攻击行动",
		ResponseType: model.ResponseSilent,
		LogicHandler: testNarrativeExtraActionSkillID,
	})

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  testNarrativeExtraActionSkillID,
	}); err != nil {
		t.Fatalf("skill action failed: %v", err)
	}

	var declared, resolved, extra *model.NarrativeTracePayload
	for i := range obs.Events {
		event := obs.Events[i]
		if event.Narrative == nil {
			continue
		}
		switch event.Narrative.NarrativeKind {
		case "skill_declared":
			declared = event.Narrative
		case "skill_resolved":
			resolved = event.Narrative
		case "extra_action_granted":
			extra = event.Narrative
		}
	}
	if declared == nil || resolved == nil || extra == nil {
		t.Fatalf("expected skill_declared, skill_resolved and extra_action_granted traces, got %+v", obs.Events)
	}
	if declared.ActionID == "" || resolved.ActionID != declared.ActionID || extra.ActionID != declared.ActionID {
		t.Fatalf("expected skill chain action id to stay stable, declared=%+v resolved=%+v extra=%+v", declared, resolved, extra)
	}
	if extra.VisualKind != "action_marker" || extra.ExtraActionType != "Attack" {
		t.Fatalf("unexpected extra action trace %+v", extra)
	}
}

func TestNarrativeTrace_PublicDiscardIsVisibleButHiddenDiscardIsNot(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newNarrativeTraceGame(obs)

	game.NotifyCardRevealed("p1", []model.Card{
		{ID: "discard-1", Name: "公开弃牌", Type: model.CardTypeMagic, Element: model.ElementWater},
	}, "discard")
	game.NotifyCardHidden("p1", []model.Card{
		{ID: "hidden-1", Name: "隐藏弃牌", Type: model.CardTypeMagic, Element: model.ElementWater},
	}, "discard")

	cardEvents := narrativeEvents(obs, model.EventCardRevealed)
	if len(cardEvents) != 2 || cardEvents[0].Narrative == nil || cardEvents[1].Narrative == nil {
		t.Fatalf("expected two discard card events with narrative traces, got %+v", cardEvents)
	}
	if cardEvents[0].Narrative.CardRole != "discard" || cardEvents[0].Narrative.VisualKind != "card" {
		t.Fatalf("public discard should be visible as a narrative card, got %+v", cardEvents[0].Narrative)
	}
	if cardEvents[1].Narrative.CardRole != "discard" || cardEvents[1].Narrative.VisualKind != "none" {
		t.Fatalf("hidden discard should not reveal a narrative card, got %+v", cardEvents[1].Narrative)
	}
}

func TestNarrativeTrace_FailedSkillDoesNotPublishDeclaredMarker(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newNarrativeTraceGame(obs)

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "missing_skill",
	}); err == nil {
		t.Fatalf("expected missing skill action to fail")
	}

	for _, event := range obs.Events {
		if event.Type != model.EventTimelineMarker || event.Narrative == nil {
			continue
		}
		if event.Narrative.NarrativeKind == "skill_declared" {
			t.Fatalf("failed skill should not publish skill_declared marker, got %+v", event)
		}
	}
}
