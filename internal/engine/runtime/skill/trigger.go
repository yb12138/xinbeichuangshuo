// gameflow: 技能候选排序与中断推送决策（旧 processSkills）。

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

	for _, sk := range skillBatch {
		if ctx != nil && ctx.Timing == model.TimingBeforeMoraleLoss && sk.ID == "ss_soul_devour" {
			continue
		}

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

		switch sk.ResponseType {
		case model.ResponseOptional:
			handler := skillhandlers.GetHandler(ResolveHandlerID(sk))
			if handler == nil || !handler.CanUse(ctx) {
				continue
			}
			optionalSkillIDs = append(optionalSkillIDs, sk.ID)
			sharedCtx = ctx

		case model.ResponseSilent, model.ResponseMandatory:
			t.exec.ExecuteSkill(h, sk, ctx)
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
