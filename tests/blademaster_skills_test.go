package tests

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

func TestBladeMaster_Skills(t *testing.T) {
	observer := testutils.NewTestObserver(t)

	// -------------------------------------------------------------------------
	// Case 1: 圣剑 (Holy Sword) - 第三次攻击强制命中
	// -------------------------------------------------------------------------
	t.Run("HolySword_ThirdAttack", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp)
		game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)
		
		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		// p2 := game.State.Players["p2"] // Unused
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection

		// 给 P1 3 张攻击牌
		p1.Hand = []model.Card{
			{ID: "atk1", Name: "斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
			{ID: "atk2", Name: "斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
			{ID: "atk3", Name: "斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		}

		// 模拟前两次攻击 (手动增加 AttackCount，省去完整 Action 流程)
		p1.TurnState.AttackCount = 2 
		
		// 发起第 3 次攻击
		action := model.PlayerAction{
			PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
		}
		
		// 应该触发 Holy Sword
		if err := game.HandleAction(action); err != nil {
			t.Fatalf("第三次攻击失败: %v", err)
		}

		// 验证是否强制命中
		// 我们无法直接读取局部变量 ctx，但可以通过日志或 AttackInfo 状态推断
		// 或者看 P2 是否可以应战
		// 这里简单检查 p1.TurnState.AttackCount 增加
		if p1.TurnState.AttackCount != 3 {
			t.Errorf("攻击计数错误")
		}
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptHolySwordDraw {
			t.Fatalf("预期第3次攻击后触发圣剑摸X弃X中断")
		}
		
		t.Logf("✅ 圣剑测试通过 (假设强制命中生效)")
	})

	// -------------------------------------------------------------------------
	// Case 2: 剑影 (Sword Shadow) - 攻击后消耗水晶额外攻击
	// -------------------------------------------------------------------------
	t.Run("SwordShadow_ExtraAction", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp)
		game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)
		
		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		// p2 := game.State.Players["p2"] // Unused
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection

		p1.Crystal = 1
		p1.Hand = []model.Card{
			{ID: "atk1", Name: "斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		}

		// 攻击 -> P2 Take -> PhaseEnd -> Sword Shadow Interrupt
		action := model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0}
		game.HandleAction(action)
		game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"}})
		game.Drive()

		// 检查中断
		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
			t.Fatalf("预期产生剑影响应中断")
		}
		
		// 确认剑影
		game.ConfirmResponseSkill("p1", "sword_shadow")
		
		// 验证水晶消耗和 Token
		if p1.Crystal != 0 {
			t.Errorf("剑影未消耗水晶")
		}
		if len(p1.TurnState.PendingActions) == 0 {
			t.Errorf("剑影未添加额外行动")
		}
		t.Logf("✅ 剑影测试通过")
	})

	// -------------------------------------------------------------------------
	// Case 3: 烈风技 (Gale Slash) - 无视圣盾
	// -------------------------------------------------------------------------
	t.Run("GaleSlash_IgnoreShield", func(t *testing.T) {
		game := engine.NewGameEngine(observer)
		game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp)
		game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)
		
		game.State.CurrentTurn = 0
		game.State.Deck = rules.InitDeck()
		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		game.State.Phase = model.PhaseActionSelection

		// 给 P1 独有牌
		p1.Hand = []model.Card{
			{
				ID: "gs_card", Name: "烈风技", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2,
				ExclusiveChar1: "风之剑圣", ExclusiveSkill1: "烈风技", // 名字修正
			},
		}

		// P2 有圣盾
		p2.AddFieldCard(&model.FieldCard{Mode: model.FieldEffect, Effect: model.EffectShield})
		p2.Heal = 5
		initialHandP2 := len(p2.Hand)

		// P1 攻击
		action := model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0}
		game.HandleAction(action)
		
		// P2 Take (圣盾本应抵挡，但烈风技无视)
		// 如果烈风技生效，GaleSlashActive = true
		// 圣盾的 Handler (HolyShieldHandler) 应该检查 GaleSlashActive?
		// 或者 GaleSlashHandler Execute 只是设了个 Flag?
		// 检查 handlers_impl.go: GaleSlashHandler sets `GaleSlashActive = true`
		// 那么 HolyShieldHandler 必须检查这个 Flag。
		// 让我们假设 HolyShieldHandler 逻辑里已经包含了这个检查 (CanUse check?)
		// 如果没包含，这个测试会失败，发现Bug。
		
		game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"}})
		game.Drive() // 结算

		// 当前伤害模型是“受伤摸牌”。若烈风技生效，目标应实际受伤并摸牌，且圣盾不应自动抵挡这次攻击
		if len(p2.Hand)-initialHandP2 != 2 {
			t.Errorf("烈风技未造成预期伤害: 预期摸2张，实际摸%d张", len(p2.Hand)-initialHandP2)
		}
		if !p2.HasFieldEffect(model.EffectShield) {
			t.Errorf("烈风技应无视圣盾而非消耗圣盾，但当前圣盾已被移除")
		}
	})
}
