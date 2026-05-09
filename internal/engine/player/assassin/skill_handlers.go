// gameflow: 暗杀者技能处理器。

package assassin

import (
	"fmt"
	"sort"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type BaseHandler = engineplayer.BaseHandler

// --- Assassin Skill Handlers ---

type BacklashHandler struct{ BaseHandler }

func (h *BacklashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.Timing != model.TimingOnDamageTaken || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.DamageVal == nil || *ctx.EventCtx.DamageVal <= 0 {
		return false
	}
	if ctx.Flags["IsMagicDamage"] {
		return false
	}
	if ctx.EventCtx.SourceID == "" || ctx.User == nil || ctx.EventCtx.SourceID == ctx.User.ID {
		return false
	}
	return true
}

func (h *BacklashHandler) Execute(ctx *model.Context) error {
	attackerID := ctx.EventCtx.SourceID
	attackerName := attackerID
	for _, p := range ctx.Game.GetAllPlayers() {
		if p.ID == attackerID {
			attackerName = model.GetPlayerDisplayName(p)
			break
		}
	}
	ctx.Game.NotifyActionStep(fmt.Sprintf("%s发动被动技反噬，%s强制摸1张牌", model.GetPlayerDisplayName(ctx.User), attackerName))
	ctx.Game.DrawCards(attackerID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [反噬] 发动，%s 强制摸1张牌", ctx.User.Name, attackerID))
	return nil
}

type WaterShadowHandler struct{ BaseHandler }

func (h *WaterShadowHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if !ctx.BeforeDrawPhase() {
		return false
	}
	if ctx.EventCtx.TargetID != "" && ctx.EventCtx.TargetID != ctx.User.ID {
		return false
	}
	if ctx.EventCtx.DrawCount == nil || *ctx.EventCtx.DrawCount <= 0 {
		return false
	}
	if ctx.EventCtx.ActionType == model.ActionBuy ||
		ctx.EventCtx.ActionType == model.ActionSynthesize ||
		ctx.EventCtx.ActionType == model.ActionExtract {
		return false
	}
	return ctx.User.HasElement(model.ElementWater)
}

func (h *WaterShadowHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return fmt.Errorf("水影上下文无效")
	}
	if !ctx.BeforeDrawPhase() {
		return fmt.Errorf("水影只能在摸牌前发动")
	}

	selection, exists := ctx.Selections["discard_indices"]
	if !exists {
		return fmt.Errorf("没有弃牌选择")
	}

	discardIndices, ok := selection.([]int)
	if !ok {
		return fmt.Errorf("弃牌选择格式错误")
	}

	if len(discardIndices) == 0 {
		return fmt.Errorf("至少需要弃1张牌")
	}

	player := ctx.User
	usedIndices := make(map[int]bool)
	waterCards := 0
	magicCards := 0

	for _, idx := range discardIndices {
		if idx < 0 || idx >= len(player.Hand) {
			return fmt.Errorf("牌索引越界: %d", idx)
		}
		if usedIndices[idx] {
			return fmt.Errorf("不能重复选择同一张牌: %d", idx)
		}
		usedIndices[idx] = true

		if player.Hand[idx].Element == model.ElementWater {
			waterCards++
		} else if player.Hand[idx].Type == model.CardTypeMagic {
			magicCards++
		} else {
			return fmt.Errorf("选择的牌既不是水系牌也不是法术牌: %s", player.Hand[idx].Name)
		}
	}

	isStealthed := engineplayer.HasForm(player, model.FormAssassinStealth)

	if waterCards == 0 {
		return fmt.Errorf("至少需要弃1张水系牌")
	}
	if !isStealthed && magicCards > 0 {
		return fmt.Errorf("不在潜行状态下不能弃法术牌")
	}

	if isStealthed && magicCards > 1 {
		return fmt.Errorf("潜行状态下最多只能弃1张法术牌")
	}

	sort.Sort(sort.Reverse(sort.IntSlice(discardIndices)))

	discardedCards := make([]model.Card, 0, len(discardIndices))
	for _, idx := range discardIndices {
		discardedCards = append(discardedCards, player.Hand[idx])
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}

	ctx.Game.NotifyCardRevealed(player.ID, discardedCards, "discard")

	ctx.Selections["discardedCards"] = discardedCards

	ctx.Game.Log(fmt.Sprintf("%s 发动 [水影]，展示并弃置了 %d 张水系牌", player.Name, waterCards+magicCards))
	if magicCards > 0 {
		ctx.Game.Log(fmt.Sprintf("%s 处于[潜行]，额外展示并弃置了 %d 张法术牌", player.Name, magicCards))
	}

	return nil
}

type StealthHandler struct{ BaseHandler }

func (h *StealthHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if ctx.Timing != model.TimingOnTurnStart && ctx.Timing != model.TimingStartup {
		return false
	}
	if ctx.User.Gem < 1 {
		return false
	}
	return !engineplayer.HasForm(ctx.User, model.FormAssassinStealth)
}

func (h *StealthHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("潜行上下文无效")
	}
	if ctx.User.Gem < 1 {
		return fmt.Errorf("宝石不足，无法发动潜行")
	}
	if engineplayer.HasForm(ctx.User, model.FormAssassinStealth) {
		return fmt.Errorf("已处于潜行状态")
	}
	ctx.User.Gem -= 1

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "assassin_stealth_draw",
			"user_id":       ctx.User.ID,
			"waiting_phase": model.TurnStageActionStart,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [潜行]，消耗1宝石，等待选择是否摸1张牌后进入潜行状态", ctx.User.Name))
	return nil
}
