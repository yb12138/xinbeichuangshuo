// gameflow: 冒险家角色选择流（声明式 ChoiceSpec 重构）。

package adventurer

import (
	"fmt"
	"strconv"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ChoiceSpecs 声明式选择流程条目。
func ChoiceSpecs() []engineplayer.ChoiceSpec {
	return []engineplayer.ChoiceSpec{
		{ChoiceType: "adventurer_fraud_pick", BuildPrompt: buildFraudPickPrompt, HandleChoice: handleFraudPick},
		{ChoiceType: "adventurer_fraud_attack_element", BuildPrompt: buildFraudElementPrompt, HandleChoice: handleFraudElement},
		{ChoiceType: "adventurer_paradise_pick", BuildPrompt: buildParadisePickPrompt, HandleChoice: handleParadisePick},
		{ChoiceType: "adventurer_paradise_target", BuildPrompt: buildParadiseTargetPrompt, HandleChoice: handleParadiseTarget},
		{ChoiceType: "adventurer_steal_sky_mode", BuildPrompt: buildStealSkyModePrompt, HandleChoice: handleStealSkyMode},
	}
}

// NewChoiceHandler 创建声明式处理器。
func NewChoiceHandler() engineplayer.ChoiceHandler {
	return engineplayer.NewSpecChoiceHandler(ChoiceSpecs())
}

// ---------------------------------------------------------------------------
// 偷天换日 - 简单两步骤流程
// ---------------------------------------------------------------------------

func buildStealSkyModePrompt(_ engineplayer.ChoiceRuntime, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	return engineplayer.NewPrompt(playerID, "【偷天换日】请选择效果：").
		OptionsFromLabels("转移对方战绩区1红宝石到我方", "将我方战绩区全部蓝水晶转换成红宝石").
		Build()
}

func handleStealSkyMode(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		enemyCamp, _ := ctxData["enemy_camp"].(string)
		selfCamp, _ := ctxData["self_camp"].(string)
		if rt.GetCampGems(enemyCamp) <= 0 {
			rt.Log(fmt.Sprintf("%s 的 [偷天换日] 失败：对方战绩区无红宝石", user.Name))
			rt.PopInterrupt()
			return true, nil
		}
		rt.ModifyGem(enemyCamp, -1)
		rt.ModifyGem(selfCamp, 1)
		rt.Log(fmt.Sprintf("%s 的 [偷天换日]：转移对方1红宝石到我方", user.Name))
	case 1:
		selfCamp, _ := ctxData["self_camp"].(string)
		crystals := rt.GetCampCrystals(selfCamp)
		if crystals <= 0 {
			rt.Log(fmt.Sprintf("%s 的 [偷天换日] 失败：我方战绩区无蓝水晶", user.Name))
			rt.PopInterrupt()
			return true, nil
		}
		rt.ModifyCrystal(selfCamp, -crystals)
		rt.ModifyGem(selfCamp, crystals)
		rt.Log(fmt.Sprintf("%s 的 [偷天换日]：将%d蓝水晶转换为红宝石", user.Name, crystals))
	default:
		return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	model.AppendExtraAction(user, "偷天换日", "")
	rt.PopInterrupt()
	rt.EnterExtraActionStage()
	return true, nil
}

// ---------------------------------------------------------------------------
// 冒险者天堂 - 转移资源
// ---------------------------------------------------------------------------

func buildParadisePickPrompt(_ engineplayer.ChoiceRuntime, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	labels := make([]string, 5)
	for i := 0; i <= 4; i++ {
		labels[i] = fmt.Sprintf("转移%d个提炼成果给队友", i)
	}
	return engineplayer.NewPrompt(playerID, "【冒险者天堂】选择要转移的提炼成果数量：").
		OptionsFromLabels(labels...).
		Build()
}

func handleParadisePick(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	ctxData["paradise_transfer_count"] = selectionIndex
	ctxData["choice_type"] = "adventurer_paradise_target"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return true, nil
}

func buildParadiseTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	eligibleIDs := runtimeutil.ParseStringSliceContextValue(data["eligible_targets"])
	if len(eligibleIDs) == 0 {
		eligibleIDs = runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
	}
	if len(eligibleIDs) == 0 {
		return nil
	}
	opts := make([]engineplayer.PromptOptionSpec, 0, len(eligibleIDs))
	for _, targetID := range eligibleIDs {
		target := rt.GetPlayers()[targetID]
		if target != nil {
			opts = append(opts, engineplayer.Option(targetID, target.Name))
		}
	}
	return engineplayer.NewPrompt(playerID, "【冒险者天堂】选择转移目标：").Options(opts...).Build()
}

func handleParadiseTarget(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	eligibleIDs := runtimeutil.ParseStringSliceContextValue(ctxData["eligible_targets"])
	if len(eligibleIDs) == 0 {
		eligibleIDs = runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	}
	if selectionIndex < 0 || selectionIndex >= len(eligibleIDs) {
		return true, fmt.Errorf("无效的目标索引")
	}
	targetID := eligibleIDs[selectionIndex]
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return true, fmt.Errorf("目标不存在")
	}
	transferCount := runtimeutil.ToIntContextValue(ctxData["paradise_transfer_count"])
	transferTotal := runtimeutil.ToIntContextValue(ctxData["transfer_total"])
	if transferTotal <= 0 {
		transferTotal = transferCount
	}
	fromPending, _ := ctxData["from_pending"].(bool)
	transferGem := runtimeutil.ToIntContextValue(ctxData["transfer_gem"])
	transferCrystal := runtimeutil.ToIntContextValue(ctxData["transfer_crystal"])
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]

	if fromPending {
		target.Gem += transferGem
		target.Crystal += transferCrystal
	} else if user != nil {
		if transferGem > 0 && user.Gem >= transferGem {
			user.Gem -= transferGem
			target.Gem += transferGem
		}
		if transferCrystal > 0 && user.Crystal >= transferCrystal {
			user.Crystal -= transferCrystal
			target.Crystal += transferCrystal
		}
	}
	if user != nil {
		if user.Gem > 0 {
			user.Gem--
		} else if user.Crystal > 0 {
			user.Crystal--
		}
	}
	rt.Log(fmt.Sprintf("[冒险者天堂] 转移 %d 个提炼成果给 %s", transferTotal, target.Name))
	rt.PopInterrupt()
	return true, nil
}

// ---------------------------------------------------------------------------
// 欺诈 - 选手牌（2张→五系选择 / 3张→暗灭）
// ---------------------------------------------------------------------------

func buildFraudPickPrompt(_ engineplayer.ChoiceRuntime, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	if player == nil {
		return nil
	}
	remainingIndices := runtimeutil.ParseChoiceIntSlice(data["remaining_indices"])
	if len(remainingIndices) == 0 {
		remainingIndices = allHandIndices(player)
	}
	selectedIndices := runtimeutil.ParseChoiceIntSlice(data["selected_indices"])
	needCount := runtimeutil.ToIntContextValue(data["need_count"])
	if needCount <= 0 {
		needCount = 2
	}
	remaining := needCount - len(selectedIndices)
	if remaining <= 0 {
		return nil
	}
	opts := make([]engineplayer.PromptOptionSpec, 0, len(remainingIndices))
	for _, idx := range remainingIndices {
		if idx >= 0 && idx < len(player.Hand) {
			opts = append(opts, engineplayer.Option(strconv.Itoa(idx), promptfmt.FormatCardInfo(player.Hand[idx])))
		}
	}
	return engineplayer.NewPrompt(playerID, fmt.Sprintf("【欺诈】请选择手牌（还需选择%d张同系牌）：", remaining)).
		Options(opts...).Build()
}

func handleFraudPick(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	selectedIndices := runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])
	remainingIndices := runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	if len(remainingIndices) == 0 {
		remainingIndices = allHandIndices(user)
	}

	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remainingIndices)
	if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
		return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	for _, idx := range selectedIndices {
		if idx == cardIdx {
			return true, fmt.Errorf("这张牌已经被选择过")
		}
	}

	selectedIndices = append(selectedIndices, cardIdx)
	remainingIndices = removeInt(remainingIndices, cardIdx)
	ctxData["selected_indices"] = selectedIndices
	ctxData["remaining_indices"] = remainingIndices

	needCount := runtimeutil.ToIntContextValue(ctxData["need_count"])
	if needCount <= 0 {
		needCount = 2
	}

	// 还没选够，继续选
	if len(selectedIndices) < needCount {
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil
	}

	// 选够后：检查是否还有顺序选牌未完成（多选批次中间步骤）
	seqRemaining := runtimeutil.ToIntContextValue(ctxData["sequential_remaining"])
	if seqRemaining > 0 {
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil
	}

	// 判断元素同系
	cards := collectCards(user, selectedIndices)
	commonElement := findCommonElement(cards)

	if len(selectedIndices) >= 3 {
		// 3张 → 暗灭（不要求同系）
		return resolveFraudAttack(rt, user, selectedIndices, ctxData, model.ElementDark)
	}
	if commonElement == "" {
		// 2张不同系，重新选
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil
	}
	// 2张同系 → 进入五系选择
	ctxData["choice_type"] = "adventurer_fraud_attack_element"
	ctxData["selected_element"] = string(commonElement)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return true, nil
}

func buildFraudElementPrompt(_ engineplayer.ChoiceRuntime, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	return engineplayer.NewPrompt(playerID, "【欺诈】请选择攻击系别（不含光/暗）：").
		Options(
			engineplayer.Option(string(model.ElementWater), "水"),
			engineplayer.Option(string(model.ElementFire), "火"),
			engineplayer.Option(string(model.ElementEarth), "地"),
			engineplayer.Option(string(model.ElementWind), "风"),
			engineplayer.Option(string(model.ElementThunder), "雷"),
		).
		Build()
}

func handleFraudElement(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	elements := []model.Element{model.ElementWater, model.ElementFire, model.ElementEarth, model.ElementWind, model.ElementThunder}
	if selectionIndex < 0 || selectionIndex >= len(elements) {
		return true, fmt.Errorf("无效的元素索引: %d", selectionIndex)
	}
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	selectedIndices := runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])
	return resolveFraudAttack(rt, user, selectedIndices, ctxData, elements[selectionIndex])
}

// resolveFraudAttack 统一处理欺诈攻击结算：弃牌、构建虚拟攻击、入队。
func resolveFraudAttack(rt engineplayer.ChoiceRuntime, user *model.Player, indices []int, ctxData map[string]interface{}, element model.Element) (bool, error) {
	targetID, _ := ctxData["fraud_target_id"].(string)
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return true, fmt.Errorf("欺诈目标不存在")
	}
	removed := removeCardsByIndicesFromHand(user, indices)
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)
	virtualCard := model.Card{
		ID:          fmt.Sprintf("fraud_%s_%d", user.ID, len(removed)),
		Name:        "欺诈",
		Type:        model.CardTypeAttack,
		Element:     element,
		Damage:      2,
		Description: "由欺诈视为的攻击",
	}
	rt.EnqueueVirtualAttack(user.ID, target.ID, virtualCard, "adventurer_fraud")
	if element == model.ElementDark {
		rt.Log(fmt.Sprintf("%s 发动 [欺诈]：选%d张牌自动转化为暗系攻击", user.Name, len(removed)))
	} else {
		rt.Log(fmt.Sprintf("%s 发动 [欺诈]：对 %s 发起%s系攻击", user.Name, target.Name, element))
	}
	rt.PopInterrupt()
	rt.EnterActionExecutionStage()
	return true, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func allHandIndices(player *model.Player) []int {
	indices := make([]int, 0, len(player.Hand))
	for i := range player.Hand {
		indices = append(indices, i)
	}
	return indices
}

func collectCards(player *model.Player, indices []int) []model.Card {
	cards := make([]model.Card, 0, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < len(player.Hand) {
			cards = append(cards, player.Hand[idx])
		}
	}
	return cards
}

func findCommonElement(cards []model.Card) model.Element {
	counts := map[model.Element]int{}
	for _, c := range cards {
		if c.Element != "" {
			counts[c.Element]++
		}
	}
	for ele, cnt := range counts {
		if cnt >= 2 {
			return ele
		}
	}
	return ""
}

func removeInt(slice []int, val int) []int {
	out := make([]int, 0, len(slice)-1)
	for _, v := range slice {
		if v != val {
			out = append(out, v)
		}
	}
	return out
}

func removeCardsByIndicesFromHand(player *model.Player, indices []int) []model.Card {
	if len(indices) == 0 || player == nil {
		return nil
	}
	sorted := make([]int, len(indices))
	copy(sorted, indices)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] > sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	removed := make([]model.Card, 0, len(sorted))
	for _, removeIdx := range sorted {
		if removeIdx >= 0 && removeIdx < len(player.Hand) {
			removed = append(removed, player.Hand[removeIdx])
			player.Hand = append(player.Hand[:removeIdx], player.Hand[removeIdx+1:]...)
		}
	}
	return removed
}
