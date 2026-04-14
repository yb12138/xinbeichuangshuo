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

func TestChoiceCatalogRouteSpecTableMatchesCatalogFile(t *testing.T) {
	types := choiceCatalogTypesFromFile(t)
	seen := make(map[string]bool, len(types))
	for _, typ := range types {
		seen[typ] = true
		spec, ok := catalogChoiceRouteSpecTable[typ]
		if !ok {
			t.Fatalf("catalog type %q missing from catalogChoiceRouteSpecTable", typ)
		}
		if !spec.valid() {
			t.Fatalf("catalog type %q has invalid route spec: %+v", typ, spec)
		}
		p := catalogChoiceBinding(typ)
		if p.build == nil || p.sel == nil {
			t.Fatalf("catalogChoiceBinding(%q) incomplete plan", typ)
		}
	}
	for typ := range catalogChoiceRouteSpecTable {
		if !seen[typ] {
			t.Fatalf("catalogChoiceRouteSpecTable has extra type %q not in choice_type_catalog.txt", typ)
		}
	}
	if len(catalogChoiceRouteSpecTable) != len(types) {
		t.Fatalf("map size %d != catalog lines %d", len(catalogChoiceRouteSpecTable), len(types))
	}
}
