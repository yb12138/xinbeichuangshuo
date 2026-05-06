// Package skills 注册并实现各角色技能的 CanUse / Execute（由 SkillDispatcher 在对应 FlowTiming 下调用）。
//
// 技能与数据配置的对应关系见 internal/data 角色卡与 SkillDefinition.LogicHandler；
// handlers_*.go 按职业/批次拆分，registry.go 在启动时 Register 各 handlerID。
//
// field_status_resolver.go：解析玩家场上盖牌/效果牌是否满足某类技能前置（如形态、指示物）。
package skills
