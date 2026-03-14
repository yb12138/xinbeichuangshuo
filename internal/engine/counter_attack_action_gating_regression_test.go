package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func hasSkillID(skillIDs []string, want string) bool {
	for _, id := range skillIDs {
		if id == want {
			return true
		}
	}
	return false
}

func mustDo(t *testing.T, g *GameEngine, act model.PlayerAction) {
	t.Helper()
	if err := g.HandleAction(act); err != nil {
		t.Fatalf("handle action failed (%+v): %v", act, err)
	}
}

// 回归：应战攻击命中不应被当作“主动攻击命中”触发主动攻击类被动（如血色荆棘）。
func TestCounterHit_DoesNotTriggerActiveOnlyOnAttackHitSkills(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Counter", "crimson_sword_spirit", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Deck = rules.InitDeck()
	g.State.Phase = model.PhaseActionSelection

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p3 := g.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p3.Heal = 0
	p3.Hand = nil

	p1.Hand = []model.Card{
		{ID: "atk-p1-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "atk-p2-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"counter"},
		CardIndex: 0,
		TargetID:  "p3",
	})
	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p3",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if got := p2.Tokens["css_blood"]; got != 0 {
		t.Fatalf("expected css_blood=0 on counter hit, got %d", got)
	}
}

// 回归：应战攻击未命中不应触发“主动攻击未命中”类技能（如贯穿射击）。
func TestCounterMiss_DoesNotTriggerActiveOnlyOnAttackMissSkills(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "CounterArcher", "archer", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Deck = rules.InitDeck()
	g.State.Phase = model.PhaseActionSelection

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p3 := g.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()

	p1.Hand = []model.Card{
		{ID: "atk-p1-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "atk-p2-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "m-p2-water", Name: "水弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 1},
	}
	p3.Hand = []model.Card{
		{ID: "m-p3-holy", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	}

	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"counter"},
		CardIndex: 0,
		TargetID:  "p3",
	})
	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p3",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"defend"},
		CardIndex: 0,
	})

	if g.State.PendingInterrupt != nil &&
		g.State.PendingInterrupt.Type == model.InterruptResponseSkill &&
		g.State.PendingInterrupt.PlayerID == "p2" &&
		hasSkillID(g.State.PendingInterrupt.SkillIDs, "piercing_shot") {
		t.Fatalf("piercing_shot should not trigger on counter-attack miss")
	}
}

// 回归：应战命中后即使触发命中响应，也不应在“攻击行动结束”阶段继续触发主动攻击类连击技能。
func TestCounterHit_PhaseEndSkillsNotTriggeredForCounterAction(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "CounterValkyrie", "valkyrie", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Deck = rules.InitDeck()
	g.State.Phase = model.PhaseActionSelection

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p3 := g.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p2.Crystal = 1
	p2.Heal = 1 // 若误判为攻击行动结束，会弹出神圣追击

	p1.Hand = []model.Card{
		{ID: "atk-p1-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "atk-p2-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"counter"},
		CardIndex: 0,
		TargetID:  "p3",
	})
	mustDo(t, g, model.PlayerAction{
		PlayerID:  "p3",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	// 先确认应战命中后的英灵召唤（命中响应）路径，覆盖 finishTakeHit -> OnPhaseEnd 的分支。
	if g.State.PendingInterrupt == nil ||
		g.State.PendingInterrupt.Type != model.InterruptResponseSkill ||
		g.State.PendingInterrupt.PlayerID != "p2" ||
		!hasSkillID(g.State.PendingInterrupt.SkillIDs, "valkyrie_heroic_summon") {
		t.Fatalf("expected valkyrie_heroic_summon response after counter hit, got %+v", g.State.PendingInterrupt)
	}
	idx := -1
	for i, sid := range g.State.PendingInterrupt.SkillIDs {
		if sid == "valkyrie_heroic_summon" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("valkyrie_heroic_summon not found in pending skills")
	}
	mustDo(t, g, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{idx},
	})

	if g.State.PendingInterrupt != nil &&
		g.State.PendingInterrupt.Type == model.InterruptResponseSkill &&
		g.State.PendingInterrupt.PlayerID == "p2" &&
		hasSkillID(g.State.PendingInterrupt.SkillIDs, "valkyrie_divine_pursuit") {
		t.Fatalf("valkyrie_divine_pursuit should not trigger from counter-attack phase end")
	}
}
