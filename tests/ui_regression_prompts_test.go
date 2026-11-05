package tests

import (
	"strings"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

type uiPromptObserver struct {
	lastPrompt *model.Prompt
}

func (o *uiPromptObserver) OnGameEvent(event model.GameEvent) {
	if event.Type != model.EventAskInput {
		return
	}
	if event.Prompt != nil {
		o.lastPrompt = event.Prompt
	}
}

func hasPromptOptionUI(prompt *model.Prompt, id string) bool {
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

func interruptHasSkillIDUI(intr *model.Interrupt, skillID string) bool {
	if intr == nil {
		return false
	}
	for _, sid := range intr.SkillIDs {
		if sid == skillID {
			return true
		}
	}
	return false
}

// UI回归：黄泉震颤触发后，攻击响应弹框中不应出现“应战(counter)”按钮。
func TestUIRegression_YellowSpring_HidesCounterOptionInResponsePrompt(t *testing.T) {
	obs := &uiPromptObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "魔剑士", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "防守方", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	// 增加同阵营队友，确保“正常情况下”防守方是存在应战反弹目标的。
	if err := game.AddPlayer("p3", "同阵营队友", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p2.IsActive = false
	p1.Gem = 1 // 黄泉震颤消耗条件
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "counter_card", Name: "暗灭", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 1},
		// 防御选项只在手里有圣光时出现
		{ID: "light_card", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   "atk",
	}); err != nil {
		t.Fatalf("发起攻击失败: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("预期出现黄泉震颤响应中断，实际: %+v", game.State.PendingInterrupt)
	}
	if err := game.ConfirmResponseSkill("p1", "ms_yellow_spring"); err != nil {
		t.Fatalf("确认黄泉震颤失败: %v", err)
	}

	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatalf("预期收到防守方响应弹框")
	}
	if obs.lastPrompt.PlayerID != "p2" {
		t.Fatalf("预期弹框玩家为 p2，实际: %s", obs.lastPrompt.PlayerID)
	}
	if !strings.Contains(obs.lastPrompt.Message, "需要响应来自") {
		t.Fatalf("预期为攻击响应弹框，实际消息: %q", obs.lastPrompt.Message)
	}
	if hasPromptOptionUI(obs.lastPrompt, "counter") {
		t.Fatalf("黄泉震颤后不应出现应战按钮，实际选项: %+v", obs.lastPrompt.Options)
	}
	if !hasPromptOptionUI(obs.lastPrompt, "take") || !hasPromptOptionUI(obs.lastPrompt, "defend") {
		t.Fatalf("预期仍有承受/防御按钮，实际选项: %+v", obs.lastPrompt.Options)
	}
}

// UI回归：血气屏障确认后本次伤害立即-1，随后弹出不可取消的选目标提示（仅对手），选完后原伤害继续结算。
func TestUIRegression_CrimsonBloodBarrier_PromptsEnemyTargetThenResumesDamage(t *testing.T) {
	game := engine.NewGameEngine(nil)
	if err := game.AddPlayer("p1", "血色剑灵", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "敌人", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = false
	p2.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 1
	p2.Hand = []model.Card{
		{ID: "mb1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = nil
	for i := 0; i < 10; i++ {
		game.State.Deck = append(game.State.Deck, model.Card{ID: "bb-deck-" + string(rune('a'+i)), Name: "补牌", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2})
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2",
		Type:     model.CmdMagic,
		CardID:   "mb1",
	}); err != nil {
		t.Fatalf("敌方打出魔弹失败: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("血色剑灵承受魔弹失败: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("预期出现血气屏障响应技能中断，实际: %+v", game.State.PendingInterrupt)
	}
	if !interruptHasSkillIDUI(game.State.PendingInterrupt, "css_blood_barrier") {
		t.Fatalf("预期包含 css_blood_barrier，实际技能列表: %+v", game.State.PendingInterrupt.SkillIDs)
	}
	if err := game.ConfirmResponseSkill("p1", "css_blood_barrier"); err != nil {
		t.Fatalf("确认血气屏障失败: %v", err)
	}

	if got := p1.Tokens["css_blood"]; got != 0 {
		t.Fatalf("预期确认后立即移除1鲜血，实际剩余 %d", got)
	}
	intr := game.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptChoice || intr.PlayerID != "p1" {
		t.Fatalf("预期出现血气屏障选目标提示，实际: %+v", intr)
	}
	prompt := game.GetCurrentPrompt()
	if prompt == nil || len(prompt.Options) != 1 || prompt.Options[0].TargetID != "p2" {
		t.Fatalf("预期仅可选对手 p2 且不可取消，实际: %+v", prompt)
	}
	p1HandBefore := len(p1.Hand)
	p2HandBefore := len(p2.Hand)
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}}); err != nil {
		t.Fatalf("选择反伤目标失败: %v", err)
	}
	game.Drive()
	if got := len(p1.Hand) - p1HandBefore; got != 1 {
		t.Fatalf("预期原魔弹伤害 -1 后结算为1点，实际 p1 摸牌 %d", got)
	}
	if got := len(p2.Hand) - p2HandBefore; got != 1 {
		t.Fatalf("预期 p2 承受1点反伤，实际摸牌 %d", got)
	}
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("预期伤害队列已清空，实际: %+v", game.State.PendingDamageQueue)
	}
}
