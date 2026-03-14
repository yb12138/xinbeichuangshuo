package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归：防御阶段只能打出【圣光】；手牌【圣盾】不能当作防御牌打出。
func TestCombatDefend_CannotPlayShieldFromHand(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "shield1", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"defend"},
		CardIndex: 0,
	})
	if err == nil {
		t.Fatalf("expected defend with shield in hand to fail")
	}
	if !strings.Contains(err.Error(), "圣盾") {
		t.Fatalf("expected shield-specific error, got: %v", err)
	}
	if len(p2.Hand) != 1 {
		t.Fatalf("shield card should not be consumed on invalid defend, hand=%d", len(p2.Hand))
	}
}

// 回归：防御阶段打出【圣光】仍应正常生效。
func TestCombatDefend_HolyLightStillValid(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "holy1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"defend"},
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("defend with holy light should succeed: %v", err)
	}

	if len(p2.Hand) != 0 {
		t.Fatalf("holy light should be consumed after defend, hand=%d", len(p2.Hand))
	}
	if len(game.State.CombatStack) != 0 {
		t.Fatalf("combat stack should be cleared after successful defend")
	}
}

// 回归：魔弹响应防御时，手牌【圣盾】不能打出，必须使用【圣光】。
func TestMagicBulletDefend_CannotPlayShieldFromHand(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Caster", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	p1.Hand = []model.Card{
		{ID: "mb1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "shield1", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementEarth},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("magic bullet failed: %v", err)
	}

	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"defend"},
		CardIndex: 0,
	})
	if err == nil {
		t.Fatalf("expected defend with shield in hand to fail for magic bullet")
	}
	if !strings.Contains(err.Error(), "圣盾") {
		t.Fatalf("expected shield-specific error, got: %v", err)
	}
	if len(p2.Hand) != 1 {
		t.Fatalf("shield card should not be consumed on invalid magic defend, hand=%d", len(p2.Hand))
	}
	if game.State.MagicBulletChain == nil {
		t.Fatalf("magic bullet chain should remain pending after invalid defend")
	}
}

// 回归：场上有【圣盾】时，魔弹不会提前自动结算；玩家先选择响应，选择承受后才触发圣盾。
func TestMagicBullet_FieldShieldAutoBlocks(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Caster", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 3

	p1.Hand = []model.Card{
		{ID: "mb1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p2.Field = []*model.FieldCard{
		{
			Card: model.Card{
				ID:      "shield_field_1",
				Name:    "圣盾",
				Type:    model.CardTypeMagic,
				Element: model.ElementEarth,
			},
			OwnerID:  "p2",
			SourceID: "p2",
			Mode:     model.FieldEffect,
			Effect:   model.EffectShield,
			Trigger:  model.EffectTriggerOnDamaged,
			Duration: 1,
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("magic bullet failed: %v", err)
	}

	if len(p2.Field) != 1 {
		t.Fatalf("field shield should not be consumed before response choice, field=%d", len(p2.Field))
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("magic bullet should wait for target response")
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take response failed: %v", err)
	}

	if len(p2.Field) != 0 {
		t.Fatalf("field shield should be consumed when choosing take, field=%d", len(p2.Field))
	}
	if p2.Heal != 3 {
		t.Fatalf("target heal should be unchanged after shield fallback, heal=%d", p2.Heal)
	}
	if game.State.MagicBulletChain != nil {
		t.Fatalf("magic bullet chain should end after shield fallback")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("no interrupt should remain after shield fallback")
	}
}

// 回归：场上有【圣盾】时，玩家仍可选择【圣光】抵挡；此时不应消耗圣盾。
func TestMagicBullet_FieldShieldCanStillDefendWithHolyLight(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Caster", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 3

	p1.Hand = []model.Card{
		{ID: "mb1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "holy1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}
	p2.Field = []*model.FieldCard{
		{
			Card: model.Card{
				ID:      "shield_field_1",
				Name:    "圣盾",
				Type:    model.CardTypeMagic,
				Element: model.ElementEarth,
			},
			OwnerID:  "p2",
			SourceID: "p2",
			Mode:     model.FieldEffect,
			Effect:   model.EffectShield,
			Trigger:  model.EffectTriggerOnDamaged,
			Duration: 1,
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("magic bullet failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"defend"},
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("defend with holy light should succeed: %v", err)
	}

	if len(p2.Hand) != 0 {
		t.Fatalf("holy light should be consumed after defend, hand=%d", len(p2.Hand))
	}
	if len(p2.Field) != 1 {
		t.Fatalf("field shield should remain after holy light defend, field=%d", len(p2.Field))
	}
	if p2.Heal != 3 {
		t.Fatalf("defend should not reduce heal, heal=%d", p2.Heal)
	}
	if game.State.MagicBulletChain != nil {
		t.Fatalf("magic bullet chain should end after defend")
	}
}

// 回归：魔弹传递到“下一个带圣盾目标”时，必须先弹响应框；不应提前自动结算圣盾。
func TestMagicBullet_PassToShieldedNextTargetNeedsPromptFirst(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Caster", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Relay", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "BlueMate", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	game.State.PlayerOrder = []string{"p1", "p3", "p2"}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p1.Heal = 3

	p1.Hand = []model.Card{
		{ID: "mb1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "mb2", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementFire, Damage: 2},
	}
	p1.Field = []*model.FieldCard{
		{
			Card: model.Card{
				ID:      "shield_field_1",
				Name:    "圣盾",
				Type:    model.CardTypeMagic,
				Element: model.ElementEarth,
			},
			OwnerID:  "p1",
			SourceID: "p1",
			Mode:     model.FieldEffect,
			Effect:   model.EffectShield,
			Trigger:  model.EffectTriggerOnDamaged,
			Duration: 1,
		},
	}

	// p1 发起魔弹，目标是 p2
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("magic bullet failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected first response prompt for p2")
	}

	// p2 使用魔弹传递给下一位敌方（p1，且 p1 有圣盾）
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"counter"},
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("counter response failed: %v", err)
	}

	if game.State.MagicBulletChain == nil {
		t.Fatalf("magic bullet chain should continue after counter")
	}
	if game.State.MagicBulletChain.TargetID != "p1" {
		t.Fatalf("expected next target p1, got %s", game.State.MagicBulletChain.TargetID)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicMissile || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected prompt for shielded next target p1 before resolution")
	}
	if len(p1.Field) != 1 {
		t.Fatalf("shield should not be consumed before p1 chooses response, field=%d", len(p1.Field))
	}

	// p1 选择承受，才触发圣盾自动抵挡
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take response failed: %v", err)
	}
	if len(p1.Field) != 0 {
		t.Fatalf("shield should be consumed after p1 chooses take, field=%d", len(p1.Field))
	}
	if game.State.MagicBulletChain != nil {
		t.Fatalf("magic bullet chain should end after shield fallback")
	}
}

// 回归：攻击响应阶段应先给目标玩家选择（应战/防御/承受），不应因场上有圣盾而提前自动抵挡。
func TestCombatShield_WaitsForPlayerChoice(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Field = []*model.FieldCard{
		{
			Card: model.Card{
				ID:      "shield_field_1",
				Name:    "圣盾",
				Type:    model.CardTypeMagic,
				Element: model.ElementEarth,
			},
			OwnerID:  "p2",
			SourceID: "p2",
			Mode:     model.FieldEffect,
			Effect:   model.EffectShield,
			Trigger:  model.EffectTriggerOnDamaged,
			Duration: 1,
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	if game.State.Phase != model.PhaseCombatInteraction {
		t.Fatalf("expected phase CombatInteraction, got=%s", game.State.Phase)
	}
	if len(game.State.CombatStack) != 1 {
		t.Fatalf("combat stack should wait for response, got=%d", len(game.State.CombatStack))
	}
	if len(p2.Field) != 1 {
		t.Fatalf("shield should not be auto consumed before choice, field=%d", len(p2.Field))
	}
}

// 回归：目标选择承受时，场上圣盾应抵挡本次攻击并被消耗。
func TestCombatShield_ConsumeOnTake(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 3

	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Field = []*model.FieldCard{
		{
			Card: model.Card{
				ID:      "shield_field_1",
				Name:    "圣盾",
				Type:    model.CardTypeMagic,
				Element: model.ElementEarth,
			},
			OwnerID:  "p2",
			SourceID: "p2",
			Mode:     model.FieldEffect,
			Effect:   model.EffectShield,
			Trigger:  model.EffectTriggerOnDamaged,
			Duration: 1,
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take response failed: %v", err)
	}

	if len(p2.Field) != 0 {
		t.Fatalf("shield should be consumed after take fallback, field=%d", len(p2.Field))
	}
	if p2.Heal != 3 {
		t.Fatalf("shield fallback should prevent damage, heal=%d", p2.Heal)
	}
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("no pending damage expected after shield fallback, queue=%d", len(game.State.PendingDamageQueue))
	}
}

// 回归：目标身上有圣盾时，仍可主动选择应战；应战成功后不应消耗圣盾。
func TestCombatShield_CounterChoiceKeepsShield(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AttackerMate", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()

	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "counter1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Field = []*model.FieldCard{
		{
			Card: model.Card{
				ID:      "shield_field_1",
				Name:    "圣盾",
				Type:    model.CardTypeMagic,
				Element: model.ElementEarth,
			},
			OwnerID:  "p2",
			SourceID: "p2",
			Mode:     model.FieldEffect,
			Effect:   model.EffectShield,
			Trigger:  model.EffectTriggerOnDamaged,
			Duration: 1,
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"counter"},
		CardIndex: 0,
		TargetID:  "p3",
	}); err != nil {
		t.Fatalf("counter response failed: %v", err)
	}

	if len(p2.Field) != 1 {
		t.Fatalf("shield should remain after counter choice, field=%d", len(p2.Field))
	}
	if len(game.State.CombatStack) == 0 {
		t.Fatalf("counter should create follow-up combat request")
	}
}
