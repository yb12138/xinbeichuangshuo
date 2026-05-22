package elf_archer

import (
	"fmt"
	"strings"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// QueueAnimalResponse checks and queues the elf archer animal companion response.
func QueueAnimalResponse(rt engineplayer.ChoiceRuntime, source, target *model.Player, pd *model.PendingDamage) bool {
	if rt == nil || source == nil || target == nil || pd == nil {
		return false
	}
	if !engineplayer.IsCharacter(source, "elf_archer") || !source.IsActive {
		return false
	}
	if pd.TargetID == "" || pd.TargetID == source.ID {
		return false
	}
	if !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) || pd.Card == nil || pd.IsCounter {
		return false
	}

	damageVal := pd.Damage
	ctx := rt.BuildContext(source, target, model.TimingDamageTaken, &model.EventContext{
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
		if rt.IsSkillStillUsable(skillID, source, ctx) {
			skillIDs = append(skillIDs, skillID)
		}
	}
	if len(skillIDs) == 0 {
		return false
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: source.ID,
		SkillIDs: skillIDs,
		Context:  ctx,
	})
	rt.Log(fmt.Sprintf("%s 的 [动物伙伴] 响应窗口开启", source.Name))
	return true
}
