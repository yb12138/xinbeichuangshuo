// gameflow: 格斗家 handler。

package fighter

import (
	"fmt"
	"sort"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const fighterQiCap = 8

type FighterPsiFieldHandler struct{ engineplayer.BaseHandler }

type FighterChargeStrikeHandler struct{ engineplayer.BaseHandler }

type FighterPsiBulletHandler struct{ engineplayer.BaseHandler }

type FighterHundredDragonHandler struct{ engineplayer.BaseHandler }

type FighterBurstCrashHandler struct{ engineplayer.BaseHandler }

type FighterWarGodDriveHandler struct{ engineplayer.BaseHandler }

func (h *FighterPsiFieldHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil {
		return false
	}
	return ctx.Timing == model.TimingOnDamageTaken && *ctx.EventCtx.DamageVal > 4
}

func (h *FighterPsiFieldHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.EventCtx == nil || ctx.EventCtx.DamageVal == nil || ctx.User == nil || ctx.Game == nil {
		return nil
	}
	if *ctx.EventCtx.DamageVal > 4 {
		*ctx.EventCtx.DamageVal = 4
		ctx.Game.Log(fmt.Sprintf("%s 的 [念气力场] 生效：本次伤害被限制为4", ctx.User.Name))
	}
	return nil
}

func (h *FighterChargeStrikeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnAttackDeclared {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if hasForm(ctx.User, model.FormFighterHundredDragon) {
		return false
	}
	if getSkillFlow(ctx.User, "fighter_attack_start_skill_lock") > 0 {
		return false
	}
	return getToken(ctx.User, "fighter_qi") < fighterQiCap
}

func (h *FighterChargeStrikeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("蓄力一击上下文无效")
	}
	if getToken(ctx.User, "fighter_qi") >= fighterQiCap {
		return fmt.Errorf("斗气已达上限，不能发动蓄力一击")
	}
	qi := addToken(ctx.User, "fighter_qi", 1, 0, fighterQiCap)
	setSkillFlow(ctx.User, "fighter_attack_start_skill_lock", 1)
	setSkillFlow(ctx.User, "fighter_charge_pending", 1)
	ctx.Game.ApplyNextAttackDamageRule(ctx.User.ID, "fighter_charge_attack_bonus", "fighter_charge_strike", 1, model.RuleLifeThisEffectChain)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [蓄力一击]：斗气+1（当前%d），本次攻击伤害额外+1；若未命中将按斗气自伤", ctx.User.Name, qi))
	return nil
}

func (h *FighterPsiBulletHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnActionEnd {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionMagic {
		return false
	}
	if getToken(ctx.User, "fighter_qi") >= fighterQiCap {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp != ctx.User.Camp {
			return true
		}
	}
	return false
}

func (h *FighterPsiBulletHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("念弹上下文无效")
	}
	if getToken(ctx.User, "fighter_qi") >= fighterQiCap {
		return fmt.Errorf("斗气已达上限，不能发动念弹")
	}
	targetIDs := make([]string, 0)
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil || p.Camp == ctx.User.Camp {
			continue
		}
		targetIDs = append(targetIDs, p.ID)
	}
	if len(targetIDs) == 0 {
		return nil
	}
	qi := addToken(ctx.User, "fighter_qi", 1, 0, fighterQiCap)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "fighter_psi_bullet_target",
			"user_id":     ctx.User.ID,
			"target_ids":  targetIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [念弹]：斗气+1（当前%d），请选择目标对手", ctx.User.Name, qi))
	return nil
}

func (h *FighterHundredDragonHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	if hasForm(ctx.User, model.FormFighterHundredDragon) {
		return false
	}
	if getToken(ctx.User, "fighter_qi") < 3 || ctx.Game == nil {
		return false
	}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p != nil && p.Camp != ctx.User.Camp {
			return true
		}
	}
	return false
}

func (h *FighterHundredDragonHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("百式幻龙拳上下文无效")
	}
	if getToken(ctx.User, "fighter_qi") < 3 {
		return fmt.Errorf("斗气不足3，无法发动百式幻龙拳")
	}
	targetIDs := make([]string, 0)
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil || p.Camp == ctx.User.Camp {
			continue
		}
		targetIDs = append(targetIDs, p.ID)
	}
	if len(targetIDs) == 0 {
		return fmt.Errorf("百式幻龙拳没有可锁定的敌方目标")
	}
	qi := addToken(ctx.User, "fighter_qi", -3, 0, fighterQiCap)
	enterForm(ctx.User, model.FormFighterHundredDragon)
	setSkillFlow(ctx.User, "fighter_hundred_dragon_target_order", 0)
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "fighter_hundred_dragon_target",
			"user_id":       ctx.User.ID,
			"target_ids":    targetIDs,
			"waiting_phase": model.TurnStageActionExecution,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [百式幻龙拳]：移除3斗气（剩余%d），进入持续形态，请选择本行动阶段锁定目标", ctx.User.Name, qi))
	return nil
}

func (h *FighterBurstCrashHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingOnAttackDeclared {
		return false
	}
	if ctx.EventCtx.AttackInfo != nil && ctx.EventCtx.AttackInfo.CounterInitiator != "" {
		return false
	}
	if getSkillFlow(ctx.User, "fighter_attack_start_skill_lock") > 0 {
		return false
	}
	return getToken(ctx.User, "fighter_qi") > 0
}

func (h *FighterBurstCrashHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("气绝崩击上下文无效")
	}
	if getToken(ctx.User, "fighter_qi") <= 0 {
		return fmt.Errorf("斗气不足，无法发动气绝崩击")
	}
	qi := addToken(ctx.User, "fighter_qi", -1, 0, fighterQiCap)
	setSkillFlow(ctx.User, "fighter_attack_start_skill_lock", 2)
	setSkillFlow(ctx.User, "fighter_qiburst_force_no_counter", 1)
	if ctx.EventCtx != nil && ctx.EventCtx.AttackInfo != nil {
		ctx.EventCtx.AttackInfo.CanBeResponded = false
	}
	if qi > 0 {
		ctx.Game.AddPendingDamage(model.PendingDamage{
			SourceID:   ctx.User.ID,
			TargetID:   ctx.User.ID,
			Damage:     qi,
			DamageType: model.MagicAttack,
		})
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [气绝崩击]：移除1斗气（剩余%d），本次攻击不可应战，并对自己造成%d点法术伤害", ctx.User.Name, qi, qi))
	return nil
}

func (h *FighterWarGodDriveHandler) CanUse(ctx *model.Context) bool {
	return canPayCrystalLike(ctx, 1)
}

func (h *FighterWarGodDriveHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("斗神天驱上下文无效")
	}
	if !spendCrystalLike(ctx, 1) {
		return fmt.Errorf("斗神天驱需要1点蓝水晶（红宝石可替代）")
	}
	discardCount := len(ctx.User.Hand) - 3
	if discardCount > 0 {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type":                 "system_discard_cards",
				"discard_subflow":             true,
				"flow_continuation_role_id":   "fighter",
				"flow_continuation_player_id": ctx.User.ID,
				"flow_continuation_skill_id":  "fighter_war_god_drive",
				"discard_count":               discardCount,
				"stay_in_turn":                true,
				"prompt":                      "【斗神天驱】请选择需要弃置的手牌：",
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [斗神天驱]：请先弃置%d张牌（弃到3张）", ctx.User.Name, discardCount))
		return nil
	}
	ctx.Game.Heal(ctx.User.ID, 2)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [斗神天驱]：手牌无需弃置，获得2点治疗", ctx.User.Name))
	return nil
}

// Helper functions

func getToken(p *model.Player, key string) int {
	if p == nil {
		return 0
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	return p.Tokens[key]
}

func setToken(p *model.Player, key string, v int) {
	if p == nil {
		return
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens[key] = v
}

func getSkillFlow(p *model.Player, key string) int {
	if p == nil || p.TurnState.SkillFlowState == nil {
		return 0
	}
	return p.TurnState.SkillFlowState[key]
}

func setSkillFlow(p *model.Player, key string, v int) {
	if p == nil {
		return
	}
	if p.TurnState.SkillFlowState == nil {
		p.TurnState.SkillFlowState = make(map[string]int)
	}
	p.TurnState.SkillFlowState[key] = v
}

func addToken(p *model.Player, key string, delta int, minV int, maxV int) int {
	cur := getToken(p, key)
	cur += delta
	if cur < minV {
		cur = minV
	}
	if maxV >= minV && cur > maxV {
		cur = maxV
	}
	setToken(p, key, cur)
	return cur
}

func hasForm(p *model.Player, form string) bool {
	return p != nil && p.Form == form
}

func enterForm(p *model.Player, form string) {
	if p == nil {
		return
	}
	p.Orientation = model.OrientationTapped
	p.Form = form
}

// 工具：为了避免不同阶段删除手牌导致索引错乱，批量删除时统一按降序处理。
func removeHandByIndices(p *model.Player, indices []int) []model.Card {
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	var out []model.Card
	for _, i := range indices {
		if i < 0 || i >= len(p.Hand) {
			continue
		}
		out = append(out, p.Hand[i])
		p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
	}
	return out
}

// 红宝石可替代蓝水晶（仅水晶消耗方向）
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
