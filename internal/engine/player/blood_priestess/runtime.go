package blood_priestess

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// FindSharedLife 查找血祭司的同生共死场牌。
func FindSharedLife(rt engineplayer.ChoiceRuntime, priestess *model.Player) (*model.Player, *model.FieldCard) {
	return rt.FindExclusiveEffectCard(priestess, model.EffectBloodSharedLife)
}

// DetachSharedLife 分离血祭司的同生共死场牌。
func DetachSharedLife(rt engineplayer.ChoiceRuntime, priestess *model.Player) (*model.Player, model.Card, bool) {
	return rt.DetachExclusiveEffectCard(priestess, model.EffectBloodSharedLife)
}

// RemoveSharedLife 移除血祭司的同生共死场牌。
func RemoveSharedLife(rt engineplayer.ChoiceRuntime, priestess *model.Player, restoreCard bool) bool {
	return rt.RemoveExclusiveEffectCard(priestess, model.EffectBloodSharedLife, restoreCard)
}

// PlaceSharedLife 放置血祭司的同生共死场牌。
func PlaceSharedLife(rt engineplayer.ChoiceRuntime, priestess, target *model.Player, card model.Card) error {
	if priestess == nil || target == nil {
		return fmt.Errorf("放置同生共死时角色不存在")
	}
	return rt.AttachExclusiveEffectCard(priestess.ID, target.ID, model.EffectBloodSharedLife, card)
}

// HasFixedMaxHandCap 判断玩家是否有固定手牌上限。
func HasFixedMaxHandCap(rt engineplayer.ChoiceRuntime, player *model.Player) bool {
	if player == nil {
		return false
	}
	if _, ok := rt.RoleFixedMaxHandCapValue(player); ok {
		return true
	}
	return rt.HasMercyFixedMaxHandCap(player)
}

// SharedLifeDeltaFor 计算同生共死对指定玩家手牌上限的增量。
func SharedLifeDeltaFor(rt engineplayer.ChoiceRuntime, player *model.Player) int {
	if player == nil {
		return 0
	}
	delta := 0
	for _, pid := range rt.PlayerOrder() {
		holder := rt.LookupPlayer(pid)
		if holder == nil {
			continue
		}
		for _, fc := range holder.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
				continue
			}
			source := rt.LookupPlayer(fc.SourceID)
			if source == nil || !engineplayer.IsCharacter(source, "blood_priestess") {
				continue
			}
			change := -2
			if InBleedingForm(source) {
				change = 1
			}
			if source.ID == player.ID {
				delta += change
				continue
			}
			if fc.OwnerID == player.ID && !HasFixedMaxHandCap(rt, player) {
				delta += change
			}
		}
	}
	return delta
}

// EnterBleedingFormWithLog 进入流血形态并记录日志。
func EnterBleedingFormWithLog(rt engineplayer.ChoiceRuntime, player *model.Player, reason string) bool {
	if player == nil || !engineplayer.IsCharacter(player, "blood_priestess") {
		return false
	}
	if InBleedingForm(player) {
		return false
	}
	beforePoses := rt.SnapshotPlayerPoses()
	EnterBleedingForm(player)
	if reason == "" {
		reason = "因承受伤害导致我方士气下降"
	}
	rt.Log(fmt.Sprintf("%s 的 [流血] 触发：%s，进入流血形态", player.Name, reason))
	rt.DispatchOrientationChanges(beforePoses)
	return true
}

// LeaveBleedingFormWithLog 脱离流血形态并记录日志。
func LeaveBleedingFormWithLog(rt engineplayer.ChoiceRuntime, player *model.Player, reason string) bool {
	if player == nil || !engineplayer.IsCharacter(player, "blood_priestess") {
		return false
	}
	if !InBleedingForm(player) {
		return false
	}
	beforePoses := rt.SnapshotPlayerPoses()
	LeaveBleedingForm(player)
	if reason == "" {
		reason = "行动结束时手牌少于3"
	}
	rt.Log(fmt.Sprintf("%s 的 [流血·手牌不足脱离] 生效：%s，脱离流血形态", player.Name, reason))
	rt.DispatchOrientationChanges(beforePoses)
	return true
}

// ResolveBleedExitOnActionEnd 在行动结束时检查所有血祭司是否需要脱离流血形态。
func ResolveBleedExitOnActionEnd(rt engineplayer.ChoiceRuntime) bool {
	released := false
	for _, pid := range rt.PlayerOrder() {
		player := rt.LookupPlayer(pid)
		if player == nil || !engineplayer.IsCharacter(player, "blood_priestess") {
			continue
		}
		if len(player.Hand) >= 3 {
			continue
		}
		if LeaveBleedingFormWithLog(rt, player, "行动结束时手牌<3") {
			released = true
		}
	}
	return released
}
