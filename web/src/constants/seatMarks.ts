/**
 * 座位上的对战标记清单（场上状态 / 盖牌 / 指示物），座位芯片与悬停展开面板共用。
 */
import type { PlayerView } from '../types/game'
import { COVER_COUNT_TOKEN_KEYS, COVER_MARKS, STATUS_MARKS, TOKEN_MARKS } from './battleMarks'
import { PERSISTENT_EFFECT_KEYS } from './persistentEffects'

/** 形态类键：由形态徽章展示，不进指示物段。 */
export const FORM_TOKEN_KEYS = new Set([
  'arbiter_form',
  'elf_ritual_form',
  'ms_shadow_form',
  'hom_burst_form',
  'prayer_form',
  'crk_hot_form',
  'onmyoji_form',
  'bw_flame_form',
  'hb_form',
  'ml_phantom_form',
  'bd_prisoner_form',
  'hero_exhaustion_form',
  'fighter_hundred_dragon_form',
  'mg_dark_form',
  'bp_bleed_form',
])

/** 没有图标定义时的文字胶囊（流程类提示等）。 */
export const TOKEN_DISPLAY: Record<string, { label: string; cls: string }> = {
  element: { label: '元素', cls: 'bg-cyan-800/70 text-cyan-100 border-cyan-500/40' },
  judgment: { label: '审判', cls: 'bg-fuchsia-800/70 text-fuchsia-100 border-fuchsia-500/40' },
  valkyrie_spirit: { label: '英灵', cls: 'bg-amber-800/70 text-amber-100 border-amber-500/40' },
  elf_blessing_count: { label: '祝福', cls: 'bg-teal-800/70 text-teal-100 border-teal-500/40' },
  css_blood: { label: '鲜血', cls: 'bg-rose-800/70 text-rose-100 border-rose-500/40' },
  prayer_rune: { label: '祈祷符文', cls: 'bg-yellow-900/70 text-yellow-100 border-yellow-500/40' },
  crk_blood_mark: { label: '血印', cls: 'bg-red-900/70 text-red-100 border-red-500/40' },
  hom_war_rune: { label: '战纹', cls: 'bg-indigo-900/70 text-indigo-100 border-indigo-500/40' },
  hom_magic_rune: { label: '魔纹', cls: 'bg-violet-900/70 text-violet-100 border-violet-500/40' },
  onmyoji_ghost_fire: { label: '鬼火', cls: 'bg-teal-900/70 text-teal-100 border-teal-500/40' },
  bw_rebirth: { label: '重生', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  mb_charge_count: { label: '充能', cls: 'bg-indigo-900/70 text-indigo-100 border-indigo-500/40' },
  bd_inspiration: { label: '灵感', cls: 'bg-violet-900/70 text-violet-100 border-violet-500/40' },
  ml_dark_release_next_attack_bonus: { label: '下次主动攻+伤', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  ml_dark_release_lock_turn: { label: '本回合锁技能', cls: 'bg-zinc-800/75 text-zinc-100 border-zinc-500/40' },
  hero_anger: { label: '怒气', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  hero_wisdom: { label: '知性', cls: 'bg-sky-900/70 text-sky-100 border-sky-500/40' },
  hero_calm_end_crystal_pending: { label: '止水回晶', cls: 'bg-cyan-900/70 text-cyan-100 border-cyan-500/40' },
  fighter_qi: { label: '斗气', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40' },
  hb_cannon: { label: '圣煌辉光炮', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  hb_faith: { label: '信仰', cls: 'bg-yellow-900/70 text-yellow-100 border-yellow-500/40' },
  ss_blue_soul: { label: '蓝色灵魂', cls: 'bg-blue-900/70 text-blue-100 border-blue-500/40' },
  ss_yellow_soul: { label: '黄色灵魂', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  mg_new_moon: { label: '新月', cls: 'bg-cyan-900/70 text-cyan-100 border-cyan-500/40' },
  mg_petrify: { label: '石化', cls: 'bg-zinc-900/70 text-zinc-100 border-zinc-500/40' },
  mg_dark_moon_count: { label: '暗月', cls: 'bg-indigo-900/70 text-indigo-100 border-indigo-500/40' },
  mg_next_attack_no_counter: { label: '下次攻不可应战', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  bp_shared_life_active: { label: '同生共死在场', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  bp_shared_life_bound: { label: '同生共死绑定', cls: 'bg-rose-950/75 text-rose-100 border-rose-400/50' },
  se_sword_qi: { label: '剑气', cls: 'bg-sky-900/70 text-sky-100 border-sky-500/40' },
  se_sword_soul_count: { label: '剑魂', cls: 'bg-slate-900/75 text-slate-100 border-slate-500/40' },
  bt_pupa: { label: '蛹', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  bt_cocoon_count: { label: '茧', cls: 'bg-indigo-900/70 text-indigo-100 border-indigo-500/40' },
  bt_wither_active: { label: '凋零生效', cls: 'bg-red-900/70 text-red-100 border-red-500/40' },
  bs_beast_soul: { label: '兽魂', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40' },
  bs_zanshin: { label: '残心', cls: 'bg-cyan-900/70 text-cyan-100 border-cyan-500/40' },
}

/** 内部流程标记：不展示给玩家。 */
export const HIDDEN_TOKEN_KEYS = new Set([
  'arbiter_skip_forced_doomsday',
  'arbiter_forced_doomsday_done_turn',
  'holy_lancer_block_sacred_strike',
  'holy_lancer_prayer_used_turn',
  // 新角色内部流程标记：不应直接展示给玩家
  'elf_ritual_release_waiting',
  'plague_block_immortal',
  'ms_yellow_spring_pending',
  'css_blood_cap',
  'css_blood_barrier_lock',
  'prayer_power_blessing_used',
  'prayer_swift_blessing_used',
  'bw_flame_release_pending',
  'bw_substitute_lock',
  'bw_mana_inversion_lock',
  'bw_pain_link_pending_discard',
  'bw_pain_link_pending_hits',
  'mb_magic_pierce_pending',
  'bd_descent_used_turn',
  'ml_stardust_pending',
  'ml_stardust_wait_discard',
  'ml_stardust_morale_before',
  'ml_fullness_next_attack_bonus',
  'bp_shared_life_active',
  'bp_shared_life_bound',
  'hero_exhaustion_release_pending',
  'hero_roar_active',
  'hero_calm_force_no_counter',
  'hero_dead_duel_pending',
  'fighter_attack_start_skill_lock',
  'fighter_charge_pending',
  'fighter_qiburst_force_no_counter',
  'hb_special_used_turn',
  'hb_auto_fill_done_turn',
  'hb_shard_miss_pending',
  // 灵符师妖力是盖牌数量派生展示，不作为头像 token chip 展示
  'sc_power_count',
  // 魔枪的回合态展示放在行动面板，不在头像 token 区重复显示
  'ml_dark_release_next_attack_bonus',
  'ml_dark_release_lock_turn',
  'mg_blasphemy_used_turn',
  'mg_blasphemy_pending',
  'bt_wither_pending',
  'adventurer_extract_last_gem',
  'adventurer_extract_last_crystal',
])

export interface SeatStatusMark {
  key: string
  label: string
  hint: string
  count: number
}

export interface SeatCoverMark {
  key: string
  label: string
  hint: string
  count: number
}

export interface SeatTokenMark {
  key: string
  label: string
  hint: string
  value: number
  /** 有图标定义：画菱形徽章；否则用文字胶囊（cls）。 */
  mark: boolean
  cap?: number
  cls: string
}

const FALLBACK_TOKEN_CLS = 'bg-gray-700/70 text-gray-100 border-gray-500/40'

/** 场上状态（Mode=Effect），同种合并计数。 */
export function seatStatusMarks(player: PlayerView, hidden: ReadonlySet<string> = new Set()): SeatStatusMark[] {
  const items = new Map<string, SeatStatusMark>()
  for (const fc of player.field ?? []) {
    if (fc.mode !== 'Effect' || !fc.effect || hidden.has(fc.effect)) continue
    const existing = items.get(fc.effect)
    if (existing) {
      existing.count += 1
      continue
    }
    const def = STATUS_MARKS[fc.effect]
    items.set(fc.effect, { key: fc.effect, label: def?.label ?? fc.effect, hint: def?.hint ?? '', count: 1 })
  }
  return [...items.values()]
}

/** 盖牌（Mode=Cover）：按种类合并为数量，内容已由服务端遮蔽。 */
export function seatCoverMarks(player: PlayerView): SeatCoverMark[] {
  const counts = new Map<string, number>()
  for (const fc of player.field ?? []) {
    if (fc.mode !== 'Cover' || !fc.effect) continue
    counts.set(fc.effect, (counts.get(fc.effect) ?? 0) + 1)
  }
  return [...counts.entries()].map(([key, count]) => {
    const def = COVER_MARKS[key]
    return { key, count, label: def?.label ?? '盖牌', hint: def?.hint ?? '' }
  })
}

/** 指示物：隐藏内部键、形态键，以及已由盖牌段表示的计数键。 */
export function seatTokenMarks(player: PlayerView): SeatTokenMark[] {
  const source = { ...(player.tokens ?? {}), ...(player.indicators ?? {}) }
  const hasCovers = (player.field ?? []).some((fc) => fc.mode === 'Cover')
  return Object.entries(source)
    .filter(
      ([key, value]) =>
        !HIDDEN_TOKEN_KEYS.has(key) &&
        !FORM_TOKEN_KEYS.has(key) &&
        // 持续效果由 seatPersistentEffects 单独展示
        !PERSISTENT_EFFECT_KEYS.has(key) &&
        !(COVER_COUNT_TOKEN_KEYS.has(key) && hasCovers) &&
        typeof value === 'number' &&
        value > 0,
    )
    .map(([key, value]) => {
      const mark = TOKEN_MARKS[key]
      const cfg = TOKEN_DISPLAY[key]
      return {
        key,
        value: value as number,
        label: mark?.label ?? cfg?.label ?? key,
        hint: mark?.hint ?? '',
        mark: !!mark,
        cap: mark?.cap,
        cls: cfg?.cls ?? FALLBACK_TOKEN_CLS,
      }
    })
}
