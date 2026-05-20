// gameflow: 贤者技能处理器。

package sage

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type BaseHandler = engineplayer.BaseHandler

// --- Sage Helpers ---

func sageDistinctElements(user *model.Player) map[model.Element]int {
	out := map[model.Element]int{}
	if user == nil {
		return out
	}
	for _, c := range user.Hand {
		if c.Element == "" {
			continue
		}
		out[c.Element]++
	}
	return out
}

// --- Sage Skill Handlers ---

type SageWisdomCodexHandler struct{ BaseHandler }

func (h *SageWisdomCodexHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SageWisdomCodexHandler) Execute(ctx *model.Context) error { return nil }

type SageMagicReboundHandler struct{ BaseHandler }

func (h *SageMagicReboundHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SageMagicReboundHandler) Execute(ctx *model.Context) error { return nil }

type SageArcaneCodexHandler struct{ BaseHandler }

func (h *SageArcaneCodexHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && len(sageDistinctElements(ctx.User)) >= 2
}

func (h *SageArcaneCodexHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("魔道法典上下文无效")
	}
	distinct := sageDistinctElements(ctx.User)
	maxX := len(distinct)
	if maxX < 2 {
		return fmt.Errorf("魔道法典需要至少2种不同元素手牌")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":              "sage_arcane_cards",
			"user_id":                  ctx.User.ID,
			model.PromptFlowContextKey: sageArcaneFlowRuntime.Begin(),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [魔道法典]，请选择异系牌", ctx.User.Name))
	return nil
}

type SageHolyCodexHandler struct{ BaseHandler }

func (h *SageHolyCodexHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return ctx.User.Gem > 0 && len(sageDistinctElements(ctx.User)) >= 3
}

func (h *SageHolyCodexHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("圣洁法典上下文无效")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":              "sage_holy_cards",
			"user_id":                  ctx.User.ID,
			model.PromptFlowContextKey: sageHolyFlowRuntime.Begin(),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [圣洁法典]，请选择异系牌", ctx.User.Name))
	return nil
}
