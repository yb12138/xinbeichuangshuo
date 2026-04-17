// gameflow: 剑帝角色选择流。

package sword_emperor

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(_ engineplayer.ChoiceRuntime, choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "se_sword_qi_slash_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			maxX = 1
		}
		options := make([]model.PromptOption, 0, maxX)
		for xValue := 1; xValue <= maxX; xValue++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", xValue), Label: fmt.Sprintf("移除%d点剑气，对另一名角色造成%d点法术伤害", xValue, xValue)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【剑气斩】请选择X值：", Options: options, Min: 1, Max: 1}
	}
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "se_sword_qi_slash_x":
		return true, handleSwordEmperorSwordQiSlashXChoice(rt, selectionIndex, ctxData)
	case "se_sword_qi_slash_target":
		return true, handleSwordEmperorSwordQiSlashTargetChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleSwordEmperorSwordQiSlashXChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	if maxX <= 0 {
		return fmt.Errorf("剑气斩没有可选X值")
	}
	if selectionIndex < 0 || selectionIndex >= maxX {
		return fmt.Errorf("无效的X值选项: %d", selectionIndex)
	}
	xValue := selectionIndex + 1
	ctxData["x_value"] = xValue
	ctxData["choice_type"] = "se_sword_qi_slash_target"
	if err := rt.ReplacePendingInterruptContext(ctxData); err != nil {
		return err
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleSwordEmperorSwordQiSlashTargetChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := rt.LookupPlayer(targetID)
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	if xValue <= 0 {
		return fmt.Errorf("剑气斩的X值无效")
	}
	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	if rawCtx != nil && rawCtx.EventCtx != nil && rawCtx.EventCtx.TargetID == targetID {
		return fmt.Errorf("剑气斩不能选择当前攻击目标")
	}

	nowQi := addSwordEmperorSwordQi(user, -xValue)
	rt.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: xValue, DamageType: model.MagicAttack})
	rt.Log(fmt.Sprintf("%s 发动 [剑气斩]：移除%d点剑气（当前%d），对 %s 造成%d点法术伤害", user.Name, xValue, nowQi, target.Name, xValue))
	rt.PopInterrupt()
	if !rt.HasPendingInterrupt() {
		rt.ResumePendingAttackHit(ctxData)
	}
	return nil
}

func swordEmperorSwordQi(player *model.Player) int {
	if player == nil || player.Tokens == nil {
		return 0
	}
	return player.Tokens["se_sword_qi"]
}

func addSwordEmperorSwordQi(player *model.Player, delta int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	newVal := player.Tokens["se_sword_qi"] + delta
	if newVal < 0 {
		newVal = 0
	}
	player.Tokens["se_sword_qi"] = newVal
	return newVal
}
