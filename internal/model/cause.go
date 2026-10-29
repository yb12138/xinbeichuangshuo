// gameflow: 效果因果（Cause）——把每个公开事件 / 状态变化归属到引起它的技能或行动。

package model

// Cause 描述一个效果的来源。引擎维护“当前因果”：
//   - 同步嵌套：进入技能 / hook 时开子因果，退出时恢复（Go 调用栈天然保证不串）。
//   - 跨中断延续：Interrupt / PendingDamage / FieldCard / CombatRequest 创建时记下当前因果，
//     恢复执行时再取回，所以中断前后的效果都落在同一个因果名下。
//
// 前端按 ID 聚合同一次技能的宣言、结算与状态变化，按 ParentID 画触发链，按 OriginID
// 找到延迟效果（如封印）当初是被哪次技能放下的。
type Cause struct {
	ID        string `json:"id"`
	SkillID   string `json:"skill_id,omitempty"`   // 空 = 非技能来源（基础行动、规则结算）
	SkillName string `json:"skill_name,omitempty"` // 演出用名称
	OwnerID   string `json:"owner_id,omitempty"`   // 演出归属的玩家（封印引爆时是封印师）
	ParentID  string `json:"parent_id,omitempty"`  // 触发链：这次效果是被哪个因果引起的
	OriginID  string `json:"origin_id,omitempty"`  // 放置来源：延迟效果当初由哪个因果放下
}

// CloneCause 返回因果的浅拷贝（nil 安全）。
func CloneCause(c *Cause) *Cause {
	if c == nil {
		return nil
	}
	out := *c
	return &out
}
