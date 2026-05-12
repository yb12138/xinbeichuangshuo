package bard_test

import (
	"fmt"
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	bardpkg "starcup-engine/internal/engine/player/bard"

	"starcup-engine/internal/data"
	"starcup-engine/internal/model"
	"strings"
)

func bardTestCard(id, name string, cardType model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     ele,
		Damage:      1,
		Description: name,
	}
}

func addBardExclusiveCardsForTest(p *model.Player, titles ...string) {
	if p == nil {
		return
	}
	// Now bard only has one exclusive card "永恒乐章"
	for _, title := range titles {
		switch title {
		case "永恒乐章":
			p.RestoreExclusiveCard(model.Card{
				ID:             fmt.Sprintf("starter-%s-bd_eternal_movement", p.ID),
				Name:           "永恒乐章",
				Type:           model.CardTypeMagic,
				Element:        model.ElementDark,
				Description:    "吟游诗人专属牌",
				ExclusiveChar1: p.Character.ID,
			})
		}
	}
}

func findFieldEffectCard(p *model.Player, effect model.EffectType) *model.FieldCard {
	if p == nil {
		return nil
	}
	for _, fc := range p.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == effect {
			return fc
		}
	}
	return nil
}

func TestBardDescentConcerto_RunsAndResolves(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Hand = []model.Card{
		bardTestCard("f_magic", "火法术", model.CardTypeMagic, model.ElementFire),
		bardTestCard("f_attack", "火攻击", model.CardTypeAttack, model.ElementFire),
		bardTestCard("w_attack", "水攻击", model.CardTypeAttack, model.ElementWater),
	}

	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p1", TargetID: "p3", Damage: 1, DamageType: model.MagicAttack,
	}); paused {
		t.Fatalf("first magic damage should not dispatch descent yet")
	}
	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p1", TargetID: "p4", Damage: 1, DamageType: model.MagicAttack,
	}); !paused {
		t.Fatalf("second self magic damage should dispatch descent interrupt")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_descent_element")

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 选火系
		t.Fatalf("choose descent element failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_descent_cards")

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 第1张火牌
		t.Fatalf("choose first discard failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_descent_cards")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 第2张火牌
		t.Fatalf("choose second discard failed: %v", err)
	}

	testutils.RequireChoicePrompt(t, game, "p1", "bd_descent_target")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose descent bonus target failed: %v", err)
	}

	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected inspiration=1, got %d", got)
	}
	if got := bard.TurnState.UsedSkillCounts["bd_descent"]; got != 1 {
		t.Fatalf("expected descent used flag=1, got %d", got)
	}
	if got := len(bard.Hand); got != 1 {
		t.Fatalf("expected bard hand reduced to 1, got %d", got)
	}
	if got := len(game.State.PendingDamageQueue); got != 1 {
		t.Fatalf("expected one bonus pending damage, got %d", got)
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.SourceID != "p1" || pd.DamageType != "magic" || pd.Damage != 1 {
		t.Fatalf("unexpected bonus damage payload: %+v", pd)
	}
}

func TestBardDescentConcerto_DoesNotTimingOnAllyMagicDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	ally := game.State.Players["p2"]
	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()

	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p2", TargetID: "p3", Damage: 1, DamageType: model.MagicAttack,
	}); paused {
		t.Fatalf("ally magic damage should not dispatch bard descent on first hit")
	}
	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p2", TargetID: "p4", Damage: 1, DamageType: model.MagicAttack,
	}); paused {
		t.Fatalf("ally magic damage should not dispatch bard descent on second hit")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt, got %+v", game.State.PendingInterrupt)
	}
}

func TestBardDissonanceChord_DrawModeAndReleasePrisoner(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	enemy := game.State.Players["p2"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Tokens["bd_inspiration"] = 3
	bard.Form = model.FormBardEternalPrisoner
	bard.Hand = []model.Card{
		bardTestCard("h1", "手牌1", model.CardTypeAttack, model.ElementFire),
	}
	enemy.Hand = []model.Card{
		bardTestCard("e1", "敌方牌1", model.CardTypeAttack, model.ElementWater),
	}
	game.State.Deck = []model.Card{
		bardTestCard("d1", "牌堆1", model.CardTypeAttack, model.ElementWind),
		bardTestCard("d2", "牌堆2", model.CardTypeAttack, model.ElementThunder),
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "bd_dissonance_chord", nil, nil); err != nil {
		t.Fatalf("use dissonance failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_dissonance_x")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // X=2
		t.Fatalf("choose X failed: %v", err)
	}
	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected inspiration consumed to 1, got %d", got)
	}
	if got := bard.Form; got != "" {
		t.Fatalf("expected prisoner form released, got %q", got)
	}

	testutils.RequireChoicePrompt(t, game, "p1", "bd_dissonance_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 摸牌分支
		t.Fatalf("choose mode failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_dissonance_target")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil { // 目标选 p2
		t.Fatalf("choose target failed: %v", err)
	}

	if got := len(bard.Hand); got != 2 {
		t.Fatalf("expected bard drew 1 card, hand=%d", got)
	}
	if got := len(enemy.Hand); got != 2 {
		t.Fatalf("expected target drew 1 card, hand=%d", got)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected dissonance flow completed, got pending interrupt %+v", game.State.PendingInterrupt)
	}
}

func TestBardHopeFugue_PlaceUsesPlayedCardAsEternalMovement(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	// No longer need exclusive cards for hope_fugue - it's a character skill now
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Crystal = 1
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	if err := game.UseSkill("p1", "bd_hope_fugue", nil, nil); err != nil {
		t.Fatalf("use hope fugue failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_draw_confirm")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil { // 不摸牌
		t.Fatalf("choose draw confirm failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 放置分支
		t.Fatalf("choose hope mode failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_place_target")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 目标队友 p2
		t.Fatalf("choose place target failed: %v", err)
	}

	if holder := bardpkg.EternalHolderID(engine.NewRoleChoiceRuntime(game), bard); holder != "p2" {
		t.Fatalf("expected eternal movement holder p2, got %q", holder)
	}
	fieldCard := findFieldEffectCard(ally, model.EffectBardEternalMovement)
	if fieldCard == nil {
		t.Fatalf("expected ally to hold eternal movement field card")
	}
	if fieldCard.Card.Name != "永恒乐章" {
		t.Fatalf("expected eternal movement field entity, got %s", fieldCard.Card.Name)
	}
}

func TestBardHopeFugue_TransferMovesExistingEternalMovementAndGainsInspiration(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "AllyA", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AllyB", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	allyA := game.State.Players["p2"]
	allyB := game.State.Players["p3"]
	// No longer need exclusive cards - hope_fugue is a character skill now
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Crystal = 1
	bard.Hand = []model.Card{
		bardTestCard("discard", "弃牌", model.CardTypeAttack, model.ElementFire),
	}
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, allyA); err != nil {
		t.Fatalf("place initial eternal movement failed: %v", err)
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	if err := game.UseSkill("p1", "bd_hope_fugue", nil, nil); err != nil {
		t.Fatalf("use hope fugue failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_draw_confirm")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil { // 不摸牌
		t.Fatalf("choose draw confirm failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{2}}); err != nil { // 转移并+1灵感
		t.Fatalf("choose hope mode failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_transfer_target")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 目标队友 p3
		t.Fatalf("choose transfer target failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_transfer_discard")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose transfer discard failed: %v", err)
	}

	if holder := bardpkg.EternalHolderID(engine.NewRoleChoiceRuntime(game), bard); holder != "p3" {
		t.Fatalf("expected eternal movement holder p3 after transfer, got %q", holder)
	}
	if findFieldEffectCard(allyA, model.EffectBardEternalMovement) != nil {
		t.Fatalf("expected allyA to no longer hold eternal movement")
	}
	if findFieldEffectCard(allyB, model.EffectBardEternalMovement) == nil {
		t.Fatalf("expected allyB to hold transferred eternal movement")
	}
	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected bard inspiration +1 after transfer mode 2, got %d", got)
	}
	// Only the discarded hand card is in discard pile (no exclusive card consumed)
	if got := len(game.State.DiscardPile); got != 1 {
		t.Fatalf("expected discard pile contain only discarded hand card, got %d", got)
	}
}

func TestBardRousingRhapsody_OnAllyTurnStartRunsForbiddenVerse(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	// 永恒乐章放置在队友身上，队友回合开始时触发响应询问
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}

	// ally 是当前回合玩家（永恒乐章持有者）
	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	ally.TurnState.HasProcessedTurnStart = false // 确保 TurnStart hooks 能被触发
	game.State.CurrentTurn = 1 // p2 是 index 1
	game.State.TurnStage = model.TurnStageTurnStart // 回合开始阶段触发激昂狂想曲
	game.State.PendingInterrupt = nil

	game.Drive()

	// 响应询问发给吟游诗人（bard），因为 bard 才是技能的主人
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	testutils.ChooseResponseSkillByID(t, game, "p1", "bd_rousing_rhapsody")
	testutils.RequireChoicePrompt(t, game, "p1", "bd_rousing_mode")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 选伤害分支
		t.Fatalf("choose rousing mode failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_rousing_targets")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 先选 p3
		t.Fatalf("choose rousing first target failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_rousing_targets")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 再选 p4
		t.Fatalf("choose rousing second target failed: %v", err)
	}

	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected forbidden verse add inspiration to 1, got %d", got)
	}
	if holder := bardpkg.EternalHolderID(engine.NewRoleChoiceRuntime(game), bard); holder != "" {
		t.Fatalf("expected eternal movement removed by forbidden verse, holder=%q", holder)
	}
	if got := len(game.State.PendingDamageQueue); got != 2 {
		t.Fatalf("expected rousing queued 2 magic damages, got %d", got)
	}
	if game.State.CombatStage != model.CombatStageCalcDamage || game.State.ReturnTurnStage != model.TurnStageActionStart {
		t.Fatalf("expected damage resolution then return to action start, combat=%s return_turn=%s", game.State.CombatStage, game.State.ReturnTurnStage)
	}
}

func TestBardVictorySymphony_AtInspirationCapEntersPrisonerAndSelfDamages(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	// Victory symphony no longer requires exclusive card - only needs eternal movement on field
	bard.Tokens["bd_inspiration"] = 3
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}

	// ally 是当前回合玩家（永恒乐章持有者），回合结束时触发响应询问
	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 1 // p2 是 index 1
	game.State.TurnStage = model.TurnStageTurnEnd // 回合结束阶段触发胜利交响诗
	game.State.PendingInterrupt = nil

	game.Drive()

	testutils.RequireResponseSkillPrompt(t, game, "p1")
	testutils.ChooseResponseSkillByID(t, game, "p1", "bd_victory_symphony")
	testutils.RequireChoicePrompt(t, game, "p1", "bd_victory_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil { // 分支②
		t.Fatalf("choose victory mode failed: %v", err)
	}

	if got := bard.Form; got != model.FormBardEternalPrisoner {
		t.Fatalf("expected bard enter prisoner form at inspiration cap, got %q", got)
	}
	if got := len(game.State.PendingDamageQueue); got != 1 {
		t.Fatalf("expected one self magic damage from forbidden verse, got %d", got)
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.TargetID != "p1" || pd.DamageType != "magic" || pd.Damage != 3 {
		t.Fatalf("unexpected self-damage payload: %+v", pd)
	}
}

func TestBardVictorySymphony_ExtractStoneChoosesGemOrCrystal(t *testing.T) {
	tests := []struct {
		name        string
		gems        int
		crystals    int
		choiceIndex int
		wantGem     int
		wantCrystal int
	}{
		{name: "gem", gems: 1, crystals: 0, choiceIndex: 0, wantGem: 1},
		{name: "crystal", gems: 0, crystals: 1, choiceIndex: 0, wantCrystal: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			game := engine.NewGameEngine(testutils.NoopObserver{})
			if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
				t.Fatal(err)
			}

			bard := game.State.Players["p1"]
			ally := game.State.Players["p2"]
			// Victory symphony no longer requires exclusive card
			if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
				t.Fatalf("place eternal movement failed: %v", err)
			}
			game.State.RedGems = tc.gems
			game.State.RedCrystals = tc.crystals

			ally.IsActive = true
			ally.TurnState = model.NewPlayerTurnState()
			game.State.CurrentTurn = 1 // p2 是 index 1
			game.State.TurnStage = model.TurnStageTurnEnd

			game.Drive()
			testutils.RequireResponseSkillPrompt(t, game, "p1")
			testutils.ChooseResponseSkillByID(t, game, "p1", "bd_victory_symphony")
			testutils.RequireChoicePrompt(t, game, "p1", "bd_victory_mode")
			if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
				t.Fatalf("choose extract mode failed: %v", err)
			}
			testutils.RequireChoicePrompt(t, game, "p1", "bd_victory_extract_stone")
			if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{tc.choiceIndex}}); err != nil {
				t.Fatalf("choose stone failed: %v", err)
			}

			if bard.Gem != tc.wantGem || bard.Crystal != tc.wantCrystal {
				t.Fatalf("unexpected energy after extract: gem=%d crystal=%d", bard.Gem, bard.Crystal)
			}
		})
	}
}

func TestBardConfig_MetadataAlignsWithDocument(t *testing.T) {
	characters := data.GetCharacters()
	var bard *model.Character
	for _, character := range characters {
		if character.ID == "bard" {
			copy := character
			bard = &copy
			break
		}
	}
	if bard == nil {
		t.Fatalf("bard character not found")
	}

	var descent *model.SkillDefinition
	var dissonance *model.SkillDefinition
	var rousing *model.SkillDefinition
	var victory *model.SkillDefinition
	var hope *model.SkillDefinition
	for i := range bard.Skills {
		switch bard.Skills[i].ID {
		case "bd_descent_concerto":
			descent = &bard.Skills[i]
		case "bd_dissonance_chord":
			dissonance = &bard.Skills[i]
		case "bd_rousing_rhapsody":
			rousing = &bard.Skills[i]
		case "bd_victory_symphony":
			victory = &bard.Skills[i]
		case "bd_hope_fugue":
			hope = &bard.Skills[i]
		}
	}
	if descent == nil || dissonance == nil || rousing == nil || victory == nil || hope == nil {
		t.Fatalf("expected bard core skills present")
	}
	if descent.TargetType != model.TargetEnemy || descent.MinTargets != 1 || descent.MaxTargets != 1 {
		t.Fatalf("expected descent target metadata enemy(1), got type=%v min=%d max=%d", descent.TargetType, descent.MinTargets, descent.MaxTargets)
	}
	if dissonance.TargetType != model.TargetAny || dissonance.MinTargets != 1 || dissonance.MaxTargets != 1 {
		t.Fatalf("expected dissonance target metadata any(1), got type=%v min=%d max=%d", dissonance.TargetType, dissonance.MinTargets, dissonance.MaxTargets)
	}
	// Rousing, Victory, and Hope are now character skills (RequireExclusive=false)
	if rousing.RequireExclusive || rousing.TargetType != model.TargetEnemy || rousing.MinTargets != 0 || rousing.MaxTargets != 2 {
		t.Fatalf("expected rousing metadata NO exclusive + enemy(0..2), got requireExclusive=%v type=%v min=%d max=%d",
			rousing.RequireExclusive, rousing.TargetType, rousing.MinTargets, rousing.MaxTargets)
	}
	if victory.RequireExclusive {
		t.Fatalf("expected victory symphony NO require exclusive")
	}
	if hope.RequireExclusive || hope.TargetType != model.TargetAlly || hope.MinTargets != 1 || hope.MaxTargets != 1 {
		t.Fatalf("expected hope metadata require exclusive + ally(1), got requireExclusive=%v type=%v min=%d max=%d",
			hope.RequireExclusive, hope.TargetType, hope.MinTargets, hope.MaxTargets)
	}
}

func TestBardStarterExclusiveCards_NotInHand(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	if err := game.StartGame(); err != nil {
		t.Fatalf("start game failed: %v", err)
	}

	bard := game.State.Players["p1"]
	if bard == nil {
		t.Fatalf("bard player missing")
	}
	if got := len(bard.Hand); got != 4 {
		t.Fatalf("expected bard starting hand remain 4, got %d", got)
	}
	// Bard now has only one exclusive card: 永恒乐章
	if len(bard.ExclusiveCards) != 1 {
		t.Fatalf("expected exactly 1 exclusive card, got %d", len(bard.ExclusiveCards))
	}
	if bard.ExclusiveCards[0].Name != "永恒乐章" {
		t.Fatalf("expected bard starter exclusive card 永恒乐章, got %s", bard.ExclusiveCards[0].Name)
	}
}

// TestBardRousingRhapsody_BleedTickRunsFirst 验证当血之巫女在流血形态且持有永恒乐章时，
// 回合开始先触发流血效果（伤害结算），再触发激昂狂想曲的响应询问。
func TestBardRousingRhapsody_BleedTickRunsFirst(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "BloodWitch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	witch := game.State.Players["p2"]

	// 永恒乐章放置在血之巫女身上
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, witch); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}

	// 血之巫女进入流血形态
	witch.Form = model.FormBloodPriestessBleeding

	// 血之巫女回合开始
	witch.IsActive = true
	witch.TurnState = model.NewPlayerTurnState()
	witch.TurnState.HasProcessedTurnStart = false
	game.State.CurrentTurn = 1 // p2 是 index 1
	game.State.TurnStage = model.TurnStageTurnStart
	game.State.PendingInterrupt = nil

	game.Drive()

	// Drive 会自动处理：流血伤害先结算 → 然后激昂狂想曲弹出响应询问
	// 最终状态应为等待吟游诗人响应技能
	testutils.RequireResponseSkillPrompt(t, game, "p1")

	// 验证日志顺序：流血在前，激昂狂想曲在后
	var bleedIdx, rousingIdx int = -1, -1
	for i, e := range obs.Events {
		if e.Type == model.EventLog {
			if bleedIdx < 0 && strings.Contains(e.Message, "流血") {
				bleedIdx = i
			}
			if rousingIdx < 0 && strings.Contains(e.Message, "激昂狂想曲") {
				rousingIdx = i
			}
		}
	}
	if bleedIdx < 0 {
		t.Fatal("expected bleed log entry")
	}
	if rousingIdx < 0 {
		t.Fatal("expected rousing rhapsody log entry")
	}
	if bleedIdx > rousingIdx {
		t.Fatalf("expected bleed log before rousing log, got bleed at %d and rousing at %d", bleedIdx, rousingIdx)
	}
}
