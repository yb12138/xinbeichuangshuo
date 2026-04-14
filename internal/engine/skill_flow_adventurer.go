// gameflow: 冒险家：地下法则、欺诈攻击等。

package engine

import (
	"fmt"
	"sort"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) isForcedAdventurerParadiseResponse(playerID string) bool {
	intr := e.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill || intr.PlayerID != playerID {
		return false
	}
	player := e.State.Players[playerID]
	if player == nil || player.TurnState.SkillFlowState == nil || player.TurnState.SkillFlowState["adventurer_extract_requires_paradise"] <= 0 {
		return false
	}
	for _, sid := range intr.SkillIDs {
		if sid == "adventurer_paradise" {
			return true
		}
	}
	return false
}

func (e *GameEngine) resolveAdventurerLuckyFortuneFromFraud(user *model.Player) {
	if user == nil {
		return
	}
	user.Crystal++
	e.Log(fmt.Sprintf("%s 的 [强运] 触发，获得1蓝水晶", user.Name))
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: 强运", user.Name))
}

func (e *GameEngine) resolveAdventurerUndergroundLaw(user *model.Player) {
	if user == nil {
		return
	}
	e.ModifyGem(string(user.Camp), 2)
	e.Log(fmt.Sprintf("%s 的 [地下法则] 生效，本次购买改为战绩区+2宝石", user.Name))
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: 地下法则", user.Name))
}

func fraudAttackElementPool() []string {
	return []string{
		string(model.ElementWater),
		string(model.ElementFire),
		string(model.ElementEarth),
		string(model.ElementWind),
		string(model.ElementThunder),
	}
}

func adventurerFraudPickCandidateIndices(user *model.Player) []int {
	if user == nil || len(user.Hand) == 0 {
		return nil
	}
	counts := map[model.Element]int{}
	for _, c := range user.Hand {
		counts[c.Element]++
	}
	out := make([]int, 0, len(user.Hand))
	for idx, c := range user.Hand {
		if c.Element == "" {
			continue
		}
		if counts[c.Element] >= 2 {
			out = append(out, idx)
		}
	}
	return out
}

func normalizeDistinctSortedIndices(indices []int) []int {
	if len(indices) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(indices))
	out := make([]int, 0, len(indices))
	for _, idx := range indices {
		if _, ok := seen[idx]; ok {
			continue
		}
		seen[idx] = struct{}{}
		out = append(out, idx)
	}
	sort.Ints(out)
	return out
}

func (e *GameEngine) buildAdventurerChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "adventurer_fraud_pick":
		if player == nil {
			return nil
		}
		candidateIndices := adventurerFraudPickCandidateIndices(player)
		options := make([]model.PromptOption, 0, len(candidateIndices))
		for _, idx := range candidateIndices {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:       model.PromptChooseCards,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【欺诈】请在手牌区选择2~3张同系牌（3张自动转为暗灭攻击；2张后可选择五系攻击）：",
			Options:    options,
			Min:        2,
			Max:        3,
		}

	case "adventurer_fraud_attack_element":
		options := make([]model.PromptOption, 0, len(fraudAttackElementPool()))
		for _, ele := range fraudAttackElementPool() {
			options = append(options, model.PromptOption{ID: ele, Label: fmt.Sprintf("%s系", elementNameForPrompt(ele))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: "【欺诈】请选择本次攻击系别（五系，不可选暗灭）：", Options: options, Min: 1, Max: 1}

	case "adventurer_paradise_target":
		allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
		options := make([]model.PromptOption, 0, len(allyIDs))
		for _, allyID := range allyIDs {
			if target := e.State.Players[allyID]; target != nil {
				options = append(options, model.PromptOption{ID: allyID, Label: target.Name})
			}
		}
		transferGem := runtimeutil.ToIntContextValue(data["transfer_gem"])
		transferCrystal := runtimeutil.ToIntContextValue(data["transfer_crystal"])
		transferTotal := runtimeutil.ToIntContextValue(data["transfer_total"])
		if transferTotal <= 0 {
			transferTotal = transferGem + transferCrystal
		}
		message := "【冒险者天堂】请选择接收能量的队友："
		if transferTotal > 0 {
			message = fmt.Sprintf("【冒险者天堂】请选择接收提炼结果的队友（共%d点：宝石%d / 水晶%d）：", transferTotal, transferGem, transferCrystal)
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1}

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
	}

	return nil
}

func (e *GameEngine) resolveAdventurerFraudAttack(ctxData map[string]interface{}, discardIndices []int, attackElement model.Element, canBeResponded bool) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	if attackElement == "" {
		return true, fmt.Errorf("欺诈攻击元素无效")
	}

	indices := normalizeDistinctSortedIndices(discardIndices)
	if len(indices) == 0 {
		return true, fmt.Errorf("欺诈弃牌为空")
	}
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	for _, idx := range indices {
		if idx < 0 || idx >= len(user.Hand) {
			return true, fmt.Errorf("弃牌索引越界")
		}
		card := user.Hand[idx]
		e.NotifyCardRevealed(userID, []model.Card{card}, "discard")
		user.Hand = append(user.Hand[:idx], user.Hand[idx+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, card)
	}

	rawCtx, ok := ctxData["user_ctx"].(*model.Context)
	if ok && rawCtx != nil && rawCtx.EventCtx != nil && rawCtx.EventCtx.Card != nil && rawCtx.EventCtx.AttackInfo != nil {
		rawCtx.EventCtx.Card.Faction = ""
		rawCtx.EventCtx.Card.Element = attackElement
		rawCtx.EventCtx.Card.Damage = 2
		rawCtx.EventCtx.AttackInfo.CanBeResponded = canBeResponded
		e.Log(fmt.Sprintf("%s 发动[欺诈]完成，弃同系牌并将本次攻击改为 %s", user.Name, attackElement))
		e.resolveAdventurerLuckyFortuneFromFraud(user)
		e.PopInterrupt()
		return true, nil
	}

	targetID, _ := ctxData["fraud_target_id"].(string)
	if targetID == "" || e.State.Players[targetID] == nil {
		return true, fmt.Errorf("欺诈目标无效")
	}
	virtualCard := model.Card{
		ID:      "fraud_virtual_attack",
		Name:    "欺诈",
		Type:    model.CardTypeAttack,
		Element: attackElement,
		Faction: "",
		Damage:  2,
	}
	e.State.ActionQueue = append(e.State.ActionQueue, model.QueuedAction{
		SourceID:        userID,
		TargetID:        targetID,
		Type:            model.ActionAttack,
		Element:         attackElement,
		Card:            &virtualCard,
		CardIndex:       -1,
		SourceSkill:     "adventurer_fraud",
		UsesVirtualCard: true,
	})
	e.resolveAdventurerLuckyFortuneFromFraud(user)
	e.enterActionExecutionStage()
	e.Log(fmt.Sprintf("%s 发动[欺诈]完成，弃同系牌并对 %s 发起%s系主动攻击", user.Name, e.State.Players[targetID].Name, attackElement))
	e.PopInterrupt()
	return true, nil
}

func (e *GameEngine) handleAdventurerFraudPickSelections(playerID string, selections []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("没有待处理的欺诈选择")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	userID, _ := ctxData["user_id"].(string)
	if userID == "" {
		userID = playerID
	}
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	allowed := map[int]struct{}{}
	for _, idx := range adventurerFraudPickCandidateIndices(user) {
		allowed[idx] = struct{}{}
	}

	picked := normalizeDistinctSortedIndices(selections)
	if len(picked) < 2 || len(picked) > 3 {
		return fmt.Errorf("【欺诈】需要选择2~3张同系牌")
	}
	var selectedElement model.Element
	for _, idx := range picked {
		if _, ok := allowed[idx]; !ok {
			return fmt.Errorf("【欺诈】请选择可用于发动的同系手牌")
		}
		if idx < 0 || idx >= len(user.Hand) {
			return fmt.Errorf("欺诈弃牌索引越界")
		}
		card := user.Hand[idx]
		if selectedElement == "" {
			selectedElement = card.Element
			continue
		}
		if card.Element != selectedElement {
			return fmt.Errorf("【欺诈】必须选择同系牌")
		}
	}
	if selectedElement == "" {
		return fmt.Errorf("【欺诈】所选卡牌元素无效")
	}

	if len(picked) == 3 {
		_, err := e.resolveAdventurerFraudAttack(ctxData, picked, model.ElementDark, false)
		return err
	}

	ctxData["choice_type"] = "adventurer_fraud_attack_element"
	ctxData["fraud_selected_indices"] = picked
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handleAdventurerChoiceInput(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "adventurer_fraud_pick":
		return true, e.handleAdventurerFraudPickSelections(playerID, []int{selectionIndex})

	case "adventurer_fraud_attack_element":
		attackElems := fraudAttackElementPool()
		if selectionIndex < 0 || selectionIndex >= len(attackElems) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected := runtimeutil.ParseChoiceIntSlice(ctxData["fraud_selected_indices"])
		if len(selected) == 0 {
			return true, fmt.Errorf("【欺诈】缺少已选手牌，请重新发动")
		}
		return e.resolveAdventurerFraudAttack(ctxData, selected, model.Element(attackElems[selectionIndex]), true)

	case "adventurer_paradise_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ally := e.State.Players[allyIDs[selectionIndex]]
		if ally == nil {
			return true, fmt.Errorf("队友不存在")
		}

		transferGem := runtimeutil.ToIntContextValue(ctxData["transfer_gem"])
		transferCrystal := runtimeutil.ToIntContextValue(ctxData["transfer_crystal"])
		transferTotal := runtimeutil.ToIntContextValue(ctxData["transfer_total"])
		if transferTotal <= 0 {
			transferTotal = transferGem + transferCrystal
		}
		fromPending, _ := ctxData["from_pending"].(bool)
		if transferTotal <= 0 {
			e.clearAdventurerExtractState(user)
			e.PopInterrupt()
			return true, nil
		}

		capLeft := e.getPlayerEnergyCap(ally) - (ally.Gem + ally.Crystal)
		if capLeft < transferTotal {
			return true, fmt.Errorf("%s 能量空间不足，无法接收全部提炼结果", ally.Name)
		}
		if !fromPending {
			if user.Gem < transferGem || user.Crystal < transferCrystal {
				return true, fmt.Errorf("自身提炼结果异常，无法转移")
			}
			user.Gem -= transferGem
			user.Crystal -= transferCrystal
		}
		ally.Gem += transferGem
		ally.Crystal += transferCrystal

		removedEnergy := false
		if user.Crystal > 0 {
			user.Crystal--
			removedEnergy = true
		} else if user.Gem > 0 {
			user.Gem--
			removedEnergy = true
		}
		e.clearAdventurerExtractState(user)
		if removedEnergy {
			e.Log(fmt.Sprintf("%s 发动[冒险者天堂]，将提炼结果交给 %s（宝石%d/水晶%d），并移除自身1点能量", user.Name, ally.Name, transferGem, transferCrystal))
		} else {
			e.Log(fmt.Sprintf("%s 发动[冒险者天堂]，将提炼结果交给 %s（宝石%d/水晶%d）", user.Name, ally.Name, transferGem, transferCrystal))
		}
		e.PopInterrupt()
		return true, nil

	case "adventurer_steal_sky_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		enemyCamp, _ := ctxData["enemy_camp"].(string)
		selfCamp, _ := ctxData["self_camp"].(string)
		switch selectionIndex {
		case 0:
			if e.GetCampGems(enemyCamp) > 0 {
				e.ModifyGem(enemyCamp, -1)
				e.ModifyGem(selfCamp, 1)
			}
		case 1:
			crystals := e.GetCampCrystals(selfCamp)
			if crystals > 0 {
				e.ModifyCrystal(selfCamp, -crystals)
				e.ModifyGem(selfCamp, crystals)
			}
		default:
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.State.PendingInterrupt.Context = map[string]interface{}{
			"choice_type": "adventurer_steal_sky_extra_action",
			"user_id":     userID,
		}
		e.notifyInterruptPrompt()
		e.Log(fmt.Sprintf("%s 完成[偷天换日]主效果，等待选择额外行动", user.Name))
		return true, nil

	case "adventurer_steal_sky_extra_action":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
		e.PopInterrupt()
		return true, nil
	}

	return false, nil
}
