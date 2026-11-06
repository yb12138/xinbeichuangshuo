import type { AvailableSkill, PlayerView } from '../types/game'

/**
 * 技能 / 攻击目标的前端合法性预筛。
 * 只镜像后端能用已下发数据判断的规则（internal/engine/skill_target_rules.go、
 * skill_use_validation.go、各角色 module.go 里的 TargetRules）；判断不了的一律放行，交给后端。
 */

/** 与后端 model.TargetType 对齐。 */
export const SKILL_TARGET_TYPE = {
  None: 0,
  Self: 1,
  Enemy: 2,
  Ally: 3, // 队友，不含自己
  AllySelf: 4, // 我方角色，含自己
  Any: 5,
  Specific: 6, // 特殊逻辑，由后端 TargetRules 约束
} as const

/** 刺客潜行形态（model.FormAssassinStealth）：不能成为主动攻击目标。 */
export const STEALTH_FORM = 'assassin_stealth_form'

const BASIC_EFFECTS = new Set([
  'Shield', 'Weak', 'Poison',
  'SealFire', 'SealWater', 'SealEarth', 'SealWind', 'SealThunder',
  'PowerBlessing', 'SwiftBlessing',
])

export function isBasicEffect(effect?: string | null): boolean {
  return !!effect && BASIC_EFFECTS.has(effect)
}

export function hasFieldEffect(player: PlayerView | undefined, effect: string): boolean {
  if (!player || !Array.isArray(player.field)) return false
  return player.field.some((fc) => fc?.mode === 'Effect' && fc.effect === effect)
}

function hasAnyBasicFieldEffect(player: PlayerView): boolean {
  return Array.isArray(player.field) && player.field.some((fc) => fc?.mode === 'Effect' && isBasicEffect(fc.effect))
}

/**
 * 不能成为主动攻击目标：以服务端下发的 not_attack_targetable 为准（与后端 CanBeActiveAttackTarget 同源）。
 * 旧数据或测试里没有该字段时，回落到潜行形态 / 潜行效果的本地判断。
 */
export function cannotBeActiveAttackTarget(player: PlayerView): boolean {
  if (typeof player.not_attack_targetable === 'boolean') return player.not_attack_targetable
  return player.form === STEALTH_FORM || hasFieldEffect(player, 'Stealth')
}

type SlotRule = {
  /** 不能以自己为目标（TargetSelfOther）。 */
  notSelf?: boolean
  /** 目标治疗至少为多少（TargetCheckTargetMinHeal）。 */
  minHeal?: number
}

/**
 * 后端 policy.TargetRules 里针对单个目标的额外限制。
 * 目前这些技能都只有 1 个目标，规则对每个候选都适用。
 */
const SKILL_SLOT_RULES: Record<string, SlotRule> = {
  holy_lancer_punishment: { notSelf: true, minHeal: 1 }, // 惩戒
  bw_blazing_codex: { notSelf: true }, // 苍炎法典
  bw_heavenfire_cleave: { notSelf: true }, // 天火断空
}

export type SkillTargetContext = {
  me: string
  camp: string
}

function matchesTargetType(player: PlayerView, targetType: number, ctx: SkillTargetContext): boolean {
  switch (targetType) {
    case SKILL_TARGET_TYPE.None:
      return false
    case SKILL_TARGET_TYPE.Self:
      return player.id === ctx.me
    case SKILL_TARGET_TYPE.Enemy:
      return player.camp !== ctx.camp
    case SKILL_TARGET_TYPE.Ally:
      return player.camp === ctx.camp && player.id !== ctx.me
    case SKILL_TARGET_TYPE.AllySelf:
      return player.camp === ctx.camp
    default:
      // Any / Specific / 未知：阵营不限，具体限制交给后端。
      return true
  }
}

/** 单个玩家能否作为该技能的目标候选。 */
export function isLegalSkillTarget(skill: AvailableSkill, player: PlayerView, ctx: SkillTargetContext): boolean {
  if (!matchesTargetType(player, skill.target_type ?? 0, ctx)) return false

  const slot = SKILL_SLOT_RULES[skill.id]
  if (slot?.notSelf && player.id === ctx.me) return false
  if (slot?.minHeal !== undefined && (player.heal ?? 0) < slot.minHeal) return false

  // 封印破碎：只能选面前有基础效果的角色。
  if (skill.id === 'seal_break' && !hasAnyBasicFieldEffect(player)) return false

  // 放置基础效果：目标面前已有同种基础效果则不能再放（validateSkillFieldPlacement）。
  if (skill.place_card && isBasicEffect(skill.place_effect) && hasFieldEffect(player, skill.place_effect || '')) {
    return false
  }
  return true
}

export function legalSkillTargets(skill: AvailableSkill | null | undefined, players: PlayerView[], ctx: SkillTargetContext): PlayerView[] {
  if (!skill) return []
  return players.filter((player) => isLegalSkillTarget(skill, player, ctx))
}

/**
 * 当前已选目标能否提交：数量在上下限内、没有重复、每个都仍是合法候选。
 * bounds 由 skillTargetBounds 给出。
 */
export function isSkillTargetSelectionValid(
  selected: readonly string[],
  bounds: { min: number; max: number },
  candidateIds: ReadonlySet<string>,
): boolean {
  if (selected.length < bounds.min || selected.length > bounds.max) return false
  if (new Set(selected).size !== selected.length) return false
  return selected.every((id) => candidateIds.has(id))
}
