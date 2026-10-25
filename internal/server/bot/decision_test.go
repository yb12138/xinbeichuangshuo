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
