package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func soulSorcererTestCard(id, name string, cardType model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     ele,
		Faction:     "幻",
		Damage:      2,
		Description: name,
	}
}

func soulSorcererExclusiveCard(charName, skillTitle string) model.Card {
	return model.Card{
		ID:              "ss_ex_" + skillTitle,
		Name:            skillTitle,
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         "幻",
		Damage:          0,
		Description:     "灵魂术士测试专属卡",
		ExclusiveChar1:  charName,
		ExclusiveSkill1: skillTitle,
	}
}

func setupSoulSorcererActionTurn(t *testing.T) (*GameEngine, *model.Player, *model.Player) {
	t.Helper()
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	return game, p1, p2
}

func TestSoulSorcerer_StartGameInitAndStarterCard(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.StartGame(); err != nil {
		t.Fatalf("start game failed: %v", err)
	}
	p1 := game.State.Players["p1"]
	if p1 == nil {
		t.Fatalf("player p1 not found")
	}
	if got := p1.Tokens["ss_blue_soul"]; got != 0 {
		t.Fatalf("expected ss_blue_soul=0, got %d", got)
	}
	if got := p1.Tokens["ss_yellow_soul"]; got != 0 {
		t.Fatalf("expected ss_yellow_soul=0, got %d", got)
	}
	if got := p1.Tokens["ss_link_active"]; got != 0 {
		t.Fatalf("expected ss_link_active=0, got %d", got)
	}
	if p1.Character == nil {
		t.Fatalf("character missing")
	}
	if !p1.HasExclusiveCard(p1.Character.Name, "灵魂链接") {
		t.Fatalf("expected starter exclusive card 【灵魂链接】")
	}
}

func TestSoulSorcererSoulRecall_PickThenDone(t *testing.T) {
	game, p1, _ := setupSoulSorcererActionTurn(t)
	p1.Hand = []model.Card{
		soulSorcererTestCard("m1", "法术A", model.CardTypeMagic, model.ElementWater),
		soulSorcererTestCard("m2", "法术B", model.CardTypeMagic, model.ElementThunder),
		soulSorcererTestCard("a1", "攻击A", model.CardTypeAttack, model.ElementFire),
	}

	if err := game.UseSkill("p1", "ss_soul_recall", nil, nil); err != nil {
		t.Fatalf("use ss_soul_recall failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "ss_recall_pick")

	// 先选择第1张法术牌（选项0是“完成”，所以牌从1开始）。
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose recall card failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "ss_recall_pick")

	// 完成选择并结算。
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("finish recall failed: %v", err)
	}

	if got := p1.Tokens["ss_blue_soul"]; got != 1 {
		t.Fatalf("expected blue soul +1, got %d", got)
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected hand size 2 after discarding 1 card, got %d", got)
	}
	if got := len(game.State.DiscardPile); got != 1 {
		t.Fatalf("expected discard pile +1, got %d", got)
	}
}

func TestSoulSorcererSoulMirror_DrawUpToMaxHand(t *testing.T) {
	game, p1, p2 := setupSoulSorcererActionTurn(t)
	p1.Tokens["ss_yellow_soul"] = 2
	p1.Hand = []model.Card{
		soulSorcererTestCard("h1", "弃牌1", model.CardTypeAttack, model.ElementFire),
		soulSorcererTestCard("h2", "弃牌2", model.CardTypeMagic, model.ElementWater),
		soulSorcererTestCard("h3", "保留", model.CardTypeAttack, model.ElementWind),
	}
	p2.MaxHand = 6
	p2.Hand = []model.Card{
		soulSorcererTestCard("t1", "现有1", model.CardTypeAttack, model.ElementFire),
		soulSorcererTestCard("t2", "现有2", model.CardTypeAttack, model.ElementFire),
		soulSorcererTestCard("t3", "现有3", model.CardTypeAttack, model.ElementFire),
		soulSorcererTestCard("t4", "现有4", model.CardTypeAttack, model.ElementFire),
		soulSorcererTestCard("t5", "现有5", model.CardTypeAttack, model.ElementFire),
	}
	game.State.Deck = []model.Card{
		soulSorcererTestCard("d1", "补牌1", model.CardTypeMagic, model.ElementLight),
		soulSorcererTestCard("d2", "补牌2", model.CardTypeMagic, model.ElementDark),
	}

	if err := game.UseSkill("p1", "ss_soul_mirror", []string{"p2"}, []int{0, 1}); err != nil {
		t.Fatalf("use ss_soul_mirror failed: %v", err)
	}

	if got := p1.Tokens["ss_yellow_soul"]; got != 0 {
		t.Fatalf("expected yellow soul -2 to 0, got %d", got)
	}
	if got := len(p2.Hand); got != 6 {
		t.Fatalf("target should draw only up to max hand (6), got %d", got)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected user discarded 2 cards and kept 1, got %d", got)
	}
}

func TestSoulSorcererSoulBlast_ConditionalBonusDamage(t *testing.T) {
	game, p1, p2 := setupSoulSorcererActionTurn(t)
	p1.Tokens["ss_yellow_soul"] = 3
	p1.ExclusiveCards = append(p1.ExclusiveCards, soulSorcererExclusiveCard(p1.Character.Name, "灵魂震爆"))
	p2.MaxHand = 6
	p2.Hand = []model.Card{
		soulSorcererTestCard("t1", "少牌1", model.CardTypeAttack, model.ElementFire),
		soulSorcererTestCard("t2", "少牌2", model.CardTypeAttack, model.ElementWater),
	}

	if err := game.UseSkill("p1", "ss_soul_blast", []string{"p2"}, nil); err != nil {
		t.Fatalf("use ss_soul_blast failed: %v", err)
	}
	if got := p1.Tokens["ss_yellow_soul"]; got != 0 {
		t.Fatalf("expected yellow soul -3 to 0, got %d", got)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending damage from soul blast")
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.TargetID != "p2" || pd.DamageType != "magic" || pd.Damage != 5 {
		t.Fatalf("unexpected soul blast pending damage: %+v", pd)
	}
}

func TestSoulSorcererSoulGrant_RespectsEnergyCap(t *testing.T) {
	game, p1, p2 := setupSoulSorcererActionTurn(t)
	p1.Tokens["ss_blue_soul"] = 3
	p1.ExclusiveCards = append(p1.ExclusiveCards, soulSorcererExclusiveCard(p1.Character.Name, "灵魂赐予"))
	p2.Gem = 2
	p2.Crystal = 0

	if err := game.UseSkill("p1", "ss_soul_grant", []string{"p2"}, nil); err != nil {
		t.Fatalf("use ss_soul_grant failed: %v", err)
	}

	if got := p1.Tokens["ss_blue_soul"]; got != 0 {
		t.Fatalf("expected blue soul -3 to 0, got %d", got)
	}
	if got := p2.Gem + p2.Crystal; got != 3 {
		t.Fatalf("target energy should cap at 3, got %d", got)
	}
}

func TestSoulSorcererSoulLink_TransferDamageBeforeResolve(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AllyA", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "AllyB", "saintess", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["ss_blue_soul"] = 2
	p1.Tokens["ss_yellow_soul"] = 1
	p1.ExclusiveCards = append(p1.ExclusiveCards, soulSorcererExclusiveCard(p1.Character.Name, "灵魂链接"))
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	if err := game.UseSkill("p1", "ss_soul_link", nil, nil); err != nil {
		t.Fatalf("use ss_soul_link failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "ss_link_target")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose soul link target failed: %v", err)
	}

	linkedHolder, _ := game.findSoulLink(p1)
	if linkedHolder == nil {
		t.Fatalf("expected soul link placed on ally")
	}

	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   "p2",
			TargetID:   linkedHolder.ID,
			Damage:     2,
			DamageType: "Attack",
			Stage:      0,
		},
	}
	if !game.maybeTriggerSoulLinkTransfer(&game.State.PendingDamageQueue[0]) {
		t.Fatalf("expected soul link transfer prompt")
	}
	requireChoicePrompt(t, game, "p1", "ss_link_transfer_x")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose transfer x failed: %v", err)
	}

	if got := p1.Tokens["ss_blue_soul"]; got != 0 {
		t.Fatalf("expected blue soul consumed by transfer, got %d", got)
	}
	if got := game.State.PendingDamageQueue[0].Damage; got != 1 {
		t.Fatalf("expected original damage reduced to 1, got %d", got)
	}
	foundTransfer := false
	for _, pd := range game.State.PendingDamageQueue {
		if pd.TargetID == "p1" && pd.DamageType == "magic" && pd.Damage == 1 && pd.FromSoulLink {
			foundTransfer = true
			break
		}
	}
	if !foundTransfer {
		t.Fatalf("expected transferred magic damage to soul sorcerer")
	}
}

func TestSoulSorcererSoulLink_Replay_TransferSorcererToAlly_NoRecursiveLinkPrompt(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AllyA", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "AllyB", "saintess", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["ss_blue_soul"] = 3
	p1.Tokens["ss_yellow_soul"] = 1
	p1.ExclusiveCards = append(p1.ExclusiveCards, soulSorcererExclusiveCard(p1.Character.Name, "灵魂链接"))
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	// 1) 启动阶段先放置灵魂链接给队友。
	if err := game.UseSkill("p1", "ss_soul_link", nil, nil); err != nil {
		t.Fatalf("use ss_soul_link failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "ss_link_target")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose soul link target failed: %v", err)
	}
	linkedHolder, _ := game.findSoulLink(p1)
	if linkedHolder == nil {
		t.Fatalf("expected soul link placed on ally")
	}

	// 2) 灵魂术士本人承伤，进入“承伤前”灵魂链接转伤选择。
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   "p2",
			TargetID:   "p1",
			Damage:     2,
			DamageType: "Attack",
			Stage:      0,
		},
	}
	if interrupted := game.processPendingDamages(); !interrupted {
		t.Fatalf("expected ss_link_transfer_x interrupt")
	}
	requireChoicePrompt(t, game, "p1", "ss_link_transfer_x")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose transfer x failed: %v", err)
	}

	if got := p1.Tokens["ss_blue_soul"]; got != 1 {
		t.Fatalf("expected blue soul left 1 after using link and transfer, got %d", got)
	}
	if len(game.State.PendingDamageQueue) < 2 {
		t.Fatalf("expected original + transferred pending damage, got %d", len(game.State.PendingDamageQueue))
	}
	original := game.State.PendingDamageQueue[0]
	if original.TargetID != "p1" || original.Damage != 1 {
		t.Fatalf("expected original damage reduced to 1 on p1, got %+v", original)
	}
	transferred := game.State.PendingDamageQueue[1]
	if transferred.TargetID != linkedHolder.ID || transferred.DamageType != "magic" || transferred.Damage != 1 || !transferred.FromSoulLink {
		t.Fatalf("unexpected transferred damage: %+v", transferred)
	}

	// 转移出来的伤害不应再次触发灵魂链接（防递归）。
	if game.maybeTriggerSoulLinkTransfer(&game.State.PendingDamageQueue[1]) {
		t.Fatalf("transferred damage should not retrigger soul link transfer")
	}
}

func TestSoulSorcererSoulLink_Replay_TransferAllyToSorcerer_NoRecursiveLinkPrompt(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AllyA", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "AllyB", "saintess", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["ss_blue_soul"] = 3
	p1.Tokens["ss_yellow_soul"] = 1
	p1.ExclusiveCards = append(p1.ExclusiveCards, soulSorcererExclusiveCard(p1.Character.Name, "灵魂链接"))
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	if err := game.UseSkill("p1", "ss_soul_link", nil, nil); err != nil {
		t.Fatalf("use ss_soul_link failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "ss_link_target")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose soul link target failed: %v", err)
	}
	linkedHolder, _ := game.findSoulLink(p1)
	if linkedHolder == nil {
		t.Fatalf("expected soul link placed on ally")
	}

	// 队友承伤 -> 转移给灵魂术士。
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   "p2",
			TargetID:   linkedHolder.ID,
			Damage:     2,
			DamageType: "Attack",
			Stage:      0,
		},
	}
	if interrupted := game.processPendingDamages(); !interrupted {
		t.Fatalf("expected ss_link_transfer_x interrupt")
	}
	requireChoicePrompt(t, game, "p1", "ss_link_transfer_x")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose transfer x failed: %v", err)
	}

	if got := p1.Tokens["ss_blue_soul"]; got != 1 {
		t.Fatalf("expected blue soul left 1 after using link and transfer, got %d", got)
	}
	if len(game.State.PendingDamageQueue) < 2 {
		t.Fatalf("expected original + transferred pending damage, got %d", len(game.State.PendingDamageQueue))
	}
	original := game.State.PendingDamageQueue[0]
	if original.TargetID != linkedHolder.ID || original.Damage != 1 {
		t.Fatalf("expected original damage reduced to 1 on ally, got %+v", original)
	}
	transferred := game.State.PendingDamageQueue[1]
	if transferred.TargetID != "p1" || transferred.DamageType != "magic" || transferred.Damage != 1 || !transferred.FromSoulLink {
		t.Fatalf("unexpected transferred damage: %+v", transferred)
	}
	if game.maybeTriggerSoulLinkTransfer(&game.State.PendingDamageQueue[1]) {
		t.Fatalf("transferred damage should not retrigger soul link transfer")
	}
}

func TestSoulSorcererSoulLink_Replay_TransferDamageThenTriggersResponseChain(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "HeroAlly", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "AllyB", "saintess", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["ss_blue_soul"] = 3
	p1.Tokens["ss_yellow_soul"] = 1
	p1.ExclusiveCards = append(p1.ExclusiveCards, soulSorcererExclusiveCard(p1.Character.Name, "灵魂链接"))
	p3.Gem = 1 // 让死斗可触发
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	// 放置链接给勇者队友。
	if err := game.UseSkill("p1", "ss_soul_link", nil, nil); err != nil {
		t.Fatalf("use ss_soul_link failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "ss_link_target")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose soul link target failed: %v", err)
	}
	linkedHolder, _ := game.findSoulLink(p1)
	if linkedHolder == nil || linkedHolder.ID != "p3" {
		t.Fatalf("expected soul link on hero ally p3, got %+v", linkedHolder)
	}

	// 灵魂术士承伤 -> 转伤给勇者队友（法伤） -> 触发勇者死斗响应链。
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   "p2",
			TargetID:   "p1",
			Damage:     2,
			DamageType: "Attack",
			Stage:      0,
		},
	}
	if interrupted := game.processPendingDamages(); !interrupted {
		t.Fatalf("expected ss_link_transfer_x interrupt")
	}
	requireChoicePrompt(t, game, "p1", "ss_link_transfer_x")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose transfer x failed: %v", err)
	}

	if interrupted := game.processPendingDamages(); !interrupted {
		t.Fatalf("expected response-skill interrupt on transferred magic damage")
	}
	requireResponseSkillPrompt(t, game, "p3")
	chooseResponseSkillByID(t, game, "p3", "hero_dead_duel")
	if got := p3.Gem; got != 0 {
		t.Fatalf("expected hero ally consumed 1 gem by dead duel, got %d", got)
	}
	if got := p3.Tokens["hero_anger"]; got != 3 {
		t.Fatalf("expected hero ally anger +3 by dead duel, got %d", got)
	}
}
