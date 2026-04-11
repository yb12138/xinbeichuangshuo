// gameflow: 新角色：形态、身份切换相关判定。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

const elfBlessingPrefix = "elf_blessing:"

func isCharacter(player *model.Player, charID string) bool {
	return player != nil && player.Character != nil && player.Character.ID == charID
}

func (e *GameEngine) isElfArcher(player *model.Player) bool {
	return isCharacter(player, "elf_archer")
}

func (e *GameEngine) isPlagueMage(player *model.Player) bool {
	return isCharacter(player, "plague_mage")
}

func (e *GameEngine) isMagicSwordsman(player *model.Player) bool {
	return isCharacter(player, "magic_swordsman")
}

func (e *GameEngine) maybeReleaseMagicSwordsmanShadowAtActionStart(player *model.Player) bool {
	if e == nil || player == nil || !e.isMagicSwordsman(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.TurnState.HasUsedActionSkill {
		return false
	}
	if !hasMagicSwordsmanShadowForm(player) {
		return false
	}
	leaveMagicSwordsmanShadowForm(player)
	e.Log(fmt.Sprintf("%s 脱离暗影形态并转正", player.Name))
	return true
}

func (e *GameEngine) isCrimsonSwordSpirit(player *model.Player) bool {
	return isCharacter(player, "crimson_sword_spirit")
}

func (e *GameEngine) isPrayerMaster(player *model.Player) bool {
	return isCharacter(player, "prayer_master")
}

func (e *GameEngine) isCrimsonKnight(player *model.Player) bool {
	return isCharacter(player, "crimson_knight")
}

func (e *GameEngine) isWarHomunculus(player *model.Player) bool {
	return isCharacter(player, "war_homunculus")
}

func (e *GameEngine) isPriest(player *model.Player) bool {
	return isCharacter(player, "priest")
}

func (e *GameEngine) isOnmyoji(player *model.Player) bool {
	return isCharacter(player, "onmyoji")
}

func (e *GameEngine) isBlazeWitch(player *model.Player) bool {
	return isCharacter(player, "blaze_witch")
}

func (e *GameEngine) isSage(player *model.Player) bool {
	return isCharacter(player, "sage")
}

func (e *GameEngine) isMagicBow(player *model.Player) bool {
	return isCharacter(player, "magic_bow")
}

func (e *GameEngine) isMagicLancer(player *model.Player) bool {
	return isCharacter(player, "magic_lancer")
}

func (e *GameEngine) isSpiritCaster(player *model.Player) bool {
	return isCharacter(player, "spirit_caster")
}

func (e *GameEngine) isBard(player *model.Player) bool {
	return isCharacter(player, "bard")
}

func (e *GameEngine) isHero(player *model.Player) bool {
	return isCharacter(player, "hero")
}

func (e *GameEngine) isFighter(player *model.Player) bool {
	return isCharacter(player, "fighter")
}

func (e *GameEngine) isHolyBow(player *model.Player) bool {
	return isCharacter(player, "holy_bow")
}

func (e *GameEngine) isSwordEmperor(player *model.Player) bool {
	return isCharacter(player, "sword_emperor")
}

func (e *GameEngine) isBeastSamurai(player *model.Player) bool {
	return isCharacter(player, "beast_samurai")
}

func (e *GameEngine) isHolyLancer(player *model.Player) bool {
	return isCharacter(player, "holy_lancer")
}

func (e *GameEngine) isSoulSorcerer(player *model.Player) bool {
	return isCharacter(player, "soul_sorcerer")
}

func (e *GameEngine) isMoonGoddess(player *model.Player) bool {
	return isCharacter(player, "moon_goddess")
}

func (e *GameEngine) isBloodPriestess(player *model.Player) bool {
	return isCharacter(player, "blood_priestess")
}

func (e *GameEngine) isButterflyDancer(player *model.Player) bool {
	return isCharacter(player, "butterfly_dancer")
}

const magicBowChargeCapEngine = 8
const bardInspirationCapEngine = 3
const heroTokenCapEngine = 4
const holyBowFaithCapEngine = 10
const holyBowCannonCapEngine = 1
const standardCampMoraleCapEngine = 15
const swordEmperorSwordQiCapEngine = 5
const swordEmperorSwordSoulCapEngine = 3
const beastSamuraiZanshinCapEngine = 4
const beastSamuraiBeastSoulCapEngine = 2
const soulSorcererBlueCapEngine = 6
const soulSorcererYellowCapEngine = 6
const moonGoddessNewMoonCapEngine = 2
const moonGoddessPetrifyCapEngine = 3
const butterflyCocoonCapEngine = 8

type poseSnapshot struct {
	Orientation model.CharacterOrientation
	Form        string
}

func legacyPlayerPose(player *model.Player) (poseSnapshot, bool) {
	if player == nil {
		return poseSnapshot{}, false
	}
	if player.Tokens == nil {
		return poseSnapshot{}, false
	}
	switch {
	case player.Tokens["valkyrie_spirit"] > 0:
		return poseSnapshot{Orientation: model.OrientationTapped, Form: "heroic_form"}, true
	}
	return poseSnapshot{}, false
}

func effectivePlayerOrientation(player *model.Player) model.CharacterOrientation {
	if player == nil {
		return model.OrientationNormal
	}
	if legacy, ok := legacyPlayerPose(player); ok {
		return legacy.Orientation
	}
	if player.Orientation != "" {
		return player.Orientation
	}
	if player.Form != "" {
		return model.OrientationTapped
	}
	return model.OrientationNormal
}

func effectivePlayerForm(player *model.Player) string {
	if player == nil {
		return ""
	}
	if legacy, ok := legacyPlayerPose(player); ok {
		return legacy.Form
	}
	return player.Form
}

func playerHasForm(player *model.Player, form string) bool {
	return player != nil && effectivePlayerForm(player) == form
}

func setPlayerForm(player *model.Player, form string) bool {
	if player == nil {
		return false
	}
	changed := effectivePlayerOrientation(player) != model.OrientationTapped || effectivePlayerForm(player) != form
	player.Orientation = model.OrientationTapped
	player.Form = form
	return changed
}

func clearPlayerForm(player *model.Player, form string) bool {
	if player == nil {
		return false
	}
	if form != "" && effectivePlayerForm(player) != form {
		return false
	}
	changed := effectivePlayerOrientation(player) != model.OrientationNormal || effectivePlayerForm(player) != ""
	player.Orientation = model.OrientationNormal
	player.Form = ""
	return changed
}

func hasPrayerMasterPrayerForm(player *model.Player) bool {
	return playerHasForm(player, model.FormPrayerMasterPrayer)
}

func hasAssassinStealthForm(player *model.Player) bool {
	return playerHasForm(player, model.FormAssassinStealth)
}

func enterAssassinStealthForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormAssassinStealth)
}

func leaveAssassinStealthForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormAssassinStealth)
}

func enterPrayerMasterPrayerForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormPrayerMasterPrayer)
}

func hasCrimsonKnightHotBloodedForm(player *model.Player) bool {
	return playerHasForm(player, model.FormCrimsonKnightHotBlooded)
}

func enterCrimsonKnightHotBloodedForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormCrimsonKnightHotBlooded)
}

func leaveCrimsonKnightHotBloodedForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormCrimsonKnightHotBlooded)
}

func hasOnmyojiShikigamiForm(player *model.Player) bool {
	return playerHasForm(player, model.FormOnmyojiShikigami)
}

func enterOnmyojiShikigamiForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormOnmyojiShikigami)
}

func leaveOnmyojiShikigamiForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormOnmyojiShikigami)
}

func hasBlazeWitchFlameForm(player *model.Player) bool {
	return playerHasForm(player, model.FormBlazeWitchFlame)
}

func enterBlazeWitchFlameForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormBlazeWitchFlame)
}

func leaveBlazeWitchFlameForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormBlazeWitchFlame)
}

func hasHolyBowHolyGloryForm(player *model.Player) bool {
	return playerHasForm(player, model.FormHolyBowHolyGlory)
}

func hasArbiterJudgmentForm(player *model.Player) bool {
	return playerHasForm(player, model.FormArbiterJudgment)
}

func enterArbiterJudgmentForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormArbiterJudgment)
}

func leaveArbiterJudgmentForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormArbiterJudgment)
}

func hasElfArcherRitualForm(player *model.Player) bool {
	return playerHasForm(player, model.FormElfArcherRitual)
}

func enterElfArcherRitualForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormElfArcherRitual)
}

func leaveElfArcherRitualForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormElfArcherRitual)
}

func hasMagicSwordsmanShadowForm(player *model.Player) bool {
	return playerHasForm(player, model.FormMagicSwordsmanShadow)
}

func enterMagicSwordsmanShadowForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormMagicSwordsmanShadow)
}

func leaveMagicSwordsmanShadowForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormMagicSwordsmanShadow)
}

func hasWarHomunculusBurstForm(player *model.Player) bool {
	return playerHasForm(player, model.FormWarHomunculusBurst)
}

func enterWarHomunculusBurstForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormWarHomunculusBurst)
}

func leaveWarHomunculusBurstForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormWarHomunculusBurst)
}

func enterHolyBowHolyGloryForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormHolyBowHolyGlory)
}

func leaveHolyBowHolyGloryForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormHolyBowHolyGlory)
}

func hasMagicLancerPhantomForm(player *model.Player) bool {
	return playerHasForm(player, model.FormMagicLancerPhantom)
}

func enterMagicLancerPhantomForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormMagicLancerPhantom)
}

func leaveMagicLancerPhantomForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormMagicLancerPhantom)
}

func hasBardEternalPrisonerForm(player *model.Player) bool {
	return playerHasForm(player, model.FormBardEternalPrisoner)
}

func enterBardEternalPrisonerForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormBardEternalPrisoner)
}

func leaveBardEternalPrisonerForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormBardEternalPrisoner)
}

func hasHeroExhaustionForm(player *model.Player) bool {
	return playerHasForm(player, model.FormHeroExhaustion)
}

func enterHeroExhaustionForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormHeroExhaustion)
}

func leaveHeroExhaustionForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormHeroExhaustion)
}

func hasFighterHundredDragonForm(player *model.Player) bool {
	return playerHasForm(player, model.FormFighterHundredDragon)
}

func enterFighterHundredDragonForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormFighterHundredDragon)
}

func leaveFighterHundredDragonForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormFighterHundredDragon)
}

func hasMoonGoddessDarkMoonForm(player *model.Player) bool {
	return playerHasForm(player, model.FormMoonGoddessDarkMoon)
}

func enterMoonGoddessDarkMoonForm(player *model.Player) bool {
	return setPlayerForm(player, model.FormMoonGoddessDarkMoon)
}

func leaveMoonGoddessDarkMoonForm(player *model.Player) bool {
	return clearPlayerForm(player, model.FormMoonGoddessDarkMoon)
}

func hasBloodPriestessBleedingForm(player *model.Player) bool {
	return playerHasForm(player, model.FormBloodPriestessBleeding)
}

func enterBloodPriestessBleedingFormState(player *model.Player) bool {
	return setPlayerForm(player, model.FormBloodPriestessBleeding)
}

func leaveBloodPriestessBleedingFormState(player *model.Player) bool {
	return clearPlayerForm(player, model.FormBloodPriestessBleeding)
}

func (e *GameEngine) snapshotPlayerPoses() map[string]poseSnapshot {
	snapshots := make(map[string]poseSnapshot, len(e.State.Players))
	for id, player := range e.State.Players {
		snapshots[id] = poseSnapshot{
			Orientation: effectivePlayerOrientation(player),
			Form:        effectivePlayerForm(player),
		}
	}
	return snapshots
}

func (e *GameEngine) dispatchOrientationChanges(before map[string]poseSnapshot) {
	if e == nil || len(before) == 0 {
		return
	}
	orderedIDs := append([]string{}, e.State.PlayerOrder...)
	seen := make(map[string]bool, len(orderedIDs))
	for _, id := range orderedIDs {
		seen[id] = true
	}
	for id := range e.State.Players {
		if !seen[id] {
			orderedIDs = append(orderedIDs, id)
		}
	}
	for _, playerID := range orderedIDs {
		player := e.State.Players[playerID]
		if player == nil {
			continue
		}
		prev := before[playerID]
		current := poseSnapshot{
			Orientation: effectivePlayerOrientation(player),
			Form:        effectivePlayerForm(player),
		}
		if prev == current {
			continue
		}
		eventCtx := &model.EventContext{
			Type:            model.EventNone,
			SourceID:        playerID,
			TargetID:        playerID,
			OperatorID:      playerID,
			PrevOrientation: prev.Orientation,
			NewOrientation:  current.Orientation,
			PrevForm:        prev.Form,
			NewForm:         current.Form,
		}
		ctx := e.buildContext(player, player, model.TimingOnOrientationChanged, eventCtx)
		e.dispatcher.OnTiming(ctx.Timing, ctx)
	}
}
