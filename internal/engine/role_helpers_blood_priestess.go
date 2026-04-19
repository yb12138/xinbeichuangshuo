// gameflow: 血祭司：同生共死与流血形态辅助。
package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) findBloodPriestessSharedLife(priestess *model.Player) (*model.Player, *model.FieldCard) {
	return e.findExclusiveEffectCard(priestess, model.EffectBloodSharedLife)
}

func (e *GameEngine) detachBloodPriestessSharedLife(priestess *model.Player) (*model.Player, model.Card, bool) {
	return e.detachExclusiveEffectCard(priestess, model.EffectBloodSharedLife)
}

func (e *GameEngine) removeBloodPriestessSharedLife(priestess *model.Player, restoreCard bool) bool {
	return e.removeExclusiveEffectCard(priestess, model.EffectBloodSharedLife, restoreCard)
}

func (e *GameEngine) placeBloodPriestessSharedLife(priestess *model.Player, target *model.Player, card model.Card) error {
	if priestess == nil || target == nil {
		return fmt.Errorf("放置同生共死时角色不存在")
	}

	return e.attachExclusiveEffectCard(priestess, target, model.EffectBloodSharedLife, card)
}

func (e *GameEngine) hasFixedMaxHandCap(player *model.Player) bool {
	if player == nil {
		return false
	}
	if _, ok := e.roleFixedMaxHandCapValue(player); ok {
		return true
	}
	return e.hasMercyFixedMaxHandCap(player)
}

func (e *GameEngine) bloodPriestessSharedLifeDeltaFor(player *model.Player) int {
	if player == nil {
		return 0
	}
	delta := 0
	for _, pid := range e.State.PlayerOrder {
		holder := e.State.Players[pid]
		if holder == nil {
			continue
		}
		for _, fc := range holder.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
				continue
			}
			source := e.State.Players[fc.SourceID]
			if source == nil || !e.isBloodPriestess(source) {
				continue
			}
			change := -2
			if hasBloodPriestessBleedingForm(source) {
				change = 1
			}
			if source.ID == player.ID {
				delta += change
				continue
			}
			if fc.OwnerID == player.ID && !e.hasFixedMaxHandCap(player) {
				delta += change
			}
		}
	}
	return delta
}

func (e *GameEngine) enterBloodPriestessBleedingForm(player *model.Player, reason string) bool {
	if player == nil || !e.isBloodPriestess(player) {
		return false
	}
	if hasBloodPriestessBleedingForm(player) {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	enterBloodPriestessBleedingFormState(player)
	if reason == "" {
		reason = "因承受伤害导致我方士气下降"
	}
	e.Log(fmt.Sprintf("%s 的 [流血] 触发：%s，进入流血形态", player.Name, reason))
	e.dispatchOrientationChanges(beforePoses)
	return true
}

func (e *GameEngine) leaveBloodPriestessBleedingForm(player *model.Player, reason string) bool {
	if player == nil || !e.isBloodPriestess(player) {
		return false
	}
	if !hasBloodPriestessBleedingForm(player) {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	leaveBloodPriestessBleedingFormState(player)
	if reason == "" {
		reason = "行动结束时手牌少于3"
	}
	e.Log(fmt.Sprintf("%s 的 [流血·手牌不足脱离] 生效：%s，脱离流血形态", player.Name, reason))
	e.dispatchOrientationChanges(beforePoses)
	return true
}

func (e *GameEngine) resolveBloodPriestessBleedExitOnActionEnd() bool {
	if e == nil || e.State == nil {
		return false
	}
	released := false
	for _, pid := range e.State.PlayerOrder {
		player := e.State.Players[pid]
		if player == nil || !e.isBloodPriestess(player) {
			continue
		}
		if len(player.Hand) >= 3 {
			continue
		}
		if e.leaveBloodPriestessBleedingForm(player, "行动结束时手牌<3") {
			released = true
		}
	}
	return released
}

// gameflow: 血祭司：血咒、仪式等。

func buildBloodPriestessDeferredFollowupHandlers() map[string]deferredFollowupHandler {
	return map[string]deferredFollowupHandler{
		"blood_priestess_shared_life_place": {
			label:   "BloodPriestess",
			resolve: (*GameEngine).resolveBloodPriestessSharedLifePlaceFollowup,
		},
		"blood_priestess_blood_sorrow_apply": {
			label:   "BloodPriestess",
			resolve: (*GameEngine).resolveBloodPriestessBloodSorrowFollowup,
		},
		"blood_priestess_curse_discard": {
			label:   "BloodPriestess",
			resolve: (*GameEngine).resolveBloodPriestessCurseDiscardFollowup,
		},
	}
}

func (e *GameEngine) resolveBloodPriestessSharedLifePlaceFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	if !e.isBloodPriestess(user) {
		return fmt.Errorf("仅血之巫女可执行同生共死后续")
	}
	if len(f.TargetIDs) != 1 {
		return fmt.Errorf("同生共死后续目标数量错误: %d", len(f.TargetIDs))
	}
	target := e.State.Players[f.TargetIDs[0]]
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

	if err := e.placeBloodPriestessSharedLife(user, target, card); err != nil {
		user.RestoreExclusiveCard(card)
		return err
	}
	e.Log(fmt.Sprintf("%s 的 [同生共死] 生效：放置于 %s 面前", user.Name, target.Name))

	e.checkHandLimit(user, nil)
	if target.ID != user.ID {
		e.checkHandLimit(target, nil)
	}
	return nil
}

func (e *GameEngine) resolveBloodPriestessBloodSorrowFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
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
		e.checkHandLimit(player, nil)
	}

	switch mode {
	case "remove":
		if !e.removeBloodPriestessSharedLife(user, true) {
			return fmt.Errorf("当前没有可移除的同生共死")
		}
		e.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：移除【同生共死】", user.Name))
		checkCap(user)
		return nil
	case "transfer":
		targetID := ""
		if f.Data != nil {
			if raw, ok := f.Data["target_id"].(string); ok {
				targetID = raw
			}
		}
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("转移目标不存在: %s", targetID)
		}
		holder, card, ok := e.detachBloodPriestessSharedLife(user)
		if !ok {
			return fmt.Errorf("当前没有可转移的同生共死")
		}
		if err := e.attachExclusiveEffectCard(user, target, model.EffectBloodSharedLife, card); err != nil {
			if holder != nil {
				_ = e.attachExclusiveEffectCard(user, holder, model.EffectBloodSharedLife, card)
			}
			return err
		}
		e.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：将【同生共死】转移至 %s", user.Name, target.Name))
		checkCap(user)
		checkCap(holder)
		checkCap(target)
		return nil
	default:
		return fmt.Errorf("未知的血之哀伤后续模式: %s", mode)
	}
}

func (e *GameEngine) resolveBloodPriestessCurseDiscardFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	discardNeed := 3
	if len(user.Hand) < discardNeed {
		discardNeed = len(user.Hand)
	}
	if discardNeed <= 0 {
		e.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：无需弃牌", user.Name))
		return nil
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type":   "bp_curse_discard",
			"user_id":       user.ID,
			"discard_count": discardNeed,
		},
	})
	e.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：伤害结算完成，请弃置%d张牌", user.Name, discardNeed))
	return nil
}
