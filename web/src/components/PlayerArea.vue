<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { PlayerView } from '../types/game'
import { useGameStore } from '../stores/gameStore'

const store = useGameStore()
const props = defineProps<{
  player: PlayerView
  isMe?: boolean
  isOpponent?: boolean
  selectable?: boolean
  selected?: boolean
  compact?: boolean
  turnOrder?: number
}>()

const emit = defineEmits<{
  select: [playerId: string]
}>()

const charInfo = computed(() => (props.player.role ? store.getCharacter(props.player.role) : null))
const roleDisplayName = computed(() => store.getRoleDisplayName(props.player.role))
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
  if (!props.player.role) return
  if (store.skillModalCharacterId === props.player.role) {
    store.openSkillModal(null)
    return
  }
  const trigger = event?.currentTarget as HTMLElement | null
  if (trigger) {
    const rect = trigger.getBoundingClientRect()
    store.openSkillModal(props.player.role, {
      x: rect.left,
      y: rect.top,
      width: rect.width,
      height: rect.height,
    })
    return
  }
  store.openSkillModal(props.player.role)
}

const campClass = computed(() => {
  return props.player.camp === 'Red' ? 'camp-red' : 'camp-blue'
})

const handCount = computed(() => {
  return props.isMe ? (props.player.hand?.length ?? props.player.hand_count) : props.player.hand_count
})

const isActive = computed(() => !!props.player.is_active)

const EFFECT_DISPLAY: Record<string, { icon: string; label: string; cls: string }> = {
  Shield: { icon: '🛡️', label: '圣盾', cls: 'bg-yellow-800/60' },
  Poison: { icon: '☠️', label: '中毒', cls: 'bg-green-800/60' },
  Weak: { icon: '💫', label: '虚弱', cls: 'bg-purple-800/60' },
  SealFire: { icon: '🔥', label: '火封印', cls: 'bg-red-800/60' },
  SealWater: { icon: '💧', label: '水封印', cls: 'bg-blue-800/60' },
  SealEarth: { icon: '🪨', label: '地封印', cls: 'bg-amber-800/60' },
  SealWind: { icon: '🌪️', label: '风封印', cls: 'bg-teal-800/60' },
  SealThunder: { icon: '⚡', label: '雷封印', cls: 'bg-indigo-800/60' },
  FiveElementsBind: { icon: '⛓️', label: '五系束缚', cls: 'bg-gray-700/80' },
  RoseCourtyard: { icon: '🌹', label: '血蔷薇庭院', cls: 'bg-rose-900/75' },
  PowerBlessing: { icon: '✨', label: '威力赐福', cls: 'bg-orange-900/75' },
  SwiftBlessing: { icon: '🪽', label: '迅捷赐福', cls: 'bg-cyan-900/75' },
  Stealth: { icon: '👤', label: '潜行', cls: 'bg-gray-700/80' },
  BardEternalMovement: { icon: '🎼', label: '永恒乐章', cls: 'bg-violet-900/75' },
  HeroTaunt: { icon: '⚔️', label: '挑衅', cls: 'bg-red-900/75' },
  SoulLink: { icon: '🧿', label: '灵魂链接', cls: 'bg-indigo-900/75' },
  BloodSharedLife: { icon: '🩸', label: '同生共死', cls: 'bg-rose-900/75' },
}

const fieldEffects = computed(() => {
  if (!props.player.field?.length) return []
  return props.player.field
    .filter(fc => fc.mode === 'Effect' && fc.effect)
    .map(fc => EFFECT_DISPLAY[fc.effect] || { icon: '✦', label: fc.effect, cls: 'bg-gray-700/60' })
})

// 当前玩家身上的伤害特效（暴血）
const myDamageEffects = computed(() =>
  store.damageEffects.filter(d => d.targetId === props.player.id)
)

const TOKEN_DISPLAY: Record<string, { label: string; cls: string }> = {
  element: { label: '元素', cls: 'bg-cyan-800/70 text-cyan-100 border-cyan-500/40' },
  judgment: { label: '审判', cls: 'bg-fuchsia-800/70 text-fuchsia-100 border-fuchsia-500/40' },
  valkyrie_spirit: { label: '英灵', cls: 'bg-amber-800/70 text-amber-100 border-amber-500/40' },
  arbiter_form: { label: '审判形态', cls: 'bg-violet-800/70 text-violet-100 border-violet-500/40' },
  elf_ritual_form: { label: '祝福形态', cls: 'bg-emerald-800/70 text-emerald-100 border-emerald-500/40' },
  elf_blessing_count: { label: '祝福', cls: 'bg-teal-800/70 text-teal-100 border-teal-500/40' },
  ms_shadow_form: { label: '暗影形态', cls: 'bg-slate-800/80 text-slate-100 border-slate-500/40' },
  css_blood: { label: '鲜血', cls: 'bg-rose-800/70 text-rose-100 border-rose-500/40' },
  css_blood_cap: { label: '鲜血上限', cls: 'bg-rose-900/60 text-rose-100 border-rose-600/40' },
  css_rose_courtyard_active: { label: '庭院在场', cls: 'bg-red-900/65 text-red-100 border-red-500/40' },
  prayer_form: { label: '祈祷形态', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  prayer_rune: { label: '祈祷符文', cls: 'bg-yellow-900/70 text-yellow-100 border-yellow-500/40' },
  crk_blood_mark: { label: '血印', cls: 'bg-red-900/70 text-red-100 border-red-500/40' },
  crk_hot_form: { label: '热血形态', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40' },
  hom_war_rune: { label: '战纹', cls: 'bg-indigo-900/70 text-indigo-100 border-indigo-500/40' },
  hom_magic_rune: { label: '魔纹', cls: 'bg-violet-900/70 text-violet-100 border-violet-500/40' },
  hom_burst_form: { label: '蓄势形态', cls: 'bg-purple-900/70 text-purple-100 border-purple-500/40' },
  onmyoji_form: { label: '式神形态', cls: 'bg-sky-900/70 text-sky-100 border-sky-500/40' },
  onmyoji_ghost_fire: { label: '鬼火', cls: 'bg-teal-900/70 text-teal-100 border-teal-500/40' },
  bw_rebirth: { label: '重生', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  bw_flame_form: { label: '烈焰形态', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40' },
  mb_charge_count: { label: '充能', cls: 'bg-indigo-900/70 text-indigo-100 border-indigo-500/40' },
  bd_inspiration: { label: '灵感', cls: 'bg-violet-900/70 text-violet-100 border-violet-500/40' },
  bd_prisoner_form: { label: '囚徒形态', cls: 'bg-purple-900/70 text-purple-100 border-purple-500/40' },
  ml_phantom_form: { label: '幻影形态', cls: 'bg-slate-900/75 text-slate-100 border-slate-500/40' },
  ml_dark_release_next_attack_bonus: { label: '下次主动攻+伤', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  ml_fullness_next_attack_bonus: { label: '充盈下次攻+伤', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40' },
  ml_dark_release_lock_turn: { label: '本回合锁技能', cls: 'bg-zinc-800/75 text-zinc-100 border-zinc-500/40' },
  hero_anger: { label: '怒气', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  hero_wisdom: { label: '知性', cls: 'bg-sky-900/70 text-sky-100 border-sky-500/40' },
  hero_exhaustion_form: { label: '精疲力竭', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  hero_calm_end_crystal_pending: { label: '止水回晶', cls: 'bg-cyan-900/70 text-cyan-100 border-cyan-500/40' },
  fighter_qi: { label: '斗气', cls: 'bg-orange-900/70 text-orange-100 border-orange-500/40' },
  fighter_hundred_dragon_form: { label: '幻龙拳形态', cls: 'bg-red-900/70 text-red-100 border-red-500/40' },
  hb_cannon: { label: '圣煌辉光炮', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  hb_faith: { label: '信仰', cls: 'bg-yellow-900/70 text-yellow-100 border-yellow-500/40' },
  hb_form: { label: '圣煌形态', cls: 'bg-sky-900/70 text-sky-100 border-sky-500/40' },
  ss_blue_soul: { label: '蓝色灵魂', cls: 'bg-blue-900/70 text-blue-100 border-blue-500/40' },
  ss_yellow_soul: { label: '黄色灵魂', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  ss_link_active: { label: '灵魂链接在场', cls: 'bg-indigo-900/70 text-indigo-100 border-indigo-500/40' },
  mg_dark_form: { label: '暗月形态', cls: 'bg-slate-900/75 text-slate-100 border-slate-500/40' },
  mg_new_moon: { label: '新月', cls: 'bg-cyan-900/70 text-cyan-100 border-cyan-500/40' },
  mg_petrify: { label: '石化', cls: 'bg-zinc-900/70 text-zinc-100 border-zinc-500/40' },
  mg_dark_moon_count: { label: '暗月', cls: 'bg-indigo-900/70 text-indigo-100 border-indigo-500/40' },
  mg_next_attack_no_counter: { label: '下次攻不可应战', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  bp_bleed_form: { label: '流血形态', cls: 'bg-red-900/70 text-red-100 border-red-500/40' },
  bp_shared_life_active: { label: '同生共死在场', cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40' },
  bt_pupa: { label: '蛹', cls: 'bg-amber-900/70 text-amber-100 border-amber-500/40' },
  bt_cocoon_count: { label: '茧', cls: 'bg-indigo-900/70 text-indigo-100 border-indigo-500/40' },
  bt_wither_active: { label: '凋零生效', cls: 'bg-red-900/70 text-red-100 border-red-500/40' },
}

const HIDDEN_TOKEN_KEYS = new Set([
  'arbiter_law_inited',
  'arbiter_skip_forced_doomsday',
  'arbiter_forced_doomsday_done_turn',
  'holy_lancer_block_sacred_strike',
  'holy_lancer_prayer_used_turn',
  // 新角色内部流程标记：不应直接展示给玩家
  'elf_elemental_shot_fire_pending',
  'elf_elemental_shot_water_pending',
  'elf_elemental_shot_earth_pending',
  'elf_elemental_shot_thunder_pending',
  'elf_ritual_release_waiting',
  'elf_ritual_suppress_overflow',
  'plague_block_immortal',
  'ms_shadow_release_pending',
  'ms_yellow_spring_pending',
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
  'hero_exhaustion_release_pending',
  'hero_roar_active',
  'hero_roar_damage_pending',
  'hero_calm_force_no_counter',
  'hero_dead_duel_pending',
  'fighter_attack_start_skill_lock',
  'fighter_charge_pending',
  'fighter_charge_damage_pending',
  'fighter_qiburst_force_no_counter',
  'fighter_hundred_dragon_target_order',
  'hb_special_used_turn',
  'hb_auto_fill_done_turn',
  'hb_shard_miss_pending',
  'mg_blasphemy_used_turn',
  'mg_blasphemy_pending',
  'mg_extra_turn_pending',
  'bp_bleed_tick_done_turn',
  'bt_wither_pending',
  'adventurer_extract_last_gem',
  'adventurer_extract_last_crystal',
])

const tokenIndicators = computed(() => {
  const entries = Object.entries(props.player.tokens ?? {})
    .filter(([key, value]) => !HIDDEN_TOKEN_KEYS.has(key) && typeof value === 'number' && value > 0)
    .filter(([key, value]) => !(key === 'css_blood_cap' && value <= 3))
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

const isFighter = computed(() => props.player.role === 'fighter')
const orderedPlayerIds = computed(() => {
  const ids: string[] = []
  const seen = new Set<string>()
  for (const p of store.roomPlayers) {
    if (store.players[p.id] && !seen.has(p.id)) {
      ids.push(p.id)
      seen.add(p.id)
    }
  }
  for (const id of Object.keys(store.players).sort()) {
    if (!seen.has(id)) {
      ids.push(id)
      seen.add(id)
    }
  }
  return ids
})
const playerNameByTurnOrder = computed<Record<number, string>>(() => {
  const map: Record<number, string> = {}
  orderedPlayerIds.value.forEach((id, idx) => {
    const p = store.players[id]
    map[idx + 1] = p?.name || id
  })
  return map
})
const playerByTurnOrder = computed<Record<number, PlayerView | undefined>>(() => {
  const map: Record<number, PlayerView | undefined> = {}
  orderedPlayerIds.value.forEach((id, idx) => {
    map[idx + 1] = store.players[id]
  })
  return map
})

const fighterStatusLines = computed(() => {
  if (!isFighter.value) return []
  const t = props.player.tokens ?? {}
  const lines: string[] = []
  const qi = t.fighter_qi ?? 0
  const inForm = (t.fighter_hundred_dragon_form ?? 0) > 0
  const targetOrder = t.fighter_hundred_dragon_target_order ?? 0
  const chargePending = (t.fighter_charge_pending ?? 0) > 0
  const burstNoCounter = (t.fighter_qiburst_force_no_counter ?? 0) > 0

  lines.push(`斗气：${qi}/8`)
  if (inForm) {
    lines.push('当前：百式幻龙拳形态（主动攻击+2，应战攻击+1）')
    lines.push('本阶段限制：仅可执行攻击；不能执行法术/特殊行动；不能发动【蓄力一击】')
    if (targetOrder > 0) {
      const targetPlayer = playerByTurnOrder.value[targetOrder]
      if (targetPlayer && targetPlayer.camp !== props.player.camp) {
        const targetName = playerNameByTurnOrder.value[targetOrder] || targetPlayer.id
        lines.push(`目标锁定：${targetName}（顺序 #${targetOrder}，主动攻击需保持同一目标）`)
      } else {
        lines.push('目标锁定：无（仅可锁定敌方目标）')
      }
    } else {
      lines.push('目标锁定：将在本回合首次主动攻击后确定')
    }
  } else {
    lines.push('当前：普通形态')
  }
  if (chargePending) {
    lines.push('蓄力一击生效中：若本次攻击未命中，将按当前斗气对自己造成法术伤害')
  }
  if (burstNoCounter) {
    lines.push('气绝崩击生效中：本次攻击无法应战')
  }
  return lines
})

const fighterSkillBriefs = computed(() => {
  if (!isFighter.value) return []
  return [
    { title: '念气力场', desc: '被动：所有对你造成的单次伤害最高为4点。' },
    { title: '蓄力一击', desc: '响应（主动攻击前）：斗气+1，本次攻击伤害+1；若未命中，按当前斗气对自己造成法术伤害。' },
    { title: '念弹', desc: '响应（法术行动后）：斗气+1并对目标敌方造成1点法术伤害；若其治疗为0，你按当前斗气自伤。' },
    { title: '百式幻龙拳', desc: '启动：移除3斗气进入持续形态；主动攻+2、应战攻+1；本阶段仅可攻击且锁定同一目标。' },
    { title: '气绝崩击', desc: '响应（主动攻击前）：移除1斗气，本次攻击无法应战，再按当前斗气对自己造成法术伤害。' },
    { title: '斗神天驱', desc: '启动（消耗1水晶）：弃到3张牌并+2治疗。' }
  ]
})

const isHolyBow = computed(() => props.player.role === 'holy_bow')

const holyBowStatusLines = computed(() => {
  if (!isHolyBow.value) return []
  const t = props.player.tokens ?? {}
  const lines: string[] = []
  const faith = t.hb_faith ?? 0
  const cannon = t.hb_cannon ?? 0
  const inForm = (t.hb_form ?? 0) > 0
  const usedSpecial = (t.hb_special_used_turn ?? 0) > 0

  lines.push(`信仰：${faith}/10`)
  lines.push(`圣煌辉光炮：${cannon}/1`)
  if (inForm) {
    lines.push('当前：圣煌形态（可用圣光爆裂/流星圣弹/圣煌辉光炮）')
    lines.push('形态规则：执行特殊行动会立即脱离并+1治疗')
  } else {
    lines.push('当前：普通形态')
  }
  if (usedSpecial) {
    lines.push('本回合已执行特殊行动（回合结束不触发自动填充）')
  } else {
    lines.push('本回合未执行特殊行动（回合结束可触发自动填充）')
  }
  return lines
})

const holyBowSkillBriefs = computed(() => {
  if (!isHolyBow.value) return []
  return [
    { title: '天之弓', desc: '被动：初始+1辉光炮、+2水晶、治疗上限+1；非圣命格主动攻击伤害-1；圣命格主动攻击命中+1信仰。' },
    { title: '圣屑飓暴', desc: '法术：弃2张同系攻击牌，视为同系圣命格主动攻击；未命中可移除治疗并令队友弃牌。' },
    { title: '圣煌降临', desc: '法术：移除2治疗或2信仰，进入圣煌形态并额外+1法术行动。' },
    { title: '圣光爆裂', desc: '法术（仅圣煌）：分支①摸1并增益治疗/信仰；分支②移除X治疗并弃X牌后，对最多X名对手造成攻击伤害。' },
    { title: '流星圣弹', desc: '响应（仅圣煌）：主动攻击前移除1治疗或1信仰，使1名我方角色+1治疗。' },
    { title: '圣煌辉光炮', desc: '法术（仅圣煌）：消耗辉光炮与信仰，全员手牌调至4、我方星杯+1，并选择士气对齐方向。' },
    { title: '自动填充', desc: '被动：回合结束且未执行特殊行动时，可消耗资源获得信仰/治疗增益。' }
  ]
})

const isSoulSorcerer = computed(() => props.player.role === 'soul_sorcerer')

const soulSorcererStatusLines = computed(() => {
  if (!isSoulSorcerer.value) return []
  const t = props.player.tokens ?? {}
  const lines: string[] = []
  const blue = t.ss_blue_soul ?? 0
  const yellow = t.ss_yellow_soul ?? 0
  const linkActive = (t.ss_link_active ?? 0) > 0

  lines.push(`蓝色灵魂：${blue}/6`)
  lines.push(`黄色灵魂：${yellow}/6`)
  if (linkActive) {
    lines.push('灵魂链接：已放置（承伤前可移除蓝魂转移伤害）')
  } else {
    lines.push('灵魂链接：未放置')
  }
  return lines
})

const soulSorcererSkillBriefs = computed(() => {
  if (!isSoulSorcerer.value) return []
  return [
    { title: '灵魂吞噬', desc: '被动：我方每下降1点士气，你+1黄色灵魂。' },
    { title: '灵魂召还', desc: '法术：弃X张法术牌，+X蓝色灵魂。' },
    { title: '灵魂转换', desc: '响应（每次主动攻击前）：可将1点蓝/黄灵魂互转。' },
    { title: '灵魂镜像', desc: '法术：移除2黄魂并弃2张牌，目标摸2张（最多补至手牌上限）。' },
    { title: '灵魂链接', desc: '启动：移除1黄+1蓝并放置链接；你或链接队友承伤前可移除X蓝魂转移X点伤害给另一方（转移后为法术伤害）。' }
  ]
})

const showStealthBlockedHint = computed(() => {
  if (props.selectable) return false
  if (!props.isOpponent) return false
  if (store.actionMode !== 'attack') return false
  if (store.selectedCardForAction === null) return false
  return !!props.player.field?.some((fc) => fc.mode === 'Effect' && fc.effect === 'Stealth')
})

function handleClick(e: MouseEvent) {
  if (props.selectable) {
    // 点击技能按钮不触发选目标
    if ((e.target as HTMLElement).closest('.btn-show-skills')) return
    emit('select', props.player.id)
  }
}
</script>

<template>
  <div
    class="player-area border transition-all duration-300 rounded-xl overflow-hidden relative"
    :class="[
      compact
        ? 'player-area--compact min-w-[108px] max-w-[124px] sm:min-w-[120px] sm:max-w-[140px] 2xl:min-w-[142px] 2xl:max-w-[166px]'
        : 'player-area--full min-w-[124px] sm:min-w-[140px] 2xl:min-w-[166px]',
      campClass,
      isActive ? 'player-area--active' : '',
      showStealthBlockedHint ? 'opacity-60 grayscale saturate-75' : '',
      selectable ? 'cursor-pointer hover:scale-[1.03] hover:ring-2 hover:ring-yellow-400 hover:shadow-lg hover:shadow-yellow-500/20' : '',
      selected ? 'player-area--selected' : ''
    ]"
    @click="handleClick"
  >
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

    <div v-if="typeof turnOrder === 'number'" class="turn-order-badge" :title="`行动顺序 #${turnOrder}`">
      #{{ turnOrder }}
    </div>

    <div class="player-overlay">
      <div class="player-overlay-title-row">
        <span class="text-sm" aria-hidden="true">{{ player.camp === 'Red' ? '🔴' : '🔵' }}</span>
        <span class="player-overlay-role" :title="roleDisplayName">{{ roleDisplayName }}</span>
        <button
          v-if="charInfo?.skills?.length"
          type="button"
          class="btn-show-skills px-1 py-0.5 rounded text-[9px] bg-gray-700/80 hover:bg-gray-600 text-amber-400/90 hover:text-amber-300 shrink-0"
          @click.stop="openSkillModal($event)"
        >
          技能
        </button>
      </div>

      <div class="player-overlay-player" :title="`玩家：${playerDisplayName}`">
        玩家：{{ playerDisplayName }}
      </div>

      <div class="player-overlay-stats">
        <div class="player-overlay-stat">
          <span aria-hidden="true">💖</span>
          <span>{{ player.heal }}/{{ player.max_heal }}</span>
        </div>
        <div class="player-overlay-stat">
          <span aria-hidden="true">🃏</span>
          <span>{{ handCount }}</span>
        </div>
        <div v-if="(player.gem || 0) + (player.crystal || 0) > 0" class="player-overlay-resource">
          <span v-if="player.gem" class="flex items-center gap-0.5"><span aria-hidden="true" class="text-red-300">♦</span>{{ player.gem }}</span>
          <span v-if="player.crystal" class="flex items-center gap-0.5"><span aria-hidden="true" class="text-blue-300">🔷</span>{{ player.crystal }}</span>
        </div>
      </div>

      <div v-if="showStealthBlockedHint" class="text-[10px] text-gray-200/90 bg-black/35 border border-gray-400/30 rounded px-1 py-0.5 w-fit">
        潜行状态无法选中
      </div>

      <div v-if="fieldEffects.length" class="player-overlay-effects">
        <span
          v-for="(eff, i) in fieldEffects"
          :key="i"
          :title="eff.label"
          class="text-[10px] px-1 rounded"
          :class="eff.cls"
        >{{ eff.icon }}</span>
      </div>

      <div v-if="tokenIndicators.length" class="player-overlay-tokens">
        <span
          v-for="token in tokenIndicators"
          :key="token.key"
          class="inline-flex items-center gap-1 px-1 py-0.5 rounded border text-[9px] leading-none"
          :class="token.cls"
          :title="`${token.label}: ${token.value}`"
        >
          <span>{{ token.label }}</span>
          <span class="font-bold">×{{ token.value }}</span>
        </span>
      </div>

      <div v-if="fighterStatusLines.length" class="player-overlay-ml-status">
        <div
          v-for="(line, idx) in fighterStatusLines"
          :key="`fighter-status-${idx}`"
          class="player-overlay-ml-line"
        >
          {{ line }}
        </div>
      </div>

      <div v-if="fighterSkillBriefs.length" class="player-overlay-ml-skills">
        <div class="player-overlay-ml-title">格斗家状态说明</div>
        <div
          v-for="item in fighterSkillBriefs"
          :key="item.title"
          class="player-overlay-ml-skill-item"
        >
          <span class="player-overlay-ml-skill-name">{{ item.title }}：</span>
          <span class="player-overlay-ml-skill-desc">{{ item.desc }}</span>
        </div>
      </div>

      <div v-if="holyBowStatusLines.length" class="player-overlay-ml-status">
        <div
          v-for="(line, idx) in holyBowStatusLines"
          :key="`holy-bow-status-${idx}`"
          class="player-overlay-ml-line"
        >
          {{ line }}
        </div>
      </div>

      <div v-if="holyBowSkillBriefs.length" class="player-overlay-ml-skills">
        <div class="player-overlay-ml-title">圣弓状态说明</div>
        <div
          v-for="item in holyBowSkillBriefs"
          :key="item.title"
          class="player-overlay-ml-skill-item"
        >
          <span class="player-overlay-ml-skill-name">{{ item.title }}：</span>
          <span class="player-overlay-ml-skill-desc">{{ item.desc }}</span>
        </div>
      </div>

      <div v-if="soulSorcererStatusLines.length" class="player-overlay-ml-status">
        <div
          v-for="(line, idx) in soulSorcererStatusLines"
          :key="`soul-sorcerer-status-${idx}`"
          class="player-overlay-ml-line"
        >
          {{ line }}
        </div>
      </div>

      <div v-if="soulSorcererSkillBriefs.length" class="player-overlay-ml-skills">
        <div class="player-overlay-ml-title">灵魂术士状态说明</div>
        <div
          v-for="item in soulSorcererSkillBriefs"
          :key="item.title"
          class="player-overlay-ml-skill-item"
        >
          <span class="player-overlay-ml-skill-name">{{ item.title }}：</span>
          <span class="player-overlay-ml-skill-desc">{{ item.desc }}</span>
        </div>
      </div>
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
.player-area--selected {
  transform: scale(1.05);
  border-color: rgba(233, 196, 136, 0.9) !important;
  box-shadow:
    0 0 0 1px rgba(246, 224, 181, 0.7),
    0 0 16px rgba(216, 173, 107, 0.48),
    0 0 28px rgba(164, 124, 72, 0.32);
  animation: frameGlowPulse 1.8s ease-in-out infinite;
}

.player-area--selected::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  border: 1px solid rgba(248, 222, 174, 0.68);
  box-shadow: inset 0 0 12px rgba(216, 173, 107, 0.32);
  pointer-events: none;
}

.player-area {
  padding: 0;
  isolation: isolate;
  background: rgba(9, 18, 30, 0.74);
}

.player-area::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(120% 88% at 50% 10%, rgba(241, 248, 255, 0.13), rgba(15, 24, 36, 0.22) 56%, rgba(8, 13, 22, 0.4) 100%);
  pointer-events: none;
  z-index: 2;
}

.player-area--compact {
  height: 196px;
}

.player-area--full {
  height: 232px;
}

.character-portrait-fill {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 32%;
  z-index: 1;
}

.character-portrait-placeholder-fill {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(47, 66, 84, 0.74), rgba(30, 47, 63, 0.9));
  color: rgba(236, 202, 143, 0.88);
  font-weight: 700;
  font-size: 28px;
  z-index: 1;
}

.turn-order-badge {
  position: absolute;
  top: 6px;
  left: 6px;
  z-index: 5;
  min-width: 24px;
  height: 18px;
  border-radius: 999px;
  border: 1px solid rgba(235, 203, 144, 0.76);
  background: linear-gradient(180deg, rgba(110, 78, 35, 0.92), rgba(71, 49, 21, 0.92));
  color: #ffe8be;
  font-size: 10px;
  font-weight: 800;
  line-height: 16px;
  text-align: center;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.62);
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.34);
}

.player-overlay {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  min-height: 52%;
  padding: 6px 6px 7px;
  display: flex;
  flex-direction: column;
  gap: 3px;
  justify-content: flex-end;
  background:
    linear-gradient(180deg, rgba(4, 10, 18, 0.04) 0%, rgba(8, 16, 27, 0.78) 32%, rgba(6, 12, 21, 0.94) 100%);
  backdrop-filter: blur(2px);
  z-index: 4;
}

.player-overlay-title-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-height: 16px;
}

.player-overlay-role {
  font-size: 11px;
  line-height: 1.15;
  font-weight: 700;
  color: rgba(241, 247, 255, 0.96);
  max-width: calc(100% - 38px);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-shadow: 0 1px 2px rgba(3, 10, 20, 0.68);
}

.player-overlay-player {
  font-size: 10px;
  line-height: 1.1;
  text-align: center;
  color: rgba(196, 213, 227, 0.92);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.player-overlay-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 2px 4px;
  font-size: 10px;
  color: rgba(228, 237, 245, 0.92);
}

.player-overlay-stat {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
}

.player-overlay-resource {
  grid-column: 1 / -1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.player-overlay-effects {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 2px;
  max-height: 20px;
  overflow: hidden;
}

.player-overlay-tokens {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 3px;
  max-height: 32px;
  overflow: hidden;
}

.player-overlay-ml-status {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 3px 4px;
  border: 1px solid rgba(148, 163, 184, 0.34);
  border-radius: 6px;
  background: rgba(15, 23, 42, 0.6);
}

.player-overlay-ml-line {
  font-size: 10px;
  line-height: 1.22;
  color: rgba(226, 232, 240, 0.96);
  text-align: left;
}

.player-overlay-ml-skills {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 3px 4px;
  border: 1px solid rgba(217, 119, 6, 0.36);
  border-radius: 6px;
  background: rgba(35, 24, 12, 0.62);
  max-height: 88px;
  overflow-y: auto;
}

.player-overlay-ml-title {
  font-size: 10px;
  line-height: 1.2;
  font-weight: 700;
  color: rgba(253, 230, 138, 0.95);
  text-align: left;
}

.player-overlay-ml-skill-item {
  font-size: 9px;
  line-height: 1.24;
  text-align: left;
  color: rgba(243, 244, 246, 0.95);
}

.player-overlay-ml-skill-name {
  font-weight: 700;
  color: rgba(251, 191, 36, 0.96);
}

.player-overlay-ml-skill-desc {
  color: rgba(241, 245, 249, 0.93);
}

.btn-show-skills {
  border: 1px solid rgba(223, 188, 124, 0.36);
  background: rgba(79, 57, 30, 0.46);
  color: rgba(248, 221, 171, 0.95);
  transition: background-color 0.2s ease, transform 0.2s ease;
}

.btn-show-skills:hover {
  background: rgba(103, 72, 38, 0.58);
  transform: translateY(-1px);
}

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
  color: #ff9186;
  text-shadow: 0 0 8px rgba(255, 128, 118, 0.68), 0 0 16px rgba(255, 80, 69, 0.45);
  animation: numberPop 0.44s ease-out;
}
.damage-splash {
  position: absolute;
  inset: -16px;
  background: radial-gradient(circle, rgba(255, 83, 83, 0.34) 0%, rgba(232, 59, 59, 0.2) 36%, transparent 72%);
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

@keyframes frameGlowPulse {
  0%, 100% {
    box-shadow:
      0 0 0 1px rgba(246, 224, 181, 0.64),
      0 0 12px rgba(213, 171, 106, 0.44),
      0 0 24px rgba(157, 117, 71, 0.3);
  }
  50% {
    box-shadow:
      0 0 0 1px rgba(249, 232, 200, 0.86),
      0 0 22px rgba(226, 187, 124, 0.62),
      0 0 34px rgba(172, 129, 81, 0.4);
  }
}

@media (min-width: 1800px) {
  .player-area--compact {
    height: 212px;
  }

  .player-area--full {
    height: 252px;
  }
}

@media (max-width: 900px) {
  .player-area--compact {
    height: 184px;
  }

  .player-area--full {
    height: 214px;
  }

  .player-overlay {
    padding: 5px 5px 6px;
    min-height: 56%;
  }

  .turn-order-badge {
    top: 5px;
    left: 5px;
  }

  .player-overlay-role {
    font-size: 10px;
  }

  .player-overlay-player,
  .player-overlay-stats {
    font-size: 9px;
  }
}

@media (max-width: 640px) {
  .player-area--compact {
    height: 176px;
  }

  .player-area--full {
    height: 202px;
  }

  .player-overlay {
    padding: 4px 4px 5px;
    gap: 2px;
  }

  .player-overlay-role {
    font-size: 9px;
    max-width: calc(100% - 34px);
  }

  .player-overlay-player {
    font-size: 8px;
  }

  .player-overlay-stats {
    font-size: 8px;
    gap: 1px 3px;
  }

  .player-overlay-effects {
    max-height: 16px;
  }

  .player-overlay-tokens {
    max-height: 26px;
  }

  .player-overlay-ml-line,
  .player-overlay-ml-title,
  .player-overlay-ml-skill-item {
    font-size: 8px;
  }

  .player-overlay-ml-skills {
    max-height: 70px;
  }

  .btn-show-skills {
    font-size: 8px;
    padding: 1px 3px;
  }

  .turn-order-badge {
    min-width: 22px;
    height: 16px;
    font-size: 9px;
    line-height: 14px;
  }
}
</style>
