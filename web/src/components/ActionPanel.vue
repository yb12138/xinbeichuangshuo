<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useBattleInteractionState } from '../composables/useBattleInteractionState'
import { useSubmitAction } from '../composables/useSubmitAction'
import { useInteractionController } from '../composables/useInteractionController'
import {
    skillDiscardEffectiveElement,
    skillCanUseDiscardCard,
    skillCostTextOverride,
    skillDiscardGuideText as skillDiscardGuideTextById,
    skillSpecificDisabledReason,
} from '../constants/skillDisplayRules'
import type { AvailableSkill, PromptOption, PlayerView } from '../types/game'
import { cardArtSrc } from '../constants/cardArt'
import { characterArtSrc } from '../constants/characterArt'
import PromptDialog from './PromptDialog.vue'

const interruptStore = useInterruptStore()
const sessionStore = useSessionStore()
const snapshotStore = useSnapshotStore()
const actions = useSubmitAction()
const interaction = useInteractionController()
const {
    myPlayer,
    myHand,
    targetablePlayers,
    effectiveAvailableSkills,
    canConfirmSkill,
    cardMatchesExclusive,
    selectedActionCard,
} = useBattleInteractionState()

const prompt = computed(() => interruptStore.currentPrompt)
const isPromptForMe = computed(() => prompt.value?.player_id === sessionStore.myPlayerId)

const debugOpen = ref(false)
const debugTargetPlayerId = ref('')
const debugSetField = ref<'gem' | 'crystal' | 'heal' | 'max_heal'>('gem')
const debugSetValue = ref(1)
const debugEffectType = ref<'Shield' | 'Poison' | 'Weak' | 'PowerBlessing' | 'SwiftBlessing'>('Shield')
const debugEffectCount = ref(1)
const debugTokenKey = ref('')
const debugTokenValue = ref(1)
const debugExclusiveRoleId = ref('')
const debugExclusiveSkillId = ref('')
const debugExclusiveCount = ref(1)
const debugElement = ref<'Water' | 'Fire' | 'Earth' | 'Wind' | 'Thunder' | 'Light' | 'Dark'>('Fire')
const debugElementCount = ref(1)
const debugFaction = ref('圣')
const debugFactionCount = ref(1)
const debugMagicCardName = ref('')
const debugMagicCardCount = ref(1)
const debugDiscardCount = ref(1)
const debugStatus = ref('')

const debugAvailable = computed(() => {
    if (typeof window === 'undefined') return false
    const query = new URLSearchParams(window.location.search)
    return import.meta.env.DEV || query.has('debug') || query.has('debug_target')
})

type SpecialActionId = 'buy' | 'synthesize' | 'extract'

interface SpecialActionMeta {
    id: SpecialActionId
    label: string
    summary: string
    detail: string
    icon: string
}

interface SpecialActionDisplayItem extends SpecialActionMeta {
    available: boolean
    disabledReason: string
    promptLabel: string
}

const SPECIAL_ACTION_CATALOG: SpecialActionMeta[] = [
    {
        id: 'buy',
        label: '购买',
        summary: '消耗 1 点士气并摸 3 张牌',
        detail: '用于快速补牌，通常在手牌资源偏少时使用。',
        icon: '🛍',
    },
    {
        id: 'synthesize',
        label: '合成',
        summary: '阵营消耗 3 点资源，补充 3 张牌',
        detail: '适合资源充足、且你需要立刻扩充可打出的牌。',
        icon: '⚗',
    },
    {
        id: 'extract',
        label: '提炼',
        summary: '将阵营资源提炼为个人能量',
        detail: '用于启动高消耗技能；个人能量上限为 3。',
        icon: '⛏',
    },
]
const isMyTurn = computed(() => snapshotStore.currentPlayer === sessionStore.myPlayerId)
/** 等待面板文案："封印师（测试玩家4）"；拿不到角色名时只显示玩家名 */
function seatLabel(playerId?: string): string {
    if (!playerId) return ''
    const player = snapshotStore.players[playerId]
    const name = player?.name || playerId
    const role = snapshotStore.characters[player?.role || '']?.name
    return role ? `${role}（${name}）` : name
}
const specialActionModalVisible = ref(false)
const isActionSelectionSnapshot = computed(() =>
    snapshotStore.turnStage === 'ActionExecution' &&
    !snapshotStore.combatStage &&
    !snapshotStore.subflow
)
const isIdleMainTurnPanel = computed(() =>
    isMyTurn.value &&
    isActionSelectionSnapshot.value &&
    !prompt.value &&
    interruptStore.actionMode === 'none' &&
    interruptStore.skillMode === 'none'
)

type ActionHubOptionId = 'attack' | 'magic' | 'special' | 'cannot_act'

function normalizeActionHubOptionId(option: PromptOption): ActionHubOptionId | null {
    const id = (option.id || '').trim()
    if (id === 'attack' || id === 'magic' || id === 'special' || id === 'cannot_act') {
        return id
    }
    return null
}

const isActionSelectionPrompt = computed(() => {
    const p = prompt.value
    if (!p || !isPromptForMe.value) return false
    if (p.presentation?.kind !== 'action_hub') return false
    return (p.options || []).some((opt) => normalizeActionHubOptionId(opt) !== null)
})

const isActionHubContext = computed(() =>
    (isIdleMainTurnPanel.value || isActionSelectionPrompt.value) &&
    interruptStore.actionMode === 'none' &&
    interruptStore.skillMode === 'none'
)

// 回合内状态展示（仅当前回合玩家且是自己时显示）
// 包括：魔枪幻影形态的攻击加成、技能锁定等
const turnStateIndicators = computed(() => {
    if (!myPlayer.value || !myPlayer.value.is_active) return []
    const entries: Array<{ key: string; label: string; cls: string }> = []

    // 魔枪幻影形态状态
    const indicators = myPlayer.value.indicators ?? {}
    const attackBonus = indicators.ml_dark_release_next_attack_bonus
    if (attackBonus && attackBonus > 0) {
        entries.push({
            key: 'ml_dark_release_next_attack_bonus',
            label: `下次主动攻击伤害+${attackBonus}`,
            cls: 'bg-rose-900/70 text-rose-100 border-rose-500/40'
        })
    }
    const lockTurn = indicators.ml_dark_release_lock_turn
    if (lockTurn && lockTurn > 0) {
        entries.push({
            key: 'ml_dark_release_lock_turn',
            label: '本回合技能锁定',
            cls: 'bg-zinc-800/75 text-zinc-100 border-zinc-500/40'
        })
    }

    return entries
})

const isInlinePromptContext = computed(() =>
    !!prompt.value &&
    isPromptForMe.value &&
    !isActionHubContext.value &&
    interruptStore.actionMode === 'none' &&
    interruptStore.skillMode === 'none'
)

const actionPanelRootClass = computed(() => ({
    'action-panel-root--hub': isActionHubContext.value,
    'action-panel-root--panel': !isActionHubContext.value,
    'action-panel-root--prompt-inline': isInlinePromptContext.value,
}))

const actionPromptOptions = computed(() => isActionSelectionPrompt.value ? (prompt.value?.options ?? []) : [])
const normalizedActionPromptOptionMap = computed(() => {
    const map = new Map<ActionHubOptionId, PromptOption>()
    for (const option of actionPromptOptions.value) {
        const normalized = normalizeActionHubOptionId(option)
        if (!normalized || map.has(normalized)) continue
        map.set(normalized, option)
    }
    return map
})
const actionPromptOptionIdSet = computed(() => new Set(normalizedActionPromptOptionMap.value.keys()))
const actionPromptOptionMap = computed(() => {
    const map = new Map<ActionHubOptionId, string>()
    for (const [id, option] of normalizedActionPromptOptionMap.value.entries()) {
        map.set(id, option.label)
    }
    return map
})
const actionHubPromptNotice = computed(() => {
    if (!prompt.value || !isPromptForMe.value || !isActionSelectionPrompt.value) return ''
    const message = String(prompt.value.message || '').trim()
    if (!message) return ''
    return message
})
function extractFullnessAttackHint(message: string): string {
    return message.match(/【充盈】[^）)]+/)?.[0] || ''
}

const attackActionPromptNotice = computed(() => {
    if (interruptStore.actionMode !== 'attack') return ''
    return extractFullnessAttackHint(actionHubPromptNotice.value)
})

const specialActionOptions = computed<PromptOption[]>(() => {
    if (snapshotStore.hasPerformedStartup) {
        return []
    }
    if (isActionSelectionPrompt.value) {
        const fromSpecial = prompt.value?.special_options ?? []
        if (fromSpecial.length > 0) return fromSpecial
        return []
    }
    return [
        { id: 'buy', label: '购买', button_label: '购买' },
        { id: 'synthesize', label: '合成', button_label: '合成' },
        { id: 'extract', label: '提炼', button_label: '提炼' },
    ]
})

const hasHubSpecialActions = computed(() => specialActionOptions.value.length > 0)
const showSpecialHubEntry = computed(() => {
    if (!isActionHubContext.value) return false
    if (!isActionSelectionPrompt.value) return true
    return hasActionPromptOption('special') || hasHubSpecialActions.value
})
const isStartupSpecialLocked = computed(() => snapshotStore.hasPerformedStartup)
const isExtraActionPrompt = computed(() =>
    isActionSelectionPrompt.value &&
    !actionPromptOptionIdSet.value.has('special') &&
    actionPromptOptionIdSet.value.has('cannot_act') &&
    (actionPromptOptionIdSet.value.has('attack') || actionPromptOptionIdSet.value.has('magic')) &&
    (prompt.value?.special_options ?? []).length === 0
)
const cannotActButtonLabel = computed(() =>
    actionPromptLabel('cannot_act', isExtraActionPrompt.value ? '跳过额外行动' : '无法行动')
)
const teamStoneCount = computed(() =>
    sessionStore.myCamp === 'Red'
        ? snapshotStore.redGems + snapshotStore.redCrystals
        : snapshotStore.blueGems + snapshotStore.blueCrystals
)
const personalEnergy = computed(() => {
    const me = myPlayer.value
    if (!me) return 0
    return (me.gem || 0) + (me.crystal || 0)
})
const specialActionOptionSet = computed(() => new Set(specialActionOptions.value.map((option) => option.id)))
const specialActionLabelMap = computed(() => {
    const map = new Map<string, string>()
    for (const option of specialActionOptions.value) {
        map.set(option.id, option.label)
    }
    return map
})
const specialActionDisplayItems = computed<SpecialActionDisplayItem[]>(() => {
    return SPECIAL_ACTION_CATALOG.map((meta) => {
        const available = specialActionOptionSet.value.has(meta.id)
        const promptLabel = specialActionLabelMap.value.get(meta.id) || meta.label
        return {
            ...meta,
            promptLabel,
            available,
            disabledReason: available ? '' : resolveSpecialActionDisabledReason(meta.id),
        }
    })
})

const debugTargetPlayers = computed(() =>
    Object.values(snapshotStore.players).sort((a, b) => a.name.localeCompare(b.name, 'zh-Hans-CN'))
)

const debugRoleList = computed(() =>
    Object.values(snapshotStore.characters).sort((a, b) => a.name.localeCompare(b.name, 'zh-Hans-CN'))
)

const debugExclusiveSkillOptions = computed(() => {
    const role = snapshotStore.characters[debugExclusiveRoleId.value]
    if (!role || !Array.isArray(role.skills)) return []
    return role.skills
})

function hasActionPromptOption(optionId: string): boolean {
    if (!isActionSelectionPrompt.value) return true
    return actionPromptOptionIdSet.value.has(optionId as ActionHubOptionId)
}

function actionPromptLabel(optionId: string, fallback: string): string {
    if (!isActionSelectionPrompt.value) return fallback
    return actionPromptOptionMap.value.get(optionId as ActionHubOptionId) || fallback
}

function actionPromptRawOptionId(optionId: string): string {
    if (!isActionSelectionPrompt.value) return optionId
    const option = normalizedActionPromptOptionMap.value.get(optionId as ActionHubOptionId)
    return option?.id || optionId
}

function resolveSpecialActionDisabledReason(optionId: SpecialActionId): string {
    if (isStartupSpecialLocked.value) {
        return '你本回合已执行启动技能，不能执行特殊行动。'
    }
    if (isExtraActionPrompt.value) {
        return '当前为额外行动阶段，只能执行攻击或法术。'
    }
    if (optionId === 'synthesize' && teamStoneCount.value < 3) {
        return `阵营资源不足：合成需要至少 3 点资源（当前 ${teamStoneCount.value}）。`
    }
    if (optionId === 'extract') {
        if (teamStoneCount.value <= 0) {
            return '阵营没有可提炼资源。'
        }
        if (personalEnergy.value >= 3) {
            return '你的个人能量已满（上限 3），无法继续提炼。'
        }
    }
    if (optionId === 'buy' || optionId === 'synthesize') {
        return '手牌空间不足（该行动会额外摸 3 张牌），或本回合阶段限制未开放。'
    }
    if (!hasActionPromptOption('special') && isActionSelectionPrompt.value) {
        return '当前行动阶段未开放该特殊操作。'
    }
    return '当前条件不足，无法执行该行动。'
}

function invokeActionHubOption(optionId: string) {
    specialActionModalVisible.value = false
    if (isActionSelectionPrompt.value) {
        handlePromptOption(actionPromptRawOptionId(optionId))
        return
    }
    switch (optionId) {
        case 'attack':
            openActionHubAttack()
            return
        case 'magic':
            openActionHubMagic()
            return
        case 'buy':
            openBuyAction()
            return
        case 'synthesize':
            openSynthesizeAction()
            return
        case 'extract':
            openExtractAction()
            return
        case 'cannot_act':
            actions.submitCannotAct()
            return
        case 'pass':
            openPassAction()
            return
        case 'skill':
            openSkillAction()
            return
        default:
            return
    }
}

function openSpecialActionModal() {
    if (isStartupSpecialLocked.value) {
        interruptStore.showError('你本回合已执行启动技能，不能执行特殊行动')
        return
    }
    specialActionModalVisible.value = true
}

function closeSpecialActionModal() {
    specialActionModalVisible.value = false
}

function chooseSpecialAction(optionId: string) {
    const chosen = specialActionDisplayItems.value.find((item) => item.id === optionId)
    if (chosen && !chosen.available) {
        interruptStore.showError(chosen.disabledReason || '该特殊行动当前不可执行')
        return
    }
    specialActionModalVisible.value = false
    invokeActionHubOption(optionId)
}

function openActionHubAttack() {
    interruptStore.setActionMode('attack')
}

function openActionHubMagic() {
    interruptStore.setActionMode('magic')
}

function openSkillAction() {
    if (effectiveAvailableSkills.value.length === 0) {
        interruptStore.showError('当前没有可发动技能')
        return
    }
    interruptStore.setSkillMode('choosing_skill')
}

function openBuyAction() {
    actions.submitBuy()
}

function openSynthesizeAction() {
    actions.submitSynthesize()
}

function openExtractAction() {
    actions.submitExtract()
}

function openPassAction() {
    actions.submitPass()
}

watch(isActionHubContext, (isOpen) => {
    if (!isOpen) {
        specialActionModalVisible.value = false
    }
})

watch(debugExclusiveRoleId, () => {
    const options = debugExclusiveSkillOptions.value
    if (!options.some((skill) => skill.id === debugExclusiveSkillId.value)) {
        debugExclusiveSkillId.value = options[0]?.id || ''
    }
})

function isMagicMissileResponsePrompt(): boolean {
    return prompt.value?.presentation?.kind === 'response'
}

function handlePromptOption(optionId: string) {
    if (!prompt.value) return
    if (optionId === 'special') {
        if (isStartupSpecialLocked.value) {
            interruptStore.showError('你本回合已执行启动技能，不能执行特殊行动')
            return
        }
        openSpecialActionModal()
        return
    } else if (optionId === 'buy') {
        actions.submitBuy()
    } else if (optionId === 'extract') {
        actions.submitExtract()
    } else if (optionId === 'synthesize') {
        actions.submitSynthesize()
    } else if (optionId === 'attack') {
        interruptStore.setActionMode('attack')
    } else if (optionId === 'magic') {
        interruptStore.setActionMode('magic')
    } else if (optionId === 'cannot_act') {
        actions.submitCannotAct()
    } else if (optionId === 'skip' || optionId === 'cancel') {
        interaction.cancelPrompt()
        return
    } else if (optionId === 'confirm') {
        interaction.submitConfirm()
    } else if (optionId === 'take') {
        interaction.submitRespondTake()
    } else if (optionId === 'counter') {
        if (!interaction.submitRespondCounter(isMagicMissileResponsePrompt())) return
    } else if (optionId === 'defend') {
        if (!interaction.submitRespondDefend()) return
    } else if (optionId === 'yes' || optionId === 'no') {
        // 魔弹融合等确认选项：yes=0, no=1
        interaction.submitOptionIndex(optionId === 'yes' ? 0 : 1)
    } else if (optionId === 'normal' || optionId === 'reverse') {
        // 魔弹掌控方向选择：normal=0, reverse=1
        interaction.submitOptionIndex(optionId === 'normal' ? 0 : 1)
    } else if (prompt.value.presentation?.kind === 'skill_choice') {
        const idx = prompt.value.options.findIndex((o: { id: string }) => o.id === optionId)
        if (idx >= 0) {
            interaction.submitOptionIndex(idx)
        } else {
            interaction.cancelPrompt()
            return
        }
    } else {
        const index = parseInt(optionId)
        if (!isNaN(index)) {
            interaction.submitOptionIndex(index)
        } else {
            interruptStore.showError('当前选项缺少可提交的 option index，请刷新后重试')
        }
    }
    // 不在此处清除 prompt：等待后端 state_update（成功）或新 prompt 到达后再清除
    // 若后端报错，prompt 保持显示，用户可重新选择
}

function backFromMagicCard() {
    interruptStore.setMagicSubChoice('none')
    interruptStore.setSelectedHandIndexForAction(null)
}

const attackTargetCandidates = computed(() => {
    if (interruptStore.actionMode !== 'attack') return []
    return Object.values(snapshotStore.players).filter((p) => p.camp !== sessionStore.myCamp)
})

const actionTargets = computed<PlayerView[]>(() => {
    if (interruptStore.actionMode === 'attack') return attackTargetCandidates.value
    if (interruptStore.actionMode === 'magic') return targetablePlayers.value
    return []
})

const hasActionTargets = computed(() => actionTargets.value.length > 0)

function isStealthBlockedTarget(playerId: string): boolean {
    if (interruptStore.actionMode !== 'attack') return false
    if (targetablePlayers.value.some((p) => p.id === playerId)) return false
    const p = snapshotStore.players[playerId]
    if (!p || !Array.isArray(p.field)) return false
    return p.field.some((fc) => fc.mode === 'Effect' && fc.effect === 'Stealth')
}

const hasStealthBlockedAttackTarget = computed(() =>
    attackTargetCandidates.value.some((p) => isStealthBlockedTarget(p.id))
)

const SKILL_REQUIRE_MANUAL_TARGET_CONFIRM_IDS = new Set([
    'water_seal',
    'fire_seal',
    'earth_seal',
    'wind_seal',
    'thunder_seal',
])
const SKILL_SUBMIT_BEFORE_TARGET_IDS = new Set([
    'mb_thunder_scatter',
])
const SKILL_BACKEND_PROMPT_FLOW_IDS = new Set([
    'elementalist_freeze',
    'elementalist_thunder_strike',
    'elementalist_wind_blade',
    'elementalist_meteor',
    'elementalist_fireball',
    'css_blood_rose',
    'sage_arcane_codex',
    'sage_holy_codex',
    'mb_thunder_scatter',
    'bd_dissonance_chord',
])
const ELEMENTALIST_EXCLUSIVE_PROMPT_FLOW_SKILL_IDS = new Set([
    'elementalist_thunder_strike',
    'elementalist_freeze',
    'elementalist_wind_blade',
    'elementalist_meteor',
    'elementalist_fireball',
])
const isManualTargetConfirmSkillFlow = computed(() => {
    const skillId = interruptStore.selectedSkill?.id
    if (!skillId) return false
    return SKILL_REQUIRE_MANUAL_TARGET_CONFIRM_IDS.has(skillId)
})

function requiresExclusiveCardSelectionBeforeTarget(skill?: AvailableSkill | null): boolean {
    return !!skill &&
        !!skill.require_exclusive &&
        (skill.target_type ?? 0) !== 0 &&
        ((skill.cost_discards ?? 0) === 0 || ELEMENTALIST_EXCLUSIVE_PROMPT_FLOW_SKILL_IDS.has(skill.id))
}

function isBackendPromptFlowSkill(skill?: AvailableSkill | null): boolean {
    if (!skill) return false
    return isServerPublishedAvailableSkill(skill) && SKILL_BACKEND_PROMPT_FLOW_IDS.has(skill.id)
}

function isElementalistExclusivePromptFlowSkill(skill?: AvailableSkill | null): boolean {
    if (!skill) return false
    return isServerPublishedAvailableSkill(skill) && ELEMENTALIST_EXCLUSIVE_PROMPT_FLOW_SKILL_IDS.has(skill.id)
}

function isSelectedExclusiveCardValid(skill?: AvailableSkill | null): boolean {
    if (!skill) return false
    if (interruptStore.skillDiscardHandIndexes.length !== 1) return false
    const selectedIndex = interruptStore.skillDiscardHandIndexes[0]
    if (selectedIndex === undefined) return false
    const card = myHand.value[selectedIndex]
    if (!card) return false
    const roleId = sessionStore.myCharRole || myPlayer.value?.role || ''
    return cardMatchesExclusive(card, roleId, skill.title)
}

const selectedExclusiveSkillCard = computed(() => {
    const selectedIndex = interruptStore.skillDiscardHandIndexes[0]
    if (selectedIndex === undefined) return null
    return myHand.value[selectedIndex] || null
})

const canConfirmExclusiveSkillSelection = computed(() => {
    const skill = interruptStore.selectedSkill
    if (!requiresExclusiveCardSelectionBeforeTarget(skill)) return false
    return isSelectedExclusiveCardValid(skill)
})

function hasMatchingExclusiveCard(skill: AvailableSkill): boolean {
    if (!skill?.require_exclusive) return true
    const roleId = resolveMyRoleIdForExclusive()
    if (!roleId) return false
    return myHand.value.some(card => cardMatchesExclusive(card, roleId, skill.title)) ||
        (myPlayer.value?.exclusive_cards || []).some(card => cardMatchesExclusive(card, roleId, skill.title))
}

function selectSkill(skill: AvailableSkill) {
    if (!canSelectSkill(skill)) {
        interruptStore.showError(skillDisabledReason(skill))
        return
    }
    interruptStore.setSelectedSkill(skill)
    // A few skills intentionally start a backend prompt flow before costs,
    // branches, or targets are known. Regular server-published target skills
    // still use the shared client-side target picker below.
    if (isElementalistExclusivePromptFlowSkill(skill)) {
        interruptStore.setSkillMode('choosing_exclusive')
        return
    }
    if (isBackendPromptFlowSkill(skill)) {
        actions.submitUseSkill(skill.id, [], undefined, { clearSkillMode: true })
        return
    }
    if (requiresExclusiveCardSelectionBeforeTarget(skill)) {
        interruptStore.setSkillMode('choosing_exclusive')
        return
    }
    // 如果技能需要弃牌，先进入弃牌选择模式
    if (skill.cost_discards > 0) {
        const required = requiredDiscardCount(skill)
        if (required <= 0) {
            proceedAfterDiscard(skill)
            return
        }
        interruptStore.setSkillMode('choosing_discard')
        return
    }
    // 无需弃牌，直接进入目标选择或发动
    proceedAfterDiscard(skill)
}

function resolveMyRoleIdForExclusive(): string {
    const roleId = (sessionStore.myCharRole || myPlayer.value?.role || '').trim()
    return roleId
}

function cardMatchesSkillDiscard(card: { type: string; element: string; faction?: string; exclusive_char1?: string; exclusive_char2?: string; exclusive_skill1?: string; exclusive_skill2?: string }, skill: AvailableSkill): boolean {
    if (skill.require_exclusive) {
        const roleId = resolveMyRoleIdForExclusive()
        if (!roleId || !cardMatchesExclusive(card, roleId, skill.title)) return false
    }
    if (skill.discard_type && card.type !== skill.discard_type) return false
    const effectiveElement = skillDiscardEffectiveElement(card, resolveMyRoleIdForExclusive(), myPlayer.value?.form)
    if (skill.discard_element) return effectiveElement === skill.discard_element
    if (!skillCanUseDiscardCard(skill.id, card)) return false
    return true
}

function countSkillDiscardCandidates(skill: AvailableSkill): number {
    if (!skill || skill.cost_discards <= 0) return 0
    return myHand.value.filter(card => cardMatchesSkillDiscard(card, skill)).length
}

function hasOnmyojiSameFactionPair(): boolean {
    const countByFaction = new Map<string, number>()
    for (const card of myHand.value) {
        if (!card.faction) continue
        const next = (countByFaction.get(card.faction) || 0) + 1
        if (next >= 2) return true
        countByFaction.set(card.faction, next)
    }
    return false
}

type SkillTokenRule = {
    token: string
    min: number
    label: string
}

const SKILL_TOKEN_RULES: Record<string, SkillTokenRule[]> = {
    crk_killing_feast: [{ token: 'crk_blood_mark', min: 1, label: '血印' }],
    crk_crimson_cross: [{ token: 'crk_blood_mark', min: 1, label: '血印' }],
    css_blood_thorns: [{ token: 'css_blood', min: 1, label: '鲜血' }],
    css_crimson_flash: [{ token: 'css_blood', min: 1, label: '鲜血' }],
    css_blood_rose: [{ token: 'css_blood', min: 2, label: '鲜血' }],
    css_blood_barrier: [{ token: 'css_blood', min: 1, label: '鲜血' }],
    hom_rage_suppress: [{ token: 'hom_war_rune', min: 1, label: '战纹' }],
    hom_rune_smash: [{ token: 'hom_magic_rune', min: 1, label: '魔纹' }],
    hom_glyph_fusion: [{ token: 'hom_magic_rune', min: 2, label: '魔纹' }],
    hero_roar: [{ token: 'hero_anger', min: 1, label: '怒气' }],
    hero_calm_mind: [{ token: 'hero_wisdom', min: 4, label: '知性' }],
    hero_taunt: [{ token: 'hero_anger', min: 1, label: '怒气' }],
    fighter_hundred_dragon: [{ token: 'fighter_qi', min: 3, label: '斗气' }],
    fighter_burst_crash: [{ token: 'fighter_qi', min: 1, label: '斗气' }],
    ss_soul_mirror: [{ token: 'ss_yellow_soul', min: 2, label: '黄色灵魂' }],
    ss_soul_blast: [{ token: 'ss_yellow_soul', min: 3, label: '黄色灵魂' }],
    ss_soul_grant: [{ token: 'ss_blue_soul', min: 3, label: '蓝色灵魂' }],
    ss_soul_link: [
        { token: 'ss_yellow_soul', min: 1, label: '黄色灵魂' },
        { token: 'ss_blue_soul', min: 1, label: '蓝色灵魂' },
    ],
    arbiter_balance: [{ token: 'judgment', min: 1, label: '审判' }],
    bw_heavenfire_cleave: [{ token: 'bw_rebirth', min: 1, label: '重生' }],
}

function getMyTokenValue(token: string): number {
    return myPlayer.value?.tokens?.[token] ?? 0
}

function hasMyForm(form: string): boolean {
    return myPlayer.value?.form === form
}

function skillTokenDisabledReason(skill: AvailableSkill): string {
    const rules = SKILL_TOKEN_RULES[skill.id] || []
    for (const rule of rules) {
        const cur = getMyTokenValue(rule.token)
        if (cur < rule.min) {
            return `${rule.label}不足（需要 ${rule.min}，当前 ${cur}）。`
        }
    }
    if (skill.id === 'hb_radiant_descent') {
        const form = hasMyForm('holy_bow_holy_glory_form')
        const faith = getMyTokenValue('hb_faith')
        const heal = myPlayer.value?.heal ?? 0
        if (form) return '已处于圣煌形态，无法再次发动。'
        if (heal < 2 && faith < 2) return '治疗与信仰均不足2，无法发动。'
    }
    if (skill.id === 'hb_light_burst') {
        if (!hasMyForm('holy_bow_holy_glory_form')) return '仅圣煌形态下可发动。'
    }
    if (skill.id === 'hb_radiant_cannon') {
        const cannon = getMyTokenValue('hb_cannon')
        const faith = getMyTokenValue('hb_faith')
        const myMorale = sessionStore.myCamp === 'Red' ? snapshotStore.redMorale : snapshotStore.blueMorale
        const enemyMorale = sessionStore.myCamp === 'Red' ? snapshotStore.blueMorale : snapshotStore.redMorale
        const moraleGap = Math.max(0, enemyMorale - myMorale)
        const requiredFaith = 4 + moraleGap
        if (!hasMyForm('holy_bow_holy_glory_form')) return '仅圣煌形态下可发动。'
        if (cannon <= 0) return '圣煌辉光炮指示物不足。'
        if (moraleGap <= 0) return '我方士气未落后敌方，无法发动。'
        if (faith < requiredFaith) return `信仰不足（需要 ${requiredFaith}，当前 ${faith}）。`
    }
    if (skill.id === 'ms_shadow_meteor' && !hasMyForm('magic_swordsman_shadow_form')) {
        return '仅暗影形态下可发动。'
    }
    return ''
}

function canPaySkillEnergy(skill: AvailableSkill): boolean {
    const me = myPlayer.value
    if (!me) return false
    const gemNeed = skill.cost_gem || 0
    const crystalNeed = skill.cost_crystal || 0
    if (me.gem < gemNeed) return false
    const usableCrystal = me.crystal + (me.gem - gemNeed)
    return usableCrystal >= crystalNeed
}

function isServerPublishedAvailableSkill(skill: AvailableSkill): boolean {
    if (!skill) return false
    const list = snapshotStore.availableSkills || []
    if (list.length === 0) return false
    const normalize = (v: unknown) => String(v ?? '').trim().toLowerCase()
    const sid = normalize(skill.id)
    if (!sid) return false
    return list.some((item) => normalize(item.id) === sid)
}

function canSelectSkill(skill: AvailableSkill): boolean {
    if (!skill) return false
    // 服务端已下发 available_skills 时，以后端可用态为准，避免前端本地预检与后端规则漂移导致误置灰。
    if (isServerPublishedAvailableSkill(skill)) return true
    if (!canPaySkillEnergy(skill)) return false
    if (skill.require_exclusive && !hasMatchingExclusiveCard(skill)) return false
    const tokenReason = skillTokenDisabledReason(skill)
    if (tokenReason) return false
    if (skill.id === 'prayer_radiant_faith' || skill.id === 'prayer_dark_curse') {
        const prayerForm = hasMyForm('prayer_master_prayer_form')
        const prayerRune = myPlayer.value?.tokens?.prayer_rune ?? 0
        if (!prayerForm || prayerRune <= 0) return false
    }
    if (skill.id === 'elementalist_ignite') {
        const element = myPlayer.value?.tokens?.element ?? 0
        if (element < 3) return false
    }
    if (skill.id === 'onmyoji_shikigami_descend') {
        return hasOnmyojiSameFactionPair()
    }
    if (skill.cost_discards > 0) {
        const required = requiredDiscardCount(skill)
        if (required > 0 && countSkillDiscardCandidates(skill) < required) return false
    }
    return true
}

function skillDisabledReason(skill: AvailableSkill): string {
    if (!skill) return '当前不可发动该技能'
    if (!canPaySkillEnergy(skill)) {
        return `能量不足（需要 ${skill.cost_gem || 0} 宝石、${skill.cost_crystal || 0} 水晶）。`
    }
    const tokenReason = skillTokenDisabledReason(skill)
    if (tokenReason) return tokenReason
    const specificReason = skillSpecificDisabledReason(skill.id)
    if (specificReason) return specificReason
    if (skill.id === 'prayer_radiant_faith' || skill.id === 'prayer_dark_curse') {
        const prayerForm = hasMyForm('prayer_master_prayer_form')
        const prayerRune = myPlayer.value?.tokens?.prayer_rune ?? 0
        if (!prayerForm) return '仅祈祷形态下可发动。'
        if (prayerRune <= 0) return '祈祷符文不足，无法发动。'
    }
    if (skill.id === 'elementalist_ignite') {
        return '元素不足3点，无法发动【元素点燃】。'
    }
    if (skill.id === 'angel_blessing') {
        return '手牌中没有水系牌，无法发动【天使祝福】。'
    }
    if (skill.id === 'angel_cleanse') {
        return '手牌中没有风系牌，无法发动【风之洁净】。'
    }
    if (skill.require_exclusive && !hasMatchingExclusiveCard(skill)) {
        return `缺少可用于发动的「${skill.title}」独有技手牌。`
    }
    if (skill.cost_discards > 0) {
        const required = requiredDiscardCount(skill)
        return `可用于弃置的手牌不足，至少需要 ${required} 张。`
    }
    return '当前不可发动该技能'
}

function proceedAfterDiscard(skill: AvailableSkill) {
    if (requiresExclusiveCardSelectionBeforeTarget(skill)) {
        return
    }
    // 魔弓【雷光散射】先移除雷系充能，再根据额外移除数决定是否选择目标。
    if (SKILL_SUBMIT_BEFORE_TARGET_IDS.has(skill.id)) {
        const selections = interruptStore.skillDiscardHandIndexes.length > 0 ? [...interruptStore.skillDiscardHandIndexes] : undefined
        actions.submitUseSkill(skill.id, [], selections, { clearSkillMode: true })
        return
    }
    // target_type=0 (None): 无需目标，直接发动
    if (skill.target_type === 0) {
        const selections = interruptStore.skillDiscardHandIndexes.length > 0 ? [...interruptStore.skillDiscardHandIndexes] : undefined
        actions.submitUseSkill(skill.id, [], selections, { clearSkillMode: true })
        return
    }
    // target_type=1 (Self): 自动选中自己并发动
    if (skill.target_type === 1) {
        const selections = interruptStore.skillDiscardHandIndexes.length > 0 ? [...interruptStore.skillDiscardHandIndexes] : undefined
        actions.submitUseSkill(skill.id, [sessionStore.myPlayerId], selections, { clearSkillMode: true })
        return
    }
    interruptStore.setSkillMode('choosing_target')
}

function confirmSkill() {
    const skill = interruptStore.selectedSkill
    if (!skill) return
    if (!canConfirmSkill.value) {
        const minTargets = (skill.min_targets || 0) > 0 ? (skill.min_targets || 0) : (skill.target_type >= 2 ? 1 : 0)
        const maxTargets = (skill.max_targets || 0) > 0 ? (skill.max_targets || 0) : 1
        if (minTargets === maxTargets) {
            interruptStore.showError(`请先选择 ${minTargets} 个目标`)
        } else {
            interruptStore.showError(`请选择 ${minTargets}-${maxTargets} 个目标`)
        }
        return
    }
    const selections = interruptStore.skillDiscardHandIndexes.length > 0 ? [...interruptStore.skillDiscardHandIndexes] : undefined
    actions.submitUseSkill(skill.id, [...interruptStore.skillTargetIds], selections, { clearSkillMode: true })
}

function confirmExclusiveSkillSelection() {
    const skill = interruptStore.selectedSkill
    if (!skill || !requiresExclusiveCardSelectionBeforeTarget(skill)) return
    if (!canConfirmExclusiveSkillSelection.value) {
      interruptStore.showError(`请先选择对应的「${skill.title}」独有技手牌`)
      return
    }
    if (isElementalistExclusivePromptFlowSkill(skill)) {
        const selections = interruptStore.skillDiscardHandIndexes.length > 0 ? [...interruptStore.skillDiscardHandIndexes] : undefined
        actions.submitUseSkill(skill.id, [], selections, { clearSkillMode: true })
        return
    }
    interruptStore.setSkillMode('choosing_target')
}

watch(
    () => [interruptStore.skillMode, interruptStore.selectedSkill?.id, interruptStore.skillDiscardHandIndexes.length] as const,
    ([mode, _skillId, selectedCount]) => {
        if (mode !== 'choosing_discard') return
        const skill = interruptStore.selectedSkill
        if (!skill) return
        const required = requiredDiscardCount(skill)
        if (required > 0 && selectedCount >= required) {
            proceedAfterDiscard(skill)
        }
    }
)

function skillCostText(skill: AvailableSkill): string {
    const override = skillCostTextOverride(skill.id)
    if (override) {
        return override
    }
    const parts: string[] = []
    if (skill.cost_gem > 0) parts.push(`${skill.cost_gem}宝石`)
    if (skill.cost_crystal > 0) parts.push(`${skill.cost_crystal}水晶`)
    if (skill.cost_discards > 0) {
        parts.push(skill.require_exclusive ? `弃${skill.cost_discards}独有技手牌` : `弃${skill.cost_discards}牌`)
    } else if (skill.require_exclusive) {
        parts.push('独有技手牌')
    }
    return parts.length > 0 ? parts.join('+') : '免费'
}

function openDebugPanel() {
    debugOpen.value = true
    debugStatus.value = ''
    if (!debugTargetPlayerId.value) {
        debugTargetPlayerId.value = sessionStore.myPlayerId
    }
    if (!debugExclusiveRoleId.value) {
        debugExclusiveRoleId.value = sessionStore.myCharRole || debugRoleList.value[0]?.id || ''
    }
    if (!debugExclusiveSkillId.value) {
        debugExclusiveSkillId.value = debugExclusiveSkillOptions.value[0]?.id || ''
    }
}

function closeDebugPanel() {
    debugOpen.value = false
}

function ensureDebugTargetPlayerId(): string | null {
    const pid = debugTargetPlayerId.value || sessionStore.myPlayerId
    if (!pid || !snapshotStore.players[pid]) {
        interruptStore.showError('请选择有效的目标角色')
        return null
    }
    return pid
}

function debugTargetName(pid: string): string {
    return snapshotStore.players[pid]?.name || pid
}

function applyDebugEffect() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const count = Number(debugEffectCount.value)
    if (!Number.isFinite(count) || count < 0) {
        interruptStore.showError('基础效果数量需为 >= 0 的数字')
        return
    }
    actions.cheatEffect(pid, debugEffectType.value, Math.floor(count))
    debugStatus.value = `已设置 ${debugTargetName(pid)} 的基础效果 ${debugEffectType.value}=${Math.floor(count)}`
}

function applyDebugSet() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const value = Number(debugSetValue.value)
    if (!Number.isFinite(value)) {
        interruptStore.showError('请输入有效数字')
        return
    }
    actions.cheatSet(pid, debugSetField.value, Math.floor(value))
    debugStatus.value = `已设置 ${debugTargetName(pid)} 的 ${debugSetField.value}=${Math.floor(value)}`
}

/** 调试：按当前快照为房间内每名玩家宝石 +3（相对增量） */
function applyDebugAddGemsEveryone() {
    const players = debugTargetPlayers.value
    if (players.length === 0) {
        interruptStore.showError('房间中暂无角色')
        return
    }
    const delta = 3
    for (const p of players) {
        const cur = Math.floor(Number(p.gem) || 0)
        actions.cheatSet(p.id, 'gem', cur + delta)
    }
    debugStatus.value = `已为 ${players.length} 名角色各增加 ${delta} 宝石`
}

function applyDebugToken() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const key = debugTokenKey.value.trim()
    const value = Number(debugTokenValue.value)
    if (!key) {
        interruptStore.showError('请输入指示物 key')
        return
    }
    if (!Number.isFinite(value)) {
        interruptStore.showError('请输入有效数字')
        return
    }
    actions.cheatToken(pid, key, Math.floor(value))
    debugStatus.value = `已设置 ${debugTargetName(pid)} 的指示物 ${key}=${Math.floor(value)}`
}

function applyDebugExclusiveCard() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    if (!debugExclusiveRoleId.value) {
        interruptStore.showError('请选择角色来源')
        return
    }
    if (!debugExclusiveSkillId.value) {
        interruptStore.showError('请选择独有技')
        return
    }
    const count = Number(debugExclusiveCount.value)
    if (!Number.isFinite(count) || count <= 0) {
        interruptStore.showError('独有技手牌数量需为 > 0 的数字')
        return
    }
    actions.cheatGiveExclusive(pid, debugExclusiveRoleId.value, debugExclusiveSkillId.value, Math.floor(count))
    debugStatus.value = `已给 ${debugTargetName(pid)} 添加 ${Math.floor(count)} 张独有技手牌`
}

function applyDebugElementCards() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const count = Number(debugElementCount.value)
    if (!Number.isFinite(count) || count <= 0) {
        interruptStore.showError('系别手牌数量需为 > 0 的数字')
        return
    }
    actions.cheatGiveByElement(pid, debugElement.value, Math.floor(count))
    debugStatus.value = `已给 ${debugTargetName(pid)} 添加 ${Math.floor(count)} 张 ${elementName(debugElement.value)}手牌`
}

function applyDebugFactionCards() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const faction = debugFaction.value.trim()
    if (!faction) {
        interruptStore.showError('请输入命格')
        return
    }
    const count = Number(debugFactionCount.value)
    if (!Number.isFinite(count) || count <= 0) {
        interruptStore.showError('命格手牌数量需为 > 0 的数字')
        return
    }
    actions.cheatGiveByFaction(pid, faction, Math.floor(count))
    debugStatus.value = `已给 ${debugTargetName(pid)} 添加 ${Math.floor(count)} 张 ${faction}命格手牌`
}

function applyDebugMagicCard() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const cardName = debugMagicCardName.value.trim()
    if (!cardName) {
        interruptStore.showError('请输入法术牌名称')
        return
    }
    const count = Number(debugMagicCardCount.value)
    if (!Number.isFinite(count) || count <= 0) {
        interruptStore.showError('法术牌数量需为 > 0 的数字')
        return
    }
    actions.cheatGiveMagicByName(pid, cardName, Math.floor(count))
    debugStatus.value = `已给 ${debugTargetName(pid)} 添加 ${Math.floor(count)} 张法术牌【${cardName}】`
}

function applyDebugDiscard() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const count = Number(debugDiscardCount.value)
    if (!Number.isFinite(count) || count <= 0) {
        interruptStore.showError('弃牌数量需为 > 0 的数字')
        return
    }
    actions.cheatDiscard(pid, Math.floor(count))
    debugStatus.value = `已让 ${debugTargetName(pid)} 弃置最多 ${Math.floor(count)} 张手牌`
}

function requiredDiscardCount(skill: AvailableSkill): number {
    if (!skill || skill.cost_discards <= 0) return 0
    // 神官-神圣领域：手牌不足2时，改为弃全部手牌。
    if (skill.id === 'priest_divine_domain') {
        return Math.min(skill.cost_discards, myHand.value.length)
    }
    // 神官-水之神力：若弃完水系后无剩余手牌，则仅需弃1张水系牌。
    if (skill.id === 'priest_water_power') {
        return Math.min(skill.cost_discards, myHand.value.length)
    }
    return skill.cost_discards
}

function skillDiscardGuideText(skill: AvailableSkill): string {
    if (skill.require_exclusive) {
        return `弃标有「${skill.title}」的独有技手牌`
    }
    if (skill.discard_element) {
        return `弃置${elementName(skill.discard_element)}牌`
    }
    if (skill.discard_type) {
        return `弃置${skill.discard_type === 'Magic' ? '法术牌' : '攻击牌'}`
    }
    const guide = skillDiscardGuideTextById(skill.id)
    if (guide) {
        return guide
    }
    return '请在下方手牌区选择要弃置的牌'
}

function elementName(el: string): string {
    const map: Record<string, string> = { Water: '水系', Fire: '火系', Wind: '风系', Earth: '土系', Dark: '暗灭' }
    return map[el] || el
}

</script>

<template>
    <div
        class="action-panel-root"
        :class="actionPanelRootClass"
    >
        <!-- 攻击/法术模式 -->
        <div v-if="interruptStore.actionMode !== 'none'" class="dock-panel action-mode-panel">
            <i class="dock-corner dock-corner--tl" aria-hidden="true" />
            <i class="dock-corner dock-corner--tr" aria-hidden="true" />
            <i class="dock-corner dock-corner--bl" aria-hidden="true" />
            <i class="dock-corner dock-corner--br" aria-hidden="true" />
            <header class="dock-hd">
                <span class="dock-hd-left">
                    <img class="dock-ico dock-ico--img" :src="`/assets/icons/${interruptStore.actionMode === 'attack' ? 'attack' : 'magic'}.webp`" alt="" aria-hidden="true" draggable="false" />
                    <span class="dock-title">
                        {{ interruptStore.actionMode === 'attack' ? '攻击' : (interruptStore.magicSubChoice === 'none' ? '法术' : '法术 · 出牌') }}
                    </span>
                </span>
                <nav v-if="interruptStore.actionMode === 'attack' || interruptStore.magicSubChoice !== 'none'" class="dock-steps" aria-hidden="true">
                    <span :class="interruptStore.selectedHandIndexForAction === null ? 'on' : 'done'">① 选牌{{ interruptStore.selectedHandIndexForAction === null ? '' : ' ✓' }}</span>
                    <i class="dock-steps-sep">›</i>
                    <span :class="interruptStore.selectedHandIndexForAction !== null ? 'on' : ''">② 选目标</span>
                </nav>
            </header>

            <!-- 法术行动：先选择 出牌 或 发动技能 -->
            <template v-if="interruptStore.actionMode === 'magic' && interruptStore.magicSubChoice === 'none'">
                <div class="dock-bd">
                    
                    <div class="dock-medal-row">
                        <button type="button" class="dock-medal" title="打出法术牌" @click="interruptStore.setMagicSubChoice('card')">
                            <span class="dock-medal-ring">
                                <img class="dock-medal-ico dock-medal-ico--img" src="/assets/icons/card.webp" alt="" aria-hidden="true" draggable="false" />
                            </span>
                            <span class="dock-medal-label">打出法术牌</span>
                        </button>
                        <button
                            v-if="effectiveAvailableSkills.length > 0"
                            type="button"
                            class="dock-medal"
                            data-testid="action-skill"
                            title="发动技能"
                            @click="interruptStore.setSkillMode('choosing_skill'); interruptStore.clearActionMode()"
                        >
                            <span class="dock-medal-ring">
                                <img class="dock-medal-ico dock-medal-ico--img" src="/assets/icons/skill.webp" alt="" aria-hidden="true" draggable="false" />
                            </span>
                            <span class="dock-medal-label">发动技能</span>
                        </button>
                    </div>
                </div>
                <footer class="dock-ft">
                    <button type="button" class="dock-btn dock-btn--ghost" title="取消" @click="interruptStore.clearActionMode()"><img class="dock-btn-ico" src="/assets/icons/cancel.webp" alt="" aria-hidden="true" />取消</button>
                </footer>
            </template>
            <!-- 攻击模式 或 法术已选「出牌」：选牌 + 选目标 -->
            <template v-else>
                <div class="dock-bd">
                    <div v-if="attackActionPromptNotice" class="dock-notice">
                        {{ attackActionPromptNotice }}
                    </div>
                    <div class="dock-flow" aria-hidden="true">
                        <span class="dock-slot dock-slot--card" :class="interruptStore.selectedHandIndexForAction === null ? 'dock-slot--pulse' : 'dock-slot--filled'">
                            <svg v-if="interruptStore.selectedHandIndexForAction === null" class="dock-slot-ico" viewBox="0 0 24 24">
                                <rect x="6.5" y="4" width="11" height="16" rx="1.6" fill="none" stroke="currentColor" stroke-width="1.3" />
                                <path d="M9.5 9h5M9.5 12h5" stroke="currentColor" stroke-width="1.1" stroke-linecap="round" fill="none" />
                            </svg>
                            <img v-else-if="selectedActionCard" class="dock-slot-art" :src="cardArtSrc(selectedActionCard)" :alt="selectedActionCard.name" />
                            <svg v-else class="dock-slot-ico" viewBox="0 0 24 24">
                                <path d="M5 12.5l4.5 4.5L19 7.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                            </svg>
                        </span>
                        <span class="dock-flow-arrow">➤</span>
                        <span class="dock-slot dock-slot--target" :class="interruptStore.selectedHandIndexForAction !== null ? 'dock-slot--pulse' : ''">
                            <svg class="dock-slot-ico" viewBox="0 0 24 24">
                                <circle cx="12" cy="12" r="6.4" fill="none" stroke="currentColor" stroke-width="1.3" />
                                <circle cx="12" cy="12" r="2.2" fill="none" stroke="currentColor" stroke-width="1.3" />
                                <path d="M12 2.6v3.2M12 18.2v3.2M2.6 12h3.2M18.2 12h3.2" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" fill="none" />
                            </svg>
                        </span>
                    </div>
                    <div v-if="interruptStore.selectedHandIndexForAction !== null" class="dock-texts">
                        <div v-if="hasActionTargets" class="dock-hint">
                            已选牌，请直接点击战场角色立绘完成{{ interruptStore.actionMode === 'attack' ? '攻击' : '施法' }}。
                        </div>
                        <div v-else class="dock-hint">当前法术无需手动选目标，将按规则自动结算。</div>
                        <div v-if="hasStealthBlockedAttackTarget" class="dock-hint dock-hint--warn">
                            潜行状态无法选中
                        </div>
                    </div>
                    <div v-else class="dock-hint">
                        先在下方手牌选一张{{ interruptStore.actionMode === 'attack' ? '攻击' : '法术' }}牌
                    </div>
                </div>
                <footer class="dock-ft">
                    <button type="button" class="dock-btn dock-btn--ghost" :title="interruptStore.actionMode === 'magic' ? '返回' : '取消'" @click="interruptStore.actionMode === 'magic' ? backFromMagicCard() : interruptStore.clearActionMode()">
                        <img class="dock-btn-ico" src="/assets/icons/cancel.webp" alt="" aria-hidden="true" />{{ interruptStore.actionMode === 'magic' ? '返回' : '取消' }}
                    </button>
                    <button
                        v-if="interruptStore.actionMode === 'magic' && effectiveAvailableSkills.length > 0"
                        type="button"
                        class="dock-btn dock-btn--ghost dock-btn--right"
                        title="改用技能"
                        @click="interruptStore.setSkillMode('choosing_skill'); interruptStore.clearActionMode()"
                    >
                        改用技能 ›
                    </button>
                </footer>
            </template>
        </div>

        <!-- 技能发动流程：选择技能 -->
        <div v-else-if="interruptStore.skillMode === 'choosing_skill'" class="dock-panel skill-select-panel" data-testid="skill-select-panel">
            <i class="dock-corner dock-corner--tl" aria-hidden="true" />
            <i class="dock-corner dock-corner--tr" aria-hidden="true" />
            <i class="dock-corner dock-corner--bl" aria-hidden="true" />
            <i class="dock-corner dock-corner--br" aria-hidden="true" />
            <header class="dock-hd">
                <span class="dock-hd-left">
                    <img class="dock-ico dock-ico--img" src="/assets/icons/skill.webp" alt="" aria-hidden="true" draggable="false" />
                    <span class="dock-title">发动技能</span>
                </span>
                <span class="dock-hd-meta">共 {{ effectiveAvailableSkills.length }} 个</span>
            </header>
            <div class="dock-bd dock-bd--list">
                <button
                    v-for="skill in effectiveAvailableSkills"
                    :key="skill.id"
                    type="button"
                    class="sk"
                    :class="{ 'sk--disabled': !canSelectSkill(skill) }"
                    :data-testid="`skill-${skill.id}`"
                    :title="canSelectSkill(skill) ? skill.description : skillDisabledReason(skill)"
                    :disabled="!canSelectSkill(skill)"
                    @click="selectSkill(skill)"
                >
                    <span class="sk-ico" aria-hidden="true"><img src="/assets/icons/skill.webp" alt="" draggable="false" /></span>
                    <span class="sk-main">
                        <span class="sk-name">{{ skill.title }}</span>
                        <span v-if="skill.description" class="sk-desc" :title="skill.description">{{ skill.description }}</span>
                        <span v-if="!canSelectSkill(skill)" class="sk-reason">{{ skillDisabledReason(skill) }}</span>
                    </span>
                    <span class="sk-cost">{{ skillCostText(skill) }}</span>
                </button>
            </div>
            <footer class="dock-ft">
                <button type="button" class="dock-btn dock-btn--ghost" data-testid="skill-cancel-btn" title="取消" @click="interruptStore.clearSkillMode()"><img class="dock-btn-ico" src="/assets/icons/cancel.webp" alt="" aria-hidden="true" />取消</button>
            </footer>
        </div>

        <!-- 技能发动流程：选择独有技手牌 -->
        <div v-else-if="interruptStore.skillMode === 'choosing_exclusive' && interruptStore.selectedSkill" class="dock-panel skill-discard-panel" data-testid="skill-exclusive-select-panel">
            <i class="dock-corner dock-corner--tl" aria-hidden="true" />
            <i class="dock-corner dock-corner--tr" aria-hidden="true" />
            <i class="dock-corner dock-corner--bl" aria-hidden="true" />
            <i class="dock-corner dock-corner--br" aria-hidden="true" />
            <header class="dock-hd">
                <span class="dock-hd-left">
                    <img class="dock-ico dock-ico--img" src="/assets/icons/skill.webp" alt="" aria-hidden="true" draggable="false" />
                    <span class="dock-title">{{ interruptStore.selectedSkill.title }}</span>
                </span>
                <nav class="dock-steps" aria-hidden="true">
                    <span class="on">① 选牌</span>
                    <i class="dock-steps-sep">›</i>
                    <span>② 选目标</span>
                </nav>
            </header>
            <div class="dock-bd">
                <p v-if="interruptStore.selectedSkill.description" class="dock-hint dock-hint--desc">{{ interruptStore.selectedSkill.description }}</p>
                <div class="dock-hint">
                    请在手牌区选择对应的独有技手牌，然后点击确认。
                </div>
                <div class="dock-progress" aria-hidden="true">
                    <i class="dock-dot" :class="{ 'dock-dot--on': interruptStore.skillDiscardHandIndexes.length > 0 }" />
                </div>
                <div v-if="selectedExclusiveSkillCard" class="dock-accent">
                    已选独有技手牌：{{ selectedExclusiveSkillCard.name }}
                </div>
            </div>
            <footer class="dock-ft">
                <button type="button" class="dock-btn dock-btn--ghost" title="取消" @click="interruptStore.clearSkillMode()"><img class="dock-btn-ico" src="/assets/icons/cancel.webp" alt="" aria-hidden="true" />取消</button>
                <button
                    type="button"
                    class="dock-btn dock-btn--primary"
                    :disabled="!canConfirmExclusiveSkillSelection"
                    data-testid="skill-exclusive-confirm-btn"
                    title="确认"
                    @click="confirmExclusiveSkillSelection"
                ><img class="dock-btn-ico" src="/assets/icons/confirm.webp" alt="" aria-hidden="true" />确认</button>
            </footer>
        </div>

        <!-- 技能发动流程：选择弃牌 -->
        <div v-else-if="interruptStore.skillMode === 'choosing_discard' && interruptStore.selectedSkill" class="dock-panel skill-discard-panel">
            <i class="dock-corner dock-corner--tl" aria-hidden="true" />
            <i class="dock-corner dock-corner--tr" aria-hidden="true" />
            <i class="dock-corner dock-corner--bl" aria-hidden="true" />
            <i class="dock-corner dock-corner--br" aria-hidden="true" />
            <header class="dock-hd">
                <span class="dock-hd-left">
                    <img class="dock-ico dock-ico--img" src="/assets/icons/skill.webp" alt="" aria-hidden="true" draggable="false" />
                    <span class="dock-title">{{ interruptStore.selectedSkill.title }}</span>
                </span>
                <nav class="dock-steps" aria-hidden="true">
                    <span class="on">① 弃牌</span>
                    <i class="dock-steps-sep">›</i>
                    <span>② 选目标</span>
                </nav>
            </header>
            <div class="dock-bd">
                <div class="dock-hint">
                    {{ skillDiscardGuideText(interruptStore.selectedSkill) }}
                </div>
                <div class="dock-progress" aria-hidden="true">
                    <i
                        v-for="n in requiredDiscardCount(interruptStore.selectedSkill)"
                        :key="n"
                        class="dock-dot"
                        :class="{ 'dock-dot--on': n <= interruptStore.skillDiscardHandIndexes.length }"
                    />
                </div>
                <div class="dock-hint dock-hint--sm">请在下方手牌区点击选择，选满后自动进入下一步</div>
            </div>
            <footer class="dock-ft">
                <button type="button" class="dock-btn dock-btn--ghost" title="取消" @click="interruptStore.clearSkillMode()"><img class="dock-btn-ico" src="/assets/icons/cancel.webp" alt="" aria-hidden="true" />取消</button>
                <span class="dock-ft-center">已选 {{ interruptStore.skillDiscardHandIndexes.length }} / {{ requiredDiscardCount(interruptStore.selectedSkill) }}</span>
            </footer>
        </div>

        <!-- 技能发动流程：选择目标 -->
        <div v-else-if="interruptStore.skillMode === 'choosing_target' && interruptStore.selectedSkill" class="dock-panel skill-target-panel">
            <i class="dock-corner dock-corner--tl" aria-hidden="true" />
            <i class="dock-corner dock-corner--tr" aria-hidden="true" />
            <i class="dock-corner dock-corner--bl" aria-hidden="true" />
            <i class="dock-corner dock-corner--br" aria-hidden="true" />
            <header class="dock-hd">
                <span class="dock-hd-left">
                    <img class="dock-ico dock-ico--img" src="/assets/icons/skill.webp" alt="" aria-hidden="true" draggable="false" />
                    <span class="dock-title">{{ interruptStore.selectedSkill.title }}</span>
                </span>
                <nav v-if="interruptStore.selectedSkill.cost_discards > 0 || interruptStore.selectedSkill.require_exclusive" class="dock-steps" aria-hidden="true">
                    <span class="done">① {{ interruptStore.selectedSkill.cost_discards > 0 ? '弃牌' : '选牌' }}</span>
                    <i class="dock-steps-sep">›</i>
                    <span class="on">② 选目标</span>
                </nav>
                <span v-else class="dock-hd-meta">选目标 {{ interruptStore.skillTargetIds.length }}/{{ interruptStore.selectedSkill.max_targets > 0 ? interruptStore.selectedSkill.max_targets : 1 }}</span>
            </header>
            <div class="dock-bd">
                <p v-if="interruptStore.selectedSkill.description" class="dock-hint dock-hint--desc">{{ interruptStore.selectedSkill.description }}</p>
                <div class="dock-hint">
                    <template v-if="isManualTargetConfirmSkillFlow">
                        点击角色头像选择目标，然后点击“确认发动”
                    </template>
                    <template v-else>
                        请直接点击角色头像选择目标并自动发动
                        <span v-if="(interruptStore.selectedSkill.max_targets || 1) === 1">（单目标：点击即发动）</span>
                        <span v-else-if="(interruptStore.selectedSkill.min_targets || 0) >= (interruptStore.selectedSkill.max_targets || 1)">
                            （选满 {{ interruptStore.selectedSkill.max_targets || 1 }} 个目标后自动发动）
                        </span>
                        <span v-else>
                            （多目标：先点头像选择，再点击任一已选头像确认发动）
                        </span>
                    </template>
                </div>
                <div class="dock-progress" aria-hidden="true">
                    <i
                        v-for="n in (interruptStore.selectedSkill.max_targets > 0 ? interruptStore.selectedSkill.max_targets : 1)"
                        :key="n"
                        class="dock-dot"
                        :class="{ 'dock-dot--on': n <= interruptStore.skillTargetIds.length }"
                    />
                </div>
                <div v-if="interruptStore.skillTargetIds.length > 0" class="dock-accent">
                    已选目标：
                    {{ interruptStore.skillTargetIds.map((id) => snapshotStore.players[id]?.name || id).join('、') }}
                </div>
            </div>
            <footer class="dock-ft">
                <button type="button" class="dock-btn dock-btn--ghost" title="取消" @click="interruptStore.clearSkillMode()"><img class="dock-btn-ico" src="/assets/icons/cancel.webp" alt="" aria-hidden="true" />取消</button>
                <button
                    v-if="isManualTargetConfirmSkillFlow"
                    type="button"
                    class="dock-btn dock-btn--primary"
                    :disabled="!canConfirmSkill"
                    title="确认发动"
                    @click="confirmSkill()"
                >
                    确认发动 ✓
                </button>
            </footer>
        </div>

        <!-- 等待提示 -->
        <div v-else-if="interruptStore.waitingFor && !prompt" class="dock-panel dock-panel--waiting">
            <i class="dock-corner dock-corner--tl" aria-hidden="true" />
            <i class="dock-corner dock-corner--tr" aria-hidden="true" />
            <i class="dock-corner dock-corner--bl" aria-hidden="true" />
            <i class="dock-corner dock-corner--br" aria-hidden="true" />
            <header class="dock-hd">
                <span class="dock-hd-left">
                    <img class="dock-ico dock-ico--img" src="/assets/icons/waiting.webp" alt="" aria-hidden="true" draggable="false" />
                    <span class="dock-title">对方回合</span>
                </span>
            </header>
            <div class="dock-bd dock-bd--waiting">
                <img
                    v-if="interruptStore.waitingFor && characterArtSrc(snapshotStore.players[interruptStore.waitingFor]?.role)"
                    class="dock-wait-portrait"
                    :src="characterArtSrc(snapshotStore.players[interruptStore.waitingFor]?.role)"
                    alt=""
                />
                <span v-else class="dock-wait-portrait dock-wait-portrait--empty" aria-hidden="true">◷</span>
                <div class="dock-wait-name">{{ seatLabel(interruptStore.waitingFor) }} 正在操作</div>
                <div class="dock-wait-bar" aria-hidden="true"><i /></div>
            </div>
        </div>

        <!-- 非行动类 Prompt：在行动区内直接显示按钮操作 -->
        <div
            v-else-if="prompt && isPromptForMe && !isActionHubContext"
            class="dock-panel dock-panel--prompt"
            :class="{ 'dock-panel--danger': (prompt.presentation?.discard_reason === 'hand_overflow' || (!prompt.presentation?.discard_reason && prompt.presentation?.card_filter === 'overflow_discard')) }"
        >
            <i class="dock-corner dock-corner--tl" aria-hidden="true" />
            <i class="dock-corner dock-corner--tr" aria-hidden="true" />
            <i class="dock-corner dock-corner--bl" aria-hidden="true" />
            <i class="dock-corner dock-corner--br" aria-hidden="true" />
            <template v-if="(prompt.presentation?.discard_reason === 'hand_overflow' || (!prompt.presentation?.discard_reason && prompt.presentation?.card_filter === 'overflow_discard'))">
                <header class="dock-hd">
                    <span class="dock-hd-left">
                        <img class="dock-ico dock-ico--img" src="/assets/icons/overflow.webp" alt="" aria-hidden="true" draggable="false" />
                        <span class="dock-title">手牌超出上限</span>
                    </span>
                    <span v-if="myPlayer" class="dock-hd-meta dock-hd-meta--danger">{{ myHand.length || myPlayer.hand_count }}/{{ myPlayer.max_hand }}</span>
                </header>
                <div class="dock-bd dock-bd--overflow">
                    <div class="dock-hint dock-hint--big">从手牌中选择{{ prompt.min > 0 && prompt.min === prompt.max ? ` ${prompt.min} 张` : '' }}弃置</div>
                    <div v-if="prompt.min > 0 && prompt.min === prompt.max" class="dock-progress" aria-hidden="true">
                        <i
                            v-for="n in prompt.min"
                            :key="n"
                            class="dock-dot dock-dot--danger"
                            :class="{ 'dock-dot--on': n <= interruptStore.selectedHandIndexes.length }"
                        />
                    </div>
                    <div class="dock-counter">已选 {{ interruptStore.selectedHandIndexes.length }}{{ prompt.min > 0 && prompt.min === prompt.max ? ` / ${prompt.min}` : '' }}</div>
                    <div class="prompt-inline-host">
                        <PromptDialog />
                    </div>
                </div>
            </template>
            <div v-else class="prompt-inline-host">
                <PromptDialog />
            </div>
        </div>

        <!-- 行动区域 -->
        <div v-else-if="isActionHubContext" class="dock-panel dock-panel--hub action-hub-desktop">
            <i class="dock-corner dock-corner--tl" aria-hidden="true" />
            <i class="dock-corner dock-corner--tr" aria-hidden="true" />
            <i class="dock-corner dock-corner--bl" aria-hidden="true" />
            <i class="dock-corner dock-corner--br" aria-hidden="true" />
            <header class="dock-hd">
                <span class="dock-hd-left">
                    <i class="dock-dot-gold" aria-hidden="true" />
                    <span v-if="isExtraActionPrompt" class="dock-title">额外{{ hasActionPromptOption('magic') && !hasActionPromptOption('attack') ? '法术' : '攻击' }}<template v-if="myPlayer?.current_extra_action"> · 来自【{{ myPlayer.current_extra_action }}】</template></span>
                    <span v-else class="dock-title">你的回合 · 行动阶段</span>
                </span>
            </header>
            <div class="dock-bd">
                <div v-if="actionHubPromptNotice" class="dock-notice dock-notice--gold">
                    {{ actionHubPromptNotice }}
                </div>
                <div class="dock-medal-row action-hub-desktop-main">
                    <button
                        v-if="hasActionPromptOption('attack')"
                        type="button"
                        class="dock-medal"
                        data-testid="action-attack"
                        :title="actionPromptLabel('attack', '攻击')"
                        :aria-label="actionPromptLabel('attack', '攻击')"
                        @click="invokeActionHubOption('attack')"
                    >
                        <span class="dock-medal-ring">
                            <img class="dock-medal-ico dock-medal-ico--img" src="/assets/icons/attack.webp" alt="" aria-hidden="true" draggable="false" />
                        </span>
                        <span class="dock-medal-label">{{ actionPromptLabel('attack', '攻击') }}</span>
                    </button>
                    <button
                        v-if="hasActionPromptOption('magic')"
                        type="button"
                        class="dock-medal"
                        data-testid="action-magic"
                        :title="actionPromptLabel('magic', '法术')"
                        :aria-label="actionPromptLabel('magic', '法术')"
                        @click="invokeActionHubOption('magic')"
                    >
                        <span class="dock-medal-ring">
                            <img class="dock-medal-ico dock-medal-ico--img" src="/assets/icons/magic.webp" alt="" aria-hidden="true" draggable="false" />
                        </span>
                        <span class="dock-medal-label">{{ actionPromptLabel('magic', '法术') }}</span>
                    </button>
                    <button
                        v-if="showSpecialHubEntry"
                        type="button"
                        class="dock-medal"
                        data-testid="action-special"
                        :class="{ 'dock-medal--muted': !hasHubSpecialActions || isStartupSpecialLocked }"
                        :title="isStartupSpecialLocked ? '本回合已执行启动技能，特殊行动已禁用' : actionPromptLabel('special', '特殊')"
                        :aria-label="actionPromptLabel('special', '特殊')"
                        :disabled="!hasHubSpecialActions || isStartupSpecialLocked"
                        @click="openSpecialActionModal"
                    >
                        <span class="dock-medal-ring">
                            <img class="dock-medal-ico dock-medal-ico--img" src="/assets/icons/special.webp" alt="" aria-hidden="true" draggable="false" />
                        </span>
                        <span class="dock-medal-label">{{ actionPromptLabel('special', '特殊') }}</span>
                    </button>
                    <button
                        v-if="isActionSelectionPrompt && !isExtraActionPrompt && hasActionPromptOption('cannot_act')"
                        type="button"
                        class="dock-medal dock-medal--dim"
                        data-testid="action-cannot-act"
                        :title="isExtraActionPrompt ? '跳过' : cannotActButtonLabel"
                        :aria-label="isExtraActionPrompt ? '跳过' : cannotActButtonLabel"
                        @click="invokeActionHubOption('cannot_act')"
                    >
                        <span class="dock-medal-ring">
                            <img class="dock-medal-ico dock-medal-ico--img" src="/assets/icons/overflow.webp" alt="" aria-hidden="true" draggable="false" />
                        </span>
                        <span class="dock-medal-label">{{ isExtraActionPrompt ? '跳过' : cannotActButtonLabel }}</span>
                    </button>
                </div>
                <div
                    v-if="isActionSelectionPrompt && !hasActionPromptOption('attack') && !hasActionPromptOption('magic') && !hasHubSpecialActions && !hasActionPromptOption('cannot_act')"
                    class="dock-hint dock-hint--center"
                >
                    当前无可执行行动，请等待下一步结算
                </div>
            </div>
            <footer class="dock-ft">
                <span v-if="turnStateIndicators.length > 0" class="turn-state-indicators dock-ft-left">
                    <span
                        v-for="indicator in turnStateIndicators"
                        :key="indicator.key"
                        class="turn-chip"
                        :class="`turn-chip--${indicator.key}`"
                        :title="indicator.label"
                    >
                        {{ indicator.label }}
                    </span>
                </span>
                <button
                    v-if="!isActionSelectionPrompt"
                    type="button"
                    class="dock-btn dock-btn--dark dock-btn--right"
                    data-testid="action-pass"
                    title="结束回合"
                    aria-label="结束回合"
                    @click="invokeActionHubOption('pass')"
                ><img class="dock-btn-ico" src="/assets/icons/end-turn.webp" alt="" aria-hidden="true" />结束回合</button>
                <button
                    v-else-if="isExtraActionPrompt"
                    type="button"
                    class="dock-btn dock-btn--dark dock-btn--right"
                    data-testid="action-cannot-act"
                    title="放弃额外行动"
                    aria-label="放弃额外行动"
                    @click="invokeActionHubOption('cannot_act')"
                ><img class="dock-btn-ico" src="/assets/icons/skip-extra.webp" alt="" aria-hidden="true" />放弃额外行动</button>
            </footer>
        </div>

        <!-- 非我的回合 -->
        <div v-else class="dock-panel dock-panel--waiting">
            <i class="dock-corner dock-corner--tl" aria-hidden="true" />
            <i class="dock-corner dock-corner--tr" aria-hidden="true" />
            <i class="dock-corner dock-corner--bl" aria-hidden="true" />
            <i class="dock-corner dock-corner--br" aria-hidden="true" />
            <header class="dock-hd">
                <span class="dock-hd-left">
                    <img class="dock-ico dock-ico--img" src="/assets/icons/waiting.webp" alt="" aria-hidden="true" draggable="false" />
                    <span class="dock-title">对方回合</span>
                </span>
            </header>
            <div class="dock-bd dock-bd--waiting">
                <img
                    v-if="characterArtSrc(snapshotStore.players[snapshotStore.currentPlayer]?.role)"
                    class="dock-wait-portrait"
                    :src="characterArtSrc(snapshotStore.players[snapshotStore.currentPlayer]?.role)"
                    alt=""
                />
                <span v-else class="dock-wait-portrait dock-wait-portrait--empty" aria-hidden="true">◷</span>
                <div class="dock-wait-name">{{ snapshotStore.currentPlayer ? `${seatLabel(snapshotStore.currentPlayer)} 的回合` : '等待其他玩家回合' }}</div>
                <div class="dock-wait-bar" aria-hidden="true"><i /></div>
            </div>
        </div>

        <button
            v-if="debugAvailable"
            class="dock-gear debug-toggle-btn"
            type="button"
            title="调试"
            aria-label="打开调试控制台"
            @click="openDebugPanel"
        >
            <img class="dock-ico dock-ico--img" src="/assets/icons/settings.webp" alt="" aria-hidden="true" draggable="false" />
        </button>
    </div>

    <Teleport to="body">
        <Transition name="special-modal-fade">
            <div
                v-if="specialActionModalVisible"
                class="special-modal-mask"
                @click.self="closeSpecialActionModal"
            >
                <div class="special-modal-card">
                    <div class="special-modal-title">选择特殊行动</div>
                    <p class="special-modal-subtitle">资源向行动：用于补牌、合成和提炼能量。不可用选项会标明原因。</p>
                    <div class="special-modal-group-title">资源调度</div>
                    <div class="special-modal-actions">
                        <div
                            v-for="item in specialActionDisplayItems"
                            :key="item.id"
                            class="special-action-card"
                            :class="{ 'special-action-card--disabled': !item.available }"
                        >
                            <div class="special-action-head">
                                <span class="special-action-icon">{{ item.icon }}</span>
                                <div class="special-action-meta">
                                    <div class="special-action-label">{{ item.promptLabel }}</div>
                                    <div class="special-action-summary">{{ item.summary }}</div>
                                </div>
                            </div>
                            <div class="special-action-detail">{{ item.detail }}</div>
                            <div v-if="!item.available" class="special-action-reason">
                                不可用：{{ item.disabledReason }}
                            </div>
                            <button
                                class="btn-economy special-modal-btn"
                                :disabled="!item.available"
                                :class="{ 'special-modal-btn--disabled': !item.available }"
                                @click="chooseSpecialAction(item.id)"
                            >
                                {{ item.available ? '执行' : '不可执行' }}
                            </button>
                        </div>
                    </div>
                    <button class="btn-secondary special-modal-cancel" @click="closeSpecialActionModal">
                        取消
                    </button>
                </div>
            </div>
        </Transition>

        <Transition name="special-modal-fade">
            <div v-if="debugOpen" class="debug-modal-mask" @click.self="closeDebugPanel">
                <div class="debug-modal-card">
                    <div class="debug-modal-header">
                        <div>
                            <div class="debug-modal-title">调试控制台</div>
                            <div class="debug-modal-subtitle">可给任意角色设置基础效果、资源/治疗/指示物，并按条件补牌或强制弃牌。</div>
                        </div>
                        <button class="debug-modal-close" type="button" @click="closeDebugPanel">关闭</button>
                    </div>

                    <div class="debug-modal-controls">
                        <select v-model="debugTargetPlayerId" class="debug-select">
                            <option value="">选择目标角色</option>
                            <option v-for="player in debugTargetPlayers" :key="player.id" :value="player.id">
                                {{ player.name }} ({{ player.id }})
                            </option>
                        </select>
                        <div class="debug-status" v-if="debugStatus">{{ debugStatus }}</div>
                    </div>

                    <div class="debug-modal-body">
                        <div class="debug-manual">
                            <div class="debug-manual-title">基础效果与数值</div>
                            <div class="debug-manual-row">
                                <select v-model="debugEffectType" class="debug-select">
                                    <option value="Shield">圣盾</option>
                                    <option value="Poison">中毒</option>
                                    <option value="Weak">虚弱</option>
                                    <option value="PowerBlessing">威力赐福</option>
                                    <option value="SwiftBlessing">迅捷赐福</option>
                                </select>
                                <input v-model="debugEffectCount" type="number" min="0" class="debug-input" />
                                <button class="debug-mini-btn" type="button" @click="applyDebugEffect">设置基础效果</button>
                            </div>
                            <div class="debug-manual-row">
                                <select v-model="debugSetField" class="debug-select">
                                    <option value="gem">宝石</option>
                                    <option value="crystal">水晶</option>
                                    <option value="heal">治疗</option>
                                    <option value="max_heal">治疗上限</option>
                                </select>
                                <input v-model="debugSetValue" type="number" class="debug-input" />
                                <button class="debug-mini-btn" type="button" @click="applyDebugSet">设置</button>
                                <button
                                    class="debug-mini-btn"
                                    type="button"
                                    title="按当前人数，每人宝石在现有基础上 +3"
                                    @click="applyDebugAddGemsEveryone"
                                >
                                    全员+3♦
                                </button>
                            </div>
                            <div class="debug-manual-row">
                                <input v-model="debugTokenKey" class="debug-input" placeholder="指示物 key" />
                                <input v-model="debugTokenValue" type="number" class="debug-input" />
                                <button class="debug-mini-btn" type="button" @click="applyDebugToken">设置指示物</button>
                            </div>
                        </div>

                        <div class="debug-manual">
                            <div class="debug-manual-title">定向补牌</div>
                            <div class="debug-manual-row">
                                <select v-model="debugExclusiveRoleId" class="debug-select">
                                    <option value="">独有技角色</option>
                                    <option v-for="role in debugRoleList" :key="role.id" :value="role.id">
                                        {{ role.name }}
                                    </option>
                                </select>
                                <select v-model="debugExclusiveSkillId" class="debug-select">
                                    <option value="">独有技</option>
                                    <option v-for="skill in debugExclusiveSkillOptions" :key="skill.id" :value="skill.id">
                                        {{ skill.title }}
                                    </option>
                                </select>
                                <input v-model="debugExclusiveCount" type="number" min="1" class="debug-input" />
                                <button class="debug-mini-btn" type="button" @click="applyDebugExclusiveCard">获取独有技手牌</button>
                            </div>

                            <div class="debug-manual-row">
                                <select v-model="debugElement" class="debug-select">
                                    <option value="Water">水系</option>
                                    <option value="Fire">火系</option>
                                    <option value="Earth">土系</option>
                                    <option value="Wind">风系</option>
                                    <option value="Thunder">雷系</option>
                                    <option value="Light">光系</option>
                                    <option value="Dark">暗系</option>
                                </select>
                                <input v-model="debugElementCount" type="number" min="1" class="debug-input" />
                                <button class="debug-mini-btn" type="button" @click="applyDebugElementCards">获取指定系别手牌</button>
                            </div>

                            <div class="debug-manual-row">
                                <select v-model="debugFaction" class="debug-select">
                                    <option value="圣">圣</option>
                                    <option value="血">血</option>
                                    <option value="幻">幻</option>
                                    <option value="咏">咏</option>
                                    <option value="技">技</option>
                                </select>
                                <input v-model="debugFactionCount" type="number" min="1" class="debug-input" />
                                <button class="debug-mini-btn" type="button" @click="applyDebugFactionCards">获取指定命格手牌</button>
                            </div>

                            <div class="debug-manual-row">
                                <input v-model="debugMagicCardName" class="debug-input" placeholder="法术牌名称（如：魔弹）" />
                                <input v-model="debugMagicCardCount" type="number" min="1" class="debug-input" />
                                <button class="debug-mini-btn" type="button" @click="applyDebugMagicCard">获取指定法术牌</button>
                            </div>

                            <div class="debug-manual-row">
                                <input v-model="debugDiscardCount" type="number" min="1" class="debug-input" placeholder="弃牌数量" />
                                <button class="debug-mini-btn" type="button" @click="applyDebugDiscard">强制弃牌</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped>
/* ================= 根容器 ================= */
.action-panel-root {
    position: relative;
    width: 100%;
    height: 100%;
    min-height: 0;
    display: flex;
    flex-direction: column;
}

.action-panel-root--hub {
    z-index: 1202;
}

/* ================= 外壳：黑漆底板 + 镀铬细边 + 四角饰件 ================= */
.dock-panel {
    --dock-frame: var(--bt-hairline);
    --dock-frame-glow: transparent;
    position: relative;
    flex: 1 1 auto;
    min-height: 0;
    display: flex;
    flex-direction: column;
    border-radius: var(--bt-radius);
    background:
        radial-gradient(120% 80% at 50% 0%, var(--bt-panel-raised), transparent 70%),
        linear-gradient(180deg, var(--bt-bg-1), var(--bt-bg-0));
    box-shadow:
        inset 0 0 0 1px var(--dock-frame),
        inset 0 0 0 2px var(--bt-bg-0),
        inset 0 1px 0 2px var(--bt-hairline-soft),
        0 0 12px var(--dock-frame-glow);
    color: var(--bt-text);
    overflow: hidden;
    transition: box-shadow 0.2s ease;
}

/* 轮到我且需要操作：金框 */
.dock-panel--hub,
.action-mode-panel,
.skill-select-panel,
.skill-discard-panel,
.skill-target-panel,
.dock-panel--prompt {
    --dock-frame: var(--bt-active);
    --dock-frame-glow: var(--bt-active-glow);
}

/* 爆牌弃牌：红框 */
.dock-panel--prompt.dock-panel--danger {
    --dock-frame: var(--bt-danger);
    --dock-frame-glow: var(--bt-red-soft);
}

.dock-corner {
    position: absolute;
    z-index: 2;
    width: 10px;
    height: 10px;
    pointer-events: none;
    border-color: var(--bt-chrome);
    border-style: solid;
    border-width: 0;
}

.dock-corner::after {
    /* 角上的镀铬小菱形 */
    content: '';
    position: absolute;
    width: 4px;
    height: 4px;
    background: var(--bt-chrome-hi);
    transform: rotate(45deg);
    box-shadow: 0 0 0 1px var(--bt-bg-0);
}

.dock-corner--tl { top: 3px; left: 3px; border-top-width: 1px; border-left-width: 1px; border-top-left-radius: 3px; }
.dock-corner--tr { top: 3px; right: 3px; border-top-width: 1px; border-right-width: 1px; border-top-right-radius: 3px; }
.dock-corner--bl { bottom: 3px; left: 3px; border-bottom-width: 1px; border-left-width: 1px; border-bottom-left-radius: 3px; }
.dock-corner--br { bottom: 3px; right: 3px; border-bottom-width: 1px; border-right-width: 1px; border-bottom-right-radius: 3px; }
.dock-corner--tl::after { top: -3px; left: -3px; }
.dock-corner--tr::after { top: -3px; right: -3px; }
.dock-corner--bl::after { bottom: -3px; left: -3px; }
.dock-corner--br::after { bottom: -3px; right: -3px; }

/* ================= 头部 ================= */
.dock-hd {
    flex: 0 0 36px;
    height: 36px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 34px 0 12px;
    border-bottom: 1px solid var(--bt-hairline-soft);
    background: linear-gradient(180deg, var(--bt-panel-raised), transparent);
    min-width: 0;
}

.dock-hd-left {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    flex: 0 1 auto;
}

.dock-ico {
    width: 16px;
    height: 16px;
    flex: none;
    color: var(--bt-chrome);
}

.dock-panel--hub .dock-hd .dock-ico,
.action-mode-panel .dock-hd .dock-ico,
.skill-select-panel .dock-hd .dock-ico,
.skill-discard-panel .dock-hd .dock-ico,
.skill-target-panel .dock-hd .dock-ico {
    color: var(--bt-active);
}

.dock-panel--danger .dock-hd .dock-ico,
.dock-panel--danger .dock-title {
    color: var(--bt-danger);
}

.dock-dot-gold {
    width: 9px;
    height: 9px;
    flex: none;
    border-radius: 999px;
    background: var(--bt-active);
    box-shadow: 0 0 0 2px var(--bt-bg-0), 0 0 0 3px var(--bt-chrome-lo), 0 0 8px var(--bt-active-glow);
}

.dock-title {
    font-size: 13px;
    font-weight: 700;
    color: var(--bt-chrome-hi);
    letter-spacing: 0.04em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.dock-hd-meta {
    margin-left: auto;
    font-size: 11px;
    color: var(--bt-text-sub);
    white-space: nowrap;
}

.dock-hd-meta--danger {
    font-family: var(--bt-font-num);
    font-size: 14px;
    font-weight: 700;
    color: var(--bt-danger);
}

.dock-steps {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    color: var(--bt-text-dim);
    white-space: nowrap;
}

.dock-steps .on {
    color: var(--bt-active);
    font-weight: 700;
}

.dock-steps .done {
    color: var(--bt-select);
}

.dock-steps-sep {
    font-style: normal;
    color: var(--bt-chrome-lo);
}

/* 调试齿轮：固定在头部右侧 */
.dock-gear {
    position: absolute;
    z-index: 3;
    top: 8px;
    right: 9px;
    width: 20px;
    height: 20px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    border: none;
    border-radius: 999px;
    background: transparent;
    color: var(--bt-chrome-mid);
    cursor: pointer;
}

.dock-gear .dock-ico {
    color: inherit;
}

.dock-gear:hover,
.dock-gear:focus-visible {
    color: var(--bt-chrome-hi);
    outline: none;
    box-shadow: 0 0 0 1px var(--bt-hairline);
}

/* ================= 主体 ================= */
.dock-bd {
    flex: 1 1 auto;
    min-height: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 8px 12px;
    overflow-y: auto;
    overflow-x: hidden;
    scrollbar-width: thin;
    text-align: center;
}

.dock-bd--list {
    justify-content: flex-start;
    align-items: stretch;
    gap: 5px;
    text-align: left;
}

.dock-hint {
    font-size: 12px;
    line-height: 1.4;
    color: var(--bt-text-sub);
}

.dock-hint--sm {
    font-size: 11px;
    color: var(--bt-text-dim);
}

.dock-hint--big {
    font-size: 14px;
    font-weight: 700;
    color: var(--bt-text);
}

.dock-hint--warn {
    color: var(--bt-danger);
}

.dock-hint--desc {
    max-height: 4.2em;
    overflow-y: auto;
    white-space: pre-wrap;
    word-break: break-word;
    font-size: 11px;
    color: var(--bt-text-dim);
}

.dock-texts {
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.dock-accent {
    font-size: 12px;
    color: var(--bt-active);
}

.dock-notice {
    width: 100%;
    font-size: 11px;
    line-height: 1.35;
    padding: 4px 8px;
    border-radius: var(--bt-radius-sm);
    color: var(--bt-danger);
    background: var(--bt-bg-0);
    box-shadow: inset 0 0 0 1px var(--bt-red-soft);
}

.dock-notice--gold {
    color: var(--bt-active);
    box-shadow: inset 0 0 0 1px var(--bt-hairline-soft);
}

/* ---------- 镀铬圆章 ---------- */
.dock-medal-row {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    align-items: flex-start;
    gap: 6px clamp(8px, 1vw, 16px);
}

.dock-medal {
    --medal-size: clamp(52px, 4.6vw, 72px);
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    gap: 5px;
    padding: 0;
    border: none;
    background: transparent;
    color: var(--bt-chrome-hi);
    cursor: pointer;
}

.dock-medal-row:has(> :nth-child(4)) .dock-medal {
    --medal-size: clamp(44px, 3.6vw, 58px);
}

.dock-medal-ring {
    position: relative;
    width: var(--medal-size);
    height: var(--medal-size);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    padding: 3px;
    background: var(--bt-chrome-gradient);
    box-shadow:
        0 0 0 1px var(--bt-bg-0),
        0 0 0 2px var(--bt-chrome-lo),
        0 4px 10px var(--bt-bg-0);
    transition: transform 0.16s ease, box-shadow 0.16s ease;
}

.dock-medal-ring::before {
    /* 内圈黑漆盘面 */
    content: '';
    position: absolute;
    inset: 3px;
    border-radius: 999px;
    background: radial-gradient(circle at 50% 32%, var(--bt-blue-soft), var(--bt-bg-0) 72%);
    box-shadow:
        inset 0 0 0 1px var(--bt-bg-0),
        inset 0 0 0 2px var(--bt-hairline-soft),
        inset 0 -6px 12px var(--bt-bg-0);
}

.dock-ico--img {
    width: 18px;
    height: 18px;
    object-fit: contain;
    flex: none;
    filter: drop-shadow(0 0 3px rgba(80, 140, 255, 0.35));
    user-select: none;
}

.dock-btn-ico {
    width: 16px;
    height: 16px;
    object-fit: contain;
    flex: none;
    margin-right: 4px;
    user-select: none;
    pointer-events: none;
}

.dock-btn--primary .dock-btn-ico {
    /* 金色主按钮上用深色图标更清楚 */
    filter: brightness(0.25) contrast(1.4);
}

.dock-medal-ico--img {
    width: 78%;
    height: 78%;
    object-fit: contain;
    filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.8));
    user-select: none;
    pointer-events: none;
}

.dock-medal-ico {
    position: relative;
    width: 62%;
    height: 62%;
    color: var(--bt-chrome-hi);
    filter: drop-shadow(0 0 3px var(--bt-blue-soft));
}

.dock-medal-label {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.08em;
    color: var(--bt-chrome-hi);
    text-shadow: 0 1px 2px var(--bt-bg-0);
    white-space: nowrap;
}

.dock-medal:not(:disabled):hover .dock-medal-ring,
.dock-medal:focus-visible .dock-medal-ring {
    transform: translateY(-2px);
    box-shadow:
        0 0 0 1px var(--bt-bg-0),
        0 0 0 2px var(--bt-select),
        0 0 14px var(--bt-select-glow),
        0 6px 12px var(--bt-bg-0);
}

.dock-medal:focus-visible {
    outline: none;
}

.dock-medal:focus-visible .dock-medal-label {
    color: var(--bt-select);
}

.dock-medal--dim .dock-medal-ico {
    color: var(--bt-chrome-mid);
}

/* 禁用：压暗圆环 + 锁 */
.dock-medal:disabled,
.dock-medal--muted {
    cursor: not-allowed;
}

.dock-medal:disabled .dock-medal-ring,
.dock-medal--muted .dock-medal-ring {
    background: linear-gradient(180deg, var(--bt-chrome-mid), var(--bt-chrome-lo));
}

.dock-medal:disabled .dock-medal-ico,
.dock-medal--muted .dock-medal-ico {
    color: var(--bt-text-dim);
    filter: none;
}

.dock-medal:disabled .dock-medal-ring::after,
.dock-medal--muted .dock-medal-ring::after {
    /* 锁图标：右下角小圆 + 锁形 */
    content: '🔒';
    position: absolute;
    right: -2px;
    bottom: -2px;
    width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    border-radius: 999px;
    background: var(--bt-bg-0);
    box-shadow: 0 0 0 1px var(--bt-chrome-lo);
}

.dock-medal:disabled .dock-medal-label,
.dock-medal--muted .dock-medal-label {
    color: var(--bt-text-dim);
}

/* ---------- 流程图：牌槽 → 目标槽 ---------- */
.dock-flow {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
}

.dock-slot {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--bt-bg-0);
    color: var(--bt-text-dim);
    box-shadow: inset 0 0 0 1px var(--bt-hairline-soft);
}

.dock-slot--card {
    width: 38px;
    height: 52px;
    border-radius: var(--bt-radius-sm);
}

.dock-slot--target {
    width: 46px;
    height: 46px;
    border-radius: 999px;
}

.dock-slot-ico {
    width: 24px;
    height: 24px;
}

.dock-slot-art {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: inherit;
}

.dock-slot--filled {
    color: var(--bt-select);
    background: linear-gradient(180deg, var(--bt-blue-soft), var(--bt-bg-0));
    box-shadow: inset 0 0 0 1px var(--bt-select), 0 0 8px var(--bt-select-glow);
}

.dock-slot--pulse {
    color: var(--bt-active);
    box-shadow: inset 0 0 0 1px var(--bt-active), 0 0 0 0 var(--bt-active-glow);
    animation: dockSlotPulse 1.4s ease-in-out infinite;
}

.dock-flow-arrow {
    font-size: 12px;
    color: var(--bt-chrome-mid);
}

@keyframes dockSlotPulse {
    0%, 100% { box-shadow: inset 0 0 0 1px var(--bt-active), 0 0 0 0 var(--bt-active-glow); }
    50% { box-shadow: inset 0 0 0 1px var(--bt-active), 0 0 12px 2px var(--bt-active-glow); }
}

/* ---------- 进度点 ---------- */
.dock-progress {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 6px;
}

.dock-dot {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    background: var(--bt-bg-0);
    box-shadow: inset 0 0 0 1px var(--bt-chrome-mid);
    transition: background 0.16s ease, box-shadow 0.16s ease;
}

.dock-dot--on {
    background: var(--bt-active);
    box-shadow: 0 0 6px var(--bt-active-glow);
}

.dock-dot--danger.dock-dot--on {
    background: var(--bt-danger);
    box-shadow: 0 0 6px var(--bt-red-soft);
}

.dock-counter {
    font-family: var(--bt-font-num);
    font-size: 13px;
    color: var(--bt-text-sub);
}

/* ---------- 技能列表行 ---------- */
.sk {
    display: grid;
    grid-template-columns: 22px minmax(0, 1fr) auto;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 6px 8px;
    border: none;
    border-radius: var(--bt-radius-sm);
    background: var(--bt-panel-raised);
    box-shadow: inset 0 0 0 1px var(--bt-hairline-soft);
    color: var(--bt-text);
    text-align: left;
    cursor: pointer;
    transition: box-shadow 0.16s ease, background 0.16s ease;
}

.sk:not(:disabled):hover,
.sk:focus-visible {
    outline: none;
    background: linear-gradient(90deg, var(--bt-blue-soft), var(--bt-panel-raised));
    box-shadow: inset 0 0 0 1px var(--bt-select), 0 0 8px var(--bt-select-glow);
}

.sk-ico {
    width: 30px;
    height: 30px;
    flex: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    background: radial-gradient(circle at 50% 35%, var(--bt-bg-1), var(--bt-bg-0) 75%);
    box-shadow: 0 0 0 1px var(--bt-chrome-lo), inset 0 0 0 1px var(--bt-hairline-soft);
}

.sk-ico img {
    width: 22px;
    height: 22px;
    object-fit: contain;
}

.sk--disabled .sk-ico img {
    filter: grayscale(1) brightness(0.6);
}

.sk-main {
    display: flex;
    flex-direction: column;
    min-width: 0;
}

.sk-name {
    font-size: 12px;
    font-weight: 700;
    color: var(--bt-chrome-hi);
}

.sk-desc {
    font-size: 10px;
    line-height: 1.3;
    color: var(--bt-text-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.sk-reason {
    font-size: 10px;
    line-height: 1.3;
    color: var(--bt-danger);
}

.sk-cost {
    font-size: 10px;
    padding: 2px 7px;
    border-radius: 999px;
    white-space: nowrap;
    color: var(--bt-active);
    background: var(--bt-bg-0);
    box-shadow: inset 0 0 0 1px var(--bt-hairline-soft);
}

.sk--disabled {
    cursor: not-allowed;
    opacity: 0.5;
}

.sk--disabled .sk-ico {
    background: var(--bt-chrome-lo);
    color: var(--bt-text-dim);
}

/* ---------- 等待 ---------- */
.dock-bd--waiting {
    gap: 6px;
}

.dock-wait-portrait {
    width: 52px;
    height: 52px;
    border-radius: 999px;
    object-fit: cover;
    object-position: 50% 20%;
    background: var(--bt-bg-0);
    box-shadow: 0 0 0 2px var(--bt-bg-0), 0 0 0 3px var(--bt-chrome-mid);
}

.dock-wait-portrait--empty {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 22px;
    color: var(--bt-chrome-mid);
}

.dock-wait-name {
    font-size: 12px;
    color: var(--bt-text-sub);
}

.dock-wait-bar {
    position: relative;
    width: 70%;
    height: 3px;
    border-radius: 999px;
    overflow: hidden;
    background: var(--bt-bg-0);
    box-shadow: inset 0 0 0 1px var(--bt-hairline-soft);
}

.dock-wait-bar > i {
    position: absolute;
    inset: 0 auto 0 0;
    width: 35%;
    border-radius: 999px;
    background: linear-gradient(90deg, transparent, var(--bt-select), transparent);
    animation: dockWaitSweep 1.8s ease-in-out infinite;
}

@keyframes dockWaitSweep {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(300%); }
}

/* ---------- 爆牌 ---------- */
.dock-bd--overflow {
    justify-content: flex-start;
}

/* ================= 底栏 ================= */
.dock-ft {
    flex: 0 0 44px;
    height: 44px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 12px;
    border-top: 1px solid var(--bt-hairline-soft);
    background: linear-gradient(0deg, var(--bt-panel-raised), transparent);
    min-width: 0;
}

.dock-ft:empty {
    display: none;
}

.dock-ft-left {
    min-width: 0;
}

.dock-ft-center {
    flex: 1 1 auto;
    text-align: center;
    font-family: var(--bt-font-num);
    font-size: 12px;
    color: var(--bt-text-sub);
}

.dock-btn {
    height: 28px;
    padding: 0 14px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 2px;
    border: none;
    border-radius: 999px;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.04em;
    white-space: nowrap;
    cursor: pointer;
    transition: box-shadow 0.16s ease, filter 0.16s ease;
}

.dock-btn--right,
.dock-btn--primary {
    margin-left: auto;
}

.dock-btn--ghost {
    background: transparent;
    color: var(--bt-text-sub);
    box-shadow: inset 0 0 0 1px var(--bt-hairline-soft);
}

.dock-btn--ghost:hover,
.dock-btn--ghost:focus-visible {
    outline: none;
    color: var(--bt-chrome-hi);
    box-shadow: inset 0 0 0 1px var(--bt-hairline);
}

.dock-btn--ghost + .dock-btn--ghost.dock-btn--right {
    margin-left: auto;
}

.dock-btn--primary {
    color: var(--bt-bg-0);
    background: linear-gradient(180deg, var(--bt-chrome-hi), var(--bt-active) 45%, var(--bt-active));
    box-shadow: 0 0 0 1px var(--bt-bg-0), 0 0 10px var(--bt-active-glow);
}

.dock-btn--primary:not(:disabled):hover,
.dock-btn--primary:focus-visible {
    outline: none;
    filter: brightness(1.08);
    box-shadow: 0 0 0 1px var(--bt-bg-0), 0 0 0 2px var(--bt-active), 0 0 14px var(--bt-active-glow);
}

.dock-btn--primary:disabled {
    cursor: not-allowed;
    color: var(--bt-text-dim);
    background: var(--bt-chrome-lo);
    box-shadow: none;
}

.dock-btn--dark {
    color: var(--bt-chrome-hi);
    background: linear-gradient(180deg, var(--bt-panel-raised), var(--bt-bg-0));
    box-shadow: inset 0 0 0 1px var(--bt-hairline);
}

.dock-btn--dark:hover,
.dock-btn--dark:focus-visible {
    outline: none;
    box-shadow: inset 0 0 0 1px var(--bt-select), 0 0 10px var(--bt-select-glow);
}

/* 回合状态小标签 */
.turn-state-indicators {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    overflow: hidden;
    max-height: 40px;
}

.turn-chip {
    display: inline-flex;
    align-items: center;
    padding: 2px 7px;
    border-radius: 999px;
    font-size: 10px;
    line-height: 1.3;
    white-space: nowrap;
    color: var(--bt-text);
    background: var(--bt-bg-0);
    box-shadow: inset 0 0 0 1px var(--bt-hairline-soft);
}

.turn-chip--ml_dark_release_next_attack_bonus {
    color: var(--bt-danger);
    box-shadow: inset 0 0 0 1px var(--bt-red-soft);
}

.turn-chip--ml_dark_release_lock_turn {
    color: var(--bt-text-sub);
}

/* ================= 非行动类 Prompt ================= */
.prompt-inline-host {
    width: 100%;
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    padding: 6px;
}

.dock-bd--overflow .prompt-inline-host {
    padding: 0;
}

.prompt-inline-host :deep(.prompt-inline-root) {
    width: 100%;
}

.prompt-inline-host :deep(.prompt-inline-surface) {
    width: 100%;
    border: none !important;
    border-radius: 0;
    background: transparent !important;
    box-shadow: none !important;
    padding: 0;
}

/* ================= 特殊行动弹窗 ================= */
.special-modal-mask {
    position: fixed;
    inset: 0;
    z-index: 2150;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    background: var(--bt-panel);
    backdrop-filter: blur(4px);
}

.special-modal-card {
    width: min(430px, calc(100vw - 28px));
    border-radius: var(--bt-radius);
    background: linear-gradient(180deg, var(--bt-bg-1), var(--bt-bg-0));
    box-shadow: var(--bt-frame-shadow), inset 0 1px 0 var(--bt-hairline-soft);
    padding: 14px 12px 12px;
}

.special-modal-title {
    font-size: 15px;
    font-weight: 700;
    color: var(--bt-chrome-hi);
    text-align: center;
}

.special-modal-subtitle {
    margin-top: 6px;
    margin-bottom: 8px;
    text-align: center;
    font-size: 12px;
    color: var(--bt-text-sub);
    line-height: 1.4;
}

.special-modal-group-title {
    font-size: 11px;
    letter-spacing: 0.08em;
    color: var(--bt-active);
    font-weight: 700;
    margin-bottom: 8px;
}

.special-modal-actions {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.special-action-card {
    border-radius: var(--bt-radius-sm);
    background: var(--bt-panel-raised);
    box-shadow: inset 0 0 0 1px var(--bt-hairline-soft);
    padding: 8px;
}

.special-action-card--disabled {
    background: var(--bt-bg-0);
}

.special-action-head {
    display: flex;
    align-items: center;
    gap: 8px;
}

.special-action-icon {
    width: 24px;
    height: 24px;
    border-radius: 999px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--bt-bg-0);
    color: var(--bt-active);
    font-size: 14px;
    box-shadow: inset 0 0 0 1px var(--bt-hairline);
}

.special-action-meta {
    min-width: 0;
}

.special-action-label {
    font-size: 13px;
    font-weight: 700;
    color: var(--bt-chrome-hi);
    line-height: 1.1;
}

.special-action-summary,
.special-action-detail {
    margin-top: 2px;
    font-size: 11px;
    color: var(--bt-text-sub);
    line-height: 1.35;
}

.special-action-detail {
    margin-top: 6px;
    color: var(--bt-text-dim);
}

.special-action-reason {
    margin-top: 5px;
    font-size: 11px;
    color: var(--bt-danger);
    line-height: 1.35;
}

.special-modal-btn,
.special-modal-cancel {
    width: 100%;
    min-height: 32px;
    font-size: 12px;
    margin-top: 6px;
}

.special-modal-btn--disabled {
    opacity: 0.55;
    cursor: not-allowed;
}

.special-modal-cancel {
    margin-top: 10px;
}

.special-modal-fade-enter-active,
.special-modal-fade-leave-active {
    transition: opacity 0.18s ease;
}

.special-modal-fade-enter-from,
.special-modal-fade-leave-to {
    opacity: 0;
}

/* ================= 调试弹窗 ================= */
.debug-modal-mask {
    position: fixed;
    inset: 0;
    background: var(--bt-panel);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2400;
}

.debug-modal-card {
    width: min(980px, 94vw);
    max-height: 88vh;
    background: linear-gradient(180deg, var(--bt-bg-1), var(--bt-bg-0));
    border-radius: var(--bt-radius);
    box-shadow: var(--bt-frame-shadow);
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.debug-modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 16px 18px 10px;
    border-bottom: 1px solid var(--bt-hairline-soft);
}

.debug-modal-title {
    font-size: 16px;
    font-weight: 700;
    color: var(--bt-chrome-hi);
}

.debug-modal-subtitle {
    font-size: 12px;
    color: var(--bt-text-sub);
}

.debug-modal-close,
.debug-mini-btn {
    padding: 6px 12px;
    border-radius: var(--bt-radius-sm);
    border: 1px solid var(--bt-hairline-soft);
    background: var(--bt-panel-raised);
    color: var(--bt-text);
    font-size: 12px;
}

.debug-mini-btn {
    padding: 6px 10px;
}

.debug-modal-controls {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    padding: 12px 18px;
    border-bottom: 1px solid var(--bt-hairline-soft);
}

.debug-input,
.debug-select {
    background: var(--bt-bg-0);
    border: 1px solid var(--bt-hairline-soft);
    color: var(--bt-text);
    padding: 6px 10px;
    border-radius: var(--bt-radius-sm);
    font-size: 12px;
    min-width: 140px;
}

.debug-status {
    margin-left: auto;
    font-size: 12px;
    color: var(--bt-text-sub);
}

.debug-modal-body {
    padding: 12px 18px 18px;
    overflow: auto;
}

.debug-manual {
    margin-top: 16px;
    padding-top: 12px;
    border-top: 1px solid var(--bt-hairline-soft);
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.debug-manual-title {
    font-size: 12px;
    font-weight: 600;
    color: var(--bt-text);
}

.debug-manual-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
}

/* ================= 响应式 ================= */
@media (max-width: 640px) {
    .dock-hd {
        padding-left: 8px;
    }

    .dock-steps {
        display: none;
    }

    .dock-bd {
        padding: 6px 8px;
    }

    .dock-ft {
        padding: 0 8px;
    }

    .dock-btn {
        padding: 0 10px;
        font-size: 11px;
    }

    .prompt-inline-host :deep(.prompt-inline-grid--2),
    .prompt-inline-host :deep(.prompt-inline-grid--3),
    .prompt-inline-host :deep(.prompt-inline-grid--4) {
        grid-template-columns: 1fr;
    }

    .special-modal-card {
        width: min(360px, calc(100vw - 20px));
        padding: 12px 10px 10px;
    }
}

@media (prefers-reduced-motion: reduce) {
    .dock-panel,
    .dock-medal-ring,
    .dock-dot,
    .sk,
    .dock-btn {
        transition: none;
    }

    .dock-medal:not(:disabled):hover .dock-medal-ring,
    .dock-medal:focus-visible .dock-medal-ring {
        transform: none;
    }

    .dock-slot--pulse {
        animation: none;
        box-shadow: inset 0 0 0 1px var(--bt-active), 0 0 8px var(--bt-active-glow);
    }

    .dock-wait-bar > i {
        animation: none;
        width: 100%;
        opacity: 0.5;
    }
}
</style>
