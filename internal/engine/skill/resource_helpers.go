// gameflow: 技能内共用的资源校验/消耗辅助（委托到 player 包）。

package skills

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CanPayCrystalLike 红宝石可替代蓝水晶（仅水晶消耗方向）
func CanPayCrystalLike(ctx *model.Context, amount int) bool {
	return player.CanPayCrystalLike(ctx, amount)
}

// SpendCrystalLike 红宝石可替代蓝水晶消耗
func SpendCrystalLike(ctx *model.Context, amount int) bool {
	return player.SpendCrystalLike(ctx, amount)
}
