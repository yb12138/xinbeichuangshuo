// gameflow: 兽灵武士 Timing Hook 实现。

package beast_samurai

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func postAttackHitHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if ctx.IsCounter || InIaijutsuForm(p) {
		return player.TimingHookResult{}
	}
	before := BeastSoul(p)
	after := AddBeastSoul(p, 1, false)
	if after > before {
		rt.Log(fmt.Sprintf("%s 的 [兽魂意念] 生效：普通形态主动攻击命中，兽魂+1（当前%d）", p.Name, after))
	}
	return player.TimingHookResult{}
}

// postDamageResolvedHook 伤害结算完成后：清除攻击指示物 + 居合形态退场。
func postDamageResolvedHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	source := rt.GetPlayer(ctx.SourceID)
	if source == nil {
		return player.TimingHookResult{}
	}
	// Clear attack tokens for attack damage
	if strings.EqualFold(string(ctx.DamageType), string(model.AttackDamage)) {
		ClearAttackTokens(source)
	}
	// Leave iaijutsu form on damage dealt
	if ctx.Damage > 0 && InIaijutsuForm(source) {
		before := rt.SnapshotPlayerPoses()
		if LeaveIaijutsuForm(source) {
			rt.Log(fmt.Sprintf("%s 的 [御魂流居合形态·造成伤害退场] 生效：转正并脱离御魂流居合形态", source.Name))
		}
		rt.DispatchOrientationChanges(before)
	}
	return player.TimingHookResult{}
}
