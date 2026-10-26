package butterfly_dancer_test

import (
	"testing"

	"starcup-engine/internal/engine"
	butterflydancer "starcup-engine/internal/engine/player/butterfly_dancer"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func newButterflySkillEventGame(t *testing.T, extra ...string) (*engine.GameEngine, *testutils.CaptureObserver) {
	t.Helper()
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	for _, id := range extra {
		if err := game.AddPlayer(id, "Enemy-"+id, "berserker", model.BlueCamp); err != nil {
			t.Fatal(err)
		}
	}
	game.State.Deck = rules.InitDeck()
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.ReturnTurnStage = model.TurnStageExtraAction
	return game, obs
}

func requireNoSkillActivation(t *testing.T, obs *testutils.CaptureObserver, skillIDs ...string) {
	t.Helper()
	for _, id := range skillIDs {
		if n := len(obs.SkillActivated(id)); n != 0 {
			t.Fatalf("expected no skill_activated for %s, got %d", id, n)
		}
	}
}

// drainButterflyFlow 拒绝所有后续可选响应并推进结算。
func drainButterflyFlow(t *testing.T, game *engine.GameEngine) {
	t.Helper()
	for i := 0; i < 30 && (len(game.State.PendingDamageQueue) > 0 || game.State.PendingInterrupt != nil); i++ {
		if intr := game.State.PendingInterrupt; intr != nil && intr.Type == model.InterruptChoice {
			sel := 0
			if ct, _ := intr.Context.(map[string]interface{})["choice_type"].(string); ct == "bt_wither_confirm" || ct == "bt_pilgrimage_confirm" {
				sel = 1
			}
			testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: intr.PlayerID, Type: model.CmdSelect, Selections: []int{sel}})
			continue
		}
		game.Drive()
	}
}

func TestButterflyPoisonPowder_EmitsSkillActivatedOnce(t *testing.T) {
	game, obs := newButterflySkillEventGame(t, "p2")
	p1 := game.State.Players["p1"]
	butterflydancer.AddCocoonCards(p1, []model.Card{butterflyTestCard("c1", model.CardTypeAttack, model.ElementWater)})
	game.AddPendingDamage(model.PendingDamage{SourceID: "p1", TargetID: "p2", Damage: 1, DamageType: model.MagicAttack})

	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "bt_poison_pick")
	requireNoSkillActivation(t, obs, "bt_poison_powder")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	drainButterflyFlow(t, game)

	testutils.RequireSingleSkillActivation(t, obs, "bt_poison_powder", "p1", "p2")
}

func TestButterflyPilgrimage_EmitsSkillActivatedOnce(t *testing.T) {
	game, obs := newButterflySkillEventGame(t, "p2")
	p1 := game.State.Players["p1"]
	butterflydancer.AddCocoonCards(p1, []model.Card{butterflyTestCard("c1", model.CardTypeAttack, model.ElementWater)})
	game.AddPendingDamage(model.PendingDamage{SourceID: "p2", TargetID: "p1", Damage: 1, DamageType: model.MagicAttack})

	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "bt_poison_pick")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}}) // 不发动毒粉
	testutils.RequireChoicePrompt(t, game, "p1", "bt_pilgrimage_confirm")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p1", "bt_pilgrimage_pick")
	requireNoSkillActivation(t, obs, "bt_pilgrimage")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	drainButterflyFlow(t, game)

	testutils.RequireSingleSkillActivation(t, obs, "bt_pilgrimage", "p1", "p1")
	requireNoSkillActivation(t, obs, "bt_poison_powder")
}

func TestButterflyMirror_EmitsSkillActivatedOnce(t *testing.T) {
	game, obs := newButterflySkillEventGame(t, "p2", "p3")
	p1 := game.State.Players["p1"]
	butterflydancer.AddCocoonCards(p1, []model.Card{
		butterflyTestCard("m1", model.CardTypeAttack, model.ElementFire),
		butterflyTestCard("m2", model.CardTypeAttack, model.ElementFire),
	})
	game.AddPendingDamage(model.PendingDamage{SourceID: "p2", TargetID: "p3", Damage: 2, DamageType: model.MagicAttack})

	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "bt_mirror_pair")
	requireNoSkillActivation(t, obs, "bt_mirror")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	drainButterflyFlow(t, game)

	testutils.RequireSingleSkillActivation(t, obs, "bt_mirror", "p1", "p3")
	requireNoSkillActivation(t, obs, "bt_wither", "bt_poison_powder")
}

// 镜花水月移除2张法术茧 → 凋零（发动1次、拒绝1次）：每个实际生效的技能各发一次事件。
func TestButterflyMirrorWitherChain_EachActivationEmitsOnce(t *testing.T) {
	game, obs := newButterflySkillEventGame(t, "p2", "p3")
	p1 := game.State.Players["p1"]
	butterflydancer.AddCocoonCards(p1, []model.Card{
		butterflyTestCard("m1", model.CardTypeMagic, model.ElementFire),
		butterflyTestCard("m2", model.CardTypeMagic, model.ElementFire),
	})
	game.AddPendingDamage(model.PendingDamage{SourceID: "p2", TargetID: "p3", Damage: 2, DamageType: model.MagicAttack})

	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "bt_mirror_pair")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	testutils.RequireChoicePrompt(t, game, "p1", "bt_wither_confirm")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p1", "bt_wither_target")
	requireNoSkillActivation(t, obs, "bt_wither")
	// 目标列表顺序 p1, p2, p3：选择 p3。
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{2}})
	// 第二张法术茧对应的凋零：拒绝。
	testutils.RequireChoicePrompt(t, game, "p1", "bt_wither_confirm")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	drainButterflyFlow(t, game)

	testutils.RequireSingleSkillActivation(t, obs, "bt_mirror", "p1", "p3")
	testutils.RequireSingleSkillActivation(t, obs, "bt_wither", "p1", "p3", "p1")
	requireNoSkillActivation(t, obs, "bt_poison_powder", "bt_pilgrimage")
}
