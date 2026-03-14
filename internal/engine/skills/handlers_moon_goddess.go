package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

const (
	moonGoddessNewMoonCap = 2
	moonGoddessPetrifyCap = 3
)

type MoonGoddessNewMoonShelterHandler struct{ BaseHandler }

type MoonGoddessDarkMoonCurseHandler struct{ BaseHandler }

type MoonGoddessMedusaEyeHandler struct{ BaseHandler }

type MoonGoddessMoonCycleHandler struct{ BaseHandler }

type MoonGoddessBlasphemyHandler struct{ BaseHandler }

type MoonGoddessDarkMoonSlashHandler struct{ BaseHandler }

type MoonGoddessPaleMoonHandler struct{ BaseHandler }

func moonGoddessEnemyIDs(game model.IGameEngine, user *model.Player) []string {
	if game == nil || user == nil {
		return nil
	}
	var ids []string
	for _, p := range game.GetAllPlayers() {
		if p == nil || p.Camp == user.Camp {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

func moonGoddessDarkMoonCount(user *model.Player) int {
	if user == nil {
		return 0
	}
	count := 0
	for _, fc := range user.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
			continue
		}
		count++
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	user.Tokens["mg_dark_moon_count"] = count
	return count
}

func addMoonGoddessNewMoon(user *model.Player, delta int) int {
	return addToken(user, "mg_new_moon", delta, 0, moonGoddessNewMoonCap)
}

func addMoonGoddessPetrify(user *model.Player, delta int) int {
	return addToken(user, "mg_petrify", delta, 0, moonGoddessPetrifyCap)
}

func (h *MoonGoddessNewMoonShelterHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil {
		return false
	}
	if ctx.Trigger != model.TriggerBeforeMoraleLoss {
		return false
	}
	if *ctx.TriggerCtx.DamageVal <= 0 {
		return false
	}
	if ctx.Selections == nil {
		return false
	}
	fromDamage, _ := ctx.Selections["from_damage_draw"].(bool)
	if !fromDamage {
		return false
	}
	if _, used := ctx.Selections["mg_new_moon_absorb_by"].(string); used {
		return false
	}
	cards, ok := ctx.Selections["discarded_cards"].([]model.Card)
	return ok && len(cards) > 0
}

func (h *MoonGoddessNewMoonShelterHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil {
		return fmt.Errorf("新月庇护上下文无效")
	}
	cards, ok := ctx.Selections["discarded_cards"].([]model.Card)
	if !ok || len(cards) == 0 {
		return fmt.Errorf("新月庇护未找到可转化的爆牌")
	}
	if ctx.User.Tokens == nil {
		ctx.User.Tokens = map[string]int{}
	}
	ctx.User.Tokens["mg_dark_form"] = 1
	added := 0
	for _, c := range cards {
		ctx.User.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  ctx.User.ID,
			SourceID: ctx.User.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectMoonDarkMoon,
			Trigger:  model.EffectTriggerManual,
		})
		added++
	}
	moonGoddessDarkMoonCount(ctx.User)
	ctx.Selections["mg_new_moon_absorb_by"] = ctx.User.ID
	*ctx.TriggerCtx.DamageVal = 0
	ctx.Game.Log(fmt.Sprintf("%s 的 [新月庇护] 触发：进入暗月形态并吸收%d张爆牌为暗月，本次士气不下降",
		ctx.User.Name, added))
	return nil
}

func (h *MoonGoddessDarkMoonCurseHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MoonGoddessDarkMoonCurseHandler) Execute(ctx *model.Context) error { return nil }

func (h *MoonGoddessMedusaEyeHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MoonGoddessMedusaEyeHandler) Execute(ctx *model.Context) error { return nil }

func (h *MoonGoddessMoonCycleHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MoonGoddessMoonCycleHandler) Execute(ctx *model.Context) error { return nil }

func (h *MoonGoddessBlasphemyHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MoonGoddessBlasphemyHandler) Execute(ctx *model.Context) error { return nil }

func (h *MoonGoddessDarkMoonSlashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackHit {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if getToken(ctx.User, "mg_dark_form") <= 0 {
		return false
	}
	if moonGoddessDarkMoonCount(ctx.User) <= 0 {
		return false
	}
	return ctx.Game.CanPayCrystalCost(ctx.User.ID, 1)
}

func (h *MoonGoddessDarkMoonSlashHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("暗月斩上下文无效")
	}
	if getToken(ctx.User, "mg_dark_form") <= 0 {
		return fmt.Errorf("仅暗月形态可发动暗月斩")
	}
	if moonGoddessDarkMoonCount(ctx.User) <= 0 {
		return fmt.Errorf("暗月不足，无法发动暗月斩")
	}
	if !ctx.Game.ConsumeCrystalCost(ctx.User.ID, 1) {
		return fmt.Errorf("暗月斩需要1点蓝水晶（红宝石可替代）")
	}
	maxX := moonGoddessDarkMoonCount(ctx.User)
	if maxX > 2 {
		maxX = 2
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mg_darkmoon_slash_x",
			"user_id":     ctx.User.ID,
			"max_x":       maxX,
			"user_ctx":    ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [暗月斩]：消耗1水晶，选择移除暗月数量X（0~%d）", ctx.User.Name, maxX))
	return nil
}

func (h *MoonGoddessPaleMoonHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	branch1 := getToken(ctx.User, "mg_petrify") >= 3
	branch2 := len(ctx.User.Hand) > 0 && len(moonGoddessEnemyIDs(ctx.Game, ctx.User)) > 0
	return branch1 || branch2
}

func (h *MoonGoddessPaleMoonHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("苍白之月上下文无效")
	}
	var modes []string
	if getToken(ctx.User, "mg_petrify") >= 3 {
		modes = append(modes, "branch1")
	}
	if len(ctx.User.Hand) > 0 && len(moonGoddessEnemyIDs(ctx.Game, ctx.User)) > 0 {
		modes = append(modes, "branch2")
	}
	if len(modes) == 0 {
		return fmt.Errorf("当前条件不满足苍白之月任一分支")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mg_pale_moon_mode",
			"user_id":     ctx.User.ID,
			"modes":       modes,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [苍白之月]：请选择分支", ctx.User.Name))
	return nil
}
