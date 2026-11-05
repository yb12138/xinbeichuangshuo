package holy_lancer_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func newHolyStrikeGame(t *testing.T, heal int, withEarthSpear bool) (*engine.GameEngine, *model.Player) {
	t.Helper()
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "HolyLancer", "holy_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens = map[string]int{}
	p2.TurnState = model.NewPlayerTurnState()
	p2.Hand = nil
	p1.MaxHeal = 3
	p1.Heal = heal
	p1.Hand = []model.Card{
		{ID: "atk-fire-1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	if !withEarthSpear {
		// 复制角色定义后去掉地枪，模拟"本次命中不会弹出地枪"的情形。
		char := *p1.Character
		skills := make([]model.SkillDefinition, 0, len(char.Skills))
		for _, sk := range char.Skills {
			if sk.ID != "holy_lancer_earth_spear" {
				skills = append(skills, sk)
			}
		}
		char.Skills = skills
		p1.Character = &char
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	return game, p1
}

// 回归：有治疗但地枪未被提供时，圣击 +1 仍应在命中时结算，而不是因为等不到"跳过地枪"而丢失。
func TestHolyLancerHolyStrike_AppliesWhenEarthSpearNotOffered(t *testing.T) {
	game, p1 := newHolyStrikeGame(t, 1, false)
	if intr := game.State.PendingInterrupt; intr != nil && intr.Type == model.InterruptResponseSkill && intr.PlayerID == "p1" {
		t.Fatalf("expected no response prompt for p1, got %+v", intr.SkillIDs)
	}
	if p1.Heal != 2 {
		t.Fatalf("expected holy strike +1 heal (1->2), got %d", p1.Heal)
	}
}

// 回归：治疗为 0（地枪不可用）时圣击照常 +1。
func TestHolyLancerHolyStrike_AppliesWithZeroHeal(t *testing.T) {
	_, p1 := newHolyStrikeGame(t, 0, true)
	if p1.Heal != 1 {
		t.Fatalf("expected holy strike +1 heal (0->1), got %d", p1.Heal)
	}
}

// 回归：地枪被提供并跳过时，圣击只 +1 一次。
func TestHolyLancerHolyStrike_SkipEarthSpearAppliesOnce(t *testing.T) {
	game, p1 := newHolyStrikeGame(t, 1, true)
	intr := game.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill || findSkillIndex(intr.SkillIDs, "holy_lancer_earth_spear") < 0 {
		t.Fatalf("expected earth spear response prompt, got %+v", intr)
	}
	if p1.Heal != 1 {
		t.Fatalf("holy strike should wait for earth spear decision, got heal=%d", p1.Heal)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	if p1.Heal != 2 {
		t.Fatalf("expected exactly +1 heal after skipping earth spear, got %d", p1.Heal)
	}
}
