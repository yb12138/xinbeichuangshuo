package skills

import (
	"fmt"
	"sort"
	"starcup-engine/internal/model"
)

const (
	magicBowChargeCap = 8
)

func magicBowChargeCovers(user *model.Player) []*model.FieldCard {
	if user == nil {
		return nil
	}
	var out []*model.FieldCard
	for _, fc := range user.Field {
		if fc == nil || fc.Mode != model.FieldCover {
			continue
		}
		if fc.Effect != model.EffectMagicBowCharge {
			continue
		}
		out = append(out, fc)
	}
	return out
}

func magicBowChargeCount(user *model.Player, element model.Element) int {
	count := 0
	for _, fc := range magicBowChargeCovers(user) {
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

func syncMagicBowChargeToken(user *model.Player) {
	if user == nil {
		return
	}
	setToken(user, "mb_charge_count", magicBowChargeCount(user, ""))
}

func removeMagicBowChargeByElement(user *model.Player, element model.Element) (model.Card, bool) {
	if user == nil {
		return model.Card{}, false
	}
	for _, fc := range user.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMagicBowCharge {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		card := fc.Card
		user.RemoveFieldCard(fc)
		syncMagicBowChargeToken(user)
		return card, true
	}
	return model.Card{}, false
}

func addMagicBowChargeCards(user *model.Player, cards []model.Card) int {
	if user == nil || len(cards) == 0 {
		return 0
	}
	room := magicBowChargeCap - magicBowChargeCount(user, "")
	if room <= 0 {
		return 0
	}
	added := 0
	for _, card := range cards {
		if added >= room {
			break
		}
		user.AddFieldCard(&model.FieldCard{
			Card:     card,
			OwnerID:  user.ID,
			SourceID: user.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectMagicBowCharge,
		})
		added++
	}
	syncMagicBowChargeToken(user)
	return added
}

func removeCardsByHandIndices(user *model.Player, indices []int) ([]model.Card, error) {
	if user == nil {
		return nil, fmt.Errorf("玩家不存在")
	}
	if len(indices) == 0 {
		return nil, nil
	}
	seen := map[int]bool{}
	for _, idx := range indices {
		if idx < 0 || idx >= len(user.Hand) {
			return nil, fmt.Errorf("无效的手牌索引: %d", idx)
		}
		if seen[idx] {
			return nil, fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	removed := make([]model.Card, 0, len(indices))
	for _, idx := range indices {
		removed = append(removed, user.Hand[idx])
		user.Hand = append(user.Hand[:idx], user.Hand[idx+1:]...)
	}
	return removed, nil
}

// --- 魔弓 ---

type MagicBowMagicPierceHandler struct{ BaseHandler }

type MagicBowThunderScatterHandler struct{ BaseHandler }

type MagicBowMultiShotHandler struct{ BaseHandler }

type MagicBowChargeHandler struct{ BaseHandler }

type MagicBowDemonEyeHandler struct{ BaseHandler }

// 内部回调：用于“充能”在弃牌后继续执行摸牌/置充能流程。
type MagicBowChargeFollowupDiscardHandler struct{ BaseHandler }

func (h *MagicBowMagicPierceHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["mb_multi_shot_used_turn"] > 0 {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["mb_charge_lock_turn"] > 0 {
		return false
	}
	if magicBowChargeCount(ctx.User, model.ElementFire) <= 0 {
		return false
	}
	// 魔贯冲击发动时，不能选择“手牌达到上限”的目标。
	if len(ctx.Target.Hand) >= ctx.Target.MaxHand {
		return false
	}
	return true
}

func (h *MagicBowMagicPierceHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.Card == nil {
		return fmt.Errorf("魔贯冲击上下文无效")
	}
	if _, ok := removeMagicBowChargeByElement(ctx.User, model.ElementFire); !ok {
		return fmt.Errorf("火系充能不足")
	}
	ctx.User.TurnState.UsedSkillCounts["mb_magic_pierce_used_turn"]++
	setToken(ctx.User, "mb_magic_pierce_pending", 1)
	ctx.TriggerCtx.Card.Damage++
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔贯冲击]：移除1个火系充能，本次攻击伤害+1", ctx.User.Name))
	return nil
}

func (h *MagicBowThunderScatterHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["mb_charge_lock_turn"] > 0 {
		return false
	}
	return magicBowChargeCount(ctx.User, model.ElementThunder) > 0
}

func (h *MagicBowThunderScatterHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("雷光散射上下文无效")
	}
	if ctx.User.TurnState.UsedSkillCounts["mb_charge_lock_turn"] > 0 {
		return fmt.Errorf("本回合已发动[充能]，不能发动雷光散射")
	}
	if _, ok := removeMagicBowChargeByElement(ctx.User, model.ElementThunder); !ok {
		return fmt.Errorf("雷系充能不足")
	}
	enemyIDs := make([]string, 0)
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil || p.Camp == ctx.User.Camp {
			continue
		}
		enemyIDs = append(enemyIDs, p.ID)
	}
	if len(enemyIDs) == 0 {
		ctx.Game.Log(fmt.Sprintf("%s 发动 [雷光散射]：无可选对手", ctx.User.Name))
		return nil
	}
	maxExtra := magicBowChargeCount(ctx.User, model.ElementThunder)
	if maxExtra <= 0 {
		for _, enemyID := range enemyIDs {
			ctx.Game.AddPendingDamage(model.PendingDamage{
				SourceID:   ctx.User.ID,
				TargetID:   enemyID,
				Damage:     1,
				DamageType: "magic",
				Stage:      0,
			})
		}
		ctx.Game.Log(fmt.Sprintf("%s 发动 [雷光散射]：对所有对手各造成1点法术伤害", ctx.User.Name))
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_thunder_scatter_extra",
			"user_id":     ctx.User.ID,
			"target_ids":  enemyIDs,
			"max_extra":   maxExtra,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [雷光散射]：可额外移除0~%d个雷系充能并指定目标", ctx.User.Name, maxExtra))
	return nil
}

func (h *MagicBowMultiShotHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnPhaseEnd {
		return false
	}
	if ctx.TriggerCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if ctx.User.TurnState.UsedSkillCounts["mb_magic_pierce_used_turn"] > 0 {
		return false
	}
	if magicBowChargeCount(ctx.User, model.ElementWind) <= 0 {
		return false
	}
	prevOrder := ctx.User.TurnState.UsedSkillCounts["mb_last_attack_target_order"]
	for i, p := range ctx.Game.GetAllPlayers() {
		if p == nil || p.Camp == ctx.User.Camp {
			continue
		}
		if prevOrder > 0 && prevOrder == i+1 {
			continue
		}
		return true
	}
	return false
}

func (h *MagicBowMultiShotHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("多重射击上下文无效")
	}
	enemyIDs := make([]string, 0)
	prevOrder := ctx.User.TurnState.UsedSkillCounts["mb_last_attack_target_order"]
	for i, p := range ctx.Game.GetAllPlayers() {
		if p == nil || p.Camp == ctx.User.Camp {
			continue
		}
		if prevOrder > 0 && prevOrder == i+1 {
			continue
		}
		enemyIDs = append(enemyIDs, p.ID)
	}
	if len(enemyIDs) == 0 {
		ctx.Game.Log(fmt.Sprintf("%s 发动 [多重射击] 失败：无可攻击目标", ctx.User.Name))
		return nil
	}
	if _, ok := removeMagicBowChargeByElement(ctx.User, model.ElementWind); !ok {
		return fmt.Errorf("风系充能不足")
	}
	ctx.User.TurnState.UsedSkillCounts["mb_multi_shot_used_turn"]++

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_multi_shot_target",
			"user_id":     ctx.User.ID,
			"target_ids":  enemyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [多重射击]：请选择暗系追加攻击目标", ctx.User.Name))
	return nil
}

func (h *MagicBowChargeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *MagicBowChargeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("充能上下文无效")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("充能需要1蓝水晶（红宝石可替代）")
	}
	ctx.User.TurnState.UsedSkillCounts["mb_charge_lock_turn"] = 1

	discardNeed := len(ctx.User.Hand) - 4
	if discardNeed < 0 {
		discardNeed = 0
	}
	if discardNeed > 0 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptDiscard,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"skill_id": "mb_charge_followup_discard",
				"user_ctx": ctx,
				"min":      discardNeed,
				"max":      discardNeed,
				"prompt":   fmt.Sprintf("【充能】请先弃置%d张手牌至4张：", discardNeed),
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [充能]：先弃至4张，再选择摸牌数量X（0~4）", ctx.User.Name))
		return nil
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_charge_draw_x",
			"user_id":     ctx.User.ID,
			"max_draw":    4,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [充能]：请选择摸牌数量X（0~4）", ctx.User.Name))
	return nil
}

func (h *MagicBowDemonEyeHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && ctx.User.Gem > 0
}

func (h *MagicBowDemonEyeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔眼上下文无效")
	}
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("魔眼需要1个红宝石")
	}
	ctx.User.Gem--
	targetIDs := make([]string, 0)
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		targetIDs = append(targetIDs, p.ID)
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_demon_eye_mode",
			"user_id":     ctx.User.ID,
			"target_ids":  targetIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔眼]：请选择分支", ctx.User.Name))
	return nil
}

func (h *MagicBowChargeFollowupDiscardHandler) CanUse(ctx *model.Context) bool { return false }

func (h *MagicBowChargeFollowupDiscardHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("充能后续上下文无效")
	}
	discardRaw, _ := ctx.Selections["discard_indices"]
	var discardIndices []int
	switch arr := discardRaw.(type) {
	case []int:
		discardIndices = append(discardIndices, arr...)
	case []interface{}:
		for _, v := range arr {
			if f, ok := v.(float64); ok {
				discardIndices = append(discardIndices, int(f))
			}
		}
	}
	if len(discardIndices) > 0 {
		removed, err := removeCardsByHandIndices(ctx.User, discardIndices)
		if err != nil {
			return err
		}
		ctx.Selections["discardedCards"] = removed
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_charge_draw_x",
			"user_id":     ctx.User.ID,
			"max_draw":    4,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 的 [充能] 已弃至4张，请选择摸牌数量X（0~4）", ctx.User.Name))
	return nil
}
