package butterfly_dancer

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const PupaCap = -1

func Pupa(p *model.Player) int {
	return engineplayer.TokenValue(p, "bt_pupa", PupaCap)
}

func AddPupa(p *model.Player, delta int) int {
	return engineplayer.AddToken(p, "bt_pupa", delta, PupaCap)
}

// WitherActive 检查蝴蝶舞者凋零状态是否激活。
func WitherActive(p *model.Player) bool {
	if p == nil || !engineplayer.IsCharacter(p, "butterfly_dancer") {
		return false
	}
	engineplayer.EnsurePlayerSkillFlowState(p)
	return p.TurnState.SkillFlowState["bt_wither_active"] > 0
}
