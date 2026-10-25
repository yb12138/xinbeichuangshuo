package magical_girl_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

type noopMagicBulletObserver struct{}

func (noopMagicBulletObserver) OnGameEvent(event model.GameEvent) {}

func buildMagicActionGame(t *testing.T) *engine.GameEngine {
	t.Helper()

	game := engine.NewGameEngine(noopMagicBulletObserver{})
	game.State.Deck = rules.InitDeck()

	if err := game.AddPlayer("p1", "P1", "berserker", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "P2", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.PlayerOrder = []string{"p1", "p2"}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()

	return game
}

func TestMagicBullet_AllowsMagicWithoutExplicitTarget(t *testing.T) {
	game := buildMagicActionGame(t)
	p1 := game.State.Players["p1"]

	p1.Hand = []model.Card{
		{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}

	err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdMagic,
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	if err != nil {
		t.Fatalf("magic bullet without target should succeed, got: %v", err)
	}

	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected magic missile interrupt, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("expected interrupt type %s, got %s", model.InterruptMagicMissile, game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected first bullet target p2, got %s", game.State.PendingInterrupt.PlayerID)
	}
}

func TestMagicWithoutTarget_StillRequiresTargetForNonMagicBullet(t *testing.T) {
	game := buildMagicActionGame(t)
	p1 := game.State.Players["p1"]

	p1.Hand = []model.Card{
		{ID: "shield", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	}

	err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdMagic,
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	if err == nil {
		t.Fatalf("non-magic-bullet without target should fail")
	}
}

func magicMissilePromptOptionIDs(t *testing.T, game *engine.GameEngine) (map[string]bool, *model.Prompt) {
	t.Helper()
	prompt := game.GetCurrentPrompt()
	if prompt == nil {
		t.Fatalf("expected magic missile prompt")
	}
	ids := map[string]bool{}
	for _, o := range prompt.Options {
		ids[o.ID] = true
	}
	return ids, prompt
}

// 被魔弹击中但手里没有魔弹/圣光时，只能承受——不应出现传递/抵挡选项。
func TestMagicMissilePrompt_NoUsableCards_OnlyTake(t *testing.T) {
	game := buildMagicActionGame(t)
	game.State.Players["p1"].Hand = []model.Card{
		{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	game.State.Players["p2"].Hand = []model.Card{
		{ID: "fire", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "shield", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
	}
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: "mb"}); err != nil {
		t.Fatalf("cast magic bullet: %v", err)
	}
	ids, prompt := magicMissilePromptOptionIDs(t, game)
	if !ids["take"] || ids["counter"] || ids["defend"] {
		t.Fatalf("expected only take, got %+v", prompt.Options)
	}
}

func TestMagicMissilePrompt_ListsBulletAndHolyLight(t *testing.T) {
	game := buildMagicActionGame(t)
	game.State.Players["p1"].Hand = []model.Card{
		{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	game.State.Players["p2"].Hand = []model.Card{
		{ID: "mb2", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementFire, Damage: 2},
		{ID: "light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: "mb"}); err != nil {
		t.Fatalf("cast magic bullet: %v", err)
	}
	ids, prompt := magicMissilePromptOptionIDs(t, game)
	if !ids["counter"] || !ids["defend"] {
		t.Fatalf("expected counter+defend, got %+v", prompt.Options)
	}
	if len(prompt.CounterCardIDs) != 1 || prompt.CounterCardIDs[0] != "mb2" {
		t.Fatalf("expected counter card mb2, got %v", prompt.CounterCardIDs)
	}
	if len(prompt.DefendCardIDs) != 1 || prompt.DefendCardIDs[0] != "light" {
		t.Fatalf("expected defend card light, got %v", prompt.DefendCardIDs)
	}
}
