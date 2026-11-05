package priest_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 神官-5：神圣领域伤害分支文案为"目标角色"，可以指定自己。
func TestPriestDivineDomain_DamageBranchCanTargetSelf(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Priest", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 2
	p1.Hand = nil

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "priest_divine_domain"})
	testutils.RequireChoicePrompt(t, game, "p1", "priest_divine_domain_mode")
	// 没有队友：只有伤害分支
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	ctx := testutils.RequireChoiceContext(t, game, "p1", "priest_divine_domain_damage_target")
	idx := testutils.ChoiceIndexForTarget(t, ctx, "p1")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{idx}})

	if p1.Heal != 1 {
		t.Fatalf("expected priest to remove 1 heal, got %d", p1.Heal)
	}
	found := false
	for _, pd := range game.State.PendingDamageQueue {
		if pd.TargetID == "p1" && pd.Damage == 2 {
			found = true
		}
	}
	if !found && len(p1.Hand) == 0 {
		t.Fatalf("expected 2 magic damage on self, queue=%+v hand=%d", game.State.PendingDamageQueue, len(p1.Hand))
	}
}
