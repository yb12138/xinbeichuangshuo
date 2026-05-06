package model

// Context 上 Timing-first 的语义查询。

// AttackDeclaredPhase 主动攻击宣告/攻击开始窗口（TimingOnAttackDeclared）。
func (ctx *Context) AttackDeclaredPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingOnAttackDeclared
}

// BeforeDrawPhase 表示“摸牌前”窗口。
func (ctx *Context) BeforeDrawPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingBeforeCardDrawn
}

// AfterDrawPhase 表示“摸牌后”窗口。
func (ctx *Context) AfterDrawPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingOnCardDrawn
}

// ResumeAttackHitPhase 响应恢复：攻击命中分支（TimingOnHitCheck 下需配合 AttackInfo.IsHit）。
func (ctx *Context) ResumeAttackHitPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingOnHitCheck && ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.IsHit
}

// ResumeAttackMissPhase 响应恢复：攻击未命中分支。
func (ctx *Context) ResumeAttackMissPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingOnHitCheck && ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil && !ctx.EventCtx.AttackInfo.IsHit
}

// ResumeDamageTakenPhase 承伤响应恢复上下文。
func (ctx *Context) ResumeDamageTakenPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingOnDamageTaken
}

// ResumeBeforeMoraleLossPhase 士气下降前响应恢复上下文。
func (ctx *Context) ResumeBeforeMoraleLossPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingBeforeMoraleLoss
}

// ResumeActionEndPhase 行动结束（阶段结束）响应恢复上下文。
func (ctx *Context) ResumeActionEndPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingOnActionEnd
}

// TurnStartOrStartupWindow 回合开始或启动技能窗口（弃牌后续恢复等）。
func (ctx *Context) TurnStartOrStartupWindow() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingOnTurnStart || ctx.Timing == TimingStartup
}
