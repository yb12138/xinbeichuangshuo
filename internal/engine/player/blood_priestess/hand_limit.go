// gameflow: 血之巫女同生共死手牌上限修改器。

package blood_priestess

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// SharedLifeHandLimitModifier 同生共死的手牌上限修改。
// 遍历场上所有同生共死效果，根据血之巫女形态计算 delta。
func SharedLifeHandLimitModifier(engine engineplayer.HandLimitModifierEngine, target *model.Player) int {
	delta := 0
	for _, holder := range engine.GetAllPlayers() {
		if holder == nil {
			continue
		}
		for _, fieldCard := range holder.Field {
			if fieldCard == nil || fieldCard.Mode != model.FieldEffect || fieldCard.Effect != model.EffectBloodSharedLife {
				continue
			}
			source := engine.LookupPlayer(fieldCard.SourceID)
			if source == nil || !engineplayer.IsCharacter(source, "blood_priestess") {
				continue
			}
			change := -2
			if engineplayer.HasForm(source, model.FormBloodPriestessBleeding) {
				change = 1
			}
			if source.ID == target.ID {
				delta += change
				continue
			}
			if fieldCard.OwnerID == target.ID {
				if !engine.HasFixedMaxHandCap(target) {
					delta += change
				}
			}
		}
	}
	return delta
}
