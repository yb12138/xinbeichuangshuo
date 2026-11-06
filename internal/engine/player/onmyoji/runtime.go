package onmyoji

import (
	"fmt"
	"strings"

	promptfmt "starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// darkRitualTargetIDs 黑暗祭礼的候选目标：文案为"目标角色"，场上任意角色（含自己）。
func darkRitualTargetIDs(order []string, exists func(string) bool) []string {
	var ids []string
	for _, pid := range order {
		if exists(pid) {
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
	targetIDs := darkRitualTargetIDs(rt.GetPlayerOrder(), func(pid string) bool { return rt.GetPlayers()[pid] != nil })
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
// attackerID 为被应战的原攻击者，作为阴阳转换 skill_activated 的目标。
func ApplyFactionCounterBonuses(rt engineplayer.ChoiceRuntime, actor *model.Player, card *model.Card, attackerID string) {
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
	var targetIDs []string
	if attackerID != "" {
		targetIDs = []string{attackerID}
	}
	engineplayer.NotifySkillActivated(rt, actor, actor.ID, "onmyoji_yinyang_shift", targetIDs...)

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
		rt.WithSkillCause("onmyoji_shikigami_shift", actor.ID, func() {
			handler.Execute(ctx)
		})
	}
}

// CanPayBindingCost 检查阵营是否有资源支付式神咒束代价：1宝石 + 1星石（星石可用宝石或水晶）。
func CanPayBindingCost(rt engineplayer.ChoiceRuntime, camp model.Camp) bool {
	gems := rt.GetCampGems(string(camp))
	crystals := rt.GetCampCrystals(string(camp))
	return gems >= 2 || (gems >= 1 && crystals >= 1)
}

// PayBindingCost 从战绩区扣除式神咒束代价：1宝石 + 1星石（优先用水晶，没有水晶时用宝石）。
func PayBindingCost(rt engineplayer.ChoiceRuntime, camp model.Camp) bool {
	if !CanPayBindingCost(rt, camp) {
		return false
	}
	rt.ModifyGem(string(camp), -1)
	if rt.GetCampCrystals(string(camp)) >= 1 {
		rt.ModifyCrystal(string(camp), -1)
	} else {
		rt.ModifyGem(string(camp), -1)
	}
	return true
}

// CollectCounterOptions 收集阴阳师可用的应战牌选项。
func CollectCounterOptions(p *model.Player, incoming *model.Card) []map[string]interface{} {
	if p == nil || incoming == nil {
		return nil
	}
	var options []map[string]interface{}
	for _, c := range p.Hand {
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
		label := counterOptionLabel(c, useFaction)
		options = append(options, map[string]interface{}{
			"card_id":     c.ID,
			"use_faction": useFaction,
			"label":       label,
		})
	}
	return options
}

// CollectYinYangOptions 收集可用于【阴阳转换】的牌：与攻击牌命格相同的攻击牌（不论系别，暗灭除外）。
// 同系同命格的牌既可普通应战，也可在此发动阴阳转换；是否发动由玩家在确认框里选择，
// 不在确认框里发动时走普通应战，不会被强制套上阴阳转换。
func CollectYinYangOptions(p *model.Player, incoming *model.Card) []map[string]interface{} {
	if p == nil || incoming == nil || !CanUseFactionCounter(incoming) {
		return nil
	}
	var options []map[string]interface{}
	for _, c := range p.Hand {
		if c.Type != model.CardTypeAttack || c.Element == model.ElementDark {
			continue
		}
		if c.Faction == "" || c.Faction != incoming.Faction {
			continue
		}
		options = append(options, map[string]interface{}{
			"card_id":     c.ID,
			"use_faction": true,
			"label":       counterOptionLabel(c, true),
		})
	}
	return options
}

// counterOptionLabel 应战选牌的短标签："水涟斩 · 水系 · 咏命格"；按命格应战时注明攻击会转系。
// 不用 promptfmt.FormatCardInfo：那是带类型 / 伤害 / 独有技的调试格式，放进选牌面板太长。
func counterOptionLabel(c model.Card, useFaction bool) string {
	parts := []string{c.Name}
	if element := promptfmt.ElementName(string(c.Element)); element != "" {
		parts = append(parts, element+"系")
	}
	if c.Faction != "" {
		parts = append(parts, c.Faction+"命格")
	}
	label := strings.Join(parts, " · ")
	if useFaction {
		label += "（阴阳转换：转为" + promptfmt.ElementName(string(c.Element)) + "系）"
	}
	return label
}
