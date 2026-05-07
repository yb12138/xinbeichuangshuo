package angel_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
)

func TestAngelSong_RunsAsTurnStartResponseAndResumesActionSelection(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p2.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "weak-1", Name: "虚弱", Type: model.CardTypeMagic, Element: model.ElementWind},
		OwnerID:  p2.ID,
		SourceID: "p3",
		Mode:     model.FieldEffect,
		Effect:   model.EffectWeak,
		Hook:     model.FieldHookOnBeforeAction,
	})

	game.Drive()

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected angel_song response interrupt at startup, got %+v", game.State.PendingInterrupt)
	}
	if !testutils.InterruptHasSkillID(game.State.PendingInterrupt, "angel_song") {
		t.Fatalf("expected angel_song in pending skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm angel_song failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected basic effect choice after angel_song, got %+v", game.State.PendingInterrupt)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctxData["choice_type"].(string); got != "basic_effect_pick" {
		t.Fatalf("expected basic_effect_pick, got %q", got)
	}
	if p1.Crystal != 0 {
		t.Fatalf("angel_song should consume 1 crystal before selection, got %d", p1.Crystal)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("resolve angel_song basic effect pick failed: %v", err)
	}
	if got := testutils.CountFieldEffect(p2, model.EffectWeak); got != 0 {
		t.Fatalf("expected weakness removed by angel_song, got %d", got)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected angel bond follow-up choice after angel_song removal, got %+v", game.State.PendingInterrupt)
	}
	ctxData, _ = game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctxData["choice_type"].(string); got != "angel_bond_heal_target" {
		t.Fatalf("expected angel_bond_heal_target, got %q", got)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("resolve angel bond after angel_song failed: %v", err)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after angel bond resolution, got %+v", game.State.PendingInterrupt)
	}
	if game.State.TurnStage != model.TurnStageActionExecution || game.State.CombatStage != model.CombatStageNone || game.State.Subflow != model.SubflowNone {
		t.Fatalf("expected turn to continue to action execution window, got turn=%s combat=%s subflow=%s", game.State.TurnStage, game.State.CombatStage, game.State.Subflow)
	}
}

func TestGodProtection_PromptsForXAndPartiallyMitigatesMoraleLoss(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.RedMorale = 10

	p1 := game.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 3

	moraleLoss := 3
	lossCtx := game.BuildContext(p1, nil, model.TimingBeforeMoraleLoss, &model.EventContext{
		Type:      model.EventDamage,
		DamageVal: &moraleLoss,
	})
	lossCtx.Flags["IsMagicDamage"] = true
	lossCtx.Selections = map[string]any{
		"morale_loss_pending":              true,
		"morale_loss_value":                3,
		"is_magic":                         true,
		"from_damage_draw":                 false,
		"overflow_morale_loss_fixed":       0,
		"discarded_cards":                  []model.Card{},
		"victim_id":                        p1.ID,
		"discard_player_id":                p1.ID,
		"morale_loss_stay_in_turn":         false,
		"morale_loss_is_damage_resolution": false,
	}

	game.Dispatcher().OnTiming(lossCtx.Timing, lossCtx)
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected god_protection response interrupt, got %+v", game.State.PendingInterrupt)
	}
	if !testutils.InterruptHasSkillID(game.State.PendingInterrupt, "god_protection") {
		t.Fatalf("expected god_protection in pending skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm god_protection failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected X choice after god_protection confirmation, got %+v", game.State.PendingInterrupt)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctxData["choice_type"].(string); got != "god_protection_x" {
		t.Fatalf("expected god_protection_x, got %q", got)
	}
	if moraleLoss != 3 || p1.Crystal != 3 {
		t.Fatalf("god_protection should wait for X selection before consuming resources, moraleLoss=%d crystal=%d", moraleLoss, p1.Crystal)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("resolve god_protection X failed: %v", err)
	}
	if moraleLoss != 1 {
		t.Fatalf("expected morale loss reduced to 1 after choosing X=2, got %d", moraleLoss)
	}
	if p1.Crystal != 1 {
		t.Fatalf("expected 2 crystal consumed, got crystal=%d", p1.Crystal)
	}
	if game.State.RedMorale != 9 {
		t.Fatalf("expected red morale 10 -> 9 after partial mitigation, got %d", game.State.RedMorale)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after god_protection resolution, got %+v", game.State.PendingInterrupt)
	}
}

func TestGodProtection_TriggersWithGemFallbackWhenCrystalInsufficient(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.RedMorale = 10

	p1 := game.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 0
	p1.Gem = 2

	moraleLoss := 2
	lossCtx := game.BuildContext(p1, nil, model.TimingBeforeMoraleLoss, &model.EventContext{
		Type:      model.EventDamage,
		DamageVal: &moraleLoss,
	})
	lossCtx.Flags["IsMagicDamage"] = true
	lossCtx.Selections = map[string]any{
		"morale_loss_pending":              true,
		"morale_loss_value":                2,
		"is_magic":                         true,
		"from_damage_draw":                 false,
		"overflow_morale_loss_fixed":       0,
		"discarded_cards":                  []model.Card{},
		"victim_id":                        p1.ID,
		"discard_player_id":                p1.ID,
		"morale_loss_stay_in_turn":         false,
		"morale_loss_is_damage_resolution": false,
	}

	game.Dispatcher().OnTiming(lossCtx.Timing, lossCtx)
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected god_protection response interrupt, got %+v", game.State.PendingInterrupt)
	}
	if !testutils.InterruptHasSkillID(game.State.PendingInterrupt, "god_protection") {
		t.Fatalf("expected god_protection in pending skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm god_protection failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected X choice after god_protection confirmation, got %+v", game.State.PendingInterrupt)
	}

	// 选择 X=2（option index 1），应消耗2红宝石并完全抵御本次士气下降。
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("resolve god_protection X failed: %v", err)
	}
	if moraleLoss != 0 {
		t.Fatalf("expected morale loss reduced to 0 after choosing X=2, got %d", moraleLoss)
	}
	if p1.Gem != 0 || p1.Crystal != 0 {
		t.Fatalf("expected gem fallback consumed as crystal-like cost, gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
	if game.State.RedMorale != 10 {
		t.Fatalf("expected red morale unchanged after full mitigation, got %d", game.State.RedMorale)
	}
}

func TestGodProtection_DoesNotTriggerOnNonMagicMoraleLoss(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.RedMorale = 10

	p1 := game.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 0
	p1.Gem = 2

	moraleLoss := 2
	lossCtx := game.BuildContext(p1, nil, model.TimingBeforeMoraleLoss, &model.EventContext{
		Type:      model.EventDamage,
		DamageVal: &moraleLoss,
	})
	// 非法术来源：即便有可替代水晶的红宝石，也不应触发神之庇护。
	lossCtx.Flags["IsMagicDamage"] = false
	lossCtx.Selections = map[string]any{
		"morale_loss_pending":              true,
		"morale_loss_value":                2,
		"is_magic":                         false,
		"from_damage_draw":                 true,
		"overflow_morale_loss_fixed":       0,
		"discarded_cards":                  []model.Card{},
		"victim_id":                        p1.ID,
		"discard_player_id":                p1.ID,
		"morale_loss_stay_in_turn":         false,
		"morale_loss_is_damage_resolution": false,
	}

	game.Dispatcher().OnTiming(lossCtx.Timing, lossCtx)
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no god_protection prompt on non-magic morale loss, got %+v", game.State.PendingInterrupt)
	}
	if p1.Gem != 2 || p1.Crystal != 0 {
		t.Fatalf("expected no crystal-like cost consumption on non-magic morale-loss timing, gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
}

func TestAngelBond_AfterShieldDoesNotReopenActionSelection(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p2.IsActive = false
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "shield-1", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "weak-1", Name: "虚弱", Type: model.CardTypeMagic, Element: model.ElementWind},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("cast shield should succeed, got err=%v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected angel bond choice interrupt after shield, got %+v", game.State.PendingInterrupt)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctxData["choice_type"].(string); got != "angel_bond_heal_target" {
		t.Fatalf("expected angel_bond_heal_target after shield, got %q", got)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("resolve angel bond choice failed: %v", err)
	}

	if got := p1.TurnState.LastActionType; got != "" {
		t.Fatalf("expected action-end catchup consumed last action, got %q", got)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p2",
		CardIndex: 0,
	}); err == nil {
		t.Fatalf("expected p1 cannot cast another magic immediately after shield+bond resolution")
	}
}

func TestAngelBlessing_ReceiverOverHandLimitTriggersOverflowDiscard(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p2.IsActive = false
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	// 先把手牌压到上限：发动技能前 6 张，弃置 1 张后为 5；再收 2 张会超上限 1 张。
	p1.Hand = []model.Card{
		{ID: "w-1", Name: "水牌", Type: model.CardTypeMagic, Element: model.ElementWater},
		{ID: "f-1", Name: "火牌", Type: model.CardTypeMagic, Element: model.ElementFire},
		{ID: "e-1", Name: "地牌", Type: model.CardTypeMagic, Element: model.ElementEarth},
		{ID: "wi-1", Name: "风牌", Type: model.CardTypeMagic, Element: model.ElementWind},
		{ID: "l-1", Name: "光牌", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "d-1", Name: "暗牌", Type: model.CardTypeMagic, Element: model.ElementDark},
	}
	p2.Hand = []model.Card{
		{ID: "g-1", Name: "给牌1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "g-2", Name: "给牌2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "angel_blessing",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("angel_blessing should succeed, got err=%v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptGiveCards {
		t.Fatalf("expected give-cards interrupt, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	}); err != nil {
		t.Fatalf("giver should be able to give 2 cards, got err=%v", err)
	}

	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected overflow discard interrupt for receiver, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected overflow discard owned by p1, got %s", game.State.PendingInterrupt.PlayerID)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctxData["discard_count"].(int); got != 1 {
		t.Fatalf("expected discard_count=1 after angel_blessing overflow, got %v", ctxData["discard_count"])
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("resolve overflow discard failed: %v", err)
	}

	maxHand := game.GetMaxHand(p1)
	if got := len(p1.Hand); got != maxHand {
		t.Fatalf("expected hand size restored to max hand=%d after overflow discard, got %d", maxHand, got)
	}
}
