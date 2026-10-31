package bot

import (
	"testing"

	"starcup-engine/internal/model"
)

// 引擎不接受 Pass（无效的行动类型），机器人在行动选择无事可做时必须宣告无法行动，
// 否则会被引擎拒绝而卡局。
func TestDecideActionSelectionNeverPasses(t *testing.T) {
	state := StateSnapshot{
		CurrentPlayer: "p1",
		Players: map[string]PlayerSnapshot{
			"p1": {ID: "p1", Camp: "Blue", MaxHand: 6},
			"p2": {ID: "p2", Camp: "Red", MaxHand: 6},
		},
	}
	// 只给出 attack/magic 选项，但手上没有任何攻击/法术牌、也没有可用技能
	prompt := &model.Prompt{
		Type:    model.PromptConfirm,
		Options: []model.PromptOption{{ID: "attack"}, {ID: "magic"}},
	}

	action, ok := DecideAction(DecisionInput{PlayerID: "p1", State: state, Prompt: prompt})
	if !ok {
		t.Fatalf("expected an action")
	}
	if action.Type == model.CmdPass {
		t.Fatalf("bot must not send Pass; engine rejects it")
	}
	if action.Type != model.CmdCannotAct {
		t.Fatalf("expected CannotAct fallback, got %s", action.Type)
	}
}

func TestBuildFallbackActionNeverPasses(t *testing.T) {
	state := StateSnapshot{
		Players: map[string]PlayerSnapshot{
			"p1": {ID: "p1", Camp: "Blue", MaxHand: 6},
		},
	}
	prompt := &model.Prompt{
		Type:    model.PromptConfirm,
		Options: []model.PromptOption{{ID: "attack"}},
	}

	action, ok := BuildFallbackAction("p1", prompt, state)
	if !ok {
		t.Fatalf("expected a fallback action")
	}
	if action.Type == model.CmdPass {
		t.Fatalf("fallback must not send Pass; engine rejects it")
	}
}

// 额外行动现在总带"跳过"选项；机器人手里有攻击牌时应继续攻击，而不是白白放弃额外行动。
func TestDecideActionSelectionUsesExtraActionInsteadOfSkipping(t *testing.T) {
	state := StateSnapshot{
		CurrentPlayer: "p1",
		Players: map[string]PlayerSnapshot{
			"p1": {ID: "p1", Camp: "Blue", MaxHand: 6, Hand: []model.Card{
				// 低收益攻击：走不到"高收益优先"分支，才能检验"跳过"不会抢在攻击前面
				{ID: "a1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
			}},
			"p2": {ID: "p2", Camp: "Red", MaxHand: 6, HandCount: 6},
		},
	}
	prompt := &model.Prompt{
		Type:    model.PromptConfirm,
		Options: []model.PromptOption{{ID: "attack"}, {ID: "cannot_act", Label: "跳过额外行动"}},
	}

	action, ok := DecideAction(DecisionInput{PlayerID: "p1", State: state, Prompt: prompt})
	if !ok {
		t.Fatalf("expected an action")
	}
	if action.Type != model.CmdAttack {
		t.Fatalf("expected the bot to attack with its extra action, got %s", action.Type)
	}
}
