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
	registry.Register(engineplayer.RoleEntry{
		ID:               "angel",
		Choices:          angelplayer.NewChoiceHandler(),
		Skills:           angelplayer.SkillEntries(),
		ChoiceRouteSpecs: angelplayer.ChoiceRouteSpecs(),
	})
}

func registerArbiterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "arbiter",
		Choices:          arbiterplayer.NewChoiceHandler(),
		Skills:           arbiterplayer.SkillEntries(),
		ChoiceRouteSpecs: arbiterplayer.ChoiceRouteSpecs(),
	})
}

func registerBlazeWitchRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "blaze_witch",
		Choices:          blazewitchplayer.NewChoiceHandler(),
		Skills:           blazewitchplayer.SkillEntries(),
		ChoiceRouteSpecs: blazewitchplayer.ChoiceRouteSpecs(),
	})
}

func registerCrimsonSwordSpiritRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "crimson_sword_spirit",
		Choices:          crimsonswordspiritplayer.NewChoiceHandler(),
		Skills:           crimsonswordspiritplayer.SkillEntries(),
		ChoiceRouteSpecs: crimsonswordspiritplayer.ChoiceRouteSpecs(),
	})
}

func registerMoonGoddessRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "moon_goddess",
		Choices:          moonplayer.NewChoiceHandler(),
		Skills:           moonplayer.SkillEntries(),
		ChoiceRouteSpecs: moonplayer.ChoiceRouteSpecs(),
	})
}

func registerOnmyojiRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "onmyoji",
		Choices:          onmyojiproplayer.NewChoiceHandler(),
		Skills:           onmyojiproplayer.SkillEntries(),
		ChoiceRouteSpecs: onmyojiproplayer.ChoiceRouteSpecs(),
	})
}

func registerHolyLancerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "holy_lancer",
		Choices:          holylancerplayer.NewChoiceHandler(),
		Skills:           holylancerplayer.SkillEntries(),
		ChoiceRouteSpecs: holylancerplayer.ChoiceRouteSpecs(),
	})
}

func registerPlagueMageRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "plague_mage",
		Choices:          plagueplayer.NewChoiceHandler(),
		Skills:           plagueplayer.SkillEntries(),
		ChoiceRouteSpecs: plagueplayer.ChoiceRouteSpecs(),
	})
}

func registerMagicSwordsmanRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "magic_swordsman",
		Choices:          magicswordsmanplayer.NewChoiceHandler(),
		Skills:           magicswordsmanplayer.SkillEntries(),
		ChoiceRouteSpecs: magicswordsmanplayer.ChoiceRouteSpecs(),
	})
}

func registerSageRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "sage",
		Choices:          sageplayer.NewChoiceHandler(),
		Skills:           sageplayer.SkillEntries(),
		ChoiceRouteSpecs: sageplayer.ChoiceRouteSpecs(),
	})
}

func registerSealerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "sealer",
		Choices:          sealerplayer.NewChoiceHandler(),
		Skills:           sealerplayer.SkillEntries(),
		ChoiceRouteSpecs: sealerplayer.ChoiceRouteSpecs(),
	})
}

func registerSaintessRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "saintess",
		Choices:          saintessplayer.NewChoiceHandler(),
		Skills:           saintessplayer.SkillEntries(),
		ChoiceRouteSpecs: saintessplayer.ChoiceRouteSpecs(),
	})
}

func registerSpiritCasterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:      "spirit_caster",
		Choices: spiritcasterplayer.NewChoiceHandler(),
		Skills:  spiritcasterplayer.SkillEntries(),
	})
}

func registerAdventurerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "adventurer",
		Choices:          adventurerplayer.NewChoiceHandler(),
		Skills:           adventurerplayer.SkillEntries(),
		ChoiceRouteSpecs: adventurerplayer.ChoiceRouteSpecs(),
	})
}

func registerAssassinRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "assassin",
		Choices:          assassinplayer.NewChoiceHandler(),
		Skills:           assassinplayer.SkillEntries(),
		ChoiceRouteSpecs: assassinplayer.ChoiceRouteSpecs(),
	})
}

func registerHeroRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "hero",
		Choices:          heroplayer.NewChoiceHandler(),
		Skills:           heroplayer.SkillEntries(),
		ChoiceRouteSpecs: heroplayer.ChoiceRouteSpecs(),
	})
}

func registerBerserkerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:      "berserker",
		Choices: berserkerplayer.NewChoiceHandler(),
		Skills:  berserkerplayer.SkillEntries(),
	})
}

func registerMagicalGirlRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:      "magical_girl",
		Choices: magicalgirlplayer.NewChoiceHandler(),
		Skills:  magicalgirlplayer.SkillEntries(),
	})
}

func registerBladeMasterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:      "blade_master",
		Choices: blademasterplayer.NewChoiceHandler(),
		Skills:  blademasterplayer.SkillEntries(),
	})
}

func registerArcherRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:      "archer",
		Choices: archerplayer.NewChoiceHandler(),
		Skills:  archerplayer.SkillEntries(),
	})
}

func registerValkyrieRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "valkyrie",
		Choices:          valkyrieplayer.NewChoiceHandler(),
		Skills:           valkyrieplayer.SkillEntries(),
		ChoiceRouteSpecs: valkyrieplayer.ChoiceRouteSpecs(),
	})
}

func registerElementalistRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "elementalist",
		Choices:          elementalistplayer.NewChoiceHandler(),
		Skills:           elementalistplayer.SkillEntries(),
		ChoiceRouteSpecs: elementalistplayer.ChoiceRouteSpecs(),
	})
}

func registerBardRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "bard",
		Choices:          bardplayer.NewChoiceHandler(),
		Skills:           bardplayer.SkillEntries(),
		ChoiceRouteSpecs: bardplayer.ChoiceRouteSpecs(),
	})
}

func registerBeastSamuraiRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "beast_samurai",
		Choices:          beastsamuraiplayer.NewChoiceHandler(),
		Skills:           beastsamuraiplayer.SkillEntries(),
		ChoiceRouteSpecs: beastsamuraiplayer.ChoiceRouteSpecs(),
	})
}

func registerBloodPriestessRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "blood_priestess",
		Choices:          bloodpriestessplayer.NewChoiceHandler(),
		Skills:           bloodpriestessplayer.SkillEntries(),
		ChoiceRouteSpecs: bloodpriestessplayer.ChoiceRouteSpecs(),
	})
}

func registerButterflyDancerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "butterfly_dancer",
		Choices:          butterflydancerplayer.NewChoiceHandler(),
		Skills:           butterflydancerplayer.SkillEntries(),
		ChoiceRouteSpecs: butterflydancerplayer.ChoiceRouteSpecs(),
	})
}

func registerCrimsonKnightRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "crimson_knight",
		Choices:          crimsonknightplayer.NewChoiceHandler(),
		Skills:           crimsonknightplayer.SkillEntries(),
		ChoiceRouteSpecs: crimsonknightplayer.ChoiceRouteSpecs(),
	})
}

func registerElfArcherRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "elf_archer",
		Choices:          elfarcherplayer.NewChoiceHandler(),
		Skills:           elfarcherplayer.SkillEntries(),
		ChoiceRouteSpecs: elfarcherplayer.ChoiceRouteSpecs(),
	})
}

func registerFighterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "fighter",
		Choices:          fighterplayer.NewChoiceHandler(),
		Skills:           fighterplayer.SkillEntries(),
		ChoiceRouteSpecs: fighterplayer.ChoiceRouteSpecs(),
	})
}

func registerHolyBowRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "holy_bow",
		Choices:          holybowplayer.NewChoiceHandler(),
		Skills:           holybowplayer.SkillEntries(),
		ChoiceRouteSpecs: holybowplayer.ChoiceRouteSpecs(),
	})
}

func registerMagicBowRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "magic_bow",
		Choices:          magicbowplayer.NewChoiceHandler(),
		Skills:           magicbowplayer.SkillEntries(),
		ChoiceRouteSpecs: magicbowplayer.ChoiceRouteSpecs(),
	})
}

func registerMagicLancerRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "magic_lancer",
		Choices:          magiclancerplayer.NewChoiceHandler(),
		Skills:           magiclancerplayer.SkillEntries(),
		ChoiceRouteSpecs: magiclancerplayer.ChoiceRouteSpecs(),
	})
}

func registerPrayerMasterRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "prayer_master",
		Choices:          prayermasterplayer.NewChoiceHandler(),
		Skills:           prayermasterplayer.SkillEntries(),
		ChoiceRouteSpecs: prayermasterplayer.ChoiceRouteSpecs(),
	})
}

func registerPriestRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "priest",
		Choices:          priestplayer.NewChoiceHandler(),
		Skills:           priestplayer.SkillEntries(),
		ChoiceRouteSpecs: priestplayer.ChoiceRouteSpecs(),
	})
}

func registerSoulSorcererRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "soul_sorcerer",
		Choices:          soulsocererplayer.NewChoiceHandler(),
		Skills:           soulsocererplayer.SkillEntries(),
		ChoiceRouteSpecs: soulsocererplayer.ChoiceRouteSpecs(),
	})
}

func registerSwordEmperorRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "sword_emperor",
		Choices:          swordemperorplayer.NewChoiceHandler(),
		Skills:           swordemperorplayer.SkillEntries(),
		ChoiceRouteSpecs: swordemperorplayer.ChoiceRouteSpecs(),
	})
}

func registerWarHomunculusRoleEntry(registry *engineplayer.RoleRegistry) {
	if registry == nil {
		return
	}
	registry.Register(engineplayer.RoleEntry{
		ID:               "war_homunculus",
		Choices:          warhomunculusplayer.NewChoiceHandler(),
		Skills:           warhomunculusplayer.SkillEntries(),
		ChoiceRouteSpecs: warhomunculusplayer.ChoiceRouteSpecs(),
	})
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
