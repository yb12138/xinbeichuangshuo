package server

import "starcup-engine/internal/server/catalog"

var availableRoles = catalog.AvailableRoles()

func isValidRole(role string) bool {
	return catalog.IsValidRole(role)
}

func buildCharacterViews() []CharacterView {
	return catalog.BuildCharacterViews()
}
