package model

// 本文件：PlayerTurnState 上「本回合流程标记」的集中读写，供引擎与视图层调用，避免散落在 UsedSkillCounts 等结构中。

// HasStartupSkillOrSpecialActionsLocked 本回合是否不可再选择购买/合成/提炼（已用启动技，或因无法行动等路径被锁定特殊行动）。
func (t *PlayerTurnState) HasStartupSkillOrSpecialActionsLocked() bool {
	if t == nil {
		return false
	}
	return t.HasUsedActionSkill || t.SpecialActionsLockedThisTurn
}

// LockSpecialActionsForRemainderOfTurn 锁定本回合的特殊行动（买/合/提），例如宣告无法行动后的规则结果。
func (t *PlayerTurnState) LockSpecialActionsForRemainderOfTurn() {
	if t == nil {
		return
	}
	t.SpecialActionsLockedThisTurn = true
}
