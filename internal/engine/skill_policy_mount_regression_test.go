package engine

import (
	"testing"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/types"
)

func TestSkillPolicyMountedFromRoleEntriesForPilotRoles(t *testing.T) {
	saintHeal := resolveSkillUsePolicy("saint_heal")
	if !saintHeal.SkipAutoPhaseEnd {
		t.Fatalf("expected saint_heal SkipAutoPhaseEnd from player role entry")
	}

	// 灵符师的 AfterConsume 已迁移为框架自动的技能恢复状态；
	// sc_talisman_thunder 和 sc_talisman_wind 不再有 AfterConsume policy。
	thunder := resolveSkillUsePolicy("sc_talisman_thunder")
	if thunder.AfterConsume != nil {
		t.Fatalf("expected sc_talisman_thunder AfterConsume to be nil after migration to skill_resume")
	}
	wind := resolveSkillUsePolicy("sc_talisman_wind")
	if wind.AfterConsume != nil {
		t.Fatalf("expected sc_talisman_wind AfterConsume to be nil after migration to skill_resume")
	}
}

func TestMountPlayerSkillPolicySpecsRoleEntryOverridesLegacy(t *testing.T) {
	orig := roleRegistry
	t.Cleanup(func() {
		roleRegistry = orig
	})

	reg := engineplayer.NewRoleRegistry()
	reg.Register(engineplayer.RoleEntry{
		ID: "test_role",
		Skills: []engineplayer.SkillEntry{
			{
				ID: "test_skill",
				Policy: types.SkillPolicy{
					SkipAutoPhaseEnd: true,
				},
			},
		},
	})
	roleRegistry = reg

	table := map[string]SkillPolicy{
		"test_skill": {
			ManualExclusiveCard: true,
		},
	}
	mountPlayerSkillPolicySpecs(table)

	got := table["test_skill"]
	if !got.SkipAutoPhaseEnd {
		t.Fatalf("expected mounted policy to enable SkipAutoPhaseEnd")
	}
	if got.ManualExclusiveCard {
		t.Fatalf("expected role-entry policy to override legacy policy instead of merging stale fields")
	}
}
