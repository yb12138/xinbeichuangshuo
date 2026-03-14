package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
)

func countFieldEffect(p *model.Player, effect model.EffectType) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == effect {
			count++
		}
	}
	return count
}

func TestPerformMagic_PoisonCannotStackOnSameTarget(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "A", "angel", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "B", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.Hand = []model.Card{
		{ID: "poison-1", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementEarth},
		{ID: "poison-2", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	game.State.Phase = model.PhaseActionExecution

	if err := game.PerformMagic("p1", "p2", 0); err != nil {
		t.Fatalf("first poison should succeed, got err=%v", err)
	}
	if got := countFieldEffect(p2, model.EffectPoison); got != 1 {
		t.Fatalf("expected 1 poison after first cast, got %d", got)
	}

	if err := game.PerformMagic("p1", "p2", 0); err == nil || !strings.Contains(err.Error(), "已有中毒") {
		t.Fatalf("second poison should be rejected by duplicate rule, got err=%v", err)
	}
	if got := countFieldEffect(p2, model.EffectPoison); got != 1 {
		t.Fatalf("expected poison to remain single instance, got %d", got)
	}
}

func TestUseSkill_BasicEffectPlacementCannotStack(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Ally", "saintess", model.RedCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.Hand = []model.Card{
		{
			ID:              "wall-a",
			Name:            "圣盾",
			Type:            model.CardTypeMagic,
			Element:         model.ElementLight,
			Faction:         "圣",
			ExclusiveChar1:  "天使",
			ExclusiveSkill1: "天使之墙",
		},
		{
			ID:              "wall-b",
			Name:            "圣盾",
			Type:            model.CardTypeMagic,
			Element:         model.ElementEarth,
			Faction:         "圣",
			ExclusiveChar1:  "天使",
			ExclusiveSkill1: "天使之墙",
		},
	}
	game.State.Phase = model.PhaseActionSelection

	// 第一次放置【天使之墙】成功。
	if err := game.UseSkill("p1", "angel_wall", []string{"p2"}, []int{0}); err != nil {
		t.Fatalf("first angel_wall should succeed, got err=%v", err)
	}
	if got := countFieldEffect(p2, model.EffectShield); got != 1 {
		t.Fatalf("expected 1 shield after first angel_wall, got %d", got)
	}

	// 回到行动阶段后再次尝试同目标放置，应被“基础效果不可叠加”拦截。
	p1.IsActive = true
	game.State.Phase = model.PhaseActionSelection
	if err := game.UseSkill("p1", "angel_wall", []string{"p2"}, []int{0}); err == nil || !strings.Contains(err.Error(), "同种基础效果") {
		t.Fatalf("second angel_wall should be rejected by duplicate rule, got err=%v", err)
	}
	if got := countFieldEffect(p2, model.EffectShield); got != 1 {
		t.Fatalf("expected shield to remain single instance, got %d", got)
	}
}
