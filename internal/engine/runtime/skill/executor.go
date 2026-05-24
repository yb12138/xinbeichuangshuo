// gameflow: 技能 Execute 统一入口（被动/响应/静默与主动共用核心）。

package skill

import (
	"fmt"
	skillhandlers "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// Executor 负责调用 SkillHandler.Execute 与引擎后置回调。
type Executor struct{}

// NewExecutor 创建执行器。
func NewExecutor() *Executor {
	return &Executor{}
}

// ExecuteSkill 执行单个技能。
func (x *Executor) ExecuteSkill(h Host, skill model.SkillDefinition, ctx *model.Context) {
	handler := skillhandlers.GetHandler(ResolveHandlerID(skill))
	if handler == nil {
		return
	}
	beforePoses := h.SnapshotPlayerPoses()

	if model.ContainsSkillTag(skill.Tags, model.TagTurnLimit) && ctx != nil && ctx.User != nil {
		ctx.User.TurnState.UsedSkillCounts[skill.ID]++
	}

	err := handler.Execute(ctx)
	if err != nil {
		if ctx != nil && ctx.Game != nil {
			ctx.Game.Log(fmt.Sprintf("[Skill Error] %s 执行失败: %v", skill.Title, err))
		}
		fmt.Printf("[Skill Error] %s 执行失败: %v\n", skill.Title, err)
		return
	}
	h.SyncPendingDamageFromContext(ctx)
	h.DispatchOrientationChanges(beforePoses)

	if ctx != nil && ctx.Game != nil && ctx.User != nil {
		h.RecordSkillUsage(ctx.User.ID, skill.Title, skill.Type)
		h.NotifySkillActivated(ctx.User.ID, skill.ID, skill.Title, skill.Description, targetIDsFromContext(ctx))
	}

	if ctx != nil && ctx.Game != nil {
		ctx.Game.Log(fmt.Sprintf("[Skill] %s 使用了技能: %s", ctx.User.Name, skill.Title))
	}
	if ctx != nil && ctx.User != nil {
		fmt.Printf("[Skill] %s 发动 [%s]\n", ctx.User.Name, skill.Title)
	}
}

func targetIDsFromContext(ctx *model.Context) []string {
	if ctx == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var ids []string
	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if ctx.Target != nil {
		add(ctx.Target.ID)
	}
	for _, target := range ctx.Targets {
		if target != nil {
			add(target.ID)
		}
	}
	return ids
}
