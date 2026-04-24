// gameflow: 血之巫女延迟后续处理。

package blood_priestess

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// FollowupSpecs 导出角色延迟后续声明。
func FollowupSpecs() map[string]engineplayer.FollowupSpec {
	return map[string]engineplayer.FollowupSpec{
		"blood_priestess_shared_life_place":  {Label: "BloodPriestess", Resolve: resolveSharedLifePlace},
		"blood_priestess_blood_sorrow_apply": {Label: "BloodPriestess", Resolve: resolveBloodSorrowApply},
		"blood_priestess_curse_discard":      {Label: "BloodPriestess", Resolve: resolveCurseDiscard},
	}
}

func resolveSharedLifePlace(rt engineplayer.ChoiceRuntime, f model.DeferredFollowup) error {
	user := rt.LookupPlayer(f.UserID)
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	if !engineplayer.IsCharacter(user, "blood_priestess") {
		return fmt.Errorf("仅血之巫女可执行同生共死后续")
	}
	if len(f.TargetIDs) != 1 {
		return fmt.Errorf("同生共死后续目标数量错误: %d", len(f.TargetIDs))
	}
	target := rt.LookupPlayer(f.TargetIDs[0])
	if target == nil {
		return fmt.Errorf("同生共死目标不存在: %s", f.TargetIDs[0])
	}

	var card model.Card
	if f.Data != nil {
		if v, ok := f.Data["card"]; ok {
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

	if err := rt.AttachExclusiveEffectCard(user.ID, target.ID, model.EffectBloodSharedLife, card); err != nil {
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

func resolveBloodSorrowApply(rt engineplayer.ChoiceRuntime, f model.DeferredFollowup) error {
	user := rt.LookupPlayer(f.UserID)
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	mode := ""
	if f.Data != nil {
		if raw, ok := f.Data["mode"].(string); ok {
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
		if !rt.RemoveExclusiveEffectCard(user, model.EffectBloodSharedLife, true) {
			return fmt.Errorf("当前没有可移除的同生共死")
		}
		rt.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：移除【同生共死】", user.Name))
		checkCap(user)
		return nil
	case "transfer":
		targetID := ""
		if f.Data != nil {
			if raw, ok := f.Data["target_id"].(string); ok {
				targetID = raw
			}
		}
		target := rt.LookupPlayer(targetID)
		if target == nil {
			return fmt.Errorf("转移目标不存在: %s", targetID)
		}
		holder, card, ok := rt.DetachExclusiveEffectCard(user, model.EffectBloodSharedLife)
		if !ok {
			return fmt.Errorf("当前没有可转移的同生共死")
		}
		if err := rt.AttachExclusiveEffectCard(user.ID, target.ID, model.EffectBloodSharedLife, card); err != nil {
			if holder != nil {
				_ = rt.AttachExclusiveEffectCard(user.ID, holder.ID, model.EffectBloodSharedLife, card)
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

func resolveCurseDiscard(rt engineplayer.ChoiceRuntime, f model.DeferredFollowup) error {
	user := rt.LookupPlayer(f.UserID)
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
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
		Context: map[string]interface{}{
			"choice_type":   "bp_curse_discard",
			"user_id":       user.ID,
			"discard_count": discardNeed,
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：伤害结算完成，请弃置%d张牌", user.Name, discardNeed))
	return nil
}
