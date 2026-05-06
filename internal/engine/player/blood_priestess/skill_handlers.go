// gameflow: 血祭司 handler。

package blood_priestess

import (
	"fmt"
	"starcup-engine/internal/model"
)

type BloodSorrowHandler struct{}

type BleedingHandler struct{}

type BackflowHandler struct{}

type BloodWailHandler struct{}

type SharedLifeHandler struct{}

type BloodCurseHandler struct{}

func allTargetIDs(game model.IGameEngine) []string {
	if game == nil {
		return nil
	}
	var ids []string
	for _, p := range game.GetAllPlayers() {
		if p == nil {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

// --- BloodSorrow (血之哀伤) ---

func (h *BloodSorrowHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	_, fc := ctx.Game.FindFieldEffectBySource(model.EffectBloodSharedLife, ctx.User.ID)
	return fc != nil
}

func (h *BloodSorrowHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("血之哀伤上下文无效")
	}
	_, fc := ctx.Game.FindFieldEffectBySource(model.EffectBloodSharedLife, ctx.User.ID)
	if fc == nil {
		return fmt.Errorf("当前没有【同生共死】可转移或移除")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bp_blood_sorrow_mode",
			"user_id":     ctx.User.ID,
			"target_ids":  allTargetIDs(ctx.Game),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血之哀伤]：请选择后续效果，结算时会先对自己造成2点法术伤害", ctx.User.Name))
	return nil
}

// --- Bleeding (流血) ---

func (h *BleedingHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BleedingHandler) Execute(ctx *model.Context) error { return nil }

// --- Backflow (逆流) ---

func (h *BackflowHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && InBleedingForm(ctx.User) && len(ctx.User.Hand) >= 2
}

func (h *BackflowHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("逆流上下文无效")
	}
	if !InBleedingForm(ctx.User) {
		return fmt.Errorf("仅流血形态下可发动逆流")
	}
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [逆流]：弃2张牌并获得1点治疗", ctx.User.Name))
	return nil
}

// --- BloodWail (血之悲鸣) ---

func (h *BloodWailHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && InBleedingForm(ctx.User)
}

func (h *BloodWailHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("血之悲鸣上下文无效")
	}
	if !InBleedingForm(ctx.User) {
		return fmt.Errorf("仅流血形态下可发动血之悲鸣")
	}
	target := ctx.Target
	if target == nil && len(ctx.Targets) > 0 {
		target = ctx.Targets[0]
	}
	if target == nil {
		return fmt.Errorf("血之悲鸣需要目标角色")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bp_blood_wail_x",
			"user_id":     ctx.User.ID,
			"target_id":   target.ID,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血之悲鸣]：请选择X值（0~2）", ctx.User.Name))
	return nil
}

// --- SharedLife (同生共死) ---

func (h *SharedLifeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return false
	}
	return ctx.User.HasExclusiveCard(ctx.User.Character.ID, "同生共死")
}

func (h *SharedLifeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return fmt.Errorf("同生共死上下文无效")
	}
	if !ctx.User.HasExclusiveCard(ctx.User.Character.ID, "同生共死") {
		return fmt.Errorf("未找到【同生共死】专属技能卡")
	}
	targetIDs := allTargetIDs(ctx.Game)
	if len(targetIDs) == 0 {
		return fmt.Errorf("没有可选目标")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bp_shared_life_target",
			"user_id":     ctx.User.ID,
			"target_ids":  targetIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [同生共死]：选择目标后先摸2张牌，再放置【同生共死】", ctx.User.Name))
	return nil
}

// --- BloodCurse (血之诅咒) ---

func (h *BloodCurseHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && ctx.Game != nil
}

func (h *BloodCurseHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("血之诅咒上下文无效")
	}
	target := ctx.Target
	if target == nil && len(ctx.Targets) > 0 {
		target = ctx.Targets[0]
	}
	if target == nil {
		return fmt.Errorf("血之诅咒需要目标角色")
	}
	ctx.Game.AddPendingDamage(model.PendingDamage{
		SourceID:   ctx.User.ID,
		TargetID:   target.ID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	ctx.Game.AppendFlowContinuation(model.FlowContinuation{
		Kind:     model.FlowContinuationAfterDamage,
		RoleID:   "blood_priestess",
		PlayerID: ctx.User.ID,
		SkillID:  "bp_blood_curse",
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血之诅咒]：先对 %s 造成2点法术伤害，伤害结算后再弃3张牌", ctx.User.Name, target.Name))
	return nil
}
