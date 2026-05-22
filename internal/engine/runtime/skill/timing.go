// gameflow: FlowTiming 窗口内要检查的玩家与身份，以及 OnTiming 主循环。

package skill

import (
	"sort"

	"starcup-engine/internal/model"
)

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

type timingStateHost interface {
	GameState() *model.GameState
}

// runOnTiming 在某个 Timing 窗口触发技能分发。
func (r *Runtime) runOnTiming(h Host, timing model.FlowTiming, ctx *model.Context) {
	if ctx == nil || h == nil {
		return
	}
	ctx.Timing = timing

	targetsToCheck := targetsForTiming(h, timing, ctx)

	if timingPriorityOrdered(timing) {
		for _, item := range r.collectTargetsWithSkillsByPriority(h, targetsToCheck, timing, ctx) {
			r.trig.ProcessSkillBatch(h, item.skills, item.ctx)
		}
		return
	}

	for _, target := range targetsToCheck {
		if target.Player == nil || target.Player.Character == nil {
			continue
		}
		skillCtx := *ctx
		skillCtx.User = target.Player

		skills := r.elig.CollectCandidates(target.Player, timing, &skillCtx, target.Role)
		r.trig.ProcessSkillBatch(h, skills, &skillCtx)
	}
}

func targetsForTiming(h timingStateHost, timing model.FlowTiming, ctx *model.Context) []checkTarget {
	if h == nil || h.GameState() == nil || ctx == nil {
		return nil
	}
	var targetsToCheck []checkTarget
	switch timing {
	case model.TimingOnTurnStart, model.TimingStartup:
		return currentPlayerTarget(h)

	case model.TimingOnFieldMarkChanged:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: ctx.User, Role: model.RoleAny})
		}
		for _, p := range h.GameState().Players {
			if p != nil && ctx.User != nil && p.ID != ctx.User.ID {
				targetsToCheck = append(targetsToCheck, checkTarget{Player: p, Role: model.RoleAny})
			}
		}
		return targetsToCheck

	case model.TimingOnOrientationChanged:
		return allPlayersInSeatOrder(h)
	}

	switch timing {
	case model.TimingTurnStart, model.TimingActionStart:
		return currentPlayerTarget(h)

	case model.TimingAttackDeclare,
		model.TimingAttackSelectTarget,
		model.TimingAttackPlayCard,
		model.TimingAttackModifyCard,
		model.TimingAttackCommitted,
		model.TimingAttackForceHitCheck,
		model.TimingAttackNoResponseCheck,
		model.TimingAttackResponse,
		model.TimingAttackHit,
		model.TimingAttackMiss,
		model.TimingDamageSourceDeal:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.User,
				Role:   model.RoleAttacker,
			})
		}
		if ctx.Target != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.Target,
				Role:   model.RoleDefender,
			})
		}

	case model.TimingDamageTaken:
		if ctx.User != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.User,
				Role:   model.RoleDefender,
			})
		}
		if ctx.Target != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{
				Player: ctx.Target,
				Role:   model.RoleAttacker,
			})
		}

	case model.TimingMoraleLossCheck:
		if ctx.User != nil {
			for _, p := range campPlayersInSeatOrder(h, ctx.User.Camp) {
				targetsToCheck = append(targetsToCheck, checkTarget{
					Player: p,
					Role:   model.RoleAny,
				})
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
	return targetsToCheck
}

func currentPlayerTarget(h timingStateHost) []checkTarget {
	if h == nil || h.GameState() == nil {
		return nil
	}
	gs := h.GameState()
	if gs.CurrentTurn < 0 || gs.CurrentTurn >= len(gs.PlayerOrder) {
		return nil
	}
	currentPid := gs.PlayerOrder[gs.CurrentTurn]
	if player := gs.Players[currentPid]; player != nil {
		return []checkTarget{{Player: player, Role: model.RoleAny}}
	}
	return nil
}

func allPlayersInSeatOrder(h timingStateHost) []checkTarget {
	if h == nil || h.GameState() == nil {
		return nil
	}
	gs := h.GameState()
	targets := make([]checkTarget, 0, len(gs.Players))
	seen := make(map[string]struct{}, len(gs.Players))
	for _, pid := range gs.PlayerOrder {
		player := gs.Players[pid]
		if player == nil {
			continue
		}
		targets = append(targets, checkTarget{Player: player, Role: model.RoleAny})
		seen[pid] = struct{}{}
	}
	for pid, player := range gs.Players {
		if player == nil {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		targets = append(targets, checkTarget{Player: player, Role: model.RoleAny})
	}
	return targets
}

func campPlayersInSeatOrder(h timingStateHost, camp model.Camp) []*model.Player {
	if h == nil || h.GameState() == nil {
		return nil
	}
	gs := h.GameState()
	orderedPlayers := make([]*model.Player, 0, len(gs.Players))
	seen := make(map[string]struct{}, len(gs.Players))
	for _, pid := range gs.PlayerOrder {
		p := gs.Players[pid]
		if p == nil || p.Camp != camp {
			continue
		}
		orderedPlayers = append(orderedPlayers, p)
		seen[pid] = struct{}{}
	}
	for pid, p := range gs.Players {
		if p == nil || p.Camp != camp {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		orderedPlayers = append(orderedPlayers, p)
		seen[pid] = struct{}{}
	}
	return orderedPlayers
}

func timingPriorityOrdered(timing model.FlowTiming) bool {
	return timing == model.TimingMoraleLossCheck || timing == model.TimingDamageTaken
}

func (r *Runtime) collectTargetsWithSkillsByPriority(h Host, targets []checkTarget, timing model.FlowTiming, ctx *model.Context) []targetSkillsBatch {
	if len(targets) == 0 || ctx == nil || h == nil {
		return nil
	}
	gs := h.GameState()
	if gs == nil {
		return nil
	}

	seatOrders := make(map[string]int, len(gs.PlayerOrder))
	for idx, pid := range gs.PlayerOrder {
		seatOrders[pid] = idx
	}

	items := make([]targetSkillsBatch, 0, len(targets))
	for _, target := range targets {
		if target.Player == nil || target.Player.Character == nil {
			continue
		}
		skillCtx := *ctx
		skillCtx.User = target.Player
		activated := r.elig.CollectCandidates(target.Player, timing, &skillCtx, target.Role)
		if len(activated) == 0 {
			continue
		}
		maxPriority := 0
		for _, sk := range activated {
			if sk.Priority > maxPriority {
				maxPriority = sk.Priority
			}
		}
		seat, ok := seatOrders[target.Player.ID]
		if !ok {
			seat = len(gs.PlayerOrder) + len(items)
		}
		items = append(items, targetSkillsBatch{
			target:    target,
			ctx:       &skillCtx,
			skills:    activated,
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
