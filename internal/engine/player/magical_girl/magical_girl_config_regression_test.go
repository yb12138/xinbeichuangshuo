package magical_girl_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestMagicalGirl_MagicBulletFusion_ReverseChainUsesConfiguredDirection(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Ally", "saintess", model.RedCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p4 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.PlayerOrder = []string{"p1", "p2", "p3", "p4"}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "poison", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementEarth},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p4",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("cast earth magic failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicBulletFusion {
		t.Fatalf("expected fusion interrupt, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm fusion failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicBulletDirection {
		t.Fatalf("expected direction interrupt after fusion, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("choose reverse direction failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("expected magic missile interrupt after direction choice, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected reverse chain to target p3, got %s", game.State.PendingInterrupt.PlayerID)
	}
	if game.State.MagicBulletChain == nil || !game.State.MagicBulletChain.Reverse {
		t.Fatalf("expected reverse magic bullet chain, got %+v", game.State.MagicBulletChain)
	}
	if game.State.MagicBulletChain == nil || !game.State.MagicBulletChain.IsFusion {
		t.Fatalf("expected fusion magic bullet chain, got %+v", game.State.MagicBulletChain)
	}
}

func TestMagicalGirl_MagicBulletFusion_DeclineKeepsOriginalSpell(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "shield", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementEarth},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("cast earth magic failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("decline fusion failed: %v", err)
	}

	if game.State.MagicBulletChain != nil {
		t.Fatalf("expected no magic bullet chain after declining fusion, got %+v", game.State.MagicBulletChain)
	}
	if !p2.HasFieldEffect(model.EffectShield) {
		t.Fatalf("expected original shield effect to be placed on target")
	}
}

func TestMagicalGirl_MagicBlast_DiscardsAfterEachFailedTarget(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	game.State.Deck = rules.InitDeck()
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "cost", Name: "法术代价", Type: model.CardTypeMagic, Element: model.ElementWater},
		{ID: "a1", Name: "弃牌1", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "m1", Name: "弃牌2", Type: model.CardTypeMagic, Element: model.ElementEarth},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "magic_blast",
		TargetIDs:  []string{"p2", "p3"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("magic blast failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected first target interrupt on p2, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdCancel}); err != nil {
		t.Fatalf("p2 decline failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected caster forced discard after first decline, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("caster first discard failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected second target interrupt on p3, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p3", Type: model.CmdCancel}); err != nil {
		t.Fatalf("p3 decline failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected caster forced discard after second decline, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("caster second discard failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected magic blast flow to end, got %+v", game.State.PendingInterrupt)
	}
	if got := len(p1.Hand); got != 0 {
		t.Fatalf("expected caster hand to be empty after 1 cost + 2 forced discards, got %d", got)
	}
	if got := game.State.RedGems; got != 1 {
		t.Fatalf("expected team gem +1 from magic blast, got %d", got)
	}

	game.Drive()
	game.Drive()

	if got := len(p2.Hand); got != 2 {
		t.Fatalf("expected p2 draw 2 after taking magic damage, got %d", got)
	}
	if got := len(p3.Hand); got != 2 {
		t.Fatalf("expected p3 draw 2 after taking magic damage, got %d", got)
	}
}

func TestMagicalGirl_MagicBlast_TargetCanSelectMagicByHandIndex(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "cost", Name: "法术代价", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	p2.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "a2", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth},
		{ID: "a3", Name: "风刃", Type: model.CardTypeAttack, Element: model.ElementWind},
		{ID: "m1", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "magic_blast",
		TargetIDs:  []string{"p2", "p3"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("magic blast failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected first target interrupt on p2, got %+v", game.State.PendingInterrupt)
	}

	// 前端 choose_cards 会提交手牌真实索引；这里验证后端兼容该编码。
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{3},
	}); err != nil {
		t.Fatalf("expected selecting magic card by hand index to work, got error: %v", err)
	}

	if got := len(p2.Hand); got != 3 {
		t.Fatalf("expected p2 to discard one magic card, got hand size %d", got)
	}
	for _, card := range p2.Hand {
		if card.Type == model.CardTypeMagic {
			t.Fatalf("expected p2 magic card to be discarded, remaining hand=%+v", p2.Hand)
		}
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected flow to advance to second target p3, got %+v", game.State.PendingInterrupt)
	}
}

func TestMagicalGirl_DestructionStorm_RequiresTwoTargetsAndCostsOneGem(t *testing.T) {
	t.Run("requires_two_targets", func(t *testing.T) {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
			t.Fatalf("add p1 failed: %v", err)
		}
		if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
			t.Fatalf("add p2 failed: %v", err)
		}
		if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
			t.Fatalf("add p3 failed: %v", err)
		}

		game.State.CurrentTurn = 0
		game.State.TurnStage = model.TurnStageActionExecution

		p1 := game.State.Players["p1"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		p1.Gem = 1

		err := game.HandleAction(model.PlayerAction{
			PlayerID:  "p1",
			Type:      model.CmdSkill,
			SkillID:   "destruction_storm",
			TargetIDs: []string{"p2"},
		})
		if err == nil {
			t.Fatalf("expected destruction storm to reject fewer than two targets")
		}
	})

	t.Run("costs_one_gem", func(t *testing.T) {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		game.State.Deck = rules.InitDeck()
		if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
			t.Fatalf("add p1 failed: %v", err)
		}
		if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
			t.Fatalf("add p2 failed: %v", err)
		}
		if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
			t.Fatalf("add p3 failed: %v", err)
		}

		game.State.CurrentTurn = 0
		game.State.TurnStage = model.TurnStageActionExecution

		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p3 := game.State.Players["p3"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		p2.TurnState = model.NewPlayerTurnState()
		p3.TurnState = model.NewPlayerTurnState()
		p1.Gem = 1

		if err := game.HandleAction(model.PlayerAction{
			PlayerID:  "p1",
			Type:      model.CmdSkill,
			SkillID:   "destruction_storm",
			TargetIDs: []string{"p2", "p3"},
		}); err != nil {
			t.Fatalf("destruction storm failed: %v", err)
		}

		if got := p1.Gem; got != 0 {
			t.Fatalf("expected destruction storm to consume exactly 1 gem, got %d", got)
		}

		game.Drive()
		game.Drive()

		if got := len(p2.Hand); got != 2 {
			t.Fatalf("expected p2 draw 2 after destruction storm, got %d", got)
		}
		if got := len(p3.Hand); got != 2 {
			t.Fatalf("expected p3 draw 2 after destruction storm, got %d", got)
		}
	})
}
