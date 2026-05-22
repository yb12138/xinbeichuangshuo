package moon_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/data"
	moonplayer "starcup-engine/internal/engine/player/moon_goddess"
	"starcup-engine/internal/model"
)

func moonTestCard(id, name string, cardType model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     ele,
		Faction:     "圣",
		Damage:      2,
		Description: name,
	}
}

func TestMoonGoddessConfig_SkillIDsAndLogicHandlersMatchModule(t *testing.T) {
	var moonCharacter *model.Character
	for _, character := range data.GetCharacters() {
		if character.ID == "moon_goddess" {
			c := character
			moonCharacter = &c
			break
		}
	}
	if moonCharacter == nil {
		t.Fatalf("moon_goddess character config not found")
	}

	moduleSkills := make(map[string]bool)
	for _, entry := range moonplayer.SkillEntries() {
		if entry.ID == "" {
			t.Fatalf("moon goddess module contains empty skill id")
		}
		if entry.Handler == nil {
			t.Fatalf("moon goddess module skill %q has nil handler", entry.ID)
		}
		if moduleSkills[entry.ID] {
			t.Fatalf("moon goddess module registers duplicate skill %q", entry.ID)
		}
		moduleSkills[entry.ID] = true
	}

	dataSkills := make(map[string]bool)
	for _, skill := range moonCharacter.Skills {
		dataSkills[skill.ID] = true
		if !moduleSkills[skill.ID] {
			t.Fatalf("moon goddess skill %q is in character data but not registered in module", skill.ID)
		}
		if skill.LogicHandler == "" {
			t.Fatalf("moon goddess skill %q has empty LogicHandler", skill.ID)
		}
		if skill.LogicHandler != skill.ID {
			t.Fatalf("moon goddess skill %q uses LogicHandler %q; this role expects exact ID/handler parity", skill.ID, skill.LogicHandler)
		}
		if !moduleSkills[skill.LogicHandler] {
			t.Fatalf("moon goddess LogicHandler %q for skill %q is not registered in module", skill.LogicHandler, skill.ID)
		}
	}

	for skillID := range moduleSkills {
		if !dataSkills[skillID] {
			t.Fatalf("moon goddess module registers skill %q but character data does not define it", skillID)
		}
	}
}

func TestMoonGoddessNewMoonShelter_AbsorbsOverflowAndPreventsMoraleLoss(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	ally.MaxHand = 4
	ally.Hand = []model.Card{
		moonTestCard("a1", "牌1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("a2", "牌2", model.CardTypeAttack, model.ElementWater),
		moonTestCard("a3", "牌3", model.CardTypeAttack, model.ElementWind),
		moonTestCard("a4", "牌4", model.CardTypeAttack, model.ElementThunder),
		moonTestCard("a5", "牌5", model.CardTypeMagic, model.ElementDark),
		moonTestCard("a6", "牌6", model.CardTypeMagic, model.ElementLight),
	}

	damageOverflowCtx := game.BuildContext(ally, nil, model.TimingActionDuring, nil)
	damageOverflowCtx.Flags["FromDamageDraw"] = true
	damageOverflowCtx.Flags["IsMagicDamage"] = false
	game.CheckHandLimitCtx(ally, damageOverflowCtx)

	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt from overflow")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{4, 5},
	})

	// 新月庇护改为 ResponseOptional 后，玩家需在 choose_skill 弹框中显式确认发动。
	testutils.ChooseResponseSkillByID(t, game, "p1", "mg_new_moon_shelter")

	if got := game.State.RedMorale; got != 15 {
		t.Fatalf("expected red morale unchanged by 新月庇护, got %d", got)
	}
	if got := moon.Form; got != model.FormMoonGoddessDarkMoon {
		t.Fatalf("expected moon enter dark form, got %q", got)
	}
	if got := moonplayer.DarkMoonCount(moon); got != 2 {
		t.Fatalf("expected 2 dark moons absorbed, got %d", got)
	}
	if got := len(game.State.DiscardPile); got != 0 {
		t.Fatalf("expected absorbed cards not in discard pile, got %d", got)
	}
}

func TestMoonGoddessNewMoonShelter_PromptsAfterBloodPriestessDamageOverflow(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Blood", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	blood := game.State.Players["p2"]
	blood.MaxHand = 6
	blood.Hand = []model.Card{
		moonTestCard("b1", "牌1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("b2", "牌2", model.CardTypeAttack, model.ElementWater),
		moonTestCard("b3", "牌3", model.CardTypeAttack, model.ElementWind),
		moonTestCard("b4", "牌4", model.CardTypeAttack, model.ElementThunder),
		moonTestCard("b5", "牌5", model.CardTypeMagic, model.ElementDark),
		moonTestCard("b6", "牌6", model.CardTypeMagic, model.ElementLight),
	}
	game.State.Deck = []model.Card{
		moonTestCard("drawn", "伤害摸牌", model.CardTypeAttack, model.ElementEarth),
	}

	game.AddPendingDamage(model.PendingDamage{
		SourceID:   "p3",
		TargetID:   "p2",
		Damage:     1,
		DamageType: model.AttackDamage,
	})
	if paused := game.ProcessPendingDamages(); !paused {
		t.Fatalf("expected damage flow to pause for overflow discard")
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt from damage overflow, got %+v", game.State.PendingInterrupt)
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{6},
	})

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected new moon shelter response after blood priestess damage overflow, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected moon goddess response prompt, got player %s", game.State.PendingInterrupt.PlayerID)
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "mg_new_moon_shelter" {
		t.Fatalf("expected only new moon shelter response, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	foundPrompt := false
	for _, event := range obs.Events {
		if event.Type != model.EventAskInput {
			continue
		}
		prompt := event.Prompt
		if prompt != nil &&
			prompt.PlayerID == "p1" &&
			prompt.Type == model.PromptChooseSkill &&
			len(prompt.Options) > 0 &&
			prompt.Options[0].ID == "mg_new_moon_shelter" {
			foundPrompt = true
			break
		}
	}
	if !foundPrompt {
		t.Fatalf("expected observer to emit new moon shelter choose_skill prompt")
	}

	testutils.ChooseResponseSkillByID(t, game, "p1", "mg_new_moon_shelter")

	if got := game.State.RedMorale; got != 15 {
		t.Fatalf("expected red morale unchanged by 新月庇护, got %d", got)
	}
	if got := moon.Form; got != model.FormMoonGoddessDarkMoon {
		t.Fatalf("expected moon enter dark form, got %q", got)
	}
	if got := moonplayer.DarkMoonCount(moon); got != 1 {
		t.Fatalf("expected 1 dark moon absorbed, got %d", got)
	}
	if intr := game.State.PendingInterrupt; intr != nil && intr.Type == model.InterruptChoice {
		if data, ok := intr.Context.(map[string]interface{}); ok && data["choice_type"] == "mg_moon_cycle_mode" {
			testutils.MustHandleAction(t, game, model.PlayerAction{
				PlayerID:   "p1",
				Type:       model.CmdSelect,
				Selections: []int{0},
			})
		}
	}

	game.State.Deck = []model.Card{
		moonTestCard("drawn-2", "第二次伤害摸牌", model.CardTypeAttack, model.ElementFire),
	}
	game.AddPendingDamage(model.PendingDamage{
		SourceID:   "p3",
		TargetID:   "p2",
		Damage:     1,
		DamageType: model.AttackDamage,
	})
	if paused := game.ProcessPendingDamages(); !paused {
		t.Fatalf("expected second damage flow to pause for overflow discard")
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected second discard interrupt from damage overflow, got %+v", game.State.PendingInterrupt)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{6},
	})
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptResponseSkill {
		t.Fatalf("expected no new moon shelter response while moon goddess is in dark form, got %+v", game.State.PendingInterrupt)
	}
	if got := game.State.RedMorale; got != 14 {
		t.Fatalf("expected second damage overflow to reduce red morale while in dark form, got %d", got)
	}
	if got := moon.Form; got != model.FormMoonGoddessDarkMoon {
		t.Fatalf("expected moon stay in dark form after second overflow, got %q", got)
	}
	if got := moonplayer.DarkMoonCount(moon); got != 1 {
		t.Fatalf("expected dark moon count unchanged after skipped shelter, got %d", got)
	}
}

func TestMoonGoddessNewMoonShelter_NestedBeastSoulAlertRunsDeferredResume(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Beast", "beast_samurai", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Victim", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	beast := game.State.Players["p2"]
	victim := game.State.Players["p3"]
	moon.Hand = []model.Card{
		moonTestCard("moon-magic", "月术", model.CardTypeMagic, model.ElementDark),
	}
	beast.Tokens["bs_beast_soul"] = 1
	victim.MaxHand = 5
	victim.Hand = []model.Card{
		moonTestCard("h1", "牌1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("h2", "牌2", model.CardTypeAttack, model.ElementWater),
		moonTestCard("h3", "牌3", model.CardTypeAttack, model.ElementWind),
		moonTestCard("h4", "牌4", model.CardTypeAttack, model.ElementThunder),
		moonTestCard("m1", "法术", model.CardTypeMagic, model.ElementDark),
	}
	game.State.Deck = []model.Card{
		moonTestCard("drawn", "伤害摸牌", model.CardTypeAttack, model.ElementEarth),
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	game.AddPendingDamage(model.PendingDamage{
		SourceID:   "p4",
		TargetID:   "p3",
		Damage:     1,
		DamageType: model.AttackDamage,
	})
	if paused := game.ProcessPendingDamages(); !paused {
		t.Fatalf("expected damage flow to pause for overflow discard")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p3", Selections: []int{5}})

	testutils.ChooseResponseSkillByID(t, game, "p1", "mg_new_moon_shelter")
	testutils.RequireResponseSkillPrompt(t, game, "p2")
	testutils.ChooseResponseSkillByID(t, game, "p2", "bs_beast_soul_alert")
	requireMoonDiscardPrompt(t, game, "p1", "bs_alert_source_discard")

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("resolve beast soul alert discard failed: %v", err)
	}

	if game.State.Subflow == model.SubflowResponse && game.State.PendingInterrupt == nil {
		t.Fatalf("flow should not stay in empty response after nested beast soul alert")
	}
	if game.State.TurnStage == "" {
		t.Fatalf("turn stage should remain concrete after nested response recovery")
	}
	if got := game.State.RedMorale; got != 15 {
		t.Fatalf("expected red morale unchanged by deferred 新月庇护 resume, got %d", got)
	}
	if got := moonplayer.DarkMoonCount(moon); got != 1 {
		t.Fatalf("expected moon to absorb overflow card after deferred resume, got %d", got)
	}
	if got := moon.Form; got != model.FormMoonGoddessDarkMoon {
		t.Fatalf("expected moon enter dark form, got %q", got)
	}
}

func requireMoonDiscardPrompt(t *testing.T, game *engine.GameEngine, playerID, choiceType string) {
	t.Helper()
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected pending discard interrupt, got %+v", game.State.PendingInterrupt)
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
		t.Fatalf("expected choice_type=%s, got %s", choiceType, got)
	}
}

func TestMoonGoddessNewMoonShelter_NoSoulDevourGainWhenMoraleLossPrevented(t *testing.T) {
	// 重复多次覆盖 map 迭代随机性，确保“暗月抵消士气后，灵魂术士不加黄魂”稳定成立。
	for i := 0; i < 24; i++ {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p2", "Moon", "moon_goddess", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
			t.Fatal(err)
		}

		soul := game.State.Players["p1"]
		moon := game.State.Players["p2"]
		soul.MaxHand = 4
		soul.Hand = []model.Card{
			moonTestCard("s1", "牌1", model.CardTypeAttack, model.ElementFire),
			moonTestCard("s2", "牌2", model.CardTypeAttack, model.ElementWater),
			moonTestCard("s3", "牌3", model.CardTypeAttack, model.ElementWind),
			moonTestCard("s4", "牌4", model.CardTypeAttack, model.ElementThunder),
			moonTestCard("s5", "牌5", model.CardTypeMagic, model.ElementDark),
			moonTestCard("s6", "牌6", model.CardTypeMagic, model.ElementLight),
		}

		damageOverflowCtx := game.BuildContext(soul, nil, model.TimingActionDuring, nil)
		damageOverflowCtx.Flags["FromDamageDraw"] = true
		damageOverflowCtx.Flags["IsMagicDamage"] = false
		game.CheckHandLimitCtx(soul, damageOverflowCtx)

		if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
			t.Fatalf("round %d: expected discard interrupt from overflow", i)
		}
		testutils.MustHandleAction(t, game, model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{4, 5},
		})

		// 新月庇护改为 ResponseOptional 后，由月之女神（p2）在 choose_skill 弹框中确认发动。
		testutils.ChooseResponseSkillByID(t, game, "p2", "mg_new_moon_shelter")

		if got := game.State.RedMorale; got != 15 {
			t.Fatalf("round %d: expected red morale unchanged by 新月庇护, got %d", i, got)
		}
		if got := soul.Tokens["ss_yellow_soul"]; got != 0 {
			t.Fatalf("round %d: expected soul devour no yellow gain when morale loss prevented, got %d", i, got)
		}
		if got := moonplayer.DarkMoonCount(moon); got != 2 {
			t.Fatalf("round %d: expected 2 dark moons absorbed, got %d", i, got)
		}
	}
}

func TestMoonGoddessMoonCycle_DeclineSkipsSkill(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 2
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	if !moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("expected moon cycle interrupt")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_moon_cycle_mode")
	prompt := game.GetCurrentPrompt()
	if prompt == nil || len(prompt.Options) == 0 || prompt.Options[0].Label != "不发动" {
		t.Fatalf("expected decline option first, got %+v", prompt.Options)
	}
	if prompt.Presentation == nil || prompt.Presentation.CancelPolicy != "" || prompt.Presentation.HasDecline {
		t.Fatalf("expected decline to render as a normal branch option, got %+v", prompt.Presentation)
	}
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("decline moon cycle failed: %v", err)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after decline, got %+v", game.State.PendingInterrupt)
	}
	if got := moonplayer.DarkMoonCount(moon); got != 1 {
		t.Fatalf("decline should not remove dark moon, got %d", got)
	}
	if got := moon.Heal; got != 2 {
		t.Fatalf("decline should not consume heal, got %d", got)
	}
}

func TestMoonGoddessMoonCycle_Branch1AppliesCurseAndHeal(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 0 // 仅保留分支①
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	ally.Heal = 0
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	if !moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("expected moon cycle interrupt")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_moon_cycle_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose moon cycle mode failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_moon_cycle_heal_target")
	if prompt := game.GetCurrentPrompt(); prompt == nil || prompt.Presentation == nil || prompt.Presentation.Kind != model.PresentationTargetPicker {
		t.Fatalf("expected moon cycle heal target to use target_picker presentation, got %+v", prompt)
	}
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose moon cycle heal target failed: %v", err)
	}

	if got := moonplayer.DarkMoonCount(moon); got != 0 {
		t.Fatalf("expected dark moon removed by branch1, got %d", got)
	}
	if got := moon.Form; got != "" {
		t.Fatalf("expected leave dark form when no dark moon, got %q", got)
	}
	if got := game.State.RedMorale; got != 14 {
		t.Fatalf("expected curse morale loss 1, got %d", got)
	}
	if got := ally.Heal; got != 1 {
		t.Fatalf("expected ally heal +1, got %d", got)
	}
}

func TestMoonGoddessMoonCycle_OnlyOncePerTurn(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 1
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	if !moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("expected moon cycle first dispatch")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_moon_cycle_mode")
	// 分支①
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose moon cycle mode branch1 failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_moon_cycle_heal_target")
	// 选自己，确保治疗>0，若无一次/回合门闩会继续出现分支②弹窗。
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose moon cycle heal target failed: %v", err)
	}
	if got := moon.TurnState.UsedSkillCounts["mg_moon_cycle"]; got != 1 {
		t.Fatalf("expected moon cycle used flag=1 in current turn, got %d", got)
	}

	if moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("moon cycle should not dispatch twice in same turn")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after second dispatch attempt")
	}
}

func TestMoonGoddessMoonCycle_Branch1NoRepromptBranch2InDriveFlow(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 1 // 同时满足分支①/②，验证选①后不会再弹②
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "mg_moon_cycle_mode")

	// 选择分支①
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})
	testutils.RequireChoicePrompt(t, game, "p1", "mg_moon_cycle_heal_target")

	// 选择治疗目标并完成分支①
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})

	if intr := game.State.PendingInterrupt; intr != nil && intr.Type == model.InterruptChoice {
		if data, ok := intr.Context.(map[string]interface{}); ok {
			if ct, _ := data["choice_type"].(string); ct == "mg_moon_cycle_mode" && intr.PlayerID == "p1" {
				t.Fatalf("moon cycle should not reprompt branch mode after branch1 resolved")
			}
		}
	}
	for _, intr := range game.State.InterruptQueue {
		if intr == nil || intr.Type != model.InterruptChoice || intr.PlayerID != "p1" {
			continue
		}
		if data, ok := intr.Context.(map[string]interface{}); ok {
			if ct, _ := data["choice_type"].(string); ct == "mg_moon_cycle_mode" {
				t.Fatalf("moon cycle mode should not stay queued after branch1 resolved")
			}
		}
	}
}

func TestMoonGoddessMoonCycle_TurnStateLatchPreventsRepromptWhenTokenResets(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 1
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	if !moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("expected moon cycle first dispatch")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_moon_cycle_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose moon cycle mode branch1 failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_moon_cycle_heal_target")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose moon cycle heal target failed: %v", err)
	}

	if moonplayer.MaybeMoonCycleAtTurnEnd(engine.NewRoleChoiceRuntime(game), moon) {
		t.Fatalf("moon cycle should stay blocked by turnstate latch")
	}
	if game.State.PendingInterrupt != nil {
		if data, ok := game.State.PendingInterrupt.Context.(map[string]interface{}); ok {
			if ct, _ := data["choice_type"].(string); ct == "mg_moon_cycle_mode" {
				t.Fatalf("unexpected moon cycle reprompt after turnstate latch")
			}
		}
	}
}

func TestMoonGoddessDarkMoonSlash_AddsDamageAndConsumesDarkMoon(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	enemy := game.State.Players["p2"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Crystal = 1
	moon.Form = model.FormMoonGoddessDarkMoon
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("dm2", "暗月2", model.CardTypeMagic, model.ElementWater),
	})
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   moon.ID,
			TargetID:   enemy.ID,
			Damage:     2,
			DamageType: model.AttackDamage,
		},
	}

	ctx := game.BuildContext(moon, enemy, model.TimingAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: moon.ID,
		TargetID: enemy.ID,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			IsHit:            true,
			CounterInitiator: "",
		},
	})
	h := &moonplayer.MoonGoddessDarkMoonSlashHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected dark moon slash can use")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute dark moon slash failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_darkmoon_slash_x")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose X=2 failed: %v", err)
	}

	if got := game.State.PendingDamageQueue[0].Damage; got != 4 {
		t.Fatalf("expected attack damage +2 (2->4), got %d", got)
	}
	if got := moonplayer.DarkMoonCount(moon); got != 0 {
		t.Fatalf("expected all dark moon removed, got %d", got)
	}
	if got := game.State.RedMorale; got != 13 {
		t.Fatalf("expected curse morale loss 2, got %d", got)
	}
	if got := moon.Crystal; got != 0 {
		t.Fatalf("expected consume 1 crystal, got %d", got)
	}
}

func TestMoonGoddessDarkMoonSlash_PromptsOnActiveAttackHit(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	enemy := game.State.Players["p2"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Crystal = 1
	moon.Form = model.FormMoonGoddessDarkMoon
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("dm2", "暗月2", model.CardTypeMagic, model.ElementWater),
		moonTestCard("dm3", "暗月3", model.CardTypeAttack, model.ElementWind),
		moonTestCard("dm4", "暗月4", model.CardTypeMagic, model.ElementLight),
	})
	attackCard := moonTestCard("atk", "主动攻击", model.CardTypeAttack, model.ElementFire)
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   moon.ID,
			TargetID:   enemy.ID,
			Damage:     2,
			DamageType: model.AttackDamage,
			Card:       &attackCard,
		},
	}

	if paused := game.ProcessPendingDamages(); !paused {
		t.Fatalf("expected attack hit to pause for dark moon slash response")
	}
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if got := game.State.PendingInterrupt.SkillIDs; len(got) != 1 || got[0] != "mg_darkmoon_slash" {
		t.Fatalf("expected dark moon slash response, got %+v", got)
	}
}

func TestMoonGoddessMedusa_ExcludesConvertedAttacks(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	enemy := game.State.Players["p3"]
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm_fire", "火暗月", model.CardTypeAttack, model.ElementFire),
	})
	attackCard := moonTestCard("atk", "火斩", model.CardTypeAttack, model.ElementFire)
	attackStartCtx := game.BuildContext(enemy, ally, model.TimingAttackDeclare, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: enemy.ID,
		TargetID: ally.ID,
		Card:     &attackCard,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
		},
	})

	if moonplayer.MaybeMedusa(engine.NewRoleChoiceRuntime(game), enemy, ally, "adventurer_fraud", &attackCard, attackStartCtx) {
		t.Fatalf("fraud converted attack should not dispatch medusa")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt for fraud converted attack")
	}
	if moonplayer.MaybeMedusa(engine.NewRoleChoiceRuntime(game), enemy, ally, "hb_holy_shard_storm", &attackCard, attackStartCtx) {
		t.Fatalf("holy shard storm converted attack should not dispatch medusa")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt for holy shard storm converted attack")
	}

	if !moonplayer.MaybeMedusa(engine.NewRoleChoiceRuntime(game), enemy, ally, "", &attackCard, attackStartCtx) {
		t.Fatalf("normal attack should dispatch medusa when matching dark moon exists")
	}
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	testutils.ChooseResponseSkillByID(t, game, "p1", "mg_medusa_eye")
	testutils.RequireChoicePrompt(t, game, "p1", "mg_medusa_darkmoon_pick")
}

func TestMoonGoddessMedusa_OnlyAtAttackStart(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	enemy := game.State.Players["p3"]
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm_fire", "火暗月", model.CardTypeAttack, model.ElementFire),
	})
	attackCard := moonTestCard("atk", "火斩", model.CardTypeAttack, model.ElementFire)

	// 非攻击开始上下文：不应触发。
	if moonplayer.MaybeMedusa(engine.NewRoleChoiceRuntime(game), enemy, ally, "", &attackCard, nil) {
		t.Fatalf("medusa should not dispatch without attack-start context")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt without attack-start context")
	}

	nonStartCtx := game.BuildContext(enemy, ally, model.TimingAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: enemy.ID,
		TargetID: ally.ID,
		Card:     &attackCard,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
		},
	})
	if moonplayer.MaybeMedusa(engine.NewRoleChoiceRuntime(game), enemy, ally, "", &attackCard, nonStartCtx) {
		t.Fatalf("medusa should not dispatch outside attack-start dispatch")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt for non-attack-start dispatch")
	}

	// 攻击开始上下文：可触发。
	attackStartCtx := game.BuildContext(enemy, ally, model.TimingAttackDeclare, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: enemy.ID,
		TargetID: ally.ID,
		Card:     &attackCard,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
		},
	})
	if !moonplayer.MaybeMedusa(engine.NewRoleChoiceRuntime(game), enemy, ally, "", &attackCard, attackStartCtx) {
		t.Fatalf("medusa should dispatch at attack start with matching dark moon")
	}
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	testutils.ChooseResponseSkillByID(t, game, "p1", "mg_medusa_eye")
	testutils.RequireChoicePrompt(t, game, "p1", "mg_medusa_darkmoon_pick")
}

func TestMoonGoddessMedusa_MagicDarkMoonExtraDamageTargetsAttackerOnly(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Attacker", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "OtherEnemy", "hero", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	attacker := game.State.Players["p3"]
	moonplayer.AddDarkMoonCards(moon, []model.Card{
		moonTestCard("dm_fire_magic", "火法术闇月", model.CardTypeMagic, model.ElementFire),
	})
	moon.Hand = []model.Card{
		moonTestCard("discard1", "弃牌", model.CardTypeAttack, model.ElementWater),
	}
	attackCard := moonTestCard("atk", "火斩", model.CardTypeAttack, model.ElementFire)
	attackStartCtx := game.BuildContext(attacker, ally, model.TimingAttackDeclare, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: attacker.ID,
		TargetID: ally.ID,
		Card:     &attackCard,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
		},
	})

	if !moonplayer.MaybeMedusa(engine.NewRoleChoiceRuntime(game), attacker, ally, "", &attackCard, attackStartCtx) {
		t.Fatalf("expected medusa dispatch")
	}
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	testutils.ChooseResponseSkillByID(t, game, "p1", "mg_medusa_eye")
	testutils.RequireChoicePrompt(t, game, "p1", "mg_medusa_darkmoon_pick")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose medusa dark moon failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_medusa_magic_discard")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("resolve medusa discard failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no extra target prompt after medusa discard, got %+v", game.State.PendingInterrupt)
	}
	if got := game.State.RedMorale; got != 14 {
		t.Fatalf("expected dark moon curse to reduce red morale by 1, got %d", got)
	}
	if got := obs.CountLogContains("[暗月诅咒]"); got != 1 {
		t.Fatalf("expected one dark moon curse log after medusa removal, got %d", got)
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected one pending damage from medusa extra effect, got %d", len(game.State.PendingDamageQueue))
	}
	if pd := game.State.PendingDamageQueue[0]; pd.TargetID != attacker.ID || pd.DamageType != "magic" || pd.Damage != 1 {
		t.Fatalf("expected medusa extra damage locked to attacker %s, got %+v", attacker.ID, pd)
	}
}

func TestMoonGoddessBlasphemy_OncePerTurnAndResetNextTurn(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.Heal = 2
	pd := model.PendingDamage{
		SourceID:   "p1",
		TargetID:   "p2",
		Damage:     1,
		DamageType: model.MagicAttack,
	}

	if !moonplayer.TryQueueBlasphemy(engine.NewRoleChoiceRuntime(game), &pd) {
		t.Fatalf("expected first blasphemy queue success")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_blasphemy_target")
	// 选第1个目标（最后一项为“不发动”）。
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("resolve blasphemy target failed: %v", err)
	}
	if got := moon.TurnState.UsedSkillCounts["mg_blasphemy"]; got != 1 {
		t.Fatalf("expected blasphemy used flag=1, got %d", got)
	}
	if got := moon.TurnState.SkillFlowState["mg_blasphemy_pending"]; got != 0 {
		t.Fatalf("expected blasphemy pending reset to 0, got %d", got)
	}
	if moonplayer.TryQueueBlasphemy(engine.NewRoleChoiceRuntime(game), &pd) {
		t.Fatalf("blasphemy should be blocked after used once in same turn")
	}

	moon.IsActive = true
	game.State.CurrentTurn = 0
	game.NextTurn() // p2's turn — moon's TurnState not reset yet
	game.NextTurn() // moon's turn — TurnState rebuilt via NewPlayerTurnState
	if got := moon.TurnState.UsedSkillCounts["mg_blasphemy"]; got != 0 {
		t.Fatalf("expected blasphemy used flag reset on own next turn, got %d", got)
	}
	if !moonplayer.TryQueueBlasphemy(engine.NewRoleChoiceRuntime(game), &pd) {
		t.Fatalf("expected blasphemy can queue again after turn reset")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_blasphemy_target")
}

func TestMoonGoddessBlasphemy_TargetLockedToDamagedEnemyAndSelfTurn(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "hero", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.Heal = 1
	pd := model.PendingDamage{
		SourceID:   "p1",
		TargetID:   "p2",
		Damage:     1,
		DamageType: model.MagicAttack,
	}

	if !moonplayer.TryQueueBlasphemy(engine.NewRoleChoiceRuntime(game), &pd) {
		t.Fatalf("expected blasphemy queue success in self turn")
	}
	prompt := game.GetCurrentPrompt()
	if prompt == nil || prompt.ChoiceType != "mg_blasphemy_target" {
		t.Fatalf("expected blasphemy prompt, got %+v", prompt)
	}
	if got := len(prompt.Options); got != 2 {
		t.Fatalf("expected only current damaged enemy + decline, got %d options", got)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("resolve blasphemy target failed: %v", err)
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected one pending damage from blasphemy, got %d", len(game.State.PendingDamageQueue))
	}
	if pd := game.State.PendingDamageQueue[0]; pd.TargetID != "p2" || pd.DamageType != "magic" || pd.Damage != 1 {
		t.Fatalf("expected blasphemy locked to damaged enemy p2, got %+v", pd)
	}

	moon.IsActive = false
	game.State.CurrentTurn = 1
	moon.TurnState.UsedSkillCounts["mg_blasphemy"] = 0
	moon.TurnState.SkillFlowState["mg_blasphemy_pending"] = 0
	moon.Heal = 1
	game.State.PendingInterrupt = nil
	game.State.PendingDamageQueue = nil
	if moonplayer.TryQueueBlasphemy(engine.NewRoleChoiceRuntime(game), &pd) {
		t.Fatalf("blasphemy should not queue outside self turn")
	}
}

func TestMoonGoddessPaleMoon_Branch1GrantsExtraTurn(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Gem = 1
	moon.Tokens["mg_petrify"] = 3
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "mg_pale_moon",
	})
	testutils.RequireChoicePrompt(t, game, "p1", "mg_pale_moon_mode")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose pale moon branch1 failed: %v", err)
	}
	if got := moon.TurnState.UsedSkillCounts["mg_next_attack_no_counter"]; got != 1 {
		t.Fatalf("expected next attack no-counter token=1, got %d", got)
	}
	if got := moon.TurnState.SkillFlowState["mg_extra_turn_pending"]; got != 1 {
		t.Fatalf("expected extra-turn pending=1, got %d", got)
	}
	if len(moon.TurnState.PendingActions) == 0 || moon.TurnState.PendingActions[0].MustType != "Attack" {
		t.Fatalf("expected pale moon branch1 to queue one extra Attack action")
	}

	game.NextTurn()
	if got := game.State.CurrentTurn; got != 0 {
		t.Fatalf("expected extra turn keeps current turn index at 0, got %d", got)
	}
	if got := moon.TurnState.SkillFlowState["mg_extra_turn_pending"]; got != 0 {
		t.Fatalf("expected extra-turn pending consumed after NextTurn, got %d", got)
	}
	if !moon.IsActive {
		t.Fatalf("expected moon still active in extra turn")
	}
}

func TestMoonGoddessPaleMoon_Branch2RequiresNewMoonAndXStartsAtOne(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Gem = 2
	moon.Hand = []model.Card{
		moonTestCard("h1", "弃牌", model.CardTypeAttack, model.ElementFire),
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "mg_pale_moon", nil, nil); err == nil {
		t.Fatalf("expected pale moon branch2 unavailable without any new moon")
	}

	moon.Gem = 2
	moon.Tokens["mg_new_moon"] = 2
	if err := game.UseSkill("p1", "mg_pale_moon", nil, nil); err != nil {
		t.Fatalf("use mg_pale_moon failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_pale_moon_mode")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose pale moon branch2 failed: %v", err)
	}
	prompt := game.GetCurrentPrompt()
	if prompt == nil || prompt.ChoiceType != "mg_pale_moon_x" {
		t.Fatalf("expected pale moon x prompt, got %+v", prompt)
	}
	if got := len(prompt.Options); got != 2 {
		t.Fatalf("expected X options only for 1..2, got %d", got)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose X=1 failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_pale_moon_target")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose pale moon target failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mg_pale_moon_discard")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("resolve pale moon discard failed: %v", err)
	}
	if got := moon.Tokens["mg_new_moon"]; got != 1 {
		t.Fatalf("expected consume 1 new moon, got %d", got)
	}
	if got := moon.Tokens["mg_petrify"]; got != 1 {
		t.Fatalf("expected petrify +1, got %d", got)
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected one pending damage from pale moon branch2, got %d", len(game.State.PendingDamageQueue))
	}
	if pd := game.State.PendingDamageQueue[0]; pd.TargetID != "p2" || pd.Damage != 2 || pd.DamageType != "magic" {
		t.Fatalf("expected pale moon branch2 deal 2 magic damage to p2, got %+v", pd)
	}
}

func TestMoonGoddessNewMoonShelter_NotDispatchWhenActualMoraleWillNotDrop(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]

	// 红莲骑士热血形态：伤害导致爆牌不掉士气，因此新月庇护不应触发。
	ally.Form = model.FormCrimsonKnightHotBlooded
	ally.MaxHand = 4
	ally.Hand = []model.Card{
		moonTestCard("h1", "牌1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("h2", "牌2", model.CardTypeAttack, model.ElementWater),
		moonTestCard("h3", "牌3", model.CardTypeAttack, model.ElementWind),
		moonTestCard("h4", "牌4", model.CardTypeAttack, model.ElementThunder),
		moonTestCard("h5", "牌5", model.CardTypeMagic, model.ElementDark),
		moonTestCard("h6", "牌6", model.CardTypeMagic, model.ElementLight),
	}

	damageOverflowCtx := game.BuildContext(ally, nil, model.TimingActionDuring, nil)
	damageOverflowCtx.Flags["FromDamageDraw"] = true
	damageOverflowCtx.Flags["IsMagicDamage"] = false
	game.CheckHandLimitCtx(ally, damageOverflowCtx)

	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt from overflow")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{4, 5},
	})

	if got := game.State.RedMorale; got != 15 {
		t.Fatalf("expected red morale unchanged, got %d", got)
	}
	if got := moonplayer.DarkMoonCount(moon); got != 0 {
		t.Fatalf("expected new moon shelter not dispatch, dark moon count=%d", got)
	}
	if got := moon.Form; got != "" {
		t.Fatalf("expected moon goddess stay non-dark form, got %q", got)
	}
	if got := len(game.State.DiscardPile); got != 2 {
		t.Fatalf("expected overflow cards enter discard pile, got %d", got)
	}
}

func TestMoonGoddessDarkMoonSlash_XBoundaries_CurseAndDamage(t *testing.T) {
	cases := []struct {
		name              string
		x                 int
		wantDamage        int
		wantRedMorale     int
		wantDarkMoonCount int
	}{
		{
			name:              "x1",
			x:                 1,
			wantDamage:        3,
			wantRedMorale:     14,
			wantDarkMoonCount: 1,
		},
		{
			name:              "x2",
			x:                 2,
			wantDamage:        4,
			wantRedMorale:     13,
			wantDarkMoonCount: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obs := &testutils.CaptureObserver{}
			game := engine.NewGameEngine(obs)
			if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
				t.Fatal(err)
			}

			moon := game.State.Players["p1"]
			enemy := game.State.Players["p2"]
			moon.IsActive = true
			moon.TurnState = model.NewPlayerTurnState()
			moon.Crystal = 1
			moon.Form = model.FormMoonGoddessDarkMoon
			moonplayer.AddDarkMoonCards(moon, []model.Card{
				moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
				moonTestCard("dm2", "暗月2", model.CardTypeMagic, model.ElementWater),
			})
			game.State.PendingDamageQueue = []model.PendingDamage{
				{
					SourceID:   moon.ID,
					TargetID:   enemy.ID,
					Damage:     2,
					DamageType: model.AttackDamage,
				},
			}

			ctx := game.BuildContext(moon, enemy, model.TimingAttackHit, &model.EventContext{
				Type:     model.EventAttack,
				SourceID: moon.ID,
				TargetID: enemy.ID,
				AttackInfo: &model.AttackEventInfo{
					ActionType:       string(model.ActionAttack),
					IsHit:            true,
					CounterInitiator: "",
				},
			})

			h := &moonplayer.MoonGoddessDarkMoonSlashHandler{}
			if !h.CanUse(ctx) {
				t.Fatalf("expected dark moon slash can use")
			}
			if err := h.Execute(ctx); err != nil {
				t.Fatalf("execute dark moon slash failed: %v", err)
			}
			testutils.RequireChoicePrompt(t, game, "p1", "mg_darkmoon_slash_x")
			prompt := game.GetCurrentPrompt()
			if prompt == nil {
				t.Fatalf("expected dark moon slash prompt")
			}
			if got := len(prompt.Options); got != 2 {
				t.Fatalf("expected only X=1..2 options, got %d", got)
			}
			if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{tc.x - 1}}); err != nil {
				t.Fatalf("choose x=%d failed: %v", tc.x, err)
			}

			if got := game.State.PendingDamageQueue[0].Damage; got != tc.wantDamage {
				t.Fatalf("expected damage=%d, got %d", tc.wantDamage, got)
			}
			if got := game.State.RedMorale; got != tc.wantRedMorale {
				t.Fatalf("expected red morale=%d, got %d", tc.wantRedMorale, got)
			}
			if got := obs.CountLogContains("[暗月诅咒]"); got != 1 {
				t.Fatalf("expected one dark moon curse log, got %d", got)
			}
			if tc.wantDarkMoonCount == 0 && moon.Form != "" {
				t.Fatalf("expected leave dark form immediately when all dark moons are removed, got %q", moon.Form)
			}
			if got := moonplayer.DarkMoonCount(moon); got != tc.wantDarkMoonCount {
				t.Fatalf("expected dark moon count=%d, got %d", tc.wantDarkMoonCount, got)
			}
			if got := moon.Crystal; got != 0 {
				t.Fatalf("expected consume 1 crystal in all branches, got %d", got)
			}
		})
	}
}
