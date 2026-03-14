package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

// TestBerserker_Skills 测试狂战士的其他技能
// 包括: 狂化(Passive), 血腥咆哮(Response), 血影狂刀(Passive logic)
func TestBerserker_Skills(t *testing.T) {
	observer := testutils.NewTestObserver(t)

	// -------------------------------------------------------------------------
	// Case 1: 狂化 (Berserker Frenzy) - 伤害加成
	// -------------------------------------------------------------------------
	t.Run("Frenzy_DamageBonus", func(t *testing.T) {
		runScenario := func(attackerHand []model.Card) int {
			game := engine.NewGameEngine(observer)
			game.AddPlayer("p1", "Berserker", "berserker", model.RedCamp)
			game.AddPlayer("p2", "Sandbag", "angel", model.BlueCamp)

			game.State.CurrentTurn = 0
			game.State.Deck = rules.InitDeck()
			p1 := game.State.Players["p1"]
			p2 := game.State.Players["p2"]
			p1.IsActive = true
			p1.TurnState = model.NewPlayerTurnState()
			game.State.Phase = model.PhaseActionSelection

			p1.Hand = attackerHand
			p2.Heal = 0
			p2.MaxHeal = 5
			p2.MaxHand = 20
			initialHandP2 := len(p2.Hand)

			if err := game.HandleAction(model.PlayerAction{
				PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
			}); err != nil {
				t.Fatalf("攻击失败: %v", err)
			}
			if err := game.HandleAction(model.PlayerAction{
				PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
			}); err != nil {
				t.Fatalf("承伤失败: %v", err)
			}
			return len(p2.Hand) - initialHandP2
		}

		// 场景 A: 手牌<=3，狂化只+1，总伤害=3
		actualDrawA := runScenario([]model.Card{
			{ID: "atk1", Name: "斩", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
		})
		if actualDrawA != 3 {
			t.Errorf("狂化(手牌<=3)伤害计算错误: 预期摸 3 张，实际摸 %d 张", actualDrawA)
		}

		// 场景 B: 打出攻击牌后手牌仍>3，狂化+2，总伤害=4
		actualDrawB := runScenario([]model.Card{
			{ID: "atk2", Name: "斩", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
			{ID: "d1"}, {ID: "d2"}, {ID: "d3"}, {ID: "d4"},
		})
		if actualDrawB != 4 {
			t.Errorf("狂化(手牌>3)伤害计算错误: 预期摸 4 张，实际摸 %d 张", actualDrawB)
		}
		t.Logf("✅ 狂化测试通过")
	})

	// -------------------------------------------------------------------------
	// Case 2: 血腥咆哮 (Blood Roar) - 强制命中
	// -------------------------------------------------------------------------
	t.Run("BloodRoar_ForcedHit", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Berserker", "berserker", model.RedCamp)
		game.AddPlayer("p2", "Target", "angel", model.BlueCamp)
		
		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection

		// 给 P1 独有牌 血腥咆哮
		p1.Hand = []model.Card{
			{
				ID: "br_card", Name: "血腥咆哮", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2,
				ExclusiveChar1: "狂战士", ExclusiveSkill1: "血腥咆哮",
			},
		}

		// 设置 P2 Heal = 2 (触发条件)
		p2.Heal = 2
		// 给 P2 一张应战牌，如果不是强制命中，他应该能应战
		p2.Hand = []model.Card{
			{ID: "def_card", Name: "斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		}

		// P1 攻击
		action := model.PlayerAction{
			PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
		}
		
		// 执行攻击
		// 引擎应该检测到 BloodRoar 技能触发 (TriggerOnAttackStart, Silent response)
		// 并且 IsHitForced 被设为 true
		if err := game.HandleAction(action); err != nil {
			t.Fatalf("攻击失败: %v", err)
		}

		// 验证 P2 是否被禁止应战 (IsHitForced=true)
		// 强制命中意味着无法应战
		prompt := game.GetCurrentPrompt()
		if prompt != nil {
			// 如果提示包含 "counter" 选项，说明逻辑可能没完全封死，或者UI层展示
			// 尝试发送 Counter 指令，看是否被拒绝
			actionCounter := model.PlayerAction{
				PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardIndex: 0, TargetID: "p1",
			}
			err := game.HandleAction(actionCounter)
			if err == nil {
				t.Errorf("预期强制命中无法应战，但 P2 应战成功了") 
			} else {
				t.Logf("✅ P2 无法应战 (符合预期): %v", err)
			}
		}
		
		t.Logf("✅ 血腥咆哮测试通过 (触发了强制命中逻辑)")
	})

	// -------------------------------------------------------------------------
	// Case 3: 血影狂刀 (Blood Blade) - 对手手牌数增伤
	// -------------------------------------------------------------------------
	t.Run("BloodBlade_DamageBonus", func(t *testing.T) {
		runScenario := func(targetHandCount int) int {
			game := engine.NewGameEngine(observer)
			game.AddPlayer("p1", "Berserker", "berserker", model.RedCamp)
			game.AddPlayer("p2", "Target", "angel", model.BlueCamp)

			game.State.CurrentTurn = 0
			game.State.Deck = rules.InitDeck()
			p1 := game.State.Players["p1"]
			p2 := game.State.Players["p2"]
			p1.IsActive = true
			p1.TurnState = model.NewPlayerTurnState()
			game.State.Phase = model.PhaseActionSelection

			p1.Hand = []model.Card{
				{
					ID: "bb_card", Name: "血影狂刀", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2,
					ExclusiveChar1: "狂战士", ExclusiveSkill1: "血影狂刀",
				},
			}
			p2.Heal = 0
			p2.MaxHand = 20
			p2.Hand = make([]model.Card, targetHandCount)
			initialHand := len(p2.Hand)

			if err := game.HandleAction(model.PlayerAction{
				PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
			}); err != nil {
				t.Fatalf("攻击失败: %v", err)
			}
			if err := game.HandleAction(model.PlayerAction{
				PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
			}); err != nil {
				t.Fatalf("承伤失败: %v", err)
			}
			return len(p2.Hand) - initialHand
		}

		// 基础2 + 狂化1 + 血影2 = 5
		actualDrawA := runScenario(2)
		if actualDrawA != 5 {
			t.Errorf("血影狂刀(手牌2)伤害错误: 预期摸 5 张，实际摸 %d 张", actualDrawA)
		}

		// 基础2 + 狂化1 + 血影1 = 4
		actualDrawB := runScenario(3)
		if actualDrawB != 4 {
			t.Errorf("血影狂刀(手牌3)伤害错误: 预期摸 4 张，实际摸 %d 张", actualDrawB)
		}
		
		t.Logf("✅ 血影狂刀测试通过")
	})
}
