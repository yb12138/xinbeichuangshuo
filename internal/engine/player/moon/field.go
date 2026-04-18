// gameflow: 月神场牌（FieldCard）管理。

package moon

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// DarkMoonCovers 按场上顺序收集暗月盖牌。
func DarkMoonCovers(p *model.Player) []*model.FieldCard {
	return player.CoverCardsByEffect(p, model.EffectMoonDarkMoon)
}

// DarkMoonCount 统计暗月盖牌数量；若数量降至 0 且玩家非 nil 则自动退出暗月形态。
func DarkMoonCount(p *model.Player) int {
	covers := player.CoverCardsByEffect(p, model.EffectMoonDarkMoon)
	count := len(covers)
	if count <= 0 && p != nil {
		LeaveDarkMoonForm(p)
	}
	return count
}

// AddDarkMoonCards 将卡牌以暗月盖牌形式放入玩家场区，并在添加后尝试进入暗月形态。
func AddDarkMoonCards(p *model.Player, cards []model.Card) int {
	added := 0
	for _, c := range cards {
		p.AddFieldCard(&model.FieldCard{
			Card:     c,
			Mode:     model.FieldCover,
			Effect:   model.EffectMoonDarkMoon,
			Hook:     model.FieldHookManual,
			OwnerID:  p.ID,
			SourceID:  p.ID,
		})
		added++
	}
	if added > 0 {
		EnterDarkMoonForm(p)
	}
	return DarkMoonCount(p)
}
