// gameflow: SkillDispatcher、技能注册、策略钩子、后置钩子。

package engine

import (
	"fmt"

	playerpkg "starcup-engine/internal/engine/player"
	skillrt "starcup-engine/internal/engine/runtime/skill"
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// ---------- SkillDispatcher（薄 facade） ----------

// SkillDispatcher 统一技能调度器（薄 facade）。
type SkillDispatcher struct {
	engine  *GameEngine
	runtime *skillrt.Runtime
}

// NewSkillDispatcher 创建技能调度器。
func NewSkillDispatcher(engine *GameEngine) *SkillDispatcher {
	return &SkillDispatcher{
		engine: engine,
	}
}

// SetRuntime 设置 skill.Runtime（在 runtime 初始化后调用）。
func (sd *SkillDispatcher) SetRuntime(rt *skillrt.Runtime) {
	sd.runtime = rt
	if sd.runtime != nil {
		sd.runtime.SetSkillPolicies(skillUsePolicies)
	}
}

// OnTiming 在某个 Timing 窗口触发技能分发。
func (sd *SkillDispatcher) OnTiming(timing model.FlowTiming, ctx *model.Context) {
	if sd == nil || sd.runtime == nil || ctx == nil {
		return
	}
	sd.runtime.OnTiming(sd.skillHost(), timing, ctx)
}

// ConfirmStartupSkill 确认执行启动技能。
func (sd *SkillDispatcher) ConfirmStartupSkill(playerID string, skillID string) error {
	if sd == nil || sd.runtime == nil {
		return fmt.Errorf("技能运行时未初始化")
	}
	return sd.runtime.ConfirmStartupSkill(sd.skillHost(), playerID, skillID)
}

func (sd *SkillDispatcher) ConfirmStartupSkillAction(playerID string, skillID string) (skillrt.InterruptActionResult, error) {
	if sd == nil || sd.runtime == nil {
		return skillrt.InterruptActionResult{}, fmt.Errorf("技能运行时未初始化")
	}
	return sd.runtime.ConfirmStartupSkillAction(sd.skillHost(), playerID, skillID)
}

// SkipStartupSkill 跳过启动技能。
func (sd *SkillDispatcher) SkipStartupSkill(playerID string) error {
	if sd == nil || sd.runtime == nil {
		return fmt.Errorf("技能运行时未初始化")
	}
	return sd.runtime.SkipStartupSkill(sd.skillHost(), playerID)
}

func (sd *SkillDispatcher) SkipStartupSkillAction(playerID string) (skillrt.InterruptActionResult, error) {
	if sd == nil || sd.runtime == nil {
		return skillrt.InterruptActionResult{}, fmt.Errorf("技能运行时未初始化")
	}
	return sd.runtime.SkipStartupSkillAction(sd.skillHost(), playerID)
}

// ConfirmResponseSkill 确认执行响应技能。
func (sd *SkillDispatcher) ConfirmResponseSkill(playerID string, skillID string) error {
	if sd == nil || sd.runtime == nil {
		return fmt.Errorf("技能运行时未初始化")
	}
	return sd.runtime.ConfirmResponseSkill(sd.skillHost(), playerID, skillID)
}

func (sd *SkillDispatcher) ConfirmResponseSkillAction(playerID string, skillID string) (skillrt.InterruptActionResult, error) {
	if sd == nil || sd.runtime == nil {
		return skillrt.InterruptActionResult{}, fmt.Errorf("技能运行时未初始化")
	}
	return sd.runtime.ConfirmResponseSkillAction(sd.skillHost(), playerID, skillID)
}

// processSkills 处理收集到的技能（调试/作弊等内部路径）。
func (sd *SkillDispatcher) processSkills(skillBatch []model.SkillDefinition, ctx *model.Context) {
	if sd == nil || sd.runtime == nil || ctx == nil {
		return
	}
	sd.runtime.ProcessSkillBatch(sd.skillHost(), skillBatch, ctx)
}

// IsSkillStillUsable 检查技能在响应链中是否仍可用。
func (sd *SkillDispatcher) IsSkillStillUsable(skillID string, user *model.Player, ctx *model.Context) bool {
	if sd == nil || sd.runtime == nil {
		return false
	}
	return sd.runtime.IsSkillStillUsable(skillID, user, ctx)
}

// getOtherUsableSkills 获取除当前技能外，中断列表中仍可用的响应技能 ID。
func (sd *SkillDispatcher) getOtherUsableSkills(currentSkillID string, player *model.Player, ctx *model.Context) []string {
	if sd == nil || sd.runtime == nil || sd.engine == nil || sd.engine.State == nil || sd.engine.State.PendingInterrupt == nil {
		return nil
	}
	return sd.runtime.GetOtherUsableResponseSkills(
		currentSkillID,
		player,
		ctx,
		sd.engine.State.PendingInterrupt.SkillIDs,
	)
}

// ---------- 技能注册（原 skill_registry.go） ----------

type playerSkillRegistrarAdapter struct{}

func (playerSkillRegistrarAdapter) Register(id string, handler model.SkillHandler) {
	skills.Register(id, handler)
}

func registerRoleEntrySkills() {
	for _, entry := range roleRegistry.Entries() {
		for _, skill := range entry.Skills {
			if skill.ID != "" && skill.Handler != nil {
				skills.Register(skill.ID, skill.Handler)
			}
		}
	}
}

func init() {
	registerRoleEntrySkills()
}

// ---------- 技能通用工具（原 skill_runtime_utils.go） ----------


// ---------- 技能策略钩子（原 skill_use_policy.go） ----------

// 技能策略类型直接复用 types 包定义。
type SkillPolicy = types.SkillPolicy
type PolicyContext = types.PolicyContext
type PolicyHost = types.PolicyHost

var skillUsePolicies = map[string]SkillPolicy{}

func init() {
	mountPlayerSkillPolicySpecs(skillUsePolicies)
}

func resolveSkillUsePolicy(skillID string) SkillPolicy {
	if policy, ok := skillUsePolicies[skillID]; ok {
		return policy
	}
	return SkillPolicy{}
}

// ---------- 技能后置钩子 ----------

func (e *GameEngine) runTimingActionEndSkillPost(use *skillUseRequest) {
	if e == nil || use == nil || use.player == nil {
		return
	}
	ctx := playerpkg.TimingHookContext{
		Player:  use.player,
		SkillID: use.skillID,
	}
	e.dispatchAllRoleTimingHooks(playerpkg.TimingOnSkillPost, ctx)
}
