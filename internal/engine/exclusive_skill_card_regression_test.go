package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func hasFieldEffect(player *model.Player, effect model.EffectType) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == effect {
			return true
		}
	}
	return false
}

func TestSealerStarterExclusiveCard_NotInHand(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	if err := g.StartGame(); err != nil {
		t.Fatalf("start game failed: %v", err)
	}

	p1 := g.State.Players["p1"]
	if p1 == nil || p1.Character == nil {
		t.Fatalf("sealer player not initialized")
	}
	if len(p1.Hand) != 4 {
		t.Fatalf("expected sealer opening hand=4 (starter card not in hand), got %d", len(p1.Hand))
	}
	for _, c := range p1.Hand {
		if c.MatchExclusive(p1.Character.Name, "五系束缚") {
			t.Fatalf("starter five-elements card should not stay in hand")
		}
	}
	if !p1.HasExclusiveCard(p1.Character.Name, "五系束缚") {
		t.Fatalf("expected five-elements starter card in exclusive zone")
	}
}

func TestFiveElementsBind_UsesExclusiveZoneCard(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseActionSelection
	g.State.Deck = rules.InitDeck()

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.ExclusiveCards = []model.Card{
		{
			ID:              "starter-p1-five_elements_bind",
			Name:            "五系束缚",
			Type:            model.CardTypeMagic,
			Element:         model.ElementLight,
			ExclusiveChar1:  "封印师",
			ExclusiveSkill1: "五系束缚",
		},
	}

	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "five_elements_bind",
		TargetIDs: []string{"p2"},
	})

	if p1.Crystal != 0 {
		t.Fatalf("expected crystal consumed by five-elements bind, got %d", p1.Crystal)
	}
	if p1.HasExclusiveCard(p1.Character.Name, "五系束缚") {
		t.Fatalf("expected five-elements exclusive card consumed from exclusive zone")
	}
	if !hasFieldEffect(p2, model.EffectFiveElementsBind) {
		t.Fatalf("expected target to have FiveElementsBind field effect")
	}
}

func TestCrimsonDance_UsesAndReturnsRoseCourtyardExclusiveCard(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseStartup
	g.State.Deck = rules.InitDeck()

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.ExclusiveCards = []model.Card{
		{
			ID:              "starter-p1-css_rose_courtyard",
			Name:            "血蔷薇庭院",
			Type:            model.CardTypeMagic,
			Element:         model.ElementDark,
			ExclusiveChar1:  "血色剑灵",
			ExclusiveSkill1: "血蔷薇庭院",
		},
	}

	if err := g.UseSkill("p1", "css_dance", nil, nil); err != nil {
		t.Fatalf("use css_dance failed: %v", err)
	}
	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected css_dance choice interrupt, got %+v", g.State.PendingInterrupt)
	}

	mustDo(t, g, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if p1.Tokens["css_rose_courtyard_active"] != 1 {
		t.Fatalf("expected courtyard active after dance")
	}
	if p1.HasExclusiveCard(p1.Character.Name, "血蔷薇庭院") {
		t.Fatalf("expected courtyard card moved out of exclusive zone while active")
	}
	if !hasFieldEffect(p1, model.EffectRoseCourtyard) {
		t.Fatalf("expected rose courtyard field card on board")
	}

	g.State.Phase = model.PhaseTurnEnd
	g.Drive()

	if p1.Tokens["css_rose_courtyard_active"] != 0 {
		t.Fatalf("expected courtyard inactive after turn end")
	}
	if hasFieldEffect(p1, model.EffectRoseCourtyard) {
		t.Fatalf("expected rose courtyard field card removed at turn end")
	}
	if !p1.HasExclusiveCard(p1.Character.Name, "血蔷薇庭院") {
		t.Fatalf("expected courtyard card returned to exclusive zone")
	}
}

func TestPreciseShot_NotTriggeredByNonOwnerCharacter(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "MG", "magical_girl", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseActionSelection

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{
			ID:              "atk-1",
			Name:            "风神斩",
			Type:            model.CardTypeAttack,
			Element:         model.ElementWind,
			Damage:          2,
			Faction:         "技",
			ExclusiveChar1:  "风之剑圣",
			ExclusiveChar2:  "神箭手",
			ExclusiveSkill1: "烈风技",
			ExclusiveSkill2: "精准射击",
		},
	}

	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})

	if got := len(g.State.CombatStack); got != 1 {
		t.Fatalf("expected one combat request, got %d", got)
	}
	req := g.State.CombatStack[len(g.State.CombatStack)-1]
	if req.IsForcedHit {
		t.Fatalf("expected non-owner card not to trigger precise shot forced hit")
	}
	if p1.TurnState.PreciseShotActive {
		t.Fatalf("expected precise shot flag remain false for non-owner character")
	}
}

func TestPreciseShotFlagMismatchCard_DoesNotReduceDamage(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "MG", "magical_girl", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.PreciseShotActive = true

	card := model.Card{
		ID:              "atk-2",
		Name:            "雷光斩",
		Type:            model.CardTypeAttack,
		Element:         model.ElementThunder,
		Damage:          2,
		Faction:         "技",
		ExclusiveChar1:  "风之剑圣",
		ExclusiveChar2:  "神箭手",
		ExclusiveSkill1: "烈风技",
		ExclusiveSkill2: "精准射击",
	}

	damage := g.applyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID: p1.ID,
		TargetID: p2.ID,
		Type:     model.ActionAttack,
		Card:     &card,
	})
	if damage != 2 {
		t.Fatalf("expected mismatch precise shot flag not to reduce damage, got %d", damage)
	}
	if p1.TurnState.PreciseShotActive {
		t.Fatalf("expected mismatch precise shot flag to be cleared")
	}
}
