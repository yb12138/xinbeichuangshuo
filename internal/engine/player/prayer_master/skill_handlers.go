// gameflow: 祈祷师技能处理器。

package prayer_master

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

func canPayCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.CanPayCrystalCost(ctx.User.ID, amount)
}

func spendCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.ConsumeCrystalCost(ctx.User.ID, amount)
}

func addMagicAction(p *model.Player, source string) {
	model.AppendMagicAction(p, source)
}

// --- 18. 祈祷师 ---

type PrayerEnterFormHandler struct{ engineplayer.BaseHandler }

type PrayerRuneGainHandler struct{ engineplayer.BaseHandler }

type PrayerRadiantFaithHandler struct{ engineplayer.BaseHandler }

type PrayerDarkCurseHandler struct{ engineplayer.BaseHandler }

type PrayerPowerBlessingHandler struct{ engineplayer.BaseHandler }

type PrayerSwiftBlessingHandler struct{ engineplayer.BaseHandler }

type PrayerManaTideHandler struct{ engineplayer.BaseHandler }

func (h *PrayerEnterFormHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && !hasForm(ctx.User, model.FormPrayerMasterPrayer)
}

func (h *PrayerEnterFormHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil {
		return fmt.Errorf("上下文无效")
	}
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("祈祷需要至少1个红宝石")
	}
	if hasForm(ctx.User, model.FormPrayerMasterPrayer) {
		return nil
	}
	ctx.User.Gem--
	enterForm(ctx.User, model.FormPrayerMasterPrayer)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [祈祷]，进入祈祷形态", ctx.User.Name))
	return nil
}

func (h *PrayerRuneGainHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if !hasForm(ctx.User, model.FormPrayerMasterPrayer) {
		return false
	}
	if ctx.Timing != model.TimingOnAttackDeclared {
		return false
	}
	// 仅主动攻击
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return true
}

func (h *PrayerRuneGainHandler) Execute(ctx *model.Context) error {
	v := addToken(ctx.User, "prayer_rune", 2, 0, 3)
	ctx.Game.Log(fmt.Sprintf("%s 的 [祈祷符文] 触发，祈祷符文=%d", ctx.User.Name, v))
	return nil
}

func (h *PrayerRadiantFaithHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return hasForm(ctx.User, model.FormPrayerMasterPrayer) && getToken(ctx.User, "prayer_rune") > 0
}

func (h *PrayerRadiantFaithHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	if !hasForm(ctx.User, model.FormPrayerMasterPrayer) {
		return fmt.Errorf("不在祈祷形态，无法发动光辉信仰")
	}
	if getToken(ctx.User, "prayer_rune") <= 0 {
		return fmt.Errorf("祈祷符文不足")
	}
	target := ctx.Target
	if target == nil || target.Camp != ctx.User.Camp || target.ID == ctx.User.ID {
		return fmt.Errorf("光辉信仰需要1名其他队友目标")
	}
	addToken(ctx.User, "prayer_rune", -1, 0, 3)
	ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	ctx.Game.Heal(target.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [光辉信仰]，移除1祈祷符文，战绩区+1红宝石，并治疗 %s 1点", ctx.User.Name, target.Name))
	return nil
}

func (h *PrayerDarkCurseHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return hasForm(ctx.User, model.FormPrayerMasterPrayer) && getToken(ctx.User, "prayer_rune") > 0
}

func (h *PrayerDarkCurseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("黑暗诅咒需要目标")
	}
	if !hasForm(ctx.User, model.FormPrayerMasterPrayer) {
		return fmt.Errorf("不在祈祷形态，无法发动黑暗诅咒")
	}
	if getToken(ctx.User, "prayer_rune") <= 0 {
		return fmt.Errorf("祈祷符文不足")
	}
	addToken(ctx.User, "prayer_rune", -1, 0, 3)
	// 先结算对方，再结算自己
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
	ctx.Game.Log(fmt.Sprintf("%s 发动 [黑暗诅咒]，先对 %s 再对自己各造成2点法术伤害", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *PrayerPowerBlessingHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil {
		return fmt.Errorf("威力赐福需要目标")
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [威力赐福]，在 %s 面前放置威力赐福", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *PrayerSwiftBlessingHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil {
		return fmt.Errorf("迅捷赐福需要目标")
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [迅捷赐福]，在 %s 面前放置迅捷赐福", ctx.User.Name, ctx.Target.Name))
	return nil
}

func (h *PrayerManaTideHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnActionEnd {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionMagic {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *PrayerManaTideHandler) Execute(ctx *model.Context) error {
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("法力潮汐需要1蓝水晶（红宝石可替代）")
	}
	addMagicAction(ctx.User, "法力潮汐")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [法力潮汐]，额外获得1次法术行动", ctx.User.Name))
	return nil
}
