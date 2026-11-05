// gameflow: 血色剑灵技能处理器。

package crimson_sword_spirit

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type BaseHandler = engineplayer.BaseHandler

// --- Crimson Sword Spirit Skill Handlers ---

type CrimsonBloodThornsHandler struct{ BaseHandler }

func (h *CrimsonBloodThornsHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || !ctx.AttackHitPhase() || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	return info.ActionType == string(model.ActionAttack)
}

func (h *CrimsonBloodThornsHandler) Execute(ctx *model.Context) error {
	cur := addBlood(ctx.User, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [血色荆棘] 触发，鲜血=%d", ctx.User.Name, cur))
	return nil
}

type CrimsonFlashHandler struct{ BaseHandler }

func (h *CrimsonFlashHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingActionEnd || ctx.EventCtx == nil {
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
	if engineplayer.GetToken(ctx.User, "css_blood") < 2 {
		return fmt.Errorf("鲜血不足")
	}

	// 分步选择流程：先选移除治疗目标，再选队友获得治疗
	allPlayerIDs := make([]string, 0, len(ctx.Game.GetPlayers()))
	for _, p := range ctx.Game.GetAllPlayers() {
		allPlayerIDs = append(allPlayerIDs, p.ID)
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "css_blood_rose_remove_heal_target",
			"user_id":     ctx.User.ID,
			"target_ids":  allPlayerIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血染蔷薇]，等待选择移除治疗目标", ctx.User.Name))
	return nil
}

type CrimsonBloodBarrierHandler struct{ BaseHandler }

func (h *CrimsonBloodBarrierHandler) CanUse(ctx *model.Context) bool {
	if !ctx.DamageTakenPhase() || ctx.EventCtx == nil {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	if ctx.EventCtx.SourceID == "" || ctx.EventCtx.SourceID == ctx.User.ID || ctx.EventCtx.TargetID != ctx.User.ID {
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
	// 费用与减伤在发动时立即生效；反伤目标为由血色剑灵指定的任意一名对手。
	enemyIDs := make([]string, 0, len(ctx.Game.GetAllPlayers()))
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp != ctx.User.Camp {
			enemyIDs = append(enemyIDs, p.ID)
		}
	}
	if len(enemyIDs) == 0 {
		ctx.Game.Log(fmt.Sprintf("%s 发动 [血气屏障]，本次法术伤害-1（无可选对手，跳过反伤）", ctx.User.Name))
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "css_blood_barrier_target",
			"user_id":     ctx.User.ID,
			"target_ids":  enemyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血气屏障]，本次法术伤害-1，请选择一名对手承受1点法术伤害", ctx.User.Name))
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
