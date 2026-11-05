// gameflow: 回合切换、NextTurn、先后手更新。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

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
	e.dispatchAllRoleTimingHooks(engineplayer.TimingTurnEndFinal, engineplayer.TimingHookContext{
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

func (e *GameEngine) prepareNextTurnRuntime(nextPlayer *model.Player) {
	if nextPlayer == nil {
		return
	}

	nextPlayer.IsActive = true
	nextPlayer.TurnState = model.NewPlayerTurnState()
	e.resetTurnLimitedSkillCounts(nextPlayer)

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

	// 回合开始的结算（虚弱 / 中毒、希望赋格曲等）是新行动者的独立一手：不开新的叙事行动，
	// 它们会沿用上一名玩家"回合结束"那一手的 ActionID，被并进上一手的战斗链（如沉沦协奏曲之后）。
	e.beginNarrativeAction("turn_start", nextPlayer.ID)
}

// turnLimitAuxKeys 回合限定技能除技能 ID 以外、同样按"每个回合一次"计的辅助标记。
var turnLimitAuxKeys = map[string][]string{
	// 沉沦协奏曲：询问过（含拒绝）即算本回合已发动。
	"bd_descent_concerto": {"bd_descent", "bd_descent_prompted"},
}

// resetTurnLimitedSkillCounts 回合限定按"任意一个回合各限一次"计：每个玩家的回合开始时，
// 清掉其他玩家的回合限定计数（回合玩家自己的 TurnState 已整体重置）。
// 只清回合限定技能的计数，不动其他跨回合存活到"自己下回合"的标记。
func (e *GameEngine) resetTurnLimitedSkillCounts(turnPlayer *model.Player) {
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p == turnPlayer || p.Character == nil || p.TurnState.UsedSkillCounts == nil {
			continue
		}
		for _, skill := range p.Character.Skills {
			if !model.ContainsSkillTag(skill.Tags, model.TagTurnLimit) {
				continue
			}
			delete(p.TurnState.UsedSkillCounts, skill.ID)
			for _, key := range turnLimitAuxKeys[skill.ID] {
				delete(p.TurnState.UsedSkillCounts, key)
			}
		}
	}
}
