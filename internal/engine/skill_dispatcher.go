package engine

import (
	"fmt"
	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

// SkillDispatcher 统一技能调度器
type SkillDispatcher struct {
	engine *GameEngine
}

// NewSkillDispatcher 创建技能调度器
func NewSkillDispatcher(engine *GameEngine) *SkillDispatcher {
	return &SkillDispatcher{
		engine: engine,
	}
}

// 辅助结构：用于在 OnTrigger 中临时存储 玩家+角色 的对应关系
type checkTarget struct {
	Player *model.Player
	Role   model.SkillRole
}

// OnEvent 在某个事件发生时调用，统一处理技能触发
func (sd *SkillDispatcher) OnTrigger(trigger model.TriggerType, ctx *model.Context) {
	ctx.Trigger = trigger

	// 1. 收集触发的技能
	// 使用 checkTarget 结构来明确当前检查的玩家是什么身份
	var targetsToCheck []checkTarget

	switch trigger {
	case model.TriggerOnTurnStart:
		// 回合开始：只检查当前玩家
		currentPid := sd.engine.State.PlayerOrder[sd.engine.State.CurrentTurn]
		if player := sd.engine.State.Players[currentPid]; player != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: player, Role: model.RoleAny})
		}

	case model.TriggerOnAttackStart, model.TriggerOnAttackHit, model.TriggerOnAttackMiss:
		// 上下文中的 User 是攻击发起者 -> 身份标记为 Attacker
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.User,
				Role:   model.RoleAttacker,
			})
		}
		// 上下文中的 Target 是受击者 -> 身份标记为 Defender
		if ctx.Target != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.Target,
				Role:   model.RoleDefender,
			})
		}

	case model.TriggerOnDamageTaken:
		// 在 combat.go 的 handleTakeHit 中，TriggerOnDamageTaken 的 ctx.User 是受害者
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.User,
				Role:   model.RoleDefender,
			})
		}
		// 允许“造成伤害后触发”的技能（如元素吸收）以攻击者身份检查
		if ctx.Target != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.Target,
				Role:   model.RoleAttacker,
			})
		}

	case model.TriggerOnCardUsed:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.User, Role: model.RoleAny})
		}

	case model.TriggerBeforeDraw, model.TriggerAfterDraw:
		if ctx.User != nil {
			// 摸牌通常没有明确的攻防关系，但在水影的case里，是在受击时触发摸牌
			// 为了安全，这里给 RoleAny，或者根据 TriggerBeforeDraw 的调用源头来定
			// 简单起见，RoleDefender 的技能也能在 RoleAny 状态下触发（如果逻辑允许），
			// 但这里我们主要限制 "Defender 不要在攻击时触发"。
			// 对于水影，TriggerBeforeDraw 是个通用事件。
			// 我们可以让 collectTriggeredSkills 宽容处理：RoleAny 的上下文可以触发任何技能，
			// 只有当上下文明确是 Attacker 时才屏蔽 Defender 技能。
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.User, Role: model.RoleAny})
		}

	case model.TriggerOnPhaseEnd, model.TriggerOnBuffRemoved:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.User, Role: model.RoleAny})
		}
		// 对于 BuffRemoved，可能需要检查全场玩家的被动技能 (如天使羁绊)
		if trigger == model.TriggerOnBuffRemoved {
			for _, p := range sd.engine.State.Players {
				if p.ID != ctx.User.ID {
					targetsToCheck = append(targetsToCheck, checkTarget{Player: p, Role: model.RoleAny})
				}
			}
		}

	case model.TriggerBeforeMoraleLoss:
		// 上下文的 User 是导致士气下降的受害者 (Victim)
		if ctx.User != nil {
			// 遍历所有玩家，找出同阵营的队友 (包括自己)
			for _, p := range sd.engine.State.Players {
				if p.Camp == ctx.User.Camp {
					// 这里的 RoleAny 表示队友是以“旁观者/盟友”的身份介入
					targetsToCheck = append(targetsToCheck, checkTarget{
						Player: p,
						Role:   model.RoleAny,
					})
				}
			}
		}

	default:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.User, Role: model.RoleAny})
		}
		if ctx.Target != nil && ctx.Target != ctx.User {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.Target, Role: model.RoleAny})
		}
	}

	// 3. 执行检查
	for _, target := range targetsToCheck {
		if target.Player == nil || target.Player.Character == nil {
			continue
		}
		// 创建上下文副本，确保 User 是当前技能持有者
		skillCtx := *ctx
		skillCtx.User = target.Player

		// 【关键】传入当前玩家在事件中的角色
		skills := sd.collectTriggeredSkills(target.Player, trigger, &skillCtx, target.Role)
		sd.processSkills(skills, &skillCtx)
	}
}

// collectTriggeredSkills 收集指定玩家在指定触发时机下可触发的技能
func (sd *SkillDispatcher) collectTriggeredSkills(player *model.Player,
	trigger model.TriggerType, ctx *model.Context, currentRole model.SkillRole) []model.SkillDefinition {
	var triggeredSkills []model.SkillDefinition

	for _, skill := range player.Character.Skills {
		// 1. 触发匹配检查 (支持主触发器 OR 额外触发器)
		isMatch := skill.Trigger == trigger
		if !isMatch {
			for _, t := range skill.ExtraTriggers {
				if t == trigger {
					isMatch = true
					break
				}
			}
		}
		if !isMatch {
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
				Trigger:      trigger,
			}

			// 如果 Handler 认为可以用，就加入列表
			triggeredSkills = append(triggeredSkills, fieldSkill)
		}
	}

	return triggeredSkills
}

// processSkills 处理收集到的技能，根据ResponseType决定执行方式
func (sd *SkillDispatcher) processSkills(triggeredSkills []model.SkillDefinition, ctx *model.Context) {
	var startupSkillIDs []string
	var optionalSkillIDs []string
	// 用于保存可选技能的上下文，假设所有并发触发的技能共享同一个上下文结构
	// (在星杯中，同一时机的技能通常共享 TriggerCtx)
	var sharedCtx *model.Context
	for _, skill := range triggeredSkills {
		// 【新增】特别处理 Startup 技能
		if skill.Type == model.SkillTypeStartup {
			// 本回合启动技只处理一次（无论发动或跳过），避免在 Startup 阶段重复弹窗。
			if ctx != nil && ctx.User != nil && ctx.User.TurnState.HasUsedTriggerSkill {
				continue
			}
			// 同一回合内同一启动技能至多发动一次，避免重复触发同一技能导致空转。
			if ctx != nil && ctx.User != nil && ctx.User.TurnState.UsedSkillCounts[skill.ID] > 0 {
				continue
			}
			// 检查是否可用 (资源等)
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

	// 如果有 Startup 技能，推送 Startup 中断 (优先于 Response)
	if len(startupSkillIDs) > 0 {
		sd.engine.State.PendingInterrupt = &model.Interrupt{
			Type:     model.InterruptStartupSkill,
			PlayerID: ctx.User.ID,
			SkillIDs: startupSkillIDs,
			Context:  sharedCtx,
		}
		sd.engine.State.Phase = model.PhaseStartup // Ensure phase is correct
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
	return ctx.TriggerCtx.Card.MatchExclusive(player.Character.Name, skill.Title)
}

// executeSkill 执行单个技能
func (sd *SkillDispatcher) executeSkill(skill model.SkillDefinition, ctx *model.Context) {
	handler := skills.GetHandler(skill.LogicHandler)
	if handler == nil {
		return
	}

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

// ConfirmStartupSkill 确认执行启动技能
func (sd *SkillDispatcher) ConfirmStartupSkill(playerID string, skillID string) error {
	intr := sd.engine.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptStartupSkill {
		return fmt.Errorf("当前没有可确认的启动技能")
	}

	if intr.PlayerID != playerID {
		return fmt.Errorf("不是你的启动阶段")
	}

	ctx, ok := intr.Context.(*model.Context)
	if !ok {
		return fmt.Errorf("上下文无效")
	}

	// 查找技能
	player := sd.engine.State.Players[playerID]
	var skillDef *model.SkillDefinition
	for _, s := range player.Character.Skills {
		if s.ID == skillID {
			skillDef = &s
			break
		}
	}
	if skillDef == nil {
		return fmt.Errorf("技能不存在")
	}

	// 执行技能
	sd.executeSkill(*skillDef, ctx)
	if skillID == "elf_ritual" {
		sd.dropQueuedOverflowDiscardForPlayer(playerID)
	}

	// 记录本回合已发动该启动技，避免同技能反复触发导致循环。
	if player.TurnState.UsedSkillCounts == nil {
		player.TurnState.UsedSkillCounts = map[string]int{}
	}
	player.TurnState.UsedSkillCounts[skillID]++
	// 启动阶段每回合只允许选择一次：确认发动后即视为本回合已处理启动技能。
	player.TurnState.HasUsedTriggerSkill = true
	// 本回合一旦执行过启动技能，则禁止特殊行动（购买/合成/提炼）。
	sd.engine.State.HasPerformedStartup = true

	// 若技能执行过程中产生了新的中断（如摸牌溢出弃牌），不要把它清掉。
	if sd.engine.State.PendingInterrupt != nil && sd.engine.State.PendingInterrupt.Type == model.InterruptStartupSkill {
		// 使用 PopInterrupt 处理队列
		sd.engine.PopInterrupt()
	}

	return nil
}

// dropQueuedOverflowDiscardForPlayer 清理“已转入祝福后仍残留”的爆牌弃牌中断。
// 仅用于精灵密仪确认后兜底，避免出现过期的 DiscardSelection。
func (sd *SkillDispatcher) dropQueuedOverflowDiscardForPlayer(playerID string) {
	player := sd.engine.State.Players[playerID]
	if player == nil {
		return
	}
	if len(player.Hand) > sd.engine.GetMaxHand(player) {
		// 仍然超限，说明确实需要弃牌，不做清理。
		return
	}
	filtered := make([]*model.Interrupt, 0, len(sd.engine.State.InterruptQueue))
	for _, intr := range sd.engine.State.InterruptQueue {
		if intr == nil || intr.Type != model.InterruptDiscard || intr.PlayerID != playerID {
			filtered = append(filtered, intr)
			continue
		}
		data, ok := intr.Context.(map[string]interface{})
		if !ok {
			filtered = append(filtered, intr)
			continue
		}
		victimID, _ := data["victim_id"].(string)
		if victimID == playerID {
			sd.engine.Log(fmt.Sprintf("[System] 清理过期中断: %s 的爆牌弃牌请求", player.Name))
			continue
		}
		filtered = append(filtered, intr)
	}
	sd.engine.State.InterruptQueue = filtered
}

// SkipStartupSkill 跳过启动技能
func (sd *SkillDispatcher) SkipStartupSkill(playerID string) error {
	intr := sd.engine.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptStartupSkill {
		return fmt.Errorf("当前没有可跳过的启动技能")
	}

	if intr.PlayerID != playerID {
		return fmt.Errorf("不是你的回合")
	}

	if player := sd.engine.State.Players[playerID]; player != nil {
		player.TurnState.HasUsedTriggerSkill = true
	}

	// 使用 PopInterrupt 处理队列
	sd.engine.PopInterrupt()

	return nil
}

// isSkillStillUsable 检查技能是否仍然可用
func (sd *SkillDispatcher) isSkillStillUsable(skillID string, user *model.Player, ctx *model.Context) bool {
	// 1. 查找技能定义
	var skillDef *model.SkillDefinition
	if user.Character != nil {
		for _, s := range user.Character.Skills {
			if s.ID == skillID {
				skillDef = &s
				break
			}
		}
	}
	if skillDef == nil {
		return false
	}

	// 2. 检查资源 (这是最重要的，因为前一个技能可能耗光了资源)
	if !canPaySkillEnergyCost(user, skillDef.CostGem, skillDef.CostCrystal) {
		return false
	}
	if !sd.uniqueSkillCardMatches(user, *skillDef, ctx) {
		return false
	}

	// 3. 检查 Handler 的 CanUse (逻辑条件)
	handler := skills.GetHandler(skillDef.LogicHandler)
	if handler == nil {
		return false
	}

	return handler.CanUse(ctx)
}

// ConfirmResponseSkill 确认执行响应技能
func (sd *SkillDispatcher) ConfirmResponseSkill(playerID string, skillID string) error {
	// 校验中断状态
	if sd.engine.State.PendingInterrupt == nil {
		return fmt.Errorf("当前没有待处理的响应技能")
	}

	if sd.engine.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		return fmt.Errorf("当前中断不是响应技能类型")
	}

	if sd.engine.State.PendingInterrupt.PlayerID != playerID {
		return fmt.Errorf("不是你的响应回合")
	}

	// 检查技能是否在可用列表中
	found := false
	for _, availableSkillID := range sd.engine.State.PendingInterrupt.SkillIDs {
		if availableSkillID == skillID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("该技能不可用")
	}

	// 获取上下文
	ctx, ok := sd.engine.State.PendingInterrupt.Context.(*model.Context)
	if !ok {
		return fmt.Errorf("技能上下文无效")
	}

	// 找到技能定义
	var skillDef *model.SkillDefinition
	player := sd.engine.State.Players[playerID]
	if player != nil && player.Character != nil {
		for _, skill := range player.Character.Skills {
			if skill.ID == skillID {
				skillDef = &skill
				break
			}
		}
	}
	if skillDef == nil {
		return fmt.Errorf("技能不存在")
	}
	if !sd.uniqueSkillCardMatches(player, *skillDef, ctx) {
		return fmt.Errorf("该独有技与当前打出的牌不匹配")
	}

	// 资源检查
	if player.Gem < skillDef.CostGem {
		return fmt.Errorf("宝石不足 (需要 %d, 拥有 %d)", skillDef.CostGem, player.Gem)
	}
	usableCrystal := player.Crystal + (player.Gem - skillDef.CostGem)
	if usableCrystal < skillDef.CostCrystal {
		return fmt.Errorf(
			"水晶不足 (需要 %d, 可用 %d = 水晶%d + 可替代宝石%d)",
			skillDef.CostCrystal, usableCrystal, player.Crystal, player.Gem-skillDef.CostGem,
		)
	}

	// 根据交互类型处理
	switch skillDef.InteractionType {
	case model.InteractionDiscard:
		// Do NOT Pop. Replace current interrupt directly to maintain stack order.
		sd.engine.State.PendingInterrupt = &model.Interrupt{
			Type:     model.InterruptDiscard,
			PlayerID: playerID,
			Context: map[string]interface{}{
				"skill_id":        skillID,
				"user_ctx":        ctx, // 传递当前上下文
				"min":             skillDef.InteractionConfig.MinSelect,
				"max":             skillDef.InteractionConfig.MaxSelect,
				"prompt":          skillDef.InteractionConfig.Prompt,
				"discard_type":    skillDef.DiscardType,    // 新增
				"discard_element": skillDef.DiscardElement, // 新增
				// 可以在这里把剩余的 SkillIDs 传进去，以便 ConfirmDiscard 恢复
				"remaining_skills": sd.getOtherUsableSkills(skillID, player, ctx),
			},
		}
		sd.engine.State.Phase = model.PhaseDiscardSelection
		sd.engine.Log(fmt.Sprintf("%s 确认发动 [%s]，请选择弃牌", player.Name, skillDef.Title))
		sd.engine.notifyInterruptPrompt() // 新增：发送弃牌选择 prompt 到前端
		return nil

	case model.InteractionNone:
		// 无交互：直接执行技能
		sd.executeSkill(*skillDef, ctx)

		// 2. 检查是否需要恢复暂停的逻辑
		if ctx.Trigger == model.TriggerBeforeDraw {
			sd.engine.resumePendingDraw(ctx)
		} else if ctx.Trigger == model.TriggerOnAttackHit {
			// 恢复战斗流程：推进 PendingDamage 到 Stage1，继续统一伤害管线。
			sd.engine.advancePendingAttackDamageStageAfterHit(ctx)
		}

		// 3. 【核心逻辑】执行完当前技能后，检查是否还有其他技能可以发动
		// (例如：刚发动了风怒，看看剑影是否还能发动)
		remainingSkillIDs := sd.getOtherUsableSkills(skillID, player, ctx)

		// 决策分支
		if len(remainingSkillIDs) > 0 {
			// 更新中断，保持响应阶段
			sd.engine.State.PendingInterrupt.SkillIDs = remainingSkillIDs
			sd.engine.Log(fmt.Sprintf("[System] %s 技能发动成功，检测到还有其他可用响应技能，请继续选择", skillDef.Title))
			return nil // 不弹出中断，保持在响应阶段
		}

		// 没有剩余技能，继续原有流程
		// 清除中断，恢复游戏流程
		sd.engine.PopInterrupt()
		if sd.engine.State.PendingInterrupt == nil && ctx.Trigger == model.TriggerOnAttackMiss {
			if sd.engine.resumePendingAttackMiss(ctx) {
				return nil
			}
		}
		if sd.engine.State.PendingInterrupt == nil && ctx.Trigger == model.TriggerBeforeMoraleLoss {
			if sd.engine.resumePendingMoraleLoss(ctx) {
				return nil
			}
		}

		// 检查是否还有其他待处理的响应 (队列为空时才尝试 NextTurn)
		if sd.engine.State.PendingInterrupt == nil {
			if len(sd.engine.State.ActionStack) > 0 {
				// 优先级 1: 如果栈里还有 Action (比如战斗响应中触发技能)，继续响应
				sd.engine.State.Phase = model.PhaseResponse
			} else if len(sd.engine.State.ActionQueue) > 0 {
				// 优先级 2: 如果队列里还有行动 (说明是在 PhaseBeforeAction 中触发的技能)
				// 必须回到 BeforeAction 继续执行该行动 (如攻击结算)
				sd.engine.State.Phase = model.PhaseBeforeAction
			} else {
				// 优先级 3: 既没响应也没行动，才去回合结束检查 (处理额外行动 Token)
				if len(sd.engine.State.PendingDamageQueue) > 0 {
					sd.engine.State.Phase = model.PhasePendingDamageResolution
					// 保留既有 ReturnPhase（例如战斗承伤路径预设的 ExtraAction），避免误跳过攻击后流程。
					// 但若 ReturnPhase 意外为 Response（常见于在响应阶段入队延迟伤害），
					// 伤害结算后回到 Response 会导致无中断可处理而停滞。
					if sd.engine.State.ReturnPhase == "" || sd.engine.State.ReturnPhase == model.PhaseResponse {
						if ctx.Trigger == model.TriggerOnPhaseEnd {
							// 行动结束响应（如赤色一闪）结算完应回到额外行动阶段。
							sd.engine.State.ReturnPhase = model.PhaseExtraAction
						} else {
							sd.engine.State.ReturnPhase = model.PhaseTurnEnd
						}
					}
				} else {
					sd.engine.State.Phase = model.PhaseTurnEnd
				}
			}
		}

	default:
		return fmt.Errorf("未知的交互类型: %s", skillDef.InteractionType)
	}

	return nil
}

// 辅助函数：获取除去当前技能外，其他仍然可用的技能 ID 列表
func (sd *SkillDispatcher) getOtherUsableSkills(currentSkillID string, player *model.Player, ctx *model.Context) []string {
	var remainingSkillIDs []string

	// 获取当前中断里的技能列表
	currentInterruptSkills := sd.engine.State.PendingInterrupt.SkillIDs

	for _, sid := range currentInterruptSkills {
		if sid == currentSkillID {
			continue // 跳过刚才执行过的技能
		}

		// 重新验证技能是否仍然可用 (因为刚才执行的技能可能消耗了水晶/宝石)
		if sd.isSkillStillUsable(sid, player, ctx) {
			remainingSkillIDs = append(remainingSkillIDs, sid)
		}
	}
	return remainingSkillIDs
}
