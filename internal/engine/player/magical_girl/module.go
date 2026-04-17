// gameflow: 魔法少女模块入口声明。

package magical_girl

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "magic_bullet_control", Handler: &skills.MagicBulletControlHandler{}},
		{ID: "magic_bullet_fusion", Handler: &skills.MagicBulletFusionHandler{}},
		{ID: "magic_blast", Handler: &skills.MagicBlastHandler{}},
		{ID: "destruction_storm", Handler: &skills.DestructionStormHandler{}},
	}
}
