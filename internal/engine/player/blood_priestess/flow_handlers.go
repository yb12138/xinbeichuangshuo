// gameflow: 血之巫女 FlowContinuation 处理函数。

package blood_priestess

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// handleSharedLifeAfterDraw 处理 after_draw 流程边界：摸牌后放置同生共死效果卡。
func handleSharedLifeAfterDraw(rt engineplayer.ChoiceRuntime, cont model.FlowContinuation) error {
	user := rt.LookupPlayer(cont.PlayerID)
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", cont.PlayerID)
	}
	if !engineplayer.IsCharacter(user, "blood_priestess") {
		return fmt.Errorf("仅血之巫女可执行同生共死后续")
	}
	if len(cont.TargetIDs) != 1 {
		return fmt.Errorf("同生共死后续目标数量错误: %d", len(cont.TargetIDs))
	}
	target := rt.LookupPlayer(cont.TargetIDs[0])
	if target == nil {
		return fmt.Errorf("同生共死目标不存在: %s", cont.TargetIDs[0])
	}

	var card model.Card
	if cont.Data != nil {
		if v, ok := cont.Data["card"]; ok {
			switch c := v.(type) {
			case model.Card:
				card = c
			case *model.Card:
				if c != nil {
					card = *c
				}
			}
		}
	}
	if card.ID == "" || card.Name == "" {
		return fmt.Errorf("同生共死后续缺少原始专属卡")
	}

	if err := rt.AttachEffectCard(user, target, model.EffectBloodSharedLife, card); err != nil {
		user.RestoreExclusiveCard(card)
		return err
	}
	rt.Log(fmt.Sprintf("%s 的 [同生共死] 生效：放置于 %s 面前", user.Name, target.Name))

	rt.CheckHandLimit(user.ID, false)
	if target.ID != user.ID {
		rt.CheckHandLimit(target.ID, false)
	}
	return nil
}

// handleBloodSorrowAfterDamage 处理 after_damage 流程边界：伤害结算后移除/转移同生共死。
func handleBloodSorrowAfterDamage(rt engineplayer.ChoiceRuntime, cont model.FlowContinuation) error {
	user := rt.LookupPlayer(cont.PlayerID)
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", cont.PlayerID)
	}
	mode := ""
	if cont.Data != nil {
		if raw, ok := cont.Data["mode"].(string); ok {
			mode = raw
		}
	}
	if mode == "" {
		return fmt.Errorf("血之哀伤后续模式缺失")
	}
	checked := map[string]bool{}
	checkCap := func(player *model.Player) {
		if player == nil || checked[player.ID] {
			return
		}
		checked[player.ID] = true
		rt.CheckHandLimit(player.ID, false)
	}

	switch mode {
	case "remove":
		if !rt.RemoveEffectCard(user, model.EffectBloodSharedLife, true) {
			return fmt.Errorf("当前没有可移除的同生共死")
		}
		rt.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：移除【同生共死】", user.Name))
		checkCap(user)
		return nil
	case "transfer":
		targetID := ""
		if cont.Data != nil {
			if raw, ok := cont.Data["target_id"].(string); ok {
				targetID = raw
			}
		}
		target := rt.LookupPlayer(targetID)
		if target == nil {
			return fmt.Errorf("转移目标不存在: %s", targetID)
		}
		holder, card, ok := rt.DetachEffectCard(user, model.EffectBloodSharedLife)
		if !ok {
			return fmt.Errorf("当前没有可转移的同生共死")
		}
		if err := rt.AttachEffectCard(user, target, model.EffectBloodSharedLife, card); err != nil {
			if holder != nil {
				_ = rt.AttachEffectCard(user, holder, model.EffectBloodSharedLife, card)
			}
			return err
		}
		rt.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：将【同生共死】转移至 %s", user.Name, target.Name))
		checkCap(user)
		checkCap(holder)
		checkCap(target)
		return nil
	default:
		return fmt.Errorf("未知的血之哀伤后续模式: %s", mode)
	}
}

// handleBloodPriestessAfterDamage 统一的 after_damage handler，根据 SkillID 分支。
// - SkillID == "bp_blood_curse" → 执行血之诅咒弃牌逻辑
// - SkillID == "bp_blood_sorrow" → 执行血之哀伤移除/转移逻辑（从 Data["mode"] 判断）
func handleBloodPriestessAfterDamage(rt engineplayer.ChoiceRuntime, cont model.FlowContinuation) error {
	switch cont.SkillID {
	case "bp_blood_curse":
		// 血之诅咒：伤害结算后推送弃牌选择
		user := rt.LookupPlayer(cont.PlayerID)
		if user == nil {
			return fmt.Errorf("执行者不存在: %s", cont.PlayerID)
		}
		discardNeed := 3
		if len(user.Hand) < discardNeed {
			discardNeed = len(user.Hand)
		}
		if discardNeed <= 0 {
			rt.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：无需弃牌", user.Name))
			return nil
		}
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: func() map[string]any {
				data := map[string]any{
					"choice_type":     "bp_curse_discard",
					"discard_subflow": true,
					"user_id":         user.ID,
					"discard_count":   discardNeed,
				}
				model.SetPromptFlowContext(data, initBloodCurseDiscardFlow(discardNeed))
				return data
			}(),
		})
		rt.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：伤害结算完成，请弃置%d张牌", user.Name, discardNeed))
		return nil

	case "bp_blood_sorrow":
		// 血之哀伤：移除/转移同生共死
		return handleBloodSorrowApplyInternal(rt, cont)

	default:
		return fmt.Errorf("未知的 after_damage continuation SkillID: %s", cont.SkillID)
	}
}

// handleBloodSorrowApplyInternal 原有血之哀伤逻辑（从 handleBloodSorrowAfterDamage 提取）。
func handleBloodSorrowApplyInternal(rt engineplayer.ChoiceRuntime, cont model.FlowContinuation) error {
	user := rt.LookupPlayer(cont.PlayerID)
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", cont.PlayerID)
	}
	mode := ""
	if cont.Data != nil {
		if raw, ok := cont.Data["mode"].(string); ok {
			mode = raw
		}
	}
	if mode == "" {
		return fmt.Errorf("血之哀伤后续模式缺失")
	}
	checked := map[string]bool{}
	checkCap := func(player *model.Player) {
		if player == nil || checked[player.ID] {
			return
		}
		checked[player.ID] = true
		rt.CheckHandLimit(player.ID, false)
	}

	switch mode {
	case "remove":
		if !rt.RemoveEffectCard(user, model.EffectBloodSharedLife, true) {
			return fmt.Errorf("当前没有可移除的同生共死")
		}
		rt.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：移除【同生共死】", user.Name))
		checkCap(user)
		return nil
	case "transfer":
		targetID := ""
		if cont.Data != nil {
			if raw, ok := cont.Data["target_id"].(string); ok {
				targetID = raw
			}
		}
		target := rt.LookupPlayer(targetID)
		if target == nil {
			return fmt.Errorf("转移目标不存在: %s", targetID)
		}
		holder, card, ok := rt.DetachEffectCard(user, model.EffectBloodSharedLife)
		if !ok {
			return fmt.Errorf("当前没有可转移的同生共死")
		}
		if err := rt.AttachEffectCard(user, target, model.EffectBloodSharedLife, card); err != nil {
			if holder != nil {
				_ = rt.AttachEffectCard(user, holder, model.EffectBloodSharedLife, card)
			}
			return err
		}
		rt.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：将【同生共死】转移至 %s", user.Name, target.Name))
		checkCap(user)
		checkCap(holder)
		checkCap(target)
		return nil
	default:
		return fmt.Errorf("未知的血之哀伤后续模式: %s", mode)
	}
}
