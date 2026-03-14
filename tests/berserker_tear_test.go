package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

// TestBerserker_Tear 测试狂战士的撕裂技能
// 撕裂：[宝石] 攻击命中后发动，本次攻击伤害额外+2
func TestBerserker_Tear(t *testing.T) {
	observer := testutils.NewTestObserver(t)
	game := engine.NewGameEngine(observer)

	t.Logf("\n========== 🚀 测试开始：狂战士 撕裂 ==========")

	// 1. 初始化玩家
	// P1 狂战士 (Red)
	game.AddPlayer("p1", "Berserker", "berserker", model.RedCamp)
	// P2 沙袋 (Blue)
	game.AddPlayer("p2", "Sandbag", "angel", model.BlueCamp) // 用天使做沙袋，血多点

	// 初始化牌库
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.Phase = model.PhaseActionSelection

	// 2. 准备状态
	// P1 需要 1 宝石，1 张攻击牌
	p1.Gem = 1
	p1.Crystal = 0
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "重斩", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
	}
	// P2 满血
	p2.Heal = 3
	p2.MaxHeal = 5 // 增加MaxHeal以避免Heal被cap到2 (如果Angel MaxHeal default 2)

	t.Logf("✅ [Setup] P1: Gem=1, Hand=1 Attack(2dmg). P2: Heal=3")

	// 3. P1 发起攻击
	t.Logf("\n👉 [Step 1] P1 发起攻击 -> P2")
	actionAtk := model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}
	if err := game.HandleAction(actionAtk); err != nil {
		t.Fatalf("攻击失败: %v", err)
	}

	// 4. P2 选择承受 (Take)
	// 这应该触发 攻击命中 -> 撕裂 (TriggerOnAttackHit)
	t.Logf("\n👉 [Step 2] P2 选择承受 (Take)")
	actionTake := model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}

	// 执行响应。如果逻辑正确，引擎应该检测到 TriggerOnAttackHit，
	// 发现 P1 有撕裂技能且满足条件，推送“可选响应”中断给 P1。
	if err := game.HandleAction(actionTake); err != nil {
		t.Fatalf("承受伤害失败: %v", err)
	}

	// 5. 先由 P1 选择是否发动撕裂
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("❌ 预期进入撕裂响应中断，实际: %+v", game.State.PendingInterrupt)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 选择发动撕裂
	}); err != nil {
		t.Fatalf("发动撕裂失败: %v", err)
	}

	// 6. 发动撕裂后，进入受伤方治疗选择
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("❌ 预期进入治疗选择中断，实际: %+v", game.State.PendingInterrupt)
	}
	t.Logf("✅ 撕裂技能触发，进入受伤方治疗选择")

	// 7. P2 选择不使用治疗（索引0），继续完成伤害结算
	t.Logf("\n👉 [Step 3] P2 选择不使用治疗")
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("治疗选择失败: %v", err)
	}

	// 8. 验证结果
	// P1 宝石应 -1 (变为 0)
	if p1.Gem != 0 {
		t.Errorf("❌ P1 宝石未扣除，剩余: %d", p1.Gem)
	}

	// 当前规则：伤害=基础2 + 狂化1 + 撕裂2 = 5；
	// 本步骤选择“不使用治疗”，因此治疗值不变，摸牌5张
	if p2.Heal != 3 {
		t.Errorf("❌ P2 治疗值错误，预期 3，实际 %d", p2.Heal)
	}
	if len(p2.Hand) != 5 {
		t.Errorf("❌ P2 手牌数错误，预期 5，实际 %d", len(p2.Hand))
	}

	t.Logf("\n========== ✅ 测试成功：撕裂技能逻辑验证通过 ==========\n")
}
