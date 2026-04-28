// gameflow: 魔法少女模块入口声明。

package magical_girl

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:          "magical_girl",
		MagicBullet: player.MagicBulletAbilities{CanFuse: true, CanDirect: true},
		Choices:     NewChoiceHandler(),
		Skills:      SkillEntries(),
		InterruptSpecs: []player.InterruptSpec{
			{
				Type:                 model.InterruptMagicMissile,
				BuildPrompt:          buildMagicMissilePrompt,
				HandleAction:         handleMagicMissileAction,
				AllowedActionTypes:   []model.PlayerActionType{model.CmdRespond},
				InvalidActionMessage: "当前为【魔弹】响应阶段，请使用响应指令",
			},
			{
				Type:                 model.InterruptMagicBulletFusion,
				BuildPrompt:          buildMagicBulletFusionPrompt,
				HandleAction:         handleMagicBulletFusionAction,
				AllowedActionTypes:   []model.PlayerActionType{model.CmdSelect},
				InvalidActionMessage: "当前为【魔弹融合】确认阶段，请选择是否发动",
			},
			{
				Type:                 model.InterruptMagicBulletDirection,
				BuildPrompt:          buildMagicBulletDirectionPrompt,
				HandleAction:         handleMagicBulletDirectionAction,
				AllowedActionTypes:   []model.PlayerActionType{model.CmdSelect},
				InvalidActionMessage: "当前为【魔弹掌控】方向选择阶段，请提交选择",
			},
			{
				Type:                 model.InterruptMagicBlast,
				BuildPrompt:          buildMagicBlastPrompt,
				HandleAction:         handleMagicBlastAction,
				AllowedActionTypes:   []model.PlayerActionType{model.CmdSelect, model.CmdCancel},
				InvalidActionMessage: "当前为【魔爆冲击】响应阶段，请选择弃牌或取消",
			},
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "magic_bullet_control", Handler: &MagicBulletControlHandler{}},
		{ID: "magic_bullet_fusion", Handler: &MagicBulletFusionHandler{}},
		{ID: "magic_blast", Handler: &MagicBlastHandler{}},
		{ID: "destruction_storm", Handler: &DestructionStormHandler{}},
	}
}
