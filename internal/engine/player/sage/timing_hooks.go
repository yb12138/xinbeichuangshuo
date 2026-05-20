// gameflow: 贤者 Timing Hook 实现。

package sage

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// isMagicDamage 内联法术伤害判定（等价于 runtimeutil.IsMagicDamageType）。
func isMagicDamage(dt model.DamageType) bool {
	return !strings.EqualFold(string(dt), string(model.AttackDamage))
}

// postDamageResolvedHook 智慧法典 + 法术反弹。
func postDamageResolvedHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	if ctx.Damage <= 0 || !isMagicDamage(ctx.DamageType) {
		return player.TimingHookResult{}
	}
	target := rt.GetPlayer(ctx.TargetID)
	if target == nil || !player.IsCharacter(target, "sage") {
		return player.TimingHookResult{}
	}
	// Wisdom Codex: damage > 3
	if ctx.Damage > 3 {
		maxEnergy := rt.GetPlayerEnergyCap(target)
		if target.Gem+target.Crystal < maxEnergy {
			room := maxEnergy - (target.Gem + target.Crystal)
			gain := 2
			if gain > room {
				gain = room
			}
			target.Gem += gain
			if gain > 0 {
				rt.Log(fmt.Sprintf("%s 的 [智慧法典] 触发：获得%d点红宝石", target.Name, gain))
			}
		} else {
			rt.Log(fmt.Sprintf("%s 的 [智慧法典] 触发：能量已满，红宝石未增加", target.Name))
		}
		if len(target.Hand) > 0 {
			rt.PushDiscardChoiceInterrupt(target.ID, map[string]interface{}{
				"discard_count":        1,
				"stay_in_turn":         true,
				"is_damage_resolution": true,
				"prompt":               "【智慧法典】请选择弃置1张手牌：",
			})
			return player.TimingHookResult{Interrupted: true}
		}
	}
	// Magic Rebound: damage == 1
	if ctx.Damage == 1 {
		sameCount := player.MaxSameElementCount(target)
		if sameCount >= 2 {
			rt.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: target.ID,
				Context: map[string]interface{}{
					"choice_type":              "sage_magic_rebound_confirm",
					"user_id":                  target.ID,
					model.PromptFlowContextKey: sageMagicReboundFlowRuntime.Begin(),
				},
			})
			rt.Log(fmt.Sprintf("%s 的 [法术反弹] 可触发：承受1点法术伤害，最大同系手牌=%d", target.Name, sameCount))
			return player.TimingHookResult{Interrupted: true}
		}
		rt.Log(fmt.Sprintf("%s 的 [法术反弹] 未触发：承受1点法术伤害但同系手牌不足2（当前最大同系=%d）", target.Name, sameCount))
	}
	return player.TimingHookResult{}
}
