package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

// TestAngel_Skills 测试天使的其他技能
// 包括: 天使羁绊(Passive), 风之洁净(Action), 天使之歌(Startup), 天使之墙(Action)
func TestAngel_Skills(t *testing.T) {
	observer := testutils.NewTestObserver(t)

	// -------------------------------------------------------------------------
	// Case 1: 天使之墙 (Angel Wall) - 放置圣盾
	// -------------------------------------------------------------------------
	t.Run("AngelWall_PlaceShield", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Angel", "angel", model.RedCamp)
		game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)

		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.TurnStage = model.TurnStageActionExecution

		// 给 P1 发一张独有牌 "天使之墙" (在ExclusiveCards中定义为 angel_wall，但这里模拟手牌匹配)
		// 注意：RequiresExclusive: true, 需要牌名匹配 Skill Title 或 ExclusiveCards 配置
		// Angel 的 ExclusiveCards: ["angel_wall"]
		// 技能 Angel Wall 的 Title: "天使之墙"
		// 牌名需要匹配 Exclusive 检查逻辑。
		// 简单起见，我们构造一张满足 MatchExclusive 的牌
		card := model.Card{
			ID: "c1", Name: "天使之墙", Type: model.CardTypeMagic, Element: model.ElementLight,
			ExclusiveChar1: "angel", ExclusiveSkill1: "天使之墙",
		}
		p1.Hand = []model.Card{card}

		// 执行技能
		action := model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSkill,
			SkillID:    "angel_wall",
			TargetIDs:  []string{"p1"}, // 给自己套盾
			Selections: []int{0},       // 弃这张牌
		}

		if err := game.HandleAction(action); err != nil {
			t.Fatalf("天使之墙发动失败: %v", err)
		}

		// 验证场上是否有圣盾
		hasShield := false
		for _, fc := range p1.Field {
			if fc.Effect == model.EffectShield {
				hasShield = true
				break
			}
		}
		if !hasShield {
			t.Errorf("发动天使之墙后，场上未找到圣盾效果")
		}
		t.Logf("✅ 天使之墙测试通过")
	})

	// -------------------------------------------------------------------------
	// Case 2: 风之洁净 (Angel Cleanse) - 移除基础效果
	// -------------------------------------------------------------------------
	t.Run("AngelCleanse_RemoveBuff", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Angel", "angel", model.RedCamp)
		game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)

		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.TurnStage = model.TurnStageActionExecution

		// 给 P1 风系牌
		p1.Hand = []model.Card{
			{ID: "c1", Name: "风牌", Type: model.CardTypeMagic, Element: model.ElementWind},
		}

		// 给 P2 上个虚弱 (Weak)
		p2.AddFieldCard(&model.FieldCard{
			Mode: model.FieldEffect, Effect: model.EffectWeak, SourceID: "p1",
		})

		// 执行技能
		action := model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSkill,
			SkillID:    "angel_cleanse",
			TargetIDs:  []string{"p2"}, // 目标 P2
			Selections: []int{0},       // 弃风牌
		}

		if err := game.HandleAction(action); err != nil {
			t.Fatalf("风之洁净发动失败: %v", err)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{0},
		}); err != nil {
			t.Fatalf("处理风之洁净移除目标失败: %v", err)
		}

		// 验证 P2 虚弱是否移除
		hasWeak := false
		for _, fc := range p2.Field {
			if fc.Effect == model.EffectWeak {
				hasWeak = true
				break
			}
		}
		if hasWeak {
			t.Errorf("发动风之洁净后，目标身上的虚弱未被移除")
		}
		t.Logf("✅ 风之洁净测试通过")
	})

	// -------------------------------------------------------------------------
	// Case 3: 天使之歌 (Angel Song) - 回合开始响应技移除Buff
	// -------------------------------------------------------------------------
	t.Run("AngelSong_Startup", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Angel", "angel", model.RedCamp)
		game.AddPlayer("p2", "Friend", "berserker", model.RedCamp)

		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.TurnStage = model.TurnStageActionStart // 回合开始阶段

		// P1 需要 1 水晶
		p1.Crystal = 1
		// P2 身上有中毒
		p2.AddFieldCard(&model.FieldCard{
			Mode: model.FieldEffect, Effect: model.EffectPoison, SourceID: "enemy",
		})

		// Startup 阶段会自动检查回合开始响应技并推送中断
		game.Drive()

		// 检查中断
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
			t.Fatalf("预期产生响应技能中断，实际: %v", game.State.PendingInterrupt)
		}
		if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "angel_song" {
			t.Fatalf("预期响应技能为 angel_song，实际: %+v", game.State.PendingInterrupt.SkillIDs)
		}

		// 先确认发动响应技（选择 angel_song）
		action := model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{0},
		}

		if err := game.HandleAction(action); err != nil {
			t.Fatalf("天使之歌确认失败: %v", err)
		}

		// AngelSong 会继续进入 Choice 让玩家选择要移除的基础效果
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
			t.Fatalf("预期进入天使之歌选择中断，实际: %v", game.State.PendingInterrupt)
		}
		ctxData, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
		if !ok {
			t.Fatalf("天使之歌选择上下文类型错误: %T", game.State.PendingInterrupt.Context)
		}
		if got, _ := ctxData["choice_type"].(string); got != "basic_effect_pick" {
			t.Fatalf("预期 choice_type=basic_effect_pick，实际: %q", got)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{0},
		}); err != nil {
			t.Fatalf("天使之歌选择移除目标失败: %v", err)
		}

		// 移除基础效果后会继续触发天使羁绊的治疗选择。
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
			t.Fatalf("预期进入天使羁绊选择中断，实际: %v", game.State.PendingInterrupt)
		}
		ctxData, ok = game.State.PendingInterrupt.Context.(map[string]interface{})
		if !ok {
			t.Fatalf("天使羁绊选择上下文类型错误: %T", game.State.PendingInterrupt.Context)
		}
		if got, _ := ctxData["choice_type"].(string); got != "angel_bond_heal_target" {
			t.Fatalf("预期 choice_type=angel_bond_heal_target，实际: %q", got)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{0},
		}); err != nil {
			t.Fatalf("天使羁绊选择治疗目标失败: %v", err)
		}

		// 验证水晶消耗
		if p1.Crystal != 0 {
			t.Errorf("天使之歌应消耗1水晶，实际剩余: %d", p1.Crystal)
		}

		// 验证中毒移除
		hasPoison := false
		for _, fc := range p2.Field {
			if fc.Effect == model.EffectPoison {
				hasPoison = true
				break
			}
		}
		if hasPoison {
			t.Errorf("天使之歌未移除目标的中毒效果")
		}
		if game.State.PendingInterrupt != nil {
			t.Fatalf("天使之歌结算后不应残留中断，实际: %v", game.State.PendingInterrupt)
		}
		if !(game.State.TurnStage == model.TurnStageActionExecution &&
			game.State.CombatStage == model.CombatStageNone &&
			game.State.Subflow == model.SubflowNone) {
			t.Fatalf("天使之歌结算后应回到行动选择窗口，实际: turn=%s combat=%s subflow=%s", game.State.TurnStage, game.State.CombatStage, game.State.Subflow)
		}
		t.Logf("✅ 天使之歌测试通过")
	})

	// -------------------------------------------------------------------------
	// Case 4: 天使羁绊 (Angel Bond) - 被动治疗
	// -------------------------------------------------------------------------
	t.Run("AngelBond_PassiveHeal", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "Angel", "angel", model.RedCamp)
		game.AddPlayer("p2", "Friend", "berserker", model.RedCamp)

		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.TurnStage = model.TurnStageActionExecution

		// 场景：移除 Buff 触发羁绊
		// 给 P2 放置一个虚弱，然后移除它
		p2.AddFieldCard(&model.FieldCard{
			Mode: model.FieldEffect, Effect: model.EffectWeak, SourceID: "enemy",
		})

		// 记录 P2 初始治疗 (Heal)
		p2.Heal = 1
		p2.MaxHeal = 5 // 确保能加血

		t.Logf("模拟由天使移除 P2 的虚弱效果...")
		// 由天使本人移除基础效果，应该触发 TimingOnBuffRemoved -> Angel Bond
		game.RemoveFieldCardBy("p2", model.EffectWeak, "p1")

		// AngelBond 当前为“弹框选目标”实现，先选择目标 p2
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
			t.Fatalf("预期触发天使羁绊选择中断，实际: %v", game.State.PendingInterrupt)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{1}, // [p1, p2] -> 选择 p2
		}); err != nil {
			t.Fatalf("处理天使羁绊治疗目标失败: %v", err)
		}

		// 验证 P2 Heal +1
		if p2.Heal != 2 {
			t.Errorf("天使羁绊未触发：移除Buff后 P2 Heal 应为 2，实际为 %d", p2.Heal)
		}
		t.Logf("✅ 天使羁绊(移除Buff)测试通过")

		// 场景：使用圣盾触发羁绊
		// Angel 手动打出一张【圣盾】牌
		p1.Hand = []model.Card{
			{ID: "shield_c", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
		}

		// 重置 Heal
		p2.Heal = 1

		// P1 对 P2 使用圣盾
		action := model.PlayerAction{
			PlayerID:  "p1",
			Type:      model.CmdMagic,
			CardID:    testutils.PlayableCardID(t, game, "p1", 0),
			TargetIDs: []string{"p2"}, // 给 P2 贴膜
		}

		if err := game.HandleAction(action); err != nil {
			t.Fatalf("使用圣盾失败: %v", err)
		}

		// 使用圣盾后也应触发天使羁绊弹框
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
			t.Fatalf("预期触发天使羁绊选择中断，实际: %v", game.State.PendingInterrupt)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{1}, // 选择 p2
		}); err != nil {
			t.Fatalf("处理圣盾后的天使羁绊选择失败: %v", err)
		}

		// Angel Bond: 使用 [圣盾] 时，目标 +1 Heal
		// P2 应该 Heal +1 -> 2
		if p2.Heal != 2 {
			t.Errorf("天使羁绊未触发：使用圣盾后 P2 Heal 应为 2，实际为 %d", p2.Heal)
		}
		t.Logf("✅ 天使羁绊(使用圣盾)测试通过")
	})
}
