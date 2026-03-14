package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func findSkillIndex(skillIDs []string, want string) int {
	for i, sid := range skillIDs {
		if sid == want {
			return i
		}
	}
	return -1
}

// 回归：圣枪骑士在“当前治疗高于MaxHeal”的场景下，地枪X上限应以当前治疗值为准（最多4）。
func TestHolyLancerEarthSpear_MaxXUsesCurrentHealValue(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyLancer", "holy_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens = map[string]int{
		"holy_lancer_prayer_used_turn": 1, // 屏蔽天枪前置询问，聚焦地枪链路
	}
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0
	p2.Hand = nil

	// 构造“当前治疗超过MaxHeal”的常见场景（例如圣光祈愈后）。
	p1.MaxHeal = 3
	p1.Heal = 4
	p1.Hand = []model.Card{
		{ID: "atk-fire-1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	mustDo(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustDo(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response-skill interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected interrupt for p1, got %s", game.State.PendingInterrupt.PlayerID)
	}
	idx := findSkillIndex(game.State.PendingInterrupt.SkillIDs, "holy_lancer_earth_spear")
	if idx < 0 {
		t.Fatalf("expected holy_lancer_earth_spear in pending skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	mustDo(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{idx},
	})

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected earth-spear choice interrupt, got %+v", game.State.PendingInterrupt)
	}
	ctxData, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("earth-spear context type mismatch")
	}
	ct, _ := ctxData["choice_type"].(string)
	if ct != "holy_lancer_earth_spear_x" {
		t.Fatalf("expected choice_type holy_lancer_earth_spear_x, got %q", ct)
	}

	maxX := 0
	if v, ok := ctxData["max_x"].(int); ok {
		maxX = v
	} else if f, ok := ctxData["max_x"].(float64); ok {
		maxX = int(f)
	}
	if maxX != 4 {
		t.Fatalf("expected earth-spear max_x=4, got %d", maxX)
	}

	prompt := game.GetCurrentPrompt()
	if prompt == nil {
		t.Fatalf("expected prompt for earth-spear choice")
	}
	if len(prompt.Options) != 4 {
		t.Fatalf("expected 4 options for earth-spear X, got %d", len(prompt.Options))
	}
}

// 回归：地枪选择X后，应恢复攻击命中后续结算流程，不应卡在ActionSelection空转。
func TestHolyLancerEarthSpear_SelectXResumesAttackFlow(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyLancer", "holy_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens = map[string]int{
		"holy_lancer_prayer_used_turn": 1, // 屏蔽天枪前置询问，聚焦地枪链路
	}
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0
	p2.Hand = nil

	p1.MaxHeal = 3
	p1.Heal = 4
	p1.Hand = []model.Card{
		{ID: "atk-fire-1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	mustDo(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustDo(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response-skill interrupt, got %+v", game.State.PendingInterrupt)
	}
	idx := findSkillIndex(game.State.PendingInterrupt.SkillIDs, "holy_lancer_earth_spear")
	if idx < 0 {
		t.Fatalf("expected holy_lancer_earth_spear in pending skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	mustDo(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{idx},
	})

	// 选第3项（0-based索引=2），期望X=3。
	mustDo(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{2},
	})

	if p1.Heal != 1 {
		t.Fatalf("expected p1 heal to decrease to 1 after X=3, got %d", p1.Heal)
	}
	// 基础1伤害 + 地枪3 = 4，受击者应摸4张牌。
	if len(p2.Hand) != 4 {
		t.Fatalf("expected p2 hand=4 after earth-spear boosted hit, got %d", len(p2.Hand))
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after earth-spear resolution, got %+v", game.State.PendingInterrupt)
	}
}

// 回归：圣光祈愈“本回合已用”标记应在真正结束回合时清理，额外行动未消耗前不得提前清理。
func TestHolyLancerPrayerToken_ClearedAtRealTurnEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyLancer", "holy_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseTurnEnd

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	if p1.Tokens == nil {
		p1.Tokens = map[string]int{}
	}
	p1.Tokens["holy_lancer_prayer_used_turn"] = 1
	p1.TurnState.PendingActions = []model.ActionContext{
		{Source: "TestExtraAction", MustType: "Attack"},
	}
	p1.Hand = []model.Card{
		{ID: "atk-extra-1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	// 第一次进入TurnEnd时，因仍有额外行动，不应清理标记。
	game.Drive()
	if got := p1.Tokens["holy_lancer_prayer_used_turn"]; got != 1 {
		t.Fatalf("expected prayer token to remain during pending extra actions, got %d", got)
	}
	if game.State.Phase != model.PhaseActionSelection {
		t.Fatalf("expected to enter ActionSelection for extra action, got %s", game.State.Phase)
	}

	// 额外行动耗尽后再次进入TurnEnd，才应清理并切到下一回合。
	p1.TurnState.PendingActions = nil
	game.State.Phase = model.PhaseTurnEnd
	game.Drive()

	if got := p1.Tokens["holy_lancer_prayer_used_turn"]; got != 0 {
		t.Fatalf("expected prayer token cleared at real turn end, got %d", got)
	}
	if game.State.CurrentTurn != 1 {
		t.Fatalf("expected turn to advance to next player, got current turn index %d", game.State.CurrentTurn)
	}
}

// 回归：响应技能弹窗构建前应实时剔除已失效技能（例如治疗不足时的天枪）。
func TestResponsePrompt_PrunesInvalidHolyLancerSkySpear(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyLancer", "holy_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.Heal = 1 // 天枪要求治疗>=2，此处应被实时剔除
	p1.TurnState = model.NewPlayerTurnState()
	if p1.Tokens == nil {
		p1.Tokens = map[string]int{}
	}

	game.State.Phase = model.PhaseResponse
	game.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: "p1",
		SkillIDs: []string{"holy_lancer_sky_spear", "holy_lancer_holy_strike"},
		Context: &model.Context{
			Game:    game,
			User:    p1,
			Trigger: model.TriggerOnAttackStart,
			TriggerCtx: &model.EventContext{
				AttackInfo: &model.AttackEventInfo{
					CounterInitiator: "",
				},
			},
		},
	}

	prompt := game.GetCurrentPrompt()
	if prompt == nil {
		t.Fatalf("expected response prompt after pruning invalid skills")
	}

	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "holy_lancer_holy_strike" {
		t.Fatalf("expected only holy_lancer_holy_strike to remain, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	for _, opt := range prompt.Options {
		if opt.ID == "holy_lancer_sky_spear" {
			t.Fatalf("sky spear should be pruned when heal is insufficient")
		}
	}
}
