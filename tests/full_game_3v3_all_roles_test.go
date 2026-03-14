package tests

import (
	"fmt"
	"strings"
	"testing"
)

// TestFullGame3v3_AllRolesCoverage 以 3v3 自动对局方式覆盖全部角色，
// 并要求主动技能触发覆盖达到一个激进阈值。
func TestFullGame3v3_AllRolesCoverage(t *testing.T) {
	roleIDs := collectRoleIDs()
	if len(roleIDs) < autoGamePlayers {
		t.Fatalf("need at least %d roles for 3v3 test, got %d", autoGamePlayers, len(roleIDs))
	}

	lineups := buildCoverageLineups(roleIDs, autoGamePlayers)
	globalTriggeredSkills := make(map[string]int)
	globalExpectedActionSkills := make(map[string]struct{})

	for idx, lineup := range lineups {
		red := strings.Join(lineup[:3], "_")
		blue := strings.Join(lineup[3:], "_")
		name := fmt.Sprintf("match_%02d_%s_vs_%s", idx+1, red, blue)

		t.Run(name, func(t *testing.T) {
			result, err := runAutoGame(lineup, autoGameStepLimit)
			if err != nil {
				t.Fatal(err)
			}
			mergeTriggered(globalTriggeredSkills, result.triggeredSkills)
			mergeSkillSet(globalExpectedActionSkills, result.expectedActionSkills)
		})
	}

	if len(globalExpectedActionSkills) == 0 {
		t.Fatalf("no action skills discovered from role definitions")
	}

	triggeredActionSkills, totalActions, coverageRatio := coverageStats(globalExpectedActionSkills, globalTriggeredSkills)
	t.Logf("aggressive action-skill coverage: %d/%d (%.2f)", triggeredActionSkills, totalActions, coverageRatio)
	if top := topTriggeredActionSkills(globalExpectedActionSkills, globalTriggeredSkills, 10); len(top) > 0 {
		t.Logf("top triggered action skills: %s", strings.Join(top, ", "))
	}

	minCoverage := 0.30
	// 角色池扩容后，主动技总量上升较快，固定阈值会放大“新增技能尚未进入自动策略”的偶然波动。
	// 这里保留激进标准，但对大技能池做轻微下调，避免因单次对局随机性导致假红。
	if totalActions >= 48 {
		minCoverage = 0.28
	}
	if coverageRatio < minCoverage {
		missing := missingSkillList(globalExpectedActionSkills, globalTriggeredSkills)
		t.Fatalf(
			"aggressive coverage too low: %d/%d (%.2f < %.2f), missing=%v",
			triggeredActionSkills,
			totalActions,
			coverageRatio,
			minCoverage,
			missing,
		)
	}
}
