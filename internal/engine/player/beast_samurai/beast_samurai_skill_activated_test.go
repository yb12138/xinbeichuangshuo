package beast_samurai_test

import (
	"reflect"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestBeastSamurai_BeastSoulWillEmitsSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game, _, _ := newBeastSamuraiTestEngine(t, obs, "")
	attackCard := model.Card{ID: "bs-hit", Name: "试斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "武", Damage: 2}
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.PendingDamageQueue = []model.PendingDamage{
		{SourceID: "p1", TargetID: "p2", Damage: 2, DamageType: model.AttackDamage, Card: &attackCard},
	}

	if paused := game.ProcessPendingDamages(); paused {
		t.Fatalf("expected hit to resolve without extra interrupt")
	}

	var events []*model.SkillActivatedPayload
	for _, ev := range obs.Events {
		if ev.Type == model.EventSkillActivated && ev.SkillActivated != nil && ev.SkillActivated.SkillID == "bs_beast_soul_will" {
			events = append(events, ev.SkillActivated)
		}
	}
	if len(events) != 1 {
		t.Fatalf("expected exactly one bs_beast_soul_will skill_activated, got %d", len(events))
	}
	if events[0].PlayerID != "p1" || !reflect.DeepEqual(events[0].TargetIDs, []string{"p1"}) {
		t.Fatalf("unexpected payload: %+v", events[0])
	}
}
