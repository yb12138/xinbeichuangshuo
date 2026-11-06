// gameflow: 神官技能处理器。

package priest

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// BaseHandler provides empty defaults for CanUse (always true) and Execute (no-op).
type BaseHandler struct{}

func (h *BaseHandler) CanUse(ctx *model.Context) bool   { return true }
func (h *BaseHandler) Execute(ctx *model.Context) error { return nil }

// --- Priest Skill Handlers ---

type PriestDivineRevelationHandler struct{ BaseHandler }

type PriestDivineBlessHandler struct{ BaseHandler }

type PriestWaterPowerHandler struct{ BaseHandler }

type PriestGuardianHandler struct{ BaseHandler }

type PriestDivineContractHandler struct{ BaseHandler }

type PriestDivineDomainHandler struct{ BaseHandler }

func (h *PriestDivineRevelationHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.EventCtx == nil {
		return false
	}
	if ctx.Timing != model.TimingActionEnd {
		return false
	}
	return ctx.EventCtx.ActionType == model.ActionBuy ||
		ctx.EventCtx.ActionType == model.ActionSynthesize ||
		ctx.EventCtx.ActionType == model.ActionExtract
}

func (h *PriestDivineRevelationHandler) Execute(ctx *model.Context) error {
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 的 [神圣启示] 触发，+1治疗", ctx.User.Name))
	return nil
}

func (h *PriestDivineBlessHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	cnt := 0
	for _, c := range ctx.User.Hand {
		if c.Type == model.CardTypeMagic {
			cnt++
		}
	}
	return cnt >= 2
}

func (h *PriestDivineBlessHandler) Execute(ctx *model.Context) error {
	ctx.Game.Heal(ctx.User.ID, 2)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [神圣祈福]，恢复2点治疗", ctx.User.Name))
	return nil
}

func (h *PriestWaterPowerHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	// 只剩 1 张水系牌时也可发动（部分发动：只弃这 1 张，不交牌）。
	return hasElementCard(ctx.User, model.ElementWater)
}

func (h *PriestWaterPowerHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("上下文无效")
	}
	target := ctx.Target
	if target == nil || target.Camp != ctx.User.Camp || target.ID == ctx.User.ID {
		return fmt.Errorf("水之神力需要指定队友")
	}

	discardedAny, _ := ctx.Selections["discardedCards"]
	discarded, _ := discardedAny.([]model.Card)
	if len(discarded) == 0 || discarded[0].Element != model.ElementWater {
		return fmt.Errorf("水之神力需要先弃置1张水系牌")
	}
	ctx.Game.Log(fmt.Sprintf("%s 为 [水之神力] 弃置了 %s", ctx.User.Name, discarded[0].Name))

	gave := false
	if len(discarded) >= 2 {
		give := discarded[1]
		target.Hand = append(target.Hand, give)
		gave = true
		// 交给队友的牌不公开牌面。
		ctx.Game.Log(fmt.Sprintf("%s 将1张手牌交给了 %s", ctx.User.Name, target.Name))
	} else {
		ctx.Game.Log(fmt.Sprintf("%s 没有其他手牌可交给 %s（部分发动）", ctx.User.Name, target.Name))
	}
	ctx.Game.Heal(ctx.User.ID, 1)
	ctx.Game.Heal(target.ID, 1)
	ctx.Game.Log(fmt.Sprintf("%s 发动 [水之神力]，与 %s 各+1治疗", ctx.User.Name, target.Name))
	// 队友收牌后检查手牌上限。
	if gave {
		ctx.Game.CheckHandLimit(target.ID, true)
	}
	return nil
}

func (h *PriestGuardianHandler) Execute(ctx *model.Context) error { return nil }

func (h *PriestDivineContractHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return false
	}
	return ctx.User.Heal > 0 && engineplayer.CanPayCrystalLike(ctx, 1) && len(priestDivineContractTargets(ctx.Game, ctx.User)) > 0
}

func (h *PriestDivineContractHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("神圣契约上下文无效")
	}
	if ctx.User.Heal <= 0 {
		return fmt.Errorf("神圣契约需要可转移治疗")
	}
	targetIDs := priestDivineContractTargets(ctx.Game, ctx.User)
	if len(targetIDs) == 0 {
		return fmt.Errorf("神圣契约需要至少1名其他队友")
	}
	// CostCrystal 已在 ConfirmStartupSkillAction 由框架统一扣减（见 skill definition CostCrystal: 1）
	waitingPhase := priestDivineContractWaitingPhase(ctx)
	resumePhase := priestDivineContractResumePhase(ctx)
	if ctx.Target == nil {
		ctx.Game.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: ctx.User.ID,
			Context: map[string]interface{}{
				"choice_type":   "priest_divine_contract_target",
				"user_id":       ctx.User.ID,
				"ally_ids":      targetIDs,
				"max_x":         ctx.User.Heal,
				"waiting_phase": waitingPhase,
				"resume_phase":  resumePhase,
			},
		})
		ctx.Game.Log(fmt.Sprintf("%s 发动 [神圣契约]，请选择目标队友", ctx.User.Name))
		return nil
	}
	if ctx.Target.Camp != ctx.User.Camp || ctx.Target.ID == ctx.User.ID {
		return fmt.Errorf("神圣契约目标必须是其他队友")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":   "priest_divine_contract_x",
			"user_id":       ctx.User.ID,
			"target_id":     ctx.Target.ID,
			"max_x":         ctx.User.Heal,
			"waiting_phase": waitingPhase,
			"resume_phase":  resumePhase,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [神圣契约]，请选择转移治疗值X（目标：%s）", ctx.User.Name, ctx.Target.Name))
	return nil
}

func priestDivineContractTargets(game model.IGameEngine, user *model.Player) []string {
	if game == nil || user == nil {
		return nil
	}
	targetIDs := make([]string, 0)
	for _, player := range game.GetAllPlayers() {
		if player == nil || player.ID == user.ID || player.Camp != user.Camp {
			continue
		}
		targetIDs = append(targetIDs, player.ID)
	}
	return targetIDs
}

func priestDivineContractWaitingPhase(_ *model.Context) model.TurnStage {
	return model.TurnStageActionStart
}

func priestDivineContractResumePhase(_ *model.Context) model.TurnStage {
	return model.TurnStageActionExecution
}

func (h *PriestDivineDomainHandler) CanUse(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	return engineplayer.CanPayCrystalLike(ctx, 1)
}

func (h *PriestDivineDomainHandler) Execute(ctx *model.Context) error {
	if ctx == nil || ctx.User == nil || ctx.Game == nil {
		return fmt.Errorf("神圣领域上下文无效")
	}
	allyIDs := []string{}
	allTargetIDs := []string{}
	for _, p := range ctx.Game.GetAllPlayers() {
		if p == nil {
			continue
		}
		// 伤害分支文案为"目标角色"：可指定包括自己在内的任意角色。
		allTargetIDs = append(allTargetIDs, p.ID)
		if p.Camp == ctx.User.Camp && p.ID != ctx.User.ID {
			allyIDs = append(allyIDs, p.ID)
		}
	}
	modeOptions := []string{}
	if ctx.User.Heal > 0 {
		modeOptions = append(modeOptions, "damage")
	}
	if len(allyIDs) > 0 {
		modeOptions = append(modeOptions, "heal")
	}
	if len(modeOptions) == 0 {
		return fmt.Errorf("神圣领域当前无可用分支（伤害分支需至少1点治疗，治疗分支需至少1名队友）")
	}
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type":     "priest_divine_domain_mode",
			"user_id":         ctx.User.ID,
			"mode_options":    modeOptions,
			"all_target_ids":  allTargetIDs,
			"ally_target_ids": allyIDs,
		},
	})
	ctx.Game.Log(fmt.Sprintf("%s 发动 [神圣领域]，等待选择分支", ctx.User.Name))
	return nil
}

// --- Helper functions for priest skills ---

func hasElementCard(p *model.Player, element model.Element) bool {
	for _, c := range p.Hand {
		if c.Element == element {
			return true
		}
	}
	return false
}
