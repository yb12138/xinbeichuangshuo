package magic_lancer_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"strings"
	"testing"

	"starcup-engine/internal/data"
	"starcup-engine/internal/engine/player/magic_lancer"
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

func magicLancerTestCard(id, name string, cardType model.CardType, element model.Element, damage int) model.Card {
	if damage <= 0 {
		damage = 2
	}
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     element,
		Damage:      damage,
		Description: name,
	}
}

func TestMagicLancerDarkRelease_HandCapAndAttackBonusAndLock(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart
	if err := game.UseSkill("p1", "ml_dark_release", nil, nil); err != nil {
		t.Fatalf("use ml_dark_release failed: %v", err)
	}

	if got := p1.Form; got != model.FormMagicLancerPhantom {
		t.Fatalf("expected magic lancer phantom form, got %q", got)
	}
	if got := game.GetMaxHand(p1); got != 5 {
		t.Fatalf("expected max hand=5 in phantom form, got %d", got)
	}

	fullnessHandler := skills.GetHandler("ml_fullness")
	if fullnessHandler == nil {
		t.Fatal("ml_fullness handler not found")
	}
	ctx := game.BuildContext(p1, nil, model.TimingActionDuring, nil)
	if fullnessHandler.CanUse(ctx) {
		t.Fatal("ml_fullness should be locked in the same turn after dark release")
	}
	if !game.IsSkillBlocked("p1", "ml_fullness") {
		t.Fatal("expected ml_fullness to be blocked by active rule modifier")
	}

	blackSpearHandler := skills.GetHandler("ml_black_spear")
	if blackSpearHandler == nil {
		t.Fatal("ml_black_spear handler not found")
	}
	p2.Hand = []model.Card{magicLancerTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire, 2)}
	hitCtx := game.BuildContext(p1, p2, model.TimingAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: p1.ID,
		TargetID: p2.ID,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       "Attack",
			IsHit:            true,
			CounterInitiator: "",
		},
	})
	if blackSpearHandler.CanUse(hitCtx) {
		t.Fatal("ml_black_spear should be locked in the same turn after dark release")
	}
	if !game.IsSkillBlocked("p1", "ml_black_spear") {
		t.Fatal("expected ml_black_spear to be blocked by active rule modifier")
	}

	// 攻击伤害 +1 的消耗时机见 TestMagicLancerDarkRelease_BonusConsumedAtFirstAttackDeclare*。
	if got := engine.AttackDamageRuleBonusForModifier(p1, "ml_dark_release_next_attack_bonus"); got != 1 {
		t.Fatalf("expected dark release bonus armed before first attack, got %d", got)
	}

	game.NextTurn()
	if game.IsSkillBlocked("p1", "ml_fullness") || game.IsSkillBlocked("p1", "ml_black_spear") {
		t.Fatal("expected dark release lock to expire at turn end")
	}
}

func TestMagicLancerDarkRelease_TriggersOverflowDiscardWith6Cards(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	// 6张手牌，超过幻影形态的上限5张
	p1.Hand = []model.Card{
		magicLancerTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire, 2),
		magicLancerTestCard("h2", "水流斩", model.CardTypeAttack, model.ElementWater, 2),
		magicLancerTestCard("h3", "风神斩", model.CardTypeAttack, model.ElementWind, 2),
		magicLancerTestCard("h4", "雷光斩", model.CardTypeAttack, model.ElementThunder, 2),
		magicLancerTestCard("h5", "地裂斩", model.CardTypeAttack, model.ElementEarth, 2),
		magicLancerTestCard("h6", "暗灭斩", model.CardTypeAttack, model.ElementDark, 2),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart
	if err := game.UseSkill("p1", "ml_dark_release", nil, nil); err != nil {
		t.Fatalf("use ml_dark_release failed: %v", err)
	}

	if p1.Form != model.FormMagicLancerPhantom {
		t.Fatalf("expected magic lancer phantom form, got %q", p1.Form)
	}
	// 6张手牌 -> 上限5张 -> 需要弃1张牌，应该触发弃牌中断
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard selection interrupt after dark release with 6 cards, got %+v", game.State.PendingInterrupt)
	}
}

func TestMagicLancerPhantomStardust_LeaveFormAndPromptTarget(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormMagicLancerPhantom
	p1.Hand = []model.Card{}
	p2.Hand = []model.Card{
		magicLancerTestCard("seed", "起始牌", model.CardTypeAttack, model.ElementFire, 2),
	}
	game.State.Deck = []model.Card{
		magicLancerTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		magicLancerTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		magicLancerTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWind, 2),
		magicLancerTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementThunder, 2),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart
	if err := game.UseSkill("p1", "ml_phantom_stardust", nil, nil); err != nil {
		t.Fatalf("use ml_phantom_stardust failed: %v", err)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending self magic damage from ml_phantom_stardust")
	}

	interrupted := game.ProcessPendingDamages()
	if !interrupted {
		t.Fatalf("expected ProcessPendingDamages to pause on stardust target prompt")
	}
	if got := p1.Form; got != "" {
		t.Fatalf("expected leave phantom form after stardust self damage, got %q", got)
	}
	if got := p1.TurnState.SkillFlowState["ml_stardust_pending"]; got != 0 {
		t.Fatalf("expected ml_stardust_pending cleared, got %d", got)
	}
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending interrupt for stardust target")
	}
	if got := testutils.ChoiceTypeOfInterrupt(game.State.PendingInterrupt); got != "ml_stardust_target" {
		t.Fatalf("expected choice_type ml_stardust_target, got %q", got)
	}
	targetIDs := testutils.PendingChoiceTargetIDs(game.State.PendingInterrupt)
	if len(targetIDs) != 1 || targetIDs[0] != "p2" {
		t.Fatalf("expected stardust target pool to include enemy only, got %v", targetIDs)
	}
}

func TestMagicLancerPhantomStardust_PreselectedEnemySkipsTargetPrompt(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormMagicLancerPhantom
	p1.Hand = []model.Card{}
	p2.Hand = []model.Card{
		magicLancerTestCard("seed", "起始牌", model.CardTypeAttack, model.ElementFire, 2),
	}
	game.State.Deck = []model.Card{
		magicLancerTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		magicLancerTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		magicLancerTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWind, 2),
		magicLancerTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementThunder, 2),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart
	if err := game.UseSkill("p1", "ml_phantom_stardust", []string{"p2"}, nil); err != nil {
		t.Fatalf("use ml_phantom_stardust with preselected target failed: %v", err)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending self magic damage from ml_phantom_stardust")
	}

	interrupted := game.ProcessPendingDamages()
	if interrupted {
		t.Fatalf("expected preselected stardust target to skip second target prompt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after preselected stardust target, got %+v", game.State.PendingInterrupt)
	}
	if got := len(game.State.Players["p2"].Hand); got != 3 {
		t.Fatalf("expected preselected enemy to immediately resolve 2 magic damage, got hand=%d", got)
	}
}

func TestMagicLancerDarkBind_BlocksMagicUseAndDefend(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{magicLancerTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0)}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	if err := game.PerformMagic("p1", "p2", 0); err == nil || !strings.Contains(err.Error(), "法术牌") {
		t.Fatalf("expected dark bind to block PerformMagic, got err=%v", err)
	}

	game.State.CombatStack = []model.CombatRequest{{
		AttackerID:     "p2",
		TargetID:       "p1",
		Card:           &model.Card{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		CanBeResponded: true,
	}}
	game.State.CombatStage = model.CombatStageDeclare
	err := game.HandleCombatResponse(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdRespond,
		CardID:    testutils.PlayableCardID(t, game, "p1", 0),
		ExtraArgs: []string{"defend"},
	})
	if err == nil || !strings.Contains(err.Error(), "黑暗束缚") {
		t.Fatalf("expected dark bind to block defend, got err=%v", err)
	}
}

func TestMagicLancerFullness_FlowBonusAndExtraAttack(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		magicLancerTestCard("cost", "圣光", model.CardTypeMagic, model.ElementLight, 0),
		magicLancerTestCard("atk", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
	}
	p2.Hand = []model.Card{magicLancerTestCard("ally", "雷击", model.CardTypeAttack, model.ElementThunder, 2)}
	p3.Hand = []model.Card{magicLancerTestCard("enemy", "圣光", model.CardTypeMagic, model.ElementLight, 0)}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	if err := game.UseSkill("p1", "ml_fullness", []string{"p2"}, nil); err != nil {
		t.Fatalf("use ml_fullness failed: %v", err)
	}
	if got := testutils.ChoiceTypeOfInterrupt(game.State.PendingInterrupt); got != "ml_fullness_cost_card" {
		t.Fatalf("expected ml_fullness_cost_card prompt, got %q", got)
	}

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose fullness cost card failed: %v", err)
	}
	if got := testutils.ChoiceTypeOfInterrupt(game.State.PendingInterrupt); got != "ml_fullness_discard_step" {
		t.Fatalf("expected ml_fullness_discard_step prompt, got %q", got)
	}

	// 敌方：必须弃牌（PlayerID = p3 自己），仅有1项可选。
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p3", Selections: []int{0}}); err != nil {
		t.Fatalf("enemy discard failed: %v", err)
	}
	// 队友：可选择不弃（PlayerID = p2 自己），index 0 = "不弃置"。
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil {
		t.Fatalf("ally skip failed: %v", err)
	}

	if got := engine.AttackDamageRuleBonusForModifier(p1, "ml_fullness_next_attack_bonus"); got != 1 {
		t.Fatalf("expected ml_fullness_next_attack_bonus=1, got %d", got)
	}
	if len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected extra attack action token from ml_fullness")
	}
	last := p1.TurnState.PendingActions[len(p1.TurnState.PendingActions)-1]
	if last.MustType != "Attack" {
		t.Fatalf("expected extra action type Attack, got %+v", last)
	}
	if len(p3.Hand) != 0 {
		t.Fatalf("expected enemy hand to be discarded, got %d cards", len(p3.Hand))
	}

	dmg := game.ApplyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID:         p1.ID,
		TargetID:         p2.ID,
		Type:             model.ActionAttack,
		CounterInitiator: "",
		Card:             &model.Card{ID: "atk", Name: "雷斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
	})
	if dmg != 3 {
		t.Fatalf("expected fullness bonus damage to apply once (2+1), got %d", dmg)
	}
	if got := engine.AttackDamageRuleBonusForModifier(p1, "ml_fullness_next_attack_bonus"); got != 0 {
		t.Fatalf("expected fullness bonus consumed, got %d", got)
	}
}

func TestMagicLancerConfig_MetadataAlignsWithDocument(t *testing.T) {
	characters := data.GetCharacters()
	var lancer *model.Character
	for _, character := range characters {
		if character.ID == "magic_lancer" {
			copy := character
			lancer = &copy
			break
		}
	}
	if lancer == nil {
		t.Fatalf("magic_lancer character not found")
	}

	var stardust *model.SkillDefinition
	var fullness *model.SkillDefinition
	for i := range lancer.Skills {
		switch lancer.Skills[i].ID {
		case "ml_phantom_stardust":
			stardust = &lancer.Skills[i]
		case "ml_fullness":
			fullness = &lancer.Skills[i]
		}
	}
	if stardust == nil || fullness == nil {
		t.Fatalf("expected phantom stardust and fullness skills present")
	}
	if stardust.TargetType != model.TargetEnemy || stardust.MinTargets != 1 || stardust.MaxTargets != 1 {
		t.Fatalf("expected phantom stardust target metadata enemy(1), got type=%v min=%d max=%d", stardust.TargetType, stardust.MinTargets, stardust.MaxTargets)
	}
	if fullness.TargetType != model.TargetNone {
		t.Fatalf("expected fullness target metadata none, got type=%v", fullness.TargetType)
	}
}

func TestMagicLancerBlackSpear_ConsumesCrystalAndAddsDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormMagicLancerPhantom
	p1.Crystal = 2
	p2.Hand = []model.Card{magicLancerTestCard("h1", "火斩", model.CardTypeAttack, model.ElementFire, 2)}

	handler := skills.GetHandler("ml_black_spear")
	if handler == nil {
		t.Fatal("ml_black_spear handler not found")
	}
	ctx := game.BuildContext(p1, p2, model.TimingAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: p1.ID,
		TargetID: p2.ID,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       "Attack",
			IsHit:            true,
			CounterInitiator: "",
		},
	})
	if !handler.CanUse(ctx) {
		t.Fatal("expected ml_black_spear can use on active hit vs hand 1/2 target")
	}
	game.State.PendingDamageQueue = []model.PendingDamage{{
		SourceID:   p1.ID,
		TargetID:   p2.ID,
		Damage:     2,
		DamageType: model.AttackDamage,
	}}
	if err := handler.Execute(ctx); err != nil {
		t.Fatalf("execute ml_black_spear failed: %v", err)
	}
	if got := testutils.ChoiceTypeOfInterrupt(game.State.PendingInterrupt); got != "ml_black_spear_x" {
		t.Fatalf("expected ml_black_spear_x prompt, got %q", got)
	}

	// selection=1 -> X=2
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose black spear x failed: %v", err)
	}
	if p1.Crystal != 0 || p1.Gem != 0 {
		t.Fatalf("expected consume 2 crystal-like resources, got gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending attack damage entry")
	}
	if got := game.State.PendingDamageQueue[0].Damage; got != 6 {
		t.Fatalf("expected attack damage increased to 6 (2 + (2+2)), got %d", got)
	}
}

func TestMagicLancerFullness_AllyDiscardCountsBonus(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		magicLancerTestCard("cost", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	p2.Hand = []model.Card{magicLancerTestCard("ally", "雷光斩", model.CardTypeAttack, model.ElementThunder, 2)} // 队友有雷系牌
	p3.Hand = []model.Card{magicLancerTestCard("enemy", "圣光", model.CardTypeMagic, model.ElementLight, 0)}    // 敌方有法术牌

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	if err := game.UseSkill("p1", "ml_fullness", []string{"p2"}, nil); err != nil {
		t.Fatalf("use ml_fullness failed: %v", err)
	}
	// p1 弃置发动费用
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose fullness cost card failed: %v", err)
	}

	// 敌方 p3 必须弃牌，index=0（敌方无"不弃置"选项，options 只有手牌）
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p3", Selections: []int{0}}); err != nil {
		t.Fatalf("enemy discard failed: %v", err)
	}
	// 队友 p2 选择弃牌（雷系牌），index=1（index=0是"不弃置"，index=1是第一张手牌）
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{1}}); err != nil {
		t.Fatalf("ally discard failed: %v", err)
	}

	// 敌方弃法术牌 +1，队友弃雷系牌 +1 = 总共 +2
	if got := engine.AttackDamageRuleBonusForModifier(p1, "ml_fullness_next_attack_bonus"); got != 2 {
		t.Fatalf("expected ml_fullness_next_attack_bonus=2 (enemy magic + ally thunder), got %d", got)
	}
	if len(p2.Hand) != 0 {
		t.Fatalf("expected ally hand to be discarded, got %d cards", len(p2.Hand))
	}
}

func TestMagicLancerFullness_AllyCanCancelAndFirstCardSelectsDiscard(t *testing.T) {
	t.Run("skip exposes cancel metadata and keeps hand intact", func(t *testing.T) {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p3", "Enemy", "angel", model.BlueCamp); err != nil {
			t.Fatal(err)
		}

		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p3 := game.State.Players["p3"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		p1.Hand = []model.Card{
			magicLancerTestCard("cost", "圣光", model.CardTypeMagic, model.ElementLight, 0),
		}
		p2.Hand = []model.Card{
			magicLancerTestCard("ally-0", "雷击", model.CardTypeAttack, model.ElementThunder, 2),
			magicLancerTestCard("ally-1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
		}
		p3.Hand = []model.Card{
			magicLancerTestCard("enemy-0", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
		}

		game.State.CurrentTurn = 0
		game.State.TurnStage = model.TurnStageActionExecution
		if err := game.UseSkill("p1", "ml_fullness", []string{"p2"}, nil); err != nil {
			t.Fatalf("use ml_fullness failed: %v", err)
		}
		if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
			t.Fatalf("choose fullness cost card failed: %v", err)
		}
		if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p3", Selections: []int{0}}); err != nil {
			t.Fatalf("enemy discard failed: %v", err)
		}

		prompt := game.BuildChoicePrompt()
		if prompt == nil || prompt.Presentation == nil {
			t.Fatal("expected ally prompt with presentation")
		}
		if prompt.Presentation.CancelPolicy != "decline" || !prompt.Presentation.HasDecline || prompt.Presentation.DeclineIndex != 0 {
			t.Fatalf("expected decline cancel metadata, got %+v", prompt.Presentation)
		}
		if len(prompt.Options) < 2 {
			t.Fatalf("expected skip plus card options, got %+v", prompt.Options)
		}
		if prompt.Options[0].ID != "-1" {
			t.Fatalf("expected first option to be skip, got %+v", prompt.Options[0])
		}
		if prompt.Options[1].ID != "1" || prompt.Options[1].CardID != p2.Hand[0].ID {
			t.Fatalf("expected first card option to be index 1 and first hand card, got %+v", prompt.Options[1])
		}

		if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdCancel, PlayerID: "p2"}); err != nil {
			t.Fatalf("ally skip via cancel failed: %v", err)
		}
		if game.State.PendingInterrupt != nil {
			t.Fatalf("expected prompt to finish after skip, got %+v", game.State.PendingInterrupt)
		}
		if len(p2.Hand) != 2 {
			t.Fatalf("expected ally hand to remain untouched on skip, got %d cards", len(p2.Hand))
		}
		if got := engine.AttackDamageRuleBonusForModifier(p1, "ml_fullness_next_attack_bonus"); got != 1 {
			t.Fatalf("expected only enemy discard bonus from skip flow, got %d", got)
		}
	})

	t.Run("first card is discarded by card id", func(t *testing.T) {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p3", "Enemy", "angel", model.BlueCamp); err != nil {
			t.Fatal(err)
		}

		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p3 := game.State.Players["p3"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		p1.Hand = []model.Card{
			magicLancerTestCard("cost", "圣光", model.CardTypeMagic, model.ElementLight, 0),
		}
		p2.Hand = []model.Card{
			magicLancerTestCard("ally-0", "雷击", model.CardTypeAttack, model.ElementThunder, 2),
			magicLancerTestCard("ally-1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
		}
		p3.Hand = []model.Card{
			magicLancerTestCard("enemy-0", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
		}

		game.State.CurrentTurn = 0
		game.State.TurnStage = model.TurnStageActionExecution
		if err := game.UseSkill("p1", "ml_fullness", []string{"p2"}, nil); err != nil {
			t.Fatalf("use ml_fullness failed: %v", err)
		}
		if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
			t.Fatalf("choose fullness cost card failed: %v", err)
		}
		if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p3", Selections: []int{0}}); err != nil {
			t.Fatalf("enemy discard failed: %v", err)
		}

		if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", CardIDs: []string{"ally-0"}}); err != nil {
			t.Fatalf("ally discard by card id failed: %v", err)
		}
		if len(p2.Hand) != 1 {
			t.Fatalf("expected ally to discard the first card, got %d cards", len(p2.Hand))
		}
		if p2.Hand[0].ID != "ally-1" {
			t.Fatalf("expected only second card to remain, got %+v", p2.Hand)
		}
		if got := engine.AttackDamageRuleBonusForModifier(p1, "ml_fullness_next_attack_bonus"); got != 2 {
			t.Fatalf("expected enemy discard plus first ally discard to grant 2 bonus, got %d", got)
		}
	})
}

// 魔枪-1：幻影星尘自伤摸牌爆牌时，弃牌结束后应继续结算（转正、清理待结算标记）。
func TestMagicLancerPhantomStardust_ContinuesAfterOverflowDiscard(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormMagicLancerPhantom
	// 手里不放雷系/法术牌，避免触发暗之障壁响应
	// 幻影形态手牌上限5：4张 + 自伤摸2 = 6，需弃1张
	p1.Hand = []model.Card{
		magicLancerTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire, 2),
		magicLancerTestCard("h2", "水流斩", model.CardTypeAttack, model.ElementWater, 2),
		magicLancerTestCard("h3", "风神斩", model.CardTypeAttack, model.ElementWind, 2),
		magicLancerTestCard("h4", "地裂斩", model.CardTypeAttack, model.ElementEarth, 2),
	}
	game.State.Deck = []model.Card{
		magicLancerTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		magicLancerTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		magicLancerTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWind, 2),
		magicLancerTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementThunder, 2),
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart
	if err := game.UseSkill("p1", "ml_phantom_stardust", []string{"p2"}, nil); err != nil {
		t.Fatalf("use ml_phantom_stardust failed: %v", err)
	}
	game.ProcessPendingDamages()
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected overflow discard interrupt, got %+v", game.State.PendingInterrupt)
	}
	if p1.TurnState.SkillFlowState["ml_stardust_wait_discard"] != 1 {
		t.Fatalf("expected stardust waiting for discard")
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})

	if got := p1.Form; got != "" {
		t.Fatalf("expected leave phantom form after discard resolves, got %q", got)
	}
	if got := p1.TurnState.SkillFlowState["ml_stardust_pending"]; got != 0 {
		t.Fatalf("expected ml_stardust_pending cleared after discard, got %d", got)
	}
	if got := p1.TurnState.SkillFlowState["ml_stardust_wait_discard"]; got != 0 {
		t.Fatalf("expected ml_stardust_wait_discard cleared after discard, got %d", got)
	}
}

func newDarkReleaseAttackGame(t *testing.T) (*engine.GameEngine, *model.Player, *model.Player) {
	t.Helper()
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormMagicLancerPhantom
	game.ApplyNextAttackDamageRule(p1.ID, "ml_dark_release_next_attack_bonus", "ml_dark_release", 1, model.RuleLifeUntilTurnEnd)
	p1.Hand = []model.Card{
		magicLancerTestCard("atk1", "地裂斩", model.CardTypeAttack, model.ElementEarth, 2),
		magicLancerTestCard("atk2", "地裂斩", model.CardTypeAttack, model.ElementEarth, 2),
	}
	p2.Heal = 0
	for i := 0; i < 10; i++ {
		game.State.Deck = append(game.State.Deck, magicLancerTestCard("d"+string(rune('a'+i)), "补牌", model.CardTypeAttack, model.ElementFire, 2))
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	return game, p1, p2
}

// 魔枪-3：第一次主动攻击命中，伤害 +1。
func TestMagicLancerDarkRelease_BonusConsumedAtFirstAttackDeclareHit(t *testing.T) {
	game, p1, p2 := newDarkReleaseAttackGame(t)
	p2.Hand = nil
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0)})
	if got := engine.AttackDamageRuleBonusForModifier(p1, "ml_dark_release_next_attack_bonus"); got != 0 {
		t.Fatalf("expected bonus consumed at attack declare, got %d", got)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"}})
	if got := len(p2.Hand); got != 3 {
		t.Fatalf("expected first attack damage 2+1=3, p2 hand=%d", got)
	}
}

// 魔枪-3：第一次主动攻击未命中（被圣光防御），+1 也随之作废，不留给之后的攻击。
func TestMagicLancerDarkRelease_BonusConsumedAtFirstAttackDeclareEvenIfMiss(t *testing.T) {
	game, p1, p2 := newDarkReleaseAttackGame(t)
	p2.Hand = []model.Card{{ID: "hl", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight}}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0)})
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"defend"}, CardID: testutils.PlayableCardID(t, game, "p2", 0)})
	if got := engine.AttackDamageRuleBonusForModifier(p1, "ml_dark_release_next_attack_bonus"); got != 0 {
		t.Fatalf("expected bonus gone after missed first attack, got %d", got)
	}
	if got := p1.TurnState.SkillFlowState["ml_dark_release_armed_bonus"]; got != 0 {
		t.Fatalf("expected armed bonus not carried over after miss, got %d", got)
	}
	// 直接计算第二次主动攻击的伤害：不应再 +1。
	card := magicLancerTestCard("atk2", "地裂斩", model.CardTypeAttack, model.ElementEarth, 2)
	if got := game.ApplyPassiveAttackEffects(p1, p2, 2, model.Action{SourceID: p1.ID, TargetID: p2.ID, Type: model.ActionAttack, Card: &card}); got != 2 {
		t.Fatalf("expected later attack without dark release bonus, got %d", got)
	}
}

// 魔枪-4/5：充盈选代价牌时可以取消，退回行动标记，手牌不变，回到行动选择。
func TestMagicLancerFullness_CancelAtCostCardRefunds(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{magicLancerTestCard("t1", "雷斩", model.CardTypeAttack, model.ElementThunder, 2)}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "ml_fullness"})
	testutils.RequireChoicePrompt(t, game, "p1", "ml_fullness_cost_card")
	if prompt := game.GetCurrentPrompt(); prompt == nil || prompt.Presentation == nil || prompt.Presentation.CancelPolicy != model.CancelPolicyAbort {
		t.Fatalf("expected cancelable cost-card prompt, got %+v", prompt)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected flow aborted, got %+v", game.State.PendingInterrupt)
	}
	if len(p1.Hand) != 1 {
		t.Fatalf("expected hand untouched, got %d", len(p1.Hand))
	}
	if p1.TurnState.HasActed || p1.TurnState.UsedSkillCounts["ml_fullness"] != 0 {
		t.Fatalf("expected action marks rolled back, got %+v", p1.TurnState)
	}
	if !game.IsActionSelectionWindow() {
		t.Fatalf("expected action selection window, stage=%s", game.State.TurnStage)
	}
}

// 魔枪-4/5：漆黑之枪选 X 时可以取消，不扣水晶，攻击伤害照常结算。
func TestMagicLancerBlackSpear_CancelAtXKeepsCrystal(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormMagicLancerPhantom
	p1.Crystal = 2
	p2.Hand = []model.Card{magicLancerTestCard("h1", "火斩", model.CardTypeAttack, model.ElementFire, 2)}
	game.State.PendingDamageQueue = []model.PendingDamage{{SourceID: p1.ID, TargetID: p2.ID, Damage: 2, DamageType: model.AttackDamage}}
	ctx := game.BuildContext(p1, p2, model.TimingAttackHit, &model.EventContext{
		Type: model.EventAttack, SourceID: p1.ID, TargetID: p2.ID,
		AttackInfo: &model.AttackEventInfo{ActionType: "Attack", IsHit: true},
	})
	if err := skills.GetHandler("ml_black_spear").Execute(ctx); err != nil {
		t.Fatalf("execute black spear failed: %v", err)
	}
	if prompt := game.GetCurrentPrompt(); prompt == nil || prompt.Presentation == nil || prompt.Presentation.CancelPolicy != model.CancelPolicyAbort {
		t.Fatalf("expected cancelable X prompt, got %+v", prompt)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdCancel, PlayerID: "p1"}); err != nil {
		t.Fatalf("cancel black spear failed: %v", err)
	}
	if testutils.ChoiceTypeOfInterrupt(game.State.PendingInterrupt) == "ml_black_spear_x" {
		t.Fatalf("expected X prompt closed after cancel")
	}
	if p1.Crystal != 2 {
		t.Fatalf("expected crystal untouched, got %d", p1.Crystal)
	}
	for _, pd := range game.State.PendingDamageQueue {
		if pd.SourceID == p1.ID && pd.Damage != 2 {
			t.Fatalf("expected attack damage unchanged, got %d", pd.Damage)
		}
	}
}

// 魔枪-7：充盈不再带"其他队友"目标规则，不指定目标也能发动。
func TestMagicLancerFullness_NoLeftoverTargetRule(t *testing.T) {
	for _, entry := range magic_lancer.SkillEntries() {
		if entry.ID == "ml_fullness" && (entry.Policy.TargetRules.Count.Max != 0 || len(entry.Policy.TargetRules.Slots) != 0) {
			t.Fatalf("expected no target rule on ml_fullness, got %+v", entry.Policy.TargetRules)
		}
	}
}
