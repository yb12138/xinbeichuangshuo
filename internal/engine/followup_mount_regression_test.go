package engine

import (
	"testing"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func TestMountPlayerDeferredFollowupSpecs_MountsAndExecutes(t *testing.T) {
	orig := roleRegistry
	t.Cleanup(func() {
		roleRegistry = orig
	})

	called := false
	reg := engineplayer.NewRoleRegistry()
	reg.Register(engineplayer.RoleEntry{
		ID: "test_role",
		FollowupSpecs: map[string]engineplayer.FollowupSpec{
			"test_role_followup": {
				Label: "RoleFollowupTest",
				Resolve: func(rt engineplayer.ChoiceRuntime, f model.DeferredFollowup) error {
					called = true
					if rt.LookupPlayer(f.UserID) == nil {
						t.Fatalf("expected followup runtime to resolve player %s", f.UserID)
					}
					rt.Log("test role followup executed")
					return nil
				},
			},
		},
	})
	roleRegistry = reg

	table := map[string]deferredFollowupHandler{}
	mountPlayerDeferredFollowupSpecs(table)

	handler, ok := table["test_role_followup"]
	if !ok {
		t.Fatalf("expected test_role_followup mounted from role entry")
	}
	if handler.label != "RoleFollowupTest" {
		t.Fatalf("expected mounted label RoleFollowupTest, got %q", handler.label)
	}

	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Tester", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := handler.resolve(game, model.DeferredFollowup{Type: "test_role_followup", UserID: "p1"}); err != nil {
		t.Fatalf("expected mounted followup resolve success, got %v", err)
	}
	if !called {
		t.Fatalf("expected mounted followup resolver to be called")
	}
}

func TestMountPlayerDeferredFollowupSpecs_SkipsEmptyResolver(t *testing.T) {
	orig := roleRegistry
	t.Cleanup(func() {
		roleRegistry = orig
	})

	reg := engineplayer.NewRoleRegistry()
	reg.Register(engineplayer.RoleEntry{
		ID: "test_role",
		FollowupSpecs: map[string]engineplayer.FollowupSpec{
			"ignored_followup": {},
		},
	})
	roleRegistry = reg

	table := map[string]deferredFollowupHandler{}
	mountPlayerDeferredFollowupSpecs(table)

	if _, ok := table["ignored_followup"]; ok {
		t.Fatalf("expected followup with empty resolver to be skipped")
	}
}

func TestSpiritCasterFollowupMountedFromRoleEntry(t *testing.T) {
	// 灵符师的 followup 已迁移为框架自动的 skill_effect_resume；
	// 验证 spirit_caster_talisman 不再挂载（角色不再注册 FollowupSpecs）。
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "SpiritCaster", "spirit_caster", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	handled, _, _ := game.resolveDeferredFollowup(model.DeferredFollowup{
		Type:      "spirit_caster_talisman",
		UserID:    "p1",
		SkillID:   "sc_talisman_thunder",
		TargetIDs: []string{"p2", "p3"},
	})
	if handled {
		t.Fatalf("expected spirit_caster_talisman followup to no longer be mounted after migration")
	}
}
