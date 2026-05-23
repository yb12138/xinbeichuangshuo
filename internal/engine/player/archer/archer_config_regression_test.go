package archer_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestArcher_PiercingShotDiscard_IsPublicReveal(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	p1.Hand = []model.Card{
		{ID: "atk", Name: "箭", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		{ID: "spell", Name: "水镜", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	p2.Hand = []model.Card{
		{ID: "counter", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("archer attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: testutils.PlayableCardID(t, game, "p2", 0), TargetID: "p3",
	}); err != nil {
		t.Fatalf("counter response failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm piercing shot failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0},
	}); err != nil {
		t.Fatalf("discard for piercing shot failed: %v", err)
	}

	reveal := testutils.FindPublicDiscardReveal(obs, "p1")
	if reveal == nil {
		t.Fatalf("expected piercing shot discard to emit a public discard reveal event")
	}
	if len(p1.Hand) != 0 {
		t.Fatalf("expected piercing shot discard to consume the spell card, got %d cards", len(p1.Hand))
	}
}

func TestArcher_PiercingShot_NotPromptedOnHit(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	p1.Hand = []model.Card{
		{ID: "atk", Name: "箭", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		{ID: "spell", Name: "水镜", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("archer attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take response failed: %v", err)
	}

	if game.State.PendingInterrupt != nil &&
		game.State.PendingInterrupt.Type == model.InterruptResponseSkill &&
		game.State.PendingInterrupt.PlayerID == "p1" &&
		testutils.InterruptHasSkillID(game.State.PendingInterrupt, "piercing_shot") {
		t.Fatalf("piercing_shot should not dispatch on hit branch")
	}
}

func TestArcher_PiercingShot_PromptedWhenMissByShieldBlock(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	p1.Hand = []model.Card{
		{ID: "atk", Name: "箭", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		{ID: "spell", Name: "水镜", Type: model.CardTypeMagic, Element: model.ElementWater},
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
			Hook:     model.FieldHookOnDamaged,
			Duration: 1,
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("archer attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take response failed: %v", err)
	}

	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if !testutils.InterruptHasSkillID(game.State.PendingInterrupt, "piercing_shot") {
		t.Fatalf("expected piercing_shot in response skills when attack is blocked by shield, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
}

func TestArcher_LightningArrow_DisablesCounterButAllowsDefend(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	p1.Hand = []model.Card{
		{ID: "thunder-atk", Name: "雷矢", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "holy-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "counter", Name: "雷刃", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("lightning arrow attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: testutils.PlayableCardID(t, game, "p2", 1), TargetID: "p3",
	}); err == nil {
		t.Fatalf("expected lightning arrow to forbid counter response")
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"defend"}, CardID: testutils.PlayableCardID(t, game, "p2", 0),
	}); err != nil {
		t.Fatalf("expected lightning arrow attack to still allow defend: %v", err)
	}

	if len(p2.Hand) != 1 || p2.Hand[0].Name != "雷刃" {
		t.Fatalf("expected only holy light to be consumed on defend, remaining hand=%+v", p2.Hand)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected defend path to resolve cleanly, got pending interrupt %+v", game.State.PendingInterrupt)
	}
}

func TestArcher_PreciseShot_OptionalResponseConfirm(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	p1.Hand = []model.Card{{
		ID:              "precise-shot-card",
		Name:            "精准射击",
		Type:            model.CardTypeAttack,
		Element:         model.ElementWind,
		Damage:          2,
		ExclusiveChar1:  "archer",
		ExclusiveSkill1: "精准射击",
	}}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("precise shot attack failed: %v", err)
	}

	// 攻击宣言时应弹出响应技能中断（精准射击为可选响应）
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected precise shot response prompt after attack declared, got no interrupt")
	}
	if game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected InterruptResponseSkill, got %+v", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected response prompt for p1, got %s", game.State.PendingInterrupt.PlayerID)
	}
	if !testutils.InterruptHasSkillID(game.State.PendingInterrupt, "precise_shot") {
		t.Fatalf("expected precise_shot in response skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	// 玩家确认发动精准射击
	testutils.ChooseResponseSkillByID(t, game, "p1", "precise_shot")
	game.Drive()

	// 强制命中：p2 摸牌（伤害=2-1=1）
	if len(p2.Hand) != 1 {
		t.Fatalf("expected precise shot force-hit for 1 damage draw, got hand=%d", len(p2.Hand))
	}
}

func TestArcher_PreciseShot_OptionalResponseSkip(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	p1.Hand = []model.Card{{
		ID:              "precise-shot-card",
		Name:            "精准射击",
		Type:            model.CardTypeAttack,
		Element:         model.ElementWind,
		Damage:          2,
		ExclusiveChar1:  "archer",
		ExclusiveSkill1: "精准射击",
	}}
	p2.Hand = []model.Card{
		{ID: "holy-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("precise shot attack failed: %v", err)
	}

	// 攻击宣言时应弹出响应技能中断
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected precise shot response prompt after attack declared")
	}
	if !testutils.InterruptHasSkillID(game.State.PendingInterrupt, "precise_shot") {
		t.Fatalf("expected precise_shot in response skills")
	}

	// 玩家选择跳过精准射击
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdCancel,
	}); err != nil {
		t.Fatalf("skip precise shot failed: %v", err)
	}
	game.Drive()

	// 跳过精准射击后走正常攻击流程：目标应能选择响应方式
	if len(game.State.CombatStack) == 0 {
		t.Fatalf("expected combat stack to have entries after skipping precise shot (normal attack flow)")
	}
}

func TestArcher_PreciseShot_ForceHitSkipsShield(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	p1.Hand = []model.Card{{
		ID:              "precise-shot-card",
		Name:            "精准射击",
		Type:            model.CardTypeAttack,
		Element:         model.ElementWind,
		Damage:          2,
		ExclusiveChar1:  "archer",
		ExclusiveSkill1: "精准射击",
	}}
	p2.Field = append(p2.Field, &model.FieldCard{
		Card:   model.Card{ID: "shield-1", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
		Mode:   model.FieldEffect,
		Effect: model.EffectShield,
	})
	redGemsBefore := game.State.RedGems

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("precise shot attack failed: %v", err)
	}

	// 确认发动精准射击
	testutils.ChooseResponseSkillByID(t, game, "p1", "precise_shot")
	game.Drive()

	if game.State.RedGems <= redGemsBefore {
		t.Fatalf("expected precise shot force-hit to count as hit and add gem, gems before=%d after=%d", redGemsBefore, game.State.RedGems)
	}
	if !p2.HasFieldEffect(model.EffectShield) {
		t.Fatalf("expected shield to remain when force-hit ignores shield")
	}
}

func TestArcher_PreciseShot_DamageReductionScopedToAttackModifier(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	card := model.Card{
		ID:              "precise-shot-card",
		Name:            "精准射击",
		Type:            model.CardTypeAttack,
		Element:         model.ElementWind,
		Damage:          2,
		ExclusiveChar1:  "archer",
		ExclusiveSkill1: "精准射击",
	}

	game.ApplyNextAttackDamageRule(p1.ID, "archer_precise_shot_damage_delta", "precise_shot", -1, model.RuleLifeThisEffectChain)
	modifiedDamage := game.ApplyAttackDamageModifiers(p1, p2, 2, model.Action{
		SourceID: p1.ID,
		TargetID: p2.ID,
		Type:     model.ActionAttack,
		Card:     &card,
	})
	if modifiedDamage != 1 {
		t.Fatalf("expected precise shot attack modifier to deal 1 damage, got %d", modifiedDamage)
	}

	p1.TurnState.SkillFlowState["precise_shot_confirmed"] = 1
	untaggedDamage := game.ApplyAttackDamageModifiers(p1, p2, 2, model.Action{
		SourceID: p1.ID,
		TargetID: p2.ID,
		Type:     model.ActionAttack,
		Card:     &card,
	})
	if untaggedDamage != 2 {
		t.Fatalf("expected untagged later attack not to inherit precise shot reduction, got %d", untaggedDamage)
	}
}
