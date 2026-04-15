package server

import "starcup-engine/internal/server/catalog"

func isValidRole(role string) bool {
	return catalog.IsValidRole(role)
}

func buildCharacterViews() []CharacterView {
	return catalog.BuildCharacterViews()
}
