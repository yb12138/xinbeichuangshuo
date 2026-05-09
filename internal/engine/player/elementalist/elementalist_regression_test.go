package elementalist_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"strings"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func elementalistExclusiveCard(owner *model.Player, skillTitle string, element model.Element) model.Card {
	charID := "elementalist"
	faction := "咏"
	if owner != nil && owner.Character != nil {
		charID = owner.Character.ID
		faction = owner.Character.Faction
	}
	return model.Card{
		ID:              "elem-exclusive-" + skillTitle,
		Name:            skillTitle,
		Type:            model.CardTypeMagic,
		Element:         element,
		Faction:         faction,
		Damage:          0,
		Description:     "元素师独有技测试卡",
		ExclusiveChar1:  charID,
		ExclusiveSkill1: skillTitle,
	}
}

func TestElementalistFreeze_StepByStepTargetSelection(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		elementalistExclusiveCard(p1, "冰冻", model.ElementFire),
	}
	p2.Hand = nil
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = rules.InitDeck()

	// 发动冰冻，应进入分步选择流程（不再需要一次性选2个目标）
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "elementalist_freeze",
		TargetIDs:  nil,  // 分步模式不需要预先选目标
		Selections: []int{0},
	})

	// 检查第一步：选择法术伤害目标
	testutils.RequireChoiceContext(t, game, "p1", "elementalist_freeze_damage_target")
	prompt := game.BuildPendingInterruptPrompt()
	if prompt == nil {
		t.Fatalf("expected freeze damage target prompt")
	}
	if !strings.Contains(prompt.Message, "法术伤害目标") {
		t.Fatalf("expected damage target hint in message, got: %s", prompt.Message)
	}

	// 选择伤害目标（p2）
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},  // p2 是第二个选项（p1=0, p2=1）
	})

	// 检查第二步：选择治疗目标
	testutils.RequireChoiceContext(t, game, "p1", "elementalist_freeze_heal_target")
	prompt = game.BuildPendingInterruptPrompt()
	if prompt == nil {
		t.Fatalf("expected freeze heal target prompt")
	}
	if !strings.Contains(prompt.Message, "治疗目标") {
		t.Fatalf("expected heal target hint in message, got: %s", prompt.Message)
	}

	// 选择治疗目标（p1 自己）
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},  // p1 是第一个选项
	})

	// 验证效果：p1 治疗+1
	if got := p1.Heal; got != 1 {
		t.Fatalf("expected freeze heal target gain 1 heal, got %d", got)
	}
}

func TestElementalistMoonlight_ConsumesGemAndRequiresGem(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Hand = nil
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = rules.InitDeck()

	p1.Gem = 0
	p1.Crystal = 3
	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_moonlight",
		TargetIDs: []string{"p2"},
	})
	if err == nil || !strings.Contains(err.Error(), "资源不足") {
		t.Fatalf("expected moonlight require gem, got err=%v", err)
	}

	p1.Gem = 1
	p1.Crystal = 2
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_moonlight",
		TargetIDs: []string{"p2"},
	})

	if got := p1.Gem; got != 0 {
		t.Fatalf("expected moonlight consume 1 gem, got %d", got)
	}
	if got := len(p2.Hand); got != 3 {
		t.Fatalf("expected moonlight damage=3 after paying gem (remaining energy 2), got hand=%d", got)
	}
}

func TestElementalistIgnite_RequiresThreeElement(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["element"] = 2
	p2.Hand = nil
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = rules.InitDeck()

	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_ignite",
		TargetIDs: []string{"p2"},
	})
	if err == nil || !strings.Contains(err.Error(), "元素不足") {
		t.Fatalf("expected ignite reject when element<3, got err=%v", err)
	}

	p1.Tokens["element"] = 3
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_ignite",
		TargetIDs: []string{"p2"},
	})

	if got := p1.Tokens["element"]; got != 0 {
		t.Fatalf("expected ignite consume 3 element, got %d", got)
	}
	if got := len(p2.Hand); got != 2 {
		t.Fatalf("expected ignite deal 2 damage (draw 2), got hand=%d", got)
	}
	if p1.TurnState.CurrentExtraAction != "Magic" && len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected ignite grant extra magic action, current=%q pending=%d", p1.TurnState.CurrentExtraAction, len(p1.TurnState.PendingActions))
	}
}

func TestElementalistOffenseSkills_RejectAllyTargets(t *testing.T) {
	skillCard := func(owner *model.Player, skillTitle string, element model.Element) []int {
		owner.Hand = []model.Card{elementalistExclusiveCard(owner, skillTitle, element)}
		return []int{0}
	}

	tests := []struct {
		name    string
		skillID string
		setup   func(p1 *model.Player) []int
	}{
		{
			name:    "ignite",
			skillID: "elementalist_ignite",
			setup: func(p1 *model.Player) []int {
				p1.Tokens["element"] = 3
				return nil
			},
		},
		{
			name:    "thunder_strike",
			skillID: "elementalist_thunder_strike",
			setup: func(p1 *model.Player) []int {
				return skillCard(p1, "雷击", model.ElementThunder)
			},
		},
		{
			name:    "wind_blade",
			skillID: "elementalist_wind_blade",
			setup: func(p1 *model.Player) []int {
				return skillCard(p1, "风刃", model.ElementWind)
			},
		},
		{
			name:    "meteor",
			skillID: "elementalist_meteor",
			setup: func(p1 *model.Player) []int {
				return skillCard(p1, "陨石", model.ElementEarth)
			},
		},
		{
			name:    "fireball",
			skillID: "elementalist_fireball",
			setup: func(p1 *model.Player) []int {
				return skillCard(p1, "火球", model.ElementFire)
			},
		},
		{
			name:    "moonlight",
			skillID: "elementalist_moonlight",
			setup: func(p1 *model.Player) []int {
				p1.Gem = 1
				p1.Crystal = 0
				return nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			game := engine.NewGameEngine(testutils.NoopObserver{})
			if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
				t.Fatal(err)
			}

			p1 := game.State.Players["p1"]
			p1.IsActive = true
			p1.TurnState = model.NewPlayerTurnState()
			game.State.CurrentTurn = 0
			game.State.TurnStage = model.TurnStageActionExecution
			selections := tc.setup(p1)

			err := game.HandleAction(model.PlayerAction{
				PlayerID:   "p1",
				Type:       model.CmdSkill,
				SkillID:    tc.skillID,
				TargetIDs:  []string{"p2"},
				Selections: selections,
			})
			if err == nil || !strings.Contains(err.Error(), "skill can only target enemies") {
				t.Fatalf("expected ally target rejection, got err=%v", err)
			}
		})
	}
}

func TestElementalistThunderStrike_BonusFlow_DirectCardPickWithCancel(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		elementalistExclusiveCard(p1, "雷击", model.ElementThunder),
		{ID: "thunder-bonus", Name: "雷系弃牌", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
	}
	p2.Hand = nil
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = rules.InitDeck()

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "elementalist_thunder_strike",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	})
	testutils.RequireChoiceContext(t, game, "p1", "elementalist_bonus_card")
	prompt := game.BuildPendingInterruptPrompt()
	if prompt == nil {
		t.Fatalf("expected elementalist bonus prompt")
	}
	if prompt.Type != model.PromptChooseCards {
		t.Fatalf("expected choose_cards prompt, got %s", prompt.Type)
	}
	if got := len(prompt.Options); got < 2 {
		t.Fatalf("expected card options plus cancel option, got %d", got)
	}

	beforeCancelHand := len(p1.Hand)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})

	if got := len(p1.Hand); got != beforeCancelHand {
		t.Fatalf("expected cancel keep hand unchanged, got %d -> %d", beforeCancelHand, got)
	}
	if got := len(p2.Hand); got != 1 {
		t.Fatalf("expected cancel path resolve base 1 magic damage, got hand=%d", got)
	}
}

func TestElementalistThunderStrike_BonusFlow_DirectCardPickWithConfirm(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		elementalistExclusiveCard(p1, "雷击", model.ElementThunder),
		{ID: "thunder-bonus", Name: "雷系弃牌", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
	}
	p2.Hand = nil
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = rules.InitDeck()

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "elementalist_thunder_strike",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	})
	testutils.RequireChoiceContext(t, game, "p1", "elementalist_bonus_card")
	beforeSelectHand := len(p1.Hand)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})

	if got := len(p1.Hand); got != beforeSelectHand-1 {
		t.Fatalf("expected bonus discard consume 1 card, got %d -> %d", beforeSelectHand, got)
	}
	if got := len(p2.Hand); got != 2 {
		t.Fatalf("expected bonus path resolve 2 magic damage, got hand=%d", got)
	}
}
