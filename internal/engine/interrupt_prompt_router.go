package engine

import "starcup-engine/internal/model"

var interruptPromptBuilders = map[model.InterruptType]func(*GameEngine) *model.Prompt{
	model.InterruptResponseSkill:        (*GameEngine).buildResponseSkillPrompt,
	model.InterruptStartupSkill:         (*GameEngine).buildStartupSkillPrompt,
	model.InterruptDiscard:              (*GameEngine).buildDiscardPrompt,
	model.InterruptChoice:               (*GameEngine).buildChoicePrompt,
	model.InterruptMagicMissile:         (*GameEngine).buildMagicMissilePrompt,
	model.InterruptGiveCards:            (*GameEngine).buildGiveCardsPrompt,
	model.InterruptMagicBulletFusion:    (*GameEngine).buildMagicBulletFusionPrompt,
	model.InterruptMagicBulletDirection: (*GameEngine).buildMagicBulletDirectionPrompt,
	model.InterruptHolySwordDraw:        (*GameEngine).buildHolySwordDrawPrompt,
	model.InterruptSaintHeal:            (*GameEngine).buildSaintHealPrompt,
	model.InterruptMagicBlast:           (*GameEngine).buildMagicBlastPrompt,
}

func (e *GameEngine) buildPendingInterruptPrompt() *model.Prompt {
	intr := e.State.PendingInterrupt
	if intr == nil {
		return nil
	}
	builder, ok := interruptPromptBuilders[intr.Type]
	if !ok {
		return nil
	}
	return builder(e)
}
