package engine

import (
	"fmt"
	"starcup-engine/internal/engine/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) handleFighterChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"fighter_psi_bullet_target":     e.handleFighterPsiBulletTargetChoice,
		"fighter_hundred_dragon_target": e.handleFighterHundredDragonTargetChoice,
	})
}

func (e *GameEngine) resolveFighterChoiceTarget(selectionIndex int, ctxData map[string]interface{}) (*model.Player, *model.Player, string, error) {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return nil, nil, "", fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return nil, nil, "", fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := e.State.Players[targetID]
	if target == nil {
		return nil, nil, "", fmt.Errorf("目标不存在")
	}
	return user, target, targetID, nil
}

func (e *GameEngine) handleFighterPsiBulletTargetChoice(selectionIndex int, ctxData map[string]interface{}) error {
	user, target, targetID, err := e.resolveFighterChoiceTarget(selectionIndex, ctxData)
	if err != nil {
		return err
	}
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   targetID,
		Damage:     1,
		DamageType: "magic",
	})
	selfDamage := 0
	if target.Heal <= 0 {
		selfDamage = user.Tokens["fighter_qi"]
		if selfDamage > 0 {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   user.ID,
				Damage:     selfDamage,
				DamageType: "magic",
			})
		}
	}
	if selfDamage > 0 {
		e.Log(fmt.Sprintf("%s 的 [念弹] 生效：对 %s 造成1点法术伤害；目标治疗为0，自己额外承受%d点法术伤害", user.Name, target.Name, selfDamage))
	} else {
		e.Log(fmt.Sprintf("%s 的 [念弹] 生效：对 %s 造成1点法术伤害", user.Name, target.Name))
	}
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.routePendingDamageOr(model.TurnStageExtraAction, func() {
			e.enterExtraActionStage()
		})
	}
	return nil
}

func (e *GameEngine) handleFighterHundredDragonTargetChoice(selectionIndex int, ctxData map[string]interface{}) error {
	user, target, targetID, err := e.resolveFighterChoiceTarget(selectionIndex, ctxData)
	if err != nil {
		return err
	}
	targetOrder := e.playerOrderPosition(targetID)
	if targetOrder <= 0 {
		return fmt.Errorf("目标不存在")
	}
	ensurePlayerTokensMap(user)
	user.Tokens["fighter_hundred_dragon_target_order"] = targetOrder
	e.Log(fmt.Sprintf("%s 的 [百式幻龙拳] 锁定目标：%s", user.Name, target.Name))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		// 规则：百式幻龙拳的“锁定目标”仅是中间步骤，结算后必须回到 waiting_phase 指定的行动窗口。
		e.applyChoiceResumePoint(mustChoiceResumePointFromMap(ctxData, "waiting_phase"))
	}
	return nil
}
