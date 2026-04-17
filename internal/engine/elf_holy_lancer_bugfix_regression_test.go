package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func skillIndex(skillIDs []string, want string) int {
	for i, sid := range skillIDs {
		if sid == want {
			return i
		}
	}
	return -1
}

func findLatestCombatPromptForPlayer(obs *captureObserver, playerID string) *model.Prompt {
	if obs == nil {
		return nil
	}
	for i := len(obs.events) - 1; i >= 0; i-- {
		event := obs.events[i]
		if event.Type != model.EventAskInput {
			continue
		}
		prompt, ok := event.Data.(*model.Prompt)
		if !ok || prompt == nil || prompt.PlayerID != playerID {
			continue
		}
		if promptHasOptionID(prompt, "take") || promptHasOptionID(prompt, "defend") || promptHasOptionID(prompt, "counter") {
			copied := *prompt
			copied.Options = append([]model.PromptOption(nil), prompt.Options...)
			return &copied
		}
	}
	return nil
}

func promptHasEffectHintContains(prompt *model.Prompt, wantFragment string) bool {
	if prompt == nil {
		return false
	}
	for _, hint := range prompt.EffectHints {
		if hint == "" {
			continue
		}
		if strings.Contains(hint, wantFragment) {
			return true
		}
	}
	return false
}

// 回归：精灵射手元素射击触发雷之矢后，本次攻击应不可应战。
func TestElfElementalShotThunder_DisablesCounterResponse(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Elf", "elf_archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	p1.Hand = []model.Card{
		{ID: "atk-thunder", Name: "雷斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
		{ID: "magic-cost", Name: "水疗术", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	p2.Hand = []model.Card{
		{ID: "def-thunder", Name: "雷斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
	}

	mustDo(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected elemental-shot response interrupt, got %+v", game.State.PendingInterrupt)
	}
	shotIdx := skillIndex(game.State.PendingInterrupt.SkillIDs, "elf_elemental_shot")
	if shotIdx < 0 {
		t.Fatalf("expected elf_elemental_shot in pending skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	mustDo(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{shotIdx},
	})
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected elemental-shot cost choice interrupt, got %+v", game.State.PendingInterrupt)
	}
	mustDo(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 选择“弃法术牌”
	})
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected elemental-shot discard choice interrupt, got %+v", game.State.PendingInterrupt)
	}
	mustDo(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0}, // 弃唯一法术牌
	})

	if got := len(game.State.CombatStack); got != 1 {
		t.Fatalf("expected combat stack size 1, got %d", got)
	}
	if game.State.CombatStack[0].CanBeResponded {
		t.Fatalf("expected thunder elemental shot to disable counter response")
	}
	if p1.Tokens["elf_elemental_shot_thunder_pending"] != 0 {
		t.Fatalf("expected thunder shot to stop using token state, got %d", p1.Tokens["elf_elemental_shot_thunder_pending"])
	}
	for _, modifier := range p1.ActiveRuleModifiers {
		if modifier != nil && modifier.ModifierID == "elf_elemental_shot_thunder_attack_tag" {
			t.Fatalf("expected thunder combat-policy rule consumed before combat request, modifier still present: %+v", modifier)
		}
	}
}

// 回归：宠物强化目标摸牌触发爆牌时，不应再追加第二次“弃1”中断。
func TestElfPetEmpower_OverflowConsumesDiscardOnlyOnce(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Elf", "elf_archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Hand = []model.Card{
		{ID: "atk-dark", Name: "暗斩", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 1},
	}
	p2.Hand = make([]model.Card, game.GetMaxHand(p2)-1)
	for i := range p2.Hand {
		p2.Hand[i] = model.Card{
			ID:      "h" + string(rune('a'+i)),
			Name:    "手牌",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  1,
		}
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
	requireResponseSkillPrompt(t, game, "p1")
	if skillIndex(game.State.PendingInterrupt.SkillIDs, "elf_animal_companion") < 0 ||
		skillIndex(game.State.PendingInterrupt.SkillIDs, "elf_pet_empower") < 0 {
		t.Fatalf("expected animal companion and pet empower in pending skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	mustDo(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSelect,
		Selections: []int{
			skillIndex(game.State.PendingInterrupt.SkillIDs, "elf_pet_empower"),
		},
	})

	if game.State.PendingInterrupt == nil || !isDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected only overflow discard interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected discard owner p2, got %s", game.State.PendingInterrupt.PlayerID)
	}
	if len(game.State.InterruptQueue) != 0 {
		t.Fatalf("expected no extra queued discard interrupt, got queue size %d", len(game.State.InterruptQueue))
	}
	if p1.Crystal != 0 {
		t.Fatalf("expected pet empower to consume exactly 1 crystal, got %d", p1.Crystal)
	}
}

// 回归：圣枪骑士攻击命中后，地枪与圣击互斥；若跳过地枪，应补触发圣击+1治疗。
func TestHolyLancer_EarthSpearAndHolyStrikeMutualExclusion(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "HolyLancer", "holy_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 1
	p1.Tokens = map[string]int{}
	p1.TurnState.UsedSkillCounts["holy_lancer_prayer"] = 1
	p1.Hand = []model.Card{
		{ID: "atk-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
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
		t.Fatalf("expected attack-hit response-skill interrupt, got %+v", game.State.PendingInterrupt)
	}
	if skillIndex(game.State.PendingInterrupt.SkillIDs, "holy_lancer_earth_spear") < 0 {
		t.Fatalf("expected holy_lancer_earth_spear in skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	if skillIndex(game.State.PendingInterrupt.SkillIDs, "holy_lancer_holy_strike") >= 0 {
		t.Fatalf("holy_lancer_holy_strike should not be offered when earth spear is available")
	}

	// 选择“跳过响应”（索引=技能数），应补触发圣击+1治疗。
	mustDo(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{len(game.State.PendingInterrupt.SkillIDs)},
	})

	if p1.Heal != 2 {
		t.Fatalf("expected holy strike fallback to heal +1 after skipping earth spear, got heal=%d", p1.Heal)
	}
}

// 回归：圣枪骑士天枪响应后，本次攻击应不可应战。
func TestHolyLancer_SkySpearDisablesCounterResponse(t *testing.T) {
	obs := &captureObserver{}
	game := NewGameEngine(obs)
	if err := game.AddPlayer("p1", "HolyLancer", "holy_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 3
	p1.Hand = []model.Card{
		{ID: "atk-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "counter-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	mustDo(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected attack-start response-skill interrupt, got %+v", game.State.PendingInterrupt)
	}
	skyIdx := skillIndex(game.State.PendingInterrupt.SkillIDs, "holy_lancer_sky_spear")
	if skyIdx < 0 {
		t.Fatalf("expected holy_lancer_sky_spear in skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	mustDo(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{skyIdx},
	})

	if p1.Heal != 1 {
		t.Fatalf("expected sky spear to consume 2 heal, got heal=%d", p1.Heal)
	}
	if got := len(game.State.CombatStack); got != 1 {
		t.Fatalf("expected combat stack size 1, got %d", got)
	}
	if game.State.CombatStack[0].CanBeResponded {
		t.Fatalf("expected sky spear to disable counter response")
	}
	prompt := findLatestCombatPromptForPlayer(obs, "p2")
	if prompt == nil {
		t.Fatalf("expected combat response prompt for p2")
	}
	if got := prompt.AttackElement; got != string(model.ElementDark) {
		t.Fatalf("expected sky spear prompt attack_element=Dark, got %q", got)
	}
	if !promptHasEffectHintContains(prompt, "无法应战") {
		t.Fatalf("expected sky spear prompt to include unrespondable hint, got %+v", prompt.EffectHints)
	}
	if promptHasOptionID(prompt, "counter") {
		t.Fatalf("expected no counter option after sky spear, got %+v", prompt.Options)
	}
}
