// gameflow: 鬼术师技能处理器。

package onmyoji

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"

	"starcup-engine/internal/model"
)

// --- Onmyoji Handlers ---

type OnmyojiShikigamiDescendHandler struct{ engineplayer.BaseHandler }

func (h *OnmyojiShikigamiDescendHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if engineplayer.HasForm(ctx.User, model.FormOnmyojiShikigami) {
		return false
	}
	if len(ctx.User.Hand) < 2 {
		return false
	}
	factionCount := map[string]int{}
	for _, c := range ctx.User.Hand {
		if c.Faction == "" {
			continue
		}
		factionCount[c.Faction]++
		if factionCount[c.Faction] >= 2 {
			return true
		}
	}
	return false
}

func (h *OnmyojiShikigamiDescendHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	if engineplayer.HasForm(ctx.User, model.FormOnmyojiShikigami) {
		return fmt.Errorf("已处于式神形态")
	}
	engineplayer.SetForm(ctx.User, model.FormOnmyojiShikigami)
	engineplayer.AddToken(ctx.User, "onmyoji_ghost_fire", 1, 3)
	model.AppendAttackAction(ctx.User, "式神降临")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [式神降临]，弃2张同命格手牌后进入式神形态并+1鬼火，获得额外攻击行动", ctx.User.Name))
	return nil
}

type OnmyojiYinYangShiftHandler struct{ engineplayer.BaseHandler }

func (h *OnmyojiYinYangShiftHandler) CanUse(ctx *model.Context) bool { return false }

func (h *OnmyojiYinYangShiftHandler) Execute(ctx *model.Context) error { return nil }

type OnmyojiShikigamiShiftHandler struct{ engineplayer.BaseHandler }

func (h *OnmyojiShikigamiShiftHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	// 仅在阴阳转换结算期间触发，防止标准响应技能系统误触发
	return ctx.Flags["yinyang_counter_active"]
}

func (h *OnmyojiShikigamiShiftHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	ctx.Game.DrawCards(ctx.User.ID, 1)
	engineplayer.AddToken(ctx.User, "onmyoji_ghost_fire", 1, 3)
	ctx.Game.Log(fmt.Sprintf("%s 的 [式神转换] 触发：摸1并鬼火+1", ctx.User.Name))
	return nil
}

type OnmyojiDarkRitualHandler struct{ engineplayer.BaseHandler }

func (h *OnmyojiDarkRitualHandler) Execute(ctx *model.Context) error { return nil }

type OnmyojiBindingHandler struct{ engineplayer.BaseHandler }

func (h *OnmyojiBindingHandler) CanUse(ctx *model.Context) bool { return false }

func (h *OnmyojiBindingHandler) Execute(ctx *model.Context) error { return nil }

type OnmyojiLifeBarrierHandler struct{ engineplayer.BaseHandler }

func (h *OnmyojiLifeBarrierHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	if !engineplayer.CanPayCrystalLike(ctx, 1) {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			return true
		}
	}
	return false
}

func (h *OnmyojiLifeBarrierHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("生命结界上下文无效")
	}
	gf := engineplayer.AddToken(ctx.User, "onmyoji_ghost_fire", 1, 3)

	// 分支①可选队友（不含自己）
	var supportTargetIDs []string
	// 分支②可选队友（需有手牌可弃）
	var releaseTargetIDs []string
	lockedTargetID := ""
	if ctx.Target != nil {
		if ctx.Target.Camp != ctx.User.Camp || ctx.Target.ID == ctx.User.ID {
			return fmt.Errorf("生命结界目标必须是其他队友")
		}
		lockedTargetID = ctx.Target.ID
		supportTargetIDs = append(supportTargetIDs, ctx.Target.ID)
		if len(ctx.Target.Hand) > 0 {
			releaseTargetIDs = append(releaseTargetIDs, ctx.Target.ID)
		}
	} else {
		for _, p := range ctx.Game.GetAllPlayers() {
			if p == nil || p.Camp != ctx.User.Camp || p.ID == ctx.User.ID {
				continue
			}
			supportTargetIDs = append(supportTargetIDs, p.ID)
			if len(p.Hand) > 0 {
				releaseTargetIDs = append(releaseTargetIDs, p.ID)
			}
		}
	}
	if len(supportTargetIDs) == 0 {
		return fmt.Errorf("生命结界没有可选队友目标")
	}

	// 分支②：式神形态 + 手牌中存在"2张同命格"组合 + 有队友可弃牌
	var releaseCombos []string
	if engineplayer.HasForm(ctx.User, model.FormOnmyojiShikigami) && len(releaseTargetIDs) > 0 {
		for i := 0; i < len(ctx.User.Hand); i++ {
			if ctx.User.Hand[i].Faction == "" {
				continue
			}
			for j := i + 1; j < len(ctx.User.Hand); j++ {
				if ctx.User.Hand[i].Faction == ctx.User.Hand[j].Faction {
					releaseCombos = append(releaseCombos, fmt.Sprintf("%d,%d", i, j))
				}
			}
		}
	}

	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":         "onmyoji_life_barrier_mode",
			"user_id":             ctx.User.ID,
			"ghost_fire":          gf,
			"locked_target_id":    lockedTargetID,
			"support_target_ids":  supportTargetIDs,
			"release_target_ids":  releaseTargetIDs,
			"release_card_combos": releaseCombos,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [生命结界]，鬼火+1（当前%d），请选择分支效果", ctx.User.Name, gf))
	return nil
}
