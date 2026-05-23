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

const (
	testMagicTimelineRecorderID = "test_magic_timeline_recorder"
	testMagicHitCheckRecorderID = "test_magic_hit_check_recorder"
)

var testMagicTimelineRecords = struct {
	sync.Mutex
	timings []model.Timing
}{}

type magicTimelineRecorder struct{}

func (magicTimelineRecorder) CanUse(ctx *model.Context) bool {
	return ctx != nil &&
		ctx.EventCtx != nil &&
		ctx.EventCtx.ActionType == model.ActionMagic &&
		ctx.Selections["magic_timeline"] == true
}

func (magicTimelineRecorder) Execute(ctx *model.Context) error {
	testMagicTimelineRecords.Lock()
	defer testMagicTimelineRecords.Unlock()
	testMagicTimelineRecords.timings = append(testMagicTimelineRecords.timings, ctx.Timing)
	return nil
}

func init() {
	skillhandlers.Register(testMagicTimelineRecorderID, magicTimelineRecorder{})
	skillhandlers.Register(testMagicHitCheckRecorderID, magicTimelineRecorder{})
}

func resetMagicTimelineRecords() {
	testMagicTimelineRecords.Lock()
	defer testMagicTimelineRecords.Unlock()
	testMagicTimelineRecords.timings = nil
}

func readMagicTimelineRecords() []model.Timing {
	testMagicTimelineRecords.Lock()
	defer testMagicTimelineRecords.Unlock()
	return append([]model.Timing(nil), testMagicTimelineRecords.timings...)
}

func TestMagicActionDispatchesRulebookTimelineWithoutProtocolChanges(t *testing.T) {
	resetMagicTimelineRecords()

	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Caster", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Character.Skills = append(p1.Character.Skills, model.SkillDefinition{
		ID:           testMagicTimelineRecorderID,
		Title:        "测试法术时间轴记录",
		Type:         model.SkillTypePassive,
		ResponseType: model.ResponseSilent,
		LogicHandler: testMagicTimelineRecorderID,
		Timings: []model.FlowTiming{
			model.TimingMagicDeclare,
			model.TimingMagicSelectTarget,
			model.TimingMagicValidate,
			model.TimingMagicResolve,
		},
	})
	p1.Hand = []model.Card{
		{ID: "poison-1", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementDark},
	}

	if err := game.PerformMagicByID("p1", "p2", testutils.PlayableCardID(t, game, "p1", 0)); err != nil {
		t.Fatalf("magic action failed: %v", err)
	}

	want := []model.Timing{
		model.TimingMagicDeclare,
		model.TimingMagicSelectTarget,
		model.TimingMagicValidate,
		model.TimingMagicResolve,
	}
	if got := readMagicTimelineRecords(); !reflect.DeepEqual(got, want) {
		t.Fatalf("magic timeline = %v, want %v", got, want)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("rulebook magic timeline should not change prompt/interrupt protocol, got %+v", game.State.PendingInterrupt)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("TurnStage changed to %q", game.State.TurnStage)
	}
	if !p2.HasFieldEffect(model.EffectPoison) {
		t.Fatalf("poison magic effect should still resolve")
	}
}

func TestMagicActionDoesNotEnterAttackHitCheckTimeline(t *testing.T) {
	resetMagicTimelineRecords()

	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Caster", "berserker", model.RedCamp); err != nil {
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
		ID:           testMagicHitCheckRecorderID,
		Title:        "测试命中判定不触发",
		Type:         model.SkillTypePassive,
		ResponseType: model.ResponseSilent,
		LogicHandler: testMagicHitCheckRecorderID,
		Timings:      []model.FlowTiming{model.TimingAttackHit},
	})
	p1.Hand = []model.Card{
		{ID: "weak-1", Name: "虚弱", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	if err := game.PerformMagicByID("p1", "p2", testutils.PlayableCardID(t, game, "p1", 0)); err != nil {
		t.Fatalf("magic action failed: %v", err)
	}

	if got := readMagicTimelineRecords(); len(got) != 0 {
		t.Fatalf("magic action should not dispatch hit-check timing, got %v", got)
	}
}
