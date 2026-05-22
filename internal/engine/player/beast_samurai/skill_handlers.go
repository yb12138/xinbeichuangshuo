// gameflow: 兽魂武士 handler。

package beast_samurai

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const (
	beastSamuraiZanshinCap   = 4
	beastSamuraiBeastSoulCap = 2
)

type BeastSamuraiWarriorZanshinHandler struct{ engineplayer.BaseHandler }

type BeastSamuraiOneStrikeNoThoughtHandler struct{ engineplayer.BaseHandler }

type BeastSamuraiBeastSoulWillHandler struct{ engineplayer.BaseHandler }

type BeastSamuraiBeastSoulAlertHandler struct{ engineplayer.BaseHandler }

type BeastSamuraiBeastReturnHandler struct{ engineplayer.BaseHandler }

type BeastSamuraiReversalIaijutsuSlashHandler struct{ engineplayer.BaseHandler }

type BeastSamuraiIaijutsuStyleHandler struct{ engineplayer.BaseHandler }

func beastSamuraiResumePhase(ctx *model.Context) interface{} {
	if ctx == nil || ctx.Selections == nil {
		return nil
	}
	if raw, ok := ctx.Selections["current_resume_point"].(model.TurnStage); ok && raw != "" && model.IsKnownTurnStage(raw) {
		return raw
	}
	if raw, ok := ctx.Selections["current_resume_point"].(model.CombatStage); ok && raw != model.CombatStageNone && model.IsKnownCombatStage(raw) {
		return raw
	}
	if raw, ok := ctx.Selections["current_resume_point"].(model.Subflow); ok && raw != model.SubflowNone && model.IsKnownSubflow(raw) {
		return raw
	}
	if stage, ok := ctx.Selections["current_turn_stage"].(model.TurnStage); ok && stage != "" && model.IsKnownTurnStage(stage) {
		return stage
	}
	if stage, ok := ctx.Selections["current_combat_stage"].(model.CombatStage); ok && stage != model.CombatStageNone && model.IsKnownCombatStage(stage) {
		return stage
	}
	if subflow, ok := ctx.Selections["current_subflow"].(model.Subflow); ok && subflow != model.SubflowNone && model.IsKnownSubflow(subflow) {
		return subflow
	}
	return nil
}

func (h *BeastSamuraiWarriorZanshinHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnActionEnd {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *BeastSamuraiWarriorZanshinHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("武者残心上下文无效")
	}
	now := engineplayer.AddToken(ctx.User, "bs_zanshin", 1, beastSamuraiZanshinCap)
	ctx.Game.Log(fmt.Sprintf("%s 的 [武者残心] 生效：残心+1（当前%d）", ctx.User.Name, now))
	return nil
}

func (h *BeastSamuraiOneStrikeNoThoughtHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnActionEnd {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return engineplayer.GetToken(ctx.User, "bs_zanshin") >= beastSamuraiZanshinCap
}

func (h *BeastSamuraiOneStrikeNoThoughtHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("一击无念上下文无效")
	}
	if engineplayer.GetToken(ctx.User, "bs_zanshin") < beastSamuraiZanshinCap {
		return fmt.Errorf("残心不足4点，无法发动一击无念")
	}
	left := engineplayer.AddToken(ctx.User, "bs_zanshin", -4, beastSamuraiZanshinCap)
	ctx.User.TurnState.UsedSkillCounts["bs_one_strike_armed"] = 1
	model.AppendAttackAction(ctx.User, "一击无念")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [一击无念]：移除4点残心（剩余%d），额外获得1次攻击行动并挂载下次攻击劫持", ctx.User.Name, left))
	return nil
}

func (h *BeastSamuraiBeastSoulWillHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BeastSamuraiBeastSoulWillHandler) Execute(ctx *model.Context) error { return nil }

func (h *BeastSamuraiBeastSoulAlertHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnOrientationChanged {
		return false
	}
	if ctx.EventCtx.OperatorID == "" || ctx.EventCtx.OperatorID == ctx.User.ID {
		return false
	}
	if ctx.EventCtx.NewOrientation != model.OrientationTapped {
		return false
	}
	if InIaijutsuForm(ctx.User) {
		return false
	}
	return engineplayer.GetToken(ctx.User, "bs_beast_soul") >= 1
}

func (h *BeastSamuraiBeastSoulAlertHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return fmt.Errorf("兽魂警戒上下文无效")
	}
	if engineplayer.GetToken(ctx.User, "bs_beast_soul") <= 0 {
		return fmt.Errorf("兽魂不足，无法发动兽魂警戒")
	}
	if InIaijutsuForm(ctx.User) {
		return fmt.Errorf("已处于御魂流居合形态，无法发动兽魂警戒")
	}
	leftSoul := engineplayer.AddToken(ctx.User, "bs_beast_soul", -1, beastSamuraiBeastSoulCap)
	nowZanshin := engineplayer.AddToken(ctx.User, "bs_zanshin", 1, beastSamuraiZanshinCap)
	ctx.User.Orientation = model.OrientationTapped
	ctx.User.Form = "beast_samurai_iaijutsu_form"
	ctx.Game.Log(fmt.Sprintf("%s 发动 [兽魂警戒]：移除1点兽魂（剩余%d），残心+1（当前%d），进入御魂流居合形态", ctx.User.Name, leftSoul, nowZanshin))

	// 令触发横置的角色（EventCtx.OperatorID）展示并弃 1 张手牌，不由兽灵武士另选目标。
	operatorID := ctx.EventCtx.OperatorID
	if operatorID == "" || operatorID == ctx.User.ID {
		return nil
	}
	var operator *model.Player
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.ID == operatorID {
			operator = p
			break
		}
	}
	if operator == nil || len(operator.Hand) == 0 {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: operatorID,
		Context: map[string]interface{}{
			"choice_type":     "bs_alert_source_discard",
			"user_id":         ctx.User.ID,
			"actor_id":        operatorID,
			"discard_count":   1,
			"discard_subflow": true,
			"prompt":          "【兽魂警戒】请选择并展示弃置1张手牌：",
			"resume_phase":    beastSamuraiResumePhase(ctx),
		},
	})
	return nil
}

func (h *BeastSamuraiBeastReturnHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return false
	}
	if !ctx.DamageTakenPhase() {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	if ctx.Target == nil || ctx.Target.ID == ctx.User.ID {
		return false
	}
	return engineplayer.GetToken(ctx.User, "bs_beast_soul") > 0
}

func (h *BeastSamuraiBeastReturnHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.Target == nil {
		return fmt.Errorf("兽返上下文无效")
	}
	maxX := engineplayer.GetToken(ctx.User, "bs_beast_soul")
	if maxX <= 0 {
		return fmt.Errorf("兽返需要至少1点兽魂")
	}
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
	ctx.Game.Log(fmt.Sprintf("%s 发动 [兽返]：请选择X（1~%d）", ctx.User.Name, maxX))
	return nil
}

func (h *BeastSamuraiReversalIaijutsuSlashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return false
	}
	if !ctx.AttackHitPhase() {
		return false
	}
	if ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.ActionType != string(model.ActionAttack) {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
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
	maxX := engineplayer.GetToken(ctx.User, "bs_beast_soul")
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
