package war_homunculus_test

import (
	"testing"

	"starcup-engine/internal/engine"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 英灵人形-4：双重回响不限于攻击/法术伤害，中毒等其他伤害类型也能触发。
func TestHomDualEcho_NotLimitedByDamageType(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Hom", "war_homunculus", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1

	h := skills.GetHandler("hom_dual_echo")
	if h == nil {
		t.Fatalf("hom_dual_echo handler not found")
	}
	damageVal := 1
	ctx := g.BuildContext(p1, g.State.Players["p2"], model.TimingDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  p1.ID,
		TargetID:  "p2",
		DamageVal: &damageVal,
	})
	if ctx.Selections == nil {
		ctx.Selections = map[string]interface{}{}
	}
	ctx.Selections["damage_type"] = "poison"
	if !h.CanUse(ctx) {
		t.Fatalf("expected dual echo usable for poison damage")
	}
}
