// gameflow: 技能目标合法性规则辅助。

package engine

import (
	"fmt"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// 技能目标规则类型直接复用 types 包定义。
type TargetCampRule = types.TargetCampRule
type TargetSelfRule = types.TargetSelfRule
type TargetCheckKind = types.TargetCheckKind
type TargetCountRule = types.TargetCountRule
type TargetSlotRule = types.TargetSlotRule
type TargetCheckRule = types.TargetCheckRule
type TargetRuleSet = types.TargetRuleSet

// 目标规则常量。
const (
	TargetCampAny                    = types.TargetCampAny
	TargetCampAlly                   = types.TargetCampAlly
	TargetCampEnemy                  = types.TargetCampEnemy
	TargetSelfAny                    = types.TargetSelfAny
	TargetSelfOnly                   = types.TargetSelfOnly
	TargetSelfOther                  = types.TargetSelfOther
	TargetCheckHasBasicFieldOnTarget = types.TargetCheckHasBasicFieldOnTarget
	TargetCheckTargetMinHeal         = types.TargetCheckTargetMinHeal
	TargetCheckAnyBasicFieldWhenNone = types.TargetCheckAnyBasicFieldWhenNone
)

// validateTargetRules 校验技能目标规则。
func validateTargetRules(rules TargetRuleSet, use *skillUseRequest) error {
	if use == nil {
		return fmt.Errorf("技能目标校验上下文缺失")
	}
	if err := validateTargetDistinct(rules, use); err != nil {
		return err
	}
	if err := validateTargetSlots(rules, use); err != nil {
		return err
	}
	return validateTargetChecks(rules, use)
}

func validateTargetDistinct(rules TargetRuleSet, use *skillUseRequest) error {
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

func validateTargetSlots(rules TargetRuleSet, use *skillUseRequest) error {
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

func validateTargetChecks(rules TargetRuleSet, use *skillUseRequest) error {
	for _, check := range rules.Checks {
		switch check.Kind {
		case TargetCheckHasBasicFieldOnTarget:
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
		case TargetCheckTargetMinHeal:
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
		case TargetCheckAnyBasicFieldWhenNone:
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

func matchTargetCamp(user *model.Player, target *model.Player, rule TargetCampRule) bool {
	switch rule {
	case "", TargetCampAny:
		return true
	case TargetCampAlly:
		return user != nil && target != nil && target.Camp == user.Camp
	case TargetCampEnemy:
		return user != nil && target != nil && target.Camp != user.Camp
	default:
		return false
	}
}

func matchTargetSelf(user *model.Player, target *model.Player, rule TargetSelfRule) bool {
	switch rule {
	case "", TargetSelfAny:
		return true
	case TargetSelfOnly:
		return user != nil && target != nil && target.ID == user.ID
	case TargetSelfOther:
		return user != nil && target != nil && target.ID != user.ID
	default:
		return false
	}
}

func hasBasicFieldEffect(player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect {
			continue
		}
		if model.IsBasicEffect(string(fc.Effect)) {
			return true
		}
	}
	return false
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
	if use.policy.TargetRules.HasCountOverride() {
		min = use.policy.TargetRules.Count.Min
		max = use.policy.TargetRules.Count.Max
		customErr = use.policy.TargetRules.Count.Err
	}
	return min, max, customErr
}

func formatTargetCountError(selected, min, max int, customErr string) error {
	// 分步选择模式：由后端流程控制目标，前端不需要预先选择
	if customErr == "分步选择" {
		return nil
	}
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
