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

func TestMagicalGirl_MagicBulletFusionChain_ResponseSkillPassesAlongDirection(t *testing.T) {
	game := buildMagicBulletFusionChainGame(t, "magical_girl")
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Hand = []model.Card{
		{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "fire-attack", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire},
	}

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: testutils.PlayableCardID(t, game, "p1", 0)}); err != nil {
		t.Fatalf("cast magic bullet failed: %v", err)
	}
	assertResponseSkillInterrupt(t, game, "p2", "magic_bullet_fusion_chain")

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdSelect, Selections: []int{0}}); err != nil {
		t.Fatalf("select fusion chain response skill failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected discard choice after selecting chain fusion, got %+v", game.State.PendingInterrupt)
	}
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdSelect, Selections: []int{0}}); err != nil {
		t.Fatalf("discard fusion chain card failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("expected magic missile interrupt after chain fusion, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected original direction to continue from p2 to p3, got %s", game.State.PendingInterrupt.PlayerID)
	}
	if game.State.PendingInterrupt.Type == model.InterruptMagicBulletDirection {
		t.Fatalf("chain fusion must not trigger magic bullet direction")
	}
	chain := game.State.MagicBulletChain
	if chain == nil {
		t.Fatalf("expected magic bullet chain to continue")
	}
	if got := chain.CurrentDamage; got != 3 {
		t.Fatalf("expected chain damage +1 after fusion, got %d", got)
	}
	if chain.SourcePlayerID != "p2" {
		t.Fatalf("expected chain source to become p2, got %s", chain.SourcePlayerID)
	}
	if chain.IsFusion || chain.FusionCard != nil {
		t.Fatalf("chain response fusion should not rewrite active fusion metadata, got %+v", chain)
	}
	if len(p2.Hand) != 0 {
		t.Fatalf("expected fusion card consumed from hand, got %+v", p2.Hand)
	}
	if len(game.State.DiscardPile) == 0 || game.State.DiscardPile[len(game.State.DiscardPile)-1].ID != "fire-attack" {
		t.Fatalf("expected fusion card in discard pile, got %+v", game.State.DiscardPile)
	}
}

func TestMagicalGirl_MagicBulletFusionChain_SkipRestoresNormalMissilePrompt(t *testing.T) {
	game := buildMagicBulletFusionChainGame(t, "magical_girl")
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Hand = []model.Card{{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}
	p2.Hand = []model.Card{{ID: "earth", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth}}

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: testutils.PlayableCardID(t, game, "p1", 0)}); err != nil {
		t.Fatalf("cast magic bullet failed: %v", err)
	}
	assertResponseSkillInterrupt(t, game, "p2", "magic_bullet_fusion_chain")

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdCancel}); err != nil {
		t.Fatalf("skip fusion chain response skill failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("expected original missile prompt after skip, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected missile prompt to remain on p2 after skip, got %s", game.State.PendingInterrupt.PlayerID)
	}
}

func TestMagicalGirl_MagicBulletFusionChain_HiddenAfterParticipated(t *testing.T) {
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

	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: testutils.PlayableCardID(t, game, "p1", 0)}); err != nil {
		t.Fatalf("cast magic bullet failed: %v", err)
	}
	assertResponseSkillInterrupt(t, game, "p2", "magic_bullet_fusion_chain")
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdCancel}); err != nil {
		t.Fatalf("skip fusion chain response skill failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: testutils.PlayableCardID(t, game, "p2", 1)}); err != nil {
		t.Fatalf("p2 counter magic bullet failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicMissile || game.State.PendingInterrupt.PlayerID != "p3" {
		t.Fatalf("expected missile to pass to p3, got %+v", game.State.PendingInterrupt)
	}
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p3", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: testutils.PlayableCardID(t, game, "p3", 0)}); err != nil {
		t.Fatalf("p3 counter magic bullet failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("expected normal missile prompt for already participated magical girl, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected p2 to be targeted again, got %s", game.State.PendingInterrupt.PlayerID)
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
	game := engine.NewGameEngine(testutils.NoopObserver{})
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

	if got := len(p2.Hand); got != 2 {
		t.Fatalf("expected p2 draw 2 after taking magic damage, got %d", got)
	}
	if got := len(p3.Hand); got != 2 {
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
