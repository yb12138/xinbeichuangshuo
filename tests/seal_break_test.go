package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

// TestSealer_SealBreak 测试封印师的封印破碎技能
// 封印破碎：[水晶] 将场上任意一张基础效果牌收入自己手中
func TestSealer_SealBreak(t *testing.T) {
	observer := testutils.NewTestObserver(t)
	game := engine.NewGameEngine(observer)

	t.Logf("========== 🚀 测试开始：封印师 封印破碎 ==========")

	// 1. 初始化玩家
	// P1 封印师 (Red)
	game.AddPlayer("p1", "Sealer", "sealer", model.RedCamp)
	// P2 受害者 (Blue)
	game.AddPlayer("p2", "Victim", "angel", model.BlueCamp)

	// 初始化牌库
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.Phase = model.PhaseActionSelection

	// 2. 准备状态
	// P1 需要 1 水晶
	p1.Crystal = 1
	p1.Gem = 0

	// P2 面前有一个【圣盾】效果
	shieldCard := &model.FieldCard{
		Card: model.Card{
			ID:      "shield_card",
			Name:    "圣盾",
			Type:    model.CardTypeMagic,
			Element: model.ElementLight,
		},
		OwnerID:  "p2",
		SourceID: "p2", // 假设自己上的
		Mode:     model.FieldEffect,
		Effect:   model.EffectShield,
		Trigger:  model.EffectTriggerOnDamaged, // 圣盾通常是OnDamaged或者特殊Trigger
	}
	p2.AddFieldCard(shieldCard)

	t.Logf("✅ [Setup] P1: Crystal=1. P2: Has Shield Effect.")

	// 3. P1 发动封印破碎 -> 目标 P2
	// 封印破碎是主动技能 (Action Skill)
	t.Logf("\n👉 [Step 1] P1 发动 [封印破碎] -> P2")
	actionSkill := model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "seal_break",
		TargetIDs: []string{"p2"},
	}

	if err := game.HandleAction(actionSkill); err != nil {
		t.Fatalf("技能发动失败: %v", err)
	}

	// 4. 验证结果
	// P1 水晶应 -1 (变为 0)
	if p1.Crystal != 0 {
		t.Errorf("❌ P1 水晶未扣除，剩余: %d", p1.Crystal)
	}

	// P2 场上的圣盾应该被移除
	hasShield := false
	for _, fc := range p2.Field {
		if fc.Effect == model.EffectShield {
			hasShield = true
			break
		}
	}
	if hasShield {
		t.Errorf("❌ P2 场上的圣盾未被移除")
	}

	// P1 手牌应该增加 1 张 (收入手中的圣盾)
	// 初始手牌为 0 (没发牌), 加上收回的 1 张 = 1
	foundShieldInHand := false
	for _, c := range p1.Hand {
		if c.Name == "圣盾" {
			foundShieldInHand = true
			break
		}
	}
	if !foundShieldInHand {
		t.Errorf("❌ P1 手中未找到收入的圣盾牌")
	}

	t.Logf("========== ✅ 测试成功：封印破碎技能逻辑验证通过 ==========\n")
}
