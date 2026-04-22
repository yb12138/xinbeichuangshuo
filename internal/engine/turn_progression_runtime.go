// gameflow: 回合切换、NextTurn、先后手更新。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

var turnScopedResetKeys = []string{
	// 全部迁出到 PlayerTurnState（UsedSkillCounts / SkillFlowState），
	// 由 NewPlayerTurnState() 在活跃玩家回合开始时自动清零。
}

// SetNextTurnPlayer 设置下回合玩家（用于额外回合等特殊机制）。
func (e *GameEngine) SetNextTurnPlayer(playerID string) {
	e.State.NextTurnPlayerOverride = playerID
}

func (e *GameEngine) NextTurn() {
	// Guard against turn progression during interrupt phases
	if e.State.PendingInterrupt != nil {
		e.Log("[Debug] NextTurn 被阻止：存在 PendingInterrupt")
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

	// 触发回合结束 TimingHookSpec（角色可在此设置额外回合）
	e.dispatchAllRoleTimingHooks(engineplayer.TimingOnTurnEndFinal, engineplayer.TimingHookContext{
		SourceID: player.ID,
	})

	// 真正的回合结束：清理当前玩家状态
	player.IsActive = false
	e.Log(fmt.Sprintf("[Turn] %s 结束回合", player.Name))

	// 切换到下一个玩家
	nextPid := currentPid
	if e.State.NextTurnPlayerOverride != "" {
		// 使用覆盖指定的下回合玩家
		nextPid = e.State.NextTurnPlayerOverride
		e.State.NextTurnPlayerOverride = "" // 清空覆盖
		// 找到该玩家在 PlayerOrder 中的索引
		for i, pid := range e.State.PlayerOrder {
			if pid == nextPid {
				e.State.CurrentTurn = i
				break
			}
		}
	} else {
		// 正常轮换
		e.State.CurrentTurn = (e.State.CurrentTurn + 1) % len(e.State.PlayerOrder)
		nextPid = e.State.PlayerOrder[e.State.CurrentTurn]
	}
	e.Log(fmt.Sprintf("[Debug] NextTurn 切换结果: from=%s to=%s", currentPid, nextPid))

	e.prepareNextTurnRuntime(e.State.Players[nextPid])
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
