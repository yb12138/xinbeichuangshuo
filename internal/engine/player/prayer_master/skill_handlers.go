// gameflow: 祈祷师技能处理器。

package prayer_master

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// --- Helper functions ---

func addMagicAction(p *model.Player, source string) {
	model.AppendMagicAction(p, source)
}

// --- 18. 祈祷师 ---

type PrayerEnterFormHandler struct{ engineplayer.BaseHandler }

type PrayerRadiantFaithHandler struct{ engineplayer.BaseHandler }

type PrayerDarkCurseHandler struct{ engineplayer.BaseHandler }

type PrayerPowerBlessingHandler struct{ engineplayer.BaseHandler }

type PrayerSwiftBlessingHandler struct{ engineplayer.BaseHandler }

type PrayerManaTideHandler struct{ engineplayer.BaseHandler }

func (h *PrayerEnterFormHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && !engineplayer.HasForm(ctx.User, model.FormPrayerMasterPrayer)
}

func (h *PrayerEnterFormHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil {
		return fmt.Errorf("上下文无效")
	}
	// CostGem 已在 ConfirmStartupSkillAction 由框架统一扣减（见 skill definition CostGem: 1）
	if engineplayer.HasForm(ctx.User, model.FormPrayerMasterPrayer) {
		return nil
	}
	engineplayer.SetForm(ctx.User, model.FormPrayerMasterPrayer)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [祈祷]，进入祈祷形态", ctx.User.Name))
	return nil
}

func (h *PrayerRadiantFaithHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return engineplayer.HasForm(ctx.User, model.FormPrayerMasterPrayer) && engineplayer.GetToken(ctx.User, "prayer_rune") > 0
}

func (h *PrayerRadiantFaithHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	if !engineplayer.HasForm(ctx.User, model.FormPrayerMasterPrayer) {
		return fmt.Errorf("不在祈祷形态，无法发动光辉信仰")
	}
	if engineplayer.GetToken(ctx.User, "prayer_rune") <= 0 {
		return fmt.Errorf("祈祷符文不足")
	}
	target := ctx.Target
	if target == nil || target.Camp != ctx.User.Camp || target.ID == ctx.User.ID {
		return fmt.Errorf("光辉信仰需要1名其他队友目标")
	}
	engineplayer.AddToken(ctx.User, "prayer_rune", -1, 3)
	ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	ctx.Game.Heal(target.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [光辉信仰]，移除1祈祷符文，战绩区+1红宝石，并治疗 %s 1点", ctx.User.Name, target.Name))
	return nil
}

func (h *PrayerDarkCurseHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return engineplayer.HasForm(ctx.User, model.FormPrayerMasterPrayer) && engineplayer.GetToken(ctx.User, "prayer_rune") > 0
}

func (h *PrayerDarkCurseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.Game == nil {
		return fmt.Errorf("黑暗诅咒需要目标")
	}
	if !engineplayer.HasForm(ctx.User, model.FormPrayerMasterPrayer) {
		return fmt.Errorf("不在祈祷形态，无法发动黑暗诅咒")
	}
	if engineplayer.GetToken(ctx.User, "prayer_rune") <= 0 {
		return fmt.Errorf("祈祷符文不足")
	}
	engineplayer.AddToken(ctx.User, "prayer_rune", -1, 3)
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
	if ctx.Timing != model.TimingActionEnd {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionMagic {
		return false
	}
	return engineplayer.CanPayCrystalLike(ctx, 1)
}

func (h *PrayerManaTideHandler) Execute(ctx *model.Context) error {
	if !engineplayer.SpendCrystalLike(ctx, 1) {
		return fmt.Errorf("法力潮汐需要1蓝水晶（红宝石可替代）")
	}
	addMagicAction(ctx.User, "法力潮汐")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [法力潮汐]，额外获得1次法术行动", ctx.User.Name))
	return nil
}
