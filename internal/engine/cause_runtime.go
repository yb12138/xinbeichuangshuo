// gameflow: 效果因果（Cause）运行时——维护“当前因果”，并在汇聚点写入 / 恢复。
//
// 规则只有两条：
//  1. 同步嵌套由调用栈负责：withCause 进入时设置当前因果，defer 恢复。
//  2. 跨中断延续由数据负责：Interrupt / PendingDamage / FieldCard / CombatRequest 创建时记下
//     当前因果，恢复执行时（选择结算、伤害结算、场上牌触发）再取回。
//
// 角色代码不直接操作因果；需要在一个 hook 内区分多个技能时调用 WithSkillCause。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type causeState struct {
	current   *model.Cause
	seq       int
	damageSeq int64
	diffing   map[string]bool // 正在比较战斗状态的因果 ID
}

func (e *GameEngine) causes() *causeState {
	if e.causeRuntime == nil {
		e.causeRuntime = &causeState{}
	}
	return e.causeRuntime
}

// CurrentCause 返回当前因果（可能为 nil：基础行动或规则结算）。
func (e *GameEngine) CurrentCause() *model.Cause {
	if e == nil || e.causeRuntime == nil {
		return nil
	}
	return e.causeRuntime.current
}

// withCause 在 cause 下同步执行 fn，结束后恢复原因果。cause 为 nil 表示“无技能来源”（基础行动 /
// 规则结算），同样会生效——恢复中断或结算伤害时必须精确还原当时的因果，不能继承外层。
func (e *GameEngine) withCause(cause *model.Cause, fn func()) {
	st := e.causes()
	prev := st.current
	if sameCause(prev, cause) {
		fn()
		return
	}
	e.emitCauseBoundary()
	st.current = cause
	defer func() {
		e.emitCauseBoundary()
		st.current = prev
	}()
	fn()
}

func sameCause(a, b *model.Cause) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.ID == b.ID
}

// emitCauseBoundary 在切换因果前通知观察者：服务端据此把公开状态差值切到正确的因果名下。
func (e *GameEngine) emitCauseBoundary() {
	if e == nil || e.observer == nil {
		return
	}
	e.emitGameEvent(model.GameEvent{Type: model.EventCauseBoundary})
}

// newSkillCause 以当前因果为父，开一个技能子因果。
func (e *GameEngine) newSkillCause(skillID, ownerID string) *model.Cause {
	st := e.causes()
	st.seq++
	cause := &model.Cause{
		ID:        fmt.Sprintf("c%d", st.seq),
		SkillID:   skillID,
		SkillName: e.skillTitleFor(ownerID, skillID),
		OwnerID:   ownerID,
	}
	if parent := st.current; parent != nil {
		cause.ParentID = parent.ID
	}
	return cause
}

// WithSkillCause 在一个技能子因果下执行 fn。供引擎入口与“一个 hook 负责多个技能”的角色代码使用。
// 若当前因果已是同一玩家的同一技能（如宣言后的结算、同一技能的后续选择），直接复用，不再开新因果。
func (e *GameEngine) WithSkillCause(skillID, ownerID string, fn func()) {
	if e == nil {
		fn()
		return
	}
	if skillID == "" {
		fn()
		return
	}
	run := func() { e.runWithCombatDiff(fn) }
	if cur := e.CurrentCause(); cur != nil && cur.SkillID == skillID && cur.OwnerID == ownerID {
		// 同一技能的后续步骤（宣言后的结算、多步选择）沿用同一因果。
		run()
		return
	}
	e.withCause(e.newSkillCause(skillID, ownerID), run)
}

// withOriginCause 延迟效果被引爆：新因果归属放置者，父因果为当前（引爆它的行动），Origin 指向放置时的因果。
func (e *GameEngine) withOriginCause(origin *model.Cause, fn func()) {
	if origin == nil {
		fn()
		return
	}
	cause := e.newSkillCause(origin.SkillID, origin.OwnerID)
	cause.OriginID = origin.ID
	e.withCause(cause, func() { e.runWithCombatDiff(fn) })
}

func (e *GameEngine) skillTitleFor(playerID, skillID string) string {
	if e == nil || e.State == nil || skillID == "" {
		return ""
	}
	if p := e.State.Players[playerID]; p != nil && p.Character != nil {
		for _, def := range p.Character.Skills {
			if def.ID == skillID {
				return def.Title
			}
		}
	}
	return ""
}

// stampCause 把当前因果写进尚未记录来源的延迟对象。已有因果的对象保持不变（例如重新入队的伤害）。
func (e *GameEngine) stampInterruptCause(intr *model.Interrupt) {
	if intr != nil && intr.Cause == nil {
		intr.Cause = model.CloneCause(e.CurrentCause())
	}
}

func (e *GameEngine) stampPendingDamageCause(pd *model.PendingDamage) {
	if pd == nil {
		return
	}
	if pd.Cause == nil {
		pd.Cause = model.CloneCause(e.CurrentCause())
	}
	if pd.Seq == 0 {
		st := e.causes()
		st.damageSeq++
		pd.Seq = st.damageSeq
	}
}

func (e *GameEngine) stampFieldCardCause(fc *model.FieldCard) {
	if fc != nil && fc.Cause == nil {
		fc.Cause = model.CloneCause(e.CurrentCause())
	}
}

// interruptCause 返回待处理中断记录的因果，供恢复执行时取回。
func (e *GameEngine) pendingInterruptCause() *model.Cause {
	if e == nil || e.State == nil || e.State.PendingInterrupt == nil {
		return nil
	}
	return e.State.PendingInterrupt.Cause
}
