package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

func TestBladeMaster_WindFury_ExtraAttack(t *testing.T) {
	// 1. 初始化带日志的观察者
	observer := testutils.NewTestObserver(t)
	game := engine.NewGameEngine(observer)

	t.Logf("\n========== 🚀 测试开始：风之剑圣 风怒追击 ==========")

	// 2. 添加玩家
	game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp)
	game.AddPlayer("p2", "Sandbag", "berserker", model.BlueCamp)

	// 3. 强制设置初始状态
	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck() // 【修复】初始化牌库
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.Phase = model.PhaseActionSelection

	// 4. 准备手牌：两张风系攻击牌
	cardWind1 := model.Card{ID: "c1", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2}
	cardWind2 := model.Card{ID: "c2", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2}
	p1.Hand = []model.Card{cardWind1, cardWind2}

	t.Logf("✅ [Setup] P1 手牌: %d 张 (均为风系), 初始阶段: %s", len(p1.Hand), game.State.Phase)

	// =============================================================
	t.Logf("\n👉 [Step 1] P1 发起第一次攻击")
	// =============================================================
	actionAtk1 := model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}

	if err := game.HandleAction(actionAtk1); err != nil {
		t.Fatalf("第一次攻击失败: %v", err)
	}

	// =============================================================
	t.Logf("\n👉 [Step 2] P2 承受伤害 (触发风怒判定)")
	// =============================================================
	actionTake := model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}

	// 此时引擎会处理伤害 -> 触发 TriggerOnPhaseEnd -> 挂起风怒中断 -> 暂停
	if err := game.HandleAction(actionTake); err != nil {
		t.Fatalf("P2 承受伤害失败: %v", err)
	}

	// 检查中断
	if game.State.PendingInterrupt != nil {
		t.Logf("⚡ [检测到中断] 类型: %s, 玩家: %s, 技能: %v",
			game.State.PendingInterrupt.Type,
			game.State.PendingInterrupt.PlayerID,
			game.State.PendingInterrupt.SkillIDs)
	}

	// =============================================================
	t.Logf("\n👉 [Step 3] P1 确认发动 [风怒追击]")
	// =============================================================
	// 模拟玩家输入 choose wind_fury
	if err := game.ConfirmResponseSkill("p1", "wind_fury"); err != nil {
		t.Fatalf("确认风怒失败: %v", err)
	}

	// 验证 PendingActions
	if len(p1.TurnState.PendingActions) > 0 {
		token := p1.TurnState.PendingActions[0]
		t.Logf("✅ [状态检查] P1 获得额外行动 Token: 类型=%s, 限制=%v", token.MustType, token.MustElement)
	} else {
		t.Error("❌ [状态检查] P1 未获得额外行动 Token")
	}

	// =============================================================
	t.Logf("\n👉 [Step 4] 驱动引擎 (结算 TurnEnd -> 进入 ExtraAction)")
	// =============================================================
	// HandleAction 结束后引擎处于暂停状态，测试代码手动推一把
	game.Drive()

	t.Logf("🔄 [状态流转] 当前阶段变为: %s", game.State.Phase)
	t.Logf("🔄 [状态流转] P1 当前限制行动: %s", p1.TurnState.CurrentExtraAction)

	// =============================================================
	t.Logf("\n👉 [Step 5] P1 执行额外攻击")
	// =============================================================
	// 此时手牌 Index 变为 0
	actionAtk2 := model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}

	if err := game.HandleAction(actionAtk2); err != nil {
		t.Fatalf("额外攻击失败: %v", err)
	}

	t.Logf("\n========== ✅ 测试成功：流程完整执行 ==========\n")
}

// TestSealer_FiveSeals 测试封印师的五系封印逻辑
func TestSealer_FiveSeals(t *testing.T) {
	// 定义测试用例结构
	type testCase struct {
		name           string           // 测试用例名称
		sealSkillID    string           // 封印师发动的技能ID
		sealEffectType model.EffectType // 期望场上生成的Effect类型
		triggerCard    model.Card       // 目标玩家使用的卡牌
		shouldTrigger  bool             // 是否应该触发封印
	}

	// 构造五种属性的测试数据
	testCases := []testCase{
		{
			name:           "水之封印-触发",
			sealSkillID:    "water_seal",
			sealEffectType: model.EffectSealWater,
			triggerCard:    model.Card{ID: "c_water", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
			shouldTrigger:  true,
		},
		{
			name:           "火之封印-触发",
			sealSkillID:    "fire_seal",
			sealEffectType: model.EffectSealFire,
			triggerCard:    model.Card{ID: "c_fire", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
			shouldTrigger:  true,
		},
		{
			name:           "地之封印-触发",
			sealSkillID:    "earth_seal",
			sealEffectType: model.EffectSealEarth,
			triggerCard:    model.Card{ID: "c_earth", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
			shouldTrigger:  true,
		},
		{
			name:           "风之封印-触发",
			sealSkillID:    "wind_seal",
			sealEffectType: model.EffectSealWind,
			triggerCard:    model.Card{ID: "c_wind", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
			shouldTrigger:  true,
		},
		{
			name:           "雷之封印-触发",
			sealSkillID:    "thunder_seal",
			sealEffectType: model.EffectSealThunder,
			triggerCard:    model.Card{ID: "c_thunder", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
			shouldTrigger:  true,
		},
		{
			name:           "水之封印-不触发(属性不匹配)",
			sealSkillID:    "water_seal",
			sealEffectType: model.EffectSealWater,
			triggerCard:    model.Card{ID: "c_fire_mismatch", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2}, // 用火系牌
			shouldTrigger:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. 初始化引擎
			observer := testutils.NewTestObserver(t) // 使用 testutils.TestObserver
			game := engine.NewGameEngine(observer)

			// 添加玩家：P1 封印师(红)，P2 狂战士(蓝，作为受害者)
			game.AddPlayer("p1", "Sealer", "sealer", model.RedCamp)
			game.AddPlayer("p2", "Victim", "berserker", model.BlueCamp)

			// 强制设置初始状态：P1 回合，行动阶段
			game.State.CurrentTurn = 0
			game.State.Deck = rules.InitDeck() // 【修复】初始化牌库，否则摸牌无效
			p1 := game.State.Players["p1"]
			p2 := game.State.Players["p2"]
			p1.IsActive = true
			p1.TurnState = model.NewPlayerTurnState()
			game.State.Phase = model.PhaseActionSelection

			// 给 P1 发一张对应封印需要的弃牌 (根据配置，部分封印需要弃牌)
			// 为了简化，我们假设技能消耗已满足，或者给 P1 塞满各种牌
			// 实际上你的配置里：水/火/地/风/雷 之封印 都需要弃 1 张对应属性的牌
			// 我们给 P1 发一张万能牌或者对应牌
			var skillTitle string
			var skillElement model.Element

			switch tc.sealSkillID {
			case "water_seal":
				skillTitle = "水之封印"
				skillElement = model.ElementWater
			case "fire_seal":
				skillTitle = "火之封印"
				skillElement = model.ElementFire
			case "earth_seal":
				skillTitle = "地之封印"
				skillElement = model.ElementEarth
			case "wind_seal":
				skillTitle = "风之封印"
				skillElement = model.ElementWind
			case "thunder_seal":
				skillTitle = "雷之封印"
				skillElement = model.ElementThunder
			}

			// 2. 构造独有牌
			discardCard := model.Card{
				ID:      "mock_exclusive_" + tc.sealSkillID,
				Name:    skillTitle + "·独有牌", // 名字仅供日志显示
				Type:    model.CardTypeMagic,
				Element: skillElement, // 元素必须匹配

				// 【核心修正点】
				// 引擎校验的是: card.MatchExclusive(player.Character.Name, skillDef.Title)
				// 所以这里必须填中文名 "封印师" 和 "水之封印"
				ExclusiveChar1:  "封印师",      // 匹配 player.Character.Name
				ExclusiveSkill1: skillTitle, // 匹配 skillDef.Title (例如 "水之封印")
			}
			// 注意：如果是不触发的case，弃牌属性也要跟技能匹配
			if !tc.shouldTrigger {
				// 对于不触发的case，比如水封印，需要弃水牌，但受害者用火牌
				discardCard.Element = model.ElementWater // 强行修正消耗牌属性，确保技能能发出来
			}

			// 针对 specific element 修正 discard card
			switch tc.sealSkillID {
			case "water_seal":
				discardCard.Element = model.ElementWater
			case "fire_seal":
				discardCard.Element = model.ElementFire
			case "earth_seal":
				discardCard.Element = model.ElementEarth
			case "wind_seal":
				discardCard.Element = model.ElementWind
			case "thunder_seal":
				discardCard.Element = model.ElementThunder
			}

			p1.Hand = []model.Card{discardCard}

			t.Logf("👉 [Step 1] P1 发动封印技能: %s -> 目标 P2", tc.sealSkillID)

			// 模拟 P1 发动技能
			// 参数: 目标ID="p2", 弃牌索引="0"
			actionSkill := model.PlayerAction{
				PlayerID:   "p1",
				Type:       model.CmdSkill,
				SkillID:    tc.sealSkillID,
				TargetIDs:  []string{"p2"},
				Selections: []int{0}, // 弃掉第1张牌
			}

			if err := game.HandleAction(actionSkill); err != nil {
				t.Fatalf("发动封印失败: %v", err)
			}

			// ========================= 【新增修复代码】 开始 =========================
			// 封印技能属于“法术行动”，结束后会触发封印师的【法术激荡】响应。
			// 此时引擎会挂起，等待 P1 选择。我们需要模拟 P1 选择“跳过”以继续流程。
			if game.State.PendingInterrupt != nil {
				t.Logf("👀 检测到中断 (法术激荡), 准备发送 CmdSelect 跳过...")

				// 1. 获取当前待选技能列表
				pendingSkills := game.State.PendingInterrupt.SkillIDs

				// 2. 计算“跳过”的索引
				// 在 handleInterruptAction 的逻辑中：
				// 如果 idx == len(SkillIDs)，则视为跳过
				skipIndex := len(pendingSkills)

				// 3. 构造真实的“选择”指令
				// 模拟客户端用户选择了最后一个选项（即跳过）
				actionSkip := model.PlayerAction{
					PlayerID:   "p1",
					Type:       model.CmdSelect,  // 使用 Select 指令
					Selections: []int{skipIndex}, // 传入跳过对应的索引
				}

				// 4. 发送指令 (走完整的 HandleAction 路由)
				if err := game.HandleAction(actionSkip); err != nil {
					t.Fatalf("跳过响应失败: %v", err)
				}

				// 5. 驱动后续流程
				game.Drive()
			}

			// 断言：P2 场上应该有了对应的封印效果
			hasSeal := false
			for _, fc := range p2.Field {
				if fc.Mode == model.FieldEffect && fc.Effect == tc.sealEffectType {
					hasSeal = true
					break
				}
			}
			if !hasSeal {
				t.Fatalf("预期 P2 场上存在 %s，但未找到", tc.sealEffectType)
			}
			t.Logf("✅ P2 成功被挂上封印: %s", tc.sealEffectType)

			// 切换回合到 P2
			game.NextTurn()
			// 注意：P1发动技能后可能进入 ExtraAction，这里我们简化流程，假设P1结束回合
			// 实际上 NextTurn 会把 Active 设给 P2，Phase 设为 BuffResolve -> Startup -> ActionSelection
			// 我们需要确保 P2 进入 ActionSelection
			game.State.CurrentTurn = 1 // P2
			p1.IsActive = false
			p2.IsActive = true
			p2.TurnState = model.NewPlayerTurnState()
			game.State.Phase = model.PhaseActionSelection

			// 给 P2 发触发牌
			p2.Hand = []model.Card{tc.triggerCard}
			initialHandSize := len(p2.Hand)

			t.Logf("👉 [Step 2] P2 使用卡牌: %s (%s)", tc.triggerCard.Name, tc.triggerCard.Element)

			// P2 发起攻击 (这将触发 TriggerOnCardUsed)
			actionAtk := model.PlayerAction{
				PlayerID:  "p2",
				Type:      model.CmdAttack,
				TargetID:  "p1",
				CardIndex: 0,
			}

			// 记录 P2 当前手牌数 (为了验证是否摸了3张牌)
			// 注意：打出牌后手牌-1，如果封印触发造成3点伤害，会摸3张牌
			// 预期手牌变化：
			// 触发：Initial(1) - 1(打出) + 3(伤害摸牌) = 3
			// 不触发：Initial(1) - 1(打出) = 0

			// 执行攻击
			// 注意：HandleAction 内部会触发 TriggerOnCardUsed -> 检测封印 -> 造成伤害
			if err := game.HandleAction(actionAtk); err != nil {
				t.Fatalf("P2 攻击失败: %v", err)
			}

			// ========================= 【新增修复代码】 开始 =========================
			// 此时 P1 需要响应战斗（承担伤害），否则流程会卡住
			if game.State.PendingInterrupt == nil && game.State.Phase == model.PhaseCombatInteraction {
				t.Logf("👀 检测到战斗交互中断，P1 承担伤害...")
				actionTakeHit := model.PlayerAction{
					PlayerID:  "p1",
					Type:      model.CmdRespond,
					ExtraArgs: []string{"take"},
				}
				if err := game.HandleAction(actionTakeHit); err != nil {
					t.Fatalf("P1 承担伤害失败: %v", err)
				}
			}
			// ========================= 【新增修复代码】 结束 =========================

			// 验证结果
			if tc.shouldTrigger {
				// 1. 验证伤害摸牌
				// P2 应该受到 3 点伤害，即摸 3 张牌
				// 此时 P2 应该处于 CombatInteraction 阶段，但伤害已经结算完了
				// 检查手牌数量
				expectedHand := initialHandSize - 1 + 3
				if len(p2.Hand) != expectedHand {
					t.Errorf("封印触发后手牌数错误：预期 %d，实际 %d (原手牌 %d)", expectedHand, len(p2.Hand), initialHandSize)
				} else {
					t.Logf("✅ 伤害结算正确：P2 摸了 3 张牌")
				}

				// 2. 验证封印移除
				hasSealAfter := false
				for _, fc := range p2.Field {
					if fc.Mode == model.FieldEffect && fc.Effect == tc.sealEffectType {
						hasSealAfter = true
						break
					}
				}
				if hasSealAfter {
					t.Errorf("封印触发后应该被移除，但 P2 场上仍有 %s", tc.sealEffectType)
				} else {
					t.Logf("✅ 封印已正确移除")
				}

			} else {
				// 不应该触发
				// 1. 验证没有额外摸牌
				expectedHand := initialHandSize - 1
				if len(p2.Hand) != expectedHand {
					t.Errorf("封印不应触发，但手牌数异常：预期 %d，实际 %d", expectedHand, len(p2.Hand))
				}

				// 2. 验证封印保留
				hasSealAfter := false
				for _, fc := range p2.Field {
					if fc.Mode == model.FieldEffect && fc.Effect == tc.sealEffectType {
						hasSealAfter = true
						break
					}
				}
				if !hasSealAfter {
					t.Errorf("封印不应触发，但 P2 场上的 %s 消失了", tc.sealEffectType)
				} else {
					t.Logf("✅ 属性不匹配，封印未触发且保留在场上")
				}
			}
		})
	}
}

// TestAngel_AngelBlessing 测试天使祝福技能：弃1张水系牌，指定1名玩家选2张牌给你
func TestAngel_AngelBlessing(t *testing.T) {
	observer := testutils.NewTestObserver(t)
	game := engine.NewGameEngine(observer)

	t.Logf("\n========== 🚀 测试开始：天使祝福 ==========")

	// 添加玩家：P1 天使(红)，P2 狂战士(蓝，作为给牌目标)
	game.AddPlayer("p1", "Angel", "angel", model.RedCamp)
	game.AddPlayer("p2", "Victim", "berserker", model.BlueCamp)

		// 设置 P1 回合，行动阶段
		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck() // 【修复】初始化牌库
		game.State.HasPerformedStartup = true
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.Phase = model.PhaseActionSelection

	// P1 手牌：1 张水系牌用于发动技能
	waterCard := model.Card{
		ID:      "c_water",
		Name:    "水涟斩",
		Type:    model.CardTypeMagic,
		Element: model.ElementWater,
		Damage:  0,
	}
	p1.Hand = []model.Card{waterCard}

	// P2 手牌：3 张牌，需要选 2 张交给 P1
	p2.Hand = []model.Card{
		{ID: "c1", Name: "牌1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "c2", Name: "牌2", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
		{ID: "c3", Name: "牌3", Type: model.CardTypeMagic, Element: model.ElementWind, Damage: 0},
	}

	t.Logf("✅ [Setup] P1 手牌: 1 张水系, P2 手牌: 3 张")

	// Step 1: P1 发动天使祝福，目标 P2，弃第 1 张牌 (index 0)
	t.Logf("\n👉 [Step 1] P1 发动天使祝福 -> 目标 P2, 弃牌 index 0")
	actionSkill := model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "angel_blessing",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	}

	if err := game.HandleAction(actionSkill); err != nil {
		t.Fatalf("发动天使祝福失败: %v", err)
	}

	// 处理可能的中断（如法术激荡）
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type != model.InterruptGiveCards {
		t.Logf("👀 检测到中断 %s, 跳过...", game.State.PendingInterrupt.Type)
		if game.State.PendingInterrupt.Type == model.InterruptResponseSkill {
			game.SkipResponse()
		}
		game.Drive()
	}

	// 断言：应产生 InterruptGiveCards，P2 需选 2 张牌
	if game.State.PendingInterrupt == nil {
		t.Fatalf("预期产生给牌中断，但 PendingInterrupt 为空")
	}
	if game.State.PendingInterrupt.Type != model.InterruptGiveCards {
		t.Fatalf("预期中断类型为 GiveCards，实际为 %s", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("预期给牌者为 p2，实际为 %s", game.State.PendingInterrupt.PlayerID)
	}

	data, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	giveCount, _ := data["give_count"].(int)
	if giveCount != 2 {
		t.Fatalf("预期需给 2 张牌，实际为 %d", giveCount)
	}
	t.Logf("✅ P2 需选择 2 张牌交给 P1")

	// Step 2: P2 选择第 1、2 张牌交给 P1 (indices 0, 1)
	t.Logf("\n👉 [Step 2] P2 选择牌 0 和 1 交给 P1")
	actionGive := model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	}

	if err := game.HandleAction(actionGive); err != nil {
		t.Fatalf("P2 给牌失败: %v", err)
	}

	// 验证：P1 收到 2 张牌，P2 剩 1 张
	if len(p1.Hand) != 2 {
		t.Errorf("P1 手牌数错误：预期 2 张 (收到2张)，实际 %d", len(p1.Hand))
	}
	if len(p2.Hand) != 1 {
		t.Errorf("P2 手牌数错误：预期 1 张 (原有3-给出2)，实际 %d", len(p2.Hand))
	}
	t.Logf("✅ P1 手牌: %d 张, P2 手牌: %d 张", len(p1.Hand), len(p2.Hand))

	t.Logf("\n========== ✅ 测试成功：天使祝福流程完整执行 ==========\n")
}
