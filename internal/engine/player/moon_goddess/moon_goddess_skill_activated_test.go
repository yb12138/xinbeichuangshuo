package moon_test

import (
	"testing"

	"starcup-engine/internal/engine"
	moonplayer "starcup-engine/internal/engine/player/moon_goddess"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestMoonGoddessSkillActivated_MoonCycleBranch1AndDarkMoonCurse(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 0
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	if !moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("expected moon cycle interrupt")
	}
	if got := obs.SkillActivated("mg_moon_cycle"); len(got) != 0 {
		t.Fatalf("moon cycle must not emit while only prompting, got %+v", got)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}})
	if got := obs.SkillActivated("mg_moon_cycle"); len(got) != 0 {
		t.Fatalf("moon cycle must not emit before target chosen, got %+v", got)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}})

	testutils.RequireSingleSkillActivation(t, obs, "mg_moon_cycle", "p1", "p2")
	testutils.RequireSingleSkillActivation(t, obs, "mg_dark_moon_curse", "p1", "p1")
}

func TestMoonGoddessSkillActivated_MoonCycleBranch2TargetsSelf(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 1 // 仅分支②
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	if !moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("expected moon cycle interrupt")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}})

	testutils.RequireSingleSkillActivation(t, obs, "mg_moon_cycle", "p1", "p1")
	if got := obs.SkillActivated("mg_dark_moon_curse"); len(got) != 0 {
		t.Fatalf("branch2 must not trigger dark moon curse, got %+v", got)
	}
}

func TestMoonGoddessSkillActivated_MoonCycleDeclineEmitsNothing(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 1
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	if !moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("expected moon cycle interrupt")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}})
	if got := obs.SkillActivated("mg_moon_cycle"); len(got) != 0 {
		t.Fatalf("declined moon cycle must not emit, got %+v", got)
	}
}

func TestMoonGoddessSkillActivated_BlasphemyTargetsDamagedEnemy(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.Heal = 1
	pd := model.PendingDamage{SourceID: "p1", TargetID: "p2", Damage: 1, DamageType: model.MagicAttack}

	if !moonplayer.TryQueueBlasphemy(engine.NewRoleChoiceRuntime(game), &pd) {
		t.Fatalf("expected blasphemy queue success")
	}
	if got := obs.SkillActivated("mg_blasphemy"); len(got) != 0 {
		t.Fatalf("blasphemy must not emit while prompting, got %+v", got)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("resolve blasphemy failed: %v", err)
	}
	testutils.RequireSingleSkillActivation(t, obs, "mg_blasphemy", "p1", "p2")
}

func TestMoonGoddessSkillActivated_BlasphemySkipEmitsNothing(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.Heal = 1
	pd := model.PendingDamage{SourceID: "p1", TargetID: "p2", Damage: 1, DamageType: model.MagicAttack}

	if !moonplayer.TryQueueBlasphemy(engine.NewRoleChoiceRuntime(game), &pd) {
		t.Fatalf("expected blasphemy queue success")
	}
	// 最后一项为“不发动”。
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("skip blasphemy failed: %v", err)
	}
	if got := obs.SkillActivated("mg_blasphemy"); len(got) != 0 {
		t.Fatalf("skipped blasphemy must not emit, got %+v", got)
	}
}
