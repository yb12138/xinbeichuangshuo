package engine

import (
	"sort"
	"starcup-engine/internal/model"
)

// OnTiming 在某个 Timing 窗口触发技能分发。
// 流程：先按规则收集本时机需要检查的玩家，再按技能可触发条件执行或挂起中断。
func (sd *SkillDispatcher) OnTiming(timing model.TriggerTiming, ctx *model.Context) {
	ctx.Timing = timing

	// 1. 收集触发的技能
	// 使用 checkTarget 结构来明确当前检查的玩家是什么身份
	var targetsToCheck []checkTarget

	switch timing {
	case model.TimingOnTurnStart, model.TimingStartup:
		// 回合开始：只检查当前玩家
		currentPid := sd.engine.State.PlayerOrder[sd.engine.State.CurrentTurn]
		if player := sd.engine.State.Players[currentPid]; player != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: player, Role: model.RoleAny})
		}

	case model.TimingOnAttackDeclared, model.TimingOnHitCheck, model.TimingOnDamageCalculated:
		// 上下文中的 User 是攻击发起者 -> 身份标记为 Attacker
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.User,
				Role:   model.RoleAttacker,
			})
		}
		// 上下文中的 Target 是受击者 -> 身份标记为 Defender
		if ctx.Target != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.Target,
				Role:   model.RoleDefender,
			})
		}

	case model.TimingOnDamageTaken:
		// 在 combat.go 的 handleTakeHit 中，TriggerOnDamageTaken 的 ctx.User 是受害者
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.User,
				Role:   model.RoleDefender,
			})
		}
		// 允许“造成伤害后触发”的技能（如元素吸收）以攻击者身份检查
		if ctx.Target != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.Target,
				Role:   model.RoleAttacker,
			})
		}

	case model.TimingOnCardPlayedOrRevealed:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.User, Role: model.RoleAny})
		}

	case model.TimingBeforeCardDrawn, model.TimingOnCardDrawn:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.User, Role: model.RoleAny})
		}

	case model.TimingOnActionEnd, model.TimingOnFieldMarkChanged:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.User, Role: model.RoleAny})
		}
		// 对于 FieldMarkChanged，可能需要检查全场玩家的被动技能 (如天使羁绊)
		if timing == model.TimingOnFieldMarkChanged {
			for _, p := range sd.engine.State.Players {
				if p.ID != ctx.User.ID {
					targetsToCheck = append(targetsToCheck, checkTarget{Player: p, Role: model.RoleAny})
				}
			}
		}

	case model.TimingOnOrientationChanged:
		for _, pid := range sd.engine.State.PlayerOrder {
			if player := sd.engine.State.Players[pid]; player != nil {
				targetsToCheck = append(targetsToCheck, checkTarget{Player: player, Role: model.RoleAny})
			}
		}
		for pid, player := range sd.engine.State.Players {
			if player == nil {
				continue
			}
			seen := false
			for _, queued := range targetsToCheck {
				if queued.Player != nil && queued.Player.ID == pid {
					seen = true
					break
				}
			}
			if !seen {
				targetsToCheck = append(targetsToCheck, checkTarget{Player: player, Role: model.RoleAny})
			}
		}

	case model.TimingBeforeMoraleLoss:
		// 上下文的 User 是导致士气下降的受害者 (Victim)
		if ctx.User != nil {
			// TriggerBeforeMoraleLoss：先按座次收集同阵营目标，后续再按技能优先级排序执行。
			orderedPlayers := make([]*model.Player, 0, len(sd.engine.State.Players))
			seen := make(map[string]struct{}, len(sd.engine.State.Players))
			for _, pid := range sd.engine.State.PlayerOrder {
				p := sd.engine.State.Players[pid]
				if p == nil || p.Camp != ctx.User.Camp {
					continue
				}
				orderedPlayers = append(orderedPlayers, p)
				seen[pid] = struct{}{}
			}
			// 兜底：补齐不在 PlayerOrder 的同阵营玩家（通常不会发生）。
			for pid, p := range sd.engine.State.Players {
				if p == nil || p.Camp != ctx.User.Camp {
					continue
				}
				if _, ok := seen[pid]; ok {
					continue
				}
				orderedPlayers = append(orderedPlayers, p)
				seen[pid] = struct{}{}
			}

			appendTarget := func(p *model.Player) {
				targetsToCheck = append(targetsToCheck, checkTarget{
					Player: p,
					Role:   model.RoleAny,
				})
			}
			for _, p := range orderedPlayers {
				appendTarget(p)
			}
		}

	default:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.User, Role: model.RoleAny})
		}
		if ctx.Target != nil && ctx.Target != ctx.User {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.Target, Role: model.RoleAny})
		}
	}

	// 同一时点按技能优先级（高->低）执行；同优先级按座次稳定。
	if timing == model.TimingBeforeMoraleLoss {
		for _, item := range sd.collectTargetsWithSkillsByPriority(targetsToCheck, timing, ctx) {
			sd.processSkills(item.skills, item.ctx)
		}
		return
	}

	// 3. 执行检查
	for _, target := range targetsToCheck {
		if target.Player == nil || target.Player.Character == nil {
			continue
		}
		// 创建上下文副本，确保 User 是当前技能持有者
		skillCtx := *ctx
		skillCtx.User = target.Player

		// 【关键】传入当前玩家在事件中的角色
		skills := sd.collectTriggeredSkills(target.Player, timing, &skillCtx, target.Role)
		sd.processSkills(skills, &skillCtx)
	}
}

func (sd *SkillDispatcher) collectTargetsWithSkillsByPriority(targets []checkTarget, timing model.TriggerTiming, ctx *model.Context) []targetTriggeredSkills {
	if len(targets) == 0 || ctx == nil || sd == nil || sd.engine == nil || sd.engine.State == nil {
		return nil
	}

	seatOrders := make(map[string]int, len(sd.engine.State.PlayerOrder))
	for idx, pid := range sd.engine.State.PlayerOrder {
		seatOrders[pid] = idx
	}

	items := make([]targetTriggeredSkills, 0, len(targets))
	for _, target := range targets {
		if target.Player == nil || target.Player.Character == nil {
			continue
		}
		skillCtx := *ctx
		skillCtx.User = target.Player
		triggered := sd.collectTriggeredSkills(target.Player, timing, &skillCtx, target.Role)
		if len(triggered) == 0 {
			continue
		}
		maxPriority := 0
		for _, skill := range triggered {
			if skill.Priority > maxPriority {
				maxPriority = skill.Priority
			}
		}
		seat, ok := seatOrders[target.Player.ID]
		if !ok {
			seat = len(sd.engine.State.PlayerOrder) + len(items)
		}
		items = append(items, targetTriggeredSkills{
			target:    target,
			ctx:       &skillCtx,
			skills:    triggered,
			priority:  maxPriority,
			seatOrder: seat,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].priority != items[j].priority {
			return items[i].priority > items[j].priority
		}
		if items[i].seatOrder != items[j].seatOrder {
			return items[i].seatOrder < items[j].seatOrder
		}
		return items[i].target.Player.ID < items[j].target.Player.ID
	})
	return items
}

// OnTrigger 兼容入口：迁移期保留，内部统一转 Timing。
func (sd *SkillDispatcher) OnTrigger(trigger model.TriggerType, ctx *model.Context) {
	if ctx == nil {
		return
	}
	prevTrigger := ctx.Trigger
	ctx.Trigger = trigger
	if ctx.Timing == model.TimingUnknown || prevTrigger != trigger {
		ctx.Timing = model.LegacyTriggerToTiming(trigger)
	}
	sd.OnTiming(ctx.Timing, ctx)
}
