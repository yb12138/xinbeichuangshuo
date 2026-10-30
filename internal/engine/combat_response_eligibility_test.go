package engine

import (
	"testing"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// 构造一次 p1(红) 攻击 p2(蓝) 的应战窗口；p3 是 p1 的队友（应战反弹目标）。
func newCombatEligibilityGame(t *testing.T, defenderRole string, attack model.Card, defenderHand []model.Card) *GameEngine {
	t.Helper()
	game := NewGameEngine(nil)
	if err := game.AddPlayer("p1", "攻击方", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "防守方", defenderRole, model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "攻击方队友", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	game.State.Players["p2"].Hand = defenderHand
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.CombatStage = model.CombatStageHitCheck
	game.State.Subflow = model.SubflowNone
	game.State.PendingInterrupt = nil
	game.State.CombatStack = []model.CombatRequest{{
		AttackerID:     "p1",
		TargetID:       "p2",
		Card:           &attack,
		CanBeResponded: true,
	}}
	return game
}

func promptOptionSet(p *model.Prompt) map[string]bool {
	out := map[string]bool{}
	for _, o := range p.Options {
		out[o.ID] = true
	}
	return out
}

var thunderAttack = model.Card{ID: "atk", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2}

// 截图场景：被雷系攻击，手里只有虚弱 + 水/风/火攻击牌——既不能防御也不能应战。
func TestCombatResponsePrompt_NoUsableCards_OnlyTake(t *testing.T) {
	game := newCombatEligibilityGame(t, "elementalist", thunderAttack, []model.Card{
		{ID: "weak", Name: "虚弱", Type: model.CardTypeMagic, Element: model.ElementEarth},
		{ID: "water", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		{ID: "wind", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		{ID: "fire", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	})
	prompt := game.GetCurrentPrompt()
	if prompt == nil {
		t.Fatal("expected combat prompt")
	}
	opts := promptOptionSet(prompt)
	if !opts["take"] || opts["defend"] || opts["counter"] {
		t.Fatalf("expected only take, got %+v", prompt.Options)
	}
	if len(prompt.DefendCardIDs) != 0 || len(prompt.CounterCardIDs) != 0 {
		t.Fatalf("expected no eligible cards, got defend=%v counter=%v", prompt.DefendCardIDs, prompt.CounterCardIDs)
	}
}

func TestCombatResponsePrompt_SameElementAndDarkCanCounter(t *testing.T) {
	game := newCombatEligibilityGame(t, "elementalist", thunderAttack, []model.Card{
		{ID: "thunder", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
		{ID: "dark", Name: "暗灭", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 2},
		{ID: "fire", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	})
	prompt := game.GetCurrentPrompt()
	opts := promptOptionSet(prompt)
	if !opts["counter"] || opts["defend"] {
		t.Fatalf("expected take+counter, got %+v", prompt.Options)
	}
	if got := prompt.CounterCardIDs; len(got) != 2 || got[0] != "thunder" || got[1] != "dark" {
		t.Fatalf("expected counter cards [thunder dark], got %v", got)
	}
}

func TestCombatResponsePrompt_HolyLightOnlyDefends(t *testing.T) {
	game := newCombatEligibilityGame(t, "elementalist", thunderAttack, []model.Card{
		{ID: "light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "shield", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
	})
	prompt := game.GetCurrentPrompt()
	opts := promptOptionSet(prompt)
	if !opts["defend"] || opts["counter"] {
		t.Fatalf("expected take+defend, got %+v", prompt.Options)
	}
	if got := prompt.DefendCardIDs; len(got) != 1 || got[0] != "light" {
		t.Fatalf("expected defend cards [light] (圣盾 must be placed beforehand), got %v", got)
	}
}

// 暗灭攻击不能应战，即使手里有暗灭。
func TestCombatResponsePrompt_DarkAttackCannotBeCountered(t *testing.T) {
	dark := model.Card{ID: "atk", Name: "暗灭", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 2}
	game := newCombatEligibilityGame(t, "elementalist", dark, []model.Card{
		{ID: "dark", Name: "暗灭", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 2},
	})
	prompt := game.GetCurrentPrompt()
	if promptOptionSet(prompt)["counter"] {
		t.Fatalf("dark attack must not offer counter, got %+v", prompt.Options)
	}
}

// 烈焰形态的苍炎魔女：非水/暗攻击牌视为火系，可应战火系攻击。
func TestCombatResponsePrompt_BlazeWitchFlameTransformsCounterCard(t *testing.T) {
	fire := model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2}
	game := newCombatEligibilityGame(t, "blaze_witch", fire, []model.Card{
		{ID: "earth", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
		{ID: "water", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
	})
	engineplayer.SetForm(game.State.Players["p2"], model.FormBlazeWitchFlame)
	prompt := game.GetCurrentPrompt()
	if got := prompt.CounterCardIDs; len(got) != 1 || got[0] != "earth" {
		t.Fatalf("expected flame-form earth attack to count as fire, got %v", got)
	}
}

// 魔剑士暗影抗拒：非自己回合可用【魔弹】应战。
func TestCombatResponsePrompt_MagicSwordsmanMagicBulletCounter(t *testing.T) {
	game := newCombatEligibilityGame(t, "magic_swordsman", thunderAttack, []model.Card{
		{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementFire, Damage: 2},
	})
	prompt := game.GetCurrentPrompt()
	if got := prompt.CounterCardIDs; len(got) != 1 || got[0] != "mb" {
		t.Fatalf("expected 魔弹 to be a valid counter card for magic swordsman off-turn, got %v (options %+v)", got, prompt.Options)
	}
}

// 提交校验与提示同源：提示里列出的应战牌一定能提交成功。
func TestCombatResponsePrompt_ListedCounterCardIsAccepted(t *testing.T) {
	game := newCombatEligibilityGame(t, "elementalist", thunderAttack, []model.Card{
		{ID: "thunder", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
	})
	prompt := game.GetCurrentPrompt()
	if len(prompt.CounterCardIDs) != 1 {
		t.Fatalf("expected one counter card, got %v", prompt.CounterCardIDs)
	}
	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"counter"},
		CardID:    prompt.CounterCardIDs[0],
		TargetID:  "p3",
	})
	if err != nil {
		t.Fatalf("listed counter card should be accepted, got %v", err)
	}
}

// 阴阳师可用同命格攻击牌应战：提示带上攻击牌命格，前端才能写"同系、同命格或暗灭"。
func TestCombatResponsePrompt_OnmyojiExposesFactionCounter(t *testing.T) {
	fire := model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2, Faction: "咏"}
	game := newCombatEligibilityGame(t, "onmyoji", fire, []model.Card{
		{ID: "water", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2, Faction: "咏"},
	})
	prompt := game.GetCurrentPrompt()
	if prompt.AttackFaction != "咏" {
		t.Fatalf("onmyoji should see the attack faction, got %q", prompt.AttackFaction)
	}
	if !promptOptionSet(prompt)["counter"] || len(prompt.CounterCardIDs) != 1 {
		t.Fatalf("same-faction off-element card should be a counter card, got options=%v cards=%v", prompt.Options, prompt.CounterCardIDs)
	}
}

// 其他角色没有命格应战规则：不带命格，避免误导。
func TestCombatResponsePrompt_NonOnmyojiHasNoFaction(t *testing.T) {
	fire := model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2, Faction: "咏"}
	game := newCombatEligibilityGame(t, "butterfly_dancer", fire, []model.Card{
		{ID: "fire2", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	})
	if got := game.GetCurrentPrompt().AttackFaction; got != "" {
		t.Fatalf("non-onmyoji should not get a faction hint, got %q", got)
	}
}
