package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestChoiceEngineUnknownChoiceTypeErrors(t *testing.T) {
	g := NewGameEngine(nil)
	g.State.Players["p1"] = &model.Player{ID: "p1", Name: "P1"}
	g.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: "p1",
		Context: map[string]interface{}{
			"choice_type": "totally_unknown_choice_type_xyz",
			"user_id":     "p1",
		},
	}
	err := g.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}})
	if err == nil {
		t.Fatal("expected error for unregistered choice_type")
	}
}
