package beast_samurai_test

import (
	"reflect"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestBeastSamurai_BeastSoulWillEmitsSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game, _, _ := newBeastSamuraiTestEngine(t, obs, "")
	attackCard := model.Card{ID: "bs-hit", Name: "试斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "武", Damage: 2}
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.PendingDamageQueue = []model.PendingDamage{
		{SourceID: "p1", TargetID: "p2", Damage: 2, DamageType: model.AttackDamage, Card: &attackCard},
	}

	if paused := game.ProcessPendingDamages(); paused {
		t.Fatalf("expected hit to resolve without extra interrupt")
	}

	var events []*model.SkillActivatedPayload
	for _, ev := range obs.Events {
		if ev.Type == model.EventSkillActivated && ev.SkillActivated != nil && ev.SkillActivated.SkillID == "bs_beast_soul_will" {
			events = append(events, ev.SkillActivated)
		}
	}
	if len(events) != 1 {
		t.Fatalf("expected exactly one bs_beast_soul_will skill_activated, got %d", len(events))
	}
	if events[0].PlayerID != "p1" || !reflect.DeepEqual(events[0].TargetIDs, []string{"p1"}) {
		t.Fatalf("unexpected payload: %+v", events[0])
	}
}

// 兽魂意念只在兽灵武士自己主动攻击命中时触发。
// 命中后 hook 按"攻击方或目标任一是兽灵武士"分派，被别人打中时 SourceID 是攻击者，不能把兽魂加到攻击者身上。
func TestBeastSamurai_BeastSoulWillIgnoresHitsOnTheSamurai(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game, samurai, attacker := newBeastSamuraiTestEngine(t, obs, "holy_bow")
	samuraiSoulBefore := samurai.Tokens["bs_beast_soul"]
	attackerSoulBefore := attacker.Tokens["bs_beast_soul"]
	attackCard := model.Card{ID: "hb-hit", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Faction: "圣", Damage: 2}
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.PendingDamageQueue = []model.PendingDamage{
		{SourceID: attacker.ID, TargetID: samurai.ID, Damage: 2, DamageType: model.AttackDamage, Card: &attackCard},
	}

	game.ProcessPendingDamages()

	for _, ev := range obs.Events {
		if ev.Type == model.EventSkillActivated && ev.SkillActivated != nil && ev.SkillActivated.SkillID == "bs_beast_soul_will" {
			t.Fatalf("兽魂意念 should not fire when the samurai is the one hit: %+v", ev.SkillActivated)
		}
	}
	if got := samurai.Tokens["bs_beast_soul"]; got != samuraiSoulBefore {
		t.Fatalf("samurai beast soul changed from %d to %d", samuraiSoulBefore, got)
	}
	if got := attacker.Tokens["bs_beast_soul"]; got != attackerSoulBefore {
		t.Fatalf("attacker (holy_bow) gained beast soul: %d -> %d", attackerSoulBefore, got)
	}
}
