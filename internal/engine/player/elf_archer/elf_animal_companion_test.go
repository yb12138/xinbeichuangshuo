package elf_archer_test

import (
	"fmt"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func TestElfAnimalCompanion_DrawOneDiscardOne(t *testing.T) {
	game := engine.NewGameEngine(&testutils.CaptureObserver{})
	if err := game.AddPlayer("p1", "Elf", "elf_archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	// 精灵射手有1张攻击牌和1张其他牌（不要有法术牌，否则元素射击会弹出）
	p1.Hand = []model.Card{
		{ID: "atk-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "existing-card", Name: "其他牌", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "def-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	// 发起攻击
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})

	// Drive 进入战斗响应阶段
	game.Drive()
	fmt.Printf("After attack Drive: TurnStage=%s, PendingInterrupt=%+v, CombatStage=%s\n", game.State.TurnStage, game.State.PendingInterrupt, game.State.CombatStage)

	// 目标承受伤害（跳过防御/应战）
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	// Drive 处理伤害结算和后续
	game.Drive()
	fmt.Printf("After take Drive: TurnStage=%s, PendingInterrupt=%+v\n", game.State.TurnStage, game.State.PendingInterrupt)

	// 伤害结算后应出现动物伙伴响应技能中断
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected response skill prompt after damage resolution, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected ResponseSkill interrupt, got %+v", game.State.PendingInterrupt)
	}

	testutils.RequireResponseSkillPrompt(t, game, "p1")
	found := false
	for _, skillID := range game.State.PendingInterrupt.SkillIDs {
		if skillID == "elf_animal_companion" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected elf_animal_companion in response skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	// 选择发动动物伙伴
	testutils.ChooseResponseSkillByID(t, game, "p1", "elf_animal_companion")

	// Drive 推进，弹出响应技能中断，激活弃牌中断
	game.Drive()
	fmt.Printf("After animal companion Drive: TurnStage=%s, PendingInterrupt=%+v, Hand=%d\n", game.State.TurnStage, game.State.PendingInterrupt, len(p1.Hand))

	// 动物伙伴执行：摸1牌后应出现弃牌选择
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected discard prompt after animal companion draw, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected choice interrupt for discard, got %+v", game.State.PendingInterrupt)
	}
	ctxData, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("expected context map, got %+v", game.State.PendingInterrupt.Context)
	}
	if got, _ := ctxData["choice_type"].(string); got != "system_discard_cards" {
		t.Fatalf("expected system_discard_cards choice_type, got %q", got)
	}

	// 验证摸牌：手牌应该有2张（原有1张 + 摸1张，攻击牌已打出）
	if len(p1.Hand) != 2 {
		t.Fatalf("expected 2 cards in hand after draw (original 1 + draw 1), got %d", len(p1.Hand))
	}

	// 选择弃牌（弃掉摸到的牌或原有的牌）
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	fmt.Printf("After discard: Hand=%d\n", len(p1.Hand))

	// 弃牌完成后，手牌应该有1张
	if len(p1.Hand) != 1 {
		t.Fatalf("expected 1 card in hand after discard, got %d cards: %+v", len(p1.Hand), p1.Hand)
	}

	// 验证游戏继续，进入正常流程
	game.Drive()
	fmt.Printf("Final Drive: TurnStage=%s\n", game.State.TurnStage)
	if game.State.TurnStage == "" {
		t.Fatalf("expected turn stage to be set after discard, got empty")
	}
}