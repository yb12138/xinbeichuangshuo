// gameflow: 格斗家角色选择流。

package fighter

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

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	// Fighter choices use target prompts built elsewhere (via ChoiceRouteTargetPrompt)
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "fighter_psi_bullet_target":
		return true, handleFighterPsiBulletTargetChoice(rt, selectionIndex, ctxData)
	case "fighter_hundred_dragon_target":
		return true, handleFighterHundredDragonTargetChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func resolveFighterChoiceTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) (*model.Player, *model.Player, string, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return nil, nil, "", fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return nil, nil, "", fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := rt.LookupPlayer(targetID)
	if target == nil {
		return nil, nil, "", fmt.Errorf("目标不存在")
	}
	return user, target, targetID, nil
}

func handleFighterPsiBulletTargetChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	user, target, targetID, err := resolveFighterChoiceTarget(rt, selectionIndex, ctxData)
	if err != nil {
		return err
	}
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   targetID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	selfDamage := 0
	if target.Heal <= 0 {
		selfDamage = user.Tokens["fighter_qi"]
		if selfDamage > 0 {
			rt.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   user.ID,
				Damage:     selfDamage,
				DamageType: model.MagicAttack,
			})
		}
	}
	if selfDamage > 0 {
		rt.Log(fmt.Sprintf("%s 的 [念弹] 生效：对 %s 造成1点法术伤害；目标治疗为0，自己额外承受%d点法术伤害", user.Name, target.Name, selfDamage))
	} else {
		rt.Log(fmt.Sprintf("%s 的 [念弹] 生效：对 %s 造成1点法术伤害", user.Name, target.Name))
	}
	rt.PopInterrupt()
	if !rt.HasPendingInterrupt() {
		rt.RoutePendingDamageOr(model.TurnStageExtraAction, func() {
			rt.EnterExtraActionStage()
		})
	}
	return nil
}

func handleFighterHundredDragonTargetChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	user, target, targetID, err := resolveFighterChoiceTarget(rt, selectionIndex, ctxData)
	if err != nil {
		return err
	}
	targetOrder := rt.PlayerOrderPosition(targetID)
	if targetOrder <= 0 {
		return fmt.Errorf("目标不存在")
	}
	engineplayer.EnsurePlayerSkillFlowState(user)
	user.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"] = targetOrder
	rt.Log(fmt.Sprintf("%s 的 [百式幻龙拳] 锁定目标：%s", user.Name, target.Name))
	rt.PopInterrupt()
	if !rt.HasPendingInterrupt() {
		// 规则：百式幻龙拳的"锁定目标"仅是中间步骤，结算后必须回到 waiting_phase 指定的行动窗口。
		rt.ApplyChoiceResumePoint(mustChoiceResumePointFromMap(ctxData, "waiting_phase"))
	}
	return nil
}

func ensurePlayerTokensMap(player *model.Player) {
	if player != nil && player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
}

func mustChoiceResumePointFromMap(data map[string]interface{}, key string) interface{} {
	if data == nil {
		panic(fmt.Sprintf("missing resume point map for key %q", key))
	}
	raw, ok := data[key]
	if !ok {
		panic(fmt.Sprintf("missing resume point key %q", key))
	}
	return mustChoiceResumePoint(raw, key)
}

func mustChoiceResumePoint(raw interface{}, key string) interface{} {
	if raw == nil {
		panic(fmt.Sprintf("nil resume point for key %q", key))
	}
	return raw
}
