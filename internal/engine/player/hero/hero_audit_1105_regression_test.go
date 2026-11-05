package hero_test

import (
	"testing"

	"starcup-engine/internal/engine"
	heroplayer "starcup-engine/internal/engine/player/hero"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 勇者-3：禁断之力缺少攻击上下文 / 命中分支缺少伤害值时整体不结算：不扣水晶、不弃牌、不自伤、不进精疲力竭。
func TestHeroForbiddenPower_MissingContextDoesNotHalfResolve(t *testing.T) {
	cases := []struct {
		name  string
		event *model.EventContext
	}{
		{name: "nil_attack_info", event: &model.EventContext{Type: model.EventAttack, SourceID: "p1", TargetID: "p2"}},
		{name: "hit_nil_damage_val", event: &model.EventContext{
			Type: model.EventAttack, SourceID: "p1", TargetID: "p2",
			AttackInfo: &model.AttackEventInfo{ActionType: string(model.ActionAttack), IsHit: true},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			game := engine.NewGameEngine(testutils.NoopObserver{})
			if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
				t.Fatal(err)
			}
			p1 := game.State.Players["p1"]
			p1.Crystal = 1
			p1.Gem = 0
			p1.Hand = []model.Card{
				heroTestCard("f1", "火法术", model.CardTypeMagic, model.ElementFire, 0),
				heroTestCard("f2", "火攻击", model.CardTypeAttack, model.ElementFire, 2),
			}
			ctx := game.BuildContext(p1, game.State.Players["p2"], model.TimingAttackHit, tc.event)
			h := &heroplayer.HeroForbiddenPowerHandler{}
			if err := h.Execute(ctx); err == nil {
				t.Fatalf("expected error when context is incomplete")
			}
			if p1.Crystal != 1 {
				t.Fatalf("crystal must not be spent, got %d", p1.Crystal)
			}
			if len(p1.Hand) != 2 {
				t.Fatalf("hand must not be discarded, got %d", len(p1.Hand))
			}
			if len(game.State.PendingDamageQueue) != 0 {
				t.Fatalf("self damage must not be queued, got %+v", game.State.PendingDamageQueue)
			}
			if p1.Form == model.FormHeroExhaustion || p1.Tokens["hero_anger"] != 0 {
				t.Fatalf("must not enter exhaustion or gain anger: form=%q anger=%d", p1.Form, p1.Tokens["hero_anger"])
			}
		})
	}
}

// 勇者-4（复现尝试）：精疲力竭在完整回合轮转中，下个行动阶段开始时即解除，不被 HasUsedActionSkill 推迟。
func TestHeroExhaustion_ReleasedAtNextActionStartThroughTurnRotation(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Form = model.FormHeroExhaustion
	p1.Tokens["hero_exhaustion_release_pending"] = 1
	p1.Hand = nil
	p2.Hand = nil
	game.State.Deck = []model.Card{
		heroTestCard("d1", "补1", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("d2", "补2", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d3", "补3", model.CardTypeAttack, model.ElementEarth, 2),
	}
	// 从 p2 的回合结束轮转到 p1 的新回合
	p2.IsActive = true
	p2.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 1
	game.NextTurn()
	game.Drive()

	if game.State.CurrentTurn != 0 {
		t.Fatalf("expected hero turn, got %d", game.State.CurrentTurn)
	}
	if p1.Form == model.FormHeroExhaustion {
		t.Fatalf("expected exhaustion released at hero's action start, still exhausted (state=%s)", game.RuntimeStateLabel())
	}
	if !game.IsActionSelectionWindow() {
		t.Fatalf("expected action selection after release, got %s", game.RuntimeStateLabel())
	}
}

// 勇者-5（复现尝试）：明镜止水的攻击未命中（被防御）时，攻击结束同样 +1 水晶，待发计数清零不残留。
func TestHeroCalmMind_MissStillGrantsCrystalAndNoResidue(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 0
	p1.Gem = 0
	p1.Tokens["hero_wisdom"] = 4
	p1.Hand = []model.Card{heroTestCard("atk", "雷斩", model.CardTypeAttack, model.ElementThunder, 2)}
	p2.Hand = []model.Card{{ID: "sh", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Description: "圣光"}}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0)})
	testutils.ChooseResponseSkillByID(t, game, "p1", "hero_calm_mind")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"defend"}, CardID: testutils.PlayableCardID(t, game, "p2", 0)})
	// 未命中后可能弹出禁断之力等响应，统统跳过
	for i := 0; i < 4 && game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptResponseSkill && game.State.PendingInterrupt.PlayerID == "p1"; i++ {
		testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	}
	if got := p1.Crystal; got != 1 {
		t.Fatalf("expected calm mind +1 crystal at attack end even on miss, got %d (state=%s)", got, game.RuntimeStateLabel())
	}
	if got := p1.TurnState.SkillFlowState["hero_calm_end_crystal_pending"]; got != 0 {
		t.Fatalf("expected calm mind pending cleared, got %d", got)
	}
}

// 勇者-6（复现尝试）：被挑衅者攻击勇者是合法的主动攻击目标，挑衅随有效攻击移除。
func TestHeroTaunt_TauntedPlayerCanAttackHero(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hero_anger"] = 1
	p1.ExclusiveCards = []model.Card{heroTauntExclusiveCard(p1)}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	if err := game.UseSkill("p1", "hero_taunt", []string{"p2"}, nil); err != nil {
		t.Fatalf("use hero_taunt failed: %v", err)
	}
	if !game.CanBeActiveAttackTarget("p1") {
		t.Fatalf("hero should always be a valid active attack target")
	}
	p1.IsActive = false
	p2.IsActive = true
	p2.TurnState = model.NewPlayerTurnState()
	p2.Hand = []model.Card{heroTestCard("a1", "光刃", model.CardTypeAttack, model.ElementLight, 2)}
	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageTurnStart
	game.Drive()
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdAttack, TargetID: "p1", CardID: testutils.PlayableCardID(t, game, "p2", 0)})
	if testutils.GetFieldEffectCard(p2, model.EffectHeroTaunt) != nil {
		t.Fatalf("expected taunt removed after attacking hero")
	}
}
