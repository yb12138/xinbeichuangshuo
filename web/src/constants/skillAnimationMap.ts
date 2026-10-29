// 技能 → 动画模板映射（纯数据）。生成依据与推断等级见 docs/skill_animation_map.md。
export type SkillOpening = 'active' | 'trigger'
export type SkillEffectTemplate = 'strike' | 'heal' | 'status' | 'resource' | 'form' | 'global'
export interface SkillAnimationSpec { opening: SkillOpening; effects: SkillEffectTemplate[]; element?: 'Water'|'Fire'|'Earth'|'Wind'|'Thunder'|'Light'|'Dark'; signature?: boolean }

/** key = skill ID from internal/data/characters.go */
export const SKILL_ANIMATION_MAP: Record<string, SkillAnimationSpec> = {
  // 天使 (angel)
  angel_bond: { opening: 'trigger', effects: ['heal'] }, // 天使羁绊
  angel_blessing: { opening: 'active', effects: ['resource'], element: 'Water' }, // 天使祝福
  angel_cleanse: { opening: 'active', effects: ['status'], element: 'Wind' }, // 风之洁净
  angel_song: { opening: 'trigger', effects: ['status'] }, // 天使之歌
  god_protection: { opening: 'trigger', effects: ['global'] }, // 神之庇护
  angel_wall: { opening: 'active', effects: ['status'] }, // 天使之墙
  // 狂战士 (berserker)
  berserker_frenzy: { opening: 'trigger', effects: ['strike'] }, // 狂化
  berserker_tear: { opening: 'trigger', effects: ['strike'] }, // 撕裂
  blood_roar: { opening: 'trigger', effects: ['strike'] }, // 血腥咆哮
  blood_blade: { opening: 'trigger', effects: ['strike'] }, // 血影狂刀
  // 封印师 (sealer)
  magic_surge: { opening: 'trigger', effects: ['resource'] }, // 法术激荡
  seal_break: { opening: 'active', effects: ['status', 'resource'] }, // 封印破碎
  five_elements_bind: { opening: 'active', effects: ['status', 'resource'] }, // 五系束缚
  water_seal: { opening: 'active', effects: ['status', 'strike'], element: 'Water' }, // 水之封印
  fire_seal: { opening: 'active', effects: ['status', 'strike'], element: 'Fire' }, // 火之封印
  earth_seal: { opening: 'active', effects: ['status', 'strike'], element: 'Earth' }, // 地之封印
  wind_seal: { opening: 'active', effects: ['status', 'strike'], element: 'Wind' }, // 风之封印
  thunder_seal: { opening: 'active', effects: ['status', 'strike'], element: 'Thunder' }, // 雷之封印
  // 风之剑圣 (blade_master)
  wind_fury: { opening: 'trigger', effects: ['resource'], element: 'Wind' }, // 风怒追击
  holy_sword: { opening: 'trigger', effects: ['status', 'resource'] }, // 圣剑
  sword_shadow: { opening: 'trigger', effects: ['resource'] }, // 剑影
  gale_skill: { opening: 'trigger', effects: ['resource'], element: 'Wind' }, // 疾风技
  gale_slash: { opening: 'trigger', effects: ['status'], element: 'Wind' }, // 列风技
  // 神箭手 (archer)
  piercing_shot: { opening: 'trigger', effects: ['strike'] }, // 贯穿射击
  lightning_arrow: { opening: 'trigger', effects: ['strike'], element: 'Thunder' }, // 闪电箭
  snipe: { opening: 'active', effects: ['resource', 'form'] }, // 狙击
  precise_shot: { opening: 'trigger', effects: ['strike'] }, // 精准射击
  flash_trap: { opening: 'active', effects: ['strike'] }, // 闪光陷阱
  // 暗杀者 (assassin)
  backlash: { opening: 'trigger', effects: ['resource'] }, // 反噬
  water_shadow: { opening: 'trigger', effects: ['resource'], element: 'Water' }, // 水影
  stealth: { opening: 'active', effects: ['form', 'resource'] }, // 潜行
  // 圣女 (saintess)
  frost_prayer: { opening: 'trigger', effects: ['heal'], element: 'Water' }, // 冰霜祷言
  healing_light: { opening: 'active', effects: ['heal'], element: 'Light' }, // 治愈之光
  heal: { opening: 'active', effects: ['heal'], element: 'Light' }, // 治疗术
  saint_heal: { opening: 'active', effects: ['heal', 'resource'], element: 'Light', signature: true }, // 圣疗
  mercy: { opening: 'active', effects: ['resource', 'form'] }, // 怜悯
  // 魔法少女 (magical_girl)
  magic_bullet_control: { opening: 'active', effects: [] }, // 魔弹掌控
  magic_bullet_fusion: { opening: 'active', effects: ['resource', 'strike'] }, // 魔弹融合
  magic_bullet_fusion_chain: { opening: 'trigger', effects: ['resource', 'strike'] }, // 魔弹融合
  magic_blast: { opening: 'active', effects: ['resource', 'global', 'strike'] }, // 魔爆冲击
  destruction_storm: { opening: 'active', effects: ['strike'], signature: true }, // 毁灭风暴
  // 女武神 (valkyrie)
  valkyrie_divine_pursuit: { opening: 'trigger', effects: ['resource'] }, // 神圣追击
  valkyrie_order_seal: { opening: 'active', effects: ['resource', 'heal'] }, // 秩序之印
  valkyrie_peace_walker: { opening: 'trigger', effects: ['form'] }, // 和平行者
  valkyrie_military_glory: { opening: 'active', effects: ['heal', 'form', 'global'] }, // 军威神光
  valkyrie_heroic_summon: { opening: 'trigger', effects: ['strike', 'form', 'heal'] }, // 英灵召唤
  // 元素师 (elementalist)
  elementalist_absorb: { opening: 'trigger', effects: ['form'] }, // 元素吸收
  elementalist_ignite: { opening: 'active', effects: ['form', 'strike', 'resource'], element: 'Fire' }, // 元素点燃
  elementalist_thunder_strike: { opening: 'active', effects: ['strike', 'global'], element: 'Thunder' }, // 雷击
  elementalist_freeze: { opening: 'active', effects: ['strike', 'heal'], element: 'Water' }, // 冰冻
  elementalist_wind_blade: { opening: 'active', effects: ['strike', 'resource'], element: 'Wind' }, // 风刃
  elementalist_meteor: { opening: 'active', effects: ['strike', 'resource'], element: 'Earth' }, // 陨石
  elementalist_fireball: { opening: 'active', effects: ['strike'], element: 'Fire' }, // 火球
  elementalist_moonlight: { opening: 'active', effects: ['strike'], signature: true }, // 月光
  // 仲裁者 (arbiter)
  arbiter_law: { opening: 'trigger', effects: ['resource'] }, // 仲裁法则
  arbiter_judgment_tide: { opening: 'trigger', effects: ['form'] }, // 审判浪潮
  arbiter_ritual: { opening: 'active', effects: ['form'] }, // 仲裁仪式
  arbiter_ritual_break: { opening: 'active', effects: ['form', 'global'] }, // 仪式中断
  arbiter_doomsday: { opening: 'active', effects: ['form', 'strike'], signature: true }, // 末日审判
  arbiter_balance: { opening: 'active', effects: ['form', 'resource', 'global'] }, // 判决天平
  // 冒险家 (adventurer)
  adventurer_fraud: { opening: 'active', effects: ['strike'] }, // 欺诈
  adventurer_lucky_fortune: { opening: 'trigger', effects: ['resource'] }, // 强运
  adventurer_underground_law: { opening: 'trigger', effects: ['global'] }, // 地下法则
  adventurer_paradise: { opening: 'trigger', effects: ['resource'] }, // 冒险者天堂
  adventurer_steal_sky: { opening: 'active', effects: ['global', 'form'] }, // 偷天换日
  // 圣枪骑士 (holy_lancer)
  holy_lancer_revelation: { opening: 'trigger', effects: ['form'] }, // 神圣启示
  holy_lancer_radiance: { opening: 'active', effects: ['resource', 'heal'], element: 'Water' }, // 辉耀
  holy_lancer_punishment: { opening: 'active', effects: ['resource', 'heal'] }, // 惩戒
  holy_lancer_holy_strike: { opening: 'trigger', effects: ['heal'] }, // 圣击
  holy_lancer_sky_spear: { opening: 'trigger', effects: ['form', 'status'], element: 'Dark' }, // 天枪
  holy_lancer_earth_spear: { opening: 'trigger', effects: ['form', 'strike'], element: 'Earth' }, // 地枪
  holy_lancer_prayer: { opening: 'active', effects: ['heal', 'resource'], element: 'Light' }, // 圣光祈愈
  // 精灵射手 (elf_archer)
  elf_elemental_shot: { opening: 'trigger', effects: ['resource', 'status'] }, // 元素射击
  elf_animal_companion: { opening: 'trigger', effects: ['resource'] }, // 动物伙伴
  elf_ritual: { opening: 'active', effects: ['form', 'status'] }, // 精灵密仪
  elf_pet_empower: { opening: 'trigger', effects: ['resource'] }, // 宠物强化
  // 瘟疫法师 (plague_mage)
  plague_immortal: { opening: 'trigger', effects: ['heal'] }, // 不朽
  plague_blasphemy: { opening: 'trigger', effects: ['status'] }, // 圣渎
  plague_outbreak: { opening: 'active', effects: ['resource', 'strike', 'heal'], element: 'Earth' }, // 瘟疫
  plague_death_touch: { opening: 'active', effects: ['heal', 'resource', 'strike'] }, // 死亡之触
  plague_toxic_nova: { opening: 'active', effects: ['resource', 'strike', 'heal'], signature: true }, // 剧毒新星
  // 魔剑士 (magic_swordsman)
  ms_asura_combo: { opening: 'trigger', effects: ['resource'], element: 'Fire' }, // 修罗连斩
  ms_shadow_gather: { opening: 'active', effects: ['strike', 'form'], element: 'Dark' }, // 暗影凝聚
  ms_shadow_power: { opening: 'trigger', effects: ['strike'], element: 'Dark' }, // 暗影之力
  ms_shadow_reject: { opening: 'trigger', effects: ['status'] }, // 暗影抗拒
  ms_shadow_meteor: { opening: 'active', effects: ['resource', 'strike', 'global', 'form'], element: 'Dark' }, // 暗影流星
  ms_yellow_spring: { opening: 'trigger', effects: ['resource', 'status'] }, // 黄泉震颤
  // 血色剑灵 (crimson_sword_spirit)
  css_blood_thorns: { opening: 'trigger', effects: ['form'] }, // 血色荆棘
  css_crimson_flash: { opening: 'trigger', effects: ['form', 'strike', 'resource'] }, // 赤色一闪
  css_blood_rose: { opening: 'active', effects: ['form', 'status', 'resource', 'heal', 'global'] }, // 血染蔷薇
  css_blood_barrier: { opening: 'trigger', effects: ['form', 'status', 'strike'] }, // 血气屏障
  css_rose_courtyard: { opening: 'trigger', effects: ['status', 'global'] }, // 血蔷薇庭院
  css_dance: { opening: 'active', effects: ['status', 'form', 'resource', 'global'] }, // 散华轮舞
  // 祈祷师 (prayer_master)
  prayer_enter_form: { opening: 'active', effects: ['resource', 'form'] }, // 祈祷
  prayer_radiant_faith: { opening: 'active', effects: ['form', 'resource', 'global', 'heal'] }, // 光辉信仰
  prayer_dark_curse: { opening: 'active', effects: ['form', 'strike'], element: 'Dark' }, // 黑暗诅咒
  prayer_power_blessing: { opening: 'active', effects: ['status', 'strike'] }, // 威力赐福
  prayer_swift_blessing: { opening: 'active', effects: ['status', 'resource'] }, // 迅捷赐福
  prayer_mana_tide: { opening: 'trigger', effects: ['resource'] }, // 法力潮汐
  // 红莲骑士 (crimson_knight)
  crk_crimson_pact: { opening: 'trigger', effects: ['heal'] }, // 腥红圣约
  crk_crimson_faith: { opening: 'trigger', effects: ['status'] }, // 腥红信仰
  crk_bloody_prayer: { opening: 'active', effects: ['strike', 'heal', 'form'] }, // 血腥祷言
  crk_killing_feast: { opening: 'trigger', effects: ['form', 'strike'] }, // 杀戮盛宴
  crk_hot_blood: { opening: 'trigger', effects: ['form', 'heal'] }, // 热血沸腾
  crk_calm_mind: { opening: 'trigger', effects: ['form', 'resource'] }, // 戒骄戒躁
  crk_crimson_cross: { opening: 'active', effects: ['form', 'strike'], signature: true }, // 腥红十字
  // 英灵人形 (war_homunculus)
  hom_battle_pattern: { opening: 'trigger', effects: ['form'] }, // 战纹掌控
  hom_rage_suppress: { opening: 'trigger', effects: ['form'] }, // 怒火压制
  hom_rune_smash: { opening: 'trigger', effects: ['resource', 'form', 'strike'] }, // 战纹碎击
  hom_glyph_fusion: { opening: 'trigger', effects: ['resource', 'form', 'strike'] }, // 魔纹融合
  hom_rune_reforge: { opening: 'active', effects: ['form', 'resource'] }, // 符文改造
  hom_dual_echo: { opening: 'trigger', effects: ['strike', 'resource'] }, // 双重回响
  // 神官 (priest)
  priest_divine_revelation: { opening: 'trigger', effects: ['heal'] }, // 神圣启示
  priest_divine_bless: { opening: 'active', effects: ['resource', 'heal'] }, // 神圣祈福
  priest_water_power: { opening: 'active', effects: ['resource', 'heal'], element: 'Water' }, // 水之神力
  priest_guardian: { opening: 'trigger', effects: ['status'] }, // 圣使守护
  priest_divine_contract: { opening: 'active', effects: ['resource', 'heal'] }, // 神圣契约
  priest_divine_domain: { opening: 'active', effects: ['resource', 'heal', 'strike'] }, // 神圣领域
  // 阴阳师 (onmyoji)
  onmyoji_shikigami_descend: { opening: 'active', effects: ['resource', 'form'] }, // 式神降临
  onmyoji_yinyang_shift: { opening: 'trigger', effects: ['resource', 'form', 'strike'] }, // 阴阳转换
  onmyoji_shikigami_shift: { opening: 'trigger', effects: ['resource', 'form'] }, // 式神转换
  onmyoji_dark_ritual: { opening: 'trigger', effects: ['form', 'strike'], element: 'Dark' }, // 黑暗祭礼
  onmyoji_binding: { opening: 'trigger', effects: ['global', 'strike'] }, // 式神咒束
  onmyoji_life_barrier: { opening: 'active', effects: ['form', 'resource', 'heal', 'strike', 'status'], signature: true }, // 生命结界
  // 苍炎魔女 (blaze_witch)
  bw_rebirth_clock: { opening: 'trigger', effects: ['form'] }, // 永生银时计
  bw_blazing_codex: { opening: 'active', effects: ['strike'], element: 'Fire' }, // 苍炎法典
  bw_heavenfire_cleave: { opening: 'active', effects: ['strike', 'form'], element: 'Fire' }, // 天火断空
  bw_witch_wrath: { opening: 'active', effects: ['form', 'resource'], element: 'Fire' }, // 魔女之怒
  bw_substitute_doll: { opening: 'trigger', effects: ['resource'] }, // 替身玩偶
  bw_pain_link: { opening: 'active', effects: ['strike', 'resource'] }, // 痛苦链接
  bw_mana_inversion: { opening: 'trigger', effects: ['strike'] }, // 魔能反转
  // 贤者 (sage)
  sage_wisdom_codex: { opening: 'trigger', effects: ['resource'] }, // 智慧法典
  sage_magic_rebound: { opening: 'trigger', effects: ['resource', 'strike'] }, // 法术反弹
  sage_arcane_codex: { opening: 'active', effects: ['resource', 'strike'], signature: true }, // 魔道法典
  sage_holy_codex: { opening: 'active', effects: ['resource', 'heal', 'strike'] }, // 圣洁法典
  // 魔弓 (magic_bow)
  mb_magic_pierce: { opening: 'trigger', effects: ['form', 'strike'], element: 'Fire' }, // 魔贯冲击
  mb_thunder_scatter: { opening: 'active', effects: ['form', 'strike'], element: 'Thunder' }, // 雷光散射
  mb_multi_shot: { opening: 'trigger', effects: ['form', 'strike'], element: 'Wind' }, // 多重射击
  mb_charge: { opening: 'active', effects: ['resource', 'form'] }, // 充能
  mb_demon_eye: { opening: 'active', effects: ['resource', 'form'] }, // 魔眼
  // 魔枪 (magic_lancer)
  ml_dark_release: { opening: 'active', effects: ['form'], element: 'Dark' }, // 暗之解放
  ml_phantom_stardust: { opening: 'active', effects: ['strike', 'form'], element: 'Dark' }, // 幻影星尘
  ml_dark_bind: { opening: 'trigger', effects: ['status'] }, // 黑暗束缚
  ml_dark_barrier: { opening: 'trigger', effects: ['resource'], element: 'Thunder' }, // 暗之障壁
  ml_fullness: { opening: 'active', effects: ['resource', 'global', 'form'] }, // 充盈
  ml_black_spear: { opening: 'trigger', effects: ['resource', 'strike'], element: 'Dark', signature: true }, // 漆黑之枪
  // 灵符师 (spirit_caster)
  sc_talisman_thunder: { opening: 'active', effects: ['strike'], element: 'Thunder' }, // 灵符-雷鸣
  sc_talisman_wind: { opening: 'active', effects: ['resource'], element: 'Wind' }, // 灵符-风行
  sc_incantation: { opening: 'trigger', effects: ['form'] }, // 念咒
  sc_hundred_night: { opening: 'trigger', effects: ['form', 'strike'] }, // 百鬼夜行
  sc_spiritual_collapse: { opening: 'trigger', effects: ['resource', 'strike'] }, // 灵力崩解
  // 吟游诗人 (bard)
  bd_descent_concerto: { opening: 'trigger', effects: ['resource', 'form', 'strike'] }, // 沉沦协奏曲
  bd_dissonance_chord: { opening: 'active', effects: ['form', 'resource'] }, // 不谐和弦
  bd_forbidden_verse: { opening: 'trigger', effects: ['form', 'status', 'strike'] }, // 禁忌诗篇
  bd_rousing_rhapsody: { opening: 'trigger', effects: ['strike', 'resource'] }, // 激昂狂想曲
  bd_victory_symphony: { opening: 'trigger', effects: ['resource', 'global', 'heal'] }, // 胜利交响诗
  bd_hope_fugue: { opening: 'active', effects: ['resource', 'status', 'heal'] }, // 希望赋格曲
  // 勇者 (hero)
  hero_heart: { opening: 'trigger', effects: ['resource'] }, // 勇者之心
  hero_roar: { opening: 'trigger', effects: ['form', 'resource', 'strike'] }, // 怒吼
  hero_forbidden_power: { opening: 'trigger', effects: ['resource', 'form', 'strike'], element: 'Fire' }, // 禁断之力
  hero_exhaustion: { opening: 'trigger', effects: ['form', 'resource', 'strike'] }, // 精疲力竭
  hero_calm_mind: { opening: 'trigger', effects: ['form', 'resource'] }, // 明镜止水
  hero_taunt: { opening: 'active', effects: ['form', 'status'] }, // 挑衅
  hero_dead_duel: { opening: 'trigger', effects: ['resource', 'form', 'global'] }, // 死斗
  // 格斗家 (fighter)
  fighter_psi_field: { opening: 'trigger', effects: ['status'] }, // 念气力场
  fighter_charge_strike: { opening: 'trigger', effects: ['form', 'strike'] }, // 蓄力一击
  fighter_psi_bullet: { opening: 'trigger', effects: ['form', 'strike'] }, // 念弹
  fighter_hundred_dragon: { opening: 'active', effects: ['form', 'status'] }, // 百式幻龙拳
  fighter_burst_crash: { opening: 'trigger', effects: ['form', 'strike'] }, // 气绝崩击
  fighter_war_god_drive: { opening: 'active', effects: ['resource', 'heal'] }, // 斗神天驱
  // 圣弓 (holy_bow)
  hb_heavenly_bow: { opening: 'trigger', effects: ['resource', 'form', 'strike'] }, // 天之弓
  hb_holy_shard_storm: { opening: 'active', effects: ['resource', 'strike'] }, // 圣屑飓暴
  hb_radiant_descent: { opening: 'active', effects: ['resource', 'form'], element: 'Light' }, // 圣煌降临
  hb_light_burst: { opening: 'active', effects: ['resource', 'heal', 'form', 'strike'], element: 'Light' }, // 圣光爆裂
  hb_meteor_bullet: { opening: 'trigger', effects: ['resource', 'heal'] }, // 流星圣弹
  hb_radiant_cannon: { opening: 'active', effects: ['form', 'resource', 'global'], element: 'Light', signature: true }, // 圣煌辉光炮
  hb_auto_fill: { opening: 'trigger', effects: ['resource', 'form', 'heal'] }, // 自动填充
  // 剑帝 (sword_emperor)
  se_sword_soul_guard: { opening: 'trigger', effects: ['form'] }, // 剑魂守护
  se_feint: { opening: 'trigger', effects: ['form'] }, // 佯攻
  se_sword_qi_slash: { opening: 'trigger', effects: ['form', 'strike'] }, // 剑气斩
  se_angel_soul: { opening: 'trigger', effects: ['form', 'heal', 'global'], element: 'Light' }, // 天使之魂
  se_demon_soul: { opening: 'trigger', effects: ['form', 'strike'], element: 'Dark' }, // 恶魔之魂
  se_indomitable_will: { opening: 'trigger', effects: ['resource', 'form'] }, // 不屈意志
  // 兽灵武士 (beast_samurai)
  bs_warrior_zanshin: { opening: 'trigger', effects: ['form'] }, // 武者残心
  bs_one_strike_no_thought: { opening: 'trigger', effects: ['form'] }, // 一击无念
  bs_beast_soul_will: { opening: 'trigger', effects: ['form'] }, // 兽魂意念
  bs_beast_soul_alert: { opening: 'trigger', effects: ['form', 'resource'] }, // 兽魂警戒
  bs_beast_return: { opening: 'trigger', effects: ['form', 'resource'] }, // 兽返
  bs_reversal_iaijutsu: { opening: 'trigger', effects: ['form', 'resource', 'global'] }, // 逆反居合斩
  bs_iaijutsu_style: { opening: 'active', effects: ['form', 'resource'] }, // 御魂流居合式
  // 灵魂术士 (soul_sorcerer)
  ss_soul_devour: { opening: 'trigger', effects: ['form'] }, // 灵魂吞噬
  ss_soul_recall: { opening: 'active', effects: ['resource', 'form'] }, // 灵魂召还
  ss_soul_convert: { opening: 'trigger', effects: ['form'] }, // 灵魂转换
  ss_soul_mirror: { opening: 'active', effects: ['resource', 'form'] }, // 灵魂镜像
  ss_soul_blast: { opening: 'active', effects: ['form', 'strike'] }, // 灵魂震爆
  ss_soul_grant: { opening: 'active', effects: ['form', 'resource'] }, // 灵魂赐予
  ss_soul_link: { opening: 'active', effects: ['form', 'status'] }, // 灵魂链接
  ss_soul_amp: { opening: 'active', effects: ['form'] }, // 灵魂增幅
  // 月之女神 (moon_goddess)
  mg_new_moon_shelter: { opening: 'trigger', effects: ['form', 'status', 'global'] }, // 新月庇护
  mg_dark_moon_curse: { opening: 'trigger', effects: ['global', 'form'], element: 'Dark' }, // 闇月诅咒
  mg_medusa_eye: { opening: 'trigger', effects: ['status', 'heal', 'form', 'global', 'strike'] }, // 美杜莎之眼
  mg_moon_cycle: { opening: 'trigger', effects: ['status', 'heal', 'global', 'form'] }, // 月之轮回
  mg_blasphemy: { opening: 'trigger', effects: ['heal', 'strike'] }, // 月渎
  mg_darkmoon_slash: { opening: 'trigger', effects: ['form', 'global', 'strike'], element: 'Dark' }, // 闇月斩
  mg_pale_moon: { opening: 'active', effects: ['form', 'resource', 'status', 'strike'], signature: true }, // 苍白之月
  // 血之巫女 (blood_priestess)
  bp_blood_sorrow: { opening: 'active', effects: ['strike', 'status'] }, // 血之哀伤
  bp_bleeding: { opening: 'trigger', effects: ['form', 'heal', 'strike'] }, // 流血
  bp_backflow: { opening: 'active', effects: ['resource', 'heal'] }, // 逆流
  bp_blood_wail: { opening: 'active', effects: ['strike'] }, // 血之悲鸣
  bp_shared_life: { opening: 'active', effects: ['resource', 'status'] }, // 同生共死
  bp_blood_curse: { opening: 'active', effects: ['strike', 'resource'] }, // 血之诅咒
  // 蝶舞者 (butterfly_dancer)
  bt_life_fire: { opening: 'trigger', effects: ['status'] }, // 生命之火
  bt_dance: { opening: 'active', effects: ['resource', 'status'] }, // 舞动
  bt_poison_powder: { opening: 'trigger', effects: ['status', 'strike'] }, // 毒粉
  bt_pilgrimage: { opening: 'trigger', effects: ['status'] }, // 朝圣
  bt_mirror: { opening: 'trigger', effects: ['status', 'strike'] }, // 镜花水月
  bt_wither: { opening: 'trigger', effects: ['strike', 'global'] }, // 凋零
  bt_chrysalis: { opening: 'active', effects: ['form', 'status'] }, // 蛹化
  bt_reverse_butterfly: { opening: 'active', effects: ['resource', 'strike', 'form'], signature: true }, // 倒逆之蝶
}
