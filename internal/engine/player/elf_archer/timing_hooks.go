// gameflow: 精灵射手 Timing Hook 实现。

package elf_archer

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// postActionEndHook 攻击行动结束后：风之矢额外攻击 + 清理元素射击状态。
func postActionEndHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack {
		return player.TimingHookResult{}
	}
	if p.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] > 0 {
		model.AppendAttackAction(p, "风之矢")
		rt.Log(fmt.Sprintf("%s 的 [元素射击·风之矢] 结算：额外获得1次攻击行动", p.Name))
	}
	player.ClearElfElementalShotCombatState(p)
	return player.TimingHookResult{}
}

// postAttackHitHook 攻击命中后：元素射击水之矢（治疗）与地之矢（追加法术伤害）。
func postAttackHitHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	// Water arrow: heal target 1
	if p.TurnState.SkillFlowState["elf_elemental_shot_water_pending"] > 0 {
		p.TurnState.SkillFlowState["elf_elemental_shot_water_pending"] = 0
		rt.Heal(ctx.TargetID, 1)
		target := rt.GetPlayer(ctx.TargetID)
		rt.Log(fmt.Sprintf("%s 的 [元素射击·水之矢] 生效：%s +1治疗", p.Name, model.GetPlayerDisplayName(target)))
	}
	// Earth arrow: add 1 magic damage to target
	if p.TurnState.SkillFlowState["elf_elemental_shot_earth_pending"] > 0 {
		p.TurnState.SkillFlowState["elf_elemental_shot_earth_pending"] = 0
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   p.ID,
			TargetID:   ctx.TargetID,
			Damage:     1,
			DamageType: model.MagicDamage,
		})
		target := rt.GetPlayer(ctx.TargetID)
		rt.Log(fmt.Sprintf("%s 的 [元素射击·地之矢] 生效：对 %s 追加1点法术伤害", p.Name, model.GetPlayerDisplayName(target)))
	}
	return player.TimingHookResult{}
}
