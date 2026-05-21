// gameflow: 月女神角色选择流。

package moon

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "mg_medusa_darkmoon_pick":
		var indices []int
		if arr, ok := data["darkmoon_indices"].([]int); ok {
			indices = append(indices, arr...)
		} else if arr, ok := data["darkmoon_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					indices = append(indices, int(f))
				}
			}
		}
		var options []model.PromptOption
		for optionIdx, idx := range indices {
			if player == nil || idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
				continue
			}
			fc := player.Field[idx]
			fieldIndex := idx
			options = append(options, model.PromptOption{
				ID:         fmt.Sprintf("%d", optionIdx),
				Label:      fmt.Sprintf("移除闇月[%s/%s/%s]", fc.Card.Name, fc.Card.Type, fc.Card.Element),
				FieldIndex: &fieldIndex,
			})
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【美杜莎之眼】请选择要展示并移除的同系闇月：",
			Options:    options,
			Min:        1,
			Max:        1,
			Presentation: &model.PromptPresentation{
				Kind:       model.PresentationCardPicker,
				Layout:     "field_cover",
				CardSource: "field",
				CardFilter: "effect:MoonDarkMoon",
			},
		}

	case "mg_medusa_magic_discard":
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{
				ID:     fmt.Sprintf("%d", idx),
				Label:  fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(c)),
				CardID: c.ID,
			})
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【美杜莎之眼】因移除了法术闇月，请弃1张手牌：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}

	case "mg_moon_cycle_mode":
		var modes []string
		if arr, ok := data["modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := data["modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		options := []model.PromptOption{{ID: "decline", Label: "不发动"}}
		for _, mode := range modes {
			switch mode {
			case "branch1":
				options = append(options, model.PromptOption{ID: "branch1", Label: "分支①：移除1个闇月，令目标角色+1治疗"})
			case "branch2":
				options = append(options, model.PromptOption{ID: "branch2", Label: "分支②：移除1点治疗，你+1新月"})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【月之轮回】请选择发动分支：",
			Options:    options,
			Min:        1,
			Max:        1,
			Presentation: &model.PromptPresentation{
				Kind:   model.PresentationBranchSelect,
				Layout: "overlay",
			},
		}

	case "mg_moon_cycle_heal_target":
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := rt.GetPlayers()[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name, TargetID: tid})
			}
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【月之轮回】请选择获得1点治疗的角色：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}

	case "mg_blasphemy_target":
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		options := make([]model.PromptOption, 0, len(targetIDs)+1)
		for _, tid := range targetIDs {
			if p := rt.GetPlayers()[tid]; p != nil {
				options = append(options, model.PromptOption{
					ID:    fmt.Sprintf("%d", len(options)),
					Label: fmt.Sprintf("对 %s 造成1点法术伤害", p.Name),
				})
			}
		}
		options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "不发动", ButtonLabel: "不发动"})
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【月渎】请选择目标或不发动：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay", HasDecline: true, DeclineIndex: len(options) - 1},
		}

	case "mg_darkmoon_slash_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			return nil
		}
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("移除%d个闇月，本次攻击伤害额外+%d", x, x),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【闇月斩】请选择X值：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
		}

	case "mg_pale_moon_mode":
		var modes []string
		if arr, ok := data["modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := data["modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		var options []model.PromptOption
		for _, mode := range modes {
			switch mode {
			case "branch1":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支①：移除3石化，强化下次主动攻击并获得额外回合"})
			case "branch2":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支②：移除X新月，弃1张牌并造成(X+1)法术伤害"})
			}
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【苍白之月】请选择分支：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}

	case "mg_pale_moon_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			return nil
		}
		var options []model.PromptOption
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d（目标法术伤害=%d）", x, x+1),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【苍白之月】分支②请选择X值：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
		}

	case "mg_pale_moon_target":
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := rt.GetPlayers()[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name, TargetID: tid})
			}
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【苍白之月】分支②请选择目标对手：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}

	case "mg_pale_moon_discard":
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{
				ID:     fmt.Sprintf("%d", idx),
				Label:  fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(c)),
				CardID: c.ID,
			})
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【苍白之月】分支②请弃1张牌：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}
	}

	return nil
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, _ string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "mg_moon_cycle_mode" {
		return false, nil
	}
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return false, fmt.Errorf("玩家不存在")
	}
	return finishMoonCycleDecline(rt, user), nil
}

func finishMoonCycleDecline(rt engineplayer.ChoiceRuntime, user *model.Player) bool {
	if user != nil {
		rt.Log(fmt.Sprintf("%s 选择不发动 [月之轮回]", user.Name))
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterTurnEndStage()
	}
	return true
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "mg_medusa_darkmoon_pick":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		var indices []int
		if arr, ok := ctxData["darkmoon_indices"].([]int); ok {
			indices = append(indices, arr...)
		} else if arr, ok := ctxData["darkmoon_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					indices = append(indices, int(f))
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(indices) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		fieldIdx := indices[selectionIndex]
		card, ok := RemoveDarkMoonByFieldIndex(rt, user, fieldIdx)
		if !ok {
			return true, fmt.Errorf("请选择可用的闇月")
		}
		rt.Heal(user.ID, 1)
		nowPetrify := addMoonGoddessPetrify(user, 1)
		rt.Log(fmt.Sprintf("%s 发动 [美杜莎之眼]：移除%s系闇月，治疗+1，石化+1（当前%d）",
			user.Name, card.Element, nowPetrify))

		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		attackerID, _ := ctxData["attacker_id"].(string)
		if card.Type == model.CardTypeMagic {
			if len(user.Hand) > 0 {
				ctxData["choice_type"] = "mg_medusa_magic_discard"
				intr := rt.GetPendingInterrupt()
				if intr != nil {
					intr.Context = ctxData
				}
				rt.NotifyInterruptPrompt()
				return true, nil
			}
			queueMoonGoddessMedusaMagicDamage(rt, user, attackerID)
		}
		finishMoonGoddessMedusa(rt, rawCtx)
		return true, nil

	case "mg_medusa_magic_discard":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		cardIdx := selectionIndex
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的弃牌索引: %d", selectionIndex)
		}
		card := user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		rt.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		rt.AppendToDiscard([]model.Card{card})
		rt.Log(fmt.Sprintf("%s 的 [美杜莎之眼] 额外效果：弃置1张手牌", user.Name))
		attackerID, _ := ctxData["attacker_id"].(string)
		queueMoonGoddessMedusaMagicDamage(rt, user, attackerID)

		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		finishMoonGoddessMedusa(rt, rawCtx)
		return true, nil

	case "mg_moon_cycle_mode":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			return finishMoonCycleDecline(rt, user), nil
		}
		var modes []string
		if arr, ok := ctxData["modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := ctxData["modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		modeIdx := selectionIndex - 1
		if modeIdx < 0 || modeIdx >= len(modes) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[modeIdx]
		switch mode {
		case "branch1":
			if moonGoddessDarkMoonCount(user) <= 0 {
				return true, fmt.Errorf("闇月不足，无法发动分支①")
			}
			ctxData["choice_type"] = "mg_moon_cycle_heal_target"
			intr := rt.GetPendingInterrupt()
			if intr != nil {
				intr.Context = ctxData
			}
			rt.NotifyInterruptPrompt()
			return true, nil
		case "branch2":
			if user.Heal <= 0 {
				return true, fmt.Errorf("治疗不足，无法发动分支②")
			}
			user.Heal--
			now := addMoonGoddessNewMoon(user, 1)
			rt.Log(fmt.Sprintf("%s 发动 [月之轮回] 分支②：移除1治疗，+1新月（当前%d）", user.Name, now))
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil {
				rt.EnterTurnEndStage()
			}
			return true, nil
		default:
			return true, fmt.Errorf("无效分支")
		}

	case "mg_moon_cycle_heal_target":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if moonGoddessDarkMoonCount(user) <= 0 {
			return true, fmt.Errorf("闇月不足，无法发动分支①")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := rt.GetPlayers()[targetIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("目标角色不存在")
		}
		RemoveDarkMoonAny(rt, user, 1)
		rt.Heal(target.ID, 1)
		rt.Log(fmt.Sprintf("%s 发动 [月之轮回] 分支①：移除1闇月并令 %s +1治疗", user.Name, target.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterTurnEndStage()
		}
		return true, nil

	case "mg_blasphemy_target":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex == len(targetIDs) {
			if user.TurnState.SkillFlowState == nil {
				user.TurnState.SkillFlowState = map[string]int{}
			}
			user.TurnState.SkillFlowState["mg_blasphemy_pending"] = 0
			rt.Log(fmt.Sprintf("%s 选择跳过 [月渎]", user.Name))
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil {
				rt.EnterDamageResolution(nil)
			}
			return true, nil
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := rt.GetPlayers()[targetIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("目标角色不存在")
		}
		if user.Heal <= 0 {
			return true, fmt.Errorf("治疗不足，无法发动月渎")
		}
		user.Heal--
		if user.TurnState.SkillFlowState == nil {
			user.TurnState.SkillFlowState = map[string]int{}
		}
		user.TurnState.SkillFlowState["mg_blasphemy_pending"] = 0
		if user.TurnState.UsedSkillCounts == nil {
			user.TurnState.UsedSkillCounts = map[string]int{}
		}
		user.TurnState.UsedSkillCounts["mg_blasphemy"] = 1
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     1,
			DamageType: model.MagicAttack,
		})
		rt.Log(fmt.Sprintf("%s 发动 [月渎]：移除1治疗，对 %s 造成1点法术伤害", user.Name, target.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterDamageResolution(nil)
		}
		return true, nil

	case "mg_darkmoon_slash_x":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if maxX < 1 {
			return true, fmt.Errorf("闇月斩没有可选X值")
		}
		if selectionIndex < 0 || selectionIndex >= maxX {
			return true, fmt.Errorf("无效的X值")
		}
		x := selectionIndex + 1
		if x > moonGoddessDarkMoonCount(user) {
			return true, fmt.Errorf("闇月不足，无法移除%d个", x)
		}
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		pd := moonGoddessFindPendingAttackDamage(rt, rawCtx)
		if pd == nil {
			return true, fmt.Errorf("未找到对应的攻击伤害结算")
		}
		RemoveDarkMoonAny(rt, user, x)
		pd.Damage += x
		rt.Log(fmt.Sprintf("%s 的 [闇月斩] 生效：移除%d个闇月，本次攻击伤害额外+%d", user.Name, x, x))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			if rawCtx != nil && rawCtx.ResumeAttackHitPhase() {
				rt.ResumePendingAttackHit(ctxData)
			}
			rt.EnterDamageResolution(nil)
		}
		return true, nil

	case "mg_pale_moon_mode":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		var modes []string
		if arr, ok := ctxData["modes"].([]string); ok {
			modes = append(modes, arr...)
		} else if arr, ok := ctxData["modes"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modes = append(modes, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		switch mode {
		case "branch1":
			if moonGoddessPetrify(user) < 3 {
				return true, fmt.Errorf("石化不足3点，无法发动分支①")
			}
			addMoonGoddessPetrify(user, -3)
			if user.TurnState.UsedSkillCounts == nil {
				user.TurnState.UsedSkillCounts = map[string]int{}
			}
			user.TurnState.UsedSkillCounts["mg_next_attack_no_counter"]++
			engineplayer.EnsurePlayerSkillFlowState(user)
			user.TurnState.SkillFlowState["mg_extra_turn_pending"]++
			model.AppendAttackAction(user, "苍白之月")
			rt.Log(fmt.Sprintf("%s 发动 [苍白之月] 分支①：移除3石化，下次主动攻击不可应战，额外+1攻击行动并获得额外回合", user.Name))
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil {
				rt.EnterExtraActionStage()
			}
			return true, nil
		case "branch2":
			if moonGoddessNewMoon(user) <= 0 {
				return true, fmt.Errorf("新月不足，无法发动分支②")
			}
			if len(user.Hand) <= 0 {
				return true, fmt.Errorf("手牌不足，无法发动分支②")
			}
			maxX := moonGoddessNewMoon(user)
			ctxData["choice_type"] = "mg_pale_moon_x"
			ctxData["max_x"] = maxX
			intr := rt.GetPendingInterrupt()
			if intr != nil {
				intr.Context = ctxData
			}
			rt.NotifyInterruptPrompt()
			return true, nil
		default:
			return true, fmt.Errorf("无效分支")
		}

	case "mg_pale_moon_x":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if maxX < 1 {
			return true, fmt.Errorf("没有可选的新月数量")
		}
		if selectionIndex < 0 || selectionIndex >= maxX {
			return true, fmt.Errorf("无效的X值")
		}
		targetIDs := moonGoddessEnemyIDs(rt, user)
		if len(targetIDs) == 0 {
			return true, fmt.Errorf("没有可选对手")
		}
		ctxData["x"] = selectionIndex + 1
		ctxData["target_ids"] = targetIDs
		ctxData["choice_type"] = "mg_pale_moon_target"
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "mg_pale_moon_target":
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["target_id"] = targetIDs[selectionIndex]
		ctxData["choice_type"] = "mg_pale_moon_discard"
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "mg_pale_moon_discard":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		cardIdx := selectionIndex
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的弃牌索引: %d", selectionIndex)
		}
		targetID, _ := ctxData["target_id"].(string)
		target := rt.GetPlayers()[targetID]
		if target == nil {
			return true, fmt.Errorf("目标角色不存在")
		}
		x := runtimeutil.ToIntContextValue(ctxData["x"])
		if x < 1 {
			return true, fmt.Errorf("苍白之月分支②的X至少为1")
		}
		if x > moonGoddessNewMoon(user) {
			return true, fmt.Errorf("新月不足，无法移除%d点", x)
		}
		card := user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		rt.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		rt.AppendToDiscard([]model.Card{card})
		addMoonGoddessNewMoon(user, -x)
		nowPetrify := addMoonGoddessPetrify(user, 1)
		damage := x + 1
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     damage,
			DamageType: model.MagicAttack,
		})
		rt.Log(fmt.Sprintf("%s 发动 [苍白之月] 分支②：移除%d新月，石化+1（当前%d），弃1张牌并对 %s 造成%d点法术伤害",
			user.Name, x, nowPetrify, target.Name, damage))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageExtraAction)
			rt.EnterDamageResolution(nil)
		}
		return true, nil
	}

	return false, nil
}

// Helper functions for moon

const (
	moonGoddessNewMoonCapEngine = 10
	moonGoddessPetrifyCapEngine = 10
)

func addTokenValueBounded(player *model.Player, key string, delta int, cap int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	current := player.Tokens[key]
	newVal := current + delta
	if newVal < 0 {
		newVal = 0
	}
	if cap > 0 && newVal > cap {
		newVal = cap
	}
	player.Tokens[key] = newVal
	return newVal
}

func addMoonGoddessNewMoon(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "mg_new_moon", delta, moonGoddessNewMoonCapEngine)
}

func moonGoddessPetrify(player *model.Player) int {
	return engineplayer.TokenValue(player, "mg_petrify", moonGoddessPetrifyCapEngine)
}

func addMoonGoddessPetrify(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "mg_petrify", delta, moonGoddessPetrifyCapEngine)
}

func moonGoddessNewMoon(player *model.Player) int {
	return engineplayer.TokenValue(player, "mg_new_moon", moonGoddessNewMoonCapEngine)
}

func moonGoddessDarkMoonCovers(player *model.Player) []*model.FieldCard {
	if player == nil {
		return nil
	}
	var out []*model.FieldCard
	for _, fc := range player.Field {
		if fc != nil && fc.Effect == model.EffectMoonDarkMoon {
			out = append(out, fc)
		}
	}
	return out
}

func moonGoddessDarkMoonCount(player *model.Player) int {
	count := len(moonGoddessDarkMoonCovers(player))
	if player != nil && count <= 0 {
		leaveMoonGoddessDarkMoonForm(player)
	}
	return count
}

func leaveMoonGoddessDarkMoonForm(player *model.Player) bool {
	return engineplayer.ClearForm(player, model.FormMoonGoddessDarkMoon)
}

func queueMoonGoddessMedusaMagicDamage(rt engineplayer.ChoiceRuntime, user *model.Player, attackerID string) {
	if user == nil || attackerID == "" {
		return
	}
	attacker := rt.GetPlayers()[attackerID]
	if attacker == nil {
		return
	}
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   attacker.ID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	rt.Log(fmt.Sprintf("%s 的 [美杜莎之眼] 额外效果：对 %s 造成1点法术伤害", user.Name, attacker.Name))
}

func finishMoonGoddessMedusa(rt engineplayer.ChoiceRuntime, rawCtx *model.Context) {
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() != nil {
		return
	}
	defaultReturn := interface{}(nil)
	if rawCtx != nil && rawCtx.AttackDeclaredPhase() {
		defaultReturn = model.TurnStageActionExecution
	}
	if rt.RoutePendingDamageWithReturn(defaultReturn) {
		return
	}
	if rawCtx != nil && rawCtx.AttackDeclaredPhase() {
		if len(rt.GetActionQueue()) > 0 {
			rt.EnterActionExecutionStage()
		} else {
			// enterResponseWindow is not exposed on ChoiceRuntime
		}
		return
	}
	// enterResponseWindow is not exposed on ChoiceRuntime
}

func moonGoddessFindPendingAttackDamage(rt engineplayer.ChoiceRuntime, _ *model.Context) *model.PendingDamage {
	queue := rt.GetPendingDamageQueue()
	for i := range queue {
		if queue[i].Damage > 0 {
			return &queue[i]
		}
	}
	return nil
}

var _ engineplayer.CancelChoiceHandler = choiceHandler{}

func moonGoddessEnemyIDs(rt engineplayer.ChoiceRuntime, user *model.Player) []string {
	if user == nil {
		return nil
	}
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		p := rt.GetPlayers()[pid]
		if p == nil || p.Camp == user.Camp {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}
