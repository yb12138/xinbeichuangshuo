<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useSubmitAction } from '../composables/useSubmitAction'
import { ROLE_NAME_MAP } from '../constants/roleNameMap'
import type { PlayerView } from '../types/game'

const interruptStore = useInterruptStore()
const sessionStore = useSessionStore()
const snapshotStore = useSnapshotStore()
const actions = useSubmitAction()

const prompt = computed(() => interruptStore.currentPrompt)
const myPlayerId = computed(() => sessionStore.myPlayerId)
const playerViews = computed(() => snapshotStore.players)
const myHand = computed(() => playerViews.value[myPlayerId.value]?.hand || [])

function getRoleDisplayName(roleId?: string): string {
  if (!roleId) return '未知角色'
  return snapshotStore.characters[roleId]?.name || ROLE_NAME_MAP[roleId] || '未知角色'
}

function showPromptError(message: string) {
  interruptStore.showError(message)
}

// 行动选择（攻击/法术/购买/提取/合成）不在这里显示，由 ActionPanel 承载
const isActionSelectionPrompt = computed(() => {
  if (!prompt.value) return false
  if (prompt.value.ui_mode === 'action_hub') return true
  if (!prompt.value.message) return false
  return prompt.value.message.includes('行动类型')
})

const isVisible = computed(() =>
  prompt.value !== null && prompt.value.player_id === myPlayerId.value && !isActionSelectionPrompt.value
)

const selectedExtractIndices = ref<number[]>([])
const selectedInlineCardOptionIndices = ref<number[]>([])
const autoResolvedPromptKey = ref('')

watch(() => prompt.value, () => {
  interruptStore.setPromptCounterTarget('')
  selectedExtractIndices.value = []
  selectedInlineCardOptionIndices.value = []
  if (!prompt.value) {
    autoResolvedPromptKey.value = ''
  }
})

type ResponseOptionKind = 'take' | 'counter' | 'defend' | null

function responseOptionKind(option: { id?: string; label?: string; button_label?: string }): ResponseOptionKind {
  const id = String(option.id || '').trim().toLowerCase()
  // 魔弹掌控方向选择不是战斗应答（take/defend/counter），避免被“传递”关键词误判为应战。
  if (id === 'normal' || id === 'reverse') return null
  const label = String(option.label || '').trim()
  const buttonLabel = String(option.button_label || '').trim()
  const text = `${label} ${buttonLabel}`.toLowerCase()
  if (id === 'take' || id === 'take_damage' || text.includes('承受') || text.includes('命中')) return 'take'
  if (id === 'defend' || text.includes('防御')) return 'defend'
  if (id === 'counter' || text.includes('应战')) return 'counter'
  return null
}

const hasCounterOption = computed(() => {
  if (!prompt.value?.options?.length) return false
  return prompt.value.options.some((o: { id: string; label: string; button_label?: string }) => responseOptionKind(o) === 'counter')
})

const hasDefendOption = computed(() => {
  if (!prompt.value?.options?.length) return false
  return prompt.value.options.some((o: { id: string; label: string; button_label?: string }) => responseOptionKind(o) === 'defend')
})

const hasCounterOrDefend = computed(() => {
  if (!prompt.value?.options?.length) return false
  return hasCounterOption.value || hasDefendOption.value
})

function promptAttackElementName(raw: string): string {
  const lower = String(raw || '').trim().toLowerCase()
  if (lower === 'water') return '水系'
  if (lower === 'fire') return '火系'
  if (lower === 'earth') return '地系'
  if (lower === 'wind') return '风系'
  if (lower === 'thunder') return '雷系'
  if (lower === 'light') return '光系'
  if (lower === 'dark') return '暗灭'
  return String(raw || '').trim()
}

const responseAttackElementHintText = computed(() => {
  if (!prompt.value || !hasCounterOrDefend.value) return ''
  const attackElement = String(prompt.value.attack_element || '').trim()
  if (!attackElement) return ''
  const displayName = promptAttackElementName(attackElement)
  if (!displayName) return ''
  const lower = attackElement.toLowerCase()
  if (lower === 'dark') return `此次攻击系别：${displayName}`
  if (hasCounterOption.value) return `此次攻击系别：${displayName}（应战需同系或暗灭）`
  return `此次攻击系别：${displayName}`
})

const isDarkAttackResponsePrompt = computed(() => {
  if (!prompt.value || !hasCounterOrDefend.value) return false
  return String(prompt.value.attack_element || '').trim().toLowerCase() === 'dark'
})

const isMagicMissilePrompt = computed(() => {
  const msg = prompt.value?.message ?? ''
  return msg.includes('魔弹')
})

const needsCardSelection = computed(() => {
  if (!prompt.value) return false
  if (prompt.value.type === 'choose_card' || prompt.value.type === 'choose_cards') return true
  if (hasCounterOrDefend.value) return true
  return false
})

const needsTargetSelection = computed(() => {
  if (!prompt.value) return false
  return prompt.value.type === 'choose_target'
})

const needsCounterTargetSelection = computed(() => {
  if (!prompt.value) return false
  const ids = prompt.value.counter_target_ids
  return hasCounterOrDefend.value && ids && ids.length > 0
})

const isConfirmType = computed(() => {
  if (!prompt.value) return false
  return prompt.value.type === 'confirm' || (prompt.value.options.length > 0 && prompt.value.type !== 'choose_extract')
})

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
  actions.submitSelect(sel)
}

function resolveOptionPlayerId(option: { id: string; label: string }): string | null {
  if (playerViews.value[option.id]) return option.id
  const label = String(option.label || '')
  if (!label) return null
  const lowLabel = label.toLowerCase()

  const markersFor = (playerId: string): string[] => {
    const p = playerViews.value[playerId]
    if (!p) return []
    const markers = new Set<string>()
    if (p.id) markers.add(p.id)
    if (p.name) markers.add(p.name)
    if (p.role) {
      markers.add(p.role)
      const roleName = getRoleDisplayName(p.role)
      if (roleName && roleName !== '未知角色') markers.add(roleName)
    }
    return [...markers]
  }

  const matched = Object.values(playerViews.value).filter((p) => {
    const markers = markersFor(p.id)
    return markers.some((marker) => {
      const token = marker.trim().toLowerCase()
      return !!token && lowLabel.includes(token)
    })
  })

  if (matched.length !== 1) return null
  return matched[0]?.id || null
}

const playerOptionEntries = computed(() => {
  if (!prompt.value?.options?.length) return []
  return prompt.value.options
    .map((option, index) => {
      const playerId = resolveOptionPlayerId(option)
      if (!playerId) return null
      const player = playerViews.value[playerId]
      if (!player) return null
      return { index, option, player }
    })
    .filter((entry): entry is { index: number; option: { id: string; label: string }; player: PlayerView } => entry != null)
})

const playerOptionIndexSet = computed(() => {
  const set = new Set<number>()
  for (const entry of playerOptionEntries.value) {
    set.add(entry.index)
  }
  return set
})

const nonPlayerOptions = computed(() => {
  const options = prompt.value?.options ?? []
  return options.filter((_, idx) => !playerOptionIndexSet.value.has(idx))
})

const showConfirmButtonSection = computed(() => {
  return (
    isConfirmType.value &&
    !!prompt.value?.options?.length &&
    prompt.value?.type !== 'choose_cards' &&
    prompt.value?.type !== 'choose_card' &&
    !needsCardSelection.value &&
    !needsTargetSelection.value
  )
})

const isResponseSkillConfirmPrompt = computed(() => {
  if (!prompt.value || prompt.value.type !== 'confirm') return false
  const message = String(prompt.value.message || '').trim()
  if (!message) return false
  if (message.includes('响应技能')) return true
  if (message.includes('是否发动')) return true
  return /是否发动【.+】/.test(message) || /【.+】是否发动/.test(message)
})

const canCancelPrompt = computed(() => {
  if (!prompt.value) return false
  if (prompt.value.type === 'choose_skill' || isResponseSkillConfirmPrompt.value) return true
  return (prompt.value.options ?? []).some((option: { id: string }) =>
    option.id === 'skip' || option.id === 'cancel' || option.id === 'refuse'
  )
})

function handleOptionClick(optionId: string) {
  if (optionId === 'counter_disabled') {
    showPromptError('此攻击无法应战')
    return
  }
  if (prompt.value?.type === 'choose_skill') {
    const idx = prompt.value.options.findIndex((o: { id: string }) => o.id === optionId)
    if (idx >= 0) {
      actions.submitSelect([idx])
    } else {
      if (!canCancelPrompt.value) {
        showPromptError('当前步骤不可取消，请先完成本次操作')
        return
      }
      actions.submitCancel()
    }
    return
  }
  if (optionId === 'skip' || optionId === 'cancel') {
    if (!canCancelPrompt.value) {
      showPromptError('当前步骤不可取消，请先完成本次操作')
      return
    }
    actions.submitCancel()
    return
  }
  if (optionId === 'refuse') {
    // 魔爆冲击“不弃牌”是规则内显式选项，直接走 Cancel 语义。
    actions.submitCancel()
    return
  }
  if (optionId === 'confirm') {
    actions.submitConfirm()
    return
  }
  // 魔弹融合等确认选项：yes=0, no=1
  if (optionId === 'yes' || optionId === 'no') {
    actions.submitSelect([optionId === 'yes' ? 0 : 1])
    return
  }
  // 魔弹掌控方向选择：normal=0, reverse=1
  if (optionId === 'normal' || optionId === 'reverse') {
    actions.submitSelect([optionId === 'normal' ? 0 : 1])
    return
  }
  if (optionId === 'take') {
    actions.submitRespondTake()
    return
  }
  if (optionId === 'counter') {
    if (interruptStore.selectedCards.length === 0) {
      showPromptError(isMagicMissilePrompt.value ? '请先选择一张【魔弹】进行传递' : '请先选择一张攻击牌进行应战')
      return
    }
    if (needsCounterTargetSelection.value && !interruptStore.promptCounterTarget) {
      showPromptError('请先选择反弹目标（攻击方的队友）')
      return
    }
    if (!actions.submitRespondCounter(isMagicMissilePrompt.value)) return
    return
  }
  if (optionId === 'defend') {
    if (interruptStore.selectedCards.length === 0) {
      showPromptError('请先选择一张【圣光】进行防御（圣盾需提前放置）')
      return
    }
    if (!actions.submitRespondDefend()) return
    return
  }
  if (isNonHandChooseCardsMultiMode.value && isNonHandChooseCardOption(optionId)) {
    toggleInlineCardOption(optionId)
    return
  }
  {
    const optionIndex = prompt.value?.options?.findIndex((o: { id: string }) => o.id === optionId) ?? -1
    if (optionIndex >= 0) {
      actions.submitSelect([optionIndex])
    } else {
      const index = parseInt(optionId, 10)
      if (!Number.isNaN(index)) {
        actions.submitSelect([index])
      } else {
        actions.submitAction({
          player_id: myPlayerId.value,
          type: 'Select',
          skill_id: optionId
        })
      }
    }
  }
}

const canConfirmPrompt = computed(() => {
  if (!prompt.value) return false
  if (prompt.value.type === 'choose_target') {
    const tCount = interruptStore.selectedTargets.length
    return tCount >= prompt.value.min && tCount <= prompt.value.max
  }
  if (prompt.value.type === 'choose_card' || prompt.value.type === 'choose_cards') {
    const cCount = isNonHandChooseCardsMultiMode.value
      ? selectedInlineCardOptionIndices.value.length
      : interruptStore.selectedCards.length
    return cCount >= prompt.value.min && cCount <= prompt.value.max
  }
  return true
})

function confirmPromptAction() {
  if (!canConfirmPrompt.value) return

  if (prompt.value?.type === 'choose_target' && interruptStore.selectedTargets.length > 0) {
    if (interruptStore.selectedTargets.length === 1) {
      const targetId = interruptStore.selectedTargets[0]
      if (!targetId) return
      actions.submitPromptTarget(targetId)
    } else {
      actions.submitAction({
        player_id: myPlayerId.value,
        type: 'Select',
        target_ids: interruptStore.selectedTargets
      })
    }
    return
  }

  const indices = isNonHandChooseCardsMultiMode.value
    ? selectedInlineCardOptionIndices.value
    : interruptStore.selectedCards
  if (indices.length > 0) {
    actions.submitSelect(indices)
  }
}

function parsePromptCardIndex(optionId: string): number | null {
  const normalized = String(optionId || '').trim()
  if (!/^-?\d+$/.test(normalized)) return null
  const parsed = Number.parseInt(normalized, 10)
  if (!Number.isFinite(parsed)) return null
  return parsed
}

function parseHandIndexFromOptionLabel(label: string): number | null {
  const matched = String(label || '').trim().match(/^(\d+)\s*:/)
  if (!matched) return null
  const displayIndex = Number.parseInt(matched[1] || '', 10)
  if (!Number.isFinite(displayIndex) || displayIndex <= 0) return null
  return displayIndex - 1
}

function parseCocoonFieldIndexFromOptionLabel(label: string): number | null {
  const matched = String(label || '').match(/茧\[(\d+)\]/)
  if (!matched) return null
  const parsed = Number.parseInt(matched[1] || '', 10)
  if (!Number.isFinite(parsed) || parsed < 0) return null
  return parsed
}

function isIndexedCocoonOption(option: { label?: string }): boolean {
  return parseCocoonFieldIndexFromOptionLabel(String(option.label || '')) !== null
}

function isPromptHandCardOption(option: { id: string; label: string }): boolean {
  const idx = parsePromptCardIndex(option.id)
  if (idx === null || idx < 0 || idx >= myHand.value.length) return false
  const labelIndex = parseHandIndexFromOptionLabel(option.label)
  return labelIndex === idx
}

const promptCardOptionIndexSet = computed(() => {
  const set = new Set<number>()
  if (!prompt.value?.options?.length) return set
  for (const option of prompt.value.options) {
    if (!isPromptHandCardOption(option)) continue
    const idx = parsePromptCardIndex(option.id)
    if (idx !== null) set.add(idx)
  }
  return set
})

const hasIndexedCocoonOptions = computed(() => {
  if (!prompt.value?.options?.length) return false
  return prompt.value.options.some((option) => isIndexedCocoonOption(option))
})

const isNonHandChooseCardsMultiMode = computed(() => {
  if (!prompt.value) return false
  if (prompt.value.type !== 'choose_cards') return false
  if (hasCounterOrDefend.value) return false
  if (promptCardOptionIndexSet.value.size > 0) return false
  if (!prompt.value.options?.length) return false
  if ((prompt.value.max ?? 1) <= 1) return false
  return prompt.value.options.every((option) => parsePromptCardIndex(option.id) !== null)
})

function isNonHandChooseCardOption(optionId: string): boolean {
  if (!isNonHandChooseCardsMultiMode.value || !prompt.value?.options?.length) return false
  const idx = parsePromptCardIndex(optionId)
  if (idx === null) return false
  return prompt.value.options.some((option) => option.id === optionId)
}

function toggleInlineCardOption(optionId: string) {
  if (!isNonHandChooseCardOption(optionId)) return
  const idx = parsePromptCardIndex(optionId)
  if (idx === null) return
  const pos = selectedInlineCardOptionIndices.value.indexOf(idx)
  if (pos >= 0) {
    selectedInlineCardOptionIndices.value.splice(pos, 1)
    return
  }
  const max = prompt.value?.max ?? 1
  if (selectedInlineCardOptionIndices.value.length >= max) return
  selectedInlineCardOptionIndices.value.push(idx)
  selectedInlineCardOptionIndices.value.sort((a, b) => a - b)
}

function isInlineCardOptionSelected(optionId: string): boolean {
  if (!isNonHandChooseCardOption(optionId)) return false
  const idx = parsePromptCardIndex(optionId)
  if (idx === null) return false
  return selectedInlineCardOptionIndices.value.includes(idx)
}

type RawDockOption = {
  id: string
  label: string
  button_label?: string
  hint?: string
  disabled?: boolean
}

type DockButtonOption = {
  id: string
  label: string
  buttonLabel: string
  hint: string
  disabled?: boolean
  numeric: boolean
}

type SkillPromptEntry = {
  id: string
  promptText: string
  buttonLabel: string
  disabled: boolean
}

type SkillPromptButton = {
  id: string
  label: string
  disabled: boolean
  cancel: boolean
}

type FraudElementCardOption = {
  id: string
  title: string
  glyph: string
  tone: string
}

type PromptImageButtonKind = 'take' | 'counter' | 'defend' | 'cancel' | 'confirm' | 'card' | 'action'

const PROMPT_IMAGE_BUTTON_CANDIDATES: Record<PromptImageButtonKind, string[]> = {
  take: ['/assets/ui/prompt_btn_take.png'],
  counter: ['/assets/ui/prompt_btn_counter.png'],
  defend: ['/assets/ui/prompt_btn_defend.png'],
  cancel: ['/assets/ui/action_cancel_btn.png'],
  confirm: ['/assets/ui/action_confirm.png'],
  card: ['/assets/ui/action_card.png'],
  action: ['/assets/ui/action_special_btn.png'],
}

const promptImageButtonIndex = ref<Record<PromptImageButtonKind, number>>({
  take: 0,
  counter: 0,
  defend: 0,
  cancel: 0,
  confirm: 0,
  card: 0,
  action: 0,
})

const promptImageButtonFailed = ref<Record<PromptImageButtonKind, boolean>>({
  take: false,
  counter: false,
  defend: false,
  cancel: false,
  confirm: false,
  card: false,
  action: false,
})

const optionButtonLabelById: Record<string, string> = {
  confirm: '发动',
  yes: '发动',
  no: '取消',
  cancel: '取消',
  skip: '取消',
  take: '命中',
  counter: '应战',
  defend: '防御',
  normal: '顺时针',
  reverse: '逆时针',
  refuse: '不弃牌',
  cannot_act: '取消',
  pass: '取消',
}

const plainNoHintButtons = new Set(['发动', '确认', '确定', '是', '取消', '不弃牌', '应战', '防御', '命中', '顺序', '反向'])

function parseNonNegativeOptionId(optionId: string): number | null {
  const normalized = String(optionId || '').trim()
  if (!/^\d+$/.test(normalized)) return null
  const value = Number.parseInt(normalized, 10)
  if (!Number.isFinite(value) || value < 0) return null
  return value
}

function shouldUseNumericButtonMode(options: RawDockOption[]): { useNumeric: boolean; plusOne: boolean } {
  if (!prompt.value || options.length < 2) return { useNumeric: false, plusOne: false }
  if (prompt.value.type === 'choose_card' || prompt.value.type === 'choose_cards') return { useNumeric: false, plusOne: false }
  const numericIds: number[] = []
  let hasLongLabel = false
  let hasXHint = /[xXＸ]/.test(String(prompt.value.message || ''))
  for (const option of options) {
    const n = parseNonNegativeOptionId(option.id)
    if (n !== null) numericIds.push(n)
    const label = String(option.label || '').trim()
    if (label.length >= 8 || label.includes('分支')) hasLongLabel = true
    if (/[xXＸ]\s*=/.test(label) || /[xXＸ]值/.test(label) || /[xXＸ]/.test(label)) hasXHint = true
  }
  if (numericIds.length < 2 || (!hasLongLabel && !hasXHint)) {
    return { useNumeric: false, plusOne: false }
  }
  const minNumeric = Math.min(...numericIds)
  return { useNumeric: true, plusOne: minNumeric === 0 }
}

function isDeclineLabel(label: string): boolean {
  const text = String(label || '').trim()
  if (!text) return false
  const compact = text.replace(/\s+/g, '')
  const lower = compact.toLowerCase()
  if (compact.includes('不发动') || compact.includes('无法行动') || compact.includes('不弃牌')) return true
  if (compact.startsWith('放弃') || compact.startsWith('跳过') || compact.startsWith('拒绝')) return true
  if (compact === '取消' || compact.startsWith('取消并') || compact.startsWith('取消本次') || compact.startsWith('取消行动')) return true
  if (lower === 'cancel' || lower === 'pass' || lower === 'skip' || lower === 'refuse') return true
  return false
}

function isConfirmLikeLabel(label: string): boolean {
  const text = String(label || '').trim().replace(/\s+/g, '')
  if (!text) return false
  if (text === '是' || text === '发动' || text === '确认' || text === '确定') return true
  if (text.startsWith('发动') || text.startsWith('确认') || text.startsWith('确定')) return true
  return false
}

function isCardSelectionLikeText(text: string): boolean {
  const normalized = String(text || '').trim()
  if (!normalized) return false
  if (/^\d+\s*[:：]/.test(normalized)) return true
  if (/第\d+张\s*[:：]/.test(normalized)) return true
  if (/^茧\[\d+\]\s*[:：]/.test(normalized)) return true
  if (/^移除茧\[\d+\]\s*[:：]/.test(normalized)) return true
  return false
}

function promptImageButtonKindByOption(option: { id?: string; label?: string; buttonLabel?: string }): PromptImageButtonKind {
  const id = String(option.id || '').trim().toLowerCase()
  const label = String(option.label || '').trim()
  const buttonLabel = String(option.buttonLabel || '').trim()
  const combinedText = `${label} ${buttonLabel}`
  const hasExplicitResponseText =
    combinedText.includes('命中') ||
    combinedText.includes('承受') ||
    combinedText.includes('防御') ||
    combinedText.includes('应战') ||
    combinedText.includes('传递')
  if (isCardSelectionLikeText(label) || isCardSelectionLikeText(buttonLabel)) {
    return 'card'
  }
  if ((isConfirmLikeLabel(buttonLabel) || isConfirmLikeLabel(label)) && !hasExplicitResponseText) {
    return 'confirm'
  }
  const responseKind = responseOptionKind({ id, label, button_label: buttonLabel })
  if (responseKind) return responseKind
  if (id === 'confirm' || id === 'yes') return 'confirm'
  if (id === 'skip' || id === 'cancel' || id === 'no' || id === 'pass' || id === 'cannot_act') return 'cancel'
  if (buttonLabel === '取消' || isDeclineLabel(buttonLabel) || isDeclineLabel(label)) return 'cancel'
  return 'action'
}

function promptImageButtonAsset(kind: PromptImageButtonKind): string {
  const candidates = PROMPT_IMAGE_BUTTON_CANDIDATES[kind]
  const index = promptImageButtonIndex.value[kind]
  return candidates[Math.min(index, candidates.length - 1)] || ''
}

function isPromptImageButtonReady(kind: PromptImageButtonKind): boolean {
  return !promptImageButtonFailed.value[kind]
}

function onPromptImageButtonError(kind: PromptImageButtonKind) {
  const candidates = PROMPT_IMAGE_BUTTON_CANDIDATES[kind]
  const nextIndex = promptImageButtonIndex.value[kind] + 1
  if (nextIndex < candidates.length) {
    promptImageButtonIndex.value[kind] = nextIndex
    return
  }
  promptImageButtonFailed.value[kind] = true
}

function promptImageButtonFallbackText(kind: PromptImageButtonKind | null, buttonLabel: string = ''): string {
  if (kind === 'take') return '命'
  if (kind === 'defend') return '防'
  if (kind === 'counter') return '应'
  if (kind === 'cancel') return '消'
  if (kind === 'confirm') return '确'
  if (kind === 'card') return '牌'
  if (kind === 'action') return buttonLabel ? buttonLabel.charAt(0) : '动'
  return ''
}

function dockButtonImageKind(option: DockButtonOption): PromptImageButtonKind | null {
  if (option.numeric) return null
  return promptImageButtonKindByOption({
    id: option.id,
    label: option.label,
    buttonLabel: option.buttonLabel
  })
}

function isDockButtonImageStyle(option: DockButtonOption): boolean {
  return dockButtonImageKind(option) !== null
}

function dockButtonImageSrc(option: DockButtonOption): string {
  const kind = dockButtonImageKind(option)
  return kind ? promptImageButtonAsset(kind) : ''
}

function isDockButtonImageReady(option: DockButtonOption): boolean {
  const kind = dockButtonImageKind(option)
  if (!kind) return false
  return isPromptImageButtonReady(kind)
}

function onDockButtonImageError(option: DockButtonOption) {
  const kind = dockButtonImageKind(option)
  if (!kind) return
  onPromptImageButtonError(kind)
}

function dockButtonFallbackText(option: DockButtonOption): string {
  return promptImageButtonFallbackText(dockButtonImageKind(option), option.buttonLabel)
}

function isPromptConfirmImageReady(): boolean {
  return isPromptImageButtonReady('confirm')
}

function promptConfirmImageSrc(): string {
  return promptImageButtonAsset('confirm')
}

function onPromptConfirmImageError() {
  onPromptImageButtonError('confirm')
}

function skillButtonImageSrc(option: SkillPromptButton): string {
  return option.cancel ? promptImageButtonAsset('cancel') : promptConfirmImageSrc()
}

function isSkillButtonImageReady(option: SkillPromptButton): boolean {
  return option.cancel ? isPromptImageButtonReady('cancel') : isPromptConfirmImageReady()
}

function onSkillButtonImageError(option: SkillPromptButton) {
  if (option.cancel) {
    onPromptImageButtonError('cancel')
    return
  }
  onPromptConfirmImageError()
}

function skillButtonFallbackText(option: SkillPromptButton): string {
  return option.cancel ? '消' : '确'
}

function normalizeButtonLabel(rawLabel: string, optionId: string, optionLabel: string, responseKind: ResponseOptionKind): string {
  const text = String(rawLabel || '').trim()
  const lowerId = String(optionId || '').trim().toLowerCase()
  if (responseKind === 'take' || lowerId === 'take' || lowerId === 'take_damage' || text.includes('承受') || text.includes('命中')) {
    return '命中'
  }
  if (responseKind === 'counter' || lowerId === 'counter' || text.includes('应战')) {
    return '应战'
  }
  if (responseKind === 'defend' || lowerId === 'defend' || text.includes('防御')) {
    return '防御'
  }
  if (
    lowerId === 'cancel' ||
    lowerId === 'skip' ||
    lowerId === 'refuse' ||
    lowerId === 'no' ||
    lowerId === 'pass' ||
    lowerId === 'cannot_act' ||
    isDeclineLabel(text) ||
    isDeclineLabel(optionLabel)
  ) {
    return '取消'
  }
  return text
}

function normalizeDockOption(option: RawDockOption, useNumeric: boolean, plusOne: boolean): DockButtonOption {
  const id = String(option.id || '').trim()
  const label = String(option.label || '').trim()
  const lowerID = id.toLowerCase()
  const responseKind = responseOptionKind(option)
  let buttonLabel = normalizeButtonLabel(String(option.button_label || ''), id, label, responseKind)
  let hint = String(option.hint || '').trim()

  if (!buttonLabel && optionButtonLabelById[lowerID]) {
    buttonLabel = optionButtonLabelById[lowerID]
  }
  if (!buttonLabel && prompt.value?.type === 'choose_skill') {
    buttonLabel = '发动'
  }
  if (!buttonLabel && lowerID === '-1') {
    buttonLabel = label.includes('完成') || label.includes('结束') ? '完成' : '取消'
  }
  if (!buttonLabel && useNumeric) {
    const n = parseNonNegativeOptionId(id)
    if (n !== null) {
      buttonLabel = String(plusOne ? n + 1 : n)
    }
  }
  if (!buttonLabel && isCardSelectionLikeText(label)) {
    buttonLabel = '打出卡牌'
  }
  if (!buttonLabel && isDeclineLabel(label)) {
    buttonLabel = '取消'
  }
  if (!buttonLabel && responseKind === 'take') {
    buttonLabel = '命中'
  }
  if (!buttonLabel && responseKind === 'defend') {
    buttonLabel = '防御'
  }
  if (!buttonLabel && responseKind === 'counter') {
    buttonLabel = '应战'
  }
  if (!buttonLabel) {
    if (prompt.value?.type === 'confirm') {
      buttonLabel = '确认'
    } else {
      buttonLabel = label && label.length <= 6 ? label : '执行'
    }
  }

  if (responseOptionKind({ id, label, button_label: buttonLabel }) !== null) {
    hint = ''
  }

  if (!hint && label && label !== buttonLabel) {
    if (!(plainNoHintButtons.has(buttonLabel) && (label === buttonLabel || isDeclineLabel(label)))) {
      hint = label
    }
  }

  return {
    id,
    label,
    buttonLabel,
    hint,
    disabled: option.disabled,
    numeric: /^\d+$/.test(buttonLabel),
  }
}

function buildDockButtons(options: RawDockOption[]): DockButtonOption[] {
  if (options.length === 0) return []
  const mode = shouldUseNumericButtonMode(options)
  return options.map((option) => normalizeDockOption(option, mode.useNumeric, mode.plusOne))
}

const fraudElementCardMetaById: Record<string, Omit<FraudElementCardOption, 'id' | 'title'>> = {
  water: { glyph: '水', tone: 'prompt-fraud-card--water' },
  fire: { glyph: '火', tone: 'prompt-fraud-card--fire' },
  earth: { glyph: '地', tone: 'prompt-fraud-card--earth' },
  wind: { glyph: '风', tone: 'prompt-fraud-card--wind' },
  thunder: { glyph: '雷', tone: 'prompt-fraud-card--thunder' },
}

function fraudAttackCardName(optionId: string, fallback: string): string {
  const lower = String(optionId || '').trim().toLowerCase()
  if (lower === 'water') return '水涟斩'
  if (lower === 'fire') return '火焰斩'
  if (lower === 'earth') return '地裂斩'
  if (lower === 'wind') return '风神斩'
  if (lower === 'thunder') return '雷光斩'
  const normalizedFallback = String(fallback || '').trim()
  if (normalizedFallback) return normalizedFallback
  return String(optionId || '').trim() || '系别'
}

const isFraudElementCardPickerPrompt = computed(() =>
  prompt.value?.choice_type === 'adventurer_fraud_attack_element' && (prompt.value?.options?.length ?? 0) > 0
)

const fraudElementCardOptions = computed<FraudElementCardOption[]>(() => {
  if (!isFraudElementCardPickerPrompt.value || !prompt.value?.options?.length) return []
  return prompt.value.options.map((option) => {
    const lower = String(option.id || '').trim().toLowerCase()
    const meta = fraudElementCardMetaById[lower] ?? {
      glyph: '系',
      tone: 'prompt-fraud-card--generic',
    }
    return {
      id: option.id,
      title: fraudAttackCardName(option.id, option.label),
      glyph: meta.glyph,
      tone: meta.tone,
    }
  })
})

const cardFooterOptions = computed<RawDockOption[]>(() => {
  if (!prompt.value?.options || !needsCardSelection.value) return []
  if (hasCounterOrDefend.value) {
    const responseOrder: Record<Exclude<ResponseOptionKind, null>, number> = {
      take: 0,
      defend: 1,
      counter: 2,
    }
    const responseRank = (kind: ResponseOptionKind): number => {
      if (!kind) return 99
      return responseOrder[kind]
    }
    return prompt.value.options
      .filter((option: { id: string; label: string; button_label?: string }) => responseOptionKind(option) !== null)
      .sort((a, b) => {
        const rankA = responseRank(responseOptionKind(a))
        const rankB = responseRank(responseOptionKind(b))
        return rankA - rankB
      })
      .map((option) => ({
        id: option.id,
        label: option.label,
        button_label: option.button_label,
        hint: option.hint,
        disabled: false
      }))
  }
  if (prompt.value.type !== 'choose_card' && prompt.value.type !== 'choose_cards') return []
  // 选牌类提示统一在手牌区完成，行动区不再展示“选哪张牌”的按钮。
  return []
})

const promptNeedsHandCardConfirm = computed(() => {
  if (!prompt.value || !needsCardSelection.value || hasCounterOrDefend.value) return false
  if (isNonHandChooseCardsMultiMode.value) return false
  if (promptCardOptionIndexSet.value.size > 0) return true
  return !prompt.value.options?.length
})

const promptNeedsInlineCardOptionConfirm = computed(() =>
  isNonHandChooseCardsMultiMode.value && !hasIndexedCocoonOptions.value
)

const promptNeedsCardConfirm = computed(() =>
  promptNeedsHandCardConfirm.value || promptNeedsInlineCardOptionConfirm.value
)

const cardConfirmHintText = computed(() => {
  if (promptNeedsInlineCardOptionConfirm.value) return '完成选择后点击发动'
  if (prompt.value?.choice_type === 'adventurer_fraud_pick') return '请选择2~3张同系牌，3张将自动转为暗灭攻击'
  return '完成选牌后点击发动'
})

const cardConfirmPromptMessage = computed(() => {
  const message = String(prompt.value?.message || '').trim()
  if (message) return message
  return cardConfirmHintText.value
})

const showCardConfirmCancelRow = computed(() =>
  promptNeedsCardConfirm.value && canCancelPrompt.value && !isSkillChoicePrompt.value
)

const inlinePrimaryButtons = computed<DockButtonOption[]>(() => {
  if (isExtractPrompt.value) return []
  if (isFraudElementCardPickerPrompt.value) return []
  if (needsCardSelection.value) return buildDockButtons(cardFooterOptions.value)
  if (showConfirmButtonSection.value) {
    const options = nonPlayerOptions.value
      .filter((option) => option.id !== 'cancel' && option.id !== 'skip')
      .filter((option) => !isIndexedCocoonOption(option))
      .map((option) => ({
      id: option.id,
      label: option.label,
      button_label: option.button_label,
      hint: option.hint,
      disabled: false
      }))
    return buildDockButtons(options)
  }
  return []
})

const isSkillChoicePrompt = computed(() => {
  if (!prompt.value) return false
  return prompt.value.type === 'choose_skill' || isResponseSkillConfirmPrompt.value
})

function parseSkillTitle(option: DockButtonOption, index: number): string {
  const rawLabel = String(option.label || '').trim()
  let title = rawLabel || `技能 ${index + 1}`

  // 兼容旧服务端：若 label 仍为“标题：说明”，前端兜底只截标题。
  const separatorIndex = rawLabel.indexOf('：')
  if (separatorIndex > 0) {
    const parsedTitle = rawLabel.slice(0, separatorIndex).trim()
    if (parsedTitle) title = parsedTitle
  }

  // 去掉前缀序号与尾部消耗标记，按钮中尽量只保留技能名。
  title = title.replace(/^\d+\s*[.)、]\s*/, '').trim()
  title = title.replace(/\s*\[[^\]]+\]\s*$/, '').trim()

  return title
}

const skillPromptEntries = computed<SkillPromptEntry[]>(() => {
  if (!isSkillChoicePrompt.value || inlinePrimaryButtons.value.length === 0) return []
  return inlinePrimaryButtons.value.map((option, index) => {
    const title = parseSkillTitle(option, index)
    return {
      id: option.id,
      promptText: `是否发动【${title}】`,
      buttonLabel: option.buttonLabel || '发动',
      disabled: !!option.disabled
    }
  })
})

const skillPromptTitle = computed(() => {
  if (!isSkillChoicePrompt.value || skillPromptEntries.value.length === 0) return ''
  if (skillPromptEntries.value.length === 1) return skillPromptEntries.value[0]?.promptText || ''
  const message = String(prompt.value?.message || '').trim()
  return message || '请选择要发动的技能'
})

const skillPromptButtons = computed<SkillPromptButton[]>(() => {
  if (!isSkillChoicePrompt.value || skillPromptEntries.value.length === 0) return []
  const skillCount = skillPromptEntries.value.length
  const buttons: SkillPromptButton[] = skillPromptEntries.value.map((entry, index) => {
    let label = entry.buttonLabel || '发动'
    if (prompt.value?.type === 'choose_skill' && skillCount > 1) {
      const option = inlinePrimaryButtons.value[index]
      label = option ? parseSkillTitle(option, index) : `技能 ${index + 1}`
    } else if (skillCount > 1 && (label === '发动' || label === '确认')) {
      label = String(index + 1)
    }
    return {
      id: entry.id,
      label,
      disabled: !!entry.disabled,
      cancel: false
    }
  })

  const hasCancelLike = buttons.some((btn) => btn.id === 'cancel' || btn.id === 'skip' || btn.cancel)
  if (canCancelPrompt.value && !hasCancelLike) {
    buttons.push({
      id: 'cancel',
      label: '取消',
      disabled: false,
      cancel: true
    })
  }
  return buttons
})

const isMultiSkillNameChoiceMode = computed(() =>
  prompt.value?.type === 'choose_skill' && skillPromptEntries.value.length > 1
)

const inlinePrimaryGridClass = computed(() => {
  const count = inlinePrimaryButtons.value.length
  if (count <= 1) return 'prompt-inline-grid--1'
  if (count === 2) return 'prompt-inline-grid--2'
  if (count === 3) return 'prompt-inline-grid--3'
  return 'prompt-inline-grid--4'
})

function buildPromptAutoResolveKey(p: NonNullable<typeof prompt.value>): string {
  const options = (p.options || [])
    .map((option) => `${option.id}|${option.label}|${option.button_label || ''}|${option.hint || ''}`)
    .join('||')
  return `${p.type}::${p.player_id}::${p.message}::${options}`
}

const autoResolveOptionId = computed(() => {
  if (!isVisible.value || !prompt.value) return ''
  if (isExtractPrompt.value || isSkillChoicePrompt.value) return ''
  if (hasIndexedCocoonOptions.value) return ''
  // 存在角色目标选项时禁止自动确认，避免“完成选择”类按钮被误触发。
  if (playerOptionEntries.value.length > 0) return ''
  if (needsCardSelection.value || needsTargetSelection.value || needsCounterTargetSelection.value) return ''
  // 有取消/跳过时表示存在真实分支，不做自动确认。
  if (canCancelPrompt.value) return ''
  const candidates = inlinePrimaryButtons.value.filter((option) => !option.disabled)
  if (candidates.length !== 1) return ''
  const onlyOption = candidates[0]
  if (!onlyOption) return ''
  return onlyOption.id
})

const hasAnyInlineButton = computed(() => {
  if (!isVisible.value) return false
  if (isFraudElementCardPickerPrompt.value) return false
  if (isExtractPrompt.value && !!prompt.value?.options?.length) return true
  if (inlinePrimaryButtons.value.length > 0) return true
  if (promptNeedsCardConfirm.value) return true
  if (canCancelPrompt.value) return true
  return false
})

const cancelDockButton = computed<DockButtonOption>(() => {
  const promptOptions = prompt.value?.options ?? []
  const refuseOption = promptOptions.find((option) => option.id === 'refuse')
  const cancelOption = promptOptions.find((option) => option.id === 'cancel')
  const skipOption = promptOptions.find((option) => option.id === 'skip')
  const option = refuseOption ?? cancelOption ?? skipOption ?? {
    id: 'cancel',
    label: canCancelPrompt.value ? '取消' : ''
  }
  return normalizeDockOption(
    {
      id: option.id,
      label: option.label,
      button_label: option.button_label,
      hint: option.hint
    },
    false,
    false
  )
})

function getDockButtonClass(optionId: string): string {
  const lowerOptionId = String(optionId || '').trim().toLowerCase()
  const kind = responseOptionKind({ id: lowerOptionId })
  if (kind === 'take') return 'prompt-inline-btn--take'
  if (kind === 'counter') return 'prompt-inline-btn--counter'
  if (kind === 'defend') return 'prompt-inline-btn--defend'
  if (lowerOptionId === 'confirm' || lowerOptionId === 'yes') return 'prompt-inline-btn--success'
  if (lowerOptionId === 'skip' || lowerOptionId === 'cancel' || lowerOptionId === 'refuse' || lowerOptionId === 'no' || lowerOptionId === 'pass' || lowerOptionId === 'cannot_act') {
    return 'prompt-inline-btn--cancel'
  }
  return 'prompt-inline-btn--normal'
}

function shouldHideOptionHint(option: DockButtonOption): boolean {
  return responseOptionKind({ id: option.id, label: option.label, button_label: option.buttonLabel }) !== null
}

function shouldEnlargeResponseActionButton(option: DockButtonOption): boolean {
  if (!isDarkAttackResponsePrompt.value) return false
  const kind = responseOptionKind({ id: option.id, label: option.label, button_label: option.buttonLabel })
  return kind === 'take' || kind === 'defend'
}

watch(autoResolveOptionId, (optionId) => {
  if (!optionId || !prompt.value) return
  const key = buildPromptAutoResolveKey(prompt.value)
  if (autoResolvedPromptKey.value === key) return
  autoResolvedPromptKey.value = key
  handleOptionClick(optionId)
})
</script>

<template>
  <Transition name="prompt-inline-pop">
    <div v-if="hasAnyInlineButton" class="prompt-inline-root">
      <div class="prompt-inline-surface">
        <template v-if="isExtractPrompt && prompt?.options?.length">
          <div class="prompt-inline-grid prompt-inline-grid--2">
            <button
              v-for="(option, idx) in prompt.options"
              :key="option.id"
              class="prompt-inline-btn prompt-inline-btn--extract"
              :class="{ 'prompt-inline-btn--selected': selectedExtractIndices.includes(idx) }"
              @click="toggleExtractOption(idx)"
            >
              {{ option.label === '红宝石' ? '♦ 红宝石' : '🔷 蓝水晶' }}
            </button>
          </div>
          <div class="flex justify-center mt-2">
            <button
              class="prompt-inline-btn prompt-inline-btn--success action-image-btn"
              :class="{ 'prompt-inline-btn--disabled': selectedExtractIndices.length < (prompt?.min ?? 1) || selectedExtractIndices.length > (prompt?.max ?? 2) }"
              :disabled="selectedExtractIndices.length < (prompt?.min ?? 1) || selectedExtractIndices.length > (prompt?.max ?? 2)"
              @click="confirmExtractSelection"
              :title="`确认提炼（${selectedExtractIndices.length}/${prompt?.max ?? 2}）`"
              :aria-label="`确认提炼（${selectedExtractIndices.length}/${prompt?.max ?? 2}）`"
            >
              <img
                v-if="isPromptConfirmImageReady()"
                class="action-image-btn-fill"
                :src="promptConfirmImageSrc()"
                alt=""
                @error="onPromptConfirmImageError"
              />
              <span v-else class="action-image-fallback-text">确</span>
            </button>
          </div>
        </template>

        <template v-else>
          <div v-if="isSkillChoicePrompt && skillPromptButtons.length > 0" class="prompt-skill-list">
            <div class="prompt-skill-row">
              <div class="prompt-skill-text" :title="skillPromptTitle">{{ skillPromptTitle }}</div>
              <div class="prompt-skill-actions">
                <button
                  v-for="option in skillPromptButtons"
                  :key="option.id"
                  class="prompt-inline-btn prompt-skill-action"
                  :class="[
                    isMultiSkillNameChoiceMode ? 'prompt-inline-btn--normal prompt-skill-action--plain' : '',
                    !isMultiSkillNameChoiceMode ? 'action-image-btn' : '',
                    !isMultiSkillNameChoiceMode ? (option.cancel ? 'prompt-inline-btn--cancel' : 'prompt-inline-btn--success') : '',
                    option.disabled ? 'prompt-inline-btn--disabled' : ''
                  ]"
                  :disabled="option.disabled"
                  :title="!isMultiSkillNameChoiceMode ? option.label : undefined"
                  :aria-label="!isMultiSkillNameChoiceMode ? option.label : undefined"
                  @click="handleOptionClick(option.id)"
                >
                  <template v-if="!isMultiSkillNameChoiceMode">
                    <img
                      v-if="isSkillButtonImageReady(option)"
                      class="action-image-btn-fill"
                      :src="skillButtonImageSrc(option)"
                      alt=""
                      @error="onSkillButtonImageError(option)"
                    />
                    <span v-else class="action-image-fallback-text">{{ skillButtonFallbackText(option) }}</span>
                  </template>
                  <template v-else>
                    {{ option.label }}
                  </template>
                </button>
              </div>
            </div>
          </div>

          <div v-else-if="inlinePrimaryButtons.length > 0">
            <div v-if="responseAttackElementHintText" class="prompt-inline-hint prompt-inline-hint--attack-element">
              {{ responseAttackElementHintText }}
            </div>
            <div class="prompt-inline-grid" :class="inlinePrimaryGridClass">
            <div
              v-for="option in inlinePrimaryButtons"
              :key="option.id"
              class="prompt-inline-entry"
              :class="{ 'prompt-inline-entry--hinted-numeric': option.numeric && !!option.hint }"
            >
              <div
                v-if="option.hint && !shouldHideOptionHint(option)"
                class="prompt-inline-hint"
                :class="{ 'prompt-inline-hint--hinted-numeric': option.numeric }"
              >
                {{ option.hint }}
              </div>
              <button
                class="prompt-inline-btn"
                :class="[
                  isDockButtonImageStyle(option) ? 'action-image-btn' : '',
                  getDockButtonClass(option.id),
                  shouldEnlargeResponseActionButton(option) ? 'prompt-inline-btn--response-large' : '',
                  option.numeric ? 'prompt-inline-btn--numeric' : '',
                  option.numeric && !!option.hint ? 'prompt-inline-btn--hinted-numeric' : '',
                  isInlineCardOptionSelected(option.id) ? 'prompt-inline-btn--selected' : '',
                  option.disabled ? 'prompt-inline-btn--disabled' : ''
                ]"
                :disabled="!!option.disabled"
                :title="isDockButtonImageStyle(option) ? option.buttonLabel : undefined"
                :aria-label="isDockButtonImageStyle(option) ? option.buttonLabel : undefined"
                @click="handleOptionClick(option.id)"
              >
                <template v-if="isDockButtonImageStyle(option)">
                  <img
                    v-if="isDockButtonImageReady(option)"
                    class="action-image-btn-fill"
                    :src="dockButtonImageSrc(option)"
                    alt=""
                    @error="onDockButtonImageError(option)"
                  />
                  <span v-else class="action-image-fallback-text">{{ dockButtonFallbackText(option) }}</span>
                </template>
                <template v-else>
                  <template v-if="option.numeric && !!option.hint">
                    <span class="prompt-inline-btn-number">{{ option.buttonLabel }}</span>
                    <span class="prompt-inline-btn-caption">选择</span>
                  </template>
                  <template v-else>
                    {{ option.buttonLabel }}
                  </template>
                </template>
              </button>
            </div>
          </div>
          </div>

          <div v-if="showCardConfirmCancelRow" class="prompt-inline-entry">
            <div class="prompt-inline-hint">{{ cardConfirmPromptMessage }}</div>
            <div class="prompt-inline-actions-row">
              <button
                class="prompt-inline-btn prompt-inline-btn--success action-image-btn"
                :class="{ 'prompt-inline-btn--disabled': !canConfirmPrompt }"
                :disabled="!canConfirmPrompt"
                title="发动"
                aria-label="发动"
                @click="confirmPromptAction"
              >
                <img
                  v-if="isPromptConfirmImageReady()"
                  class="action-image-btn-fill"
                  :src="promptConfirmImageSrc()"
                  alt=""
                  @error="onPromptConfirmImageError"
                />
                <span v-else class="action-image-fallback-text">确</span>
              </button>
              <button
                class="prompt-inline-btn prompt-inline-btn--cancel action-image-btn"
                :title="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
                :aria-label="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
                @click="handleOptionClick(cancelDockButton.id)"
              >
                <img
                  v-if="isDockButtonImageReady(cancelDockButton)"
                  class="action-image-btn-fill"
                  :src="dockButtonImageSrc(cancelDockButton)"
                  alt=""
                  @error="onDockButtonImageError(cancelDockButton)"
                />
                <span v-else class="action-image-fallback-text">{{ dockButtonFallbackText(cancelDockButton) }}</span>
              </button>
            </div>
          </div>

          <div v-else-if="promptNeedsCardConfirm" class="prompt-inline-entry">
            <div class="prompt-inline-hint">{{ cardConfirmHintText }}</div>
            <button
              class="prompt-inline-btn prompt-inline-btn--success action-image-btn"
              :class="{ 'prompt-inline-btn--disabled': !canConfirmPrompt }"
              :disabled="!canConfirmPrompt"
              title="发动"
              aria-label="发动"
              @click="confirmPromptAction"
            >
              <img
                v-if="isPromptConfirmImageReady()"
                class="action-image-btn-fill"
                :src="promptConfirmImageSrc()"
                alt=""
                @error="onPromptConfirmImageError"
              />
              <span v-else class="action-image-fallback-text">确</span>
            </button>
          </div>
        </template>

        <div v-if="canCancelPrompt && !isSkillChoicePrompt && !showCardConfirmCancelRow" class="prompt-inline-entry">
          <div v-if="cancelDockButton.hint" class="prompt-inline-hint">{{ cancelDockButton.hint }}</div>
          <button
            class="prompt-inline-btn prompt-inline-btn--cancel"
            :class="isDockButtonImageStyle(cancelDockButton) ? 'action-image-btn' : ''"
            :title="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
            :aria-label="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
            @click="handleOptionClick(cancelDockButton.id)"
          >
            <template v-if="isDockButtonImageStyle(cancelDockButton)">
              <img
                v-if="isDockButtonImageReady(cancelDockButton)"
                class="action-image-btn-fill"
                :src="dockButtonImageSrc(cancelDockButton)"
                alt=""
                @error="onDockButtonImageError(cancelDockButton)"
              />
              <span v-else class="action-image-fallback-text">{{ dockButtonFallbackText(cancelDockButton) }}</span>
            </template>
            <template v-else>
              {{ cancelDockButton.buttonLabel }}
            </template>
          </button>
        </div>
      </div>
    </div>
  </Transition>

  <Teleport to="body">
    <Transition name="prompt-fraud-side-pop">
      <div
        v-if="isFraudElementCardPickerPrompt && fraudElementCardOptions.length > 0"
        class="prompt-fraud-global-layer"
      >
        <div class="prompt-fraud-global-panel">
          <div class="prompt-fraud-dialog prompt-fraud-dialog--global">
            <div class="prompt-fraud-title">{{ prompt?.message || '请选择本次攻击系别' }}</div>
            <div class="prompt-fraud-grid">
              <button
                v-for="option in fraudElementCardOptions"
                :key="option.id"
                class="prompt-fraud-card"
                :class="option.tone"
                :title="option.title"
                :aria-label="option.title"
                @click="handleOptionClick(option.id)"
              >
                <span class="prompt-fraud-card-title-banner">
                  <span class="prompt-fraud-card-title">{{ option.title }}</span>
                </span>
                <span class="prompt-fraud-card-medal">
                  <span>{{ option.glyph }}</span>
                </span>
                <span class="prompt-fraud-card-art">
                  <span class="prompt-fraud-card-glyph">{{ option.glyph }}</span>
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.prompt-inline-root {
  width: 100%;
  display: flex;
  justify-content: center;
  pointer-events: auto;
}

.prompt-inline-surface {
  width: min(760px, 100%);
  border-radius: 14px;
  border: 1px solid rgba(146, 183, 207, 0.36);
  background:
    linear-gradient(180deg, rgba(12, 24, 40, 0.94), rgba(7, 16, 27, 0.96));
  box-shadow:
    0 16px 28px rgba(2, 8, 18, 0.44),
    inset 0 1px 0 rgba(236, 246, 254, 0.1);
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.prompt-inline-grid {
  display: grid;
  gap: 8px;
}

.prompt-skill-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 2px;
}

.prompt-skill-row {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  padding: 2px;
}

.prompt-skill-row + .prompt-skill-row {
  border-top: 1px dashed rgba(138, 171, 192, 0.28);
  padding-top: 9px;
}

.prompt-skill-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  align-items: center;
}

.prompt-skill-text {
  font-size: 13px;
  line-height: 1.3;
  color: rgba(221, 237, 248, 0.94);
  letter-spacing: 0.01em;
  text-align: center;
  white-space: normal;
  word-break: break-word;
}

.prompt-skill-action {
  justify-self: center;
}

.prompt-skill-action--plain {
  justify-self: stretch;
  width: 100%;
  min-height: 42px;
}

.prompt-skill-action:hover:not(:disabled) {
  filter: brightness(1.08);
}

.prompt-fraud-global-layer {
  position: fixed;
  inset: 0;
  z-index: 60;
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(7, 14, 24, 0.38);
  backdrop-filter: blur(7px) saturate(0.98);
  -webkit-backdrop-filter: blur(7px) saturate(0.98);
  padding:
    max(16px, calc(var(--safe-top, 0px) + 8px))
    max(16px, calc(var(--safe-right, 0px) + 8px))
    max(16px, calc(var(--safe-bottom, 0px) + 8px))
    max(16px, calc(var(--safe-left, 0px) + 8px));
}

.prompt-fraud-global-panel {
  pointer-events: auto;
  width: min(860px, calc(100vw - 40px));
  max-height: calc(100vh - 40px);
  overflow: auto;
  border-radius: 14px;
  border: 1px solid rgba(146, 183, 207, 0.42);
  background:
    linear-gradient(180deg, rgba(9, 20, 34, 0.96), rgba(6, 14, 24, 0.97));
  box-shadow:
    0 18px 34px rgba(2, 8, 18, 0.52),
    inset 0 1px 0 rgba(236, 246, 254, 0.12);
  padding: 10px;
}

.prompt-fraud-dialog {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 4px 2px 2px;
}

.prompt-fraud-dialog--global {
  padding: 4px;
}

.prompt-fraud-title {
  font-size: 13px;
  line-height: 1.4;
  color: rgba(225, 238, 249, 0.96);
  text-align: center;
  letter-spacing: 0.01em;
}

.prompt-fraud-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
}

.prompt-fraud-card {
  --fraud-edge-color: rgba(185, 152, 102, 0.78);
  --fraud-edge-glow: rgba(232, 191, 121, 0.38);
  --fraud-base-top: #2f2520;
  --fraud-base-bottom: #17120f;
  --fraud-ribbon-start: #8c5a2f;
  --fraud-ribbon-end: #60401f;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #f1d79b, #c6924f 58%, #784d1d);
  --fraud-medal-fg: #fff7ea;
  --fraud-art-top: rgba(214, 174, 116, 0.3);
  --fraud-art-bottom: rgba(71, 45, 24, 0.82);
  display: flex;
  flex-direction: column;
  align-items: stretch;
  justify-content: flex-start;
  position: relative;
  min-height: 136px;
  border-radius: 10px;
  border: 2px solid var(--fraud-edge-color);
  background: linear-gradient(180deg, var(--fraud-base-top), var(--fraud-base-bottom));
  color: rgba(245, 250, 255, 0.97);
  box-shadow:
    0 8px 16px rgba(0, 0, 0, 0.55),
    0 0 12px var(--fraud-edge-glow),
    inset 0 0 0 1px rgba(255, 244, 214, 0.24),
    inset 0 0 10px rgba(255, 255, 255, 0.08);
  transition: transform 0.16s ease, filter 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
  overflow: hidden;
  text-align: center;
}

.prompt-fraud-card:hover:not(:disabled) {
  transform: translateY(-2px);
  border-color: rgba(255, 228, 170, 0.96);
  box-shadow:
    0 12px 22px rgba(0, 0, 0, 0.62),
    0 0 16px rgba(255, 214, 138, 0.44),
    inset 0 0 0 1px rgba(255, 247, 229, 0.38);
}

.prompt-fraud-card-title-banner {
  position: absolute;
  top: 6px;
  left: 16px;
  right: 16px;
  height: 20px;
  border-radius: 4px;
  border: 1px solid rgba(202, 184, 148, 0.88);
  background: linear-gradient(180deg, rgba(251, 249, 243, 0.96), rgba(222, 214, 197, 0.94));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  z-index: 4;
}

.prompt-fraud-card-title {
  color: rgba(46, 34, 22, 0.94);
  font-size: 11px;
  font-weight: 820;
  letter-spacing: 0.04em;
  line-height: 1;
}

.prompt-fraud-card-medal {
  position: absolute;
  top: 2px;
  left: 3px;
  width: 28px;
  height: 28px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.7);
  background: radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.95), rgba(201, 191, 176, 0.8) 42%, rgba(61, 53, 47, 0.88));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  z-index: 8;
  box-shadow:
    0 2px 6px rgba(0, 0, 0, 0.55),
    0 0 8px rgba(255, 244, 189, 0.44);
}

.prompt-fraud-card-medal > span {
  width: 72%;
  height: 72%;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--fraud-medal-bg);
  color: var(--fraud-medal-fg);
  font-size: 13px;
  font-weight: 900;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
}

.prompt-fraud-card-art {
  margin: 30px 7px 8px;
  height: 88px;
  border-radius: 4px;
  border: 1px solid rgba(175, 161, 132, 0.8);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(100% 120% at 50% 0%, var(--fraud-art-top), transparent 56%),
    linear-gradient(180deg, rgba(255, 250, 235, 0.12), var(--fraud-art-bottom));
  box-shadow:
    inset 0 0 0 1px rgba(255, 236, 196, 0.3),
    inset 0 0 16px rgba(0, 0, 0, 0.38);
}

.prompt-fraud-card-glyph {
  font-size: 38px;
  font-weight: 900;
  line-height: 1;
  color: rgba(247, 253, 255, 0.96);
  text-shadow:
    0 0 10px rgba(255, 255, 255, 0.24),
    0 2px 5px rgba(0, 0, 0, 0.6);
}

.prompt-fraud-card--water {
  --fraud-edge-color: rgba(102, 152, 196, 0.78);
  --fraud-edge-glow: rgba(124, 196, 255, 0.38);
  --fraud-base-top: #1a2a3e;
  --fraud-base-bottom: #0f1826;
  --fraud-ribbon-start: #25689f;
  --fraud-ribbon-end: #1b446c;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #ccf3ff, #4ea1d6 58%, #195580);
  --fraud-medal-fg: #effbff;
  --fraud-art-top: rgba(138, 206, 255, 0.4);
  --fraud-art-bottom: rgba(18, 48, 78, 0.84);
}

.prompt-fraud-card--fire {
  --fraud-edge-color: rgba(205, 123, 82, 0.78);
  --fraud-edge-glow: rgba(255, 140, 98, 0.4);
  --fraud-base-top: #3c1f18;
  --fraud-base-bottom: #1c120f;
  --fraud-ribbon-start: #c6352f;
  --fraud-ribbon-end: #8e1b17;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #ffca7f, #f36d33 58%, #9b2e1a);
  --fraud-medal-fg: #fff8eb;
  --fraud-art-top: rgba(255, 176, 126, 0.42);
  --fraud-art-bottom: rgba(88, 30, 20, 0.84);
}

.prompt-fraud-card--earth {
  --fraud-edge-color: rgba(174, 138, 93, 0.8);
  --fraud-edge-glow: rgba(225, 186, 113, 0.34);
  --fraud-base-top: #32261a;
  --fraud-base-bottom: #1a1410;
  --fraud-ribbon-start: #8a5d2f;
  --fraud-ribbon-end: #60401f;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #f1d79b, #c6924f 58%, #784d1d);
  --fraud-medal-fg: #fff6e4;
  --fraud-art-top: rgba(236, 199, 128, 0.36);
  --fraud-art-bottom: rgba(85, 57, 22, 0.84);
}

.prompt-fraud-card--wind {
  --fraud-edge-color: rgba(96, 169, 145, 0.78);
  --fraud-edge-glow: rgba(116, 223, 181, 0.34);
  --fraud-base-top: #183329;
  --fraud-base-bottom: #101f1b;
  --fraud-ribbon-start: #237258;
  --fraud-ribbon-end: #194e3d;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #c8f9e6, #55b68d 58%, #216f54);
  --fraud-medal-fg: #edfff6;
  --fraud-art-top: rgba(145, 241, 205, 0.36);
  --fraud-art-bottom: rgba(25, 73, 56, 0.84);
}

.prompt-fraud-card--thunder {
  --fraud-edge-color: rgba(140, 124, 200, 0.8);
  --fraud-edge-glow: rgba(183, 148, 255, 0.36);
  --fraud-base-top: #24213d;
  --fraud-base-bottom: #171427;
  --fraud-ribbon-start: #5f4a99;
  --fraud-ribbon-end: #40306f;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #efe2ff, #9c79dc 58%, #4e3385);
  --fraud-medal-fg: #faf3ff;
  --fraud-art-top: rgba(208, 180, 255, 0.38);
  --fraud-art-bottom: rgba(54, 36, 89, 0.84);
}

.prompt-fraud-card--generic {
  --fraud-edge-color: rgba(170, 190, 216, 0.56);
  --fraud-edge-glow: rgba(166, 193, 227, 0.34);
  --fraud-base-top: #223044;
  --fraud-base-bottom: #121c29;
  --fraud-ribbon-start: #44648a;
  --fraud-ribbon-end: #2d4159;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #dfe9f7, #8ba5cc 58%, #4d6283);
  --fraud-medal-fg: #f2f7ff;
  --fraud-art-top: rgba(170, 200, 235, 0.34);
  --fraud-art-bottom: rgba(35, 51, 72, 0.84);
}

.prompt-inline-entry {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
}

.prompt-inline-actions-row {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  justify-items: center;
  align-items: center;
}

.prompt-inline-entry--hinted-numeric {
  padding: 6px;
  border-radius: 12px;
  border: 1px solid rgba(114, 150, 177, 0.3);
  background:
    linear-gradient(180deg, rgba(26, 41, 58, 0.78), rgba(17, 30, 45, 0.82));
  box-shadow:
    inset 0 1px 0 rgba(208, 226, 241, 0.08),
    0 6px 14px rgba(3, 9, 18, 0.26);
}

.prompt-inline-btn.action-image-btn {
  width: 72px;
  height: 72px;
  min-height: 0;
  max-width: 72px;
  aspect-ratio: 1 / 1;
  border-radius: 12px !important;
  align-self: center;
  justify-self: center;
  flex-shrink: 0;
}

.prompt-inline-btn.action-image-btn.prompt-inline-btn--response-large {
  width: 96px;
  height: 96px;
  max-width: 96px;
  border-radius: 14px !important;
}

.prompt-inline-hint {
  min-height: 18px;
  padding: 0 4px;
  text-align: center;
  color: rgba(199, 219, 237, 0.88);
  font-size: 11px;
  line-height: 1.35;
  letter-spacing: 0.01em;
}

.prompt-inline-hint--hinted-numeric {
  min-height: 0;
  padding: 2px 6px 0;
  font-size: 11.5px;
  font-weight: 600;
  color: rgba(214, 231, 246, 0.96);
}

.prompt-inline-hint--attack-element {
  min-height: 0;
  margin: 0 0 6px;
  padding: 4px 8px;
  border-radius: 9px;
  border: 1px solid rgba(132, 165, 187, 0.36);
  background: linear-gradient(180deg, rgba(28, 43, 61, 0.72), rgba(17, 30, 45, 0.75));
  color: rgba(223, 239, 250, 0.96);
  font-size: 12px;
  font-weight: 640;
  letter-spacing: 0.01em;
}

.prompt-inline-grid--1 {
  grid-template-columns: 1fr;
}

.prompt-inline-grid--2 {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.prompt-inline-grid--3 {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.prompt-inline-grid--4 {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.prompt-inline-btn {
  min-height: 40px;
  width: 100%;
  max-width: 100%;
  border-radius: 10px;
  border: 1px solid rgba(137, 167, 186, 0.42);
  background: rgba(32, 48, 67, 0.68);
  color: #e3eef8;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.01em;
  transition: transform 0.16s ease, border-color 0.16s ease, filter 0.16s ease, background 0.16s ease;
}

.action-image-btn {
  -webkit-appearance: none !important;
  appearance: none !important;
  border: none !important;
  background: transparent !important;
  box-shadow: none !important;
  padding: 0 !important;
  overflow: hidden;
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.action-image-btn:focus,
.action-image-btn:focus-visible {
  outline: none !important;
  box-shadow: none !important;
}

.action-image-btn-fill {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  pointer-events: none;
  user-select: none;
}

.prompt-inline-btn.action-image-btn .action-image-btn-fill {
  transform-origin: center;
}

.action-image-fallback-text {
  position: relative;
  z-index: 1;
  font-size: 14px;
  font-weight: 700;
  color: #f2f8ff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.45);
}

.prompt-inline-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  border-color: rgba(180, 210, 227, 0.66);
  filter: brightness(1.05);
}

.prompt-inline-btn--normal {
  background: linear-gradient(180deg, rgba(47, 67, 88, 0.86), rgba(35, 53, 72, 0.88));
}

.prompt-inline-btn--numeric {
  font-size: 15px;
  font-weight: 800;
}

.prompt-inline-btn--hinted-numeric {
  min-height: 42px;
  border-radius: 10px;
  border-color: rgba(175, 197, 223, 0.52);
  background:
    linear-gradient(180deg, rgba(74, 101, 128, 0.92), rgba(54, 77, 99, 0.94));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow:
    inset 0 1px 0 rgba(228, 239, 252, 0.26),
    0 6px 12px rgba(5, 12, 21, 0.28);
}

.prompt-inline-btn-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.8em;
  min-height: 1.8em;
  padding: 0 0.45em;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 900;
  line-height: 1;
  color: #f6fbff;
  background: rgba(16, 29, 45, 0.46);
  border: 1px solid rgba(197, 220, 245, 0.38);
}

.prompt-inline-btn-caption {
  font-size: 12px;
  font-weight: 650;
  letter-spacing: 0.03em;
  color: rgba(234, 244, 255, 0.94);
}

.prompt-inline-btn--hinted-numeric:hover:not(:disabled) {
  filter: brightness(1.06);
  border-color: rgba(197, 220, 246, 0.75);
}

.prompt-inline-btn--success {
  border-color: rgba(111, 185, 141, 0.52);
  background: linear-gradient(180deg, rgba(30, 109, 74, 0.9), rgba(22, 78, 55, 0.9));
}

.prompt-inline-btn--take,
.prompt-inline-btn--counter,
.prompt-inline-btn--defend,
.prompt-inline-btn--cancel {
  color: #f6ecda;
  text-shadow: 0 1px 2px rgba(8, 8, 12, 0.65);
  background-repeat: no-repeat;
  background-size: cover;
  background-position: center;
  box-shadow:
    inset 0 1px 0 rgba(255, 238, 206, 0.22),
    0 8px 18px rgba(6, 7, 14, 0.32);
}

.prompt-inline-btn--take {
  border-color: rgba(205, 171, 113, 0.68);
  background-image: url('/assets/ui/prompt_btn_take.png');
}

.prompt-inline-btn--counter {
  border-color: rgba(157, 141, 228, 0.56);
  background-image: url('/assets/ui/prompt_btn_counter.png');
}

.prompt-inline-btn--defend {
  border-color: rgba(111, 170, 225, 0.6);
  background-image: url('/assets/ui/prompt_btn_defend.png');
}

.prompt-inline-btn--cancel {
  border-color: rgba(196, 152, 102, 0.56);
  background-image: url('/assets/ui/action_cancel_btn.png');
}

.prompt-inline-btn--extract {
  border-color: rgba(183, 154, 105, 0.56);
  background: linear-gradient(180deg, rgba(91, 69, 38, 0.9), rgba(68, 50, 28, 0.92));
}

.prompt-inline-btn--selected {
  box-shadow:
    0 0 0 2px rgba(241, 211, 150, 0.74),
    0 0 18px rgba(189, 152, 90, 0.38);
}

.prompt-inline-btn--disabled {
  opacity: 0.45;
  cursor: not-allowed;
  transform: none !important;
  filter: none !important;
}

.prompt-inline-pop-enter-active,
.prompt-inline-pop-leave-active {
  transition: opacity 0.22s ease, transform 0.22s ease;
}

.prompt-inline-pop-enter-from,
.prompt-inline-pop-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.98);
}

.prompt-fraud-side-pop-enter-active,
.prompt-fraud-side-pop-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.prompt-fraud-side-pop-enter-from,
.prompt-fraud-side-pop-leave-to {
  opacity: 0;
  transform: translateX(18px) scale(0.98);
}

@media (max-width: 900px) {
  .prompt-inline-surface {
    width: min(92vw, 680px);
    padding: 8px;
    gap: 7px;
  }

  .prompt-inline-grid--3,
  .prompt-inline-grid--4 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .prompt-inline-btn {
    min-height: 38px;
    font-size: 12px;
  }

  .prompt-inline-btn.action-image-btn {
    width: 72px;
    height: 72px;
    min-height: 0;
    max-width: 72px;
    justify-self: center;
  }

  .prompt-inline-btn.action-image-btn.prompt-inline-btn--response-large {
    width: 88px;
    height: 88px;
    max-width: 88px;
  }

  .prompt-inline-hint {
    min-height: 16px;
    font-size: 10px;
  }

  .prompt-inline-hint--hinted-numeric {
    font-size: 10.5px;
  }

  .prompt-inline-btn--hinted-numeric {
    min-height: 38px;
    gap: 6px;
  }

  .prompt-inline-btn-number {
    font-size: 12px;
  }

  .prompt-inline-btn-caption {
    font-size: 11px;
  }

  .prompt-skill-text {
    font-size: 12px;
  }

  .prompt-fraud-global-panel {
    width: min(760px, calc(100vw - 28px));
    max-height: calc(100vh - 28px);
  }

  .prompt-fraud-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .prompt-fraud-card {
    min-height: 126px;
  }
}

@media (max-width: 560px) {
  .prompt-inline-surface {
    width: 100%;
    padding: 7px;
    border-radius: 12px;
  }

  .prompt-inline-grid--2,
  .prompt-inline-grid--3,
  .prompt-inline-grid--4 {
    grid-template-columns: 1fr;
  }

  .prompt-skill-row {
    flex-direction: column;
    align-items: stretch;
    gap: 6px;
  }

  .prompt-skill-text {
    text-align: center;
  }

  .prompt-fraud-global-layer {
    justify-content: center;
    align-items: center;
    padding:
      max(10px, calc(var(--safe-top, 0px) + 4px))
      max(8px, calc(var(--safe-right, 0px) + 4px))
      max(10px, calc(var(--safe-bottom, 0px) + 4px))
      max(8px, calc(var(--safe-left, 0px) + 4px));
  }

  .prompt-fraud-global-panel {
    width: min(100%, 620px);
    border-radius: 12px;
    padding: 8px;
  }

  .prompt-inline-btn.action-image-btn {
    width: 72px;
    height: 72px;
    min-height: 0;
    max-width: 72px;
    justify-self: center;
  }

  .prompt-inline-btn.action-image-btn.prompt-inline-btn--response-large {
    width: 84px;
    height: 84px;
    max-width: 84px;
  }

  .prompt-fraud-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .prompt-fraud-card {
    min-height: 120px;
  }

  .prompt-fraud-card-glyph {
    font-size: 32px;
  }
}
</style>
