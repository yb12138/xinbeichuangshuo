package model

// Context 上 Timing-first 的语义查询。

func (ctx *Context) timingPhase(rulebook Timing, aliases ...Timing) bool {
	if ctx == nil {
		return false
	}
	if ctx.Timing == rulebook {
		return true
	}
	for _, t := range aliases {
		if ctx.Timing == t {
			return true
		}
	}
	return false
}

// AttackDeclarePhase 主动攻击宣告/攻击开始窗口。
func (ctx *Context) AttackDeclarePhase() bool {
	return ctx.timingPhase(TimingAttackDeclare)
}

// AttackDeclaredPhase 主动攻击宣告/攻击开始窗口。
func (ctx *Context) AttackDeclaredPhase() bool {
	return ctx.AttackDeclarePhase()
}

// AttackSelectTargetPhase 表示主动攻击选择目标窗口。
func (ctx *Context) AttackSelectTargetPhase() bool {
	return ctx.timingPhase(TimingAttackSelectTarget)
}

// AttackPlayCardPhase 表示主动攻击出牌窗口。
func (ctx *Context) AttackPlayCardPhase() bool {
	return ctx.timingPhase(TimingAttackPlayCard)
}

// AttackModifyCardPhase 表示主动攻击牌面修正窗口。
func (ctx *Context) AttackModifyCardPhase() bool {
	return ctx.timingPhase(TimingAttackModifyCard)
}

// AttackCommittedPhase 表示一次主动/应战攻击已经完成阶段①提交。
func (ctx *Context) AttackCommittedPhase() bool {
	return ctx.timingPhase(TimingAttackCommitted)
}

// AttackForceHitCheckPhase 表示强制命中检查窗口。
func (ctx *Context) AttackForceHitCheckPhase() bool {
	return ctx.timingPhase(TimingAttackForceHitCheck)
}

// AttackNoResponseCheckPhase 表示不可响应/免响应检查窗口。
func (ctx *Context) AttackNoResponseCheckPhase() bool {
	return ctx.timingPhase(TimingAttackNoResponseCheck)
}

// AttackResponsePhase 表示攻击响应窗口。
func (ctx *Context) AttackResponsePhase() bool {
	return ctx.timingPhase(TimingAttackResponse)
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
	return ctx.timingPhase(TimingAttackHit)
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
	return ctx.timingPhase(TimingAttackMiss)
}

// ResumeAttackMissPhase 响应恢复：攻击未命中分支。
func (ctx *Context) ResumeAttackMissPhase() bool {
	return ctx.AttackMissPhase()
}

// MagicDeclarePhase 表示主动法术宣告窗口。
func (ctx *Context) MagicDeclarePhase() bool {
	return ctx.timingPhase(TimingMagicDeclare, TimingOnMagicDeclared)
}

// MagicResolvePhase 表示主动法术进入效果结算。
func (ctx *Context) MagicResolvePhase() bool {
	return ctx.timingPhase(TimingMagicResolve, TimingOnMagicDeclared)
}

// MagicMissileResponsePhase 表示魔弹基础响应窗口。
func (ctx *Context) MagicMissileResponsePhase() bool {
	return ctx.timingPhase(TimingMagicMissileResponse)
}

// MagicMissileDefendPhase 表示魔弹防御判定窗口。
func (ctx *Context) MagicMissileDefendPhase() bool {
	return ctx.timingPhase(TimingMagicMissileDefend, Timing("on_magic_missile_defend"))
}

// MagicMissileCounterPhase 表示魔弹传递/反击判定窗口。
func (ctx *Context) MagicMissileCounterPhase() bool {
	return ctx.timingPhase(TimingMagicMissileCounter, Timing("on_magic_missile_counter"))
}

// MagicMissileResponseSkillPhase 表示魔弹响应技能窗口。
func (ctx *Context) MagicMissileResponseSkillPhase() bool {
	return ctx.timingPhase(TimingMagicMissileResponseSkill, Timing("on_magic_missile_response_skill_aug"))
}

// DamageSourceDealPhase 表示伤害来源造成伤害窗口。
func (ctx *Context) DamageSourceDealPhase() bool {
	return ctx.timingPhase(TimingDamageSourceDeal)
}

// DamageTargetBeforePhase 表示伤害目标承伤前窗口。
func (ctx *Context) DamageTargetBeforePhase() bool {
	return ctx.timingPhase(TimingDamageTargetBefore, Timing("on_damage_before_taken"))
}

// DamageAppliedPhase 表示伤害已应用窗口。
func (ctx *Context) DamageAppliedPhase() bool {
	return ctx.timingPhase(TimingDamageApplied, TimingOnDamageApplied, Timing("on_damage_applied"))
}

// DamageTakenPhase 表示伤害目标已承伤窗口。
func (ctx *Context) DamageTakenPhase() bool {
	return ctx.timingPhase(TimingDamageTaken)
}

// DamageResolvedPhase 表示伤害流程完成窗口。
func (ctx *Context) DamageResolvedPhase() bool {
	return ctx.timingPhase(TimingDamageResolved, Timing("post_damage_resolved"), Timing("on_damage_after_apply"))
}

// ResumeDamageTakenPhase 承伤响应恢复上下文。
func (ctx *Context) ResumeDamageTakenPhase() bool {
	return ctx.DamageTakenPhase()
}

// ResumeBeforeMoraleLossPhase 士气下降前响应恢复上下文。
func (ctx *Context) ResumeBeforeMoraleLossPhase() bool {
	if ctx == nil {
		return false
	}
	return ctx.Timing == TimingMoraleLossCheck
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
