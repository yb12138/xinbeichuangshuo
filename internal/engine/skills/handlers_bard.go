package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

func bardInspirationCount(user *model.Player) int {
	if user == nil {
		return 0
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	v := user.Tokens["bd_inspiration"]
	if v < 0 {
		v = 0
	}
	if v > 3 {
		v = 3
	}
	user.Tokens["bd_inspiration"] = v
	return v
}

type BardDescentConcertoHandler struct{ BaseHandler }

type BardDissonanceChordHandler struct{ BaseHandler }

type BardForbiddenVerseHandler struct{ BaseHandler }

type BardRousingRhapsodyHandler struct{ BaseHandler }

type BardVictorySymphonyHandler struct{ BaseHandler }

type BardHopeFugueHandler struct{ BaseHandler }

func (h *BardDescentConcertoHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BardDescentConcertoHandler) Execute(ctx *model.Context) error { return nil }

func (h *BardDissonanceChordHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && bardInspirationCount(ctx.User) > 1
}

func (h *BardDissonanceChordHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("不谐和弦上下文无效")
	}
	inspiration := bardInspirationCount(ctx.User)
	if inspiration <= 1 {
		return fmt.Errorf("灵感不足，至少需要2点")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_dissonance_x",
			"user_id":     ctx.User.ID,
			"max_x":       inspiration,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [不谐和弦]，请选择X值（2~%d）", ctx.User.Name, inspiration))
	return nil
}

func (h *BardForbiddenVerseHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BardForbiddenVerseHandler) Execute(ctx *model.Context) error { return nil }

func (h *BardRousingRhapsodyHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BardRousingRhapsodyHandler) Execute(ctx *model.Context) error { return nil }

func (h *BardVictorySymphonyHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BardVictorySymphonyHandler) Execute(ctx *model.Context) error { return nil }

func (h *BardHopeFugueHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *BardHopeFugueHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("希望赋格曲上下文无效")
	}
	// 资源已由 UseSkill 统一结算，这里只负责进入技能交互流程。
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_hope_draw_confirm",
			"user_id":     ctx.User.ID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [希望赋格曲]，请先选择是否摸1张牌", ctx.User.Name))
	return nil
}
