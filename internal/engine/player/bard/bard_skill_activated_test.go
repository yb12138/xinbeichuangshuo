package bard_test

import (
	"testing"

	"starcup-engine/internal/engine"
	bardpkg "starcup-engine/internal/engine/player/bard"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func newBardFourPlayerGame(t *testing.T, obs model.GameObserver) *engine.GameEngine {
	t.Helper()
	game := engine.NewGameEngine(obs)
	for _, p := range []struct {
		id, name, role string
		camp           model.Camp
	}{
		{"p1", "Bard", "bard", model.RedCamp},
		{"p2", "Ally", "priest", model.RedCamp},
		{"p3", "EnemyA", "berserker", model.BlueCamp},
		{"p4", "EnemyB", "angel", model.BlueCamp},
	} {
		if err := game.AddPlayer(p.id, p.name, p.role, p.camp); err != nil {
			t.Fatal(err)
		}
	}
	return game
}

func TestBardDescentConcerto_EmitsSkillActivatedWithBonusTarget(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newBardFourPlayerGame(t, obs)
	bard := game.State.Players["p1"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Hand = []model.Card{
		bardTestCard("f_magic", "火法术", model.CardTypeMagic, model.ElementFire),
		bardTestCard("f_attack", "火攻击", model.CardTypeAttack, model.ElementFire),
	}
	game.HandlePostDamageResolved(&model.PendingDamage{SourceID: "p1", TargetID: "p3", Damage: 1, DamageType: model.MagicAttack})
	game.HandlePostDamageResolved(&model.PendingDamage{SourceID: "p1", TargetID: "p4", Damage: 1, DamageType: model.MagicAttack})
	if paused := game.RunTurnEndTimingStageHooks(bard, engine.TimingTurnEndPreExtra); !paused {
		t.Fatalf("descent hook should trigger")
	}
	if n := len(obs.SkillActivated("bd_descent_concerto")); n != 0 {
		t.Fatalf("prompt should not emit skill_activated, got %d", n)
	}
	// 确认发动 → 弃火法术 → 弃火攻击 → 追加伤害目标 p3
	for step := 0; step < 4; step++ {
		if step == 3 {
			testutils.RequireChoiceContext(t, game, "p1", "bd_descent_target")
			if n := len(obs.SkillActivated("bd_descent_concerto")); n != 0 {
				t.Fatalf("should wait for bonus target before emitting, got %d", n)
			}
		}
		if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
			t.Fatalf("step %d failed: %v", step, err)
		}
	}
	testutils.RequireSingleSkillActivation(t, obs, "bd_descent_concerto", "p1", "p3")
}

func TestBardDescentConcerto_DeclineDoesNotEmitSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newBardFourPlayerGame(t, obs)
	bard := game.State.Players["p1"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Hand = []model.Card{
		bardTestCard("f_magic", "火法术", model.CardTypeMagic, model.ElementFire),
		bardTestCard("f_attack", "火攻击", model.CardTypeAttack, model.ElementFire),
	}
	game.HandlePostDamageResolved(&model.PendingDamage{SourceID: "p1", TargetID: "p3", Damage: 1, DamageType: model.MagicAttack})
	game.HandlePostDamageResolved(&model.PendingDamage{SourceID: "p1", TargetID: "p4", Damage: 1, DamageType: model.MagicAttack})
	game.RunTurnEndTimingStageHooks(bard, engine.TimingTurnEndPreExtra)
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("decline failed: %v", err)
	}
	if n := len(obs.SkillActivated("bd_descent_concerto")); n != 0 {
		t.Fatalf("declined descent should not emit skill_activated, got %d", n)
	}
}

func TestBardRousingRhapsody_EmitsSkillActivatedForHolderAndBard(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newBardFourPlayerGame(t, obs)
	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}
	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageTurnStart
	game.State.PendingInterrupt = nil
	game.Drive()

	testutils.RequireChoiceContext(t, game, "p2", "bd_rousing_mode")
	if n := len(obs.SkillActivated("bd_rousing_rhapsody")); n != 0 {
		t.Fatalf("prompt should not emit skill_activated, got %d", n)
	}
	steps := []struct {
		pid string
		sel int
	}{{"p2", 0}, {"p1", 0}, {"p1", 0}}
	for _, s := range steps {
		if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: s.pid, Selections: []int{s.sel}}); err != nil {
			t.Fatalf("select failed: %v", err)
		}
	}
	testutils.RequireSingleSkillActivation(t, obs, "bd_rousing_rhapsody", "p2", "p2", "p1")
	// 狂想曲结算后禁忌诗篇随之生效
	testutils.RequireSingleSkillActivation(t, obs, "bd_forbidden_verse", "p1", "p1")
}

func TestBardVictorySymphony_EmitsSkillActivatedForHolderAndBard(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}
	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageTurnEnd
	game.State.PendingInterrupt = nil
	game.Drive()

	testutils.RequireChoicePrompt(t, game, "p2", "bd_victory_confirm")
	if n := len(obs.SkillActivated("bd_victory_symphony")); n != 0 {
		t.Fatalf("prompt should not emit skill_activated, got %d", n)
	}
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{1}}); err != nil {
		t.Fatalf("choose victory branch failed: %v", err)
	}
	testutils.RequireSingleSkillActivation(t, obs, "bd_victory_symphony", "p2", "p2", "p1")
	testutils.RequireSingleSkillActivation(t, obs, "bd_forbidden_verse", "p1", "p1")
}

func TestBardVictorySymphony_CancelDoesNotEmitSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}
	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageTurnEnd
	game.State.PendingInterrupt = nil
	game.Drive()
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{2}}); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	for _, id := range []string{"bd_victory_symphony", "bd_forbidden_verse"} {
		if n := len(obs.SkillActivated(id)); n != 0 {
			t.Fatalf("cancel should not emit %s, got %d", id, n)
		}
	}
}
