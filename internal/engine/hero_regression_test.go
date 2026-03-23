package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func heroTestCard(id, name string, cardType model.CardType, element model.Element, damage int) model.Card {
	if damage <= 0 {
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

func heroTauntExclusiveCard(owner *model.Player) model.Card {
	charName := "勇者"
	faction := "血"
	if owner != nil && owner.Character != nil {
		charName = owner.Character.Name
		faction = owner.Character.Faction
	}
	return model.Card{
		ID:              "starter-hero-taunt-test",
		Name:            "挑衅",
		Type:            model.CardTypeMagic,
		Element:         model.ElementFire,
		Faction:         faction,
		Damage:          0,
		Description:     "勇者专属测试卡",
		ExclusiveChar1:  charName,
		ExclusiveSkill1: "挑衅",
	}
}

func TestHeroHeart_InitialCrystalPlusTwo(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	if p1 == nil {
		t.Fatal("hero player not found")
	}
	if got := p1.Crystal; got != 2 {
		t.Fatalf("expected hero initial crystal=2, got %d", got)
	}
}

func TestHeroRoar_HitDamagePlusTwoAndCleared(t *testing.T) {
	game := NewGameEngine(noopObserver{})
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
	p1.Crystal = 0
	p1.Gem = 0
	p1.Hand = []model.Card{
		heroTestCard("a1", "火焰斩", model.CardTypeAttack, model.ElementFire, 2),
	}
	p2.Hand = nil
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d4", "抽4", model.CardTypeAttack, model.ElementWater, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "hero_roar")
	requireChoicePrompt(t, game, "p1", "hero_roar_draw")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if got := len(p2.Hand); got != 4 {
		t.Fatalf("expected target draw 4 cards from roar-boosted attack, got %d", got)
	}
	if got := p1.Tokens["hero_roar_active"]; got != 0 {
		t.Fatalf("expected hero_roar_active cleared after hit, got %d", got)
	}
	if got := p1.Tokens["hero_roar_damage_pending"]; got != 0 {
		t.Fatalf("expected hero_roar_damage_pending cleared after hit, got %d", got)
	}
	if got := p1.Tokens["hero_wisdom"]; got != 0 {
		t.Fatalf("expected wisdom unchanged on hit branch, got %d", got)
	}
}

func TestHeroRoar_MissAddsWisdom(t *testing.T) {
	game := NewGameEngine(noopObserver{})
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
	p1.Crystal = 0
	p1.Gem = 0
	p1.Hand = []model.Card{
		heroTestCard("a1", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
	}
	p2.Hand = []model.Card{
		heroTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "hero_roar")
	requireChoicePrompt(t, game, "p1", "hero_roar_draw")
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

	if got := p1.Tokens["hero_wisdom"]; got != 1 {
		t.Fatalf("expected roar miss grant wisdom=1, got %d", got)
	}
	if got := p1.Tokens["hero_roar_active"]; got != 0 {
		t.Fatalf("expected roar active cleared after miss, got %d", got)
	}
	if got := p1.Tokens["hero_roar_damage_pending"]; got != 0 {
		t.Fatalf("expected roar damage marker cleared after miss, got %d", got)
	}
}

func TestHeroForbiddenPower_HitBranch(t *testing.T) {
	game := NewGameEngine(noopObserver{})
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
	p1.Hand = []model.Card{
		heroTestCard("atk", "起手攻击", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("mf", "火法术", model.CardTypeMagic, model.ElementFire, 0),
		heroTestCard("mw", "水法术", model.CardTypeMagic, model.ElementWater, 0),
		heroTestCard("af", "火攻击", model.CardTypeAttack, model.ElementFire, 2),
	}
	p2.Hand = nil
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d4", "抽4", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d5", "抽5", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d6", "抽6", model.CardTypeAttack, model.ElementEarth, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	chooseResponseSkillByID(t, game, "p1", "hero_forbidden_power")

	if got := p1.Tokens["hero_anger"]; got != 2 {
		t.Fatalf("expected anger +2 from discarded magic cards, got %d", got)
	}
	if got := p1.Crystal; got != 1 {
		t.Fatalf("expected forbidden power consume 1 crystal-like, crystal=%d", got)
	}
	if got := p1.Tokens["hero_exhaustion_form"]; got != 1 {
		t.Fatalf("expected exhaustion form active, got %d", got)
	}
	if got := p1.Tokens["hero_exhaustion_release_pending"]; got != 1 {
		t.Fatalf("expected exhaustion pending flag=1, got %d", got)
	}
	if got := len(p2.Hand); got != 4 {
		t.Fatalf("expected hit branch add fire-count bonus to attack damage (target draw 4), got %d", got)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" && len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected exhaustion grant extra attack action")
	}
}

func TestHeroForbiddenPower_MissBranchWaterToWisdom(t *testing.T) {
	game := NewGameEngine(noopObserver{})
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
	p1.Hand = []model.Card{
		heroTestCard("atk", "起手攻击", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("mw", "水法术", model.CardTypeMagic, model.ElementWater, 0),
		heroTestCard("ml", "光法术", model.CardTypeMagic, model.ElementLight, 0),
		heroTestCard("aw", "水攻击", model.CardTypeAttack, model.ElementWater, 2),
	}
	p2.Hand = []model.Card{
		heroTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	})
	chooseResponseSkillByID(t, game, "p1", "hero_forbidden_power")

	if got := p1.Tokens["hero_anger"]; got != 2 {
		t.Fatalf("expected anger +2 from discarded magic cards on miss branch, got %d", got)
	}
	if got := p1.Tokens["hero_wisdom"]; got != 2 {
		t.Fatalf("expected wisdom +2 from discarded water cards on miss branch, got %d", got)
	}
	if got := p1.Tokens["hero_exhaustion_form"]; got != 1 {
		t.Fatalf("expected exhaustion form active after forbidden power miss branch, got %d", got)
	}
}

// 用户场景1：
// 弃牌为 [水系攻击, 水系法术, 地系攻击]，攻击未命中时应 +1怒气、+2知性。
func TestHeroForbiddenPower_UserScenario_Miss_WaterAttackAndMagicToWisdom(t *testing.T) {
	game := NewGameEngine(noopObserver{})
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
	p1.Tokens["hero_anger"] = 0
	p1.Tokens["hero_wisdom"] = 0
	// 第1张用于本次主动攻击，触发禁断之力后应展示并弃掉剩余3张：
	// 水攻 + 水法 + 地攻
	p1.Hand = []model.Card{
		heroTestCard("atk", "起手攻击", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("aw", "水攻击", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("mw", "水法术", model.CardTypeMagic, model.ElementWater, 0),
		heroTestCard("ae", "地攻击", model.CardTypeAttack, model.ElementEarth, 2),
	}
	p2.Hand = []model.Card{
		heroTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	// 用圣光防御，制造“未命中”分支
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	})
	chooseResponseSkillByID(t, game, "p1", "hero_forbidden_power")

	if got := p1.Tokens["hero_anger"]; got != 1 {
		t.Fatalf("expected anger +1 from one discarded magic card, got %d", got)
	}
	if got := p1.Tokens["hero_wisdom"]; got != 2 {
		t.Fatalf("expected wisdom +2 from two discarded water cards (water attack + water magic), got %d", got)
	}
}

// 用户场景2：
// 弃牌为 [水系法术, 火系法术, 火系攻击]，攻击命中时应：
// 本次伤害+2，自身承受2点法术伤害；法术牌2张带来怒气+2（受上限影响时实际增加可能小于2）。
func TestHeroForbiddenPower_UserScenario_Hit_FireCardsBonusAndSelfDamage(t *testing.T) {
	game := NewGameEngine(noopObserver{})
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
	// 将怒气预置为3，便于验证“2张法术应加2，但受上限4约束，实际只+1”。
	p1.Tokens["hero_anger"] = 3
	p1.Tokens["hero_wisdom"] = 0
	// 第1张用于攻击；禁断之力弃剩余3张：水法 + 火法 + 火攻（火牌共2）
	p1.Hand = []model.Card{
		heroTestCard("atk", "起手攻击", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("mw", "水法术", model.CardTypeMagic, model.ElementWater, 0),
		heroTestCard("mf", "火法术", model.CardTypeMagic, model.ElementFire, 0),
		heroTestCard("af", "火攻击", model.CardTypeAttack, model.ElementFire, 2),
	}
	p2.Hand = nil
	// p2命中摸牌4张 + p1自伤摸牌2张
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d4", "抽4", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d5", "抽5", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d6", "抽6", model.CardTypeAttack, model.ElementEarth, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	// 攻击前会先进入勇者响应技能（如怒吼）窗口；本用例聚焦禁断之力命中分支，这里显式跳过前置响应。
	requireResponseSkillPrompt(t, game, "p1")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{len(game.State.PendingInterrupt.SkillIDs)},
	})
	// 目标承受，制造“命中”分支
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	chooseResponseSkillByID(t, game, "p1", "hero_forbidden_power")

	if got := p1.Tokens["hero_anger"]; got != 4 {
		t.Fatalf("expected anger reach cap=4 after two discarded magic cards (start 3), got %d", got)
	}
	if got := len(p2.Hand); got != 4 {
		t.Fatalf("expected hit branch add +2 damage from two discarded fire cards (target draw 4), got %d", got)
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected self take 2 magic damage from fire card count (self draw 2), got hand=%d", got)
	}
}

func requireResponseSkillContains(t *testing.T, game *GameEngine, playerID, skillID string) {
	t.Helper()
	requireResponseSkillPrompt(t, game, playerID)
	for _, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == skillID {
			return
		}
	}
	t.Fatalf("expected pending response skills for %s contain %s, got %+v", playerID, skillID, game.State.PendingInterrupt.SkillIDs)
}

func TestHeroRoar_AfterHitStillPromptsForbiddenPower(t *testing.T) {
	game := NewGameEngine(noopObserver{})
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
	p1.Crystal = 1
	p1.Gem = 0
	p1.Hand = []model.Card{
		heroTestCard("atk", "火斩", model.CardTypeAttack, model.ElementFire, 2),
	}
	p2.Hand = nil
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d4", "抽4", model.CardTypeAttack, model.ElementWater, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})

	chooseResponseSkillByID(t, game, "p1", "hero_roar")
	requireChoicePrompt(t, game, "p1", "hero_roar_draw")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	requireResponseSkillContains(t, game, "p1", "hero_forbidden_power")
}

func TestHeroRoar_AfterMissStillPromptsForbiddenPower(t *testing.T) {
	game := NewGameEngine(noopObserver{})
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
	p1.Crystal = 1
	p1.Gem = 0
	p1.Hand = []model.Card{
		heroTestCard("atk", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
	}
	p2.Hand = []model.Card{
		heroTestCard("guard", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})

	chooseResponseSkillByID(t, game, "p1", "hero_roar")
	requireChoicePrompt(t, game, "p1", "hero_roar_draw")
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

	requireResponseSkillContains(t, game, "p1", "hero_forbidden_power")
}

func TestHeroRoar_DrawOneWithOverflow_StillContinuesAttackAndPromptsForbiddenPower(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hero_anger"] = 1
	p1.Crystal = 1
	p1.Hand = []model.Card{
		heroTestCard("atk", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
		heroTestCard("m1", "法1", model.CardTypeMagic, model.ElementFire, 0),
		heroTestCard("m2", "法2", model.CardTypeMagic, model.ElementWater, 0),
		heroTestCard("m3", "法3", model.CardTypeMagic, model.ElementEarth, 0),
		heroTestCard("m4", "法4", model.CardTypeMagic, model.ElementWind, 0),
		heroTestCard("m5", "法5", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("d4", "抽4", model.CardTypeAttack, model.ElementFire, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "hero_roar")
	requireChoicePrompt(t, game, "p1", "hero_roar_draw")

	// 选择“摸1张”触发爆牌弃1
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected overflow discard interrupt after roar draw1, got %+v", game.State.PendingInterrupt)
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// 弃牌后攻击流程应继续到战斗响应，而不是直接结束回合
	if game.State.Phase != model.PhaseCombatInteraction {
		t.Fatalf("expected combat interaction after resolving overflow discard, got phase=%s", game.State.Phase)
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	requireResponseSkillContains(t, game, "p1", "hero_forbidden_power")
}

func TestHeroExhaustion_ReleaseAtTurnStart_Draw3Damage3_StillCanAct(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hero_exhaustion_form"] = 1
	p1.Tokens["hero_exhaustion_release_pending"] = 1
	p1.Hand = nil
	game.State.Deck = []model.Card{
		heroTestCard("d1", "起手攻击", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("d2", "补牌2", model.CardTypeMagic, model.ElementWater, 0),
		heroTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementLight, 2),
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	game.Drive()

	if got := p1.Tokens["hero_exhaustion_form"]; got != 0 {
		t.Fatalf("expected exhaustion form released at startup, got %d", got)
	}
	if got := p1.Tokens["hero_exhaustion_release_pending"]; got != 0 {
		t.Fatalf("expected exhaustion release pending flag cleared, got %d", got)
	}
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected release settlement draw (3 + damage draw 1 with short deck), got hand=%d", got)
	}
	if got := len(game.State.PendingDamageQueue); got != 0 {
		t.Fatalf("expected pending damage resolved before action phase, got %d", got)
	}
	if game.State.CurrentTurn != 0 {
		t.Fatalf("expected still hero turn after release settlement, got turn index %d", game.State.CurrentTurn)
	}
	if game.State.Phase != model.PhaseActionSelection {
		t.Fatalf("expected return to action selection after release settlement, got phase=%s", game.State.Phase)
	}

	attackIdx := firstAttackCardIndex(p1)
	if attackIdx < 0 {
		t.Fatalf("expected an attack card available after release settlement")
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: attackIdx,
	})
}

func TestHeroExhaustion_ReleaseWithOverflow_StillStartsTurnNormally(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hero_exhaustion_form"] = 1
	p1.Tokens["hero_exhaustion_release_pending"] = 1
	p1.Hand = []model.Card{
		heroTestCard("h1", "手牌1", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h2", "手牌2", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("h3", "手牌3", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("h4", "手牌4", model.CardTypeAttack, model.ElementWind, 2),
		heroTestCard("h5", "手牌5", model.CardTypeAttack, model.ElementThunder, 2),
		heroTestCard("h6", "手牌6", model.CardTypeAttack, model.ElementLight, 2),
	}
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementEarth, 2),
		heroTestCard("d4", "抽4", model.CardTypeAttack, model.ElementWind, 2),
		heroTestCard("d5", "抽5", model.CardTypeAttack, model.ElementThunder, 2),
		heroTestCard("d6", "抽6", model.CardTypeAttack, model.ElementLight, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected first overflow discard during exhaustion release, got %+v", game.State.PendingInterrupt)
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1, 2},
	})

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf(
			"expected second overflow discard during exhaustion self-damage draw, got intr=%+v phase=%s pendingDamage=%d hand=%d maxHand=%d returnPhase=%s",
			game.State.PendingInterrupt,
			game.State.Phase,
			len(game.State.PendingDamageQueue),
			len(p1.Hand),
			game.GetMaxHand(p1),
			game.State.ReturnPhase,
		)
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1, 2},
	})

	if game.State.CurrentTurn != 0 {
		t.Fatalf("expected still p1 turn after exhaustion overflow settlement, got turn=%d", game.State.CurrentTurn)
	}
	if game.State.Phase != model.PhaseActionSelection {
		t.Fatalf("expected action selection after exhaustion overflow settlement, got phase=%s", game.State.Phase)
	}
}

func TestHeroCalmMind_DisablesCounterAndAttackEndGainCrystal(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 0
	p1.Gem = 0
	p1.Tokens["hero_wisdom"] = 4
	p1.Hand = []model.Card{
		heroTestCard("atk", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "hero_calm_mind")

	if len(game.State.CombatStack) == 0 {
		t.Fatalf("expected combat stack after selecting calm mind")
	}
	top := game.State.CombatStack[len(game.State.CombatStack)-1]
	if top.CanBeResponded {
		t.Fatalf("expected calm mind to disable counter response")
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if got := p1.Tokens["hero_wisdom"]; got != 0 {
		t.Fatalf("expected wisdom consumed to 0, got %d", got)
	}
	if got := p1.Crystal; got != 1 {
		t.Fatalf("expected calm mind grant +1 crystal at attack end, got %d", got)
	}
}

func TestHeroTaunt_NonAttackActionSkipsAndRemoves(t *testing.T) {
	game := NewGameEngine(noopObserver{})
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
	game.State.Phase = model.PhaseActionSelection

	if err := game.UseSkill("p1", "hero_taunt", []string{"p2"}, nil); err != nil {
		t.Fatalf("use hero_taunt failed: %v", err)
	}

	if getFieldEffectCard(p2, model.EffectHeroTaunt) == nil {
		t.Fatalf("expected taunt effect placed on target")
	}

	p1.IsActive = false
	p2.IsActive = true
	p2.TurnState = model.NewPlayerTurnState()
	p2.Hand = []model.Card{
		heroTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 1
	game.State.Phase = model.PhaseActionSelection

	beforeHand := len(p2.Hand)
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdMagic,
		TargetID:  "p1",
		CardIndex: 0,
	})
	if got := len(p2.Hand); got != beforeHand {
		t.Fatalf("expected taunted player non-attack action be skipped without card use, hand %d -> %d", beforeHand, got)
	}
	if getFieldEffectCard(p2, model.EffectHeroTaunt) != nil {
		t.Fatalf("expected taunt effect removed after forced skip")
	}
}

func TestHeroDeadDuel_MagicOverflowMoraleLossFlooredToOne(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.Hand = []model.Card{
		heroTestCard("h1", "卡1", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h2", "卡2", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h3", "卡3", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h4", "卡4", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h5", "卡5", model.CardTypeAttack, model.ElementFire, 2),
		heroTestCard("h6", "卡6", model.CardTypeAttack, model.ElementFire, 2),
	}
	p1.Gem = 1
	p1.Tokens["hero_anger"] = 0
	game.State.Deck = []model.Card{
		heroTestCard("d1", "抽1", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d2", "抽2", model.CardTypeAttack, model.ElementWater, 2),
		heroTestCard("d3", "抽3", model.CardTypeAttack, model.ElementWater, 2),
	}

	game.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     3,
		DamageType: "magic",
		Stage:      0,
	})

	if interrupted := game.processPendingDamages(); !interrupted {
		t.Fatalf("expected dead-duel response interrupt")
	}
	chooseResponseSkillByID(t, game, "p1", "hero_dead_duel")

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt after magic overflow, got %+v", game.State.PendingInterrupt)
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1, 2},
	})

	if got := game.State.RedMorale; got != 14 {
		t.Fatalf("expected dead duel floor morale loss to 1 (15->14), got %d", got)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected dead duel consume 1 gem, got %d", got)
	}
	if got := p1.Tokens["hero_anger"]; got != 3 {
		t.Fatalf("expected dead duel add 3 anger, got %d", got)
	}
}
