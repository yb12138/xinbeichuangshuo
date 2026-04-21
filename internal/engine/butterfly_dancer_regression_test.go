package engine

import (
	"testing"

	engineplayer "starcup-engine/internal/engine/player"
	butterflydancer "starcup-engine/internal/engine/player/butterfly_dancer"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func butterflyTestCard(id string, typ model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        "测试牌",
		Type:        typ,
		Element:     ele,
		Faction:     "咏",
		Damage:      1,
		Description: "test",
	}
}

func TestButterflyLifeFire_MaxHandFloor(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]

	p1.Tokens["bt_pupa"] = 0
	if got := game.GetMaxHand(p1); got != 6 {
		t.Fatalf("expected max hand 6 at pupa=0, got %d", got)
	}
	p1.Tokens["bt_pupa"] = 2
	if got := game.GetMaxHand(p1); got != 4 {
		t.Fatalf("expected max hand 4 at pupa=2, got %d", got)
	}
	p1.Tokens["bt_pupa"] = 20
	if got := game.GetMaxHand(p1); got != 3 {
		t.Fatalf("expected max hand floor 3, got %d", got)
	}
}

func TestButterflyDance_DrawAndGainCocoon(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		butterflyTestCard("h1", model.CardTypeAttack, model.ElementFire),
	}
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bt_dance",
	})
	requireChoicePrompt(t, game, "p1", "bt_dance_mode")

	// 选择“摸1张牌”。
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	if got := butterflydancer.CocoonCount(p1); got != 1 {
		t.Fatalf("expected 1 cocoon after dance, got %d", got)
	}
	if len(p1.Hand) != 2 {
		t.Fatalf("expected hand size 2 after draw mode, got %d", len(p1.Hand))
	}
}

func TestButterflyChrysalis_RunsOverflowDiscardWhenPupaLowersHandLimit(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		butterflyTestCard("h1", model.CardTypeAttack, model.ElementFire),
		butterflyTestCard("h2", model.CardTypeAttack, model.ElementWater),
		butterflyTestCard("h3", model.CardTypeAttack, model.ElementWind),
		butterflyTestCard("h4", model.CardTypeAttack, model.ElementThunder),
		butterflyTestCard("h5", model.CardTypeMagic, model.ElementDark),
		butterflyTestCard("h6", model.CardTypeMagic, model.ElementLight),
	}
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bt_chrysalis",
	})

	if got := p1.Tokens["bt_pupa"]; got != 1 {
		t.Fatalf("expected pupa +1 after chrysalis, got %d", got)
	}
	if got := game.GetMaxHand(p1); got != 5 {
		t.Fatalf("expected max hand 5 after pupa +1, got %d", got)
	}
	if game.State.PendingInterrupt == nil || !isDiscardSelectionInterrupt(game.State.PendingInterrupt) || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected overflow discard interrupt after chrysalis, got %+v", game.State.PendingInterrupt)
	}
	data, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if dc, _ := data["discard_count"].(int); dc != 1 {
		t.Fatalf("expected discard_count=1 after hand limit shrink, got %v", data["discard_count"])
	}
}

func TestButterflyReverse_RequiresTwoDiscardCardsByConfig(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Hand = []model.Card{
		butterflyTestCard("h1", model.CardTypeAttack, model.ElementFire),
	}
	game.State.TurnStage = model.TurnStageActionExecution

	sd := findCharacterSkill(p1.Character, "bt_reverse_butterfly")
	if sd == nil {
		t.Fatal("expected bt_reverse_butterfly skill definition")
	}
	if !game.canSatisfyActionSkillDiscardRequirement(p1, *sd) {
		// 满足这里说明当前实现会在可用技能层面阻止手牌不足时发动。
	} else {
		t.Fatalf("expected reverse butterfly discard requirement to fail with only 1 hand card")
	}

	err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bt_reverse_butterfly",
	})
	if err == nil {
		t.Fatal("expected using reverse butterfly with only 1 hand card to fail")
	}
}

func TestButterflyReverse_UsesUnifiedDiscardCostBeforeBranchChoice(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Hand = []model.Card{
		butterflyTestCard("h1", model.CardTypeAttack, model.ElementFire),
		butterflyTestCard("h2", model.CardTypeMagic, model.ElementWater),
	}
	game.State.TurnStage = model.TurnStageActionExecution

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bt_reverse_butterfly",
	})

	if game.State.PendingInterrupt == nil || !isDiscardSelectionInterrupt(game.State.PendingInterrupt) || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected reverse butterfly to first enter discard interrupt, got %+v", game.State.PendingInterrupt)
	}
	data, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if skillID, _ := data["skill_id"].(string); skillID != "bt_reverse_butterfly" {
		t.Fatalf("expected discard interrupt skill_id=bt_reverse_butterfly, got %v", data["skill_id"])
	}
	if discardCount, _ := data["discard_count"].(int); discardCount != 2 {
		t.Fatalf("expected discard_count=2, got %v", data["discard_count"])
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	})
	requireChoicePrompt(t, game, "p1", "bt_reverse_mode")
	if got := len(p1.Hand); got != 0 {
		t.Fatalf("expected 2 cards consumed before branch choice, got hand=%d", got)
	}
}

func TestButterflyPilgrimage_ResistOneDamage(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	butterflydancer.AddCocoonCards(p1, []model.Card{
		butterflyTestCard("c1", model.CardTypeAttack, model.ElementWater),
	})

	game.AddPendingDamage(model.PendingDamage{
		SourceID:   p2.ID,
		TargetID:   p1.ID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.ReturnTurnStage = model.TurnStageExtraAction

	game.Drive()
	requireChoicePrompt(t, game, "p1", "bt_pilgrimage_pick")

	// 选择移除第一个茧（选项0为不发动，因此这里选1）。
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})
	if got := butterflydancer.CocoonCount(p1); got != 0 {
		t.Fatalf("expected cocoon consumed by pilgrimage, got %d", got)
	}

	// 继续驱动直到伤害队列清空。
	for i := 0; i < 8 && len(game.State.PendingDamageQueue) > 0; i++ {
		if game.State.PendingInterrupt != nil {
			// 本用例不应再有必选中断；若出现则跳过。
			if game.State.PendingInterrupt.Type == model.InterruptChoice {
				mustHandleAction(t, game, model.PlayerAction{
					PlayerID:   game.State.PendingInterrupt.PlayerID,
					Type:       model.CmdSelect,
					Selections: []int{0},
				})
				continue
			}
		}
		game.Drive()
	}
	if got := len(game.State.PendingDamageQueue); got != 0 {
		t.Fatalf("pending damage queue not drained, len=%d", got)
	}
}

func TestButterflyMirror_ReplaceTwoDamageToTwoHits(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	game.State.Deck = rules.InitDeck()

	butterflydancer.AddCocoonCards(p1, []model.Card{
		butterflyTestCard("m1", model.CardTypeAttack, model.ElementFire),
		butterflyTestCard("m2", model.CardTypeAttack, model.ElementFire),
	})

	game.AddPendingDamage(model.PendingDamage{
		SourceID:   p2.ID,
		TargetID:   p3.ID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.ReturnTurnStage = model.TurnStageExtraAction

	game.Drive()
	requireChoicePrompt(t, game, "p1", "bt_mirror_pair")

	// 选项0为不发动，选项1为第一组同系茧。
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})
	if got := butterflydancer.CocoonCount(p1); got != 0 {
		t.Fatalf("expected 2 cocoons consumed by mirror, got remaining=%d", got)
	}

	// 推进结算，原2点伤害应被替换为两次1点。
	for i := 0; i < 20 && (len(game.State.PendingDamageQueue) > 0 || game.State.PendingInterrupt != nil); i++ {
		if game.State.PendingInterrupt != nil {
			mustHandleAction(t, game, model.PlayerAction{
				PlayerID:   game.State.PendingInterrupt.PlayerID,
				Type:       model.CmdSelect,
				Selections: []int{0},
			})
			continue
		}
		game.Drive()
	}
	if got := len(game.State.PendingDamageQueue); got != 0 {
		t.Fatalf("pending damage queue not drained, len=%d", got)
	}
	if got := len(p2.Hand); got != 2 {
		t.Fatalf("expected damage source draw 2 cards from two 1-damage hits, got hand=%d", got)
	}
	if got := len(p3.Hand); got != 0 {
		t.Fatalf("expected original target take no replacement damage, got hand=%d", got)
	}
}

func TestButterflyWither_MoraleFloorAtOne(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	engineplayer.EnsurePlayerSkillFlowState(p1)
	p1.TurnState.SkillFlowState["bt_wither_active"] = 1

	game.State.BlueMorale = 1
	if got := game.applyCampMoraleLoss(model.BlueCamp, 3); got != 0 {
		t.Fatalf("expected morale loss 0 at floor, got %d", got)
	}
	if game.State.BlueMorale != 1 {
		t.Fatalf("expected blue morale stay at 1, got %d", game.State.BlueMorale)
	}

	game.State.BlueMorale = 3
	if got := game.applyCampMoraleLoss(model.BlueCamp, 5); got != 2 {
		t.Fatalf("expected clamped morale loss 2, got %d", got)
	}
	if game.State.BlueMorale != 1 {
		t.Fatalf("expected blue morale clamped to 1, got %d", game.State.BlueMorale)
	}
}

func TestButterflyWither_CanTargetAnyCharacter(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	butterflydancer.QueueWitherFollowup(newRoleChoiceRuntime(game), p1)
	requireChoicePrompt(t, game, "p1", "bt_wither_confirm")

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	requireChoicePrompt(t, game, "p1", "bt_wither_target")
	prompt := game.GetCurrentPrompt()
	if prompt == nil {
		t.Fatalf("expected wither target prompt")
	}
	foundAlly := false
	for _, opt := range prompt.Options {
		if opt.Label == "Ally" {
			foundAlly = true
			break
		}
	}
	if !foundAlly {
		t.Fatalf("expected wither target prompt to include ally target, got %+v", prompt.Options)
	}

	// 玩家顺序为 p1, p2, p3；这里选择己方队友 p2，验证目标不限于敌人。
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})

	if got := len(game.State.PendingDamageQueue); got != 0 {
		t.Fatalf("expected wither damage chain fully resolved, got %d pending entries", got)
	}
	if got := p1.TurnState.SkillFlowState["bt_wither_active"]; got != 1 {
		t.Fatalf("expected wither floor effect active, got %d", got)
	}
}
