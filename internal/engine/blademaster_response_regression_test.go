package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

type noopObserver struct{}

func (noopObserver) OnGameEvent(event model.GameEvent) {}

// 回归测试：剑影在同回合后续攻击结束时应继续询问（直到本回合真正发动）
func TestBladeMaster_SwordShadow_ReAskOnEachAttackEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p2.Heal = 0

	// 两张非风系攻击牌（避免风怒追击干扰）；有蓝水晶满足剑影条件
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p1.Crystal = 1
	// 人为放一个额外攻击 token，确保同回合发生第二次攻击行动
	model.AppendAttackAction(p1, "test-token")

	// 第一次攻击
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("first attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("first take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after first attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "sword_shadow" {
		t.Fatalf("expected only sword_shadow after first attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	// 第一次不发动（跳过）
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdCancel,
	}); err != nil {
		t.Fatalf("skip first response failed: %v", err)
	}

	// 推进到第二次攻击行动
	game.Drive()

	// 第二次攻击
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("second attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("second take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt again after second attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "sword_shadow" {
		t.Fatalf("expected only sword_shadow after second attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
}

// 回归测试：风怒追击在同回合后续攻击结束时应继续询问（直到本回合真正发动）
func TestBladeMaster_WindFury_ReAskOnEachAttackEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// 三张风系攻击牌：前两张用于两次攻击，保留一张确保第二次攻击结束时仍满足风怒可用条件
	p1.Hand = []model.Card{
		{ID: "a1", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
		{ID: "a2", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
		{ID: "a3", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}
	p1.Gem = 0
	p1.Crystal = 0
	model.AppendAttackAction(p1, "test-token")

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("first attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("first take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after first attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after first attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdCancel,
	}); err != nil {
		t.Fatalf("skip first response failed: %v", err)
	}

	game.Drive()

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("second attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("second take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt again after second attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after second attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
}

func TestBladeMaster_WindFury_StillRunsWithoutRemainingWindAttack(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// 只有这一张风系攻击牌；按文档，风怒追击仍应在攻击结束时触发，
	// 随后若没有合法风系攻击可执行，可在额外行动阶段选择“无法行动”跳过。
	p1.Hand = []model.Card{
		{ID: "a1", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("wind attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after attack end, got %+v", game.State.PendingInterrupt)
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.ConfirmResponseSkill("p1", "wind_fury"); err != nil {
		t.Fatalf("confirm wind_fury failed: %v", err)
	}

	game.Drive()
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected extra action to enter action execution window, got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected extra attack constraint, got %q", p1.TurnState.CurrentExtraAction)
	}
	if p1.TurnState.CurrentExtraElement == nil || len(p1.TurnState.CurrentExtraElement) != 1 || p1.TurnState.CurrentExtraElement[0] != model.ElementWind {
		t.Fatalf("expected wind-only extra attack constraint, got %+v", p1.TurnState.CurrentExtraElement)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCannotAct,
	}); err != nil {
		t.Fatalf("expected cannot_act to skip unusable wind-only extra action, got %v", err)
	}
	if p1.TurnState.CurrentExtraAction != "" {
		t.Fatalf("expected extra action constraint cleared after cannot_act, got %q", p1.TurnState.CurrentExtraAction)
	}
}

// 回归测试：若疾风技已提供额外攻击行动，取消风怒追击不应吞掉这次额外行动。
func TestBladeMaster_GaleSkillExtraAction_PreservedWhenWindFuryCanceled(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// 第一张为带“疾风技”的独有攻击牌；第二张用于验证额外行动仍可执行。
	p1.Hand = []model.Card{
		{
			ID:              "gale-attack",
			Name:            "疾风斩",
			Type:            model.CardTypeAttack,
			Element:         model.ElementWind,
			Damage:          1,
			ExclusiveChar1:  "blade_master",
			ExclusiveSkill1: "疾风技",
		},
		{
			ID:      "follow-attack",
			Name:    "火斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  1,
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("gale attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after attack end, got %+v", game.State.PendingInterrupt)
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	if len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected gale_skill to append extra attack token before wind_fury response")
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdCancel,
	}); err != nil {
		t.Fatalf("cancel wind_fury failed: %v", err)
	}

	// HandleAction 会自动 Drive；取消后应继续保留疾风技带来的额外行动窗口。
	if game.State.CurrentTurn != 0 {
		t.Fatalf("expected turn to stay on p1 after cancel, got turn=%d", game.State.CurrentTurn)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected to resume action execution for extra action, got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected current extra action Attack after cancel, got %q", p1.TurnState.CurrentExtraAction)
	}
}

// 回归测试：攻击被圣盾抵挡（未命中）时，取消风怒追击也不应吞掉疾风技额外行动。
func TestBladeMaster_GaleSkillExtraAction_PreservedWhenWindFuryCanceled_AfterShieldMiss(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0
	p2.AddFieldCard(&model.FieldCard{
		Mode:   model.FieldEffect,
		Effect: model.EffectShield,
		Card: model.Card{
			ID:      "shield-card",
			Name:    "圣盾",
			Type:    model.CardTypeMagic,
			Element: model.ElementLight,
		},
	})

	p1.Hand = []model.Card{
		{
			ID:              "gale-attack",
			Name:            "疾风斩",
			Type:            model.CardTypeAttack,
			Element:         model.ElementWind,
			Damage:          1,
			ExclusiveChar1:  "blade_master",
			ExclusiveSkill1: "疾风技",
		},
		{
			ID:      "follow-attack",
			Name:    "火斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  1,
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("gale attack failed: %v", err)
	}
	// 选择承受，随后由圣盾自动抵挡并走未命中分支。
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after shield miss action end, got %+v", game.State.PendingInterrupt)
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after shield miss action end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	if len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected gale_skill extra attack token before cancel")
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdCancel,
	}); err != nil {
		t.Fatalf("cancel wind_fury failed: %v", err)
	}

	if game.State.CurrentTurn != 0 {
		t.Fatalf("expected turn to stay on p1 after cancel, got turn=%d", game.State.CurrentTurn)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected to resume action execution for extra action, got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected current extra action Attack after cancel, got %q", p1.TurnState.CurrentExtraAction)
	}
}

// 回归测试：若历史分支遗留 LastActionType，取消风怒追击后不应再次重入同一 ActionEnd 响应窗口。
func TestBladeMaster_WindFuryCancel_WithStaleLastActionType_NoRepromptAndKeepExtraAction(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	// 模拟疾风技已提供额外攻击行动。
	model.AppendAttackAction(p1, "疾风技")
	// 模拟某旧分支遗留了 LastActionType，导致状态机会尝试 ActionEnd 补结算。
	p1.TurnState.LastActionType = string(model.ActionAttack)

	eventCtx := &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   "p1",
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	}
	userCtx := game.buildContext(p1, nil, model.TimingOnActionEnd, eventCtx)
	game.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: "p1",
		SkillIDs: []string{"wind_fury"},
		Context:  userCtx,
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	}); err != nil {
		t.Fatalf("cancel wind_fury failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no reprompt after cancel, got %+v", game.State.PendingInterrupt)
	}
	if game.State.CurrentTurn != 0 {
		t.Fatalf("expected turn stay on p1, got %d", game.State.CurrentTurn)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected resume to action execution(extra action), got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected current extra action Attack, got %q", p1.TurnState.CurrentExtraAction)
	}
}

// 回归测试：弃牌型恢复链路若落在 ActionEnd，上下文恢复应清理 LastActionType，避免同轮重复 ActionEnd 补结算。
func TestBladeMaster_DiscardContextResume_OnActionEnd_ClearsLastActionCatchup(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "wind-1", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}
	p1.TurnState.LastActionType = string(model.ActionAttack)
	p1.TurnState.LastActionCard = &model.Card{ID: "attack-ended", Name: "已结束攻击", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1}

	ctx := game.buildContext(p1, nil, model.TimingOnActionEnd, &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   p1.ID,
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	})
	ctx.Selections["response_resume_phase"] = model.TurnStageActionEnd

	if ok := game.resumePhaseAfterSkillDiscardContext(ctx); !ok {
		t.Fatalf("expected discard-context resume handled")
	}
	if got := p1.TurnState.LastActionType; got != "" {
		t.Fatalf("expected LastActionType cleared during resume, got %q", got)
	}
	if p1.TurnState.LastActionCard != nil {
		t.Fatalf("expected LastActionCard cleared during resume")
	}

	game.Drive()
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no duplicated action-end response reprompt, got %+v", game.State.PendingInterrupt)
	}
}

// 回归测试：多个可响应技能时，确认其中一个后应先结算并弹出当前响应中断，
// 再继续后续中断/剩余响应技能，而不是停留在同一个响应中断里原地重选。
func TestBladeMaster_MultiResponse_ConfirmOneSettlesBeforeRemaining(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.Crystal = 1
	p1.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionEnd
	game.State.Subflow = model.SubflowResponse

	ctx := game.buildContext(p1, nil, model.TimingOnActionEnd, &model.EventContext{
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			IsHit:            true,
			CounterInitiator: "",
		},
	})

	game.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: "p1",
		SkillIDs: []string{"sword_shadow", "wind_fury"},
		Context:  ctx,
	}
	game.State.InterruptQueue = []*model.Interrupt{
		{
			Type:     model.InterruptChoice,
			PlayerID: "p1",
			Context: map[string]interface{}{
				"choice_type": "test_followup_choice",
			},
		},
	}

	if err := game.ConfirmResponseSkill("p1", "sword_shadow"); err != nil {
		t.Fatalf("confirm sword_shadow failed: %v", err)
	}

	if got := p1.TurnState.UsedSkillCounts["sword_shadow"]; got != 1 {
		t.Fatalf("expected sword_shadow used count=1, got %d", got)
	}
	if p1.Crystal != 0 {
		t.Fatalf("expected sword_shadow to consume 1 crystal, got %d", p1.Crystal)
	}
	if len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected sword_shadow to append an extra attack action")
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected queued follow-up interrupt to become pending first, got %+v", game.State.PendingInterrupt)
	}
	if len(game.State.InterruptQueue) != 1 {
		t.Fatalf("expected exactly one queued interrupt (remaining response), got %d", len(game.State.InterruptQueue))
	}
	next := game.State.InterruptQueue[0]
	if next == nil || next.Type != model.InterruptResponseSkill || next.PlayerID != "p1" {
		t.Fatalf("expected queued remaining response interrupt for p1, got %+v", next)
	}
	if len(next.SkillIDs) != 1 || next.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected remaining response skill to be only wind_fury, got %+v", next.SkillIDs)
	}
}

// 集成回归：真实走一轮“攻击命中后双响应 -> 先选剑影 -> 再问风怒追击 -> 再选风怒追击”。
func TestBladeMaster_ResponseChain_SwordShadowThenWindFury_Integration(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p2.Heal = 0
	p1.Hand = []model.Card{
		{ID: "atk-1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response-skill prompt after attack end, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected response prompt for p1, got %s", game.State.PendingInterrupt.PlayerID)
	}

	hasSwordShadow := false
	hasWindFury := false
	for _, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == "sword_shadow" {
			hasSwordShadow = true
		}
		if sid == "wind_fury" {
			hasWindFury = true
		}
	}
	if !hasSwordShadow || !hasWindFury {
		t.Fatalf("expected both sword_shadow and wind_fury in response skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	swordShadowIdx := -1
	for i, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == "sword_shadow" {
			swordShadowIdx = i
			break
		}
	}
	if swordShadowIdx < 0 {
		t.Fatalf("sword_shadow not found in response options: %+v", game.State.PendingInterrupt.SkillIDs)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{swordShadowIdx},
	}); err != nil {
		t.Fatalf("select sword_shadow failed: %v", err)
	}

	if got := p1.TurnState.UsedSkillCounts["sword_shadow"]; got != 1 {
		t.Fatalf("expected sword_shadow used count=1, got %d", got)
	}
	if p1.Crystal != 0 {
		t.Fatalf("expected sword_shadow to consume crystal, got %d", p1.Crystal)
	}
	if len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected sword_shadow to append an extra action token")
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected reprompt for remaining response skill, got %+v", game.State.PendingInterrupt)
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after sword_shadow resolves, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("select wind_fury failed: %v", err)
	}

	if got := p1.TurnState.UsedSkillCounts["wind_fury"]; got != 1 {
		t.Fatalf("expected wind_fury used count=1, got %d", got)
	}
	if intr := game.State.PendingInterrupt; intr != nil && intr.Type == model.InterruptResponseSkill {
		for _, sid := range intr.SkillIDs {
			if sid == "wind_fury" {
				t.Fatalf("wind_fury should not be reprompted after it is confirmed")
			}
		}
	}
	extraActionTotal := len(p1.TurnState.PendingActions)
	if p1.TurnState.CurrentExtraAction != "" {
		extraActionTotal++
	}
	if extraActionTotal < 2 {
		t.Fatalf("expected at least 2 extra attack opportunities from sword_shadow + wind_fury, got %d", extraActionTotal)
	}
}

func TestBladeMaster_HolySwordDraw_X0ResumesExtraAction(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.AttackCount = 3
	game.State.TurnStage = model.TurnStageActionExecution

	if !game.holySwordDrawInterruptIfNeeded(p1) {
		t.Fatalf("expected holy sword draw interrupt to dispatch")
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptHolySwordDraw {
		t.Fatalf("expected holy sword draw interrupt, got %+v", game.State.PendingInterrupt)
	}

	if err := game.handleInterruptAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("holy sword x=0 response failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after holy sword x=0, got %+v", game.State.PendingInterrupt)
	}
	if game.State.TurnStage != model.TurnStageExtraAction {
		t.Fatalf("expected holy sword x=0 to resume extra action, got %s", game.State.TurnStage)
	}
}

func TestBladeMaster_HolySwordDiscardResumesExtraAction(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.AttackCount = 3
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	model.AppendAttackAction(p1, "holy-sword-followup")
	game.State.CurrentTurn = 0

	if !game.holySwordDrawInterruptIfNeeded(p1) {
		t.Fatalf("expected holy sword draw interrupt to dispatch")
	}
	if err := game.handleInterruptAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("holy sword x=1 response failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || !isDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt after holy sword x=1, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("holy sword discard confirmation failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after holy sword discard, got %+v", game.State.PendingInterrupt)
	}
	if game.State.CurrentTurn != 0 {
		t.Fatalf("expected holy sword aftermath to keep current turn when extra action remains, got turn=%d", game.State.CurrentTurn)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected holy sword discard to continue into extra action execution window, got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected extra attack constraint to be restored, got %q", p1.TurnState.CurrentExtraAction)
	}
	if len(p1.TurnState.PendingActions) != 0 {
		t.Fatalf("expected extra action token to be consumed into current extra action, got %d pending", len(p1.TurnState.PendingActions))
	}
}

func TestBladeMaster_GaleSlash_DisablesCounterButAllowsDefend(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0
	p2.Hand = []model.Card{
		{ID: "holy-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "counter-wind", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}
	p2.AddFieldCard(&model.FieldCard{
		Mode:   model.FieldEffect,
		Effect: model.EffectShield,
		Card:   model.Card{ID: "shield", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
	})
	p1.Hand = []model.Card{{
		ID:              "gale-slash-card",
		Name:            "列风技",
		Type:            model.CardTypeAttack,
		Element:         model.ElementWind,
		Damage:          2,
		ExclusiveChar1:  "blade_master",
		ExclusiveSkill1: "列风技",
	}}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("gale slash attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardIndex: 1, TargetID: "p3",
	}); err == nil {
		t.Fatalf("expected gale slash to forbid counter response")
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"defend"}, CardIndex: 0,
	}); err != nil {
		t.Fatalf("expected gale slash to still allow defend: %v", err)
	}

	if !p2.HasFieldEffect(model.EffectShield) {
		t.Fatalf("expected shield to remain because gale slash only ignores it on take")
	}
	if len(p2.Hand) != 1 || p2.Hand[0].Name != "风斩" {
		t.Fatalf("expected only holy light to be consumed on defend, remaining hand=%+v", p2.Hand)
	}
	if p2.Heal != 0 {
		t.Fatalf("expected defend to prevent damage draw, got heal=%d", p2.Heal)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected defend path to resolve cleanly, got pending interrupt %+v", game.State.PendingInterrupt)
	}
}
