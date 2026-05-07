// gameflow: 女武神技能处理器。

package valkyrie

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- 女武神技能处理器 ---

type ValkyrieDivinePursuitHandler struct{ engineplayer.BaseHandler }

func (h *ValkyrieDivinePursuitHandler) CanUse(ctx *model.Context) bool {
	if ctx.Timing != model.TimingOnActionEnd || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack && ctx.EventCtx.ActionType != model.ActionMagic {
		return false
	}
	// 攻击行动仅指主动攻击；应战攻击结束不触发神圣追击。
	if ctx.EventCtx.ActionType == model.ActionAttack &&
		ctx.EventCtx.AttackInfo != nil &&
		ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return ctx.User.Heal > 0
}

func (h *ValkyrieDivinePursuitHandler) Execute(ctx *model.Context) error {
	ctx.User.Heal--
	model.AppendAttackAction(ctx.User, "神圣追击")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [神圣追击]，移除1点治疗并获得额外攻击行动", ctx.User.Name))
	return nil
}

type ValkyrieOrderSealHandler struct{ engineplayer.BaseHandler }

func (h *ValkyrieOrderSealHandler) Execute(ctx *model.Context) error {
	ctx.Game.DrawCards(ctx.User.ID, 2)
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.User.Crystal++
	ctx.Game.Log(fmt.Sprintf("%s 发动 [秩序之印]，摸2并获得1治疗+1蓝水晶", ctx.User.Name))
	return nil
}

type ValkyriePeaceWalkerHandler struct{ engineplayer.BaseHandler }

func (h *ValkyriePeaceWalkerHandler) CanUse(ctx *model.Context) bool {
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return engineplayer.HasForm(ctx.User, model.FormValkyrieHeroic)
}

func (h *ValkyriePeaceWalkerHandler) Execute(ctx *model.Context) error {
	if !engineplayer.HasForm(ctx.User, model.FormValkyrieHeroic) {
		return nil
	}
	engineplayer.ClearForm(ctx.User, model.FormValkyrieHeroic)
	ctx.Game.Log(fmt.Sprintf("%s 的 [和平行者] 触发，脱离英灵形态", ctx.User.Name))
	return nil
}

type ValkyrieMilitaryGloryHandler struct{ engineplayer.BaseHandler }

func (h *ValkyrieMilitaryGloryHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.Timing == model.TimingOnTurnStart && engineplayer.HasForm(ctx.User, model.FormValkyrieHeroic)
}

func (h *ValkyrieMilitaryGloryHandler) Execute(ctx *model.Context) error {
	camp := string(ctx.User.Camp)
	energy := ctx.Game.GetCampCrystals(camp) + ctx.Game.GetCampGems(camp)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "valkyrie_military_glory_mode",
			"user_id":     ctx.User.ID,
			"camp":        camp,
			"max_x":       minInt(2, energy),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 的 [军威神光] 触发，等待选择效果", ctx.User.Name))
	return nil
}

type ValkyrieHeroicSummonHandler struct{ engineplayer.BaseHandler }

func (h *ValkyrieHeroicSummonHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.Timing != model.TimingOnHitCheck {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	if info.ActionType != string(model.ActionAttack) || !info.IsHit {
		return false
	}
	return engineplayer.CanPayCrystalLike(ctx, 1)
}

func (h *ValkyrieHeroicSummonHandler) Execute(ctx *model.Context) error {
	if !engineplayer.CanPayCrystalLike(ctx, 1) {
		return nil
	}
	if !engineplayer.SpendCrystalLike(ctx, 1) {
		return fmt.Errorf("英灵召唤发动失败：水晶不足（红宝石可替代）")
	}
	if ctx.EventCtx != nil && ctx.EventCtx.DamageVal != nil {
		*ctx.EventCtx.DamageVal += 1
	}
	hasMagic := false
	magicIndices := make([]int, 0)
	for i, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			hasMagic = true
			magicIndices = append(magicIndices, i)
		}
	}
	if hasMagic {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type":   "valkyrie_heroic_discard_card",
				"user_id":       ctx.User.ID,
				"user_ctx":      ctx,
				"magic_indices": magicIndices,
			},
		})
	}
	// 仅在自己的行动回合内，英灵召唤才会令女武神进入英灵形态；
	// 应战命中依然可以发动该技能，但不会入形态。
	if ctx.User.IsActive {
		engineplayer.SetForm(ctx.User, model.FormValkyrieHeroic)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [英灵召唤]，伤害+1并进入英灵形态", ctx.User.Name))
		return nil
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [英灵召唤]，伤害+1", ctx.User.Name))
	return nil
}
