package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

// TestMagicBullet_ChainAndDamage 测试魔弹的传递和伤害递增机制
func TestMagicBullet_ChainAndDamage(t *testing.T) {
	// 1. 初始化游戏引擎
	game := engine.NewGameEngine(testutils.NewTestObserver(t))

	// 初始化牌库，防止摸牌出错
	game.State.Deck = rules.InitDeck()

	// 添加4名玩家: P1(Red), P2(Blue), P3(Red), P4(Blue)
	game.AddPlayer("p1", "Player1", "Hero", model.RedCamp)
	game.AddPlayer("p2", "Player2", "Mage", model.BlueCamp)
	game.AddPlayer("p3", "Player3", "Warrior", model.RedCamp)
	game.AddPlayer("p4", "Player4", "Archer", model.BlueCamp)

	game.StartGame()

	p1 := game.State.Players["p1"]
	p3 := game.State.Players["p3"]
	p4 := game.State.Players["p4"]

	// 2. 准备卡牌
	// P1: 一张魔弹 (发起者)
	p1.Hand = []model.Card{
		{ID: "c1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	// P4: 一张魔弹 (用于传递，当前规则魔弹默认先指向 p4)
	p4.Hand = []model.Card{
		{ID: "c4", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementFire, Damage: 2},
	}
	// P3: 没有魔弹，只能承伤
	p3.Hand = []model.Card{}
	p3.Heal = 4 // 提供可用治疗点数

	// 3. P1 回合，使用魔弹
	game.State.CurrentTurn = 0 // P1's turn
	game.State.Phase = model.PhaseActionSelection
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()

	t.Logf("Step 1: P1 发动魔弹（当前规则默认右手方向，首个目标应为 P4）")
	// 正常来说 Magic 需要指定 target，但代码中魔弹逻辑是 findNextMagicBulletTarget
	// 不过 HandleAction CmdMagic 还是需要一个初始 TargetID 才能过校验，
	// PerformMagic 内部会重新计算 nextTargetID (如果是魔弹的话)
	// 让我们看一眼 PerformMagic 的实现：
	// case "魔弹": nextTargetID := e.findNextMagicBulletTarget(player.ID) ...
	// 所以初始指定的 TargetID 其实会被忽略，或者说我们应该指定第一个合法的？
	// 为了通过 CmdMagic 的校验，我们指定 P2。

	err := game.HandleAction(model.PlayerAction{
		Type:      model.CmdMagic,
		PlayerID:  "p1",
		TargetID:  "p2",
		CardIndex: 0, // P1's Magic Bullet
	})
	if err != nil {
		t.Fatalf("P1 使用魔弹失败: %v", err)
	}

	// 4. 验证 P4 收到中断
	if game.State.PendingInterrupt == nil {
		t.Fatalf("期望 P2 收到中断，但 PendingInterrupt 为 nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Errorf("中断类型错误，期望 MagicMissile，实际 %s", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != "p4" {
		t.Errorf("中断目标错误，期望 p4，实际 %s", game.State.PendingInterrupt.PlayerID)
	}

	// 验证当前伤害
	chain := game.State.MagicBulletChain
	if chain == nil {
		t.Fatalf("魔弹链条未创建")
	}
	if chain.CurrentDamage != 2 {
		t.Errorf("初始伤害错误，期望 2，实际 %d", chain.CurrentDamage)
	}

	t.Logf("Step 2: P4 使用手中的魔弹进行传递 (Counter)")
	// P4 的魔弹在索引 0
	err = game.HandleAction(model.PlayerAction{
		Type:      model.CmdRespond,
		PlayerID:  "p4",
		ExtraArgs: []string{"counter"},
		CardIndex: 0,
	})
	if err != nil {
		t.Fatalf("P4 传递魔弹失败: %v", err)
	}

	// 5. 验证 P4 魔弹被消耗
	if len(p4.Hand) != 0 {
		t.Errorf("P4 的魔弹未被消耗")
	}

	// 6. 验证魔弹传递给 P3 (P4 的下一个敌方是 P3)
	if game.State.PendingInterrupt == nil {
		t.Fatalf("期望 P3 收到中断，但 nil")
	}
	if game.State.PendingInterrupt.PlayerID != "p3" {
		t.Errorf("传递目标错误，期望 p3，实际 %s", game.State.PendingInterrupt.PlayerID)
	}

	// 验证伤害递增
	if chain.CurrentDamage != 3 {
		t.Errorf("传递后伤害未递增，期望 3，实际 %d", chain.CurrentDamage)
	}

	t.Logf("Step 3: P3 选择承受伤害 (Take)")
	err = game.HandleAction(model.PlayerAction{
		Type:      model.CmdRespond,
		PlayerID:  "p3",
		ExtraArgs: []string{"take"},
	})
	if err != nil {
		t.Fatalf("P3 承受伤害失败: %v", err)
	}

	// 当前伤害流程会先进入“是否使用治疗抵消”中断，这里选择不使用
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("期望 P3 进入治疗选择中断，实际: %+v", game.State.PendingInterrupt)
	}
	err = game.HandleAction(model.PlayerAction{
		Type:       model.CmdSelect,
		PlayerID:   "p3",
		Selections: []int{0},
	})
	if err != nil {
		t.Fatalf("P3 治疗选择失败: %v", err)
	}

	// 7. 验证 P3 受到伤害：当前引擎语义为“伤害=摸牌”，并非扣 Heal
	// 预期手牌: 0 + 3 = 3
	if len(p3.Hand) != 3 {
		t.Errorf("P3 手牌错误，期望 3（承受3点伤害摸3张），实际 %d", len(p3.Hand))
	}
	// 且选择不使用治疗后，治疗储备不变
	if p3.Heal != 4 {
		t.Errorf("P3 治疗储备错误，期望 4，实际 %d", p3.Heal)
	}

	// 8. 验证链条清除
	if game.State.MagicBulletChain != nil {
		t.Errorf("魔弹结算后链条未清除")
	}
}

// TestMagicBullet_Defend 测试圣盾/圣光抵挡魔弹
func TestMagicBullet_Defend(t *testing.T) {
	game := engine.NewGameEngine(testutils.NewTestObserver(t))
	game.State.Deck = rules.InitDeck()

	game.AddPlayer("p1", "Player1", "Hero", model.RedCamp)
	game.AddPlayer("p2", "Player2", "Mage", model.BlueCamp)
	game.StartGame()

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]

	// P1: 魔弹
	p1.Hand = []model.Card{
		{ID: "c1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	// P2: 圣光 (手牌)
	p2.Hand = []model.Card{
		{ID: "c2", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}
	p2.Heal = 3

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	p1.IsActive = true

	t.Logf("Step 1: P1 对 P2 使用魔弹")
	game.HandleAction(model.PlayerAction{
		Type:      model.CmdMagic,
		PlayerID:  "p1",
		TargetID:  "p2",
		CardIndex: 0,
	})

	t.Logf("Step 2: P2 使用圣光抵挡")
	// 查找圣光索引
	cardIdx := -1
	for i, c := range p2.Hand {
		if c.Name == "圣光" {
			cardIdx = i
			break
		}
	}

	err := game.HandleAction(model.PlayerAction{
		Type:      model.CmdRespond,
		PlayerID:  "p2",
		ExtraArgs: []string{"defend"},
		CardIndex: cardIdx,
	})
	if err != nil {
		t.Fatalf("P2 抵挡失败: %v", err)
	}

	// 验证无伤害
	if p2.Heal != 3 {
		t.Errorf("P2 血量不应变化，期望 3，实际 %d", p2.Heal)
	}

	// 验证圣光被消耗
	if len(p2.Hand) != 0 {
		t.Errorf("P2 的圣光未被消耗")
	}

	if game.State.MagicBulletChain != nil {
		t.Errorf("链条未清除")
	}
}

// TestMagicBullet_CounterEndsWhenRoundCovered
// 回归：当当前传递会补齐“本轮全员已参与”时，链条应直接结束，不能继续进入下一轮。
func TestMagicBullet_CounterEndsWhenRoundCovered(t *testing.T) {
	game := engine.NewGameEngine(testutils.NewTestObserver(t))
	game.State.Deck = rules.InitDeck()

	game.AddPlayer("p1", "Player1", "Hero", model.RedCamp)
	game.AddPlayer("p2", "Player2", "Mage", model.BlueCamp)
	if err := game.StartGame(); err != nil {
		t.Fatalf("start game failed: %v", err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Hand = []model.Card{
		{ID: "mb1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "mb2", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementFire, Damage: 2},
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()

	if err := game.HandleAction(model.PlayerAction{
		Type:      model.CmdMagic,
		PlayerID:  "p1",
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("p1 magic bullet failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected p2 to respond magic bullet, got: %+v", game.State.PendingInterrupt)
	}

	// 两人局下，p2 传递后将覆盖全员，本轮应直接结束。
	if err := game.HandleAction(model.PlayerAction{
		Type:      model.CmdRespond,
		PlayerID:  "p2",
		ExtraArgs: []string{"counter"},
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("p2 counter failed: %v", err)
	}

	if game.State.MagicBulletChain != nil {
		t.Fatalf("magic bullet chain should end after full round coverage")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("pending interrupt should be cleared, got: %+v", game.State.PendingInterrupt)
	}
	if len(p2.Hand) != 0 {
		t.Fatalf("p2 counter card should be consumed")
	}
}
