// gameflow: 自 choice_type_catalog.txt 批量注册 ChoiceSpec（与 specials 互补）。

package engine

import (
	_ "embed"
	"fmt"
	"strings"

	choicert "starcup-engine/internal/engine/runtime/choice"
	"starcup-engine/internal/model"
)

//go:embed choice_type_catalog.txt
var choiceTypeCatalogFile string

type catalogSpecPlan struct {
	build func(*GameEngine, string, string, *model.Player, map[string]any) *model.Prompt
	sel   func(*GameEngine, string, int, map[string]any) (bool, error)
	multi func(*GameEngine, string, []int) error
}

func multiSequential(typ string) func(*GameEngine, string, []int) error {
	if !choicert.HasSequentialMulti(typ) {
		return nil
	}
	return func(ge *GameEngine, pid string, sels []int) error {
		return ge.runSequentialChoiceSelections(pid, typ, sels)
	}
}

func registerCatalogPlan(reg *choicert.SpecRegistry, typ string, p catalogSpecPlan) {
	if p.build == nil || p.sel == nil {
		panic(fmt.Sprintf("choice catalog: incomplete spec for %q", typ))
	}
	if p.multi != nil {
		regSpecSingleAndMulti(reg, typ, p.build, p.sel, p.multi)
		return
	}
	regSpecSingle(reg, typ, p.build, p.sel)
}

func (e *GameEngine) bootstrapChoiceSpecsFromCatalog(reg *choicert.SpecRegistry) {
	for _, line := range strings.Split(strings.TrimSpace(choiceTypeCatalogFile), "\n") {
		typ := strings.TrimSpace(line)
		if typ == "" || strings.HasPrefix(typ, "#") {
			continue
		}
		if reg.Has(typ) {
			continue
		}
		p := catalogChoiceBinding(typ)
		registerCatalogPlan(reg, typ, p)
	}
}
