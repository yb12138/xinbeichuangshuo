package soul_sorcerer_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestSoulSorcererSoulConvert_EmitsSkillActivatedOnce(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1.Tokens["ss_blue_soul"] = 1
	p1.Tokens["ss_yellow_soul"] = 0
	p1.Hand = []model.Card{soulSorcererTestCard("atk1", "暗袭", model.CardTypeAttack, model.ElementDark)}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.RequireChoicePrompt(t, game, "p1", "ss_convert_color")
	if n := len(obs.SkillActivated("ss_soul_convert")); n != 0 {
		t.Fatalf("expected no skill_activated while prompting, got %d", n)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})

	testutils.RequireSingleSkillActivation(t, obs, "ss_soul_convert", "p1", "p1")
}

func TestSoulSorcererSoulConvert_CancelDoesNotEmit(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1.Tokens["ss_blue_soul"] = 1
	p1.Hand = []model.Card{soulSorcererTestCard("atk1", "暗袭", model.CardTypeAttack, model.ElementDark)}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.RequireChoicePrompt(t, game, "p1", "ss_convert_color")
	// mode_order 仅含 b2y，索引 1 为取消。
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	if n := len(obs.SkillActivated("ss_soul_convert")); n != 0 {
		t.Fatalf("expected no skill_activated on cancel, got %d", n)
	}
}

func TestSoulSorcererSoulDevour_EmitsSkillActivatedOnce(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	ally := game.State.Players["p2"]
	loss := 2
	lossCtx := game.BuildContext(ally, nil, model.TimingMoraleLossCheck, &model.EventContext{
		Type:      model.EventDamage,
		DamageVal: &loss,
	})
	lossCtx.Selections = map[string]interface{}{
		"from_damage_draw":  true,
		"victim_id":         ally.ID,
		"discard_player_id": ally.ID,
	}
	game.Dispatcher().OnTiming(lossCtx.Timing, lossCtx)
	if n := len(obs.SkillActivated("ss_soul_devour")); n != 0 {
		t.Fatalf("expected no skill_activated during timing window, got %d", n)
	}
	game.ApplyMoraleLossAfterTimingWindow(ally, loss, false, true, 0, nil, lossCtx)

	if got := game.State.Players["p1"].Tokens["ss_yellow_soul"]; got != 2 {
		t.Fatalf("expected yellow soul 2, got %d", got)
	}
	testutils.RequireSingleSkillActivation(t, obs, "ss_soul_devour", "p1", "p2")
}
