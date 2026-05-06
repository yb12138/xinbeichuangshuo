// gameflow: 形态（Form）管理基础设施。

package player

import "starcup-engine/internal/model"

// EffectiveOrientation 读取有效朝向。
func EffectiveOrientation(p *model.Player) model.CharacterOrientation {
	if p == nil {
		return model.OrientationNormal
	}
	if p.Orientation != "" {
		return p.Orientation
	}
	return model.OrientationNormal
}

// EffectiveForm 读取有效形态。
func EffectiveForm(p *model.Player) string {
	if p == nil {
		return ""
	}
	return p.Form
}

// HasForm 判断是否处于指定形态。
func HasForm(p *model.Player, form string) bool {
	return p != nil && EffectiveForm(p) == form
}

// SetForm 进入形态（横置 + 设置形态名），返回是否有变化。
func SetForm(p *model.Player, form string) bool {
	if p == nil {
		return false
	}
	changed := EffectiveOrientation(p) != model.OrientationTapped || EffectiveForm(p) != form
	p.Orientation = model.OrientationTapped
	p.Form = form
	return changed
}

// ClearForm 退出形态（恢复直立 + 清空形态名），返回是否有变化。
func ClearForm(p *model.Player, form string) bool {
	if p == nil {
		return false
	}
	if form != "" && EffectiveForm(p) != form {
		return false
	}
	changed := EffectiveOrientation(p) != model.OrientationNormal || EffectiveForm(p) != ""
	p.Orientation = model.OrientationNormal
	p.Form = ""
	return changed
}
