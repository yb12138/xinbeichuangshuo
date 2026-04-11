// gameflow: 回合切换、NextTurn、先后手更新。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

var turnScopedResetKeys = []string{
	// 全部迁出到 PlayerTurnState（UsedSkillCounts / SkillFlowState），
	// 由 NewPlayerTurnState() 在活跃玩家回合开始时自动清零。
}

func (e *GameEngine) NextTurn() {
	// Guard against turn progression during interrupt phases
	if e.State.PendingInterrupt != nil {
		return // Silently prevent turn progression during interrupts
	}
	if e.actionSummaryTurn <= 0 {
		e.actionSummaryTurn = 1
	} else {
		e.actionSummaryTurn++
	}

	currentPid := e.State.PlayerOrder[e.State.CurrentTurn]
	player := e.State.Players[currentPid]

	e.expireRuleModifiersByLifetime(player, model.RuleLifeUntilTurnEnd)

	extraTurn := e.consumePendingMoonGoddessExtraTurn(player)

	// 真正的回合结束：清理当前玩家状态
	player.IsActive = false
	e.Log(fmt.Sprintf("[Turn] %s 结束回合", player.Name))

	// 切换到下一个玩家（若有额外回合则保持当前玩家）
	nextPid := currentPid
	if !extraTurn {
		e.State.CurrentTurn = (e.State.CurrentTurn + 1) % len(e.State.PlayerOrder)
		nextPid = e.State.PlayerOrder[e.State.CurrentTurn]
	} else {
		e.Log(fmt.Sprintf("%s 的 [苍白之月] 生效：立即获得额外回合", player.Name))
	}

	e.prepareNextTurnRuntime(e.State.Players[nextPid])
}

func (e *GameEngine) consumePendingMoonGoddessExtraTurn(player *model.Player) bool {
	if player == nil || !e.isMoonGoddess(player) || player.Tokens == nil || player.Tokens["mg_extra_turn_pending"] <= 0 {
		return false
	}
	player.Tokens["mg_extra_turn_pending"]--
	if player.Tokens["mg_extra_turn_pending"] < 0 {
		player.Tokens["mg_extra_turn_pending"] = 0
	}
	return true
}

func (e *GameEngine) resetTurnScopedPlayerTokens() {
	for _, player := range e.State.Players {
		if player == nil {
			continue
		}
		ensurePlayerTokensMap(player)
		for _, key := range turnScopedResetKeys {
			player.Tokens[key] = 0
		}
	}
}

func (e *GameEngine) prepareNextTurnRuntime(nextPlayer *model.Player) {
	if nextPlayer == nil {
		return
	}

	nextPlayer.IsActive = true
	nextPlayer.TurnState = model.NewPlayerTurnState()

	e.resetTurnScopedPlayerTokens()
	e.resetTurnMagicDamageTracker()

	// 重置 Engine / FSM 级状态
	e.State.ActionQueue = []model.QueuedAction{}
	e.State.CombatStack = []model.CombatRequest{}
	e.clearSubflow()

	e.Log(fmt.Sprintf("[Turn] %s 回合开始 (Hand:%d Gem:%d Cry:%d)",
		nextPlayer.Name, len(nextPlayer.Hand), nextPlayer.Gem, nextPlayer.Crystal))

	// 设置阶段为第1步：Buff结算
	e.setTurnStage(model.TurnStageTurnBeforeStart)
	e.clearCombatStage()
}
