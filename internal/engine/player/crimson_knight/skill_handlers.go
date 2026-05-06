// gameflow: 红莲骑士技能处理器。

package crimson_knight

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// --- Helper functions ---

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

func canPayCrystalLike(ctx *model.Context, amount int) bool {
	return engineplayer.CanPayCrystalLike(ctx, amount)
}

func spendCrystalLike(ctx *model.Context, amount int) bool {
	return engineplayer.SpendCrystalLike(ctx, amount)
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

func addAttackAction(p *model.Player, source string) {
	model.AppendAttackAction(p, source)
}

// --- 19. 红莲骑士 ---

type CrimsonKnightCrimsonPactHandler struct{ engineplayer.BaseHandler }

type CrimsonKnightCrimsonFaithHandler struct{ engineplayer.BaseHandler }

type CrimsonKnightBloodyPrayerHandler struct{ engineplayer.BaseHandler }

type CrimsonKnightKillingFeastHandler struct{ engineplayer.BaseHandler }

type CrimsonKnightHotBloodHandler struct{ engineplayer.BaseHandler }

type CrimsonKnightCalmMindHandler struct{ engineplayer.BaseHandler }

type CrimsonKnightCrimsonCrossHandler struct{ engineplayer.BaseHandler }

func (h *CrimsonKnightCrimsonPactHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnAttackDeclared {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *CrimsonKnightCrimsonPactHandler) Execute(ctx *model.Context) error {
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [腥红圣约] 触发，+1治疗", ctx.User.Name))
	return nil
}

func (h *CrimsonKnightCrimsonFaithHandler) Execute(ctx *model.Context) error { return nil }

func (h *CrimsonKnightBloodyPrayerHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if ctx.User.Heal <= 0 {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			return true
		}
	}
	return false
}

func (h *CrimsonKnightBloodyPrayerHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	if ctx.User.Heal <= 0 {
		return fmt.Errorf("血腥祷言需要至少1点治疗")
	}
	var allyIDs []string
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			allyIDs = append(allyIDs, p.ID)
		}
	}
	if len(allyIDs) == 0 {
		return fmt.Errorf("没有可分配治疗的队友")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "crk_bloody_prayer_x",
			"user_id":     ctx.User.ID,
			"max_x":       ctx.User.Heal,
			"ally_ids":    allyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血腥祷言]，请选择X与治疗队友", ctx.User.Name))
	return nil
}

func (h *CrimsonKnightKillingFeastHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
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
	return getToken(ctx.User, "crk_blood_mark") > 0
}

func (h *CrimsonKnightKillingFeastHandler) Execute(ctx *model.Context) error {
	if getToken(ctx.User, "crk_blood_mark") <= 0 {
		return nil
	}
	addToken(ctx.User, "crk_blood_mark", -1, 0, 3)
	// 先提升本次命中伤害，再追加自伤到 PendingDamageQueue。
	// 否则 append 触发底层扩容时，DamageVal 可能指向旧切片元素导致加伤丢失。
	if ctx.EventCtx != nil && ctx.EventCtx.DamageVal != nil {
		*ctx.EventCtx.DamageVal += 2
	}
	// 规则：先结算本技能自伤，再结算本次攻击命中伤害。
	ctx.Game.AddPendingDamageFront(model.PendingDamage{
		SourceID:              ctx.User.ID,
		TargetID:              ctx.User.ID,
		Damage:                4,
		DamageType:            model.MagicAttack,
		AllowCrimsonFaithHeal: true,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [杀戮盛宴]，移除1血印并对自己造成4伤害，本次攻击伤害+2", ctx.User.Name))
	return nil
}

func (h *CrimsonKnightHotBloodHandler) Execute(ctx *model.Context) error { return nil }

func (h *CrimsonKnightCalmMindHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnActionEnd {
		return false
	}
	if !hasForm(ctx.User, model.FormCrimsonKnightHotBlooded) {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack && ctx.EventCtx.ActionType != model.ActionMagic {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *CrimsonKnightCalmMindHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("戒骄戒躁上下文无效")
	}
	if ctx.EventCtx == nil {
		return fmt.Errorf("戒骄戒躁缺少行动结束上下文")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("戒骄戒躁需要1蓝水晶（红宝石可替代）")
	}
	leaveForm(ctx.User, model.FormCrimsonKnightHotBlooded)

	actionType := ctx.EventCtx.ActionType
	if actionType != model.ActionAttack && actionType != model.ActionMagic {
		return fmt.Errorf("戒骄戒躁只支持攻击/法术行动结束后触发")
	}

	actionLabel := "攻击"
	if actionType == model.ActionMagic {
		actionLabel = "法术"
	}
	model.AppendExtraAction(ctx.User, "戒骄戒躁", string(actionType))
	ctx.Game.Log(fmt.Sprintf("%s 发动 [戒骄戒躁]，脱离热血沸腾形态并额外获得1次%s行动", ctx.User.Name, actionLabel))
	return nil
}

func (h *CrimsonKnightCrimsonCrossHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Target == nil {
		return false
	}
	if ctx.Target.Camp == ctx.User.Camp {
		return false
	}
	if getToken(ctx.User, "crk_blood_mark") <= 0 {
		return false
	}
	magicCount := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicCount++
		}
	}
	return magicCount >= 2
}

func (h *CrimsonKnightCrimsonCrossHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("腥红十字需要目标")
	}
	if ctx.Target.Camp == ctx.User.Camp {
		return fmt.Errorf("腥红十字只能指定敌方角色")
	}
	if getToken(ctx.User, "crk_blood_mark") <= 0 {
		return fmt.Errorf("血印不足")
	}
	addToken(ctx.User, "crk_blood_mark", -1, 0, 3)
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:              ctx.User.ID,
		TargetID:              ctx.User.ID,
		Damage:                4,
		DamageType:            model.MagicAttack,
		AllowCrimsonFaithHeal: true,
	})
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   ctx.Target.ID,
		Damage:     3,
		DamageType: model.MagicAttack,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [腥红十字]，对自己造成4点法术伤害，并对 %s 造成3点法术伤害", ctx.User.Name, ctx.Target.Name))
	return nil
}
