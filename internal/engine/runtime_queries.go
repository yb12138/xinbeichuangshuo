// gameflow: 引擎只读查询（供 handler 或 UI）。

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
