// gameflow: 冒险家角色选择流。

package adventurer

import (
	"fmt"
	"strconv"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

// ---------------------------------------------------------------------------
// BuildPrompt
// ---------------------------------------------------------------------------

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "adventurer_fraud_pick":
		return buildFraudPickPrompt(rt, playerID, player, data)
	case "adventurer_fraud_attack_element":
		return buildFraudAttackElementPrompt(playerID, data)
	case "adventurer_paradise_pick":
		return buildParadisePickPrompt(rt, playerID, player, data)
	case "adventurer_paradise_target":
		return buildParadiseTargetPrompt(rt, playerID, player, data)
	case "adventurer_steal_sky_mode":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【偷天换日】请选择效果：",
			Options: []model.PromptOption{
				{ID: "0", Label: "转移对方战绩区1红宝石到我方"},
				{ID: "1", Label: "将我方战绩区全部蓝水晶转换成红宝石"},
			},
			Min: 1,
			Max: 1,
		}
	case "adventurer_steal_sky_extra_action":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【偷天换日】请选择额外行动类型：",
			Options: []model.PromptOption{
				{ID: "0", Label: "额外+1攻击行动"},
				{ID: "1", Label: "额外+1法术行动"},
			},
			Min: 1,
			Max: 1,
		}
	default:
		return nil
	}
}

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
		if len(selectedIndices) == 2 && len(remainingIndices) > 0 {
			remaining = 1
		} else {
			return nil
		}
	}
	options := make([]model.PromptOption, 0, len(remainingIndices))
	for _, idx := range remainingIndices {
		if idx < 0 || idx >= len(player.Hand) {
			continue
		}
		card := player.Hand[idx]
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(idx),
			Label: promptfmt.FormatCardInfo(card),
		})
	}
	message := fmt.Sprintf("【欺诈】请选择手牌（还需选择%d张同系牌）：", remaining)
	if len(selectedIndices) >= 2 {
		message = "【欺诈】可选第三张牌转化为暗系攻击（或跳过）："
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func buildFraudAttackElementPrompt(playerID string, _ map[string]interface{}) *model.Prompt {
	options := []model.PromptOption{
		{ID: string(model.ElementWater), Label: "水"},
		{ID: string(model.ElementFire), Label: "火"},
		{ID: string(model.ElementEarth), Label: "地"},
		{ID: string(model.ElementWind), Label: "风"},
		{ID: string(model.ElementThunder), Label: "雷"},
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【欺诈】请选择攻击系别（不含光/暗）：",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func buildParadisePickPrompt(_ engineplayer.ChoiceRuntime, playerID string, player *model.Player, _ map[string]interface{}) *model.Prompt {
	if player == nil {
		return nil
	}
	options := make([]model.PromptOption, 0, 5)
	for i := 0; i <= 4; i++ {
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i),
			Label: fmt.Sprintf("转移%d个提炼成果给队友", i),
		})
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【冒险者天堂】选择要转移的提炼成果数量：",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func buildParadiseTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	eligibleIDs := runtimeutil.ParseStringSliceContextValue(data["eligible_targets"])
	if len(eligibleIDs) == 0 {
		eligibleIDs = runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
	}
	if len(eligibleIDs) == 0 {
		return nil
	}
	options := make([]model.PromptOption, 0, len(eligibleIDs))
	for _, targetID := range eligibleIDs {
		target := rt.LookupPlayer(targetID)
		if target == nil {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    targetID,
			Label: target.Name,
		})
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【冒险者天堂】选择转移目标：",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

// ---------------------------------------------------------------------------
// HandleChoice
// ---------------------------------------------------------------------------

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "adventurer_fraud_pick":
		return handleFraudPickChoice(rt, playerID, selectionIndex, ctxData)
	case "adventurer_fraud_attack_element":
		return handleFraudAttackElementChoice(rt, playerID, selectionIndex, ctxData)
	case "adventurer_paradise_pick":
		return handleParadisePickChoice(rt, playerID, selectionIndex, ctxData)
	case "adventurer_paradise_target":
		return handleParadiseTargetChoice(rt, playerID, selectionIndex, ctxData)
	case "adventurer_steal_sky_mode":
		return handleStealSkyModeChoice(rt, playerID, selectionIndex, ctxData)
	case "adventurer_steal_sky_extra_action":
		return handleStealSkyExtraActionChoice(rt, playerID, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleFraudPickChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	user := rt.LookupPlayer(playerID)
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	selectedIndices := runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])
	remainingIndices := runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	if len(remainingIndices) == 0 {
		remainingIndices = allHandIndices(user)
	}
	cardIdx := selectionIndex
	validCardIdx := false
	for _, idx := range remainingIndices {
		if idx == cardIdx {
			validCardIdx = true
			break
		}
	}
	if !validCardIdx && len(remainingIndices) > 0 {
		cardIdx, validCardIdx = runtimeutil.ResolveSelectionToCandidate(selectionIndex, remainingIndices)
	}
	if !validCardIdx || cardIdx < 0 || cardIdx >= len(user.Hand) {
		return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	for _, idx := range selectedIndices {
		if idx == cardIdx {
			return true, fmt.Errorf("这张牌已经被选择过")
		}
	}
	selectedIndices = append(selectedIndices, cardIdx)
	nextRemaining := make([]int, 0, len(remainingIndices)-1)
	for _, idx := range remainingIndices {
		if idx != cardIdx {
			nextRemaining = append(nextRemaining, idx)
		}
	}
	ctxData["selected_indices"] = selectedIndices
	ctxData["remaining_indices"] = nextRemaining

	needCount := runtimeutil.ToIntContextValue(ctxData["need_count"])
	if needCount <= 0 {
		needCount = 2
	}

	// 还需要选择更多牌
	if len(selectedIndices) < needCount {
		_ = rt.ReplacePendingInterruptContext(ctxData)
		rt.NotifyInterruptPrompt()
		return true, nil
	}

	// 已选>=2张：检查是否已选3张（自动暗系）
	if len(selectedIndices) >= 3 {
		return resolveFraudDarkAttack(rt, user, selectedIndices, ctxData)
	}

	// 已选2张：检查同系
	cards := make([]model.Card, 0, len(selectedIndices))
	for _, idx := range selectedIndices {
		if idx < len(user.Hand) {
			cards = append(cards, user.Hand[idx])
		}
	}
	commonElement := findCommonElement(cards)
	if commonElement == "" {
		// 无同系，继续选择第三张
		_ = rt.ReplacePendingInterruptContext(ctxData)
		rt.NotifyInterruptPrompt()
		return true, nil
	}
	// 有同系牌
	seqRemaining := runtimeutil.ToIntContextValue(ctxData["sequential_remaining"])
	if seqRemaining > 0 {
		// 顺序多选还有牌要处理，保持 choice_type 不变
		ctxData["selected_element"] = string(commonElement)
		_ = rt.ReplacePendingInterruptContext(ctxData)
		rt.NotifyInterruptPrompt()
		return true, nil
	}
	// 这是最后一张牌，转为攻击元素选择
	ctxData["choice_type"] = "adventurer_fraud_attack_element"
	ctxData["selected_element"] = string(commonElement)
	_ = rt.ReplacePendingInterruptContext(ctxData)
	rt.NotifyInterruptPrompt()
	return true, nil
}

// resolveFraudDarkAttack 处理3张及以上牌的暗系攻击结算。
func resolveFraudDarkAttack(rt engineplayer.ChoiceRuntime, user *model.Player, selectedIndices []int, ctxData map[string]interface{}) (bool, error) {
	removed := removeCardsByIndicesFromHand(user, selectedIndices)
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)
	virtualCard := model.Card{
		ID:          fmt.Sprintf("fraud_virtual_attack_%s_%d", user.ID, len(removed)),
		Name:        "欺诈",
		Type:        model.CardTypeAttack,
		Element:     model.ElementDark,
		Damage:      2,
		Description: "由欺诈视为的暗系攻击",
	}
	targetID, _ := ctxData["fraud_target_id"].(string)
	rt.EnqueueVirtualAttack(user.ID, targetID, virtualCard, "adventurer_fraud")
	rt.Log(fmt.Sprintf("%s 发动 [欺诈]：选%d张牌自动转化为暗系攻击", user.Name, len(removed)))
	rt.PopInterrupt()
	rt.EnterActionExecutionStage()
	return true, nil
}

func handleFraudAttackElementChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	user := rt.LookupPlayer(playerID)
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	targetID, _ := ctxData["fraud_target_id"].(string)
	target := rt.LookupPlayer(targetID)
	if target == nil {
		return true, fmt.Errorf("欺诈目标不存在")
	}
	// 五元素选项索引：水=0, 火=1, 地=2, 风=3, 雷=4
	elements := []model.Element{model.ElementWater, model.ElementFire, model.ElementEarth, model.ElementWind, model.ElementThunder}
	if selectionIndex < 0 || selectionIndex >= len(elements) {
		return true, fmt.Errorf("无效的元素索引: %d", selectionIndex)
	}
	selectedElement := elements[selectionIndex]
	selectedIndices := runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])
	removed := removeCardsByIndicesFromHand(user, selectedIndices)
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)
	virtualCard := model.Card{
		ID:          fmt.Sprintf("fraud_virtual_attack_%s_%d", user.ID, len(removed)),
		Name:        "欺诈",
		Type:        model.CardTypeAttack,
		Element:     selectedElement,
		Damage:      2,
		Description: "由欺诈视为的主动攻击",
	}
	rt.EnqueueVirtualAttack(user.ID, target.ID, virtualCard, "adventurer_fraud")
	rt.Log(fmt.Sprintf("%s 发动 [欺诈]：对 %s 发起%s系攻击", user.Name, target.Name, selectedElement))
	rt.PopInterrupt()
	rt.EnterActionExecutionStage()
	return true, nil
}

func handleParadisePickChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	ctxData["paradise_transfer_count"] = selectionIndex
	ctxData["choice_type"] = "adventurer_paradise_target"
	_ = rt.ReplacePendingInterruptContext(ctxData)
	rt.NotifyInterruptPrompt()
	return true, nil
}

func handleParadiseTargetChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	// paradise_target uses ally_ids not eligible_targets
	eligibleIDs := runtimeutil.ParseStringSliceContextValue(ctxData["eligible_targets"])
	if len(eligibleIDs) == 0 {
		eligibleIDs = runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	}
	if selectionIndex < 0 || selectionIndex >= len(eligibleIDs) {
		return true, fmt.Errorf("无效的目标索引")
	}
	targetID := eligibleIDs[selectionIndex]
	target := rt.LookupPlayer(targetID)
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
	user := rt.LookupPlayer(userID)

	if fromPending {
		// Forced path: energy was deducted from camp but never added to player.
		// Add directly to target.
		target.Gem += transferGem
		target.Crystal += transferCrystal
	} else if user != nil {
		// Optional path: energy was already added to player, transfer from player to target.
		if transferGem > 0 && user.Gem >= transferGem {
			user.Gem -= transferGem
			target.Gem += transferGem
		}
		if transferCrystal > 0 && user.Crystal >= transferCrystal {
			user.Crystal -= transferCrystal
			target.Crystal += transferCrystal
		}
	}
	// Deduct 1 energy cost from user as skill requirement (applies to both paths)
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

func handleStealSkyModeChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	user := rt.LookupPlayer(playerID)
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		// 转移对方战绩区1红宝石到我方
		enemyCamp, _ := ctxData["enemy_camp"].(string)
		selfCamp, _ := ctxData["self_camp"].(string)
		enemyGems := rt.GetCampGems(enemyCamp)
		if enemyGems <= 0 {
			rt.Log(fmt.Sprintf("%s 的 [偷天换日] 失败：对方战绩区无红宝石", user.Name))
			rt.PopInterrupt()
			return true, nil
		}
		rt.ModifyGem(enemyCamp, -1)
		rt.ModifyGem(selfCamp, 1)
		rt.Log(fmt.Sprintf("%s 的 [偷天换日]：转移对方1红宝石到我方", user.Name))
	case 1:
		// 将我方战绩区全部蓝水晶转换成红宝石
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
	// 进入额外行动选择
	ctxData["choice_type"] = "adventurer_steal_sky_extra_action"
	_ = rt.ReplacePendingInterruptContext(ctxData)
	rt.NotifyInterruptPrompt()
	return true, nil
}

func handleStealSkyExtraActionChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		model.AppendAttackAction(user, "偷天换日")
	case 1:
		model.AppendMagicAction(user, "偷天换日")
	default:
		return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	rt.PopInterrupt()
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

func findCommonElement(cards []model.Card) model.Element {
	if len(cards) == 0 {
		return ""
	}
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