package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
)

type actionSummaryCaptureObserver struct {
	events []model.GameEvent
}

func (o *actionSummaryCaptureObserver) OnGameEvent(event model.GameEvent) {
	o.events = append(o.events, event)
}

func newActionSummaryTestGame(observer model.GameObserver) *GameEngine {
	game := NewGameEngine(observer)
	game.State.Players["p1"] = &model.Player{ID: "p1", Name: "Alice"}
	game.State.Players["p2"] = &model.Player{ID: "p2", Name: "Bob"}
	game.State.PlayerOrder = []string{"p1", "p2"}
	game.State.TurnStage = model.TurnStageActionEnd
	return game
}

func actionSummaryLines(events []model.GameEvent) []string {
	lines := []string{}
	for _, event := range events {
		if event.Type != model.EventActionStep || event.ActionStep == nil || event.ActionStep.Kind != "summary" {
			continue
		}
		lines = append(lines, event.ActionStep.Line)
	}
	return lines
}

func TestActionSummaryExcludesDetailActionSteps(t *testing.T) {
	observer := &actionSummaryCaptureObserver{}
	game := newActionSummaryTestGame(observer)

	game.BeginActionSummary("attack", "p1", "火焰斩", []string{"p2"})
	game.NotifyActionStep("中间过程：Bob承受伤害")
	game.NotifyDamageDealt("p1", "p2", 2, model.AttackDamage)
	game.FinalizeActionSummaryIfIdle()

	lines := actionSummaryLines(observer.events)
	if len(lines) != 1 {
		t.Fatalf("expected one summary line, got %d: %+v", len(lines), observer.events)
	}
	line := lines[0]
	if strings.Contains(line, "中间过程") {
		t.Fatalf("summary should not include detail action step, got %q", line)
	}
	if !strings.Contains(line, "Alice 使用攻击【火焰斩】 -> Bob") {
		t.Fatalf("summary should include action title, got %q", line)
	}
	if !strings.Contains(line, "Bob 受到2点伤害") {
		t.Fatalf("summary should include structured damage, got %q", line)
	}
}

func TestActionSummaryAggregatesStructuredResources(t *testing.T) {
	game := newActionSummaryTestGame(nil)

	game.BeginActionSummary("skill", "p1", "测试技能", []string{"p2"})
	game.recordSkillUsage("p2", "响应测试", model.SkillTypeResponse)
	game.addActionDraw("p1", 2)
	game.addActionDiscard("p1", 1)
	game.addActionHeal("p2", 1)

	line := game.actionSummaryMessage()
	for _, want := range []string{
		"Alice 发动技能【测试技能】 -> Bob",
		"Bob 响应技能【响应测试】",
		"Alice 摸2张牌",
		"Alice 弃1张牌",
		"Bob 获得1点治疗",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("summary %q missing %q", line, want)
		}
	}
}

func TestBeginActionSummaryClearsUnfinalizedPreviousSummary(t *testing.T) {
	game := newActionSummaryTestGame(nil)

	game.BeginActionSummary("skill", "p1", "旧技能", nil)
	game.NotifyActionSummaryNote("旧摘要")
	game.BeginActionSummary("magic", "p2", "魔弹", []string{"p1"})

	line := game.actionSummaryMessage()
	if strings.Contains(line, "旧技能") || strings.Contains(line, "旧摘要") {
		t.Fatalf("new summary should not reuse old active summary, got %q", line)
	}
	if !strings.Contains(line, "Bob 使用法术【魔弹】 -> Alice") {
		t.Fatalf("expected new action summary, got %q", line)
	}
}
