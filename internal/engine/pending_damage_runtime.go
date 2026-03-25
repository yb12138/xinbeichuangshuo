package engine

import "starcup-engine/internal/model"

// syncPendingDamageRuntimeFromContext 将响应/被动技能在当前伤害上下文里写入的运行时元数据，
// 回填到正在处理的 PendingDamage 头结点，确保中断恢复后状态仍然存在。
func (e *GameEngine) syncPendingDamageRuntimeFromContext(ctx *model.Context) {
	if e == nil || ctx == nil || ctx.Trigger != model.TriggerOnDamageTaken || len(e.State.PendingDamageQueue) == 0 {
		return
	}

	pd := &e.State.PendingDamageQueue[0]
	if ctx.TriggerCtx != nil {
		if ctx.TriggerCtx.SourceID != "" && pd.SourceID != ctx.TriggerCtx.SourceID {
			return
		}
		if ctx.TriggerCtx.TargetID != "" && pd.TargetID != ctx.TriggerCtx.TargetID {
			return
		}
	}
	if ctx.Selections == nil {
		return
	}

	if raw, ok := ctx.Selections["overflow_morale_loss_fixed"]; ok {
		pd.OverflowMoraleLossFixed = toIntContextValue(raw)
	}
}
