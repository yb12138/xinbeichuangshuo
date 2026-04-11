// gameflow: buildContext：组装 User/Target/Timing/EventCtx。

package engine

import "starcup-engine/internal/model"

func (e *GameEngine) buildContext(user *model.Player, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context {
	ctx := &model.Context{
		Game:       e,
		User:       user,
		Target:     target,
		Timing:     timing,
		EventCtx: eventCtx,
		// 初始化 map 避免 handler 写入时 panic
		Selections: make(map[string]any),
		Flags:      make(map[string]bool),
		// 当前PendingInterrupt （仅供Handler读取，不要修改）
		PendingInterrupt: e.State.PendingInterrupt,
		// 自动将单个 Target 包装进 Targets 切片，方便多目标技能处理
		Targets: []*model.Player{},
	}
	ctx.Selections["current_resume_point"] = e.currentChoiceResumePoint()
	ctx.Selections["current_turn_stage"] = e.State.TurnStage
	ctx.Selections["current_combat_stage"] = e.State.CombatStage
	ctx.Selections["current_subflow"] = e.State.Subflow

	if target != nil {
		ctx.Targets = append(ctx.Targets, target)
	}

	return ctx
}
