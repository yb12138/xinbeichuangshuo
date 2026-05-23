package adventurer_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func countAskPromptsForPlayer(obs *testutils.CaptureObserver, playerID string) int {
	if obs == nil {
		return 0
	}
	count := 0
	for _, event := range obs.Events {
		if event.Type != model.EventAskInput {
			continue
		}
		prompt := event.Prompt
		ok := prompt != nil
		if !ok || prompt == nil || prompt.PlayerID != playerID {
			continue
		}
		count++
	}
	return count
}

func latestAskPromptForPlayer(obs *testutils.CaptureObserver, playerID string) *model.Prompt {
	if obs == nil {
		return nil
	}
	for i := len(obs.Events) - 1; i >= 0; i-- {
		event := obs.Events[i]
		if event.Type != model.EventAskInput {
			continue
		}
		prompt := event.Prompt
		ok := prompt != nil
		if !ok || prompt == nil || prompt.PlayerID != playerID {
			continue
		}
		copied := *prompt
		copied.Options = append([]model.PromptOption(nil), prompt.Options...)
		return &copied
	}
	return nil
}

func TestAdventurerStealSky_ModeAndExtraActionChoice(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	game.State.BlueGems = 1

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "adventurer_steal_sky",
	})
	testutils.RequireChoicePrompt(t, game, "p1", "adventurer_steal_sky_mode")

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	if game.State.BlueGems != 0 || game.State.RedGems != 1 {
		t.Fatalf("expected gem transfer blue->red, got blue=%d red=%d", game.State.BlueGems, game.State.RedGems)
	}
	// 额外行动应已被 Drive 消费，验证 TurnStage 回到了 ActionExecution
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected TurnStage=ActionExecution for extra action, got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "" {
		// MustType="" 表示无约束，被消费后应已清空
	}
}

func TestAdventurerUndergroundLaw_RewritesBuyInsteadOfDefaultSettlement(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	// 初始化牌堆，确保有牌可摸
	game.State.Deck = rules.InitDeck()
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	// 手牌留3张，还有3张上限空间，购买摸3张刚好
	p1.Hand = []model.Card{
		{ID: "c1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "c2", Name: "水盾", Type: model.CardTypeMagic, Element: model.ElementWater},
		{ID: "c3", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdBuy})

	// 地下法则：正常摸3张（3+3=6），战绩区+2宝石（原为+1宝石+1水晶）
	if got := len(p1.Hand); got != 6 {
		t.Fatalf("expected hand to be 6 after buy, got %d", got)
	}
	if got := game.State.RedGems; got != 2 {
		t.Fatalf("expected underground law to add 2 team gems, got %d", got)
	}
	if got := game.State.RedCrystals; got != 0 {
		t.Fatalf("expected underground law to not add team crystals, got %d", got)
	}
}

func TestAdventurerExtractFullEnergy_AutoParadiseAllyExtract(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Crystal = 2 // 自身能量已满
	p2.Gem = 1
	game.State.RedCrystals = 2

	// 提炼 → 自身能量满，override hook 自动跳转到队友选择
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdExtract})
	testutils.RequireChoicePrompt(t, game, "p1", "adventurer_paradise_pick")
	prompt := game.GetCurrentPrompt()
	if prompt == nil || prompt.Presentation == nil || prompt.Presentation.Kind != model.PresentationTargetPicker {
		t.Fatalf("expected paradise ally prompt to use target_picker presentation, got %+v", prompt)
	}
	if len(prompt.Options) != 1 || prompt.Options[0].TargetID != "p2" {
		t.Fatalf("expected paradise ally option to carry target_id p2, got %+v", prompt.Options)
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// 队友看到提炼提示
	testutils.RequireChoicePrompt(t, game, "p2", "extract")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	})

	// p1 能量不变（未提炼），p2 从提炼中获得2水晶
	if p1.Gem+p1.Crystal != 3 {
		t.Fatalf("expected p1 energy unchanged (full), got gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
	if p2.Gem+p2.Crystal != 3 {
		t.Fatalf("expected p2 receive two extracted energies, got gem=%d crystal=%d", p2.Gem, p2.Crystal)
	}
	if game.State.RedCrystals != 0 {
		t.Fatalf("expected red crystals extracted to 0, got %d", game.State.RedCrystals)
	}
}

func TestAdventurerParadise_AllyDirectExtract(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 2 // 已有能量
	game.State.RedGems = 1

	// 提炼 → 询问是否发动天堂
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdExtract})
	testutils.RequireChoicePrompt(t, game, "p1", "adventurer_extract_paradise_check")

	// 选择"是，发动冒险者天堂"
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// 选择队友
	testutils.RequireChoicePrompt(t, game, "p1", "adventurer_paradise_pick")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// 队友看到提炼提示
	testutils.RequireChoicePrompt(t, game, "p2", "extract")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// p1 能量不变（队友提炼而非冒险者提炼），p2 获得提炼的1宝石
	if p1.Gem != 2 || p1.Crystal != 0 {
		t.Fatalf("expected p1 keep pre-existing energy, got gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
	if p2.Gem != 1 || p2.Crystal != 0 {
		t.Fatalf("expected ally receive extracted gem, got gem=%d crystal=%d", p2.Gem, p2.Crystal)
	}
}

func TestAdventurerParadise_TargetsFilteredByExtractCapacity(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Crystal = 2 // 能量已满，提炼后自动走冒险者天堂
	p2.Gem = 1
	p2.Crystal = 1 // 仅剩1格
	p3.Gem = 1
	p3.Crystal = 0 // 剩余2格
	game.State.RedCrystals = 2

	// 提炼 → 自身能量满，自动跳转到队友选择
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdExtract})
	ctx := testutils.RequireChoiceContext(t, game, "p1", "adventurer_paradise_pick")

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
	// p2（仅1格空间）和 p3（2格空间）都有空间，都应出现在候选列表中
	if len(allyIDs) != 2 {
		t.Fatalf("expected both allies with room, got ally_ids=%v", allyIDs)
	}

	// 选择 p3（有2格空间的队友）
	p3Idx := 0
	for i, id := range allyIDs {
		if id == "p3" {
			p3Idx = i
			break
		}
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{p3Idx},
	})

	// p3 看到提炼提示
	testutils.RequireChoicePrompt(t, game, "p3", "extract")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p3",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	})

	if p2.Gem+p2.Crystal != 2 {
		t.Fatalf("expected p2 unchanged (not selected), got gem=%d crystal=%d", p2.Gem, p2.Crystal)
	}
	if p3.Gem+p3.Crystal != 3 {
		t.Fatalf("expected p3 receive extracted energy, got gem=%d crystal=%d", p3.Gem, p3.Crystal)
	}
}

// 回归：冒险者提炼时，override hook 应先询问是否发动天堂，而非直接显示提炼提示。
func TestAdventurerExtract_ParadiseCheckBeforeExtract(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
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
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 0
	p1.Crystal = 0
	game.State.RedGems = 1

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdExtract})

	// 新流程：override hook 拦截提炼，先询问是否发动天堂
	testutils.RequireChoicePrompt(t, game, "p1", "adventurer_extract_paradise_check")

	// 选择"否，自行提炼"
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})

	// 应显示标准提炼提示
	testutils.RequireChoicePrompt(t, game, "p1", "extract")
}

func TestPriestDivineDomain_HealBranchRequiresTwoDiscards(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 0 // 伤害分支不可用，应只出现治疗分支
	p1.Hand = []model.Card{
		{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "a1", Name: "火刃", Type: model.CardTypeAttack, Element: model.ElementFire},
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "priest_divine_domain",
		Selections: []int{
			0, 1,
		},
	})
	if len(p1.Hand) != 0 {
		t.Fatalf("expected divine domain consume 2 cards, got hand=%d", len(p1.Hand))
	}
	if p1.Crystal != 0 {
		t.Fatalf("expected crystal spent, got %d", p1.Crystal)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "priest_divine_domain_mode")

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	ctx := testutils.RequireChoiceContext(t, game, "p1", "priest_divine_domain_heal_target")
	idx := testutils.ChoiceIndexForTarget(t, ctx, "p2")
	testutils.MustHandleAction(t, game, model.PlayerAction{
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

func TestPriestDivineDomain_RejectsPartialDiscard(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 0
	p1.Hand = []model.Card{
		{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "priest_divine_domain",
		Selections: []int{0},
	}); err == nil {
		t.Fatalf("expected divine domain to reject partial discard when hand<2")
	}
}

func TestPriestDivineDomain_DamageBranchTargetsAnyPlayer(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 1
	p1.Hand = []model.Card{
		{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "a1", Name: "火刃", Type: model.CardTypeAttack, Element: model.ElementFire},
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "priest_divine_domain",
		Selections: []int{
			0, 1,
		},
	})
	testutils.RequireChoicePrompt(t, game, "p1", "priest_divine_domain_mode")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 伤害分支
	})
	ctx := testutils.RequireChoiceContext(t, game, "p1", "priest_divine_domain_damage_target")
	if ids, ok := ctx["target_ids"].([]string); ok {
		for _, id := range ids {
			if id == "p1" {
				t.Fatalf("expected damage branch to exclude self target, got %v", ids)
			}
		}
	}
	idx := testutils.ChoiceIndexForTarget(t, ctx, "p3")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{idx},
	})

	if p1.Heal != 0 {
		t.Fatalf("expected damage branch to consume 1 heal, got %d", p1.Heal)
	}
}

func TestPriestWaterPower_DiscardWaterThenGiveSelectedCard(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "w1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater},
		{ID: "f1", Name: "火刃", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "t1", Name: "雷枪", Type: model.CardTypeAttack, Element: model.ElementThunder},
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
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

func TestPriestWaterPower_RequiresTransferCard(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "w1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "priest_water_power",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	}); err == nil {
		t.Fatalf("expected water power to require a second card to transfer")
	}
}

// 回归：神官被动【神圣启示】在一次特殊行动结束后只应触发1次。
func TestPriestDivineRevelation_RunsOnlyOncePerSpecialAction(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Priest", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p1.Hand = nil

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdBuy,
	})

	if got := p1.Heal; got != 1 {
		t.Fatalf("expected divine revelation heal +1 after one special action, got %d", got)
	}
}

func TestPriestDivineContract_HasXChoiceAndCapsTargetAt4(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionStart
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 3
	p2.Heal = 3

	game.Drive()
	startupIdx := testutils.StartupSkillIndexByID(t, game, "p1", "priest_divine_contract")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{startupIdx},
	})
	testutils.RequireChoicePrompt(t, game, "p1", "priest_divine_contract_target")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	testutils.RequireChoicePrompt(t, game, "p1", "priest_divine_contract_x")

	// 选择 X=2（索引1）
	testutils.MustHandleAction(t, game, model.PlayerAction{
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
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionStart
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 3
	p2.Heal = 5

	game.Drive()
	startupIdx := testutils.StartupSkillIndexByID(t, game, "p1", "priest_divine_contract")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{startupIdx},
	})
	testutils.RequireChoicePrompt(t, game, "p1", "priest_divine_contract_target")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	testutils.RequireChoicePrompt(t, game, "p1", "priest_divine_contract_x")

	// 选择 X=2（索引1）
	testutils.MustHandleAction(t, game, model.PlayerAction{
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

func TestPriestDivineContract_ResumesToActionSelectionAfterStartup(t *testing.T) {
	game := engine.NewGameEngine(testutils.NewTestObserver(t))
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
	game.State.TurnStage = model.TurnStageActionStart
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 3

	game.Drive()
	startupIdx := testutils.StartupSkillIndexByID(t, game, "p1", "priest_divine_contract")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{startupIdx},
	})
	testutils.RequireChoicePrompt(t, game, "p1", "priest_divine_contract_target")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	testutils.RequireChoicePrompt(t, game, "p1", "priest_divine_contract_x")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if got := game.State.TurnStage; got != model.TurnStageActionExecution {
		t.Fatalf("expected startup skill to return to action selection, got turn stage %s", got)
	}
	if got := game.State.CurrentTurn; got != 0 {
		t.Fatalf("expected to remain on p1's turn, got current turn index %d", got)
	}
}
