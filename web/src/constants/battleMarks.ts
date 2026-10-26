/**
 * 对战标记（场上状态 / 盖牌 / 指示物）的统一定义：名称、类别与配色。
 * 视觉依据：outimage/《对战标记图标集》UI概念图.png。
 * 图形本体由 components/marks/BattleMarkIcon.vue 按 glyph 绘制（24px 下仍可辨认的剪影）。
 */

/** 场上状态的性质决定徽章边框颜色：增益金银、减益红紫、封印取元素色、控制/链接青色。 */
export type StatusTone = 'buff' | 'debuff' | 'seal' | 'control'

export interface StatusMarkDef {
  label: string
  tone: StatusTone
  glyph: string
  /** 封印等需要单独点缀色时使用；缺省按 tone 取色。 */
  accent?: string
}

export interface CoverMarkDef {
  label: string
  glyph: string
  accent: string
}

export interface TokenMarkDef {
  label: string
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
  Shield: { label: '圣盾', tone: 'buff', glyph: 'shield' },
  PowerBlessing: { label: '威力赐福', tone: 'buff', glyph: 'sword-up' },
  SwiftBlessing: { label: '迅捷赐福', tone: 'buff', glyph: 'wing-arrow' },
  Mercy: { label: '怜悯', tone: 'buff', glyph: 'palm' },
  BloodSharedLife: { label: '同生共死', tone: 'buff', glyph: 'twin-heart' },
  Poison: { label: '中毒', tone: 'debuff', glyph: 'drop' },
  Weak: { label: '虚弱', tone: 'debuff', glyph: 'broken-sword' },
  SealFire: { label: '火之封印', tone: 'seal', glyph: 'seal-fire', accent: '#ff6a4a' },
  SealWater: { label: '水之封印', tone: 'seal', glyph: 'seal-water', accent: '#4fb0ff' },
  SealEarth: { label: '地之封印', tone: 'seal', glyph: 'seal-earth', accent: '#d8a85a' },
  SealWind: { label: '风之封印', tone: 'seal', glyph: 'seal-wind', accent: '#4fd39a' },
  SealThunder: { label: '雷之封印', tone: 'seal', glyph: 'seal-thunder', accent: '#a98bff' },
  FiveElementsBind: { label: '五系束缚', tone: 'seal', glyph: 'chain-ring', accent: '#e3c46a' },
  SoulLink: { label: '灵魂链接', tone: 'control', glyph: 'soul-link' },
  HeroTaunt: { label: '挑衅', tone: 'control', glyph: 'taunt' },
  FighterHundredDragonLock: { label: '百式幻龙拳锁定', tone: 'control', glyph: 'claw-target' },
  BardEternalMovement: { label: '永恒乐章', tone: 'control', glyph: 'note' },
  RoseCourtyard: { label: '血蔷薇庭院', tone: 'control', glyph: 'rose' },
  Stealth: { label: '潜行', tone: 'control', glyph: 'veil' },
}

/** key = FieldCard.effect（Mode=Cover）。盖牌只露牌背纹样与数量，不暴露牌面。 */
export const COVER_MARKS: Record<string, CoverMarkDef> = {
  MagicBowCharge: { label: '充能', glyph: 'crystal', accent: '#6fa8ff' },
  SpiritCasterPower: { label: '妖力', glyph: 'talisman', accent: '#e3c46a' },
  SwordSoul: { label: '剑魂', glyph: 'sword', accent: '#b9a4ff' },
  MoonDarkMoon: { label: '暗月', glyph: 'crescent', accent: '#c7b8ff' },
  ButterflyCocoon: { label: '茧', glyph: 'cocoon', accent: '#d9dee8' },
  ElfBlessing: { label: '祝福区', glyph: 'laurel', accent: '#59d39a' },
}

/**
 * key = player.tokens / player.indicators 的键。
 * 与盖牌数量重复的计数键（茧、暗月、剑魂、充能）不在这里列出，由盖牌段统一展示。
 */
export const TOKEN_MARKS: Record<string, TokenMarkDef> = {
  se_sword_qi: { label: '剑气', glyph: 'sword-qi', accent: '#6fc3ff' },
  onmyoji_ghost_fire: { label: '鬼火', glyph: 'flame', accent: '#4fd6c4', cap: 3 },
  mg_new_moon: { label: '新月', glyph: 'crescent', accent: '#d9dee8' },
  mg_petrify: { label: '石化', glyph: 'stone', accent: '#a3a9b5' },
  bd_inspiration: { label: '灵感', glyph: 'note', accent: '#b48bff' },
  hero_anger: { label: '怒气', glyph: 'anger', accent: '#ff5d5d' },
  hero_wisdom: { label: '知性', glyph: 'book', accent: '#6fa8ff' },
  fighter_qi: { label: '斗气', glyph: 'fist', accent: '#ff9a4d' },
  css_blood: { label: '鲜血', glyph: 'drop', accent: '#e2525e' },
  valkyrie_spirit: { label: '英灵', glyph: 'wing-helm', accent: '#e3c46a' },
  ss_yellow_soul: { label: '黄魂', glyph: 'soul-flame', accent: '#f0c35a' },
  ss_blue_soul: { label: '蓝魂', glyph: 'soul-flame', accent: '#4f9fff' },
  hom_war_rune: { label: '战纹', glyph: 'war-rune', accent: '#8fa6ff' },
  hom_magic_rune: { label: '魔纹', glyph: 'sigil', accent: '#b48bff' },
  prayer_rune: { label: '祈祷符文', glyph: 'pray', accent: '#e3c46a' },
  crk_blood_mark: { label: '血印', glyph: 'blood-mark', accent: '#e2525e' },
  judgment: { label: '审判', glyph: 'scale', accent: '#9fd4ff' },
  element: { label: '元素', glyph: 'prism', accent: '#4fd6c4' },
  bw_rebirth: { label: '重生', glyph: 'phoenix', accent: '#f0a35a' },
  hb_faith: { label: '信仰', glyph: 'pray', accent: '#f0d27a' },
  hb_cannon: { label: '圣煌辉光炮', glyph: 'sun', accent: '#f0d27a' },
  elf_blessing_count: { label: '祝福', glyph: 'laurel', accent: '#59d39a' },
  bs_beast_soul: { label: '兽魂', glyph: 'fang', accent: '#e39a4f' },
  bs_zanshin: { label: '残心', glyph: 'eye', accent: '#9fd4ff' },
  bt_pupa: { label: '蛹', glyph: 'cocoon', accent: '#d9dee8' },
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
