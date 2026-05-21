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

// runOnTiming 在某个 Timing 窗口触发技能分发。
func (r *Runtime) runOnTiming(h Host, timing model.FlowTiming, ctx *model.Context) {
	if ctx == nil || h == nil {
		return
	}
	ctx.Timing = timing

	var targetsToCheck []checkTarget

	switch timing {
	case model.TimingOnTurnStart, model.TimingStartup:
		currentPid := h.GameState().PlayerOrder[h.GameState().CurrentTurn]
		if player := h.GameState().Players[currentPid]; player != nil {
			targetsToCheck = append(targetsToCheck, checkTarget{Player: player, Role: model.RoleAny})
		}

	case model.TimingOnAttackDeclared, model.TimingOnHitCheck, model.TimingOnDamageCalculated:
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

	case model.TimingOnDamageTaken:
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
		if timing == model.TimingOnFieldMarkChanged {
			for _, p := range h.GameState().Players {
				if p != nil && ctx.User != nil && p.ID != ctx.User.ID {
					targetsToCheck = append(targetsToCheck, checkTarget{Player: p, Role: model.RoleAny})
				}
			}
		}

	case model.TimingOnOrientationChanged:
		for _, pid := range h.GameState().PlayerOrder {
			if player := h.GameState().Players[pid]; player != nil {
				targetsToCheck = append(targetsToCheck, checkTarget{Player: player, Role: model.RoleAny})
			}
		}
		for pid, player := range h.GameState().Players {
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
		if ctx.User != nil {
			orderedPlayers := make([]*model.Player, 0, len(h.GameState().Players))
			seen := make(map[string]struct{}, len(h.GameState().Players))
			for _, pid := range h.GameState().PlayerOrder {
				p := h.GameState().Players[pid]
				if p == nil || p.Camp != ctx.User.Camp {
					continue
				}
				orderedPlayers = append(orderedPlayers, p)
				seen[pid] = struct{}{}
			}
			for pid, p := range h.GameState().Players {
				if p == nil || p.Camp != ctx.User.Camp {
					continue
				}
				if _, ok := seen[pid]; ok {
					continue
				}
				orderedPlayers = append(orderedPlayers, p)
				seen[pid] = struct{}{}
			}
			for _, p := range orderedPlayers {
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

	if timing == model.TimingBeforeMoraleLoss || timing == model.TimingOnDamageTaken {
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
