// gameflow: 冒险者策略 Hook 声明式注册。

package adventurer

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// undergroundLawOverrideHook 地下法则特殊行动覆盖策略。
// 保持与原 PolicySpec stub 一致：不覆盖，让正常购买流程继续。
func undergroundLawOverrideHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	return engineplayer.TimingHookResult{}
}

// extractOverrideHook 冒险者天堂提炼覆盖策略。
// 当冒险者选择提炼时，询问是否发动天堂让队友直接提炼。
// 自身能量已满时自动跳过询问，直接选择队友提炼。
func extractOverrideHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	if ctx.ActionType != model.ActionExtract {
		return engineplayer.TimingHookResult{}
	}
	p := ctx.Player
	if p == nil || !rt.IsCharacter(p, "adventurer") {
		return engineplayer.TimingHookResult{}
	}
	// 玩家已拒绝天堂，不再拦截（防止 StartExtractForPlayer 触发循环）
	if p.TurnState.SkillFlowState != nil && p.TurnState.SkillFlowState["adventurer_declined_paradise"] > 0 {
		p.TurnState.SkillFlowState["adventurer_declined_paradise"] = 0
		return engineplayer.TimingHookResult{}
	}
	// 检查是否有天堂技能且阵营有可用资源
	if !hasParadiseSkill(p) {
		return engineplayer.TimingHookResult{}
	}
	// 检查是否有可承接提炼的队友
	allies := eligibleAllyIDs(rt, p)
	if len(allies) == 0 {
		return engineplayer.TimingHookResult{}
	}

	currentEnergy := p.Gem + p.Crystal
	maxEnergy := rt.GetPlayerEnergyCap(p)
	energyRoom := maxEnergy - currentEnergy

	if energyRoom <= 0 {
		// 自身能量已满，自动选择队友提炼（不询问）
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: p.ID,
			Context: map[string]interface{}{
				"choice_type": "adventurer_paradise_pick",
				"user_id":     p.ID,
				"ally_ids":    allies,
			},
		})
		return engineplayer.TimingHookResult{Handled: true, Interrupted: true}
	}

	// 有能量空间，询问是否发动天堂
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type": "adventurer_extract_paradise_check",
			"user_id":     p.ID,
		},
	})
	return engineplayer.TimingHookResult{Handled: true, Interrupted: true}
}

func hasParadiseSkill(p *model.Player) bool {
	if p == nil || p.Character == nil {
		return false
	}
	for _, s := range p.Character.Skills {
		if s.ID == "adventurer_paradise" {
			return true
		}
	}
	return false
}

func eligibleAllyIDs(rt engineplayer.HookRuntime, p *model.Player) []string {
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		ally := rt.GetPlayer(pid)
		if ally == nil || ally.Camp != p.Camp || ally.ID == p.ID {
			continue
		}
		maxEnergy := rt.GetPlayerEnergyCap(ally)
		if ally.Gem+ally.Crystal < maxEnergy {
			ids = append(ids, pid)
		}
	}
	return ids
}
