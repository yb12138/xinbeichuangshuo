// gameflow: 精灵射手角色选择流。

package elf_archer

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
	case "elf_elemental_shot_cost":
		canMagic, _ := data["can_discard_magic"].(bool)
		canBless, _ := data["can_remove_bless"].(bool)
		options := make([]model.PromptOption, 0, 2)
		if canMagic {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "弃1张法术牌发动"})
		}
		if canBless {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除1个祝福发动"})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【元素射击】请选择发动消耗：", Options: options, Min: 1, Max: 1}

	case "elf_elemental_shot_discard_magic":
		if player == nil {
			return nil
		}
		idxs := runtimeutil.ParseChoiceIntSlice(data["magic_indices"])
		options := make([]model.PromptOption, 0, len(idxs))
		for _, idx := range idxs {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx]))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【元素射击】请选择弃置的法术牌：", Options: options, Min: 1, Max: 1}

	case "elf_elemental_shot_remove_blessing":
		if player == nil {
			return nil
		}
		blessings := elfBlessingCards(player)
		idxs := runtimeutil.ParseChoiceIntSlice(data["blessing_indices"])
		options := make([]model.PromptOption, 0, len(idxs))
		for _, idx := range idxs {
			if idx < 0 || idx >= len(blessings) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(blessings[idx]))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【元素射击】请选择要移除的祝福：", Options: options, Min: 1, Max: 1}

	case "elf_animal_companion_confirm":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【动物伙伴】是否发动（摸1弃1）？", Options: []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}}, Min: 1, Max: 1}

	case "elf_pet_empower_confirm":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【宠物强化】是否消耗1蓝水晶，将效果改为任意角色摸1弃1？", Options: []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}}, Min: 1, Max: 1}

	case "elf_elemental_shot_water_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【水之矢】请选择+1治疗目标：", data, false)
	case "elf_elemental_shot_earth_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【地之矢】请选择1点法术伤害目标：", data, false)
	case "elf_pet_empower_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【宠物强化】请选择摸1弃1目标：", data, false)
	case "elf_ritual_release_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【精灵密仪】你已无祝福，转正并请选择1名敌方角色承受2点法术伤害：", data, false)
	}
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "elf_elemental_shot_cost":
		return true, handleElfElementalShotCost(rt, selectionIndex, ctxData)

	case "elf_elemental_shot_discard_magic", "elf_elemental_shot_remove_blessing":
		return true, handleElfElementalShotDiscardOrRemoveBlessing(rt, selectionIndex, ctxData)

	case "elf_animal_companion_confirm":
		return true, handleElfAnimalCompanionConfirm(rt, selectionIndex, ctxData)

	case "elf_pet_empower_confirm":
		return true, handleElfPetEmpowerConfirm(rt, selectionIndex, ctxData)

	case "elf_pet_empower_target":
		return true, handleElfPetEmpowerTarget(rt, selectionIndex, ctxData)

	case "elf_elemental_shot_water_target", "elf_elemental_shot_earth_target", "elf_ritual_release_target":
		return true, handleElfElementalShotOrRitualTarget(rt, selectionIndex, ctxData)
	}

	return false, nil
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "elf_elemental_shot_cost" {
		return false, nil
	}

	rt.PopInterrupt()
	if user := rt.GetPlayers()[playerID]; user != nil {
		rt.Log(fmt.Sprintf("%s 放弃发动 [元素射击]", user.Name))
	}
	if rt.GetPendingInterrupt() == nil {
		if len(rt.GetActionQueue()) > 0 {
			rt.EnterActionExecutionStage()
		} else if len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		}
	}
	return true, nil
}

func handleElfElementalShotCost(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	canMagic, _ := ctxData["can_discard_magic"].(bool)
	canBless, _ := ctxData["can_remove_bless"].(bool)
	modeList := make([]int, 0, 2)
	if canMagic {
		modeList = append(modeList, 0)
	}
	if canBless {
		modeList = append(modeList, 1)
	}
	if selectionIndex < 0 || selectionIndex >= len(modeList) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if modeList[selectionIndex] == 0 {
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = map[string]interface{}{
				"choice_type":    "elf_elemental_shot_discard_magic",
				"user_id":        userID,
				"magic_indices":  getCardIndicesByType(user, model.CardTypeMagic),
				"attack_element": ctxData["attack_element"],
			}
		}
	} else {
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = map[string]interface{}{
				"choice_type":      "elf_elemental_shot_remove_blessing",
				"user_id":          userID,
				"blessing_indices": elfBlessingHandIndices(user),
				"attack_element":   ctxData["attack_element"],
			}
		}
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleElfElementalShotDiscardOrRemoveBlessing(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	choiceType, _ := ctxData["choice_type"].(string)
	key := "magic_indices"
	if choiceType == "elf_elemental_shot_remove_blessing" {
		key = "blessing_indices"
	}
	candidates := runtimeutil.ParseChoiceIntSlice(ctxData[key])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, candidates)
	if !ok {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	var card model.Card
	if choiceType == "elf_elemental_shot_remove_blessing" {
		blessings := elfBlessingCards(user)
		if cardIdx < 0 || cardIdx >= len(blessings) {
			return fmt.Errorf("无效的祝福索引: %d", selectionIndex)
		}
		card = blessings[cardIdx]
		removeElfBlessingByCardID(user, card.ID)
	} else {
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的手牌索引: %d", selectionIndex)
		}
		card = user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
	}
	rt.NotifyCardRevealed(userID, []model.Card{card}, "discard")
	rt.AppendToDiscard([]model.Card{card})

	if user.TurnState.SkillFlowState == nil {
		user.TurnState.SkillFlowState = map[string]int{}
	}
	user.TurnState.SkillFlowState["elf_elemental_shot_water_pending"] = 0
	user.TurnState.SkillFlowState["elf_elemental_shot_earth_pending"] = 0
	user.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] = 0
	attackElement, _ := ctxData["attack_element"].(string)
	switch attackElement {
	case string(model.ElementFire):
		rt.ApplyNextAttackDamageRule(userID, "elf_elemental_shot_fire_attack_bonus", "elf_elemental_shot", 1, model.RuleLifeThisEffectChain)
	case string(model.ElementWater):
		user.TurnState.SkillFlowState["elf_elemental_shot_water_pending"] = 1
	case string(model.ElementWind):
		user.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] = 1
	case string(model.ElementThunder):
		rt.ApplyNextAttackInterceptTagRule(userID, "elf_elemental_shot_thunder_attack_tag", "elf_elemental_shot", model.CombatInterceptUnrespondable, model.RuleLifeThisEffectChain)
	case string(model.ElementEarth):
		user.TurnState.SkillFlowState["elf_elemental_shot_earth_pending"] = 1
	}
	rt.Log(fmt.Sprintf("%s 发动 [元素射击]（%s）", user.Name, attackElement))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if len(rt.GetActionQueue()) > 0 {
			rt.EnterActionExecutionStage()
		} else if len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		}
	}
	return nil
}

func handleElfAnimalCompanionConfirm(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex == 1 {
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		}
		return nil
	}
	if selectionIndex != 0 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if rt.CanPayCrystalCost(userID, 1) {
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = map[string]interface{}{
				"choice_type": "elf_pet_empower_confirm",
				"user_id":     userID,
			}
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	rt.DrawCards(userID, 1)
	rt.PushInterrupt(newDiscardChoiceInterrupt(userID, map[string]interface{}{
		"discard_count":     1,
		"stay_in_turn":      true,
		"prompt":            "【动物伙伴】请选择弃置1张牌：",
		"exclude_blessings": true,
	}))
	rt.PopInterrupt()
	return nil
}

func handleElfPetEmpowerConfirm(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex == 0 && rt.ConsumeCrystalCost(userID, 1) {
		targetIDs := rt.GetPlayerOrder()
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = map[string]interface{}{
				"choice_type": "elf_pet_empower_target",
				"user_id":     userID,
				"target_ids":  targetIDs,
			}
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	rt.DrawCards(userID, 1)
	rt.PushInterrupt(newDiscardChoiceInterrupt(userID, map[string]interface{}{
		"discard_count":     1,
		"stay_in_turn":      true,
		"prompt":            "【动物伙伴】请选择弃置1张牌：",
		"exclude_blessings": true,
	}))
	rt.PopInterrupt()
	return nil
}

func handleElfPetEmpowerTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	rt.DrawCards(targetID, 1)
	if len(target.Hand) > rt.GetMaxHand(target) {
		rt.Log(fmt.Sprintf("[宠物强化] %s 摸牌后触发爆牌，本次弃1由爆牌弃牌结算承担", target.Name))
		rt.PopInterrupt()
		return nil
	}
	rt.PushInterrupt(newDiscardChoiceInterrupt(targetID, map[string]interface{}{
		"discard_count":     1,
		"stay_in_turn":      true,
		"prompt":            fmt.Sprintf("【宠物强化】%s 请弃置1张牌：", target.Name),
		"exclude_blessings": isElfArcher(target),
	}))
	rt.PopInterrupt()
	return nil
}

func handleElfElementalShotOrRitualTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "elf_elemental_shot_water_target":
		rt.Heal(targetID, 1)
	case "elf_elemental_shot_earth_target":
		rt.AddPendingDamage(model.PendingDamage{SourceID: userID, TargetID: targetID, Damage: 1, DamageType: model.MagicAttack})
	case "elf_ritual_release_target":
		leaveElfArcherRitualForm(user)
		user.TurnState.SkillFlowState["elf_ritual_release_waiting"] = 0
		rt.AddPendingDamage(model.PendingDamage{SourceID: userID, TargetID: targetID, Damage: 2, DamageType: model.MagicAttack})
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

// Helper functions for elf_archer blessings

func elfBlessingCards(player *model.Player) []model.Card {
	if player == nil {
		return nil
	}
	covers := player.GetCoverCardsByEffect(model.EffectElfBlessing)
	if len(covers) == 0 {
		return nil
	}
	out := make([]model.Card, 0, len(covers))
	for _, fc := range covers {
		if fc == nil {
			continue
		}
		out = append(out, fc.Card)
	}
	return out
}

func elfBlessingHandIndices(player *model.Player) []int {
	if player == nil {
		return nil
	}
	count := countElfBlessings(player)
	if count == 0 {
		return nil
	}
	idxs := make([]int, count)
	for i := 0; i < count; i++ {
		idxs[i] = i
	}
	return idxs
}

func countElfBlessings(player *model.Player) int {
	if player == nil {
		return 0
	}
	return len(player.GetCoverCardsByEffect(model.EffectElfBlessing))
}

func removeElfBlessingByCardID(player *model.Player, cardID string) bool {
	if player == nil || cardID == "" {
		return false
	}
	removed := false
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectElfBlessing {
			continue
		}
		if !removed && fc.Card.ID == cardID {
			player.RemoveFieldCard(fc)
			removed = true
			break
		}
	}

	elfBlessingPrefix := "elf_blessing_"
	target := elfBlessingPrefix + cardID
	newZone := make([]string, 0, len(player.CharaZone))
	removedZone := false
	for _, z := range player.CharaZone {
		if !removedZone && z == target {
			removedZone = true
			continue
		}
		newZone = append(newZone, z)
	}
	player.CharaZone = newZone
	return removed || removedZone
}

func leaveElfArcherRitualForm(player *model.Player) bool {
	if player == nil {
		return false
	}
	if player.Form != model.FormElfArcherRitual {
		return false
	}
	player.Form = ""
	return true
}

func isElfArcher(player *model.Player) bool {
	if player == nil {
		return false
	}
	if player.Character != nil && player.Character.ID == "elf_archer" {
		return true
	}
	return player.Role == "elf_archer"
}

func getCardIndicesByType(player *model.Player, cardType model.CardType) []int {
	if player == nil {
		return nil
	}
	var out []int
	for i, c := range player.Hand {
		if c.Type == cardType {
			out = append(out, i)
		}
	}
	return out
}

func newDiscardChoiceInterrupt(playerID string, data map[string]interface{}) *model.Interrupt {
	return &model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: playerID,
		Context:  data,
	}
}

var _ engineplayer.CancelChoiceHandler = choiceHandler{}
