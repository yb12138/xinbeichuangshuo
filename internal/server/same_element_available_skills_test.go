package server

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

// 需要“同系 / 同命格手牌”才能发动的主动技：不满足时必须以 disabled 下发，满足后可用。
func TestBuildAvailableActionSkills_SameElementSkillsGated(t *testing.T) {
	cases := []struct {
		role, skillID string
		setup         func(p *model.Player)
		without, with []model.Card
	}{
		{
			role: "adventurer", skillID: "adventurer_fraud",
			without: []model.Card{
				{ID: "a", Type: model.CardTypeAttack, Element: model.ElementWater},
				{ID: "b", Type: model.CardTypeAttack, Element: model.ElementFire},
			},
			with: []model.Card{{ID: "c", Type: model.CardTypeAttack, Element: model.ElementWater}},
		},
		{
			role: "plague_mage", skillID: "plague_death_touch",
			setup: func(p *model.Player) { p.Heal = 2 },
			without: []model.Card{
				{ID: "a", Type: model.CardTypeAttack, Element: model.ElementWater},
				{ID: "b", Type: model.CardTypeMagic, Element: model.ElementFire},
			},
			with: []model.Card{{ID: "c", Type: model.CardTypeMagic, Element: model.ElementWater}},
		},
		{
			role: "onmyoji", skillID: "onmyoji_shikigami_descend",
			without: []model.Card{
				{ID: "a", Type: model.CardTypeAttack, Element: model.ElementWater, Faction: "圣"},
				{ID: "b", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "血"},
			},
			with: []model.Card{{ID: "c", Type: model.CardTypeAttack, Element: model.ElementWind, Faction: "圣"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.skillID, func(t *testing.T) {
			room := NewRoom("SAME")
			room.Engine = engine.NewGameEngine(room)
			_ = room.Engine.AddPlayer("p1", "Me", tc.role, model.BlueCamp)
			_ = room.Engine.AddPlayer("p2", "Enemy", "berserker", model.RedCamp)
			room.Engine.State.CurrentTurn = 0
			room.Engine.State.TurnStage = model.TurnStageActionExecution
			p1 := room.Engine.State.Players["p1"]
			p1.IsActive = true
			p1.TurnState = model.NewPlayerTurnState()
			if tc.setup != nil {
				tc.setup(p1)
			}
			p1.Hand = tc.without

			var found *AvailableSkill
			skills := room.buildAvailableActionSkills("p1")
			for i := range skills {
				if skills[i].ID == tc.skillID {
					found = &skills[i]
				}
			}
			if found == nil || !found.Disabled || found.DisabledReason == "" {
				t.Fatalf("expected %s listed as disabled with reason, got %+v", tc.skillID, found)
			}

			p1.Hand = append(append([]model.Card{}, tc.without...), tc.with...)
			if !hasAvailableSkill(room.buildAvailableActionSkills("p1"), tc.skillID) {
				t.Fatalf("expected %s usable once a matching pair is in hand", tc.skillID)
			}
		})
	}
}
