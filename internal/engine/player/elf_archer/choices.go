// gameflow: 精灵射手角色选择流。

package elf_archer

import (
	"fmt"
	"sort"

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
	case "elf_archer_elemental_shot_pick":
		if player == nil {
			return nil
		}
		attackCardID, _ := data["attack_card_id"].(string)
		idxs := elfElementalShotCandidateIndices(player, attackCardID)
		options := make([]model.PromptOption, 0, len(idxs))
		for _, idx := range idxs {
			card, ok := getElfPlayableCardByIndex(player, idx)
			if !ok {
				continue
			}
			// 手牌必须是法术牌，祝福不限制类型
			if idx < len(player.Hand) && card.Type != model.CardTypeMagic {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(card)), CardID: card.ID})
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: "【元素射击】请选择发动消耗（法术牌或祝福）：", Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand", CardFilter: "magic_or_elf_blessing", CancelPolicy: "abort"}}

	case "elf_elemental_shot_water_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【水之矢】请选择+1治疗目标：", data, false)
	case "elf_elemental_shot_earth_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【地之矢】请选择1点法术伤害目标：", data, false)
	case "elf_ritual_release_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【精灵密仪】你已无祝福，转正并请选择1名角色承受2点法术伤害：", data, false)
	}
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "elf_archer_elemental_shot_pick":
		return true, handleElfElementalShotPick(rt, selectionIndex, ctxData)

	case "elf_elemental_shot_water_target", "elf_elemental_shot_earth_target", "elf_ritual_release_target":
		return true, handleElfElementalShotOrRitualTarget(rt, selectionIndex, ctxData)
	}

	return false, nil
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "elf_archer_elemental_shot_pick" {
		return false, nil
	}

	rt.PopInterrupt()
	if user := rt.GetPlayers()[playerID]; user != nil {
		// 回合限定次数在 Execute 前已计数；取消发动时退回，本回合仍可再用。
		if user.TurnState.UsedSkillCounts["elf_elemental_shot"] > 0 {
			user.TurnState.UsedSkillCounts["elf_elemental_shot"]--
		}
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

func handleElfElementalShotPick(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	attackCardID, _ := ctxData["attack_card_id"].(string)
	candidates := elfElementalShotCandidateIndices(user, attackCardID)
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, candidates)
	if !ok {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	// 手牌必须是法术牌，祝福不限制类型
	if cardIdx < len(user.Hand) {
		card := user.Hand[cardIdx]
		if card.Type != model.CardTypeMagic {
			return fmt.Errorf("手牌区只能选择法术牌发动元素射击")
		}
	}
	selectedCard, ok := rt.GetPlayableCardByIndex(user, cardIdx)
	if !ok || selectedCard.ID == "" {
		return fmt.Errorf("无效的卡牌选项: %d", selectionIndex)
	}
	card, ok := rt.ConsumePlayableCardByCardID(userID, selectedCard.ID)
	if !ok {
		return fmt.Errorf("未找到卡牌: %s", selectedCard.ID)
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
		func() {
			defer rt.PoseChangeGuard()()
			leaveElfArcherRitualForm(user)
		}()
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
	// 用 ClearForm 同时恢复直立，配合 PoseChangeGuard 派发姿态变化。
	return engineplayer.ClearForm(player, model.FormElfArcherRitual)
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

func getElfPlayableCardByIndex(player *model.Player, idx int) (model.Card, bool) {
	if player == nil || idx < 0 {
		return model.Card{}, false
	}
	if idx < len(player.Hand) {
		return player.Hand[idx], true
	}
	offset := idx - len(player.Hand)
	covers := elfBlessingCards(player)
	if offset < 0 || offset >= len(covers) {
		return model.Card{}, false
	}
	return covers[offset], true
}

func elfElementalShotCandidateIndices(player *model.Player, excludeCardIDs ...string) []int {
	if player == nil {
		return nil
	}
	exclude := make(map[string]bool, len(excludeCardIDs))
	for _, id := range excludeCardIDs {
		if id != "" {
			exclude[id] = true
		}
	}
	var out []int
	// 手牌：只有法术牌可用于元素射击
	for i, card := range player.Hand {
		if card.Type == model.CardTypeMagic && !exclude[card.ID] {
			out = append(out, i)
		}
	}
	// 祝福：所有祝福都可用于元素射击（不限制类型），排除攻击牌
	offset := len(player.Hand)
	blessings := elfBlessingCards(player)
	for i, card := range blessings {
		if !exclude[card.ID] {
			out = append(out, offset+i)
		}
	}
	sort.Ints(out)
	return out
}

var _ engineplayer.CancelChoiceHandler = choiceHandler{}
