// gameflow: 灵魂术士 Timing Hook 实现。

package soul_sorcerer

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// damageBeforeTakenHook 灵魂链接转伤：在承伤触发前检查是否可转移伤害。
func damageBeforeTakenHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	pd := ctx.PendingDamage
	if pd == nil || pd.Damage <= 0 ||
		pd.HasCheck(model.PendingDamageCheckFromSoulLink) ||
		pd.HasCheck(model.PendingDamageCheckSoulLinkChecked) {
		return player.TimingHookResult{}
	}
	pd.SetCheck(model.PendingDamageCheckSoulLinkChecked, true)

	target := rt.GetPlayer(ctx.TargetID)
	if target == nil {
		return player.TimingHookResult{}
	}

	var sorcerer *model.Player
	var counterpart *model.Player
	// 场景1：灵魂术士本人受伤，另一方是其链接队友。
	if rt.IsCharacter(target, "soul_sorcerer") {
		holder, _ := rt.FindExclusiveEffectCard(target, model.EffectSoulLink)
		if holder != nil {
			sorcerer = target
			counterpart = holder
		}
	} else {
		// 场景2：链接队友受伤，寻找来源为该灵魂术士的链接牌。
		for _, fc := range target.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectSoulLink {
				continue
			}
			p := rt.GetPlayer(fc.SourceID)
			if p == nil || !rt.IsCharacter(p, "soul_sorcerer") {
				continue
			}
			sorcerer = p
			counterpart = p
			break
		}
	}
	if sorcerer == nil || counterpart == nil {
		return player.TimingHookResult{}
	}

	blue := BlueSoul(sorcerer)
	if blue <= 0 {
		return player.TimingHookResult{}
	}
	maxX := pd.Damage
	if blue < maxX {
		maxX = blue
	}
	if maxX <= 0 {
		return player.TimingHookResult{}
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: sorcerer.ID,
		Context: map[string]interface{}{
			"choice_type":     "ss_link_transfer_x",
			"sorcerer_id":     sorcerer.ID,
			"damage_index":    0,
			"source_id":       pd.SourceID,
			"target_id":       pd.TargetID,
			"counterpart_id":  counterpart.ID,
			"max_x":           maxX,
			"original_damage": pd.Damage,
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [灵魂链接] 可触发：是否移除蓝色灵魂转移伤害（最多%d）", sorcerer.Name, maxX))
	return player.TimingHookResult{Interrupted: true}
}
