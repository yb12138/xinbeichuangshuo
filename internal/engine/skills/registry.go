package skills

import (
	"starcup-engine/internal/model"
	"sync"
)

var (
	registry = make(map[string]model.SkillHandler)
	initOnce sync.Once
)

// Register 注册技能逻辑
func Register(id string, handler model.SkillHandler) {
	if _, exists := registry[id]; exists {
		// In a real app, maybe panic or log warning
		// fmt.Printf("Warning: Skill handler %s already registered\n", id)
		return
	}
	registry[id] = handler
}

// GetHandler 获取技能逻辑
func GetHandler(id string) model.SkillHandler {
	return registry[id]
}

// BaseHandler 基础处理器，简化实现
type BaseHandler struct{}

func (h *BaseHandler) CanUse(ctx *model.Context) bool {
	return true
}

func (h *BaseHandler) Execute(ctx *model.Context) error {
	return nil
}

// InitHandlers 初始化所有技能处理器
func InitHandlers() {
	initOnce.Do(func() {
		Register("holy_shield", &HolyShieldHandler{})
		Register("weakness", &WeaknessHandler{})
		Register("poison", &PoisonHandler{})
		// 1. 天使
		Register("angel_bond", &AngelBondHandler{})
		Register("angel_blessing", &AngelBlessingHandler{})
		Register("angel_cleanse", &AngelCleanseHandler{})
		Register("angel_song", &AngelSongHandler{})
		Register("god_protection", &GodProtectionHandler{})
		Register("angel_wall", &AngelWallHandler{})

		// 2. 狂战士
		// berserker_frenzy is passive, handled directly in game logic
		Register("berserker_tear", &BerserkerTearHandler{})
		Register("blood_roar", &BloodRoarHandler{})
		Register("blood_blade", &BloodBladeHandler{})

		// 3. 封印师
		Register("magic_surge", &MagicSurgeHandler{})
		Register("seal_break", &SealBreakHandler{})
		Register("five_elements_bind", &FiveElementsBindHandler{})
		Register("water_seal", NewWaterSealHandler())
		Register("fire_seal", NewFireSealHandler())
		Register("earth_seal", NewEarthSealHandler())
		Register("wind_seal", NewWindSealHandler())
		Register("thunder_seal", NewThunderSealHandler())

		// 4. 风之剑圣
		Register("wind_fury", &WindFuryHandler{})
		Register("holy_sword", &HolySwordHandler{})
		Register("sword_shadow", &SwordShadowHandler{})
		Register("gale_skill", &GaleSkillHandler{})
		Register("gale_slash", &GaleSlashHandler{})

		// 5. 神箭手
		Register("piercing_shot", &PiercingShotHandler{})
		Register("lightning_arrow", &LightningArrowHandler{})
		Register("snipe", &SnipeHandler{})
		Register("precise_shot", &PreciseShotHandler{})
		Register("flash_trap", &FlashTrapHandler{})

		// 6. 暗杀者
		Register("backlash", &BacklashHandler{})
		Register("water_shadow", &WaterShadowHandler{})
		Register("stealth", &StealthHandler{})

		// 7. 圣女
		Register("frost_prayer", &FrostPrayerHandler{})
		Register("healing_light", &HealingLightHandler{})
		Register("heal", &HealHandler{})
		Register("saint_heal", &SaintHealHandler{})
		Register("mercy", &MercyHandler{})

		// 8. 魔法少女
		Register("magic_bullet_control", &MagicBulletControlHandler{})
		Register("magic_bullet_fusion", &MagicBulletFusionHandler{})
		Register("magic_blast", &MagicBlastHandler{})
		Register("destruction_storm", &DestructionStormHandler{})

		// 9. 女武神
		Register("valkyrie_divine_pursuit", &ValkyrieDivinePursuitHandler{})
		Register("valkyrie_order_seal", &ValkyrieOrderSealHandler{})
		Register("valkyrie_peace_walker", &ValkyriePeaceWalkerHandler{})
		Register("valkyrie_military_glory", &ValkyrieMilitaryGloryHandler{})
		Register("valkyrie_heroic_summon", &ValkyrieHeroicSummonHandler{})

		// 10. 元素师
		Register("elementalist_absorb", &ElementalistAbsorbHandler{})
		Register("elementalist_ignite", &ElementalistIgniteHandler{})
		Register("elementalist_thunder_strike", &ElementalistThunderStrikeHandler{})
		Register("elementalist_freeze", &ElementalistFreezeHandler{})
		Register("elementalist_wind_blade", &ElementalistWindBladeHandler{})
		Register("elementalist_meteor", &ElementalistMeteorHandler{})
		Register("elementalist_fireball", &ElementalistFireballHandler{})
		Register("elementalist_moonlight", &ElementalistMoonlightHandler{})

		// 11. 仲裁者
		Register("arbiter_law", &ArbiterLawHandler{})
		Register("arbiter_judgment_tide", &ArbiterJudgmentTideHandler{})
		Register("arbiter_ritual", &ArbiterRitualHandler{})
		Register("arbiter_ritual_break", &ArbiterRitualBreakHandler{})
		Register("arbiter_doomsday", &ArbiterDoomsdayHandler{})
		Register("arbiter_balance", &ArbiterBalanceHandler{})

		// 12. 冒险家
		Register("adventurer_fraud", &AdventurerFraudHandler{})
		Register("adventurer_lucky_fortune", &AdventurerLuckyFortuneHandler{})
		Register("adventurer_underground_law", &AdventurerUndergroundLawHandler{})
		Register("adventurer_paradise", &AdventurerParadiseHandler{})
		Register("adventurer_steal_sky", &AdventurerStealSkyHandler{})

		// 13. 圣枪骑士
		Register("holy_lancer_revelation", &HolyLancerRevelationHandler{})
		Register("holy_lancer_radiance", &HolyLancerRadianceHandler{})
		Register("holy_lancer_punishment", &HolyLancerPunishmentHandler{})
		Register("holy_lancer_holy_strike", &HolyLancerHolyStrikeHandler{})
		Register("holy_lancer_sky_spear", &HolyLancerSkySpearHandler{})
		Register("holy_lancer_earth_spear", &HolyLancerEarthSpearHandler{})
		Register("holy_lancer_prayer", &HolyLancerPrayerHandler{})

		// 14. 精灵射手
		Register("elf_elemental_shot", &ElfElementalShotHandler{})
		Register("elf_animal_companion", &ElfAnimalCompanionHandler{})
		Register("elf_ritual", &ElfRitualHandler{})
		Register("elf_pet_empower", &ElfPetEmpowerHandler{})

		// 15. 瘟疫法师
		Register("plague_immortal", &PlagueImmortalHandler{})
		Register("plague_blasphemy", &PlagueBlasphemyHandler{})
		Register("plague_outbreak", &PlagueOutbreakHandler{})
		Register("plague_death_touch", &PlagueDeathTouchHandler{})
		Register("plague_toxic_nova", &PlagueToxicNovaHandler{})

		// 16. 魔剑士
		Register("ms_asura_combo", &MagicSwordsmanAsuraComboHandler{})
		Register("ms_shadow_gather", &MagicSwordsmanShadowGatherHandler{})
		Register("ms_shadow_power", &MagicSwordsmanShadowPowerHandler{})
		Register("ms_shadow_reject", &MagicSwordsmanShadowRejectHandler{})
		Register("ms_shadow_meteor", &MagicSwordsmanShadowMeteorHandler{})
		Register("ms_yellow_spring", &MagicSwordsmanYellowSpringHandler{})

		// 17. 血色剑灵
		Register("css_blood_thorns", &CrimsonBloodThornsHandler{})
		Register("css_crimson_flash", &CrimsonFlashHandler{})
		Register("css_blood_rose", &CrimsonBloodRoseHandler{})
		Register("css_blood_barrier", &CrimsonBloodBarrierHandler{})
		Register("css_rose_courtyard", &CrimsonRoseCourtyardHandler{})
		Register("css_dance", &CrimsonDanceHandler{})

		// 18. 祈祷师
		Register("prayer_enter_form", &PrayerEnterFormHandler{})
		Register("prayer_rune_gain", &PrayerRuneGainHandler{})
		Register("prayer_radiant_faith", &PrayerRadiantFaithHandler{})
		Register("prayer_dark_curse", &PrayerDarkCurseHandler{})
		Register("prayer_power_blessing", &PrayerPowerBlessingHandler{})
		Register("prayer_swift_blessing", &PrayerSwiftBlessingHandler{})
		Register("prayer_mana_tide", &PrayerManaTideHandler{})

		// 19. 红莲骑士
		Register("crk_crimson_pact", &CrimsonKnightCrimsonPactHandler{})
		Register("crk_crimson_faith", &CrimsonKnightCrimsonFaithHandler{})
		Register("crk_bloody_prayer", &CrimsonKnightBloodyPrayerHandler{})
		Register("crk_killing_feast", &CrimsonKnightKillingFeastHandler{})
		Register("crk_hot_blood", &CrimsonKnightHotBloodHandler{})
		Register("crk_calm_mind", &CrimsonKnightCalmMindHandler{})
		Register("crk_crimson_cross", &CrimsonKnightCrimsonCrossHandler{})

		// 20. 英灵人形
		Register("hom_battle_pattern", &HomunculusBattlePatternHandler{})
		Register("hom_rage_suppress", &HomunculusRageSuppressHandler{})
		Register("hom_rune_smash", &HomunculusRuneSmashHandler{})
		Register("hom_glyph_fusion", &HomunculusGlyphFusionHandler{})
		Register("hom_rune_reforge", &HomunculusRuneReforgeHandler{})
		Register("hom_dual_echo", &HomunculusDualEchoHandler{})

		// 21. 神官
		Register("priest_divine_revelation", &PriestDivineRevelationHandler{})
		Register("priest_divine_bless", &PriestDivineBlessHandler{})
		Register("priest_water_power", &PriestWaterPowerHandler{})
		Register("priest_guardian", &PriestGuardianHandler{})
		Register("priest_divine_contract", &PriestDivineContractHandler{})
		Register("priest_divine_domain", &PriestDivineDomainHandler{})

		// 22. 阴阳师
		Register("onmyoji_shikigami_descend", &OnmyojiShikigamiDescendHandler{})
		Register("onmyoji_yinyang_shift", &OnmyojiYinYangShiftHandler{})
		Register("onmyoji_shikigami_shift", &OnmyojiShikigamiShiftHandler{})
		Register("onmyoji_dark_ritual", &OnmyojiDarkRitualHandler{})
		Register("onmyoji_binding", &OnmyojiBindingHandler{})
		Register("onmyoji_life_barrier", &OnmyojiLifeBarrierHandler{})

		// 23. 苍炎魔女
		Register("bw_rebirth_clock", &BlazeWitchRebirthClockHandler{})
		Register("bw_blazing_codex", &BlazeWitchBlazingCodexHandler{})
		Register("bw_heavenfire_cleave", &BlazeWitchHeavenfireCleaveHandler{})
		Register("bw_witch_wrath", &BlazeWitchWitchWrathHandler{})
		Register("bw_substitute_doll", &BlazeWitchSubstituteDollHandler{})
		Register("bw_pain_link", &BlazeWitchPainLinkHandler{})
		Register("bw_mana_inversion", &BlazeWitchManaInversionHandler{})

		// 24. 贤者
		Register("sage_wisdom_codex", &SageWisdomCodexHandler{})
		Register("sage_magic_rebound", &SageMagicReboundHandler{})
		Register("sage_arcane_codex", &SageArcaneCodexHandler{})
		Register("sage_holy_codex", &SageHolyCodexHandler{})

		// 25. 魔弓
		Register("mb_magic_pierce", &MagicBowMagicPierceHandler{})
		Register("mb_thunder_scatter", &MagicBowThunderScatterHandler{})
		Register("mb_multi_shot", &MagicBowMultiShotHandler{})
		Register("mb_charge", &MagicBowChargeHandler{})
		Register("mb_demon_eye", &MagicBowDemonEyeHandler{})
		// 内部回调技能：用于“充能”弃牌后的继续流程
		Register("mb_charge_followup_discard", &MagicBowChargeFollowupDiscardHandler{})

		// 26. 魔枪
		Register("ml_dark_release", &MagicLancerDarkReleaseHandler{})
		Register("ml_phantom_stardust", &MagicLancerPhantomStardustHandler{})
		Register("ml_dark_bind", &MagicLancerDarkBindHandler{})
		Register("ml_dark_barrier", &MagicLancerDarkBarrierHandler{})
		Register("ml_fullness", &MagicLancerFullnessHandler{})
		Register("ml_black_spear", &MagicLancerBlackSpearHandler{})

		// 27. 灵符师
		Register("sc_talisman_thunder", &SpiritCasterTalismanThunderHandler{})
		Register("sc_talisman_wind", &SpiritCasterTalismanWindHandler{})
		Register("sc_incantation", &SpiritCasterIncantationHandler{})
		Register("sc_hundred_night", &SpiritCasterHundredNightHandler{})
		Register("sc_spiritual_collapse", &SpiritCasterSpiritualCollapseHandler{})

		// 28. 吟游诗人
		Register("bd_descent_concerto", &BardDescentConcertoHandler{})
		Register("bd_dissonance_chord", &BardDissonanceChordHandler{})
		Register("bd_forbidden_verse", &BardForbiddenVerseHandler{})
		Register("bd_rousing_rhapsody", &BardRousingRhapsodyHandler{})
		Register("bd_victory_symphony", &BardVictorySymphonyHandler{})
		Register("bd_hope_fugue", &BardHopeFugueHandler{})

		// 29. 勇者
		Register("hero_heart", &HeroHeartHandler{})
		Register("hero_roar", &HeroRoarHandler{})
		Register("hero_forbidden_power", &HeroForbiddenPowerHandler{})
		Register("hero_exhaustion", &HeroExhaustionHandler{})
		Register("hero_calm_mind", &HeroCalmMindHandler{})
		Register("hero_taunt", &HeroTauntHandler{})
		Register("hero_dead_duel", &HeroDeadDuelHandler{})

		// 30. 格斗家
		Register("fighter_psi_field", &FighterPsiFieldHandler{})
		Register("fighter_charge_strike", &FighterChargeStrikeHandler{})
		Register("fighter_psi_bullet", &FighterPsiBulletHandler{})
		Register("fighter_hundred_dragon", &FighterHundredDragonHandler{})
		Register("fighter_burst_crash", &FighterBurstCrashHandler{})
		Register("fighter_war_god_drive", &FighterWarGodDriveHandler{})
		Register("fighter_war_god_drive_followup", &FighterWarGodDriveFollowupHandler{})

		// 31. 圣弓
		Register("hb_heavenly_bow", &HolyBowHeavenlyBowHandler{})
		Register("hb_holy_shard_storm", &HolyBowShardStormHandler{})
		Register("hb_radiant_descent", &HolyBowRadiantDescentHandler{})
		Register("hb_light_burst", &HolyBowLightBurstHandler{})
		Register("hb_meteor_bullet", &HolyBowMeteorBulletHandler{})
		Register("hb_radiant_cannon", &HolyBowRadiantCannonHandler{})
		Register("hb_auto_fill", &HolyBowAutoFillHandler{})

		Register("ss_soul_devour", &SoulSorcererSoulDevourHandler{})
		Register("ss_soul_recall", &SoulSorcererSoulRecallHandler{})
		Register("ss_soul_convert", &SoulSorcererSoulConvertHandler{})
		Register("ss_soul_mirror", &SoulSorcererSoulMirrorHandler{})
		Register("ss_soul_blast", &SoulSorcererSoulBlastHandler{})
		Register("ss_soul_grant", &SoulSorcererSoulGrantHandler{})
		Register("ss_soul_link", &SoulSorcererSoulLinkHandler{})
		Register("ss_soul_amp", &SoulSorcererSoulAmpHandler{})

		// 33. 月之女神
		Register("mg_new_moon_shelter", &MoonGoddessNewMoonShelterHandler{})
		Register("mg_dark_moon_curse", &MoonGoddessDarkMoonCurseHandler{})
		Register("mg_medusa_eye", &MoonGoddessMedusaEyeHandler{})
		Register("mg_moon_cycle", &MoonGoddessMoonCycleHandler{})
		Register("mg_blasphemy", &MoonGoddessBlasphemyHandler{})
		Register("mg_dark_moon_slash", &MoonGoddessDarkMoonSlashHandler{})
		Register("mg_pale_moon", &MoonGoddessPaleMoonHandler{})

		// 34. 血之巫女
		Register("bp_blood_sorrow", &BloodPriestessBloodSorrowHandler{})
		Register("bp_bleeding", &BloodPriestessBleedingHandler{})
		Register("bp_backflow", &BloodPriestessBackflowHandler{})
		Register("bp_blood_wail", &BloodPriestessBloodWailHandler{})
		Register("bp_shared_life", &BloodPriestessSharedLifeHandler{})
		Register("bp_blood_curse", &BloodPriestessBloodCurseHandler{})

		// 35. 蝶舞者
		Register("bt_life_fire", &ButterflyLifeFireHandler{})
		Register("bt_dance", &ButterflyDanceHandler{})
		Register("bt_poison_powder", &ButterflyPoisonPowderHandler{})
		Register("bt_pilgrimage", &ButterflyPilgrimageHandler{})
		Register("bt_mirror", &ButterflyMirrorHandler{})
		Register("bt_wither", &ButterflyWitherHandler{})
		Register("bt_chrysalis", &ButterflyChrysalisHandler{})
		Register("bt_reverse_butterfly", &ButterflyReverseHandler{})
	})
}
