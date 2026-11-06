package moon_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 新月庇护：队友爆牌导致士气将下降 → 月之女神响应。
// 技能结算（形态变化、暗月）应归属新月庇护，持有者为月之女神；父因果为爆牌来源（此处无技能来源）。
func TestMoonGoddessNewMoonShelter_AttributesEffectsToShelterCause(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	_ = game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp)
	_ = game.AddPlayer("p2", "Ally", "angel", model.RedCamp)
	_ = game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp)
	ally := game.State.Players["p2"]
	ally.MaxHand = 4
	ally.Hand = []model.Card{
		moonTestCard("a1", "牌1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("a2", "牌2", model.CardTypeAttack, model.ElementWater),
		moonTestCard("a3", "牌3", model.CardTypeAttack, model.ElementWind),
		moonTestCard("a4", "牌4", model.CardTypeAttack, model.ElementThunder),
		moonTestCard("a5", "牌5", model.CardTypeMagic, model.ElementDark),
		moonTestCard("a6", "牌6", model.CardTypeMagic, model.ElementLight),
	}
	ctx := game.BuildContext(ally, nil, model.TimingActionDuring, nil)
	ctx.Flags["FromDamageDraw"] = true
	game.CheckHandLimitCtx(ally, ctx)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdSelect, Selections: []int{4, 5}})
	testutils.ChooseResponseSkillByID(t, game, "p1", "mg_new_moon_shelter")

	var shelter *model.Cause
	for _, ev := range obs.Events {
		if ev.Type == model.EventSkillActivated && ev.SkillActivated != nil && ev.SkillActivated.SkillID == "mg_new_moon_shelter" {
			shelter = ev.Cause
		}
	}
	if shelter == nil || shelter.SkillID != "mg_new_moon_shelter" || shelter.OwnerID != "p1" || shelter.SkillName != "新月庇护" {
		t.Fatalf("expected shelter skill event under its own cause owned by p1, got %+v", shelter)
	}
}
