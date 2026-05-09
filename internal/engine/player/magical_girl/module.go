// gameflow: 魔法少女模块入口声明。

package magical_girl

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:          "magical_girl",
		MagicBullet: player.MagicBulletAbilities{CanFuse: true, CanDirect: true},
		Choices:     NewChoiceHandler(),
		Skills:      SkillEntries(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnMagicMissileResponseSkillAug, Priority: 100, Hook: magicMissileResponseSkillAugHook},
		},
		InterruptSpecs: []player.InterruptSpec{
			{
				Type:                 model.InterruptMagicMissile,
				PhaseSync:            player.InterruptPhaseSyncResponseWindow,
				BuildPrompt:          buildMagicMissilePrompt,
				HandleActionResult:   handleMagicMissileAction,
				AllowedActionTypes:   []model.PlayerActionType{model.CmdRespond},
				InvalidActionMessage: "当前为【魔弹】响应阶段，请使用响应指令",
			},

			{
				Type:                 model.InterruptMagicBulletDirection,
				PhaseSync:            player.InterruptPhaseSyncActionExecution,
				BuildPrompt:          buildMagicBulletDirectionPrompt,
				HandleActionResult:   handleMagicBulletDirectionAction,
				AllowedActionTypes:   []model.PlayerActionType{model.CmdSelect},
				InvalidActionMessage: "当前为【魔弹掌控】方向选择阶段，请提交选择",
			},
			{
				Type:                 model.InterruptMagicBlast,
				PhaseSync:            player.InterruptPhaseSyncResponseWindow,
				BuildPrompt:          buildMagicBlastPrompt,
				HandleActionResult:   handleMagicBlastAction,
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
		{ID: "magic_bullet_fusion", Handler: &MagicBulletFusionHandler{}, Policy: types.SkillPolicy{
			ValidateDiscardedCards: validateFusionDiscard,
		}},
		{ID: "magic_bullet_fusion_chain", Handler: &MagicBulletFusionChainHandler{}},
		{ID: "magic_blast", Handler: &MagicBlastHandler{}},
		{ID: "destruction_storm", Handler: &DestructionStormHandler{}},
	}
}
