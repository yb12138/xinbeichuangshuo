package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func mustHandleAction(t *testing.T, game *GameEngine, act model.PlayerAction) {
	t.Helper()
	if err := game.HandleAction(act); err != nil {
		t.Fatalf("handle action failed (%+v): %v", act, err)
	}
}

func requireResponseSkillPrompt(t *testing.T, game *GameEngine, playerID string) {
	t.Helper()
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending interrupt, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response-skill interrupt, got %s", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected interrupt player %s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
}

func requireChoicePrompt(t *testing.T, game *GameEngine, playerID, choiceType string) {
	t.Helper()
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending interrupt, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected choice interrupt, got %s", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected interrupt player %s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
	ctxData, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("choice context type mismatch")
	}
	got, _ := ctxData["choice_type"].(string)
	if got != choiceType {
		t.Fatalf("expected choice_type=%s, got %s", choiceType, got)
	}
}

func chooseResponseSkillByID(t *testing.T, game *GameEngine, playerID, skillID string) {
	t.Helper()
	requireResponseSkillPrompt(t, game, playerID)
	idx := -1
	for i, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == skillID {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("skill %s not found in pending skills: %+v", skillID, game.State.PendingInterrupt.SkillIDs)
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   playerID,
		Type:       model.CmdSelect,
		Selections: []int{idx},
	})
}

func firstAttackCardIndex(p *model.Player) int {
	for i, c := range p.Hand {
		if c.Type == model.CardTypeAttack {
			return i
		}
	}
	return -1
}

// 回归测试：女武神连招应可完整执行，不在英灵召唤结算后提前断回合
func TestValkyrie_ComboChain_FullFlow(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Valkyrie", "valkyrie", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection

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
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "valkyrie_order_seal",
	})
	requireResponseSkillPrompt(t, game, "p1")
	chooseResponseSkillByID(t, game, "p1", "valkyrie_divine_pursuit")

	// 2) 神圣追击后应进入额外攻击行动
	game.Drive()
	if game.State.Phase != model.PhaseActionSelection {
		t.Fatalf("expected phase ActionSelection after divine pursuit, got %s", game.State.Phase)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected current extra action=Attack, got %s", p1.TurnState.CurrentExtraAction)
	}

	// 3) 攻击命中后应询问英灵召唤
	attackIdx := firstAttackCardIndex(p1)
	if attackIdx < 0 {
		t.Fatalf("no attack card found for first attack")
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: attackIdx,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	requireResponseSkillPrompt(t, game, "p1")
	chooseResponseSkillByID(t, game, "p1", "valkyrie_heroic_summon")

	// 4) 英灵召唤额外流程：确认弃法术 -> 选法术 -> 选治疗目标
	requireChoicePrompt(t, game, "p1", "valkyrie_heroic_extra_confirm")
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}}) // 是
	requireChoicePrompt(t, game, "p1", "valkyrie_heroic_discard_card")
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	requireChoicePrompt(t, game, "p1", "valkyrie_heroic_heal_target")
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}}) // 治疗自己

	// 5) 攻击行动结束后应再次询问神圣追击（关键回归点）
	requireResponseSkillPrompt(t, game, "p1")
	chooseResponseSkillByID(t, game, "p1", "valkyrie_divine_pursuit")

	// 6) 再次进入额外攻击，并在主动攻击开始时脱离英灵形态（和平行者）
	game.Drive()
	if game.State.Phase != model.PhaseActionSelection {
		t.Fatalf("expected phase ActionSelection before second attack, got %s", game.State.Phase)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected current extra action=Attack before second attack, got %s", p1.TurnState.CurrentExtraAction)
	}
	if p1.Tokens["valkyrie_spirit"] != 1 {
		t.Fatalf("expected valkyrie spirit=1 before second attack, got %d", p1.Tokens["valkyrie_spirit"])
	}

	secondAttackIdx := firstAttackCardIndex(p1)
	if secondAttackIdx < 0 {
		t.Fatalf("no attack card found for second attack")
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: secondAttackIdx,
	})
	if p1.Tokens["valkyrie_spirit"] != 0 {
		t.Fatalf("expected valkyrie spirit to be removed on active attack start, got %d", p1.Tokens["valkyrie_spirit"])
	}
}

// 回归测试：英灵召唤在响应阶段取消后，不应在同一次命中结算里重复弹出
func TestValkyrie_HeroicSummon_CancelDoesNotRepromptSameHit(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Valkyrie", "valkyrie", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection

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
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	requireResponseSkillPrompt(t, game, "p1")
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "valkyrie_heroic_summon" {
		t.Fatalf("expected only valkyrie_heroic_summon prompt, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	// 取消响应：不发动英灵召唤
	mustHandleAction(t, game, model.PlayerAction{
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
