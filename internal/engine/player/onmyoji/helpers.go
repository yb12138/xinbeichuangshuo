package onmyoji

import "starcup-engine/internal/model"

// CanUseFactionCounter checks if the incoming card allows faction counter.
func CanUseFactionCounter(incoming *model.Card) bool {
	if incoming == nil {
		return false
	}
	if incoming.Name == "欺诈" {
		return false
	}
	return incoming.Faction != ""
}
