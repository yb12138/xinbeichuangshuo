package bot

import (
	"testing"

	"starcup-engine/internal/model"
)

// 以下用例对应随机整局对战里机器人被引擎拒绝的几类动作：机器人应遵循引擎给出的规则判定。

func TestDecideAction_SkipsUntargetableAttackTarget(t *testing.T) {
	state := StateSnapshot{
		CurrentPlayer: "p1",
		Players: map[string]PlayerSnapshot{
			"p1": {ID: "p1", Camp: "Blue", MaxHand: 6, Hand: []model.Card{
				{ID: "a1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
			}},
			"p2": {ID: "p2", Camp: "Red", MaxHand: 6, HandCount: 1, NotAttackTargetable: true},
			"p3": {ID: "p3", Camp: "Red", MaxHand: 6, HandCount: 5},
		},
	}
	prompt := &model.Prompt{Type: model.PromptConfirm, Options: []model.PromptOption{{ID: "attack"}, {ID: "magic"}}}
	action, ok := DecideAction(DecisionInput{PlayerID: "p1", State: state, Prompt: prompt})
	if !ok || action.Type != model.CmdAttack || action.TargetID != "p3" {
		t.Fatalf("expected attack on the targetable enemy p3, got %+v", action)
	}
}

func TestDecideAction_NoMagicCardWhenFormBlocksMagic(t *testing.T) {
	state := StateSnapshot{
		CurrentPlayer: "p1",
		Players: map[string]PlayerSnapshot{
			"p1": {ID: "p1", Camp: "Blue", MaxHand: 6, CannotCastMagic: true, Hand: []model.Card{
				{ID: "m1", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementWater},
				{ID: "a1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
			}},
			"p2": {ID: "p2", Camp: "Red", MaxHand: 6, HandCount: 6},
		},
	}
	prompt := &model.Prompt{Type: model.PromptConfirm, Options: []model.PromptOption{{ID: "attack"}, {ID: "magic"}}}
	action, ok := DecideAction(DecisionInput{PlayerID: "p1", State: state, Prompt: prompt})
	if !ok || action.Type == model.CmdMagic {
		t.Fatalf("form blocks magic cards, bot must not cast one: %+v", action)
	}
}

func TestDecideAction_ForcedSkillIsUsed(t *testing.T) {
	state := StateSnapshot{
		CurrentPlayer: "p1",
		Players: map[string]PlayerSnapshot{
			"p1": {ID: "p1", Camp: "Blue", MaxHand: 6, Hand: []model.Card{
				{ID: "m1", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementWater},
			}},
			"p2": {ID: "p2", Camp: "Red", MaxHand: 6},
		},
	}
	prompt := &model.Prompt{Type: model.PromptConfirm, SkillID: "arbiter_doomsday", Options: []model.PromptOption{{ID: "magic"}}}
	skills := []AvailableSkill{{ID: "arbiter_doomsday", TargetType: int(model.TargetEnemy), MinTargets: 1, MaxTargets: 1}}
	action, ok := DecideAction(DecisionInput{PlayerID: "p1", State: state, Prompt: prompt, AvailableSkills: skills})
	if !ok || action.Type != model.CmdSkill || action.SkillID != "arbiter_doomsday" {
		t.Fatalf("expected forced 末日审判, got %+v", action)
	}
}

func TestDecideAction_MagicMissileCounterUsesEngineCards(t *testing.T) {
	state := StateSnapshot{
		Players: map[string]PlayerSnapshot{
			"p1": {ID: "p1", Camp: "Blue", MaxHand: 6, Hand: []model.Card{
				{ID: "a1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
				{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
			}},
			"p2": {ID: "p2", Camp: "Red", MaxHand: 6},
		},
	}
	prompt := &model.Prompt{
		Type:           model.PromptConfirm,
		Message:        "你成为了【魔弹】的目标，当前伤害为 2，请选择应对：",
		Options:        []model.PromptOption{{ID: "take"}, {ID: "counter"}},
		CounterCardIDs: []string{"mb"},
	}
	action, ok := DecideAction(DecisionInput{PlayerID: "p1", State: state, Prompt: prompt})
	if !ok {
		t.Fatalf("expected an action")
	}
	if action.Type == model.CmdRespond && len(action.ExtraArgs) > 0 && action.ExtraArgs[0] == "counter" && action.CardID != "mb" {
		t.Fatalf("magic missile must be passed with 魔弹, got card %s", action.CardID)
	}
}

func TestDecideAction_ChooseCardsByCardIDWhenOptionIDsAreNotHandIndexes(t *testing.T) {
	me := PlayerSnapshot{ID: "p1", Camp: "Blue", MaxHand: 6, Hand: []model.Card{
		{ID: "c0", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater},
	}}
	state := StateSnapshot{Players: map[string]PlayerSnapshot{"p1": me}}
	// 充盈：-1=不弃置，1=第一张候选（与手牌下标 0 不对应）
	prompt := &model.Prompt{
		Type: model.PromptChooseCards, Min: 1, Max: 1,
		Options: []model.PromptOption{{ID: "-1", Label: "不弃置"}, {ID: "1", CardID: "c0"}},
	}
	action, ok := DecideAction(DecisionInput{PlayerID: "p1", State: state, Prompt: prompt})
	if !ok || action.Type != model.CmdSelect || len(action.CardIDs) != 1 || action.CardIDs[0] != "c0" {
		t.Fatalf("expected select by card id c0, got %+v", action)
	}
}
