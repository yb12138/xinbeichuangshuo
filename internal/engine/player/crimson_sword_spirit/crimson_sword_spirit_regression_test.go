package crimson_sword_spirit_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
	"testing"
)

func hasFieldEffect(player *model.Player, effect model.EffectType) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == effect {
			return true
		}
	}
	return false
}

func TestRoseCourtyard_DisablesHealResistForAllPlayers(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Attacker", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Target", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p3 := g.State.Players["p3"]
	p1.Field = append(p1.Field, &model.FieldCard{
		Card: model.Card{
			ID:      "rose-courtyard",
			Name:    "血蔷薇庭院",
			Type:    model.CardTypeMagic,
			Element: model.ElementDark,
		},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectRoseCourtyard,
	})
	p3.Heal = 1
	p3.Hand = nil
	g.State.Deck = []model.Card{
		{ID: "d1", Name: "补1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "d2", Name: "补2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
	}

	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p3",
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	g.Drive()

	if intr := g.State.PendingInterrupt; intr != nil {
		t.Fatalf("rose courtyard should prevent heal-resist prompt for all players, got %+v", intr)
	}
	if got := p3.Heal; got != 1 {
		t.Fatalf("expected target heal not spent, got %d", got)
	}
	if got := len(p3.Hand); got != 2 {
		t.Fatalf("expected full damage draw because heal is disabled, got hand=%d", got)
	}
}

func TestRoseCourtyard_ReturnsOnlyOnFinalTurnEndAfterExtraActions(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood_cap"] = 4
	p1.Tokens["css_blood"] = 4
	p1.Field = append(p1.Field, &model.FieldCard{
		Card: model.Card{
			ID:              "starter-p1-css_rose_courtyard",
			Name:            "血蔷薇庭院",
			Type:            model.CardTypeMagic,
			Element:         model.ElementDark,
			ExclusiveChar1:  "crimson_sword_spirit",
			ExclusiveSkill1: "血蔷薇庭院",
		},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectRoseCourtyard,
	})
	model.AppendAttackAction(p1, "赤色一闪")

	if paused := g.RunTurnEndTimingStageHooks(p1, engine.TimingTurnEndPreExtra); paused {
		t.Fatalf("unexpected interrupt during pre-extra turn end")
	}
	if !hasFieldEffect(p1, model.EffectRoseCourtyard) {
		t.Fatalf("rose courtyard should stay on board before pending extra actions are consumed")
	}
	if p1.HasExclusiveCard(p1.Character.ID, "血蔷薇庭院") {
		t.Fatalf("rose courtyard should not return to exclusive zone before final turn end")
	}
	if p1.Tokens["css_blood_cap"] != 4 || p1.Tokens["css_blood"] != 4 {
		t.Fatalf("rose courtyard blood cap should stay active before final turn end, cap=%d blood=%d", p1.Tokens["css_blood_cap"], p1.Tokens["css_blood"])
	}
	if len(p1.TurnState.PendingActions) != 1 || p1.TurnState.PendingActions[0].Source != "赤色一闪" {
		t.Fatalf("expected pending crimson flash extra attack to remain, got %+v", p1.TurnState.PendingActions)
	}

	if paused := g.RunTurnEndTimingStageHooks(p1, engine.TimingTurnEndFinal); paused {
		t.Fatalf("unexpected interrupt during final turn end")
	}
	if hasFieldEffect(p1, model.EffectRoseCourtyard) {
		t.Fatalf("expected rose courtyard field card removed at final turn end")
	}
	if !p1.HasExclusiveCard(p1.Character.ID, "血蔷薇庭院") {
		t.Fatalf("expected rose courtyard returned to exclusive zone at final turn end")
	}
	if p1.Tokens["css_blood_cap"] != 3 {
		t.Fatalf("expected blood cap reset to 3 at final turn end, got %d", p1.Tokens["css_blood_cap"])
	}
	if p1.Tokens["css_blood"] != 3 {
		t.Fatalf("expected blood trimmed to 3 at final turn end, got %d", p1.Tokens["css_blood"])
	}
}

func TestRoseCourtyard_DriveKeepsActiveWhenTurnEndStartsCrimsonFlashExtraAction(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Field = append(p1.Field, &model.FieldCard{
		Card: model.Card{
			ID:              "starter-p1-css_rose_courtyard",
			Name:            "血蔷薇庭院",
			Type:            model.CardTypeMagic,
			Element:         model.ElementDark,
			ExclusiveChar1:  "crimson_sword_spirit",
			ExclusiveSkill1: "血蔷薇庭院",
		},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectRoseCourtyard,
	})
	model.AppendAttackAction(p1, "赤色一闪")

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageTurnEnd
	g.Drive()

	if g.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected drive to enter crimson flash extra action, got stage=%s", g.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != string(model.ActionAttack) {
		t.Fatalf("expected current extra action Attack, got %q", p1.TurnState.CurrentExtraAction)
	}
	if len(p1.TurnState.PendingActions) != 0 {
		t.Fatalf("expected pending extra action consumed into current action, got %+v", p1.TurnState.PendingActions)
	}
	if !hasFieldEffect(p1, model.EffectRoseCourtyard) {
		t.Fatalf("rose courtyard should stay on board while crimson flash extra action starts")
	}
	if p1.HasExclusiveCard(p1.Character.ID, "血蔷薇庭院") {
		t.Fatalf("rose courtyard should not return to exclusive zone before the final turn end")
	}
}

func TestCrimsonFlash_PhaseEndDamageShouldNotStall(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 1
	p1.Heal = 0
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageExtraAction
	g.State.Deck = []model.Card{
		{ID: "d1", Name: "补1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "d2", Name: "补2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		{ID: "d3", Name: "补3", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
	}

	eventCtx := &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   p1.ID,
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{ActionType: string(model.ActionAttack), CounterInitiator: ""},
	}
	ctx := g.BuildContext(p1, nil, model.TimingActionEnd, eventCtx)
	g.Dispatcher().OnTiming(ctx.Timing, ctx)

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt for css flash, got %+v", g.State.PendingInterrupt)
	}
	if err := g.ConfirmResponseSkill("p1", "css_crimson_flash"); err != nil {
		t.Fatalf("confirm response failed: %v", err)
	}

	if len(g.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected 1 pending damage from crimson flash, got %d", len(g.State.PendingDamageQueue))
	}
	if g.State.PendingDamageQueue[0].Damage != 2 || g.State.PendingDamageQueue[0].TargetID != "p1" {
		t.Fatalf("unexpected pending damage: %+v", g.State.PendingDamageQueue[0])
	}

	// Drive should resolve pending damage and should not return to response phase (would stall).
	g.Drive()

	if g.State.Subflow == model.SubflowResponse {
		t.Fatalf("flow stuck in response after crimson flash: return_subflow=%s", g.State.ReturnSubflow)
	}
	if len(g.State.PendingDamageQueue) != 0 {
		t.Fatalf("expected pending damage drained, got %d", len(g.State.PendingDamageQueue))
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected exactly 2 self-damage draw from crimson flash, got hand=%d", got)
	}
}

func TestCrimsonFlash_SpendsOnlyBloodBeforeSelfDamageResponses(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 1
	p1.Heal = 0
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageExtraAction
	g.State.Deck = []model.Card{
		{ID: "d1", Name: "补1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "d2", Name: "补2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
	}

	eventCtx := &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   p1.ID,
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{ActionType: string(model.ActionAttack), CounterInitiator: ""},
	}
	ctx := g.BuildContext(p1, nil, model.TimingActionEnd, eventCtx)
	g.Dispatcher().OnTiming(ctx.Timing, ctx)

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt for css flash, got %+v", g.State.PendingInterrupt)
	}
	if err := g.ConfirmResponseSkill("p1", "css_crimson_flash"); err != nil {
		t.Fatalf("confirm response failed: %v", err)
	}
	if got := p1.Tokens["css_blood"]; got != 0 {
		t.Fatalf("expected crimson flash to spend the only blood immediately, got %d", got)
	}

	g.Drive()

	if intr := g.State.PendingInterrupt; intr != nil {
		t.Fatalf("expected no blood barrier prompt after crimson flash spent the only blood, got %+v", intr)
	}
	if got := p1.Tokens["css_blood"]; got != 0 {
		t.Fatalf("expected blood to remain 0 after self damage, got %d", got)
	}
}

func TestCrimsonFlash_SelfDamageCannotTriggerBloodBarrier(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 2
	p1.Heal = 0
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageExtraAction
	g.State.Deck = []model.Card{
		{ID: "d1", Name: "补1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "d2", Name: "补2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
	}

	eventCtx := &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   p1.ID,
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{ActionType: string(model.ActionAttack), CounterInitiator: ""},
	}
	ctx := g.BuildContext(p1, nil, model.TimingActionEnd, eventCtx)
	g.Dispatcher().OnTiming(ctx.Timing, ctx)

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt for css flash, got %+v", g.State.PendingInterrupt)
	}
	if err := g.ConfirmResponseSkill("p1", "css_crimson_flash"); err != nil {
		t.Fatalf("confirm response failed: %v", err)
	}
	if got := p1.Tokens["css_blood"]; got != 1 {
		t.Fatalf("expected crimson flash to spend one blood, got %d", got)
	}

	g.Drive()

	if intr := g.State.PendingInterrupt; intr != nil {
		t.Fatalf("expected no blood barrier prompt for crimson flash self damage, got %+v", intr)
	}
	if got := p1.Tokens["css_blood"]; got != 1 {
		t.Fatalf("expected remaining blood not spent by blood barrier, got %d", got)
	}
}

func TestCrimsonFlash_CombatFlow_DealsExactlyTwoAndKeepsTurnProgressing(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p2.Heal = 0
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution
	g.State.Deck = []model.Card{
		{ID: "d1", Name: "补1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "d2", Name: "补2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		{ID: "d3", Name: "补3", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		{ID: "d4", Name: "补4", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
		{ID: "d5", Name: "补5", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
	}

	if err := g.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, g, "p1", 0),
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	if err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("target take failed: %v", err)
	}

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response skill prompt after attack end, got %+v", g.State.PendingInterrupt)
	}
	if err := g.ConfirmResponseSkill("p1", "css_crimson_flash"); err != nil {
		t.Fatalf("confirm crimson flash failed: %v", err)
	}

	// 触发后应仅追加一次“对自己2点法术伤害”。
	if len(g.State.PendingDamageQueue) != 1 || g.State.PendingDamageQueue[0].Damage != 2 || g.State.PendingDamageQueue[0].TargetID != "p1" {
		t.Fatalf("unexpected pending damages after crimson flash: %+v", g.State.PendingDamageQueue)
	}

	g.Drive()

	if g.State.Subflow == model.SubflowResponse {
		t.Fatalf("flow stuck in response after crimson flash in combat flow")
	}
	if len(g.State.PendingDamageQueue) != 0 {
		t.Fatalf("expected pending damage drained, got %d", len(g.State.PendingDamageQueue))
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected CSS hand=2 after self damage draw, got %d", got)
	}
}
