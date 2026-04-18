// gameflow: 开局发角色专属牌/盖牌。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func ensureExclusiveStarterCard(player *model.Player, skillTitle string, buildCard func() model.Card) bool {
	charID := player.Character.ID
	for _, c := range player.ExclusiveCards {
		if c.MatchExclusive(charID, skillTitle) {
			return false
		}
	}
	player.ExclusiveCards = append(player.ExclusiveCards, buildCard())
	return true
}

// ensureStarterRoleCards 为特定角色补充开局自带专属技能卡（置于专属卡区，不占手牌）。
func (e *GameEngine) ensureStarterRoleCards(player *model.Player) {
	if player == nil || player.Character == nil {
		return
	}
	entry := roleRegistry.Entry(player.Character.ID)
	if entry.ID == "" || entry.StarterCards == nil {
		return
	}
	cards := entry.ApplyStarterCards(player)
	for _, card := range cards {
		if ensureExclusiveStarterCard(player, card.Name, func() model.Card { return card }) {
			e.Log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【%s】（专属卡区）", player.Name, card.Name))
		}
	}
}
