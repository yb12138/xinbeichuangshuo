package server

import (
	"sort"
	"starcup-engine/internal/data"
	"starcup-engine/internal/model"
	"strings"
	"testing"
)

func TestAvailableRolesSyncWithCharacterData(t *testing.T) {
	charSet := map[string]bool{}
	for _, c := range data.GetCharacters() {
		charSet[c.ID] = true
	}

	roleSet := map[string]bool{}
	for _, roleID := range availableRoles {
		roleSet[roleID] = true
	}

	var missingInRoom []string
	for charID := range charSet {
		if !roleSet[charID] {
			missingInRoom = append(missingInRoom, charID)
		}
	}
	sort.Strings(missingInRoom)
	if len(missingInRoom) > 0 {
		t.Fatalf("availableRoles 缺少角色配置: %v", missingInRoom)
	}

	var unknownInRoom []string
	for roleID := range roleSet {
		if !charSet[roleID] {
			unknownInRoom = append(unknownInRoom, roleID)
		}
	}
	sort.Strings(unknownInRoom)
	if len(unknownInRoom) > 0 {
		t.Fatalf("availableRoles 存在无效角色ID: %v", unknownInRoom)
	}
}

func TestValidateLineupRejectsDuplicateRoles(t *testing.T) {
	room := NewRoom("DUP_ROLE")
	room.Clients = map[string]*Client{
		"p1": {PlayerID: "p1", Name: "A", Camp: model.RedCamp, CharRole: "berserker"},
		"p2": {PlayerID: "p2", Name: "B", Camp: model.BlueCamp, CharRole: "berserker"},
	}

	err := room.validateLineupLocked()
	if err == nil {
		t.Fatalf("expected duplicate role error, got nil")
	}
	if !strings.Contains(err.Error(), "角色不可重复") {
		t.Fatalf("expected duplicate role error message, got: %v", err)
	}
}
