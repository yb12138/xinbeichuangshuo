// gameflow: 红莲骑士角色选择流。

package crimson_knight

import (
	"fmt"
	"strings"

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
	case "crk_bloody_prayer_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d（移除%d治疗并对自己造成%d法伤）", x, x, x)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【血腥祷言】请选择X值：", Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 1}}

	case "crk_bloody_prayer_ally_count":
		allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
		xValue := runtimeutil.ToIntContextValue(data["x_value"])
		if xValue <= 0 {
			xValue = 1
		}
		options := []model.PromptOption{{ID: "0", Label: "选择1名队友"}}
		if len(allyIDs) >= 2 && xValue >= 2 {
			options = append(options, model.PromptOption{ID: "1", Label: "选择2名队友（治疗将分配）"})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【血腥祷言】请选择要分配治疗的队友数量：", Options: options, Min: 1, Max: 1}

	case "crk_bloody_prayer_target":
		allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
		selectedSet := runtimeutil.IDsToSet(runtimeutil.ParseStringSliceContextValue(data["selected_ally_ids"]))
		allyCount := runtimeutil.ToIntContextValue(data["ally_count"])
		if allyCount <= 0 {
			allyCount = 1
		}
		options := make([]model.PromptOption, 0, len(allyIDs))
		for _, allyID := range allyIDs {
			if selectedSet[allyID] {
				continue
			}
			if target := rt.GetPlayers()[allyID]; target != nil {
				options = append(options, model.PromptOption{ID: allyID, Label: target.Name})
			}
		}
		pickIndex := len(selectedSet) + 1
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【血腥祷言】请选择第 %d/%d 名队友：", pickIndex, allyCount), Options: options, Min: 1, Max: 1}

	case "crk_bloody_prayer_split":
		selected := runtimeutil.ParseStringSliceContextValue(data["selected_ally_ids"])
		if len(selected) != 2 {
			return nil
		}
		xValue := runtimeutil.ToIntContextValue(data["x_value"])
		if xValue < 2 {
			return nil
		}
		first := rt.GetPlayers()[selected[0]]
		second := rt.GetPlayers()[selected[1]]
		if first == nil || second == nil {
			return nil
		}
		options := make([]model.PromptOption, 0, xValue-1)
		for firstHeal := 1; firstHeal < xValue; firstHeal++ {
			secondHeal := xValue - firstHeal
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", firstHeal-1), Label: fmt.Sprintf("%s +%d，%s +%d", first.Name, firstHeal, second.Name, secondHeal)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【血腥祷言】请选择治疗分配：", Options: options, Min: 1, Max: 1}
	}

	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "crk_bloody_prayer_x":
		return true, handleCrimsonKnightBloodyPrayerX(rt, selectionIndex, ctxData)

	case "crk_bloody_prayer_ally_count":
		return true, handleCrimsonKnightBloodyPrayerAllyCount(rt, selectionIndex, ctxData)

	case "crk_bloody_prayer_target":
		return true, handleCrimsonKnightBloodyPrayerTarget(rt, selectionIndex, ctxData)

	case "crk_bloody_prayer_split":
		return true, handleCrimsonKnightBloodyPrayerSplit(rt, selectionIndex, ctxData)
	}

	return false, nil
}

func handleCrimsonKnightBloodyPrayerX(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	xValue := selectionIndex + 1
	if xValue < 1 || xValue > maxX {
		return fmt.Errorf("无效的X值")
	}
	ctxData["x_value"] = xValue
	ctxData["selected_ally_ids"] = []string{}
	allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	if len(allyIDs) == 0 {
		return fmt.Errorf("没有可分配治疗的队友")
	}
	if len(allyIDs) >= 2 && xValue >= 2 {
		ctxData["choice_type"] = "crk_bloody_prayer_ally_count"
	} else {
		ctxData["ally_count"] = 1
		ctxData["choice_type"] = "crk_bloody_prayer_target"
	}
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleCrimsonKnightBloodyPrayerAllyCount(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	maxCount := 1
	if len(allyIDs) >= 2 && xValue >= 2 {
		maxCount = 2
	}
	allyCount := selectionIndex + 1
	if allyCount < 1 || allyCount > maxCount {
		return fmt.Errorf("无效的队友数量选择")
	}
	ctxData["ally_count"] = allyCount
	ctxData["selected_ally_ids"] = []string{}
	ctxData["choice_type"] = "crk_bloody_prayer_target"
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleCrimsonKnightBloodyPrayerTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	selected := dedupeNonEmptyIDs(runtimeutil.ParseStringSliceContextValue(ctxData["selected_ally_ids"]))
	selectedSet := runtimeutil.IDsToSet(selected)
	remaining := make([]string, 0, len(allyIDs))
	for _, allyID := range allyIDs {
		if !selectedSet[allyID] {
			remaining = append(remaining, allyID)
		}
	}
	if selectionIndex < 0 || selectionIndex >= len(remaining) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	allyCount := runtimeutil.ToIntContextValue(ctxData["ally_count"])
	if allyCount <= 0 {
		allyCount = 1
	}
	chosenID := remaining[selectionIndex]
	selected = append(selected, chosenID)
	ctxData["selected_ally_ids"] = selected

	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	if xValue <= 0 || user.Heal < xValue {
		return fmt.Errorf("治疗不足，无法结算血腥祷言")
	}
	if len(selected) < allyCount {
		ctxData["choice_type"] = "crk_bloody_prayer_target"
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}

	if allyCount <= 1 {
		if err := resolveCrimsonKnightBloodyPrayer(rt, user, xValue, map[string]int{selected[0]: xValue}); err != nil {
			return err
		}
	} else {
		if xValue < 2 {
			return fmt.Errorf("X不足以分配给2名队友")
		}
		ctxData["choice_type"] = "crk_bloody_prayer_split"
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func handleCrimsonKnightBloodyPrayerSplit(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	if xValue < 2 || user.Heal < xValue {
		return fmt.Errorf("治疗不足，无法结算血腥祷言")
	}
	selected := runtimeutil.ParseStringSliceContextValue(ctxData["selected_ally_ids"])
	if len(selected) != 2 {
		return fmt.Errorf("血腥祷言分配目标数量异常")
	}
	if selectionIndex < 0 || selectionIndex >= xValue-1 {
		return fmt.Errorf("无效的分配选项")
	}
	firstHeal := selectionIndex + 1
	secondHeal := xValue - firstHeal
	alloc := map[string]int{selected[0]: firstHeal, selected[1]: secondHeal}
	if err := resolveCrimsonKnightBloodyPrayer(rt, user, xValue, alloc); err != nil {
		return err
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func resolveCrimsonKnightBloodyPrayer(rt engineplayer.ChoiceRuntime, user *model.Player, x int, allocations map[string]int) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if x <= 0 {
		return fmt.Errorf("无效的X值")
	}
	if user.Heal < x {
		return fmt.Errorf("治疗不足，无法结算血腥祷言")
	}

	user.Heal -= x
	for _, pid := range rt.GetPlayerOrder() {
		amt := allocations[pid]
		if amt <= 0 {
			continue
		}
		rt.Heal(pid, amt)
	}
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:              user.ID,
		TargetID:              user.ID,
		Damage:                x,
		DamageType:            model.MagicAttack,
		AllowCrimsonFaithHeal: true,
	})
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	user.Tokens["crk_blood_mark"]++
	if user.Tokens["crk_blood_mark"] > 3 {
		user.Tokens["crk_blood_mark"] = 3
	}

	var parts []string
	for _, pid := range rt.GetPlayerOrder() {
		amt := allocations[pid]
		if amt <= 0 {
			continue
		}
		if p := rt.GetPlayers()[pid]; p != nil {
			parts = append(parts, fmt.Sprintf("%s +%d治疗", p.Name, amt))
		}
	}
	allocText := "未分配治疗"
	if len(parts) > 0 {
		allocText = strings.Join(parts, "，")
	}
	rt.Log(fmt.Sprintf("%s 发动 [血腥祷言]：移除%d治疗并自伤%d，%s，血印+1", user.Name, x, x, allocText))
	return nil
}

func dedupeNonEmptyIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}
