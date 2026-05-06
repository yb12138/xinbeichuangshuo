package magic_swordsman

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// MaybeReleaseShadowAtActionStart checks if a magic swordsman should
// automatically leave shadow form at the start of their action phase.
// Returns true if shadow form was released.
func MaybeReleaseShadowAtActionStart(rt model.IGameEngine, p *model.Player) bool {
	if rt == nil || p == nil || !player.IsCharacter(p, "magic_swordsman") {
		return false
	}
	if p.Tokens == nil {
		return false
	}
	if p.TurnState.HasUsedActionSkill {
		return false
	}
	if !player.HasForm(p, model.FormMagicSwordsmanShadow) {
		return false
	}
	player.ClearForm(p, model.FormMagicSwordsmanShadow)
	rt.Log(fmt.Sprintf("%s 脱离暗影形态并转正", p.Name))
	return true
}
