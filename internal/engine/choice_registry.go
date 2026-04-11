// gameflow: 集中注册所有 choice handler 和 prompt builder。

package engine

func init() {
	// ==================== System Choices ====================
	RegisterChoicePrompt("weak", (*GameEngine).buildSystemChoicePrompt)
	RegisterChoicePrompt("buy_resource", (*GameEngine).buildSystemChoicePrompt)
	RegisterChoicePrompt("heal", (*GameEngine).buildSystemChoicePrompt)
	RegisterChoicePrompt("basic_effect_pick", (*GameEngine).buildSystemChoicePrompt)
	RegisterChoicePrompt("extract", (*GameEngine).buildSystemChoicePrompt)
	RegisterChoiceSingleHandler("weak", (*GameEngine).handleSystemChoiceInput)
	RegisterChoiceSingleHandler("buy_resource", (*GameEngine).handleSystemChoiceInput)
	RegisterChoiceSingleHandler("heal", (*GameEngine).handleSystemChoiceInput)
	RegisterChoiceSingleHandler("basic_effect_pick", (*GameEngine).handleSystemChoiceInput)
	RegisterChoiceMultiHandler("extract", (*GameEngine).handleExtractChoiceSelections)
	RegisterChoiceCancelHandler("extract", (*GameEngine).cancelExtractChoice)

	// ==================== Onmyoji ====================
	RegisterChoicePrompt("onmyoji_life_barrier_mode", (*GameEngine).buildOnmyojiChoicePrompt)
	RegisterChoicePrompt("onmyoji_shikigami_pick", (*GameEngine).buildOnmyojiChoicePrompt)
	RegisterChoicePrompt("onmyoji_yinyang_shift_pick", (*GameEngine).buildOnmyojiChoicePrompt)
	RegisterChoicePrompt("onmyoji_dark_ritual_pick", (*GameEngine).buildOnmyojiChoicePrompt)
	RegisterChoicePrompt("onmyoji_binding_pick", (*GameEngine).buildOnmyojiChoicePrompt)
	RegisterChoicePrompt("onmyoji_shikigami_shift_pick", (*GameEngine).buildOnmyojiChoicePrompt)
	RegisterChoiceSingleHandler("onmyoji_life_barrier_mode", func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
		return e.handleOnmyojiChoiceInput(selectionIndex, ctxData)
	})
	RegisterChoiceSingleHandler("onmyoji_shikigami_pick", func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
		return e.handleOnmyojiChoiceInput(selectionIndex, ctxData)
	})
	RegisterChoiceSingleHandler("onmyoji_yinyang_shift_pick", func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
		return e.handleOnmyojiChoiceInput(selectionIndex, ctxData)
	})
	RegisterChoiceSingleHandler("onmyoji_dark_ritual_pick", func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
		return e.handleOnmyojiChoiceInput(selectionIndex, ctxData)
	})
	RegisterChoiceSingleHandler("onmyoji_binding_pick", func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
		return e.handleOnmyojiChoiceInput(selectionIndex, ctxData)
	})
	RegisterChoiceSingleHandler("onmyoji_shikigami_shift_pick", func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
		return e.handleOnmyojiChoiceInput(selectionIndex, ctxData)
	})

	// ==================== Beast Samurai ====================
	RegisterChoicePrompt("bs_iaijutsu_draw_pick", (*GameEngine).buildBeastSamuraiChoicePrompt)
	RegisterChoicePrompt("bs_iaijutsu_mode_pick", (*GameEngine).buildBeastSamuraiChoicePrompt)
	RegisterChoiceSingleHandler("bs_iaijutsu_draw_pick", func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
		return e.handleBeastSamuraiChoiceInput(selectionIndex, ctxData)
	})
	RegisterChoiceSingleHandler("bs_iaijutsu_mode_pick", func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
		return e.handleBeastSamuraiChoiceInput(selectionIndex, ctxData)
	})

	// ==================== Sage ====================
	RegisterChoicePrompt("sage_magic_rebound_confirm", (*GameEngine).buildSageChoicePrompt)
	RegisterChoicePrompt("sage_magic_rebound_x", (*GameEngine).buildSageChoicePrompt)
	RegisterChoicePrompt("sage_magic_rebound_element", (*GameEngine).buildSageChoicePrompt)
	RegisterChoicePrompt("sage_magic_rebound_cards", (*GameEngine).buildSageChoicePrompt)
	RegisterChoicePrompt("sage_arcane_cards", (*GameEngine).buildSageChoicePrompt)
	RegisterChoicePrompt("sage_holy_cards", (*GameEngine).buildSageChoicePrompt)
	RegisterChoiceSingleHandler("sage_magic_rebound_confirm", (*GameEngine).handleSageChoiceInput)
	RegisterChoiceSingleHandler("sage_magic_rebound_x", (*GameEngine).handleSageChoiceInput)
	RegisterChoiceSingleHandler("sage_magic_rebound_element", (*GameEngine).handleSageChoiceInput)
	RegisterChoiceSingleHandler("sage_magic_rebound_cards", (*GameEngine).handleSageChoiceInput)
	RegisterChoiceSingleHandler("sage_arcane_cards", (*GameEngine).handleSageChoiceInput)
	RegisterChoiceSingleHandler("sage_holy_cards", (*GameEngine).handleSageChoiceInput)

	// ==================== Adventurer ====================
	RegisterChoicePrompt("adventurer_fraud_pick", (*GameEngine).buildAdventurerChoicePrompt)
	RegisterChoicePrompt("adventurer_paradise_pick", (*GameEngine).buildAdventurerChoicePrompt)
	RegisterChoiceSingleHandler("adventurer_fraud_pick", (*GameEngine).handleAdventurerChoiceInput)
	RegisterChoiceSingleHandler("adventurer_paradise_pick", (*GameEngine).handleAdventurerChoiceInput)

	// ==================== Priest ====================
	RegisterChoicePrompt("priest_divine_revelation_pick", (*GameEngine).buildPriestChoicePrompt)
	RegisterChoicePrompt("priest_divine_domain_pick", (*GameEngine).buildPriestChoicePrompt)
	RegisterChoiceSingleHandler("priest_divine_revelation_pick", (*GameEngine).handlePriestChoiceInput)
	RegisterChoiceSingleHandler("priest_divine_domain_pick", (*GameEngine).handlePriestChoiceInput)

	// ==================== Prayer Master ====================
	RegisterChoicePrompt("prayer_master_rune_pick", (*GameEngine).buildPrayerMasterChoicePrompt)
	RegisterChoicePrompt("prayer_master_extra_action_pick", (*GameEngine).buildPrayerMasterChoicePrompt)
	RegisterChoiceSingleHandler("prayer_master_rune_pick", (*GameEngine).handlePrayerMasterChoiceInput)
	RegisterChoiceSingleHandler("prayer_master_extra_action_pick", (*GameEngine).handlePrayerMasterChoiceInput)

	// ==================== Crimson Knight ====================
	RegisterChoicePrompt("crk_bloody_prayer_pick", (*GameEngine).buildCrimsonKnightChoicePrompt)
	RegisterChoiceSingleHandler("crk_bloody_prayer_pick", (*GameEngine).handleCrimsonKnightChoiceInput)

	// ==================== War Homunculus ====================
	RegisterChoicePrompt("hom_dual_echo_target", (*GameEngine).buildWarHomunculusChoicePrompt)
	RegisterChoicePrompt("hom_rune_reforge_pick", (*GameEngine).buildWarHomunculusChoicePrompt)
	RegisterChoicePrompt("hom_glyph_fusion_pick", (*GameEngine).buildWarHomunculusChoicePrompt)
	RegisterChoiceSingleHandler("hom_dual_echo_target", (*GameEngine).handleWarHomunculusChoiceInput)
	RegisterChoiceSingleHandler("hom_rune_reforge_pick", (*GameEngine).handleWarHomunculusChoiceInput)
	RegisterChoiceSingleHandler("hom_glyph_fusion_pick", (*GameEngine).handleWarHomunculusChoiceInput)
	RegisterChoiceCancelHandler("hom_dual_echo_target", (*GameEngine).cancelHomDualEchoChoice)

	// ==================== Valkyrie ====================
	RegisterChoicePrompt("valkyrie_heroic_summon_pick", (*GameEngine).buildValkyrieChoicePrompt)
	RegisterChoiceSingleHandler("valkyrie_heroic_summon_pick", (*GameEngine).handleValkyrieChoiceInput)

	// ==================== Elementalist ====================
	RegisterChoicePrompt("elementalist_primordial_pick", (*GameEngine).buildElementalistChoicePrompt)
	RegisterChoiceSingleHandler("elementalist_primordial_pick", (*GameEngine).handleElementalistChoiceInput)

	// ==================== Elf Archer ====================
	RegisterChoicePrompt("elf_archer_pet_pick", (*GameEngine).buildElfArcherChoicePrompt)
	RegisterChoicePrompt("elf_archer_elemental_shot_pick", (*GameEngine).buildElfArcherChoicePrompt)
	RegisterChoiceSingleHandler("elf_archer_pet_pick", (*GameEngine).handleElfArcherChoiceInput)
	RegisterChoiceSingleHandler("elf_archer_elemental_shot_pick", (*GameEngine).handleElfArcherChoiceInput)

	// ==================== Magic Bow ====================
	RegisterChoicePrompt("mb_charge_discard_pick", (*GameEngine).buildMagicBowChoicePrompt)
	RegisterChoicePrompt("mb_demon_eye_pick", (*GameEngine).buildMagicBowChoicePrompt)
	RegisterChoiceSingleHandler("mb_charge_discard_pick", (*GameEngine).handleMagicBowChoiceInput)
	RegisterChoiceSingleHandler("mb_demon_eye_pick", (*GameEngine).handleMagicBowChoiceInput)

	// ==================== Sword Emperor ====================
	RegisterChoicePrompt("se_soul_pick", (*GameEngine).buildSwordEmperorChoicePrompt)
	RegisterChoiceSingleHandler("se_soul_pick", (*GameEngine).handleSwordEmperorChoiceInput)

	// ==================== Magic Lancer ====================
	RegisterChoicePrompt("ml_dark_release_pick", (*GameEngine).buildMagicLancerChoicePrompt)
	RegisterChoicePrompt("ml_phantom_stardust_pick", (*GameEngine).buildMagicLancerChoicePrompt)
	RegisterChoiceSingleHandler("ml_dark_release_pick", (*GameEngine).handleMagicLancerChoiceInput)
	RegisterChoiceSingleHandler("ml_phantom_stardust_pick", (*GameEngine).handleMagicLancerChoiceInput)

	// ==================== Soul Sorcerer ====================
	RegisterChoicePrompt("ss_soul_devour_pick", (*GameEngine).buildSoulSorcererChoicePrompt)
	RegisterChoicePrompt("ss_soul_recall_pick", (*GameEngine).buildSoulSorcererChoicePrompt)
	RegisterChoiceSingleHandler("ss_soul_devour_pick", (*GameEngine).handleSoulSorcererChoiceInput)
	RegisterChoiceSingleHandler("ss_soul_recall_pick", (*GameEngine).handleSoulSorcererChoiceInput)
	RegisterChoiceMultiHandler("ss_recall_pick", (*GameEngine).handleSoulRecallSelections)

	// ==================== Moon Goddess ====================
	RegisterChoicePrompt("mg_dark_moon_curse_pick", (*GameEngine).buildMoonGoddessChoicePrompt)
	RegisterChoiceSingleHandler("mg_dark_moon_curse_pick", (*GameEngine).handleMoonGoddessChoiceInput)

	// ==================== Blood Priestess ====================
	RegisterChoicePrompt("bp_shared_life_pick", (*GameEngine).buildBloodPriestessChoicePrompt)
	RegisterChoicePrompt("bp_blood_curse_pick", (*GameEngine).buildBloodPriestessChoicePrompt)
	RegisterChoiceSingleHandler("bp_shared_life_pick", (*GameEngine).handleBloodPriestessChoiceInput)
	RegisterChoiceSingleHandler("bp_blood_curse_pick", (*GameEngine).handleBloodPriestessChoiceInput)
	RegisterChoiceMultiHandler("bp_curse_discard", (*GameEngine).handleBloodCurseDiscardSelections)

	// ==================== Butterfly Dancer ====================
	RegisterChoicePrompt("bt_cocoon_pick", (*GameEngine).buildButterflyChoicePrompt)
	RegisterChoicePrompt("bt_reverse_branch1_pick", (*GameEngine).buildButterflyChoicePrompt)
	RegisterChoicePrompt("bt_reverse_branch2_pick", (*GameEngine).buildButterflyChoicePrompt)
	RegisterChoiceSingleHandler("bt_cocoon_pick", (*GameEngine).handleButterflyChoiceInput)
	RegisterChoiceSingleHandler("bt_reverse_branch1_pick", (*GameEngine).handleButterflyChoiceInput)
	RegisterChoiceSingleHandler("bt_reverse_branch2_pick", (*GameEngine).handleButterflyChoiceInput)
	RegisterChoiceMultiHandler("bt_cocoon_overflow_discard", (*GameEngine).handleButterflyCocoonOverflowSelections)
	RegisterChoiceMultiHandler("bt_reverse_branch2_pick", (*GameEngine).handleButterflyReverseBranch2PickSelections)

	// ==================== Spirit Caster ====================
	RegisterChoicePrompt("sc_talisman_pick", (*GameEngine).buildSpiritCasterChoicePrompt)
	RegisterChoiceSingleHandler("sc_talisman_pick", (*GameEngine).handleSpiritCasterChoiceInput)

	// ==================== Bard ====================
	RegisterChoicePrompt("bd_dissonance_pick", (*GameEngine).buildBardChoicePrompt)
	RegisterChoicePrompt("bd_forbidden_verse_pick", (*GameEngine).buildBardChoicePrompt)
	RegisterChoiceSingleHandler("bd_dissonance_pick", (*GameEngine).handleBardChoiceInput)
	RegisterChoiceSingleHandler("bd_forbidden_verse_pick", (*GameEngine).handleBardChoiceInput)

	// ==================== Holy Bow ====================
	RegisterChoicePrompt("hb_radiant_descent_pick", (*GameEngine).buildHolyBowChoicePrompt)
	RegisterChoiceSingleHandler("hb_radiant_descent_pick", (*GameEngine).handleHolyBowChoiceInput)

	// ==================== Hero Assassin ====================
	RegisterChoicePrompt("hero_forbidden_power_pick", (*GameEngine).buildHeroAssassinChoicePrompt)
	RegisterChoicePrompt("assassin_stealth_draw", (*GameEngine).buildHeroAssassinChoicePrompt)
	RegisterChoiceSingleHandler("hero_forbidden_power_pick", (*GameEngine).handleHeroAssassinChoiceInput)
	RegisterChoiceSingleHandler("assassin_stealth_draw", (*GameEngine).handleHeroAssassinChoiceInput)

	// ==================== Arbiter ====================
	RegisterChoicePrompt("arbiter_ritual_pick", (*GameEngine).buildArbiterChoicePrompt)
	RegisterChoicePrompt("arbiter_law_pick", (*GameEngine).buildArbiterChoicePrompt)
	RegisterChoiceSingleHandler("arbiter_ritual_pick", (*GameEngine).handleArbiterChoiceInput)
	RegisterChoiceSingleHandler("arbiter_law_pick", (*GameEngine).handleArbiterChoiceInput)

	// ==================== Guardian Support ====================
	RegisterChoicePrompt("guardian_support_pick", (*GameEngine).buildGuardianSupportChoicePrompt)
	RegisterChoiceSingleHandler("guardian_support_pick", (*GameEngine).handleGuardianSupportChoiceInput)

	// ==================== Holy Lancer ====================
	RegisterChoicePrompt("holy_lancer_revelation_pick", (*GameEngine).buildHolyLancerChoicePrompt)
	RegisterChoiceSingleHandler("holy_lancer_revelation_pick", (*GameEngine).handleHolyLancerChoiceInput)

	// ==================== Sealer ====================
	RegisterChoicePrompt("sealer_five_elements_bind_pick", (*GameEngine).buildSealerChoicePrompt)
	RegisterChoiceSingleHandler("sealer_five_elements_bind_pick", (*GameEngine).handleSealerChoiceInput)

	// ==================== Plague Mage ====================
	RegisterChoicePrompt("plague_mage_toxic_nova_pick", (*GameEngine).buildPlagueMageChoicePrompt)
	RegisterChoiceSingleHandler("plague_mage_toxic_nova_pick", (*GameEngine).handlePlagueMageChoiceInput)

	// ==================== Magic Swordsman ====================
	RegisterChoicePrompt("ms_asura_combo_pick", (*GameEngine).buildMagicSwordsmanChoicePrompt)
	RegisterChoicePrompt("ms_shadow_gather_pick", (*GameEngine).buildMagicSwordsmanChoicePrompt)
	RegisterChoiceSingleHandler("ms_asura_combo_pick", (*GameEngine).handleMagicSwordsmanChoiceInput)
	RegisterChoiceSingleHandler("ms_shadow_gather_pick", (*GameEngine).handleMagicSwordsmanChoiceInput)

	// ==================== Crimson Sword Spirit ====================
	RegisterChoicePrompt("css_rose_courtyard_pick", (*GameEngine).buildCrimsonSwordSpiritChoicePrompt)
	RegisterChoiceSingleHandler("css_rose_courtyard_pick", (*GameEngine).handleCrimsonSwordSpiritChoiceInput)

	// ==================== Blaze Witch ====================
	RegisterChoicePrompt("bw_pain_link_pick", (*GameEngine).buildBlazeWitchChoicePrompt)
	RegisterChoiceSingleHandler("bw_pain_link_pick", (*GameEngine).handleBlazeWitchChoiceInput)

	// ==================== Target Choice ====================
	RegisterChoicePrompt("angel_bond_heal_target", (*GameEngine).buildTargetChoicePrompt)
	RegisterChoicePrompt("frost_prayer_target", (*GameEngine).buildTargetChoicePrompt)
	RegisterChoicePrompt("god_protection_x", (*GameEngine).buildTargetChoicePrompt)
	RegisterChoicePrompt("piercing_shot_discard_pick", (*GameEngine).buildTargetChoicePrompt)
	RegisterChoicePrompt("water_shadow_discard_pick", (*GameEngine).buildTargetChoicePrompt)
	RegisterChoicePrompt("saint_heal_pick", (*GameEngine).buildTargetChoicePrompt)
	RegisterChoicePrompt("basic_effect_pick", (*GameEngine).buildTargetChoicePrompt)
	RegisterChoicePrompt("target_choice_pick", (*GameEngine).buildTargetChoicePrompt)
}
