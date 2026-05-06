// gameflow: 技能目标校验规则类型定义（共享包）。

package types

import "fmt"

// TargetCampRule 定义目标阵营约束。
type TargetCampRule string

const (
	TargetCampAny   TargetCampRule = "any"
	TargetCampAlly  TargetCampRule = "ally"
	TargetCampEnemy TargetCampRule = "enemy"
)

// TargetSelfRule 定义目标自身关系约束。
type TargetSelfRule string

const (
	TargetSelfAny   TargetSelfRule = "any"
	TargetSelfOnly  TargetSelfRule = "self"
	TargetSelfOther TargetSelfRule = "other"
)

// TargetCheckKind 定义目标检查类型。
type TargetCheckKind string

const (
	TargetCheckHasBasicFieldOnTarget TargetCheckKind = "has_basic_field_on_target"
	TargetCheckTargetMinHeal         TargetCheckKind = "target_min_heal"
	TargetCheckAnyBasicFieldWhenNone TargetCheckKind = "any_basic_field_when_none"
)

// TargetCountRule 定义目标数量约束。
type TargetCountRule struct {
	Min int
	Max int
	Err string
}

// TargetSlotRule 定义目标槽位约束。
type TargetSlotRule struct {
	Index int
	Camp  TargetCampRule
	Self  TargetSelfRule
	Err   string
}

// TargetCheckRule 定义目标检查谓词。
type TargetCheckRule struct {
	Kind           TargetCheckKind
	Index          int
	Min            int
	Err            string
	WithTargetName bool
}

// TargetRuleSet 定义目标校验规则集合。
type TargetRuleSet struct {
	Count       TargetCountRule
	Distinct    bool
	DistinctErr string
	Slots       []TargetSlotRule
	Checks      []TargetCheckRule
}

// HasCountOverride 检查是否有自定义目标数量覆盖。
func (rules TargetRuleSet) HasCountOverride() bool {
	return rules.Count.Min > 0 || rules.Count.Max > 0 || rules.Count.Err != ""
}

// ValidateDistinct 校验目标是否不允许重复。
func (rules TargetRuleSet) ValidateDistinct(targetIDs []string) error {
	if !rules.Distinct || len(targetIDs) <= 1 {
		return nil
	}
	seen := make(map[string]struct{}, len(targetIDs))
	for _, id := range targetIDs {
		if _, ok := seen[id]; ok {
			if rules.DistinctErr != "" {
				return fmt.Errorf(rules.DistinctErr)
			}
			return fmt.Errorf("不能重复选择同一角色")
		}
		seen[id] = struct{}{}
	}
	return nil
}

// ValidateSlotCount 检查是否有槽位规则。
func (rules TargetRuleSet) HasSlots() bool {
	return len(rules.Slots) > 0
}

// ValidateSlot 校验单个槽位约束。
func (rule TargetSlotRule) ValidateSlot(userCamp, targetCamp string, userID, targetID string) error {
	if rule.Camp != "" && rule.Camp != TargetCampAny {
		if rule.Camp == TargetCampAlly && userCamp != targetCamp {
			if rule.Err != "" {
				return fmt.Errorf(rule.Err)
			}
			return fmt.Errorf("目标不符合阵营要求")
		}
		if rule.Camp == TargetCampEnemy && userCamp == targetCamp {
			if rule.Err != "" {
				return fmt.Errorf(rule.Err)
			}
			return fmt.Errorf("目标不符合阵营要求")
		}
	}
	if rule.Self != "" && rule.Self != TargetSelfAny {
		if rule.Self == TargetSelfOnly && userID != targetID {
			if rule.Err != "" {
				return fmt.Errorf(rule.Err)
			}
			return fmt.Errorf("目标不符合要求")
		}
		if rule.Self == TargetSelfOther && userID == targetID {
			if rule.Err != "" {
				return fmt.Errorf(rule.Err)
			}
			return fmt.Errorf("目标不符合要求")
		}
	}
	return nil
}
