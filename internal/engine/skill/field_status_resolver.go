// gameflow: 解析场上牌/形态是否满足技能 CanUse 条件。

package skills

import (
	"fmt"

	"starcup-engine/internal/model"
)

// fieldStatusResolverSpec 定义了场牌状态解析器的规范
// 用于统一处理不同类型的场牌效果（如五系封印）
type fieldStatusResolverSpec struct {
	matchesEffect func(effect model.EffectType) bool                  // 判断是否匹配该效果类型
	canUse        func(ctx *model.Context, fc *model.FieldCard) bool  // 检查是否可以触发
	execute       func(ctx *model.Context, fc *model.FieldCard) error // 执行效果逻辑
}

// fieldStatusResolverSpecs 注册所有场牌状态解析器
// 目前包括：五系封印
var fieldStatusResolverSpecs = []fieldStatusResolverSpec{
	{
		matchesEffect: model.IsElementalSealEffect, // 匹配任意五系封印效果
		canUse:        canResolveElementalSealStatus,
		execute:       executeElementalSealStatus,
	},
}

// canResolveFieldStatus 检查指定场牌效果是否可以在当前时机触发
func canResolveFieldStatus(ctx *model.Context, effect model.EffectType) bool {
	spec, fc := resolveFieldStatusSpec(ctx, effect)
	return spec != nil && fc != nil && spec.canUse(ctx, fc)
}

// executeFieldStatus 执行指定场牌效果的逻辑
func executeFieldStatus(ctx *model.Context, effect model.EffectType) error {
	spec, fc := resolveFieldStatusSpec(ctx, effect)
	if spec == nil || fc == nil {
		return nil
	}
	return spec.execute(ctx, fc)
}

// resolveFieldStatusSpec 根据效果类型找到对应的解析器规范和场牌
func resolveFieldStatusSpec(ctx *model.Context, effect model.EffectType) (*fieldStatusResolverSpec, *model.FieldCard) {
	if ctx == nil || ctx.User == nil {
		return nil, nil
	}
	for i := range fieldStatusResolverSpecs {
		spec := &fieldStatusResolverSpecs[i]
		if !spec.matchesEffect(effect) {
			continue
		}
		for _, fc := range ctx.User.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != effect {
				continue
			}
			return spec, fc
		}
	}
	return nil, nil
}

// ==========================================
// 五系封印核心逻辑
// ==========================================
//
// 五系封印包括：水之封印、火之封印、地之封印、风之封印、雷之封印
//
// 触发流程：
// 1. 目标玩家打出/展示对应元素的牌 → TimingOnCardPlayedOrRevealed
// 2. SkillDispatcher 收集触发技能 → collectSkillsForTiming
// 3. 检查封印是否匹配 → canResolveElementalSealStatus
// 4. 执行封印效果 → executeElementalSealStatus
// 5. 添加PendingDamage（带EffectTypeToRemove标记）
// 6. 伤害应用后由封印结算钩子移除封印（pendingDamageElementalSealCleanupHook）
// ==========================================

// canResolveElementalSealStatus 检查五系封印是否可以触发
// 触发条件：
//  1. 触发时机为 TimingOnCardPlayedOrRevealed（打出或展示牌）
//  2. 打出/展示的牌的元素与封印绑定的元素一致
func canResolveElementalSealStatus(ctx *model.Context, fc *model.FieldCard) bool {
	if ctx == nil || ctx.User == nil || fc == nil {
		return false
	}
	// 只在"打出牌"或"展示牌"时触发
	if ctx.Timing != model.TimingOnCardPlayedOrRevealed {
		return false
	}
	if ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	// 获取封印绑定的元素（从Meta中读取，或根据EffectType推断）
	boundElement := model.BoundElementForFieldCard(fc)
	// 检查打出/展示的牌的元素是否匹配封印元素
	return boundElement != "" && ctx.EventCtx.Card.Element == boundElement
}

// executeElementalSealStatus 执行五系封印效果
// 效果：
//  1. 对目标造成3点法术伤害
//  2. 伤害应用后移除该封印（通过EffectTypeToRemove标记）
func executeElementalSealStatus(ctx *model.Context, fc *model.FieldCard) error {
	if !canResolveElementalSealStatus(ctx, fc) {
		return nil
	}
	// 伤害来源：放置封印的玩家（SourceID）
	sourceID := fc.SourceID
	if sourceID == "" {
		sourceID = ctx.User.ID
	}
	actionWord := "打出"
	if ctx.Timing == model.TimingOnCardPlayedOrRevealed {
		actionWord = "展示"
	}
	ctx.Game.Log(fmt.Sprintf(
		"[Seal] %s %s了%s系牌，触发了%s",
		ctx.User.Name,
		actionWord,
		model.BoundElementForFieldCard(fc),
		elementalSealName(fc),
	))
	// 添加待结算伤害，同时标记需要移除的封印效果。
	// EffectTypeToRemove 由封印结算钩子在伤害应用后消费并移除场上封印牌。
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:           sourceID,
		TargetID:           ctx.User.ID,
		Damage:             3,
		DamageType:         model.MagicAttack,
		EffectTypeToRemove: fc.Effect, // 【关键】伤害应用后移除此封印
	})
	return nil
}

// elementalSealName 获取封印的名称
// 优先使用场牌上的卡牌名称，没有则根据EffectType返回默认名称
func elementalSealName(fc *model.FieldCard) string {
	if fc != nil && fc.Card.Name != "" {
		return fc.Card.Name
	}
	switch fc.Effect {
	case model.EffectSealWater:
		return "水之封印"
	case model.EffectSealFire:
		return "火之封印"
	case model.EffectSealEarth:
		return "地之封印"
	case model.EffectSealWind:
		return "风之封印"
	case model.EffectSealThunder:
		return "雷之封印"
	default:
		return string(fc.Effect)
	}
}
