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

	if game.maybeTriggerMoonGoddessMedusa(enemy, ally, "adventurer_fraud", &attackCard, nil) {
		t.Fatalf("fraud converted attack should not trigger medusa")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt for fraud converted attack")
	}
	if game.maybeTriggerMoonGoddessMedusa(enemy, ally, "hb_holy_shard_storm", &attackCard, nil) {
		t.Fatalf("holy shard storm converted attack should not trigger medusa")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no interrupt for holy shard storm converted attack")
	}

	if !game.maybeTriggerMoonGoddessMedusa(enemy, ally, "", &attackCard, nil) {
		t.Fatalf("normal attack should trigger medusa when matching dark moon exists")
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
