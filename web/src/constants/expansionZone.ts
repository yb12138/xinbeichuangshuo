/**
 * 扩展区（专属牌 + 盖牌）的分组配色与图形。
 * 盖牌牌背、分组标签、收起竖条共用这里的色调，保证同一种盖牌在三处看起来一致。
 */
import { COVER_MARKS } from './battleMarks'
import { FALLBACK_GLYPH, MARK_GLYPHS } from '../components/marks/markGlyphs'

/** 专属牌分组使用的键（不是 FieldCard.effect）。 */
export const EXCLUSIVE_GROUP_KEY = 'Exclusive'

/** 牌背色调：比座位标记更饱和，便于在深色抽屉里区分种类。 */
const EXPANSION_TINTS: Record<string, string> = {
  [EXCLUSIVE_GROUP_KEY]: '#e8c478',
  ButterflyCocoon: '#6f9bff',
  MoonDarkMoon: '#a77bff',
  SwordSoul: '#b4bccb',
  MagicBowCharge: '#4fa8ff',
  ElfBlessing: '#53d08f',
  SpiritCasterPower: '#e3c46a',
}

const STAR_GLYPH = 'M12 2.8l2.6 6 6.5.6-4.9 4.3 1.5 6.4L12 16.8l-5.7 3.3 1.5-6.4-4.9-4.3 6.5-.6L12 2.8z'

export type ExpansionGroupView = {
  label: string
  tint: string
  glyph: string
}

export function expansionGroupView(key: string): ExpansionGroupView {
  if (key === EXCLUSIVE_GROUP_KEY) {
    return { label: '专属牌', tint: EXPANSION_TINTS[EXCLUSIVE_GROUP_KEY]!, glyph: STAR_GLYPH }
  }
  const def = COVER_MARKS[key]
  return {
    label: def?.label ?? '盖牌',
    tint: EXPANSION_TINTS[key] ?? def?.accent ?? '#9fb3cf',
    glyph: (def && MARK_GLYPHS[def.glyph]) || FALLBACK_GLYPH,
  }
}
