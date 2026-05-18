package blade_master_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归：风之剑圣第3次主动攻击触发【圣剑】后，强制命中应无视目标场上圣盾。
func TestBladeMaster_HolySword_ThirdAttackForceHitIgnoresShield(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	// 避免行动结束后弹出无关响应中断，聚焦圣剑命中链。
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p1.TurnState.UsedSkillCounts["sword_shadow"] = 1
	p1.Gem = 0
	p1.Crystal = 0

	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "火斩2", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a3", Name: "火斩3", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	model.AppendAttackAction(p1, "test-extra-1")
	model.AppendAttackAction(p1, "test-extra-2")
	p2.Heal = 0

	// 前两次攻击正常命中，堆到第3次攻击条件。
	for i := 0; i < 2; i++ {
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
		}); err != nil {
			t.Fatalf("attack #%d failed: %v", i+1, err)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
		}); err != nil {
			t.Fatalf("respond take #%d failed: %v", i+1, err)
		}
		if game.State.PendingInterrupt != nil {
			t.Fatalf("unexpected interrupt after attack #%d: %+v", i+1, game.State.PendingInterrupt)
		}
	}

	// 第3次攻击前给目标挂圣盾，验证圣剑是否正确无视。
	p2.Field = append(p2.Field, &model.FieldCard{
		Card:   model.Card{ID: "shield-1", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
		Mode:   model.FieldEffect,
		Effect: model.EffectShield,
	})
	redGemsBeforeThird := game.State.RedGems

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("third attack failed: %v", err)
	}

	if p1.TurnState.AttackCount != 3 {
		t.Fatalf("expected attack count=3, got %d", p1.TurnState.AttackCount)
	}
	if game.State.RedGems <= redGemsBeforeThird {
		t.Fatalf("expected holy_sword third attack to count as hit and add gem, gems before=%d after=%d", redGemsBeforeThird, game.State.RedGems)
	}
	if !p2.HasFieldEffect(model.EffectShield) {
		t.Fatalf("expected shield to remain when holy_sword attack ignores shield")
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.PlayerID == "p2" {
		t.Fatalf("expected holy_sword forced hit to skip target response prompt, got %+v", game.State.PendingInterrupt)
	}
}

// 回归：风之剑圣第3次攻击结束后，通过 TimingOnActionEnd hook 自动触发圣剑摸X弃X。
// X=0 时不摸不弃，直接进入额外行动阶段。
func TestBladeMaster_HolySword_FullFlow_X0ResumesExtraAction(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	// 禁用风怒和剑影，避免额外行动干扰验证
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p1.TurnState.UsedSkillCounts["sword_shadow"] = 1
	p1.Gem = 0
	p1.Crystal = 0

	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "火斩2", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a3", Name: "火斩3", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	// 两个额外行动令牌：确保两次攻击后仍有额外行动可用
	model.AppendAttackAction(p1, "test-extra-1")
	model.AppendAttackAction(p1, "test-extra-2")
	p2.Heal = 0

	// 前两次攻击正常进行
	for i := 0; i < 2; i++ {
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
		}); err != nil {
			t.Fatalf("attack #%d failed: %v", i+1, err)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
		}); err != nil {
			t.Fatalf("respond take #%d failed: %v", i+1, err)
		}
	}

	// 第3次攻击：正常应战并承受伤害（圣剑强制命中）
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("third attack failed: %v", err)
	}

	// 第3次攻击命中后，目标不能应战（圣剑强制命中），应直接进入圣剑摸X弃X中断
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected holy sword draw interrupt for p1, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.Type != model.InterruptHolySwordDraw {
		t.Fatalf("expected InterruptHolySwordDraw type, got %s", game.State.PendingInterrupt.Type)
	}

	// 玩家选择 X=0（不摸不弃）
	if err := game.HandleInterruptAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("holy sword x=0 response failed: %v", err)
	}

	// 验证：中断已清除，进入额外行动阶段
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after holy sword x=0, got %+v", game.State.PendingInterrupt)
	}
	if game.State.TurnStage != model.TurnStageExtraAction {
		t.Fatalf("expected holy sword x=0 to resume extra action, got %s", game.State.TurnStage)
	}
}

// 回归：风之剑圣第3次攻击结束后，圣剑摸X弃X完整流程验证。
// X=1 时摸1张牌并弃1张牌，完成后进入额外行动阶段。
func TestBladeMaster_HolySword_FullFlow_DiscardResumesExtraAction(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	// 禁用风怒和剑影，避免额外行动干扰验证
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p1.TurnState.UsedSkillCounts["sword_shadow"] = 1
	p1.Gem = 0
	p1.Crystal = 0

	// 3张攻击牌，足够打3次攻击
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "火斩2", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a3", Name: "火斩3", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	// 额外行动令牌：确保3次攻击后仍有行动可执行
	model.AppendAttackAction(p1, "test-extra-1")
	model.AppendAttackAction(p1, "test-extra-2")
	p2.Heal = 0

	// 前两次攻击正常进行
	for i := 0; i < 2; i++ {
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
		}); err != nil {
			t.Fatalf("attack #%d failed: %v", i+1, err)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
		}); err != nil {
			t.Fatalf("respond take #%d failed: %v", i+1, err)
		}
	}

	handCountBeforeThird := len(p1.Hand)

	// 第3次攻击
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("third attack failed: %v", err)
	}

	// 验证弹出了圣剑摸X弃X中断
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected holy sword draw interrupt for p1, got %+v", game.State.PendingInterrupt)
	}

	// 玩家选择 X=1
	if err := game.HandleInterruptAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("holy sword x=1 response failed: %v", err)
	}

	// 验证：摸牌后出现弃牌中断
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt after holy sword x=1, got %+v", game.State.PendingInterrupt)
	}

	handAfterDraw := len(p1.Hand)
	// 第三次攻击消耗了最后一张牌（handCountBeforeThird-1），圣剑摸1张后应恢复到 handCountBeforeThird
	if handAfterDraw != handCountBeforeThird {
		t.Fatalf("expected hand count to be %d after draw (third attack consumed 1, then drew 1), got %d", handCountBeforeThird, handAfterDraw)
	}

	// 玩家弃掉1张牌
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("holy sword discard confirmation failed: %v", err)
	}

	// 验证最终状态：无残留中断、保持当前回合、正确进入额外行动阶段
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
