// gameflow: 神箭手技能处理器。

package archer

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type BaseHandler = engineplayer.BaseHandler

// --- Archer Skill Handlers ---

type PiercingShotHandler struct{ BaseHandler }

func (h *PiercingShotHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || !ctx.AttackMissPhase() || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.ActionType != string(model.ActionAttack) {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	for _, card := range ctx.User.Hand {
		if card.Type == model.CardTypeMagic {
			return true
		}
	}
	return false
}

func (h *PiercingShotHandler) Execute(ctx *model.Context) error {
	discardRaw, hasDiscard := ctx.Selections["discard_indices"]
	if !hasDiscard {
		return fmt.Errorf("贯穿射击缺少弃牌选择")
	}
	indices, ok := discardRaw.([]int)
	if !ok || len(indices) != 1 {
		return fmt.Errorf("贯穿射击需要且仅需弃置1张法术牌")
	}
	idx := indices[0]
	if idx < 0 || idx >= len(ctx.User.Hand) {
		return fmt.Errorf("贯穿射击弃牌索引无效: %d", idx)
	}
	card := ctx.User.Hand[idx]
	if card.Type != model.CardTypeMagic {
		return fmt.Errorf("贯穿射击必须弃置法术牌")
	}
	ctx.Game.NotifyCardRevealed(ctx.User.ID, []model.Card{card}, "discard")
	ctx.User.Hand = append(ctx.User.Hand[:idx], ctx.User.Hand[idx+1:]...)
	ctx.Selections["discardedCards"] = []model.Card{card}

	if ctx.Target != nil {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 2, model.MagicAttack)
		ctx.Game.Log(fmt.Sprintf("%s 发动 [贯穿射击]，对 %s 造成2点法术伤害",
			ctx.User.Name, ctx.Target.Name))
	}
	return nil
}

type LightningArrowHandler struct{ BaseHandler }

func (h *LightningArrowHandler) CanUse(ctx *model.Context) bool {
	if ctx.EventCtx != nil && ctx.EventCtx.Card != nil {
		return ctx.EventCtx.Card.Element == model.ElementThunder
	}
	return false
}

func (h *LightningArrowHandler) Execute(ctx *model.Context) error {
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		ctx.EventCtx.AttackInfo.CanBeResponded = false
		ctx.Game.Log(fmt.Sprintf("%s 发动 [闪电箭]，雷系攻击不可被应战", ctx.User.Name))
	}
	return nil
}

type SnipeHandler struct{ BaseHandler }

func (h *SnipeHandler) Execute(ctx *model.Context) error {
	if ctx.Target != nil {
		currentHand := len(ctx.Target.Hand)
		if currentHand < 5 {
			needCards := 5 - currentHand
			ctx.Game.DrawCards(ctx.Target.ID, needCards)
			ctx.Game.Log(fmt.Sprintf("%s 的 [狙击] 发动，%s 手牌补到5张", ctx.User.Name, ctx.Target.Name))
		} else {
			ctx.Game.Log(fmt.Sprintf("%s 的 [狙击] 发动，但 %s 手牌已有%d张，无事发生", ctx.User.Name, ctx.Target.Name, currentHand))
		}
		model.AppendAttackAction(ctx.User, "狙击")
		ctx.Game.Log(fmt.Sprintf("%s 发动 [狙击]，额外获得1次攻击行动", ctx.User.Name))
	}
	return nil
}

type PreciseShotHandler struct{ BaseHandler }

func (h *PreciseShotHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.User.Character == nil || ctx.EventCtx == nil || ctx.EventCtx.AttackInfo == nil || ctx.EventCtx.Card == nil {
		return false
	}
	if !ctx.AttackDeclarePhase() {
		return false
	}
	info := ctx.EventCtx.AttackInfo
	if info.ActionType != string(model.ActionAttack) || info.CounterInitiator != "" {
		return false
	}
	return ctx.EventCtx.Card.MatchExclusive(ctx.User.Character.ID, "精准射击")
}

func (h *PreciseShotHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.EventCtx == nil {
		return nil
	}
	if !ctx.AttackDeclarePhase() {
		return nil
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [精准射击]，攻击强制命中但伤害-1", ctx.User.Name))
	if ctx.EventCtx.AttackInfo != nil {
		ctx.EventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptForceHit)
	}
	if ctx.User != nil && ctx.Game != nil {
		ctx.Game.ApplyNextAttackDamageRule(ctx.User.ID, preciseShotDamageModifierID, "precise_shot", -1, model.RuleLifeThisEffectChain)
	}
	return nil
}

type FlashTrapHandler struct{ BaseHandler }

func (h *FlashTrapHandler) Execute(ctx *model.Context) error {
	if ctx.Target != nil {
		ctx.Game.InflictDamage(ctx.User.ID, ctx.Target.ID, 2, "法术")
	}
	ctx.Game.Log(fmt.Sprintf("%s 使用技能后回合结束", ctx.User.Name))
	return nil
}
