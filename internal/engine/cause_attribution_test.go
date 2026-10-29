package engine_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

// causesOf 返回带指定技能因果的事件（不含无负载的因果边界事件）。
func causesOf(obs *testutils.CaptureObserver, skillID string) []*model.Cause {
	var out []*model.Cause
	for _, ev := range obs.Events {
		if ev.Type == model.EventCauseBoundary {
			continue
		}
		if ev.Cause != nil && ev.Cause.SkillID == skillID {
			out = append(out, ev.Cause)
		}
	}
	return out
}

func firstDamageCause(obs *testutils.CaptureObserver, targetID string) *model.Cause {
	for _, ev := range obs.Events {
		if ev.Type == model.EventDamageDealt && ev.DamageDealt != nil && ev.DamageDealt.TargetID == targetID {
			return ev.Cause
		}
	}
	return nil
}

func skipPendingResponses(t *testing.T, game *engine.GameEngine) {
	t.Helper()
	for i := 0; i < 10 && game.State.PendingInterrupt != nil; i++ {
		intr := game.State.PendingInterrupt
		if intr.Type != model.InterruptResponseSkill && intr.Type != model.InterruptStartupSkill {
			return
		}
		testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: intr.PlayerID, Type: model.CmdSelect, Selections: []int{len(intr.SkillIDs)}})
	}
}

// 水之封印：封印师放下的封印在对手打出水系牌时引爆。
// 引爆的伤害应归属封印师的水之封印，Origin 指向当初放置封印的那次技能。
func TestCauseAttribution_WaterSealDetonationKeepsOrigin(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	_ = game.AddPlayer("p1", "Sealer", "sealer", model.RedCamp)
	_ = game.AddPlayer("p2", "Victim", "berserker", model.BlueCamp)
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1, p2 := game.State.Players["p1"], game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{{ID: "seal-card", Name: "水之封印·独有牌", Type: model.CardTypeMagic, Element: model.ElementWater, ExclusiveChar1: "sealer", ExclusiveSkill1: "水之封印"}}

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "water_seal", TargetIDs: []string{"p2"}, Selections: []int{0}})
	skipPendingResponses(t, game)
	if len(p2.Field) != 1 || p2.Field[0].Cause == nil || p2.Field[0].Cause.SkillID != "water_seal" {
		t.Fatalf("expected placed seal to record its placing cause, got %+v", p2.Field)
	}
	placeCause := p2.Field[0].Cause

	game.NextTurn()
	game.State.CurrentTurn = 1
	p1.IsActive = false
	p2.IsActive = true
	p2.TurnState = model.NewPlayerTurnState()
	game.State.TurnStage = model.TurnStageActionExecution
	p2.Hand = []model.Card{{ID: "c_water", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2}}
	obs.Reset()

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdAttack, TargetID: "p1", CardID: testutils.PlayableCardID(t, game, "p2", 0)})
	game.Drive()

	dmg := firstDamageCause(obs, "p2")
	if dmg == nil {
		t.Fatalf("expected seal damage on p2 with a cause")
	}
	if dmg.SkillID != "water_seal" || dmg.OwnerID != "p1" {
		t.Fatalf("expected seal damage attributed to sealer's water_seal, got %+v", dmg)
	}
	if dmg.OriginID != placeCause.ID {
		t.Fatalf("expected origin %s (placement), got %+v", placeCause.ID, dmg)
	}
}

// 朝圣：伤害 → 承伤 hook 推出确认 → 玩家选择 → 移除茧减伤。
// 跨两次中断后，移除茧的变化仍归属朝圣，且父因果是这次伤害的因果（攻击方技能）。
func TestCauseAttribution_PilgrimageAcrossInterrupts(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	_ = game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp)
	_ = game.AddPlayer("p2", "Mage", "berserker", model.BlueCamp)
	game.State.Deck = rules.InitDeck()
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.ReturnTurnStage = model.TurnStageExtraAction
	p1 := game.State.Players["p1"]
	p1.Field = append(p1.Field, &model.FieldCard{
		Card: model.Card{ID: "c1", Name: "茧", Type: model.CardTypeAttack, Element: model.ElementWater}, OwnerID: "p1", SourceID: "p1",
		Mode: model.FieldCover, Effect: model.EffectButterflyCocoon,
	})

	// 伤害来自 p2 的某个技能：用 WithSkillCause 模拟“技能造成的伤害”。
	game.WithSkillCause("berserker_frenzy", "p2", func() {
		game.AddPendingDamage(model.PendingDamage{SourceID: "p2", TargetID: "p1", Damage: 2, DamageType: model.AttackDamage})
	})
	attackCause := game.State.PendingDamageQueue[0].Cause
	if attackCause == nil || attackCause.SkillID != "berserker_frenzy" {
		t.Fatalf("expected pending damage to carry the attacker's cause, got %+v", attackCause)
	}

	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "bt_pilgrimage_confirm")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p1", "bt_pilgrimage_pick")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	for i := 0; i < 10 && (game.State.PendingInterrupt != nil || len(game.State.PendingDamageQueue) > 0); i++ {
		if intr := game.State.PendingInterrupt; intr != nil {
			testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: intr.PlayerID, Type: model.CmdSelect, Selections: []int{0}})
			continue
		}
		game.Drive()
	}

	pilgrimage := causesOf(obs, "bt_pilgrimage")
	if len(pilgrimage) == 0 {
		t.Fatalf("expected events attributed to bt_pilgrimage")
	}
	id := pilgrimage[0].ID
	for _, c := range pilgrimage {
		if c.ID != id {
			t.Fatalf("expected one pilgrimage cause across interrupts, got %s and %s", id, c.ID)
		}
		if c.OwnerID != "p1" || c.ParentID != attackCause.ID {
			t.Fatalf("expected pilgrimage owned by p1 with parent %s, got %+v", attackCause.ID, c)
		}
	}
	if dmg := firstDamageCause(obs, "p1"); dmg == nil || dmg.ID != attackCause.ID {
		t.Fatalf("expected the damage itself to stay with the attacker's cause, got %+v", dmg)
	}
}

// 毒粉：蝶舞者对 1 点法术伤害移除茧令其 +1。伤害修正应作为 damage_modified 归属毒粉。
func TestCauseAttribution_PoisonPowderEmitsDamageModified(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	_ = game.AddPlayer("p1", "Butterfly", "butterfly_dancer", model.RedCamp)
	_ = game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp)
	game.State.Deck = rules.InitDeck()
	game.State.CombatStage = model.CombatStageCalcDamage
	game.State.ReturnTurnStage = model.TurnStageExtraAction
	p1 := game.State.Players["p1"]
	p1.Field = append(p1.Field, &model.FieldCard{
		Card: model.Card{ID: "c1", Name: "茧", Type: model.CardTypeAttack, Element: model.ElementWater}, OwnerID: "p1", SourceID: "p1",
		Mode: model.FieldCover, Effect: model.EffectButterflyCocoon,
	})
	game.AddPendingDamage(model.PendingDamage{SourceID: "p1", TargetID: "p2", Damage: 1, DamageType: model.MagicAttack})
	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "bt_poison_pick")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})

	var marker *model.GameEvent
	for i := range obs.Events {
		ev := &obs.Events[i]
		if ev.Type == model.EventTimelineMarker && ev.TimelineMarker != nil && ev.TimelineMarker.NarrativeKind == "damage_modified" {
			marker = ev
		}
	}
	if marker == nil {
		t.Fatalf("expected damage_modified marker")
	}
	if marker.Cause == nil || marker.Cause.SkillID != "bt_poison_powder" || marker.TimelineMarker.Summary != "伤害 +1" {
		t.Fatalf("expected 伤害 +1 under bt_poison_powder, got cause=%+v marker=%+v", marker.Cause, marker.TimelineMarker)
	}
}
