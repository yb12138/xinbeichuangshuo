// gameflow: 精灵射手 Timing Hook 实现（PostDamageResolved）。

package elf_archer

import (
	"fmt"
	"strings"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// elfPostDamageResolvedFlag 标记动物伙伴/宠物强化的⑥时点上下文。
const elfPostDamageResolvedFlag = "elf_post_damage_resolved"

// postDamageResolvedHook 伤害结算完成后：动物伙伴响应。
func postDamageResolvedHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	players := rt.GetPlayers()
	source := players[ctx.SourceID]
	target := players[ctx.TargetID]
	if rt == nil || source == nil || target == nil || ctx.PendingDamage == nil {
		return engineplayer.TimingHookResult{}
	}
	if !engineplayer.IsCharacter(source, "elf_archer") || !source.IsActive {
		return engineplayer.TimingHookResult{}
	}
	pd := ctx.PendingDamage
	if pd.TargetID == "" || pd.TargetID == source.ID {
		return engineplayer.TimingHookResult{}
	}
	// 自己回合内造成的任意伤害（攻击/法术、主动/应战）均可触发，伤害为0不触发。
	if pd.Damage <= 0 {
		return engineplayer.TimingHookResult{}
	}

	damageVal := pd.Damage
	eventCtx := &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  pd.SourceID,
		TargetID:  pd.TargetID,
		DamageVal: &damageVal,
		Card:      pd.Card,
	}
	if strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
		eventCtx.AttackInfo = &model.AttackEventInfo{
			ActionType: "Attack",
			IsHit:      true,
		}
		if pd.IsCounter {
			eventCtx.AttackInfo.CounterInitiator = pd.SourceID
		}
	}
	cctx := rt.BuildContext(source, target, model.TimingDamageTaken, eventCtx)
	if pd.DamageType != model.AttackDamage {
		cctx.Flags["IsMagicDamage"] = true
	}
	// 标记为⑥"承受伤害后"时点，区别于伤害结算③时点的通用 DamageTaken 派发。
	cctx.Flags[elfPostDamageResolvedFlag] = true

	skillIDs := make([]string, 0, 2)
	for _, skillID := range []string{"elf_animal_companion", "elf_pet_empower"} {
		if rt.IsSkillStillUsable(skillID, source, cctx) {
			skillIDs = append(skillIDs, skillID)
		}
	}
	if len(skillIDs) == 0 {
		return engineplayer.TimingHookResult{}
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: source.ID,
		SkillIDs: skillIDs,
		Context:  cctx,
	})
	rt.Log(fmt.Sprintf("%s 的 [动物伙伴] 响应窗口开启", source.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}
