package sealer_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
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
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
		if c.MatchExclusive(p1.Character.ID, "五系束缚") {
			t.Fatalf("starter five-elements card should not stay in hand")
		}
	}
	if !p1.HasExclusiveCard(p1.Character.ID, "五系束缚") {
		t.Fatalf("expected five-elements starter card in exclusive zone")
	}
}

func TestFiveElementsBind_UsesExclusiveZoneCard(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution
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
			ExclusiveChar1:  "sealer",
			ExclusiveSkill1: "五系束缚",
		},
	}

	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "five_elements_bind",
		TargetIDs: []string{"p2"},
	})

	if p1.Crystal != 0 {
		t.Fatalf("expected crystal consumed by five-elements bind, got %d", p1.Crystal)
	}
	if p1.HasExclusiveCard(p1.Character.ID, "五系束缚") {
		t.Fatalf("expected five-elements exclusive card consumed from exclusive zone")
	}
	if !hasFieldEffect(p2, model.EffectFiveElementsBind) {
		t.Fatalf("expected target to have FiveElementsBind field effect")
	}
}

func TestCrimsonDance_UsesAndReturnsRoseCourtyardExclusiveCard(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionStart
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
			ExclusiveChar1:  "crimson_sword_spirit",
			ExclusiveSkill1: "血蔷薇庭院",
		},
	}

	if err := g.UseSkill("p1", "css_dance", nil, nil); err != nil {
		t.Fatalf("use css_dance failed: %v", err)
	}
	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected css_dance choice interrupt, got %+v", g.State.PendingInterrupt)
	}

	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if p1.HasExclusiveCard(p1.Character.ID, "血蔷薇庭院") {
		t.Fatalf("expected courtyard card moved out of exclusive zone while active")
	}
	if !hasFieldEffect(p1, model.EffectRoseCourtyard) {
		t.Fatalf("expected rose courtyard field card on board")
	}

	g.State.TurnStage = model.TurnStageTurnEnd
	g.Drive()

	if hasFieldEffect(p1, model.EffectRoseCourtyard) {
		t.Fatalf("expected rose courtyard field card removed at turn end")
	}
	if !p1.HasExclusiveCard(p1.Character.ID, "血蔷薇庭院") {
		t.Fatalf("expected courtyard card returned to exclusive zone")
	}
}

func TestCrimsonDanceTurnEnd_DoesNotSynthesizeRoseCourtyardWithoutFieldCard(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.Tokens["css_blood_cap"] = 4
	p1.Tokens["css_blood"] = 4
	p1.Field = append(p1.Field, &model.FieldCard{
		Card: model.Card{
			ID:      "css_test_rose_courtyard",
			Name:    "血蔷薇庭院",
			Type:    model.CardTypeMagic,
			Element: model.ElementDark,
		},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectRoseCourtyard,
	})
	p1.ExclusiveCards = nil

	g.RunTurnEndTimingStageHooks(p1, engine.TimingTurnEndFinal)

	if p1.Tokens["css_blood_cap"] != 3 {
		t.Fatalf("expected blood cap reset to 3, got %d", p1.Tokens["css_blood_cap"])
	}
	if p1.Tokens["css_blood"] != 3 {
		t.Fatalf("expected blood trimmed to 3, got %d", p1.Tokens["css_blood"])
	}
	if p1.HasExclusiveCard(p1.Character.ID, "血蔷薇庭院") {
		t.Fatalf("expected no synthetic courtyard card restored without field card")
	}
}

func TestPreciseShot_NotActivatedByNonOwnerCharacter(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "MG", "magical_girl", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

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
			ExclusiveChar1:  "blade_master",
			ExclusiveChar2:  "archer",
			ExclusiveSkill1: "列风技",
			ExclusiveSkill2: "精准射击",
		},
	}

	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, g, "p1", 0),
	})

	if got := len(g.State.CombatStack); got != 1 {
		t.Fatalf("expected one combat request, got %d", got)
	}
	req := g.State.CombatStack[len(g.State.CombatStack)-1]
	if req.IsForcedHit {
		t.Fatalf("expected non-owner card not to dispatch precise shot forced hit")
	}
}

func TestPreciseShot_ModifyDamage_ReducesOwnerAttack(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	p1.Hand = []model.Card{
		{
			ID:              "atk-2",
			Name:            "雷光斩",
			Type:            model.CardTypeAttack,
			Element:         model.ElementThunder,
			Damage:          2,
			Faction:         "技",
			ExclusiveChar1:  "blade_master",
			ExclusiveChar2:  "archer",
			ExclusiveSkill1: "列风技",
			ExclusiveSkill2: "精准射击",
		},
	}

	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, g, "p1", 0),
	})

	// 精准射击为可选响应，确认发动后强制命中且伤害-1
	testutils.ChooseResponseSkillByID(t, g, "p1", "precise_shot")
	g.Drive()

	// 伤害 = 2 - 1 = 1，目标摸1张牌
	if len(p2.Hand) != 1 {
		t.Fatalf("expected precise shot to reduce owner attack damage to 1 (draw 1), got hand=%d", len(p2.Hand))
	}
}

func TestPreciseShot_NonOwnerCard_DoesNotReduceDamage(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "MG", "magical_girl", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()

	card := model.Card{
		ID:              "atk-3",
		Name:            "雷光斩",
		Type:            model.CardTypeAttack,
		Element:         model.ElementThunder,
		Damage:          2,
		Faction:         "技",
		ExclusiveChar1:  "blade_master",
		ExclusiveChar2:  "archer",
		ExclusiveSkill1: "列风技",
		ExclusiveSkill2: "精准射击",
	}

	damage := g.ApplyAttackDamageModifiers(p1, p2, 2, model.Action{
		SourceID: p1.ID,
		TargetID: p2.ID,
		Type:     model.ActionAttack,
		Card:     &card,
	})
	if damage != 2 {
		t.Fatalf("expected non-owner precise-shot card not to reduce damage, got %d", damage)
	}
}
