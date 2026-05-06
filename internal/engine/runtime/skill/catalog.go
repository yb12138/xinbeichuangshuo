// gameflow: 角色技能定义查找（单一事实源入口）。

package skill

import "starcup-engine/internal/model"

// Catalog 提供技能定义查询。
type Catalog struct{}

// NewCatalog 创建目录。
func NewCatalog() *Catalog {
	return &Catalog{}
}

// FindCharacterSkill 在角色技能表中查找定义。
func (c *Catalog) FindCharacterSkill(character *model.Character, skillID string) *model.SkillDefinition {
	if character == nil {
		return nil
	}
	for i := range character.Skills {
		if character.Skills[i].ID == skillID {
			return &character.Skills[i]
		}
	}
	return nil
}

// FindCharacterSkillOnPlayer 在玩家当前角色上查找技能。
func (c *Catalog) FindCharacterSkillOnPlayer(player *model.Player, skillID string) *model.SkillDefinition {
	if player == nil || player.Character == nil {
		return nil
	}
	return c.FindCharacterSkill(player.Character, skillID)
}

// ResolveHandlerID LogicHandler 优先，否则回退技能 ID。
func ResolveHandlerID(skillDef model.SkillDefinition) string {
	if skillDef.LogicHandler != "" {
		return skillDef.LogicHandler
	}
	return skillDef.ID
}
