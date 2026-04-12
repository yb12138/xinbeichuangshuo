// gameflow: 技能运行时通用小工具（供引擎与测试复用）。

package engine

// ContainsSkillID 判断列表中是否包含指定技能 ID。
func ContainsSkillID(skillIDs []string, skillID string) bool {
	for _, id := range skillIDs {
		if id == skillID {
			return true
		}
	}
	return false
}
