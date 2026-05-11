// gameflow: 吟游诗人 handler。

package bard

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"

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

func bardHasEternalMovement(game model.IGameEngine, bard *model.Player) bool {
	if game == nil || bard == nil {
		return false
	}
	_, fc := game.FindFieldEffectBySource(model.EffectBardEternalMovement, bard.ID)
	return fc != nil
}

func bardEnemyIDs(game model.IGameEngine, bard *model.Player) []string {
	if game == nil || bard == nil {
		return nil
	}
	var ids []string
	for _, p := range game.GetAllPlayers() {
		if p == nil || p.Camp == bard.Camp {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

type BardDescentConcertoHandler struct{ engineplayer.BaseHandler }

type BardDissonanceChordHandler struct{ engineplayer.BaseHandler }

type BardForbiddenVerseHandler struct{ engineplayer.BaseHandler }

type BardRousingRhapsodyHandler struct{ engineplayer.BaseHandler }

type BardVictorySymphonyHandler struct{ engineplayer.BaseHandler }

type BardHopeFugueHandler struct{ engineplayer.BaseHandler }

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

func (h *BardRousingRhapsodyHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return false
	}
	stage, _ := ctx.Selections["bd_song_stage"].(string)
	if stage != "turn_start" || !ctx.User.IsActive {
		return false
	}
	if !bardHasEternalMovement(ctx.Game, ctx.User) {
		return false
	}
	return len(bardEnemyIDs(ctx.Game, ctx.User)) >= 2 || len(ctx.User.Hand) >= 2
}

func (h *BardRousingRhapsodyHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return fmt.Errorf("激昂狂想曲上下文无效")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_rousing_mode",
			"user_id":     ctx.User.ID,
			"target_ids":  bardEnemyIDs(ctx.Game, ctx.User),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [激昂狂想曲]，请选择效果", ctx.User.Name))
	return nil
}

func (h *BardVictorySymphonyHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return false
	}
	stage, _ := ctx.Selections["bd_song_stage"].(string)
	if stage != "turn_end" || !ctx.User.IsActive {
		return false
	}
	if !bardHasEternalMovement(ctx.Game, ctx.User) {
		return false
	}
	return true
}

func (h *BardVictorySymphonyHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return fmt.Errorf("胜利交响诗上下文无效")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_victory_mode",
			"user_id":     ctx.User.ID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [胜利交响诗]，请选择效果", ctx.User.Name))
	return nil
}

func (h *BardHopeFugueHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return false
	}
	return engineplayer.CanPayCrystalLike(ctx, 1)
}

func (h *BardHopeFugueHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return fmt.Errorf("希望赋格曲上下文无效")
	}
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
