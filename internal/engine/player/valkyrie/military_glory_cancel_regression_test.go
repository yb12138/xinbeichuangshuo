package valkyrie_test

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func requireCancelPolicy(t *testing.T, prompt *model.Prompt, want string) {
	t.Helper()
	if prompt == nil || prompt.Presentation == nil || prompt.Presentation.CancelPolicy != want {
		t.Fatalf("expected cancel_policy=%q, got %+v", want, prompt)
	}
}

// 审计 女武神-3：军威神光中间步骤可返回上一步，第一步可取消。
func TestValkyrie_MilitaryGlory_StepsCanGoBack(t *testing.T) {
	game := newMilitaryGloryGame(t, testutils.NoopObserver{})
	game.State.RedGems = 1
	game.State.RedCrystals = 1
	requireCancelPolicy(t, game.GetCurrentPrompt(), model.CancelPolicyAbort)

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	testutils.RequireChoicePrompt(t, game, "p1", "valkyrie_military_glory_x")
	requireCancelPolicy(t, game.GetCurrentPrompt(), model.CancelPolicyBack)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	testutils.RequireChoicePrompt(t, game, "p1", "valkyrie_military_glory_mode")

	chooseMilitaryGloryTarget(t, game, 1, "p2")
	testutils.RequireChoicePrompt(t, game, "p1", "valkyrie_military_glory_pay")
	requireCancelPolicy(t, game.GetCurrentPrompt(), model.CancelPolicyBack)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	testutils.RequireChoicePrompt(t, game, "p1", "valkyrie_military_glory_target")
	requireCancelPolicy(t, game.GetCurrentPrompt(), model.CancelPolicyBack)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	testutils.RequireChoicePrompt(t, game, "p1", "valkyrie_military_glory_x")
	if game.State.RedGems != 1 || game.State.RedCrystals != 1 || game.State.Players["p2"].Heal != 0 {
		t.Fatalf("going back must not pay or heal: gems=%d crystals=%d heal=%d", game.State.RedGems, game.State.RedCrystals, game.State.Players["p2"].Heal)
	}

	// 返回后重新走完流程仍可正常结算。
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	chooseMilitaryGloryTarget(t, game, 1, "p2")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	if game.State.RedGems != 0 || game.State.Players["p2"].Heal != 1 {
		t.Fatalf("expected settle after going back: gems=%d heal=%d", game.State.RedGems, game.State.Players["p2"].Heal)
	}
}

func TestValkyrie_MilitaryGlory_FirstStepCancelSkips(t *testing.T) {
	game := newMilitaryGloryGame(t, testutils.NoopObserver{})
	healBefore := game.State.Players["p1"].Heal
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	if intr := game.State.PendingInterrupt; intr != nil {
		if ctx, ok := intr.Context.(map[string]interface{}); ok && ctx["choice_type"] == "valkyrie_military_glory_mode" {
			t.Fatalf("military glory prompt should be gone after cancel")
		}
	}
	p1 := game.State.Players["p1"]
	if p1.Heal != healBefore || p1.Form != model.FormValkyrieHeroic || game.State.RedCrystals != 2 {
		t.Fatalf("cancel must have no effect: heal=%d form=%s crystals=%d", p1.Heal, p1.Form, game.State.RedCrystals)
	}
}
