// gameflow: 法术爆发类响应。

package engine

import (
	"fmt"
	"strconv"

	"starcup-engine/internal/model"
)

// handleMagicBlastResponse 处理魔爆冲击弃牌响应。
func (e *GameEngine) handleMagicBlastResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}
	if act.PlayerID != interrupt.PlayerID {
		return fmt.Errorf("不是你的响应回合")
	}

	// 流程入口：从中断上下文读取当前阶段、施法者和目标序号。
	data, casterID, targetIDs, currentTargetIdx, stage, err := parseMagicBlastInterruptContext(interrupt)
	if err != nil {
		return err
	}

	switch stage {
	case magicBlastStageTargetDiscard:
		// 阶段1：当前目标先决策“弃法术牌”或“吃2点法术伤害”。
		return e.resolveMagicBlastTargetDiscard(act, data, casterID, targetIDs, currentTargetIdx)
	case magicBlastStageCasterForcedDiscard:
		// 阶段2：若目标未弃牌，施法者额外弃1张牌，然后回到下一个目标。
		return e.resolveMagicBlastCasterForcedDiscard(act, data, targetIDs, currentTargetIdx)
	default:
		return fmt.Errorf("未知的魔爆冲击阶段: %s", stage)
	}
}

func parseMagicBlastInterruptContext(interrupt *model.Interrupt) (map[string]interface{}, string, []string, int, string, error) {
	// 中断上下文是魔爆链路的单一真相源：
	// caster_id = 施法者，targets = 目标顺序，current_target = 当前处理到的目标下标。
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

func (e *GameEngine) resolveMagicBlastTargetDiscard(
	act model.PlayerAction,
	data map[string]interface{},
	casterID string,
	targetIDs []string,
	currentTargetIdx int,
) error {
	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}
	if currentTargetIdx >= len(targetIDs) {
		return fmt.Errorf("魔爆冲击目标序号无效")
	}

	discarded, err := e.resolveMagicBlastTargetChoice(player, act)
	if err != nil {
		return err
	}

	nextTargetIdx := currentTargetIdx + 1
	data["current_target"] = nextTargetIdx
	if discarded {
		return e.advanceMagicBlastToNextTarget(data, targetIDs, nextTargetIdx)
	}

	// 规则顺序：目标不弃法术牌 -> 先受到2点法术伤害 -> 再判断施法者是否强制弃牌。
	e.InflictDamage(casterID, player.ID, 2, model.MagicAttack)
	e.Log(fmt.Sprintf("[Skill] %s 未弃法术牌，受到2点伤害", player.Name))

	caster := e.State.Players[casterID]
	if caster != nil && len(caster.Hand) > 0 {
		return e.enterMagicBlastCasterForcedDiscard(data, casterID, nextTargetIdx)
	}
	return e.advanceMagicBlastToNextTarget(data, targetIDs, nextTargetIdx)
}

func (e *GameEngine) resolveMagicBlastTargetChoice(player *model.Player, act model.PlayerAction) (bool, error) {
	// 目标响应只允许两类输入：
	// 1) CmdSelect 选择一张法术牌弃置；2) CmdCancel / 选择 refuse 代表不弃牌并承受伤害。
	if act.Type == model.CmdCancel {
		return false, nil
	}
	if act.Type != model.CmdSelect || len(act.Selections) != 1 {
		return false, fmt.Errorf("请选择弃一张法术牌，或取消并承受伤害")
	}

	options := magicBlastTargetDiscardOptions(player)
	selection := act.Selections[0]
	if selection < 0 || selection >= len(options) {
		return false, fmt.Errorf("无效的选择")
	}
	optionID := options[selection].ID
	if optionID == "refuse" {
		return false, nil
	}

	cardIdx, err := strconv.Atoi(optionID)
	if err != nil || cardIdx < 0 || cardIdx >= len(player.Hand) {
		return false, fmt.Errorf("无效的卡牌索引")
	}
	card := player.Hand[cardIdx]
	if card.Type != model.CardTypeMagic {
		return false, fmt.Errorf("只能弃置法术牌")
	}

	e.NotifyCardRevealed(player.ID, []model.Card{card}, "discard")
	player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)
	e.State.DiscardPile = append(e.State.DiscardPile, card)
	e.Log(fmt.Sprintf("[Skill] %s 弃掉了法术牌 %s", player.Name, card.Name))
	return true, nil
}

func (e *GameEngine) resolveMagicBlastCasterForcedDiscard(
	act model.PlayerAction,
	data map[string]interface{},
	targetIDs []string,
	currentTargetIdx int,
) error {
	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}
	if act.Type != model.CmdSelect || len(act.Selections) != 1 {
		return fmt.Errorf("请选择1张牌弃置")
	}

	options := magicBlastCasterForcedDiscardOptions(player)
	selection := act.Selections[0]
	if selection < 0 || selection >= len(options) {
		return fmt.Errorf("无效的选择")
	}

	cardIdx, err := strconv.Atoi(options[selection].ID)
	if err != nil || cardIdx < 0 || cardIdx >= len(player.Hand) {
		return fmt.Errorf("无效的卡牌索引")
	}
	card := player.Hand[cardIdx]
	player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)
	e.State.DiscardPile = append(e.State.DiscardPile, card)
	e.Log(fmt.Sprintf("[Skill] %s 因【魔爆冲击】弃掉了 %s", player.Name, card.Name))

	return e.advanceMagicBlastToNextTarget(data, targetIDs, currentTargetIdx)
}

func (e *GameEngine) enterMagicBlastCasterForcedDiscard(data map[string]interface{}, casterID string, nextTargetIdx int) error {
	// 当前目标拒绝弃法术后，临时切换中断响应者到施法者执行“强制弃1”。
	data["stage"] = magicBlastStageCasterForcedDiscard
	data["current_target"] = nextTargetIdx
	e.State.PendingInterrupt.PlayerID = casterID
	e.State.PendingInterrupt.Context = data
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) advanceMagicBlastToNextTarget(data map[string]interface{}, targetIDs []string, nextTargetIdx int) error {
	// 目标链路按既定顺序逐个处理，全部处理完后才结束本次魔爆中断。
	if nextTargetIdx >= len(targetIDs) {
		e.PopInterrupt()
		return nil
	}
	data["stage"] = magicBlastStageTargetDiscard
	nextTargetID := targetIDs[nextTargetIdx]
	e.State.PendingInterrupt.PlayerID = nextTargetID
	e.State.PendingInterrupt.Context = data
	if nextTarget := e.State.Players[nextTargetID]; nextTarget != nil {
		e.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", nextTarget.Name))
	}
	e.notifyInterruptPrompt()
	return nil
}
