package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
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

func TestMagicSwordsmanYellowSpring_NoCounterKeepsOriginalElementAndConsumesGem(t *testing.T) {
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
	g.State.TurnStage = model.TurnStageActionExecution

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
	if top.Card == nil || top.Card.Element != model.ElementFire {
		t.Fatalf("expected attack element remain fire, got %+v", top.Card)
	}
	if top.CanBeResponded {
		t.Fatalf("expected attack cannot be countered after yellow spring")
	}
	if p1.Gem != 0 {
		t.Fatalf("expected yellow spring consume 1 gem, got %d", p1.Gem)
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

func TestCrimsonBloodBarrier_AutoDamagesSourceWithoutPrompt(t *testing.T) {
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
	ctx := g.buildContext(p1, p1, model.TimingOnDamageTaken, &model.EventContext{
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

	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected no extra prompt for css blood barrier, got %+v", g.State.PendingInterrupt)
	}
	if damage != 1 {
		t.Fatalf("expected damage reduced to 1, got %d", damage)
	}
	if len(g.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected 1 reflected damage queued, got %+v", g.State.PendingDamageQueue)
	}
	if g.State.PendingDamageQueue[0].TargetID != "p2" {
		t.Fatalf("expected reflected damage target source p2, got %+v", g.State.PendingDamageQueue[0])
	}
}

func TestCrimsonBloodBarrier_DoesNotRetargetOtherEnemy(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "OtherEnemy", "berserker", model.BlueCamp); err != nil {
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
	ctx := g.buildContext(p1, p1, model.TimingOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  "p2",
		TargetID:  "p1",
		DamageVal: &damage,
	})
	ctx.Flags["IsMagicDamage"] = true
	if err := handler.Execute(ctx); err != nil {
		t.Fatalf("execute css_blood_barrier failed: %v", err)
	}
	if len(g.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected exactly 1 reflected damage, got %+v", g.State.PendingDamageQueue)
	}
	if g.State.PendingDamageQueue[0].TargetID != "p2" {
		t.Fatalf("expected reflected damage locked to original source p2, got %+v", g.State.PendingDamageQueue[0])
	}
}

func TestPrayerManaTide_RunsAfterMagicActionEnd(t *testing.T) {
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
	g.State.TurnStage = model.TurnStageExtraAction

	g.Drive()

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response interrupt for prayer_mana_tide, got %+v", g.State.PendingInterrupt)
	}
	if !containsSkillIDBugfix(g.State.PendingInterrupt.SkillIDs, "prayer_mana_tide") {
		t.Fatalf("expected prayer_mana_tide in interrupt skill ids, got %+v", g.State.PendingInterrupt.SkillIDs)
	}
}

func TestPrayerSwiftBlessing_StillRunsAfterPhaseEndInterrupt(t *testing.T) {
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
		Hook: model.FieldHookManual,
	})
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageExtraAction

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
	if ct := pendingChoiceType(g.State.PendingInterrupt); ct != "prayer_swift_blessing_followup" {
		t.Fatalf("expected prayer_swift_blessing_followup, got %q", ct)
	}
}

func TestPrayerSwiftBlessing_AttackFollowupSurvivesPhaseEndResponseInterrupt(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "Blade", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Deck = rules.InitDeck()
	g.State.TurnStage = model.TurnStageActionExecution

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.Crystal = 1
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p2.TurnState = model.NewPlayerTurnState()
	p1.AddFieldCard(&model.FieldCard{
		OwnerID: p1.ID,
		Mode:    model.FieldEffect,
		Effect:  model.EffectSwiftBlessing,
		Hook: model.FieldHookManual,
	})
	p1.Hand = []model.Card{{ID: "atk1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1}}

	if err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	if err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected first attack-end response interrupt, got %+v", g.State.PendingInterrupt)
	}
	if !containsSkillIDBugfix(g.State.PendingInterrupt.SkillIDs, "sword_shadow") {
		t.Fatalf("expected sword_shadow first, got %+v", g.State.PendingInterrupt.SkillIDs)
	}

	if err := g.SkipResponse(); err != nil {
		t.Fatalf("skip sword_shadow failed: %v", err)
	}
	g.Drive()

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected swift blessing choice after attack-end response, got %+v", g.State.PendingInterrupt)
	}
	if ct := pendingChoiceType(g.State.PendingInterrupt); ct != "prayer_swift_blessing_followup" {
		t.Fatalf("expected prayer_swift_blessing_followup after sword_shadow, got %q", ct)
	}

	if err := g.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm swift blessing failed: %v", err)
	}

	if p1.HasFieldEffect(model.EffectSwiftBlessing) {
		t.Fatalf("expected swift blessing to be consumed")
	}
	if g.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected extra attack action window after swift blessing, got %s", g.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != string(model.ActionAttack) {
		t.Fatalf("expected current extra action Attack, got %q", p1.TurnState.CurrentExtraAction)
	}
	if len(g.State.DeferredFollowups) != 0 {
		t.Fatalf("expected deferred followups drained, got %+v", g.State.DeferredFollowups)
	}
}
