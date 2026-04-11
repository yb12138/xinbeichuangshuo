// gameflow: 圣枪骑士：地枪、圣击等互斥与补结算。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) syncHolyLancerRevelationMaxHeal(player *model.Player) {
	if player == nil || !e.isHolyLancer(player) {
		return
	}
	enemyCamp := model.BlueCamp
	if player.Camp == model.BlueCamp {
		enemyCamp = model.RedCamp
	}
	maxHeal := 2
	if e.GetCampCups(string(player.Camp)) >= e.GetCampCups(string(enemyCamp)) {
		maxHeal = 3
	}
	player.MaxHeal = maxHeal
}

func syncHolyLancerDerivedStateOnPlayerSetup(e *GameEngine, player *model.Player) {
	e.syncHolyLancerRevelationMaxHeal(player)
}

func syncHolyLancerDerivedStateOnCampCupChanged(e *GameEngine, _ model.Camp) {
	e.refreshAllPlayerDerivedStates()
}

func (e *GameEngine) buildHolyLancerChoicePrompt(choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	if choiceType != "holy_lancer_earth_spear_x" {
		return nil
	}
	maxX := runtimeutil.ToIntContextValue(data["max_x"])
	options := make([]model.PromptOption, 0, maxX)
	for x := 1; x <= maxX; x++ {
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", x-1),
			Label: fmt.Sprintf("移除%d点治疗，本次伤害+%d", x, x),
		})
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【地枪】请选择X值：",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func (e *GameEngine) handleHolyLancerChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"holy_lancer_earth_spear_x": e.handleHolyLancerEarthSpearXChoice,
	})
}

func (e *GameEngine) handleHolyLancerEarthSpearXChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	x := selectionIndex + 1
	if x < 1 || x > maxX || x > user.Heal {
		return fmt.Errorf("无效的X值")
	}
	userCtx, ok := ctxData["user_ctx"].(*model.Context)
	if !ok || userCtx == nil || userCtx.EventCtx == nil || userCtx.EventCtx.DamageVal == nil {
		return fmt.Errorf("地枪上下文丢失")
	}
	user.Heal -= x
	*userCtx.EventCtx.DamageVal += x
	user.TurnState.UsedSkillCounts["holy_lancer_block_sacred_strike"] = 1
	e.Log(fmt.Sprintf("%s 发动 [地枪]，移除%d治疗，本次伤害+%d", user.Name, x, x))
	e.PopInterrupt()
	e.resumePendingAttackHit(ctxData)
	return nil
}
