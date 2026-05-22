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

const testSettlementTimelineRecorderID = "test_settlement_timeline_recorder"

var testSettlementTimelineRecords = struct {
	sync.Mutex
	timings []model.Timing
}{}

type settlementTimelineRecorder struct{}

func (settlementTimelineRecorder) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.EventCtx == nil || ctx.Selections["settlement_timeline"] != true {
		return false
	}
	switch ctx.Timing {
	case model.TimingDamageSourceDeal:
		return ctx.User != nil && ctx.User.ID == ctx.EventCtx.SourceID
	default:
		return ctx.User != nil && ctx.User.ID == ctx.EventCtx.TargetID
	}
}

func (settlementTimelineRecorder) Execute(ctx *model.Context) error {
	testSettlementTimelineRecords.Lock()
	defer testSettlementTimelineRecords.Unlock()
	testSettlementTimelineRecords.timings = append(testSettlementTimelineRecords.timings, ctx.Timing)
	return nil
}

func init() {
	skillhandlers.Register(testSettlementTimelineRecorderID, settlementTimelineRecorder{})
}

func resetSettlementTimelineRecords() {
	testSettlementTimelineRecords.Lock()
	defer testSettlementTimelineRecords.Unlock()
	testSettlementTimelineRecords.timings = nil
}

func readSettlementTimelineRecords() []model.Timing {
	testSettlementTimelineRecords.Lock()
	defer testSettlementTimelineRecords.Unlock()
	return append([]model.Timing(nil), testSettlementTimelineRecords.timings...)
}

func TestDamageSettlementTimelineIncludesHealAndDrawStages(t *testing.T) {
	resetSettlementTimelineRecords()

	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Source", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	game.State.Deck = rules.InitDeck()
	source := game.State.Players["p1"]
	target := game.State.Players["p2"]
	source.Character.Skills = append(source.Character.Skills, settlementRecorderSkill())
	target.Character.Skills = append(target.Character.Skills, settlementRecorderSkill())
	target.Heal = 1

	game.AddPendingDamage(model.PendingDamage{
		SourceID:   "p1",
		TargetID:   "p2",
		Damage:     2,
		DamageType: model.MagicDamage,
	})
	if !game.ProcessPendingDamages() {
		t.Fatalf("expected heal choice to pause damage settlement")
	}
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending heal interrupt")
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("heal choice failed: %v", err)
	}

	want := []model.Timing{
		model.TimingDamageSourceDeal,
		model.TimingDamageTargetBefore,
		model.TimingDamageTaken,
		model.TimingHealBefore,
		model.TimingHealCap,
		model.TimingHealUse,
		model.TimingDamageApplied,
		model.TimingSettleDraw,
		model.TimingDamageResolved,
	}
	if got := readSettlementTimelineRecords(); !reflect.DeepEqual(got, want) {
		t.Fatalf("settlement timeline = %v, want %v", got, want)
	}
	if target.Heal != 0 {
		t.Fatalf("expected heal to be consumed, got %d", target.Heal)
	}
}

func settlementRecorderSkill() model.SkillDefinition {
	return model.SkillDefinition{
		ID:           testSettlementTimelineRecorderID,
		Title:        "测试结算时间轴记录",
		Type:         model.SkillTypePassive,
		ResponseType: model.ResponseSilent,
		LogicHandler: testSettlementTimelineRecorderID,
		Timings: []model.FlowTiming{
			model.TimingDamageSourceDeal,
			model.TimingDamageTargetBefore,
			model.TimingDamageTaken,
			model.TimingHealBefore,
			model.TimingHealCap,
			model.TimingHealUse,
			model.TimingDamageApplied,
			model.TimingSettleDraw,
			model.TimingDamageResolved,
		},
	}
}
