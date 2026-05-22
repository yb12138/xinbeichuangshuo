package blood_priestess_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/testutils"
	"testing"

	bloodpriestesspkg "starcup-engine/internal/engine/player/blood_priestess"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func bloodPriestessTestCard(id string, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        "测试牌",
		Type:        model.CardTypeAttack,
		Element:     ele,
		Faction:     "血",
		Damage:      2,
		Description: "test",
	}
}

func TestBloodPriestessSharedLife_DrawBeforePlaceOverflowThenApply(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormBloodPriestessBleeding
	p1.Hand = []model.Card{
		bloodPriestessTestCard("h1", model.ElementFire),
		bloodPriestessTestCard("h2", model.ElementWater),
		bloodPriestessTestCard("h3", model.ElementWind),
		bloodPriestessTestCard("h4", model.ElementThunder),
		bloodPriestessTestCard("h5", model.ElementEarth),
		bloodPriestessTestCard("h6", model.ElementDark),
	}
	p1.ExclusiveCards = append(p1.ExclusiveCards, testutils.MakeStarterBloodSharedLifeCard(p1))
	game.State.Deck = rules.InitDeck()

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bp_shared_life",
	})
	testutils.RequireChoicePrompt(t, game, "p1", "bp_shared_life_target")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose shared-life target failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected overflow discard interrupt before placing shared life")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{6, 7},
	})

	holder, fc := bloodpriestesspkg.FindSharedLife(engine.NewRoleChoiceRuntime(game), p1)
	if holder == nil || fc == nil {
		t.Fatalf("expected shared life effect placed after overflow resolution")
	}
	if holder.ID != "p1" {
		t.Fatalf("expected shared life holder p1, got %s", holder.ID)
	}
	if got := game.State.RedMorale; got != 13 {
		t.Fatalf("expected morale down by 2 before shared-life placement, got %d", got)
	}
	if got := len(p1.Hand); got != 6 {
		t.Fatalf("expected final hand count 6, got %d", got)
	}
	if got := game.GetMaxHand(p1); got != 7 {
		t.Fatalf("expected bleed-form shared-life max hand 7, got %d", got)
	}
}

func TestBloodPriestessSharedLife_ChoiceAndDrawStayInActionExecution(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormBloodPriestessBleeding
	p1.Hand = []model.Card{
		bloodPriestessTestCard("h1", model.ElementFire),
		bloodPriestessTestCard("h2", model.ElementWater),
		bloodPriestessTestCard("h3", model.ElementWind),
		bloodPriestessTestCard("h4", model.ElementThunder),
	}
	p1.ExclusiveCards = append(p1.ExclusiveCards, testutils.MakeStarterBloodSharedLifeCard(p1))
	game.State.Deck = rules.InitDeck()

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bp_shared_life",
	})
	testutils.RequireChoicePrompt(t, game, "p1", "bp_shared_life_target")
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected shared-life target choice in action execution, got %s", game.State.TurnStage)
	}

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose shared-life target failed: %v", err)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected shared-life draw to resume action execution, got %s", game.State.TurnStage)
	}
	holder, fc := bloodpriestesspkg.FindSharedLife(engine.NewRoleChoiceRuntime(game), p1)
	if holder == nil || fc == nil {
		t.Fatalf("expected shared life placed after draw continuation")
	}
}

func TestBloodPriestessSharedLife_OverflowDiscardResumesActionExecution(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormBloodPriestessBleeding
	p1.Hand = []model.Card{
		bloodPriestessTestCard("h1", model.ElementFire),
		bloodPriestessTestCard("h2", model.ElementWater),
		bloodPriestessTestCard("h3", model.ElementWind),
		bloodPriestessTestCard("h4", model.ElementThunder),
		bloodPriestessTestCard("h5", model.ElementEarth),
		bloodPriestessTestCard("h6", model.ElementDark),
	}
	p1.ExclusiveCards = append(p1.ExclusiveCards, testutils.MakeStarterBloodSharedLifeCard(p1))
	game.State.Deck = rules.InitDeck()

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bp_shared_life",
	})
	testutils.RequireChoicePrompt(t, game, "p1", "bp_shared_life_target")
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected shared-life target choice in action execution, got %s", game.State.TurnStage)
	}

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose shared-life target failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected overflow discard interrupt after draw")
	}
	if game.State.Subflow != model.SubflowDiscardSelection {
		t.Fatalf("expected overflow to enter discard selection, got %s", game.State.Subflow)
	}

	data, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("expected discard context map")
	}
	if resumeStage := model.ParseResumePointTurnStage(data["draw_resume_phase"]); resumeStage != model.TurnStageActionExecution {
		t.Fatalf("expected overflow discard to carry action-execution resume stage, got %s", resumeStage)
	}

	if err := game.ConfirmDiscard("p1", []int{6, 7}); err != nil {
		t.Fatalf("resolve overflow discard failed: %v", err)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected overflow discard to resume action execution, got %s", game.State.TurnStage)
	}
	holder, fc := bloodpriestesspkg.FindSharedLife(engine.NewRoleChoiceRuntime(game), p1)
	if holder == nil || fc == nil {
		t.Fatalf("expected shared life placed after overflow recovery")
	}
}

func TestBloodPriestessBleeding_EnterOnMoraleLossAndReleaseOnActionEndLowHand(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.MaxHand = 6
	p1.Hand = []model.Card{
		bloodPriestessTestCard("a1", model.ElementFire),
		bloodPriestessTestCard("a2", model.ElementWater),
		bloodPriestessTestCard("a3", model.ElementWind),
		bloodPriestessTestCard("a4", model.ElementThunder),
		bloodPriestessTestCard("a5", model.ElementEarth),
		bloodPriestessTestCard("a6", model.ElementDark),
		bloodPriestessTestCard("a7", model.ElementLight),
		bloodPriestessTestCard("a8", model.ElementFire),
	}
	damageOverflowCtx := game.BuildContext(p1, nil, model.TimingActionDuring, nil)
	damageOverflowCtx.Flags["FromDamageDraw"] = true
	game.CheckHandLimitCtx(p1, damageOverflowCtx)
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{6, 7},
	})

	if got := p1.Form; got != model.FormBloodPriestessBleeding {
		t.Fatalf("expected enter bleed form, got %q", got)
	}
	if got := p1.Heal; got != 1 {
		t.Fatalf("expected heal +1 on entering bleed form, got %d", got)
	}

	// 手动降到2张，验证不会立刻脱离；而是在一次行动完整结束后再检查。
	p1.Hand = p1.Hand[:2]
	if got := p1.Form; got != model.FormBloodPriestessBleeding {
		t.Fatalf("expected still remain in bleed form before action end, got %q", got)
	}
	game.BeginActionSummary("skill", "p2", "测试行动", nil)
	game.State.TurnStage = model.TurnStageActionEnd
	p2 := game.State.Players["p2"]
	// 通过 HandlePostActionEndEffects 触发 TimingPostActionEnd hooks（包括流血形态脱离）
	game.HandlePostActionEndEffects(p2, model.ActionMagic)
	game.FinalizeActionSummaryIfIdle()
	if got := p1.Form; got != "" {
		t.Fatalf("expected release from bleed form on action end at hand<3, got %q", got)
	}
}

func TestBloodPriestessBleeding_TurnStartSelfDamageBeforeBuff(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p1.Form = model.FormBloodPriestessBleeding
	game.State.Deck = rules.InitDeck()
	p1.Hand = []model.Card{
		bloodPriestessTestCard("s1", model.ElementFire),
		bloodPriestessTestCard("s2", model.ElementWater),
		bloodPriestessTestCard("s3", model.ElementWind),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnStart
	game.Drive()

	// 承伤摸1：3 -> 4
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected turn-start self-damage draw 1 card, hand=4 got %d", got)
	}
}

func TestBloodPriestessBloodSorrow_TransferThenRemove(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		bloodPriestessTestCard("c1", model.ElementFire),
		bloodPriestessTestCard("c2", model.ElementWater),
	}
	p1.ExclusiveCards = append(p1.ExclusiveCards, testutils.MakeStarterBloodSharedLifeCard(p1))
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	// 1) 先放置同生共死到 p2。
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bp_shared_life",
	})
	testutils.RequireChoicePrompt(t, game, "p1", "bp_shared_life_target")
	// 目标列表顺序按 PlayerOrder: p1,p2,p3；这里选 p2(index=1)。
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose shared-life target p2 failed: %v", err)
	}
	game.Drive() // 触发 deferred 放置同生共死
	holder, _ := bloodpriestesspkg.FindSharedLife(engine.NewRoleChoiceRuntime(game), p1)
	if holder == nil || holder.ID != "p2" {
		t.Fatalf("expected shared life holder p2 before blood sorrow, got %+v", holder)
	}
	// 让上限足够高，避免血之哀伤自伤摸牌触发爆牌弃牌中断，聚焦转移/移除逻辑本身。
	p1.Form = model.FormBloodPriestessBleeding

	// 2) 启动血之哀伤，选择“转移”到 p3。
	game.State.CurrentTurn = 0
	p1.IsActive = true
	ctx := game.BuildContext(p1, nil, model.TimingTurnStart, nil)
	h := &bloodpriestesspkg.BloodSorrowHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected blood sorrow can use when shared life exists")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute blood sorrow failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bp_blood_sorrow_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil { // 转移分支
		t.Fatalf("choose blood sorrow transfer mode failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bp_blood_sorrow_target")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{2}}); err != nil { // 选 p3
		t.Fatalf("choose blood sorrow transfer target p3 failed: %v", err)
	}
	game.Drive() // 先结算自伤，再执行延迟的转移后续
	holder, _ = bloodpriestesspkg.FindSharedLife(engine.NewRoleChoiceRuntime(game), p1)
	if holder == nil || holder.ID != "p3" {
		t.Fatalf("expected shared life holder p3 after transfer, got %+v", holder)
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptStartupSkill {
		if err := game.SkipStartupSkill("p1"); err != nil {
			t.Fatalf("skip startup prompt after transfer failed: %v", err)
		}
	}

	// 3) 再次发动血之哀伤，选择“移除”。
	game.State.CurrentTurn = 0
	p1.IsActive = true
	ctx = game.BuildContext(p1, nil, model.TimingTurnStart, nil)
	if !h.CanUse(ctx) {
		t.Fatalf("expected blood sorrow can use before remove branch")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute blood sorrow(remove) failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bp_blood_sorrow_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 移除分支
		t.Fatalf("choose blood sorrow remove mode failed: %v", err)
	}
	game.Drive() // 先结算自伤，再执行延迟的移除后续
	if game.State.PendingInterrupt != nil && engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		data, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
		discardCount := runtimeutil.ToIntContextValue(data["discard_count"])
		if discardCount <= 0 {
			discardCount = 1
		}
		picks := make([]int, 0, discardCount)
		for i := 0; i < discardCount && i < len(p1.Hand); i++ {
			picks = append(picks, i)
		}
		testutils.MustHandleAction(t, game, model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: picks,
		})
	}
	holder, fc := bloodpriestesspkg.FindSharedLife(engine.NewRoleChoiceRuntime(game), p1)
	if holder != nil || fc != nil {
		t.Fatalf("expected shared life removed, holder=%+v card=%+v", holder, fc)
	}
	if !p1.HasExclusiveCard(p1.Character.ID, "同生共死") {
		t.Fatalf("expected shared-life card restored to exclusive zone after remove branch")
	}
}

func TestBloodPriestessBloodSorrow_Remove_ShouldEnterBleedWhenDamageCausesMoraleLoss(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p1.Form = ""
	p1.Hand = []model.Card{
		bloodPriestessTestCard("h1", model.ElementFire),
		bloodPriestessTestCard("h2", model.ElementWater),
		bloodPriestessTestCard("h3", model.ElementWind),
		bloodPriestessTestCard("h4", model.ElementThunder),
	}
	if !p1.HasExclusiveCard(p1.Character.ID, "同生共死") {
		p1.ExclusiveCards = append(p1.ExclusiveCards, testutils.MakeStarterBloodSharedLifeCard(p1))
	}
	// 保证同生共死处于生效状态，此时普通形态下巫女手牌上限应为4。
	card, ok := p1.ConsumeExclusiveCard(p1.Character.ID, "同生共死")
	if !ok {
		t.Fatalf("expected starter shared life card in exclusive zone")
	}
	if err := bloodpriestesspkg.PlaceSharedLife(engine.NewRoleChoiceRuntime(game), p1, p2, card); err != nil {
		t.Fatalf("place shared life failed: %v", err)
	}
	if got := game.GetMaxHand(p1); got != 4 {
		t.Fatalf("expected max hand=4 with shared life in normal form, got %d", got)
	}
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	// 发动血之哀伤并选择“移除同生共死”。
	ctx := game.BuildContext(p1, nil, model.TimingTurnStart, nil)
	h := &bloodpriestesspkg.BloodSorrowHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected blood sorrow can use when shared life exists")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute blood sorrow failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bp_blood_sorrow_mode")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 移除
	})

	// 自伤2应先按上限4结算承伤摸牌并触发爆牌弃牌。
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected overflow discard interrupt after blood sorrow self-damage, got %+v", game.State.PendingInterrupt)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	})

	if got := p1.Form; got != model.FormBloodPriestessBleeding {
		t.Fatalf("expected enter bleed form after morale loss from blood sorrow self-damage, got %q", got)
	}
	if got := p1.Heal; got != 1 {
		t.Fatalf("expected +1 heal when entering bleed form, got %d", got)
	}
	if got := game.State.RedMorale; got != 13 {
		t.Fatalf("expected red morale loss 2 from overflow discard, got %d", got)
	}
	holder, fc := bloodpriestesspkg.FindSharedLife(engine.NewRoleChoiceRuntime(game), p1)
	if holder != nil || fc != nil {
		t.Fatalf("expected shared life removed after blood sorrow remove branch, holder=%+v card=%+v", holder, fc)
	}
}

func TestBloodPriestessSharedLife_FixedHandCapTargetExempt(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Lancer", "magic_lancer", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p2.Form = model.FormMagicLancerPhantom // 恒定手牌上限=5
	p1.Hand = []model.Card{
		bloodPriestessTestCard("f1", model.ElementFire),
		bloodPriestessTestCard("f2", model.ElementWater),
		bloodPriestessTestCard("f3", model.ElementWind),
	}
	if !p1.HasExclusiveCard(p1.Character.ID, "同生共死") {
		p1.ExclusiveCards = append(p1.ExclusiveCards, testutils.MakeStarterBloodSharedLifeCard(p1))
	}

	card, ok := p1.ConsumeExclusiveCard(p1.Character.ID, "同生共死")
	if !ok {
		t.Fatalf("expected starter shared life card in exclusive zone")
	}
	if err := bloodpriestesspkg.PlaceSharedLife(engine.NewRoleChoiceRuntime(game), p1, p2, card); err != nil {
		t.Fatalf("place shared life failed: %v", err)
	}

	// 普通形态：同生共死对固定上限目标不生效，对血之巫女自身照常生效。
	p1.Form = ""
	if got := game.GetMaxHand(p1); got != 4 {
		t.Fatalf("expected priestess max hand 4 in normal form with shared life, got %d", got)
	}
	if got := game.GetMaxHand(p2); got != 5 {
		t.Fatalf("expected fixed-cap target keep max hand 5, got %d", got)
	}

	// 流血形态：自身改为+1；目标仍应保持固定上限不变。
	p1.Form = model.FormBloodPriestessBleeding
	if got := game.GetMaxHand(p1); got != 7 {
		t.Fatalf("expected priestess max hand 7 in bleed form with shared life, got %d", got)
	}
	if got := game.GetMaxHand(p2); got != 5 {
		t.Fatalf("expected fixed-cap target still 5 in bleed form, got %d", got)
	}
}

func TestBloodPriestessBloodCurse_DiscardPromptAndConfirm(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		bloodPriestessTestCard("c0", model.ElementFire),
		bloodPriestessTestCard("c1", model.ElementWater),
		bloodPriestessTestCard("c2", model.ElementWind),
		bloodPriestessTestCard("c3", model.ElementThunder),
		bloodPriestessTestCard("c4", model.ElementEarth),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "bp_blood_curse",
		TargetIDs: []string{"p2"},
	})

	ctxData := testutils.RequireChoiceContext(t, game, "p1", "bp_curse_discard")
	if game.State.Subflow != model.SubflowDiscardSelection {
		t.Fatalf("expected blood curse delayed discard to enter discard subflow, got %s", game.State.Subflow)
	}
	if !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected blood curse delayed discard interrupt to be recognized as discard selection")
	}
	testutils.RequirePromptFlow(t, ctxData, "bp_blood_curse_discard", "cards")
	if _, ok := ctxData["selected_indices"]; ok {
		t.Fatalf("blood curse discard should store selections in prompt flow, got legacy selected_indices in %+v", ctxData)
	}
	prompt := game.GetCurrentPrompt()
	if prompt == nil {
		t.Fatalf("expected blood curse discard prompt")
	}
	if got := len(game.State.PendingDamageQueue); got != 0 {
		t.Fatalf("expected damage queue cleared before discard prompt, got %d pending entries", got)
	}
	if prompt.Type != model.PromptChooseCards {
		t.Fatalf("expected choose_cards prompt, got %s", prompt.Type)
	}
	if prompt.Min != 3 || prompt.Max != 3 {
		t.Fatalf("expected fixed discard count 3, got min=%d max=%d", prompt.Min, prompt.Max)
	}
	if got := len(prompt.Options); got != 5 {
		t.Fatalf("expected 5 discard options from hand, got %d", got)
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{4, 1, 3},
	})

	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected hand size 2 after discarding 3 cards, got %d", got)
	}
	if p1.Hand[0].ID != "c0" || p1.Hand[1].ID != "c2" {
		t.Fatalf("unexpected remaining cards: %+v", p1.Hand)
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptChoice {
		if ctx, ok := game.State.PendingInterrupt.Context.(map[string]interface{}); ok {
			if ct, _ := ctx["choice_type"].(string); ct == "bp_curse_discard" {
				t.Fatalf("blood curse discard interrupt should be finished after confirm")
			}
		}
	}
}

func TestBloodPriestessBloodCurse_DiscardAllWhenHandInsufficient(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		bloodPriestessTestCard("s0", model.ElementFire),
		bloodPriestessTestCard("s1", model.ElementWater),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "bp_blood_curse",
		TargetIDs: []string{"p2"},
	})

	ctxData := testutils.RequireChoiceContext(t, game, "p1", "bp_curse_discard")
	if game.State.Subflow != model.SubflowDiscardSelection {
		t.Fatalf("expected blood curse delayed discard-all to enter discard subflow, got %s", game.State.Subflow)
	}
	if !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected blood curse delayed discard-all interrupt to be recognized as discard selection")
	}
	testutils.RequirePromptFlow(t, ctxData, "bp_blood_curse_discard", "cards")
	if _, ok := ctxData["selected_indices"]; ok {
		t.Fatalf("blood curse discard should store selections in prompt flow, got legacy selected_indices in %+v", ctxData)
	}
	prompt := game.GetCurrentPrompt()
	if prompt == nil {
		t.Fatalf("expected blood curse discard prompt when hand<3")
	}
	if got := len(game.State.PendingDamageQueue); got != 0 {
		t.Fatalf("expected damage queue cleared before discard prompt, got %d pending entries", got)
	}
	if prompt.Min != 2 || prompt.Max != 2 {
		t.Fatalf("expected discard-all prompt min/max=2, got min=%d max=%d", prompt.Min, prompt.Max)
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	})

	if got := len(p1.Hand); got != 0 {
		t.Fatalf("expected all cards discarded when hand<3, got %d", got)
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptChoice {
		if ctx, ok := game.State.PendingInterrupt.Context.(map[string]interface{}); ok {
			if ct, _ := ctx["choice_type"].(string); ct == "bp_curse_discard" {
				t.Fatalf("blood curse discard interrupt should be finished when hand<3 flow ends")
			}
		}
	}
}
