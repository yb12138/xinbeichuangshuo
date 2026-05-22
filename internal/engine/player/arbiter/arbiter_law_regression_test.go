package arbiter_test

import (
	"starcup-engine/internal/engine"
	playerpkg "starcup-engine/internal/engine/player"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归测试：仲裁法则应在角色初始化时结算，而不是借首个回合开始偷渡触发。
func TestArbiterLaw_GrantsInitialCrystalAndDoesNotReactivateOnTurnStart(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 0

	if p1.Crystal != 2 {
		t.Fatalf("expected crystal=2 immediately after arbiter init, got %d", p1.Crystal)
	}

	ctx := game.BuildContext(p1, nil, model.TimingTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	game.Dispatcher().OnTiming(ctx.Timing, ctx)

	if p1.Crystal != 2 {
		t.Fatalf("expected crystal to stay 2 after turn-start dispatch, got %d", p1.Crystal)
	}
	if got := obs.CountLogContains("[仲裁法则]"); got != 0 {
		t.Fatalf("expected no [仲裁法则] turn-start log, got %d", got)
	}
}

// 回归测试：审判形态的“回合开始审判+1”应独立保留，不依赖仲裁法则重复触发
func TestArbiterForm_JudgmentAutoGainAtStartup(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnStart

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 0
	p1.Tokens = map[string]int{
		"judgment": 3,
	}
	playerpkg.SetForm(p1, model.FormArbiterJudgment)

	game.Drive()

	if p1.Tokens["judgment"] != 4 {
		t.Fatalf("expected judgment to increase to 4 in startup, got %d", p1.Tokens["judgment"])
	}
	if got := obs.CountLogContains("[仲裁法则]"); got != 0 {
		t.Fatalf("expected no [仲裁法则] log in form upkeep, got %d", got)
	}
	if got := obs.CountLogContains("处于审判形态，回合开始审判+1"); got != 1 {
		t.Fatalf("expected exactly one form upkeep log, got %d", got)
	}
}

func TestArbiterRitual_EntersFormWithoutImmediateJudgment(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Tokens["judgment"] = 2

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt for ritual, got %+v", game.State.PendingInterrupt)
	}

	ritualIdx := -1
	for i, skillID := range game.State.PendingInterrupt.SkillIDs {
		if skillID == "arbiter_ritual" {
			ritualIdx = i
			break
		}
	}
	if ritualIdx < 0 {
		t.Fatalf("arbiter_ritual not found in startup interrupt: %+v", game.State.PendingInterrupt.SkillIDs)
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{ritualIdx},
	})

	if got := p1.Form; got != model.FormArbiterJudgment {
		t.Fatalf("expected ritual to enter judgment form, got %q", got)
	}
	if got := p1.MaxHand; got != 5 {
		t.Fatalf("expected ritual to fix max hand at 5, got %d", got)
	}
	if got := p1.Tokens["judgment"]; got != 2 {
		t.Fatalf("expected ritual not to grant immediate judgment, got %d", got)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected ritual to consume 1 gem, got %d", got)
	}
}

func TestArbiterRitualBreak_RestoresHandLimitAndAddsTeamGem(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.MaxHand = 5
	playerpkg.SetForm(p1, model.FormArbiterJudgment)

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt for ritual break, got %+v", game.State.PendingInterrupt)
	}

	breakIdx := -1
	for i, skillID := range game.State.PendingInterrupt.SkillIDs {
		if skillID == "arbiter_ritual_break" {
			breakIdx = i
			break
		}
	}
	if breakIdx < 0 {
		t.Fatalf("arbiter_ritual_break not found in startup interrupt: %+v", game.State.PendingInterrupt.SkillIDs)
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{breakIdx},
	})

	if got := p1.Form; got != "" {
		t.Fatalf("expected ritual break to exit judgment form, got %q", got)
	}
	if got := p1.MaxHand; got != 6 {
		t.Fatalf("expected ritual break to restore max hand to 6, got %d", got)
	}
	if got := game.State.RedGems; got != 1 {
		t.Fatalf("expected ritual break to add 1 team gem, got %d", got)
	}
}

func TestArbiterForcedDoomsday_HappensAfterStartupAndTargetsEnemiesOnly(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Tokens = map[string]int{
		"judgment": 4,
	}
	p1.Hand = []model.Card{
		{ID: "arbiter-invalid-attack", Name: "测试攻击", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt before forced doomsday, got %+v", game.State.PendingInterrupt)
	}
	startupIdx := len(game.State.PendingInterrupt.SkillIDs)
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{startupIdx},
	})

	if !game.IsActionSelectionWindow() {
		t.Fatalf("expected action selection window after startup, got %s", game.RuntimeStateLabel())
	}
	if got := p1.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"]; got != 1 {
		t.Fatalf("expected forced doomsday pending token set, got %d", got)
	}

	state := game.BuildActionSelectionOptions("p1", p1)
	if state.ActionRuleMode != engine.ActionSelectionRuleForceSkillAsMagic {
		t.Fatalf("expected forced doomsday action selection restriction")
	}
	if len(state.ValidOptions) != 1 || state.ValidOptions[0].ID != "magic" {
		t.Fatalf("expected only magic option for forced doomsday, got %+v", state.ValidOptions)
	}

	err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p3", CardID: "arbiter-invalid-attack"})
	if err == nil {
		t.Fatalf("expected non-doomsday action to be rejected while forced")
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "arbiter_doomsday",
		TargetID: "p3",
	})

	if got := p1.Tokens["judgment"]; got != 0 {
		t.Fatalf("expected forced doomsday to clear judgment, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"]; got != 0 {
		t.Fatalf("expected forced doomsday pending cleared after skill use, got %d", got)
	}
}

func TestArbiterBalance_BranchesFollowConfiguredEffects(t *testing.T) {
	t.Run("branch_zero_discards_all_hand_after_gaining_judgment", func(t *testing.T) {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
			t.Fatal(err)
		}

		game.State.CurrentTurn = 0
		game.State.TurnStage = model.TurnStageActionExecution

		p1 := game.State.Players["p1"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		p1.Crystal = 1
		p1.Hand = []model.Card{
			{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
			{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		}

		testutils.MustHandleAction(t, game, model.PlayerAction{
			PlayerID: "p1",
			Type:     model.CmdSkill,
			SkillID:  "arbiter_balance",
		})
		testutils.RequireChoicePrompt(t, game, "p1", "arbiter_balance_mode")
		testutils.MustHandleAction(t, game, model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{0},
		})

		if got := p1.Tokens["judgment"]; got != 1 {
			t.Fatalf("expected balance to add 1 judgment before branch 0, got %d", got)
		}
		if got := len(p1.Hand); got != 0 {
			t.Fatalf("expected branch 0 to discard all hand cards, got %d", got)
		}
	})

	t.Run("branch_one_draws_to_hand_limit_and_adds_team_gem", func(t *testing.T) {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		game.State.Deck = rules.InitDeck()
		if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
			t.Fatal(err)
		}

		game.State.CurrentTurn = 0
		game.State.TurnStage = model.TurnStageActionExecution

		p1 := game.State.Players["p1"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		p1.Crystal = 1
		p1.MaxHand = 5
		p1.Tokens["judgment"] = 2
		p1.Hand = []model.Card{
			{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		}

		testutils.MustHandleAction(t, game, model.PlayerAction{
			PlayerID: "p1",
			Type:     model.CmdSkill,
			SkillID:  "arbiter_balance",
		})
		testutils.RequireChoicePrompt(t, game, "p1", "arbiter_balance_mode")
		testutils.MustHandleAction(t, game, model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{1},
		})

		if got := p1.Tokens["judgment"]; got != 3 {
			t.Fatalf("expected balance to add 1 judgment before branch 1, got %d", got)
		}
		if got := len(p1.Hand); got != 5 {
			t.Fatalf("expected branch 1 to draw to max hand 5, got %d", got)
		}
		if got := game.State.RedGems; got != 1 {
			t.Fatalf("expected branch 1 to add 1 team gem, got %d", got)
		}
	})
}

func TestArbiterForcedDoomsday_IgnoresTauntAndClearsItAfterResolution(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Hero", "hero", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens = map[string]int{
		"judgment": 4,
	}
	p1.Field = []*model.FieldCard{{
		Card:     model.Card{ID: "taunt", Name: "挑衅", Type: model.CardTypeMagic, Element: model.ElementLight},
		OwnerID:  p1.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectHeroTaunt,
	}}

	game.Drive()
	if !game.IsActionSelectionWindow() {
		t.Fatalf("expected action selection window after forced doomsday startup, got %s", game.RuntimeStateLabel())
	}

	if got := p1.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"]; got != 1 {
		t.Fatalf("expected forced doomsday pending set, got %d", got)
	}
	state := game.BuildActionSelectionOptions("p1", p1)
	if state.ActionRuleMode != engine.ActionSelectionRuleForceSkillAsMagic {
		t.Fatalf("expected forced doomsday to override taunt restriction")
	}
	if state.ActionRuleMode == engine.ActionSelectionRuleForceAttackOrSkip {
		t.Fatalf("expected taunt not to control prompt while forced doomsday is pending")
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "arbiter_doomsday",
		TargetID: "p3",
	})

	if got := p1.Tokens["judgment"]; got != 0 {
		t.Fatalf("expected forced doomsday to clear judgment, got %d", got)
	}
	if got := testutils.CountFieldEffect(p1, model.EffectHeroTaunt); got != 0 {
		t.Fatalf("expected taunt to be cleared after forced doomsday consumes action phase, got %d", got)
	}
}
