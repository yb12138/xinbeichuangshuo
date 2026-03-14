package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func mustHandleActionNoErr(t *testing.T, g *GameEngine, act model.PlayerAction) {
	t.Helper()
	if err := g.HandleAction(act); err != nil {
		t.Fatalf("handle action failed: %v (act=%+v)", err, act)
	}
}

// 回归：魔剑士【暗影抗拒】仅限制“自己行动阶段打出法术牌”，
// 非自己行动阶段仍可用【圣光】进行防御响应。
func TestMagicSwordsmanShadowReject_AllowHolyLightDefendOutsideOwnTurn(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "MS", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := g.AddPlayer("p2", "ATK", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = false
	p2.IsActive = true
	p1.Tokens["ms_shadow_form"] = 1
	p1.Hand = []model.Card{
		{ID: "holy_1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}
	p2.Hand = []model.Card{
		{ID: "atk_1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	g.State.CurrentTurn = 1
	g.State.Phase = model.PhaseActionSelection

	mustHandleActionNoErr(t, g, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdAttack,
		TargetID:  "p1",
		CardIndex: 0,
	})
	mustHandleActionNoErr(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	})

	if len(p1.Hand) != 0 {
		t.Fatalf("expected holy light consumed on defend, got hand=%d", len(p1.Hand))
	}
}

// 回归：魔剑士【暗影抗拒】下，非自己行动阶段仍可用【魔弹】进行魔弹链响应。
func TestMagicSwordsmanShadowReject_AllowMagicBulletCounterOutsideOwnTurn(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "MS", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := g.AddPlayer("p2", "Caster", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = false
	p2.IsActive = true
	p1.Tokens["ms_shadow_form"] = 1
	p1.Hand = []model.Card{
		{ID: "mb_1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "mb_src", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}

	g.State.CurrentTurn = 1
	g.State.Phase = model.PhaseActionSelection

	// 发起魔弹（无需手动选择目标，按顺序自动寻找对手）。
	mustHandleActionNoErr(t, g, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdMagic,
		CardIndex: 0,
	})
	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("expected magic missile interrupt, got %+v", g.State.PendingInterrupt)
	}

	// 目标 p1 在暗影形态下仍应可打出【魔弹】传递。
	mustHandleActionNoErr(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"counter"},
	})

	if len(p1.Hand) != 0 {
		t.Fatalf("expected magic bullet consumed on counter, got hand=%d", len(p1.Hand))
	}
}
