package server

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/stateview"
)

func newCauseFlowRoom(t *testing.T, roles map[string]string, camps map[string]model.Camp, order []string) *Room {
	t.Helper()
	room := NewRoom("CAUSE_FLOW")
	room.Engine = engine.NewGameEngine(room)
	for _, id := range order {
		if err := room.Engine.AddPlayer(id, id, roles[id], camps[id]); err != nil {
			t.Fatal(err)
		}
	}
	if err := room.Engine.StartGame(); err != nil {
		t.Fatal(err)
	}
	return room
}

func setActivePlayer(room *Room, playerID string) {
	st := room.Engine.State
	for i, id := range st.PlayerOrder {
		st.Players[id].IsActive = id == playerID
		if id == playerID {
			st.CurrentTurn = i
		}
	}
}

func replayFlows(room *Room) []ActionFlowDTO {
	return buildReplayActionFlows(room.snapshotTimelineHistory(), room.timelineDisplayName)
}

func skillNodes(flow ActionFlowDTO, skillID string) []ActionFlowNodeDTO {
	var out []ActionFlowNodeDTO
	for _, node := range flow.Nodes {
		if node.Kind == "skill" && node.SkillID == skillID {
			out = append(out, node)
		}
	}
	return out
}

func hasDelta(deltas []TimelineDelta, field string, value int) bool {
	for _, d := range deltas {
		if d.Field == field && d.Value == value {
			return true
		}
	}
	return false
}

// 蛹化：宣言与结算同属一个因果 → 只有一个技能节点，且挂着该技能产生的公开状态变化。
func TestActionFlowCause_ChrysalisSingleNodeWithDeltas(t *testing.T) {
	room := newCauseFlowRoom(t,
		map[string]string{"p1": "butterfly_dancer", "p2": "berserker"},
		map[string]model.Camp{"p1": model.BlueCamp, "p2": model.RedCamp},
		[]string{"p1", "p2"})
	setActivePlayer(room, "p1")
	room.Engine.State.Players["p1"].Gem = 2

	if err := room.Engine.UseSkill("p1", "bt_chrysalis", nil, nil); err != nil {
		t.Fatal(err)
	}

	var flow *ActionFlowDTO
	for _, f := range replayFlows(room) {
		if len(skillNodes(f, "bt_chrysalis")) > 0 {
			f := f
			flow = &f
		}
	}
	if flow == nil {
		t.Fatalf("expected a flow with the chrysalis skill node")
	}
	nodes := skillNodes(*flow, "bt_chrysalis")
	if len(nodes) != 1 {
		t.Fatalf("expected exactly one chrysalis node, got %d: %+v", len(nodes), flow.Nodes)
	}
	node := nodes[0]
	if node.CauseID == "" || node.ActorUserID != "p1" || node.EffectText == "" {
		t.Fatalf("expected cause-backed node with actor and effect text, got %+v", node)
	}
	if !hasDelta(node.Deltas, "token:bt_pupa", 1) || !hasDelta(node.Deltas, "gem", -1) {
		t.Fatalf("expected pupa +1 and gem -1 attributed to chrysalis, got %+v", node.Deltas)
	}
	for _, n := range flow.Nodes {
		if n.Kind == "effect" && n.Label == "" {
			t.Fatalf("unexpected unnamed effect node: %+v", n)
		}
		if n.Kind == "effect" && n.Label == "效果" {
			t.Fatalf("unexpected placeholder effect node: %+v", n)
		}
	}
}

// 式神咒束代为应战：服务端把 attack_taken_over 作为结果挂在阴阳师的技能节点上。
func TestActionFlowCause_OnmyojiTakeoverBecomesNodeEffect(t *testing.T) {
	room := newCauseFlowRoom(t,
		map[string]string{"p1": "berserker", "p2": "angel", "p3": "onmyoji", "p4": "angel"},
		map[string]model.Camp{"p1": model.RedCamp, "p2": model.BlueCamp, "p3": model.BlueCamp, "p4": model.RedCamp},
		[]string{"p1", "p2", "p3", "p4"})
	st := room.Engine.State
	p3 := st.Players["p3"]
	p3.Form = model.FormOnmyojiShikigami
	p3.Hand = []model.Card{{ID: "c1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2}}
	st.BlueGems, st.BlueCrystals = 1, 1
	st.PendingInterrupt = nil
	st.InterruptQueue = nil
	st.CombatStack = []model.CombatRequest{{
		AttackerID: "p1", TargetID: "p2", CanBeResponded: true,
		Card: &model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
	}}
	req := st.CombatStack[0]
	if !room.Engine.RunAttackResponseCombatInteractionPolicies(&req) {
		t.Fatalf("binding should start")
	}
	for i := 0; i < 3; i++ {
		if err := room.Engine.HandleAction(model.PlayerAction{PlayerID: "p3", Type: model.CmdSelect, Selections: []int{0}}); err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}

	for _, flow := range replayFlows(room) {
		for _, node := range skillNodes(flow, "onmyoji_binding") {
			for _, eff := range node.Effects {
				if eff.Kind == "attack_taken_over" && len(eff.TargetIDs) == 2 && eff.TargetIDs[0] == "p2" && eff.TargetIDs[1] == "p4" {
					if node.ActorUserID != "p3" {
						t.Fatalf("expected binding node owned by p3, got %+v", node)
					}
					return
				}
			}
		}
	}
	t.Fatalf("expected a binding node with attack_taken_over p2→p4; flows=%+v", replayFlows(room))
}

// 凋零生效：作为公开指示物下发（同步状态与时间线快照一致），且置位那一刻的差值归属凋零。
func TestPublicIndicators_WitherIsPublicAndAttributed(t *testing.T) {
	room := newCauseFlowRoom(t,
		map[string]string{"p1": "butterfly_dancer", "p2": "berserker"},
		map[string]model.Camp{"p1": model.BlueCamp, "p2": model.RedCamp},
		[]string{"p1", "p2"})
	p1 := room.Engine.State.Players["p1"]
	room.resetPublicTimelineSnapshot()

	room.Engine.WithSkillCause("bt_wither", "p1", func() {
		p1.TurnState.SkillFlowState["bt_wither_active"] = 1
		room.Engine.Log("凋零生效")
	})

	if got := stateview.PublicIndicators(room.Engine.State, p1)["bt_wither_active"]; got != 1 {
		t.Fatalf("expected bt_wither_active public indicator, got %d", got)
	}
	found := false
	for _, ev := range room.snapshotTimelineHistory() {
		for _, d := range ev.Deltas {
			if d.Field == "indicator:bt_wither_active" && d.Value == 1 {
				if ev.Cause == nil || ev.Cause.SkillID != "bt_wither" {
					t.Fatalf("expected wither delta under bt_wither cause, got %+v", ev.Cause)
				}
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("expected an indicator delta for bt_wither_active")
	}
}

// 地下法则：冒险者购买改为阵营 +2 宝石；阵营宝石差值应归属地下法则（前端据此从冒险者座位飞出宝石）。
func TestActionFlowCause_UndergroundLawCampGemsAttributed(t *testing.T) {
	room := newCauseFlowRoom(t,
		map[string]string{"p1": "adventurer", "p2": "berserker"},
		map[string]model.Camp{"p1": model.BlueCamp, "p2": model.RedCamp},
		[]string{"p1", "p2"})
	setActivePlayer(room, "p1")
	st := room.Engine.State
	st.PendingInterrupt = nil
	st.InterruptQueue = nil
	st.TurnStage = model.TurnStageActionExecution
	// 购买摸 3 张：先清空手牌，避免超过手牌上限被拒绝。
	st.Players["p1"].Hand = nil
	room.resetPublicTimelineSnapshot()

	if err := room.Engine.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdBuy}); err != nil {
		t.Fatalf("buy: %v", err)
	}
	for _, ev := range room.snapshotTimelineHistory() {
		for _, d := range ev.Deltas {
			if d.Type == "team_gem" && d.Value == 2 {
				if ev.Cause == nil || ev.Cause.SkillID != "adventurer_underground_law" || ev.Cause.OwnerID != "p1" {
					t.Fatalf("expected team gem +2 under 地下法则 owned by p1, got %+v", ev.Cause)
				}
				return
			}
		}
	}
	t.Fatalf("expected a team_gem +2 delta after buy")
}
