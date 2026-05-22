package model

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestLegacyTimingUsageStaysQuarantined(t *testing.T) {
	legacyTokens := []string{
		"TimingOnAttackDeclared",
		"TimingOnHitCheck",
		"TimingOnDamageCalculated",
		"TimingOnDamageTaken",
	}
	allowedFiles := map[string]bool{
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
		"internal/engine/player/butterfly_dancer/module.go":             true,
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
		"internal/engine/player/elf_archer/runtime.go":                  true,
		"internal/engine/player/elf_archer/timing_hooks_post_damage.go": true,
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
	allowedCharacterLegacyTimingSkills := map[string][]string{
		// needs_manual_review: combines source damage calculation and hit-branch damage in one handler.
		"berserker_frenzy": {"TimingOnDamageCalculated", "TimingOnHitCheck"},
		// needs_manual_review: attack-hit optional response still depends on the legacy hit-check resume/response chain.
		"berserker_tear": {"TimingOnHitCheck"},
		// needs_manual_review: attack-hit damage branch still depends on the legacy hit-check resume/response chain.
		"blood_blade": {"TimingOnHitCheck"},
		// needs_manual_review: attack-miss optional response still depends on legacy hit-check resume contexts.
		"piercing_shot": {"TimingOnHitCheck"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"backlash": {"TimingOnDamageTaken"},
		// needs_manual_review: attack-hit optional response still depends on the legacy hit-check resume/response chain.
		"valkyrie_heroic_summon": {"TimingOnHitCheck"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"elementalist_absorb": {"TimingOnDamageTaken"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"arbiter_judgment_tide": {"TimingOnDamageTaken"},
		// needs_manual_review: attack-hit response ordering with earth spear still depends on the legacy hit-check chain.
		"holy_lancer_holy_strike": {"TimingOnHitCheck"},
		// needs_manual_review: attack-hit optional response still depends on the legacy hit-check resume/response chain.
		"holy_lancer_earth_spear": {"TimingOnHitCheck"},
		// needs_manual_review: post-damage helper flows currently construct legacy damage-taken contexts.
		"elf_animal_companion": {"TimingOnDamageTaken"},
		// needs_manual_review: paired with animal companion in legacy post-damage helper flows.
		"elf_pet_empower": {"TimingOnDamageTaken"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"css_blood_barrier": {"TimingOnDamageTaken"},
		// needs_manual_review: attack-hit optional response mutates current damage through the legacy hit-check chain.
		"crk_killing_feast": {"TimingOnHitCheck"},
		// needs_manual_review: attack-miss response group still depends on legacy hit-check resume contexts.
		"hom_rage_suppress": {"TimingOnHitCheck"},
		// needs_manual_review: attack-hit optional response still depends on the legacy hit-check resume/response chain.
		"hom_rune_smash": {"TimingOnHitCheck"},
		// needs_manual_review: attack-miss response group still depends on legacy hit-check resume contexts.
		"hom_glyph_fusion": {"TimingOnHitCheck"},
		// needs_manual_review: damage priority/resume chain still expects the legacy damage-taken timing.
		"hom_dual_echo": {"TimingOnDamageTaken"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"bw_substitute_doll": {"TimingOnDamageTaken"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"bw_mana_inversion": {"TimingOnDamageTaken"},
		// needs_manual_review: defensive/counter interaction is currently driven by onmyoji role hooks, not standard attack declaration collection.
		"onmyoji_yinyang_shift": {"TimingOnAttackDeclared"},
		// needs_manual_review: follows yinyang resolution via an explicit context flag; standard timing collection would be misleading.
		"onmyoji_shikigami_shift": {"TimingOnAttackDeclared"},
		// needs_manual_review: substitute counter-response flow is currently driven by onmyoji role hooks.
		"onmyoji_binding": {"TimingOnAttackDeclared"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"ml_dark_barrier": {"TimingOnDamageTaken"},
		// needs_manual_review: attack-hit optional response still depends on the legacy hit-check resume/response chain.
		"ml_black_spear": {"TimingOnHitCheck"},
		// needs_manual_review: attack-hit optional response still depends on the legacy hit-check resume/response chain.
		"sc_hundred_night": {"TimingOnHitCheck"},
		// needs_manual_review: single skill spans hit and miss outcomes, while its handler still checks the legacy response phase.
		"hero_forbidden_power": {"TimingOnHitCheck"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"hero_dead_duel": {"TimingOnDamageTaken"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"fighter_psi_field": {"TimingOnDamageTaken"},
		// needs_manual_review: attack-miss passive still depends on legacy hit-check resume contexts.
		"se_sword_soul_guard": {"TimingOnHitCheck"},
		// needs_manual_review: attack-miss passive still depends on legacy hit-check resume contexts.
		"se_feint": {"TimingOnHitCheck"},
		// needs_manual_review: attack-hit optional response still depends on the legacy hit-check resume/response chain.
		"se_sword_qi_slash": {"TimingOnHitCheck"},
		// needs_manual_review: attack-hit optional response still depends on the legacy hit-check resume/response chain.
		"bs_reversal_iaijutsu": {"TimingOnHitCheck"},
		// needs_manual_review: damage-taken dispatch/resume still expects the legacy damage-taken timing.
		"bs_beast_return": {"TimingOnDamageTaken"},
		// needs_manual_review: attack-hit optional response still depends on the legacy hit-check resume/response chain.
		"mg_darkmoon_slash": {"TimingOnHitCheck"},
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
		if rel == "internal/data/characters.go" {
			violations = append(violations, legacyCharacterTimingViolations(text, legacyTokens, allowedCharacterLegacyTimingSkills)...)
			return nil
		}
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

func legacyCharacterTimingViolations(text string, legacyTokens []string, allowed map[string][]string) []string {
	var violations []string
	skillTimingPattern := regexp.MustCompile(`ID: "([^"]+)", Timings: \[\]model\.FlowTiming\{([^}]*)\}`)
	stripped := skillTimingPattern.ReplaceAllStringFunc(text, func(decl string) string {
		matches := skillTimingPattern.FindStringSubmatch(decl)
		if len(matches) != 3 {
			return decl
		}
		skillID, timings := matches[1], matches[2]
		used := legacyTokensIn(timings, legacyTokens)
		if len(used) == 0 {
			return decl
		}
		want, ok := allowed[skillID]
		if !ok {
			violations = append(violations, "internal/data/characters.go:"+skillID)
			return ""
		}
		if !sameStringSet(used, want) {
			violations = append(violations, "internal/data/characters.go:"+skillID+" has "+strings.Join(used, "|"))
		}
		return ""
	})
	if len(legacyTokensIn(stripped, legacyTokens)) > 0 {
		violations = append(violations, "internal/data/characters.go:legacy timing outside SkillDefinition.Timings")
	}
	return violations
}

func legacyTokensIn(text string, legacyTokens []string) []string {
	var used []string
	for _, token := range legacyTokens {
		if strings.Contains(text, token) {
			used = append(used, token)
		}
	}
	return used
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, x := range a {
		seen[x]++
	}
	for _, x := range b {
		if seen[x] == 0 {
			return false
		}
		seen[x]--
	}
	return true
}

func testRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
