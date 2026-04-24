// gameflow: 运行时策略装配：combat rules。
// 策略包装器已迁移到角色包 PolicySpecs / TimingHookSpecs，核心逻辑保留在 engine 方法中。

package engine

import (
	onmyojiplayer "starcup-engine/internal/engine/player/onmyoji"
	"starcup-engine/internal/model"
)

// ---- 阴阳师战斗交互 ----

func (e *GameEngine) maybeOnmyojiDarkRitual(player *model.Player) bool {
	return onmyojiplayer.MaybeDarkRitual(newRoleChoiceRuntime(e), player)
}

func (e *GameEngine) applyOnmyojiFactionCounterBonuses(actor *model.Player, card *model.Card) {
	onmyojiplayer.ApplyFactionCounterBonuses(newRoleChoiceRuntime(e), actor, card)
}
