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

const (
	adventurerFraudFlowID      = "adventurer_fraud"
	adventurerFraudCardsStep   = "cards"
	adventurerFraudElementStep = "element"
	adventurerFraudTargetStep  = "target"
)

var adventurerFraudFlowRuntime = model.MustNewPromptFlowRuntime(adventurerFraudFlowID, []model.PromptFlowStepSpec{
	{ID: adventurerFraudCardsStep, ChoiceType: "adventurer_fraud_pick", CancelPolicy: model.CancelPolicyAbort},
	{ID: adventurerFraudElementStep, ChoiceType: "adventurer_fraud_attack_element", CancelPolicy: model.CancelPolicyBack},
	{ID: adventurerFraudTargetStep, ChoiceType: "adventurer_fraud_target", CancelPolicy: model.CancelPolicyBack},
})

// ChoiceSpecs 声明式选择流程条目。
func ChoiceSpecs() []engineplayer.ChoiceSpec {
	return []engineplayer.ChoiceSpec{
		{ChoiceType: "adventurer_extract_paradise_check", BuildPrompt: buildExtractParadiseCheckPrompt, HandleChoice: handleExtractParadiseCheck},
		{ChoiceType: "adventurer_paradise_pick", BuildPrompt: buildParadiseAllyPickPrompt, HandleChoice: handleParadiseAllyPick},
		{
			ChoiceType:        "adventurer_fraud_pick",
			BuildPrompt:       buildFraudPickPrompt,
			HandleChoice:      handleFraudPick,
			HandleMultiSelect: handleFraudPickMultiSelect,
		},
		{ChoiceType: "adventurer_fraud_attack_element", BuildPrompt: buildFraudElementPrompt, HandleChoice: handleFraudElement},
		{ChoiceType: "adventurer_fraud_target", BuildPrompt: buildFraudTargetPrompt, HandleChoice: handleFraudTarget},
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
	p := engineplayer.NewPrompt(playerID, "【偷天换日】请选择效果：").
		OptionsFromLabels("转移对方战绩区1红宝石到我方", "将我方战绩区全部蓝水晶转换成红宝石").
		Build()
	p.Presentation = &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}
	return p
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
// 冒险者天堂 - 选择队友直接提炼
// ---------------------------------------------------------------------------

func buildParadiseAllyPickPrompt(rt engineplayer.ChoiceRuntime, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
	if len(allyIDs) == 0 {
		return nil
	}
	options := make([]model.PromptOption, 0, len(allyIDs))
	for _, allyID := range allyIDs {
		ally := rt.GetPlayers()[allyID]
		if ally != nil {
			options = append(options, model.PromptOption{ID: allyID, Label: ally.Name, TargetID: allyID})
		}
	}
	return &model.Prompt{
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		Message:      "【冒险者天堂】选择队友代为提炼：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
	}
}

func handleParadiseAllyPick(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
		return true, fmt.Errorf("无效的队友索引")
	}
	allyID := allyIDs[selectionIndex]
	ally := rt.GetPlayers()[allyID]
	if ally == nil {
		return true, fmt.Errorf("队友不存在")
	}
	rt.Log(fmt.Sprintf("[冒险者天堂] %s 选择队友 %s 代为提炼", playerID, ally.Name))
	rt.PopInterrupt()
	rt.StartExtractForPlayer(allyID)
	return true, nil
}

// ---------------------------------------------------------------------------
// 冒险者天堂 - 提炼时询问是否发动
// ---------------------------------------------------------------------------

func buildExtractParadiseCheckPrompt(_ engineplayer.ChoiceRuntime, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	p := engineplayer.NewPrompt(playerID, "是否发动[冒险者天堂]，让队友代为提炼？").
		Options(
			engineplayer.Option("yes", "是，发动冒险者天堂"),
			engineplayer.Option("no", "否，自行提炼"),
		).
		Build()
	p.Presentation = &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}
	return p
}

func handleExtractParadiseCheck(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	rt.PopInterrupt()
	if selectionIndex == 0 {
		// 是 → 推送队友选择
		allies := eligibleAllyIDsForChoice(rt, user)
		if len(allies) == 0 {
			// 没有可选队友，退回自行提炼
			setDeclinedParadise(user)
			rt.StartExtractForPlayer(playerID)
			return true, nil
		}
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: playerID,
			Context: map[string]interface{}{
				"choice_type": "adventurer_paradise_pick",
				"user_id":     playerID,
				"ally_ids":    allies,
			},
		})
		rt.NotifyInterruptPrompt()
		return true, nil
	}
	// 否 → 正常提炼（设置标志防止 override hook 再次拦截）
	setDeclinedParadise(user)
	rt.StartExtractForPlayer(playerID)
	return true, nil
}

func setDeclinedParadise(p *model.Player) {
	if p == nil {
		return
	}
	if p.TurnState.SkillFlowState == nil {
		p.TurnState.SkillFlowState = map[string]int{}
	}
	p.TurnState.SkillFlowState["adventurer_declined_paradise"] = 1
}

func eligibleAllyIDsForChoice(rt engineplayer.ChoiceRuntime, p *model.Player) []string {
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		ally := rt.GetPlayers()[pid]
		if ally == nil || ally.Camp != p.Camp || ally.ID == p.ID {
			continue
		}
		maxEnergy := rt.GetPlayerEnergyCap(ally)
		if ally.Gem+ally.Crystal < maxEnergy {
			ids = append(ids, pid)
		}
	}
	return ids
}

// ---------------------------------------------------------------------------
// 欺诈 - 选手牌（2张→五系选择 / 3张→暗灭）
// ---------------------------------------------------------------------------

func buildFraudPickPrompt(_ engineplayer.ChoiceRuntime, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	if player == nil {
		return nil
	}
	if _, err := model.RequirePromptFlow(data, adventurerFraudFlowID, "欺诈"); err != nil {
		return nil
	}
	remainingIndices := runtimeutil.ParseChoiceIntSlice(data["remaining_indices"])
	if len(remainingIndices) == 0 {
		remainingIndices = engineplayer.AllHandIndices(player)
	}
	opts := make([]engineplayer.PromptOptionSpec, 0, len(remainingIndices))
	for _, idx := range remainingIndices {
		if idx >= 0 && idx < len(player.Hand) {
			opts = append(opts, engineplayer.CardOption(strconv.Itoa(idx), promptfmt.FormatCardInfo(player.Hand[idx]), player.Hand[idx].ID))
		}
	}
	p := engineplayer.NewPrompt(playerID, "【欺诈】请选择2~3张同系手牌：").
		Options(opts...).Build()
	p.Type = model.PromptChooseCards
	p.Min = 2
	p.Max = 3
	p.Presentation = &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand", CardFilter: "same_element_combo"}
	return p
}

func handleFraudPick(_ engineplayer.ChoiceRuntime, _ string, _ int, _ map[string]interface{}) (bool, error) {
	return true, fmt.Errorf("欺诈请一次选择2~3张同系手牌后确认提交")
}

func handleFraudPickMultiSelect(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	if len(selections) < 2 || len(selections) > 3 {
		return true, fmt.Errorf("欺诈需要选择2或3张同系手牌")
	}
	flow, err := model.RequirePromptFlow(ctxData, adventurerFraudFlowID, "欺诈")
	if err != nil {
		return true, err
	}
	remainingIndices := runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	if len(remainingIndices) == 0 {
		remainingIndices = engineplayer.AllHandIndices(user)
	}
	allowed := map[int]bool{}
	for _, idx := range remainingIndices {
		allowed[idx] = true
	}

	seen := map[int]bool{}
	var commonElement model.Element
	for _, idx := range selections {
		if idx < 0 || idx >= len(user.Hand) || !allowed[idx] {
			return true, fmt.Errorf("无效的选项索引: %d", idx)
		}
		if seen[idx] {
			return true, fmt.Errorf("欺诈不能重复选择同一张牌")
		}
		seen[idx] = true

		cardElement := user.Hand[idx].Element
		if cardElement == "" {
			return true, fmt.Errorf("欺诈需选择有系别的手牌")
		}
		if commonElement == "" {
			commonElement = cardElement
			continue
		}
		if cardElement != commonElement {
			return true, fmt.Errorf("欺诈需选择同系牌（2张可选五系攻击，3张自动转暗灭）")
		}
	}

	flow.PutSelection(adventurerFraudCardsStep, model.PromptFlowSelection{OptionIndexes: append([]int{}, selections...)})
	if len(selections) == 3 {
		flow.PutSelection(adventurerFraudElementStep, model.PromptFlowSelection{Element: string(model.ElementDark)})
		return advanceFraudTargetStep(rt, user, ctxData, flow)
	}

	flow.PutSelection(adventurerFraudElementStep, model.PromptFlowSelection{Element: string(commonElement)})
	if err := adventurerFraudFlowRuntime.MoveTo(flow, adventurerFraudElementStep); err != nil {
		return true, err
	}
	ctxData["choice_type"] = "adventurer_fraud_attack_element"
	engineplayer.NotifyChoiceContext(rt, ctxData)
	return true, nil
}

func buildFraudElementPrompt(_ engineplayer.ChoiceRuntime, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	p := engineplayer.NewPrompt(playerID, "【欺诈】请选择攻击系别（不含光/暗）：").
		Options(
			engineplayer.Option(string(model.ElementWater), "水"),
			engineplayer.Option(string(model.ElementFire), "火"),
			engineplayer.Option(string(model.ElementEarth), "地"),
			engineplayer.Option(string(model.ElementWind), "风"),
			engineplayer.Option(string(model.ElementThunder), "雷"),
		).
		Build()
	p.Presentation = &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "fraud_attack_element"}
	return p
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
	flow, err := model.RequirePromptFlow(ctxData, adventurerFraudFlowID, "欺诈")
	if err != nil {
		return true, err
	}
	selectedIndices := flow.Selection(adventurerFraudCardsStep).OptionIndexes
	flow.PutSelection(adventurerFraudElementStep, model.PromptFlowSelection{
		OptionIndexes: []int{selectionIndex},
		Element:       string(elements[selectionIndex]),
	})
	if len(selectedIndices) < 2 || len(selectedIndices) > 3 {
		return true, fmt.Errorf("欺诈需要选择2或3张同系手牌")
	}
	return advanceFraudTargetStep(rt, user, ctxData, flow)
}

func buildFraudTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	if player == nil {
		return nil
	}
	if _, err := model.RequirePromptFlow(data, adventurerFraudFlowID, "欺诈"); err != nil {
		return nil
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	if len(targetIDs) == 0 {
		targetIDs = fraudEnemyTargetIDs(rt, player)
		data["target_ids"] = targetIDs
	}
	if len(targetIDs) == 0 {
		return nil
	}
	return engineplayer.BuildTargetChoicePrompt(rt, "adventurer_fraud_target", playerID, "【欺诈】请选择攻击目标：", data, false)
}

func handleFraudTarget(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	flow, err := model.RequirePromptFlow(ctxData, adventurerFraudFlowID, "欺诈")
	if err != nil {
		return true, err
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if len(targetIDs) == 0 {
		targetIDs = flow.Selection(adventurerFraudTargetStep).TargetIDs
	}
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return true, fmt.Errorf("无效的目标索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return true, fmt.Errorf("欺诈目标不存在")
	}
	if target.Camp == user.Camp {
		return true, fmt.Errorf("欺诈只能选择敌方角色")
	}
	flow.PutSelection(adventurerFraudTargetStep, model.PromptFlowSelection{
		OptionIndexes: []int{selectionIndex},
		TargetIDs:     []string{targetID},
	})
	ctxData["fraud_target_id"] = targetID
	selectedIndices := flow.Selection(adventurerFraudCardsStep).OptionIndexes
	element := model.Element(flow.Selection(adventurerFraudElementStep).Element)
	if element == "" {
		return true, fmt.Errorf("欺诈攻击系别不存在")
	}
	return resolveFraudAttack(rt, user, selectedIndices, ctxData, element)
}

func advanceFraudTargetStep(rt engineplayer.ChoiceRuntime, user *model.Player, ctxData map[string]interface{}, flow *model.PromptFlowState) (bool, error) {
	targetIDs := fraudEnemyTargetIDs(rt, user)
	if len(targetIDs) == 0 {
		return true, fmt.Errorf("欺诈没有可选择的敌方目标")
	}
	ctxData["target_ids"] = targetIDs
	flow.PutSelection(adventurerFraudTargetStep, model.PromptFlowSelection{TargetIDs: append([]string{}, targetIDs...)})
	return true, engineplayer.AdvancePromptFlowRuntimeChoice(rt, ctxData, adventurerFraudFlowRuntime, flow, adventurerFraudTargetStep)
}

func fraudEnemyTargetIDs(rt engineplayer.ChoiceRuntime, user *model.Player) []string {
	if rt == nil || user == nil {
		return nil
	}
	if ids := rt.CampEnemyIDs(user.Camp); len(ids) > 0 {
		return ids
	}
	ids := make([]string, 0)
	for _, pid := range rt.GetPlayerOrder() {
		p := rt.GetPlayers()[pid]
		if p != nil && p.ID != user.ID && p.Camp != user.Camp {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// resolveFraudAttack 统一处理欺诈攻击结算：弃牌、构建虚拟攻击、入队。
func resolveFraudAttack(rt engineplayer.ChoiceRuntime, user *model.Player, indices []int, ctxData map[string]interface{}, element model.Element) (bool, error) {
	targetID, _ := ctxData["fraud_target_id"].(string)
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return true, fmt.Errorf("欺诈目标不存在")
	}
	removed, _ := engineplayer.RemoveCardsByIndicesFromHand(user, indices)
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
