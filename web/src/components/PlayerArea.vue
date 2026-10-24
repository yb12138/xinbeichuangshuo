<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, ref, watch } from 'vue'
import type { PlayerView } from '../types/game'
import { useBattleFxStore } from '../stores/battlefx.store'
import { useBattleReviewStore } from '../stores/battleReview.store'
import { useInterruptStore } from '../stores/interrupt.store'
import { useUiStore } from '../stores/ui.store'
import { useBattleInteractionState } from '../composables/useBattleInteractionState'
import StatusEffectIcon from './StatusIcons/StatusEffectIcon.vue'

const battleFxStore = useBattleFxStore()
const battleReviewStore = useBattleReviewStore()
const interruptStore = useInterruptStore()
const uiStore = useUiStore()
const { damageEffects } = storeToRefs(battleFxStore)
const { currentPrompt, actionMode, selectedHandIndexForAction, skillMode } = storeToRefs(interruptStore)
const { skillModalCharacterId } = storeToRefs(uiStore)
const { getCharacter, getRoleDisplayName, isPromptForMe } = useBattleInteractionState()

const props = defineProps<{
  player: PlayerView
  isMe?: boolean
  isOpponent?: boolean
  selectable?: boolean
  selected?: boolean
  debugTargetReason?: string
  compact?: boolean
  turnOrder?: number
  soulLinkBindingText?: string
  soulLinkBindingTitle?: string
  fighterHundredDragonText?: string
  fighterHundredDragonTitle?: string
  fighterHundredDragonRole?: 'source' | 'bound'
  bloodSharedLifeText?: string
  bloodSharedLifeTitle?: string
  bloodSharedLifeRole?: 'source' | 'bound'
  testIdSuffix?: string
}>()

const emit = defineEmits<{
  select: [playerId: string]
}>()

const charInfo = computed(() => (props.player.role ? getCharacter(props.player.role) : null))
const roleDisplayName = computed(() => getRoleDisplayName(props.player.role))
const playerDisplayName = computed(() => props.player.name || props.player.id)

// 立绘路径：/characters/{role}.png，role 与角色ID对应
const characterImageSrc = computed(() => {
  const role = props.player.role
  if (!role) return ''
  return `/characters/${role}.png`
})

const showCharacterImage = ref(true)
watch(() => props.player.role, () => {
  showCharacterImage.value = true
})
function onCharImageError() {
  showCharacterImage.value = false
}

function openSkillModal(event?: MouseEvent) {
  if (props.selectable) {
    emit('select', props.player.id)
    return
  }
  if (!props.player.role) return
  if (skillModalCharacterId.value === props.player.role) {
    uiStore.openSkillModal(null)
    return
  }
  const anchorEl = event?.currentTarget as HTMLElement | null
  if (anchorEl) {
    const rect = anchorEl.getBoundingClientRect()
    uiStore.openSkillModal(props.player.role, {
      x: rect.left,
      y: rect.top,
      width: rect.width,
      height: rect.height,
    })
    return
  }
  uiStore.openSkillModal(props.player.role)
}

const campClass = computed(() => {
  return props.player.camp === 'Red' ? 'camp-red' : 'camp-blue'
})

const handCount = computed(() => {
  return props.isMe ? (props.player.hand?.length ?? props.player.hand_count) : props.player.hand_count
})

const maxHand = computed(() => {
  const limit = props.player.max_hand
  if (typeof limit === 'number' && limit >= 0) return limit
  return handCount.value
})

const isActive = computed(() => !!props.player.is_active)

const isHandOver = computed(() => handCount.value > maxHand.value)

const campLabel = computed(() => (props.player.camp === 'Red' ? '红方' : '蓝方'))

// 治疗以心形点阵展示；上限超过 4 或为 0 时回退为 "x/y" 文本
const HEAL_PIP_LIMIT = 4
const healPips = computed<boolean[] | null>(() => {
  const max = props.player.max_heal ?? 0
  if (max <= 0 || max > HEAL_PIP_LIMIT) return null
  const heal = props.player.heal ?? 0
  return Array.from({ length: max }, (_, i) => i < heal)
})

type EffectOverlay = {
  gradient: string
  animation: string
  opacity: number
  duration: string
  backgroundSize?: string
}

const EFFECT_DISPLAY: Record<string, { icon: string; label: string; cls: string; overlay?: EffectOverlay }> = {
  Shield: { icon: '🛡️', label: '圣盾', cls: 'bg-yellow-800/60' },
  Poison: { icon: '☠️', label: '中毒', cls: 'bg-green-800/60' },
  Weak: { icon: '💫', label: '虚弱', cls: 'bg-purple-800/60' },
  SealFire: { icon: '🔥', label: '火封印', cls: 'bg-red-800/60', overlay: {
    gradient: 'conic-gradient(from 0deg, rgba(239,68,68,0.7) 0%, rgba(239,68,68,0.5) 25%, rgba(239,68,68,0.7) 50%, rgba(239,68,68,0.5) 75%, rgba(239,68,68,0.7) 100%)',
    animation: 'effectSealRotate', opacity: 0.5, duration: '12s',
  }},
  SealWater: { icon: '💧', label: '水封印', cls: 'bg-blue-800/60', overlay: {
    gradient: 'conic-gradient(from 0deg, rgba(59,130,246,0.7) 0%, rgba(59,130,246,0.5) 25%, rgba(59,130,246,0.7) 50%, rgba(59,130,246,0.5) 75%, rgba(59,130,246,0.7) 100%)',
    animation: 'effectSealRotate', opacity: 0.5, duration: '12s',
  }},
  SealEarth: { icon: '🪨', label: '地封印', cls: 'bg-amber-800/60', overlay: {
    gradient: 'conic-gradient(from 0deg, rgba(217,119,6,0.7) 0%, rgba(217,119,6,0.5) 25%, rgba(217,119,6,0.7) 50%, rgba(217,119,6,0.5) 75%, rgba(217,119,6,0.7) 100%)',
    animation: 'effectSealRotate', opacity: 0.5, duration: '12s',
  }},
  SealWind: { icon: '🌪️', label: '风封印', cls: 'bg-teal-800/60', overlay: {
    gradient: 'conic-gradient(from 0deg, rgba(20,184,166,0.7) 0%, rgba(20,184,166,0.5) 25%, rgba(20,184,166,0.7) 50%, rgba(20,184,166,0.5) 75%, rgba(20,184,166,0.7) 100%)',
    animation: 'effectSealRotate', opacity: 0.5, duration: '12s',
  }},
  SealThunder: { icon: '⚡', label: '雷封印', cls: 'bg-indigo-800/60', overlay: {
    gradient: 'conic-gradient(from 0deg, rgba(99,102,241,0.7) 0%, rgba(99,102,241,0.5) 25%, rgba(99,102,241,0.7) 50%, rgba(99,102,241,0.5) 75%, rgba(99,102,241,0.7) 100%)',
    animation: 'effectSealRotate', opacity: 0.5, duration: '12s',
  }},
  FiveElementsBind: { icon: '⛓️', label: '五系束缚', cls: 'bg-gray-700/80', overlay: {
    gradient: 'conic-gradient(from 0deg, rgba(107,114,128,0.7) 0%, rgba(107,114,128,0.5) 25%, rgba(107,114,128,0.7) 50%, rgba(107,114,128,0.5) 75%, rgba(107,114,128,0.7) 100%)',
    animation: 'effectSealRotate', opacity: 0.45, duration: '10s',
  }},
  RoseCourtyard: { icon: '🌹', label: '血蔷薇庭院', cls: 'bg-rose-900/75' },
  PowerBlessing: { icon: '✨', label: '威力赐福', cls: 'bg-orange-900/75' },
  SwiftBlessing: { icon: '🪽', label: '迅捷赐福', cls: 'bg-cyan-900/75' },
  Stealth: { icon: '👤', label: '潜行', cls: 'bg-gray-700/80' },
  BardEternalMovement: { icon: '🎼', label: '永恒乐章', cls: 'bg-violet-900/75' },
  HeroTaunt: { icon: '⚔️', label: '挑衅', cls: 'bg-red-900/75' },
  SoulLink: { icon: '🧿', label: '灵魂链接', cls: 'bg-indigo-900/75' },
  BloodSharedLife: { icon: '🩸', label: '同生共死', cls: 'bg-rose-900/75' },
}

// @ts-expect-error fieldEffects computed is used for future field effect display feature
const fieldEffects = computed(() => {
  if (!props.player.field?.length) return []
  return props.player.field
    .filter(fc => fc.mode === 'Effect' && fc.effect)
    .map(fc => EFFECT_DISPLAY[fc.effect] || { icon: '✦', label: fc.effect, cls: 'bg-gray-700/60' })
})

const activeOverlays = computed(() => {
  if (!props.player.field?.length) return []
  return props.player.field
    .filter(fc => fc.mode === 'Effect' && fc.effect)
    .map(fc => EFFECT_DISPLAY[fc.effect]?.overlay)
    .filter((o): o is EffectOverlay => !!o)
})

const hasStealth = computed(() =>
  props.player.field?.some(fc => fc.mode === 'Effect' && fc.effect === 'Stealth') ?? false
)

// 需要显示图标的状态效果（排除场地效果）
const ICON_EFFECTS = new Set([
  'Shield', 'Poison', 'Weak',
  'SealFire', 'SealWater', 'SealEarth', 'SealWind', 'SealThunder',
  'FiveElementsBind', 'Stealth',
  'PowerBlessing', 'SwiftBlessing',
  'BardEternalMovement',
])

const statusIconEffects = computed(() => {
  if (!props.player.field?.length) return []
  return props.player.field
    .filter(fc => fc.mode === 'Effect' && fc.effect && ICON_EFFECTS.has(fc.effect))
    .map(fc => ({
      effect: fc.effect!,
      title: EFFECT_DISPLAY[fc.effect]?.label || fc.effect,
    }))
    .slice(0, 3)
})

const isTargetingContext = computed(() => {
  if (props.selectable || props.selected) return true
  if ((actionMode.value === 'attack' || actionMode.value === 'magic') && selectedHandIndexForAction.value !== null) return true
  if (skillMode.value === 'choosing_target') return true
  const prompt = currentPrompt.value
  if (!prompt || !isPromptForMe.value) return false
  if (prompt.presentation?.kind === 'target_picker') return true
  if (prompt.type === 'choose_target') return true
  if ((prompt.counter_target_ids?.length ?? 0) > 0) return true
  return (prompt.options ?? []).some((option: any) => !!option?.target_id)
})

const isTargetBlocked = computed(() =>
  isTargetingContext.value && !isActive.value && !props.selectable && !props.selected && !props.isMe
)

const portraitBadges = computed(() => {
  const badges: Array<{
    key: string
    text: string
    title: string
    cls: string
    icon?: string
  }> = []

  if (typeof props.turnOrder === 'number') {
    badges.push({
      key: 'turn-order',
      text: `#${props.turnOrder}`,
      title: `行动顺序 #${props.turnOrder}`,
      cls: 'player-portrait-badge--turn-order',
    })
  }

  if (props.soulLinkBindingText) {
    badges.push({
      key: 'soul-link',
      text: props.soulLinkBindingText,
      title: props.soulLinkBindingTitle || props.soulLinkBindingText,
      cls: 'player-portrait-badge--soul-link',
    })
  }

  if (props.fighterHundredDragonText) {
    badges.push({
      key: 'fighter-hundred-dragon',
      text: props.fighterHundredDragonText,
      title: props.fighterHundredDragonTitle || props.fighterHundredDragonText,
      cls: [
        'player-portrait-badge--fighter-hundred-dragon',
        props.fighterHundredDragonRole ? `player-portrait-badge--fighter-hundred-dragon-${props.fighterHundredDragonRole}` : '',
      ].filter(Boolean).join(' '),
    })
  }

  if (props.bloodSharedLifeText) {
    badges.push({
      key: 'blood-shared-life',
      text: props.bloodSharedLifeText,
      title: props.bloodSharedLifeTitle || props.bloodSharedLifeText,
      cls: [
        'player-portrait-badge--blood-shared-life',
        props.bloodSharedLifeRole ? `player-portrait-badge--blood-shared-life-${props.bloodSharedLifeRole}` : '',
      ].filter(Boolean).join(' '),
      icon: 'BloodSharedLife',
    })
  }

  return badges
})

// 当前玩家身上的伤害特效（暴血）
const myDamageEffects = computed(() =>
  damageEffects.value.filter(d => d.targetId === props.player.id)
)

const FORM_DISPLAY: Record<string, { label: string; cls: string }> = {
  heroic_form: { label: '英灵形态', cls: 'bg-amber-800/70 text-amber-100 border-amber-500/40' },
  arbiter_judgment_form: { label: '审判形态', cls: 'bg-violet-800/70 text-violet-100 border-violet-500/40' },
  elf_archer_ritual_form: { label: '祝福形态', cls: 'bg-emerald-800/70 text-emerald-100 border-emerald-500/40' },
  magic_swordsman_shadow_form: { label: '暗影形态', cls: 'bg-slate-800/80 text-slate-100 border-slate-500/40' },
  war_homunculus_burst_form: { label: '蓄势形态', cls: 'bg-purple-900/70 text-purple-100 border-purple-500/40' },
  prayer_master_prayer_form: { label: '祈祷形态', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  crimson_knight_hot_blooded_form: { label: '热血形态', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40' },
  onmyoji_shikigami_form: { label: '式神形态', cls: 'bg-sky-900/70 text-sky-100 border-sky-500/40' },
  blaze_witch_flame_form: { label: '烈焰形态', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40' },
  beast_samurai_iaijutsu_form: { label: '居合形态', cls: 'bg-zinc-900/75 text-zinc-100 border-zinc-500/40' },
  holy_bow_holy_glory_form: { label: '圣煌形态', cls: 'bg-sky-900/70 text-sky-100 border-sky-500/40' },
  magic_lance_phantom_form: { label: '幻影形态', cls: 'bg-slate-900/75 text-slate-100 border-slate-500/40' },
  bard_eternal_prisoner_form: { label: '囚徒形态', cls: 'bg-purple-900/70 text-purple-100 border-purple-500/40' },
  hero_exhaustion_form: { label: '精疲力竭', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  fighter_hundred_dragon_form: { label: '幻龙拳形态', cls: 'bg-red-900/70 text-red-100 border-red-500/40' },
  moon_goddess_dark_moon_form: { label: '暗月形态', cls: 'bg-slate-900/75 text-slate-100 border-slate-500/40' },
  blood_witch_bleeding_form: { label: '流血形态', cls: 'bg-red-900/70 text-red-100 border-red-500/40' },
}

const FORM_TOKEN_KEYS = new Set([
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

const TOKEN_DISPLAY: Record<string, { label: string; cls: string }> = {
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

const HIDDEN_TOKEN_KEYS = new Set([
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

const tokenIndicators = computed(() => {
  const indicatorSource = {
    ...(props.player.tokens ?? {}),
    ...(props.player.indicators ?? {}),
  }
  const entries = Object.entries(indicatorSource)
    .filter(([key, value]) => !HIDDEN_TOKEN_KEYS.has(key) && !FORM_TOKEN_KEYS.has(key) && typeof value === 'number' && value > 0)
    .map(([key, value]) => {
      const cfg = TOKEN_DISPLAY[key]
      return {
        key,
        value,
        label: cfg?.label ?? key,
        cls: cfg?.cls ?? 'bg-gray-700/70 text-gray-100 border-gray-500/40'
      }
  })
  return entries
})

const allIndicators = computed(() => tokenIndicators.value)

const formIndicator = computed(() => {
  const form = props.player.form
  if (!form) return null
  const cfg = FORM_DISPLAY[form]
  if (!cfg) {
    return {
      key: form,
      label: '形态',
      cls: 'bg-gray-700/70 text-gray-100 border-gray-500/40',
    }
  }
  return {
    key: form,
    label: cfg.label,
    cls: cfg.cls,
  }
})

const showStealthBlockedHint = computed(() => {
  if (props.selectable) return false
  if (!props.isOpponent) return false
  if (actionMode.value !== 'attack') return false
  if (selectedHandIndexForAction.value === null) return false
  return !!props.player.field?.some((fc) => fc.mode === 'Effect' && fc.effect === 'Stealth')
})

const targetDebugEnabled = computed(() => {
  if (typeof window === 'undefined') return false
  const query = new URLSearchParams(window.location.search)
  return import.meta.env.DEV || query.has('debug') || query.has('debug_target')
})

function logTargetDebug(stage: string, payload?: Record<string, unknown>) {
  if (!targetDebugEnabled.value) return
  const data = {
    stage,
    playerId: props.player.id,
    playerName: props.player.name,
    selectable: !!props.selectable,
    selected: !!props.selected,
    reason: props.debugTargetReason || '',
    actionMode: actionMode.value,
    skillMode: skillMode.value,
    promptPresentationKind: currentPrompt.value?.presentation?.kind || '',
    isPromptForMe: isPromptForMe.value,
    ...payload
  }
  console.log('[TargetDebug][PlayerArea]', data)
  battleReviewStore.addLog(
    `[TargetDebug][PlayerArea] ${stage} p=${props.player.id} selectable=${String(!!props.selectable)} selected=${String(!!props.selected)} reason=${props.debugTargetReason || ''}`
  )
}

function handleClick(e: MouseEvent) {
  logTargetDebug('click_received')
  if (props.selectable) {
    // 点击技能按钮不触发选目标
    if ((e.target as HTMLElement).closest('.btn-show-skills')) {
      logTargetDebug('click_ignored_skill_button')
      return
    }
    logTargetDebug('emit_select')
    emit('select', props.player.id)
    return
  }
  logTargetDebug('click_blocked_not_selectable')
}
</script>

<template>
  <div
    class="player-area transition-all duration-300 overflow-hidden relative"
    :class="[
      compact
        ? 'player-area--compact min-w-[102px] max-w-[120px] sm:min-w-[116px] sm:max-w-[136px] 2xl:min-w-[128px] 2xl:max-w-[152px]'
        : 'player-area--full min-w-[116px] sm:min-w-[128px] 2xl:min-w-[152px]',
      campClass,
      isMe ? 'player-area--me' : '',
      isActive ? 'player-area--active' : '',
      isTargetBlocked ? 'player-area--target-blocked' : '',
      showStealthBlockedHint && !isActive ? 'opacity-60 grayscale saturate-75' : '',
      hasStealth ? 'player-area--stealth' : '',
      selectable ? 'player-area--targetable cursor-pointer hover:scale-[1.03]' : '',
      selected ? 'player-area--selected' : ''
    ]"
    :data-testid="`player-area-${player.id}${testIdSuffix ? `-${testIdSuffix}` : ''}`"
    @click="handleClick"
  >
    <!-- 立绘区：填满名牌条以上的空间 -->
    <div class="seat-portrait">
      <img
        v-if="characterImageSrc && showCharacterImage"
        :src="characterImageSrc"
        :alt="charInfo?.name || player.name"
        class="character-portrait-fill"
        @error="onCharImageError"
      >
      <div
        v-else
        class="character-portrait-placeholder-fill"
      >
        {{ (charInfo?.name || player.name || '?').charAt(0) }}
      </div>

      <!-- 状态效果叠加层 -->
      <div
        v-for="(overlay, idx) in activeOverlays"
        :key="'fx-' + idx"
        class="effect-overlay-layer"
        :style="{
          background: overlay.gradient,
          backgroundSize: overlay.backgroundSize || 'auto',
          opacity: overlay.opacity,
          'animation-name': overlay.animation,
          'animation-duration': overlay.duration,
        }"
      />

      <div class="player-state-badges">
        <span v-if="isMe" class="player-state-badge player-state-badge--me">你</span>
        <span v-if="isActive" class="player-state-badge player-state-badge--active">当前行动</span>
        <span v-if="selectable" class="player-state-badge player-state-badge--targetable">可选</span>
      </div>

      <div v-if="portraitBadges.length" class="player-portrait-badges">
        <div
          v-for="badge in portraitBadges"
          :key="badge.key"
          class="player-portrait-badge"
          :class="badge.cls"
          :title="badge.title"
        >
          <StatusEffectIcon v-if="badge.icon" :effect="badge.icon" class="player-portrait-badge-icon" />
          <span class="player-portrait-badge-text">{{ badge.text }}</span>
        </div>
      </div>

      <!-- 完整模式下的状态图标列（紧凑模式由 GameBoard 状态栏负责） -->
      <div v-if="statusIconEffects.length" class="player-overlay-effects">
        <div
          v-for="(statusEffect, idx) in statusIconEffects"
          :key="`status-${statusEffect.effect}-${idx}`"
          class="effect-icon-item"
          :title="statusEffect.title"
        >
          <StatusEffectIcon
            :effect="statusEffect.effect"
          />
        </div>
      </div>

      <!-- 名牌条上沿的小型芯片：形态 / 宝石水晶 / 指示物 -->
      <div
        v-if="formIndicator || (player.gem || 0) + (player.crystal || 0) > 0 || allIndicators.length"
        class="seat-chip-row"
      >
        <span v-if="formIndicator" class="seat-chip form-badge" :title="formIndicator.label">
          {{ formIndicator.label }}
        </span>
        <span
          v-if="(player.gem || 0) + (player.crystal || 0) > 0"
          class="player-overlay-resource"
        >
          <span v-if="player.gem" class="seat-chip seat-chip--gem" :title="`宝石 ${player.gem}`">
            <span aria-hidden="true" class="seat-chip-gem-mark" />{{ player.gem }}
          </span>
          <span v-if="player.crystal" class="seat-chip seat-chip--crystal" :title="`水晶 ${player.crystal}`">
            <span aria-hidden="true" class="seat-chip-crystal-mark" />{{ player.crystal }}
          </span>
        </span>
        <span v-if="allIndicators.length" class="player-overlay-tokens">
          <span
            v-for="token in allIndicators"
            :key="token.key"
            class="seat-chip seat-chip--token"
            :title="`${token.label}: ${token.value}`"
          >
            <span>{{ token.label }}</span>
            <span class="seat-chip-value">×{{ token.value }}</span>
          </span>
        </span>
      </div>
    </div>

    <div class="player-overlay">
      <!-- 阵营色名牌条 -->
      <div class="player-overlay-title-row">
        <span class="sr-only">{{ campLabel }}</span>
        <span class="player-overlay-role" :title="roleDisplayName">{{ roleDisplayName }}</span>
        <button
          v-if="charInfo?.skills?.length"
          type="button"
          class="btn-show-skills"
          @click.stop="openSkillModal($event)"
        >
          技能
        </button>
      </div>

      <div class="player-overlay-player" :title="`玩家：${playerDisplayName}`">
        玩家：{{ playerDisplayName }}
      </div>

      <!-- 底栏：治疗 | 手牌 -->
      <div class="player-overlay-stats">
        <div
          class="player-overlay-stat player-overlay-stat--heal"
          :title="`治疗 ${player.heal}/${player.max_heal}`"
          :aria-label="`治疗 ${player.heal}/${player.max_heal}`"
        >
          <template v-if="healPips">
            <svg
              v-for="(filled, i) in healPips"
              :key="`heal-pip-${i}`"
              class="seat-heal-pip"
              :class="filled ? 'seat-heal-pip--filled' : ''"
              viewBox="0 0 16 16"
              aria-hidden="true"
            >
              <path d="M8 14s-5.5-3.4-5.5-7.2A3 3 0 0 1 8 4.6a3 3 0 0 1 5.5 2.2C13.5 10.6 8 14 8 14z" />
            </svg>
          </template>
          <template v-else>
            <svg class="seat-stat-icon" viewBox="0 0 16 16" aria-hidden="true">
              <path d="M8 14s-5.5-3.4-5.5-7.2A3 3 0 0 1 8 4.6a3 3 0 0 1 5.5 2.2C13.5 10.6 8 14 8 14z" />
            </svg>
            <span class="seat-stat-num">{{ player.heal }}/{{ player.max_heal }}</span>
          </template>
        </div>
        <span class="seat-stat-divider" aria-hidden="true" />
        <div
          class="player-overlay-stat player-overlay-stat--hand"
          :class="isHandOver ? 'player-overlay-stat--over' : ''"
          :title="`手牌 ${handCount}/${maxHand}`"
        >
          <svg class="seat-stat-icon" viewBox="0 0 16 16" aria-hidden="true">
            <rect x="2.5" y="3.5" width="7" height="10" rx="1.2" transform="rotate(-10 6 8.5)" />
            <rect x="6.5" y="2.5" width="7" height="10" rx="1.2" transform="rotate(8 10 7.5)" />
          </svg>
          <span class="seat-stat-num">{{ handCount }}/{{ maxHand }}</span>
        </div>
      </div>

      <div v-if="showStealthBlockedHint" class="seat-stealth-hint">
        潜行状态无法选中
      </div>
    </div>

    <!-- 镀铬细边框与四角饰件 -->
    <div class="seat-frame" aria-hidden="true">
      <span class="seat-corner seat-corner--tl" />
      <span class="seat-corner seat-corner--tr" />
      <span class="seat-corner seat-corner--bl" />
      <span class="seat-corner seat-corner--br" />
    </div>

    <!-- 伤害暴血特效 overlay -->
    <div
      v-for="eff in myDamageEffects"
      :key="eff.id"
      class="damage-burst-overlay absolute inset-0 flex items-center justify-center pointer-events-none z-10"
    >
      <div class="damage-burst-inner">
        <div class="damage-number">-{{ eff.damage }}</div>
        <div class="damage-splash" />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ===== 座位卡：黑漆底板 + 镀铬细边框（颜色只取 battle-theme.css 的 --bt-*） ===== */
.player-area {
  padding: 0;
  isolation: isolate;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--bt-hairline);
  border-radius: var(--bt-radius);
  background: var(--bt-bg-0);
  box-shadow: 0 0 0 2px var(--bt-bg-0), 0 6px 18px var(--bt-panel);
}

.player-area--compact {
  height: 100%;
}

.player-area--full {
  height: 256px;
}

/* 内侧第二道细线 + 四角饰件，压在立绘之上 */
.seat-frame {
  position: absolute;
  inset: 2px;
  z-index: 7;
  border: 1px solid var(--bt-hairline-soft);
  border-radius: var(--bt-radius-sm);
  pointer-events: none;
  transition: border-color 0.3s ease;
}

.seat-corner {
  position: absolute;
  width: 8px;
  height: 8px;
  border-color: var(--bt-chrome-hi);
  border-style: solid;
  border-width: 0;
  filter: drop-shadow(0 0 1px var(--bt-bg-0));
  transition: border-color 0.3s ease;
}

.seat-corner--tl { top: -2px; left: -2px; border-top-width: 1.5px; border-left-width: 1.5px; border-top-left-radius: 3px; }
.seat-corner--tr { top: -2px; right: -2px; border-top-width: 1.5px; border-right-width: 1.5px; border-top-right-radius: 3px; }
.seat-corner--bl { bottom: -2px; left: -2px; border-bottom-width: 1.5px; border-left-width: 1.5px; border-bottom-left-radius: 3px; }
.seat-corner--br { bottom: -2px; right: -2px; border-bottom-width: 1.5px; border-right-width: 1.5px; border-bottom-right-radius: 3px; }

/* ===== 状态：当前行动（金框）/ 可选（冷光）/ 选中 / 我 ===== */
.player-area--me {
  border-color: var(--bt-chrome);
}

/* GameBoard 会把 --active 座位的 box-shadow 清空，这里用 !important 保住金色光晕 */
.player-area--active {
  border-color: var(--bt-active) !important;
  box-shadow:
    0 0 0 1px var(--bt-active),
    0 0 14px var(--bt-active-glow),
    0 0 26px var(--bt-active-glow) !important;
}

.player-area--active .seat-frame {
  border-color: var(--bt-active-glow);
}

.player-area--active .seat-corner {
  border-color: var(--bt-active);
}

.player-area--targetable {
  border-color: var(--bt-select);
  box-shadow:
    0 0 0 1px var(--bt-select-glow),
    0 0 14px var(--bt-select-glow);
}

.player-area--targetable .seat-corner {
  border-color: var(--bt-select);
}

.player-area--selected {
  transform: scale(1.05);
  border-color: var(--bt-select) !important;
  box-shadow:
    0 0 0 1px var(--bt-select),
    0 0 16px var(--bt-select-glow),
    0 0 28px var(--bt-select-glow);
}

.player-area--selected .seat-frame,
.player-area--selected .seat-corner {
  border-color: var(--bt-select);
}

@media (prefers-reduced-motion: no-preference) {
  .player-area--active {
    animation: seatActiveGlow 2.4s ease-in-out infinite;
  }

  .player-area--selected {
    animation: seatSelectGlow 1.8s ease-in-out infinite;
  }
}

@keyframes seatActiveGlow {
  0%, 100% {
    box-shadow: 0 0 0 1px var(--bt-active), 0 0 10px var(--bt-active-glow), 0 0 20px var(--bt-active-glow);
  }
  50% {
    box-shadow: 0 0 0 1px var(--bt-active), 0 0 18px var(--bt-active-glow), 0 0 32px var(--bt-active-glow);
  }
}

@keyframes seatSelectGlow {
  0%, 100% {
    box-shadow: 0 0 0 1px var(--bt-select), 0 0 12px var(--bt-select-glow), 0 0 22px var(--bt-select-glow);
  }
  50% {
    box-shadow: 0 0 0 1px var(--bt-select), 0 0 20px var(--bt-select-glow), 0 0 34px var(--bt-select-glow);
  }
}

.player-area--target-blocked {
  filter: grayscale(0.35) saturate(0.72) brightness(0.82);
}

/* 潜行整卡变灰 */
.player-area--stealth {
  filter: grayscale(0.7) brightness(0.85);
}

.player-area--active.player-area--target-blocked,
.player-area--active.player-area--stealth {
  filter: none;
}

/* ===== 立绘 ===== */
.seat-portrait {
  position: relative;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  border-radius: var(--bt-radius-sm) var(--bt-radius-sm) 0 0;
  background: var(--bt-bg-1);
}

/* 立绘底部压暗，保证芯片与名牌可读 */
.seat-portrait::after {
  content: '';
  position: absolute;
  inset: auto 0 0 0;
  height: 40%;
  background: linear-gradient(180deg, transparent, var(--bt-panel));
  pointer-events: none;
  z-index: 2;
}

.character-portrait-fill {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 15%;
  z-index: 1;
}

.character-portrait-placeholder-fill {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(circle at 50% 40%, var(--bt-panel-raised), var(--bt-bg-0));
  color: var(--bt-chrome);
  font-family: var(--bt-font-num);
  font-weight: 700;
  font-size: 28px;
  z-index: 1;
}

/* 状态效果叠加层 */
.effect-overlay-layer {
  position: absolute;
  inset: 0;
  z-index: 3;
  pointer-events: none;
  animation-timing-function: ease-in-out;
  animation-iteration-count: infinite;
}

@media (prefers-reduced-motion: reduce) {
  .effect-overlay-layer {
    animation: none !important;
  }
}

/* ===== 状态小徽章（你 / 当前行动 / 可选） ===== */
.player-state-badges {
  position: absolute;
  top: 5px;
  right: 5px;
  z-index: 6;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 3px;
  pointer-events: none;
}

.player-state-badge {
  min-height: 16px;
  padding: 1px 6px;
  border-radius: 999px;
  border: 1px solid var(--bt-hairline);
  background: var(--bt-panel);
  color: var(--bt-text);
  font-size: 9px;
  font-weight: 800;
  line-height: 1.35;
  letter-spacing: 0.04em;
  box-shadow: 0 0 0 1px var(--bt-bg-0);
}

.player-state-badge--me {
  border-color: var(--bt-chrome-hi);
  color: var(--bt-chrome-hi);
}

.player-state-badge--active {
  border-color: var(--bt-active);
  color: var(--bt-active);
}

.player-state-badge--targetable {
  border-color: var(--bt-select);
  color: var(--bt-select);
}

/* ===== 立绘徽章（行动序号 / 灵魂链接 / 幻龙 / 同生共死） ===== */
.player-portrait-badges {
  position: absolute;
  top: 5px;
  left: 5px;
  z-index: 5;
  max-width: calc(100% - 58px);
  display: flex;
  flex-direction: column;
  gap: 3px;
  pointer-events: auto;
}

.player-portrait-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  width: fit-content;
  max-width: 100%;
  min-height: 16px;
  padding: 1px 6px;
  border-radius: var(--bt-radius-sm);
  border: 1px solid var(--bt-hairline);
  background: var(--bt-panel);
  color: var(--bt-text);
  font-size: 9px;
  font-weight: 800;
  line-height: 1.3;
  box-shadow: 0 0 0 1px var(--bt-bg-0);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.player-portrait-badge-text {
  overflow: hidden;
  text-overflow: ellipsis;
}

.player-portrait-badge-icon {
  width: 12px;
  height: 12px;
  flex-shrink: 0;
}

/* 行动序号：镀铬圆章 + 衬线数字 */
.player-portrait-badge--turn-order {
  min-width: 22px;
  justify-content: center;
  border-radius: 999px;
  border-color: var(--bt-chrome);
  color: var(--bt-chrome-hi);
  font-family: var(--bt-font-num);
  font-size: 10px;
}

.player-portrait-badge--soul-link {
  border-color: var(--bt-violet);
}

.player-portrait-badge--fighter-hundred-dragon,
.player-portrait-badge--blood-shared-life {
  border-color: var(--bt-red);
}

.player-portrait-badge--fighter-hundred-dragon-source,
.player-portrait-badge--blood-shared-life-source {
  box-shadow: 0 0 0 1px var(--bt-bg-0), 0 0 8px var(--bt-red-soft);
}

.player-portrait-badge--fighter-hundred-dragon-bound,
.player-portrait-badge--blood-shared-life-bound {
  border-style: dashed;
}

/* ===== 名牌上沿芯片 ===== */
.seat-chip-row {
  position: absolute;
  left: 4px;
  right: 4px;
  bottom: 3px;
  z-index: 5;
  display: flex;
  flex-wrap: wrap-reverse;
  justify-content: center;
  align-items: center;
  gap: 2px 3px;
  max-height: 34px;
  overflow: hidden;
  pointer-events: auto;
}

.seat-chip {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 1px 4px;
  border-radius: 3px;
  border: 1px solid var(--bt-hairline-soft);
  background: var(--bt-panel);
  color: var(--bt-text-sub);
  font-size: 9px;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
}

.seat-chip-value {
  color: var(--bt-text);
  font-family: var(--bt-font-num);
}

.form-badge {
  border-color: var(--bt-violet);
  color: var(--bt-text);
}

.player-overlay-resource,
.player-overlay-tokens {
  display: contents;
}

.seat-chip--gem,
.seat-chip--crystal {
  color: var(--bt-text);
  font-family: var(--bt-font-num);
}

.seat-chip-gem-mark,
.seat-chip-crystal-mark {
  width: 6px;
  height: 6px;
  margin-right: 1px;
  transform: rotate(45deg);
  box-shadow: 0 0 0 1px var(--bt-bg-0);
}

.seat-chip-gem-mark {
  background: var(--bt-red);
}

.seat-chip-crystal-mark {
  background: var(--bt-blue);
}

/* ===== 完整模式状态图标列：内侧圆形镀铬徽章 ===== */
.player-overlay-effects {
  position: absolute;
  top: 28px;
  right: 5px;
  z-index: 5;
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
}

.effect-icon-item {
  width: 22px;
  height: 22px;
  flex-shrink: 0;
  padding: 2px;
  border-radius: 999px;
  border: 1px solid var(--bt-chrome);
  background: radial-gradient(circle at 50% 35%, var(--bt-panel-raised), var(--bt-bg-0));
  box-shadow: 0 0 0 1px var(--bt-bg-0);
}

.player-area--compact .player-overlay-resource,
.player-area--compact .player-overlay-effects,
.player-area--compact .player-overlay-tokens {
  display: none;
}

/* ===== 名牌条 + 底栏 ===== */
.player-overlay {
  position: relative;
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  z-index: 4;
  background: var(--bt-bg-0);
}

.player-overlay-title-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-height: 16px;
  padding: 1px 6px;
  border-top: 1px solid var(--bt-hairline);
  border-bottom: 1px solid var(--bt-hairline-soft);
}

.camp-blue .player-overlay-title-row {
  background: linear-gradient(180deg, var(--bt-blue), var(--bt-blue-soft));
}

.camp-red .player-overlay-title-row {
  background: linear-gradient(180deg, var(--bt-red), var(--bt-red-soft));
}

.player-overlay-role {
  min-width: 0;
  font-size: 11px;
  line-height: 1.2;
  font-weight: 700;
  letter-spacing: 0.06em;
  color: var(--bt-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-shadow: 0 1px 2px var(--bt-bg-0);
}

.player-overlay-player {
  display: none;
}

.btn-show-skills {
  flex-shrink: 0;
  padding: 0 4px;
  border-radius: 3px;
  border: 1px solid var(--bt-hairline);
  background: var(--bt-panel);
  color: var(--bt-chrome-hi);
  font-size: 9px;
  line-height: 1.4;
  transition: border-color 0.2s ease, color 0.2s ease;
}

.btn-show-skills:hover,
.btn-show-skills:focus-visible {
  border-color: var(--bt-active);
  color: var(--bt-active);
}

.btn-show-skills:focus-visible {
  outline: 1px solid var(--bt-active);
  outline-offset: 1px;
}

.player-overlay-stats {
  display: flex;
  align-items: center;
  justify-content: space-evenly;
  width: 100%;
  min-height: 18px;
  padding: 2px 6px;
  font-size: 10px;
  color: var(--bt-text);
}

.player-overlay-stat {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
  min-width: 0;
}

.seat-stat-divider {
  width: 1px;
  height: 11px;
  background: var(--bt-hairline);
}

.seat-stat-icon {
  width: 11px;
  height: 11px;
  flex-shrink: 0;
  fill: var(--bt-chrome);
  stroke: var(--bt-bg-0);
  stroke-width: 0.6;
}

.seat-stat-num {
  font-family: var(--bt-font-num);
  font-size: 11px;
  line-height: 1;
  font-variant-numeric: tabular-nums;
}

.seat-heal-pip {
  width: 10px;
  height: 10px;
  flex-shrink: 0;
  fill: none;
  stroke: var(--bt-chrome-mid);
  stroke-width: 1.4;
}

.seat-heal-pip--filled {
  fill: var(--bt-heal);
  stroke: var(--bt-heal);
}

.player-overlay-stat--over {
  color: var(--bt-danger);
}

.player-overlay-stat--over .seat-stat-icon {
  fill: var(--bt-danger);
}

.seat-stealth-hint {
  margin: 0 4px 3px;
  padding: 1px 4px;
  border-radius: 3px;
  border: 1px solid var(--bt-hairline-soft);
  color: var(--bt-text-sub);
  font-size: 9px;
  text-align: center;
}

/* ===== 伤害暴血 ===== */
.damage-burst-overlay {
  border-radius: inherit;
}
.damage-burst-inner {
  position: relative;
  animation: burstPop 0.54s ease-out forwards;
}
.damage-number {
  position: relative;
  z-index: 2;
  font-size: 1.4rem;
  font-weight: 900;
  font-family: var(--bt-font-num);
  color: var(--bt-danger);
  text-shadow: 0 0 8px var(--bt-red-soft), 0 1px 2px var(--bt-bg-0);
  animation: numberPop 0.44s ease-out;
}
.damage-splash {
  position: absolute;
  inset: -16px;
  background: radial-gradient(circle, var(--bt-red-soft) 0%, transparent 72%);
  border-radius: 50%;
  animation: splashExpand 0.54s ease-out forwards;
  transform: scale(0);
}
@keyframes burstPop {
  0% { transform: scale(0.72); opacity: 0; }
  35% { transform: scale(1.07); opacity: 1; }
  100% { transform: scale(1); opacity: 0.74; }
}
@keyframes numberPop {
  0% { transform: scale(0.5); opacity: 0; }
  55% { transform: scale(1.1); opacity: 1; }
  100% { transform: scale(1); opacity: 1; }
}
@keyframes splashExpand {
  0% { transform: scale(0.2); opacity: 0.7; }
  100% { transform: scale(1.2); opacity: 0; }
}

@keyframes effectSealRotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ===== 尺寸断点（保持原有高度） ===== */
@media (min-width: 1800px) {
  .player-area--compact {
    height: 150px;
  }

  .player-area--full {
    height: 276px;
  }
}

@media (max-width: 900px) {
  .player-area--compact {
    height: 128px;
  }

  .player-area--full {
    height: 236px;
  }

  .player-overlay-role {
    font-size: 10px;
  }

  .player-overlay-stats {
    font-size: 9px;
  }
}

@media (max-width: 640px) {
  .player-area--compact {
    height: 118px;
  }

  .player-area--full {
    height: 224px;
  }

  .player-overlay-role {
    font-size: 9px;
  }

  .seat-stat-num {
    font-size: 10px;
  }

  .seat-heal-pip {
    width: 8px;
    height: 8px;
  }

  .effect-icon-item {
    width: 20px;
    height: 20px;
  }

  .player-portrait-badges {
    max-width: calc(100% - 50px);
  }

  .player-portrait-badge {
    font-size: 8px;
    min-height: 14px;
    padding: 1px 5px;
  }

  .player-portrait-badge-icon {
    width: 10px;
    height: 10px;
  }

  .btn-show-skills {
    font-size: 8px;
    padding: 0 3px;
  }
}
</style>
