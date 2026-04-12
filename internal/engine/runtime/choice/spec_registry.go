// gameflow: ChoiceSpec 注册表（重复注册 panic）。

package choice

import "fmt"

// SpecRegistry ChoiceSpec 注册表。
type SpecRegistry struct {
	specs map[string]*ChoiceSpec
}

// NewSpecRegistry 创建注册表。
func NewSpecRegistry() *SpecRegistry {
	return &SpecRegistry{
		specs: make(map[string]*ChoiceSpec),
	}
}

// Register 注册 ChoiceSpec；同一 choice_type 重复注册会 panic。
func (r *SpecRegistry) Register(spec *ChoiceSpec) {
	if r == nil || spec == nil || spec.Type == "" {
		return
	}
	if _, exists := r.specs[spec.Type]; exists {
		panic(fmt.Sprintf("choice: duplicate ChoiceSpec for type %q", spec.Type))
	}
	r.specs[spec.Type] = spec
}

// Get 获取指定类型的 ChoiceSpec。
func (r *SpecRegistry) Get(choiceType string) *ChoiceSpec {
	if r == nil {
		return nil
	}
	return r.specs[choiceType]
}

// Has 是否已注册该 choice_type。
func (r *SpecRegistry) Has(choiceType string) bool {
	if r == nil {
		return false
	}
	_, ok := r.specs[choiceType]
	return ok
}

// ListTypes 列出已注册的 choice_type（用于测试与诊断）。
func (r *SpecRegistry) ListTypes() []string {
	if r == nil {
		return nil
	}
	out := make([]string, 0, len(r.specs))
	for t := range r.specs {
		out = append(out, t)
	}
	return out
}
