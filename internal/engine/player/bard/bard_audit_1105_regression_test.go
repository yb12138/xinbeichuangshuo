package bard_test

import (
	"testing"

	"starcup-engine/internal/engine"
	bardpkg "starcup-engine/internal/engine/player/bard"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 吟游诗人-2：不谐和弦弃牌分支中，被弃牌的对手不能取消逃过弃牌。
func TestBardDissonanceChord_DiscarderCannotCancel(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	bard := game.State.Players["p1"]
	enemy := game.State.Players["p2"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Tokens["bd_inspiration"] = 3
	bard.Hand = []model.Card{
		bardTestCard("h1", "手牌1", model.CardTypeAttack, model.ElementFire),
		bardTestCard("h2", "手牌2", model.CardTypeAttack, model.ElementWater),
	}
	enemy.Hand = []model.Card{
		bardTestCard("e1", "敌方牌1", model.CardTypeAttack, model.ElementWater),
		bardTestCard("e2", "敌方牌2", model.CardTypeAttack, model.ElementFire),
		bardTestCard("e3", "敌方牌3", model.CardTypeAttack, model.ElementWind),
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "bd_dissonance_chord", nil, nil); err != nil {
		t.Fatalf("use dissonance failed: %v", err)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}) // X=3
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}) // 弃牌分支
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}) // 目标 p2

	testutils.RequireChoicePrompt(t, game, "p1", "bd_dissonance_discard_step")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdCancel, PlayerID: "p1"}); err == nil {
		t.Fatalf("bard discard step must not be cancelable")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1}})

	testutils.RequireChoicePrompt(t, game, "p2", "bd_dissonance_discard_step")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdCancel, PlayerID: "p2"}); err == nil {
		t.Fatalf("discarder must not be able to cancel the forced discard")
	}
	testutils.RequireChoicePrompt(t, game, "p2", "bd_dissonance_discard_step")
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0, 1}})
	if len(enemy.Hand) != 1 {
		t.Fatalf("expected enemy discard 2, hand=%d", len(enemy.Hand))
	}
}

// 吟游诗人-7：能量满或战绩区无星石时，"提炼星石"选项置灰；强行选择被拒绝。
func TestBardVictorySymphony_ExtractOptionDisabledWhenUnavailable(t *testing.T) {
	cases := []struct {
		name       string
		gems       int
		crystals   int
		holderGem  int
		wantReason string
	}{
		{name: "no_stone", gems: 0, crystals: 0, wantReason: "我方战绩区没有可提炼的星石"},
		{name: "energy_full", gems: 1, crystals: 1, holderGem: 3, wantReason: "能量已达上限，无法提炼"},
		{name: "available", gems: 1, crystals: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			game := engine.NewGameEngine(testutils.NoopObserver{})
			if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			bard := game.State.Players["p1"]
			ally := game.State.Players["p2"]
			if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
				t.Fatalf("place eternal movement failed: %v", err)
			}
			game.State.RedGems = tc.gems
			game.State.RedCrystals = tc.crystals
			ally.Gem = tc.holderGem
			ally.IsActive = true
			ally.TurnState = model.NewPlayerTurnState()
			game.State.CurrentTurn = 1
			game.State.TurnStage = model.TurnStageTurnEnd
			game.Drive()

			testutils.RequireChoicePrompt(t, game, "p2", "bd_victory_confirm")
			prompt := game.BuildChoicePrompt()
			if prompt == nil || len(prompt.Options) == 0 || prompt.Options[0].ID != "0" {
				t.Fatalf("unexpected victory prompt: %+v", prompt)
			}
			opt := prompt.Options[0]
			if tc.wantReason == "" {
				if opt.Disabled {
					t.Fatalf("extract option should be enabled, got %+v", opt)
				}
				return
			}
			if !opt.Disabled || opt.DisabledReason != tc.wantReason {
				t.Fatalf("expected extract option disabled with %q, got %+v", tc.wantReason, opt)
			}
			if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err == nil {
				t.Fatalf("selecting disabled extract option should be rejected")
			}
			testutils.RequireChoicePrompt(t, game, "p2", "bd_victory_confirm")
		})
	}
}
