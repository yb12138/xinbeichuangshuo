// gameflow: 技能因果内的非公开战斗状态比较——改向 / 代为应战 / 伤害修正等语义事件。
//
// 公开状态（士气、宝石、标记、场上牌）由服务端快照差值覆盖；但战斗栈与待结算伤害不在公开快照里，
// 角色代码又常直接改字段（pd.Damage--、弹出原攻击后压入反击）。这里在技能因果的进入 / 退出处
// 各记一次，变化即发出带当前因果的 TimelineMarker，角色代码无需手动通知。

package engine

import (
	"starcup-engine/internal/model"
)

type combatSnapshot struct {
	hasCombat  bool
	attackerID string
	targetID   string

	hasDamage  bool
	damageSeq  int64
	sourceID   string
	victimID   string
	damage     int
	damageType model.DamageType
}

func (e *GameEngine) captureCombatSnapshot() combatSnapshot {
	snap := combatSnapshot{}
	if e == nil || e.State == nil {
		return snap
	}

	if n := len(e.State.CombatStack); n > 0 {
		top := &e.State.CombatStack[n-1]
		snap.hasCombat = true
		snap.attackerID = top.AttackerID
		snap.targetID = top.TargetID
	}
	if len(e.State.PendingDamageQueue) > 0 {
		head := &e.State.PendingDamageQueue[0]
		snap.hasDamage = true
		snap.damageSeq = head.Seq
		snap.sourceID = head.SourceID
		snap.victimID = head.TargetID
		snap.damage = head.Damage
		snap.damageType = head.DamageType
	}
	return snap
}

// publishCombatDiff 比较技能执行前后的战斗状态，生成语义事件（在当前技能因果下发出）。
func (e *GameEngine) publishCombatDiff(before combatSnapshot) {
	cause := e.CurrentCause()
	if cause == nil || cause.SkillID == "" {
		return
	}
	after := e.captureCombatSnapshot()

	// 攻击改向 / 代为应战：栈顶仍是“一次攻击”，但目标或攻击者变了。
	if before.hasCombat && after.hasCombat && (before.targetID != after.targetID || before.attackerID != after.attackerID) {
		kind := "attack_redirected"
		if before.targetID != after.targetID && after.attackerID == cause.OwnerID && after.attackerID != before.attackerID {
			// 原攻击被移除、由技能持有者发出新的攻击：代为应战（阴阳师式神咒束等）。
			kind = "attack_taken_over"
		}
		e.publishTimelineMarker(model.TimelineMarkerPayload{
			PlayerID:      cause.OwnerID,
			SkillID:       cause.SkillID,
			SkillName:     cause.SkillName,
			ActionType:    "skill",
			NarrativeKind: kind,
			VisualKind:    "skill_effect",
			SkillPhase:    "resolved",
			Summary:       e.redirectSummary(before, after),
			TargetIDs:     uniqueNonEmpty(before.targetID, after.targetID),
		})
	}

	// 伤害修正：同一笔待结算伤害的数值或类型被技能改了（毒粉 +1、朝圣 -1、伤害转弃牌等）。
	// 伤害降为 0 视为抵消（damage_prevented）。
	if before.hasDamage && after.hasDamage && before.damageSeq == after.damageSeq &&
		(before.damage != after.damage || before.damageType != after.damageType) {
		e.publishDamageModified(cause, before.victimID, before.damage, after.damage, before.damageType, after.damageType)
	}
}

func (e *GameEngine) publishDamageModified(cause *model.Cause, victimID string, from, to int, fromType, toType model.DamageType) {
	marker := model.TimelineMarkerPayload{
		PlayerID:      cause.OwnerID,
		SkillID:       cause.SkillID,
		SkillName:     cause.SkillName,
		ActionType:    "skill",
		NarrativeKind: "damage_modified",
		VisualKind:    "skill_effect",
		SkillPhase:    "resolved",
		TargetIDs:     uniqueNonEmpty(victimID),
		Summary:       damageModifiedSummary(from, to, fromType, toType),
	}
	if to == 0 {
		marker.NarrativeKind = "damage_prevented"
	}
	e.publishTimelineMarker(marker)
}

func (e *GameEngine) redirectSummary(before, after combatSnapshot) string {
	name := func(id string) string {
		if p := e.State.Players[id]; p != nil {
			return p.Name
		}
		return id
	}
	if before.targetID != after.targetID {
		return name(before.targetID) + " → " + name(after.targetID)
	}
	return name(before.attackerID) + " → " + name(after.attackerID)
}

func damageModifiedSummary(from, to int, fromType, toType model.DamageType) string {
	if to == 0 {
		return "伤害抵消"
	}
	if fromType != toType {
		return "伤害类型改变"
	}
	diff := to - from
	if diff > 0 {
		return "伤害 +" + itoa(diff)
	}
	return "伤害 " + itoa(diff)
}

func itoa(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	digits := []byte{}
	for {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
		if n == 0 {
			break
		}
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

func uniqueNonEmpty(ids ...string) []string {
	out := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// runWithCombatDiff 执行 fn 并比较战斗状态；同一因果只在最外层比较一次（宣言后的结算、多步选择
// 会重入同一因果，内层不重复发出）。
func (e *GameEngine) runWithCombatDiff(fn func()) {
	cause := e.CurrentCause()
	if cause == nil {
		fn()
		return
	}
	st := e.causes()
	if st.diffing[cause.ID] {
		fn()
		return
	}
	if st.diffing == nil {
		st.diffing = map[string]bool{}
	}
	st.diffing[cause.ID] = true
	defer delete(st.diffing, cause.ID)
	before := e.captureCombatSnapshot()
	fn()
	e.publishCombatDiff(before)
}
