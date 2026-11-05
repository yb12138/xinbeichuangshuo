package spirit_caster_test

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 审计 灵符师-3：雷鸣经灵力崩解结算时，伤害只路由一次，外层 EnterDamageResolution(nil) 不覆盖恢复点。
func TestSpiritCasterTalismanThunder_CollapseRoutesDamageOnce(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newSpiritCasterNoHandGame(t, obs, spiritCasterTestCard("t1", "雷符", model.CardTypeMagic, model.ElementThunder))
	game.State.Players["p1"].Crystal = 1
	if err := game.UseSkill("p1", "sc_talisman_thunder", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman thunder failed: %v", err)
	}
	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "sc_spiritual_collapse_confirm")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	game.Drive()
	for _, id := range []string{"p2", "p3"} {
		if got := countSpiritCasterDamageEvents(obs, "p1", id, 2, model.MagicAttack); got != 1 {
			t.Fatalf("expected thunder damage resolved once on %s, got %d", id, got)
		}
	}
}
