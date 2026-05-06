// gameflow: 角色身份判定基础设施。

package player

import "starcup-engine/internal/model"

// IsCharacter 判断玩家是否为指定角色。
func IsCharacter(p *model.Player, charID string) bool {
	return p != nil && p.Character != nil && p.Character.ID == charID
}
