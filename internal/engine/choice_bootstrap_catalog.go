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
	build  func(*GameEngine, string, string, *model.Player, map[string]any) *model.Prompt
	sel    func(*GameEngine, string, int, map[string]any) (bool, error)
	multi  func(*GameEngine, string, []int) error
	cancel func(*GameEngine, string, map[string]any) (bool, error)
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
	if p.cancel != nil {
		if p.multi != nil {
			reg.Register(&choicert.ChoiceSpec{
				Type: typ,
				BuildPrompt: func(h choicert.Host, ct, pid string, pl *model.Player, data map[string]any) *model.Prompt {
					ge := engFromHost(h)
					if ge == nil {
						return nil
					}
					return p.build(ge, ct, pid, pl, data)
				},
				OnSelect: func(h choicert.Host, pid string, idx int, ctx map[string]any) (bool, error) {
					ge := engFromHost(h)
					if ge == nil {
						return false, fmt.Errorf("choice: engine bridge unavailable")
					}
					return p.sel(ge, pid, idx, ctx)
				},
				OnMultiSelect: func(h choicert.Host, pid string, sel []int, _ map[string]any) (bool, error) {
					ge := engFromHost(h)
					if ge == nil {
						return false, fmt.Errorf("choice: engine bridge unavailable")
					}
					return true, p.multi(ge, pid, sel)
				},
				OnCancel: func(h choicert.Host, pid string, ctx map[string]any) (bool, error) {
					ge := engFromHost(h)
					if ge == nil {
						return false, fmt.Errorf("choice: engine bridge unavailable")
					}
					return p.cancel(ge, pid, ctx)
				},
			})
			return
		}
		reg.Register(&choicert.ChoiceSpec{
			Type: typ,
			BuildPrompt: func(h choicert.Host, ct, pid string, pl *model.Player, data map[string]any) *model.Prompt {
				ge := engFromHost(h)
				if ge == nil {
					return nil
				}
				return p.build(ge, ct, pid, pl, data)
			},
			OnSelect: func(h choicert.Host, pid string, idx int, ctx map[string]any) (bool, error) {
				ge := engFromHost(h)
				if ge == nil {
					return false, fmt.Errorf("choice: engine bridge unavailable")
				}
				return p.sel(ge, pid, idx, ctx)
			},
			OnCancel: func(h choicert.Host, pid string, ctx map[string]any) (bool, error) {
				ge := engFromHost(h)
				if ge == nil {
					return false, fmt.Errorf("choice: engine bridge unavailable")
				}
				return p.cancel(ge, pid, ctx)
			},
		})
		return
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
