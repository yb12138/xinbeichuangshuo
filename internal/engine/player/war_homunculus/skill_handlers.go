// gameflow: 英灵人形技能处理器。

package war_homunculus

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"strings"
)

// --- Helper functions ---

func canPayCrystalLike(ctx *model.Context, amount int) bool {
	return engineplayer.CanPayCrystalLike(ctx, amount)
}

func spendCrystalLike(ctx *model.Context, amount int) bool {
	return engineplayer.SpendCrystalLike(ctx, amount)
}

// --- 20. 英灵人形 ---

type HomunculusBattlePatternHandler struct{ engineplayer.BaseHandler }

type HomunculusRageSuppressHandler struct{ engineplayer.BaseHandler }

type HomunculusRuneSmashHandler struct{ engineplayer.BaseHandler }

type HomunculusGlyphFusionHandler struct{ engineplayer.BaseHandler }

type HomunculusRuneReforgeHandler struct{ engineplayer.BaseHandler }

type HomunculusDualEchoHandler struct{ engineplayer.BaseHandler }

func (h *HomunculusBattlePatternHandler) Execute(ctx *model.Context) error { return nil }

func (h *HomunculusRageSuppressHandler) CanUse(ctx *model.Context) bool {
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
	if ctx.EventCtx.AttackInfo.IsHit {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return engineplayer.GetToken(ctx.User, "hom_war_rune") > 0
}

func (h *HomunculusRageSuppressHandler) Execute(ctx *model.Context) error {
	if engineplayer.GetToken(ctx.User, "hom_war_rune") <= 0 {
		return nil
	}
	engineplayer.AddToken(ctx.User, "hom_war_rune", -1, 99)
	engineplayer.AddToken(ctx.User, "hom_magic_rune", 1, 99)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [怒火压制]，翻转1战纹为魔纹", ctx.User.Name))
	return nil
}

func (h *HomunculusRuneSmashHandler) CanUse(ctx *model.Context) bool {
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
	if engineplayer.GetToken(ctx.User, "hom_war_rune") <= 0 {
		return false
	}
	// 战纹碎击：只要有手牌就可以发动，选择的牌彼此同系即可（不要求与攻击牌同系）
	return len(ctx.User.Hand) > 0
}

func (h *HomunculusRuneSmashHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return fmt.Errorf("战纹碎击上下文无效")
	}
	if engineplayer.GetToken(ctx.User, "hom_war_rune") <= 0 {
		return fmt.Errorf("战纹不足")
	}
	// 战纹碎击：所有手牌都可作为候选，选择的牌彼此同系即可
	var candidates []int
	for i := range ctx.User.Hand {
		candidates = append(candidates, i)
	}
	if len(candidates) == 0 {
		return fmt.Errorf("没有可弃置的手牌")
	}
	maxY := 0
	if engineplayer.HasForm(ctx.User, model.FormWarHomunculusBurst) {
		warRunes := engineplayer.GetToken(ctx.User, "hom_war_rune")
		if warRunes > 1 {
			maxY = warRunes - 1
		}
	}
	// 直接弹出选牌 interrupt，跳过 X 数值选择
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":              "hom_rune_smash_cards",
			"user_id":                  ctx.User.ID,
			"user_ctx":                 ctx,
			"attack_element":           "", // 不再限制与攻击牌同系
			"candidate_indices":        candidates,
			"max_y":                    maxY,
			"min_pick":                 1,
			model.PromptFlowContextKey: model.NewPromptFlowState(runeSmashFlowID, runeChoiceStepCards),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [战纹碎击]，请选择要弃置的同系牌（所选牌彼此同系）", ctx.User.Name))
	return nil
}

func (h *HomunculusGlyphFusionHandler) CanUse(ctx *model.Context) bool {
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
	if ctx.EventCtx.AttackInfo.IsHit {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if engineplayer.GetToken(ctx.User, "hom_magic_rune") <= 0 {
		return false
	}
	// 魔纹融合：只要有至少2张手牌就可以发动，选择的牌彼此异系即可（不要求与攻击牌异系）
	// 至少需要2张牌才能选择彼此异系的牌
	return len(ctx.User.Hand) >= 2
}

func (h *HomunculusGlyphFusionHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return fmt.Errorf("魔纹融合上下文无效")
	}
	if engineplayer.GetToken(ctx.User, "hom_magic_rune") <= 0 {
		return fmt.Errorf("魔纹不足")
	}
	// 魔纹融合：所有手牌都可作为候选，选择的牌彼此异系即可
	var candidates []int
	for i := range ctx.User.Hand {
		candidates = append(candidates, i)
	}
	if len(candidates) < 2 {
		return fmt.Errorf("手牌不足2张")
	}
	maxY := 0
	if engineplayer.HasForm(ctx.User, model.FormWarHomunculusBurst) {
		magicRunes := engineplayer.GetToken(ctx.User, "hom_magic_rune")
		if magicRunes > 1 {
			maxY = magicRunes - 1
		}
	}
	// 直接弹出选牌 interrupt，跳过 X 数值选择
	// 魔纹融合至少需要2张异系牌（元素互不相同）
	minPick := 2
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":              "hom_glyph_fusion_cards",
			"user_id":                  ctx.User.ID,
			"user_ctx":                 ctx,
			"attack_element":           "", // 不再限制与攻击牌异系
			"candidate_indices":        candidates,
			"max_y":                    maxY,
			"min_pick":                 minPick,
			model.PromptFlowContextKey: model.NewPromptFlowState(glyphFusionFlowID, runeChoiceStepCards),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔纹融合]，请选择要弃置的异系牌（所选牌彼此异系）", ctx.User.Name))
	return nil
}

func (h *HomunculusRuneReforgeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && !engineplayer.HasForm(ctx.User, model.FormWarHomunculusBurst)
}

func (h *HomunculusRuneReforgeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("符文改造上下文无效")
	}
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("符文改造需要红宝石")
	}
	ctx.User.Gem--
	engineplayer.SetForm(ctx.User, model.FormWarHomunculusBurst)
	ctx.Game.DrawCards(ctx.User.ID, 1)
	totalRunes := engineplayer.GetToken(ctx.User, "hom_war_rune") + engineplayer.GetToken(ctx.User, "hom_magic_rune")
	if totalRunes <= 0 {
		totalRunes = 3
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "hom_rune_reforge_distribution",
			"user_id":     ctx.User.ID,
			"total_runes": totalRunes,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [符文改造]，进入蓄势迸发形态并摸1张牌，请调整战纹/魔纹分配", ctx.User.Name))
	return nil
}

func (h *HomunculusDualEchoHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnDamageTaken {
		return false
	}
	if ctx.EventCtx.SourceID != ctx.User.ID {
		return false
	}
	if damageType, _ := ctx.Selections["damage_type"].(string); damageType != "" {
		lower := strings.ToLower(strings.TrimSpace(damageType))
		if lower != "attack" && !strings.Contains(lower, "magic") {
			return false
		}
	}
	if ctx.EventCtx.DamageVal == nil || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *HomunculusDualEchoHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return fmt.Errorf("双重回响上下文无效")
	}
	if !canPayCrystalLike(ctx, 1) {
		return fmt.Errorf("双重回响需要1蓝水晶（红宝石可替代）")
	}
	damage := *ctx.EventCtx.DamageVal
	if damage > 3 {
		damage = 3
	}
	if damage <= 0 {
		return nil
	}
	var targetIDs []string
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil || p.ID == ctx.EventCtx.TargetID {
			continue
		}
		targetIDs = append(targetIDs, p.ID)
	}
	if len(targetIDs) == 0 {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "hom_dual_echo_target",
			"user_id":     ctx.User.ID,
			"target_ids":  targetIDs,
			"damage":      damage,
			// 成本在最终选定目标后再扣除，便于在目标弹框中取消本次响应。
			"cost_pending": 1,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [双重回响]，请选择追加伤害目标", ctx.User.Name))
	return nil
}
