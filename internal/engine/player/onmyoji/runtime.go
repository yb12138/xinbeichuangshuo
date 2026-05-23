package onmyoji

import (
	"fmt"

	promptfmt "starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// campEnemyIDs 返回与指定阵营不同的所有玩家ID。
func campEnemyIDs(rt engineplayer.ChoiceRuntime, camp model.Camp) []string {
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		if p := rt.GetPlayers()[pid]; p != nil && p.Camp != camp {
			ids = append(ids, pid)
		}
	}
	return ids
}

// MaybeDarkRitual 检查阴阳师是否满足黑暗祭礼的发动条件。
func MaybeDarkRitual(rt engineplayer.ChoiceRuntime, player *model.Player) bool {
	if player == nil || !engineplayer.IsCharacter(player, "onmyoji") || player.Tokens == nil || player.Tokens["onmyoji_ghost_fire"] < 3 {
		return false
	}
	targetIDs := campEnemyIDs(rt, player.Camp)
	if len(targetIDs) == 0 {
		return false
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "onmyoji_dark_ritual_target",
			"user_id":     player.ID,
			"target_ids":  targetIDs,
			"ghost_fire":  player.Tokens["onmyoji_ghost_fire"],
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [黑暗祭礼] 触发，等待选择2点法术伤害目标", player.Name))
	return true
}

// ApplyFactionCounterBonuses 应用阴阳转换与式神转换效果。
// 阴阳转换：鬼火+1；若处于式神形态则脱离式神形态；攻击伤害=当前鬼火数。
// 式神转换：阴阳转换生效时自动触发，强制摸1张牌，然后+1鬼火。
func ApplyFactionCounterBonuses(rt engineplayer.ChoiceRuntime, actor *model.Player, card *model.Card) {
	if actor == nil || card == nil {
		return
	}
	defer rt.PoseChangeGuard()

	// 阴阳转换：鬼火+1
	if actor.Tokens == nil {
		actor.Tokens = map[string]int{}
	}
	actor.Tokens["onmyoji_ghost_fire"]++
	if actor.Tokens["onmyoji_ghost_fire"] > 3 {
		actor.Tokens["onmyoji_ghost_fire"] = 3
	}
	rt.Log(fmt.Sprintf("%s 的 [阴阳转换] 触发，鬼火+1", actor.Name))

	// 式神转换：阴阳转换生效时自动触发（无条件）
	triggerShikigamiShift(rt, actor)

	// 阴阳转换：若处于式神形态则脱离式神形态
	if InShikigamiForm(actor) {
		LeaveShikigamiForm(actor)
		rt.Log(fmt.Sprintf("%s 的 [阴阳转换] 效果：脱离式神形态", actor.Name))
	}

	// 阴阳转换：攻击伤害=当前鬼火数
	card.Damage = actor.Tokens["onmyoji_ghost_fire"]
	if card.Damage < 0 {
		card.Damage = 0
	}
}

// triggerShikigamiShift 通过正式技能 handler 触发式神转换。
func triggerShikigamiShift(rt engineplayer.ChoiceRuntime, actor *model.Player) {
	ctx := rt.BuildContext(actor, nil, model.TimingActionDuring, nil)
	ctx.Flags["yinyang_counter_active"] = true
	handler := &OnmyojiShikigamiShiftHandler{}
	if handler.CanUse(ctx) {
		handler.Execute(ctx)
	}
}

// CanPayBindingCost 检查阵营是否有资源支付式神咒束代价。
func CanPayBindingCost(rt engineplayer.ChoiceRuntime, camp model.Camp) bool {
	gems := rt.GetCampGems(string(camp))
	crystals := rt.GetCampCrystals(string(camp))
	return gems >= 1 && crystals >= 1
}

// CollectCounterOptions 收集阴阳师可用的应战牌选项。
func CollectCounterOptions(p *model.Player, incoming *model.Card) []map[string]interface{} {
	if p == nil || incoming == nil {
		return nil
	}
	var options []map[string]interface{}
	for i, c := range p.Hand {
		if c.Type != model.CardTypeAttack {
			continue
		}
		useFaction := false
		canCounter := false
		if c.Element == incoming.Element || c.Element == model.ElementDark {
			canCounter = true
		}
		if !canCounter && CanUseFactionCounter(incoming) && c.Faction != "" && c.Faction == incoming.Faction {
			canCounter = true
			useFaction = true
		}
		if !canCounter {
			continue
		}
		label := fmt.Sprintf("%d: %s", i+1, promptfmt.FormatCardInfo(c))
		if useFaction {
			label += "（阴阳转换）"
		}
		options = append(options, map[string]interface{}{
			"card_id":     c.ID,
			"use_faction": useFaction,
			"label":       label,
		})
	}
	return options
}
