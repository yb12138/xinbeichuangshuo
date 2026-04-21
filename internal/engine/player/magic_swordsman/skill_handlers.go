// gameflow: 魔剑士技能处理器。

package magic_swordsman

import (
	"fmt"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

func getToken(p *model.Player, key string) int {
	if p == nil {
		return 0
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	return p.Tokens[key]
}

func setToken(p *model.Player, key string, v int) {
	if p == nil {
		return
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens[key] = v
}

func addToken(p *model.Player, key string, delta int, minV int, maxV int) int {
	cur := getToken(p, key)
	cur += delta
	if cur < minV {
		cur = minV
	}
	if maxV >= minV && cur > maxV {
		cur = maxV
	}
	setToken(p, key, cur)
	return cur
}

func hasForm(p *model.Player, form string) bool {
	return p != nil && p.Form == form
}

func enterForm(p *model.Player, form string) {
	if p == nil {
		return
	}
	p.Orientation = model.OrientationTapped
	p.Form = form
}

func leaveForm(p *model.Player, form string) {
	if p == nil {
		return
	}
	if form != "" && p.Form != form {
		return
	}
	p.Orientation = model.OrientationNormal
	p.Form = ""
}

type MagicSwordsmanAsuraComboHandler struct{ skills.BaseHandler }

type MagicSwordsmanShadowGatherHandler struct{ skills.BaseHandler }

type MagicSwordsmanShadowPowerHandler struct{ skills.BaseHandler }

type MagicSwordsmanShadowRejectHandler struct{ skills.BaseHandler }

type MagicSwordsmanShadowMeteorHandler struct{ skills.BaseHandler }

type MagicSwordsmanYellowSpringHandler struct{ skills.BaseHandler }

func (h *MagicSwordsmanAsuraComboHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnActionEnd || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	// 修罗连斩响应"攻击行动结束"，不应在应战攻击结束后触发。
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *MagicSwordsmanAsuraComboHandler) Execute(ctx *model.Context) error {
	model.AppendAttackAction(ctx.User, "修罗连斩", model.ElementFire)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [修罗连斩]，获得额外火系攻击行动", ctx.User.Name))
	return nil
}

func (h *MagicSwordsmanShadowGatherHandler) CanUse(ctx *model.Context) bool {
	return !hasForm(ctx.User, model.FormMagicSwordsmanShadow)
}

func (h *MagicSwordsmanShadowGatherHandler) Execute(ctx *model.Context) error {
	enterForm(ctx.User, model.FormMagicSwordsmanShadow)
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.User.ID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [暗影凝聚]，进入暗影形态并承受1点法术伤害", ctx.User.Name))
	return nil
}

func (h *MagicSwordsmanShadowPowerHandler) Execute(ctx *model.Context) error { return nil }

func (h *MagicSwordsmanShadowRejectHandler) Execute(ctx *model.Context) error { return nil }

func (h *MagicSwordsmanShadowMeteorHandler) CanUse(ctx *model.Context) bool {
	if !hasForm(ctx.User, model.FormMagicSwordsmanShadow) {
		return false
	}
	count := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			count++
		}
	}
	return count >= 2
}

func (h *MagicSwordsmanShadowMeteorHandler) Execute(ctx *model.Context) error {
	if !hasForm(ctx.User, model.FormMagicSwordsmanShadow) {
		return fmt.Errorf("暗影流星仅可在暗影形态下发动")
	}
	if ctx.Target == nil {
		return fmt.Errorf("暗影流星需要1名敌方目标")
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	camp := string(ctx.User.Camp)
	total := ctx.Game.GetCampGems(camp) + ctx.Game.GetCampCrystals(camp)
	if total >= 2 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type": "ms_shadow_meteor_release_confirm",
				"user_id":     ctx.User.ID,
				"camp":        camp,
			},
		})
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [暗影流星]，对 %s 造成2点法术伤害", ctx.User.Name, model.GetPlayerDisplayName(ctx.Target)))
	return nil
}

func (h *MagicSwordsmanYellowSpringHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnAttackDeclared || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return ctx.User.Gem > 0
}

func (h *MagicSwordsmanYellowSpringHandler) Execute(ctx *model.Context) error {
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("黄泉震颤需要至少1个红宝石")
	}
	ctx.User.Gem--
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		ctx.EventCtx.AttackInfo.CanBeResponded = false
	}
	ctx.User.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] = 1
	ctx.Game.Log(fmt.Sprintf("%s 发动 [黄泉震颤]，本次攻击不可应战", ctx.User.Name))
	return nil
}
