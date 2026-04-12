// gameflow: catalog 中每个 choice_type 显式绑定到 Build/Select（由 catalogChoiceArchetypeTable 索引，无前缀推断）。

package engine

import (
	"fmt"
)

func selectOnmyoji(ge *GameEngine, _ string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleOnmyojiChoiceInput(idx, choiceCtxAsInterfaceMap(ctx))
}

func selectBeast(ge *GameEngine, _ string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleBeastSamuraiChoiceInput(idx, choiceCtxAsInterfaceMap(ctx))
}

func selectSealerChoice(ge *GameEngine, _ string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleSealerChoiceInput("", idx, choiceCtxAsInterfaceMap(ctx))
}

func systemChoiceSelect(typ string) func(*GameEngine, string, int, map[string]any) (bool, error) {
	switch typ {
	case "weak":
		return func(ge *GameEngine, pid string, idx int, _ map[string]any) (bool, error) {
			return true, ge.handleSystemWeakChoice(pid, idx)
		}
	case "buy_resource":
		return func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
			return true, ge.handleSystemBuyResourceChoice(pid, idx, choiceCtxAsInterfaceMap(ctx))
		}
	case "heal":
		return func(ge *GameEngine, _ string, idx int, ctx map[string]any) (bool, error) {
			return true, ge.handleSystemHealChoice(idx, choiceCtxAsInterfaceMap(ctx))
		}
	case "basic_effect_pick":
		return func(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
			return true, ge.handleBasicEffectChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
		}
	default:
		panic(fmt.Sprintf("choice: systemChoiceSelect unknown type %q", typ))
	}
}

func selectChoiceTargetBard(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleBardChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
}

func selectChoiceTargetElf(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleElfArcherChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
}

func selectChoiceTargetPriest(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
	return ge.handlePriestChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
}

func selectChoiceTargetOnmyoji(ge *GameEngine, _ string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleOnmyojiChoiceInput(idx, choiceCtxAsInterfaceMap(ctx))
}

func selectChoiceTargetMagicBow(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleMagicBowChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
}

func selectChoiceTargetSpiritCaster(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleSpiritCasterChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
}

func selectChoiceTargetMagicLancer(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleMagicLancerChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
}

func selectChoiceTargetSwordEmperor(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleSwordEmperorChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
}

func selectChoiceTargetFighter(ge *GameEngine, pid string, idx int, ctx map[string]any) (bool, error) {
	return ge.handleFighterChoiceInput(pid, idx, choiceCtxAsInterfaceMap(ctx))
}

func catalogChoiceBinding(typ string) catalogSpecPlan {
	arch, ok := catalogChoiceArchetypeTable[typ]
	if !ok {
		panic(fmt.Sprintf("choice catalog: no archetype row for type %q (sync choice_catalog_archetype_map.go)", typ))
	}
	m := multiSequential(typ)
	switch arch {
	case "system":
		return catalogSpecPlan{(*GameEngine).buildSystemChoicePrompt, systemChoiceSelect(typ), m}
	case "five_elements_bind":
		return catalogSpecPlan{(*GameEngine).buildSealerChoicePrompt, selectSealerChoice, m}
	case "angel":
		return catalogSpecPlan{(*GameEngine).buildAngelChoicePrompt, (*GameEngine).handleAngelChoiceInput, m}
	case "saintess":
		return catalogSpecPlan{(*GameEngine).buildSaintessChoicePrompt, (*GameEngine).handleSaintessChoiceInput, m}
	case "target_prompt_bard":
		return catalogSpecPlan{(*GameEngine).buildTargetChoicePrompt, selectChoiceTargetBard, m}
	case "target_prompt_elf":
		return catalogSpecPlan{(*GameEngine).buildTargetChoicePrompt, selectChoiceTargetElf, m}
	case "target_prompt_priest":
		return catalogSpecPlan{(*GameEngine).buildTargetChoicePrompt, selectChoiceTargetPriest, m}
	case "target_prompt_onmyoji":
		return catalogSpecPlan{(*GameEngine).buildTargetChoicePrompt, selectChoiceTargetOnmyoji, m}
	case "target_prompt_mb":
		return catalogSpecPlan{(*GameEngine).buildTargetChoicePrompt, selectChoiceTargetMagicBow, m}
	case "target_prompt_sc":
		return catalogSpecPlan{(*GameEngine).buildTargetChoicePrompt, selectChoiceTargetSpiritCaster, m}
	case "target_prompt_ml":
		return catalogSpecPlan{(*GameEngine).buildTargetChoicePrompt, selectChoiceTargetMagicLancer, m}
	case "target_prompt_se":
		return catalogSpecPlan{(*GameEngine).buildTargetChoicePrompt, selectChoiceTargetSwordEmperor, m}
	case "target_prompt_fighter":
		return catalogSpecPlan{(*GameEngine).buildTargetChoicePrompt, selectChoiceTargetFighter, m}
	case "onmyoji":
		return catalogSpecPlan{(*GameEngine).buildOnmyojiChoicePrompt, selectOnmyoji, m}
	case "beast":
		return catalogSpecPlan{(*GameEngine).buildBeastSamuraiChoicePrompt, selectBeast, m}
	case "sage":
		return catalogSpecPlan{(*GameEngine).buildSageChoicePrompt, (*GameEngine).handleSageChoiceInput, m}
	case "adventurer":
		return catalogSpecPlan{(*GameEngine).buildAdventurerChoicePrompt, (*GameEngine).handleAdventurerChoiceInput, m}
	case "priest":
		return catalogSpecPlan{(*GameEngine).buildPriestChoicePrompt, (*GameEngine).handlePriestChoiceInput, m}
	case "prayer_master":
		return catalogSpecPlan{(*GameEngine).buildPrayerMasterChoicePrompt, (*GameEngine).handlePrayerMasterChoiceInput, m}
	case "crimson_knight":
		return catalogSpecPlan{(*GameEngine).buildCrimsonKnightChoicePrompt, (*GameEngine).handleCrimsonKnightChoiceInput, m}
	case "homunculus":
		return catalogSpecPlan{(*GameEngine).buildWarHomunculusChoicePrompt, (*GameEngine).handleWarHomunculusChoiceInput, m}
	case "valkyrie":
		return catalogSpecPlan{(*GameEngine).buildValkyrieChoicePrompt, (*GameEngine).handleValkyrieChoiceInput, m}
	case "elementalist":
		return catalogSpecPlan{(*GameEngine).buildElementalistChoicePrompt, (*GameEngine).handleElementalistChoiceInput, m}
	case "elf_archer":
		return catalogSpecPlan{(*GameEngine).buildElfArcherChoicePrompt, (*GameEngine).handleElfArcherChoiceInput, m}
	case "magic_bow":
		return catalogSpecPlan{(*GameEngine).buildMagicBowChoicePrompt, (*GameEngine).handleMagicBowChoiceInput, m}
	case "sword_emperor":
		return catalogSpecPlan{(*GameEngine).buildSwordEmperorChoicePrompt, (*GameEngine).handleSwordEmperorChoiceInput, m}
	case "magic_lancer":
		return catalogSpecPlan{(*GameEngine).buildMagicLancerChoicePrompt, (*GameEngine).handleMagicLancerChoiceInput, m}
	case "soul_sorcerer":
		return catalogSpecPlan{(*GameEngine).buildSoulSorcererChoicePrompt, (*GameEngine).handleSoulSorcererChoiceInput, m}
	case "moon_goddess":
		return catalogSpecPlan{(*GameEngine).buildMoonGoddessChoicePrompt, (*GameEngine).handleMoonGoddessChoiceInput, m}
	case "blood_priestess":
		return catalogSpecPlan{(*GameEngine).buildBloodPriestessChoicePrompt, (*GameEngine).handleBloodPriestessChoiceInput, m}
	case "butterfly":
		return catalogSpecPlan{(*GameEngine).buildButterflyChoicePrompt, (*GameEngine).handleButterflyChoiceInput, m}
	case "spirit_caster":
		return catalogSpecPlan{(*GameEngine).buildSpiritCasterChoicePrompt, (*GameEngine).handleSpiritCasterChoiceInput, m}
	case "bard":
		return catalogSpecPlan{(*GameEngine).buildBardChoicePrompt, (*GameEngine).handleBardChoiceInput, m}
	case "holy_bow":
		return catalogSpecPlan{(*GameEngine).buildHolyBowChoicePrompt, (*GameEngine).handleHolyBowChoiceInput, m}
	case "hero":
		return catalogSpecPlan{(*GameEngine).buildHeroChoicePrompt, (*GameEngine).handleHeroChoiceInput, m}
	case "assassin":
		return catalogSpecPlan{(*GameEngine).buildAssassinChoicePrompt, (*GameEngine).handleAssassinChoiceInput, m}
	case "arbiter":
		return catalogSpecPlan{(*GameEngine).buildArbiterChoicePrompt, (*GameEngine).handleArbiterChoiceInput, m}
	case "holy_lancer":
		return catalogSpecPlan{(*GameEngine).buildHolyLancerChoicePrompt, (*GameEngine).handleHolyLancerChoiceInput, m}
	case "sealer":
		return catalogSpecPlan{(*GameEngine).buildSealerChoicePrompt, (*GameEngine).handleSealerChoiceInput, m}
	case "plague_mage":
		return catalogSpecPlan{(*GameEngine).buildPlagueMageChoicePrompt, (*GameEngine).handlePlagueMageChoiceInput, m}
	case "magic_swordsman":
		return catalogSpecPlan{(*GameEngine).buildMagicSwordsmanChoicePrompt, (*GameEngine).handleMagicSwordsmanChoiceInput, m}
	case "crimson_sword_spirit":
		return catalogSpecPlan{(*GameEngine).buildCrimsonSwordSpiritChoicePrompt, (*GameEngine).handleCrimsonSwordSpiritChoiceInput, m}
	case "blaze_witch":
		return catalogSpecPlan{(*GameEngine).buildBlazeWitchChoicePrompt, (*GameEngine).handleBlazeWitchChoiceInput, m}
	default:
		panic(fmt.Sprintf("choice catalog: unknown archetype %q for type %q", arch, typ))
	}
}
