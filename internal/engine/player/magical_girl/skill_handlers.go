// gameflow: 魔法少女 skill handler。

package magical_girl

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"

	"starcup-engine/internal/model"
)

type MagicBulletControlHandler struct{ engineplayer.BaseHandler }

func (h *MagicBulletControlHandler) Execute(ctx *model.Context) error {
	// 魔弹掌控由 magic.go/game.go 中的魔弹中断链路直接处理；
	// 这里保留 handler 仅用于维持技能注册表完整。
	return nil
}

type MagicBulletFusionHandler struct{ engineplayer.BaseHandler }

func (h *MagicBulletFusionHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return hasFusionCard(ctx.User)
}

func (h *MagicBulletFusionHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔弹融合上下文无效")
	}
	discardedCards, _ := ctx.Selections["discardedCards"].([]model.Card)
	if len(discardedCards) != 1 || !isFusionElement(discardedCards[0]) {
		return fmt.Errorf("魔弹融合需要选择1张地系或火系牌")
	}
	fusionCard := discardedCards[0]
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptMagicBulletDirection,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"source_id":   ctx.User.ID,
			"is_fusion":   true,
			"fusion_card": fusionCard,
		},
	})
	ctx.Game.Log(fmt.Sprintf("[Skill] %s 发动【魔弹融合】，将 %s 当魔弹使用！", ctx.User.Name, fusionCard.Name))
	return nil
}

type MagicBulletFusionChainHandler struct{ engineplayer.BaseHandler }

func (h *MagicBulletFusionChainHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	chain, _ := ctx.Selections["magic_missile_chain"].(*model.MagicBulletChain)
	if chain == nil || chain.TargetID != ctx.User.ID {
		return false
	}
	return !hasParticipatedInMagicBulletChain(ctx.User.ID, chain) && hasFusionCard(ctx.User)
}

func (h *MagicBulletFusionChainHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔弹融合上下文无效")
	}
	chain, _ := ctx.Selections["magic_missile_chain"].(*model.MagicBulletChain)
	if chain == nil || chain.TargetID != ctx.User.ID {
		return fmt.Errorf("当前没有指向你的魔弹")
	}
	if hasParticipatedInMagicBulletChain(ctx.User.ID, chain) {
		return fmt.Errorf("你在本轮传递中已参与过，无法再次传递")
	}
	discardIndices, ok := ctx.Selections["discard_indices"].([]int)
	if !ok || len(discardIndices) != 1 {
		return fmt.Errorf("魔弹融合需要选择1张地系或火系牌")
	}
	idx := discardIndices[0]
	if idx < 0 || idx >= len(ctx.User.Hand) {
		return fmt.Errorf("牌索引越界: %d", idx)
	}
	card := ctx.User.Hand[idx]
	if !isFusionElement(card) {
		return fmt.Errorf("魔弹融合需要选择地系或火系牌")
	}

	ctx.User.Hand = append(ctx.User.Hand[:idx], ctx.User.Hand[idx+1:]...)
	ctx.Game.NotifyCardRevealed(ctx.User.ID, []model.Card{card}, "counter")
	ctx.Selections["discardedCards"] = []model.Card{card}
	ctx.Selections["magic_missile_fusion_chain_resolved"] = true
	ctx.Game.Log(fmt.Sprintf("[Skill] %s 发动【魔弹融合】，将 %s 视为魔弹继续传递", ctx.User.Name, card.Name))

	chain.CurrentDamage += 1
	chain.SourcePlayerID = ctx.User.ID
	chain.InvolvedIDs = append(chain.InvolvedIDs, ctx.User.ID)

	return nil
}

type MagicBlastHandler struct{ engineplayer.BaseHandler }

func (h *MagicBlastHandler) CanUse(ctx *model.Context) bool {
	// 需要有法术牌可弃才能发动
	for _, card := range ctx.User.Hand {
		if card.Type == model.CardTypeMagic {
			return true
		}
	}
	return false
}

func hasFusionCard(player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, card := range player.Hand {
		if isFusionElement(card) {
			return true
		}
	}
	return false
}

func hasParticipatedInMagicBulletChain(playerID string, chain *model.MagicBulletChain) bool {
	if chain == nil {
		return false
	}
	for _, pid := range chain.InvolvedIDs {
		if pid == playerID {
			return true
		}
	}
	return false
}

func (h *MagicBlastHandler) Execute(ctx *model.Context) error {
	// 弃牌代价已在 UseSkill 中处理；这里从"我方战绩区 +1 宝石"开始。
	ctx.Game.ModifyGem(string(ctx.User.Camp), 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔爆冲击]，我方战绩区+1宝石", ctx.User.Name))

	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) != 2 {
		return fmt.Errorf("魔爆冲击需要且只能指定2名敌方目标")
	}

	targetIDs := make([]string, len(targets))
	for i, t := range targets {
		targetIDs[i] = t.ID
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptMagicBlast,
		PlayerID: targetIDs[0], // 第一个目标先响应
		Context: map[string]interface{}{
			"choice_type":    "magic_blast",
			"stage":          "target_discard",
			"caster_id":      ctx.User.ID,
			"targets":        targetIDs,
			"current_target": 0,
		},
	})
	ctx.Game.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", targets[0].Name))

	return nil
}

type DestructionStormHandler struct{ engineplayer.BaseHandler }

func (h *DestructionStormHandler) Execute(ctx *model.Context) error {
	// 毁灭风暴：[宝石] 对任2名目标对手各造成2点法术伤害
	targets := ctx.Targets
	if len(targets) == 0 && ctx.Target != nil {
		targets = []*model.Player{ctx.Target}
	}
	if len(targets) != 2 {
		return fmt.Errorf("毁灭风暴需要且只能指定2名敌方目标")
	}

	for _, t := range targets {
		ctx.Game.InflictDamage(ctx.User.ID, t.ID, 2, model.MagicAttack)
	}

	ctx.Game.Log(fmt.Sprintf("%s 发动 [毁灭风暴]，对 %d 名目标造成伤害", ctx.User.Name, len(targets)))
	return nil
}
