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
import BattleMarkIcon from './marks/BattleMarkIcon.vue'
import CoverCard from './card/CoverCard.vue'
import { seatCoverMarks, seatTokenMarks } from '../constants/seatMarks'
import { seatPersistentEffects } from '../constants/persistentEffects'
import { characterArtSrc, characterFullCardSrc } from '../constants/characterArt'

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
  /** 盖牌浮层开 / 关：外层据此暂停角色立绘预览，避免两者重叠。 */
  'cover-preview': [open: boolean]
}>()

const charInfo = computed(() => (props.player.role ? getCharacter(props.player.role) : null))
const roleDisplayName = computed(() => getRoleDisplayName(props.player.role))
const playerDisplayName = computed(() => props.player.name || props.player.id)

// 座位用裁掉原卡框的插画；加载失败时先回退到完整卡图，再失败才显示占位
const useFullCardFallback = ref(false)
const characterImageSrc = computed(() => (useFullCardFallback.value ? characterFullCardSrc(props.player.role) : characterArtSrc(props.player.role)))

const showCharacterImage = ref(true)
watch(() => props.player.role, () => {
  showCharacterImage.value = true
  useFullCardFallback.value = false
})
function onCharImageError() {
  if (!useFullCardFallback.value) {
    useFullCardFallback.value = true
    return
  }
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


const tokenIndicators = computed(() => seatTokenMarks(props.player))

/** 盖牌：按种类合并为"牌背 + 数量"，对所有人可见（内容已由服务端遮蔽）。 */
const coverMarks = computed(() => seatCoverMarks(props.player))

/** 悬停某种盖牌标记时，在座位旁浮出该种盖牌（他人视角为牌背，自己视角为牌面）。 */
const hoveredCoverKey = ref<string | null>(null)
const coverPopoverStyle = ref<Record<string, string>>({})

function showCoverPopover(key: string, event: Event) {
  const el = event.currentTarget as HTMLElement | null
  if (el) {
    const rect = el.getBoundingClientRect()
    // 左半屏的座位往右弹，右半屏的往左弹，都与标记垂直居中
    const toRight = rect.left + rect.width / 2 < window.innerWidth / 2
    coverPopoverStyle.value = {
      top: `${rect.top + rect.height / 2}px`,
      ...(toRight ? { left: `${rect.right + 10}px` } : { right: `${window.innerWidth - rect.left + 10}px` }),
    }
  }
  hoveredCoverKey.value = key
  emit('cover-preview', true)
}

function hideCoverPopover() {
  hoveredCoverKey.value = null
  emit('cover-preview', false)
}
const hoveredCoverCards = computed(() => {
  const key = hoveredCoverKey.value
  if (!key) return []
  return (props.player.field ?? [])
    .map((fc, fieldIndex) => ({ fc, fieldIndex }))
    .filter(({ fc }) => fc?.mode === 'Cover' && fc.effect === key && !!fc.card)
})
const hoveredCoverLabel = computed(() => coverMarks.value.find((c) => c.key === hoveredCoverKey.value)?.label ?? '')

const allIndicators = computed(() => tokenIndicators.value)
/** 持续中的公开效果（凋零、额外回合、下击加伤……）：比普通指示物醒目，悬停看规则。 */
const persistentEffects = computed(() => seatPersistentEffects(props.player))

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
        v-if="formIndicator || (player.gem || 0) + (player.crystal || 0) > 0 || allIndicators.length || coverMarks.length || persistentEffects.length"
        class="seat-chip-row"
      >
        <span v-if="formIndicator" class="seat-chip form-badge" :title="formIndicator.label">
          {{ formIndicator.label }}
        </span>
        <span
          v-for="effect in persistentEffects"
          :key="effect.key"
          class="seat-chip seat-chip--effect"
          :title="effect.hint"
          :data-effect-key="effect.key"
          data-testid="seat-persistent-effect"
        >
          <span>{{ effect.label }}</span>
          <span v-if="effect.valueText" class="seat-chip-value">{{ effect.valueText }}</span>
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
        <span v-if="coverMarks.length" class="player-overlay-tokens">
          <span
            v-for="cover in coverMarks"
            :key="`cover-${cover.key}`"
            class="seat-mark seat-mark--cover"
            :title="`${cover.label}: ${cover.count}`"
            :data-mark-key="cover.key"
            tabindex="0"
            @mouseenter="showCoverPopover(cover.key, $event)"
            @mouseleave="hideCoverPopover"
            @focus="showCoverPopover(cover.key, $event)"
            @blur="hideCoverPopover"
          >
            <span class="seat-mark-icon seat-mark-icon--cover"><BattleMarkIcon kind="cover" :mark-key="cover.key" /></span>
            <span class="seat-chip-value">{{ cover.count }}</span>
          </span>
        </span>
        <span v-if="allIndicators.length" class="player-overlay-tokens">
          <template v-for="token in allIndicators" :key="token.key">
            <span
              v-if="token.mark"
              class="seat-mark seat-mark--token"
              :title="`${token.label}: ${token.value}`"
              :data-mark-key="token.key"
            >
              <span class="seat-mark-icon"><BattleMarkIcon kind="token" :mark-key="token.key" /></span>
              <span v-if="token.cap" class="seat-mark-pips" aria-hidden="true">
                <span v-for="n in token.cap" :key="n" class="seat-mark-pip" :class="{ 'seat-mark-pip--on': n <= token.value }" />
              </span>
              <span v-else class="seat-chip-value">{{ token.value }}</span>
            </span>
            <span v-else class="seat-chip seat-chip--token" :title="`${token.label}: ${token.value}`">
              <span>{{ token.label }}</span>
              <span class="seat-chip-value">×{{ token.value }}</span>
            </span>
          </template>
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
    <!-- 悬停盖牌标记：浮出该种盖牌。他人视角牌面已由服务端遮蔽，画种类牌背 -->
    <Teleport to="body">
      <div
        v-if="hoveredCoverKey && hoveredCoverCards.length"
        class="seat-cover-popover"
        :style="coverPopoverStyle"
        role="tooltip"
      >
        <div class="seat-cover-popover__head">
          {{ hoveredCoverLabel }}<span class="seat-cover-popover__count">×{{ hoveredCoverCards.length }}</span>
          <span class="seat-cover-popover__sub">{{ isMe ? '牌面仅自己可见' : '盖牌（牌面不可见）' }}</span>
        </div>
        <div class="seat-cover-popover__cards">
          <CoverCard
            v-for="entry in hoveredCoverCards"
            :key="`seat-cover-${entry.fieldIndex}`"
            :card="entry.fc.card"
            :effect="entry.fc.effect"
            :index="entry.fieldIndex"
            :face-up="!!isMe"
          />
        </div>
      </div>
    </Teleport>
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
  object-position: center 22%;
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

/* 持续效果：金色描边 + 缓慢呼吸光，区别于静态计数指示物。 */
.seat-chip--effect {
  color: #fff3d0;
  border-color: rgba(227, 196, 106, 0.7);
  background: linear-gradient(180deg, rgba(92, 70, 24, 0.85), rgba(40, 30, 10, 0.9));
  animation: seat-effect-glow 2.4s ease-in-out infinite;
}
@keyframes seat-effect-glow {
  0%, 100% { box-shadow: 0 0 0 rgba(227, 196, 106, 0); }
  50% { box-shadow: 0 0 8px rgba(227, 196, 106, 0.55); }
}
@media (prefers-reduced-motion: reduce) {
  .seat-chip--effect { animation: none; }
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

/* 盖牌 / 指示物：图标 + 数量（或圆点格），图形与配色来自 BattleMarkIcon */
.seat-mark {
  display: inline-flex;
  align-items: center;
  gap: 1px;
  padding: 0 3px 0 1px;
  border-radius: 999px;
  background: var(--bt-panel);
  box-shadow: 0 0 0 1px var(--bt-hairline-soft);
  font-size: 9px;
  font-weight: 700;
  line-height: 1;
  white-space: nowrap;
}
.seat-mark--cover {
  cursor: default;
}

.seat-mark--cover:focus-visible {
  outline: 2px solid var(--bt-select);
  outline-offset: 1px;
}

.seat-cover-popover {
  position: fixed;
  z-index: 80;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 16px 12px;
  transform: translateY(-50%);
  pointer-events: none;
  isolation: isolate;
  animation: seatCoverIn 0.18s ease-out;
}

/* 羽化暗影衬底，与扩展区一致，不画面板框 */
.seat-cover-popover::before {
  content: '';
  position: absolute;
  inset: -14px -24px;
  z-index: -1;
  background: radial-gradient(closest-side, rgba(3, 7, 16, 0.92), rgba(3, 7, 16, 0.72) 60%, transparent);
  filter: blur(4px);
}

@keyframes seatCoverIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .seat-cover-popover {
    animation: none;
  }
}

.seat-cover-popover__head {
  display: flex;
  align-items: baseline;
  gap: 8px;
  color: #eef3fb;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.1em;
  white-space: nowrap;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.9);
}

.seat-cover-popover__count {
  margin-left: -4px;
  color: rgba(210, 225, 245, 0.85);
  font-variant-numeric: tabular-nums;
}

.seat-cover-popover__sub {
  color: rgba(170, 192, 222, 0.7);
  font-size: 10px;
  font-weight: 400;
  letter-spacing: 0.04em;
}

.seat-cover-popover__cards {
  display: flex;
  gap: 8px;
  padding-bottom: 6px;
}

.seat-mark-icon {
  display: inline-flex;
  width: 15px;
  height: 15px;
  flex: none;
}
.seat-mark-icon--cover {
  width: 11px;
  height: 15px;
}
.seat-mark-pips {
  display: inline-flex;
  gap: 1.5px;
  margin-left: 1px;
}
.seat-mark-pip {
  width: 4px;
  height: 4px;
  border-radius: 999px;
  box-shadow: inset 0 0 0 1px var(--bt-chrome-mid);
}
.seat-mark-pip--on {
  background: var(--bt-chrome-hi);
  box-shadow: 0 0 3px var(--bt-select-glow);
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

/* 紧凑座位：宝石水晶与状态图标由 GameBoard 状态栏负责；盖牌/指示物图标足够小，留在座位上 */
.player-area--compact .player-overlay-resource,
.player-area--compact .player-overlay-effects {
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
