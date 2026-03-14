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
	if p, ok := event.Data.(*model.Prompt); ok && p != nil {
		o.lastPrompt = p
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

func choiceTypeOfPendingInterruptUI(intr *model.Interrupt) string {
	if intr == nil {
		return ""
	}
	data, _ := intr.Context.(map[string]interface{})
	v, _ := data["choice_type"].(string)
	return v
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
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
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

// UI回归：血气屏障“确认弹框”和“目标弹框”都允许取消，且取消后流程不报错不卡住。
func TestUIRegression_CrimsonBloodBarrier_TwoLevelCancelFlow(t *testing.T) {
	makeGame := func() *engine.GameEngine {
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
		game.State.Phase = model.PhaseActionSelection
		return game
	}

	triggerBloodBarrier := func(game *engine.GameEngine) {
		// 端到端：由敌方打出魔弹，血色剑灵选择承受后触发血气屏障。
		if err := game.HandleAction(model.PlayerAction{
			PlayerID:  "p2",
			Type:      model.CmdMagic,
			CardIndex: 0,
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
	}

	t.Run("确认弹框可取消", func(t *testing.T) {
		game := makeGame()
		triggerBloodBarrier(game)

		if ct := choiceTypeOfPendingInterruptUI(game.State.PendingInterrupt); ct != "css_blood_barrier_counter_confirm" {
			t.Fatalf("预期确认弹框中断，实际 choice_type=%q", ct)
		}
		prompt := game.GetCurrentPrompt()
		if !hasPromptOptionUI(prompt, "cancel") {
			t.Fatalf("确认弹框应包含取消按钮，实际: %+v", prompt)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p1",
			Type:     model.CmdCancel,
		}); err != nil {
			t.Fatalf("取消确认弹框失败: %v", err)
		}
		if game.State.PendingInterrupt != nil {
			t.Fatalf("取消后应清空中断，实际: %+v", game.State.PendingInterrupt)
		}
	})

	t.Run("目标弹框可取消", func(t *testing.T) {
		game := makeGame()
		triggerBloodBarrier(game)

		if err := game.HandleAction(model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{0}, // 先选“是”进入目标弹框
		}); err != nil {
			t.Fatalf("确认进入目标弹框失败: %v", err)
		}

		if ct := choiceTypeOfPendingInterruptUI(game.State.PendingInterrupt); ct != "css_blood_barrier_target" {
			t.Fatalf("预期目标弹框中断，实际 choice_type=%q", ct)
		}
		prompt := game.GetCurrentPrompt()
		if !hasPromptOptionUI(prompt, "cancel") {
			t.Fatalf("目标弹框应包含取消按钮，实际: %+v", prompt)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p1",
			Type:     model.CmdCancel,
		}); err != nil {
			t.Fatalf("取消目标弹框失败: %v", err)
		}
		if game.State.PendingInterrupt != nil {
			t.Fatalf("取消后应清空中断，实际: %+v", game.State.PendingInterrupt)
		}
	})
}
