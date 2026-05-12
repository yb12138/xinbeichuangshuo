// gameflow: 瘟疫法师技能处理器。

package plague_mage

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"

	"starcup-engine/internal/model"
)

// --- 辅助函数 ---

func hasElementCard(p *model.Player, element model.Element) bool {
	for _, c := range p.Hand {
		if c.Element == element {
			return true
		}
	}
	return false
}

// --- 瘟疫法师技能处理器 ---

type PlagueImmortalHandler struct{ engineplayer.BaseHandler }

func (h *PlagueImmortalHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Timing != model.TimingOnActionEnd || ctx.EventCtx == nil {
		return false
	}
	if ctx.EventCtx.ActionType != model.ActionMagic {
		return false
	}
	return ctx.User.IsActive
}

func (h *PlagueImmortalHandler) Execute(ctx *model.Context) error {
	if ctx.User.TurnState.UsedSkillCounts["plague_block_immortal"] > 0 {
		ctx.User.TurnState.UsedSkillCounts["plague_block_immortal"] = 0
		ctx.Game.Log(fmt.Sprintf("%s 的 [不朽] 本次被技能效果抑制", ctx.User.Name))
		return nil
	}
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [不朽] 触发，+1治疗", ctx.User.Name))
	return nil
}

type PlagueBlasphemyHandler struct{ engineplayer.BaseHandler }

func (h *PlagueBlasphemyHandler) Execute(ctx *model.Context) error { return nil }

type PlagueOutbreakHandler struct{ engineplayer.BaseHandler }

func (h *PlagueOutbreakHandler) CanUse(ctx *model.Context) bool {
	return hasElementCard(ctx.User, model.ElementEarth)
}

func (h *PlagueOutbreakHandler) Execute(ctx *model.Context) error {
	ordered := engineplayer.ReversePlayersFromSlice(ctx.Game.GetAllPlayers(), ctx.User.ID)
	for _, p := range ordered {
		if p.ID == ctx.User.ID {
			continue
		}
		ctx.Game.AddPendingDamage(model.PendingDamage{
			SourceID:      ctx.User.ID,
			TargetID:      p.ID,
			Damage:        1,
			DamageType:    model.MagicAttack,
			SourceSkillID: "plague_outbreak",
		})
	}
	ctx.Game.Log(fmt.Sprintf("%s 发动 [瘟疫]，按逆序对其余角色各造成1点法术伤害", ctx.User.Name))
	return nil
}

type PlagueDeathTouchHandler struct{ engineplayer.BaseHandler }

func (h *PlagueDeathTouchHandler) CanUse(ctx *model.Context) bool {
	if ctx.User.Heal < 2 {
		return false
	}
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		if c.Element != "" {
			counts[c.Element]++
		}
	}
	for _, n := range counts {
		if n >= 2 {
			return true
		}
	}
	return false
}

func (h *PlagueDeathTouchHandler) Execute(ctx *model.Context) error {
	if ctx.Target == nil {
		return fmt.Errorf("死亡之触需要1名敌方目标")
	}
	if ctx.User.Heal < 2 {
		return fmt.Errorf("死亡之触需要至少2点治疗")
	}
	counts := map[model.Element]int{}
	for _, c := range ctx.User.Hand {
		if c.Element != "" {
			counts[c.Element]++
		}
	}
	var elements []string
	for _, ele := range []model.Element{
		model.ElementEarth, model.ElementWater, model.ElementFire,
		model.ElementWind, model.ElementThunder, model.ElementLight, model.ElementDark,
	} {
		if counts[ele] >= 2 {
			elements = append(elements, string(ele))
		}
	}
	if len(elements) == 0 {
		return fmt.Errorf("死亡之触需要至少2张同系牌")
	}
	// 该技能不触发不朽：先设置抑制标记，覆盖 UseSkill 的阶段结束触发。
	ctx.User.TurnState.UsedSkillCounts["plague_block_immortal"] = 1
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":      "plague_death_touch_element",
			"user_id":          ctx.User.ID,
			"target_id":        ctx.Target.ID,
			"elements":         elements,
			"max_heal":         ctx.User.Heal,
			"element_counts":   counts,
			"selected_indices": []int{},
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [死亡之触]，等待选择X/Y与目标", ctx.User.Name))
	return nil
}

type PlagueToxicNovaHandler struct{ engineplayer.BaseHandler }

func (h *PlagueToxicNovaHandler) CanUse(ctx *model.Context) bool {
	return ctx.User.Gem > 0
}

func (h *PlagueToxicNovaHandler) Execute(ctx *model.Context) error {
	ordered := engineplayer.ReversePlayersFromSlice(ctx.Game.GetAllPlayers(), ctx.User.ID)
	for _, p := range ordered {
		if p.ID == ctx.User.ID {
			continue
		}
		ctx.Game.AddPendingDamage(model.PendingDamage{
			SourceID:   ctx.User.ID,
			TargetID:   p.ID,
			Damage:     2,
			DamageType: model.MagicAttack,
		})
	}
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [剧毒新星]，对其余角色各造成2点法术伤害", ctx.User.Name))
	return nil
}
