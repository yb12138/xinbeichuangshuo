package tests

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

// TestCauseAttributionCoverage 兜底：自动对局中日志出现“发动 [技能名]”的技能，都应有事件带着该技能的因果。
// 新角色 / 新技能若没有声明归属（ChoiceRouteSpec.WithSkill、TimingHookSpec.SkillID 或 rt.WithSkillCause），
// 这里会列出缺失并失败。临时放宽可设 LENIENT_CAUSE_COVERAGE=1（只报告不失败）。
func TestCauseAttributionCoverage(t *testing.T) {
	// 复用回归战役（多轮轮换 + 镜像阵容）与定向场景，尽量多地覆盖低频技能。
	campaign := runFullGameCampaign(t)
	directed := runDirectedScenarioCampaign(t)
	invoked := map[string]int{}
	attributed := map[string]int{}
	mergeSkillCountMaps(invoked, campaign.skillBatch)
	mergeSkillCountMaps(invoked, directed.skillBatch)
	mergeSkillCountMaps(attributed, campaign.attributed)
	mergeSkillCountMaps(attributed, directed.attributed)

	wrong := map[string]int{}
	mergeSkillCountMaps(wrong, campaign.misattrib)
	mergeSkillCountMaps(wrong, directed.misattrib)
	if len(wrong) > 0 {
		var lines []string
		for k, n := range wrong {
			lines = append(lines, fmt.Sprintf("%s(×%d)", k, n))
		}
		sort.Strings(lines)
		t.Logf("归属与日志不一致：%s", strings.Join(lines, "、"))
		if os.Getenv("LENIENT_CAUSE_COVERAGE") != "1" {
			t.Errorf("%d skill logs resolved under a different skill's cause", len(lines))
		}
	}

	var missing []string
	for name, n := range invoked {
		if attributed[name] == 0 {
			missing = append(missing, fmt.Sprintf("%s(×%d)", name, n))
		}
	}
	sort.Strings(missing)
	t.Logf("技能因果覆盖：发动过 %d 个，已归属 %d 个，缺失 %d 个", len(invoked), len(invoked)-len(missing), len(missing))
	if len(missing) > 0 {
		t.Logf("缺失归属：%s", strings.Join(missing, "、"))
		if os.Getenv("LENIENT_CAUSE_COVERAGE") != "1" {
			t.Fatalf("%d skills fired without cause attribution", len(missing))
		}
	}
}
