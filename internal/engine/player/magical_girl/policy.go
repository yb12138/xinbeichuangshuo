// gameflow: 魔法少女技能策略。

package magical_girl

import (
	"fmt"

	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

func validateFusionDiscard(ctx types.PolicyContext) error {
	if len(ctx.DiscardedCards) != 1 {
		return fmt.Errorf("魔弹融合需要选择1张地系或火系牌")
	}
	if !isFusionElement(ctx.DiscardedCards[0]) {
		return fmt.Errorf("魔弹融合需要选择地系或火系牌")
	}
	return nil
}

func isFusionElement(card model.Card) bool {
	return card.Element == model.ElementEarth || card.Element == model.ElementFire
}
