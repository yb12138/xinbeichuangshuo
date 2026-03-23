package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

const (
	soulSorcererBlueCap   = 6
	soulSorcererYellowCap = 6
)

type SoulSorcererSoulDevourHandler struct{ BaseHandler }

type SoulSorcererSoulRecallHandler struct{ BaseHandler }

type SoulSorcererSoulConvertHandler struct{ BaseHandler }

type SoulSorcererSoulMirrorHandler struct{ BaseHandler }

type SoulSorcererSoulBlastHandler struct{ BaseHandler }

type SoulSorcererSoulGrantHandler struct{ BaseHandler }

type SoulSorcererSoulLinkHandler struct{ BaseHandler }

type SoulSorcererSoulAmpHandler struct{ BaseHandler }

func soulBlue(user *model.Player) int {
	return addToken(user, "ss_blue_soul", 0, 0, soulSorcererBlueCap)
}

func soulYellow(user *model.Player) int {
	return addToken(user, "ss_yellow_soul", 0, 0, soulSorcererYellowCap)
}

func addSoulBlue(user *model.Player, delta int) int {
	return addToken(user, "ss_blue_soul", delta, 0, soulSorcererBlueCap)
}

func addSoulYellow(user *model.Player, delta int) int {
	return addToken(user, "ss_yellow_soul", delta, 0, soulSorcererYellowCap)
}

func soulSorcererAllyIDs(game model.IGameEngine, user *model.Player, includeSelf bool) []string {
	if game == nil || user == nil {
		return nil
	}
	var ids []string
	for _, p := range game.GetAllPlayers() {
		if p == nil || p.Camp != user.Camp {
			continue
		}
		if !includeSelf && p.ID == user.ID {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

func (h *SoulSorcererSoulDevourHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil {
		return false
	}
	return ctx.Trigger == model.TriggerBeforeMoraleLoss && *ctx.TriggerCtx.DamageVal > 0
}

func (h *SoulSorcererSoulDevourHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DamageVal == nil {
		return fmt.Errorf("灵魂吞噬上下文无效")
	}
	loss := *ctx.TriggerCtx.DamageVal
	if loss <= 0 {
		return nil
	}
	before := soulYellow(ctx.User)
	after := addSoulYellow(ctx.User, loss)
	ctx.Game.Log(fmt.Sprintf("%s 的 [灵魂吞噬] 触发：黄色灵魂 +%d（%d→%d）", ctx.User.Name, loss, before, after))
	return nil
}

func (h *SoulSorcererSoulRecallHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			return true
		}
	}
	return false
}

func (h *SoulSorcererSoulRecallHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("灵魂召还上下文无效")
	}
	magicIndices := make([]int, 0)
	for i, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			magicIndices = append(magicIndices, i)
		}
	}
	if len(magicIndices) == 0 {
		return fmt.Errorf("灵魂召还需要弃置法术牌")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "ss_recall_pick",
			"user_id":       ctx.User.ID,
			"magic_indices": magicIndices,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [灵魂召还]：请选择要弃置的法术牌", ctx.User.Name))
	return nil
}

func (h *SoulSorcererSoulConvertHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	y := soulYellow(ctx.User)
	b := soulBlue(ctx.User)
	canY2B := y > 0 && b < soulSorcererBlueCap
	canB2Y := b > 0 && y < soulSorcererYellowCap
	return canY2B || canB2Y
}

func (h *SoulSorcererSoulConvertHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("灵魂转换上下文无效")
	}
	y := soulYellow(ctx.User)
	b := soulBlue(ctx.User)
	modeOrder := make([]string, 0, 2)
	if y > 0 && b < soulSorcererBlueCap {
		modeOrder = append(modeOrder, "y2b")
	}
	if b > 0 && y < soulSorcererYellowCap {
		modeOrder = append(modeOrder, "b2y")
	}
	if len(modeOrder) == 0 {
		return fmt.Errorf("当前无可执行的灵魂转换")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "ss_convert_color",
			"user_id":     ctx.User.ID,
			"mode_order":  modeOrder,
			"user_ctx":    ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [灵魂转换]：请选择转换方向", ctx.User.Name))
	return nil
}

func (h *SoulSorcererSoulMirrorHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && soulYellow(ctx.User) >= 2 && len(ctx.User.Hand) >= 2
}

func (h *SoulSorcererSoulMirrorHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("灵魂镜像上下文无效")
	}
	target := ctx.Target
	if target == nil && len(ctx.Targets) > 0 {
		target = ctx.Targets[0]
	}
	if target == nil {
		return fmt.Errorf("灵魂镜像需要目标角色")
	}
	if soulYellow(ctx.User) < 2 {
		return fmt.Errorf("黄色灵魂不足2点")
	}
	addSoulYellow(ctx.User, -2)
	drawN := 2
	room := target.MaxHand - len(target.Hand)
	if room < drawN {
		drawN = room
	}
	if drawN < 0 {
		drawN = 0
	}
	if drawN > 0 {
		ctx.Game.DrawCards(target.ID, drawN)
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [灵魂镜像]：移除2点黄色灵魂，%s 摸%d张牌（上限保护）", ctx.User.Name, target.Name, drawN))
	return nil
}

func (h *SoulSorcererSoulBlastHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && soulYellow(ctx.User) >= 3
}

func (h *SoulSorcererSoulBlastHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("灵魂震爆上下文无效")
	}
	target := ctx.Target
	if target == nil && len(ctx.Targets) > 0 {
		target = ctx.Targets[0]
	}
	if target == nil {
		return fmt.Errorf("灵魂震爆需要目标角色")
	}
	if soulYellow(ctx.User) < 3 {
		return fmt.Errorf("黄色灵魂不足3点")
	}
	addSoulYellow(ctx.User, -3)
	damage := 3
	maxHand := target.MaxHand
	if gameWithDynamicMaxHand, ok := ctx.Game.(interface {
		GetMaxHand(*model.Player) int
	}); ok {
		maxHand = gameWithDynamicMaxHand.GetMaxHand(target)
	}
	if len(target.Hand) < 3 && maxHand > 5 {
		damage += 2
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   target.ID,
		Damage:     damage,
		DamageType: "magic",
		Stage:      0,
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [灵魂震爆]：对 %s 造成%d点法术伤害", ctx.User.Name, target.Name, damage))
	return nil
}

func (h *SoulSorcererSoulGrantHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && soulBlue(ctx.User) >= 3
}

func (h *SoulSorcererSoulGrantHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("灵魂赐予上下文无效")
	}
	target := ctx.Target
	if target == nil && len(ctx.Targets) > 0 {
		target = ctx.Targets[0]
	}
	if target == nil {
		return fmt.Errorf("灵魂赐予需要目标角色")
	}
	if soulBlue(ctx.User) < 3 {
		return fmt.Errorf("蓝色灵魂不足3点")
	}
	addSoulBlue(ctx.User, -3)
	cap := playerEnergyCap(target)
	room := cap - (target.Gem + target.Crystal)
	if room < 0 {
		room = 0
	}
	gain := 2
	if room < gain {
		gain = room
	}
	target.Gem += gain
	ctx.Game.Log(fmt.Sprintf("%s 发动 [灵魂赐予]：%s +%d宝石", ctx.User.Name, target.Name, gain))
	return nil
}

func (h *SoulSorcererSoulLinkHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	allyIDs := soulSorcererAllyIDs(ctx.Game, ctx.User, false)
	if len(allyIDs) <= 1 {
		return false
	}
	if soulYellow(ctx.User) < 1 || soulBlue(ctx.User) < 1 {
		return false
	}
	if ctx.User.Character == nil {
		return false
	}
	return ctx.User.HasExclusiveCard(ctx.User.Character.Name, "灵魂链接")
}

func (h *SoulSorcererSoulLinkHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("灵魂链接上下文无效")
	}
	allyIDs := soulSorcererAllyIDs(ctx.Game, ctx.User, false)
	if len(allyIDs) <= 1 {
		return fmt.Errorf("队友数量不足，无法发动灵魂链接")
	}
	if soulYellow(ctx.User) < 1 || soulBlue(ctx.User) < 1 {
		return fmt.Errorf("灵魂不足，无法发动灵魂链接")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "ss_link_target",
			"user_id":     ctx.User.ID,
			"ally_ids":    allyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [灵魂链接]：请选择目标队友", ctx.User.Name))
	return nil
}

func (h *SoulSorcererSoulAmpHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && ctx.User.Gem > 0
}

func (h *SoulSorcererSoulAmpHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("灵魂增幅上下文无效")
	}
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("灵魂增幅需要1个红宝石")
	}
	ctx.User.Gem--
	y := addSoulYellow(ctx.User, 2)
	b := addSoulBlue(ctx.User, 2)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [灵魂增幅]：黄色灵魂+2（当前%d），蓝色灵魂+2（当前%d）", ctx.User.Name, y, b))
	return nil
}
