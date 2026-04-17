// gameflow: 神箭手模块入口声明。

package archer

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "piercing_shot", Handler: &skills.PiercingShotHandler{}},
		{ID: "lightning_arrow", Handler: &skills.LightningArrowHandler{}},
		{ID: "snipe", Handler: &skills.SnipeHandler{}},
		{ID: "precise_shot", Handler: &skills.PreciseShotHandler{}},
		{ID: "flash_trap", Handler: &skills.FlashTrapHandler{}},
	}
}
