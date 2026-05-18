package beast_samurai_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func newBeastSamuraiTestEngine(t *testing.T, observer model.GameObserver, enemyRole string) (*engine.GameEngine, *model.Player, *model.Player) {
	t.Helper()
	if enemyRole == "" {
		enemyRole = "berserker"
	}
	game := engine.NewGameEngine(observer)
	if err := game.AddPlayer("p1", "Beast", "beast_samurai", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", enemyRole, model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	return game, p1, p2
}

func requireBeastSamuraiDiscardInterrupt(t *testing.T, game *engine.GameEngine, playerID, choiceType string) {
	t.Helper()
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending discard interrupt, got nil")
	}
	if !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt, got %s", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected discard interrupt for %s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
	data, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("discard context type mismatch")
	}
	got, _ := data["choice_type"].(string)
	if got != choiceType {
		t.Fatalf("expected discard choice_type=%s, got %s", choiceType, got)
	}
}

func findBeastSamuraiCombatPromptForPlayer(obs *testutils.CaptureObserver, playerID string) *model.Prompt {
	if obs == nil {
		return nil
	}
	for i := len(obs.Events) - 1; i >= 0; i-- {
		event := obs.Events[i]
		if event.Type != model.EventAskInput {
			continue
		}
		prompt, ok := event.Data.(*model.Prompt)
		if !ok || prompt == nil || prompt.PlayerID != playerID {
			continue
		}
		if testutils.PromptHasOptionID(prompt, "take") || testutils.PromptHasOptionID(prompt, "defend") || testutils.PromptHasOptionID(prompt, "counter") {
			copied := *prompt
			copied.Options = append([]model.PromptOption(nil), prompt.Options...)
			return &copied
		}
	}
	return nil
}

func TestBeastSamurai_InitTokens(t *testing.T) {
	game, p1, _ := newBeastSamuraiTestEngine(t, testutils.NoopObserver{}, "")

	if p1.Tokens["bs_zanshin"] != 0 {
		t.Fatalf("expected initial zanshin=0, got %d", p1.Tokens["bs_zanshin"])
	}
	if p1.Tokens["bs_beast_soul"] != 0 {
		t.Fatalf("expected initial beast soul=0, got %d", p1.Tokens["bs_beast_soul"])
	}
	if p1.TurnState.UsedSkillCounts["bs_one_strike_armed"] != 0 {
		t.Fatalf("expected initial one-strike flag=0, got %d", p1.TurnState.UsedSkillCounts["bs_one_strike_armed"])
	}
	if p1.Orientation != model.OrientationNormal {
		t.Fatalf("expected initial orientation normal, got %s", p1.Orientation)
	}
	if p1.Form != "" {
		t.Fatalf("expected initial form empty, got %s", p1.Form)
	}
	if game.GetPlayerOrientation("p1") != model.OrientationNormal {
		t.Fatalf("expected engine orientation normal, got %s", game.GetPlayerOrientation("p1"))
	}
}

func TestBeastSamurai_WarriorZanshinThenOneStrikeBecomesAvailable(t *testing.T) {
	game, p1, p2 := newBeastSamuraiTestEngine(t, testutils.NoopObserver{}, "")
	p1.Tokens["bs_zanshin"] = 3

	ctx := game.BuildContext(p1, p2, model.TimingOnActionEnd, &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   p1.ID,
		TargetID:   p2.ID,
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{ActionType: string(model.ActionAttack)},
	})
	game.Dispatcher().OnTiming(ctx.Timing, ctx)

	if got := p1.Tokens["bs_zanshin"]; got != 4 {
		t.Fatalf("expected zanshin=4 after [武者残心], got %d", got)
	}
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if !testutils.InterruptHasSkillID(game.State.PendingInterrupt, "bs_one_strike_no_thought") {
		t.Fatalf("expected [一击无念] to become available, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
}

func TestBeastSamurai_OneStrikeArmedSurvivesTurnEndPreExtra(t *testing.T) {
	game, p1, _ := newBeastSamuraiTestEngine(t, testutils.NoopObserver{}, "")
	p1.TurnState.UsedSkillCounts["bs_one_strike_armed"] = 1
	model.AppendAttackAction(p1, "一击无念")

	if paused := game.RunTimingOnTurnEndStageHooks(p1, engine.TimingOnTurnEndPreExtra); paused {
		t.Fatalf("unexpected interrupt during pre-extra turn end")
	}
	if p1.TurnState.UsedSkillCounts["bs_one_strike_armed"] != 1 {
		t.Fatalf("one-strike armed should survive pre-extra turn end, got %d", p1.TurnState.UsedSkillCounts["bs_one_strike_armed"])
	}
	if len(p1.TurnState.PendingActions) != 1 || p1.TurnState.PendingActions[0].Source != "一击无念" {
		t.Fatalf("expected pending one-strike attack action, got %+v", p1.TurnState.PendingActions)
	}

	if paused := game.RunTimingOnTurnEndStageHooks(p1, engine.TimingOnTurnEndFinal); paused {
		t.Fatalf("unexpected interrupt during final turn end")
	}
	if p1.TurnState.UsedSkillCounts["bs_one_strike_armed"] != 0 {
		t.Fatalf("one-strike armed should expire on final turn end when unused, got %d", p1.TurnState.UsedSkillCounts["bs_one_strike_armed"])
	}
}

func TestBeastSamurai_OneStrike_NextAttackIgnoresShieldAndHoly(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game, p1, p2 := newBeastSamuraiTestEngine(t, obs, "angel")

	attackCard := model.Card{
		ID:      "bs-attack-normal",
		Name:    "裂风斩",
		Type:    model.CardTypeAttack,
		Element: model.ElementFire,
		Faction: "武",
		Damage:  2,
	}
	p1.Hand = []model.Card{attackCard}
	p1.TurnState.UsedSkillCounts["bs_one_strike_armed"] = 1

	p2.Hand = []model.Card{
		{ID: "holy-1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Faction: "圣"},
	}
	p2.Field = append(p2.Field, &model.FieldCard{
		Card:   model.Card{ID: "shield-1", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
		Mode:   model.FieldEffect,
		Effect: model.EffectShield,
	})

	game.State.TurnStage = model.TurnStageActionExecution
	game.State.ActionQueue = []model.QueuedAction{
		{
			SourceID: p1.ID,
			TargetID: p2.ID,
			Type:     model.ActionAttack,
			Card:     &attackCard,
			CardID:   attackCard.ID,
		},
	}

	game.Drive()

	if len(game.State.CombatStack) != 1 {
		t.Fatalf("expected combat stack size 1, got %d", len(game.State.CombatStack))
	}
	if p1.TurnState.UsedSkillCounts["bs_one_strike_armed"] != 0 {
		t.Fatalf("expected one-strike armed cleared, got %d", p1.TurnState.UsedSkillCounts["bs_one_strike_armed"])
	}
	if !game.State.CombatStack[0].HasInterceptTag(model.CombatInterceptIgnoreHolyShield) {
		t.Fatalf("expected one-strike attack to carry IgnoreHolyShield tag")
	}
	if !game.State.CombatStack[0].HasInterceptTag(model.CombatInterceptIgnoreTargetHoly) {
		t.Fatalf("expected one-strike attack to carry IgnoreTargetHoly tag")
	}
	if game.HasUsableShieldForCombat(p2, game.State.CombatStack[0]) {
		t.Fatalf("expected shield to be ignored by one-strike attack")
	}
	prompt := findBeastSamuraiCombatPromptForPlayer(obs, p2.ID)
	if prompt == nil {
		t.Fatalf("expected combat prompt for target")
	}
	if testutils.PromptHasOptionID(prompt, "defend") {
		t.Fatalf("expected defend option to be removed under one-strike attack")
	}
	if err := game.HandleCombatResponse(model.PlayerAction{
		PlayerID:  p2.ID,
		Type:      model.CmdRespond,
		ExtraArgs: []string{"defend"},
		CardIndex: 0,
	}); err == nil {
		t.Fatalf("expected holy defend to be rejected under one-strike attack")
	}
}

func TestBeastSamurai_OneStrike_JiFactionForceHit(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game, p1, p2 := newBeastSamuraiTestEngine(t, obs, "")

	attackCard := model.Card{
		ID:      "bs-attack-ji",
		Name:    "瞬斩",
		Type:    model.CardTypeAttack,
		Element: model.ElementFire,
		Faction: "技",
		Damage:  2,
	}
	p1.Hand = []model.Card{attackCard}
	p1.TurnState.UsedSkillCounts["bs_one_strike_armed"] = 1
	p2.Hand = nil

	game.State.TurnStage = model.TurnStageActionExecution
	game.State.ActionQueue = []model.QueuedAction{
		{
			SourceID: p1.ID,
			TargetID: p2.ID,
			Type:     model.ActionAttack,
			Card:     &attackCard,
			CardID:   attackCard.ID,
		},
	}

	game.Drive()

	if prompt := findBeastSamuraiCombatPromptForPlayer(obs, p2.ID); prompt != nil {
		t.Fatalf("expected no combat response prompt on forced-hit one-strike attack")
	}
	if len(game.State.CombatStack) != 0 {
		t.Fatalf("expected combat stack cleared after forced hit, got %d", len(game.State.CombatStack))
	}
	if len(p2.Hand) != 2 {
		t.Fatalf("expected target to draw 2 cards from forced-hit damage, got hand=%d", len(p2.Hand))
	}
}

func TestBeastSamurai_BeastSoulWill_NormalFormHitGainBeastSoul(t *testing.T) {
	game, p1, _ := newBeastSamuraiTestEngine(t, testutils.NoopObserver{}, "")
	attackCard := model.Card{
		ID:      "bs-hit",
		Name:    "试斩",
		Type:    model.CardTypeAttack,
		Element: model.ElementFire,
		Faction: "武",
		Damage:  2,
	}

	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   "p1",
			TargetID:   "p2",
			Damage:     2,
			DamageType: model.AttackDamage,
			Card:       &attackCard,
		},
	}

	if paused := game.ProcessPendingDamages(); paused {
		t.Fatalf("expected hit to resolve without extra interrupt")
	}
	if got := p1.Tokens["bs_beast_soul"]; got != 1 {
		t.Fatalf("expected beast soul +1 on normal-form active hit, got %d", got)
	}
}

func TestBeastSamurai_BeastSoulAlert_RunsOnOtherPlayerTapped(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game, p1, p2 := newBeastSamuraiTestEngine(t, obs, "")
	p1.Tokens["bs_beast_soul"] = 1
	p2.Hand = []model.Card{
		{ID: "magic-1", Name: "火球", Type: model.CardTypeMagic, Element: model.ElementFire, Faction: "法"},
	}

	before := game.SnapshotPlayerPoses()
	p2.Orientation = model.OrientationTapped
	game.DispatchOrientationChanges(before)

	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if err := game.ConfirmResponseSkill("p1", "bs_beast_soul_alert"); err != nil {
		t.Fatalf("confirm beast soul alert failed: %v", err)
	}
	// 兽魂警戒：由触发横置的角色（p2）直接弃牌，兽灵武士不再另选目标。
	requireBeastSamuraiDiscardInterrupt(t, game, "p2", "bs_alert_source_discard")

	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm alert discard failed: %v", err)
	}

	if p1.Orientation != model.OrientationTapped {
		t.Fatalf("expected beast samurai tapped after alert, got %s", p1.Orientation)
	}
	if p1.Form != "beast_samurai_iaijutsu_form" {
		t.Fatalf("expected beast samurai to enter iaijutsu form, got %s", p1.Form)
	}
	if got := p1.Tokens["bs_zanshin"]; got != 1 {
		t.Fatalf("expected zanshin +1 from alert, got %d", got)
	}
	if got := p1.Tokens["bs_beast_soul"]; got != 1 {
		t.Fatalf("expected beast soul net stay at 1 after magic discard refund, got %d", got)
	}
	if reveal := testutils.FindPublicDiscardReveal(obs, "p2"); reveal == nil {
		t.Fatalf("expected alert source discard to be public reveal")
	}
}

func TestBeastSamurai_BeastSoulAlert_DiscardOnlyTappedOperator(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game, p1, p2 := newBeastSamuraiTestEngine(t, obs, "")
	p1.Tokens["bs_beast_soul"] = 1
	p1.Hand = []model.Card{
		{ID: "bs-hand", Name: "兽灵手牌", Type: model.CardTypeAttack, Element: model.ElementFire},
	}
	p2.Hand = []model.Card{
		{ID: "magic-1", Name: "火球", Type: model.CardTypeMagic, Element: model.ElementFire, Faction: "法"},
	}

	before := game.SnapshotPlayerPoses()
	p2.Orientation = model.OrientationTapped
	game.DispatchOrientationChanges(before)

	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if err := game.ConfirmResponseSkill("p1", "bs_beast_soul_alert"); err != nil {
		t.Fatalf("confirm beast soul alert failed: %v", err)
	}
	if game.State.PendingInterrupt != nil {
		if ctx, ok := game.State.PendingInterrupt.Context.(map[string]interface{}); ok {
			if ctx["choice_type"] == "bs_alert_target" {
				t.Fatalf("expected no beast samurai target pick step, got bs_alert_target")
			}
		}
	}
	requireBeastSamuraiDiscardInterrupt(t, game, "p2", "bs_alert_source_discard")
}

func TestBeastSamurai_BeastReturn_XFlowAndMagicDiscardGainSoul(t *testing.T) {
	game, p1, p2 := newBeastSamuraiTestEngine(t, testutils.NoopObserver{}, "")
	p1.Tokens["bs_beast_soul"] = 1
	p1.Hand = []model.Card{
		{ID: "self-discard", Name: "自弃牌", Type: model.CardTypeAttack, Element: model.ElementFire},
	}
	p2.Hand = []model.Card{
		{ID: "source-magic", Name: "来源法术", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	damage := 2
	game.State.CombatStage = model.CombatStageCalcDamage
	ctx := game.BuildContext(p1, p2, model.TimingOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  p2.ID,
		TargetID:  p1.ID,
		DamageVal: &damage,
	})
	ctx.Flags["IsMagicDamage"] = true
	game.Dispatcher().OnTiming(ctx.Timing, ctx)

	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if err := game.ConfirmResponseSkill("p1", "bs_beast_return"); err != nil {
		t.Fatalf("confirm beast return failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bs_beast_return_x")

	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose X=1 failed: %v", err)
	}
	requireBeastSamuraiDiscardInterrupt(t, game, "p1", "bs_beast_return_self_discard")

	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm self discard failed: %v", err)
	}
	requireBeastSamuraiDiscardInterrupt(t, game, "p2", "bs_beast_return_source_discard")

	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm source discard failed: %v", err)
	}

	if got := p1.Tokens["bs_zanshin"]; got != 1 {
		t.Fatalf("expected zanshin +1 from beast soul removal, got %d", got)
	}
	if got := p1.Tokens["bs_beast_soul"]; got != 1 {
		t.Fatalf("expected beast soul refunded by source magic discard, got %d", got)
	}
	if len(game.State.DiscardPile) != 2 {
		t.Fatalf("expected two discarded cards in pile, got %d", len(game.State.DiscardPile))
	}
}

func TestBeastSamurai_BeastReturnSkip_ResumesPendingDamageWithoutReprompt(t *testing.T) {
	game, p1, p2 := newBeastSamuraiTestEngine(t, testutils.NoopObserver{}, "")
	p1.Tokens["bs_beast_soul"] = 0
	p1.Heal = 0
	p2.Heal = 0
	p1.Hand = nil
	p2.Hand = nil

	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   p2.ID,
			TargetID:   p1.ID,
			Damage:     1,
			DamageType: model.MagicAttack,
		},
	}

	if paused := game.ProcessPendingDamages(); !paused {
		t.Fatalf("expected beast return response prompt on magic damage")
	}
	testutils.RequireResponseSkillPrompt(t, game, "p1")

	if err := game.SkipResponse(); err != nil {
		t.Fatalf("skip beast return failed: %v", err)
	}

	if game.State.CombatStage != model.CombatStageCalcDamage {
		t.Fatalf("expected combat stage resume to damage resolution, got %s", game.State.CombatStage)
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected pending damage to remain for resolution, got %d", len(game.State.PendingDamageQueue))
	}
	if !game.State.PendingDamageQueue[0].DamageTakenFlowDispatched {
		t.Fatalf("expected damage-taken dispatch to be marked checked after skip")
	}

	for i := 0; i < 4 && len(game.State.PendingDamageQueue) > 0; i++ {
		if paused := game.ProcessPendingDamages(); paused {
			t.Fatalf("expected pending damage to finish without reprompt, got %+v", game.State.PendingInterrupt)
		}
	}
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("expected pending damage queue drained after skip, got %d", len(game.State.PendingDamageQueue))
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after damage resolution, got %+v", game.State.PendingInterrupt)
	}
}

func TestBeastSamurai_IaijutsuTurnEndDrainAndZeroExit(t *testing.T) {
	game, p1, _ := newBeastSamuraiTestEngine(t, testutils.NoopObserver{}, "")
	p1.Orientation = model.OrientationTapped
	p1.Form = "beast_samurai_iaijutsu_form"
	p1.Tokens["bs_beast_soul"] = 1
	p1.Tokens["bs_zanshin"] = 0

	game.State.TurnStage = model.TurnStageTurnEnd
	game.Drive()

	if got := p1.Tokens["bs_beast_soul"]; got != 0 {
		t.Fatalf("expected beast soul drained to 0, got %d", got)
	}
	if got := p1.Tokens["bs_zanshin"]; got != 1 {
		t.Fatalf("expected zanshin +1 from drain, got %d", got)
	}
	if p1.Form != "" {
		t.Fatalf("expected iaijutsu form to exit when beast soul reaches 0, got %s", p1.Form)
	}
	if p1.Orientation != model.OrientationNormal {
		t.Fatalf("expected orientation normal after exit, got %s", p1.Orientation)
	}
}

func TestBeastSamurai_ReversalIaijutsu_ReplacesDamageWithDiscard(t *testing.T) {
	game, p1, p2 := newBeastSamuraiTestEngine(t, testutils.NoopObserver{}, "")
	p1.Orientation = model.OrientationTapped
	p1.Form = "beast_samurai_iaijutsu_form"
	p1.Tokens["bs_beast_soul"] = 1
	p2.Hand = []model.Card{
		{ID: "target-1", Name: "目标牌1", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "target-2", Name: "目标牌2", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	attackCard := model.Card{
		ID:      "reversal-hit",
		Name:    "居合斩",
		Type:    model.CardTypeAttack,
		Element: model.ElementFire,
		Faction: "技",
		Damage:  2,
	}
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   p1.ID,
			TargetID:   p2.ID,
			Damage:     2,
			DamageType: model.AttackDamage,
			Card:       &attackCard,
		},
	}

	if !game.ProcessPendingDamages() {
		t.Fatalf("expected reversal prompt on attack hit")
	}
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if err := game.ConfirmResponseSkill("p1", "bs_reversal_iaijutsu"); err != nil {
		t.Fatalf("confirm reversal failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bs_reversal_x")

	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose reversal X=1 failed: %v", err)
	}
	requireBeastSamuraiDiscardInterrupt(t, game, "p2", "bs_reversal_target_discard")

	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm reversal discard (1st) failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm reversal discard (2nd) failed: %v", err)
	}
	if paused := game.ProcessPendingDamages(); paused {
		t.Fatalf("expected zero-damage attack to finish after reversal discard")
	}

	if got := p1.Tokens["bs_zanshin"]; got != 1 {
		t.Fatalf("expected zanshin +1 from reversal beast soul removal, got %d", got)
	}
	if got := game.State.BlueMorale; got != 14 {
		t.Fatalf("expected blue morale reduced to 14 because actual discard < X+2, got %d", got)
	}
	if len(p2.Hand) != 0 {
		t.Fatalf("expected target hand emptied by reversal, got %d", len(p2.Hand))
	}
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("expected pending damage queue cleared, got %d", len(game.State.PendingDamageQueue))
	}
}

func TestBeastSamurai_IaijutsuStyle_CanOverflowBeastSoulAndEnterForm(t *testing.T) {
	game, p1, _ := newBeastSamuraiTestEngine(t, testutils.NoopObserver{}, "")
	p1.Gem = 1
	p1.Tokens["bs_beast_soul"] = 2

	handler := skills.GetHandler("bs_iaijutsu_style")
	if handler == nil {
		t.Fatalf("expected iaijutsu style handler")
	}
	game.State.TurnStage = model.TurnStageActionStart
	ctx := game.BuildContext(p1, nil, model.TimingOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: p1.ID,
	})
	if err := handler.Execute(ctx); err != nil {
		t.Fatalf("execute iaijutsu style failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bs_iaijutsu_style_mode")

	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose iaijutsu style draw mode failed: %v", err)
	}

	if got := p1.Tokens["bs_beast_soul"]; got != 3 {
		t.Fatalf("expected iaijutsu style to overflow beast soul to 3, got %d", got)
	}
	if p1.Orientation != model.OrientationTapped {
		t.Fatalf("expected iaijutsu style to tap self in normal form, got %s", p1.Orientation)
	}
	if p1.Form != "beast_samurai_iaijutsu_form" {
		t.Fatalf("expected iaijutsu style to enter iaijutsu form, got %s", p1.Form)
	}
}
