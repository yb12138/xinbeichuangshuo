// gameflow: 魔枪士 handler。

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
	return !hasForm(ctx.User, model.FormMagicLancerPhantom)
}

func (h *MagicLancerDarkReleaseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("暗之解放上下文无效")
	}
	if hasForm(ctx.User, model.FormMagicLancerPhantom) {
		return fmt.Errorf("已处于幻影形态，不能再次发动暗之解放")
	}
	enterForm(ctx.User, model.FormMagicLancerPhantom)
	ctx.Game.ApplyNextAttackDamageRule(ctx.User.ID, "ml_dark_release_next_attack_bonus", "ml_dark_release", 1, model.RuleLifeUntilTurnEnd)
	ctx.Game.ApplySkillGateRule(ctx.User.ID, "ml_dark_release_lock_turn", "ml_dark_release", []string{"ml_fullness", "ml_black_spear"}, model.RuleLifeUntilTurnEnd)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [暗之解放]，进入幻影形态：手牌上限恒定为5，本回合下一次主动攻击伤害+1，且本回合不能发动充盈/漆黑之枪", ctx.User.Name))
	return nil
}

func (h *MagicLancerPhantomStardustHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return hasForm(ctx.User, model.FormMagicLancerPhantom)
}

func (h *MagicLancerPhantomStardustHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("幻影星尘上下文无效")
	}
	if !hasForm(ctx.User, model.FormMagicLancerPhantom) {
		return fmt.Errorf("仅幻影形态下可发动幻影星尘")
	}
	if ctx.Target != nil {
		if ctx.Target.Camp == ctx.User.Camp || ctx.Target.ID == ctx.User.ID {
			return fmt.Errorf("幻影星尘目标必须是敌方角色")
		}
		for i, p := range ctx.Game.GetAllPlayers() {
			if p != nil && p.ID == ctx.Target.ID {
				setSkillFlow(ctx.User, "ml_stardust_locked_target_order", i+1)
				break
			}
		}
	} else {
		setSkillFlow(ctx.User, "ml_stardust_locked_target_order", 0)
	}
	before := ctx.Game.GetCampMorale(string(ctx.User.Camp))
	setSkillFlow(ctx.User, "ml_stardust_pending", 1)
	setSkillFlow(ctx.User, "ml_stardust_wait_discard", 0)
	setSkillFlow(ctx.User, "ml_stardust_morale_before", before)
	ctx.Game.InflictDamage(ctx.User.ID, ctx.User.ID, 2, model.MagicAttack)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [幻影星尘]：先对自己造成2点法术伤害，待完全结算后转正，并根据士气变化判定是否追加目标伤害", ctx.User.Name))
	return nil
}

func (h *MagicLancerDarkBindHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MagicLancerDarkBindHandler) Execute(ctx *model.Context) error { return nil }

func (h *MagicLancerDarkBarrierHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnDamageTaken {
		return false
	}
	if ctx.EventCtx.DamageVal == nil || *ctx.EventCtx.DamageVal <= 0 {
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
				if ctx.EventCtx != nil {
					return ctx.EventCtx.SourceID
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
	if ctx.Game != nil && ctx.Game.IsSkillBlocked(ctx.User.ID, "ml_fullness") {
		return false
	}
	return magicLancerHasMagicOrThunder(ctx.User)
}

func (h *MagicLancerFullnessHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("充盈上下文无效")
	}
	if ctx.Game.IsSkillBlocked(ctx.User.ID, "ml_fullness") {
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
			"locked_ally_id": func() string {
				if ctx.Target == nil {
					return ""
				}
				return ctx.Target.ID
			}(),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [充盈]，请先弃置1张法术牌或雷系牌", ctx.User.Name))
	return nil
}

func (h *MagicLancerBlackSpearHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnHitCheck {
		return false
	}
	if ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.ActionType != string(model.ActionAttack) {
		return false
	}
	if !ctx.EventCtx.AttackInfo.IsHit {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if !hasForm(ctx.User, model.FormMagicLancerPhantom) {
		return false
	}
	if ctx.Game != nil && ctx.Game.IsSkillBlocked(ctx.User.ID, "ml_black_spear") {
		return false
	}
	handCount := len(ctx.Target.Hand)
	if handCount != 1 && handCount != 2 {
		return false
	}
	return CanPayCrystalLike(ctx, 1)
}

func (h *MagicLancerBlackSpearHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("漆黑之枪上下文无效")
	}
	if !hasForm(ctx.User, model.FormMagicLancerPhantom) {
		return fmt.Errorf("仅幻影形态下可发动漆黑之枪")
	}
	if ctx.Game.IsSkillBlocked(ctx.User.ID, "ml_black_spear") {
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
