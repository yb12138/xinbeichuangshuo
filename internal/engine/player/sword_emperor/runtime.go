// gameflow: 剑帝运行时：剑魂放置。

package sword_emperor

import (
	"starcup-engine/internal/model"
)

// PlaceSwordEmperorSwordSoul 尝试将 card 作为剑魂盖牌放置到玩家场上。
func PlaceSwordEmperorSwordSoul(player *model.Player, card model.Card) bool {
	if player == nil || SwordSoulCount(player) >= SwordSoulCap {
		return false
	}
	player.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  player.ID,
		SourceID: player.ID,
		Mode:     model.FieldCover,
		Effect:   model.EffectSwordSoul,
	})
	SyncSwordSoulToken(player)
	return true
}
