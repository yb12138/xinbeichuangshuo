package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func countDamageEvents(obs *captureObserver, sourceID, targetID, damageType string) int {
	if obs == nil {
		return 0
	}
	n := 0
	for _, ev := range obs.events {
		if ev.Type != model.EventDamageDealt {
			continue
		}
		data, ok := ev.Data.(map[string]interface{})
		if !ok {
			continue
		}
		src, _ := data["source_id"].(string)
		dst, _ := data["target_id"].(string)
		dt, _ := data["damage_type"].(string)
		if sourceID != "" && src != sourceID {
			continue
		}
		if targetID != "" && dst != targetID {
			continue
		}
		if damageType != "" && !strings.EqualFold(dt, damageType) {
			continue
		}
		n++
	}
	return n
}

// 回归：反噬仅在暗杀者承受“攻击伤害”时触发；
// 承受法术伤害时不应触发，也不能出现连锁死循环。
func TestAssassinBacklash_DoesNotTriggerOnMagicDamage(t *testing.T) {
	obs := &captureObserver{}
	game := NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Caster", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Assassin", "assassin", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection
	game.State.CurrentTurn = 0
	game.State.Players["p1"].TurnState = model.NewPlayerTurnState()
	game.State.Players["p2"].TurnState = model.NewPlayerTurnState()

	game.InflictDamage("p1", "p2", 1, "magic")
	game.Drive()

	if got := countDamageEvents(obs, "p1", "p2", "magic"); got != 1 {
		t.Fatalf("expected one magic damage event p1->p2, got %d", got)
	}
	if got := countDamageEvents(obs, "p2", "p1", "magic"); got != 0 {
		t.Fatalf("backlash should not trigger on magic damage, but got %d reflected events", got)
	}
	if got := countDamageEvents(obs, "p2", "p1", "backlash"); got != 0 {
		t.Fatalf("unexpected legacy backlash pseudo-damage events: %d", got)
	}
	if got := len(game.State.Players["p1"].Hand); got != 0 {
		t.Fatalf("expected p1 hand unchanged (0), got %d", got)
	}
	if got := len(game.State.Players["p2"].Hand); got != 1 {
		t.Fatalf("expected p2 draw 1 from magic damage, got %d", got)
	}
	if len(game.State.PendingDamageQueue) != 0 || game.State.PendingInterrupt != nil {
		t.Fatalf("expected clean resolution, pendingDamage=%d pendingInterrupt=%+v",
			len(game.State.PendingDamageQueue), game.State.PendingInterrupt)
	}
}

// 回归：反噬在承受攻击伤害后触发，并强制让攻击者摸1张牌（非伤害）。
func TestAssassinBacklash_TriggersOnAttackDamage(t *testing.T) {
	obs := &captureObserver{}
	game := NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Attacker", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Assassin", "assassin", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "atk-fire-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}

	if got := countDamageEvents(obs, "p1", "p2", "Attack"); got != 1 {
		t.Fatalf("expected one attack damage event p1->p2, got %d", got)
	}
	if got := countDamageEvents(obs, "p2", "p1", "magic"); got != 0 {
		t.Fatalf("backlash should not create magic damage event, got %d", got)
	}
	if got := countDamageEvents(obs, "p2", "p1", "backlash"); got != 0 {
		t.Fatalf("unexpected legacy backlash pseudo-damage events: %d", got)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected p1 final hand 1 (attack后承受1点反噬伤害摸1), got %d", got)
	}
	if got := len(p2.Hand); got != 1 {
		t.Fatalf("expected p2 final hand 1 (承受1点攻击伤害摸1), got %d", got)
	}
	if len(game.State.PendingDamageQueue) != 0 || game.State.PendingInterrupt != nil {
		t.Fatalf("expected clean resolution, pendingDamage=%d pendingInterrupt=%+v",
			len(game.State.PendingDamageQueue), game.State.PendingInterrupt)
	}
}
