// Package rolecatalog wires concrete player role modules into the role registry.
package rolecatalog

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
	moonplayer "starcup-engine/internal/engine/player/moon_goddess"
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
)

func BuildRoleRegistry() *engineplayer.RoleRegistry {
	registry := engineplayer.NewRoleRegistry()
	registry.Register(angelplayer.RoleEntry())
	registry.Register(arbiterplayer.RoleEntry())
	registry.Register(bardplayer.RoleEntry())
	registry.Register(beastsamuraiplayer.RoleEntry())
	registry.Register(blazewitchplayer.RoleEntry())
	registry.Register(bloodpriestessplayer.RoleEntry())
	registry.Register(butterflydancerplayer.RoleEntry())
	registry.Register(crimsonknightplayer.RoleEntry())
	registry.Register(crimsonswordspiritplayer.RoleEntry())
	registry.Register(elfarcherplayer.RoleEntry())
	registry.Register(fighterplayer.RoleEntry())
	registry.Register(holybowplayer.RoleEntry())
	registry.Register(holylancerplayer.RoleEntry())
	registry.Register(magicbowplayer.RoleEntry())
	registry.Register(magiclancerplayer.RoleEntry())
	registry.Register(magicswordsmanplayer.RoleEntry())
	registry.Register(moonplayer.RoleEntry())
	registry.Register(onmyojiproplayer.RoleEntry())
	registry.Register(plagueplayer.RoleEntry())
	registry.Register(prayermasterplayer.RoleEntry())
	registry.Register(priestplayer.RoleEntry())
	registry.Register(sageplayer.RoleEntry())
	registry.Register(sealerplayer.RoleEntry())
	registry.Register(saintessplayer.RoleEntry())
	registry.Register(soulsocererplayer.RoleEntry())
	registry.Register(spiritcasterplayer.RoleEntry())
	registry.Register(swordemperorplayer.RoleEntry())
	registry.Register(warhomunculusplayer.RoleEntry())
	registry.Register(adventurerplayer.RoleEntry())
	registry.Register(assassinplayer.RoleEntry())
	registry.Register(heroplayer.RoleEntry())
	registry.Register(berserkerplayer.RoleEntry())
	registry.Register(magicalgirlplayer.RoleEntry())
	registry.Register(blademasterplayer.RoleEntry())
	registry.Register(archerplayer.RoleEntry())
	registry.Register(valkyrieplayer.RoleEntry())
	registry.Register(elementalistplayer.RoleEntry())
	return registry
}
