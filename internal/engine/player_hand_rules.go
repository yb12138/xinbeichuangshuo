// gameflow: 手牌上限、可见性等规则。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
	adventurerplayer "starcup-engine/internal/engine/player/adventurer"
	angelplayer "starcup-engine/internal/engine/player/angel"
	arbiterplayer "starcup-engine/internal/engine/player/arbiter"
	archerplayer "starcup-engine/internal/engine/player/archer"
	assassinplayer "starcup-engine/internal/engine/player/assassin"
	bardplayer "starcup-engine/internal/engine/player/bard"
	beastsamuraiplayer "starcup-engine/internal/engine/player/beast_samurai"
	berserkerplayer "starcup-engine/internal/engine/player/berserker"
	blademasterplayer "starcup-engine/internal/engine/player/blade_master"
	blazewitchplayer "starcup-engine/internal/engine/player/blaze_witch"
	bloodpriestessplayer "starcup-engine/internal/engine/player/blood_priestess"
	butterflydancerplayer "starcup-engine/internal/engine/player/butterfly_dancer"
	crimsonknightplayer "starcup-engine/internal/engine/player/crimson_knight"
	crimsonswordspiritplayer "starcup-engine/internal/engine/player/crimson_sword_spirit"
	elementalistplayer "starcup-engine/internal/engine/player/elementalist"
	elfarcherplayer "starcup-engine/internal/engine/player/elf_archer"
	fighterplayer "starcup-engine/internal/engine/player/fighter"
	heroplayer "starcup-engine/internal/engine/player/hero"
	holybowplayer "starcup-engine/internal/engine/player/holy_bow"
	holylancerplayer "starcup-engine/internal/engine/player/holy_lancer"
	magicbowplayer "starcup-engine/internal/engine/player/magic_bow"
	magiclancerplayer "starcup-engine/internal/engine/player/magic_lancer"
	magicswordsmanplayer "starcup-engine/internal/engine/player/magic_swordsman"
	magicalgirlplayer "starcup-engine/internal/engine/player/magical_girl"
	moonplayer "starcup-engine/internal/engine/player/moon"
	onmyojiproplayer "starcup-engine/internal/engine/player/onmyoji"
	plagueplayer "starcup-engine/internal/engine/player/plague_mage"
	prayermasterplayer "starcup-engine/internal/engine/player/prayer_master"
	priestplayer "starcup-engine/internal/engine/player/priest"
	sageplayer "starcup-engine/internal/engine/player/sage"
	saintessplayer "starcup-engine/internal/engine/player/saintess"
	sealerplayer "starcup-engine/internal/engine/player/sealer"
	soulsocererplayer "starcup-engine/internal/engine/player/soul_sorcerer"
	spiritcasterplayer "starcup-engine/internal/engine/player/spirit_caster"
	swordemperorplayer "starcup-engine/internal/engine/player/sword_emperor"
	valkyrieplayer "starcup-engine/internal/engine/player/valkyrie"
	warhomunculusplayer "starcup-engine/internal/engine/player/war_homunculus"
	"starcup-engine/internal/model"
)

var roleRegistry = buildRoleRegistry()

func buildRoleRegistry() *engineplayer.RoleRegistry {
	registry := engineplayer.NewRoleRegistry()
	registerAngelRoleEntry(registry)
	registerArbiterRoleEntry(registry)
	registerBardRoleEntry(registry)
	registerBeastSamuraiRoleEntry(registry)
	registerBlazeWitchRoleEntry(registry)
	registerBloodPriestessRoleEntry(registry)
	registerButterflyDancerRoleEntry(registry)
	registerCrimsonKnightRoleEntry(registry)
	registerCrimsonSwordSpiritRoleEntry(registry)
	registerElfArcherRoleEntry(registry)
	registerFighterRoleEntry(registry)
	registerHolyBowRoleEntry(registry)
	registerHolyLancerRoleEntry(registry)
	registerMagicBowRoleEntry(registry)
	registerMagicLancerRoleEntry(registry)
	registerMagicSwordsmanRoleEntry(registry)
	registerMoonGoddessRoleEntry(registry)
	registerOnmyojiRoleEntry(registry)
	registerPlagueMageRoleEntry(registry)
	registerPrayerMasterRoleEntry(registry)
	registerPriestRoleEntry(registry)
	registerSageRoleEntry(registry)
	registerSealerRoleEntry(registry)
	registerSaintessRoleEntry(registry)
	registerSoulSorcererRoleEntry(registry)
	registerSpiritCasterRoleEntry(registry)
	registerSwordEmperorRoleEntry(registry)
	registerWarHomunculusRoleEntry(registry)
	registerAdventurerRoleEntry(registry)
	registerAssassinRoleEntry(registry)
	registerHeroRoleEntry(registry)
	registerBerserkerRoleEntry(registry)
	registerMagicalGirlRoleEntry(registry)
	registerBladeMasterRoleEntry(registry)
	registerArcherRoleEntry(registry)
	registerValkyrieRoleEntry(registry)
	registerElementalistRoleEntry(registry)
	registerHandLimitRoleEntries(registry)
	return registry
}

func registerAngelRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(angelplayer.RoleEntry())
}

func registerArbiterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(arbiterplayer.RoleEntry())
}

func registerBlazeWitchRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(blazewitchplayer.RoleEntry())
}

func registerCrimsonSwordSpiritRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(crimsonswordspiritplayer.RoleEntry())
}

func registerMoonGoddessRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(moonplayer.RoleEntry())
}

func registerOnmyojiRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(onmyojiproplayer.RoleEntry())
}

func registerHolyLancerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(holylancerplayer.RoleEntry())
}

func registerPlagueMageRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(plagueplayer.RoleEntry())
}

func registerMagicSwordsmanRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(magicswordsmanplayer.RoleEntry())
}

func registerSageRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(sageplayer.RoleEntry())
}

func registerSealerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(sealerplayer.RoleEntry())
}

func registerSaintessRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(saintessplayer.RoleEntry())
}

func registerSpiritCasterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(spiritcasterplayer.RoleEntry())
}

func registerAdventurerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(adventurerplayer.RoleEntry())
}

func registerAssassinRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(assassinplayer.RoleEntry())
}

func registerHeroRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(heroplayer.RoleEntry())
}

func registerBerserkerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(berserkerplayer.RoleEntry())
}

func registerMagicalGirlRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(magicalgirlplayer.RoleEntry())
}

func registerBladeMasterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(blademasterplayer.RoleEntry())
}

func registerArcherRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(archerplayer.RoleEntry())
}

func registerValkyrieRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(valkyrieplayer.RoleEntry())
}

func registerElementalistRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(elementalistplayer.RoleEntry())
}

func registerBardRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(bardplayer.RoleEntry())
}

func registerBeastSamuraiRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(beastsamuraiplayer.RoleEntry())
}

func registerBloodPriestessRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(bloodpriestessplayer.RoleEntry())
}

func registerButterflyDancerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(butterflydancerplayer.RoleEntry())
}

func registerCrimsonKnightRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(crimsonknightplayer.RoleEntry())
}

func registerElfArcherRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(elfarcherplayer.RoleEntry())
}

func registerFighterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(fighterplayer.RoleEntry())
}

func registerHolyBowRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(holybowplayer.RoleEntry())
}

func registerMagicBowRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(magicbowplayer.RoleEntry())
}

func registerMagicLancerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(magiclancerplayer.RoleEntry())
}

func registerPrayerMasterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(prayermasterplayer.RoleEntry())
}

func registerPriestRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(priestplayer.RoleEntry())
}

func registerSoulSorcererRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(soulsocererplayer.RoleEntry())
}

func registerSwordEmperorRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(swordemperorplayer.RoleEntry())
}

func registerWarHomunculusRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(warhomunculusplayer.RoleEntry())
}

func registerHandLimitRoleEntries(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID: "prayer_master",
		HandLimit: engineplayer.HandLimitRuleFuncs{
			Hard: func(player *model.Player) (int, bool) {
				if hasPrayerMasterPrayerForm(player) {
					return 5, true
				}
				return 0, false
			},
		},
	})
	registry.Register(engineplayer.RoleEntry{
		ID: "magic_lancer",
		HandLimit: engineplayer.HandLimitRuleFuncs{
			Hard: func(player *model.Player) (int, bool) {
				if hasMagicLancerPhantomForm(player) {
					return 5, true
				}
				return 0, false
			},
		},
	})
	registry.Register(engineplayer.RoleEntry{
		ID: "hero",
		HandLimit: engineplayer.HandLimitRuleFuncs{
			Hard: func(player *model.Player) (int, bool) {
				if hasHeroExhaustionForm(player) {
					return 4, true
				}
				return 0, false
			},
		},
	})
	registry.Register(engineplayer.RoleEntry{
		ID: "war_homunculus",
		HandLimit: engineplayer.HandLimitRuleFuncs{
			Modifier: func(player *model.Player, current int) int {
				if hasWarHomunculusBurstForm(player) {
					return current + 1
				}
				return current
			},
		},
	})
	registry.Register(engineplayer.RoleEntry{
		ID: "blaze_witch",
		HandLimit: engineplayer.HandLimitRuleFuncs{
			Modifier: func(player *model.Player, current int) int {
				if hasBlazeWitchFlameForm(player) {
					return current + player.Tokens["bw_rebirth"] - 2
				}
				return current
			},
		},
	})
	registry.Register(engineplayer.RoleEntry{
		ID: "assassin",
		HandLimit: engineplayer.HandLimitRuleFuncs{
			Modifier: func(player *model.Player, current int) int {
				if hasAssassinStealthForm(player) {
					return current - 1
				}
				return current
			},
		},
	})
	registry.Register(engineplayer.RoleEntry{
		ID: "butterfly_dancer",
		HandLimit: engineplayer.HandLimitRuleFuncs{
			Modifier: func(player *model.Player, current int) int {
				current -= butterflyPupa(player)
				if current < 3 {
					return 3
				}
				return current
			},
		},
	})
}

func roleIDForHandLimitRule(player *model.Player) string {
	if player == nil {
		return ""
	}
	if player.Character != nil && player.Character.ID != "" {
		return player.Character.ID
	}
	return player.Role
}

func (e *GameEngine) roleHandLimitRule(player *model.Player) engineplayer.HandLimitRule {
	return roleRegistry.HandLimitRule(roleIDForHandLimitRule(player))
}

func (e *GameEngine) baseMaxHand(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.MaxHand > 0 {
		return player.MaxHand
	}
	// 角色基础手牌上限默认值。
	return 6
}

func (e *GameEngine) roleFixedMaxHandCapValue(player *model.Player) (int, bool) {
	if player == nil {
		return 0, false
	}
	return e.roleHandLimitRule(player).HardCap(player)
}

func (e *GameEngine) hasMercyFixedMaxHandCap(player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectMercy {
			return true
		}
	}
	return false
}

// GetMaxHand 计算玩家的动态手牌上限
func (e *GameEngine) GetMaxHand(player *model.Player) int {
	if player == nil {
		return 0
	}
	base := e.baseMaxHand(player)
	if fixed, ok := e.fixedMaxHandCapValue(player); ok {
		if fixed < 0 {
			return 0
		}
		return fixed
	}

	maxHand := e.applyRoleMaxHandModifiers(player, base)
	if maxHand < 0 {
		return 0
	}
	return maxHand
}

func (e *GameEngine) fixedMaxHandCapValue(player *model.Player) (int, bool) {
	if fixed, ok := e.roleFixedMaxHandCapValue(player); ok {
		return fixed, true
	}
	if e.hasMercyFixedMaxHandCap(player) {
		return 7, true
	}
	return 0, false
}

func (e *GameEngine) applyRoleMaxHandModifiers(player *model.Player, maxHand int) int {
	maxHand = e.roleHandLimitRule(player).ModifierCap(player, maxHand)
	// 血之巫女：同生共死根据形态动态修正手牌上限。
	maxHand += e.bloodPriestessSharedLifeDeltaFor(player)
	return maxHand
}
