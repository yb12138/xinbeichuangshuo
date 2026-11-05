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
	if len(prompt.EffectHints) == 0 || !strings.Contains(prompt.EffectHints[0], "队友") {
		t.Fatalf("expected camp constraint hint, got: %v", prompt.EffectHints)
	}

	// 队友不含自己：候选只有 p3
	if len(prompt.Options) != 1 || prompt.Options[0].TargetID != "p3" {
		t.Fatalf("expected only ally p3 (not self) as gain-heal target, got %+v", prompt.Options)
	}
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	if p3.Heal != 1 || p1.Heal != 0 {
		t.Fatalf("expected ally p3 +1 heal and self unchanged, got p3=%d p1=%d", p3.Heal, p1.Heal)
	}

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

func TestCrimsonBloodRose_SecondTargetExcludesSelf(t *testing.T) {
	g := engine.NewGameEngine(nil)
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
	p1.Tokens["css_blood"] = 2
	p2.Heal = 2
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

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
		Selections: []int{1},
	})

	// 没有队友（自己不算）：不再弹出队友选择，直接结算其余效果，自己不 +1 治疗。
	if ct := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); ct == "css_blood_rose_gain_heal_target" {
		t.Fatalf("expected no gain-heal prompt when there is no teammate")
	}
	if p1.Heal != 0 {
		t.Fatalf("expected self not to gain heal from blood rose, got %d", p1.Heal)
	}
	if p2.Heal != 0 {
		t.Fatalf("expected remove-heal target lose 2 heal, got %d", p2.Heal)
	}
	if p1.Tokens["css_blood"] != 0 {
		t.Fatalf("expected blood rose spend 2 blood, got %d", p1.Tokens["css_blood"])
	}
}
