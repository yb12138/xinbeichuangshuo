/**
 * 对战标记（场上状态 / 盖牌 / 指示物）的统一定义：名称、类别与配色。
 * 视觉依据：outimage/《对战标记图标集》UI概念图.png。
 * 图形本体由 components/marks/BattleMarkIcon.vue 按 glyph 绘制（24px 下仍可辨认的剪影）。
 */

/** 场上状态的性质决定徽章边框颜色：增益金银、减益红紫、封印取元素色、控制/链接青色。 */
export type StatusTone = 'buff' | 'debuff' | 'seal' | 'control'

export interface StatusMarkDef {
  label: string
  /** 一句话规则说明，展开面板与悬停提示用。 */
  hint: string
  tone: StatusTone
  glyph: string
  /** 封印等需要单独点缀色时使用；缺省按 tone 取色。 */
  accent?: string
}

export interface CoverMarkDef {
  label: string
  /** 一句话规则说明，展开面板与悬停提示用。 */
  hint: string
  glyph: string
  accent: string
}

export interface TokenMarkDef {
  label: string
  /** 一句话规则说明，展开面板与悬停提示用。 */
  hint: string
  glyph: string
  accent: string
  /** 上限较小的指示物用圆点格展示（剑魂、鬼火等）。 */
  cap?: number
}

export const TONE_RING: Record<StatusTone, string> = {
  buff: '#e3c46a',
  debuff: '#d0507a',
  seal: '#b9c1cd',
  control: '#4fc9d6',
}

/** key = FieldCard.effect（Mode=Effect）。 */
export const STATUS_MARKS: Record<string, StatusMarkDef> = {
  Shield: { label: '圣盾', hint: '遭受攻击或魔弹时移除，视为未命中。', tone: 'buff', glyph: 'shield' },
  PowerBlessing: { label: '威力赐福', hint: '攻击命中后可移除，本次伤害+2。', tone: 'buff', glyph: 'sword-up' },
  SwiftBlessing: { label: '迅捷赐福', hint: '攻击或法术行动结束后可移除，额外+1攻击行动。', tone: 'buff', glyph: 'wing-arrow' },
  Mercy: { label: '怜悯', hint: '手牌上限恒定为7，并获得1水晶。', tone: 'buff', glyph: 'palm' },
  BloodSharedLife: { label: '同生共死', hint: '与血之巫女同生共死，伤害与治疗相互牵连。', tone: 'buff', glyph: 'twin-heart' },
  Poison: { label: '中毒', hint: '行动阶段开始前承受1点法术伤害，然后移除。', tone: 'debuff', glyph: 'drop' },
  Weak: { label: '虚弱', hint: '行动阶段开始前选择：跳过行动阶段，或摸3张牌继续。', tone: 'debuff', glyph: 'broken-sword' },
  SealFire: { label: '火之封印', hint: '打出或展示火系牌时强制触发，承受3点法术伤害后移除。', tone: 'seal', glyph: 'seal-fire', accent: '#ff6a4a' },
  SealWater: { label: '水之封印', hint: '打出或展示水系牌时强制触发，承受3点法术伤害后移除。', tone: 'seal', glyph: 'seal-water', accent: '#4fb0ff' },
  SealEarth: { label: '地之封印', hint: '打出或展示地系牌时强制触发，承受3点法术伤害后移除。', tone: 'seal', glyph: 'seal-earth', accent: '#d8a85a' },
  SealWind: { label: '风之封印', hint: '打出或展示风系牌时强制触发，承受3点法术伤害后移除。', tone: 'seal', glyph: 'seal-wind', accent: '#4fd39a' },
  SealThunder: { label: '雷之封印', hint: '打出或展示雷系牌时强制触发，承受3点法术伤害后移除。', tone: 'seal', glyph: 'seal-thunder', accent: '#a98bff' },
  FiveElementsBind: { label: '五系束缚', hint: '跳过下个行动阶段；可改为摸（2+封印数）张牌取消。', tone: 'seal', glyph: 'chain-ring', accent: '#e3c46a' },
  SoulLink: { label: '灵魂链接', hint: '灵魂术士与队友之间，承受伤害前可移除蓝魂把伤害转给另一方。', tone: 'control', glyph: 'soul-link' },
  HeroTaunt: { label: '挑衅', hint: '下个行动阶段必须且只能主动攻击勇者，否则跳过。', tone: 'control', glyph: 'taunt' },
  FighterHundredDragonLock: { label: '百式幻龙拳锁定', hint: '格斗家百式幻龙拳锁定的目标，主动攻击必须选择此人。', tone: 'control', glyph: 'claw-target' },
  BardEternalMovement: { label: '永恒乐章', hint: '吟游诗人的乐章结算后，按灵感是否已满获得灵感或承受伤害。', tone: 'control', glyph: 'note' },
  RoseCourtyard: { label: '血蔷薇庭院', hint: '在场时所有人的治疗都不能用于抵挡伤害。', tone: 'control', glyph: 'rose' },
  Stealth: { label: '潜行', hint: '不能成为主动攻击的目标，持续到下个行动阶段开始。', tone: 'control', glyph: 'veil' },
}

/** key = FieldCard.effect（Mode=Cover）。盖牌只露牌背纹样与数量，不暴露牌面。 */
export const COVER_MARKS: Record<string, CoverMarkDef> = {
  MagicBowCharge: { label: '充能', hint: '魔弓的充能；主动攻击前可移除来增加伤害。', glyph: 'crystal', accent: '#6fa8ff' },
  SpiritCasterPower: { label: '妖力', hint: '灵符师发动灵符时盖放的妖力。', glyph: 'talisman', accent: '#e3c46a' },
  SwordSoul: { label: '剑魂', hint: '剑帝主动攻击未命中时留下的攻击牌；按能量单双视为天使之魂或恶魔之魂。', glyph: 'sword', accent: '#b9a4ff' },
  MoonDarkMoon: { label: '暗月', hint: '月之女神的闇月；可在对手攻击或回合结束时移除发动效果。', glyph: 'crescent', accent: '#c7b8ff' },
  ButterflyCocoon: { label: '茧', hint: '蝶舞者的茧；移除时触发朝圣或毒粉。', glyph: 'cocoon', accent: '#d9dee8' },
  ElfBlessing: { label: '祝福区', hint: '精灵射手的祝福区；可像手牌一样打出。', glyph: 'laurel', accent: '#59d39a' },
}

/**
 * key = player.tokens / player.indicators 的键。
 * 与盖牌数量重复的计数键（茧、暗月、剑魂、充能）不在这里列出，由盖牌段统一展示。
 */
export const TOKEN_MARKS: Record<string, TokenMarkDef> = {
  se_sword_qi: { label: '剑气', hint: '剑帝主动攻击未命中时+1。', glyph: 'sword-qi', accent: '#6fc3ff' },
  onmyoji_ghost_fire: { label: '鬼火', hint: '阴阳转换时+1，攻击伤害等于鬼火数，上限3。', glyph: 'flame', accent: '#4fd6c4', cap: 3 },
  mg_new_moon: { label: '新月', hint: '月之女神回合结束时移除治疗获得。', glyph: 'crescent', accent: '#d9dee8' },
  mg_petrify: { label: '石化', hint: '月之女神移除同系闇月时获得。', glyph: 'stone', accent: '#a3a9b5' },
  bd_inspiration: { label: '灵感', hint: '吟游诗人对多名对手造成法术伤害后获得。', glyph: 'note', accent: '#b48bff' },
  hero_anger: { label: '怒气', hint: '勇者主动攻击前移除1点，本次伤害+2。', glyph: 'anger', accent: '#ff5d5d' },
  hero_wisdom: { label: '知性', hint: '勇者攻击未命中或发动挑衅时获得。', glyph: 'book', accent: '#6fa8ff' },
  fighter_qi: { label: '斗气', hint: '格斗家蓄力一击累积；未命中时对自己造成等量伤害。', glyph: 'fist', accent: '#ff9a4d' },
  css_blood: { label: '鲜血', hint: '血色剑灵攻击命中时+1，上限3。', glyph: 'drop', accent: '#e2525e' },
  valkyrie_spirit: { label: '英灵', hint: '女武神英灵形态相关的指示物。', glyph: 'wing-helm', accent: '#e3c46a' },
  ss_yellow_soul: { label: '黄魂', hint: '我方士气每下降1点，灵魂术士+1黄魂（上限6）。', glyph: 'soul-flame', accent: '#f0c35a' },
  ss_blue_soul: { label: '蓝魂', hint: '灵魂术士弃法术牌获得，用于灵魂链接转移伤害（上限6）。', glyph: 'soul-flame', accent: '#4f9fff' },
  hom_war_rune: { label: '战纹', hint: '英灵人形的战纹。', glyph: 'war-rune', accent: '#8fa6ff' },
  hom_magic_rune: { label: '魔纹', hint: '英灵人形的魔纹。', glyph: 'sigil', accent: '#b48bff' },
  prayer_rune: { label: '祈祷符文', hint: '祈祷师在祈祷形态下主动攻击累积。', glyph: 'pray', accent: '#e3c46a' },
  crk_blood_mark: { label: '血印', hint: '红莲骑士分配治疗时获得。', glyph: 'blood-mark', accent: '#e2525e' },
  judgment: { label: '审判', hint: '仲裁者每次承受伤害+1，上限4。', glyph: 'scale', accent: '#9fd4ff' },
  element: { label: '元素', hint: '元素师的元素指示物。', glyph: 'prism', accent: '#4fd6c4' },
  bw_rebirth: { label: '重生', hint: '苍炎魔女因法术伤害士气下降时+1，上限4。', glyph: 'phoenix', accent: '#f0a35a' },
  hb_faith: { label: '信仰', hint: '圣弓以圣命格攻击命中时+1。', glyph: 'pray', accent: '#f0d27a' },
  hb_cannon: { label: '圣煌辉光炮', hint: '圣弓的圣煌辉光炮次数。', glyph: 'sun', accent: '#f0d27a' },
  elf_blessing_count: { label: '祝福', hint: '精灵射手祝福区的牌数。', glyph: 'laurel', accent: '#59d39a' },
  bs_beast_soul: { label: '兽魂', hint: '兽灵武士普通形态主动攻击命中时+1。', glyph: 'fang', accent: '#e39a4f' },
  bs_zanshin: { label: '残心', hint: '兽灵武士移除兽魂或攻击行动结束时获得。', glyph: 'eye', accent: '#9fd4ff' },
  bt_pupa: { label: '蛹', hint: '蝶舞者的蛹；手牌上限-蛹数（最低3）。', glyph: 'cocoon', accent: '#d9dee8' },
}

/** 仅作展示提示的计数键：已由盖牌段表示，避免在指示物段重复出现。 */
export const COVER_COUNT_TOKEN_KEYS = new Set([
  'bt_cocoon_count',
  'mg_dark_moon_count',
  'se_sword_soul_count',
  'mb_charge_count',
])

export function statusRingColor(def: StatusMarkDef): string {
  return def.accent ?? TONE_RING[def.tone]
}
