package engine_test

import (
	"starcup-engine/internal/engine"
	playerpkg "starcup-engine/internal/engine/player"
	"starcup-engine/internal/testutils"
	"strings"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

type actionPromptObserver struct {
	lastPrompt *model.Prompt
}

func (o *actionPromptObserver) OnGameEvent(event model.GameEvent) {
	if event.Type != model.EventAskInput {
		return
	}
	prompt := event.Prompt
	ok := prompt != nil
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

func buildActionSelectionEngine(t *testing.T, extraAction string) (*engine.GameEngine, *actionPromptObserver) {
	t.Helper()

	obs := &actionPromptObserver{}
	game := engine.NewGameEngine(obs)

	if err := game.AddPlayer("p1", "Tester", "blade_master", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

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

func buildActionSelectionElementalistEngine(t *testing.T, extraAction string) (*engine.GameEngine, *actionPromptObserver) {
	t.Helper()

	obs := &actionPromptObserver{}
	game := engine.NewGameEngine(obs)

	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

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

func TestActionSelectionPrompt_FlexibleExtraActionShowsAttackAndMagicOnly(t *testing.T) {
	game, obs := buildActionSelectionEngine(t, model.ExtraActionAny)
	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}

	options := promptOptionSet(obs.lastPrompt)
	if !options["attack"] || !options["magic"] {
		t.Fatalf("expected attack and magic options for flexible extra action, got %+v", obs.lastPrompt.Options)
	}
	if options["special"] || options["buy"] || options["extract"] || options["synthesize"] || options["cannot_act"] {
		t.Fatalf("unexpected options for flexible extra action prompt: %+v", obs.lastPrompt.Options)
	}
	if !strings.Contains(obs.lastPrompt.Message, "攻击或法术") {
		t.Fatalf("expected flexible extra-action hint in prompt message, got: %s", obs.lastPrompt.Message)
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

func TestActionSelectionPrompt_MagicLancerFullnessBonusInAttackPrompt(t *testing.T) {
	obs := &actionPromptObserver{}
	game := engine.NewGameEngine(obs)

	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.CurrentExtraAction = "Attack"
	p1.Hand = []model.Card{
		{ID: "atk", Name: "测试攻击", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
	}
	game.ApplyNextAttackDamageRule(p1.ID, "ml_fullness_next_attack_bonus", "ml_fullness", 2, model.RuleLifeUntilTurnEnd)

	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}
	options := promptOptionSet(obs.lastPrompt)
	if !options["attack"] {
		t.Fatalf("expected attack option for fullness extra attack, got %+v", obs.lastPrompt.Options)
	}
	if !strings.Contains(obs.lastPrompt.Message, "【充盈】下一次主动攻击伤害额外+2") {
		t.Fatalf("expected fullness attack bonus hint in prompt message, got: %s", obs.lastPrompt.Message)
	}
}

func TestActionSelection_ExtraActionCannotActSkipsWhenNoLegalAction(t *testing.T) {
	game, _ := buildActionSelectionEngine(t, "Attack")
	p1 := game.State.Players["p1"]
	p1.Hand = []model.Card{
		{ID: "mag-only", Name: "测试法术", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 1},
	}

	err := game.HandleActionSelection(model.PlayerAction{
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
	if game.State.TurnStage != model.TurnStageTurnEnd {
		t.Fatalf("expected turn stage turn_end after skipping extra action, got %s", game.State.TurnStage)
	}
}

func TestActionSelection_ExtraActionCannotActRejectedWhenLegalActionExists(t *testing.T) {
	game, _ := buildActionSelectionEngine(t, "Attack")

	err := game.HandleActionSelection(model.PlayerAction{
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

func TestActionQueueConsumesSelectedAttackByCardIDAfterHandReorder(t *testing.T) {
	game, _ := buildActionSelectionEngine(t, "")
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Hand = []model.Card{
		{ID: "atk-old-index", Name: "旧下标攻击", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "atk-selected", Name: "选中攻击", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
	}

	err := game.HandleActionSelection(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   "atk-selected",
	})
	if err != nil {
		t.Fatalf("attack selection failed: %v", err)
	}
	if len(game.State.ActionQueue) != 1 {
		t.Fatalf("expected one queued action, got %d", len(game.State.ActionQueue))
	}
	if got := game.State.ActionQueue[0].CardID; got != "atk-selected" {
		t.Fatalf("expected queued card id atk-selected, got %q", got)
	}

	p1.Hand[0], p1.Hand[1] = p1.Hand[1], p1.Hand[0]
	game.Drive()

	if len(p1.Hand) != 1 || p1.Hand[0].ID != "atk-old-index" {
		t.Fatalf("expected only old-index card to remain in hand, got %+v", p1.Hand)
	}
	if len(game.State.DiscardPile) == 0 || game.State.DiscardPile[len(game.State.DiscardPile)-1].ID != "atk-selected" {
		t.Fatalf("expected selected card in discard pile, got %+v", game.State.DiscardPile)
	}
	if len(game.State.CombatStack) != 1 || game.State.CombatStack[0].TargetID != p2.ID {
		t.Fatalf("expected combat against p2 after selected card resolves, got %+v", game.State.CombatStack)
	}
}

func TestActionQueueConsumesSelectedMagicByCardIDAfterHandReorder(t *testing.T) {
	game, _ := buildActionSelectionEngine(t, "")
	p1 := game.State.Players["p1"]
	p1.Hand = []model.Card{
		{ID: "magic-old-index", Name: "旧下标法术", Type: model.CardTypeMagic, Element: model.ElementFire, Damage: 1},
		{ID: "magic-selected", Name: "选中法术", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}

	err := game.HandleActionSelection(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdMagic,
		TargetID: "p2",
		CardID:   "magic-selected",
	})
	if err != nil {
		t.Fatalf("magic selection failed: %v", err)
	}
	if len(game.State.ActionQueue) != 1 {
		t.Fatalf("expected one queued action, got %d", len(game.State.ActionQueue))
	}
	if got := game.State.ActionQueue[0].CardID; got != "magic-selected" {
		t.Fatalf("expected queued card id magic-selected, got %q", got)
	}

	p1.Hand[0], p1.Hand[1] = p1.Hand[1], p1.Hand[0]
	game.Drive()

	if len(p1.Hand) != 1 || p1.Hand[0].ID != "magic-old-index" {
		t.Fatalf("expected only old-index magic card to remain in hand, got %+v", p1.Hand)
	}
	if len(game.State.DiscardPile) == 0 || game.State.DiscardPile[len(game.State.DiscardPile)-1].ID != "magic-selected" {
		t.Fatalf("expected selected magic card in discard pile, got %+v", game.State.DiscardPile)
	}
}

func TestInterruptChoiceSelectResolvesCardIDsFromPromptOptions(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Tester", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.Hand = []model.Card{
		{ID: "card-a", Name: "A", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "card-b", Name: "B", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 1},
		{ID: "card-c", Name: "C", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
	}
	game.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: "p1",
		Context: map[string]interface{}{
			"choice_type":    "system_discard_cards",
			"discard_count":  2,
			"no_morale_loss": true,
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSelect,
		CardIDs:  []string{"card-c", "card-a"},
	}); err != nil {
		t.Fatalf("card id select failed: %v", err)
	}
	if len(p1.Hand) != 1 || p1.Hand[0].ID != "card-b" {
		t.Fatalf("expected only card-b to remain in hand, got %+v", p1.Hand)
	}
	discarded := map[string]bool{}
	for _, card := range game.State.DiscardPile {
		discarded[card.ID] = true
	}
	if !discarded["card-a"] || !discarded["card-c"] || discarded["card-b"] {
		t.Fatalf("expected card-a and card-c in discard pile only, got %+v", game.State.DiscardPile)
	}
}

func TestActionSkillDiscardSelectionPublishesCardPickerPrompt(t *testing.T) {
	obs := &actionPromptObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "earth-card", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "earth_seal",
		TargetIDs: []string{"p2"},
	}); err != nil {
		t.Fatalf("earth seal should enter discard selection: %v", err)
	}

	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard selection interrupt, got %+v", game.State.PendingInterrupt)
	}
	if obs.lastPrompt == nil {
		t.Fatalf("expected discard card picker prompt to be published")
	}
	if obs.lastPrompt.PlayerID != "p1" || obs.lastPrompt.SkillID != "earth_seal" {
		t.Fatalf("unexpected prompt owner/skill: %+v", obs.lastPrompt)
	}
	if obs.lastPrompt.Presentation == nil || obs.lastPrompt.Presentation.Kind != model.PresentationCardPicker {
		t.Fatalf("expected card picker presentation, got %+v", obs.lastPrompt.Presentation)
	}
	if obs.lastPrompt.Presentation.CardFilter != "discard" {
		t.Fatalf("expected skill cost discard filter, got %+v", obs.lastPrompt.Presentation)
	}
	if obs.lastPrompt.Presentation.DiscardReason != "skill_cost" {
		t.Fatalf("expected skill_cost discard reason, got %+v", obs.lastPrompt.Presentation)
	}
	if obs.lastPrompt.ChoiceType != "system_discard_cards" {
		t.Fatalf("expected system_discard_cards choice type, got %q", obs.lastPrompt.ChoiceType)
	}
}

func TestActionSkillDiscardSelectionRejectsInvalidCardWithoutConsumingInterrupt(t *testing.T) {
	obs := &actionPromptObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Spirit", "spirit_caster", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy1", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Dummy2", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "poison", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementEarth},
		{ID: "thunder", Name: "雷击", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "sc_talisman_thunder",
		TargetIDs: []string{"p2", "p3"},
	}); err != nil {
		t.Fatalf("talisman thunder should enter discard selection: %v", err)
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard selection interrupt, got %+v", game.State.PendingInterrupt)
	}
	if obs.lastPrompt == nil {
		t.Fatalf("expected discard card picker prompt")
	}
	options := promptOptionSet(obs.lastPrompt)
	if options["0"] || !options["1"] {
		t.Fatalf("expected prompt to expose only thunder card option, got %+v", obs.lastPrompt.Options)
	}

	err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	if err == nil || !strings.Contains(err.Error(), "不符合元素要求") {
		t.Fatalf("expected invalid discard element error, got %v", err)
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("invalid discard should keep discard interrupt pending, got %+v", game.State.PendingInterrupt)
	}
	if len(p1.Hand) != 2 {
		t.Fatalf("invalid discard should not consume hand cards, got %+v", p1.Hand)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("valid thunder discard should resume skill: %v", err)
	}
	if len(p1.Hand) != 1 || p1.Hand[0].ID != "poison" {
		t.Fatalf("expected thunder card consumed and poison retained, got %+v", p1.Hand)
	}
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected talisman follow-up prompt after valid discard")
	}
	ctx, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctx["choice_type"].(string); got != "sc_incant_confirm" {
		t.Fatalf("expected incantation follow-up prompt, got %+v", game.State.PendingInterrupt)
	}
}

func TestActionSelection_ExtraMagicAllowsSkill(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = rules.InitDeck()
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.CurrentExtraAction = "Magic"
	p1.Tokens["element"] = 3

	t.Logf("Before HandleAction: TurnStage=%s, PendingInterrupt=%v, ActionQueue=%d, PlayerOrder=%v, CurrentTurn=%d",
		game.State.TurnStage, game.State.PendingInterrupt, len(game.State.ActionQueue), game.State.PlayerOrder, game.State.CurrentTurn)
	t.Logf("p1.Character=%v, p1.Tokens=%v", p1.Character != nil, p1.Tokens)
	// 测试额外法术行动时可以使用技能
	// 直接调用 HandleActionSelection 测试技能执行流程
	err := game.HandleActionSelection(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_ignite",
		TargetIDs: []string{"p2"},
	})
	if err != nil {
		t.Fatalf("expected extra magic action can use skill, got err: %v", err)
	}
	// 验证技能执行后的状态
	t.Logf("After HandleActionSelection: TurnStage=%s, PendingDamageQueue=%d, PendingActions=%d",
		game.State.TurnStage, len(game.State.PendingDamageQueue), len(p1.TurnState.PendingActions))
	// 技能执行后应该进入 ActionEnd 阶段，等待状态机推进
	if game.State.TurnStage != model.TurnStageActionEnd {
		t.Fatalf("expected TurnStage=ActionEnd, got %s", game.State.TurnStage)
	}
	// 伤害应该已入队
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending damage queued")
	}
	// 额外法术行动应该已添加
	if len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected extra magic action added to pending actions")
	}
}

func TestActionSelection_ExtraMagicCannotActRejectedWhenSkillAvailable(t *testing.T) {
	game, _ := buildActionSelectionElementalistEngine(t, "Magic")

	err := game.HandleActionSelection(model.PlayerAction{
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

func TestActionSelectionPrompt_ArbiterForcedDoomsdayOnlyShowsMagic(t *testing.T) {
	obs := &actionPromptObserver{}
	game := engine.NewGameEngine(obs)

	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["judgment"] = 4
	p1.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 1

	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}
	options := promptOptionSet(obs.lastPrompt)
	if !options["magic"] {
		t.Fatalf("expected magic option for forced doomsday, got %+v", obs.lastPrompt.Options)
	}
	if options["attack"] || options["buy"] || options["extract"] || options["synthesize"] {
		t.Fatalf("unexpected options for forced doomsday prompt: %+v", obs.lastPrompt.Options)
	}
	if obs.lastPrompt.SkillID != "arbiter_doomsday" {
		t.Fatalf("expected prompt skill id arbiter_doomsday, got %q", obs.lastPrompt.SkillID)
	}
}

func TestActionSelectionPrompt_TauntWithoutAttackOnlyShowsSkip(t *testing.T) {
	obs := &actionPromptObserver{}
	game := engine.NewGameEngine(obs)

	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = false
	p2.IsActive = true
	p2.TurnState = model.NewPlayerTurnState()
	p2.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 1
	p2.Field = []*model.FieldCard{{
		Card:     model.Card{ID: "taunt", Name: "挑衅", Type: model.CardTypeMagic, Element: model.ElementLight},
		OwnerID:  p2.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectHeroTaunt,
	}}
	p2.Hand = []model.Card{{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0}}

	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt, got nil")
	}
	options := promptOptionSet(obs.lastPrompt)
	if !options["cannot_act"] {
		t.Fatalf("expected cannot_act option for taunt without attack card, got %+v", obs.lastPrompt.Options)
	}
	if options["attack"] || options["magic"] {
		t.Fatalf("unexpected action options for taunt without attack card: %+v", obs.lastPrompt.Options)
	}
}

func TestActionSelectionPrompt_MagicSwordsmanShadowForm_StillShowsMagicWhenSkillUsable(t *testing.T) {
	obs := &actionPromptObserver{}
	game := engine.NewGameEngine(obs)

	if err := game.AddPlayer("p1", "MS", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	playerpkg.SetForm(p1, model.FormMagicSwordsmanShadow)
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
