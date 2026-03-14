package engine

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"testing"
)

type promptCaptureObserver struct {
	lastPrompt *model.Prompt
}

func (o *promptCaptureObserver) OnGameEvent(event model.GameEvent) {
	if event.Type != model.EventAskInput {
		return
	}
	if p, ok := event.Data.(*model.Prompt); ok {
		o.lastPrompt = p
	}
}

func TestPlagueMageCannotUseHealAgainstAttackDamage(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "A", "angel", model.RedCamp); err != nil {
		t.Fatalf("add player p1 failed: %v", err)
	}
	if err := g.AddPlayer("p2", "B", "plague_mage", model.BlueCamp); err != nil {
		t.Fatalf("add player p2 failed: %v", err)
	}

	victim := g.State.Players["p2"]
	victim.Heal = 2

	g.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   "p1",
			TargetID:   "p2",
			Damage:     1,
			DamageType: "Attack",
			Stage:      0,
		},
	}
	g.State.Phase = model.PhasePendingDamageResolution
	g.State.ReturnPhase = model.PhaseTurnEnd

	g.Drive()

	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected no heal choice interrupt for plague mage on attack damage, got %v", g.State.PendingInterrupt.Type)
	}
	if victim.Heal != 2 {
		t.Fatalf("expected plague mage heal unchanged when taking attack damage, got %d", victim.Heal)
	}
}

func TestPlagueMageCanUseHealAgainstMagicDamage(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "A", "angel", model.RedCamp); err != nil {
		t.Fatalf("add player p1 failed: %v", err)
	}
	if err := g.AddPlayer("p2", "B", "plague_mage", model.BlueCamp); err != nil {
		t.Fatalf("add player p2 failed: %v", err)
	}

	victim := g.State.Players["p2"]
	victim.Heal = 2

	g.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   "p1",
			TargetID:   "p2",
			Damage:     1,
			DamageType: "magic",
			Stage:      0,
		},
	}
	g.State.Phase = model.PhasePendingDamageResolution
	g.State.ReturnPhase = model.PhaseTurnEnd

	g.Drive()

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected heal choice interrupt for plague mage on magic damage")
	}
	ctx, _ := g.State.PendingInterrupt.Context.(map[string]interface{})
	if ct, _ := ctx["choice_type"].(string); ct != "heal" {
		t.Fatalf("expected heal choice_type, got %q", ct)
	}
}

func TestMagicSwordsmanShadowRejectHidesMagicOption(t *testing.T) {
	obs := &promptCaptureObserver{}
	g := NewGameEngine(obs)
	if err := g.AddPlayer("p1", "A", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatalf("add player p1 failed: %v", err)
	}
	if err := g.AddPlayer("p2", "B", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add player p2 failed: %v", err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.Tokens["ms_shadow_form"] = 1
	p1.Hand = []model.Card{
		{ID: "m1", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementEarth},
	}
	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseActionSelection

	g.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt")
	}
	hasMagic := false
	hasCannotAct := false
	for _, opt := range obs.lastPrompt.Options {
		if opt.ID == "magic" {
			hasMagic = true
		}
		if opt.ID == "cannot_act" {
			hasCannotAct = true
		}
	}
	if hasMagic {
		t.Fatalf("expected no magic option under shadow reject")
	}
	if !hasCannotAct {
		t.Fatalf("expected cannot_act option when only unplayable magic cards exist")
	}
}

func TestMagicSwordsmanShadowGather_PersistsThisTurnAndReleasesNextTurn(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "A", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatalf("add player p1 failed: %v", err)
	}
	if err := g.AddPlayer("p2", "B", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add player p2 failed: %v", err)
	}

	g.State.Deck = rules.InitDeck()
	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseStartup

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	// 启动阶段应出现暗影凝聚可选中断。
	g.Drive()
	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt for magic swordsman")
	}
	startupIntr := g.State.PendingInterrupt
	shadowIdx := -1
	for i, skillID := range startupIntr.SkillIDs {
		if skillID == "ms_shadow_gather" {
			shadowIdx = i
			break
		}
	}
	if shadowIdx < 0 {
		t.Fatalf("expected ms_shadow_gather in startup skills, got %+v", startupIntr.SkillIDs)
	}

	// 选择发动暗影凝聚。
	if err := g.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{shadowIdx},
	}); err != nil {
		t.Fatalf("confirm startup shadow gather failed: %v", err)
	}

	// 回到同回合行动阶段时，仍应保持暗影形态。
	if p1.Tokens["ms_shadow_form"] <= 0 {
		t.Fatalf("shadow form should persist in current turn after startup confirm")
	}
	if p1.Tokens["ms_shadow_release_pending"] <= 0 {
		t.Fatalf("shadow release pending flag should remain until next own startup")
	}

	// 同回合发起一次攻击，应触发暗影之力(+1)，受击方应摸2张（火焰斩基础1）。
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = nil
	p2.Heal = 0
	g.State.Phase = model.PhaseActionSelection

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
		t.Fatalf("take response failed: %v", err)
	}
	if len(p2.Hand) != 2 {
		t.Fatalf("expected target hand=2 after shadow-power boosted attack, got %d", len(p2.Hand))
	}

	// 模拟下一次自己回合开始：应自动转正脱离暗影形态。
	p1.TurnState = model.NewPlayerTurnState()
	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseStartup
	g.State.PendingInterrupt = nil
	g.Drive()

	if p1.Tokens["ms_shadow_form"] != 0 {
		t.Fatalf("shadow form should be released at next own startup")
	}
	if p1.Tokens["ms_shadow_release_pending"] != 0 {
		t.Fatalf("shadow release pending should be cleared at next own startup")
	}
}

func TestMagicSwordsmanAsuraCombo_OnlyOncePerTurn(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "A", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatalf("add player p1 failed: %v", err)
	}
	if err := g.AddPlayer("p2", "B", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add player p2 failed: %v", err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	g.State.CurrentTurn = 0

	// 第一次攻击行动结束：应出现修罗连斩可选响应。
	g.State.Phase = model.PhaseExtraAction
	p1.TurnState.LastActionType = string(model.ActionAttack)
	g.Drive()

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response interrupt for first asura combo trigger")
	}
	asuraIdx := -1
	for i, skillID := range g.State.PendingInterrupt.SkillIDs {
		if skillID == "ms_asura_combo" {
			asuraIdx = i
			break
		}
	}
	if asuraIdx < 0 {
		t.Fatalf("expected ms_asura_combo in response skills, got %+v", g.State.PendingInterrupt.SkillIDs)
	}

	if err := g.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{asuraIdx},
	}); err != nil {
		t.Fatalf("confirm asura combo failed: %v", err)
	}

	if got := p1.TurnState.UsedSkillCounts["ms_asura_combo"]; got != 1 {
		t.Fatalf("expected ms_asura_combo used count to be 1, got %d", got)
	}

	// 同回合再次“攻击行动结束”：不应再出现修罗连斩。
	g.State.PendingInterrupt = nil
	g.State.Phase = model.PhaseExtraAction
	p1.TurnState.LastActionType = string(model.ActionAttack)
	g.Drive()

	if g.State.PendingInterrupt != nil && g.State.PendingInterrupt.Type == model.InterruptResponseSkill {
		for _, skillID := range g.State.PendingInterrupt.SkillIDs {
			if skillID == "ms_asura_combo" {
				t.Fatalf("ms_asura_combo should not trigger more than once in the same turn")
			}
		}
	}
}
