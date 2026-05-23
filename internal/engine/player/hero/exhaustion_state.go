package hero

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const heroExhaustionReleasePendingKey = "hero_exhaustion_release_pending"

func markExhaustionReleasePending(p *model.Player) {
	engineplayer.SetToken(p, heroExhaustionReleasePendingKey, 1)
	engineplayer.SetSkillFlowState(p, heroExhaustionReleasePendingKey, 1)
}

func hasExhaustionReleasePending(p *model.Player) bool {
	return engineplayer.GetToken(p, heroExhaustionReleasePendingKey) > 0 ||
		engineplayer.GetSkillFlowState(p, heroExhaustionReleasePendingKey) > 0
}

func clearExhaustionReleasePending(p *model.Player) {
	engineplayer.SetToken(p, heroExhaustionReleasePendingKey, 0)
	engineplayer.SetSkillFlowState(p, heroExhaustionReleasePendingKey, 0)
}
