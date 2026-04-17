// gameflow: UseSkill 弃牌后自动等待封印等 PendingDamage 结算，再恢复技能 handler 执行。

package engine

import (
	"fmt"

	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

const followupTypeSkillEffectResume = "skill_effect_resume"

// buildSkillEffectResumeFollowupHandler 注册通用"技能效果恢复"后续。
func buildSkillEffectResumeFollowupHandler() map[string]deferredFollowupHandler {
	return map[string]deferredFollowupHandler{
		followupTypeSkillEffectResume: {
			label:   "SkillEffectResume",
			resolve: (*GameEngine).resolveSkillEffectResumeFollowup,
		},
	}
}

// resolveSkillEffectResumeFollowup 恢复技能执行：构造 Context，调用 handler.Execute()，
// 再执行 finishSkillUse 的收尾步骤。
func (e *GameEngine) resolveSkillEffectResumeFollowup(f model.DeferredFollowup) error {
	if e == nil || e.State == nil {
		return fmt.Errorf("引擎未初始化")
	}
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("技能恢复：执行者不存在: %s", f.UserID)
	}
	if user.Character == nil {
		return fmt.Errorf("技能恢复：角色未分配")
	}
	skillDef := findCharacterSkill(user.Character, f.SkillID)
	if skillDef == nil {
		return fmt.Errorf("技能恢复：技能未找到: %s", f.SkillID)
	}

	// 从 followup Data 中还原执行上下文
	var target *model.Player
	if len(f.TargetIDs) > 0 {
		target = e.State.Players[f.TargetIDs[0]]
	}
	handler := skills.GetHandler(f.SkillID)
	if handler == nil {
		return fmt.Errorf("技能恢复：handler 未找到: %s", f.SkillID)
	}

	ctx := e.buildContext(user, target, model.TimingActive, nil)
	// 还原多目标
	if len(f.TargetIDs) > 0 {
		ctx.Targets = make([]*model.Player, 0, len(f.TargetIDs))
		for _, tid := range f.TargetIDs {
			if p := e.State.Players[tid]; p != nil {
				ctx.Targets = append(ctx.Targets, p)
			}
		}
	}
	if ctx.Selections == nil {
		ctx.Selections = map[string]any{}
	}
	// 还原弃牌信息
	if discardedRaw, ok := f.Data["discarded_cards"]; ok {
		if cards, ok := discardedRaw.([]model.Card); ok {
			ctx.Selections["discardedCards"] = cards
		}
	}

	// 恢复 Policy（用于 finishSkillUse）
	policy := resolveSkillUsePolicy(f.SkillID)

	beforePoses := e.snapshotPlayerPoses()
	if err := handler.Execute(ctx); err != nil {
		return fmt.Errorf("技能恢复执行失败: %v", err)
	}
	e.dispatchOrientationChanges(beforePoses)

	// finishSkillUse 收尾逻辑
	if skillDef.PlaceCard && skillDef.PlaceMode == model.FieldEffect && len(f.TargetIDs) > 0 {
		e.emitBuffAddedDispatch(user.ID, f.TargetIDs[0], skillDef.PlaceEffect)
	}
	e.runTimingOnActionEndSkillPost(&skillUseRequest{
		engine:   e,
		player:   user,
		skillDef: skillDef,
		skillID:  f.SkillID,
		policy:   policy,
	})
	e.recordSkillUsage(user.ID, skillDef.Title, skillDef.Type)
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: %s (%s)", user.Name, skillDef.Title, skillDef.Description))

	if skillDef.Type == model.SkillTypeAction && !policy.SkipAutoPhaseEnd {
		user.TurnState.HasActed = true
		user.TurnState.LastActionType = string(model.ActionMagic)
		user.TurnState.LastActionCard = nil
	}

	return nil
}
