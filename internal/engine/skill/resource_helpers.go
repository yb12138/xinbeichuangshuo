// gameflow: 技能内共用的资源校验/消耗辅助。

package skills

import "starcup-engine/internal/model"

// CanPayCrystalLike 红宝石可替代蓝水晶（仅水晶消耗方向）
func CanPayCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.CanPayCrystalCost(ctx.User.ID, amount)
}

// SpendCrystalLike 红宝石可替代蓝水晶消耗
func SpendCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.ConsumeCrystalCost(ctx.User.ID, amount)
}

// 内部别名（保持向后兼容）
func canPayCrystalLike(ctx *model.Context, amount int) bool {
	return CanPayCrystalLike(ctx, amount)
}

func spendCrystalLike(ctx *model.Context, amount int) bool {
	return SpendCrystalLike(ctx, amount)
}
