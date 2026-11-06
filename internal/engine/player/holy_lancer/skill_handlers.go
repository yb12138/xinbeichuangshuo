// gameflow: 圣枪骑士技能处理器。

package holy_lancer

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"

	"starcup-engine/internal/model"
)

// --- 圣枪骑士技能处理器 ---

type HolyLancerRevelationHandler struct{ engineplayer.BaseHandler }

func (h *HolyLancerRevelationHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.Game == nil || ctx.User == nil {
		return nil
	}
	ctx.Game.RefreshPlayerDerivedState(ctx.User.ID)
	return nil
}

type HolyLancerRadianceHandler struct{ engineplayer.BaseHandler }

func (h *HolyLancerRadianceHandler) Execute(ctx *model.Context) error {
	for _, p := range ctx.Game.GetAllPlayers() {
		ctx.Game.Heal(p.ID, 1)
	}
	model.AppendAttackAction(ctx.User, "辉耀")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [辉耀]，全场+1治疗并获得额外攻击行动", ctx.User.Name))
	return nil
}

type HolyLancerPunishmentHandler struct{ engineplayer.BaseHandler }

func (h *HolyLancerPunishmentHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("惩戒需要目标")
	}
	if ctx.Target.ID == ctx.User.ID {
		return fmt.Errorf("惩戒目标必须是其他角色")
	}
	if ctx.Target.Heal <= 0 {
		return fmt.Errorf("惩戒目标没有治疗，无法发动")
	}
	ctx.Target.Heal--
	if ctx.User.Heal < ctx.User.MaxHeal {
		ctx.User.Heal++
	}
	model.AppendAttackAction(ctx.User, "惩戒")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [惩戒]，从 %s 转移1点治疗并获得额外攻击行动", ctx.User.Name, ctx.Target.Name))
	return nil
}

type HolyLancerHolyStrikeHandler struct{ engineplayer.BaseHandler }

func (h *HolyLancerHolyStrikeHandler) CanUse(ctx *model.Context) bool {
	// 与地枪互斥：
	// 仅当本次命中确实会弹出地枪响应窗口时，才把圣击延后到"跳过地枪"时结算；
	// 地枪不会被提供时（未持有/条件不满足），圣击在命中时直接结算，避免 +1 丢失。
	if ctx == nil || ctx.User == nil || !ctx.AttackHitPhase() || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	if info.ActionType != string(model.ActionAttack) {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["holy_lancer_block_sacred_strike"] != 0 {
		return false
	}
	return !earthSpearWillBeOffered(ctx)
}

func (h *HolyLancerHolyStrikeHandler) Execute(ctx *model.Context) error {
	// 静默结算时地枪必然未被提供，"跳过地枪"补触发路径不会再 +1，无需额外标记。
	ctx.Game.Heal(ctx.User.ID, 1)
	return nil
}

// earthSpearWillBeOffered 判断本次命中是否会提供地枪响应（持有地枪且满足发动条件）。
func earthSpearWillBeOffered(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.User.Character == nil {
		return false
	}
	hasSkill := false
	for _, sk := range ctx.User.Character.Skills {
		if sk.ID == "holy_lancer_earth_spear" {
			hasSkill = true
			break
		}
	}
	if !hasSkill {
		return false
	}
	return (&HolyLancerEarthSpearHandler{}).CanUse(ctx)
}

type HolyLancerSkySpearHandler struct{ engineplayer.BaseHandler }

func (h *HolyLancerSkySpearHandler) CanUse(ctx *model.Context) bool {
	if ctx.User.Heal < 2 {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["holy_lancer_prayer"] > 0 {
		return false
	}
	if ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	return ctx.EventCtx.AttackInfo.CounterInitiator == ""
}

func (h *HolyLancerSkySpearHandler) Execute(ctx *model.Context) error {
	ctx.User.Heal -= 2
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		ctx.EventCtx.AttackInfo.CanBeResponded = false
	}
	// 通过令牌持久化"本次攻击无法应战"，由 attackGatingHook 在流控阶段应用到 CombatRequest。
	ctx.User.TurnState.UsedSkillCounts["holy_lancer_sky_spear_no_counter"] = 1
	ctx.User.TurnState.UsedSkillCounts["holy_lancer_block_sacred_strike"] = 1
	ctx.Game.Log(fmt.Sprintf("%s 发动 [天枪]，移除2治疗，本次攻击不可应战", ctx.User.Name))
	return nil
}

type HolyLancerEarthSpearHandler struct{ engineplayer.BaseHandler }

func (h *HolyLancerEarthSpearHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if !ctx.AttackHitPhase() {
		return false
	}
	if ctx.User.Heal <= 0 || ctx.EventCtx.DamageVal == nil {
		return false
	}
	// 地枪仅可在主动攻击命中后发动。
	if ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.ActionType != string(model.ActionAttack) {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *HolyLancerEarthSpearHandler) Execute(ctx *model.Context) error {
	if ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return nil
	}
	x := ctx.User.Heal
	if x > 4 {
		x = 4
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "holy_lancer_earth_spear_x",
			"user_id":     ctx.User.ID,
			"max_x":       x,
			"user_ctx":    ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [地枪]，等待选择X值", ctx.User.Name))
	return nil
}

type HolyLancerPrayerHandler struct{ engineplayer.BaseHandler }

func (h *HolyLancerPrayerHandler) CanUse(ctx *model.Context) bool {
	return ctx.User.Gem > 0
}

func (h *HolyLancerPrayerHandler) Execute(ctx *model.Context) error {
	ctx.User.Heal += 2
	if ctx.User.Heal > 5 {
		ctx.User.Heal = 5
	}
	ctx.User.TurnState.UsedSkillCounts["holy_lancer_prayer"] = 1
	model.AppendAttackAction(ctx.User, "圣光祈愈")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣光祈愈]，治疗+2（上限5）并获得额外攻击行动", ctx.User.Name))
	return nil
}
