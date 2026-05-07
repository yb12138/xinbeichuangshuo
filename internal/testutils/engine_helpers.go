package testutils

import (
	"fmt"
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"testing"
)

// RequireChoicePrompt asserts that the game has a pending choice interrupt
// for the given playerID with the specified choiceType.
func RequireChoicePrompt(t *testing.T, game *engine.GameEngine, playerID, choiceType string) {
	t.Helper()
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending interrupt, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected choice interrupt, got %s", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected interrupt player %s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
	ctxData, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("choice context type mismatch")
	}
	got, _ := ctxData["choice_type"].(string)
	if got != choiceType {
		t.Fatalf("expected choice_type=%s, got %s", choiceType, got)
	}
}

// RequireResponseSkillPrompt asserts that the game has a pending response-skill
// interrupt for the given playerID.
func RequireResponseSkillPrompt(t *testing.T, game *engine.GameEngine, playerID string) {
	t.Helper()
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending interrupt, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response-skill interrupt, got %s", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected interrupt player %s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
}

// ChooseResponseSkillByID asserts a response-skill prompt is active, then selects
// the skill with the given ID.
func ChooseResponseSkillByID(t *testing.T, game *engine.GameEngine, playerID, skillID string) {
	t.Helper()
	RequireResponseSkillPrompt(t, game, playerID)
	idx := -1
	for i, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == skillID {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("skill %s not found in pending skills: %+v", skillID, game.State.PendingInterrupt.SkillIDs)
	}
	MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   playerID,
		Type:       model.CmdSelect,
		Selections: []int{idx},
	})
}

// MustHandleAction calls game.HandleAction and fatals on error.
func MustHandleAction(t *testing.T, game *engine.GameEngine, act model.PlayerAction) {
	t.Helper()
	if err := game.HandleAction(act); err != nil {
		t.Fatalf("handle action failed (%+v): %v", act, err)
	}
}

// StartupSkillIndexByID returns the index of skillID in the pending startup
// interrupt's skill list.
func StartupSkillIndexByID(t *testing.T, game *engine.GameEngine, playerID, skillID string) int {
	t.Helper()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected startup interrupt for %s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
	for i, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == skillID {
			return i
		}
	}
	t.Fatalf("skill %s not found in startup skills: %+v", skillID, game.State.PendingInterrupt.SkillIDs)
	return -1
}

// RequireChoiceContext asserts a choice prompt and returns the context map.
func RequireChoiceContext(t *testing.T, game *engine.GameEngine, playerID, choiceType string) map[string]interface{} {
	t.Helper()
	RequireChoicePrompt(t, game, playerID, choiceType)
	ctx, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("choice context type mismatch")
	}
	return ctx
}

// ChoiceIndexForTarget finds the index of targetID in the choice context's target_ids.
func ChoiceIndexForTarget(t *testing.T, ctx map[string]interface{}, targetID string) int {
	t.Helper()
	var targetIDs []string
	if arr, ok := ctx["target_ids"].([]string); ok {
		targetIDs = append(targetIDs, arr...)
	} else if arr, ok := ctx["target_ids"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				targetIDs = append(targetIDs, s)
			}
		}
	}
	for i, id := range targetIDs {
		if id == targetID {
			return i
		}
	}
	t.Fatalf("target %s not found in choice target_ids=%v", targetID, targetIDs)
	return -1
}

// FindPublicDiscardReveal searches the observer events in reverse for a
// public (non-hidden) discard reveal event for the given playerID.
func FindPublicDiscardReveal(obs *CaptureObserver, playerID string) map[string]interface{} {
	if obs == nil {
		return nil
	}
	for i := len(obs.Events) - 1; i >= 0; i-- {
		event := obs.Events[i]
		if event.Type != model.EventCardRevealed {
			continue
		}
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			continue
		}
		eventPlayerID, _ := data["player_id"].(string)
		actionType, _ := data["action_type"].(string)
		hidden, _ := data["hidden"].(bool)
		if eventPlayerID == playerID && actionType == "discard" && !hidden {
			return data
		}
	}
	return nil
}

// FirstAttackCardIndex returns the index of the first attack card in the player's hand, or -1.
func FirstAttackCardIndex(p *model.Player) int {
	for i, c := range p.Hand {
		if c.Type == model.CardTypeAttack {
			return i
		}
	}
	return -1
}

// InterruptHasSkillID reports whether the interrupt contains the given skill ID.
func InterruptHasSkillID(intr *model.Interrupt, skillID string) bool {
	if intr == nil {
		return false
	}
	for _, sid := range intr.SkillIDs {
		if sid == skillID {
			return true
		}
	}
	return false
}

// CountFieldEffect counts field cards on the player with the given effect type.
func CountFieldEffect(p *model.Player, effect model.EffectType) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == effect {
			count++
		}
	}
	return count
}

// HasFieldEffect reports whether the player has a field card with the given effect type.
func HasFieldEffect(player *model.Player, effect model.EffectType) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == effect {
			return true
		}
	}
	return false
}

// GetFieldEffectCard returns the first field card with the given effect type, or nil.
func GetFieldEffectCard(player *model.Player, effect model.EffectType) *model.FieldCard {
	if player == nil {
		return nil
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != effect {
			continue
		}
		return fc
	}
	return nil
}

// PromptHasOptionID reports whether the prompt has an option with the given ID.
func PromptHasOptionID(prompt *model.Prompt, optionID string) bool {
	if prompt == nil {
		return false
	}
	for _, opt := range prompt.Options {
		if opt.ID == optionID {
			return true
		}
	}
	return false
}

// ChoiceTypeOfInterrupt returns the choice_type string from an interrupt's context, or "".
func ChoiceTypeOfInterrupt(intr *model.Interrupt) string {
	if intr == nil {
		return ""
	}
	data, ok := intr.Context.(map[string]interface{})
	if !ok {
		return ""
	}
	v, _ := data["choice_type"].(string)
	return v
}

// PendingChoiceTargetIDs extracts target_ids from an interrupt's context.
func PendingChoiceTargetIDs(intr *model.Interrupt) []string {
	if intr == nil {
		return nil
	}
	data, ok := intr.Context.(map[string]interface{})
	if !ok {
		return nil
	}
	var targetIDs []string
	if arr, ok := data["target_ids"].([]string); ok {
		targetIDs = append(targetIDs, arr...)
	} else if arr, ok := data["target_ids"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				targetIDs = append(targetIDs, s)
			}
		}
	}
	return targetIDs
}

// MakeStarterBardRousingRhapsodyCard creates the bard's starter card for the player.
func MakeStarterBardRousingRhapsodyCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bd_rousing_rhapsody", player.ID),
		Name:            "激昂狂想曲",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "吟游诗人开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "激昂狂想曲",
	}
}

// MakeStarterBardVictorySymphonyCard creates the bard's victory symphony starter card for the player.
func MakeStarterBardVictorySymphonyCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bd_victory_symphony", player.ID),
		Name:            "胜利交响诗",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "吟游诗人开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "胜利交响诗",
	}
}

// MakeStarterBardHopeFugueCard creates the bard's hope fugue starter card for the player.
func MakeStarterBardHopeFugueCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bd_hope_fugue", player.ID),
		Name:            "希望赋格曲",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "吟游诗人开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "希望赋格曲",
	}
}

// MakeStarterSoulLinkCard creates the soul sorcerer's soul link starter card for the player.
func MakeStarterSoulLinkCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-soul_link", player.ID),
		Name:            "灵魂链接",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "灵魂术士开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "灵魂链接",
	}
}

// MakeStarterBloodSharedLifeCard creates the blood priestess's shared life starter card for the player.
func MakeStarterBloodSharedLifeCard(player *model.Player) model.Card {
	return model.Card{
		ID:              fmt.Sprintf("starter-%s-bp_shared_life", player.ID),
		Name:            "同生共死",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		Faction:         player.Character.Faction,
		Description:     "血之巫女开局自带专属技能卡",
		ExclusiveChar1:  player.Character.ID,
		ExclusiveSkill1: "同生共死",
	}
}
