// gameflow: 精灵射手 Timing Hook 实现（PostDamageResolved）。

package elf_archer

import (
	"fmt"
	"strings"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// postDamageResolvedHook 伤害结算完成后：动物伙伴响应。
func postDamageResolvedHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	source := rt.LookupPlayer(ctx.SourceID)
	target := rt.LookupPlayer(ctx.TargetID)
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
	if !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) || pd.Card == nil || pd.IsCounter {
		return engineplayer.TimingHookResult{}
	}

	damageVal := pd.Damage
	cctx := rt.BuildContext(source, target, model.TimingOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  pd.SourceID,
		TargetID:  pd.TargetID,
		DamageVal: &damageVal,
		Card:      pd.Card,
		AttackInfo: &model.AttackEventInfo{
			ActionType: "Attack",
			IsHit:      true,
		},
	})

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
