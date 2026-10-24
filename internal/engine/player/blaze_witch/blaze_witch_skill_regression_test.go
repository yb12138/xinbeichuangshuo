package blaze_witch_test

import (
	"fmt"
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"strings"
	"testing"

	"starcup-engine/internal/data"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func makeBlazeWitchTestCards(n int) []model.Card {
	elements := []model.Element{
		model.ElementFire,
		model.ElementWater,
		model.ElementWind,
		model.ElementThunder,
		model.ElementEarth,
		model.ElementDark,
		model.ElementLight,
	}
	cards := make([]model.Card, 0, n)
	for i := range n {
		cardType := model.CardTypeAttack
		if i%2 == 0 {
			cardType = model.CardTypeMagic
		}
		cards = append(cards, model.Card{
			ID:      fmt.Sprintf("bw_test_%d", i),
			Name:    fmt.Sprintf("测试牌%d", i+1),
			Type:    cardType,
			Element: elements[i%len(elements)],
			Faction: "血",
			Damage:  2,
		})
	}
	return cards
}

func TestBlazeWitchPainLink_ConsumesCrystalOnceAndQueuesDiscardToThree(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Hand = makeBlazeWitchTestCards(5)

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "bw_pain_link", []string{"p2"}, nil); err != nil {
		t.Fatalf("use pain link failed: %v", err)
	}
	if p1.Crystal != 0 || p1.Gem != 0 {
		t.Fatalf("expected consume exactly 1 crystal-like resource, got crystal=%d gem=%d", p1.Crystal, p1.Gem)
	}
	if got := len(game.State.PendingDamageQueue); got != 2 {
		t.Fatalf("expected 2 pending damages, got %d", got)
	}

	game.State.CombatStage = model.CombatStageCalcDamage
	for i := 0; i < 16 && game.State.PendingInterrupt == nil && len(game.State.PendingDamageQueue) > 0; i++ {
		game.ProcessPendingDamages()
	}

	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected discard interrupt from pain link")
	}
	if !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("unexpected interrupt: %+v", game.State.PendingInterrupt)
	}
	data, _ := game.State.PendingInterrupt.Context.(map[string]any)
	downTo, _ := data["discard_down_to"].(int)
	if downTo != 3 {
		t.Fatalf("expected discard_down_to=3, got %d", downTo)
	}
}

func TestBlazeWitchManaInversion_UsesPromptFlow(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Hand = []model.Card{
		{ID: "m1", Name: "烈焰术", Type: model.CardTypeMagic, Element: model.ElementFire, Faction: "血"},
		{ID: "m2", Name: "暗影术", Type: model.CardTypeMagic, Element: model.ElementDark, Faction: "血"},
		{ID: "a1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "血", Damage: 2},
	}
	p2.Hand = nil

	game.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	if paused := game.ProcessPendingDamages(); !paused {
		t.Fatalf("expected mana inversion response prompt")
	}
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if err := game.ConfirmResponseSkill("p1", "bw_mana_inversion"); err != nil {
		t.Fatalf("confirm mana inversion failed: %v", err)
	}
	ctxData := testutils.RequireChoiceContext(t, game, "p1", "bw_mana_inversion_cards")
	testutils.RequirePromptFlow(t, ctxData, "bw_mana_inversion", "cards")

	prompt := game.BuildChoicePrompt()
	if prompt == nil {
		t.Fatalf("expected mana inversion card prompt")
	}
	if prompt.Message == "【魔能反转】请选择X值：" {
		t.Fatalf("mana inversion should not ask for numeric X before discarding cards")
	}
	if prompt.Min != 2 || prompt.Max != 2 {
		t.Fatalf("expected card picker min/max 2, got min=%d max=%d", prompt.Min, prompt.Max)
	}

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1}}); err != nil {
		t.Fatalf("choose mana inversion cards failed: %v", err)
	}
	ctxData = testutils.RequireChoiceContext(t, game, "p1", "bw_mana_inversion_target")
	flow := testutils.RequirePromptFlow(t, ctxData, "bw_mana_inversion", "target")
	if got := flow.Selection("cards").OptionIndexes; len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("expected mana inversion flow to accumulate magic card indexes [0 1], got %+v in %+v", got, flow)
	}
	if got := flow.Selection("cards").Count; got != 2 {
		t.Fatalf("expected mana inversion X to equal selected card count 2, got %d", got)
	}

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose mana inversion target failed: %v", err)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected mana inversion to discard 2 magic cards, got hand=%d", got)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected mana inversion to queue target damage")
	}
}

func TestBlazeWitchHeavenfireCleave_AllowsNonFireAttackDiscardInFlameForm(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormBlazeWitchFlame
	p1.Tokens["bw_rebirth"] = 1
	p1.Hand = []model.Card{
		{ID: "a1", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Faction: "血", Damage: 2},
		{ID: "a2", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Faction: "血", Damage: 2},
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "bw_heavenfire_cleave", []string{"p2"}, []int{0, 1}); err != nil {
		t.Fatalf("heavenfire cleave should accept transformed fire discards in flame form, got: %v", err)
	}
	if got := p1.Tokens["bw_rebirth"]; got != 1 {
		t.Fatalf("expected rebirth not consumed in flame form, got %d", got)
	}
	if got := len(game.State.PendingDamageQueue); got != 2 {
		t.Fatalf("expected 2 pending damages, got %d", got)
	}
	if game.State.PendingDamageQueue[0].Damage != 3 || game.State.PendingDamageQueue[1].Damage != 3 {
		t.Fatalf("expected heavenfire base damage 3 when morale not behind, got %+v", game.State.PendingDamageQueue)
	}
}

func TestBlazeWitchFlameForm_DiscardPromptShowsEffectiveFireAttackCards(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormBlazeWitchFlame
	p1.Tokens["bw_rebirth"] = 1
	p1.Hand = []model.Card{
		{ID: "wind", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Faction: "血", Damage: 2},
		{ID: "thunder", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Faction: "血", Damage: 2},
		{ID: "water", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Faction: "血", Damage: 2},
		{ID: "dark", Name: "暗灭", Type: model.CardTypeAttack, Element: model.ElementDark, Faction: "血", Damage: 2},
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "bw_heavenfire_cleave", []string{"p2"}, nil); err != nil {
		t.Fatalf("request heavenfire discard prompt failed: %v", err)
	}
	prompt := game.BuildChoicePrompt()
	if prompt == nil {
		t.Fatalf("expected discard prompt")
	}
	gotIDs := map[string]string{}
	for _, option := range prompt.Options {
		gotIDs[option.CardID] = option.Label
	}
	if len(gotIDs) != 2 || gotIDs["wind"] == "" || gotIDs["thunder"] == "" {
		t.Fatalf("expected only wind/thunder attack cards to be selectable as effective fire, got %+v", gotIDs)
	}
	if !strings.Contains(gotIDs["wind"], "视为火系") || !strings.Contains(gotIDs["thunder"], "视为火系") {
		t.Fatalf("expected effective fire hint in option labels, got %+v", gotIDs)
	}
}

func TestBlazeWitchBlazingCodex_CostThenTargetPrompt(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "血", Damage: 2},
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "bw_blazing_codex", nil, nil); err != nil {
		t.Fatalf("request blazing codex discard prompt failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "system_discard_cards")

	if err := game.ConfirmDiscard("p1", []int{0}); err != nil {
		t.Fatalf("confirm blazing codex fire discard failed: %v", err)
	}
	ctxData := testutils.RequireChoiceContext(t, game, "p1", "bw_blazing_codex_target")
	targetIDs, _ := ctxData["target_ids"].([]string)
	if len(targetIDs) != 1 || targetIDs[0] != "p2" {
		t.Fatalf("expected target prompt for p2, got %+v", ctxData["target_ids"])
	}
	if game.State.Subflow != model.SubflowNone {
		t.Fatalf("expected discard subflow restored before target prompt, got %s", game.State.Subflow)
	}

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose blazing codex target failed: %v", err)
	}
	if got := len(p1.Hand); got != 0 {
		t.Fatalf("expected blazing codex to discard the selected fire card, got hand=%d", got)
	}
	if got := len(game.State.DiscardPile); got != 1 || game.State.DiscardPile[0].ID != "f1" {
		t.Fatalf("expected selected fire card in discard pile, got %+v", game.State.DiscardPile)
	}
	if got := len(game.State.PendingDamageQueue); got != 2 {
		t.Fatalf("expected target and self pending damage, got %d", got)
	}
	if game.State.PendingDamageQueue[0].TargetID != "p2" || game.State.PendingDamageQueue[1].TargetID != "p1" {
		t.Fatalf("expected damage order target then self, got %+v", game.State.PendingDamageQueue)
	}
}

func TestBlazeWitchRebirthClock_IncreasesOnMagicMoraleLossWithCap(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.Hand = makeBlazeWitchTestCards(8)

	damageOverflowCtx := &model.Context{
		Flags: map[string]bool{
			"FromDamageDraw": true,
			"IsMagicDamage":  true,
		},
	}
	game.CheckHandLimitCtx(p1, damageOverflowCtx)
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt, got %+v", game.State.PendingInterrupt)
	}
	if err := game.ConfirmDiscard("p1", []int{0, 1}); err != nil {
		t.Fatalf("confirm discard failed: %v", err)
	}
	if got := p1.Tokens["bw_rebirth"]; got != 1 {
		t.Fatalf("expected rebirth +1 after magic morale loss, got %d", got)
	}

	p1.Tokens["bw_rebirth"] = 4
	p1.Hand = makeBlazeWitchTestCards(8)
	game.CheckHandLimitCtx(p1, damageOverflowCtx)
	if err := game.ConfirmDiscard("p1", []int{0, 1}); err != nil {
		t.Fatalf("confirm discard at cap failed: %v", err)
	}
	if got := p1.Tokens["bw_rebirth"]; got != 4 {
		t.Fatalf("expected rebirth capped at 4, got %d", got)
	}
}

func TestBlazeWitchFlameForm_ReleasesAtTurnStart(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormBlazeWitchFlame
	p1.TurnState.SkillFlowState["bw_flame_release_pending"] = 1
	p1.Hand = makeBlazeWitchTestCards(5)

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnStart

	game.Drive()

	if got := p1.Form; got != "" {
		t.Fatalf("expected flame form released at turn start, got %q", got)
	}
	if got := p1.TurnState.SkillFlowState["bw_flame_release_pending"]; got != 0 {
		t.Fatalf("expected flame release flag cleared, got %d", got)
	}
	if got := game.State.TurnStage; got != model.TurnStageActionExecution {
		t.Fatalf("expected drive to continue into action execution after release, got %s", got)
	}
}

func TestBlazeWitchGetMaxHand_DynamicByRebirthInFlameForm(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	if got := game.GetMaxHand(p1); got != 6 {
		t.Fatalf("expected base max hand 6, got %d", got)
	}

	p1.Form = model.FormBlazeWitchFlame
	p1.Tokens["bw_rebirth"] = 0
	if got := game.GetMaxHand(p1); got != 4 {
		t.Fatalf("expected max hand 4 when rebirth=0 in flame form, got %d", got)
	}
	p1.Tokens["bw_rebirth"] = 1
	if got := game.GetMaxHand(p1); got != 5 {
		t.Fatalf("expected max hand 5 when rebirth=1 in flame form, got %d", got)
	}
	p1.Tokens["bw_rebirth"] = 3
	if got := game.GetMaxHand(p1); got != 7 {
		t.Fatalf("expected max hand 7 when rebirth=3 in flame form, got %d", got)
	}
}

func TestBlazeWitchCodexAndHeavenfire_RejectSelfTarget(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["bw_rebirth"] = 1
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "血", Damage: 2},
		{ID: "f2", Name: "烈焰术", Type: model.CardTypeMagic, Element: model.ElementFire, Faction: "血", Damage: 2},
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "bw_blazing_codex", []string{"p1"}, []int{0}); err == nil {
		t.Fatalf("expected blazing codex reject self target")
	}

	if err := game.UseSkill("p1", "bw_heavenfire_cleave", []string{"p1"}, []int{0, 1}); err == nil {
		t.Fatalf("expected heavenfire cleave reject self target")
	}
}

func TestBlazeWitchFlameForm_AttackUsesPreparedTransformedCard(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Sealer", "sealer", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormBlazeWitchFlame
	p1.Hand = []model.Card{{ID: "atk-wind", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Faction: "血", Damage: 2}}
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "seal-fire", Name: "火之封印", Type: model.CardTypeMagic, Element: model.ElementFire},
		OwnerID:  p1.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealFire,
		Hook:     model.FieldHookOnCardPlayedOrRevealed,
	})

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0)})

	if got := testutils.CountFieldEffect(p1, model.EffectSealFire); got != 0 {
		t.Fatalf("expected transformed fire attack to consume fire seal, got %d", got)
	}
	if len(game.State.CombatStack) == 0 {
		t.Fatalf("expected combat request queued after attack")
	}
	combatReq := game.State.CombatStack[len(game.State.CombatStack)-1]
	if combatReq.Card == nil || combatReq.Card.Element != model.ElementFire {
		t.Fatalf("expected combat card element Fire after flame-form transform, got %+v", combatReq.Card)
	}
}

func TestBlazeWitchConfig_MetadataAlignsWithDocument(t *testing.T) {
	characters := data.GetCharacters()
	var blaze *model.Character
	for _, character := range characters {
		if character.ID == "blaze_witch" {
			copy := character
			blaze = &copy
			break
		}
	}
	if blaze == nil {
		t.Fatalf("blaze_witch character not found")
	}

	var substitute *model.SkillDefinition
	var manaInversion *model.SkillDefinition
	for i := range blaze.Skills {
		switch blaze.Skills[i].ID {
		case "bw_substitute_doll":
			substitute = &blaze.Skills[i]
		case "bw_mana_inversion":
			manaInversion = &blaze.Skills[i]
		}
	}
	if substitute == nil || manaInversion == nil {
		t.Fatalf("expected blaze witch substitute doll and mana inversion skills")
	}

	if substitute.CostDiscards != 1 || substitute.DiscardType != model.CardTypeMagic {
		t.Fatalf("expected substitute doll metadata to require 1 magic discard, got cost=%d discardType=%s", substitute.CostDiscards, substitute.DiscardType)
	}
	if substitute.TargetType != model.TargetAlly || substitute.MinTargets != 1 || substitute.MaxTargets != 1 {
		t.Fatalf("expected substitute doll target metadata ally(1), got type=%v min=%d max=%d", substitute.TargetType, substitute.MinTargets, substitute.MaxTargets)
	}

	if manaInversion.CostCrystal != 1 {
		t.Fatalf("expected mana inversion metadata crystal cost=1, got %d", manaInversion.CostCrystal)
	}
	if manaInversion.TargetType != model.TargetEnemy || manaInversion.MinTargets != 1 || manaInversion.MaxTargets != 1 {
		t.Fatalf("expected mana inversion target metadata enemy(1), got type=%v min=%d max=%d", manaInversion.TargetType, manaInversion.MinTargets, manaInversion.MaxTargets)
	}
}
