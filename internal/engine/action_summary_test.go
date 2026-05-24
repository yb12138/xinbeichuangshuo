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
	game.State.Players["p1"] = &model.Player{ID: "p1", Name: "Alice", Camp: model.RedCamp}
	game.State.Players["p2"] = &model.Player{ID: "p2", Name: "Bob", Camp: model.BlueCamp}
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
	game.addActionMoraleLoss(model.RedCamp, 1)
	game.addActionHealUse("p2", 1)
	game.addActionHeal("p2", 1)

	line := game.actionSummaryMessage()
	for _, want := range []string{
		"Alice 发动技能【测试技能】 -> Bob",
		"Bob 响应技能【响应测试】",
		"Alice 摸2张牌",
		"Alice 弃1张牌",
		"红方士气-1",
		"Bob 使用1点治疗抵挡伤害",
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

func TestDamageDrawMoraleLossEntersActionSummary(t *testing.T) {
	game := newActionSummaryTestGame(nil)
	game.State.BlueMorale = 15

	game.BeginActionSummary("attack", "p1", "火焰斩", []string{"p2"})
	game.ApplyMoraleLossAfterTimingWindow(game.State.Players["p2"], 1, false, true, 0, nil, nil)

	line := game.actionSummaryMessage()
	if !strings.Contains(line, "蓝方士气-1") {
		t.Fatalf("summary should include damage-draw morale loss, got %q", line)
	}
}

func TestNonDamageDrawMoraleLossDoesNotEnterActionSummary(t *testing.T) {
	game := newActionSummaryTestGame(nil)
	game.State.BlueMorale = 15

	game.BeginActionSummary("skill", "p1", "测试技能", []string{"p2"})
	game.ApplyMoraleLossAfterTimingWindow(game.State.Players["p2"], 1, false, false, 0, nil, nil)

	line := game.actionSummaryMessage()
	if strings.Contains(line, "蓝方士气-1") {
		t.Fatalf("summary should not include non-damage-draw morale loss, got %q", line)
	}
}

func TestActionSummaryRecordsResourceCostsAndGains(t *testing.T) {
	game := newActionSummaryTestGame(nil)
	alice := game.State.Players["p1"]
	alice.Gem = 2
	alice.Crystal = 1
	alice.Tokens = map[string]int{"hero_anger": 1}
	game.State.RedGems = 1
	game.State.BlueCrystals = 2

	game.BeginActionSummary("skill", "p1", "资源技能", nil)
	alice.Gem--
	alice.Crystal++
	alice.Tokens["hero_anger"] += 2
	game.State.RedGems++
	game.State.BlueCrystals--

	line := game.actionSummaryMessage()
	for _, want := range []string{
		"Alice 消耗1红宝石",
		"蓝方战绩区-1蓝水晶",
		"Alice 获得1蓝水晶",
		"Alice 获得2个怒气指示物",
		"红方战绩区+1红宝石",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("summary %q missing %q", line, want)
		}
	}
}

func TestTokenDisplayNameFallsBackToRawKey(t *testing.T) {
	if got := tokenDisplayName("hero_anger"); got != "怒气" {
		t.Fatalf("expected mapped token display name, got %q", got)
	}
	if got := tokenDisplayName("unknown_token"); got != "unknown_token" {
		t.Fatalf("expected unknown token to fall back to raw key, got %q", got)
	}
}

func TestActionSummaryKeepsSeparateResourceCostAndGainCheckpoints(t *testing.T) {
	game := newActionSummaryTestGame(nil)
	alice := game.State.Players["p1"]
	alice.Gem = 1

	game.BeginActionSummary("skill", "p1", "资源返还", nil)
	alice.Gem--
	game.recordActionResourceDelta()
	alice.Gem++

	line := game.actionSummaryMessage()
	for _, want := range []string{
		"Alice 消耗1红宝石",
		"Alice 获得1红宝石",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("summary %q missing %q", line, want)
		}
	}
}
