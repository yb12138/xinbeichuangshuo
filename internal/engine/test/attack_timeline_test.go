package engine_test

import (
	"reflect"
	"sync"
	"testing"

	"starcup-engine/internal/engine"
	skillhandlers "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

const testAttackTimelineRecorderID = "test_attack_timeline_recorder"

type attackTimelineRecord struct {
	Timing model.Timing
	Kind   model.AttackKind
}

var testAttackTimelineRecords = struct {
	sync.Mutex
	records []attackTimelineRecord
}{}

type attackTimelineRecorder struct{}

func (attackTimelineRecorder) CanUse(ctx *model.Context) bool {
	return ctx != nil &&
		ctx.EventCtx != nil &&
		ctx.EventCtx.ActionType == model.ActionAttack &&
		ctx.Selections["attack_timeline"] == true
}

func (attackTimelineRecorder) Execute(ctx *model.Context) error {
	kind, _ := ctx.Selections["attack_kind"].(model.AttackKind)
	testAttackTimelineRecords.Lock()
	defer testAttackTimelineRecords.Unlock()
	testAttackTimelineRecords.records = append(testAttackTimelineRecords.records, attackTimelineRecord{
		Timing: ctx.Timing,
		Kind:   kind,
	})
	return nil
}

func init() {
	skillhandlers.Register(testAttackTimelineRecorderID, attackTimelineRecorder{})
}

func resetAttackTimelineRecords() {
	testAttackTimelineRecords.Lock()
	defer testAttackTimelineRecords.Unlock()
	testAttackTimelineRecords.records = nil
}

func readAttackTimelineRecords() []attackTimelineRecord {
	testAttackTimelineRecords.Lock()
	defer testAttackTimelineRecords.Unlock()
	return append([]attackTimelineRecord(nil), testAttackTimelineRecords.records...)
}

func TestAttackActionDispatchesRulebookTimelineInOrder(t *testing.T) {
	resetAttackTimelineRecords()

	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Character.Skills = append(p1.Character.Skills, model.SkillDefinition{
		ID:           testAttackTimelineRecorderID,
		Title:        "测试攻击时间轴记录",
		Type:         model.SkillTypePassive,
		ResponseType: model.ResponseSilent,
		LogicHandler: testAttackTimelineRecorderID,
		Timings: []model.FlowTiming{
			model.TimingAttackDeclare,
			model.TimingAttackSelectTarget,
			model.TimingAttackPlayCard,
			model.TimingAttackModifyCard,
			model.TimingAttackCommitted,
			model.TimingAttackForceHitCheck,
			model.TimingAttackNoResponseCheck,
			model.TimingAttackResponse,
			model.TimingAttackHit,
		},
	})
	p1.Hand = []model.Card{
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
	if game.State.PendingInterrupt != nil {
		t.Fatalf("rulebook attack timeline should not change response protocol, got %+v", game.State.PendingInterrupt)
	}
	if game.State.CombatStage != model.CombatStageHitCheck {
		t.Fatalf("expected existing combat response window, got %q", game.State.CombatStage)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take response failed: %v", err)
	}

	want := []attackTimelineRecord{
		{Timing: model.TimingAttackDeclare, Kind: model.AttackKindActive},
		{Timing: model.TimingAttackSelectTarget, Kind: model.AttackKindActive},
		{Timing: model.TimingAttackPlayCard, Kind: model.AttackKindActive},
		{Timing: model.TimingAttackModifyCard, Kind: model.AttackKindActive},
		{Timing: model.TimingAttackCommitted, Kind: model.AttackKindActive},
		{Timing: model.TimingAttackForceHitCheck, Kind: model.AttackKindActive},
		{Timing: model.TimingAttackNoResponseCheck, Kind: model.AttackKindActive},
		{Timing: model.TimingAttackResponse, Kind: model.AttackKindActive},
		{Timing: model.TimingAttackHit, Kind: model.AttackKindActive},
	}
	if got := readAttackTimelineRecords(); !reflect.DeepEqual(got, want) {
		t.Fatalf("attack timeline = %+v, want %+v", got, want)
	}
}

func TestCounterAttackUsesNestedRulebookAttackKind(t *testing.T) {
	resetAttackTimelineRecords()

	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Counter", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AllyTarget", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "atk-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "counter-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Character.Skills = append(p2.Character.Skills, model.SkillDefinition{
		ID:           testAttackTimelineRecorderID,
		Title:        "测试应战时间轴记录",
		Type:         model.SkillTypePassive,
		ResponseType: model.ResponseSilent,
		LogicHandler: testAttackTimelineRecorderID,
		Timings: []model.FlowTiming{
			model.TimingAttackDeclare,
			model.TimingAttackSelectTarget,
			model.TimingAttackPlayCard,
			model.TimingAttackModifyCard,
			model.TimingAttackCommitted,
			model.TimingAttackForceHitCheck,
		},
	})

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("attack action failed: %v", err)
	}
	resetAttackTimelineRecords()

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		TargetID:  "p3",
		CardID:    testutils.PlayableCardID(t, game, "p2", 0),
		ExtraArgs: []string{"counter"},
	}); err != nil {
		t.Fatalf("counter response failed: %v", err)
	}

	got := readAttackTimelineRecords()
	if len(got) < 6 {
		t.Fatalf("counter attack timeline too short: %+v", got)
	}
	wantPrefix := []attackTimelineRecord{
		{Timing: model.TimingAttackDeclare, Kind: model.AttackKindCounter},
		{Timing: model.TimingAttackSelectTarget, Kind: model.AttackKindCounter},
		{Timing: model.TimingAttackPlayCard, Kind: model.AttackKindCounter},
		{Timing: model.TimingAttackModifyCard, Kind: model.AttackKindCounter},
		{Timing: model.TimingAttackCommitted, Kind: model.AttackKindCounter},
		{Timing: model.TimingAttackForceHitCheck, Kind: model.AttackKindCounter},
	}
	if !reflect.DeepEqual(got[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("counter attack prefix = %+v, want %+v", got[:len(wantPrefix)], wantPrefix)
	}
}
