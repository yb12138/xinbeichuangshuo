package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

type MagicLancerDarkReleaseHandler struct{ BaseHandler }

type MagicLancerPhantomStardustHandler struct{ BaseHandler }

type MagicLancerDarkBindHandler struct{ BaseHandler }

type MagicLancerDarkBarrierHandler struct{ BaseHandler }

type MagicLancerFullnessHandler struct{ BaseHandler }

type MagicLancerBlackSpearHandler struct{ BaseHandler }

func magicLancerMagicCardCount(user *model.Player) int {
	if user == nil {
		return 0
	}
	count := 0
	for _, c := range user.Hand {
		if c.Type == model.CardTypeMagic {
			count++
		}
	}
	return count
}

func magicLancerThunderCardCount(user *model.Player) int {
	if user == nil {
		return 0
	}
	count := 0
	for _, c := range user.Hand {
		if c.Element == model.ElementThunder {
			count++
		}
	}
	return count
}

func magicLancerHasMagicOrThunder(user *model.Player) bool {
	if user == nil {
		return false
	}
	for _, c := range user.Hand {
		if c.Type == model.CardTypeMagic || c.Element == model.ElementThunder {
			return true
		}
	}
	return false
}

func (h *MagicLancerDarkReleaseHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return getToken(ctx.User, "ml_phantom_form") <= 0
}

func (h *MagicLancerDarkReleaseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("暗之解放上下文无效")
	}
	if getToken(ctx.User, "ml_phantom_form") > 0 {
		return fmt.Errorf("已处于幻影形态，不能再次发动暗之解放")
	}
	setToken(ctx.User, "ml_phantom_form", 1)
	if ctx.User.TurnState.UsedSkillCounts == nil {
		ctx.User.TurnState.UsedSkillCounts = map[string]int{}
	}
	ctx.User.TurnState.UsedSkillCounts["ml_dark_release_next_attack_bonus"] = 1
	ctx.User.TurnState.UsedSkillCounts["ml_dark_release_lock_turn"] = 1
	ctx.Game.Log(fmt.Sprintf("%s 发动 [暗之解放]，进入幻影形态：手牌上限恒定为5，本回合下一次主动攻击伤害+1，且本回合不能发动充盈/漆黑之枪", ctx.User.Name))
	return nil
}

func (h *MagicLancerPhantomStardustHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return getToken(ctx.User, "ml_phantom_form") > 0
}

func (h *MagicLancerPhantomStardustHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("幻影星尘上下文无效")
	}
	if getToken(ctx.User, "ml_phantom_form") <= 0 {
		return fmt.Errorf("仅幻影形态下可发动幻影星尘")
	}
	before := ctx.Game.GetCampMorale(string(ctx.User.Camp))
	setToken(ctx.User, "ml_stardust_pending", 1)
	setToken(ctx.User, "ml_stardust_wait_discard", 0)
	setToken(ctx.User, "ml_stardust_morale_before", before)
	ctx.Game.InflictDamage(ctx.User.ID, ctx.User.ID, 2, "magic")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [幻影星尘]：先对自己造成2点法术伤害，待完全结算后转正，并根据士气变化判定是否追加目标伤害", ctx.User.Name))
	return nil
}

func (h *MagicLancerDarkBindHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MagicLancerDarkBindHandler) Execute(ctx *model.Context) error { return nil }

func (h *MagicLancerDarkBarrierHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnDamageTaken {
		return false
	}
	if ctx.TriggerCtx.DamageVal == nil || *ctx.TriggerCtx.DamageVal <= 0 {
		return false
	}
	return magicLancerMagicCardCount(ctx.User) > 0 || magicLancerThunderCardCount(ctx.User) > 0
}

func (h *MagicLancerDarkBarrierHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("暗之障壁上下文无效")
	}
	magicCount := magicLancerMagicCardCount(ctx.User)
	thunderCount := magicLancerThunderCardCount(ctx.User)
	if magicCount <= 0 && thunderCount <= 0 {
		return fmt.Errorf("暗之障壁需要法术牌或雷系牌")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "ml_dark_barrier_mode",
			"user_id":     ctx.User.ID,
			"max_magic":   magicCount,
			"max_thunder": thunderCount,
			"source_player_id": func() string {
				if ctx.TriggerCtx != nil {
					return ctx.TriggerCtx.SourceID
				}
				return ""
			}(),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 可发动 [暗之障壁]，请选择弃牌类型与数量", ctx.User.Name))
	return nil
}

func (h *MagicLancerFullnessHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["ml_dark_release_lock_turn"] > 0 {
		return false
	}
	return magicLancerHasMagicOrThunder(ctx.User)
}

func (h *MagicLancerFullnessHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("充盈上下文无效")
	}
	if ctx.User.TurnState.UsedSkillCounts["ml_dark_release_lock_turn"] > 0 {
		return fmt.Errorf("本回合已发动暗之解放，不能发动充盈")
	}
	if !magicLancerHasMagicOrThunder(ctx.User) {
		return fmt.Errorf("充盈需要弃置1张法术牌或雷系牌")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "ml_fullness_cost_card",
			"user_id":     ctx.User.ID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [充盈]，请先弃置1张法术牌或雷系牌", ctx.User.Name))
	return nil
}

func (h *MagicLancerBlackSpearHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackHit {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if getToken(ctx.User, "ml_phantom_form") <= 0 {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["ml_dark_release_lock_turn"] > 0 {
		return false
	}
	handCount := len(ctx.Target.Hand)
	if handCount != 1 && handCount != 2 {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *MagicLancerBlackSpearHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("漆黑之枪上下文无效")
	}
	if getToken(ctx.User, "ml_phantom_form") <= 0 {
		return fmt.Errorf("仅幻影形态下可发动漆黑之枪")
	}
	if ctx.User.TurnState.UsedSkillCounts["ml_dark_release_lock_turn"] > 0 {
		return fmt.Errorf("本回合已发动暗之解放，不能发动漆黑之枪")
	}
	maxX := ctx.Game.GetUsableCrystal(ctx.User.ID)
	if maxX <= 0 {
		return fmt.Errorf("漆黑之枪至少需要1点蓝水晶（红宝石可替代）")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "ml_black_spear_x",
			"user_id":     ctx.User.ID,
			"target_id":   ctx.Target.ID,
			"max_x":       maxX,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [漆黑之枪]，请选择X（1~%d）", ctx.User.Name, maxX))
	return nil
}
