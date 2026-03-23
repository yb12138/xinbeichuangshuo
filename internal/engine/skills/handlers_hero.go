package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

const heroTokenCap = 4

type HeroHeartHandler struct{ BaseHandler }

type HeroRoarHandler struct{ BaseHandler }

type HeroForbiddenPowerHandler struct{ BaseHandler }

type HeroExhaustionHandler struct{ BaseHandler }

type HeroCalmMindHandler struct{ BaseHandler }

type HeroTauntHandler struct{ BaseHandler }

type HeroDeadDuelHandler struct{ BaseHandler }

func (h *HeroHeartHandler) CanUse(ctx *model.Context) bool { return false }

func (h *HeroHeartHandler) Execute(ctx *model.Context) error { return nil }

func (h *HeroExhaustionHandler) CanUse(ctx *model.Context) bool { return false }

func (h *HeroExhaustionHandler) Execute(ctx *model.Context) error { return nil }

func (h *HeroRoarHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return getToken(ctx.User, "hero_anger") > 0
}

func (h *HeroRoarHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("怒吼上下文无效")
	}
	if getToken(ctx.User, "hero_anger") <= 0 {
		return fmt.Errorf("怒气不足，无法发动怒吼")
	}
	addToken(ctx.User, "hero_anger", -1, 0, heroTokenCap)
	setToken(ctx.User, "hero_roar_active", 1)
	setToken(ctx.User, "hero_roar_damage_pending", 1)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "hero_roar_draw",
			"user_id":     ctx.User.ID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [怒吼]：移除1点怒气，本次攻击伤害额外+2，等待选择摸牌数量", ctx.User.Name))
	return nil
}

func (h *HeroForbiddenPowerHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackHit && ctx.Trigger != model.TriggerOnAttackMiss {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *HeroForbiddenPowerHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("禁断之力上下文无效")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("禁断之力需要1点蓝水晶（红宝石可替代）")
	}

	handCards := append([]model.Card{}, ctx.User.Hand...)
	magicCount := 0
	waterCount := 0
	fireCount := 0
	for _, c := range handCards {
		if c.Type == model.CardTypeMagic {
			magicCount++
		}
		if c.Element == model.ElementWater {
			waterCount++
		}
		if c.Element == model.ElementFire {
			fireCount++
		}
	}

	if len(handCards) > 0 {
		ctx.Game.NotifyCardRevealed(ctx.User.ID, handCards, "discard")
		ctx.Game.AppendToDiscard(handCards)
		ctx.User.Hand = ctx.User.Hand[:0]
	}

	anger := addToken(ctx.User, "hero_anger", magicCount, 0, heroTokenCap)
	wisdomGain := 0
	if ctx.Trigger == model.TriggerOnAttackMiss {
		before := getToken(ctx.User, "hero_wisdom")
		after := addToken(ctx.User, "hero_wisdom", waterCount, 0, heroTokenCap)
		wisdomGain = after - before
	}

	if ctx.Trigger == model.TriggerOnAttackHit && fireCount > 0 {
		if ctx.TriggerCtx != nil && ctx.TriggerCtx.DamageVal != nil {
			*ctx.TriggerCtx.DamageVal += fireCount
		}
		ctx.Game.AddPendingDamage(model.PendingDamage{
			SourceID:   ctx.User.ID,
			TargetID:   ctx.User.ID,
			Damage:     fireCount,
			DamageType: "magic",
			Stage:      0,
		})
	}

	// 精疲力竭：强制进入状态并追加1次攻击行动。
	setToken(ctx.User, "hero_exhaustion_form", 1)
	setToken(ctx.User, "hero_exhaustion_release_pending", 1)
	addAttackAction(ctx.User, "精疲力竭")

	switch ctx.Trigger {
	case model.TriggerOnAttackHit:
		ctx.Game.Log(fmt.Sprintf("%s 发动 [禁断之力]：弃掉%d张手牌（法术%d/火%d），怒气=%d；本次攻击伤害额外+%d并对自己造成%d点法术伤害；进入精疲力竭并获得额外攻击行动",
			ctx.User.Name, len(handCards), magicCount, fireCount, anger, fireCount, fireCount))
	case model.TriggerOnAttackMiss:
		ctx.Game.Log(fmt.Sprintf("%s 发动 [禁断之力]：弃掉%d张手牌（法术%d/水系%d），怒气=%d，知性+%d；进入精疲力竭并获得额外攻击行动",
			ctx.User.Name, len(handCards), magicCount, waterCount, anger, wisdomGain))
	}
	return nil
}

func (h *HeroCalmMindHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnAttackStart {
		return false
	}
	if ctx.TriggerCtx.AttackInfo != nil && ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	return getToken(ctx.User, "hero_wisdom") >= 4
}

func (h *HeroCalmMindHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return fmt.Errorf("明镜止水上下文无效")
	}
	if getToken(ctx.User, "hero_wisdom") < 4 {
		return fmt.Errorf("知性不足，无法发动明镜止水")
	}
	addToken(ctx.User, "hero_wisdom", -4, 0, heroTokenCap)
	if ctx.TriggerCtx.AttackInfo != nil {
		ctx.TriggerCtx.AttackInfo.CanBeResponded = false
	}
	setToken(ctx.User, "hero_calm_force_no_counter", 1)
	setToken(ctx.User, "hero_calm_end_crystal_pending", getToken(ctx.User, "hero_calm_end_crystal_pending")+1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [明镜止水]：移除4点知性，本次攻击无法应战（攻击结束时+1水晶）", ctx.User.Name))
	return nil
}

func (h *HeroTauntHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && getToken(ctx.User, "hero_anger") > 0
}

func (h *HeroTauntHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("挑衅上下文无效")
	}
	if getToken(ctx.User, "hero_anger") <= 0 {
		return fmt.Errorf("怒气不足，无法发动挑衅")
	}
	target := ctx.Target
	if target == nil && len(ctx.Targets) > 0 {
		target = ctx.Targets[0]
	}
	if target == nil {
		return fmt.Errorf("挑衅需要指定目标")
	}
	if target.Camp == ctx.User.Camp {
		return fmt.Errorf("挑衅只能指定敌方角色")
	}
	addToken(ctx.User, "hero_anger", -1, 0, heroTokenCap)
	wisdom := addToken(ctx.User, "hero_wisdom", 1, 0, heroTokenCap)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [挑衅] 指向 %s：移除1点怒气并使自己知性+1（当前%d）", ctx.User.Name, target.Name, wisdom))
	return nil
}

func (h *HeroDeadDuelHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil {
		return false
	}
	if ctx.Trigger != model.TriggerOnDamageTaken {
		return false
	}
	if ctx.TriggerCtx.DamageVal == nil || *ctx.TriggerCtx.DamageVal <= 0 {
		return false
	}
	if !ctx.Flags["IsMagicDamage"] {
		return false
	}
	return ctx.User.Gem > 0
}

func (h *HeroDeadDuelHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("死斗上下文无效")
	}
	if ctx.User.Gem <= 0 {
		return fmt.Errorf("死斗需要1个红宝石")
	}
	ctx.User.Gem--
	anger := addToken(ctx.User, "hero_anger", 3, 0, heroTokenCap)
	setToken(ctx.User, "hero_dead_duel_pending", 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [死斗]：消耗1红宝石，怒气+3（当前%d）；若本次法术伤害导致士气下降，则该次下降值恒定为1", ctx.User.Name, anger))
	return nil
}
