// gameflow: 剑帝 handler。

package sword_emperor

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const swordEmperorSwordQiCap = 5

type SwordEmperorSwordSoulGuardHandler struct{ engineplayer.BaseHandler }

type SwordEmperorFeintHandler struct{ engineplayer.BaseHandler }

type SwordEmperorSwordQiSlashHandler struct{ engineplayer.BaseHandler }

type SwordEmperorAngelSoulHandler struct{ engineplayer.BaseHandler }

type SwordEmperorDemonSoulHandler struct{ engineplayer.BaseHandler }

type SwordEmperorAngelSoulHitHandler struct{ engineplayer.BaseHandler }

type SwordEmperorAngelSoulMissHandler struct{ engineplayer.BaseHandler }

type SwordEmperorDemonSoulMissHandler struct{ engineplayer.BaseHandler }

type SwordEmperorIndomitableWillHandler struct{ engineplayer.BaseHandler }

func (h *SwordEmperorSwordSoulGuardHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SwordEmperorSwordSoulGuardHandler) Execute(ctx *model.Context) error { return nil }

func (h *SwordEmperorFeintHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SwordEmperorFeintHandler) Execute(ctx *model.Context) error { return nil }

func (h *SwordEmperorAngelSoulHitHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SwordEmperorAngelSoulHitHandler) Execute(ctx *model.Context) error { return nil }

func (h *SwordEmperorAngelSoulMissHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SwordEmperorAngelSoulMissHandler) Execute(ctx *model.Context) error { return nil }

func (h *SwordEmperorDemonSoulMissHandler) CanUse(ctx *model.Context) bool { return false }

func (h *SwordEmperorDemonSoulMissHandler) Execute(ctx *model.Context) error { return nil }

func swordEmperorEnergy(user *model.Player) int {
	if user == nil {
		return 0
	}
	return user.Gem + user.Crystal
}

func swordEmperorSwordQi(user *model.Player) int {
	if user == nil {
		return 0
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	v := user.Tokens["se_sword_qi"]
	if v < 0 {
		v = 0
	}
	if v > swordEmperorSwordQiCap {
		v = swordEmperorSwordQiCap
	}
	user.Tokens["se_sword_qi"] = v
	return v
}

func addSwordEmperorSwordQi(user *model.Player, delta int) int {
	if user == nil {
		return 0
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	v := user.Tokens["se_sword_qi"] + delta
	if v < 0 {
		v = 0
	}
	if v > swordEmperorSwordQiCap {
		v = swordEmperorSwordQiCap
	}
	user.Tokens["se_sword_qi"] = v
	return v
}

func swordEmperorSwordSoulCards(user *model.Player) []*model.FieldCard {
	if user == nil {
		return nil
	}
	var out []*model.FieldCard
	for _, fc := range user.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSwordSoul {
			continue
		}
		out = append(out, fc)
	}
	return out
}

func swordEmperorConsumeSwordSoul(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	cards := swordEmperorSwordSoulCards(ctx.User)
	if len(cards) == 0 {
		return false
	}
	consume := cards[0]
	if err := ctx.Game.DiscardCard(consume); err != nil {
		return false
	}
	newField := make([]*model.FieldCard, 0, len(ctx.User.Field))
	removed := false
	for _, fc := range ctx.User.Field {
		if !removed && fc == consume {
			removed = true
			continue
		}
		newField = append(newField, fc)
	}
	ctx.User.Field = newField
	// se_sword_soul_count 已移除，服务端实时计算
	return true
}

func swordEmperorSlashTargets(game model.IGameEngine, user *model.Player, excludedID string) []string {
	if game == nil || user == nil {
		return nil
	}
	var ids []string
	for _, p := range game.GetAllPlayers() {
		if p == nil || p.ID == excludedID {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

func canPayCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.CanPayCrystalCost(ctx.User.ID, amount)
}

func spendCrystalLike(ctx *model.Context, amount int) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.Game.ConsumeCrystalCost(ctx.User.ID, amount)
}

func (h *SwordEmperorSwordQiSlashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnHitCheck {
		return false
	}
	if ctx.EventCtx.AttackInfo == nil {
		return false
	}
	if ctx.EventCtx.AttackInfo.ActionType != string(model.ActionAttack) {
		return false
	}
	if !ctx.EventCtx.AttackInfo.IsHit {
		return false
	}
	if ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if swordEmperorSwordQi(ctx.User) <= 0 {
		return false
	}
	return len(swordEmperorSlashTargets(ctx.Game, ctx.User, ctx.EventCtx.TargetID)) > 0
}

func (h *SwordEmperorSwordQiSlashHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return fmt.Errorf("剑气斩上下文无效")
	}
	maxX := swordEmperorSwordQi(ctx.User)
	if maxX > 3 {
		maxX = 3
	}
	if maxX <= 0 {
		return fmt.Errorf("剑气不足，无法发动剑气斩")
	}
	targetIDs := swordEmperorSlashTargets(ctx.Game, ctx.User, ctx.EventCtx.TargetID)
	if len(targetIDs) == 0 {
		return fmt.Errorf("没有可选的剑气斩目标")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "se_sword_qi_slash_x",
			"user_id":     ctx.User.ID,
			"max_x":       maxX,
			"target_ids":  targetIDs,
			"user_ctx":    ctx,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [剑气斩]：请选择X值", ctx.User.Name))
	return nil
}

func (h *SwordEmperorAngelSoulHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnAttackDeclared {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	energy := swordEmperorEnergy(ctx.User)
	if energy <= 0 || energy%2 == 0 {
		return false
	}
	return len(swordEmperorSwordSoulCards(ctx.User)) > 0
}

func (h *SwordEmperorAngelSoulHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("天使之魂上下文无效")
	}
	if !swordEmperorConsumeSwordSoul(ctx) {
		return fmt.Errorf("没有可移除的天使之魂")
	}
	ctx.User.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] = 1
	ctx.User.TurnState.UsedSkillCounts["se_angel_soul_armed"] = 1
	ctx.Game.Log(fmt.Sprintf("%s 发动 [天使之魂]：移除1张剑魂，本次攻击命中后+2治疗，未命中则我方士气+1", ctx.User.Name))
	return nil
}

func (h *SwordEmperorDemonSoulHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnAttackDeclared {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	energy := swordEmperorEnergy(ctx.User)
	if energy <= 0 || energy%2 != 0 {
		return false
	}
	return len(swordEmperorSwordSoulCards(ctx.User)) > 0
}

func (h *SwordEmperorDemonSoulHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("恶魔之魂上下文无效")
	}
	if !swordEmperorConsumeSwordSoul(ctx) {
		return fmt.Errorf("没有可移除的恶魔之魂")
	}
	ctx.User.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] = 1
	ctx.User.TurnState.UsedSkillCounts["se_demon_soul_armed"] = 1
	ctx.Game.ApplyNextAttackDamageRule(ctx.User.ID, "se_demon_soul_attack_bonus", "se_demon_soul", 1, model.RuleLifeThisEffectChain)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [恶魔之魂]：移除1张剑魂，本次攻击伤害额外+1，未命中则+2剑气", ctx.User.Name))
	return nil
}

func (h *SwordEmperorIndomitableWillHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnActionEnd {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionAttack {
		return false
	}
	return canPayCrystalLike(ctx, 1)
}

func (h *SwordEmperorIndomitableWillHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("不屈意志上下文无效")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("不屈意志需要1点蓝水晶（红宝石可替代）")
	}
	ctx.Game.DrawCards(ctx.User.ID, 1)
	now := addSwordEmperorSwordQi(ctx.User, 1)
	model.AppendAttackAction(ctx.User, "不屈意志")
	ctx.Game.Log(fmt.Sprintf("%s 发动 [不屈意志]：摸1张牌，剑气+1（当前%d），额外获得1次攻击行动", ctx.User.Name, now))
	return nil
}
