// gameflow: 玩家在中断中的 CmdSelect 等动作路由。

package engine

import "starcup-engine/internal/model"

type interruptActionHandler func(*GameEngine, model.PlayerAction) error

type interruptActionRule struct {
	// allowed 为空表示由具体处理器自行校验输入类型。
	allowed map[model.PlayerActionType]bool
	// invalidActionMessage 用于统一提示“该中断阶段不接受此类输入”。
	invalidActionMessage string
	handler              interruptActionHandler
}

func allowedInterruptActionTypes(types ...model.PlayerActionType) map[model.PlayerActionType]bool {
	set := make(map[model.PlayerActionType]bool, len(types))
	for _, t := range types {
		set[t] = true
	}
	return set
}

var interruptActionRules = map[model.InterruptType]interruptActionRule{
	// 技能响应窗口：技能系统内部决定“发动/跳过”等输入规则。
	model.InterruptResponseSkill: {
		handler: (*GameEngine).handleInterruptResponseSkillAction,
	},
	// 启动技窗口：技能系统内部决定“发动/跳过”等输入规则。
	model.InterruptStartupSkill: {
		handler: (*GameEngine).handleInterruptStartupSkillAction,
	},
	// 弃牌窗口：弃牌处理器内部区分强制弃牌与可取消弃牌。
	model.InterruptDiscard: {
		handler: (*GameEngine).handleInterruptDiscardAction,
	},
	// 给牌窗口：给牌处理器内部校验选择数量与接收方。
	model.InterruptGiveCards: {
		handler: (*GameEngine).handleInterruptGiveCardsAction,
	},
	// 通用选择窗口：由 choice 注册器驱动具体分支。
	model.InterruptChoice: {
		handler: (*GameEngine).handleInterruptChoiceAction,
	},

	// 魔弹响应：只能 respond（承受/应战/防御）。
	model.InterruptMagicMissile: {
		allowed:              allowedInterruptActionTypes(model.CmdRespond),
		invalidActionMessage: "当前为【魔弹】响应阶段，请使用响应指令",
		handler:              (*GameEngine).handleMagicMissileResponse,
	},
	// 魔弹融合：只能 select（是否当作魔弹使用）。
	model.InterruptMagicBulletFusion: {
		allowed:              allowedInterruptActionTypes(model.CmdSelect),
		invalidActionMessage: "当前为【魔弹融合】确认阶段，请选择是否发动",
		handler:              (*GameEngine).handleMagicBulletFusionResponse,
	},
	// 魔弹掌控：只能 select（传递方向）。
	model.InterruptMagicBulletDirection: {
		allowed:              allowedInterruptActionTypes(model.CmdSelect),
		invalidActionMessage: "当前为【魔弹掌控】方向选择阶段，请提交选择",
		handler:              (*GameEngine).handleMagicBulletDirectionResponse,
	},
	// 圣剑后续：只能 select（摸X弃X的X值）。
	model.InterruptHolySwordDraw: {
		allowed:              allowedInterruptActionTypes(model.CmdSelect),
		invalidActionMessage: "当前为【圣剑】后续选择阶段，请提交选择",
		handler:              (*GameEngine).handleHolySwordDrawResponse,
	},
	// 圣疗：只能 select（治疗分配 / 额外行动类型）。
	model.InterruptSaintHeal: {
		allowed:              allowedInterruptActionTypes(model.CmdSelect),
		invalidActionMessage: "当前为【圣疗】选择阶段，请提交选择",
		handler:              (*GameEngine).handleSaintHealResponse,
	},
	// 魔爆冲击：目标可 select 弃牌，或 cancel 表示承受伤害。
	model.InterruptMagicBlast: {
		allowed:              allowedInterruptActionTypes(model.CmdSelect, model.CmdCancel),
		invalidActionMessage: "当前为【魔爆冲击】响应阶段，请选择弃牌或取消",
		handler:              (*GameEngine).handleMagicBlastResponse,
	},
}
