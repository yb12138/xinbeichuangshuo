package sword_emperor_test

import (
	"starcup-engine/internal/data"
	"starcup-engine/internal/engine"
	skillregistry "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/testutils"
	"testing"

	swordemperor "starcup-engine/internal/engine/player/sword_emperor"
	"starcup-engine/internal/model"
)

func swordEmperorTestCard(id, name string, cardType model.CardType, element model.Element, damage int) model.Card {
	if damage <= 0 && cardType == model.CardTypeAttack {
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

func giveSwordEmperorSwordSoul(t *testing.T, game *engine.GameEngine, playerID, cardID string) {
	t.Helper()
	player := game.State.Players[playerID]
	if player == nil {
		t.Fatalf("player %s not found", playerID)
	}
	card := swordEmperorTestCard(cardID, "剑魂牌", model.CardTypeAttack, model.ElementFire, 2)
	if !swordemperor.PlaceSwordEmperorSwordSoul(player, card) {
		t.Fatalf("failed to place sword soul for %s", playerID)
	}
}

func prepareSwordEmperorActiveTurn(game *engine.GameEngine, playerID string, hand []model.Card) *model.Player {
	player := game.State.Players[playerID]
	player.IsActive = true
	player.TurnState = model.NewPlayerTurnState()
	player.Hand = append([]model.Card{}, hand...)
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	return player
}

func TestSwordEmperor_InitTokens(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	if got := p1.Tokens["se_sword_qi"]; got != 0 {
		t.Fatalf("expected se_sword_qi=0, got %d", got)
	}
	if got := swordemperor.SwordSoulCount(p1); got != 0 {
		t.Fatalf("expected se_sword_soul_count=0, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"]; got != 0 {
		t.Fatalf("expected guard disable token cleared, got %d", got)
	}
}

func TestSwordEmperor_SoulSettlementBranchesAreTimingHooksNotRegisteredSkills(t *testing.T) {
	settlementIDs := map[string]bool{
		"se_angel_soul_hit":  true,
		"se_angel_soul_miss": true,
		"se_demon_soul_miss": true,
	}

	for _, entry := range swordemperor.SkillEntries() {
		if settlementIDs[entry.ID] {
			t.Fatalf("settlement branch %q should be a timing hook, not a registered skill", entry.ID)
		}
	}
	for id := range settlementIDs {
		if handler := skillregistry.GetHandler(id); handler != nil {
			t.Fatalf("settlement branch %q should not have a registered skill handler", id)
		}
	}

	var swordEmperor *model.Character
	for _, character := range data.GetCharacters() {
		if character.ID == "sword_emperor" {
			c := character
			swordEmperor = &c
			break
		}
	}
	if swordEmperor == nil {
		t.Fatal("sword_emperor character config not found")
	}
	for _, skill := range swordEmperor.Skills {
		if settlementIDs[skill.ID] {
			t.Fatalf("settlement branch %q should not be declared as a character skill", skill.ID)
		}
	}
}

func TestSwordEmperor_MissAddsSwordSoulAndSwordQi(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := prepareSwordEmperorActiveTurn(game, "p1", []model.Card{
		swordEmperorTestCard("atk1", "炎斩", model.CardTypeAttack, model.ElementFire, 2),
	})
	p2 := game.State.Players["p2"]
	p2.Hand = []model.Card{
		swordEmperorTestCard("def1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
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
		CardID:    testutils.PlayableCardID(t, game, "p2", 0),
		ExtraArgs: []string{"defend"},
	})

	if got := swordemperor.SwordSoulCount(p1); got != 1 {
		t.Fatalf("expected 1 sword soul after miss, got %d", got)
	}
	if got := swordemperor.SwordSoulCount(p1); got != 1 {
		t.Fatalf("expected sword soul token synced to 1, got %d", got)
	}
	if got := p1.Tokens["se_sword_qi"]; got != 1 {
		t.Fatalf("expected sword qi +1 after feint, got %d", got)
	}
	if len(game.State.DiscardPile) != 1 || game.State.DiscardPile[0].ID != "def1" {
		t.Fatalf("expected only defend card remain in discard, got %+v", game.State.DiscardPile)
	}
}

func TestSwordEmperor_SwordSoulGuard_StopsAtCap(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		giveSwordEmperorSwordSoul(t, game, "p1", "soul_cap_"+string(rune('a'+i)))
	}
	p1 := game.State.Players["p1"]
	attackCard := swordEmperorTestCard("atk_cap", "上限测试斩", model.CardTypeAttack, model.ElementFire, 2)
	game.State.DiscardPile = append(game.State.DiscardPile, attackCard)

	swordemperor.ResolveAttackMiss(engine.NewRoleChoiceRuntime(game), "p1", &attackCard, false)

	if got := swordemperor.SwordSoulCount(p1); got != 3 {
		t.Fatalf("expected sword soul stay at cap 3, got %d", got)
	}
	if got := swordemperor.SwordSoulCount(p1); got != 3 {
		t.Fatalf("expected sword soul token stay at 3, got %d", got)
	}
	if got := p1.Tokens["se_sword_qi"]; got != 1 {
		t.Fatalf("expected feint still add 1 sword qi at cap, got %d", got)
	}
	if len(game.State.DiscardPile) == 0 || game.State.DiscardPile[len(game.State.DiscardPile)-1].ID != "atk_cap" {
		t.Fatalf("expected attack card remain in discard when cap reached, got %+v", game.State.DiscardPile)
	}
}

func TestSwordEmperor_AngelSoul_MissDisablesGuardAndAddsMorale(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := prepareSwordEmperorActiveTurn(game, "p1", []model.Card{
		swordEmperorTestCard("atk1", "圣断", model.CardTypeAttack, model.ElementLight, 2),
	})
	p1.Crystal = 1
	giveSwordEmperorSwordSoul(t, game, "p1", "soul1")
	game.State.RedMorale = 14

	p2 := game.State.Players["p2"]
	p2.Hand = []model.Card{
		swordEmperorTestCard("def1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.ChooseResponseSkillByID(t, game, "p1", "se_angel_soul")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardID:    testutils.PlayableCardID(t, game, "p2", 0),
		ExtraArgs: []string{"defend"},
	})

	if got := swordemperor.SwordSoulCount(p1); got != 0 {
		t.Fatalf("expected sword soul consumed and guard disabled, got %d", got)
	}
	if got := p1.Tokens["se_sword_qi"]; got != 1 {
		t.Fatalf("expected feint add 1 sword qi on miss, got %d", got)
	}
	if got := game.State.RedMorale; got != 15 {
		t.Fatalf("expected red morale 14 -> 15, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"]; got != 0 {
		t.Fatalf("expected guard disable token cleared after miss, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["se_angel_soul_armed"]; got != 0 {
		t.Fatalf("expected angel soul armed token cleared after miss, got %d", got)
	}
}

func TestSwordEmperor_AngelSoul_HitHealsTwo(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := prepareSwordEmperorActiveTurn(game, "p1", []model.Card{
		swordEmperorTestCard("atk1", "光斩", model.CardTypeAttack, model.ElementLight, 2),
	})
	p1.Crystal = 1
	p1.Heal = 0
	giveSwordEmperorSwordSoul(t, game, "p1", "soul1")
	game.State.Deck = []model.Card{
		swordEmperorTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		swordEmperorTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.ChooseResponseSkillByID(t, game, "p1", "se_angel_soul")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if got := p1.Heal; got != 2 {
		t.Fatalf("expected angel soul hit heal=2, got %d", got)
	}
	if got := swordemperor.SwordSoulCount(p1); got != 0 {
		t.Fatalf("expected sword soul consumed on angel soul hit, got %d", got)
	}
	if got := len(game.State.Players["p2"].Hand); got != 2 {
		t.Fatalf("expected target draw 2 cards from attack damage, got %d", got)
	}
}

func TestSwordEmperor_DemonSoul_HitAddsDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := prepareSwordEmperorActiveTurn(game, "p1", []model.Card{
		swordEmperorTestCard("atk1", "暗斩", model.CardTypeAttack, model.ElementDark, 2),
	})
	p1.Crystal = 2
	giveSwordEmperorSwordSoul(t, game, "p1", "soul1")
	game.State.Deck = []model.Card{
		swordEmperorTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		swordEmperorTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		swordEmperorTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder, 2),
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.ChooseResponseSkillByID(t, game, "p1", "se_demon_soul")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if got := len(game.State.Players["p2"].Hand); got != 3 {
		t.Fatalf("expected demon soul hit deal 3 damage, got target draw=%d", got)
	}
}

func TestSwordEmperor_DemonSoul_MissAddsTwoSwordQi(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := prepareSwordEmperorActiveTurn(game, "p1", []model.Card{
		swordEmperorTestCard("atk1", "暗月斩", model.CardTypeAttack, model.ElementDark, 2),
	})
	p1.Crystal = 2
	giveSwordEmperorSwordSoul(t, game, "p1", "soul1")

	p2 := game.State.Players["p2"]
	p2.Hand = []model.Card{
		swordEmperorTestCard("def1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.ChooseResponseSkillByID(t, game, "p1", "se_demon_soul")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardID:    testutils.PlayableCardID(t, game, "p2", 0),
		ExtraArgs: []string{"defend"},
	})

	if got := swordemperor.SwordSoulCount(p1); got != 0 {
		t.Fatalf("expected guard remain disabled after demon soul miss, got sword soul=%d", got)
	}
	if got := p1.Tokens["se_sword_qi"]; got != 3 {
		t.Fatalf("expected miss grant feint+1 and demon+2 qi, got %d", got)
	}
}

func TestSwordEmperor_SwordQiSlash_ExcludeOriginalTargetAndDealMagicDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := prepareSwordEmperorActiveTurn(game, "p1", []model.Card{
		swordEmperorTestCard("atk1", "剑压", model.CardTypeAttack, model.ElementWind, 2),
	})
	p1.Tokens["se_sword_qi"] = 3
	game.State.Deck = []model.Card{
		swordEmperorTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		swordEmperorTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		swordEmperorTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder, 2),
		swordEmperorTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementEarth, 2),
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

	testutils.ChooseResponseSkillByID(t, game, "p1", "se_sword_qi_slash")
	ctxData := testutils.RequireChoiceContext(t, game, "p1", "se_sword_qi_slash_x")
	flow := testutils.RequirePromptFlow(t, ctxData, "se_sword_qi_slash", "x")
	targetIDs, ok := ctxData["target_ids"].([]string)
	if !ok {
		t.Fatalf("expected target_ids in sword qi slash context, got %+v", ctxData["target_ids"])
	}
	for _, tid := range targetIDs {
		if tid == "p2" {
			t.Fatalf("expected original target p2 excluded from sword qi slash targets, got %+v", targetIDs)
		}
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1}, // 选择第二个选项 => X=2
	})
	prompt := game.GetCurrentPrompt()
	if prompt == nil || prompt.Presentation == nil || prompt.Presentation.Kind != model.PresentationTargetPicker {
		t.Fatalf("expected sword qi slash target prompt to use target_picker presentation, got %+v", prompt)
	}
	for _, option := range prompt.Options {
		if option.TargetID == "" {
			t.Fatalf("expected sword qi slash target option %q to carry target_id, got %+v", option.ID, prompt.Options)
		}
	}
	ctxData = testutils.RequireChoiceContext(t, game, "p1", "se_sword_qi_slash_target")
	flow = testutils.RequirePromptFlow(t, ctxData, "se_sword_qi_slash", "target")
	if got := flow.Selection("x").Count; got != 2 {
		t.Fatalf("expected sword qi slash flow to accumulate x=2, got %d in %+v", got, flow)
	}
	targetIDs, ok = ctxData["target_ids"].([]string)
	if !ok {
		t.Fatalf("expected target_ids in sword qi slash target context, got %+v", ctxData["target_ids"])
	}
	targetSelection := -1
	for i, tid := range targetIDs {
		if tid == "p3" {
			targetSelection = i
			break
		}
	}
	if targetSelection < 0 {
		t.Fatalf("expected p3 available for sword qi slash target, got %+v", targetIDs)
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{targetSelection},
	})

	if got := p1.Tokens["se_sword_qi"]; got != 1 {
		t.Fatalf("expected sword qi 3 -> 1 after X=2 slash, got %d", got)
	}
	if got := len(game.State.Players["p2"].Hand); got != 2 {
		t.Fatalf("expected original target take 2 attack damage, got %d", got)
	}
	if got := len(game.State.Players["p3"].Hand); got != 2 {
		t.Fatalf("expected extra target take 2 magic damage, got %d", got)
	}
}

func TestSwordEmperor_SwordRainTargetPromptCarriesTargetIDs(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: "p1",
		Context: map[string]interface{}{
			"choice_type": "se_sword_rain_target",
			"user_id":     "p1",
			"target_ids":  []string{"p2"},
		},
	})

	prompt := game.GetCurrentPrompt()
	if prompt == nil || prompt.Presentation == nil || prompt.Presentation.Kind != model.PresentationTargetPicker {
		t.Fatalf("expected sword rain target prompt to use target_picker presentation, got %+v", prompt)
	}
	if len(prompt.Options) != 1 || prompt.Options[0].TargetID != "p2" {
		t.Fatalf("expected sword rain target option to carry target_id p2, got %+v", prompt.Options)
	}
}

func TestSwordEmperor_IndomitableWill_DrawQiAndExtraAttack(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.Crystal = 1
	p1.TurnState = model.NewPlayerTurnState()
	game.State.Deck = []model.Card{
		swordEmperorTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
	}

	ctx := game.BuildContext(p1, nil, model.TimingOnActionEnd, &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   "p1",
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	})
	game.Dispatcher().OnTiming(ctx.Timing, ctx)

	testutils.ChooseResponseSkillByID(t, game, "p1", "se_indomitable_will")

	if got := p1.Crystal; got != 0 {
		t.Fatalf("expected crystal spent by indomitable will, got %d", got)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected draw 1 card from indomitable will, got %d", got)
	}
	if got := p1.Tokens["se_sword_qi"]; got != 1 {
		t.Fatalf("expected sword qi +1 from indomitable will, got %d", got)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" && len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected indomitable will grant extra attack action, current=%q pending=%d", p1.TurnState.CurrentExtraAction, len(p1.TurnState.PendingActions))
	}
	if len(p1.TurnState.PendingActions) > 0 && p1.TurnState.PendingActions[0].MustType != "Attack" {
		t.Fatalf("expected pending extra action type Attack, got %s", p1.TurnState.PendingActions[0].MustType)
	}
}
