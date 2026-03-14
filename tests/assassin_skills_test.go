package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

func TestAssassin_Skills(t *testing.T) {
	observer := testutils.NewTestObserver(t)

	// -------------------------------------------------------------------------
	// Case 1: 反噬 (Backlash) - 受伤让攻击者摸牌
	// -------------------------------------------------------------------------
	t.Run("Backlash_AttackerDraws", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Assassin", "assassin", model.RedCamp)
		game.AddPlayer("p2", "Attacker", "berserker", model.BlueCamp)
		
		game.State.CurrentTurn = 1 // P2 Turn
		game.State.Deck = rules.InitDeck()
		// p1 := game.State.Players["p1"] // Unused
		p2 := game.State.Players["p2"]
		p2.IsActive = true
		p2.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection

		// P2 攻击 P1
		p2.Hand = []model.Card{{ID: "atk", Name: "斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2}}
		// initialHandSize := len(p2.Hand) // Unused

		action := model.PlayerAction{PlayerID: "p2", Type: model.CmdAttack, TargetID: "p1", CardIndex: 0}
		game.HandleAction(action)
		
		// P1 承受
		// P2 Damage = 2.
		// P1 Heal = 0 (assumed default). Damage = 2 -> Draw 2 cards.
		// TriggerOnDamageTaken (Backlash) triggers. P2 draws 1 card.
		
		game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdRespond, ExtraArgs: []string{"take"}})
		game.Drive() // 结算伤害 -> TriggerOnDamageTaken -> Backlash

		// 验证 P2 手牌
		// P2 Hand initial 1. Played 1 -> 0.
		// Backlash -> 1.
		if len(p2.Hand) != 1 {
			t.Errorf("反噬未触发: P2应有1张牌，实际 %d", len(p2.Hand))
		}
		t.Logf("✅ 反噬测试通过")
	})

	// -------------------------------------------------------------------------
	// Case 2: 潜行 (Stealth) - 启动技进入潜行
	// -------------------------------------------------------------------------
	t.Run("Stealth_EnterState", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Assassin", "assassin", model.RedCamp)
		game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)
		
		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseStartup

		p1.Gem = 1

		// 启动技需要先进入 Startup 中断，再通过 Select 确认
		game.Drive()
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
			t.Fatalf("未进入启动技能中断")
		}
		if len(game.State.PendingInterrupt.SkillIDs) == 0 || game.State.PendingInterrupt.SkillIDs[0] != "stealth" {
			t.Fatalf("启动技能列表不包含 stealth: %+v", game.State.PendingInterrupt.SkillIDs)
		}

		// 选择第 1 个启动技能（0-based 索引）
		action := model.PlayerAction{
			PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0},
		}
		if err := game.HandleAction(action); err != nil {
			t.Fatalf("确认潜行失败: %v", err)
		}

		// 验证 Gem 消耗
		if p1.Gem != 0 {
			t.Errorf("潜行未消耗宝石")
		}
		// 验证是否获得潜行效果
		hasStealth := false
		for _, fc := range p1.Field {
			if fc.Effect == model.EffectStealth {
				hasStealth = true
				break
			}
		}
		if !hasStealth {
			t.Errorf("未获得潜行状态")
		}
		t.Logf("✅ 潜行测试通过")
	})
}
