package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

const (
	beastSamuraiZanshinCap   = 4
	beastSamuraiBeastSoulCap = 2
)

type BeastSamuraiWarriorZanshinHandler struct{ BaseHandler }

type BeastSamuraiOneStrikeNoThoughtHandler struct{ BaseHandler }

type BeastSamuraiOneStrikeInterceptHandler struct{ BaseHandler }

type BeastSamuraiBeastSoulWillHandler struct{ BaseHandler }

type BeastSamuraiBeastSoulAlertHandler struct{ BaseHandler }

type BeastSamuraiBeastReturnHandler struct{ BaseHandler }

type BeastSamuraiIaijutsuTurnEndDrainHandler struct{ BaseHandler }

type BeastSamuraiIaijutsuExitOnDealDamageHandler struct{ BaseHandler }

type BeastSamuraiIaijutsuExitOnZeroHandler struct{ BaseHandler }

type BeastSamuraiIaijutsuTappedBoostHandler struct{ BaseHandler }

type BeastSamuraiReversalIaijutsuSlashHandler struct{ BaseHandler }

type BeastSamuraiIaijutsuStyleHandler struct{ BaseHandler }

func beastSamuraiResumePhase(ctx *model.Context) string {
	if ctx == nil || ctx.Selections == nil {
		return ""
	}
	if raw, ok := ctx.Selections["current_resume_point"].(string); ok {
		return model.NormalizeResumePoint(raw)
	}
	if raw, ok := ctx.Selections["current_turn_stage"].(string); ok {
		if stage := model.TurnStage(raw); model.IsKnownTurnStage(stage) {
			return model.NormalizeResumePoint(stage)
		}
	}
	if raw, ok := ctx.Selections["current_combat_stage"].(string); ok {
		if stage := model.CombatStage(raw); model.IsKnownCombatStage(stage) {
			return model.NormalizeResumePoint(stage)
		}
	}
	if raw, ok := ctx.Selections["current_subflow"].(string); ok {
		if subflow := model.Subflow(raw); model.IsKnownSubflow(subflow) {
			return model.NormalizeResumePoint(subflow)
		}
	}
	return ""
}

func (h *BeastSamuraiWarriorZanshinHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnPhaseEnd {
		return false
	}
	if ctx.TriggerCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *BeastSamuraiWarriorZanshinHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("武者残心上下文无效")
	}
	now := addToken(ctx.User, "bs_zanshin", 1, 0, beastSamuraiZanshinCap)
	ctx.Game.Log(fmt.Sprintf("%s 的 [武者残心] 生效：残心+1（当前%d）", ctx.User.Name, now))
	return nil
}

func (h *BeastSamuraiOneStrikeNoThoughtHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnPhaseEnd {
		return false
	}
	if ctx.TriggerCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return getToken(ctx.User, "bs_zanshin") >= beastSamuraiZanshinCap
}

func (h *BeastSamuraiOneStrikeNoThoughtHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("一击无念上下文无效")
	}
	if getToken(ctx.User, "bs_zanshin") < beastSamuraiZanshinCap {
		return fmt.Errorf("残心不足4点，无法发动一击无念")
	}
	left := addToken(ctx.User, "bs_zanshin", -4, 0, beastSamuraiZanshinCap)
	ctx.User.TurnState.UsedSkillCounts["bs_one_strike_armed"] = 1
	addAttackAction(ctx.User, "一击无念")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [一击无念]：移除4点残心（剩余%d），额外获得1次攻击行动并挂载下次攻击劫持", ctx.User.Name, left))
	return nil
}

func (h *BeastSamuraiOneStrikeInterceptHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BeastSamuraiOneStrikeInterceptHandler) Execute(ctx *model.Context) error { return nil }

func (h *BeastSamuraiBeastSoulWillHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BeastSamuraiBeastSoulWillHandler) Execute(ctx *model.Context) error { return nil }

func (h *BeastSamuraiBeastSoulAlertHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnOrientationChanged {
		return false
	}
	if ctx.TriggerCtx.OperatorID == "" || ctx.TriggerCtx.OperatorID == ctx.User.ID {
		return false
	}
	if ctx.TriggerCtx.NewOrientation != model.OrientationTapped {
		return false
	}
	return getToken(ctx.User, "bs_beast_soul") >= 1
}

func (h *BeastSamuraiBeastSoulAlertHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return fmt.Errorf("兽魂警戒上下文无效")
	}
	actorID := ctx.TriggerCtx.OperatorID
	if actorID == "" {
		return fmt.Errorf("兽魂警戒缺少触发角色")
	}
	actor := ctx.Target
	if actor == nil || actor.ID != actorID {
		for _, p := range ctx.Game.GetAllPlayers() {
			if p != nil && p.ID == actorID {
				actor = p
				break
			}
		}
	}
	if getToken(ctx.User, "bs_beast_soul") <= 0 {
		return fmt.Errorf("兽魂不足，无法发动兽魂警戒")
	}
	leftSoul := addToken(ctx.User, "bs_beast_soul", -1, 0, beastSamuraiBeastSoulCap)
	nowZanshin := addToken(ctx.User, "bs_zanshin", 1, 0, beastSamuraiZanshinCap)
	ctx.User.Orientation = model.OrientationTapped
	ctx.User.Form = "beast_samurai_iaijutsu_form"
	ctx.Game.Log(fmt.Sprintf("%s 发动 [兽魂警戒]：移除1点兽魂（剩余%d），残心+1（当前%d），进入御魂流居合形态", ctx.User.Name, leftSoul, nowZanshin))
	if actor == nil || len(actor.Hand) == 0 {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptDiscard,
		PlayerID: actor.ID,
		Context: map[string]interface{}{
			"choice_type":   "bs_alert_source_discard",
			"user_id":       ctx.User.ID,
			"actor_id":      actor.ID,
			"discard_count": 1,
			"prompt":        "【兽魂警戒】请选择并展示弃置1张手牌：",
			"resume_phase":  beastSamuraiResumePhase(ctx),
		},
	})
	return nil
}

func (h *BeastSamuraiBeastReturnHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnDamageTaken {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	if ctx.Target == nil || ctx.Target.ID == ctx.User.ID {
		return false
	}
	return true
}

func (h *BeastSamuraiBeastReturnHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.Target == nil {
		return fmt.Errorf("兽返上下文无效")
	}
	maxX := getToken(ctx.User, "bs_beast_soul")
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":  "bs_beast_return_x",
			"user_id":      ctx.User.ID,
			"source_id":    ctx.Target.ID,
			"max_x":        maxX,
			"resume_phase": beastSamuraiResumePhase(ctx),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [兽返]：请选择X（0~%d）", ctx.User.Name, maxX))
	return nil
}

func (h *BeastSamuraiIaijutsuTurnEndDrainHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BeastSamuraiIaijutsuTurnEndDrainHandler) Execute(ctx *model.Context) error { return nil }

func (h *BeastSamuraiIaijutsuExitOnDealDamageHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BeastSamuraiIaijutsuExitOnDealDamageHandler) Execute(ctx *model.Context) error { return nil }

func (h *BeastSamuraiIaijutsuExitOnZeroHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BeastSamuraiIaijutsuExitOnZeroHandler) Execute(ctx *model.Context) error { return nil }

func (h *BeastSamuraiIaijutsuTappedBoostHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BeastSamuraiIaijutsuTappedBoostHandler) Execute(ctx *model.Context) error { return nil }

func (h *BeastSamuraiReversalIaijutsuSlashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackHit {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if ctx.User.Form != "beast_samurai_iaijutsu_form" {
		return false
	}
	target := ctx.Target
	if target == nil {
		return false
	}
	return len(target.Hand) < 4
}

func (h *BeastSamuraiReversalIaijutsuSlashHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.Target == nil {
		return fmt.Errorf("逆反居合斩上下文无效")
	}
	maxX := getToken(ctx.User, "bs_beast_soul")
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":  "bs_reversal_x",
			"user_id":      ctx.User.ID,
			"target_id":    ctx.Target.ID,
			"max_x":        maxX,
			"user_ctx":     ctx,
			"resume_phase": beastSamuraiResumePhase(ctx),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [逆反居合斩]：请选择X（0~%d）", ctx.User.Name, maxX))
	return nil
}

func (h *BeastSamuraiIaijutsuStyleHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && ctx.User.Gem >= 1
}

func (h *BeastSamuraiIaijutsuStyleHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("御魂流居合式上下文无效")
	}
	modes := []int{0}
	if len(ctx.User.Hand) > 0 {
		modes = append(modes, 1)
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":  "bs_iaijutsu_style_mode",
			"user_id":      ctx.User.ID,
			"modes":        modes,
			"resume_phase": beastSamuraiResumePhase(ctx),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [御魂流居合式]：请选择摸1或弃1", ctx.User.Name))
	return nil
}
