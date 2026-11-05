package onmyoji_test

import (
	"strings"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func newYinYangSameElementGame(t *testing.T, obs model.GameObserver) (*engine.GameEngine, *model.Player) {
	t.Helper()
	game := engine.NewGameEngine(obs)
	for _, p := range []struct {
		id, name, role string
		camp           model.Camp
	}{
		{"p1", "Attacker", "berserker", model.RedCamp},
		{"p2", "Onmyoji", "onmyoji", model.BlueCamp},
		{"p3", "RedAlly", "angel", model.RedCamp},
	} {
		if err := game.AddPlayer(p.id, p.name, p.role, p.camp); err != nil {
			t.Fatal(err)
		}
		game.State.Players[p.id].TurnState = model.NewPlayerTurnState()
	}
	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.Hand = []model.Card{{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2}}
	// 同系、同命格：既能普通应战，也能发动阴阳转换
	p2.Hand = []model.Card{{ID: "c2", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2}}
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0)}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	return game, p2
}

// 阴阳师-6：同系同命格的牌普通应战时不强制发动阴阳转换；玩家可选择不发动。
func TestOnmyojiYinYang_SameElementCounterCanDecline(t *testing.T) {
	game, p2 := newYinYangSameElementGame(t, testutils.NoopObserver{})
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_yinyang_confirm" {
		t.Fatalf("expected yinyang confirm offered for same-element same-faction card, got %q", got)
	}
	// 不发动
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdSelect, Selections: []int{1}})
	// 普通应战
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: "c2", TargetID: "p3"}); err != nil {
		t.Fatalf("normal counter failed: %v", err)
	}
	if got := p2.Tokens["onmyoji_ghost_fire"]; got != 0 {
		t.Fatalf("normal counter must not trigger yinyang shift, ghost fire=%d", got)
	}
	top := game.State.CombatStack[len(game.State.CombatStack)-1]
	if top.Card == nil || top.Card.Damage != 2 {
		t.Fatalf("normal counter should keep card damage 2, got %+v", top.Card)
	}
}

// 阴阳师-6：同系同命格的牌也可以选择发动阴阳转换。
// 阴阳师-5：结算顺序为 阴阳转换+1鬼火 → 式神转换摸1+1鬼火 → 脱离式神形态。
func TestOnmyojiYinYang_SameElementCounterCanActivateInOfficialOrder(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game, p2 := newYinYangSameElementGame(t, obs)
	p2.Form = model.FormOnmyojiShikigami
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdSelect, Selections: []int{0}})
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdSelect, CardIDs: []string{"c2"}})
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdSelect, Selections: []int{0}})

	if got := p2.Tokens["onmyoji_ghost_fire"]; got != 2 {
		t.Fatalf("expected ghost fire 0+1+1=2, got %d", got)
	}
	if p2.Form != "" {
		t.Fatalf("expected leave shikigami form, got %q", p2.Form)
	}
	top := game.State.CombatStack[len(game.State.CombatStack)-1]
	if top.Card == nil || top.Card.Damage != 2 {
		t.Fatalf("expected counter damage = ghost fire 2, got %+v", top.Card)
	}
	order := []string{"[阴阳转换] 触发，鬼火+1", "[式神转换] 触发", "脱离式神形态"}
	last := -1
	for _, want := range order {
		idx := -1
		for i, e := range obs.Events {
			if e.Type == model.EventLog && strings.Contains(e.Message, want) {
				idx = i
				break
			}
		}
		if idx < 0 || idx < last {
			t.Fatalf("expected log %q after previous step (idx=%d last=%d)", want, idx, last)
		}
		last = idx
	}
}
