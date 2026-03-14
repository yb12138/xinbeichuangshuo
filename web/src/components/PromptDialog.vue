<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { useWebSocket } from '../composables/useWebSocket'
import CardComponent from './CardComponent.vue'
import PlayerArea from './PlayerArea.vue'

const store = useGameStore()
const ws = useWebSocket()

const prompt = computed(() => store.currentPrompt)

const ELEMENT_LABEL_MAP: Record<string, string> = {
  Water: '水',
  Fire: '火',
  Earth: '地',
  Wind: '风',
  Thunder: '雷',
  Light: '光',
  Dark: '暗'
}

// 行动选择（攻击/法术/购买/提取/合成）不弹窗，用 ActionPanel 的按钮
const isActionSelectionPrompt = computed(() => {
  if (!prompt.value) return false
  if (prompt.value.ui_mode === 'action_hub') return true
  if (!prompt.value.message) return false
  return prompt.value.message.includes('行动类型')
})

const isVisible = computed(() => 
  prompt.value !== null && store.isPromptForMe && !isActionSelectionPrompt.value
)

const selectedOptions = ref<string[]>([])
const selectedCardIndices = ref<number[]>([])
const selectedCounterTarget = ref<string | null>(null)
const selectedExtractIndices = ref<number[]>([])

// 重置选择
watch(() => prompt.value, () => {
  selectedOptions.value = []
  selectedCardIndices.value = []
  selectedCounterTarget.value = null
  selectedExtractIndices.value = []
})

// 判断是否需要选择手牌（含应战/防御时需选牌）
const hasCounterOption = computed(() => {
  if (!prompt.value?.options?.length) return false
  return prompt.value.options.some((o: { id: string }) => o.id === 'counter')
})

const hasDefendOption = computed(() => {
  if (!prompt.value?.options?.length) return false
  return prompt.value.options.some((o: { id: string }) => o.id === 'defend')
})

const hasCounterOrDefend = computed(() => {
  if (!prompt.value?.options?.length) return false
  return hasCounterOption.value || hasDefendOption.value
})

const isMagicMissilePrompt = computed(() => {
  const msg = prompt.value?.message ?? ''
  return msg.includes('魔弹')
})

const incomingAttackElement = computed(() => prompt.value?.attack_element ?? '')
const incomingAttackElementLabel = computed(() => {
  const el = incomingAttackElement.value
  if (!el) return '未知'
  return ELEMENT_LABEL_MAP[el] ?? el
})
const incomingAttackerName = computed(() => {
  const attackerId = prompt.value?.attacker_id
  if (!attackerId) return ''
  const roleId = store.players[attackerId]?.role
  return store.getRoleDisplayName(roleId)
})

const effectHints = computed(() => prompt.value?.effect_hints ?? [])
const hasShieldFallbackHint = computed(() =>
  effectHints.value.some((hint: string) => hint.includes('承受伤害') && hint.includes('圣盾'))
)

function roleDisplayName(roleId?: string): string {
  return store.getRoleDisplayName(roleId)
}

const incomingSummary = computed(() => {
  const message = prompt.value?.message ?? ''
  if (!message) return ''
  let replaced = message
  for (const [id, p] of Object.entries(store.players)) {
    const roleName = store.getRoleDisplayName(p?.role)
    if (p?.name) {
      replaced = replaced.split(p.name).join(roleName)
    }
    if (id) {
      replaced = replaced.split(id).join(roleName)
    }
  }
  return replaced
})

const needsCardSelection = computed(() => {
  if (!prompt.value) return false
  if (prompt.value.type === 'choose_card' || prompt.value.type === 'choose_cards') return true
  // 应战(counter)需选攻击牌，防御(defend)仅可选圣光（圣盾需提前放置）
  if (hasCounterOrDefend.value) return true
  return false
})

// 判断是否需要选择目标
const needsTargetSelection = computed(() => {
  if (!prompt.value) return false
  return prompt.value.type === 'choose_target'
})

// 应战是否需选反弹目标（攻击方队友，不含攻击者）
const needsCounterTargetSelection = computed(() => {
  if (!prompt.value) return false
  const ids = prompt.value.counter_target_ids
  return hasCounterOrDefend.value && ids && ids.length > 0
})

// 应战可选目标玩家（从 store 中按 counter_target_ids 解析）
const counterTargetPlayers = computed(() => {
  const ids = prompt.value?.counter_target_ids
  if (!ids?.length) return []
  return ids
    .map(id => store.players[id])
    .filter((p): p is NonNullable<typeof p> => p != null)
})

const isWaterShadowPrompt = computed(() => {
  return !hasCounterOrDefend.value && !!prompt.value?.message?.includes('水影')
})

const isStealthed = computed(() => {
  return !!store.myPlayer?.field?.some((fc) => fc.mode === 'Effect' && fc.effect === 'Stealth')
})

// 应战/防御时，仅同系攻击牌或暗灭可应战；防御仅可使用圣光
function isCardSelectableForCounterOrDefend(card: { type: string; element: string; name: string }, idx: number): boolean {
  if (!hasCounterOrDefend.value || !prompt.value?.options) {
    if (!isWaterShadowPrompt.value) return true
    const selectedCards = selectedCardIndices.value
      .map((i) => store.myHand[i])
      .filter((c): c is NonNullable<typeof c> => !!c)
    const waterCount = selectedCards.filter((c) => c.element === 'Water').length
    const magicCount = selectedCards.filter((c) => c.type === 'Magic' && c.element !== 'Water').length
    if (card.element === 'Water') return true
    if (isStealthed.value && card.type === 'Magic') {
      if (selectedCardIndices.value.includes(idx)) return true
      if (magicCount >= 1) return false
      return waterCount > 0
    }
    return false
  }
  const hasCounter = hasCounterOption.value
  const hasDefend = hasDefendOption.value
  if (isMagicMissilePrompt.value) {
    const validForCounter = hasCounter && card.type === 'Magic' && card.name === '魔弹'
    const validForDefend = hasDefend && card.type === 'Magic' && card.name === '圣光'
    return validForCounter || validForDefend
  }
  const attackEl = prompt.value.attack_element
  // 应战：同系或暗灭（若无 attack_element 则放宽为所有攻击牌，后端会校验）
  const validForCounter = hasCounter && card.type === 'Attack' &&
    (!attackEl || card.element === attackEl || card.element === 'Dark')
  const validForDefend = hasDefend && card.type === 'Magic' && card.name === '圣光'
  return validForCounter || validForDefend
}

// 判断是否为确认类型
const isConfirmType = computed(() => {
  if (!prompt.value) return false
  return prompt.value.type === 'confirm' || (prompt.value.options.length > 0 && prompt.value.type !== 'choose_extract')
})

// 提炼多选
const isExtractPrompt = computed(() => prompt.value?.type === 'choose_extract')

function toggleExtractOption(index: number) {
  const idx = selectedExtractIndices.value.indexOf(index)
  if (idx >= 0) {
    selectedExtractIndices.value.splice(idx, 1)
  } else {
    const max = prompt.value?.max ?? 2
    if (selectedExtractIndices.value.length < max) {
      selectedExtractIndices.value.push(index)
      selectedExtractIndices.value.sort((a, b) => a - b)
    }
  }
}

function confirmExtractSelection() {
  const min = prompt.value?.min ?? 1
  const max = prompt.value?.max ?? 2
  const sel = selectedExtractIndices.value
  if (sel.length < min || sel.length > max) return
  ws.select(sel)
}

// 可选目标列表
const targetPlayers = computed(() => {
  return Object.values(store.players).filter(p => p.id !== store.myPlayerId)
})

const targetSelectionPlayers = computed(() => {
  if (!prompt.value) return targetPlayers.value
  if (!prompt.value.options?.length) return targetPlayers.value
  const fromOptions = prompt.value.options
    .map(option => store.players[option.id])
    .filter((p): p is NonNullable<typeof p> => p != null)
  return fromOptions.length > 0 ? fromOptions : targetPlayers.value
})

function isAllyPlayer(player: { id: string; camp: string }): boolean {
  if (store.myCamp) return player.camp === store.myCamp
  const myCamp = store.players[store.myPlayerId]?.camp
  if (myCamp) return player.camp === myCamp
  return player.id === store.myPlayerId
}

const groupedTargetSelectionPlayers = computed(() => {
  const enemies: typeof targetSelectionPlayers.value = []
  const allies: typeof targetSelectionPlayers.value = []
  for (const player of targetSelectionPlayers.value) {
    if (isAllyPlayer(player)) allies.push(player)
    else enemies.push(player)
  }
  return { enemies, allies }
})

const playerOptionEntries = computed(() => {
  if (!prompt.value?.options?.length) return []
  return prompt.value.options
    .map(option => {
      const player = store.players[option.id]
      if (!player) return null
      return { option, player }
    })
    .filter((entry): entry is { option: { id: string; label: string }; player: NonNullable<typeof store.players[string]> } => entry != null)
})

const showPlayerOptionCards = computed(() => {
  if (!prompt.value) return false
  if (!isConfirmType.value || needsCardSelection.value || prompt.value.type === 'choose_cards') return false
  if (!prompt.value.options?.length) return false
  return playerOptionEntries.value.length === prompt.value.options.length
})

const canCancelPrompt = computed(() => {
  if (!prompt.value) return false
  if (prompt.value.type === 'choose_skill') return true
  return (prompt.value.options ?? []).some((option: { id: string }) => option.id === 'skip' || option.id === 'cancel')
})

function toggleCardSelection(index: number) {
  const idx = selectedCardIndices.value.indexOf(index)
  if (idx === -1) {
    const card = store.myHand[index]
    if (card && !isCardSelectableForCounterOrDefend(card, index)) {
      if (isWaterShadowPrompt.value) {
        store.setError(isStealthed.value ? '水影仅可弃水系牌，潜行状态下最多额外弃1张法术牌' : '水影仅可弃水系牌')
      } else {
        store.setError('当前卡牌不可选择')
      }
      return
    }
    if (prompt.value && prompt.value.max === 1) {
      selectedCardIndices.value = [index]
    } else {
      selectedCardIndices.value.push(index)
    }
  } else {
    selectedCardIndices.value.splice(idx, 1)
  }
}

function handleOptionClick(optionId: string) {
  if (optionId === 'counter_disabled') {
    store.setError('此攻击无法应战')
    return
  }
  if (prompt.value?.type === 'choose_skill') {
    const idx = prompt.value.options.findIndex((o: { id: string }) => o.id === optionId)
    if (idx >= 0) {
      ws.select([idx])
    } else {
      if (!canCancelPrompt.value) {
        store.setError('当前步骤不可取消，请先完成本次操作')
        return
      }
      ws.cancel()
    }
    return
  }
  if (optionId === 'skip' || optionId === 'cancel') {
    if (!canCancelPrompt.value) {
      store.setError('当前步骤不可取消，请先完成本次操作')
      return
    }
    ws.cancel()
    return
  }
  if (optionId === 'confirm') {
    ws.confirm()
    return
  }
  // 魔弹融合等确认选项：yes=0, no=1
  if (optionId === 'yes' || optionId === 'no') {
    ws.select([optionId === 'yes' ? 0 : 1])
    return
  }
  // 魔弹掌控方向选择：normal=0, reverse=1
  if (optionId === 'normal' || optionId === 'reverse') {
    ws.select([optionId === 'normal' ? 0 : 1])
    return
  }
  if (optionId === 'take') {
    ws.respond('take')
    return
  }
  if (optionId === 'counter') {
    if (selectedCardIndices.value.length === 0) {
      store.setError(isMagicMissilePrompt.value ? '请先选择一张【魔弹】进行传递' : '请先选择一张攻击牌进行应战')
      return
    }
    if (needsCounterTargetSelection.value && !selectedCounterTarget.value) {
      store.setError('请先选择反弹目标（攻击方的队友）')
      return
    }
    ws.respond('counter', selectedCardIndices.value[0], selectedCounterTarget.value || undefined)
    return
  }
  if (optionId === 'defend') {
    if (selectedCardIndices.value.length === 0) {
      store.setError('请先选择一张【圣光】进行防御（圣盾需提前放置）')
      return
    }
    ws.respond('defend', selectedCardIndices.value[0])
    return
  }
  {
    const optionIndex = prompt.value?.options?.findIndex((o: { id: string }) => o.id === optionId) ?? -1
    if (optionIndex >= 0) {
      ws.select([optionIndex])
    } else {
      const index = parseInt(optionId)
      if (!isNaN(index)) {
        ws.select([index])
      } else {
        ws.sendAction({
          player_id: store.myPlayerId,
          type: 'Select',
          skill_id: optionId
        })
      }
    }
  }
  // 不在此处清除 prompt：等后端 state_update 或新 prompt 到达
}

function confirmCardSelection() {
  if (selectedCardIndices.value.length > 0) {
    ws.select(selectedCardIndices.value)
  }
}

function confirmTargetSelection(targetId: string) {
  ws.sendAction({
    player_id: store.myPlayerId,
    type: 'Select',
    target_id: targetId
  })
}

// 弹窗选项按钮样式（参考 noname 类型区分）
function getDialogOptionClass(optionId: string): string {
  if (optionId === 'take' || optionId === 'confirm' || optionId === 'yes') return 'btn-success'
  if (optionId === 'counter') return 'btn-skill bg-purple-900/60 hover:bg-purple-800/80'
  if (optionId === 'defend') return 'btn-primary bg-blue-900/60 hover:bg-blue-800/80'
  if (optionId === 'skip' || optionId === 'cancel') return 'btn-secondary'
  return 'bg-gray-700 text-gray-200 hover:bg-gray-600'
}

function getElementBadgeClass(element: string): string {
  if (element === 'Water') return 'element-badge-water'
  if (element === 'Fire') return 'element-badge-fire'
  if (element === 'Earth') return 'element-badge-earth'
  if (element === 'Wind') return 'element-badge-wind'
  if (element === 'Thunder') return 'element-badge-thunder'
  if (element === 'Light') return 'element-badge-light'
  if (element === 'Dark') return 'element-badge-dark'
  return 'element-badge-neutral'
}

// 检查选择是否有效
const isSelectionValid = computed(() => {
  if (!prompt.value) return false
  
  if (needsCardSelection.value) {
    return selectedCardIndices.value.length >= prompt.value.min &&
           selectedCardIndices.value.length <= prompt.value.max
  }
  
  return true
})

const cardSelectionOptions = computed<Array<{ id: string; label: string; disabled?: boolean }>>(() => {
  if (!prompt.value?.options || !needsCardSelection.value || !hasCounterOrDefend.value) return []
  return prompt.value.options
    .filter((option: { id: string }) => option.id === 'take' || option.id === 'counter' || option.id === 'defend')
    .map((option) => ({ id: option.id, label: option.label, disabled: false }))
})

const cardActionButtonGridClass = computed(() => {
  const count = cardSelectionOptions.value.length
  if (count <= 1) return 'grid-cols-1'
  if (count === 2) return 'grid-cols-2'
  return 'grid-cols-3'
})

const selectableCards = computed(() => {
  if (hasCounterOrDefend.value) {
    return store.myPlayableCards.map(item => ({
      card: item.card,
      index: item.index
    }))
  }
  return store.myHand.map((card, index) => ({
    card,
    index
  }))
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div 
        v-if="isVisible" 
        class="prompt-overlay fixed inset-0 z-[2200] flex items-center justify-center bg-black/70 backdrop-blur-sm"
      >
        <div class="prompt-card prompt-card-shell glass-panel max-w-lg w-full mx-4 overflow-hidden dialog-pop">
        <!-- 标题栏 -->
        <div class="dialog-header prompt-header px-6 py-4">
          <h3 class="text-lg font-bold text-white">策略抉择</h3>
        </div>

        <!-- 内容区域 -->
        <div class="prompt-body p-6 space-y-4">
          <!-- 提示消息 -->
          <div v-if="!hasCounterOrDefend" class="text-gray-200 text-lg">
            {{ prompt?.message }}
          </div>

          <div v-if="hasCounterOrDefend" class="incoming-attack-card">
            <div class="incoming-title">来袭信息</div>
            <div class="incoming-summary">{{ incomingSummary }}</div>
            <div class="incoming-content">
              <span class="incoming-label">当前系刀</span>
              <span class="incoming-element-badge" :class="getElementBadgeClass(incomingAttackElement)">
                {{ incomingAttackElementLabel }}系
              </span>
              <span v-if="incomingAttackerName" class="incoming-attacker">攻击者：{{ incomingAttackerName }}</span>
            </div>
            <div v-if="effectHints.length > 0" class="incoming-hints">
              <div
                v-for="(hint, idx) in effectHints"
                :key="`hint-${idx}`"
                class="incoming-hint-item"
              >
                {{ hint }}
              </div>
            </div>
          </div>

          <!-- 提炼多选：展示战绩区星石，点击切换选择 -->
          <div v-if="isExtractPrompt && prompt?.options?.length" class="space-y-3">
            <div class="flex flex-wrap gap-2">
              <button
                v-for="(option, idx) in prompt.options"
                :key="option.id"
                class="py-3 px-4 rounded-lg font-semibold transition-all"
                :class="[
                  option.label === '红宝石' ? 'bg-red-900/60 hover:bg-red-800/80 text-red-200' : 'bg-indigo-900/60 hover:bg-indigo-800/80 text-indigo-200',
                  selectedExtractIndices.includes(idx) ? 'ring-2 ring-yellow-400 scale-105' : ''
                ]"
                @click="toggleExtractOption(idx)"
              >
                {{ option.label === '红宝石' ? '♦' : '🔷' }} {{ option.label }}
              </button>
            </div>
            <button
              class="w-full py-3 rounded-lg font-bold btn-success"
              :class="{ 'opacity-50 cursor-not-allowed': selectedExtractIndices.length < (prompt?.min ?? 1) || selectedExtractIndices.length > (prompt?.max ?? 2) }"
              :disabled="selectedExtractIndices.length < (prompt?.min ?? 1) || selectedExtractIndices.length > (prompt?.max ?? 2)"
              @click="confirmExtractSelection"
            >
              确认提炼 ({{ selectedExtractIndices.length }}/{{ prompt?.max ?? 2 }})
            </button>
          </div>
          <!-- 选项按钮（choose_cards 用下方手牌选择，不在此展示） -->
          <div v-else-if="isConfirmType && prompt?.options?.length && prompt?.type !== 'choose_cards' && !needsCardSelection" class="space-y-2">
            <div v-if="showPlayerOptionCards" class="target-player-grid grid grid-cols-2 gap-2 sm:gap-3">
              <PlayerArea
                v-for="entry in playerOptionEntries"
                :key="entry.option.id"
                :player="entry.player"
                :selectable="true"
                @select="() => handleOptionClick(entry.option.id)"
              />
            </div>
            <button
              v-else
              v-for="option in prompt.options"
              :key="option.id"
              class="w-full py-3 px-4 rounded-lg font-semibold text-left"
              :class="[
                getDialogOptionClass(option.id),
                selectedOptions.includes(option.id) ? 'ring-2 ring-yellow-400 scale-[1.02]' : ''
              ]"
              @click="handleOptionClick(option.id)"
            >
              {{ option.label }}
            </button>
          </div>

          <!-- 应战反弹目标选择（攻击方队友，不含攻击者） -->
          <div v-if="needsCounterTargetSelection" class="space-y-3">
            <div class="text-sm text-gray-400">请选择反弹目标（攻击方的队友，不能选攻击者本人）：</div>
            <div class="grid grid-cols-2 gap-2">
              <button
                v-for="player in counterTargetPlayers"
                :key="player.id"
                class="btn-target py-3 px-4 rounded-lg"
                :class="[
                  selectedCounterTarget === player.id ? 'selected ring-2 ring-yellow-400 bg-amber-900/50' : '',
                  player.camp === 'Red' ? 'bg-red-900/50 hover:bg-red-800/60' : 'bg-blue-900/50 hover:bg-blue-800/60'
                ]"
                @click="selectedCounterTarget = selectedCounterTarget === player.id ? null : player.id"
              >
                <div class="font-bold">{{ player.name }}</div>
                <div class="text-sm text-gray-400">{{ roleDisplayName(player.role) }}</div>
              </button>
            </div>
          </div>

          <!-- 手牌选择区域（选牌 / 应战选牌 / 防御选牌） -->
          <div v-if="needsCardSelection" class="space-y-3">
            <div class="text-sm text-gray-400">
              <template v-if="hasCounterOrDefend">
                <template v-if="isMagicMissilePrompt">
                  魔弹传递：请选择【魔弹】进行传递；防御仅可选【圣光】（圣盾需提前放置）。先选牌再点对应按钮：
                </template>
                <template v-else-if="hasCounterOption">
                  应战须同系攻击牌或暗灭。当前来袭为「{{ incomingAttackElementLabel }}系」；防御仅可选【圣光】（圣盾需提前放置）。先选牌再点对应按钮：
                </template>
                <template v-else>
                  此攻击无法应战。你可以选择防御（仅【圣光】；圣盾需提前放置）或承受伤害：
                </template>
                <template v-if="hasShieldFallbackHint">
                  你可先应战或防御；若选择承受伤害，将先触发场上【圣盾】抵挡本次攻击。
                </template>
              </template>
              <template v-else>
                请选择 {{ prompt?.min }}-{{ prompt?.max }} 张手牌:
              </template>
            </div>
            <div class="flex gap-2 flex-wrap justify-center">
              <CardComponent
                v-for="entry in selectableCards"
                :key="entry.index"
                :card="entry.card"
                :index="entry.index"
                :selectable="isCardSelectableForCounterOrDefend(entry.card, entry.index)"
                :selected="selectedCardIndices.includes(entry.index)"
                @click="toggleCardSelection"
              />
            </div>
            <!-- 应战/防御操作按钮合并到选牌区域，避免双区块歧义 -->
            <div v-if="cardSelectionOptions.length > 0" :class="['grid gap-2', cardActionButtonGridClass]">
              <button
                v-for="option in cardSelectionOptions"
                :key="option.id"
                class="w-full py-3 px-3 rounded-lg font-semibold text-center whitespace-nowrap"
                :class="[
                  getDialogOptionClass(option.id),
                  selectedOptions.includes(option.id) ? 'ring-2 ring-yellow-400 scale-[1.02]' : '',
                  option.disabled ? 'opacity-45 cursor-not-allowed bg-gray-700 text-gray-400 hover:bg-gray-700' : ''
                ]"
                :disabled="!!option.disabled"
                @click="handleOptionClick(option.id)"
              >
                {{ option.label }}
              </button>
            </div>
            <!-- 普通选牌使用确认按钮 -->
            <button
              v-if="!hasCounterOrDefend"
              class="w-full py-3 rounded-lg font-bold btn-success"
              :class="{ 'opacity-50 cursor-not-allowed': !isSelectionValid }"
              :disabled="!isSelectionValid"
              @click="confirmCardSelection"
            >
              确认选择 ({{ selectedCardIndices.length }}/{{ prompt?.max }})
            </button>
          </div>

          <!-- 目标选择区域 -->
          <div v-if="needsTargetSelection" class="space-y-3">
            <div class="text-sm text-gray-400">请选择目标:</div>
            <div class="prompt-target-group-stack">
              <div v-if="groupedTargetSelectionPlayers.enemies.length > 0" class="prompt-target-group-card">
                <div class="prompt-target-group-title prompt-target-group-title--enemy">敌方阵营</div>
                <div class="target-player-grid grid grid-cols-2 gap-2 sm:gap-3">
                  <PlayerArea
                    v-for="player in groupedTargetSelectionPlayers.enemies"
                    :key="player.id"
                    :player="player"
                    :selectable="true"
                    @select="() => confirmTargetSelection(player.id)"
                  />
                </div>
              </div>
              <div v-if="groupedTargetSelectionPlayers.allies.length > 0" class="prompt-target-group-card">
                <div class="prompt-target-group-title prompt-target-group-title--ally">我方阵营</div>
                <div class="target-player-grid grid grid-cols-2 gap-2 sm:gap-3">
                  <PlayerArea
                    v-for="player in groupedTargetSelectionPlayers.allies"
                    :key="player.id"
                    :player="player"
                    :selectable="true"
                    @select="() => confirmTargetSelection(player.id)"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 底部操作栏 -->
        <div class="prompt-footer px-6 py-4 flex justify-end gap-2">
          <button
            class="btn-secondary px-4 py-2"
            :class="{ 'opacity-50 cursor-not-allowed': !canCancelPrompt }"
            :disabled="!canCancelPrompt"
            @click="handleOptionClick('cancel')"
          >
            {{ canCancelPrompt ? '取消/跳过' : '当前不可取消' }}
          </button>
        </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.prompt-overlay {
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(360px 180px at 50% 48%, rgba(141, 192, 196, 0.18), transparent 70%),
    rgba(1, 7, 12, 0.76);
}

.prompt-card {
  background:
    linear-gradient(180deg, rgba(8, 20, 34, 0.92), rgba(7, 16, 28, 0.94)),
    url('/assets/ui/modal-aura.svg') center/cover no-repeat;
  border: 1px solid rgba(130, 167, 187, 0.38);
  box-shadow:
    0 24px 54px rgba(2, 8, 18, 0.62),
    inset 0 1px 0 rgba(238, 246, 252, 0.1);
}

.prompt-card-shell {
  max-height: min(90dvh, 760px);
  display: flex;
  flex-direction: column;
}

.prompt-header {
  background: linear-gradient(100deg, rgba(30, 73, 98, 0.86), rgba(95, 72, 43, 0.86));
  border-bottom: 1px solid rgba(160, 193, 211, 0.24);
}

.prompt-body {
  background: rgba(4, 15, 25, 0.35);
  overflow-y: auto;
  min-height: 0;
}

.prompt-body .text-gray-200 {
  color: #e7f0f7;
}

.prompt-body .text-gray-400 {
  color: #a7bed1;
}

.prompt-footer {
  background: rgba(7, 19, 32, 0.68);
  border-top: 1px solid rgba(118, 152, 174, 0.24);
}

.prompt-body .ring-yellow-400 {
  --tw-ring-color: rgba(231, 192, 131, 0.88);
}

.incoming-attack-card {
  border: 1px solid rgba(143, 176, 198, 0.4);
  border-radius: 10px;
  background:
    linear-gradient(180deg, rgba(11, 28, 44, 0.74), rgba(7, 18, 30, 0.8));
  padding: 10px 12px;
}

.incoming-title {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: #c7dced;
  margin-bottom: 4px;
}

.incoming-summary {
  color: #e4eef7;
  font-size: 14px;
  line-height: 1.4;
  margin-bottom: 8px;
}

.incoming-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.incoming-label {
  color: #9bb8cb;
  font-size: 12px;
}

.incoming-element-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 22px;
  border-radius: 999px;
  padding: 0 10px;
  font-size: 12px;
  font-weight: 700;
  border: 1px solid transparent;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
}

.incoming-attacker {
  color: #b5cad9;
  font-size: 12px;
}

.incoming-hints {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.incoming-hint-item {
  font-size: 12px;
  line-height: 1.35;
  color: #f6e7c8;
  border-left: 2px solid rgba(228, 185, 116, 0.75);
  padding-left: 8px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.55);
}

.prompt-target-group-stack {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.prompt-target-group-card {
  border: 1px solid rgba(126, 158, 180, 0.35);
  border-radius: 10px;
  background: rgba(8, 20, 34, 0.6);
  padding: 8px;
}

.prompt-target-group-title {
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.prompt-target-group-title--enemy {
  color: rgba(251, 113, 133, 0.92);
}

.prompt-target-group-title--ally {
  color: rgba(125, 211, 252, 0.92);
}

.element-badge-water {
  color: #c8e9ff;
  border-color: rgba(114, 182, 226, 0.56);
  background: linear-gradient(140deg, rgba(24, 77, 126, 0.9), rgba(17, 51, 84, 0.9));
}

.element-badge-fire {
  color: #ffd8d0;
  border-color: rgba(230, 126, 103, 0.56);
  background: linear-gradient(140deg, rgba(133, 44, 33, 0.9), rgba(96, 33, 26, 0.9));
}

.element-badge-earth {
  color: #f6e2ba;
  border-color: rgba(205, 156, 84, 0.56);
  background: linear-gradient(140deg, rgba(125, 84, 38, 0.9), rgba(90, 58, 24, 0.9));
}

.element-badge-wind {
  color: #cff6e8;
  border-color: rgba(97, 186, 156, 0.56);
  background: linear-gradient(140deg, rgba(30, 112, 87, 0.9), rgba(20, 78, 62, 0.9));
}

.element-badge-thunder {
  color: #efe4ff;
  border-color: rgba(160, 128, 225, 0.56);
  background: linear-gradient(140deg, rgba(83, 53, 138, 0.9), rgba(56, 35, 93, 0.9));
}

.element-badge-light {
  color: #fff6d5;
  border-color: rgba(225, 192, 118, 0.56);
  background: linear-gradient(140deg, rgba(153, 119, 44, 0.9), rgba(112, 84, 30, 0.9));
}

.element-badge-dark {
  color: #d6def7;
  border-color: rgba(120, 132, 174, 0.56);
  background: linear-gradient(140deg, rgba(34, 43, 76, 0.9), rgba(23, 30, 55, 0.9));
}

.element-badge-neutral {
  color: #d2deeb;
  border-color: rgba(136, 160, 181, 0.56);
  background: linear-gradient(140deg, rgba(40, 58, 79, 0.9), rgba(27, 38, 56, 0.9));
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.24s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .prompt-card,
.modal-leave-to .prompt-card {
  transform: scale(0.94) translateY(10px);
}

.modal-enter-active .prompt-card,
.modal-leave-active .prompt-card {
  transition: transform 0.26s ease;
}

@media (max-width: 640px) {
  .prompt-overlay {
    align-items: center;
    justify-content: center;
    padding: max(12px, var(--safe-top)) var(--safe-right) max(12px, var(--safe-bottom)) var(--safe-left);
  }

  .prompt-card-shell {
    max-width: min(100vw, 560px);
    max-height: min(88dvh, 700px);
    border-radius: 14px;
  }

  .prompt-header {
    padding: 12px 14px;
  }

  .prompt-body {
    padding: 14px;
  }

  .prompt-footer {
    padding: 10px 14px;
  }

  .prompt-body .grid.grid-cols-2 {
    grid-template-columns: 1fr;
  }
}
</style>
