import type { AvailableSkill } from '../types/game'

/**
 * 技能发动流程（选独有技牌 / 弃牌 / 选目标）的共享判定。
 * ActionPanel 负责推进步骤，GameBoard 负责手牌与头像的点选，两边必须用同一套规则，
 * 否则会出现"面板进入了选牌步骤、手牌却点不动"这类分歧。
 */

/** 元素师独有技：带 1 张弃牌成本，但仍然要先在手牌里选中对应的独有技牌。 */
export const ELEMENTALIST_EXCLUSIVE_PROMPT_FLOW_SKILL_IDS: ReadonlySet<string> = new Set([
  'elementalist_thunder_strike',
  'elementalist_freeze',
  'elementalist_wind_blade',
  'elementalist_meteor',
  'elementalist_fireball',
])

/** 是否要先在手牌里选中独有技牌，再进入后续步骤。 */
export function requiresExclusiveCardSelectionBeforeTarget(skill?: AvailableSkill | null): boolean {
  return !!skill &&
    !!skill.require_exclusive &&
    (skill.target_type ?? 0) !== 0 &&
    ((skill.cost_discards ?? 0) === 0 || ELEMENTALIST_EXCLUSIVE_PROMPT_FLOW_SKILL_IDS.has(skill.id))
}

/** 本次发动实际需要弃几张牌（部分技能在手牌不足时降为手牌数）。 */
export function requiredSkillDiscardCount(skill: AvailableSkill | null | undefined, handCount: number): number {
  if (!skill || (skill.cost_discards ?? 0) <= 0) return 0
  // 神官-神圣领域：手牌不足 2 时，改为弃全部手牌。
  // 神官-水之神力：若弃完水系后无剩余手牌，则仅需弃 1 张水系牌。
  if (skill.id === 'priest_divine_domain' || skill.id === 'priest_water_power') {
    return Math.min(skill.cost_discards, Math.max(0, handCount))
  }
  return skill.cost_discards
}

/** 技能的目标数量范围；未声明时按目标类型兜底（需要目标的至少 1 个，至多 1 个）。 */
export function skillTargetBounds(skill: AvailableSkill): { min: number; max: number } {
  const fallbackMin = (skill.target_type ?? 0) >= 2 ? 1 : 0
  const min = (skill.min_targets || 0) > 0 ? (skill.min_targets || 0) : fallbackMin
  const max = (skill.max_targets || 0) > 0 ? (skill.max_targets || 0) : 1
  return { min, max: Math.max(min, max) }
}

/**
 * 点头像后的目标列表：点已选目标 = 取消选中；选满后点新目标 = 替换最早选择的那个。
 * 只负责选择，发动统一走"确认发动"按钮。
 */
export function nextSkillTargets(current: readonly string[], playerId: string, maxTargets: number): string[] {
  if (current.includes(playerId)) {
    return current.filter((id) => id !== playerId)
  }
  const max = Math.max(1, maxTargets)
  const kept = current.length >= max ? current.slice(current.length - max + 1) : [...current]
  return [...kept, playerId]
}

/** 已选满时，下一次点新目标会被挤掉的那个（最早选择的）；没选满返回空串。 */
export function targetToBeReplaced(current: readonly string[], maxTargets: number): string {
  if (maxTargets <= 0 || current.length < maxTargets) return ''
  return current[0] || ''
}
