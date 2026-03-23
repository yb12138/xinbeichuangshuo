package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

type ButterflyLifeFireHandler struct{ BaseHandler }

type ButterflyDanceHandler struct{ BaseHandler }

type ButterflyPoisonPowderHandler struct{ BaseHandler }

type ButterflyPilgrimageHandler struct{ BaseHandler }

type ButterflyMirrorHandler struct{ BaseHandler }

type ButterflyWitherHandler struct{ BaseHandler }

type ButterflyChrysalisHandler struct{ BaseHandler }

type ButterflyReverseHandler struct{ BaseHandler }

func (h *ButterflyLifeFireHandler) CanUse(ctx *model.Context) bool { return false }

func (h *ButterflyLifeFireHandler) Execute(ctx *model.Context) error { return nil }

func (h *ButterflyDanceHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && ctx.Game != nil
}

func (h *ButterflyDanceHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("舞动上下文无效")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_dance_mode",
			"user_id":     ctx.User.ID,
			"can_discard": len(ctx.User.Hand) > 0,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [舞动]：请选择先摸1张牌或弃1张牌", ctx.User.Name))
	return nil
}

// 毒粉、朝圣、镜花水月、凋零由引擎伤害时点逻辑统一调度，不走通用 Trigger 分发。
func (h *ButterflyPoisonPowderHandler) CanUse(ctx *model.Context) bool   { return false }
func (h *ButterflyPoisonPowderHandler) Execute(ctx *model.Context) error { return nil }
func (h *ButterflyPilgrimageHandler) CanUse(ctx *model.Context) bool     { return false }
func (h *ButterflyPilgrimageHandler) Execute(ctx *model.Context) error   { return nil }
func (h *ButterflyMirrorHandler) CanUse(ctx *model.Context) bool         { return false }
func (h *ButterflyMirrorHandler) Execute(ctx *model.Context) error       { return nil }
func (h *ButterflyWitherHandler) CanUse(ctx *model.Context) bool         { return false }
func (h *ButterflyWitherHandler) Execute(ctx *model.Context) error       { return nil }

func (h *ButterflyChrysalisHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && ctx.Game != nil
}

func (h *ButterflyChrysalisHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("蛹化上下文无效")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_chrysalis_resolve",
			"user_id":     ctx.User.ID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [蛹化]：等待结算 +1蛹 并获得4张茧", ctx.User.Name))
	return nil
}

func (h *ButterflyReverseHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && ctx.Game != nil
}

func (h *ButterflyReverseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("倒逆之蝶上下文无效")
	}
	discardNeed := 2
	if len(ctx.User.Hand) < discardNeed {
		discardNeed = len(ctx.User.Hand)
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "bt_reverse_discard",
			"user_id":       ctx.User.ID,
			"discard_count": discardNeed,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [倒逆之蝶]：请先弃置%d张牌", ctx.User.Name, discardNeed))
	return nil
}
