package moon_test

import (
	"strings"
	"testing"

	"starcup-engine/internal/engine"
	moonplayer "starcup-engine/internal/engine/player/moon_goddess"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 月之女神-5：闇月斩只给本次攻击（来源=月之女神、攻击伤害、目标=被攻击者）加伤，
// 队列前面其他来源/法术伤害不受影响。
func TestMoonGoddessDarkMoonSlash_BonusTargetsOwnAttackDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Crystal = 1
	moon.Form = model.FormMoonGoddessDarkMoon
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.PendingDamageQueue = []model.PendingDamage{
		{SourceID: "p2", TargetID: "p1", Damage: 3, DamageType: model.MagicAttack},
		{SourceID: "p1", TargetID: "p3", Damage: 1, DamageType: model.MagicAttack},
		{SourceID: "p1", TargetID: "p3", Damage: 2, DamageType: model.AttackDamage},
		{SourceID: "p1", TargetID: "p2", Damage: 2, DamageType: model.AttackDamage},
	}
	ctx := game.BuildContext(moon, game.State.Players["p2"], model.TimingAttackHit, &model.EventContext{
		Type:       model.EventAttack,
		SourceID:   "p1",
		TargetID:   "p2",
		AttackInfo: &model.AttackEventInfo{ActionType: string(model.ActionAttack), IsHit: true},
	})
	h := &moonplayer.MoonGoddessDarkMoonSlashHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected dark moon slash usable")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_darkmoon_slash_x")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose X=1 failed: %v", err)
	}
	want := []int{3, 1, 2, 3}
	for i, pd := range game.State.PendingDamageQueue {
		if pd.Damage != want[i] {
			t.Fatalf("queue[%d] damage=%d want %d (queue=%+v)", i, pd.Damage, want[i], game.State.PendingDamageQueue)
		}
	}
}

// 月之女神-7①：月之轮回分支①提示写明移除闇月会令我方士气 -1。
func TestMoonGoddessMoonCycle_Branch1PromptWarnsMoraleLoss(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 2
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd
	if !moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("expected moon cycle interrupt")
	}
	prompt := game.GetCurrentPrompt()
	found := false
	for _, opt := range prompt.Options {
		if opt.ID == "branch1" {
			found = true
			if !strings.Contains(opt.Label, "移除闇月会让我方士气 -1") {
				t.Fatalf("branch1 label should warn morale loss, got %q", opt.Label)
			}
		}
	}
	if !found {
		t.Fatalf("expected branch1 option, got %+v", prompt.Options)
	}
}

// 月之女神-7②：苍白之月扣宝石后，弃牌结算前任一步取消都退回宝石与发动次数，并回到行动选择。
func TestMoonGoddessPaleMoon_CancelBeforeDiscardRefundsGem(t *testing.T) {
	steps := []struct {
		name  string
		picks []int
		want  string
	}{
		{name: "mode", picks: nil, want: "mg_pale_moon_mode"},
		{name: "x", picks: []int{0}, want: "mg_pale_moon_x"},
		{name: "target", picks: []int{0, 0}, want: "mg_pale_moon_target"},
		{name: "discard", picks: []int{0, 0, 0}, want: "mg_pale_moon_discard"},
	}
	for _, tc := range steps {
		t.Run(tc.name, func(t *testing.T) {
			game := engine.NewGameEngine(testutils.NoopObserver{})
			if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
				t.Fatal(err)
			}
			moon := game.State.Players["p1"]
			moon.IsActive = true
			moon.TurnState = model.NewPlayerTurnState()
			moon.Gem = 1
			moon.Tokens["mg_new_moon"] = 2
			moon.Hand = []model.Card{moonTestCard("h1", "弃牌", model.CardTypeAttack, model.ElementFire)}
			game.State.CurrentTurn = 0
			game.State.TurnStage = model.TurnStageActionExecution

			if err := game.UseSkill("p1", "mg_pale_moon", nil, nil); err != nil {
				t.Fatalf("use pale moon failed: %v", err)
			}
			if moon.Gem != 0 {
				t.Fatalf("expected gem paid, got %d", moon.Gem)
			}
			for _, pick := range tc.picks {
				testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{pick}})
			}
			testutils.RequireChoicePrompt(t, game, "p1", tc.want)
			testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdCancel, PlayerID: "p1"})

			if game.State.PendingInterrupt != nil {
				t.Fatalf("expected pale moon flow abandoned, got %+v", game.State.PendingInterrupt)
			}
			if moon.Gem != 1 {
				t.Fatalf("expected gem refunded, got %d", moon.Gem)
			}
			if got := moon.TurnState.UsedSkillCounts["mg_pale_moon"]; got != 0 {
				t.Fatalf("expected skill count refunded, got %d", got)
			}
			if moon.Tokens["mg_new_moon"] != 2 || len(moon.Hand) != 1 {
				t.Fatalf("cancel must not consume new moon or hand: newMoon=%d hand=%d", moon.Tokens["mg_new_moon"], len(moon.Hand))
			}
			if game.State.TurnStage != model.TurnStageActionExecution || moon.TurnState.HasActed {
				t.Fatalf("expected back to action selection, stage=%s hasActed=%v", game.State.TurnStage, moon.TurnState.HasActed)
			}
		})
	}
}

// 通用-10：新月/石化导出上限与实际生效的 10 一致。
func TestMoonGoddessTokenCaps_MatchEffectiveCap(t *testing.T) {
	if moonplayer.NewMoonCap != 10 || moonplayer.PetrifyCap != 10 {
		t.Fatalf("expected caps 10/10, got %d/%d", moonplayer.NewMoonCap, moonplayer.PetrifyCap)
	}
}
