package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归：风之剑圣第3次主动攻击触发【圣剑】后，强制命中应无视目标场上圣盾。
func TestBladeMaster_HolySword_ThirdAttackForceHitIgnoresShield(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
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
	// 避免行动结束后弹出无关响应中断，聚焦圣剑命中链。
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p1.TurnState.UsedSkillCounts["sword_shadow"] = 1
	p1.Gem = 0
	p1.Crystal = 0

	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "火斩2", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a3", Name: "火斩3", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	model.AppendAttackAction(p1, "test-extra-1")
	model.AppendAttackAction(p1, "test-extra-2")
	p2.Heal = 0

	// 前两次攻击正常命中，堆到第3次攻击条件。
	for i := 0; i < 2; i++ {
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
		}); err != nil {
			t.Fatalf("attack #%d failed: %v", i+1, err)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
		}); err != nil {
			t.Fatalf("respond take #%d failed: %v", i+1, err)
		}
		if game.State.PendingInterrupt != nil {
			t.Fatalf("unexpected interrupt after attack #%d: %+v", i+1, game.State.PendingInterrupt)
		}
	}

	// 第3次攻击前给目标挂圣盾，验证圣剑是否正确无视。
	p2.Field = append(p2.Field, &model.FieldCard{
		Card:   model.Card{ID: "shield-1", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
		Mode:   model.FieldEffect,
		Effect: model.EffectShield,
	})
	redGemsBeforeThird := game.State.RedGems

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("third attack failed: %v", err)
	}

	if p1.TurnState.AttackCount != 3 {
		t.Fatalf("expected attack count=3, got %d", p1.TurnState.AttackCount)
	}
	if game.State.RedGems <= redGemsBeforeThird {
		t.Fatalf("expected holy_sword third attack to count as hit and add gem, gems before=%d after=%d", redGemsBeforeThird, game.State.RedGems)
	}
	if !p2.HasFieldEffect(model.EffectShield) {
		t.Fatalf("expected shield to remain when holy_sword attack ignores shield")
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.PlayerID == "p2" {
		t.Fatalf("expected holy_sword forced hit to skip target response prompt, got %+v", game.State.PendingInterrupt)
	}
}
