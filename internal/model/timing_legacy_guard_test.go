package model

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLegacyTimingUsageStaysQuarantined(t *testing.T) {
	legacyTokens := []string{
		"TimingOnAttackDeclared",
		"TimingOnHitCheck",
		"TimingOnDamageCalculated",
	}
	allowedFiles := map[string]bool{
		"internal/data/characters.go":                                   true,
		"internal/model/context_timing.go":                              true,
		"internal/model/rulebook_timing.go":                             true,
		"internal/model/stage_timing.go":                                true,
		"internal/model/timing_registry.go":                             true,
		"internal/engine/attack_lifecycle.go":                           true,
		"internal/engine/attack_role_hooks.go":                          true,
		"internal/engine/before_action_phase_runtime.go":                true,
		"internal/engine/combat.go":                                     true,
		"internal/engine/combat_response_runtime.go":                    true,
		"internal/engine/damage_role_runtime_hooks.go":                  true,
		"internal/engine/debug_cheat_runtime.go":                        true,
		"internal/engine/game.go":                                       true,
		"internal/engine/game_damage_pipeline.go":                       true,
		"internal/engine/game_player_lifecycle.go":                      true,
		"internal/engine/interrupt_runtime.go":                          true,
		"internal/engine/magic.go":                                      true,
		"internal/engine/pending_damage_runtime.go":                     true,
		"internal/engine/role_choice_runtime.go":                        true,
		"internal/engine/runtime_policy_hooks.go":                       true,
		"internal/engine/skill_runtime_host.go":                         true,
		"internal/engine/timing_attack_declared_registry.go":            true,
		"internal/engine/turn_fsm_dispatcher.go":                        true,
		"internal/engine/player/timing_hook.go":                         true,
		"internal/engine/player/adventurer/skill_handlers.go":           true,
		"internal/engine/player/archer/skill_handlers.go":               true,
		"internal/engine/player/beast_samurai/skill_handlers.go":        true,
		"internal/engine/player/berserker/module.go":                    true,
		"internal/engine/player/berserker/skill_handlers.go":            true,
		"internal/engine/player/blade_master/skill_handlers.go":         true,
		"internal/engine/player/crimson_knight/skill_handlers.go":       true,
		"internal/engine/player/crimson_sword_spirit/skill_handlers.go": true,
		"internal/engine/player/elf_archer/skill_handlers.go":           true,
		"internal/engine/player/fighter/module.go":                      true,
		"internal/engine/player/fighter/skill_handlers.go":              true,
		"internal/engine/player/hero/module.go":                         true,
		"internal/engine/player/hero/skill_handlers.go":                 true,
		"internal/engine/player/holy_bow/skill_handlers.go":             true,
		"internal/engine/player/holy_lancer/skill_handlers.go":          true,
		"internal/engine/player/magic_bow/skill_handlers.go":            true,
		"internal/engine/player/magic_lancer/skill_handlers.go":         true,
		"internal/engine/player/magic_swordsman/skill_handlers.go":      true,
		"internal/engine/player/moon_goddess/module.go":                 true,
		"internal/engine/player/moon_goddess/skill_handlers.go":         true,
		"internal/engine/player/prayer_master/skill_handlers.go":        true,
		"internal/engine/player/soul_sorcerer/module.go":                true,
		"internal/engine/player/soul_sorcerer/skill_handlers.go":        true,
		"internal/engine/player/spirit_caster/skill_handlers.go":        true,
		"internal/engine/player/sword_emperor/skill_handlers.go":        true,
		"internal/engine/player/valkyrie/skill_handlers.go":             true,
		"internal/engine/player/war_homunculus/skill_handlers.go":       true,
	}

	repoRoot := testRepoRoot(t)
	var violations []string
	err := filepath.WalkDir(filepath.Join(repoRoot, "internal"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		text := string(body)
		hasLegacy := false
		for _, token := range legacyTokens {
			if strings.Contains(text, token) {
				hasLegacy = true
				break
			}
		}
		if !hasLegacy {
			return nil
		}
		rel, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if !allowedFiles[rel] {
			violations = append(violations, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) > 0 {
		t.Fatalf("legacy timing use outside quarantine allowlist: %s", strings.Join(violations, ", "))
	}
}

func testRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
