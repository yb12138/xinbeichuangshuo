package engine

import (
	"fmt"
	"strconv"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildMagicMissilePrompt() *model.Prompt {
	chain := e.State.MagicBulletChain
	if chain == nil {
		return nil
	}

	playerID := chain.TargetID
	player := e.State.Players[playerID]
	if player == nil {
		return nil
	}

	damage := chain.CurrentDamage
	hasShield := player.HasFieldEffect(model.EffectShield)
	takeLabel := "承受伤害"
	if hasShield {
		takeLabel = "承受（将触发圣盾）"
	}
	effectHints := []string{}
	if hasShield {
		effectHints = append(effectHints, "你身上有【圣盾】：可先应战/防御；若本次选择承受伤害，将自动消耗圣盾抵挡魔弹。")
	}

	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		AttackerID: chain.SourcePlayerID,
		Message:    fmt.Sprintf("你成为了【魔弹】的目标，当前伤害为 %d，请选择应对：", damage),
		Options: []model.PromptOption{
			{ID: "take", Label: takeLabel},
			{ID: "counter", Label: "打出【魔弹】传递"},
			{ID: "defend", Label: "使用【圣光】抵挡"},
		},
		EffectHints: effectHints,
		Min:         1,
		Max:         1,
	}
}

func (e *GameEngine) buildMagicBulletFusionPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return nil
	}
	playerID := interrupt.PlayerID
	player := e.State.Players[playerID]
	if player == nil {
		return nil
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return nil
	}
	cardIdx, _ := data["card_idx"].(int)
	card, _, _, cardOK := getPlayableCardByIndex(player, cardIdx)
	if !cardOK {
		return nil
	}

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  fmt.Sprintf("【魔弹融合】是否将 %s (%s系) 当魔弹使用？", card.Name, elementNameForPrompt(string(card.Element))),
		Options: []model.PromptOption{
			{ID: "yes", Label: "是 - 当魔弹使用"},
			{ID: "no", Label: "否 - 正常使用"},
		},
		Min: 1,
		Max: 1,
	}
}

func (e *GameEngine) buildMagicBulletDirectionPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return nil
	}
	playerID := interrupt.PlayerID

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【魔弹掌控】选择魔弹传递方向：",
		Options: []model.PromptOption{
			{ID: "normal", Label: "默认方向 (右手边，前一位对手)"},
			{ID: "reverse", Label: "逆向传递 (左手边，后一位对手)"},
		},
		Min: 1,
		Max: 1,
	}
}

func (e *GameEngine) buildMagicBlastPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return nil
	}

	playerID := interrupt.PlayerID
	player := e.State.Players[playerID]
	if player == nil {
		return nil
	}
	data, _ := interrupt.Context.(map[string]interface{})
	stage := magicBlastStageFromContext(data)

	if stage == magicBlastStageCasterForcedDiscard {
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  "【魔爆冲击】请选择弃1张牌：",
			Options:  magicBlastCasterForcedDiscardOptions(player),
			Min:      1,
			Max:      1,
		}
	}

	return &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: playerID,
		Message:  "【魔爆冲击】请选择弃一张法术牌，否则受到2点伤害：",
		Options:  magicBlastTargetDiscardOptions(player),
		Min:      1,
		Max:      1,
	}
}

func magicBlastStageFromContext(data map[string]interface{}) string {
	stage, _ := data["stage"].(string)
	if stage == "" {
		return magicBlastStageTargetDiscard
	}
	return stage
}

func magicBlastCasterForcedDiscardOptions(player *model.Player) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(player.Hand))
	for i, card := range player.Hand {
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i),
			Label: fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
		})
	}
	return options
}

func magicBlastTargetDiscardOptions(player *model.Player) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(player.Hand)+1)
	for i, card := range player.Hand {
		if card.Type != model.CardTypeMagic {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i),
			Label: fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
		})
	}
	options = append(options, model.PromptOption{
		ID:    "refuse",
		Label: "不弃牌 (受到2点伤害)",
	})
	return options
}
