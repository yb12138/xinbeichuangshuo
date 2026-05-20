// gameflow: 月女神 handler。

package moon

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"

	"starcup-engine/internal/model"
)

// 辅助函数（仅定义本文件专用、choices.go 中未出现的）
func skillGetToken(p *model.Player, key string) int {
	if p == nil || p.Tokens == nil {
		return 0
	}
	return p.Tokens[key]
}

func skillEnterForm(p *model.Player, form string) {
	if p == nil {
		return
	}
	p.Orientation = model.OrientationTapped
	p.Form = form
}

// moonGoddessEnemyIDsFromGame 返回敌方玩家 ID 列表（基于 IGameEngine）。
// choices.go 中已有 moonGoddessEnemyIDs 基于 ChoiceRuntime，此处为 handler 专用版本。
func moonGoddessEnemyIDsFromGame(game model.IGameEngine, user *model.Player) []string {
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

// Handlers
type MoonGoddessNewMoonShelterHandler struct{ engineplayer.BaseHandler }

type MoonGoddessDarkMoonCurseHandler struct{ engineplayer.BaseHandler }

type MoonGoddessMedusaEyeHandler struct{ engineplayer.BaseHandler }

type MoonGoddessMoonCycleHandler struct{ engineplayer.BaseHandler }

type MoonGoddessBlasphemyHandler struct{ engineplayer.BaseHandler }

type MoonGoddessDarkMoonSlashHandler struct{ engineplayer.BaseHandler }

type MoonGoddessPaleMoonHandler struct{ engineplayer.BaseHandler }

func (h *MoonGoddessNewMoonShelterHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return false
	}
	if ctx.Timing != model.TimingBeforeMoraleLoss {
		return false
	}
	if engineplayer.EffectiveForm(ctx.User) != "" {
		return false
	}
	if *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	if ctx.Selections == nil {
		return false
	}
	fromDamage, _ := ctx.Selections["from_damage_draw"].(bool)
	if !fromDamage {
		return false
	}
	if destination, _ := ctx.Selections["discard_destination"].(string); destination != "" {
		return false
	}
	cards, ok := ctx.Selections["discarded_cards"].([]model.Card)
	return ok && len(cards) > 0
}

func (h *MoonGoddessNewMoonShelterHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return fmt.Errorf("新月庇护上下文无效")
	}
	cards, ok := ctx.Selections["discarded_cards"].([]model.Card)
	if !ok || len(cards) == 0 {
		return fmt.Errorf("新月庇护未找到可转化的爆牌")
	}
	if ctx.User.Tokens == nil {
		ctx.User.Tokens = map[string]int{}
	}
	skillEnterForm(ctx.User, model.FormMoonGoddessDarkMoon)
	added := 0
	for _, c := range cards {
		ctx.User.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  ctx.User.ID,
			SourceID: ctx.User.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectMoonDarkMoon,
			Hook:     model.FieldHookManual,
		})
		added++
	}
	moonGoddessDarkMoonCount(ctx.User)
	ctx.Selections["discard_destination"] = "absorbed"
	ctx.Selections["discard_absorbed_by"] = ctx.User.ID
	*ctx.EventCtx.DamageVal = 0
	ctx.Game.Log(fmt.Sprintf("%s 的 [新月庇护] 触发：进入暗月形态并吸收%d张爆牌为暗月，本次士气不下降",
		ctx.User.Name, added))
	return nil
}

func (h *MoonGoddessDarkMoonCurseHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MoonGoddessDarkMoonCurseHandler) Execute(ctx *model.Context) error { return nil }

func (h *MoonGoddessMedusaEyeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	if !ctx.AttackDeclaredPhase() {
		return false
	}
	attackerID := ctx.EventCtx.SourceID
	if attackerID == "" && ctx.Target != nil {
		attackerID = ctx.Target.ID
	}
	if attackerID == "" || attackerID == ctx.User.ID {
		return false
	}
	attacker := ctx.Game.GetPlayers()[attackerID]
	if attacker != nil && attacker.Camp == ctx.User.Camp {
		return false
	}
	return len(medusaSelectableDarkMoonIndices(ctx.User, ctx.EventCtx.Card.Element)) > 0
}

func (h *MoonGoddessMedusaEyeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return fmt.Errorf("美杜莎之眼上下文无效")
	}
	attackCard := ctx.EventCtx.Card
	selectable := medusaSelectableDarkMoonIndices(ctx.User, attackCard.Element)
	if len(selectable) == 0 {
		return fmt.Errorf("没有同系闇月可用于美杜莎之眼")
	}
	attackerID := ctx.EventCtx.SourceID
	if attackerID == "" && ctx.Target != nil {
		attackerID = ctx.Target.ID
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":      "mg_medusa_darkmoon_pick",
			"user_id":          ctx.User.ID,
			"attacker_id":      attackerID,
			"attack_element":   string(attackCard.Element),
			"darkmoon_indices": selectable,
			"user_ctx":         ctx,
			"source_skill":     ctx.Selections["source_skill"],
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [美杜莎之眼]：请选择要展示并移除的%s系闇月", ctx.User.Name, attackCard.Element))
	return nil
}

func (h *MoonGoddessMoonCycleHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MoonGoddessMoonCycleHandler) Execute(ctx *model.Context) error { return nil }

func (h *MoonGoddessBlasphemyHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MoonGoddessBlasphemyHandler) Execute(ctx *model.Context) error { return nil }

func (h *MoonGoddessDarkMoonSlashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
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
	if !engineplayer.HasForm(ctx.User, model.FormMoonGoddessDarkMoon) {
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
	if !engineplayer.HasForm(ctx.User, model.FormMoonGoddessDarkMoon) {
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
	if maxX < 1 {
		return fmt.Errorf("暗月不足，无法发动暗月斩")
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
	ctx.Game.Log(fmt.Sprintf("%s 发动 [暗月斩]：消耗1水晶，选择移除暗月数量X（1~%d）", ctx.User.Name, maxX))
	return nil
}

func (h *MoonGoddessPaleMoonHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	branch1 := skillGetToken(ctx.User, "mg_petrify") >= 3
	branch2 := skillGetToken(ctx.User, "mg_new_moon") >= 1 &&
		len(ctx.User.Hand) > 0 &&
		len(moonGoddessEnemyIDsFromGame(ctx.Game, ctx.User)) > 0
	return branch1 || branch2
}

func (h *MoonGoddessPaleMoonHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("苍白之月上下文无效")
	}
	var modes []string
	if skillGetToken(ctx.User, "mg_petrify") >= 3 {
		modes = append(modes, "branch1")
	}
	if skillGetToken(ctx.User, "mg_new_moon") >= 1 &&
		len(ctx.User.Hand) > 0 &&
		len(moonGoddessEnemyIDsFromGame(ctx.Game, ctx.User)) > 0 {
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
