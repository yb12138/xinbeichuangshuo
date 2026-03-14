package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
)

type actionPromptObserver struct {
	lastPrompt *model.Prompt
}

func (o *actionPromptObserver) OnGameEvent(event model.GameEvent) {
	if event.Type != model.EventAskInput {
		return
	}
	prompt, ok := event.Data.(*model.Prompt)
	if !ok || prompt == nil {
		return
	}
	copied := *prompt
	copied.Options = append([]model.PromptOption(nil), prompt.Options...)
	o.lastPrompt = &copied
}

func promptOptionSet(prompt *model.Prompt) map[string]bool {
	set := make(map[string]bool, len(prompt.Options))
	for _, opt := range prompt.Options {
		set[opt.ID] = true
	}
	return set
}

func promptOptionLabel(prompt *model.Prompt, optionID string) string {
	for _, opt := range prompt.Options {
		if opt.ID == optionID {
			return opt.Label
		}
	}
	return ""
}

func buildActionSelectionEngine(t *testing.T, extraAction string) (*GameEngine, *actionPromptObserver) {
	t.Helper()

	obs := &actionPromptObserver{}
	game := NewGameEngine(obs)

	if err := game.AddPlayer("p1", "Tester", "blade_master", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.CurrentExtraAction = extraAction
	p1.Hand = []model.Card{
		{ID: "atk", Name: "测试攻击", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "mag", Name: "测试法术", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 1},
	}

	return game, obs
}

func buildActionSelectionElementalistEngine(t *testing.T, extraAction string) (*GameEngine, *actionPromptObserver) {
	t.Helper()

	obs := &actionPromptObserver{}
	game := NewGameEngine(obs)

	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.CurrentExtraAction = extraAction
	p1.Tokens["element"] = 3
	p1.Hand = []model.Card{
		{ID: "atk-only", Name: "测试攻击", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	return game, obs
}

func TestActionSelectionPrompt_ExtraAttackOnlyShowsAttack(t *testing.T) {
	game, obs := buildActionSelectionEngine(t, "Attack")
	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}

	options := promptOptionSet(obs.lastPrompt)
	if !options["attack"] {
		t.Fatalf("expected option attack, got %+v", obs.lastPrompt.Options)
	}
	if options["magic"] || options["buy"] || options["extract"] || options["synthesize"] || options["cannot_act"] {
		t.Fatalf("unexpected options for extra attack prompt: %+v", obs.lastPrompt.Options)
	}
	if !strings.Contains(obs.lastPrompt.Message, "当前为额外攻击行动") {
		t.Fatalf("expected extra-attack hint in prompt message, got: %s", obs.lastPrompt.Message)
	}
}

func TestActionSelectionPrompt_ExtraMagicOnlyShowsMagic(t *testing.T) {
	game, obs := buildActionSelectionEngine(t, "Magic")
	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}

	options := promptOptionSet(obs.lastPrompt)
	if !options["magic"] {
		t.Fatalf("expected option magic, got %+v", obs.lastPrompt.Options)
	}
	if options["attack"] || options["buy"] || options["extract"] || options["synthesize"] || options["cannot_act"] {
		t.Fatalf("unexpected options for extra magic prompt: %+v", obs.lastPrompt.Options)
	}
	if !strings.Contains(obs.lastPrompt.Message, "当前为额外法术行动") {
		t.Fatalf("expected extra-magic hint in prompt message, got: %s", obs.lastPrompt.Message)
	}
}

func TestActionSelectionPrompt_ExtraAttackNoLegalActionShowsSkip(t *testing.T) {
	game, obs := buildActionSelectionEngine(t, "Attack")
	game.State.Players["p1"].Hand = []model.Card{
		{ID: "mag-only", Name: "测试法术", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 1},
	}
	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}
	options := promptOptionSet(obs.lastPrompt)
	if !options["cannot_act"] {
		t.Fatalf("expected cannot_act option for no-legal extra attack, got %+v", obs.lastPrompt.Options)
	}
	if options["attack"] || options["magic"] || options["buy"] || options["extract"] || options["synthesize"] {
		t.Fatalf("unexpected options for no-legal extra attack prompt: %+v", obs.lastPrompt.Options)
	}
	if label := promptOptionLabel(obs.lastPrompt, "cannot_act"); !strings.Contains(label, "跳过额外行动") {
		t.Fatalf("expected cannot_act label to indicate skip extra action, got: %q", label)
	}
}

func TestActionSelectionPrompt_ExtraMagicNoLegalActionShowsSkip(t *testing.T) {
	game, obs := buildActionSelectionEngine(t, "Magic")
	game.State.Players["p1"].Hand = []model.Card{
		{ID: "atk-only", Name: "测试攻击", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}
	options := promptOptionSet(obs.lastPrompt)
	if !options["cannot_act"] {
		t.Fatalf("expected cannot_act option for no-legal extra magic, got %+v", obs.lastPrompt.Options)
	}
	if options["attack"] || options["magic"] || options["buy"] || options["extract"] || options["synthesize"] {
		t.Fatalf("unexpected options for no-legal extra magic prompt: %+v", obs.lastPrompt.Options)
	}
}

func TestActionSelectionPrompt_ExtraMagicWithSkillOnlyShowsMagic(t *testing.T) {
	game, obs := buildActionSelectionElementalistEngine(t, "Magic")
	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}
	options := promptOptionSet(obs.lastPrompt)
	if !options["magic"] {
		t.Fatalf("expected magic option when only skill is usable, got %+v", obs.lastPrompt.Options)
	}
	if options["cannot_act"] {
		t.Fatalf("did not expect cannot_act when skill is usable in extra magic, got %+v", obs.lastPrompt.Options)
	}
}

func TestActionSelection_ExtraActionCannotActSkipsWhenNoLegalAction(t *testing.T) {
	game, _ := buildActionSelectionEngine(t, "Attack")
	p1 := game.State.Players["p1"]
	p1.Hand = []model.Card{
		{ID: "mag-only", Name: "测试法术", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 1},
	}

	err := game.handleActionSelection(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCannotAct,
	})
	if err != nil {
		t.Fatalf("expected skip extra action to succeed, got err: %v", err)
	}
	if p1.TurnState.CurrentExtraAction != "" {
		t.Fatalf("expected extra-action constraint cleared, got %q", p1.TurnState.CurrentExtraAction)
	}
	if len(p1.TurnState.CurrentExtraElement) != 0 {
		t.Fatalf("expected extra-action element constraint cleared, got %+v", p1.TurnState.CurrentExtraElement)
	}
	if game.State.Phase != model.PhaseTurnEnd {
		t.Fatalf("expected phase turn_end after skipping extra action, got %s", game.State.Phase)
	}
}

func TestActionSelection_ExtraActionCannotActRejectedWhenLegalActionExists(t *testing.T) {
	game, _ := buildActionSelectionEngine(t, "Attack")

	err := game.handleActionSelection(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCannotAct,
	})
	if err == nil {
		t.Fatalf("expected skip extra action to be rejected when legal action exists")
	}
	if !strings.Contains(err.Error(), "不能跳过") {
		t.Fatalf("expected reject reason to mention cannot skip extra action, got: %v", err)
	}
}

func TestActionSelection_ExtraMagicAllowsSkill(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.CurrentExtraAction = "Magic"
	p1.Tokens["element"] = 3

	err := game.handleActionSelection(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_ignite",
		TargetIDs: []string{"p2"},
	})
	if err != nil {
		t.Fatalf("expected extra magic action can use skill, got err: %v", err)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected ignite queued pending damage")
	}
	if game.State.ReturnPhase != model.PhaseExtraAction {
		t.Fatalf("expected return phase extra action, got %s", game.State.ReturnPhase)
	}
}

func TestActionSelection_ExtraMagicCannotActRejectedWhenSkillAvailable(t *testing.T) {
	game, _ := buildActionSelectionElementalistEngine(t, "Magic")

	err := game.handleActionSelection(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCannotAct,
	})
	if err == nil {
		t.Fatalf("expected skip extra magic action to be rejected when action skill exists")
	}
	if !strings.Contains(err.Error(), "不能跳过") {
		t.Fatalf("expected reject reason to mention cannot skip extra action, got: %v", err)
	}
}

func TestActionSelectionPrompt_MagicSwordsmanShadowForm_StillShowsMagicWhenSkillUsable(t *testing.T) {
	obs := &actionPromptObserver{}
	game := NewGameEngine(obs)

	if err := game.AddPlayer("p1", "MS", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["ms_shadow_form"] = 1
	// 暗影流星需要至少2张法术牌弃置；暗影抗拒会禁用法术牌直接打出。
	p1.Hand = []model.Card{
		{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
		{ID: "m2", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}

	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}
	options := promptOptionSet(obs.lastPrompt)
	if !options["magic"] {
		t.Fatalf("expected magic option for action-skill entry, got %+v", obs.lastPrompt.Options)
	}
	if options["cannot_act"] {
		t.Fatalf("did not expect cannot_act when shadow meteor is usable, got %+v", obs.lastPrompt.Options)
	}
}
