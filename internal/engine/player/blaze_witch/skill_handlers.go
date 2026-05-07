// gameflow: 烈焰魔女技能处理器。

package blaze_witch

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"

	"starcup-engine/internal/model"
)

// --- Helper functions ---

// --- BlazeWitch Handlers ---

type BlazeWitchRebirthClockHandler struct{ engineplayer.BaseHandler }

type BlazeWitchBlazingCodexHandler struct{ engineplayer.BaseHandler }

type BlazeWitchHeavenfireCleaveHandler struct{ engineplayer.BaseHandler }

type BlazeWitchWitchWrathHandler struct{ engineplayer.BaseHandler }

type BlazeWitchSubstituteDollHandler struct{ engineplayer.BaseHandler }

type BlazeWitchPainLinkHandler struct{ engineplayer.BaseHandler }

type BlazeWitchManaInversionHandler struct{ engineplayer.BaseHandler }

func (h *BlazeWitchRebirthClockHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BlazeWitchRebirthClockHandler) Execute(ctx *model.Context) error { return nil }

func (h *BlazeWitchBlazingCodexHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("苍炎法典需要目标")
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [苍炎法典]，先对 %s 后对自己各造成2点法术伤害", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *BlazeWitchHeavenfireCleaveHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if engineplayer.HasForm(ctx.User, model.FormBlazeWitchFlame) {
		return true
	}
	return engineplayer.GetToken(ctx.User, "bw_rebirth") > 0
}

func (h *BlazeWitchHeavenfireCleaveHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("天火断空需要目标")
	}
	damage := 3
	userCampMorale := ctx.Game.GetCampMorale(string(ctx.User.Camp))
	targetCampMorale := ctx.Game.GetCampMorale(string(ctx.Target.Camp))
	if userCampMorale < targetCampMorale {
		damage++
	}
	if !engineplayer.HasForm(ctx.User, model.FormBlazeWitchFlame) {
		if engineplayer.GetToken(ctx.User, "bw_rebirth") <= 0 {
			return fmt.Errorf("天火断空需要至少1点重生")
		}
		engineplayer.AddToken(ctx.User, "bw_rebirth", -1, 4)
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     damage,
		DamageType: model.MagicAttack,
	})
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     damage,
		DamageType: model.MagicAttack,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [天火断空]，先对 %s 后对自己各造成%d点法术伤害", ctx.User.Name, ctx.Target.Name, damage))
	return nil
}

func (h *BlazeWitchWitchWrathHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return len(ctx.User.Hand) < 4
}

func (h *BlazeWitchWitchWrathHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔女之怒上下文无效")
	}
	engineplayer.SetForm(ctx.User, model.FormBlazeWitchFlame)
	engineplayer.SetSkillFlowState(ctx.User, "bw_flame_release_pending", 1)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "bw_witch_wrath_draw",
			"user_id":       ctx.User.ID,
			"waiting_phase": model.TurnStageActionStart,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔女之怒]，进入烈焰形态并选择摸牌数量", ctx.User.Name))
	return nil
}

func (h *BlazeWitchSubstituteDollHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnDamageTaken {
		return false
	}
	if engineplayer.GetSkillFlowState(ctx.User, "bw_substitute_lock") > 0 {
		return false
	}
	if ctx.Flags["IsMagicDamage"] {
		return false
	}
	if ctx.EventCtx.DamageVal == nil || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	magicCount := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicCount++
		}
	}
	if magicCount <= 0 {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			return true
		}
	}
	return false
}

func (h *BlazeWitchSubstituteDollHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("替身玩偶上下文无效")
	}
	engineplayer.SetSkillFlowState(ctx.User, "bw_substitute_lock", 1)
	var magicIndices []int
	for i, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicIndices = append(magicIndices, i)
		}
	}
	var allyIDs []string
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			allyIDs = append(allyIDs, p.ID)
		}
	}
	if len(magicIndices) == 0 || len(allyIDs) == 0 {
		engineplayer.SetSkillFlowState(ctx.User, "bw_substitute_lock", 0)
		return fmt.Errorf("替身玩偶缺少可用牌或队友")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "bw_substitute_doll_card",
			"user_id":       ctx.User.ID,
			"magic_indices": magicIndices,
			"ally_ids":      allyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [替身玩偶]，请选择要弃置的法术牌", ctx.User.Name))
	return nil
}

func (h *BlazeWitchPainLinkHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return engineplayer.CanPayCrystalLike(ctx, 1)
}

func (h *BlazeWitchPainLinkHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("痛苦链接需要目标")
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	engineplayer.SetSkillFlowState(ctx.User, "bw_pain_link_pending_discard", 1)
	engineplayer.SetSkillFlowState(ctx.User, "bw_pain_link_pending_hits", 2)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [痛苦链接]，先对 %s 后对自己各造成1点法术伤害", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *BlazeWitchManaInversionHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnDamageTaken {
		return false
	}
	if engineplayer.GetSkillFlowState(ctx.User, "bw_mana_inversion_lock") > 0 {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	if ctx.EventCtx.DamageVal == nil || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	magicCount := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicCount++
		}
	}
	if magicCount < 2 {
		return false
	}
	if !engineplayer.CanPayCrystalLike(ctx, 1) {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp != ctx.User.Camp {
			return true
		}
	}
	return false
}

func (h *BlazeWitchManaInversionHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔能反转上下文无效")
	}
	if !engineplayer.SpendCrystalLike(ctx, 1) {
		return fmt.Errorf("魔能反转需要1蓝水晶（红宝石可替代）")
	}
	engineplayer.SetSkillFlowState(ctx.User, "bw_mana_inversion_lock", 1)
	magicCount := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicCount++
		}
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bw_mana_inversion_x",
			"user_id":     ctx.User.ID,
			"max_x":       magicCount,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔能反转]，请选择弃牌数量X", ctx.User.Name))
	return nil
}
