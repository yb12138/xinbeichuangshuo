package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func requireChoiceContext(t *testing.T, game *GameEngine, playerID, choiceType string) map[string]interface{} {
	t.Helper()
	requireChoicePrompt(t, game, playerID, choiceType)
	ctx, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("choice context type mismatch")
	}
	return ctx
}

func choiceIndexForTarget(t *testing.T, ctx map[string]interface{}, targetID string) int {
	t.Helper()
	var targetIDs []string
	if arr, ok := ctx["target_ids"].([]string); ok {
		targetIDs = append(targetIDs, arr...)
	} else if arr, ok := ctx["target_ids"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				targetIDs = append(targetIDs, s)
			}
		}
	}
	for i, id := range targetIDs {
		if id == targetID {
			return i
		}
	}
	t.Fatalf("target %s not found in choice target_ids=%v", targetID, targetIDs)
	return -1
}

func TestAdventurerStealSky_ModeAndExtraActionChoice(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	game.State.BlueGems = 1

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "adventurer_steal_sky",
	})
	requireChoicePrompt(t, game, "p1", "adventurer_steal_sky_mode")

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	requireChoicePrompt(t, game, "p1", "adventurer_steal_sky_extra_action")
	if game.State.BlueGems != 0 || game.State.RedGems != 1 {
		t.Fatalf("expected gem transfer blue->red, got blue=%d red=%d", game.State.BlueGems, game.State.RedGems)
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected steal sky choices resolved, got pending interrupt %+v", game.State.PendingInterrupt)
	}
}

func TestAdventurerExtractFullEnergy_ForceParadiseTransfer(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Crystal = 2 // 自身能量已满
	p2.Gem = 1
	game.State.RedCrystals = 2

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdExtract})
	requireChoicePrompt(t, game, "p1", "extract")

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	})
	requireResponseSkillPrompt(t, game, "p1")

	skipIdx := len(game.State.PendingInterrupt.SkillIDs)
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{skipIdx},
	}); err == nil {
		t.Fatalf("expected forced paradise transfer to reject skip")
	}

	chooseResponseSkillByID(t, game, "p1", "adventurer_paradise")
	requireChoicePrompt(t, game, "p1", "adventurer_paradise_target")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if p1.Gem+p1.Crystal != 2 {
		t.Fatalf("expected p1 energy reduced to 2 after paradise cost, got gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
	if p2.Gem+p2.Crystal != 3 {
		t.Fatalf("expected p2 receive two extracted energies, got gem=%d crystal=%d", p2.Gem, p2.Crystal)
	}
	if game.State.RedCrystals != 0 {
		t.Fatalf("expected red crystals extracted to 0, got %d", game.State.RedCrystals)
	}
}

func TestAdventurerParadise_TransferOnlyExtractedEnergy(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 2 // 已有能量，不应被整体转移
	game.State.RedGems = 1

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdExtract})
	requireChoicePrompt(t, game, "p1", "extract")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	chooseResponseSkillByID(t, game, "p1", "adventurer_paradise")
	requireChoicePrompt(t, game, "p1", "adventurer_paradise_target")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if p2.Gem != 1 || p2.Crystal != 0 {
		t.Fatalf("expected ally only receive extracted gem, got gem=%d crystal=%d", p2.Gem, p2.Crystal)
	}
	if p1.Gem != 1 || p1.Crystal != 0 {
		t.Fatalf("expected p1 keep pre-existing energy except mandatory -1, got gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
}

func TestAdventurerParadise_TargetsFilteredByExtractCapacity(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "AllyLowRoom", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AllyEnoughRoom", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Crystal = 2 // 能量已满，提炼后必须走冒险者天堂转移
	p2.Gem = 1
	p2.Crystal = 1 // 仅剩1格，不可接收2点提炼
	p3.Gem = 1
	p3.Crystal = 0 // 剩余2格，可完整接收
	game.State.RedCrystals = 2

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdExtract})
	requireChoicePrompt(t, game, "p1", "extract")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	})

	chooseResponseSkillByID(t, game, "p1", "adventurer_paradise")
	ctx := requireChoiceContext(t, game, "p1", "adventurer_paradise_target")
	var allyIDs []string
	if arr, ok := ctx["ally_ids"].([]string); ok {
		allyIDs = append(allyIDs, arr...)
	} else if arr, ok := ctx["ally_ids"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				allyIDs = append(allyIDs, s)
			}
		}
	}
	if len(allyIDs) != 1 || allyIDs[0] != "p3" {
		t.Fatalf("expected only p3 can receive extracted energy, got ally_ids=%v", allyIDs)
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if p2.Gem+p2.Crystal != 2 {
		t.Fatalf("expected p2 unchanged (cannot receive), got gem=%d crystal=%d", p2.Gem, p2.Crystal)
	}
	if p3.Gem+p3.Crystal != 3 {
		t.Fatalf("expected p3 receive all extracted energy, got gem=%d crystal=%d", p3.Gem, p3.Crystal)
	}
}

func TestPriestDivineDomain_AllowsPartialDiscardAndHealBranch(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Priest", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
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
	p1.Heal = 0 // 伤害分支不可用，应只出现治疗分支
	p1.Hand = []model.Card{
		{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "priest_divine_domain",
		Selections: []int{
			0,
		},
	})
	if len(p1.Hand) != 0 {
		t.Fatalf("expected partial discard consume 1 card, got hand=%d", len(p1.Hand))
	}
	if p1.Crystal != 0 {
		t.Fatalf("expected crystal spent, got %d", p1.Crystal)
	}
	requireChoicePrompt(t, game, "p1", "priest_divine_domain_mode")

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	ctx := requireChoiceContext(t, game, "p1", "priest_divine_domain_heal_target")
	idx := choiceIndexForTarget(t, ctx, "p2")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{idx},
	})

	if p1.Heal != 2 {
		t.Fatalf("expected priest +2 heal, got %d", p1.Heal)
	}
	if p2.Heal != 1 {
		t.Fatalf("expected ally +1 heal, got %d", p2.Heal)
	}
}

func TestPriestDivineDomain_DamageBranchTargetsAnyPlayer(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Priest", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 1
	p1.Hand = []model.Card{
		{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "a1", Name: "火刃", Type: model.CardTypeAttack, Element: model.ElementFire},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "priest_divine_domain",
		Selections: []int{
			0, 1,
		},
	})
	requireChoicePrompt(t, game, "p1", "priest_divine_domain_mode")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 伤害分支
	})
	ctx := requireChoiceContext(t, game, "p1", "priest_divine_domain_damage_target")
	idx := choiceIndexForTarget(t, ctx, "p3")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{idx},
	})

	if p1.Heal != 0 {
		t.Fatalf("expected damage branch to consume 1 heal, got %d", p1.Heal)
	}
}

func TestPriestWaterPower_DiscardWaterThenGiveSelectedCard(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Priest", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "w1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater},
		{ID: "f1", Name: "火刃", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "t1", Name: "雷枪", Type: model.CardTypeAttack, Element: model.ElementThunder},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "priest_water_power",
		TargetIDs:  []string{"p2"},
		Selections: []int{0, 2}, // 先弃水，再把雷枪交给队友
	})

	if len(p1.Hand) != 1 || p1.Hand[0].ID != "f1" {
		t.Fatalf("expected p1 keeps only fire card, got hand=%+v", p1.Hand)
	}
	if len(p2.Hand) != 1 || p2.Hand[0].ID != "t1" {
		t.Fatalf("expected ally receives selected card t1, got hand=%+v", p2.Hand)
	}
	if p1.Heal != 1 || p2.Heal != 1 {
		t.Fatalf("expected both sides +1 heal, got p1=%d p2=%d", p1.Heal, p2.Heal)
	}
	if len(game.State.DiscardPile) == 0 || game.State.DiscardPile[len(game.State.DiscardPile)-1].ID != "w1" {
		t.Fatalf("expected only water cost card enters discard pile, got discard=%+v", game.State.DiscardPile)
	}
}

func TestPriestWaterPower_NoRemainingCardSkipsGiveStillHealsBoth(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Priest", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "w1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "priest_water_power",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	})

	if len(p1.Hand) != 0 {
		t.Fatalf("expected priest hand empty after paying water cost, got %d", len(p1.Hand))
	}
	if len(p2.Hand) != 0 {
		t.Fatalf("expected ally receives no card when priest has no remaining hand, got hand=%+v", p2.Hand)
	}
	if p1.Heal != 1 || p2.Heal != 1 {
		t.Fatalf("expected both sides +1 heal, got p1=%d p2=%d", p1.Heal, p2.Heal)
	}
	if len(game.State.DiscardPile) == 0 || game.State.DiscardPile[len(game.State.DiscardPile)-1].ID != "w1" {
		t.Fatalf("expected water card in discard pile, got discard=%+v", game.State.DiscardPile)
	}
}

// 回归：神官被动【神圣启示】在一次特殊行动结束后只应触发1次。
func TestPriestDivineRevelation_TriggersOnlyOncePerSpecialAction(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Priest", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p1.Hand = nil

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdBuy,
	})

	if got := p1.Heal; got != 1 {
		t.Fatalf("expected divine revelation heal +1 after one special action, got %d", got)
	}
}

func TestPriestDivineContract_HasXChoiceAndCapsTargetAt4(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Priest", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 3
	p2.Heal = 3

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "priest_divine_contract",
		TargetIDs: []string{"p2"},
	})
	requireChoicePrompt(t, game, "p1", "priest_divine_contract_x")

	// 选择 X=2（索引1）
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})

	if got := p1.Crystal; got != 0 {
		t.Fatalf("expected crystal spent, got %d", got)
	}
	if got := p1.Heal; got != 1 {
		t.Fatalf("expected priest heal reduce by X=2, got %d", got)
	}
	if got := p2.Heal; got != 4 {
		t.Fatalf("expected ally heal capped to 4, got %d", got)
	}
}

func TestPriestDivineContract_TargetAlreadyAbove4KeepsUnchanged(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Priest", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 3
	p2.Heal = 5

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "priest_divine_contract",
		TargetIDs: []string{"p2"},
	})
	requireChoicePrompt(t, game, "p1", "priest_divine_contract_x")

	// 选择 X=2（索引1）
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})

	if got := p1.Heal; got != 1 {
		t.Fatalf("expected priest heal reduce by X=2, got %d", got)
	}
	if got := p2.Heal; got != 5 {
		t.Fatalf("expected ally heal unchanged when already >4, got %d", got)
	}
}
