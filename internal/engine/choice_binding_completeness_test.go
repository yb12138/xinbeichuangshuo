package engine

import (
	"strings"
	"testing"
)

func choiceCatalogTypesFromFile(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(choiceTypeCatalogFile), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

func TestChoiceCatalogArchetypeTableMatchesCatalogFile(t *testing.T) {
	types := choiceCatalogTypesFromFile(t)
	seen := make(map[string]bool, len(types))
	for _, typ := range types {
		seen[typ] = true
		arch, ok := catalogChoiceArchetypeTable[typ]
		if !ok {
			t.Fatalf("catalog type %q missing from catalogChoiceArchetypeTable", typ)
		}
		if arch == "" {
			t.Fatalf("catalog type %q has empty archetype", typ)
		}
		if arch == "hero_assassin" || arch == "guardian" {
			t.Fatalf("catalog type %q still uses deprecated mixed-role archetype %q", typ, arch)
		}
		p := catalogChoiceBinding(typ)
		if p.build == nil || p.sel == nil {
			t.Fatalf("catalogChoiceBinding(%q) incomplete plan", typ)
		}
	}
	for typ := range catalogChoiceArchetypeTable {
		if !seen[typ] {
			t.Fatalf("catalogChoiceArchetypeTable has extra type %q not in choice_type_catalog.txt", typ)
		}
	}
	if len(catalogChoiceArchetypeTable) != len(types) {
		t.Fatalf("map size %d != catalog lines %d", len(catalogChoiceArchetypeTable), len(types))
	}
}
