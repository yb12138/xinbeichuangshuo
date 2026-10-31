package engine_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

// 手牌为空、主动技都因 Handler 条件（形态 / 指示物 / 法术牌）不能发动时，
// 必须能宣告无法行动；否则技能全灰、攻击法术又没牌可打，玩家会卡在行动选择上。
func TestCannotAct_AllowedWhenSkillsBlockedByHandlerConditions(t *testing.T) {
	for _, role := range []string{"prayer_master", "soul_sorcerer"} {
		t.Run(role, func(t *testing.T) {
			obs := &actionPromptObserver{}
			g := engine.NewGameEngine(obs)
			if err := g.AddPlayer("p1", "Me", role, model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
				t.Fatal(err)
			}
			p1 := g.State.Players["p1"]
			p1.IsActive = true
			p1.TurnState = model.NewPlayerTurnState()
			p1.Hand = nil
			g.State.CurrentTurn = 0
			g.State.TurnStage = model.TurnStageActionExecution
			g.Drive()

			if obs.lastPrompt == nil || !promptOptionSet(obs.lastPrompt)["cannot_act"] {
				t.Fatalf("expected cannot_act option with empty hand and blocked skills, got %+v", obs.lastPrompt)
			}
			if err := g.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdCannotAct}); err != nil {
				t.Fatalf("cannot act rejected: %v", err)
			}
		})
	}
}
