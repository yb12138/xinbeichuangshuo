// gameflow: catalog 中每个 choice_type 显式绑定到 Build/Select（由 catalogChoiceRouteSpecTable 索引，无前缀推断）。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func systemChoiceSelect(typ string) func(*GameEngine, string, int, map[string]any) (bool, error) {
	switch typ {
	case "weak":
		return func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
			return true, ge.handleSystemWeakChoice(pid, idx, choiceCtxAsInterfaceMap(ctx))
		}
	case "buy_resource":
		return func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
			return true, ge.handleSystemBuyResourceChoice(pid, idx, choiceCtxAsInterfaceMap(ctx))
		}
	case "heal":
		return func(ge *GameEngine, _ string, idx int, ctx map[string]any) (bool, error) {
			return true, ge.handleSystemHealChoice(idx, choiceCtxAsInterfaceMap(ctx))
		}
	case "basic_effect_pick":
		return func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
			return true, ge.handleBasicEffectChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
		}
	case choiceTypeSystemDiscardCards:
		return func(ge *GameEngine, pid string, idx int, _ map[string]any) (bool, error) {
			return true, ge.handleSystemDiscardChoiceSelections(pid, []int{idx})
		}
	default:
		panic(fmt.Sprintf("choice: systemChoiceSelect unknown type %q", typ))
	}
}

func systemChoiceMulti(typ string) func(*GameEngine, string, []int) error {
	switch typ {
	case choiceTypeSystemDiscardCards:
		return (*GameEngine).handleSystemDiscardChoiceSelections
	default:
		return nil
	}
}

func systemChoiceCancel(typ string) func(*GameEngine, string, map[string]any) (bool, error) {
	switch typ {
	case choiceTypeSystemDiscardCards:
		return func(ge *GameEngine, pid string, _ map[string]any) (bool, error) {
			if ge.State == nil || ge.State.PendingInterrupt == nil {
				return false, fmt.Errorf("当前没有待处理的弃牌操作")
			}
			ctxData, _ := ge.State.PendingInterrupt.Context.(map[string]interface{})
			return true, ge.cancelSystemDiscardChoice(pid, ctxData)
		}
	default:
		return nil
	}
}

func systemChoiceConsumes(typ string) func(map[string]any) bool {
	switch typ {
	case choiceTypeSystemDiscardCards:
		return func(ctx map[string]any) bool {
			if ctx == nil {
				return false
			}
			skillID, _ := ctx["skill_id"].(string)
			return skillID == ""
		}
	default:
		return nil
	}
}

func catalogChoiceBinding(typ string) catalogSpecPlan {
	spec, ok := catalogChoiceRouteSpecTable[typ]
	if !ok {
		panic(fmt.Sprintf("choice catalog: no route spec row for type %q (sync choice_catalog_route_map.go)", typ))
	}
	if !spec.Valid() {
		panic(fmt.Sprintf("choice catalog: invalid route spec for type %q: %+v", typ, spec))
	}
	switch spec.Kind {
	case ChoiceRouteKindSystem:
		consumesFn := systemChoiceConsumes(typ)
		autoConsume := consumesFn == nil // 只有无自定义 consumes 规则的才自动消费
		return catalogSpecPlan{
			autoConsume: autoConsume,
			consumes:    consumesFn,
			build:       (*GameEngine).buildSystemChoicePrompt,
			sel:         systemChoiceSelect(typ),
			multi:       systemChoiceMulti(typ),
			cancel:      systemChoiceCancel(typ),
			after:       systemChoiceAfterConsume(typ),
		}
	case ChoiceRouteKindRole:
		roleID := spec.Role
		return catalogSpecPlan{
			build: func(ge *GameEngine, choiceType, playerID string, player *model.Player, data map[string]any) *model.Prompt {
				return ge.buildRoleChoicePrompt(roleID, choiceType, playerID, player, choiceCtxAsInterfaceMap(data))
			},
			sel: func(ge *GameEngine, playerID string, idx int, ctx map[string]any) (bool, error) {
				return ge.handleRoleChoiceInput(roleID, playerID, idx, choiceCtxAsInterfaceMap(ctx))
			},
			cancel: func(ge *GameEngine, playerID string, ctx map[string]any) (bool, error) {
				return ge.handleRoleChoiceCancel(roleID, playerID, choiceCtxAsInterfaceMap(ctx))
			},
		}
	default:
		panic(fmt.Sprintf("choice catalog: unknown route kind %q for type %q", spec.Kind, typ))
	}
}

func systemChoiceAfterConsume(typ string) func(*GameEngine, map[string]any) {
	switch typ {
	case "weak":
		return func(ge *GameEngine, ctx map[string]any) { ge.afterSystemWeakChoice(ctx) }
	case "buy_resource":
		return func(ge *GameEngine, _ map[string]any) { ge.enterExtraActionStage() }
	case "heal":
		return func(ge *GameEngine, _ map[string]any) { ge.enterDamageResolution(nil) }
	case "basic_effect_pick":
		return func(ge *GameEngine, ctx map[string]any) { ge.afterBasicEffectChoice(ctx) }
	case choiceTypeSystemDiscardCards:
		return func(ge *GameEngine, ctx map[string]any) { ge.afterSystemDiscardChoice(ctx) }
	default:
		return nil
	}
}
