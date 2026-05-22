package model

// Context 上 Timing-first 的语义查询。

// AttackDeclaredPhase 主动攻击宣告/攻击开始窗口。
func (ctx *Context) AttackDeclaredPhase() bool {
	if ctx == nil {
		return false
	}
	switch ctx.Timing {
	case TimingAttackDeclare, TimingOnAttackDeclared:
		return true
	default:
		return false
	}
}

// AttackCommittedPhase 表示一次主动/应战攻击已经完成阶段①提交。
func (ctx *Context) AttackCommittedPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingAttackCommitted
}

// BeforeDrawPhase 表示“摸牌前”窗口。
func (ctx *Context) BeforeDrawPhase() bool {
	if ctx == nil {
		return false
	}
	switch ctx.Timing {
	case TimingBeforeCardDrawn, TimingSettleDraw:
		return true
	default:
		return false
	}
}

// AfterDrawPhase 表示“摸牌后”窗口。
func (ctx *Context) AfterDrawPhase() bool {
	if ctx == nil {
		return false
	}
	switch ctx.Timing {
	case TimingOnCardDrawn, TimingSettleDraw:
		return true
	default:
		return false
	}
}

// AttackHitPhase 响应恢复：攻击命中分支。
func (ctx *Context) AttackHitPhase() bool {
	if ctx == nil {
		return false
	}
	if ctx.Timing == TimingAttackHit {
		return true
	}
	return ctx.Timing == TimingOnHitCheck && ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.IsHit
}

// ResumeAttackHitPhase 响应恢复：攻击命中分支。
func (ctx *Context) ResumeAttackHitPhase() bool {
	return ctx.AttackHitPhase()
}

// AttackMissPhase 响应恢复：攻击未命中分支。
func (ctx *Context) AttackMissPhase() bool {
	if ctx == nil {
		return false
	}
	if ctx.Timing == TimingAttackMiss {
		return true
	}
	return ctx.Timing == TimingOnHitCheck && ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil && !ctx.EventCtx.AttackInfo.IsHit
}

// ResumeAttackMissPhase 响应恢复：攻击未命中分支。
func (ctx *Context) ResumeAttackMissPhase() bool {
	return ctx.AttackMissPhase()
}

// MagicResolvePhase 表示主动法术进入效果结算。
func (ctx *Context) MagicResolvePhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingMagicResolve
}

// ResumeDamageTakenPhase 承伤响应恢复上下文。
func (ctx *Context) ResumeDamageTakenPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingDamageTaken || ctx.Timing == TimingOnDamageTaken
}

// ResumeBeforeMoraleLossPhase 士气下降前响应恢复上下文。
func (ctx *Context) ResumeBeforeMoraleLossPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingMoraleLossCheck || ctx.Timing == TimingBeforeMoraleLoss
}

// ResumeActionEndPhase 行动结束（阶段结束）响应恢复上下文。
func (ctx *Context) ResumeActionEndPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingActionEnd || ctx.Timing == TimingOnActionEnd
}

// TurnStartOrStartupWindow 回合开始或启动技能窗口（弃牌后续恢复等）。
func (ctx *Context) TurnStartOrStartupWindow() bool {
	if ctx == nil {
		return false
	}
	switch ctx.Timing {
	case TimingTurnStart, TimingActionStart, TimingOnTurnStart, TimingStartup:
		return true
	default:
		return false
	}
}
