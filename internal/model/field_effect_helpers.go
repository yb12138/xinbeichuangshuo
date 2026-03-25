package model

// FieldMetaBoundElement 是场牌Meta中用于存储"绑定元素"的键名
// 五系封印会记录它绑定的是哪个元素（水/火/地/风/雷）
const FieldMetaBoundElement = "bound_element"

// ==========================================
// 五系封印辅助函数
// ==========================================

// IsElementalSealEffect 判断是否是五系封印效果
// 五系封印包括：水之封印、火之封印、地之封印、风之封印、雷之封印
func IsElementalSealEffect(effect EffectType) bool {
	switch effect {
	case EffectSealWater, EffectSealFire, EffectSealEarth, EffectSealWind, EffectSealThunder:
		return true
	default:
		return false
	}
}

// ElementalSealEffectElement 获取五系封印效果对应的默认元素
// 例如：EffectSealWater -> ElementWater
func ElementalSealEffectElement(effect EffectType) Element {
	switch effect {
	case EffectSealWater:
		return ElementWater
	case EffectSealFire:
		return ElementFire
	case EffectSealEarth:
		return ElementEarth
	case EffectSealWind:
		return ElementWind
	case EffectSealThunder:
		return ElementThunder
	default:
		return ""
	}
}

// BoundElementForFieldCard 获取场牌（封印）绑定的元素
// 查找优先级：
//  1. 从 Meta[FieldMetaBoundElement] 中读取（最优先，可自定义）
//  2. 根据 EffectType 推断（例如 EffectSealWater 就是水）
//  3. 从放置的卡牌元素中推断（兜底）
func BoundElementForFieldCard(fc *FieldCard) Element {
	if fc == nil {
		return ""
	}
	// 优先级1：从Meta中读取绑定元素（支持自定义绑定）
	if fc.Meta != nil {
		if raw := fc.Meta[FieldMetaBoundElement]; raw != "" {
			return Element(raw)
		}
	}
	// 优先级2：根据封印效果类型推断默认元素
	if bound := ElementalSealEffectElement(fc.Effect); bound != "" {
		return bound
	}
	// 优先级3：从放置封印时使用的卡牌元素推断
	switch fc.Card.Element {
	case ElementWater, ElementFire, ElementEarth, ElementWind, ElementThunder:
		return fc.Card.Element
	default:
		return ""
	}
}
