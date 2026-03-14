package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

const (
	holyBowFaithCap  = 10
	holyBowCannonCap = 1
)

type HolyBowHeavenlyBowHandler struct{ BaseHandler }

type HolyBowShardStormHandler struct{ BaseHandler }

type HolyBowRadiantDescentHandler struct{ BaseHandler }

type HolyBowLightBurstHandler struct{ BaseHandler }

type HolyBowMeteorBulletHandler struct{ BaseHandler }

type HolyBowRadiantCannonHandler struct{ BaseHandler }

type HolyBowAutoFillHandler struct{ BaseHandler }

func (h *HolyBowHeavenlyBowHandler) CanUse(ctx *model.Context) bool { return false }

func (h *HolyBowHeavenlyBowHandler) Execute(ctx *model.Context) error { return nil }

func (h *HolyBowAutoFillHandler) CanUse(ctx *model.Context) bool { return false }

func (h *HolyBowAutoFillHandler) Execute(ctx *model.Context) error { return nil }

func holyBowFaith(user *model.Player) int {
	if user == nil {
		return 0
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	v := user.Tokens["hb_faith"]
	if v < 0 {
		v = 0
	}
	if v > holyBowFaithCap {
		v = holyBowFaithCap
	}
	user.Tokens["hb_faith"] = v
	return v
}

func addHolyBowFaith(user *model.Player, delta int) int {
	if user == nil {
		return 0
	}
	return addToken(user, "hb_faith", delta, 0, holyBowFaithCap)
}

func holyBowCannon(user *model.Player) int {
	if user == nil {
		return 0
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	v := user.Tokens["hb_cannon"]
	if v < 0 {
		v = 0
	}
	if v > holyBowCannonCap {
		v = holyBowCannonCap
	}
	user.Tokens["hb_cannon"] = v
	return v
}

func holyBowPairCombos(user *model.Player) []string {
	if user == nil {
		return nil
	}
	elemToIdx := map[model.Element][]int{}
	for i, c := range user.Hand {
		if c.Type != model.CardTypeAttack || c.Element == "" {
			continue
		}
		elemToIdx[c.Element] = append(elemToIdx[c.Element], i)
	}
	order := []model.Element{
		model.ElementEarth, model.ElementWater, model.ElementFire,
		model.ElementWind, model.ElementThunder, model.ElementLight, model.ElementDark,
	}
	var combos []string
	for _, ele := range order {
		idxs := elemToIdx[ele]
		if len(idxs) < 2 {
			continue
		}
		for i := 0; i < len(idxs)-1; i++ {
			for j := i + 1; j < len(idxs); j++ {
				combos = append(combos, fmt.Sprintf("%s:%d,%d", ele, idxs[i], idxs[j]))
			}
		}
	}
	return combos
}

func holyBowAllies(game model.IGameEngine, user *model.Player, includeSelf bool) []string {
	if game == nil || user == nil {
		return nil
	}
	var ids []string
	for _, p := range game.GetAllPlayers() {
		if p == nil || p.Camp != user.Camp {
			continue
		}
		if !includeSelf && p.ID == user.ID {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

func holyBowEnemies(game model.IGameEngine, user *model.Player) []string {
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

func holyBowMoraleGap(game model.IGameEngine, user *model.Player) int {
	if game == nil || user == nil {
		return 0
	}
	selfMorale := game.GetCampMorale(string(user.Camp))
	enemyCamp := model.RedCamp
	if user.Camp == model.RedCamp {
		enemyCamp = model.BlueCamp
	}
	enemyMorale := game.GetCampMorale(string(enemyCamp))
	if enemyMorale <= selfMorale {
		return 0
	}
	return enemyMorale - selfMorale
}

func holyBowCanUseLightBurstModeA(user *model.Player) bool {
	if user == nil {
		return false
	}
	return user.Heal >= 1
}

func holyBowCanUseLightBurstModeB(game model.IGameEngine, user *model.Player, enemyIDs []string, maxX int) bool {
	if game == nil || user == nil || maxX <= 0 || len(enemyIDs) == 0 {
		return false
	}
	enemies := map[string]*model.Player{}
	for _, p := range game.GetAllPlayers() {
		if p == nil {
			continue
		}
		enemies[p.ID] = p
	}
	handCount := len(user.Hand)
	for x := 1; x <= maxX; x++ {
		limit := handCount - x
		for _, eid := range enemyIDs {
			if ep := enemies[eid]; ep != nil && len(ep.Hand) <= limit {
				return true
			}
		}
	}
	return false
}

func (h *HolyBowShardStormHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	if len(holyBowPairCombos(ctx.User)) == 0 {
		return false
	}
	return len(holyBowEnemies(ctx.Game, ctx.User)) > 0
}

func (h *HolyBowShardStormHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("圣屑飓暴上下文无效")
	}
	combos := holyBowPairCombos(ctx.User)
	if len(combos) == 0 {
		return fmt.Errorf("没有可弃置的同系攻击牌组合")
	}
	enemyIDs := holyBowEnemies(ctx.Game, ctx.User)
	if len(enemyIDs) == 0 {
		return fmt.Errorf("没有可选敌方目标")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "hb_holy_shard_combo",
			"user_id":     ctx.User.ID,
			"combos":      combos,
			"target_ids":  enemyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣屑飓暴]：请选择弃置的同系攻击牌组合", ctx.User.Name))
	return nil
}

func (h *HolyBowRadiantDescentHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if getToken(ctx.User, "hb_form") > 0 {
		return false
	}
	return ctx.User.Heal >= 2 || holyBowFaith(ctx.User) >= 2
}

func (h *HolyBowRadiantDescentHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("圣煌降临上下文无效")
	}
	if getToken(ctx.User, "hb_form") > 0 {
		return fmt.Errorf("已处于圣煌形态")
	}
	var costModes []string
	if ctx.User.Heal >= 2 {
		costModes = append(costModes, "heal")
	}
	if holyBowFaith(ctx.User) >= 2 {
		costModes = append(costModes, "faith")
	}
	if len(costModes) == 0 {
		return fmt.Errorf("治疗与信仰均不足2，无法发动圣煌降临")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "hb_radiant_descent_cost",
			"user_id":     ctx.User.ID,
			"cost_modes":  costModes,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣煌降临]：请选择支付方式", ctx.User.Name))
	return nil
}

func (h *HolyBowLightBurstHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	if getToken(ctx.User, "hb_form") <= 0 {
		return false
	}
	enemyIDs := holyBowEnemies(ctx.Game, ctx.User)
	maxX := ctx.User.Heal
	if len(ctx.User.Hand) < maxX {
		maxX = len(ctx.User.Hand)
	}
	return holyBowCanUseLightBurstModeA(ctx.User) || holyBowCanUseLightBurstModeB(ctx.Game, ctx.User, enemyIDs, maxX)
}

func (h *HolyBowLightBurstHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("圣光爆裂上下文无效")
	}
	if getToken(ctx.User, "hb_form") <= 0 {
		return fmt.Errorf("仅圣煌形态可发动圣光爆裂")
	}
	allyIDs := holyBowAllies(ctx.Game, ctx.User, true)
	if len(allyIDs) == 0 {
		allyIDs = append(allyIDs, ctx.User.ID)
	}
	enemyIDs := holyBowEnemies(ctx.Game, ctx.User)
	maxX := ctx.User.Heal
	if len(ctx.User.Hand) < maxX {
		maxX = len(ctx.User.Hand)
	}
	if !holyBowCanUseLightBurstModeA(ctx.User) && !holyBowCanUseLightBurstModeB(ctx.Game, ctx.User, enemyIDs, maxX) {
		return fmt.Errorf("当前不满足圣光爆裂任一分支发动条件")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "hb_light_burst_mode",
			"user_id":     ctx.User.ID,
			"ally_ids":    allyIDs,
			"enemy_ids":   enemyIDs,
			"max_x":       maxX,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣光爆裂]：请选择发动分支", ctx.User.Name))
	return nil
}

func (h *HolyBowMeteorBulletHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	if getToken(ctx.User, "hb_form") <= 0 {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if ctx.User.Heal <= 0 && holyBowFaith(ctx.User) <= 0 {
		return false
	}
	return len(holyBowAllies(ctx.Game, ctx.User, true)) > 0
}

func (h *HolyBowMeteorBulletHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("流星圣弹上下文无效")
	}
	if getToken(ctx.User, "hb_form") <= 0 {
		return fmt.Errorf("仅圣煌形态可发动流星圣弹")
	}
	var costModes []string
	if ctx.User.Heal > 0 {
		costModes = append(costModes, "heal")
	}
	if holyBowFaith(ctx.User) > 0 {
		costModes = append(costModes, "faith")
	}
	if len(costModes) == 0 {
		return fmt.Errorf("治疗与信仰均不足，无法发动流星圣弹")
	}
	allyIDs := holyBowAllies(ctx.Game, ctx.User, true)
	if len(allyIDs) == 0 {
		return fmt.Errorf("没有可选我方目标")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "hb_meteor_bullet_cost",
			"user_id":     ctx.User.ID,
			"cost_modes":  costModes,
			"ally_ids":    allyIDs,
			"user_ctx":    ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [流星圣弹]：请选择移除资源并指定我方目标", ctx.User.Name))
	return nil
}

func (h *HolyBowRadiantCannonHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	if getToken(ctx.User, "hb_form") <= 0 {
		return false
	}
	if holyBowCannon(ctx.User) <= 0 {
		return false
	}
	requiredFaith := 4 + holyBowMoraleGap(ctx.Game, ctx.User)
	return holyBowFaith(ctx.User) >= requiredFaith
}

func (h *HolyBowRadiantCannonHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("圣煌辉光炮上下文无效")
	}
	if getToken(ctx.User, "hb_form") <= 0 {
		return fmt.Errorf("仅圣煌形态可发动圣煌辉光炮")
	}
	if holyBowCannon(ctx.User) <= 0 {
		return fmt.Errorf("圣煌辉光炮指示物不足")
	}
	requiredFaith := 4 + holyBowMoraleGap(ctx.Game, ctx.User)
	if holyBowFaith(ctx.User) < requiredFaith {
		return fmt.Errorf("信仰不足，需要%d点", requiredFaith)
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":    "hb_radiant_cannon_side",
			"user_id":        ctx.User.ID,
			"required_faith": requiredFaith,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣煌辉光炮]：请选择士气对齐方向", ctx.User.Name))
	return nil
}
