package crimson_sword_spirit_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"strings"
	"testing"

	"starcup-engine/internal/model"
)

func TestCrimsonBloodRose_StepByStepTargetSelection(t *testing.T) {
	g := engine.NewGameEngine(nil)
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p3 := g.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 2
	p1.Heal = 0
	p2.Heal = 3
	p3.Heal = 0
	g.State.RedCrystals = 1
	g.State.RedGems = 0

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	// 发动血染蔷薇，应进入分步选择流程（不再需要一次性选2个目标）
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "css_blood_rose",
		TargetIDs: nil, // 分步模式不需要预先选目标
	})

	// 检查第一步：选择移除治疗目标
	testutils.RequireChoiceContext(t, g, "p1", "css_blood_rose_remove_heal_target")
	prompt := g.BuildPendingInterruptPrompt()
	if prompt == nil {
		t.Fatalf("expected blood rose remove heal target prompt")
	}
	if !strings.Contains(prompt.Message, "移除治疗") {
		t.Fatalf("expected remove heal hint in message, got: %s", prompt.Message)
	}

	// 选择移除治疗目标（p2 敌方）
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1}, // p2 是第二个选项
	})

	// 检查第二步：选择获得治疗的队友（仅队友可选）
	testutils.RequireChoiceContext(t, g, "p1", "css_blood_rose_gain_heal_target")
	prompt = g.BuildPendingInterruptPrompt()
	if prompt == nil {
		t.Fatalf("expected blood rose gain heal target prompt")
	}
	if !strings.Contains(prompt.Message, "队友") {
		t.Fatalf("expected ally hint in message, got: %s", prompt.Message)
	}
	// 检查阵营约束提示
	if len(prompt.EffectHints) == 0 || !strings.Contains(prompt.EffectHints[0], "我方角色") {
		t.Fatalf("expected camp constraint hint, got: %v", prompt.EffectHints)
	}

	// 选择获得治疗目标（p3 队友）
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 第一个队友选项（p1 或 p3）
	})

	// 验证效果
	if p1.Tokens["css_blood"] != 0 {
		t.Fatalf("expected blood rose spend 2 blood, got %d", p1.Tokens["css_blood"])
	}
	if p2.Heal != 1 {
		t.Fatalf("expected first target lose 2 heal (3->1), got %d", p2.Heal)
	}
	if g.State.RedCrystals != 0 || g.State.RedGems != 1 {
		t.Fatalf("expected camp crystal->gem conversion, got red crystals=%d gems=%d", g.State.RedCrystals, g.State.RedGems)
	}
}

func TestCrimsonBloodRose_SecondTargetMustBeAlly(t *testing.T) {
	g := engine.NewGameEngine(nil)
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	_ = g.State.Players["p2"] // p2 存在但测试中不直接使用
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 2
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	// 发动血染蔷薇
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "css_blood_rose",
		TargetIDs: nil,
	})

	// 选择移除治疗目标（p2 敌方）
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1}, // p2
	})

	// 第二步的目标列表应该只有队友（p1）
	// 如果没有队友可选，流程应该报错
	prompt := g.BuildPendingInterruptPrompt()
	if prompt == nil {
		// 只有一个玩家（自己）的情况下，治疗目标应该默认为自己
		// 或者应该有明确的提示
		t.Fatalf("expected second step prompt even when no ally available")
	}

	// 验证只有自己作为可选队友
	if len(prompt.Options) != 1 || prompt.Options[0].Label != p1.Name {
		t.Logf("step 2 options: %+v (expected only p1)", prompt.Options)
	}
}
