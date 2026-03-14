package skills

import (
	"fmt"
	"starcup-engine/internal/model"
)

type BloodPriestessBloodSorrowHandler struct{ BaseHandler }

type BloodPriestessBleedingHandler struct{ BaseHandler }

type BloodPriestessBackflowHandler struct{ BaseHandler }

type BloodPriestessBloodWailHandler struct{ BaseHandler }

type BloodPriestessSharedLifeHandler struct{ BaseHandler }

type BloodPriestessBloodCurseHandler struct{ BaseHandler }

func bloodPriestessFindSharedLife(game model.IGameEngine, sourceID string) (*model.Player, *model.FieldCard) {
	if game == nil || sourceID == "" {
		return nil, nil
	}
	for _, p := range game.GetAllPlayers() {
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
				continue
			}
			if fc.SourceID == sourceID {
				return p, fc
			}
		}
	}
	return nil, nil
}

func bloodPriestessAllTargetIDs(game model.IGameEngine) []string {
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

func (h *BloodPriestessBloodSorrowHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	_, fc := bloodPriestessFindSharedLife(ctx.Game, ctx.User.ID)
	return fc != nil
}

func (h *BloodPriestessBloodSorrowHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("血之哀伤上下文无效")
	}
	_, fc := bloodPriestessFindSharedLife(ctx.Game, ctx.User.ID)
	if fc == nil {
		return fmt.Errorf("当前没有【同生共死】可转移或移除")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "bp_blood_sorrow_mode",
			"user_id":     ctx.User.ID,
			"target_ids":  bloodPriestessAllTargetIDs(ctx.Game),
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血之哀伤]：先承受2点法术伤害，再选择转移或移除【同生共死】", ctx.User.Name))
	return nil
}

func (h *BloodPriestessBleedingHandler) CanUse(ctx *model.Context) bool { return false }

func (h *BloodPriestessBleedingHandler) Execute(ctx *model.Context) error { return nil }

func (h *BloodPriestessBackflowHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && getToken(ctx.User, "bp_bleed_form") > 0 && len(ctx.User.Hand) >= 2
}

func (h *BloodPriestessBackflowHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("逆流上下文无效")
	}
	if getToken(ctx.User, "bp_bleed_form") <= 0 {
		return fmt.Errorf("仅流血形态下可发动逆流")
	}
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [逆流]：弃2张牌并获得1点治疗", ctx.User.Name))
	return nil
}

func (h *BloodPriestessBloodWailHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && getToken(ctx.User, "bp_bleed_form") > 0
}

func (h *BloodPriestessBloodWailHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("血之悲鸣上下文无效")
	}
	if getToken(ctx.User, "bp_bleed_form") <= 0 {
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

func (h *BloodPriestessSharedLifeHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return false
	}
	return ctx.User.HasExclusiveCard(ctx.User.Character.Name, "同生共死")
}

func (h *BloodPriestessSharedLifeHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil || ctx.User.Character == nil {
		return fmt.Errorf("同生共死上下文无效")
	}
	if !ctx.User.HasExclusiveCard(ctx.User.Character.Name, "同生共死") {
		return fmt.Errorf("未找到【同生共死】专属技能卡")
	}
	targetIDs := bloodPriestessAllTargetIDs(ctx.Game)
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
	ctx.Game.Log(fmt.Sprintf("%s 发动 [同生共死]：先摸2张牌，再选择放置目标", ctx.User.Name))
	return nil
}

func (h *BloodPriestessBloodCurseHandler) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.User != nil && ctx.Game != nil
}

func (h *BloodPriestessBloodCurseHandler) Execute(ctx *model.Context) error {
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
		DamageType: "magic",
		Stage:      0,
	})
	discardNeed := 3
	if len(ctx.User.Hand) < discardNeed {
		discardNeed = len(ctx.User.Hand)
	}
	if discardNeed > 0 {
		remaining := make([]int, 0, len(ctx.User.Hand))
		for i := range ctx.User.Hand {
			remaining = append(remaining, i)
		}
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type":       "bp_curse_discard",
				"user_id":           ctx.User.ID,
				"discard_count":     discardNeed,
				"remaining_indices": remaining,
				"selected_indices":  []int{},
			},
		})
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [血之诅咒]：对 %s 造成2点法术伤害，并弃%d张牌", ctx.User.Name, target.Name, discardNeed))
	return nil
}
