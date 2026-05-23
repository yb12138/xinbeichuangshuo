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
	)

	skills := room.buildAvailableActionSkills("p1")
	if !hasAvailableSkill(skills, "water_seal") {
		t.Fatalf("expected water_seal available when exclusive card exists")
	}
}

func TestBuildAvailableActionSkills_BloodPriestessSharedLifeRequiresExclusiveCard(t *testing.T) {
	room := NewRoom("BLOOD")
	room.Engine = engine.NewGameEngine(room)

	if err := room.Engine.AddPlayer("p1", "Blood", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	room.Engine.State.CurrentTurn = 0
	room.Engine.State.TurnStage = model.TurnStageActionExecution
	p1 := room.Engine.State.Players["p1"]
	p2 := room.Engine.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = nil
	p1.ExclusiveCards = nil

	skills := room.buildAvailableActionSkills("p1")
	if hasAvailableSkill(skills, "bp_shared_life") {
		t.Fatalf("expected shared life hidden without its exclusive card")
	}

	p1.ExclusiveCards = append(p1.ExclusiveCards, makeExclusiveSkillCard("shared-life-exclusive", "blood_priestess", "同生共死"))
	skills = room.buildAvailableActionSkills("p1")
	if !hasAvailableSkill(skills, "bp_shared_life") {
		t.Fatalf("expected shared life available while its exclusive card is unplaced")
	}
	for _, skill := range skills {
		if skill.ID == "bp_shared_life" && !skill.RequireExclusive {
			t.Fatalf("expected shared life availability metadata to require exclusive card")
		}
	}

	card, ok := p1.ConsumeExclusiveCard("blood_priestess", "同生共死")
	if !ok {
		t.Fatalf("expected to consume shared life exclusive card")
	}
	p2.Field = append(p2.Field, &model.FieldCard{
		Card:     card,
		OwnerID:  p2.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectBloodSharedLife,
		Hook:     model.FieldHookManual,
		Duration: -1,
	})

	skills = room.buildAvailableActionSkills("p1")
	if hasAvailableSkill(skills, "bp_shared_life") {
		t.Fatalf("expected shared life hidden after its exclusive card has been placed on field")
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

func TestBuildAvailableActionSkills_SoulRecallRequiresMagicCard(t *testing.T) {
	room := NewRoom("SOUL")
	room.Engine = engine.NewGameEngine(room)

	if err := room.Engine.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
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

	skills := room.buildAvailableActionSkills("p1")
	if hasAvailableSkill(skills, "ss_soul_recall") {
		t.Fatalf("expected soul recall hidden when no magic card exists")
	}

	p1.Hand = append(p1.Hand, model.Card{
		ID:      "magic-card",
		Name:    "法术牌",
		Type:    model.CardTypeMagic,
		Element: model.ElementWater,
	})
	skills = room.buildAvailableActionSkills("p1")
	if !hasAvailableSkill(skills, "ss_soul_recall") {
		t.Fatalf("expected soul recall available when magic card exists")
	}
}
