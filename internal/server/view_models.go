package server

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/viewmodel"
)

type CharacterView = viewmodel.CharacterView

type SkillView = viewmodel.SkillView

type RoomEvent = viewmodel.RoomEvent

type PlayerInfo = viewmodel.PlayerInfo

type AvailableSkill = viewmodel.AvailableSkill

type GameStateUpdate = viewmodel.GameStateUpdate

type PlayerView = viewmodel.PlayerView

type lineupPlayer struct {
	id   string
	name string
	role string
	camp model.Camp
}
