package soul_sorcerer_test

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 审计 灵魂术士-2：灵魂召还选牌时可取消，手牌不动、回到行动选择。
func TestSoulSorcererSoulRecall_CancelRestoresActionWindow(t *testing.T) {
	game, p1, _ := setupSoulSorcererActionTurn(t)
	p1.Hand = []model.Card{
		soulSorcererTestCard("m1", "中毒", model.CardTypeMagic, model.ElementWater),
		soulSorcererTestCard("a1", "暗袭", model.CardTypeAttack, model.ElementDark),
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "ss_soul_recall"})
	testutils.RequireChoicePrompt(t, game, "p1", "ss_recall_pick")
	if p := game.GetCurrentPrompt(); p == nil || p.Presentation == nil || p.Presentation.CancelPolicy != model.CancelPolicyAbort {
		t.Fatalf("expected abort cancel policy, got %+v", p)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected recall aborted, got %+v", game.State.PendingInterrupt)
	}
	if len(p1.Hand) != 2 || p1.Tokens["ss_blue_soul"] != 0 {
		t.Fatalf("cancel must not discard or gain souls: hand=%d blue=%d", len(p1.Hand), p1.Tokens["ss_blue_soul"])
	}
	if p1.TurnState.HasActed || p1.TurnState.UsedSkillCounts["ss_soul_recall"] != 0 {
		t.Fatalf("expected action marks rolled back, got %+v", p1.TurnState)
	}
	if !game.IsActionSelectionWindow() {
		t.Fatalf("expected action selection window, stage=%s", game.State.TurnStage)
	}
}

// 审计 灵魂术士-4：没有可转换的灵魂（或两个方向都会溢出上限）时，主动攻击不弹灵魂转换。
func TestSoulSorcererSoulConvert_NoPromptWhenNothingToConvert(t *testing.T) {
	cases := []struct {
		name         string
		yellow, blue int
	}{
		{"no souls", 0, 0},
		{"both colors at cap", 6, 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			game, p1, _ := setupSoulSorcererActionTurn(t)
			p1.Tokens["ss_yellow_soul"] = tc.yellow
			p1.Tokens["ss_blue_soul"] = tc.blue
			p1.Hand = []model.Card{soulSorcererTestCard("atk1", "暗袭", model.CardTypeAttack, model.ElementDark)}
			testutils.MustHandleAction(t, game, model.PlayerAction{
				PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2",
				CardID: testutils.PlayableCardID(t, game, "p1", 0),
			})
			if intr := game.State.PendingInterrupt; intr != nil && testutils.ChoiceTypeOfInterrupt(intr) == "ss_convert_color" {
				t.Fatalf("soul convert must not prompt without convertible souls")
			}
		})
	}
}
