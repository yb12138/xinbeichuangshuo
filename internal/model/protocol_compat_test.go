package model

import "testing"

func TestTimingRefactorDoesNotChangeFrontendProtocolConstants(t *testing.T) {
	turnStages := map[TurnStage]string{
		TurnStageTurnBeforeStart: "TurnBeforeStart",
		TurnStageTurnStart:       "TurnStart",
		TurnStageBeforeAction:    "BeforeAction",
		TurnStageActionStart:     "ActionStart",
		TurnStageActionExecution: "ActionExecution",
		TurnStageActionEnd:       "ActionEnd",
		TurnStageExtraAction:     "ExtraAction",
		TurnStageTurnEnd:         "TurnEnd",
	}
	for got, want := range turnStages {
		if string(got) != want {
			t.Fatalf("TurnStage protocol changed: got %q, want %q", got, want)
		}
	}

	interrupts := map[InterruptType]string{
		InterruptResponseSkill:        "ResponseSkill",
		InterruptStartupSkill:         "StartupSkill",
		InterruptChoice:               "Choice",
		InterruptMagicMissile:         "MagicMissile",
		InterruptGiveCards:            "GiveCards",
		InterruptMagicBulletFusion:    "MagicBulletFusion",
		InterruptMagicBulletDirection: "MagicBulletDirection",
		InterruptHolySwordDraw:        "HolySwordDraw",
		InterruptSaintHeal:            "SaintHeal",
		InterruptMagicBlast:           "MagicBlast",
	}
	for got, want := range interrupts {
		if string(got) != want {
			t.Fatalf("InterruptType protocol changed: got %q, want %q", got, want)
		}
	}

	prompts := map[PromptType]string{
		PromptChooseCards:   "choose_cards",
		PromptChooseSkill:   "choose_skill",
		PromptConfirm:       "confirm",
		PromptChooseExtract: "choose_extract",
	}
	for got, want := range prompts {
		if string(got) != want {
			t.Fatalf("PromptType protocol changed: got %q, want %q", got, want)
		}
	}

	commands := map[PlayerActionType]string{
		CmdAttack:  "Attack",
		CmdMagic:   "Magic",
		CmdRespond: "Respond",
		CmdSkill:   "Skill",
		CmdSelect:  "Select",
	}
	for got, want := range commands {
		if string(got) != want {
			t.Fatalf("PlayerActionType protocol changed: got %q, want %q", got, want)
		}
	}

	promptChoiceTypes := []string{
		"assassin_stealth_draw",
		"heal",
		"discard",
		"counter",
		"take",
	}
	for _, choiceType := range promptChoiceTypes {
		if choiceType == "" {
			t.Fatal("ChoiceType protocol sentinel changed to empty string")
		}
	}
}
