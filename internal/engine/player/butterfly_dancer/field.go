// gameflow: 蝶舞者场牌（FieldCard）管理。

package butterfly_dancer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ButterflyCocoonCap 茧盖牌上限。
const ButterflyCocoonCap = 8

// CocoonCovers 按场上顺序收集茧盖牌。
func CocoonCovers(p *model.Player) []*model.FieldCard {
	return player.CoverCardsByEffect(p, model.EffectButterflyCocoon)
}

// CocoonCount 统计茧盖牌数量。
func CocoonCount(p *model.Player) int {
	return player.CoverCountByEffect(p, model.EffectButterflyCocoon)
}

// SyncCocoonToken 将茧盖牌数量同步到 player.Tokens。
func SyncCocoonToken(p *model.Player) {
	player.EnsurePlayerTokensMap(p)
	p.Tokens["bt_cocoon_count"] = CocoonCount(p)
}

// AddCocoonCards 将卡牌以茧盖牌形式放入玩家场区，然后同步 token。
func AddCocoonCards(p *model.Player, cards []model.Card) {
	for _, c := range cards {
		p.AddFieldCard(&model.FieldCard{
			Card:     c,
			Mode:     model.FieldCover,
			Effect:   model.EffectButterflyCocoon,
			Hook:     model.FieldHookManual,
			OwnerID:  p.ID,
			SourceID:  p.ID,
		})
	}
	SyncCocoonToken(p)
}

// CocoonFieldIndices 收集场上所有茧盖牌的索引。
func CocoonFieldIndices(p *model.Player) []int {
	if p == nil {
		return nil
	}
	var indices []int
	for i, fc := range p.Field {
		if fc != nil && fc.Mode == model.FieldCover && fc.Effect == model.EffectButterflyCocoon {
			indices = append(indices, i)
		}
	}
	return indices
}

// RemoveCocoonByFieldIndex 按场上索引移除茧盖牌；返回被移除的牌及是否成功。
func RemoveCocoonByFieldIndex(p *model.Player, idx int) (*model.FieldCard, bool) {
	if p == nil || idx < 0 || idx >= len(p.Field) {
		return nil, false
	}
	fc := p.Field[idx]
	if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
		return nil, false
	}
	p.RemoveFieldCard(fc)
	return fc, true
}

// RemoveCocoonByFieldIndices 批量按场上索引移除茧盖牌，完成后同步 token。
func RemoveCocoonByFieldIndices(p *model.Player, indices []int) {
	if p == nil || len(indices) == 0 {
		return
	}
	// 先收集要移除的 FieldCard 指针，再统一移除，避免索引漂移。
	var toRemove []*model.FieldCard
	for _, idx := range indices {
		if idx < 0 || idx >= len(p.Field) {
			continue
		}
		fc := p.Field[idx]
		if fc != nil && fc.Mode == model.FieldCover && fc.Effect == model.EffectButterflyCocoon {
			toRemove = append(toRemove, fc)
		}
	}
	for _, fc := range toRemove {
		p.RemoveFieldCard(fc)
	}
	SyncCocoonToken(p)
}

// MirrorPairDefs 找出元素相同的茧盖牌对，返回每对在场上的索引。
func MirrorPairDefs(p *model.Player) [][2]int {
	if p == nil {
		return nil
	}
	// 按元素分组收集茧盖牌的场上索引。
	groups := map[model.Element][]int{}
	for i, fc := range p.Field {
		if fc != nil && fc.Mode == model.FieldCover && fc.Effect == model.EffectButterflyCocoon {
			groups[fc.Card.Element] = append(groups[fc.Card.Element], i)
		}
	}
	var pairs [][2]int
	for _, indices := range groups {
		for len(indices) >= 2 {
			pairs = append(pairs, [2]int{indices[0], indices[1]})
			indices = indices[2:]
		}
	}
	return pairs
}
