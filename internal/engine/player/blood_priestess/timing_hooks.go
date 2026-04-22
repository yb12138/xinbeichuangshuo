// gameflow: 血祭司士气损失 Timing Hook 实现。

package blood_priestess

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
)

// moraleLossHook 血祭司：伤害导致士气下降时进入流血形态并获得1点治疗。
func moraleLossHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	if !ctx.FromDamageDraw || ctx.MoraleLoss <= 0 {
		return engineplayer.TimingHookResult{}
	}
	victim := rt.GetPlayer(ctx.TargetID)
	if victim == nil || !engineplayer.IsCharacter(victim, "blood_priestess") {
		return engineplayer.TimingHookResult{}
	}
	if InBleedingForm(victim) {
		return engineplayer.TimingHookResult{}
	}
	beforePoses := rt.SnapshotPlayerPoses()
	EnterBleedingForm(victim)
	rt.Log(fmt.Sprintf("%s 的 [流血] 触发：因承受伤害导致我方士气下降，进入流血形态", victim.Name))
	rt.DispatchOrientationChanges(beforePoses)
	rt.Heal(victim.ID, 1)
	rt.Log(fmt.Sprintf("%s 的 [流血] 触发：获得1点治疗", victim.Name))
	return engineplayer.TimingHookResult{}
}
