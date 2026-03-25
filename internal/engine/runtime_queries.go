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
