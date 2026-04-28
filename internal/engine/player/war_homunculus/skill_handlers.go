// gameflow: 英灵人形技能处理器。

package war_homunculus

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"strings"
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
	return getToken(ctx.User, "hom_war_rune") > 0
}

func (h *HomunculusRageSuppressHandler) Execute(ctx *model.Context) error {
	if getToken(ctx.User, "hom_war_rune") <= 0 {
		return nil
	}
	addToken(ctx.User, "hom_war_rune", -1, 0, 99)
	addToken(ctx.User, "hom_magic_rune", 1, 0, 99)
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
	if getToken(ctx.User, "hom_war_rune") <= 0 {
		return false
	}
	if ctx.EventCtx.Card == nil {
		return false
	}
	ele := ctx.EventCtx.Card.Element
	sameCnt := 0
	for _, c := range ctx.User.Hand {
		if c.Element == ele {
			sameCnt++
		}
	}
	return sameCnt > 0
}

func (h *HomunculusRuneSmashHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return fmt.Errorf("战纹碎击上下文无效")
	}
	if getToken(ctx.User, "hom_war_rune") <= 0 {
		return fmt.Errorf("战纹不足")
	}
	attackEle := ctx.EventCtx.Card.Element
	var candidates []int
	for i, c := range ctx.User.Hand {
		if c.Element == attackEle {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) == 0 {
		return fmt.Errorf("没有可弃置的同系牌")
	}
	maxY := 0
	if hasForm(ctx.User, model.FormWarHomunculusBurst) {
		warRunes := getToken(ctx.User, "hom_war_rune")
		if warRunes > 1 {
			maxY = warRunes - 1
		}
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":       "hom_rune_smash_x",
			"user_id":           ctx.User.ID,
			"user_ctx":          ctx,
			"attack_element":    string(attackEle),
			"max_x":             len(candidates),
			"candidate_indices": candidates,
			"max_y":             maxY,
			"selected_indices":  []int{},
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [战纹碎击]，请选择X、弃牌与Y", ctx.User.Name))
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
	if getToken(ctx.User, "hom_magic_rune") <= 0 {
		return false
	}
	attackEle := model.Element("")
	if ctx.EventCtx.Card != nil {
		attackEle = ctx.EventCtx.Card.Element
	}
	uniqueElements := map[model.Element]bool{}
	for _, c := range ctx.User.Hand {
		if c.Element != attackEle {
			uniqueElements[c.Element] = true
		}
	}
	return len(uniqueElements) >= 2
}

func (h *HomunculusGlyphFusionHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return fmt.Errorf("魔纹融合上下文无效")
	}
	if getToken(ctx.User, "hom_magic_rune") <= 0 {
		return fmt.Errorf("魔纹不足")
	}
	attackEle := model.Element("")
	if ctx.EventCtx.Card != nil {
		attackEle = ctx.EventCtx.Card.Element
	}
	var candidates []int
	for i, c := range ctx.User.Hand {
		if c.Element != attackEle {
			candidates = append(candidates, i)
		}
	}
	uniqueElements := map[model.Element]bool{}
	for _, idx := range candidates {
		uniqueElements[ctx.User.Hand[idx].Element] = true
	}
	if len(uniqueElements) < 2 {
		return fmt.Errorf("异系牌不足2张")
	}
	maxX := len(uniqueElements)
	maxY := 0
	if hasForm(ctx.User, model.FormWarHomunculusBurst) {
		magicRunes := getToken(ctx.User, "hom_magic_rune")
		if magicRunes > 1 {
			maxY = magicRunes - 1
		}
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":       "hom_glyph_fusion_x",
			"user_id":           ctx.User.ID,
			"user_ctx":          ctx,
			"attack_element":    string(attackEle),
			"max_x":             maxX,
			"candidate_indices": candidates,
			"max_y":             maxY,
			"selected_indices":  []int{},
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔纹融合]，请选择X、弃牌与Y", ctx.User.Name))
	return nil
}

func (h *HomunculusRuneReforgeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && !hasForm(ctx.User, model.FormWarHomunculusBurst)
}

func (h *HomunculusRuneReforgeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("符文改造上下文无效")
	}
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("符文改造需要红宝石")
	}
	ctx.User.Gem--
	enterForm(ctx.User, model.FormWarHomunculusBurst)
	ctx.Game.DrawCards(ctx.User.ID, 1)
	totalRunes := getToken(ctx.User, "hom_war_rune") + getToken(ctx.User, "hom_magic_rune")
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
