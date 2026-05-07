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
	switch choiceType {
	case "fighter_psi_bullet_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【念弹】请选择1名目标对手：", data, false)
	case "fighter_hundred_dragon_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【百式幻龙拳】请选择本行动阶段锁定的目标角色：", data, false)
	default:
		return nil
	}
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
	user := rt.GetPlayers()[userID]
	if user == nil {
		return nil, nil, "", fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return nil, nil, "", fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := rt.GetPlayers()[targetID]
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
	if rt.GetPendingInterrupt() == nil {
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
	targetOrder := playerOrderPosition(rt, targetID)
	if targetOrder <= 0 {
		return fmt.Errorf("目标不存在")
	}
	engineplayer.EnsurePlayerSkillFlowState(user)
	user.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"] = targetOrder
	rt.Log(fmt.Sprintf("%s 的 [百式幻龙拳] 锁定目标：%s", user.Name, target.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		// 规则：百式幻龙拳的"锁定目标"仅是中间步骤，结算后必须回到 waiting_phase 指定的行动窗口。
		rt.ApplyChoiceResumePoint(engineplayer.MustChoiceResumePointFromMap(ctxData, "waiting_phase"))
	}
	return nil
}

