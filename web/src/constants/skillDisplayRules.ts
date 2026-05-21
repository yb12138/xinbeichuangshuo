import type { Card } from '../types/game'

type SkillId = string

const SKILL_DISCARD_ELEMENT_RULES: Record<string, Card['element'][]> = {
  magic_bullet_fusion: ['Fire', 'Earth'],
  magic_bullet_fusion_chain: ['Fire', 'Earth'],
}

const SKILL_DISCARD_GUIDES: Record<string, string> = {
  magic_bullet_fusion: '需要火系或地系牌',
  magic_bullet_fusion_chain: '需要火系或地系牌',
  priest_water_power: '第一张需水系；若仍有手牌，第二张将交给目标队友',
  onmyoji_shikigami_descend: '需要2张命格相同的手牌',
}

const SKILL_DISABLED_REASONS: Record<string, string> = {
  magic_blast: '手牌中没有法术牌，无法发动【魔爆冲击】。',
  magic_bullet_fusion: '需要弃置1张火系或地系牌，才能发动【魔弹融合】。',
  magic_bullet_fusion_chain: '需要弃置1张火系或地系牌，才能发动【魔弹融合】。',
  bw_blazing_codex: '手牌中没有火系牌，无法发动【苍炎法典】。',
  onmyoji_shikigami_descend: '需要弃置2张命格相同的手牌才能发动。',
}

const SKILL_COST_TEXT_OVERRIDES: Record<string, string> = {
    priest_water_power: '弃1水牌+交1手牌(若有)',
    ss_soul_recall: '弃法术牌',
}

export function skillDiscardGuideText(skillId: SkillId): string {
  return SKILL_DISCARD_GUIDES[skillId] || ''
}

export function skillSpecificDisabledReason(skillId: SkillId): string {
  return SKILL_DISABLED_REASONS[skillId] || ''
}

export function skillCostTextOverride(skillId: SkillId): string {
  return SKILL_COST_TEXT_OVERRIDES[skillId] || ''
}

export function skillCanUseDiscardCard(skillId: SkillId, card: Pick<Card, 'element'> | { element: string }): boolean {
  const allowed = SKILL_DISCARD_ELEMENT_RULES[skillId]
  if (!allowed || allowed.length === 0) return true
  return (allowed as readonly string[]).includes(card.element)
}
