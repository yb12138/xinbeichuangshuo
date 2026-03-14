package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

func TestMagicalGirl_Skills(t *testing.T) {
	observer := testutils.NewTestObserver(t)

	// -------------------------------------------------------------------------
	// Case 1: 毁灭风暴 (Destruction Storm) - AOE 伤害
	// -------------------------------------------------------------------------
	t.Run("DestructionStorm_AOE", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "MagicalGirl", "magical_girl", model.RedCamp)
		game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp)
		game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp)
		
		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p3 := game.State.Players["p3"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection
		p1.Gem = 1
		p2.Heal = 0
		p3.Heal = 0
		initialHandP2 := len(p2.Hand)
		initialHandP3 := len(p3.Hand)

		// 发动技能
		action := model.PlayerAction{
			PlayerID: "p1", Type: model.CmdSkill, SkillID: "destruction_storm",
			TargetIDs: []string{"p2", "p3"},
		}
		
		if err := game.HandleAction(action); err != nil {
			t.Fatalf("毁灭风暴发动失败: %v", err)
		}
		
		// 结算伤害 (AOE damage might need processPendingDamages loop)
		game.Drive()
		game.Drive() 
		
		// 验证伤害 (各受到 2 点法术伤害 -> 摸 2 张牌)
		if len(p2.Hand) != initialHandP2 + 2 || len(p3.Hand) != initialHandP3 + 2 {
			t.Errorf("毁灭风暴伤害错误: p2Hand=%d, p3Hand=%d", len(p2.Hand), len(p3.Hand))
		}
		t.Logf("✅ 毁灭风暴测试通过")
	})

	// -------------------------------------------------------------------------
	// Case 2: 魔爆冲击 (Magic Blast) - 弃法术加宝石
	// -------------------------------------------------------------------------
	t.Run("MagicBlast_GainGem", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "MagicalGirl", "magical_girl", model.RedCamp)
		game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp)
		game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp)
		
		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p3 := game.State.Players["p3"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection

		// p1 需弃1张法术牌发动技能
		p1.Hand = []model.Card{
			{ID: "m1", Name: "法", Type: model.CardTypeMagic, Element: model.ElementFire},
		}
		// p2 有法术牌可弃，p3 无法术牌（取消后将受2点法术伤害 -> 摸2）
		p2.Hand = []model.Card{
			{ID: "m2", Name: "法", Type: model.CardTypeMagic, Element: model.ElementWater},
		}
		initialHandP3 := len(p3.Hand)
		initialRedGems := game.State.RedGems

		action := model.PlayerAction{
			PlayerID: "p1", Type: model.CmdSkill, SkillID: "magic_blast",
			TargetIDs: []string{"p2", "p3"},
			Selections: []int{0}, // 弃1张法术牌作为技能发动代价
		}
		
		if err := game.HandleAction(action); err != nil {
			t.Fatalf("魔爆冲击发动失败: %v", err)
		}

		// 目标1：p2 选择弃法术牌
		if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdSelect, Selections: []int{0}}); err != nil {
			t.Fatalf("p2 处理魔爆冲击失败: %v", err)
		}
		// 目标2：p3 放弃弃牌 -> 受2点法术伤害
		if err := game.HandleAction(model.PlayerAction{PlayerID: "p3", Type: model.CmdCancel}); err != nil {
			t.Fatalf("p3 处理魔爆冲击失败: %v", err)
		}
		// 施法者阶段：可选弃1张牌，这里选择跳过
		if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel}); err != nil {
			t.Fatalf("p1 处理可选弃牌失败: %v", err)
		}

		// 继续驱动伤害结算
		game.Drive()
		game.Drive()

		if game.State.RedGems != initialRedGems+1 {
			t.Errorf("魔爆冲击未给我方战绩区+1宝石")
		}
		if len(p3.Hand) != initialHandP3+2 {
			t.Errorf("p3 未正确受到2点法术伤害（摸2），当前手牌=%d", len(p3.Hand))
		}
		t.Logf("✅ 魔爆冲击测试通过")
	})
}
