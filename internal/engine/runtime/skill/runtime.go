// gameflow: Skill Runtime 统一入口（OnTiming / Confirm / 调试辅助）。

package skill

import (
	"fmt"

	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// Runtime 技能运行时：收集、触发、执行。
type Runtime struct {
	cat      *Catalog
	elig     *Eligibility
	exec     *Executor
	trig     *Trigger
	policies map[string]types.SkillPolicy
}

// NewRuntime 创建并装配技能运行时。
func NewRuntime() *Runtime {
	cat := NewCatalog()
	elig := NewEligibility(cat)
	exec := NewExecutor()
	trig := NewTrigger(exec)
	return &Runtime{
		cat:  cat,
		elig: elig,
		exec: exec,
		trig: trig,
	}
}

// SetSkillPolicies injects role-declared skill policies into the runtime.
func (r *Runtime) SetSkillPolicies(policies map[string]types.SkillPolicy) {
	if r == nil {
		return
	}
	r.policies = policies
	if r.elig != nil {
		r.elig.SetSkillPolicies(policies)
	}
}

func (r *Runtime) skillPolicy(skillID string) types.SkillPolicy {
	if r == nil || r.policies == nil || skillID == "" {
		return types.SkillPolicy{}
	}
	return r.policies[skillID]
}

// OnTiming 在某个 Timing 窗口触发技能分发。
func (r *Runtime) OnTiming(h Host, timing model.FlowTiming, ctx *model.Context) {
	if h == nil || ctx == nil {
		return
	}
	r.runOnTiming(h, timing, ctx)
}

// OnTimingWithAugment 预留与旧 dispatcher 相同的 augment 签名（augment 在 Host 的 ApplyHitCheck* 中完成）。
func (r *Runtime) OnTimingWithAugment(h Host, timing model.FlowTiming, ctx *model.Context, _ /* augmenters */ []any, _ /* normalizers */ []any) {
	r.OnTiming(h, timing, ctx)
}

// ProcessSkillBatch 供调试/作弊路径直接处理一批技能定义。
func (r *Runtime) ProcessSkillBatch(h Host, batch []model.SkillDefinition, ctx *model.Context) {
	if h == nil {
		return
	}
	r.trig.ProcessSkillBatch(h, batch, ctx)
}

// IsSkillStillUsable 与旧 dispatcher.isSkillStillUsable 一致。
func (r *Runtime) IsSkillStillUsable(skillID string, user *model.Player, ctx *model.Context) bool {
	return r.elig.IsStillUsable(skillID, user, ctx)
}

// GetOtherUsableResponseSkills 从中断提供的技能列表中，排除当前技能后仍可用且不互斥的 ID（与旧 getOtherUsableSkills 一致）。
func (r *Runtime) GetOtherUsableResponseSkills(currentSkillID string, player *model.Player, ctx *model.Context, interruptSkillIDs []string) []string {
	return r.elig.FilterRemainingUsable(currentSkillID, player, ctx, interruptSkillIDs)
}

// ExecuteSkill 执行单个技能（静默/确认后路径）。
func (r *Runtime) ExecuteSkill(h Host, skill model.SkillDefinition, ctx *model.Context) {
	if h == nil {
		return
	}
	r.exec.ExecuteSkill(h, skill, ctx)
}

type InterruptActionResult struct {
	Consumed bool
	AfterPop func()
}

func consumeInterruptAction(h Host, result InterruptActionResult, err error) error {
	if err != nil {
		return err
	}
	if result.Consumed {
		h.PopInterrupt()
		if result.AfterPop != nil {
			result.AfterPop()
		}
	}
	return nil
}

func consumeInterruptActionResult(h Host, run func() (InterruptActionResult, error)) error {
	result, err := run()
	return consumeInterruptAction(h, result, err)
}

// ConfirmStartupSkill 确认执行启动技能。
func (r *Runtime) ConfirmStartupSkill(h Host, playerID, skillID string) error {
	return consumeInterruptActionResult(h, func() (InterruptActionResult, error) {
		return r.ConfirmStartupSkillAction(h, playerID, skillID)
	})
}

// ConfirmStartupSkillAction 执行启动技能确认，消费当前中断由调用方负责。
func (r *Runtime) ConfirmStartupSkillAction(h Host, playerID, skillID string) (InterruptActionResult, error) {
	if h == nil {
		return InterruptActionResult{}, fmt.Errorf("技能运行时未初始化")
	}
	intr := h.PendingInterrupt()
	if intr == nil || intr.Type != model.InterruptStartupSkill {
		return InterruptActionResult{}, fmt.Errorf("当前没有可确认的启动技能")
	}
	if intr.PlayerID != playerID {
		return InterruptActionResult{}, fmt.Errorf("不是你的启动阶段")
	}
	ctx, ok := intr.Context.(*model.Context)
	if !ok {
		return InterruptActionResult{}, fmt.Errorf("上下文无效")
	}

	player := h.GameState().Players[playerID]
	if player == nil {
		return InterruptActionResult{}, fmt.Errorf("玩家不存在")
	}
	skillDef := r.cat.FindCharacterSkillOnPlayer(player, skillID)
	if skillDef == nil {
		return InterruptActionResult{}, fmt.Errorf("技能不存在")
	}

	r.exec.ExecuteSkill(h, *skillDef, ctx)
	if err := r.runAfterExecute(h, *skillDef, playerID); err != nil {
		return InterruptActionResult{}, err
	}

	if player.TurnState.UsedSkillCounts == nil {
		player.TurnState.UsedSkillCounts = map[string]int{}
	}
	player.TurnState.UsedSkillCounts[skillID]++
	player.TurnState.HasUsedActionSkill = true

	return InterruptActionResult{Consumed: h.PendingInterrupt() == intr}, nil
}

// SkipStartupSkill 跳过启动技能。
func (r *Runtime) SkipStartupSkill(h Host, playerID string) error {
	return consumeInterruptActionResult(h, func() (InterruptActionResult, error) {
		return r.SkipStartupSkillAction(h, playerID)
	})
}

// SkipStartupSkillAction 跳过启动技能，消费当前中断由调用方负责。
func (r *Runtime) SkipStartupSkillAction(h Host, playerID string) (InterruptActionResult, error) {
	if h == nil {
		return InterruptActionResult{}, fmt.Errorf("技能运行时未初始化")
	}
	intr := h.PendingInterrupt()
	if intr == nil || intr.Type != model.InterruptStartupSkill {
		return InterruptActionResult{}, fmt.Errorf("当前没有可跳过的启动技能")
	}
	if intr.PlayerID != playerID {
		return InterruptActionResult{}, fmt.Errorf("不是你的回合")
	}
	if player := h.GameState().Players[playerID]; player != nil {
		player.TurnState.HasUsedActionSkill = true
	}
	return InterruptActionResult{Consumed: h.PendingInterrupt() == intr}, nil
}

// ConfirmResponseSkill 确认执行响应技能。
func (r *Runtime) ConfirmResponseSkill(h Host, playerID, skillID string) error {
	return consumeInterruptActionResult(h, func() (InterruptActionResult, error) {
		return r.ConfirmResponseSkillAction(h, playerID, skillID)
	})
}

// ConfirmResponseSkillAction 执行响应技能确认，消费当前中断由调用方负责。
func (r *Runtime) ConfirmResponseSkillAction(h Host, playerID, skillID string) (InterruptActionResult, error) {
	if h == nil {
		return InterruptActionResult{}, fmt.Errorf("技能运行时未初始化")
	}
	if h.PendingInterrupt() == nil {
		return InterruptActionResult{}, fmt.Errorf("当前没有待处理的响应技能")
	}
	if h.PendingInterrupt().Type != model.InterruptResponseSkill {
		return InterruptActionResult{}, fmt.Errorf("当前中断不是响应技能类型")
	}
	if h.PendingInterrupt().PlayerID != playerID {
		return InterruptActionResult{}, fmt.Errorf("不是你的响应回合")
	}
	intr := h.PendingInterrupt()

	found := false
	for _, availableSkillID := range h.PendingInterrupt().SkillIDs {
		if availableSkillID == skillID {
			found = true
			break
		}
	}
	if !found {
		return InterruptActionResult{}, fmt.Errorf("该技能不可用")
	}

	ctx, ok := h.PendingInterrupt().Context.(*model.Context)
	if !ok {
		return InterruptActionResult{}, fmt.Errorf("技能上下文无效")
	}

	player := h.GameState().Players[playerID]
	if player == nil || player.Character == nil {
		return InterruptActionResult{}, fmt.Errorf("玩家不存在")
	}
	skillDef := r.cat.FindCharacterSkillOnPlayer(player, skillID)
	if skillDef == nil {
		return InterruptActionResult{}, fmt.Errorf("技能不存在")
	}
	if !r.elig.uniqueSkillCardMatches(player, *skillDef, ctx) {
		return InterruptActionResult{}, fmt.Errorf("该独有技与当前打出的牌不匹配")
	}

	if player.Gem < skillDef.CostGem {
		return InterruptActionResult{}, fmt.Errorf("宝石不足 (需要 %d, 拥有 %d)", skillDef.CostGem, player.Gem)
	}
	usableCrystal := player.Crystal + (player.Gem - skillDef.CostGem)
	if usableCrystal < skillDef.CostCrystal {
		return InterruptActionResult{}, fmt.Errorf(
			"水晶不足 (需要 %d, 可用 %d = 水晶%d + 可替代宝石%d)",
			skillDef.CostCrystal, usableCrystal, player.Crystal, player.Gem-skillDef.CostGem,
		)
	}

	switch skillDef.InteractionType {
	case model.InteractionDiscard:
		remaining := r.elig.FilterRemainingUsable(skillID, player, ctx, h.PendingInterrupt().SkillIDs)
		h.SetPendingInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: playerID,
			Context: map[string]interface{}{
				"choice_type":      "system_discard_cards",
				"discard_subflow":  true,
				"skill_id":         skillID,
				"user_ctx":         ctx,
				"min":              skillDef.InteractionConfig.MinSelect,
				"max":              skillDef.InteractionConfig.MaxSelect,
				"prompt":           skillDef.InteractionConfig.Prompt,
				"discard_type":     skillDef.DiscardType,
				"discard_element":  skillDef.DiscardElement,
				"remaining_skills": remaining,
			},
		})
		h.EnterDiscardSelection()
		h.NotifyInterruptPrompt()
		h.Log(fmt.Sprintf("%s 确认发动 [%s]，请选择弃牌", player.Name, skillDef.Title))
		return InterruptActionResult{}, nil

	case model.InteractionNone:
		resumeState := h.CaptureResponseResumeStateOnConfirm(skillID, ctx)
		r.exec.ExecuteSkill(h, *skillDef, ctx)
		if err := r.runAfterExecute(h, *skillDef, playerID); err != nil {
			return InterruptActionResult{}, err
		}
		h.PrepareConfirmedResponseResume(resumeState)

		remainingSkillIDs := r.elig.FilterRemainingUsable(skillID, player, ctx, h.PendingInterrupt().SkillIDs)
		afterPop := func() {
			h.RestoreConfirmedResponseAfterPop(resumeState)
			if len(remainingSkillIDs) > 0 {
				h.PublishResponseInterrupt(player, remainingSkillIDs, ctx)
				h.Log(fmt.Sprintf("[System] %s 技能结算完成，检测到还有其他可用响应技能，请继续选择", skillDef.Title))
			}
		}
		return InterruptActionResult{Consumed: h.PendingInterrupt() == intr, AfterPop: afterPop}, nil

	default:
		return InterruptActionResult{}, fmt.Errorf("未知的交互类型: %s", skillDef.InteractionType)
	}
}

func (r *Runtime) runAfterExecute(h Host, skillDef model.SkillDefinition, playerID string) error {
	policy := r.skillPolicy(skillDef.ID)
	if policy.AfterExecute == nil {
		return nil
	}
	return policy.AfterExecute(h, types.PolicyContext{
		SkillID:  skillDef.ID,
		PlayerID: playerID,
		SkillDef: skillDef,
	})
}
