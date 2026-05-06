// gameflow: 血色剑灵技能处理器。

package crimson_sword_spirit

import (
	"fmt"
	"strings"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type BaseHandler = engineplayer.BaseHandler

// --- Crimson Sword Spirit Skill Handlers ---

type CrimsonBloodThornsHandler struct{ BaseHandler }

func (h *CrimsonBloodThornsHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.Timing != model.TimingOnHitCheck || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	return info.ActionType == string(model.ActionAttack) && info.IsHit
}

func (h *CrimsonBloodThornsHandler) Execute(ctx *model.Context) error {
	cur := addBlood(ctx.User, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [血色荆棘] 触发，鲜血=%d", ctx.User.Name, cur))
	return nil
}

type CrimsonFlashHandler struct{ BaseHandler }

func (h *CrimsonFlashHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnActionEnd || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return engineplayer.GetToken(ctx.User, "css_blood") > 0
}

func (h *CrimsonFlashHandler) Execute(ctx *model.Context) error {
	if engineplayer.GetToken(ctx.User, "css_blood") <= 0 {
		return nil
	}
	addBlood(ctx.User, -1)
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	model.AppendAttackAction(ctx.User, "赤色一闪")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [赤色一闪]，移除1鲜血并获得额外攻击行动", ctx.User.Name))
	return nil
}

type CrimsonBloodRoseHandler struct{ BaseHandler }

func (h *CrimsonBloodRoseHandler) CanUse(ctx *model.Context) bool {
	return engineplayer.GetToken(ctx.User, "css_blood") >= 2
}

func (h *CrimsonBloodRoseHandler) Execute(ctx *model.Context) error {
	if len(ctx.Targets) != 2 {
		return fmt.Errorf("血染蔷薇需要恰好2名目标")
	}
	if engineplayer.GetToken(ctx.User, "css_blood") < 2 {
		return fmt.Errorf("鲜血不足")
	}
	healReduceTarget := ctx.Targets[0]
	healGainTarget := ctx.Targets[1]
	if healReduceTarget == nil || healGainTarget == nil {
		return fmt.Errorf("血染蔷薇目标无效")
	}
	if healGainTarget.Camp != ctx.User.Camp {
		return fmt.Errorf("血染蔷薇的第2个目标必须是我方角色")
	}
	addBlood(ctx.User, -2)
	if healReduceTarget.Heal > 0 {
		loss := 2
		if healReduceTarget.Heal < loss {
			loss = healReduceTarget.Heal
		}
		healReduceTarget.Heal -= loss
	}
	if ctx.Game.GetCampCrystals(string(ctx.User.Camp)) > 0 {
		ctx.Game.ModifyCrystal(string(ctx.User.Camp), -1)
		ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	}
	ctx.Game.Heal(healGainTarget.ID, 1)
	hasRoseCourtyard := false
	for _, fc := range ctx.User.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectRoseCourtyard {
			hasRoseCourtyard = true
			break
		}
	}
	if hasRoseCourtyard {
		for _, p := range ctx.Game.GetAllPlayers() {
			ctx.Game.AddPendingDamage(model.PendingDamage{
				SourceID:   ctx.User.ID,
				TargetID:   p.ID,
				Damage:     1,
				DamageType: model.MagicAttack,
			})
		}
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血染蔷薇]：%s -2治疗，阵营1水晶转1宝石，%s +1治疗", ctx.User.Name, healReduceTarget.Name, healGainTarget.Name))
	return nil
}

type CrimsonBloodBarrierHandler struct{ BaseHandler }

func (h *CrimsonBloodBarrierHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnDamageTaken || ctx.EventCtx == nil {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	if engineplayer.GetSkillFlowState(ctx.User, "css_blood_barrier_lock") > 0 {
		return false
	}
	return engineplayer.GetToken(ctx.User, "css_blood") > 0
}

func (h *CrimsonBloodBarrierHandler) Execute(ctx *model.Context) error {
	if engineplayer.GetToken(ctx.User, "css_blood") <= 0 {
		return nil
	}
	engineplayer.SetSkillFlowState(ctx.User, "css_blood_barrier_lock", 1)
	addBlood(ctx.User, -1)
	if ctx.EventCtx != nil && ctx.EventCtx.DamageVal != nil && *ctx.EventCtx.DamageVal > 0 {
		*ctx.EventCtx.DamageVal--
	}
	sourceID := ""
	if ctx.EventCtx != nil {
		sourceID = strings.TrimSpace(ctx.EventCtx.SourceID)
	}
	if sourceID != "" && sourceID != ctx.User.ID {
		ctx.Game.AddPendingDamage(model.PendingDamage{
			SourceID:   ctx.User.ID,
			TargetID:   sourceID,
			Damage:     1,
			DamageType: model.MagicAttack,
		})
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血气屏障]，本次法术伤害-1，并对伤害来源造成1点法术伤害", ctx.User.Name))
	return nil
}

type CrimsonRoseCourtyardHandler struct{ BaseHandler }

func (h *CrimsonRoseCourtyardHandler) Execute(ctx *model.Context) error { return nil }

type CrimsonDanceHandler struct{ BaseHandler }

func (h *CrimsonDanceHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.User.Character == nil {
		return false
	}
	if !(engineplayer.CanPayCrystalLike(ctx, 1) || ctx.User.Gem > 0) {
		return false
	}
	return ctx.User.HasExclusiveCard(ctx.User.Character.ID, "血蔷薇庭院")
}

func (h *CrimsonDanceHandler) Execute(ctx *model.Context) error {
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "css_dance_mode",
			"user_id":     ctx.User.ID,
			"can_crystal": engineplayer.CanPayCrystalLike(ctx, 1),
			"can_gem":     ctx.User.Gem > 0,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [散华轮舞]，等待选择模式", ctx.User.Name))
	return nil
}
