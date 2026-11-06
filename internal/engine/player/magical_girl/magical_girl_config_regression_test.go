package magical_girl_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestMagicalGirl_MagicBulletFusion_ReverseChainUsesConfiguredDirection(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Ally", "saintess", model.RedCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p4 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.PlayerOrder = []string{"p1", "p2", "p3", "p4"}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "earth-attack", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "magic_bullet_fusion",
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("active fusion failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicBulletDirection {
		t.Fatalf("expected direction interrupt after fusion, got %+v", game.State.PendingInterrupt)
	}
	if !p1.TurnState.HasActed {
		t.Fatalf("expected active fusion to count as an action")
	}
	if got := p1.TurnState.LastActionType; got != string(model.ActionMagic) {
		t.Fatalf("expected active fusion to count as magic action, got %q", got)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("choose reverse direction failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("expected magic missile interrupt after direction choice, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected reverse chain to target p3, got %s", game.State.PendingInterrupt.PlayerID)
	}
	if game.State.MagicBulletChain == nil || !game.State.MagicBulletChain.Reverse {
		t.Fatalf("expected reverse magic bullet chain, got %+v", game.State.MagicBulletChain)
	}
	if game.State.MagicBulletChain == nil || !game.State.MagicBulletChain.IsFusion {
		t.Fatalf("expected fusion magic bullet chain, got %+v", game.State.MagicBulletChain)
	}
	if game.State.MagicBulletChain.FusionCard == nil || game.State.MagicBulletChain.FusionCard.ID != "earth-attack" {
		t.Fatalf("expected chain to remember active fusion card, got %+v", game.State.MagicBulletChain.FusionCard)
	}
}

func TestMagicalGirl_MagicBulletFusion_DirectEarthMagicResolvesAsOriginalSpell(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "shield", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementEarth},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdMagic,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("cast earth magic failed: %v", err)
	}

	if game.State.MagicBulletChain != nil {
		t.Fatalf("expected no magic bullet chain after direct earth magic, got %+v", game.State.MagicBulletChain)
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptMagicBulletFusion {
		t.Fatalf("direct earth magic should not prompt old fusion interrupt")
	}
	if !p2.HasFieldEffect(model.EffectShield) {
		t.Fatalf("expected original shield effect to be placed on target")
	}
}

func TestMagicalGirl_MagicBulletFusion_RejectsNonFireEarthCard(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "water", Name: "水波", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "magic_bullet_fusion",
		Selections: []int{0},
	}); err == nil {
		t.Fatalf("expected non fire/earth fusion card to be rejected")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after rejected fusion, got %+v", game.State.PendingInterrupt)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected rejected card to remain in hand, got %d", got)
	}
}

func TestMagicalGirl_MagicBulletFusionChain_NonMagicalGirlDoesNotOfferSkill(t *testing.T) {
	game := buildMagicBulletFusionChainGame(t, "berserker")
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Hand = []model.Card{{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}
	p2.Hand = []model.Card{{ID: "earth", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth}}

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: testutils.PlayableCardID(t, game, "p1", 0)}); err != nil {
		t.Fatalf("cast magic bullet failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("expected normal missile prompt for non magical girl, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected p2 as missile target, got %s", game.State.PendingInterrupt.PlayerID)
	}
}

func buildMagicBulletFusionChainGame(t *testing.T, targetRole string) *engine.GameEngine {
	t.Helper()
	return buildMagicBulletFusionChainGameWithObs(t, targetRole, testutils.NoopObserver{})
}

func buildMagicBulletFusionChainGameWithObs(t *testing.T, targetRole string, obs model.GameObserver) *engine.GameEngine {
	t.Helper()
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Caster", "berserker", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Target", targetRole, model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Relay", "berserker", model.RedCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}
	if err := game.AddPlayer("p4", "BlueAlly", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p4 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.PlayerOrder = []string{"p1", "p3", "p4", "p2"}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	for _, pid := range []string{"p2", "p3", "p4"} {
		game.State.Players[pid].TurnState = model.NewPlayerTurnState()
	}
	return game
}

func assertResponseSkillInterrupt(t *testing.T, game *engine.GameEngine, playerID string, skillID string) {
	t.Helper()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response skill interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected response skill interrupt for %s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
	for _, got := range game.State.PendingInterrupt.SkillIDs {
		if got == skillID {
			return
		}
	}
	t.Fatalf("expected response skill %s in %+v", skillID, game.State.PendingInterrupt.SkillIDs)
}

func TestMagicalGirl_MagicBlast_DiscardsAfterEachFailedTarget(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	game.State.Deck = rules.InitDeck()
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}

	// 座位 p1,p3,p2：从 p1 起逆时针为 p2→p3（魔爆冲击按逆时针结算目标）。
	game.State.PlayerOrder = []string{"p1", "p3", "p2"}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "cost", Name: "法术代价", Type: model.CardTypeMagic, Element: model.ElementWater},
		{ID: "a1", Name: "弃牌1", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "m1", Name: "弃牌2", Type: model.CardTypeMagic, Element: model.ElementEarth},
	}
	// 目标手里有法术牌才会弹出选择（没有法术牌会自动按不弃牌处理）。
	p2.Hand = []model.Card{{ID: "p2m", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder}}
	p3.Hand = []model.Card{{ID: "p3m", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder}}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "magic_blast",
		TargetIDs:  []string{"p2", "p3"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("magic blast failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected first target interrupt on p2, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdCancel}); err != nil {
		t.Fatalf("p2 decline failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected caster forced discard after first decline, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("caster first discard failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected second target interrupt on p3, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p3", Type: model.CmdCancel}); err != nil {
		t.Fatalf("p3 decline failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected caster forced discard after second decline, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("caster second discard failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected magic blast flow to end, got %+v", game.State.PendingInterrupt)
	}
	if got := len(p1.Hand); got != 0 {
		t.Fatalf("expected caster hand to be empty after 1 cost + 2 forced discards, got %d", got)
	}
	if got := game.State.RedGems; got != 1 {
		t.Fatalf("expected team gem +1 from magic blast, got %d", got)
	}

	game.Drive()
	game.Drive()

	if got := len(p2.Hand); got != 3 {
		t.Fatalf("expected p2 draw 2 after taking magic damage, got %d", got)
	}
	if got := len(p3.Hand); got != 3 {
		t.Fatalf("expected p3 draw 2 after taking magic damage, got %d", got)
	}
}

func TestMagicalGirl_MagicBlast_TargetCanSelectMagicByHandIndex(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}

	// 座位 p1,p3,p2：从 p1 起逆时针为 p2→p3（魔爆冲击按逆时针结算目标）。
	game.State.PlayerOrder = []string{"p1", "p3", "p2"}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "cost", Name: "法术代价", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	game.State.Players["p3"].Hand = []model.Card{
		{ID: "p3m", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder},
	}
	p2.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "a2", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth},
		{ID: "a3", Name: "风刃", Type: model.CardTypeAttack, Element: model.ElementWind},
		{ID: "m1", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "magic_blast",
		TargetIDs:  []string{"p2", "p3"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("magic blast failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected first target interrupt on p2, got %+v", game.State.PendingInterrupt)
	}

	// 前端 choose_cards 会提交手牌真实索引；这里验证后端兼容该编码。
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{3},
	}); err != nil {
		t.Fatalf("expected selecting magic card by hand index to work, got error: %v", err)
	}

	if got := len(p2.Hand); got != 3 {
		t.Fatalf("expected p2 to discard one magic card, got hand size %d", got)
	}
	for _, card := range p2.Hand {
		if card.Type == model.CardTypeMagic {
			t.Fatalf("expected p2 magic card to be discarded, remaining hand=%+v", p2.Hand)
		}
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected flow to advance to second target p3, got %+v", game.State.PendingInterrupt)
	}
}

// 前端选牌面板提交的是 card_ids（不是下标）；魔爆冲击的目标弃牌必须能按 card_ids 结算，
// 否则点「确认」只会收到“请选择弃一张法术牌”的错误，提示一直不消失。
func TestMagicalGirl_MagicBlast_TargetCanSelectMagicByCardID(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}

	// 座位 p1,p3,p2：从 p1 起逆时针为 p2→p3（魔爆冲击按逆时针结算目标）。
	game.State.PlayerOrder = []string{"p1", "p3", "p2"}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "cost", Name: "法术代价", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	game.State.Players["p3"].Hand = []model.Card{
		{ID: "p3m", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder},
	}
	p2.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "m1", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder},
		{ID: "a2", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "magic_blast",
		TargetIDs:  []string{"p2", "p3"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("magic blast failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected first target interrupt on p2, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2",
		Type:     model.CmdSelect,
		CardIDs:  []string{"m1"},
	}); err != nil {
		t.Fatalf("expected selecting magic card by card id to work, got error: %v", err)
	}
	if got := len(p2.Hand); got != 2 {
		t.Fatalf("expected p2 to discard one magic card, got hand size %d", got)
	}
	for _, card := range p2.Hand {
		if card.ID == "m1" {
			t.Fatalf("expected m1 to be discarded, remaining hand=%+v", p2.Hand)
		}
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected flow to advance to second target p3, got %+v", game.State.PendingInterrupt)
	}
}

func TestMagicalGirl_DestructionStorm_RequiresTwoTargetsAndCostsOneGem(t *testing.T) {
	t.Run("requires_two_targets", func(t *testing.T) {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
			t.Fatalf("add p1 failed: %v", err)
		}
		if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
			t.Fatalf("add p2 failed: %v", err)
		}
		if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
			t.Fatalf("add p3 failed: %v", err)
		}

		game.State.CurrentTurn = 0
		game.State.TurnStage = model.TurnStageActionExecution

		p1 := game.State.Players["p1"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		p1.Gem = 1

		err := game.HandleAction(model.PlayerAction{
			PlayerID:  "p1",
			Type:      model.CmdSkill,
			SkillID:   "destruction_storm",
			TargetIDs: []string{"p2"},
		})
		if err == nil {
			t.Fatalf("expected destruction storm to reject fewer than two targets")
		}
	})

	t.Run("costs_one_gem", func(t *testing.T) {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		game.State.Deck = rules.InitDeck()
		if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
			t.Fatalf("add p1 failed: %v", err)
		}
		if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
			t.Fatalf("add p2 failed: %v", err)
		}
		if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
			t.Fatalf("add p3 failed: %v", err)
		}

		game.State.CurrentTurn = 0
		game.State.TurnStage = model.TurnStageActionExecution

		p1 := game.State.Players["p1"]
		p2 := game.State.Players["p2"]
		p3 := game.State.Players["p3"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		p2.TurnState = model.NewPlayerTurnState()
		p3.TurnState = model.NewPlayerTurnState()
		p1.Gem = 1

		if err := game.HandleAction(model.PlayerAction{
			PlayerID:  "p1",
			Type:      model.CmdSkill,
			SkillID:   "destruction_storm",
			TargetIDs: []string{"p2", "p3"},
		}); err != nil {
			t.Fatalf("destruction storm failed: %v", err)
		}

		if got := p1.Gem; got != 0 {
			t.Fatalf("expected destruction storm to consume exactly 1 gem, got %d", got)
		}

		game.Drive()
		game.Drive()

		if got := len(p2.Hand); got != 2 {
			t.Fatalf("expected p2 draw 2 after destruction storm, got %d", got)
		}
		if got := len(p3.Hand); got != 2 {
			t.Fatalf("expected p3 draw 2 after destruction storm, got %d", got)
		}
	})
}

// 魔法少女-4：目标没有法术牌时不弹选择框，直接按"不弃牌（受到2点伤害）"结算。
func TestMagicalGirl_MagicBlast_TargetWithoutMagicCardAutoRefuses(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	game.State.Deck = rules.InitDeck()
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy1", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Enemy2", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}
	// 座位 p1,p3,p2：从 p1 起逆时针为 p2→p3（魔爆冲击按逆时针结算目标）。
	game.State.PlayerOrder = []string{"p1", "p3", "p2"}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "cost", Name: "法术代价", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	// p2 只有攻击牌，p3 有法术牌
	p2.Hand = []model.Card{{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire}}
	p3.Hand = []model.Card{{ID: "m1", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder}}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "magic_blast",
		TargetIDs:  []string{"p2", "p3"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("magic blast failed: %v", err)
	}
	// 施法者此时没有手牌，不需要强制弃牌；直接轮到 p3 选择。
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected p2 auto-refused and prompt moves to p3, got %+v", game.State.PendingInterrupt)
	}
	found := false
	for _, pd := range game.State.PendingDamageQueue {
		if pd.TargetID == "p2" && pd.Damage == 2 && pd.DamageType == model.MagicAttack {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 2 magic damage queued for p2, queue=%+v", game.State.PendingDamageQueue)
	}
}

// 审计 魔法少女-3：魔爆冲击的两名目标按"从当前回合角色开始逆时针"结算，与点选顺序无关。
func TestMagicalGirl_MagicBlast_TargetsResolveCounterclockwise(t *testing.T) {
	for _, picked := range [][]string{{"p2", "p3"}, {"p3", "p2"}} {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		for _, p := range []struct {
			id, role string
			camp     model.Camp
		}{{"p1", "magical_girl", model.RedCamp}, {"p2", "berserker", model.BlueCamp}, {"p3", "angel", model.BlueCamp}} {
			if err := game.AddPlayer(p.id, p.id, p.role, p.camp); err != nil {
				t.Fatal(err)
			}
		}
		game.State.PlayerOrder = []string{"p1", "p2", "p3"} // 逆时针：p1→p3→p2
		game.State.CurrentTurn = 0
		game.State.TurnStage = model.TurnStageActionExecution
		p1 := game.State.Players["p1"]
		p1.IsActive = true
		p1.TurnState = model.NewPlayerTurnState()
		p1.Hand = []model.Card{{ID: "cost", Name: "法术代价", Type: model.CardTypeMagic, Element: model.ElementWater}}
		game.State.Players["p2"].Hand = []model.Card{{ID: "p2m", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder}}
		game.State.Players["p3"].Hand = []model.Card{{ID: "p3m", Name: "雷鸣术", Type: model.CardTypeMagic, Element: model.ElementThunder}}
		testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "magic_blast", TargetIDs: picked, Selections: []int{0}})
		if intr := game.State.PendingInterrupt; intr == nil || intr.PlayerID != "p3" {
			t.Fatalf("picked %v: expected p3 (counterclockwise first) to respond first, got %+v", picked, intr)
		}
	}
}

func missilePromptCounterIDs(t *testing.T, game *engine.GameEngine, playerID string) map[string]bool {
	t.Helper()
	intr := game.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptMagicMissile || intr.PlayerID != playerID {
		t.Fatalf("expected magic missile prompt for %s, got %+v", playerID, intr)
	}
	prompt := game.GetCurrentPrompt()
	if prompt == nil {
		t.Fatalf("expected missile prompt")
	}
	ids := map[string]bool{}
	for _, id := range prompt.CounterCardIDs {
		ids[id] = true
	}
	return ids
}

// 审计 魔法少女-2：被魔弹指向时，地/火系牌直接作为"传递"牌出现在魔弹响应面板里，一步完成魔弹融合。
func TestMagicalGirl_MagicBulletFusionChain_CounterInMissilePanel(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := buildMagicBulletFusionChainGameWithObs(t, "magical_girl", obs)
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Hand = []model.Card{{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}
	p2.Hand = []model.Card{{ID: "fire-attack", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire}}

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: testutils.PlayableCardID(t, game, "p1", 0)})
	if ids := missilePromptCounterIDs(t, game, "p2"); !ids["fire-attack"] {
		t.Fatalf("expected fire card offered as missile counter, got %v", ids)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: "fire-attack"})

	intr := game.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptMagicMissile || intr.PlayerID != "p3" {
		t.Fatalf("expected missile passed along original direction to p3, got %+v", intr)
	}
	chain := game.State.MagicBulletChain
	if chain == nil || chain.CurrentDamage != 3 || chain.SourcePlayerID != "p2" {
		t.Fatalf("expected chain damage 3 from p2, got %+v", chain)
	}
	if len(p2.Hand) != 0 || game.State.DiscardPile[len(game.State.DiscardPile)-1].ID != "fire-attack" {
		t.Fatalf("expected fusion card discarded, hand=%+v", p2.Hand)
	}
	testutils.RequireSingleSkillActivation(t, obs, "magic_bullet_fusion_chain", "p2", "p3")
}

func TestMagicalGirl_MagicBulletFusionChain_NotOfferedAfterParticipated(t *testing.T) {
	game := buildMagicBulletFusionChainGame(t, "magical_girl")
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.Hand = []model.Card{{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}
	p2.Hand = []model.Card{
		{ID: "earth", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth},
		{ID: "mb-p2", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p3.Hand = []model.Card{{ID: "mb-p3", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: testutils.PlayableCardID(t, game, "p1", 0)})
	missilePromptCounterIDs(t, game, "p2")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: "mb-p2"})
	missilePromptCounterIDs(t, game, "p3")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p3", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: "mb-p3"})
	if ids := missilePromptCounterIDs(t, game, "p2"); ids["earth"] {
		t.Fatalf("already participated magical girl must not fuse again, got %v", ids)
	}
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: "earth"}); err == nil {
		t.Fatalf("expected fusion rejected after participating")
	}
}

func TestMagicalGirl_MagicBulletFusionChain_NonMagicalGirlCannotFuse(t *testing.T) {
	game := buildMagicBulletFusionChainGame(t, "berserker")
	game.State.Players["p1"].Hand = []model.Card{{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}
	game.State.Players["p2"].Hand = []model.Card{{ID: "earth", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth}}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: testutils.PlayableCardID(t, game, "p1", 0)})
	if ids := missilePromptCounterIDs(t, game, "p2"); ids["earth"] {
		t.Fatalf("non magical girl must not fuse, got %v", ids)
	}
}
