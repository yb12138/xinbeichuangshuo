package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

func pendingChoiceType(intr *model.Interrupt) string {
	if intr == nil {
		return ""
	}
	ctx, _ := intr.Context.(map[string]interface{})
	v, _ := ctx["choice_type"].(string)
	return v
}

func containsSkillIDBugfix(list []string, id string) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}

func promptHasOption(prompt *model.Prompt, id string) bool {
	if prompt == nil {
		return false
	}
	for _, opt := range prompt.Options {
		if opt.ID == id {
			return true
		}
	}
	return false
}

func TestMagicSwordsmanYellowSpring_ForcesDarkAndNoCounter(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "MS", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Def", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p3 := g.State.Players["p3"]
	p1.IsActive = true
	p2.IsActive = false
	p3.IsActive = false
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "counter_dark", Name: "暗灭", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 1},
	}
	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseActionSelection

	if err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		CardIndex: 0,
		TargetID:  "p2",
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response interrupt for yellow spring, got %+v", g.State.PendingInterrupt)
	}
	if !containsSkillIDBugfix(g.State.PendingInterrupt.SkillIDs, "ms_yellow_spring") {
		t.Fatalf("expected ms_yellow_spring in interrupt skill ids, got %+v", g.State.PendingInterrupt.SkillIDs)
	}

	if err := g.ConfirmResponseSkill("p1", "ms_yellow_spring"); err != nil {
		t.Fatalf("confirm yellow spring failed: %v", err)
	}
	g.Drive()

	if len(g.State.CombatStack) == 0 {
		t.Fatalf("expected combat stack after confirming yellow spring")
	}
	top := g.State.CombatStack[len(g.State.CombatStack)-1]
	if top.Card == nil || top.Card.Element != model.ElementDark {
		t.Fatalf("expected attack element forced to dark, got %+v", top.Card)
	}
	if top.CanBeResponded {
		t.Fatalf("expected attack cannot be countered after yellow spring")
	}

	// 兜底校验：即使防守方手里有可应战牌，也应被禁止应战。
	err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		TargetID:  "p3",
		ExtraArgs: []string{"counter"},
	})
	if err == nil || !strings.Contains(err.Error(), "无法被应战") {
		t.Fatalf("expected counter denied after yellow spring, got err=%v", err)
	}
}

func TestCrimsonBloodBarrierPrompt_CanCancel(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 1

	handler := skills.GetHandler("css_blood_barrier")
	if handler == nil {
		t.Fatalf("css_blood_barrier handler not found")
	}
	damage := 2
	ctx := g.buildContext(p1, p1, model.TriggerOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  "p2",
		TargetID:  "p1",
		DamageVal: &damage,
	})
	ctx.Flags["IsMagicDamage"] = true
	if !handler.CanUse(ctx) {
		t.Fatalf("expected css_blood_barrier can use")
	}
	if err := handler.Execute(ctx); err != nil {
		t.Fatalf("execute css_blood_barrier failed: %v", err)
	}

	if ct := pendingChoiceType(g.State.PendingInterrupt); ct != "css_blood_barrier_counter_confirm" {
		t.Fatalf("expected css_blood_barrier_counter_confirm, got %q", ct)
	}
	prompt := g.buildChoicePrompt()
	if !promptHasOption(prompt, "cancel") {
		t.Fatalf("expected cancel option in css blood barrier confirm prompt, got %+v", prompt)
	}

	if err := g.handleInterruptAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	}); err != nil {
		t.Fatalf("cancel css blood barrier confirm failed: %v", err)
	}
	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected interrupt cleared after cancel, got %+v", g.State.PendingInterrupt)
	}
	if len(g.State.PendingDamageQueue) != 0 {
		t.Fatalf("cancel should not enqueue extra damage, got %+v", g.State.PendingDamageQueue)
	}
}

func TestCrimsonBloodBarrierTargetPrompt_CanCancel(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 1

	handler := skills.GetHandler("css_blood_barrier")
	if handler == nil {
		t.Fatalf("css_blood_barrier handler not found")
	}
	damage := 2
	ctx := g.buildContext(p1, p1, model.TriggerOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  "p2",
		TargetID:  "p1",
		DamageVal: &damage,
	})
	ctx.Flags["IsMagicDamage"] = true
	if err := handler.Execute(ctx); err != nil {
		t.Fatalf("execute css_blood_barrier failed: %v", err)
	}

	if err := g.handleInterruptAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 是 -> 进入选目标
	}); err != nil {
		t.Fatalf("confirm css blood barrier extra damage failed: %v", err)
	}

	if ct := pendingChoiceType(g.State.PendingInterrupt); ct != "css_blood_barrier_target" {
		t.Fatalf("expected css_blood_barrier_target, got %q", ct)
	}
	prompt := g.buildChoicePrompt()
	if !promptHasOption(prompt, "cancel") {
		t.Fatalf("expected cancel option in css blood barrier target prompt, got %+v", prompt)
	}

	if err := g.handleInterruptAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	}); err != nil {
		t.Fatalf("cancel css blood barrier target failed: %v", err)
	}
	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected interrupt cleared after cancel, got %+v", g.State.PendingInterrupt)
	}
	if len(g.State.PendingDamageQueue) != 0 {
		t.Fatalf("cancel target selection should not enqueue extra damage, got %+v", g.State.PendingDamageQueue)
	}
}

func TestPrayerManaTide_TriggersAfterMagicActionEnd(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.Crystal = 1
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.LastActionType = string(model.ActionMagic)
	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseExtraAction

	g.Drive()

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response interrupt for prayer_mana_tide, got %+v", g.State.PendingInterrupt)
	}
	if !containsSkillIDBugfix(g.State.PendingInterrupt.SkillIDs, "prayer_mana_tide") {
		t.Fatalf("expected prayer_mana_tide in interrupt skill ids, got %+v", g.State.PendingInterrupt.SkillIDs)
	}
}

func TestPrayerSwiftBlessing_StillTriggersAfterPhaseEndInterrupt(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.Crystal = 1 // 让法力潮汐先触发一个 OnPhaseEnd 中断
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.LastActionType = string(model.ActionMagic)
	p1.AddFieldCard(&model.FieldCard{
		OwnerID: p1.ID,
		Mode:    model.FieldEffect,
		Effect:  model.EffectSwiftBlessing,
		Trigger: model.EffectTriggerManual,
	})
	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseExtraAction

	g.Drive()
	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected first phase-end interrupt (mana tide), got %+v", g.State.PendingInterrupt)
	}
	if !containsSkillIDBugfix(g.State.PendingInterrupt.SkillIDs, "prayer_mana_tide") {
		t.Fatalf("expected prayer_mana_tide first, got %+v", g.State.PendingInterrupt.SkillIDs)
	}

	// 跳过法力潮汐后，仍应继续弹出迅捷赐福触发询问（不能被吞掉）。
	if err := g.SkipResponse(); err != nil {
		t.Fatalf("skip mana tide failed: %v", err)
	}
	g.Drive()

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected swift blessing choice interrupt after mana tide, got %+v", g.State.PendingInterrupt)
	}
	if ct := pendingChoiceType(g.State.PendingInterrupt); ct != "prayer_swift_blessing_trigger" {
		t.Fatalf("expected prayer_swift_blessing_trigger, got %q", ct)
	}
}
