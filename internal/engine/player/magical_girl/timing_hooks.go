// gameflow: 魔法少女 timing hooks。

package magical_girl

import (
	engineplayer "starcup-engine/internal/engine/player"
)

func magicMissileResponseSkillAugHook(_ engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := ctx.Player
	chain := ctx.MagicBulletChain
	if player == nil || chain == nil || chain.TargetID != player.ID {
		return engineplayer.TimingHookResult{}
	}
	if hasParticipatedInMagicBulletChain(player.ID, chain) || !hasFusionCard(player) {
		return engineplayer.TimingHookResult{SkillIDs: append([]string{}, ctx.OfferedSkillIDs...)}
	}
	skillIDs := append([]string{}, ctx.OfferedSkillIDs...)
	skillIDs = append(skillIDs, "magic_bullet_fusion_chain")
	return engineplayer.TimingHookResult{SkillIDs: skillIDs}
}
