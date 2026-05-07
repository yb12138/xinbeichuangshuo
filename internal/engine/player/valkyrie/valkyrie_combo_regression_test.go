package valkyrie_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归测试：女武神连招应可完整执行，不在英灵召唤结算后提前断回合
func TestValkyrie_ComboChain_FullFlow(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Valkyrie", "valkyrie", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0
	p1.Heal = 0
	p1.Crystal = 0
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
		{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	}

	// 1) 发动秩序之印 -> 应在法术行动结束后询问神圣追击
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "valkyrie_order_seal",
	})
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	testutils.ChooseResponseSkillByID(t, game, "p1", "valkyrie_divine_pursuit")

	// 2) 神圣追击后应进入额外攻击行动
	game.Drive()
	if !game.IsActionSelectionWindow() {
		t.Fatalf("expected action selection window after divine pursuit, got %s", game.RuntimeStateLabel())
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected current extra action=Attack, got %s", p1.TurnState.CurrentExtraAction)
	}

	// 3) 攻击命中后应询问英灵召唤
	attackIdx := testutils.FirstAttackCardIndex(p1)
	if attackIdx < 0 {
		t.Fatalf("no attack card found for first attack")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: attackIdx,
	})
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	testutils.ChooseResponseSkillByID(t, game, "p1", "valkyrie_heroic_summon")

	// 4) 英灵召唤额外流程：直接选择法术牌（可取消）-> 当前战斗目标自动+1治疗
	testutils.RequireChoicePrompt(t, game, "p1", "valkyrie_heroic_discard_card")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	if p2.Heal != 1 {
		t.Fatalf("expected heroic summon extra heal to apply to current combat target, got %d", p2.Heal)
	}

	// 5) 当前文档口径下，额外治疗只能给当前战斗目标，因此不会再给自己补治疗触发第二次神圣追击。
	if !player.HasForm(p1, model.FormValkyrieHeroic) {
		t.Fatalf("expected enter heroic form after heroic summon on self turn, got form=%q", p1.Form)
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptResponseSkill {
		for _, sid := range game.State.PendingInterrupt.SkillIDs {
			if sid == "valkyrie_divine_pursuit" {
				t.Fatalf("divine pursuit should not reprompt when heroic summon healed the combat target instead of self")
			}
		}
	}
}

// 回归测试：英灵召唤在响应阶段取消后，不应在同一次命中结算里重复弹出
func TestValkyrie_HeroicSummon_CancelDoesNotRepromptSameHit(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Valkyrie", "valkyrie", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 0
	p2.Heal = 0
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	// 攻击命中后进入英灵召唤响应询问
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "valkyrie_heroic_summon" {
		t.Fatalf("expected only valkyrie_heroic_summon prompt, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	// 取消响应：不发动英灵召唤
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	})

	// 取消后不应再次弹出同一次命中的英灵召唤响应
	if intr := game.State.PendingInterrupt; intr != nil && intr.Type == model.InterruptResponseSkill {
		for _, sid := range intr.SkillIDs {
			if sid == "valkyrie_heroic_summon" {
				t.Fatalf("heroic summon reprompted after cancel on same hit")
			}
		}
	}

	// 取消不应消耗蓝水晶，且当前延迟伤害应已继续结算完成
	if p1.Crystal != 1 {
		t.Fatalf("expected crystal remain 1 after cancel, got %d", p1.Crystal)
	}
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("expected pending damage queue drained, got %d", len(game.State.PendingDamageQueue))
	}
}
