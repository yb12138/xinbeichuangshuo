package model_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestRemovedCoarseTimingNamesStayDeleted(t *testing.T) {
	legacyTokens := []string{
		"Timing" + "OnAttackDeclared",
		"Timing" + "OnHitCheck",
		"Timing" + "OnDamageCalculated",
		"Timing" + "OnDamageTaken",
	}
	allowedCharacterLegacyTimingSkills := map[string][]string{}

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
		violations = append(violations, rel)
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
