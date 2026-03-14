package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归：暗杀者在受伤摸牌前出现【水影】响应时，选择“跳过”后必须回到伤害结算流程，
// 不能停留在 PhaseResponse 导致 Drive 空转。
func TestAssassinWaterShadowSkip_ResumesPendingDamageResolution(t *testing.T) {
	game := NewGameEngine(&captureObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Assassin", "assassin", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	game.State.Deck = rules.InitDeck()

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// p1 打出主动攻击牌；p2 预留一张水系牌，确保可触发水影可选响应。
	p1.Hand = []model.Card{
		{ID: "atk-fire-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "water-card-1", Name: "水弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 1},
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

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response-skill interrupt after damage, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected response interrupt for p2, got %s", game.State.PendingInterrupt.PlayerID)
	}
	hasWaterShadow := false
	for _, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == "water_shadow" {
			hasWaterShadow = true
			break
		}
	}
	if !hasWaterShadow {
		t.Fatalf("expected water_shadow in response skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	// 机器人路径：Select 选择“跳过”（索引等于技能数量）
	skipIdx := len(game.State.PendingInterrupt.SkillIDs)
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{skipIdx},
	}); err != nil {
		t.Fatalf("skip response failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after skip, got %+v", game.State.PendingInterrupt)
	}
	if game.State.Phase == model.PhaseResponse {
		t.Fatalf("phase should not stay in response after skip (would stall drive)")
	}
}
