// gameflow: SkillDispatcher：OnTiming 收集候选技能并 processSkills。

package engine

import "starcup-engine/internal/model"

// SkillDispatcher 统一技能调度器。
// 职责：在 FlowTiming 窗口内收集候选技能并执行（或挂起中断等待玩家确认）。
type SkillDispatcher struct {
	engine *GameEngine
}

// NewSkillDispatcher 创建技能调度器。
func NewSkillDispatcher(engine *GameEngine) *SkillDispatcher {
	return &SkillDispatcher{engine: engine}
}

// checkTarget 表示一次时窗扫描中“要检查哪个玩家、以什么身份检查”。
// 例如在攻击相关窗口中，同一个事件会分别以 Attacker/Defender 身份检查不同技能。
type checkTarget struct {
	Player *model.Player
	Role   model.SkillRole
}

type targetSkillsBatch struct {
	target    checkTarget
	ctx       *model.Context
	skills    []model.SkillDefinition
	priority  int
	seatOrder int
}
