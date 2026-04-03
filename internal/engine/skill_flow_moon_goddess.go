package engine

import (
	"fmt"
	"starcup-engine/internal/engine/runtimeutil"
	"strings"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildMoonGoddessChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
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
		for _, idx := range indices {
			if player == nil || idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
				continue
			}
			fc := player.Field[idx]
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("移除闇月[%s/%s/%s]", fc.Card.Name, fc.Card.Type, fc.Card.Element),
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
		}

	case "mg_medusa_magic_discard":
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:       model.PromptChooseCards,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【美杜莎之眼】因移除了法术闇月，请弃1张手牌：",
			Options:    options,
			Min:        1,
			Max:        1,
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
		var options []model.PromptOption
		for _, mode := range modes {
			switch mode {
			case "branch1":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支①：移除1个闇月，令目标角色+1治疗"})
			case "branch2":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支②：移除1点治疗，你+1新月"})
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
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【月之轮回】请选择获得1点治疗的角色：",
			Options:    options,
			Min:        1,
			Max:        1,
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
		options := []model.PromptOption{{ID: "0", Label: "跳过月渎"}}
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{
					ID:    fmt.Sprintf("%d", len(options)),
					Label: fmt.Sprintf("对 %s 造成1点法术伤害", p.Name),
				})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【月渎】请选择是否对当前受伤目标追加1点法术伤害：",
			Options:    options,
			Min:        1,
			Max:        1,
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
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【闇月斩】请选择X值：",
			Options:    options,
			Min:        1,
			Max:        1,
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
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【苍白之月】请选择分支：",
			Options:    options,
			Min:        1,
			Max:        1,
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
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【苍白之月】分支②请选择X值：",
			Options:    options,
			Min:        1,
			Max:        1,
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
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【苍白之月】分支②请选择目标对手：",
			Options:    options,
			Min:        1,
			Max:        1,
		}

	case "mg_pale_moon_discard":
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
			})
		}
		return &model.Prompt{
			Type:       model.PromptChooseCards,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【苍白之月】分支②请弃1张牌：",
			Options:    options,
			Min:        1,
			Max:        1,
		}
	}

	return nil
}

func (e *GameEngine) moonGoddessFindPendingAttackDamage(rawCtx *model.Context) *model.PendingDamage {
	if rawCtx == nil || rawCtx.TriggerCtx == nil {
		return nil
	}
	for i := range e.State.PendingDamageQueue {
		pd := &e.State.PendingDamageQueue[i]
		if !strings.EqualFold(pd.DamageType, "Attack") {
			continue
		}
		if pd.SourceID != rawCtx.TriggerCtx.SourceID || pd.TargetID != rawCtx.TriggerCtx.TargetID {
			continue
		}
		return pd
	}
	return nil
}

func (e *GameEngine) finishMoonGoddessMedusa(rawCtx *model.Context) {
	e.PopInterrupt()
	if e.State.PendingInterrupt != nil {
		return
	}
	defaultReturn := interface{}(nil)
	if rawCtx != nil && rawCtx.Trigger == model.TriggerOnAttackStart {
		defaultReturn = model.TurnStageActionExecution
	}
	if e.routePendingDamageWithDefaultReturn(defaultReturn) {
		return
	}
	if rawCtx != nil && rawCtx.Trigger == model.TriggerOnAttackStart {
		if len(e.State.ActionQueue) > 0 {
			e.enterActionExecutionStage()
		} else {
			e.enterResponseWindow()
		}
		return
	}
	e.enterResponseWindow()
}

func (e *GameEngine) queueMoonGoddessMedusaMagicDamage(user *model.Player, attackerID string) {
	if user == nil || attackerID == "" {
		return
	}
	attacker := e.State.Players[attackerID]
	if attacker == nil {
		return
	}
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   attacker.ID,
		Damage:     1,
		DamageType: "magic",
	})
	e.Log(fmt.Sprintf("%s 的 [美杜莎之眼] 额外效果：对 %s 造成1点法术伤害", user.Name, attacker.Name))
}

func (e *GameEngine) handleMoonGoddessChoiceInput(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "mg_medusa_darkmoon_pick":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
		card, ok := e.removeMoonGoddessDarkMoonByFieldIndex(user, fieldIdx)
		if !ok {
			return true, fmt.Errorf("请选择可用的闇月")
		}
		e.Heal(user.ID, 1)
		nowPetrify := addMoonGoddessPetrify(user, 1)
		e.Log(fmt.Sprintf("%s 发动 [美杜莎之眼]：移除%s系闇月，治疗+1，石化+1（当前%d）",
			user.Name, card.Element, nowPetrify))

		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		attackerID, _ := ctxData["attacker_id"].(string)
		if card.Type == model.CardTypeMagic {
			if len(user.Hand) > 0 {
				ctxData["choice_type"] = "mg_medusa_magic_discard"
				e.State.PendingInterrupt.Context = ctxData
				e.notifyInterruptPrompt()
				return true, nil
			}
			e.queueMoonGoddessMedusaMagicDamage(user, attackerID)
		}
		e.finishMoonGoddessMedusa(rawCtx)
		return true, nil

	case "mg_medusa_magic_discard":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		cardIdx := selectionIndex
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的弃牌索引: %d", selectionIndex)
		}
		card := user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.Log(fmt.Sprintf("%s 的 [美杜莎之眼] 额外效果：弃置1张手牌", user.Name))
		attackerID, _ := ctxData["attacker_id"].(string)
		e.queueMoonGoddessMedusaMagicDamage(user, attackerID)

		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		e.finishMoonGoddessMedusa(rawCtx)
		return true, nil

	case "mg_moon_cycle_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
			if moonGoddessDarkMoonCount(user) <= 0 {
				return true, fmt.Errorf("闇月不足，无法发动分支①")
			}
			ctxData["choice_type"] = "mg_moon_cycle_heal_target"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		case "branch2":
			if user.Heal <= 0 {
				return true, fmt.Errorf("治疗不足，无法发动分支②")
			}
			user.Heal--
			now := addMoonGoddessNewMoon(user, 1)
			e.Log(fmt.Sprintf("%s 发动 [月之轮回] 分支②：移除1治疗，+1新月（当前%d）", user.Name, now))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.enterTurnEndStage()
			}
			return true, nil
		default:
			return true, fmt.Errorf("无效分支")
		}

	case "mg_moon_cycle_heal_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("目标角色不存在")
		}
		e.removeMoonGoddessDarkMoonAny(user, 1)
		e.Heal(target.ID, 1)
		e.Log(fmt.Sprintf("%s 发动 [月之轮回] 分支①：移除1闇月并令 %s +1治疗", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterTurnEndStage()
		}
		return true, nil

	case "mg_blasphemy_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
		if selectionIndex == 0 {
			user.TurnState.SkillFlowState["mg_blasphemy_pending"] = 0
			e.Log(fmt.Sprintf("%s 选择跳过 [月渎]", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.enterDamageResolution(nil)
			}
			return true, nil
		}
		choice := selectionIndex - 1
		if choice < 0 || choice >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[choice]]
		if target == nil {
			return true, fmt.Errorf("目标角色不存在")
		}
		if user.Heal <= 0 {
			return true, fmt.Errorf("治疗不足，无法发动月渎")
		}
		user.Heal--
		user.TurnState.SkillFlowState["mg_blasphemy_pending"] = 0
		user.TurnState.UsedSkillCounts["mg_blasphemy"] = 1
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     1,
			DamageType: "magic",
		})
		e.Log(fmt.Sprintf("%s 发动 [月渎]：移除1治疗，对 %s 造成1点法术伤害", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "mg_darkmoon_slash_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
		pd := e.moonGoddessFindPendingAttackDamage(rawCtx)
		if pd == nil {
			return true, fmt.Errorf("未找到对应的攻击伤害结算")
		}
		e.removeMoonGoddessDarkMoonAny(user, x)
		pd.Damage += x
		e.Log(fmt.Sprintf("%s 的 [闇月斩] 生效：移除%d个闇月，本次攻击伤害额外+%d", user.Name, x, x))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if rawCtx != nil && rawCtx.Trigger == model.TriggerOnAttackHit {
				e.markPendingAttackDamageHitProcessed(rawCtx)
			}
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "mg_pale_moon_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
			user.TurnState.UsedSkillCounts["mg_next_attack_no_counter"]++
			ensurePlayerTokensMap(user)
			user.Tokens["mg_extra_turn_pending"]++
			model.AppendAttackAction(user, "苍白之月")
			e.Log(fmt.Sprintf("%s 发动 [苍白之月] 分支①：移除3石化，下次主动攻击不可应战，额外+1攻击行动并获得额外回合", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.enterExtraActionStage()
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
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		default:
			return true, fmt.Errorf("无效分支")
		}

	case "mg_pale_moon_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
		targetIDs := e.moonGoddessEnemyIDs(user)
		if len(targetIDs) == 0 {
			return true, fmt.Errorf("没有可选对手")
		}
		ctxData["x"] = selectionIndex + 1
		ctxData["target_ids"] = targetIDs
		ctxData["choice_type"] = "mg_pale_moon_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
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
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "mg_pale_moon_discard":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		cardIdx := selectionIndex
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的弃牌索引: %d", selectionIndex)
		}
		targetID, _ := ctxData["target_id"].(string)
		target := e.State.Players[targetID]
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
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		addMoonGoddessNewMoon(user, -x)
		nowPetrify := addMoonGoddessPetrify(user, 1)
		damage := x + 1
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     damage,
			DamageType: "magic",
		})
		e.Log(fmt.Sprintf("%s 发动 [苍白之月] 分支②：移除%d新月，石化+1（当前%d），弃1张牌并对 %s 造成%d点法术伤害",
			user.Name, x, nowPetrify, target.Name, damage))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setReturnPoint(model.TurnStageExtraAction)
			e.enterDamageResolution(nil)
		}
		return true, nil
	}

	return false, nil
}
