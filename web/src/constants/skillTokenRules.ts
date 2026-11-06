/**
 * 前端兜底的技能指示物门槛：只在服务端没下发该技能时生效（下发了就以服务端 disabled 为准）。
 * 只列"发动要消耗/要求持有"的指示物；像判决天平这种只增加指示物的技能不应出现在这里。
 */
export type SkillTokenRule = {
  token: string
  min: number
  label: string
}

export const SKILL_TOKEN_RULES: Record<string, SkillTokenRule[]> = {
  crk_killing_feast: [{ token: 'crk_blood_mark', min: 1, label: '血印' }],
  crk_crimson_cross: [{ token: 'crk_blood_mark', min: 1, label: '血印' }],
  css_blood_thorns: [{ token: 'css_blood', min: 1, label: '鲜血' }],
  css_crimson_flash: [{ token: 'css_blood', min: 1, label: '鲜血' }],
  css_blood_rose: [{ token: 'css_blood', min: 2, label: '鲜血' }],
  css_blood_barrier: [{ token: 'css_blood', min: 1, label: '鲜血' }],
  hom_rage_suppress: [{ token: 'hom_war_rune', min: 1, label: '战纹' }],
  hom_rune_smash: [{ token: 'hom_war_rune', min: 1, label: '战纹' }],
  hom_glyph_fusion: [{ token: 'hom_magic_rune', min: 1, label: '魔纹' }],
  hero_roar: [{ token: 'hero_anger', min: 1, label: '怒气' }],
  hero_calm_mind: [{ token: 'hero_wisdom', min: 4, label: '知性' }],
  hero_taunt: [{ token: 'hero_anger', min: 1, label: '怒气' }],
  fighter_hundred_dragon: [{ token: 'fighter_qi', min: 3, label: '斗气' }],
  fighter_burst_crash: [{ token: 'fighter_qi', min: 1, label: '斗气' }],
  ss_soul_mirror: [{ token: 'ss_yellow_soul', min: 2, label: '黄色灵魂' }],
  ss_soul_blast: [{ token: 'ss_yellow_soul', min: 3, label: '黄色灵魂' }],
  ss_soul_grant: [{ token: 'ss_blue_soul', min: 3, label: '蓝色灵魂' }],
  ss_soul_link: [
      { token: 'ss_yellow_soul', min: 1, label: '黄色灵魂' },
      { token: 'ss_blue_soul', min: 1, label: '蓝色灵魂' },
  ],
  bw_heavenfire_cleave: [{ token: 'bw_rebirth', min: 1, label: '重生' }],
}

/** 指示物不足时返回提示文案，否则返回空串。 */
export function skillTokenShortfallReason(skillId: string, tokens?: Record<string, number> | null): string {
  for (const rule of SKILL_TOKEN_RULES[skillId] || []) {
    const cur = tokens?.[rule.token] ?? 0
    if (cur < rule.min) {
      return `${rule.label}不足（需要 ${rule.min}，当前 ${cur}）。`
    }
  }
  return ''
}
