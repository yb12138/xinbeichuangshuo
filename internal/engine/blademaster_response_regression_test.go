package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

type noopObserver struct{}

func (noopObserver) OnGameEvent(event model.GameEvent) {}

// 回归测试：剑影在同回合后续攻击结束时应继续询问（直到本回合真正发动）
func TestBladeMaster_SwordShadow_ReAskOnEachAttackEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p2.Heal = 0

	// 两张非风系攻击牌（避免风怒追击干扰）；有蓝水晶满足剑影条件
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p1.Crystal = 1
	// 人为放一个额外攻击 token，确保同回合发生第二次攻击行动
	model.AppendAttackAction(p1, "test-token")

	// 第一次攻击
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("first attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("first take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after first attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "sword_shadow" {
		t.Fatalf("expected only sword_shadow after first attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	// 第一次不发动（跳过）
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdCancel,
	}); err != nil {
		t.Fatalf("skip first response failed: %v", err)
	}

	// 推进到第二次攻击行动
	game.Drive()

	// 第二次攻击
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("second attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("second take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt again after second attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "sword_shadow" {
		t.Fatalf("expected only sword_shadow after second attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
}

// 回归测试：风怒追击在同回合后续攻击结束时应继续询问（直到本回合真正发动）
func TestBladeMaster_WindFury_ReAskOnEachAttackEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// 三张风系攻击牌：前两张用于两次攻击，保留一张确保第二次攻击结束时仍满足风怒可用条件
	p1.Hand = []model.Card{
		{ID: "a1", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
		{ID: "a2", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
		{ID: "a3", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}
	p1.Gem = 0
	p1.Crystal = 0
	model.AppendAttackAction(p1, "test-token")

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("first attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("first take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after first attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after first attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdCancel,
	}); err != nil {
		t.Fatalf("skip first response failed: %v", err)
	}

	game.Drive()

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("second attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("second take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt again after second attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after second attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
}

func TestBladeMaster_WindFury_StillTriggersWithoutRemainingWindAttack(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// 只有这一张风系攻击牌；按文档，风怒追击仍应在攻击结束时触发，
	// 随后若没有合法风系攻击可执行，可在额外行动阶段选择“无法行动”跳过。
	p1.Hand = []model.Card{
		{ID: "a1", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("wind attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after attack end, got %+v", game.State.PendingInterrupt)
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.ConfirmResponseSkill("p1", "wind_fury"); err != nil {
		t.Fatalf("confirm wind_fury failed: %v", err)
	}

	game.Drive()
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected extra action to enter action execution window, got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected extra attack constraint, got %q", p1.TurnState.CurrentExtraAction)
	}
	if p1.TurnState.CurrentExtraElement == nil || len(p1.TurnState.CurrentExtraElement) != 1 || p1.TurnState.CurrentExtraElement[0] != model.ElementWind {
		t.Fatalf("expected wind-only extra attack constraint, got %+v", p1.TurnState.CurrentExtraElement)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCannotAct,
	}); err != nil {
		t.Fatalf("expected cannot_act to skip unusable wind-only extra action, got %v", err)
	}
	if p1.TurnState.CurrentExtraAction != "" {
		t.Fatalf("expected extra action constraint cleared after cannot_act, got %q", p1.TurnState.CurrentExtraAction)
	}
}

func TestBladeMaster_HolySwordDraw_X0ResumesExtraAction(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.AttackCount = 3
	game.State.TurnStage = model.TurnStageActionExecution

	if !game.triggerHolySwordDrawIfNeeded(p1) {
		t.Fatalf("expected holy sword draw interrupt to trigger")
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptHolySwordDraw {
		t.Fatalf("expected holy sword draw interrupt, got %+v", game.State.PendingInterrupt)
	}

	if err := game.handleHolySwordDrawResponse(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("holy sword x=0 response failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after holy sword x=0, got %+v", game.State.PendingInterrupt)
	}
	if game.State.TurnStage != model.TurnStageExtraAction {
		t.Fatalf("expected holy sword x=0 to resume extra action, got %s", game.State.TurnStage)
	}
}

func TestBladeMaster_HolySwordDiscardResumesExtraAction(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.AttackCount = 3
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	model.AppendAttackAction(p1, "holy-sword-followup")
	game.State.CurrentTurn = 0

	if !game.triggerHolySwordDrawIfNeeded(p1) {
		t.Fatalf("expected holy sword draw interrupt to trigger")
	}
	if err := game.handleHolySwordDrawResponse(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("holy sword x=1 response failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt after holy sword x=1, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("holy sword discard confirmation failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after holy sword discard, got %+v", game.State.PendingInterrupt)
	}
	if game.State.CurrentTurn != 0 {
		t.Fatalf("expected holy sword aftermath to keep current turn when extra action remains, got turn=%d", game.State.CurrentTurn)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected holy sword discard to continue into extra action execution window, got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected extra attack constraint to be restored, got %q", p1.TurnState.CurrentExtraAction)
	}
	if len(p1.TurnState.PendingActions) != 0 {
		t.Fatalf("expected extra action token to be consumed into current extra action, got %d pending", len(p1.TurnState.PendingActions))
	}
}

func TestBladeMaster_GaleSlash_DisablesCounterButAllowsDefend(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
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
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0
	p2.Hand = []model.Card{
		{ID: "holy-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "counter-wind", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}
	p2.AddFieldCard(&model.FieldCard{
		Mode:   model.FieldEffect,
		Effect: model.EffectShield,
		Card:   model.Card{ID: "shield", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
	})
	p1.Hand = []model.Card{{
		ID:              "gale-slash-card",
		Name:            "列风技",
		Type:            model.CardTypeAttack,
		Element:         model.ElementWind,
		Damage:          2,
		ExclusiveChar1:  "blade_master",
		ExclusiveSkill1: "列风技",
	}}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("gale slash attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardIndex: 1, TargetID: "p3",
	}); err == nil {
		t.Fatalf("expected gale slash to forbid counter response")
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"defend"}, CardIndex: 0,
	}); err != nil {
		t.Fatalf("expected gale slash to still allow defend: %v", err)
	}

	if !p2.HasFieldEffect(model.EffectShield) {
		t.Fatalf("expected shield to remain because gale slash only ignores it on take")
	}
	if len(p2.Hand) != 1 || p2.Hand[0].Name != "风斩" {
		t.Fatalf("expected only holy light to be consumed on defend, remaining hand=%+v", p2.Hand)
	}
	if p2.Heal != 0 {
		t.Fatalf("expected defend to prevent damage draw, got heal=%d", p2.Heal)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected defend path to resolve cleanly, got pending interrupt %+v", game.State.PendingInterrupt)
	}
}
