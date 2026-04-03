package engine

import "starcup-engine/internal/model"

func (e *GameEngine) FindFieldEffectBySource(effect model.EffectType, sourceID string) (*model.Player, *model.FieldCard) {
	if e == nil || e.State == nil || sourceID == "" {
		return nil, nil
	}
	for _, pid := range e.State.PlayerOrder {
		holder := e.State.Players[pid]
		if holder == nil {
			continue
		}
		for _, fc := range holder.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != effect {
				continue
			}
			if fc.SourceID == sourceID {
				return holder, fc
			}
		}
	}
	return nil, nil
}

func buildElementCardIndexMap(player *model.Player) map[model.Element][]int {
	out := map[model.Element][]int{}
	if player == nil {
		return out
	}
	for i, c := range player.Hand {
		if c.Element == "" {
			continue
		}
		out[c.Element] = append(out[c.Element], i)
	}
	return out
}

func maxSameElementCount(player *model.Player) int {
	maxCount := 0
	for _, idxs := range buildElementCardIndexMap(player) {
		if len(idxs) > maxCount {
			maxCount = len(idxs)
		}
	}
	return maxCount
}

func distinctElementCount(player *model.Player) int {
	return len(buildElementCardIndexMap(player))
}

func elementOrderForPrompt() []model.Element {
	return []model.Element{
		model.ElementEarth,
		model.ElementWater,
		model.ElementFire,
		model.ElementWind,
		model.ElementThunder,
		model.ElementLight,
		model.ElementDark,
	}
}

func availableElementsByMinCount(player *model.Player, minCount int) []string {
	if minCount <= 0 {
		minCount = 1
	}
	elemMap := buildElementCardIndexMap(player)
	var out []string
	for _, ele := range elementOrderForPrompt() {
		if len(elemMap[ele]) >= minCount {
			out = append(out, string(ele))
		}
	}
	return out
}

func allHandIndices(player *model.Player) []int {
	if player == nil {
		return nil
	}
	out := make([]int, 0, len(player.Hand))
	for i := range player.Hand {
		out = append(out, i)
	}
	return out
}

func removeElementIndices(indices []int, player *model.Player, element model.Element, keepIndex int) []int {
	if len(indices) == 0 {
		return nil
	}
	var out []int
	for _, idx := range indices {
		if idx == keepIndex {
			continue
		}
		if idx < 0 || player == nil || idx >= len(player.Hand) {
			continue
		}
		if player.Hand[idx].Element == element {
			continue
		}
		out = append(out, idx)
	}
	return out
}

func (e *GameEngine) campEnemyIDs(camp model.Camp) []string {
	var ids []string
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp == camp {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

func (e *GameEngine) allOtherPlayerIDs(userID string) []string {
	var ids []string
	for _, pid := range e.State.PlayerOrder {
		if pid == userID {
			continue
		}
		if p := e.State.Players[pid]; p != nil {
			ids = append(ids, p.ID)
		}
	}
	return ids
}
