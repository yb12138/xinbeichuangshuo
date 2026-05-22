// gameflow: 天使技能处理器。

package angel

import (
	"fmt"
	"strings"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type BaseHandler = engineplayer.BaseHandler

// --- Angel Helpers ---

type basicEffectOption struct {
	TargetID    string
	TargetName  string
	FieldIndex  int
	Effect      model.EffectType
	DisplayName string
	Label       string
}

func basicEffectLabel(effect model.EffectType) string {
	switch effect {
	case model.EffectShield:
		return "圣盾"
	case model.EffectWeak:
		return "虚弱"
	case model.EffectPoison:
		return "中毒"
	case model.EffectSealFire:
		return "火之封印"
	case model.EffectSealWater:
		return "水之封印"
	case model.EffectSealEarth:
		return "地之封印"
	case model.EffectSealWind:
		return "风之封印"
	case model.EffectSealThunder:
		return "雷之封印"
	case model.EffectPowerBlessing:
		return "威力赐福"
	case model.EffectSwiftBlessing:
		return "迅捷赐福"
	default:
		return string(effect)
	}
}

func collectBasicEffectOptions(players ...*model.Player) []basicEffectOption {
	options := make([]basicEffectOption, 0)
	for _, player := range players {
		if player == nil {
			continue
		}
		for idx, fc := range player.Field {
			if fc == nil || fc.Mode != model.FieldEffect || !model.IsBasicEffect(string(fc.Effect)) {
				continue
			}
			displayName := basicEffectLabel(fc.Effect)
			options = append(options, basicEffectOption{
				TargetID:    player.ID,
				TargetName:  player.Name,
				FieldIndex:  idx,
				Effect:      fc.Effect,
				DisplayName: displayName,
				Label:       fmt.Sprintf("%s：%s", player.Name, displayName),
			})
		}
	}
	return options
}

func encodeBasicEffectOptions(options []basicEffectOption) []map[string]interface{} {
	encoded := make([]map[string]interface{}, 0, len(options))
	for _, option := range options {
		encoded = append(encoded, map[string]interface{}{
			"id":           fmt.Sprintf("%s|%d|%s", option.TargetID, option.FieldIndex, option.Effect),
			"target_id":    option.TargetID,
			"field_index":  option.FieldIndex,
			"effect":       string(option.Effect),
			"display_name": option.DisplayName,
			"label":        option.Label,
		})
	}
	return encoded
}

// --- Angel Skill Handlers ---

type HolyShieldHandler struct{}

func (h *HolyShieldHandler) CanUse(ctx *model.Context) bool {
	if !ctx.DamageTakenPhase() {
		return false
	}
	if ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	if ctx.Flags["IsMagicDamage"] {
		if ctx.EventCtx.Card == nil || strings.TrimSpace(ctx.EventCtx.Card.Name) != "魔弹" {
			return false
		}
	} else {
		return false
	}
	if ctx.Flags["ignore_shield"] {
		return false
	}
	hasShield := false
	for _, fc := range ctx.User.Field {
		if fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			hasShield = true
			break
		}
	}
	return hasShield
}

func (h *HolyShieldHandler) Execute(ctx *model.Context) error {
	originalDamage := *ctx.EventCtx.DamageVal
	*ctx.EventCtx.DamageVal = 0

	if ctx.Selections != nil {
		ctx.Selections["holy_shield_applied"] = true
	}
	ctx.Game.Log(fmt.Sprintf("[Shield] %s 的【圣盾】自动触发，抵消了 %d 点伤害！", ctx.User.Name, originalDamage))
	ctx.Game.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】触发，抵消了 %d 点伤害", ctx.User.Name, originalDamage))

	newField := make([]*model.FieldCard, 0)
	removed := false
	for _, fc := range ctx.User.Field {
		if !removed && fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			removed = true
			ctx.Game.DiscardCard(fc)
			continue
		}
		newField = append(newField, fc)
	}
	ctx.User.Field = newField
	return nil
}

type AngelBondHandler struct{ BaseHandler }

func (h *AngelBondHandler) CanUse(ctx *model.Context) bool {
	if ctx.EventCtx == nil || ctx.User == nil {
		return false
	}
	if ctx.Timing == model.TimingOnFieldMarkChanged {
		if ctx.EventCtx.SourceID != ctx.User.ID {
			return false
		}
		return model.IsBasicEffect(ctx.EventCtx.BuffID)
	}
	return false
}

func (h *AngelBondHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return nil
	}
	targetIDs := make([]string, 0)
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
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
			"choice_type":  "angel_bond_heal_target",
			"user_id":      ctx.User.ID,
			"target_ids":   targetIDs,
			"buff_name":    ctx.EventCtx.BuffID,
			"resume_phase": ctx.Selections["current_resume_point"],
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 的 [天使羁绊] 触发：请选择1名角色获得+1治疗", ctx.User.Name))
	return nil
}

type AngelBlessingHandler struct{ BaseHandler }

func (h *AngelBlessingHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	for _, card := range ctx.User.Hand {
		if card.Element == model.ElementWater {
			return true
		}
	}
	return false
}

func (h *AngelBlessingHandler) Execute(ctx *model.Context) error {
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) == 0 {
		return fmt.Errorf("天使祝福需要指定目标")
	}

	receiverID := ctx.User.ID

	if len(targets) == 1 {
		target := targets[0]
		giveCount := 2
		if len(target.Hand) < giveCount {
			giveCount = len(target.Hand)
		}
		if giveCount > 0 {
			ctx.Game.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptGiveCards,
				PlayerID: target.ID,
				Context: map[string]interface{}{
					"give_count":   giveCount,
					"receiver_id":  receiverID,
					"stay_in_turn": true,
					"resume_phase": ctx.Selections["current_resume_point"],
				},
			})
			ctx.Game.Log(fmt.Sprintf("%s 发动天使祝福，%s 需选择 %d 张牌交给 %s", ctx.User.Name, target.Name, giveCount, ctx.User.Name))
		} else {
			ctx.Game.Log(fmt.Sprintf("%s 发动天使祝福，但 %s 没有手牌可交", ctx.User.Name, target.Name))
		}
	} else if len(targets) == 2 {
		for i := len(targets) - 1; i >= 0; i-- {
			t := targets[i]
			if len(t.Hand) >= 1 {
				ctx.Game.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptGiveCards,
					PlayerID: t.ID,
					Context: map[string]interface{}{
						"give_count":   1,
						"receiver_id":  receiverID,
						"stay_in_turn": true,
						"resume_phase": ctx.Selections["current_resume_point"],
					},
				})
			} else {
				ctx.Game.Log(fmt.Sprintf("%s 没有手牌可交给 %s", t.Name, ctx.User.Name))
			}
		}
		ctx.Game.Log(fmt.Sprintf("%s 发动天使祝福，%s 和 %s 需各选择 1 张牌交给 %s",
			ctx.User.Name, targets[0].Name, targets[1].Name, ctx.User.Name))
	} else {
		return fmt.Errorf("天使祝福只能指定 1 名或 2 名目标")
	}
	return nil
}

type AngelCleanseHandler struct{ BaseHandler }

func (h *AngelCleanseHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	for _, card := range ctx.User.Hand {
		if card.Element == model.ElementWind {
			return true
		}
	}
	return false
}

func (h *AngelCleanseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("风之洁净上下文无效")
	}
	options := collectBasicEffectOptions(ctx.Game.GetAllPlayers()...)
	if len(options) == 0 {
		ctx.Game.Log(fmt.Sprintf("%s 的 [风之洁净] 发动：场上没有可移除的基础效果，跳过移除", ctx.User.Name))
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "basic_effect_pick",
			"user_id":       ctx.User.ID,
			"skill_name":    "风之洁净",
			"operation":     "remove",
			"cancel_policy": "decline",
			"resume_phase":  model.TurnStageActionExecution,
			"prompt":        "【风之洁净】请选择要移除的基础效果：",
			"options":       encodeBasicEffectOptions(options),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [风之洁净]，请选择要移除的基础效果", ctx.User.Name))
	return nil
}

type AngelSongHandler struct{ BaseHandler }

func (h *AngelSongHandler) CanUse(ctx *model.Context) bool {
	if !engineplayer.CanPayCrystalLike(ctx, 1) {
		return false
	}
	if ctx == nil || ctx.Game == nil {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc.Mode == model.FieldEffect && model.IsBasicEffect(string(fc.Effect)) {
				return true
			}
		}
	}
	return false
}

func (h *AngelSongHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("天使之歌上下文无效")
	}
	options := collectBasicEffectOptions(ctx.Game.GetAllPlayers()...)
	if len(options) == 0 {
		return fmt.Errorf("发动天使之歌失败：场上没有可移除的基础效果")
	}
	if !engineplayer.SpendCrystalLike(ctx, 1) {
		return fmt.Errorf("发动天使之歌失败：水晶不足（红宝石可替代）")
	}
	ctx.Game.Log(fmt.Sprintf("%s 消耗 1 水晶（可由红宝石替代）发动 [天使之歌]，请选择要移除的基础效果", ctx.User.Name))
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "basic_effect_pick",
			"user_id":       ctx.User.ID,
			"skill_name":    "天使之歌",
			"operation":     "remove",
			"resume_phase":  model.TurnStageActionStart,
			"waiting_phase": model.TurnStageActionStart,
			"prompt":        "【天使之歌】请选择要移除的基础效果：",
			"options":       encodeBasicEffectOptions(options),
		},
	})
	return nil
}

type GodProtectionHandler struct{ BaseHandler }

func (h *GodProtectionHandler) CanUse(ctx *model.Context) bool {
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	if !engineplayer.CanPayCrystalLike(ctx, 1) {
		return false
	}
	if ctx.EventCtx.DamageVal == nil || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	return true
}

func (h *GodProtectionHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return nil
	}
	angel := ctx.User
	loss := *ctx.EventCtx.DamageVal
	usable := ctx.Game.GetUsableCrystal(angel.ID)
	maxX := loss
	if maxX > usable {
		maxX = usable
	}
	if maxX <= 0 {
		return nil
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: angel.ID,
		Context: map[string]interface{}{
			"choice_type": "god_protection_x",
			"user_id":     angel.ID,
			"max_x":       maxX,
			"user_ctx":    ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 触发 [神之庇护]：请选择要抵御的士气下降值（最多%d）", angel.Name, maxX))
	return nil
}

type AngelWallHandler struct{ BaseHandler }

func (h *AngelWallHandler) Execute(ctx *model.Context) error {
	targetName := ctx.Target.Name
	if ctx.User.ID == ctx.Target.ID {
		ctx.Game.Log(fmt.Sprintf("%s 发动 [天使之墙]，自己获得圣盾保护", ctx.User.Name))
	} else {
		ctx.Game.Log(fmt.Sprintf("%s 发动 [天使之墙]，给 %s 提供圣盾保护", ctx.User.Name, targetName))
	}
	return nil
}
