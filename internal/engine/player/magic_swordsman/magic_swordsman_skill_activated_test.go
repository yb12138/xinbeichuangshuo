package magic_swordsman_test

import (
	"reflect"
	"testing"

	"starcup-engine/internal/engine"
	playerpkg "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func TestMagicSwordsmanShadowPower_EmitsSkillActivatedOnBonusDamage(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "MS", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.Players["p2"].TurnState = model.NewPlayerTurnState()
	playerpkg.SetForm(p1, model.FormMagicSwordsmanShadow)

	attackCard := model.Card{ID: "ms-hit", Name: "暗灭", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 2}
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.PendingDamageQueue = []model.PendingDamage{
		{SourceID: "p1", TargetID: "p2", Damage: 2, DamageType: model.AttackDamage, Card: &attackCard},
	}
	game.ProcessPendingDamages()

	var events []*model.SkillActivatedPayload
	for _, ev := range obs.Events {
		if ev.Type == model.EventSkillActivated && ev.SkillActivated != nil && ev.SkillActivated.SkillID == "ms_shadow_power" {
			events = append(events, ev.SkillActivated)
		}
	}
	if len(events) != 1 {
		t.Fatalf("expected exactly one ms_shadow_power skill_activated, got %d", len(events))
	}
	if events[0].PlayerID != "p1" || !reflect.DeepEqual(events[0].TargetIDs, []string{"p2"}) {
		t.Fatalf("unexpected payload: %+v", events[0])
	}
}
