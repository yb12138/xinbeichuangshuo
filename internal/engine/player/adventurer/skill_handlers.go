// gameflow: 冒险家技能处理器。

package adventurer

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// --- 冒险家技能处理器 ---

type AdventurerFraudHandler struct{ engineplayer.BaseHandler }

func (h *AdventurerFraudHandler) CanUse(ctx *model.Context) bool {
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		counts[c.Element]++
	}
	for ele, n := range counts {
		// 弃2同系仅要求有同系牌；攻击系别在后续弹窗中单独选择（不含光/暗）
		if ele != "" && n >= 2 {
			return true
		}
		if n >= 3 {
			return true
		}
	}
	return false
}

func (h *AdventurerFraudHandler) Execute(ctx *model.Context) error {
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		counts[c.Element]++
	}
	canPick := false
	for ele, n := range counts {
		if ele == "" {
			continue
		}
		if n >= 2 {
			canPick = true
			break
		}
	}
	if !canPick {
		return fmt.Errorf("欺诈需要至少2张同系手牌")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context:  fraudChoiceContext(ctx),
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [欺诈]，请先选择同系手牌", ctx.User.Name))
	return nil
}

func fraudChoiceContext(ctx *model.Context) map[string]interface{} {
	data := map[string]interface{}{
		"choice_type": "adventurer_fraud_pick",
		"user_id":     ctx.User.ID,
		"user_ctx":    ctx,
		"fraud_target_id": func() string {
			if ctx.Target != nil {
				return ctx.Target.ID
			}
			return ""
		}(),
		"fraud_from_skill": true,
	}
	flow := adventurerFraudFlowRuntime.MustBeginAt(adventurerFraudCardsStep)
	flow.PutSelection(adventurerFraudCardsStep, model.PromptFlowSelection{})
	model.SetPromptFlowContext(data, flow)
	return data
}

type AdventurerLuckyFortuneHandler struct{ engineplayer.BaseHandler }

func (h *AdventurerLuckyFortuneHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || !ctx.AttackDeclarePhase() {
		return false
	}
	if ctx.EventCtx == nil || ctx.EventCtx.Card == nil {
		return false
	}
	card := ctx.EventCtx.Card
	// 强运仅在"欺诈转化出的攻击"开始时自动触发。
	return card.ID == "fraud_virtual_attack" || card.Name == "欺诈"
}

func (h *AdventurerLuckyFortuneHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return nil
	}
	ctx.User.Crystal++
	ctx.Game.Log(fmt.Sprintf("%s 的 [强运] 触发，获得1蓝水晶", ctx.User.Name))
	return nil
}

type AdventurerUndergroundLawHandler struct{ engineplayer.BaseHandler }

func (h *AdventurerUndergroundLawHandler) CanUse(ctx *model.Context) bool {
	return ctx.EventCtx != nil && ctx.EventCtx.ActionType == model.ActionBuy
}

func (h *AdventurerUndergroundLawHandler) Execute(ctx *model.Context) error {
	ctx.Game.ModifyGem(string(ctx.User.Camp), 2)
	ctx.Game.Log(fmt.Sprintf("%s 的 [地下法则] 触发，战绩区+2红宝石", ctx.User.Name))
	return nil
}

type AdventurerStealSkyHandler struct{ engineplayer.BaseHandler }

func (h *AdventurerStealSkyHandler) CanUse(ctx *model.Context) bool {
	return engineplayer.CanPayCrystalLike(ctx, 1)
}

func (h *AdventurerStealSkyHandler) Execute(ctx *model.Context) error {
	enemy := model.BlueCamp
	if ctx.User.Camp == model.BlueCamp {
		enemy = model.RedCamp
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "adventurer_steal_sky_mode",
			"user_id":     ctx.User.ID,
			"enemy_camp":  string(enemy),
			"self_camp":   string(ctx.User.Camp),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [偷天换日]，等待选择效果", ctx.User.Name))
	return nil
}
