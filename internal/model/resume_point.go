package model

import (
	"fmt"
	"strings"
)

const (
	// Resume point 协议前缀：用于把“恢复到哪个流程节点”编码成稳定字符串。
	// 这些字符串会跨模块流转（中断上下文、选择上下文、恢复点存储），
	// 所以必须固定且可判别，不能依赖调用方的自由文本。
	ResumePointTurnStagePrefix   = "turn:"
	ResumePointCombatStagePrefix = "combat:"
	ResumePointSubflowPrefix     = "subflow:"
)

// NormalizeResumePoint 只做“输入归一化”，不做规则校验。
// 规则校验（前缀、枚举合法性、维度互斥）统一在 ParseResumePoint 中完成。
//
// 设计上保留 string 直通，是为了允许上层把已编码好的协议值再次传入。
// 例如：从中断上下文读出的 "turn:BeforeAction" 可以直接参与后续解析。
func NormalizeResumePoint(raw interface{}) string {
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case TurnStage:
		if value == "" {
			return ""
		}
		return ResumePointTurnStagePrefix + string(value)
	case CombatStage:
		if value == "" {
			return ""
		}
		return ResumePointCombatStagePrefix + string(value)
	case Subflow:
		if value == "" {
			return ""
		}
		return ResumePointSubflowPrefix + string(value)
	default:
		return ""
	}
}

func IsKnownTurnStage(stage TurnStage) bool {
	switch stage {
	case TurnStageTurnBeforeStart,
		TurnStageTurnStart,
		TurnStageBeforeAction,
		TurnStageActionStart,
		TurnStageActionExecution,
		TurnStageActionEnd,
		TurnStageExtraAction,
		TurnStageTurnEnd:
		return true
	default:
		return false
	}
}

func IsKnownCombatStage(stage CombatStage) bool {
	switch stage {
	case CombatStageNone,
		CombatStageDeclare,
		CombatStageHitCheck,
		CombatStageCalcDamage,
		CombatStageHeal,
		CombatStageApply,
		CombatStageDraw:
		return true
	default:
		return false
	}
}

func IsKnownSubflow(subflow Subflow) bool {
	switch subflow {
	case SubflowNone,
		SubflowResponse,
		SubflowDiscardSelection:
		return true
	default:
		return false
	}
}

// ParseResumePoint 规则语义（游戏流程）：
// - resume point 表示“中断处理完成后，状态机回到哪里继续执行”。
// - 恢复目标必须是一个明确且可执行的流程节点，不能是默认值或模糊值。
// - 因此这里不做兼容兜底，任何非法值直接 panic，让错误在源头暴露。
//
// 常见来源：
// - 回合阶段中断恢复（TurnStage）
// - 战斗结算链中断恢复（CombatStage）
// - 子流程窗口恢复（Subflow，如响应窗/弃牌选择）
//
// 注意：Strict-Mode 下仅接受强类型枚举输入，不接受字符串（包括 "turn:xxx" 格式）。
func ParseResumePoint(raw interface{}) (TurnStage, CombatStage, Subflow) {
	// 强类型分支：调用方直接传枚举值时，不走字符串前缀解析。
	switch value := raw.(type) {
	case TurnStage:
		if !IsKnownTurnStage(value) {
			panic(fmt.Sprintf("invalid turn-stage resume point: %q", value))
		}
		return value, CombatStageNone, SubflowNone
	case CombatStage:
		if !IsKnownCombatStage(value) || value == CombatStageNone {
			panic(fmt.Sprintf("invalid combat-stage resume point: %q", value))
		}
		return "", value, SubflowNone
	case Subflow:
		if !IsKnownSubflow(value) || value == SubflowNone {
			panic(fmt.Sprintf("invalid subflow resume point: %q", value))
		}
		return "", CombatStageNone, value
	}
	panic(fmt.Sprintf("invalid resume point payload type %T", raw))
}

// ParseResumePointTurnStage 是强约束包装：
// - 用于“我明确只接受回合主流程恢复点”的场景。
// - resume point 必须是 turn 维度，否则直接 panic。
func ParseResumePointTurnStage(raw interface{}) TurnStage {
	stage, combat, subflow := ParseResumePoint(raw)
	if stage == "" || combat != CombatStageNone || subflow != SubflowNone {
		panic(fmt.Sprintf("resume point is not a turn stage: %v", raw))
	}
	return stage
}

// ParseResumePointCombatStage 是强约束包装：
// - 用于“我明确只接受战斗子流程恢复点”的场景。
// - resume point 必须是 combat 维度，且不能是 CombatStageNone。
func ParseResumePointCombatStage(raw interface{}) CombatStage {
	stage, combat, subflow := ParseResumePoint(raw)
	if stage != "" || combat == CombatStageNone || subflow != SubflowNone {
		panic(fmt.Sprintf("resume point is not a combat stage: %v", raw))
	}
	return combat
}

// ParseResumePointSubflow 是强约束包装：
// - 用于“我明确只接受子流程窗口恢复点”的场景。
// - resume point 必须是 subflow 维度，且不能是 SubflowNone。
func ParseResumePointSubflow(raw interface{}) Subflow {
	stage, combat, subflow := ParseResumePoint(raw)
	if stage != "" || combat != CombatStageNone || subflow == SubflowNone {
		panic(fmt.Sprintf("resume point is not a subflow: %v", raw))
	}
	return subflow
}
