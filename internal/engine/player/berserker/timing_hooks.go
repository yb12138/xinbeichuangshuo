// gameflow: 狂战士 Timing Hook 实现。

package berserker

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// hitCheckHook 狂战士血腥咆哮：目标治疗剂为2时强制命中且无视圣盾。
func hitCheckHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	attacker := rt.GetPlayer(ctx.SourceID)
	victim := rt.GetPlayer(ctx.TargetID)
	if attacker == nil || victim == nil || ctx.IsCounter || !rt.IsCharacter(attacker, "berserker") || victim.Heal != 2 {
		return player.TimingHookResult{}
	}
	// 命中判定时修改 PendingDamage
	if ctx.PendingDamage != nil {
		ctx.PendingDamage.SetInterceptTag(model.CombatInterceptForceHit)
		ctx.PendingDamage.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
		ctx.PendingDamage.IgnoreHeal = true
		rt.Log(fmt.Sprintf("%s 发动 [血腥咆哮]！目标治疗剂为2，强制命中且无视圣盾", attacker.Name))
	}
	return player.TimingHookResult{}
}
