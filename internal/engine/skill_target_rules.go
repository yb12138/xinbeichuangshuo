// gameflow: 技能目标合法性规则辅助。

package engine

import (
	"fmt"
	"starcup-engine/internal/model"
)

type targetCampRule string

const (
	targetCampAny   targetCampRule = "any"
	targetCampAlly  targetCampRule = "ally"
	targetCampEnemy targetCampRule = "enemy"
)

type targetSelfRule string

const (
	targetSelfAny   targetSelfRule = "any"
	targetSelfOnly  targetSelfRule = "self"
	targetSelfOther targetSelfRule = "other"
)

type targetCheckKind string

const (
	targetCheckHasBasicFieldOnTarget targetCheckKind = "has_basic_field_on_target"
	targetCheckTargetMinHeal         targetCheckKind = "target_min_heal"
	targetCheckAnyBasicFieldWhenNone targetCheckKind = "any_basic_field_when_none"
)

type targetCountRule struct {
	Min int
	Max int
	Err string
}

type targetSlotRule struct {
	Index int
	Camp  targetCampRule
	Self  targetSelfRule
	Err   string
}

type targetCheckRule struct {
	Kind           targetCheckKind
	Index          int
	Min            int
	Err            string
	WithTargetName bool
}

type targetRuleSet struct {
	Count       targetCountRule
	Distinct    bool
	DistinctErr string
	Slots       []targetSlotRule
	Checks      []targetCheckRule
}

func (rules targetRuleSet) hasCountOverride() bool {
	return rules.Count.Min > 0 || rules.Count.Max > 0 || rules.Count.Err != ""
}

func (rules targetRuleSet) validate(use *skillUseRequest) error {
	if use == nil {
		return fmt.Errorf("技能目标校验上下文缺失")
	}
	if err := rules.validateDistinct(use); err != nil {
		return err
	}
	if err := rules.validateSlots(use); err != nil {
		return err
	}
	return rules.validateChecks(use)
}

func (rules targetRuleSet) validateDistinct(use *skillUseRequest) error {
	if !rules.Distinct || len(use.actualTargets) <= 1 {
		return nil
	}
	seen := make(map[string]struct{}, len(use.actualTargets))
	for _, target := range use.actualTargets {
		if target == nil {
			continue
		}
		if _, ok := seen[target.ID]; ok {
			if rules.DistinctErr != "" {
				return fmt.Errorf(rules.DistinctErr)
			}
			return fmt.Errorf("不能重复选择同一角色")
		}
		seen[target.ID] = struct{}{}
	}
	return nil
}

func (rules targetRuleSet) validateSlots(use *skillUseRequest) error {
	for _, rule := range rules.Slots {
		if rule.Index < 0 || rule.Index >= len(use.actualTargets) {
			continue
		}
		target := use.actualTargets[rule.Index]
		if target == nil {
			return fmt.Errorf("目标无效")
		}
		if !matchTargetCamp(use.player, target, rule.Camp) || !matchTargetSelf(use.player, target, rule.Self) {
			if rule.Err != "" {
				return fmt.Errorf(rule.Err)
			}
			return fmt.Errorf("目标不符合要求")
		}
	}
	return nil
}

func (rules targetRuleSet) validateChecks(use *skillUseRequest) error {
	for _, check := range rules.Checks {
		switch check.Kind {
		case targetCheckHasBasicFieldOnTarget:
			if check.Index < 0 || check.Index >= len(use.actualTargets) {
				continue
			}
			target := use.actualTargets[check.Index]
			if target == nil {
				return fmt.Errorf("目标无效")
			}
			if hasBasicFieldEffect(target) {
				continue
			}
			if check.WithTargetName {
				return fmt.Errorf(check.Err, target.Name)
			}
			return fmt.Errorf(check.Err)
		case targetCheckTargetMinHeal:
			if check.Index < 0 || check.Index >= len(use.actualTargets) {
				continue
			}
			target := use.actualTargets[check.Index]
			if target == nil {
				return fmt.Errorf("目标无效")
			}
			if target.Heal < check.Min {
				return fmt.Errorf(check.Err)
			}
		case targetCheckAnyBasicFieldWhenNone:
			if len(use.actualTargets) > 0 {
				continue
			}
			if use.engine == nil {
				return fmt.Errorf("封印破碎缺少引擎上下文")
			}
			for _, player := range use.engine.GetAllPlayers() {
				if hasBasicFieldEffect(player) {
					return nil
				}
			}
			return fmt.Errorf(check.Err)
		default:
			return fmt.Errorf("未知目标校验规则: %s", check.Kind)
		}
	}
	return nil
}

func matchTargetCamp(user *model.Player, target *model.Player, rule targetCampRule) bool {
	switch rule {
	case "", targetCampAny:
		return true
	case targetCampAlly:
		return user != nil && target != nil && target.Camp == user.Camp
	case targetCampEnemy:
		return user != nil && target != nil && target.Camp != user.Camp
	default:
		return false
	}
}

func matchTargetSelf(user *model.Player, target *model.Player, rule targetSelfRule) bool {
	switch rule {
	case "", targetSelfAny:
		return true
	case targetSelfOnly:
		return user != nil && target != nil && target.ID == user.ID
	case targetSelfOther:
		return user != nil && target != nil && target.ID != user.ID
	default:
		return false
	}
}

func effectiveTargetCountRange(use *skillUseRequest) (min int, max int, customErr string) {
	if use == nil || use.skillDef == nil {
		return 0, 0, ""
	}
	if use.skillDef.TargetType == model.TargetNone {
		return 0, 0, ""
	}
	min = use.skillDef.MinTargets
	if min <= 0 {
		min = 1
	}
	max = use.skillDef.MaxTargets
	if use.policy.targetRules.hasCountOverride() {
		min = use.policy.targetRules.Count.Min
		max = use.policy.targetRules.Count.Max
		customErr = use.policy.targetRules.Count.Err
	}
	return min, max, customErr
}

func formatTargetCountError(selected, min, max int, customErr string) error {
	if customErr != "" {
		return fmt.Errorf(customErr)
	}
	if min > 0 && selected == 0 {
		return fmt.Errorf("skill requires target(s)")
	}
	if max > 0 && selected > max {
		return fmt.Errorf("技能最多只能指定 %d 个目标，你指定了 %d 个", max, selected)
	}
	if min > 0 && selected < min {
		return fmt.Errorf("技能最少需要指定 %d 个目标，你指定了 %d 个", min, selected)
	}
	return fmt.Errorf("目标数量不合法")
}

func validateTargetTypeConstraints(use *skillUseRequest) error {
	if use == nil || use.skillDef == nil {
		return fmt.Errorf("技能目标校验上下文缺失")
	}
	for _, target := range use.actualTargets {
		switch use.skillDef.TargetType {
		case model.TargetSelf:
			if target.ID != use.player.ID {
				return fmt.Errorf("skill can only target self")
			}
		case model.TargetEnemy:
			if target.Camp == use.player.Camp {
				return fmt.Errorf("skill can only target enemies")
			}
		case model.TargetAlly:
			if target.Camp != use.player.Camp {
				return fmt.Errorf("skill can only target allies")
			}
		case model.TargetAllySelf:
			if target.Camp != use.player.Camp {
				return fmt.Errorf("skill can only target allies or self")
			}
		case model.TargetAny:
			// no-op
		default:
			// TargetSpecific 等交由 targetRules 继续约束。
		}
	}
	return nil
}
