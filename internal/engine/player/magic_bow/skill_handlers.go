// gameflow: 魔弓手 handler。

package magic_bow

import (
	"fmt"
	"sort"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
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

func magicBowChargeFieldIndices(user *model.Player, element model.Element) []int {
	if user == nil {
		return nil
	}
	out := make([]int, 0)
	for idx, fc := range user.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMagicBowCharge {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		out = append(out, idx)
	}
	return out
}

func syncMagicBowChargeToken(user *model.Player) {
	// no-op: mb_charge_count 在服务端 buildStateForPlayer 中按场上盖牌派生写入 PlayerView.indicators
}

func removeMagicBowChargeAtFieldIndex(user *model.Player, fieldIndex int, element model.Element) (model.Card, bool) {
	if user == nil || fieldIndex < 0 || fieldIndex >= len(user.Field) {
		return model.Card{}, false
	}
	fc := user.Field[fieldIndex]
	if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMagicBowCharge {
		return model.Card{}, false
	}
	if element != "" && fc.Card.Element != element {
		return model.Card{}, false
	}
	card := fc.Card
	user.RemoveFieldCard(fc)
	syncMagicBowChargeToken(user)
	return card, true
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

func magicBowAllHandIndices(user *model.Player) []int {
	if user == nil {
		return nil
	}
	out := make([]int, 0, len(user.Hand))
	for i := range user.Hand {
		out = append(out, i)
	}
	return out
}

// --- 魔弓 ---

type MagicBowMagicPierceHandler struct{ engineplayer.BaseHandler }

type MagicBowThunderScatterHandler struct{ engineplayer.BaseHandler }

type MagicBowMultiShotHandler struct{ engineplayer.BaseHandler }

type MagicBowChargeHandler struct{ engineplayer.BaseHandler }

type MagicBowDemonEyeHandler struct{ engineplayer.BaseHandler }

func (h *MagicBowMagicPierceHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Target == nil || ctx.EventCtx == nil {
		return false
	}
	if !ctx.AttackDeclarePhase() {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
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
	// 魔贯冲击发动时，不能选择"手牌达到上限"的目标。
	if len(ctx.Target.Hand) >= ctx.Target.MaxHand {
		return false
	}
	return true
}

func (h *MagicBowMagicPierceHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return fmt.Errorf("魔贯冲击上下文无效")
	}
	if magicBowChargeCount(ctx.User, model.ElementFire) <= 0 {
		return fmt.Errorf("火系充能不足")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_magic_pierce_charge",
			"user_id":     ctx.User.ID,
			"user_ctx":    ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 确认发动 [魔贯冲击]：请选择移除1个火系充能", ctx.User.Name))
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
	if magicBowChargeCount(ctx.User, model.ElementThunder) <= 0 {
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
	lockedTargetID := ""
	if ctx.Target != nil {
		if ctx.Target.Camp == ctx.User.Camp || ctx.Target.ID == ctx.User.ID {
			return fmt.Errorf("雷光散射的额外目标必须是敌方角色")
		}
		lockedTargetID = ctx.Target.ID
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":      "mb_thunder_scatter_base_charge",
			"user_id":          ctx.User.ID,
			"target_ids":       enemyIDs,
			"locked_target_id": lockedTargetID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [雷光散射]：请选择移除1个雷系充能", ctx.User.Name))
	return nil
}

func (h *MagicBowMultiShotHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingActionEnd {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
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
	if magicBowChargeCount(ctx.User, model.ElementWind) <= 0 {
		return fmt.Errorf("风系充能不足")
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_multi_shot_charge",
			"user_id":     ctx.User.ID,
			"target_ids":  enemyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 确认发动 [多重射击]：请选择移除1个风系充能", ctx.User.Name))
	return nil
}

func (h *MagicBowChargeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil {
		return false
	}
	// 启动技在 ActionStart 阶段检查可用性时，
	// runtime 会先检查 skillDef.CostCrystal，确认玩家有足够资源后才会创建中断。
	// CanUse 不需要重复检查资源（否则会导致红宝石替代水晶的判定与 runtime 不一致）。
	// 对于响应技等其他时点，依赖 CanPayCrystalLike 检查。
	if ctx.Timing == model.TimingActionStart || ctx.Timing == model.TimingActive {
		return true
	}
	return engineplayer.CanPayCrystalLike(ctx, 1)
}

func (h *MagicBowChargeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("充能上下文无效")
	}
	// 启动技和行动技的能耗已由 runtime/UseSkill 流程在调用 Execute 前扣减，
	// 仅响应技等其他时点需要 handler 内自行扣减。
	if ctx.Timing != model.TimingActionStart && ctx.Timing != model.TimingActive && !engineplayer.SpendCrystalLike(ctx, 1) {
		return fmt.Errorf("充能需要1蓝水晶（红宝石可替代）")
	}
	ctx.User.TurnState.UsedSkillCounts["mb_charge_lock_turn"] = 1

	if len(ctx.User.Hand) > 4 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type":                 "system_discard_cards",
				"discard_subflow":             true,
				"flow_continuation_role_id":   "magic_bow",
				"flow_continuation_player_id": ctx.User.ID,
				"flow_continuation_skill_id":  "mb_charge",
				"discard_down_to":             4,
				"stay_in_turn":                true,
				"discard_forced":              true,
				"forced_reason":               "【充能】弃牌为强制步骤，不能取消",
				"prompt":                      "【充能】请先弃置手牌至4张：",
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
	return ctx != nil && ctx.User != nil && ctx.User.Gem > 0 && len(ctx.User.Hand) > 0
}

func (h *MagicBowDemonEyeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔眼上下文无效")
	}
	if len(ctx.User.Hand) == 0 {
		return fmt.Errorf("魔眼需要至少1张手牌作为充能")
	}
	if ctx.Timing != model.TimingActive && ctx.User.Gem <= 0 {
		return fmt.Errorf("魔眼需要1个红宝石")
	}
	if ctx.Timing != model.TimingActive {
		ctx.User.Gem--
	}
	// Build target list for branch 1 (includes self)
	targetIDs := make([]string, 0)
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		targetIDs = append(targetIDs, p.ID)
	}
	// Push branch selection first
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_demon_eye_mode",
			"user_id":     ctx.User.ID,
			"target_ids":  targetIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔眼]，请选择分支", ctx.User.Name))
	return nil
}
