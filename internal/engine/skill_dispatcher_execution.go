package engine

import (
	"fmt"
	"sort"
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// collectTriggeredSkills 收集当前时机可触发技能。
// 该阶段只做“能否触发”判断，不做玩家交互。
// collectTriggeredSkills 收集指定玩家在指定触发时机下可触发的技能
func (sd *SkillDispatcher) collectTriggeredSkills(player *model.Player,
	timing model.TriggerTiming, ctx *model.Context, currentRole model.SkillRole) []model.SkillDefinition {
	var triggeredSkills []model.SkillDefinition

	for _, skill := range player.Character.Skills {
		if ctx != nil && ctx.Timing == model.TimingStartup {
			if skill.Type != model.SkillTypeStartup {
				continue
			}
		} else if skill.Type == model.SkillTypeStartup {
			continue
		}

		if !skillMatchesTiming(skill, timing) {
			continue
		}
		// 基本筛选条件
		// if skill.Trigger != trigger {
		// 	continue
		// }
		// 上面的 ExtraTriggers 逻辑已经处理了匹配问题
		// 如果 isMatch 为 false，已经在上面 continue 了
		// 这里不需要再次检查 skill.Trigger == trigger，否则会过滤掉 ExtraTriggers 匹配的情况

		// 2. [核心修改] 身份匹配机制
		// 逻辑：
		// 如果 Context 里的角色是明确的 Attacker，那么不能触发 RoleDefender 的技能。
		// 如果 Context 里的角色是明确的 Defender，那么不能触发 RoleAttacker 的技能。
		// 如果 Context 里的角色是 Any，则均可触发 (或视具体需求而定)。
		if skill.RequiredRole != model.RoleAny && currentRole != model.RoleAny {
			if skill.RequiredRole != currentRole {
				continue // 身份不符：跳过
			}
		}

		// 跳过主动技能（主动技能需要手动使用）
		if skill.Type == model.SkillTypeAction {
			continue
		}

		if !canPaySkillEnergyCost(player, skill.CostGem, skill.CostCrystal) {
			continue
		}
		if skill.CostCoverCards > 0 {
			if len(player.GetCoverCards()) < skill.CostCoverCards {
				continue
			}
		}

		// 检查回合使用限制
		if model.ContainsSkillTag(skill.Tags, model.TagTurnLimit) {
			if count, exists := player.TurnState.UsedSkillCounts[skill.ID]; exists && count > 0 {
				continue // 本回合已使用过
			}
		}

		// 独有技必须由“当前角色打出了匹配该技能的独有牌”才能触发。
		if !sd.uniqueSkillCardMatches(player, skill, ctx) {
			continue
		}

		// 检查技能是否可用（通过SkillHandler.CanUse）
		handler := skills.GetHandler(skill.LogicHandler)
		if handler == nil {
			continue
		}

		if !handler.CanUse(ctx) {
			continue
		}

		triggeredSkills = append(triggeredSkills, skill)
	}

	if ctx != nil && ctx.Timing == model.TimingStartup {
		return triggeredSkills
	}

	for _, fc := range player.Field {
		// 必须是 Effect 模式且未锁定
		if fc.Mode != model.FieldEffect || fc.Locked {
			continue
		}

		// 映射枚举到 Handler ID
		handlerID := model.GetHandlerIDByEffect(fc.Effect)
		if handlerID == "" {
			continue
		}

		// 获取 Handler
		handler := skills.GetHandler(handlerID)
		if handler == nil {
			continue
		}

		// 检查 CanUse
		// 注意：FieldCard 相当于一个被动技能，我们临时构建一个 SkillDefinition 包装它
		// 这样下游的 processSkills 就可以统一处理了
		if handler.CanUse(ctx) {

			// 临时构建一个技能定义，代表这张场上卡
			fieldSkill := model.SkillDefinition{
				ID:    handlerID,
				Title: fc.Card.Name,
				Type:  model.SkillTypePassive,

				// 【关键修改】：设置为静默执行或强制执行
				// 这样 processSkills 方法会直接调用 executeSkill，而不会 PushInterrupt
				ResponseType: model.ResponseSilent,

				LogicHandler: handlerID,
				Timings:      []model.TriggerTiming{timing},
			}

			// 如果 Handler 认为可以用，就加入列表
			triggeredSkills = append(triggeredSkills, fieldSkill)
		}
	}

	return triggeredSkills
}

func skillMatchesTiming(skill model.SkillDefinition, timing model.TriggerTiming) bool {
	if timing == model.TimingStartup && skill.Type == model.SkillTypeStartup {
		// Startup 技能在独立窗口下按技能类型匹配，不再依赖 Trigger 枚举。
		return true
	}

	if len(skill.Timings) > 0 {
		for _, t := range skill.Timings {
			if t == timing {
				return true
			}
		}
		return false
	}

	// 兼容旧配置：迁移完成后删除。
	if legacyTriggerToTiming(skill.Trigger) == timing {
		return true
	}
	for _, t := range skill.ExtraTriggers {
		if legacyTriggerToTiming(t) == timing {
			return true
		}
	}
	return false
}

// processSkills 处理收集到的技能，根据ResponseType决定执行方式
func (sd *SkillDispatcher) processSkills(triggeredSkills []model.SkillDefinition, ctx *model.Context) {
	sort.SliceStable(triggeredSkills, func(i, j int) bool {
		return triggeredSkills[i].Priority > triggeredSkills[j].Priority
	})

	var startupSkillIDs []string
	var optionalSkillIDs []string
	// 用于保存可选技能的上下文，假设所有并发触发的技能共享同一个上下文结构
	// (在星杯中，同一时机的技能通常共享 TriggerCtx)
	var sharedCtx *model.Context
	for _, skill := range triggeredSkills {
		// 灵魂吞噬按文档要求基于“最终实际士气下降”结算，
		// 因此统一在 applyMoraleLossAfterTrigger 中处理，避免被响应修改前抢先加魂。
		if ctx != nil && ctx.Timing == model.TimingBeforeMoraleLoss && skill.ID == "ss_soul_devour" {
			continue
		}

		// 启动技只在 TimingStartup 窗口收集，统一走 StartupInterrupt。
		if skill.Type == model.SkillTypeStartup {
			if ctx != nil && ctx.User != nil && ctx.User.TurnState.HasUsedActionSkill {
				continue
			}
			if ctx != nil && ctx.User != nil && ctx.User.TurnState.UsedSkillCounts[skill.ID] > 0 {
				continue
			}
			handler := skills.GetHandler(skill.LogicHandler)
			if handler != nil && handler.CanUse(ctx) {
				startupSkillIDs = append(startupSkillIDs, skill.ID)
				sharedCtx = ctx
			}
			continue
		}

		switch skill.ResponseType {
		case model.ResponseOptional:
			// 可选响应：检查CanUse，如果可以则通过中断系统处理
			handler := skills.GetHandler(skill.LogicHandler)
			if handler == nil || !handler.CanUse(ctx) {
				continue
			}

			optionalSkillIDs = append(optionalSkillIDs, skill.ID)
			sharedCtx = ctx // 记录上下文

		case model.ResponseSilent:
			// 静默执行：直接执行
			sd.executeSkill(skill, ctx)

		case model.ResponseMandatory:
			// 强制响应：直接执行（通常用于被动效果）
			sd.executeSkill(skill, ctx)
		}
	}

	optionalSkillIDs = sd.dispatchTimingOnHitCheckSkillIDs(optionalSkillIDs, ctx, timingOnHitCheckSkillAugment)
	optionalSkillIDs = sd.dispatchTimingOnHitCheckSkillIDs(optionalSkillIDs, ctx, timingOnHitCheckSkillNormalize)
	if len(optionalSkillIDs) > 0 && sharedCtx == nil {
		sharedCtx = ctx
	}

	// 如果有 Startup 技能，推送 Startup 中断 (优先于 Response)
	if len(startupSkillIDs) > 0 {
		sd.engine.State.PendingInterrupt = &model.Interrupt{
			Type:     model.InterruptStartupSkill,
			PlayerID: ctx.User.ID,
			SkillIDs: startupSkillIDs,
			Context:  sharedCtx,
		}
		sd.engine.setTurnStage(model.TurnStageActionStart)
		sd.engine.clearCombatStage()
		sd.engine.clearSubflow()
		sd.engine.Log(fmt.Sprintf("[Startup] %s 有 %d 个启动技能可以发动", ctx.User.Name, len(startupSkillIDs)))
		return // 暂不处理其他中断，一次只处理一种类型
	}

	// 如果有收集到可选技能，推送【单次】中断，包含所有技能 ID
	if len(optionalSkillIDs) > 0 {
		sd.engine.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptResponseSkill,
			PlayerID: ctx.User.ID,
			SkillIDs: optionalSkillIDs, // 【关键】传入列表
			Context:  sharedCtx,
		})
		sd.engine.Log(fmt.Sprintf("%s 有 %d 个响应技能可以发动", ctx.User.Name, len(optionalSkillIDs)))
	}
}

// uniqueSkillCardMatches 校验独有技是否由当前角色打出对应独有牌触发。
func (sd *SkillDispatcher) uniqueSkillCardMatches(player *model.Player, skill model.SkillDefinition, ctx *model.Context) bool {
	if !model.ContainsSkillTag(skill.Tags, model.TagUnique) {
		return true
	}
	if player == nil || player.Character == nil || ctx == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.Card == nil {
		return false
	}
	return ctx.TriggerCtx.Card.MatchExclusive(player.Character.ID, skill.Title)
}

// executeSkill 执行单个技能
func (sd *SkillDispatcher) executeSkill(skill model.SkillDefinition, ctx *model.Context) {
	handler := skills.GetHandler(skill.LogicHandler)
	if handler == nil {
		return
	}
	beforePoses := sd.engine.snapshotPlayerPoses()

	// 记录回合使用次数
	if model.ContainsSkillTag(skill.Tags, model.TagTurnLimit) {
		ctx.User.TurnState.UsedSkillCounts[skill.ID]++
	}

	// 【修正】如果是独有技，且不是由打出该牌触发的，需要提醒玩家选择手里的独有牌
	if model.ContainsSkillTag(skill.Tags, model.TagUnique) {
		isConsumingTrigger := ctx.Trigger == model.TriggerOnAttackStart ||
			ctx.Trigger == model.TriggerOnCardUsed

		if !isConsumingTrigger {
			// 如果已经在响应确认中断中，且是独有技，我们需要在 Execute 之前确保弃牌
			// 这里的逻辑较为复杂，因为 dispatcher 是同步执行的。
			// 暂且维持现状：如果玩家手里有多张合法独有牌，在执行确认时由 ConfirmResponseSkill 处理
		}
	}

	// 执行技能
	err := handler.Execute(ctx)
	if err != nil {
		if ctx != nil && ctx.Game != nil {
			ctx.Game.Log(fmt.Sprintf("[Skill Error] %s 执行失败: %v", skill.Title, err))
		}
		fmt.Printf("[Skill Error] %s 执行失败: %v\n", skill.Title, err)
		return
	}
	sd.engine.syncPendingDamageRuntimeFromContext(ctx)
	sd.engine.dispatchOrientationChanges(beforePoses)

	if ctx != nil && ctx.Game != nil && ctx.User != nil {
		if engine, ok := ctx.Game.(*GameEngine); ok {
			engine.recordSkillUsage(ctx.User.ID, skill.Title, skill.Type)
		}
	}

	// 打印执行日志
	if ctx != nil && ctx.Game != nil {
		// 事件流使用“使用了技能”格式，避免与技能内日志的“发动 [技能]”重复冲突。
		ctx.Game.Log(fmt.Sprintf("[Skill] %s 使用了技能: %s", ctx.User.Name, skill.Title))
	}
	fmt.Printf("[Skill] %s 发动 [%s]\n", ctx.User.Name, skill.Title)
}

func containsSkillID(skillIDs []string, skillID string) bool {
	for _, id := range skillIDs {
		if id == skillID {
			return true
		}
	}
	return false
}
