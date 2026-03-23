package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func holyBowTestCard(id, name string, cardType model.CardType, element model.Element, damage int) model.Card {
	if damage <= 0 && cardType == model.CardTypeAttack {
		damage = 2
	}
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     element,
		Damage:      damage,
		Description: name,
	}
}

func TestHolyBow_InitStatsAndTokens(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	if got := p1.Crystal; got != 2 {
		t.Fatalf("expected initial crystal=2, got %d", got)
	}
	if got := p1.MaxHeal; got != 3 {
		t.Fatalf("expected max heal=3, got %d", got)
	}
	if got := p1.Tokens["hb_cannon"]; got != 1 {
		t.Fatalf("expected hb_cannon=1, got %d", got)
	}
	if got := p1.Tokens["hb_faith"]; got != 0 {
		t.Fatalf("expected hb_faith=0, got %d", got)
	}
	if got := p1.Tokens["hb_form"]; got != 0 {
		t.Fatalf("expected hb_form=0, got %d", got)
	}
}

func TestHolyBow_HeavenlyBowDamageAdjustments(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]

	nonHoly := holyBowTestCard("atk1", "火斩", model.CardTypeAttack, model.ElementFire, 2)
	nonHoly.Faction = "幻"
	if got := game.applyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID: p1.ID,
		TargetID: p2.ID,
		Type:     model.ActionAttack,
		Card:     &nonHoly,
	}); got != 1 {
		t.Fatalf("expected non-holy active attack damage=1, got %d", got)
	}

	holy := holyBowTestCard("atk2", "圣斩", model.CardTypeAttack, model.ElementLight, 2)
	holy.Faction = "圣"
	if got := game.applyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID: p1.ID,
		TargetID: p2.ID,
		Type:     model.ActionAttack,
		Card:     &holy,
	}); got != 2 {
		t.Fatalf("expected holy active attack damage keep 2, got %d", got)
	}
}

func TestHolyBow_HeavenlyBowHolyHitGainFaith(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	holy := holyBowTestCard("atk1", "圣斩", model.CardTypeAttack, model.ElementLight, 2)
	holy.Faction = "圣"
	game.handlePostAttackHitEffects(&model.PendingDamage{
		SourceID:   "p1",
		TargetID:   "p2",
		Damage:     2,
		DamageType: "Attack",
		IsCounter:  false,
		Card:       &holy,
	})
	if got := p1.Tokens["hb_faith"]; got != 1 {
		t.Fatalf("expected faith+1 on holy active hit, got %d", got)
	}

	nonHoly := holyBowTestCard("atk2", "火斩", model.CardTypeAttack, model.ElementFire, 2)
	nonHoly.Faction = "幻"
	game.handlePostAttackHitEffects(&model.PendingDamage{
		SourceID:   "p1",
		TargetID:   "p2",
		Damage:     2,
		DamageType: "Attack",
		IsCounter:  false,
		Card:       &nonHoly,
	})
	if got := p1.Tokens["hb_faith"]; got != 1 {
		t.Fatalf("expected no extra faith on non-holy hit, got %d", got)
	}

	game.handlePostAttackHitEffects(&model.PendingDamage{
		SourceID:   "p1",
		TargetID:   "p2",
		Damage:     2,
		DamageType: "Attack",
		IsCounter:  true,
		Card:       &holy,
	})
	if got := p1.Tokens["hb_faith"]; got != 1 {
		t.Fatalf("expected no extra faith on counter hit, got %d", got)
	}
}

func TestHolyBow_RadiantDescentAndSpecialExitForm(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 3
	p1.Tokens["hb_faith"] = 0
	p1.Hand = []model.Card{
		holyBowTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "hb_radiant_descent",
	})
	requireChoicePrompt(t, game, "p1", "hb_radiant_descent_cost")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 移除2点治疗
	})

	if got := p1.Tokens["hb_form"]; got != 1 {
		t.Fatalf("expected enter hb_form=1 after radiant descent, got %d", got)
	}
	if got := p1.Heal; got != 1 {
		t.Fatalf("expected heal reduced to 1, got %d", got)
	}
	if p1.TurnState.CurrentExtraAction != "Magic" && len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected extra magic action from radiant descent, current=%s pending=%d", p1.TurnState.CurrentExtraAction, len(p1.TurnState.PendingActions))
	}

	// 清理额外行动约束后执行特殊行动，验证圣煌形态会脱离并+1治疗。
	p1.TurnState.CurrentExtraAction = ""
	p1.TurnState.CurrentExtraElement = nil
	p1.TurnState.PendingActions = nil
	game.State.Phase = model.PhaseActionSelection
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdBuy,
	})
	if got := p1.Tokens["hb_form"]; got != 0 {
		t.Fatalf("expected holy bow form cleared after special action, got %d", got)
	}
	if got := p1.Heal; got != 2 {
		t.Fatalf("expected +1 heal after exiting form by special action, got %d", got)
	}
}

func TestHolyBow_AutoFillTriggeredAtTurnEndWithoutSpecial(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hb_special_used_turn"] = 0
	p1.Tokens["hb_auto_fill_done_turn"] = 0
	p1.Crystal = 1
	p1.Gem = 0
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseTurnEnd

	game.Drive()
	requireChoicePrompt(t, game, "p1", "hb_auto_fill_resource")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // crystal 分支
	})
	requireChoicePrompt(t, game, "p1", "hb_auto_fill_gain")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // +1 faith
	})

	if got := p1.Tokens["hb_faith"]; got != 1 {
		t.Fatalf("expected auto-fill to add 1 faith, got %d", got)
	}
	if got := p1.Crystal; got != 0 {
		t.Fatalf("expected crystal consumed by auto-fill branch 1, got %d", got)
	}
}

func TestHolyBow_HolyShardStormMiss_NoBranch(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 3
	p1.Hand = []model.Card{
		holyBowTestCard("hb_a1", "火斩1", model.CardTypeAttack, model.ElementFire, 2),
		holyBowTestCard("hb_a2", "火斩2", model.CardTypeAttack, model.ElementFire, 2),
	}
	p2.Hand = []model.Card{
		holyBowTestCard("e_m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	p3.Hand = []model.Card{
		holyBowTestCard("al_m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
		holyBowTestCard("al_m2", "魔弹", model.CardTypeMagic, model.ElementDark, 0),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "hb_holy_shard_storm",
	})
	requireChoicePrompt(t, game, "p1", "hb_holy_shard_combo")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	requireChoicePrompt(t, game, "p1", "hb_holy_shard_target")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	})

	requireChoicePrompt(t, game, "p1", "hb_holy_shard_miss_confirm")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1}, // 否
	})

	if got := p1.Heal; got != 3 {
		t.Fatalf("expected heal unchanged on miss-no branch, got %d", got)
	}
	if got := len(p3.Hand); got != 2 {
		t.Fatalf("expected ally hand unchanged on miss-no branch, got %d", got)
	}
	if got := p1.Tokens["hb_shard_miss_pending"]; got != 0 {
		t.Fatalf("expected shard_miss_pending cleared, got %d", got)
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptDiscard {
		t.Fatalf("did not expect discard interrupt on miss-no branch")
	}
}

func TestHolyBow_HolyShardStormMiss_YesBranch(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 3
	p1.Hand = []model.Card{
		holyBowTestCard("hb_a1", "火斩1", model.CardTypeAttack, model.ElementFire, 2),
		holyBowTestCard("hb_a2", "火斩2", model.CardTypeAttack, model.ElementFire, 2),
	}
	p2.Hand = []model.Card{
		holyBowTestCard("e_m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	p3.Hand = []model.Card{
		holyBowTestCard("al_m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
		holyBowTestCard("al_m2", "魔弹", model.CardTypeMagic, model.ElementDark, 0),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "hb_holy_shard_storm",
	})
	requireChoicePrompt(t, game, "p1", "hb_holy_shard_combo")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	requireChoicePrompt(t, game, "p1", "hb_holy_shard_target")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	})

	requireChoicePrompt(t, game, "p1", "hb_holy_shard_miss_confirm")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 是
	})
	requireChoicePrompt(t, game, "p1", "hb_holy_shard_miss_x")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1}, // X=2（上限边界）
	})
	requireChoicePrompt(t, game, "p1", "hb_holy_shard_miss_ally_target")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 指定 ally
	})

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard || game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected ally discard interrupt for p3, got %+v", game.State.PendingInterrupt)
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p3",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	})

	if got := p1.Heal; got != 1 {
		t.Fatalf("expected heal reduced by X=2, got %d", got)
	}
	if got := len(p3.Hand); got != 0 {
		t.Fatalf("expected ally discarded 2 cards, got hand=%d", got)
	}
	if got := p1.Tokens["hb_shard_miss_pending"]; got != 0 {
		t.Fatalf("expected shard_miss_pending cleared, got %d", got)
	}
}

func TestHolyBow_LightBurstModeB_XYBoundaries(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hb_form"] = 1
	p1.Heal = 2
	p1.Hand = []model.Card{
		holyBowTestCard("lb_c1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
		holyBowTestCard("lb_c2", "魔弹", model.CardTypeMagic, model.ElementDark, 0),
	}
	p2.Hand = nil
	p3.Hand = nil
	p2.Heal = 1
	p3.Heal = 0
	game.State.Deck = []model.Card{
		holyBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		holyBowTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		holyBowTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder, 2),
		holyBowTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementWind, 2),
		holyBowTestCard("d5", "补牌5", model.CardTypeAttack, model.ElementEarth, 2),
		holyBowTestCard("d6", "补牌6", model.CardTypeAttack, model.ElementLight, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "hb_light_burst",
	})
	requireChoicePrompt(t, game, "p1", "hb_light_burst_mode")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1}, // 分支②
	})
	requireChoicePrompt(t, game, "p1", "hb_light_burst_mode_b_x")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1}, // X=2（最大）
	})
	requireChoicePrompt(t, game, "p1", "hb_light_burst_mode_b_targets")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 先选第1名目标
	})
	requireChoicePrompt(t, game, "p1", "hb_light_burst_mode_b_targets")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1}, // 点击“完成目标选择”（至多X名）
	})
	requireChoicePrompt(t, game, "p1", "hb_light_burst_mode_b_discard")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	requireChoicePrompt(t, game, "p1", "hb_light_burst_mode_b_discard")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// 处理伤害结算过程中的“治疗抵消”选择（目标顺序取决于前面的选目标顺序）。
	for game.State.PendingInterrupt != nil {
		if game.State.PendingInterrupt.Type != model.InterruptChoice {
			t.Fatalf("unexpected pending interrupt during damage resolution: %+v", game.State.PendingInterrupt)
		}
		data, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
		choiceType, _ := data["choice_type"].(string)
		if choiceType == "hb_auto_fill_resource" {
			// 说明圣光爆裂分支②的伤害已完成结算，进入了回合结束时的自动填充询问。
			break
		}
		if choiceType != "heal" {
			t.Fatalf("unexpected pending choice during damage resolution: %s", choiceType)
		}
		maxHeal := 0
		if v, ok := data["max_heal"].(int); ok {
			maxHeal = v
		} else if f, ok := data["max_heal"].(float64); ok {
			maxHeal = int(f)
		}
		useHeal := maxHeal
		if useHeal < 0 {
			useHeal = 0
		}
		mustHandleAction(t, game, model.PlayerAction{
			PlayerID:   game.State.PendingInterrupt.PlayerID,
			Type:       model.CmdSelect,
			Selections: []int{useHeal},
		})
	}

	if got := p1.Heal; got != 0 {
		t.Fatalf("expected healer cost X=2 consumed, got heal=%d", got)
	}
	if got := len(p1.Hand); got != 0 {
		t.Fatalf("expected discard X=2 from hand, got hand=%d", got)
	}
	// 本用例在 X=2 时提前完成，仅选择了 1 名目标（p2）。
	// Y=1（已选目标中仅 p2 有治疗），故目标伤害为 Y+2=3。
	// p2 先抵消1点治疗后摸2张；p3未被指定，不应受到伤害。
	if got := len(p2.Hand); got != 2 {
		t.Fatalf("expected enemyA hand draw=2 with heal mitigation, got %d", got)
	}
	if got := len(p3.Hand); got != 0 {
		t.Fatalf("expected enemyB not targeted and hand unchanged, got %d", got)
	}
}

// 回归：圣光爆裂若两个分支都不满足发动条件，不应进入空选项弹窗。
func TestHolyBow_LightBurst_NoAvailableModeCannotUse(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hb_form"] = 1
	p1.Heal = 0
	p1.Hand = []model.Card{
		holyBowTestCard("hb_lb_block", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	// 分支②要求 X>=1（需治疗可移除），此处故意让 maxX=0；分支①同样要求至少1点治疗。
	p2.Hand = nil

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "hb_light_burst",
	})
	if err == nil {
		t.Fatalf("expected hb_light_burst to be unusable when no mode is available")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt, got %+v", game.State.PendingInterrupt)
	}
}

func TestHolyBow_RadiantCannon_MoraleAlignBothSides(t *testing.T) {
	type tc struct {
		name       string
		sideSelect int
		wantRed    int
		wantBlue   int
	}
	cases := []tc{
		{name: "align_red_to_blue", sideSelect: 0, wantRed: 6, wantBlue: 6},
		{name: "align_blue_to_red", sideSelect: 1, wantRed: 9, wantBlue: 9},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			game := NewGameEngine(noopObserver{})
			if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
				t.Fatal(err)
			}

			p1 := game.State.Players["p1"]
			p2 := game.State.Players["p2"]
			p3 := game.State.Players["p3"]
			p1.IsActive = true
			p1.TurnState = model.NewPlayerTurnState()
			p1.Tokens["hb_form"] = 1
			p1.Tokens["hb_cannon"] = 1
			p1.Tokens["hb_faith"] = 6
			p1.Hand = []model.Card{
				holyBowTestCard("c1", "卡1", model.CardTypeAttack, model.ElementFire, 2),
				holyBowTestCard("c2", "卡2", model.CardTypeAttack, model.ElementWater, 2),
				holyBowTestCard("c3", "卡3", model.CardTypeAttack, model.ElementThunder, 2),
				holyBowTestCard("c4", "卡4", model.CardTypeMagic, model.ElementLight, 0),
				holyBowTestCard("c5", "卡5", model.CardTypeMagic, model.ElementDark, 0),
			}
			p2.Hand = []model.Card{
				holyBowTestCard("e1", "敌卡1", model.CardTypeAttack, model.ElementFire, 2),
				holyBowTestCard("e2", "敌卡2", model.CardTypeMagic, model.ElementLight, 0),
			}
			p3.Hand = []model.Card{
				holyBowTestCard("a1", "队卡1", model.CardTypeAttack, model.ElementFire, 2),
				holyBowTestCard("a2", "队卡2", model.CardTypeAttack, model.ElementWater, 2),
				holyBowTestCard("a3", "队卡3", model.CardTypeAttack, model.ElementEarth, 2),
				holyBowTestCard("a4", "队卡4", model.CardTypeMagic, model.ElementDark, 0),
			}
			game.State.Deck = []model.Card{
				holyBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
				holyBowTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
				holyBowTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder, 2),
				holyBowTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementWind, 2),
			}
			game.State.RedMorale = 9
			game.State.BlueMorale = 6
			game.State.RedCups = 0
			game.State.BlueCups = 0
			game.State.CurrentTurn = 0
			game.State.Phase = model.PhaseActionSelection

			mustHandleAction(t, game, model.PlayerAction{
				PlayerID: "p1",
				Type:     model.CmdSkill,
				SkillID:  "hb_radiant_cannon",
			})
			requireChoicePrompt(t, game, "p1", "hb_radiant_cannon_side")
			mustHandleAction(t, game, model.PlayerAction{
				PlayerID:   "p1",
				Type:       model.CmdSelect,
				Selections: []int{c.sideSelect},
			})

			if got := p1.Tokens["hb_cannon"]; got != 0 {
				t.Fatalf("expected cannon consumed to 0, got %d", got)
			}
			if got := p1.Tokens["hb_faith"]; got != 2 {
				t.Fatalf("expected faith cost 4, got %d", got)
			}
			if got := game.State.RedCups; got != 1 {
				t.Fatalf("expected red camp cups +1, got %d", got)
			}
			if got := game.State.RedMorale; got != c.wantRed {
				t.Fatalf("unexpected red morale, got %d want %d", got, c.wantRed)
			}
			if got := game.State.BlueMorale; got != c.wantBlue {
				t.Fatalf("unexpected blue morale, got %d want %d", got, c.wantBlue)
			}
			if got := len(p1.Hand); got != 4 {
				t.Fatalf("expected p1 hand adjusted to 4, got %d", got)
			}
			if got := len(p2.Hand); got != 4 {
				t.Fatalf("expected p2 hand adjusted to 4, got %d", got)
			}
			if got := len(p3.Hand); got != 4 {
				t.Fatalf("expected p3 hand adjusted to 4, got %d", got)
			}
		})
	}
}
