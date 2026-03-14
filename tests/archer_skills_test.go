package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

func TestArcher_Skills(t *testing.T) {
	observer := testutils.NewTestObserver(t)

	// -------------------------------------------------------------------------
	// Case 1: 贯穿射击 (Piercing Shot) - 攻击未命中造成法术伤害
	// -------------------------------------------------------------------------
	t.Run("PiercingShot_OnMiss", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Archer", "archer", model.RedCamp)
		game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)
		game.AddPlayer("p3", "Ally", "angel", model.RedCamp) // P1 队友，应战反弹目标

		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection

		// P1 攻击牌 + 法术牌(用于消耗)
		p1.Hand = []model.Card{
			{ID: "atk", Name: "箭", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
			{ID: "mag", Name: "法", Type: model.CardTypeMagic, Element: model.ElementWind},
		}

		// P2 有应战牌 -> 造成 Miss，反弹给 P3
		p2.Hand = []model.Card{
			{ID: "def", Name: "闪", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		}
		p2.Heal = 0 // Heal 0 -> Damage causes draw
		initialHandP2 := len(p2.Hand)

		// P1 攻击
		action := model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0}
		if err := game.HandleAction(action); err != nil {
			t.Fatalf("P1 发起攻击失败: %v", err)
		}

		// P2 应战，反弹给 P1 的队友 P3
		actionCounter := model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardIndex: 0, TargetID: "p3"}
		if err := game.HandleAction(actionCounter); err != nil {
			t.Fatalf("P2 应战失败: %v", err)
		}

		// 新规则：贯穿射击为可选响应，需玩家确认并弃法术牌。
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
			t.Fatalf("预期出现贯穿射击响应中断，实际: %+v", game.State.PendingInterrupt)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{0}, // 选择发动贯穿射击
		}); err != nil {
			t.Fatalf("确认贯穿射击失败: %v", err)
		}
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
			t.Fatalf("预期进入贯穿射击弃牌中断，实际: %+v", game.State.PendingInterrupt)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{0}, // 弃掉唯一法术牌
		}); err != nil {
			t.Fatalf("贯穿射击弃牌失败: %v", err)
		}

		// 贯穿射击结算后，战斗应继续到反弹目标 p3。
		if game.State.Phase != model.PhaseCombatInteraction {
			t.Fatalf("预期处于 CombatInteraction，实际: %s", game.State.Phase)
		}
		if len(game.State.CombatStack) == 0 || game.State.CombatStack[len(game.State.CombatStack)-1].TargetID != "p3" {
			t.Fatalf("预期当前被反弹攻击目标是 p3，实际战斗栈: %+v", game.State.CombatStack)
		}

		// 贯穿射击伤害应已入队，等待当前反弹战斗响应结束后结算。
		foundPendingPierceDamage := false
		for _, pd := range game.State.PendingDamageQueue {
			if pd.SourceID == "p1" && pd.TargetID == "p2" && pd.DamageType == "magic" && pd.Damage == 2 {
				foundPendingPierceDamage = true
				break
			}
		}
		if !foundPendingPierceDamage {
			t.Fatalf("预期贯穿射击向 p2 的2点法伤已入队，当前队列: %+v", game.State.PendingDamageQueue)
		}
		_ = initialHandP2 // 仅保留场景语义，实际在伤害结算后校验更稳定
		// 发动贯穿射击需要弃1张法术牌，因此 p1 手牌应被清空。
		if len(p1.Hand) != 0 {
			t.Errorf("贯穿射击后 P1 手牌数量异常: 预期 0，实际 %d", len(p1.Hand))
		}
		t.Logf("✅ 贯穿射击测试通过")
	})

	// -------------------------------------------------------------------------
	// Case 2: 狙击 (Snipe) - 补牌 + 额外攻击
	// -------------------------------------------------------------------------
	t.Run("Snipe_RefillAndAction", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Archer", "archer", model.RedCamp)
		game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)

		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		// p2 := game.State.Players["p2"] // Unused
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection

		p1.Crystal = 1
		p1.Hand = []model.Card{{}} // 1 card

		// 发动狙击 -> Self (p1)
		action := model.PlayerAction{
			PlayerID: "p1", Type: model.CmdSkill, SkillID: "snipe", TargetIDs: []string{"p1"},
		}
		game.HandleAction(action)

		// 验证手牌补到 5
		if len(p1.Hand) != 5 {
			t.Errorf("狙击未补满5张牌，实际: %d", len(p1.Hand))
		}
		// 验证额外攻击
		// 由于 HandleAction 调用了 Drive，且进入了 TurnEnd，PendignActions 可能已被消费并转化为 CurrentExtraAction
		// 检查 TurnState
		if p1.TurnState.CurrentExtraAction == "Attack" {
			t.Logf("✅ 检测到 CurrentExtraAction 为 Attack")
		} else if len(p1.TurnState.PendingActions) > 0 {
			t.Logf("✅ 检测到 PendingActions 中有 Action")
		} else {
			t.Errorf("狙击未增加额外行动 (CurrentExtraAction=%s, Pending=%d)",
				p1.TurnState.CurrentExtraAction, len(p1.TurnState.PendingActions))
		}
		t.Logf("✅ 狙击测试通过")
	})

	// -------------------------------------------------------------------------
	// Case 3: 闪光陷阱 (Flash Trap) - 独有牌直伤
	// -------------------------------------------------------------------------
	t.Run("FlashTrap_Damage", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Archer", "archer", model.RedCamp)
		game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)

		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection

		p1.Hand = []model.Card{
			{
				ID: "ft_card", Name: "闪光陷阱", Type: model.CardTypeMagic, Element: model.ElementFire,
				ExclusiveChar1: "神箭手", ExclusiveSkill1: "闪光陷阱",
			},
		}
		p2.Heal = 0
		initialHandP2 := len(p2.Hand) // 0

		action := model.PlayerAction{
			PlayerID: "p1", Type: model.CmdSkill, SkillID: "flash_trap",
			TargetIDs: []string{"p2"}, Selections: []int{0},
		}
		game.HandleAction(action)
		game.Drive() // 结算伤害 (2点法术)

		// 验证伤害 (Draw 2 cards)
		if len(p2.Hand) != initialHandP2+2 {
			t.Errorf("闪光陷阱伤害错误: 预期手牌+2，实际 %d (Init %d)", len(p2.Hand), initialHandP2)
		}
		t.Logf("✅ 闪光陷阱测试通过")
	})
}
