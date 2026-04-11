// gameflow: 集中注册所有 skill handler。

package engine

import (
	"starcup-engine/internal/engine/skill"
)

func init() {
	skills.Register("holy_shield", &skills.HolyShieldHandler{})
	// 1. 天使
	skills.Register("angel_bond", &skills.AngelBondHandler{})
	skills.Register("angel_blessing", &skills.AngelBlessingHandler{})
	skills.Register("angel_cleanse", &skills.AngelCleanseHandler{})
	skills.Register("angel_song", &skills.AngelSongHandler{})
	skills.Register("god_protection", &skills.GodProtectionHandler{})
	skills.Register("angel_wall", &skills.AngelWallHandler{})

	// 2. 狂战士
	skills.Register("berserker_frenzy", &skills.BerserkerFrenzyHandler{})
	skills.Register("berserker_tear", &skills.BerserkerTearHandler{})
	skills.Register("blood_roar", &skills.BloodRoarHandler{})
	skills.Register("blood_blade", &skills.BloodBladeHandler{})

	// 3. 封印师
	skills.Register("magic_surge", &skills.MagicSurgeHandler{})
	skills.Register("seal_break", &skills.SealBreakHandler{})
	skills.Register("five_elements_bind", &skills.FiveElementsBindHandler{})
	skills.Register("water_seal", skills.NewWaterSealHandler())
	skills.Register("fire_seal", skills.NewFireSealHandler())
	skills.Register("earth_seal", skills.NewEarthSealHandler())
	skills.Register("wind_seal", skills.NewWindSealHandler())
	skills.Register("thunder_seal", skills.NewThunderSealHandler())

	// 4. 风之剑圣
	skills.Register("wind_fury", &skills.WindFuryHandler{})
	skills.Register("holy_sword", &skills.HolySwordHandler{})
	skills.Register("sword_shadow", &skills.SwordShadowHandler{})
	skills.Register("gale_skill", &skills.GaleSkillHandler{})
	skills.Register("gale_slash", &skills.GaleSlashHandler{})

	// 5. 神箭手
	skills.Register("piercing_shot", &skills.PiercingShotHandler{})
	skills.Register("lightning_arrow", &skills.LightningArrowHandler{})
	skills.Register("snipe", &skills.SnipeHandler{})
	skills.Register("precise_shot", &skills.PreciseShotHandler{})
	skills.Register("flash_trap", &skills.FlashTrapHandler{})

	// 6. 暗杀者
	skills.Register("backlash", &skills.BacklashHandler{})
	skills.Register("water_shadow", &skills.WaterShadowHandler{})
	skills.Register("stealth", &skills.StealthHandler{})

	// 7. 圣女
	skills.Register("frost_prayer", &skills.FrostPrayerHandler{})
	skills.Register("healing_light", &skills.HealingLightHandler{})
	skills.Register("heal", &skills.HealHandler{})
	skills.Register("saint_heal", &skills.SaintHealHandler{})
	skills.Register("mercy", &skills.MercyHandler{})

	// 8. 魔法少女
	skills.Register("magic_bullet_control", &skills.MagicBulletControlHandler{})
	skills.Register("magic_bullet_fusion", &skills.MagicBulletFusionHandler{})
	skills.Register("magic_blast", &skills.MagicBlastHandler{})
	skills.Register("destruction_storm", &skills.DestructionStormHandler{})

	// 9. 女武神
	skills.Register("valkyrie_divine_pursuit", &skills.ValkyrieDivinePursuitHandler{})
	skills.Register("valkyrie_order_seal", &skills.ValkyrieOrderSealHandler{})
	skills.Register("valkyrie_peace_walker", &skills.ValkyriePeaceWalkerHandler{})
	skills.Register("valkyrie_military_glory", &skills.ValkyrieMilitaryGloryHandler{})
	skills.Register("valkyrie_heroic_summon", &skills.ValkyrieHeroicSummonHandler{})

	// 10. 元素师
	skills.Register("elementalist_absorb", &skills.ElementalistAbsorbHandler{})
	skills.Register("elementalist_ignite", &skills.ElementalistIgniteHandler{})
	skills.Register("elementalist_thunder_strike", &skills.ElementalistThunderStrikeHandler{})
	skills.Register("elementalist_freeze", &skills.ElementalistFreezeHandler{})
	skills.Register("elementalist_wind_blade", &skills.ElementalistWindBladeHandler{})
	skills.Register("elementalist_meteor", &skills.ElementalistMeteorHandler{})
	skills.Register("elementalist_fireball", &skills.ElementalistFireballHandler{})
	skills.Register("elementalist_moonlight", &skills.ElementalistMoonlightHandler{})

	// 11. 仲裁者
	skills.Register("arbiter_law", &skills.ArbiterLawHandler{})
	skills.Register("arbiter_judgment_tide", &skills.ArbiterJudgmentTideHandler{})
	skills.Register("arbiter_ritual", &skills.ArbiterRitualHandler{})
	skills.Register("arbiter_ritual_break", &skills.ArbiterRitualBreakHandler{})
	skills.Register("arbiter_doomsday", &skills.ArbiterDoomsdayHandler{})
	skills.Register("arbiter_balance", &skills.ArbiterBalanceHandler{})

	// 12. 冒险家
	skills.Register("adventurer_fraud", &skills.AdventurerFraudHandler{})
	skills.Register("adventurer_lucky_fortune", &skills.AdventurerLuckyFortuneHandler{})
	skills.Register("adventurer_underground_law", &skills.AdventurerUndergroundLawHandler{})
	skills.Register("adventurer_paradise", &skills.AdventurerParadiseHandler{})
	skills.Register("adventurer_steal_sky", &skills.AdventurerStealSkyHandler{})

	// 13. 圣枪骑士
	skills.Register("holy_lancer_revelation", &skills.HolyLancerRevelationHandler{})
	skills.Register("holy_lancer_radiance", &skills.HolyLancerRadianceHandler{})
	skills.Register("holy_lancer_punishment", &skills.HolyLancerPunishmentHandler{})
	skills.Register("holy_lancer_holy_strike", &skills.HolyLancerHolyStrikeHandler{})
	skills.Register("holy_lancer_sky_spear", &skills.HolyLancerSkySpearHandler{})
	skills.Register("holy_lancer_earth_spear", &skills.HolyLancerEarthSpearHandler{})
	skills.Register("holy_lancer_prayer", &skills.HolyLancerPrayerHandler{})

	// 14. 精灵射手
	skills.Register("elf_elemental_shot", &skills.ElfElementalShotHandler{})
	skills.Register("elf_animal_companion", &skills.ElfAnimalCompanionHandler{})
	skills.Register("elf_ritual", &skills.ElfRitualHandler{})
	skills.Register("elf_pet_empower", &skills.ElfPetEmpowerHandler{})

	// 15. 瘟疫法师
	skills.Register("plague_immortal", &skills.PlagueImmortalHandler{})
	skills.Register("plague_blasphemy", &skills.PlagueBlasphemyHandler{})
	skills.Register("plague_outbreak", &skills.PlagueOutbreakHandler{})
	skills.Register("plague_death_touch", &skills.PlagueDeathTouchHandler{})
	skills.Register("plague_toxic_nova", &skills.PlagueToxicNovaHandler{})

	// 16. 魔剑士
	skills.Register("ms_asura_combo", &skills.MagicSwordsmanAsuraComboHandler{})
	skills.Register("ms_shadow_gather", &skills.MagicSwordsmanShadowGatherHandler{})
	skills.Register("ms_shadow_power", &skills.MagicSwordsmanShadowPowerHandler{})
	skills.Register("ms_shadow_reject", &skills.MagicSwordsmanShadowRejectHandler{})
	skills.Register("ms_shadow_meteor", &skills.MagicSwordsmanShadowMeteorHandler{})
	skills.Register("ms_yellow_spring", &skills.MagicSwordsmanYellowSpringHandler{})

	// 17. 血色剑灵
	skills.Register("css_blood_thorns", &skills.CrimsonBloodThornsHandler{})
	skills.Register("css_crimson_flash", &skills.CrimsonFlashHandler{})
	skills.Register("css_blood_rose", &skills.CrimsonBloodRoseHandler{})
	skills.Register("css_blood_barrier", &skills.CrimsonBloodBarrierHandler{})
	skills.Register("css_rose_courtyard", &skills.CrimsonRoseCourtyardHandler{})
	skills.Register("css_dance", &skills.CrimsonDanceHandler{})

	// 18. 祈祷师
	skills.Register("prayer_enter_form", &skills.PrayerEnterFormHandler{})
	skills.Register("prayer_rune_gain", &skills.PrayerRuneGainHandler{})
	skills.Register("prayer_radiant_faith", &skills.PrayerRadiantFaithHandler{})
	skills.Register("prayer_dark_curse", &skills.PrayerDarkCurseHandler{})
	skills.Register("prayer_power_blessing", &skills.PrayerPowerBlessingHandler{})
	skills.Register("prayer_swift_blessing", &skills.PrayerSwiftBlessingHandler{})
	skills.Register("prayer_mana_tide", &skills.PrayerManaTideHandler{})

	// 19. 红莲骑士
	skills.Register("crk_crimson_pact", &skills.CrimsonKnightCrimsonPactHandler{})
	skills.Register("crk_crimson_faith", &skills.CrimsonKnightCrimsonFaithHandler{})
	skills.Register("crk_bloody_prayer", &skills.CrimsonKnightBloodyPrayerHandler{})
	skills.Register("crk_killing_feast", &skills.CrimsonKnightKillingFeastHandler{})
	skills.Register("crk_hot_blood", &skills.CrimsonKnightHotBloodHandler{})
	skills.Register("crk_calm_mind", &skills.CrimsonKnightCalmMindHandler{})
	skills.Register("crk_crimson_cross", &skills.CrimsonKnightCrimsonCrossHandler{})

	// 20. 英灵人形
	skills.Register("hom_battle_pattern", &skills.HomunculusBattlePatternHandler{})
	skills.Register("hom_rage_suppress", &skills.HomunculusRageSuppressHandler{})
	skills.Register("hom_rune_smash", &skills.HomunculusRuneSmashHandler{})
	skills.Register("hom_glyph_fusion", &skills.HomunculusGlyphFusionHandler{})
	skills.Register("hom_rune_reforge", &skills.HomunculusRuneReforgeHandler{})
	skills.Register("hom_dual_echo", &skills.HomunculusDualEchoHandler{})

	// 21. 神官
	skills.Register("priest_divine_revelation", &skills.PriestDivineRevelationHandler{})
	skills.Register("priest_divine_bless", &skills.PriestDivineBlessHandler{})
	skills.Register("priest_water_power", &skills.PriestWaterPowerHandler{})
	skills.Register("priest_guardian", &skills.PriestGuardianHandler{})
	skills.Register("priest_divine_contract", &skills.PriestDivineContractHandler{})
	skills.Register("priest_divine_domain", &skills.PriestDivineDomainHandler{})

	// 22. 阴阳师
	skills.Register("onmyoji_shikigami_descend", &skills.OnmyojiShikigamiDescendHandler{})
	skills.Register("onmyoji_yinyang_shift", &skills.OnmyojiYinYangShiftHandler{})
	skills.Register("onmyoji_shikigami_shift", &skills.OnmyojiShikigamiShiftHandler{})
	skills.Register("onmyoji_dark_ritual", &skills.OnmyojiDarkRitualHandler{})
	skills.Register("onmyoji_binding", &skills.OnmyojiBindingHandler{})
	skills.Register("onmyoji_life_barrier", &skills.OnmyojiLifeBarrierHandler{})

	// 23. 苍炎魔女
	skills.Register("bw_rebirth_clock", &skills.BlazeWitchRebirthClockHandler{})
	skills.Register("bw_blazing_codex", &skills.BlazeWitchBlazingCodexHandler{})
	skills.Register("bw_heavenfire_cleave", &skills.BlazeWitchHeavenfireCleaveHandler{})
	skills.Register("bw_witch_wrath", &skills.BlazeWitchWitchWrathHandler{})
	skills.Register("bw_substitute_doll", &skills.BlazeWitchSubstituteDollHandler{})
	skills.Register("bw_pain_link", &skills.BlazeWitchPainLinkHandler{})
	skills.Register("bw_mana_inversion", &skills.BlazeWitchManaInversionHandler{})

	// 24. 贤者
	skills.Register("sage_wisdom_codex", &skills.SageWisdomCodexHandler{})
	skills.Register("sage_magic_rebound", &skills.SageMagicReboundHandler{})
	skills.Register("sage_arcane_codex", &skills.SageArcaneCodexHandler{})
	skills.Register("sage_holy_codex", &skills.SageHolyCodexHandler{})

	// 25. 魔弓
	skills.Register("mb_magic_pierce", &skills.MagicBowMagicPierceHandler{})
	skills.Register("mb_thunder_scatter", &skills.MagicBowThunderScatterHandler{})
	skills.Register("mb_multi_shot", &skills.MagicBowMultiShotHandler{})
	skills.Register("mb_charge", &skills.MagicBowChargeHandler{})
	skills.Register("mb_demon_eye", &skills.MagicBowDemonEyeHandler{})
	// 内部回调技能：用于"充能"弃牌后的继续流程
	skills.Register("mb_charge_followup_discard", &skills.MagicBowChargeFollowupDiscardHandler{})

	// 26. 魔枪
	skills.Register("ml_dark_release", &skills.MagicLancerDarkReleaseHandler{})
	skills.Register("ml_phantom_stardust", &skills.MagicLancerPhantomStardustHandler{})
	skills.Register("ml_dark_bind", &skills.MagicLancerDarkBindHandler{})
	skills.Register("ml_dark_barrier", &skills.MagicLancerDarkBarrierHandler{})
	skills.Register("ml_fullness", &skills.MagicLancerFullnessHandler{})
	skills.Register("ml_black_spear", &skills.MagicLancerBlackSpearHandler{})

	// 27. 灵符师
	skills.Register("sc_talisman_thunder", &skills.SpiritCasterTalismanThunderHandler{})
	skills.Register("sc_talisman_wind", &skills.SpiritCasterTalismanWindHandler{})
	skills.Register("sc_incantation", &skills.SpiritCasterIncantationHandler{})
	skills.Register("sc_hundred_night", &skills.SpiritCasterHundredNightHandler{})
	skills.Register("sc_spiritual_collapse", &skills.SpiritCasterSpiritualCollapseHandler{})

	// 28. 吟游诗人
	skills.Register("bd_descent_concerto", &skills.BardDescentConcertoHandler{})
	skills.Register("bd_dissonance_chord", &skills.BardDissonanceChordHandler{})
	skills.Register("bd_forbidden_verse", &skills.BardForbiddenVerseHandler{})
	skills.Register("bd_rousing_rhapsody", &skills.BardRousingRhapsodyHandler{})
	skills.Register("bd_victory_symphony", &skills.BardVictorySymphonyHandler{})
	skills.Register("bd_hope_fugue", &skills.BardHopeFugueHandler{})

	// 29. 勇者
	skills.Register("hero_heart", &skills.HeroHeartHandler{})
	skills.Register("hero_roar", &skills.HeroRoarHandler{})
	skills.Register("hero_forbidden_power", &skills.HeroForbiddenPowerHandler{})
	skills.Register("hero_exhaustion", &skills.HeroExhaustionHandler{})
	skills.Register("hero_calm_mind", &skills.HeroCalmMindHandler{})
	skills.Register("hero_taunt", &skills.HeroTauntHandler{})
	skills.Register("hero_dead_duel", &skills.HeroDeadDuelHandler{})

	// 30. 格斗家
	skills.Register("fighter_psi_field", &skills.FighterPsiFieldHandler{})
	skills.Register("fighter_charge_strike", &skills.FighterChargeStrikeHandler{})
	skills.Register("fighter_psi_bullet", &skills.FighterPsiBulletHandler{})
	skills.Register("fighter_hundred_dragon", &skills.FighterHundredDragonHandler{})
	skills.Register("fighter_burst_crash", &skills.FighterBurstCrashHandler{})
	skills.Register("fighter_war_god_drive", &skills.FighterWarGodDriveHandler{})
	skills.Register("fighter_war_god_drive_followup", &skills.FighterWarGodDriveFollowupHandler{})

	// 31. 圣弓
	skills.Register("hb_heavenly_bow", &skills.HolyBowHeavenlyBowHandler{})
	skills.Register("hb_holy_shard_storm", &skills.HolyBowShardStormHandler{})
	skills.Register("hb_radiant_descent", &skills.HolyBowRadiantDescentHandler{})
	skills.Register("hb_light_burst", &skills.HolyBowLightBurstHandler{})
	skills.Register("hb_meteor_bullet", &skills.HolyBowMeteorBulletHandler{})
	skills.Register("hb_radiant_cannon", &skills.HolyBowRadiantCannonHandler{})
	skills.Register("hb_auto_fill", &skills.HolyBowAutoFillHandler{})

	// 32. 剑帝
	skills.Register("se_sword_soul_guard", &skills.SwordEmperorSwordSoulGuardHandler{})
	skills.Register("se_feint", &skills.SwordEmperorFeintHandler{})
	skills.Register("se_sword_qi_slash", &skills.SwordEmperorSwordQiSlashHandler{})
	skills.Register("se_angel_soul", &skills.SwordEmperorAngelSoulHandler{})
	skills.Register("se_demon_soul", &skills.SwordEmperorDemonSoulHandler{})
	skills.Register("se_angel_soul_hit", &skills.SwordEmperorAngelSoulHitHandler{})
	skills.Register("se_angel_soul_miss", &skills.SwordEmperorAngelSoulMissHandler{})
	skills.Register("se_demon_soul_miss", &skills.SwordEmperorDemonSoulMissHandler{})
	skills.Register("se_indomitable_will", &skills.SwordEmperorIndomitableWillHandler{})

	// 33. 兽灵武士
	skills.Register("bs_warrior_zanshin", &skills.BeastSamuraiWarriorZanshinHandler{})
	skills.Register("bs_one_strike_no_thought", &skills.BeastSamuraiOneStrikeNoThoughtHandler{})
	skills.Register("bs_one_strike_intercept", &skills.BeastSamuraiOneStrikeInterceptHandler{})
	skills.Register("bs_beast_soul_will", &skills.BeastSamuraiBeastSoulWillHandler{})
	skills.Register("bs_beast_soul_alert", &skills.BeastSamuraiBeastSoulAlertHandler{})
	skills.Register("bs_beast_return", &skills.BeastSamuraiBeastReturnHandler{})
	skills.Register("bs_iaijutsu_turn_end_drain", &skills.BeastSamuraiIaijutsuTurnEndDrainHandler{})
	skills.Register("bs_iaijutsu_exit_on_deal_damage", &skills.BeastSamuraiIaijutsuExitOnDealDamageHandler{})
	skills.Register("bs_iaijutsu_exit_on_zero", &skills.BeastSamuraiIaijutsuExitOnZeroHandler{})
	skills.Register("bs_iaijutsu_tapped_target_boost", &skills.BeastSamuraiIaijutsuTappedBoostHandler{})
	skills.Register("bs_reversal_iaijutsu", &skills.BeastSamuraiReversalIaijutsuSlashHandler{})
	skills.Register("bs_iaijutsu_style", &skills.BeastSamuraiIaijutsuStyleHandler{})

	// 34. 灵魂术士
	skills.Register("ss_soul_devour", &skills.SoulSorcererSoulDevourHandler{})
	skills.Register("ss_soul_recall", &skills.SoulSorcererSoulRecallHandler{})
	skills.Register("ss_soul_convert", &skills.SoulSorcererSoulConvertHandler{})
	skills.Register("ss_soul_mirror", &skills.SoulSorcererSoulMirrorHandler{})
	skills.Register("ss_soul_blast", &skills.SoulSorcererSoulBlastHandler{})
	skills.Register("ss_soul_grant", &skills.SoulSorcererSoulGrantHandler{})
	skills.Register("ss_soul_link", &skills.SoulSorcererSoulLinkHandler{})
	skills.Register("ss_soul_amp", &skills.SoulSorcererSoulAmpHandler{})

	// 35. 月之女神
	skills.Register("mg_new_moon_shelter", &skills.MoonGoddessNewMoonShelterHandler{})
	skills.Register("mg_dark_moon_curse", &skills.MoonGoddessDarkMoonCurseHandler{})
	skills.Register("mg_medusa_eye", &skills.MoonGoddessMedusaEyeHandler{})
	skills.Register("mg_moon_cycle", &skills.MoonGoddessMoonCycleHandler{})
	skills.Register("mg_blasphemy", &skills.MoonGoddessBlasphemyHandler{})
	skills.Register("mg_dark_moon_slash", &skills.MoonGoddessDarkMoonSlashHandler{})
	skills.Register("mg_pale_moon", &skills.MoonGoddessPaleMoonHandler{})

	// 36. 血之巫女
	skills.Register("bp_blood_sorrow", &skills.BloodPriestessBloodSorrowHandler{})
	skills.Register("bp_bleeding", &skills.BloodPriestessBleedingHandler{})
	skills.Register("bp_backflow", &skills.BloodPriestessBackflowHandler{})
	skills.Register("bp_blood_wail", &skills.BloodPriestessBloodWailHandler{})
	skills.Register("bp_shared_life", &skills.BloodPriestessSharedLifeHandler{})
	skills.Register("bp_blood_curse", &skills.BloodPriestessBloodCurseHandler{})

	// 37. 蝶舞者
	skills.Register("bt_life_fire", &skills.ButterflyLifeFireHandler{})
	skills.Register("bt_dance", &skills.ButterflyDanceHandler{})
	skills.Register("bt_poison_pow", &skills.ButterflyPoisonPowderHandler{})
	skills.Register("bt_pilgrimage", &skills.ButterflyPilgrimageHandler{})
	skills.Register("bt_mirror", &skills.ButterflyMirrorHandler{})
	skills.Register("bt_wither", &skills.ButterflyWitherHandler{})
	skills.Register("bt_chrysalis", &skills.ButterflyChrysalisHandler{})
	skills.Register("bt_reverse_butterfly", &skills.ButterflyReverseHandler{})
}
