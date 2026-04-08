package engine

import (
	"fmt"
	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

// ConfirmStartupSkill / ConfirmResponseSkill 属于“技能交互确认阶段”。
// 这些函数负责把玩家在中断窗口的选择，转回技能执行流水线。
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
	player.TurnState.HasUsedActionSkill = true

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
		player.TurnState.HasUsedActionSkill = true
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
		sd.engine.enterDiscardSelection()
		sd.engine.Log(fmt.Sprintf("%s 确认发动 [%s]，请选择弃牌", player.Name, skillDef.Title))
		sd.engine.notifyInterruptPrompt() // 新增：发送弃牌选择 prompt 到前端
		return nil

	case model.InteractionNone:
		resumeState := sd.engine.captureResponseResumeStateFromContext(responseCompletionConfirm, skillID, ctx)

		// 无交互：直接执行技能
		sd.executeSkill(*skillDef, ctx)

		// 2. 检查是否需要恢复暂停的逻辑
		sd.engine.prepareConfirmedResponseResume(resumeState)

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
		sd.engine.restoreConfirmedResponseAfterPop(resumeState)

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
		if isMutuallyExclusiveResponseSkill(currentSkillID, sid) {
			continue
		}

		// 重新验证技能是否仍然可用 (因为刚才执行的技能可能消耗了水晶/宝石)
		if sd.isSkillStillUsable(sid, player, ctx) {
			remainingSkillIDs = append(remainingSkillIDs, sid)
		}
	}
	return remainingSkillIDs
}

func isMutuallyExclusiveResponseSkill(currentSkillID string, otherSkillID string) bool {
	if currentSkillID == "" || otherSkillID == "" {
		return false
	}
	if (currentSkillID == "elf_animal_companion" || currentSkillID == "elf_pet_empower") &&
		(otherSkillID == "elf_animal_companion" || otherSkillID == "elf_pet_empower") {
		return true
	}
	if (currentSkillID == "hom_rage_suppress" || currentSkillID == "hom_glyph_fusion") &&
		(otherSkillID == "hom_rage_suppress" || otherSkillID == "hom_glyph_fusion") {
		return true
	}
	return false
}
