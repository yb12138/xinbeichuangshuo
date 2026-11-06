/**
 * 持续中的公开效果：技能结算后仍在生效、会影响后续决策的状态（凋零保护、额外回合、下一击加成……）。
 * 服务端以指示物（player.indicators）下发，键与 internal/server/stateview.PublicIndicators 一致。
 *
 * scope:
 *   - 'seat'      显示在持有者座位上
 *   - 'enemyCamp' 作用于持有者的对方阵营：显示在对方阵营士气旁（同时在持有者座位上标出来源）
 */
import type { PlayerView } from '../types/game'

export interface PersistentEffectDef {
  label: string
  hint: string
  scope: 'seat' | 'enemyCamp'
  /** 数值有意义时（+N 伤害、N 个额外回合）显示数值。 */
  showValue?: boolean
  /** 数值格式：'bonus' → "+N"，'count' → "×N"。 */
  valueFormat?: 'bonus' | 'count'
}

export const PERSISTENT_EFFECTS: Record<string, PersistentEffectDef> = {
  bt_wither_active: {
    label: '凋零',
    hint: '凋零生效中：直到蝶舞者下回合开始前，对方阵营士气最低为1。',
    scope: 'enemyCamp',
  },
  mg_extra_turn_pending: {
    label: '额外回合',
    hint: '苍白之月：本回合结束后，月之女神立即再进行一个回合。',
    scope: 'seat',
    showValue: true,
    valueFormat: 'count',
  },
  mg_next_attack_no_counter: {
    label: '不可应战',
    hint: '苍白之月：本回合下次主动攻击不可应战。',
    scope: 'seat',
  },
  ml_dark_release_next_attack_bonus: {
    label: '下击加伤',
    hint: '暗之解放：本回合下次主动攻击伤害额外增加。',
    scope: 'seat',
    showValue: true,
    valueFormat: 'bonus',
  },
  ml_fullness_next_attack_bonus: {
    label: '下击加伤',
    hint: '充盈：本回合下次主动攻击伤害额外增加。',
    scope: 'seat',
    showValue: true,
    valueFormat: 'bonus',
  },
}

export const PERSISTENT_EFFECT_KEYS = new Set(Object.keys(PERSISTENT_EFFECTS))

export interface PersistentEffectView {
  key: string
  label: string
  hint: string
  valueText: string
}

function valueText(def: PersistentEffectDef, value: number): string {
  if (!def.showValue) return ''
  if (def.valueFormat === 'bonus') return `+${value}`
  return value > 1 ? `×${value}` : ''
}

/** 持有者座位上显示的持续效果（含作用于对方阵营的效果来源）。 */
export function seatPersistentEffects(player: PlayerView): PersistentEffectView[] {
  const out: PersistentEffectView[] = []
  for (const [key, raw] of Object.entries(player.indicators ?? {})) {
    const def = PERSISTENT_EFFECTS[key]
    const value = typeof raw === 'number' ? raw : 0
    if (!def || value <= 0) continue
    out.push({ key, label: def.label, hint: def.hint, valueText: valueText(def, value) })
  }
  return out
}

export interface CampEffectView {
  key: string
  label: string
  hint: string
  sourceId: string
}

/** 作用于某阵营的持续效果（如凋零：对方阵营士气最低为 1），按受影响阵营汇总。 */
export function campPersistentEffects(players: Record<string, PlayerView>, camp: 'Red' | 'Blue'): CampEffectView[] {
  const out: CampEffectView[] = []
  for (const player of Object.values(players)) {
    for (const [key, raw] of Object.entries(player.indicators ?? {})) {
      const def = PERSISTENT_EFFECTS[key]
      if (!def || def.scope !== 'enemyCamp' || typeof raw !== 'number' || raw <= 0) continue
      if (player.camp === camp) continue
      out.push({ key, label: def.label, hint: def.hint, sourceId: player.id })
    }
  }
  return out
}
