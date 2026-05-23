// gameflow: 魔法少女中断 prompt 与 action 处理。

package magical_girl

import (
	"fmt"
	"strconv"

	"starcup-engine/internal/engine/hook/promptfmt"
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const (
	magicBlastStageTargetDiscard       = "target_discard"
	magicBlastStageCasterForcedDiscard = "caster_forced_discard"
)

// --- Prompt builders ---

func buildMagicMissilePrompt(rt player.ChoiceRuntime) *model.Prompt {
	chain := rt.GetMagicBulletChain()
	if chain == nil {
		return nil
	}

	playerID := chain.TargetID
	p := rt.GetPlayers()[playerID]
	if p == nil {
		return nil
	}

	damage := chain.CurrentDamage
	hasShield := p.HasFieldEffect(model.EffectShield)
	takeLabel := "承受伤害"
	if hasShield {
		takeLabel = "承受（将触发圣盾）"
	}
	effectHints := []string{}
	if hasShield {
		effectHints = append(effectHints, "你身上有【圣盾】：可先应战/防御；若本次选择承受伤害，将自动消耗圣盾抵挡魔弹。")
	}

	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		AttackerID: chain.SourcePlayerID,
		Message:    fmt.Sprintf("你成为了【魔弹】的目标，当前伤害为 %d，请选择应对：", damage),
		Options: []model.PromptOption{
			{ID: "take", Label: takeLabel},
			{ID: "counter", Label: "打出【魔弹】传递"},
			{ID: "defend", Label: "使用【圣光】抵挡"},
		},
		EffectHints:  effectHints,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationResponse, Layout: "magic_missile"},
	}
}

func buildMagicBulletDirectionPrompt(rt player.ChoiceRuntime) *model.Prompt {
	interrupt := rt.GetPendingInterrupt()
	if interrupt == nil {
		return nil
	}
	playerID := interrupt.PlayerID

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【魔弹掌控】选择魔弹传递方向：",
		Options: []model.PromptOption{
			{ID: "normal", Label: "默认方向 (右手边，前一位对手)"},
			{ID: "reverse", Label: "逆向传递 (左手边，后一位对手)"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
	}
}

func buildMagicBlastPrompt(rt player.ChoiceRuntime) *model.Prompt {
	interrupt := rt.GetPendingInterrupt()
	if interrupt == nil {
		return nil
	}

	playerID := interrupt.PlayerID
	p := rt.GetPlayers()[playerID]
	if p == nil {
		return nil
	}
	data, _ := interrupt.Context.(map[string]interface{})
	stage := magicBlastStageFromContext(data)

	if stage == magicBlastStageCasterForcedDiscard {
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【魔爆冲击】请选择弃1张牌：",
			Options:      magicBlastCasterForcedDiscardOptions(p),
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}
	}

	return &model.Prompt{
		Type:         model.PromptChooseCards,
		PlayerID:     playerID,
		Message:      "【魔爆冲击】请选择弃一张法术牌，否则受到2点伤害：",
		Options:      magicBlastTargetDiscardOptions(p),
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
	}
}

// --- MagicBlast prompt helpers ---

func magicBlastStageFromContext(data map[string]interface{}) string {
	stage, _ := data["stage"].(string)
	if stage == "" {
		return magicBlastStageTargetDiscard
	}
	return stage
}

func magicBlastCasterForcedDiscardOptions(p *model.Player) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(p.Hand))
	for i, card := range p.Hand {
		options = append(options, model.PromptOption{
			ID:     strconv.Itoa(i),
			Label:  fmt.Sprintf("%d: %s", i+1, promptfmt.FormatCardInfo(card)),
			CardID: card.ID,
		})
	}
	return options
}

func magicBlastTargetDiscardOptions(p *model.Player) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(p.Hand)+1)
	for i, card := range p.Hand {
		if card.Type != model.CardTypeMagic {
			continue
		}
		options = append(options, model.PromptOption{
			ID:     strconv.Itoa(i),
			Label:  fmt.Sprintf("%d: %s", i+1, promptfmt.FormatCardInfo(card)),
			CardID: card.ID,
		})
	}
	options = append(options, model.PromptOption{
		ID:    "refuse",
		Label: "不弃牌 (受到2点伤害)",
	})
	return options
}

// --- Action handlers ---

func handleMagicMissileAction(rt player.ChoiceRuntime, act model.PlayerAction) (player.InterruptActionResult, error) {
	return resultAfterMagicGirlInterrupt(rt, func() error {
		return resolveMagicMissileAction(rt, act)
	})
}

func resolveMagicMissileAction(rt player.ChoiceRuntime, act model.PlayerAction) error {
	chain := rt.GetMagicBulletChain()
	if chain == nil {
		return fmt.Errorf("魔弹链条不存在")
	}
	if act.PlayerID != chain.TargetID {
		return fmt.Errorf("不是你的响应回合")
	}

	respType := ""
	if len(act.ExtraArgs) > 0 {
		respType = act.ExtraArgs[0]
	} else {
		return fmt.Errorf("缺少响应类型")
	}

	p := rt.GetPlayers()[act.PlayerID]
	if p == nil {
		return fmt.Errorf("玩家不存在")
	}

	switch respType {
	case "take":
		return resolveMagicMissileTake(rt, p, chain)
	case "counter":
		return resolveMagicMissileCounter(rt, p, chain, act)
	case "defend":
		return resolveMagicMissileDefend(rt, p, chain, act)
	default:
		return fmt.Errorf("未知的响应类型: %s", respType)
	}
}

func resolveMagicMissileTake(rt player.ChoiceRuntime, p *model.Player, chain *model.MagicBulletChain) error {
	// 尝试自动消耗圣盾
	if p.HasFieldEffect(model.EffectShield) {
		removed := false
		for _, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectShield {
				continue
			}
			p.RemoveFieldCard(fc)
			rt.AppendToDiscard([]model.Card{fc.Card})
			removed = true
			break
		}
		if removed {
			rt.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】触发，自动抵挡了魔弹", p.Name))
			rt.Log(fmt.Sprintf("[Magic] %s 选择承受，触发【圣盾】自动抵挡魔弹", p.Name))
			rt.SetMagicBulletChain(nil)
			rt.PopInterrupt()
			return nil
		}
	}

	damage := chain.CurrentDamage
	rt.Log(fmt.Sprintf("[Magic] %s 选择承受魔弹伤害 (%d点)", p.Name, damage))
	magicCard := &model.Card{
		Name:        "魔弹",
		Type:        model.CardTypeMagic,
		Damage:      damage,
		Description: "魔弹伤害",
	}

	rt.PopInterrupt()
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   chain.SourcePlayerID,
		TargetID:   p.ID,
		Damage:     damage,
		DamageType: model.MagicDamage,
		Card:       magicCard,
	})
	rt.EnterDamageResolution(nil)
	rt.SetMagicBulletChain(nil)
	return nil
}

func resolveMagicMissileCounter(rt player.ChoiceRuntime, p *model.Player, chain *model.MagicBulletChain, act model.PlayerAction) error {
	card, cardOK := playableCardForAction(rt, p, act)
	if !cardOK {
		return fmt.Errorf("无效的卡牌ID")
	}
	if err := rt.DispatchHitCheckMagicMissileCounter(p, chain, &card); err != nil {
		return err
	}
	if card.Name != "魔弹" {
		return fmt.Errorf("必须使用【魔弹】进行传递")
	}

	hasParticipated := false
	for _, pid := range chain.InvolvedIDs {
		if pid == p.ID {
			hasParticipated = true
			break
		}
	}
	if hasParticipated {
		return fmt.Errorf("你在本轮传递中已参与过，无法再次传递")
	}

	consumed, err := consumePlayableCardForAction(rt, p, act)
	if err != nil {
		return err
	}
	rt.AppendToDiscard([]model.Card{consumed})
	rt.Log(fmt.Sprintf("[Magic] %s 打出魔弹，将伤害传递给下一位！伤害+1", p.Name))

	return passMagicMissileToNext(rt, p, chain)
}

func resolveMagicMissileDefend(rt player.ChoiceRuntime, p *model.Player, chain *model.MagicBulletChain, act model.PlayerAction) error {
	if err := rt.DispatchHitCheckMagicMissileDefend(p, chain); err != nil {
		return err
	}
	card, cardOK := playableCardForAction(rt, p, act)
	if !cardOK {
		return fmt.Errorf("无效的卡牌ID")
	}
	if card.Name == "圣盾" {
		return fmt.Errorf("【圣盾】不能在防御时打出，请提前放置到场上触发")
	}
	if card.Name != "圣光" {
		return fmt.Errorf("必须使用【圣光】抵挡")
	}
	rt.Log(fmt.Sprintf("[Magic] %s 使用【圣光】，抵挡了魔弹", p.Name))
	consumed, err := consumePlayableCardForAction(rt, p, act)
	if err != nil {
		return err
	}
	rt.AppendToDiscard([]model.Card{consumed})

	rt.SetMagicBulletChain(nil)
	rt.PopInterrupt()
	return nil
}

func playableCardForAction(rt player.ChoiceRuntime, p *model.Player, act model.PlayerAction) (model.Card, bool) {
	if act.CardID == "" {
		return model.Card{}, false
	}
	return rt.GetPlayableCardByCardID(p, act.CardID)
}

func consumePlayableCardForAction(rt player.ChoiceRuntime, p *model.Player, act model.PlayerAction) (model.Card, error) {
	if act.CardID == "" {
		return model.Card{}, fmt.Errorf("无效的卡牌ID")
	}
	card, ok := rt.ConsumePlayableCardByCardID(p.ID, act.CardID)
	if !ok {
		return model.Card{}, fmt.Errorf("无效的卡牌ID")
	}
	return card, nil
}

func passMagicMissileToNext(rt player.ChoiceRuntime, p *model.Player, chain *model.MagicBulletChain) error {
	chain.CurrentDamage += 1
	chain.SourcePlayerID = p.ID
	chain.InvolvedIDs = append(chain.InvolvedIDs, p.ID)
	aliveCount := len(rt.GetPlayerOrder())
	if len(chain.InvolvedIDs) >= aliveCount {
		rt.Log("[Magic] 本轮魔弹传递已覆盖所有角色，魔弹结算结束")
		rt.SetMagicBulletChain(nil)
		rt.PopInterrupt()
		return nil
	}

	nextTargetID := rt.FindNextMagicBulletTarget(p.ID)
	if nextTargetID == "" {
		rt.Log("[Magic] 没有下一个目标，魔弹失效")
		rt.SetMagicBulletChain(nil)
		rt.PopInterrupt()
		return nil
	}

	nextTarget := rt.GetPlayers()[nextTargetID]
	chain.TargetID = nextTargetID
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.PlayerID = nextTargetID
		intr.Context = map[string]interface{}{
			"damage":    chain.CurrentDamage,
			"source_id": p.ID,
		}
	}
	rt.OfferMagicMissileResponseSkills()
	if intr := rt.GetPendingInterrupt(); intr == nil || intr.Type != model.InterruptMagicMissile {
		if nextTarget != nil {
			rt.Log(fmt.Sprintf("[Magic] 魔弹指向 %s (伤害: %d)，等待响应...", nextTarget.Name, chain.CurrentDamage))
		}
		return nil
	}
	rt.NotifyInterruptPrompt()
	if nextTarget != nil {
		rt.Log(fmt.Sprintf("[Magic] 魔弹指向 %s (伤害: %d)，等待响应...", nextTarget.Name, chain.CurrentDamage))
	}
	return nil
}

func handleMagicBulletDirectionAction(rt player.ChoiceRuntime, act model.PlayerAction) (player.InterruptActionResult, error) {
	return resultAfterMagicGirlInterrupt(rt, func() error {
		return resolveMagicBulletDirectionAction(rt, act)
	})
}

func resolveMagicBulletDirectionAction(rt player.ChoiceRuntime, act model.PlayerAction) error {
	interrupt := rt.GetPendingInterrupt()
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	p := rt.GetPlayers()[act.PlayerID]
	if p == nil {
		return fmt.Errorf("玩家不存在")
	}

	reverse := false
	if len(act.Selections) > 0 && act.Selections[0] == 1 {
		reverse = true
	}

	isFusion, _ := data["is_fusion"].(bool)
	var fusionCard *model.Card
	if isFusion {
		if fc, ok := data["fusion_card"].(model.Card); ok {
			fusionCard = &fc
		}
	}

	rt.PopInterrupt()
	direction := "顺时针"
	if reverse {
		direction = "逆时针"
		rt.Log(fmt.Sprintf("[Skill] %s 发动【魔弹掌控】，魔弹将%s传递！", p.Name, direction))
	}
	return rt.ExecuteMagicBullet(p, reverse, isFusion, fusionCard)
}

func handleMagicBlastAction(rt player.ChoiceRuntime, act model.PlayerAction) (player.InterruptActionResult, error) {
	return resultAfterMagicGirlInterrupt(rt, func() error {
		return resolveMagicBlastAction(rt, act)
	})
}

func resolveMagicBlastAction(rt player.ChoiceRuntime, act model.PlayerAction) error {
	interrupt := rt.GetPendingInterrupt()
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}
	if act.PlayerID != interrupt.PlayerID {
		return fmt.Errorf("不是你的响应回合")
	}

	data, casterID, targetIDs, currentTargetIdx, stage, err := parseMagicBlastInterruptContext(interrupt)
	if err != nil {
		return err
	}

	switch stage {
	case magicBlastStageTargetDiscard:
		return resolveMagicBlastTargetDiscard(rt, act, data, casterID, targetIDs, currentTargetIdx)
	case magicBlastStageCasterForcedDiscard:
		return resolveMagicBlastCasterForcedDiscard(rt, act, data, targetIDs, currentTargetIdx)
	default:
		return fmt.Errorf("未知的魔爆冲击阶段: %s", stage)
	}
}

func resultAfterMagicGirlInterrupt(rt player.ChoiceRuntime, run func() error) (player.InterruptActionResult, error) {
	before := rt.GetPendingInterrupt()
	if err := run(); err != nil {
		return player.InterruptActionResult{}, err
	}
	if before != nil && rt.GetPendingInterrupt() == before {
		return player.InterruptActionResult{}, nil
	}
	return player.InterruptActionResult{Consumed: true}, nil
}

// --- MagicBlast helpers ---

func parseMagicBlastInterruptContext(interrupt *model.Interrupt) (map[string]interface{}, string, []string, int, string, error) {
	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return nil, "", nil, 0, "", fmt.Errorf("中断上下文格式错误")
	}
	casterID, _ := data["caster_id"].(string)
	if casterID == "" {
		return nil, "", nil, 0, "", fmt.Errorf("魔爆冲击缺少施法者")
	}
	targetIDs, ok := data["targets"].([]string)
	if !ok || len(targetIDs) == 0 {
		return nil, "", nil, 0, "", fmt.Errorf("魔爆冲击缺少目标")
	}
	currentTargetIdx := 0
	if v, ok := data["current_target"].(int); ok {
		currentTargetIdx = v
	}
	if currentTargetIdx < 0 || currentTargetIdx > len(targetIDs) {
		return nil, "", nil, 0, "", fmt.Errorf("魔爆冲击目标序号无效")
	}
	stage := magicBlastStageFromContext(data)
	return data, casterID, targetIDs, currentTargetIdx, stage, nil
}

func resolveMagicBlastTargetDiscard(
	rt player.ChoiceRuntime,
	act model.PlayerAction,
	data map[string]interface{},
	casterID string,
	targetIDs []string,
	currentTargetIdx int,
) error {
	p := rt.GetPlayers()[act.PlayerID]
	if p == nil {
		return fmt.Errorf("玩家不存在")
	}
	if currentTargetIdx >= len(targetIDs) {
		return fmt.Errorf("魔爆冲击目标序号无效")
	}

	discarded, err := resolveMagicBlastTargetChoice(rt, p, act)
	if err != nil {
		return err
	}

	nextTargetIdx := currentTargetIdx + 1
	data["current_target"] = nextTargetIdx
	if discarded {
		return advanceMagicBlastToNextTarget(rt, data, targetIDs, nextTargetIdx)
	}

	rt.InflictDamage(casterID, p.ID, 2, model.MagicAttack)
	rt.Log(fmt.Sprintf("[Skill] %s 未弃法术牌，受到2点伤害", p.Name))

	caster := rt.GetPlayers()[casterID]
	if caster != nil && len(caster.Hand) > 0 {
		return enterMagicBlastCasterForcedDiscard(rt, data, casterID, nextTargetIdx)
	}
	return advanceMagicBlastToNextTarget(rt, data, targetIDs, nextTargetIdx)
}

func resolveMagicBlastTargetChoice(rt player.ChoiceRuntime, p *model.Player, act model.PlayerAction) (bool, error) {
	if act.Type == model.CmdCancel {
		return false, nil
	}
	if act.Type != model.CmdSelect || len(act.Selections) != 1 {
		return false, fmt.Errorf("请选择弃一张法术牌，或取消并承受伤害")
	}

	selection := act.Selections[0]
	return true, discardMagicBlastMagicCard(rt, p, selection)
}

func discardMagicBlastMagicCard(rt player.ChoiceRuntime, p *model.Player, cardIdx int) error {
	if cardIdx < 0 || cardIdx >= len(p.Hand) {
		return fmt.Errorf("无效的卡牌索引")
	}
	card := p.Hand[cardIdx]
	if card.Type != model.CardTypeMagic {
		return fmt.Errorf("只能弃置法术牌")
	}
	rt.NotifyCardRevealed(p.ID, []model.Card{card}, "discard")
	p.Hand = append(p.Hand[:cardIdx], p.Hand[cardIdx+1:]...)
	rt.AppendToDiscard([]model.Card{card})
	rt.Log(fmt.Sprintf("[Skill] %s 弃掉了法术牌 %s", p.Name, card.Name))
	return nil
}

func resolveMagicBlastCasterForcedDiscard(
	rt player.ChoiceRuntime,
	act model.PlayerAction,
	data map[string]interface{},
	targetIDs []string,
	currentTargetIdx int,
) error {
	p := rt.GetPlayers()[act.PlayerID]
	if p == nil {
		return fmt.Errorf("玩家不存在")
	}
	if act.Type != model.CmdSelect || len(act.Selections) != 1 {
		return fmt.Errorf("请选择1张牌弃置")
	}

	cardIdx := act.Selections[0]
	if cardIdx < 0 || cardIdx >= len(p.Hand) {
		return fmt.Errorf("无效的卡牌索引")
	}
	card := p.Hand[cardIdx]
	p.Hand = append(p.Hand[:cardIdx], p.Hand[cardIdx+1:]...)
	rt.AppendToDiscard([]model.Card{card})
	rt.Log(fmt.Sprintf("[Skill] %s 因【魔爆冲击】弃掉了 %s", p.Name, card.Name))

	return advanceMagicBlastToNextTarget(rt, data, targetIDs, currentTargetIdx)
}

func enterMagicBlastCasterForcedDiscard(rt player.ChoiceRuntime, data map[string]interface{}, casterID string, nextTargetIdx int) error {
	data["stage"] = magicBlastStageCasterForcedDiscard
	data["current_target"] = nextTargetIdx
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.PlayerID = casterID
	}
	func() {
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = data
		}
	}()
	rt.NotifyInterruptPrompt()
	return nil
}

func advanceMagicBlastToNextTarget(rt player.ChoiceRuntime, data map[string]interface{}, targetIDs []string, nextTargetIdx int) error {
	if nextTargetIdx >= len(targetIDs) {
		rt.PopInterrupt()
		return nil
	}
	data["stage"] = magicBlastStageTargetDiscard
	nextTargetID := targetIDs[nextTargetIdx]
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.PlayerID = nextTargetID
	}
	func() {
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = data
		}
	}()
	if nextTarget := rt.GetPlayers()[nextTargetID]; nextTarget != nil {
		rt.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", nextTarget.Name))
	}
	rt.NotifyInterruptPrompt()
	return nil
}
