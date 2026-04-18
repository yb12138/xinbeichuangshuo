// gameflow: 新角色：形态、身份切换相关判定。
// 形态基础设施已迁移到 player 包，此文件保留 engine 专属逻辑。

package engine

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const elfBlessingPrefix = "elf_blessing:"

func isCharacter(p *model.Player, charID string) bool {
	return player.IsCharacter(p, charID)
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

// ---- Token cap 常量 ----

const magicBowChargeCapEngine = 8
const bardInspirationCapEngine = 3
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

// ---- 形态基础设施（委托到 player 包） ----

func effectivePlayerOrientation(p *model.Player) model.CharacterOrientation {
	return player.EffectiveOrientation(p)
}

func effectivePlayerForm(p *model.Player) string {
	return player.EffectiveForm(p)
}

func playerHasForm(p *model.Player, form string) bool {
	return player.HasForm(p, form)
}

func setPlayerForm(p *model.Player, form string) bool {
	return player.SetForm(p, form)
}

func clearPlayerForm(p *model.Player, form string) bool {
	return player.ClearForm(p, form)
}

// ---- 角色形态三元组（委托到 player 包） ----

func hasPrayerMasterPrayerForm(p *model.Player) bool {
	return playerHasForm(p, model.FormPrayerMasterPrayer)
}

func hasValkyrieHeroicForm(p *model.Player) bool {
	return playerHasForm(p, model.FormValkyrieHeroic)
}

func enterValkyrieHeroicForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormValkyrieHeroic)
}

func leaveValkyrieHeroicForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormValkyrieHeroic)
}

func hasAssassinStealthForm(p *model.Player) bool {
	return playerHasForm(p, model.FormAssassinStealth)
}

func enterAssassinStealthForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormAssassinStealth)
}

func leaveAssassinStealthForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormAssassinStealth)
}

func hasCrimsonKnightHotBloodedForm(p *model.Player) bool {
	return playerHasForm(p, model.FormCrimsonKnightHotBlooded)
}

func enterCrimsonKnightHotBloodedForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormCrimsonKnightHotBlooded)
}

func leaveCrimsonKnightHotBloodedForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormCrimsonKnightHotBlooded)
}

func hasOnmyojiShikigamiForm(p *model.Player) bool {
	return playerHasForm(p, model.FormOnmyojiShikigami)
}

func leaveOnmyojiShikigamiForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormOnmyojiShikigami)
}

func hasBlazeWitchFlameForm(p *model.Player) bool {
	return playerHasForm(p, model.FormBlazeWitchFlame)
}

func leaveBlazeWitchFlameForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormBlazeWitchFlame)
}

func hasHolyBowHolyGloryForm(p *model.Player) bool {
	return playerHasForm(p, model.FormHolyBowHolyGlory)
}

func hasArbiterJudgmentForm(p *model.Player) bool {
	return playerHasForm(p, model.FormArbiterJudgment)
}

func enterArbiterJudgmentForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormArbiterJudgment)
}

func hasElfArcherRitualForm(p *model.Player) bool {
	return playerHasForm(p, model.FormElfArcherRitual)
}

func enterElfArcherRitualForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormElfArcherRitual)
}

func leaveElfArcherRitualForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormElfArcherRitual)
}

func hasMagicSwordsmanShadowForm(p *model.Player) bool {
	return playerHasForm(p, model.FormMagicSwordsmanShadow)
}

func enterMagicSwordsmanShadowForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormMagicSwordsmanShadow)
}

func leaveMagicSwordsmanShadowForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormMagicSwordsmanShadow)
}

func hasWarHomunculusBurstForm(p *model.Player) bool {
	return playerHasForm(p, model.FormWarHomunculusBurst)
}

func enterWarHomunculusBurstForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormWarHomunculusBurst)
}

func leaveWarHomunculusBurstForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormWarHomunculusBurst)
}

func enterHolyBowHolyGloryForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormHolyBowHolyGlory)
}

func leaveHolyBowHolyGloryForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormHolyBowHolyGlory)
}

func hasMagicLancerPhantomForm(p *model.Player) bool {
	return playerHasForm(p, model.FormMagicLancerPhantom)
}

func leaveMagicLancerPhantomForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormMagicLancerPhantom)
}

func hasBardEternalPrisonerForm(p *model.Player) bool {
	return playerHasForm(p, model.FormBardEternalPrisoner)
}

func enterBardEternalPrisonerForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormBardEternalPrisoner)
}

func leaveBardEternalPrisonerForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormBardEternalPrisoner)
}

func hasHeroExhaustionForm(p *model.Player) bool {
	return playerHasForm(p, model.FormHeroExhaustion)
}

func leaveHeroExhaustionForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormHeroExhaustion)
}

func hasFighterHundredDragonForm(p *model.Player) bool {
	return playerHasForm(p, model.FormFighterHundredDragon)
}

func leaveFighterHundredDragonForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormFighterHundredDragon)
}

func enterMoonGoddessDarkMoonForm(p *model.Player) bool {
	return setPlayerForm(p, model.FormMoonGoddessDarkMoon)
}

func leaveMoonGoddessDarkMoonForm(p *model.Player) bool {
	return clearPlayerForm(p, model.FormMoonGoddessDarkMoon)
}

func hasBloodPriestessBleedingForm(p *model.Player) bool {
	return playerHasForm(p, model.FormBloodPriestessBleeding)
}

func enterBloodPriestessBleedingFormState(p *model.Player) bool {
	return setPlayerForm(p, model.FormBloodPriestessBleeding)
}

func leaveBloodPriestessBleedingFormState(p *model.Player) bool {
	return clearPlayerForm(p, model.FormBloodPriestessBleeding)
}

// ---- 引擎级形态基础设施（保留在 engine） ----

type poseSnapshot struct {
	Orientation model.CharacterOrientation
	Form        string
}

func (e *GameEngine) snapshotPlayerPoses() map[string]poseSnapshot {
	snapshots := make(map[string]poseSnapshot, len(e.State.Players))
	for id, p := range e.State.Players {
		snapshots[id] = poseSnapshot{
			Orientation: effectivePlayerOrientation(p),
			Form:        effectivePlayerForm(p),
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
		p := e.State.Players[playerID]
		if p == nil {
			continue
		}
		prev := before[playerID]
		current := poseSnapshot{
			Orientation: effectivePlayerOrientation(p),
			Form:        effectivePlayerForm(p),
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
		ctx := e.buildContext(p, p, model.TimingOnOrientationChanged, eventCtx)
		e.dispatcher.OnTiming(ctx.Timing, ctx)
	}
}
