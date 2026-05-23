// gameflow: 无法行动判断逻辑。

package engine

import (
	"fmt"
	playerpkg "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// DefaultCannotActChecker 默认无法行动判断。
// 检查玩家是否有可用的攻击牌或法术牌，或是否有可用的行动技能。
func DefaultCannotActChecker(e *GameEngine, player *model.Player) (bool, string) {
	// 1. 检查手牌是否有可执行动作
	if len(player.Hand) > 0 {
		canUseMagic := e.canCastMagicInAction(player)
		total := e.playableCardCount(player)
		for idx := 0; idx < total; idx++ {
			card, _, _, ok := e.getPlayableCardByIndex(player, idx)
			if !ok {
				continue
			}
			if card.Type == model.CardTypeAttack {
				return false, "有攻击牌可用"
			}
			if card.Type == model.CardTypeMagic && canUseMagic {
				return false, "有法术牌可用"
			}
		}
	}

	// 2. 其他情况：可以宣告无法行动
	return true, "没有可用的攻击牌、法术牌或行动技能"
}

// checkPlayerCannotAct 检查玩家是否可以宣告无法行动。
// 判断流程：
//  1. 如果有可用的行动技能，则不能宣告无法行动（硬性条件，优先级最高）
//  2. 调用角色自定义hook，若返回true则可以宣告无法行动
//  3. 无hook或hook未拦截时，使用默认判断
func (e *GameEngine) checkPlayerCannotAct(player *model.Player) (bool, string) {
	// 硬性条件：有可用的行动技能时，任何角色都不能宣告无法行动
	if e.hasUsableActionSkillForExtraMagic(player) {
		return false, "有行动技能可用"
	}

	roleID := ""
	if player.Character != nil {
		roleID = player.Character.ID
	}

	// 角色自定义hook
	checker := roleRegistry.CannotActChecker(roleID)
	if checker != nil {
		canCannotAct, reason := checker(player)
		if canCannotAct {
			return true, "[角色规则] " + reason
		}
		// 角色hook返回false时，继续走默认判断
	}

	// 默认判断
	return DefaultCannotActChecker(e, player)
}

// skipExtraAction 跳过额外行动。
func (e *GameEngine) skipExtraAction(player *model.Player) {
	constraintInfo := e.buildConstraintInfo(player.TurnState.CurrentExtraAction, player.TurnState.CurrentExtraElement)
	e.BeginActionSummary("cannot_act", player.ID, "跳过额外行动", nil)
	e.Log(fmt.Sprintf("[Turn] %s 宣告【无法行动】，跳过本次额外行动%s", player.Name, constraintInfo))
	player.TurnState.CurrentExtraAction = ""
	player.TurnState.CurrentExtraElement = nil
	e.enterTurnEndStage()
}

// executeCannotActFlow 执行无法行动流程（展示、弃牌、重摸）。
func (e *GameEngine) executeCannotActFlow(player *model.Player) {
	e.BeginActionSummary("cannot_act", player.ID, "无法行动", nil)
	handCount := len(player.Hand)

	if handCount == 0 {
		e.Log(fmt.Sprintf("[Action] %s 宣告【无法行动】（无手牌），结束本回合行动阶段", player.Name))
		player.TurnState.LockSpecialActionsForRemainderOfTurn()
		e.enterTurnEndStage()
		return
	}

	// 展示并弃掉全部手牌
	e.Log(fmt.Sprintf("[Action] %s 宣告【无法行动】，展示并弃掉全部手牌(%d张)", player.Name, handCount))
	e.NotifyCardRevealed(player.ID, append([]model.Card{}, player.Hand...), "discard")
	for _, c := range player.Hand {
		e.State.DiscardPile = append(e.State.DiscardPile, c)
	}
	player.Hand = player.Hand[:0]

	// 重摸相同数量的牌
	cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, handCount)
	e.State.Deck = newDeck
	e.State.DiscardPile = newDiscard
	player.Hand = append(player.Hand, cards...)
	e.NotifyDrawCards(player.ID, handCount, "cannot_act_redraw")

	// 角色特定后续处理（如魔剑士全法术重摸）
	e.dispatchAllRoleTimingHooks(playerpkg.TimingAfterCannotAct, playerpkg.TimingHookContext{
		Player: player,
	})

	e.Log(fmt.Sprintf("[Action] %s 重新摸了%d张牌，且本回合不可执行特殊行动", player.Name, handCount))
	player.TurnState.LockSpecialActionsForRemainderOfTurn()
	e.enterActionExecutionStage()
}
