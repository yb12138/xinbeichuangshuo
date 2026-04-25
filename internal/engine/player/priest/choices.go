// gameflow: 神官角色选择流。

package priest

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "priest_divine_contract_target":
		allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
		options := make([]model.PromptOption, 0, len(allyIDs))
		for _, allyID := range allyIDs {
			if target := rt.GetPlayers()[allyID]; target != nil {
				options = append(options, model.PromptOption{ID: allyID, Label: target.Name})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【神圣契约】请选择1名队友：", Options: options, Min: 1, Max: 1}

	case "priest_divine_contract_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		targetID, _ := data["target_id"].(string)
		targetName := targetID
		targetHeal := -1
		if target := rt.GetPlayers()[targetID]; target != nil {
			targetName = target.Name
			targetHeal = target.Heal
		}
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("转移 %d 点治疗", x)})
		}
		message := "【神圣契约】请选择转移治疗值X："
		if targetName != "" {
			if targetHeal >= 0 {
				message = fmt.Sprintf("【神圣契约】请选择转移治疗值X（目标：%s，当前治疗%d）：", targetName, targetHeal)
			} else {
				message = fmt.Sprintf("【神圣契约】请选择转移治疗值X（目标：%s）：", targetName)
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1}

	case "priest_divine_domain_mode":
		modeOptions := runtimeutil.ParseStringSliceContextValue(data["mode_options"])
		options := make([]model.PromptOption, 0, len(modeOptions))
		for _, mode := range modeOptions {
			switch mode {
			case "damage":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支①：移除1治疗，对任意角色造成2点法术伤害"})
			case "heal":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支②：你+2治疗，1名队友+1治疗"})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【神圣领域】请选择发动分支：", Options: options, Min: 1, Max: 1}

	case "priest_divine_domain_damage_target", "priest_divine_domain_heal_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if target := rt.GetPlayers()[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
			}
		}
		message := "请选择目标："
		if choiceType == "priest_divine_domain_damage_target" {
			message = "【神圣领域·分支①】请选择2点法术伤害目标："
		} else {
			message = "【神圣领域·分支②】请选择+1治疗的队友："
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1}
	}

	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "priest_divine_contract_target":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := rt.GetPlayers()[allyIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("队友不存在")
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if maxX <= 0 {
			maxX = user.Heal
		}
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"choice_type":   "priest_divine_contract_x",
				"user_id":       user.ID,
				"target_id":     target.ID,
				"max_x":         maxX,
				"waiting_phase": ctxData["waiting_phase"],
				"resume_phase":  ctxData["resume_phase"],
			},
		})
		rt.Log(fmt.Sprintf("%s 的 [神圣契约] 选择 %s 为目标，继续选择转移治疗值X", user.Name, target.Name))
		rt.PopInterrupt()
		return true, nil

	case "priest_divine_contract_x":
		userID, _ := ctxData["user_id"].(string)
		targetID, _ := ctxData["target_id"].(string)
		user := rt.GetPlayers()[userID]
		target := rt.GetPlayers()[targetID]
		if user == nil || target == nil {
			return true, fmt.Errorf("神圣契约目标不存在")
		}
		if target.Camp != user.Camp {
			return true, fmt.Errorf("神圣契约目标必须是队友")
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		xValue := selectionIndex + 1
		if xValue < 1 || xValue > maxX {
			return true, fmt.Errorf("无效的X值")
		}
		if xValue > user.Heal {
			return true, fmt.Errorf("当前治疗不足，无法转移%d点治疗", xValue)
		}

		before := target.Heal
		user.Heal -= xValue
		if before <= 4 {
			target.Heal = before + xValue
			if target.Heal > 4 {
				target.Heal = 4
			}
		}
		after := target.Heal
		if before > 4 {
			rt.Log(fmt.Sprintf("%s 的 [神圣契约] 生效：移除自身%d点治疗；目标 %s 当前治疗已超过4（%d），保持不变", user.Name, xValue, target.Name, before))
		} else {
			rt.Log(fmt.Sprintf("%s 的 [神圣契约] 生效：移除自身%d点治疗并转移给 %s（%d -> %d）", user.Name, xValue, target.Name, before, after))
		}

		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			// 规则：神圣契约是"选择目标+选择X"的两段式结算，最终恢复点必须由上游显式给出。
			rt.ApplyChoiceResumePoint(mustChoiceResumePointFromMap(ctxData, "resume_phase"))
		}
		return true, nil

	case "priest_divine_domain_mode":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		modeOptions := runtimeutil.ParseStringSliceContextValue(ctxData["mode_options"])
		if selectionIndex < 0 || selectionIndex >= len(modeOptions) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modeOptions[selectionIndex]
		switch mode {
		case "damage":
			allTargets := runtimeutil.ParseStringSliceContextValue(ctxData["all_target_ids"])
			if len(allTargets) == 0 {
				return true, fmt.Errorf("无可选伤害目标")
			}
			ctxData["choice_type"] = "priest_divine_domain_damage_target"
			ctxData["target_ids"] = allTargets
			intr := rt.GetPendingInterrupt()
			if intr != nil {
				intr.Context = ctxData
			}
			rt.NotifyInterruptPrompt()
			return true, nil
		case "heal":
			allyTargets := runtimeutil.ParseStringSliceContextValue(ctxData["ally_target_ids"])
			if len(allyTargets) == 0 {
				return true, fmt.Errorf("无可选队友目标")
			}
			ctxData["choice_type"] = "priest_divine_domain_heal_target"
			ctxData["target_ids"] = allyTargets
			intr := rt.GetPendingInterrupt()
			if intr != nil {
				intr.Context = ctxData
			}
			rt.NotifyInterruptPrompt()
			return true, nil
		default:
			return true, fmt.Errorf("无效的神圣领域分支")
		}

	case "priest_divine_domain_damage_target", "priest_divine_domain_heal_target":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := rt.GetPlayers()[targetID]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}

		if choiceType == "priest_divine_domain_damage_target" {
			if user.Heal <= 0 {
				return true, fmt.Errorf("神圣领域分支①需要至少1点治疗")
			}
			user.Heal--
			rt.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: 2, DamageType: model.MagicAttack})
			rt.Log(fmt.Sprintf("%s 的 [神圣领域] 分支①生效：移除1点治疗，对 %s 造成2点法术伤害", user.Name, target.Name))
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil {
				rt.RoutePendingDamageOr(model.TurnStageExtraAction, func() {
					rt.EnterExtraActionStage()
				})
			}
			return true, nil
		}

		if target.Camp != user.Camp || target.ID == user.ID {
			return true, fmt.Errorf("神圣领域分支②目标必须是其他队友")
		}
		rt.Heal(user.ID, 2)
		rt.Heal(targetID, 1)
		rt.Log(fmt.Sprintf("%s 的 [神圣领域] 分支②生效：自身+2治疗，%s +1治疗", user.Name, target.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterExtraActionStage()
		}
		return true, nil
	}

	return false, nil
}

// Helper functions for priest

func mustChoiceResumePointFromMap(data map[string]interface{}, key string) interface{} {
	if data == nil {
		return nil
	}
	return data[key]
}
