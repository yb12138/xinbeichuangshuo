package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func heroTestCard(id, name string, cardType model.CardType, element model.Element, damage int) model.Card {
	if damage <= 0 {
		damage = 2
	}
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     element,
		Damage:      damage,
		Description: name,
	}
}

func heroTauntExclusiveCard(owner *model.Player) model.Card {
	charName := "勇者"
	faction := "血"
	if owner != nil && owner.Character != nil {
		charName = owner.Character.Name
		faction = owner.Character.Faction
	}
	return model.Card{
		ID:              "starter-hero-taunt-test",
		Name:            "挑衅",
		Type:            model.CardTypeMagic,
		Element:         model.ElementFire,
		Faction:         faction,
		Damage:          0,
		Description:     "勇者专属测试卡",
		ExclusiveChar1:  charName,
		ExclusiveSkill1: "挑衅",
	}
}

func TestHeroHeart_InitialCrystalPlusTwo(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	if p1 == nil {
		t.Fatal("hero player not found")
	}
	if got := p1.Crystal; got != 2 {
		t.Fatalf("expected hero initial crystal=2, got %d", got)
	}
}

func TestHeroRoar_HitDamagePlusTwoAndCleared(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hero_anger"] = 1
	p1.Crystal = 0
	p1.Gem = 0
	p1.Hand = []model.Card{
		heroTestCard("a1", "火焰斩", model.CardTypeAttack, model.ElementFire, 2),
	}
	p2.Hand = nil
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d4", "抽4", model.CardTypeAttack, model.ElementWater, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "hero_roar")
	requireChoicePrompt(t, game, "p1", "hero_roar_draw")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if got := len(p2.Hand); got != 4 {
		t.Fatalf("expected target draw 4 cards from roar-boosted attack, got %d", got)
	}
	if got := p1.Tokens["hero_roar_active"]; got != 0 {
		t.Fatalf("expected hero_roar_active cleared after hit, got %d", got)
	}
	if got := p1.Tokens["hero_roar_damage_pending"]; got != 0 {
		t.Fatalf("expected hero_roar_damage_pending cleared after hit, got %d", got)
	}
	if got := p1.Tokens["hero_wisdom"]; got != 0 {
		t.Fatalf("expected wisdom unchanged on hit branch, got %d", got)
	}
}

func TestHeroRoar_MissAddsWisdom(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hero_anger"] = 1
	p1.Crystal = 0
	p1.Gem = 0
	p1.Hand = []model.Card{
		heroTestCard("a1", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
	}
	p2.Hand = []model.Card{
		heroTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "hero_roar")
	requireChoicePrompt(t, game, "p1", "hero_roar_draw")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	})

	if got := p1.Tokens["hero_wisdom"]; got != 1 {
		t.Fatalf("expected roar miss grant wisdom=1, got %d", got)
	}
	if got := p1.Tokens["hero_roar_active"]; got != 0 {
		t.Fatalf("expected roar active cleared after miss, got %d", got)
	}
	if got := p1.Tokens["hero_roar_damage_pending"]; got != 0 {
		t.Fatalf("expected roar damage marker cleared after miss, got %d", got)
	}
}

func TestHeroForbiddenPower_HitBranch(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		heroTestCard("atk", "起手攻击", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("mf", "火法术", model.CardTypeMagic, model.ElementFire, 0),
		heroTestCard("mw", "水法术", model.CardTypeMagic, model.ElementWater, 0),
		heroTestCard("af", "火攻击", model.CardTypeAttack, model.ElementFire, 2),
	}
	p2.Hand = nil
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d4", "抽4", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d5", "抽5", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d6", "抽6", model.CardTypeAttack, model.ElementEarth, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	chooseResponseSkillByID(t, game, "p1", "hero_forbidden_power")

	if got := p1.Tokens["hero_anger"]; got != 2 {
		t.Fatalf("expected anger +2 from discarded magic cards, got %d", got)
	}
	if got := p1.Crystal; got != 1 {
		t.Fatalf("expected forbidden power consume 1 crystal-like, crystal=%d", got)
	}
	if got := p1.Tokens["hero_exhaustion_form"]; got != 1 {
		t.Fatalf("expected exhaustion form active, got %d", got)
	}
	if got := p1.Tokens["hero_exhaustion_release_pending"]; got != 1 {
		t.Fatalf("expected exhaustion pending flag=1, got %d", got)
	}
	if got := len(p2.Hand); got != 4 {
		t.Fatalf("expected hit branch add fire-count bonus to attack damage (target draw 4), got %d", got)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" && len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected exhaustion grant extra attack action")
	}
}

func TestHeroForbiddenPower_MissBranchWaterToWisdom(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		heroTestCard("atk", "起手攻击", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("mw", "水法术", model.CardTypeMagic, model.ElementWater, 0),
		heroTestCard("aw", "水攻击", model.CardTypeAttack, model.ElementWater, 2),
	}
	p2.Hand = []model.Card{
		heroTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	})
	chooseResponseSkillByID(t, game, "p1", "hero_forbidden_power")

	if got := p1.Tokens["hero_anger"]; got != 1 {
		t.Fatalf("expected anger +1 from discarded magic cards on miss branch, got %d", got)
	}
	if got := p1.Tokens["hero_wisdom"]; got != 2 {
		t.Fatalf("expected wisdom +2 from discarded water cards on miss branch, got %d", got)
	}
	if got := p1.Tokens["hero_exhaustion_form"]; got != 1 {
		t.Fatalf("expected exhaustion form active after forbidden power miss branch, got %d", got)
	}
}

func TestHeroCalmMind_DisablesCounterAndAttackEndGainCrystal(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 0
	p1.Gem = 0
	p1.Tokens["hero_wisdom"] = 4
	p1.Hand = []model.Card{
		heroTestCard("atk", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "hero_calm_mind")

	if len(game.State.CombatStack) == 0 {
		t.Fatalf("expected combat stack after selecting calm mind")
	}
	top := game.State.CombatStack[len(game.State.CombatStack)-1]
	if top.CanBeResponded {
		t.Fatalf("expected calm mind to disable counter response")
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if got := p1.Tokens["hero_wisdom"]; got != 0 {
		t.Fatalf("expected wisdom consumed to 0, got %d", got)
	}
	if got := p1.Crystal; got != 1 {
		t.Fatalf("expected calm mind grant +1 crystal at attack end, got %d", got)
	}
}

func TestHeroTaunt_NonAttackActionSkipsAndRemoves(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hero_anger"] = 1
	p1.ExclusiveCards = []model.Card{heroTauntExclusiveCard(p1)}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	if err := game.UseSkill("p1", "hero_taunt", []string{"p2"}, nil); err != nil {
		t.Fatalf("use hero_taunt failed: %v", err)
	}

	if getFieldEffectCard(p2, model.EffectHeroTaunt) == nil {
		t.Fatalf("expected taunt effect placed on target")
	}

	p1.IsActive = false
	p2.IsActive = true
	p2.TurnState = model.NewPlayerTurnState()
	p2.Hand = []model.Card{
		heroTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 1
	game.State.Phase = model.PhaseActionSelection

	beforeHand := len(p2.Hand)
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdMagic,
		TargetID:  "p1",
		CardIndex: 0,
	})
	if got := len(p2.Hand); got != beforeHand {
		t.Fatalf("expected taunted player non-attack action be skipped without card use, hand %d -> %d", beforeHand, got)
	}
	if getFieldEffectCard(p2, model.EffectHeroTaunt) != nil {
		t.Fatalf("expected taunt effect removed after forced skip")
	}
}

func TestHeroDeadDuel_MagicOverflowMoraleLossFlooredToOne(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.Hand = []model.Card{
		heroTestCard("h1", "卡1", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h2", "卡2", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h3", "卡3", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h4", "卡4", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h5", "卡5", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h6", "卡6", model.CardTypeAttack, model.ElementFire, 2),
	}
	p1.Gem = 1
	p1.Tokens["hero_anger"] = 0
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementWater, 2),
	}

	game.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     3,
		DamageType: "magic",
		Stage:      0,
	})

	if interrupted := game.processPendingDamages(); !interrupted {
		t.Fatalf("expected dead-duel response interrupt")
	}
	chooseResponseSkillByID(t, game, "p1", "hero_dead_duel")

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt after magic overflow, got %+v", game.State.PendingInterrupt)
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1, 2},
	})

	if got := game.State.RedMorale; got != 14 {
		t.Fatalf("expected dead duel floor morale loss to 1 (15->14), got %d", got)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected dead duel consume 1 gem, got %d", got)
	}
	if got := p1.Tokens["hero_anger"]; got != 3 {
		t.Fatalf("expected dead duel add 3 anger, got %d", got)
	}
}
