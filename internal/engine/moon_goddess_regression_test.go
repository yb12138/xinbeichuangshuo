package engine

import (
	"testing"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

func moonTestCard(id, name string, cardType model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     ele,
		Faction:     "圣",
		Damage:      2,
		Description: name,
	}
}

func TestMoonGoddessNewMoonShelter_AbsorbsOverflowAndPreventsMoraleLoss(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	ally.MaxHand = 4
	ally.Hand = []model.Card{
		moonTestCard("a1", "牌1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("a2", "牌2", model.CardTypeAttack, model.ElementWater),
		moonTestCard("a3", "牌3", model.CardTypeAttack, model.ElementWind),
		moonTestCard("a4", "牌4", model.CardTypeAttack, model.ElementThunder),
		moonTestCard("a5", "牌5", model.CardTypeMagic, model.ElementDark),
		moonTestCard("a6", "牌6", model.CardTypeMagic, model.ElementLight),
	}

	damageOverflowCtx := game.buildContext(ally, nil, model.TriggerNone, nil)
	damageOverflowCtx.Flags["FromDamageDraw"] = true
	damageOverflowCtx.Flags["IsMagicDamage"] = false
	game.checkHandLimit(ally, damageOverflowCtx)

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt from overflow")
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{4, 5},
	})

	if got := game.State.RedMorale; got != 15 {
		t.Fatalf("expected red morale unchanged by 新月庇护, got %d", got)
	}
	if got := moon.Tokens["mg_dark_form"]; got != 1 {
		t.Fatalf("expected moon enter dark form, got %d", got)
	}
	if got := moonGoddessDarkMoonCount(moon); got != 2 {
		t.Fatalf("expected 2 dark moons absorbed, got %d", got)
	}
	if got := len(game.State.DiscardPile); got != 0 {
		t.Fatalf("expected absorbed cards not in discard pile, got %d", got)
	}
}

func TestMoonGoddessNewMoonShelter_NoSoulDevourGainWhenMoraleLossPrevented(t *testing.T) {
	// 重复多次覆盖 map 迭代随机性，确保“暗月抵消士气后，灵魂术士不加黄魂”稳定成立。
	for i := 0; i < 24; i++ {
		game := NewGameEngine(noopObserver{})
		if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p2", "Moon", "moon_goddess", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
			t.Fatal(err)
		}

		soul := game.State.Players["p1"]
		moon := game.State.Players["p2"]
		soul.MaxHand = 4
		soul.Hand = []model.Card{
			moonTestCard("s1", "牌1", model.CardTypeAttack, model.ElementFire),
			moonTestCard("s2", "牌2", model.CardTypeAttack, model.ElementWater),
			moonTestCard("s3", "牌3", model.CardTypeAttack, model.ElementWind),
			moonTestCard("s4", "牌4", model.CardTypeAttack, model.ElementThunder),
			moonTestCard("s5", "牌5", model.CardTypeMagic, model.ElementDark),
			moonTestCard("s6", "牌6", model.CardTypeMagic, model.ElementLight),
		}

		damageOverflowCtx := game.buildContext(soul, nil, model.TriggerNone, nil)
		damageOverflowCtx.Flags["FromDamageDraw"] = true
		damageOverflowCtx.Flags["IsMagicDamage"] = false
		game.checkHandLimit(soul, damageOverflowCtx)

		if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
			t.Fatalf("round %d: expected discard interrupt from overflow", i)
		}
		mustHandleAction(t, game, model.PlayerAction{
			PlayerID:   "p1",
			Type:       model.CmdSelect,
			Selections: []int{4, 5},
		})

		if got := game.State.RedMorale; got != 15 {
			t.Fatalf("round %d: expected red morale unchanged by 新月庇护, got %d", i, got)
		}
		if got := soul.Tokens["ss_yellow_soul"]; got != 0 {
			t.Fatalf("round %d: expected soul devour no yellow gain when morale loss prevented, got %d", i, got)
		}
		if got := moonGoddessDarkMoonCount(moon); got != 2 {
			t.Fatalf("round %d: expected 2 dark moons absorbed, got %d", i, got)
		}
	}
}

func TestMoonGoddessMoonCycle_Branch1AppliesCurseAndHeal(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 0 // 仅保留分支①
	addMoonGoddessDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	ally.Heal = 0
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseTurnEnd

	if !game.maybeTriggerMoonGoddessMoonCycleAtTurnEnd(moon) {
		t.Fatalf("expected moon cycle interrupt")
	}
	requireChoicePrompt(t, game, "p1", "mg_moon_cycle_mode")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose moon cycle mode failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mg_moon_cycle_heal_target")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose moon cycle heal target failed: %v", err)
	}

	if got := moonGoddessDarkMoonCount(moon); got != 0 {
		t.Fatalf("expected dark moon removed by branch1, got %d", got)
	}
	if got := moon.Tokens["mg_dark_form"]; got != 0 {
		t.Fatalf("expected leave dark form when no dark moon, got %d", got)
	}
	if got := game.State.RedMorale; got != 14 {
		t.Fatalf("expected curse morale loss 1, got %d", got)
	}
	if got := ally.Heal; got != 1 {
		t.Fatalf("expected ally heal +1, got %d", got)
	}
}

func TestMoonGoddessMoonCycle_OnlyOncePerTurn(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 1
	addMoonGoddessDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseTurnEnd

	if !game.maybeTriggerMoonGoddessMoonCycleAtTurnEnd(moon) {
		t.Fatalf("expected moon cycle first trigger")
	}
	requireChoicePrompt(t, game, "p1", "mg_moon_cycle_mode")
	// 分支①
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose moon cycle mode branch1 failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mg_moon_cycle_heal_target")
	// 选自己，确保治疗>0，若无一次/回合门闩会继续出现分支②弹窗。
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose moon cycle heal target failed: %v", err)
	}
	if got := moon.Tokens["mg_moon_cycle_used_turn"]; got != 1 {
		t.Fatalf("expected moon cycle used flag=1 in current turn, got %d", got)
	}

	if game.maybeTriggerMoonGoddessMoonCycleAtTurnEnd(moon) {
		t.Fatalf("moon cycle should not trigger twice in same turn")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after second trigger attempt")
	}
}

func TestMoonGoddessMoonCycle_Branch1NoRepromptBranch2InDriveFlow(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 1 // 同时满足分支①/②，验证选①后不会再弹②
	addMoonGoddessDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseTurnEnd

	game.Drive()
	requireChoicePrompt(t, game, "p1", "mg_moon_cycle_mode")

	// 选择分支①
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	requireChoicePrompt(t, game, "p1", "mg_moon_cycle_heal_target")

	// 选择治疗目标并完成分支①
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})

	if intr := game.State.PendingInterrupt; intr != nil && intr.Type == model.InterruptChoice {
		if data, ok := intr.Context.(map[string]interface{}); ok {
			if ct, _ := data["choice_type"].(string); ct == "mg_moon_cycle_mode" && intr.PlayerID == "p1" {
				t.Fatalf("moon cycle should not reprompt branch mode after branch1 resolved")
			}
		}
	}
	for _, intr := range game.State.InterruptQueue {
		if intr == nil || intr.Type != model.InterruptChoice || intr.PlayerID != "p1" {
			continue
		}
		if data, ok := intr.Context.(map[string]interface{}); ok {
			if ct, _ := data["choice_type"].(string); ct == "mg_moon_cycle_mode" {
				t.Fatalf("moon cycle mode should not stay queued after branch1 resolved")
			}
		}
	}
}

func TestMoonGoddessMoonCycle_TurnStateLatchPreventsRepromptWhenTokenResets(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Heal = 1
	addMoonGoddessDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
	})
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseTurnEnd

	if !game.maybeTriggerMoonGoddessMoonCycleAtTurnEnd(moon) {
		t.Fatalf("expected moon cycle first trigger")
	}
	requireChoicePrompt(t, game, "p1", "mg_moon_cycle_mode")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose moon cycle mode branch1 failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mg_moon_cycle_heal_target")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose moon cycle heal target failed: %v", err)
	}

	// 模拟异常链路将 token 意外清零，仍应被本回合 TurnState 门闩拦住。
	moon.Tokens["mg_moon_cycle_used_turn"] = 0
	if game.maybeTriggerMoonGoddessMoonCycleAtTurnEnd(moon) {
		t.Fatalf("moon cycle should stay blocked by turnstate latch even if token resets unexpectedly")
	}
	if game.State.PendingInterrupt != nil {
		if data, ok := game.State.PendingInterrupt.Context.(map[string]interface{}); ok {
			if ct, _ := data["choice_type"].(string); ct == "mg_moon_cycle_mode" {
				t.Fatalf("unexpected moon cycle reprompt after turnstate latch")
			}
		}
	}
}

func TestMoonGoddessDarkMoonSlash_AddsDamageAndConsumesDarkMoon(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	enemy := game.State.Players["p2"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Crystal = 1
	moon.Tokens["mg_dark_form"] = 1
	addMoonGoddessDarkMoonCards(moon, []model.Card{
		moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("dm2", "暗月2", model.CardTypeMagic, model.ElementWater),
	})
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   moon.ID,
			TargetID:   enemy.ID,
			Damage:     2,
			DamageType: "Attack",
			Stage:      1,
		},
	}

	ctx := game.buildContext(moon, enemy, model.TriggerOnAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: moon.ID,
		TargetID: enemy.ID,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	})
	h := &skills.MoonGoddessDarkMoonSlashHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected dark moon slash can use")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute dark moon slash failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mg_darkmoon_slash_x")
	if err := game.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose X=2 failed: %v", err)
	}

	if got := game.State.PendingDamageQueue[0].Damage; got != 4 {
		t.Fatalf("expected attack damage +2 (2->4), got %d", got)
	}
	if got := moonGoddessDarkMoonCount(moon); got != 0 {
		t.Fatalf("expected all dark moon removed, got %d", got)
	}
	if got := game.State.RedMorale; got != 13 {
		t.Fatalf("expected curse morale loss 2, got %d", got)
	}
	if got := moon.Crystal; got != 0 {
		t.Fatalf("expected consume 1 crystal, got %d", got)
	}
}

func TestMoonGoddessMedusa_ExcludesConvertedAttacks(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	enemy := game.State.Players["p3"]
	addMoonGoddessDarkMoonCards(moon, []model.Card{
		moonTestCard("dm_fire", "火暗月", model.CardTypeAttack, model.ElementFire),
	})
	attackCard := moonTestCard("atk", "火斩", model.CardTypeAttack, model.ElementFire)
	attackStartCtx := game.buildContext(enemy, ally, model.TriggerOnAttackStart, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: enemy.ID,
		TargetID: ally.ID,
		Card:     &attackCard,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
		},
	})

	if game.maybeTriggerMoonGoddessMedusa(enemy, ally, "adventurer_fraud", &attackCard, attackStartCtx) {
		t.Fatalf("fraud converted attack should not trigger medusa")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt for fraud converted attack")
	}
	if game.maybeTriggerMoonGoddessMedusa(enemy, ally, "hb_holy_shard_storm", &attackCard, attackStartCtx) {
		t.Fatalf("holy shard storm converted attack should not trigger medusa")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt for holy shard storm converted attack")
	}

	if !game.maybeTriggerMoonGoddessMedusa(enemy, ally, "", &attackCard, attackStartCtx) {
		t.Fatalf("normal attack should trigger medusa when matching dark moon exists")
	}
	requireChoicePrompt(t, game, "p1", "mg_medusa_darkmoon_pick")
}

func TestMoonGoddessMedusa_OnlyAtAttackStart(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	enemy := game.State.Players["p3"]
	addMoonGoddessDarkMoonCards(moon, []model.Card{
		moonTestCard("dm_fire", "火暗月", model.CardTypeAttack, model.ElementFire),
	})
	attackCard := moonTestCard("atk", "火斩", model.CardTypeAttack, model.ElementFire)

	// 非攻击开始上下文：不应触发。
	if game.maybeTriggerMoonGoddessMedusa(enemy, ally, "", &attackCard, nil) {
		t.Fatalf("medusa should not trigger without attack-start context")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt without attack-start context")
	}

	nonStartCtx := game.buildContext(enemy, ally, model.TriggerOnAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: enemy.ID,
		TargetID: ally.ID,
		Card:     &attackCard,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
		},
	})
	if game.maybeTriggerMoonGoddessMedusa(enemy, ally, "", &attackCard, nonStartCtx) {
		t.Fatalf("medusa should not trigger outside attack-start trigger")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt for non-attack-start trigger")
	}

	// 攻击开始上下文：可触发。
	attackStartCtx := game.buildContext(enemy, ally, model.TriggerOnAttackStart, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: enemy.ID,
		TargetID: ally.ID,
		Card:     &attackCard,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
		},
	})
	if !game.maybeTriggerMoonGoddessMedusa(enemy, ally, "", &attackCard, attackStartCtx) {
		t.Fatalf("medusa should trigger at attack start with matching dark moon")
	}
	requireChoicePrompt(t, game, "p1", "mg_medusa_darkmoon_pick")
}

func TestMoonGoddessBlasphemy_OncePerTurnAndResetNextTurn(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.Heal = 2
	pd := model.PendingDamage{
		SourceID:   "p1",
		TargetID:   "p2",
		Damage:     1,
		DamageType: "magic",
	}

	if !game.tryQueueMoonGoddessBlasphemy(&pd) {
		t.Fatalf("expected first blasphemy queue success")
	}
	requireChoicePrompt(t, game, "p1", "mg_blasphemy_target")
	// 选第1个目标（index=0 为“跳过”）。
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("resolve blasphemy target failed: %v", err)
	}
	if got := moon.Tokens["mg_blasphemy_used_turn"]; got != 1 {
		t.Fatalf("expected blasphemy used flag=1, got %d", got)
	}
	if got := moon.Tokens["mg_blasphemy_pending"]; got != 0 {
		t.Fatalf("expected blasphemy pending reset to 0, got %d", got)
	}
	if game.tryQueueMoonGoddessBlasphemy(&pd) {
		t.Fatalf("blasphemy should be blocked after used once in same turn")
	}

	moon.IsActive = true
	game.State.CurrentTurn = 0
	game.NextTurn()
	if got := moon.Tokens["mg_blasphemy_used_turn"]; got != 0 {
		t.Fatalf("expected blasphemy used flag reset on next turn, got %d", got)
	}
	if !game.tryQueueMoonGoddessBlasphemy(&pd) {
		t.Fatalf("expected blasphemy can queue again after turn reset")
	}
	requireChoicePrompt(t, game, "p1", "mg_blasphemy_target")
}

func TestMoonGoddessPaleMoon_Branch1GrantsExtraTurn(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	moon.IsActive = true
	moon.TurnState = model.NewPlayerTurnState()
	moon.Gem = 1
	moon.Tokens["mg_petrify"] = 3
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "mg_pale_moon",
	})
	requireChoicePrompt(t, game, "p1", "mg_pale_moon_mode")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose pale moon branch1 failed: %v", err)
	}
	if got := moon.Tokens["mg_next_attack_no_counter"]; got != 1 {
		t.Fatalf("expected next attack no-counter token=1, got %d", got)
	}
	if got := moon.Tokens["mg_extra_turn_pending"]; got != 1 {
		t.Fatalf("expected extra-turn pending=1, got %d", got)
	}
	if len(moon.TurnState.PendingActions) == 0 || moon.TurnState.PendingActions[0].MustType != "Attack" {
		t.Fatalf("expected pale moon branch1 to queue one extra Attack action")
	}

	game.NextTurn()
	if got := game.State.CurrentTurn; got != 0 {
		t.Fatalf("expected extra turn keeps current turn index at 0, got %d", got)
	}
	if got := moon.Tokens["mg_extra_turn_pending"]; got != 0 {
		t.Fatalf("expected extra-turn pending consumed after NextTurn, got %d", got)
	}
	if !moon.IsActive {
		t.Fatalf("expected moon still active in extra turn")
	}
}

func TestMoonGoddessNewMoonShelter_NotTriggerWhenActualMoraleWillNotDrop(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	moon := game.State.Players["p1"]
	ally := game.State.Players["p2"]

	// 红莲骑士热血形态：伤害导致爆牌不掉士气，因此新月庇护不应触发。
	ally.Tokens["crk_hot_form"] = 1
	ally.MaxHand = 4
	ally.Hand = []model.Card{
		moonTestCard("h1", "牌1", model.CardTypeAttack, model.ElementFire),
		moonTestCard("h2", "牌2", model.CardTypeAttack, model.ElementWater),
		moonTestCard("h3", "牌3", model.CardTypeAttack, model.ElementWind),
		moonTestCard("h4", "牌4", model.CardTypeAttack, model.ElementThunder),
		moonTestCard("h5", "牌5", model.CardTypeMagic, model.ElementDark),
		moonTestCard("h6", "牌6", model.CardTypeMagic, model.ElementLight),
	}

	damageOverflowCtx := game.buildContext(ally, nil, model.TriggerNone, nil)
	damageOverflowCtx.Flags["FromDamageDraw"] = true
	damageOverflowCtx.Flags["IsMagicDamage"] = false
	game.checkHandLimit(ally, damageOverflowCtx)

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt from overflow")
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{4, 5},
	})

	if got := game.State.RedMorale; got != 15 {
		t.Fatalf("expected red morale unchanged, got %d", got)
	}
	if got := moonGoddessDarkMoonCount(moon); got != 0 {
		t.Fatalf("expected new moon shelter not trigger, dark moon count=%d", got)
	}
	if got := moon.Tokens["mg_dark_form"]; got != 0 {
		t.Fatalf("expected moon goddess stay non-dark form, got %d", got)
	}
	if got := len(game.State.DiscardPile); got != 2 {
		t.Fatalf("expected overflow cards enter discard pile, got %d", got)
	}
}

func TestMoonGoddessDarkMoonSlash_XBoundaries_CurseAndDamage(t *testing.T) {
	cases := []struct {
		name              string
		x                 int
		wantDamage        int
		wantRedMorale     int
		wantDarkMoonCount int
	}{
		{
			name:              "x0",
			x:                 0,
			wantDamage:        2,
			wantRedMorale:     15,
			wantDarkMoonCount: 2,
		},
		{
			name:              "x1",
			x:                 1,
			wantDamage:        3,
			wantRedMorale:     14,
			wantDarkMoonCount: 1,
		},
		{
			name:              "x2",
			x:                 2,
			wantDamage:        4,
			wantRedMorale:     13,
			wantDarkMoonCount: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			game := NewGameEngine(noopObserver{})
			if err := game.AddPlayer("p1", "Moon", "moon_goddess", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
				t.Fatal(err)
			}

			moon := game.State.Players["p1"]
			enemy := game.State.Players["p2"]
			moon.IsActive = true
			moon.TurnState = model.NewPlayerTurnState()
			moon.Crystal = 1
			moon.Tokens["mg_dark_form"] = 1
			addMoonGoddessDarkMoonCards(moon, []model.Card{
				moonTestCard("dm1", "暗月1", model.CardTypeAttack, model.ElementFire),
				moonTestCard("dm2", "暗月2", model.CardTypeMagic, model.ElementWater),
			})
			game.State.PendingDamageQueue = []model.PendingDamage{
				{
					SourceID:   moon.ID,
					TargetID:   enemy.ID,
					Damage:     2,
					DamageType: "Attack",
					Stage:      1,
				},
			}

			ctx := game.buildContext(moon, enemy, model.TriggerOnAttackHit, &model.EventContext{
				Type:     model.EventAttack,
				SourceID: moon.ID,
				TargetID: enemy.ID,
				AttackInfo: &model.AttackEventInfo{
					ActionType:       string(model.ActionAttack),
					CounterInitiator: "",
				},
			})

			h := &skills.MoonGoddessDarkMoonSlashHandler{}
			if !h.CanUse(ctx) {
				t.Fatalf("expected dark moon slash can use")
			}
			if err := h.Execute(ctx); err != nil {
				t.Fatalf("execute dark moon slash failed: %v", err)
			}
			requireChoicePrompt(t, game, "p1", "mg_darkmoon_slash_x")
			if err := game.handleWeakChoiceInput("p1", tc.x); err != nil {
				t.Fatalf("choose x=%d failed: %v", tc.x, err)
			}

			if got := game.State.PendingDamageQueue[0].Damage; got != tc.wantDamage {
				t.Fatalf("expected damage=%d, got %d", tc.wantDamage, got)
			}
			if got := game.State.RedMorale; got != tc.wantRedMorale {
				t.Fatalf("expected red morale=%d, got %d", tc.wantRedMorale, got)
			}
			if got := moonGoddessDarkMoonCount(moon); got != tc.wantDarkMoonCount {
				t.Fatalf("expected dark moon count=%d, got %d", tc.wantDarkMoonCount, got)
			}
			if got := moon.Crystal; got != 0 {
				t.Fatalf("expected consume 1 crystal in all branches, got %d", got)
			}
		})
	}
}
