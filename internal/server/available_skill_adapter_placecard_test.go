package server

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

func makeExclusiveSkillCard(id, charID, skillTitle string) model.Card {
	return model.Card{
		ID:              id,
		Name:            skillTitle,
		Element:         model.ElementWater,
		ExclusiveChar1:  charID,
		ExclusiveSkill1: skillTitle,
	}
}

func TestBuildAvailableActionSkills_SealerPlaceCardSkillsNotBlockedByDeferredHandlerCanUse(t *testing.T) {
	room := NewRoom("SEALER")
	room.Engine = engine.NewGameEngine(room)

	if err := room.Engine.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	room.Engine.State.CurrentTurn = 0
	room.Engine.State.TurnStage = model.TurnStageActionExecution
	p1 := room.Engine.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()

	// 仅持有水/雷封印独有牌：预期可见这两个封印。
	p1.Hand = append(p1.Hand,
		makeExclusiveSkillCard("water-exclusive", "sealer", "水之封印"),
		makeExclusiveSkillCard("thunder-exclusive", "sealer", "雷之封印"),
	)

	skills := room.buildAvailableActionSkills("p1")
	if !hasAvailableSkill(skills, "water_seal") {
		t.Fatalf("expected water_seal available when exclusive card exists")
	}
	if !hasAvailableSkill(skills, "thunder_seal") {
		t.Fatalf("expected thunder_seal available when exclusive card exists")
	}
}

func TestBuildAvailableActionSkills_HeroTauntStillRequiresAngerToken(t *testing.T) {
	room := NewRoom("HERO")
	room.Engine = engine.NewGameEngine(room)

	if err := room.Engine.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	room.Engine.State.CurrentTurn = 0
	room.Engine.State.TurnStage = model.TurnStageActionExecution
	p1 := room.Engine.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = append(p1.Hand, makeExclusiveSkillCard("hero-taunt-exclusive", "hero", "挑衅"))

	// 怒气不足时不可见（防止“PlaceCard 跳过 CanUse”误伤 Manual 类技能）。
	p1.Tokens["hero_anger"] = 0
	skills := room.buildAvailableActionSkills("p1")
	if hasAvailableSkill(skills, "hero_taunt") {
		t.Fatalf("expected hero_taunt hidden when hero_anger=0")
	}

	p1.Tokens["hero_anger"] = 1
	skills = room.buildAvailableActionSkills("p1")
	if !hasAvailableSkill(skills, "hero_taunt") {
		t.Fatalf("expected hero_taunt available when hero_anger>0 and exclusive card exists")
	}
}

func TestBuildAvailableActionSkills_AngelCleanseExposedAsNoTargetSkill(t *testing.T) {
	room := NewRoom("ANGEL")
	room.Engine = engine.NewGameEngine(room)

	if err := room.Engine.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	room.Engine.State.CurrentTurn = 0
	room.Engine.State.TurnStage = model.TurnStageActionExecution
	p1 := room.Engine.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = append(p1.Hand, model.Card{
		ID:      "wind-card",
		Name:    "风牌",
		Type:    model.CardTypeMagic,
		Element: model.ElementWind,
	})

	skills := room.buildAvailableActionSkills("p1")
	for _, skill := range skills {
		if skill.ID != "angel_cleanse" {
			continue
		}
		if skill.TargetType != int(model.TargetNone) || skill.MinTargets != 0 || skill.MaxTargets != 0 {
			t.Fatalf("expected angel_cleanse no-target metadata, got target_type=%d min=%d max=%d", skill.TargetType, skill.MinTargets, skill.MaxTargets)
		}
		return
	}
	t.Fatalf("expected angel_cleanse available when owning wind card")
}
