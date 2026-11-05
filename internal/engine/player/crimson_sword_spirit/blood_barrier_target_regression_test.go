package crimson_sword_spirit_test

import (
	"fmt"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// setupBloodBarrierGame: p1 血色剑灵(红)、p3 队友(红)、p2/p4 对手(蓝)。
func setupBloodBarrierGame(t *testing.T) *engine.GameEngine {
	t.Helper()
	g := engine.NewGameEngine(testutils.NoopObserver{})
	for _, spec := range []struct {
		id, role string
		camp     model.Camp
	}{
		{"p1", "crimson_sword_spirit", model.RedCamp},
		{"p2", "angel", model.BlueCamp},
		{"p3", "berserker", model.RedCamp},
		{"p4", "berserker", model.BlueCamp},
	} {
		if err := g.AddPlayer(spec.id, spec.id, spec.role, spec.camp); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range g.State.Players {
		p.TurnState = model.NewPlayerTurnState()
		p.Heal = 0
		p.Hand = nil
	}
	g.State.Players["p1"].Tokens["css_blood"] = 1
	deck := make([]model.Card, 0, 20)
	for i := 0; i < 20; i++ {
		deck = append(deck, model.Card{ID: fmt.Sprintf("d%d", i), Name: "补牌", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2})
	}
	g.State.Deck = deck
	return g
}

func runBloodBarrierToTargetPrompt(t *testing.T, g *engine.GameEngine, sourceID string) *model.Prompt {
	t.Helper()
	g.AddPendingDamage(model.PendingDamage{SourceID: sourceID, TargetID: "p1", Damage: 2, DamageType: model.MagicAttack})
	if paused := g.ProcessPendingDamages(); !paused {
		t.Fatalf("expected blood barrier response prompt")
	}
	testutils.RequireResponseSkillPrompt(t, g, "p1")
	if err := g.ConfirmResponseSkill("p1", "css_blood_barrier"); err != nil {
		t.Fatalf("confirm blood barrier failed: %v", err)
	}
	if got := g.State.Players["p1"].Tokens["css_blood"]; got != 0 {
		t.Fatalf("expected blood spent immediately, got %d", got)
	}
	if len(g.State.PendingDamageQueue) == 0 || g.State.PendingDamageQueue[0].Damage != 1 {
		t.Fatalf("expected original damage reduced to 1 immediately, got %+v", g.State.PendingDamageQueue)
	}
	testutils.RequireChoiceContext(t, g, "p1", "css_blood_barrier_target")
	prompt := g.BuildPendingInterruptPrompt()
	if prompt == nil {
		t.Fatalf("expected blood barrier target prompt")
	}
	for _, opt := range prompt.Options {
		if opt.ID == "cancel" {
			t.Fatalf("blood barrier target prompt must not be cancellable, got %+v", prompt.Options)
		}
	}
	return prompt
}

func optionTargets(prompt *model.Prompt) []string {
	ids := make([]string, 0, len(prompt.Options))
	for _, opt := range prompt.Options {
		ids = append(ids, opt.TargetID)
	}
	return ids
}

func selectTarget(t *testing.T, g *engine.GameEngine, prompt *model.Prompt, targetID string) {
	t.Helper()
	for i, opt := range prompt.Options {
		if opt.TargetID == targetID {
			testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{i}})
			return
		}
	}
	t.Fatalf("target %s not in options %+v", targetID, optionTargets(prompt))
}

func TestCrimsonBloodBarrier_EnemySourceCanRetargetAnotherEnemy(t *testing.T) {
	g := setupBloodBarrierGame(t)
	prompt := runBloodBarrierToTargetPrompt(t, g, "p2")
	if got := optionTargets(prompt); len(got) != 2 || got[0] == got[1] || (got[0] != "p2" && got[0] != "p4") || (got[1] != "p2" && got[1] != "p4") {
		t.Fatalf("expected options to be exactly enemies p2/p4, got %+v", got)
	}
	selectTarget(t, g, prompt, "p4")
	g.Drive()

	if got := len(g.State.Players["p1"].Hand); got != 1 {
		t.Fatalf("expected original damage resolved as 1 (p1 draws 1), got hand=%d", got)
	}
	if got := len(g.State.Players["p4"].Hand); got != 1 {
		t.Fatalf("expected reflected 1 magic damage on p4, got hand=%d", got)
	}
	if got := len(g.State.Players["p2"].Hand); got != 0 {
		t.Fatalf("expected original source p2 untouched, got hand=%d", got)
	}
	if len(g.State.PendingDamageQueue) != 0 || g.State.PendingInterrupt != nil {
		t.Fatalf("expected damage flow finished, queue=%+v intr=%+v", g.State.PendingDamageQueue, g.State.PendingInterrupt)
	}
}

func TestCrimsonBloodBarrier_AllySourceOnlyTargetsEnemies(t *testing.T) {
	g := setupBloodBarrierGame(t)
	prompt := runBloodBarrierToTargetPrompt(t, g, "p3")
	for _, id := range optionTargets(prompt) {
		if id != "p2" && id != "p4" {
			t.Fatalf("expected only enemies as options when source is ally, got %+v", optionTargets(prompt))
		}
	}
	if len(prompt.Options) != 2 {
		t.Fatalf("expected 2 enemy options, got %+v", optionTargets(prompt))
	}
	// 非法索引（越界）应被拒绝
	if err := g.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{5}}); err == nil {
		t.Fatalf("expected invalid selection rejected")
	}
	selectTarget(t, g, prompt, "p2")
	g.Drive()

	if got := len(g.State.Players["p1"].Hand); got != 1 {
		t.Fatalf("expected original damage resolved as 1, got hand=%d", got)
	}
	if got := len(g.State.Players["p2"].Hand); got != 1 {
		t.Fatalf("expected reflected damage on enemy p2, got hand=%d", got)
	}
	if got := len(g.State.Players["p3"].Hand); got != 0 {
		t.Fatalf("expected ally source p3 not hit, got hand=%d", got)
	}
	if len(g.State.PendingDamageQueue) != 0 || g.State.PendingInterrupt != nil {
		t.Fatalf("expected damage flow finished, queue=%+v intr=%+v", g.State.PendingDamageQueue, g.State.PendingInterrupt)
	}
}
