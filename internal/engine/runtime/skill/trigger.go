// gameflow: 技能候选排序与中断推送决策。

package skill

import (
	"fmt"
	"sort"

	skillhandlers "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// Trigger 处理技能批次：静默执行或推送启动/响应中断。
type Trigger struct {
	exec *Executor
}

// NewTrigger 创建触发器。
func NewTrigger(exec *Executor) *Trigger {
	return &Trigger{exec: exec}
}

// ProcessSkillBatch 处理收集到的技能列表。
func (t *Trigger) ProcessSkillBatch(h Host, skillBatch []model.SkillDefinition, ctx *model.Context) {
	sort.SliceStable(skillBatch, func(i, j int) bool {
		return skillBatch[i].Priority > skillBatch[j].Priority
	})

	var startupSkillIDs []string
	var optionalSkillIDs []string
	var sharedCtx *model.Context

	// 第一轮：执行启动技能（收集）和静默/强制技能（立即执行）
	for _, sk := range skillBatch {
		if sk.Type == model.SkillTypeStartup {
			if ctx != nil && ctx.User != nil && ctx.User.TurnState.HasUsedActionSkill {
				continue
			}
			if ctx != nil && ctx.User != nil && ctx.User.TurnState.UsedSkillCounts[sk.ID] > 0 {
				continue
			}
			handler := skillhandlers.GetHandler(ResolveHandlerID(sk))
			if handler != nil && handler.CanUse(ctx) {
				startupSkillIDs = append(startupSkillIDs, sk.ID)
				sharedCtx = ctx
			}
			continue
		}

		// 静默和强制响应技能立即执行
		if sk.ResponseType == model.ResponseSilent || sk.ResponseType == model.ResponseMandatory {
			t.exec.ExecuteSkill(h, sk, ctx)
		}
	}

	// 第二轮：在静默技能执行后，收集可选响应技能（此时可选技能的 CanUse 会反映最新状态）
	for _, sk := range skillBatch {
		if sk.Type == model.SkillTypeStartup {
			continue
		}

		if sk.ResponseType == model.ResponseOptional {
			handler := skillhandlers.GetHandler(ResolveHandlerID(sk))
			if handler != nil && handler.CanUse(ctx) {
				optionalSkillIDs = append(optionalSkillIDs, sk.ID)
				sharedCtx = ctx
			}
		}
	}

	optionalSkillIDs = h.ApplyHitCheckAugment(optionalSkillIDs, ctx)
	optionalSkillIDs = h.ApplyHitCheckNormalize(optionalSkillIDs, ctx)
	optionalSkillIDs = dedupeSkillIDs(optionalSkillIDs)
	startupSkillIDs = dedupeSkillIDs(startupSkillIDs)

	if len(optionalSkillIDs) > 0 && sharedCtx == nil {
		sharedCtx = ctx
	}

	if len(startupSkillIDs) > 0 && ctx != nil && ctx.User != nil {
		h.PublishStartupInterrupt(ctx.User.ID, startupSkillIDs, sharedCtx)
		h.OnStartupInterruptPublished()
		h.Log(fmt.Sprintf("[Startup] %s 有 %d 个启动技能可以发动", ctx.User.Name, len(startupSkillIDs)))
		return
	}

	if len(optionalSkillIDs) > 0 && ctx != nil && ctx.User != nil {
		h.PublishResponseInterrupt(ctx.User, optionalSkillIDs, sharedCtx)
		h.Log(fmt.Sprintf("%s 有 %d 个响应技能可以发动", ctx.User.Name, len(optionalSkillIDs)))
	}
}

func dedupeSkillIDs(skillIDs []string) []string {
	if len(skillIDs) <= 1 {
		return skillIDs
	}
	out := make([]string, 0, len(skillIDs))
	seen := make(map[string]struct{}, len(skillIDs))
	for _, id := range skillIDs {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
