<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useBattleInteractionState } from '../composables/useBattleInteractionState'
import { useSubmitAction } from '../composables/useSubmitAction'
import type { AvailableSkill, PromptOption, PlayerView } from '../types/game'
import CardComponent from './CardComponent.vue'
import PromptDialog from './PromptDialog.vue'

const interruptStore = useInterruptStore()
const sessionStore = useSessionStore()
const snapshotStore = useSnapshotStore()
const actions = useSubmitAction()
const {
    myPlayer,
    myHand,
    targetablePlayers,
    effectiveAvailableSkills,
    canConfirmSkill,
    cardMatchesExclusive,
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

type MainActionIconId = 'attack' | 'magic' | 'special' | 'cannot_act' | 'skill' | 'pass' | 'cancel' | 'confirm' | 'card'

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

const MAIN_ACTION_IMAGE_CANDIDATES: Record<MainActionIconId, string[]> = {
    attack: ['/assets/ui/action_attack.png'],
    magic: ['/assets/ui/action_magic_btn.png'],
    special: ['/assets/ui/action_special_btn.png'],
    cannot_act: ['/assets/ui/action_cannot_act.png'],
    skill: ['/assets/ui/action_skill.png'],
    pass: ['/assets/ui/action_pass_btn.png'],
    cancel: ['/assets/ui/action_cancel_btn.png'],
    confirm: ['/assets/ui/action_confirm.png'],
    card: ['/assets/ui/action_card.png'],
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
const waitingName = computed(() => {
    if (!interruptStore.waitingFor) return ''
    return snapshotStore.players[interruptStore.waitingFor]?.name || interruptStore.waitingFor
})
const specialActionModalVisible = ref(false)
const isIdleMainTurnPanel = computed(() =>
    isMyTurn.value &&
    !prompt.value &&
    interruptStore.actionMode === 'none' &&
    interruptStore.skillMode === 'none'
)

function isActionSelectionPromptMessage(message: string): boolean {
    const normalized = String(message || '').trim()
    // 仅识别主流程“请选择行动类型”提示；
    // 避免把【圣疗】“请选择额外行动类型”误判为行动枢纽。
    return normalized.includes('请选择行动类型')
}

type ActionHubOptionId = 'attack' | 'magic' | 'special' | 'cannot_act'

function normalizeActionHubOptionId(option: PromptOption): ActionHubOptionId | null {
    const id = (option.id || '').trim()
    if (id === 'attack' || id === 'magic' || id === 'special' || id === 'cannot_act') {
        return id
    }
    const label = (option.label || '').trim()
    if (!label) return null
    if (label.includes('攻击行动') || label.includes('攻击')) return 'attack'
    if (label.includes('法术行动') || label.includes('法术')) return 'magic'
    if (label.includes('跳过额外行动') || label.includes('无法行动')) return 'cannot_act'
    if (label.includes('特殊')) return 'special'
    return null
}

const isActionSelectionPrompt = computed(() => {
    const p = prompt.value
    if (!p || !isPromptForMe.value) return false
    if (p.ui_mode === 'action_hub') return true
    if (p.type !== 'confirm') return false
    if (!isActionSelectionPromptMessage(p.message || '')) return false
    return (p.options || []).some((opt) => normalizeActionHubOptionId(opt) !== null)
})

const isActionHubContext = computed(() =>
    (isIdleMainTurnPanel.value || isActionSelectionPrompt.value) &&
    interruptStore.actionMode === 'none' &&
    interruptStore.skillMode === 'none'
)

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

const specialActionOptions = computed<PromptOption[]>(() => {
    if (snapshotStore.hasPerformedStartup) {
        return []
    }
    if (isActionSelectionPrompt.value) {
        const fromSpecial = prompt.value?.special_options ?? []
        if (fromSpecial.length > 0) return fromSpecial
        // 兼容旧后端：若仍直接下发 buy/synthesize/extract，则前端照样合并展示
        return (prompt.value?.options ?? []).filter((opt) =>
            opt.id === 'buy' || opt.id === 'synthesize' || opt.id === 'extract'
        )
    }
    return [
        { id: 'buy', label: '购买' },
        { id: 'synthesize', label: '合成' },
        { id: 'extract', label: '提炼' },
    ]
})

const hasHubSpecialActions = computed(() => specialActionOptions.value.length > 0)
const showSpecialHubEntry = computed(() => isActionHubContext.value)
const isStartupSpecialLocked = computed(() => snapshotStore.hasPerformedStartup)
const isExtraActionPrompt = computed(() => {
    const message = prompt.value?.message ?? ''
    return message.includes('额外攻击行动') || message.includes('额外法术行动') || message.includes('额外行动阶段')
})
const cannotActButtonLabel = computed(() =>
    isExtraActionPrompt.value
        ? actionPromptLabel('cannot_act', '跳过额外行动')
        : actionPromptLabel('cannot_act', '无法行动')
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

const mainActionImageIndex = ref<Record<MainActionIconId, number>>({
    attack: 0,
    magic: 0,
    special: 0,
    cannot_act: 0,
    skill: 0,
    pass: 0,
    cancel: 0,
    confirm: 0,
    card: 0,
})
const mainActionImageFailed = ref<Record<MainActionIconId, boolean>>({
    attack: false,
    magic: false,
    special: false,
    cannot_act: false,
    skill: false,
    pass: false,
    cancel: false,
    confirm: false,
    card: false,
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

function mainActionButtonImage(optionId: MainActionIconId): string {
    const candidates = MAIN_ACTION_IMAGE_CANDIDATES[optionId]
    const idx = mainActionImageIndex.value[optionId]
    const image = candidates[Math.min(idx, candidates.length - 1)]
    return image || ''
}

function isMainActionImageReady(optionId: MainActionIconId): boolean {
    return !mainActionImageFailed.value[optionId]
}

function onMainActionImageError(optionId: MainActionIconId) {
    const candidates = MAIN_ACTION_IMAGE_CANDIDATES[optionId]
    const nextIndex = mainActionImageIndex.value[optionId] + 1
    if (nextIndex < candidates.length) {
        mainActionImageIndex.value[optionId] = nextIndex
        return
    }
    mainActionImageFailed.value[optionId] = true
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

function isMagicMissilePromptMessage(): boolean {
    return (prompt.value?.message ?? '').includes('魔弹')
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
        actions.submitCancel()
        return
    } else if (optionId === 'confirm') {
        actions.submitConfirm()
    } else if (optionId === 'take') {
        actions.submitRespondTake()
    } else if (optionId === 'counter') {
        if (!actions.submitRespondCounter(isMagicMissilePromptMessage())) return
    } else if (optionId === 'defend') {
        if (!actions.submitRespondDefend()) return
    } else if (optionId === 'yes' || optionId === 'no') {
        // 魔弹融合等确认选项：yes=0, no=1
        actions.submitSelect([optionId === 'yes' ? 0 : 1])
    } else if (optionId === 'normal' || optionId === 'reverse') {
        // 魔弹掌控方向选择：normal=0, reverse=1
        actions.submitSelect([optionId === 'normal' ? 0 : 1])
    } else if (prompt.value.type === 'choose_skill') {
        const idx = prompt.value.options.findIndex((o: { id: string }) => o.id === optionId)
        if (idx >= 0) {
            actions.submitSelect([idx])
        } else {
            actions.submitCancel()
            return
        }
    } else {
        const index = parseInt(optionId)
        if (!isNaN(index)) {
            actions.submitSelect([index])
        } else {
            actions.submitAction({
                player_id: sessionStore.myPlayerId,
                type: 'Select',
                skill_id: optionId
            })
        }
    }
    // 不在此处清除 prompt：等待后端 state_update（成功）或新 prompt 到达后再清除
    // 若后端报错，prompt 保持显示，用户可重新选择
}

function backFromMagicCard() {
    interruptStore.setMagicSubChoice('none')
    interruptStore.setSelectedCardForAction(null)
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

const BOARD_GUIDED_SKILL_IDS = new Set([
    'ss_soul_mirror',
    'water_seal',
    'fire_seal',
    'earth_seal',
    'wind_seal',
    'thunder_seal',
])
const SKILL_REQUIRE_MANUAL_TARGET_CONFIRM_IDS = new Set([
    'water_seal',
    'fire_seal',
    'earth_seal',
    'wind_seal',
    'thunder_seal',
])
const isBoardGuidedSkillFlow = computed(() => {
    const skillId = interruptStore.selectedSkill?.id
    if (!skillId) return false
    return BOARD_GUIDED_SKILL_IDS.has(skillId)
})
const isManualTargetConfirmSkillFlow = computed(() => {
    const skillId = interruptStore.selectedSkill?.id
    if (!skillId) return false
    return SKILL_REQUIRE_MANUAL_TARGET_CONFIRM_IDS.has(skillId)
})

function selectSkill(skill: AvailableSkill) {
    if (!canSelectSkill(skill)) {
        interruptStore.showError(skillDisabledReason(skill))
        return
    }
    interruptStore.setSelectedSkill(skill)
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
    if (skill.discard_element) return card.element === skill.discard_element
    if (skill.id === 'magic_bullet_fusion') return card.element === 'Fire' || card.element === 'Earth'
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
    if (skill.id === 'onmyoji_shikigami_descend') {
        return '需要弃置2张命格相同的手牌才能发动。'
    }
    if (skill.id === 'magic_blast') {
        return '手牌中没有法术牌，无法发动【魔爆冲击】。'
    }
    if (skill.id === 'magic_bullet_fusion') {
        return '需要弃置1张火系或地系牌，才能发动【魔弹融合】。'
    }
    if (skill.cost_discards > 0) {
        const required = requiredDiscardCount(skill)
        return `可用于弃置的手牌不足，至少需要 ${required} 张。`
    }
    return '当前不可发动该技能'
}

function proceedAfterDiscard(skill: AvailableSkill) {
    // target_type=0 (None): 无需目标，直接发动
    if (skill.target_type === 0) {
        const selections = interruptStore.skillDiscardIndices.length > 0 ? [...interruptStore.skillDiscardIndices] : undefined
        actions.submitUseSkill(skill.id, [], selections, { clearSkillMode: true })
        return
    }
    // target_type=1 (Self): 自动选中自己并发动
    if (skill.target_type === 1) {
        const selections = interruptStore.skillDiscardIndices.length > 0 ? [...interruptStore.skillDiscardIndices] : undefined
        actions.submitUseSkill(skill.id, [sessionStore.myPlayerId], selections, { clearSkillMode: true })
        return
    }
    interruptStore.setSkillMode('choosing_target')
}

function confirmSkillDiscard() {
    const skill = interruptStore.selectedSkill
    if (!skill) return
    const required = requiredDiscardCount(skill)
    if (interruptStore.skillDiscardIndices.length < required) {
        interruptStore.showError(`请选择 ${required} 张牌`)
        return
    }
    proceedAfterDiscard(skill)
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
    const selections = interruptStore.skillDiscardIndices.length > 0 ? [...interruptStore.skillDiscardIndices] : undefined
    actions.submitUseSkill(skill.id, [...interruptStore.skillTargetIds], selections, { clearSkillMode: true })
}

watch(
    () => [interruptStore.skillMode, interruptStore.selectedSkill?.id, interruptStore.skillDiscardIndices.length] as const,
    ([mode, skillId, selectedCount]) => {
        if (mode !== 'choosing_discard' || !skillId || !BOARD_GUIDED_SKILL_IDS.has(skillId)) return
        const skill = interruptStore.selectedSkill
        if (!skill) return
        const required = requiredDiscardCount(skill)
        if (required > 0 && selectedCount >= required) {
            proceedAfterDiscard(skill)
        }
    }
)

function skillCostText(skill: AvailableSkill): string {
    if (skill.id === 'priest_water_power') {
        return '弃1水牌+交1手牌(若有)'
    }
    const parts: string[] = []
    if (skill.cost_gem > 0) parts.push(`${skill.cost_gem}宝石`)
    if (skill.cost_crystal > 0) parts.push(`${skill.cost_crystal}水晶`)
    if (skill.cost_discards > 0) {
        parts.push(skill.require_exclusive ? `弃${skill.cost_discards}独有牌` : `弃${skill.cost_discards}牌`)
    } else if (skill.require_exclusive) {
        parts.push('专属技能卡')
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
        interruptStore.showError('独有牌数量需为 > 0 的数字')
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

function isCardSelectableForSkillDiscard(card: { type: string; element: string; faction?: string; exclusive_char1?: string; exclusive_char2?: string; exclusive_skill1?: string; exclusive_skill2?: string }): boolean {
    const skill = interruptStore.selectedSkill
    if (!skill) return false
    if (skill.id === 'priest_water_power') {
        const selected = interruptStore.skillDiscardIndices
            .map((i) => myHand.value[i])
            .filter((c): c is NonNullable<typeof c> => !!c)
        if (selected.length === 0) {
            return card.element === 'Water'
        }
        // 第一张已是水系后，第二张可为任意手牌（但上限仍由 requiredDiscardCount 控制）。
        return selected[0]?.element === 'Water'
    }
    // 独有技：必须使用卡牌下标了该技能名的牌
    if (skill.require_exclusive) {
        const roleId = resolveMyRoleIdForExclusive()
        if (!roleId || !cardMatchesExclusive(card, roleId, skill.title)) return false
    }
    if (skill.discard_type && card.type !== skill.discard_type) return false
    // 元素要求
    if (skill.discard_element) return card.element === skill.discard_element
    if (skill.id === 'magic_bullet_fusion') {
        return card.element === 'Fire' || card.element === 'Earth'
    }
    // 阴阳师：式神降临需要两张同命格手牌
    if (skill.id === 'onmyoji_shikigami_descend') {
        if (!card.faction) return false
        const selected = interruptStore.skillDiscardIndices
            .map((i) => myHand.value[i])
            .filter((c): c is NonNullable<typeof c> => !!c)
        if (selected.length > 0) {
            const reqFaction = selected[0]?.faction
            if (reqFaction && card.faction !== reqFaction) return false
        }
    }
    return true
}

function toggleSkillDiscardCard(idx: number) {
    const skill = interruptStore.selectedSkill
    if (!skill) return
    const required = requiredDiscardCount(skill)
    const card = myHand.value[idx]
    if (!card) return
    // 独有技：必须使用卡牌下标了该技能名的牌
    if (skill.require_exclusive) {
        const roleId = resolveMyRoleIdForExclusive()
        if (!roleId || !cardMatchesExclusive(card, roleId, skill.title)) {
            interruptStore.showError('必须使用标有该技能名的独有牌')
            return
        }
    }
    // 检查元素要求
    if (skill.discard_element && card.element !== skill.discard_element) {
        interruptStore.showError(`需要弃置${elementName(skill.discard_element)}牌`)
        return
    }
    if (skill.discard_type && card.type !== skill.discard_type) {
        interruptStore.showError(`需要弃置${skill.discard_type === 'Magic' ? '法术' : '攻击'}牌`)
        return
    }
    if (skill.id === 'priest_water_power' && !interruptStore.skillDiscardIndices.includes(idx)) {
        const selected = interruptStore.skillDiscardIndices
            .map((i) => myHand.value[i])
            .filter((c): c is NonNullable<typeof c> => !!c)
        if (selected.length === 0 && card.element !== 'Water') {
            interruptStore.showError('水之神力第一张需弃置水系牌')
            return
        }
        if (selected.length > 0 && selected[0]?.element !== 'Water') {
            interruptStore.showError('水之神力第一张需弃置水系牌')
            return
        }
    }
    if (skill.id === 'magic_bullet_fusion' && card.element !== 'Fire' && card.element !== 'Earth') {
        interruptStore.showError('魔弹融合需要弃置1张火系或地系牌')
        return
    }
    // 阴阳师：式神降临必须弃置两张同命格手牌
    if (skill.id === 'onmyoji_shikigami_descend' && !interruptStore.skillDiscardIndices.includes(idx)) {
        if (!card.faction) {
            interruptStore.showError('式神降临需要弃置有命格的手牌')
            return
        }
        const selected = interruptStore.skillDiscardIndices
            .map((i) => myHand.value[i])
            .filter((c): c is NonNullable<typeof c> => !!c)
        if (selected.length > 0) {
            const reqFaction = selected[0]?.faction
            if (reqFaction && card.faction !== reqFaction) {
                interruptStore.showError('式神降临需要弃置2张命格相同的手牌')
                return
            }
        }
    }
    // 如果已选满且不是取消选择，不允许继续选
    if (!interruptStore.skillDiscardIndices.includes(idx) && interruptStore.skillDiscardIndices.length >= required) {
        interruptStore.showError(`最多选择 ${required} 张牌`)
        return
    }
    interruptStore.toggleSkillDiscard(idx)
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
        <div v-if="interruptStore.actionMode !== 'none'" class="space-y-3 action-mode-panel">
            <!-- 法术行动：先选择 出牌 或 发动技能 -->
            <template v-if="interruptStore.actionMode === 'magic' && interruptStore.magicSubChoice === 'none'">
                <div class="text-amber-400 text-sm font-bold">✨ 法术行动</div>
                <div class="text-xs text-gray-400">发动法术有两种方式：打出法术牌或发动角色技能</div>
                <div class="flex gap-3 justify-center mt-2">
                    <button class="action-image-btn w-16 sm:w-20" title="打出法术牌" @click="interruptStore.setMagicSubChoice('card')">
                        <img v-if="isMainActionImageReady('card')" class="action-image-btn-fill" :src="mainActionButtonImage('card')" alt="" @error="onMainActionImageError('card')" />
                        <span v-else class="action-image-fallback-text">牌</span>
                        <span class="action-image-btn-label">打出法术牌</span>
                    </button>
                    <button
                        v-if="effectiveAvailableSkills.length > 0"
                        class="action-image-btn w-16 sm:w-20"
                        title="发动技能"
                        @click="interruptStore.setSkillMode('choosing_skill'); interruptStore.clearActionMode()"
                    >
                        <img v-if="isMainActionImageReady('skill')" class="action-image-btn-fill" :src="mainActionButtonImage('skill')" alt="" @error="onMainActionImageError('skill')" />
                        <span v-else class="action-image-fallback-text">技</span>
                        <span class="action-image-btn-label">发动技能</span>
                    </button>
                    <button class="action-image-btn w-16 sm:w-20" title="取消" @click="interruptStore.clearActionMode()">
                        <img v-if="isMainActionImageReady('cancel')" class="action-image-btn-fill" :src="mainActionButtonImage('cancel')" alt="" @error="onMainActionImageError('cancel')" />
                        <span v-else class="action-image-fallback-text">消</span>
                        <span class="action-image-btn-label">取消</span>
                    </button>
                </div>
            </template>
            <!-- 攻击模式 或 法术已选「出牌」：选牌 + 选目标 -->
            <template v-else>
                <div class="flex items-center justify-between">
          <span class="text-amber-400 text-sm font-bold">
            {{ interruptStore.actionMode === 'attack' ? '⚔️ 攻击模式' : '✨ 法术模式' }}
          </span>
                    <span class="step-indicator">
            {{ interruptStore.selectedCardForAction === null ? '步骤 1/2' : '步骤 2/2' }} · {{ interruptStore.selectedCardForAction === null ? '选牌' : '选目标' }}
          </span>
                </div>
                <div v-if="interruptStore.selectedCardForAction !== null" class="space-y-2">
                    <div
                        v-if="hasActionTargets"
                        class="text-xs text-gray-400"
                    >
                        请直接点击战场角色立绘完成{{ interruptStore.actionMode === 'attack' ? '攻击' : '施法' }}。
                    </div>
                    <div v-else class="text-xs text-gray-400">当前法术无需手动选目标，将按规则自动结算。</div>
                    <div v-if="hasStealthBlockedAttackTarget" class="text-[11px] text-gray-400">
                        潜行状态无法选中
                    </div>
                </div>
                <div v-else class="text-xs text-gray-400 py-1">
                    先在下方手牌选一张{{ interruptStore.actionMode === 'attack' ? '攻击' : '法术' }}牌
                </div>
                <div class="flex gap-3 justify-center mt-2">
                    <button class="action-image-btn w-16 sm:w-20" :title="interruptStore.actionMode === 'magic' ? '返回' : '取消'" @click="interruptStore.actionMode === 'magic' ? backFromMagicCard() : interruptStore.clearActionMode()">
                        <img v-if="isMainActionImageReady('cancel')" class="action-image-btn-fill" :src="mainActionButtonImage('cancel')" alt="" @error="onMainActionImageError('cancel')" />
                        <span v-else class="action-image-fallback-text">{{ interruptStore.actionMode === 'magic' ? '返' : '消' }}</span>
                        <span class="action-image-btn-label">{{ interruptStore.actionMode === 'magic' ? '返回' : '取消' }}</span>
                    </button>
                    <button
                        v-if="interruptStore.actionMode === 'magic' && effectiveAvailableSkills.length > 0"
                        class="action-image-btn w-16 sm:w-20"
                        title="改用技能"
                        @click="interruptStore.setSkillMode('choosing_skill'); interruptStore.clearActionMode()"
                    >
                        <img v-if="isMainActionImageReady('skill')" class="action-image-btn-fill" :src="mainActionButtonImage('skill')" alt="" @error="onMainActionImageError('skill')" />
                        <span v-else class="action-image-fallback-text">技</span>
                        <span class="action-image-btn-label">改用技能</span>
                    </button>
                </div>
            </template>
        </div>

        <!-- 技能发动流程：选择技能 -->
        <div v-else-if="interruptStore.skillMode === 'choosing_skill'" class="space-y-3 skill-select-panel">
            <div class="text-amber-400 text-sm font-bold">选择要发动的技能</div>
            <div class="flex flex-col gap-2">
                <button
                    v-for="skill in effectiveAvailableSkills"
                    :key="skill.id"
                    class="btn-skill px-4 py-2.5 rounded-lg text-sm text-left w-full"
                    :class="{ 'skill-btn-disabled': !canSelectSkill(skill) }"
                    :title="canSelectSkill(skill) ? skill.description : skillDisabledReason(skill)"
                    :disabled="!canSelectSkill(skill)"
                    @click="selectSkill(skill)"
                >
                    <div class="flex items-center justify-between">
                        <span class="font-semibold">{{ skill.title }}</span>
                        <span class="text-[10px] opacity-70 ml-2 whitespace-nowrap">{{ skillCostText(skill) }}</span>
                    </div>
                    <span v-if="skill.description" class="block text-xs opacity-80 mt-0.5" :title="skill.description">{{ skill.description }}</span>
                    <span v-if="!canSelectSkill(skill)" class="block text-[11px] text-gray-400 mt-1">{{ skillDisabledReason(skill) }}</span>
                </button>
            </div>
            <div class="flex gap-3 justify-center mt-2">
                <button class="action-image-btn w-16 sm:w-20" title="取消" @click="interruptStore.clearSkillMode()">
                    <img v-if="isMainActionImageReady('cancel')" class="action-image-btn-fill" :src="mainActionButtonImage('cancel')" alt="" @error="onMainActionImageError('cancel')" />
                    <span v-else class="action-image-fallback-text">消</span>
                    <span class="action-image-btn-label">取消</span>
                </button>
            </div>
        </div>

        <!-- 技能发动流程：选择弃牌 -->
        <div v-else-if="interruptStore.skillMode === 'choosing_discard' && interruptStore.selectedSkill" class="space-y-3 skill-discard-panel">
            <div class="flex items-center justify-between">
                <span class="text-amber-400 text-sm font-bold">{{ interruptStore.selectedSkill.title }}</span>
                <span class="step-indicator">
          {{ interruptStore.skillDiscardIndices.length }}/{{ requiredDiscardCount(interruptStore.selectedSkill) }}
        </span>
            </div>
            <template v-if="isBoardGuidedSkillFlow">
                <div class="text-xs text-gray-400">
                    <span v-if="interruptStore.selectedSkill.require_exclusive">
                        弃标有「{{ interruptStore.selectedSkill.title }}」的独有牌
                    </span>
                    <span v-else>
                        请在下方手牌区选择要弃置的牌
                    </span>
                    <span class="text-amber-300">（已选 {{ interruptStore.skillDiscardIndices.length }}/{{ requiredDiscardCount(interruptStore.selectedSkill) }}）</span>
                </div>
                <div class="text-[11px] text-gray-500">选满后将自动进入目标选择</div>
                <div class="flex gap-3 justify-center">
                    <button class="action-image-btn w-16 sm:w-20" title="取消" @click="interruptStore.clearSkillMode()">
                        <img v-if="isMainActionImageReady('cancel')" class="action-image-btn-fill" :src="mainActionButtonImage('cancel')" alt="" @error="onMainActionImageError('cancel')" />
                        <span v-else class="action-image-fallback-text">消</span>
                        <span class="action-image-btn-label">取消</span>
                    </button>
                </div>
            </template>
            <template v-else>
                <div class="text-xs text-gray-400">
                    请选择要弃置的牌
                    <span v-if="interruptStore.selectedSkill.require_exclusive" class="text-amber-300">
          （须为标有「{{ interruptStore.selectedSkill.title }}」的独有牌）
        </span>
                    <span v-else-if="interruptStore.selectedSkill.discard_element" class="text-amber-300">
          （需要{{ elementName(interruptStore.selectedSkill.discard_element) }}牌）
                    </span>
                    <span v-else-if="interruptStore.selectedSkill.discard_type" class="text-amber-300">
          （需要{{ interruptStore.selectedSkill.discard_type === 'Magic' ? '法术牌' : '攻击牌' }}）
                    </span>
                    <span v-else-if="interruptStore.selectedSkill.id === 'priest_water_power'" class="text-amber-300">
          （第一张需水系；若仍有手牌，第二张将交给目标队友）
                    </span>
                    <span v-else-if="interruptStore.selectedSkill.id === 'magic_bullet_fusion'" class="text-amber-300">
          （需要火系或地系牌）
                    </span>
                    <span v-else-if="interruptStore.selectedSkill.id === 'onmyoji_shikigami_descend'" class="text-amber-300">
          （需要2张命格相同的手牌）
                    </span>
                </div>
                <div class="flex gap-1 flex-wrap justify-center skill-discard-card-row">
                    <CardComponent
                        v-for="(card, idx) in myHand"
                        :key="idx"
                        :card="card"
                        :index="idx"
                        medium
                        :selectable="isCardSelectableForSkillDiscard(card)"
                        :selected="interruptStore.skillDiscardIndices.includes(idx)"
                        @click="toggleSkillDiscardCard(idx)"
                    />
                </div>
            <div class="flex gap-3 justify-center mt-2">
                <button
                    class="action-image-btn w-16 sm:w-20"
                    :class="{ 'opacity-50 cursor-not-allowed': interruptStore.skillDiscardIndices.length < requiredDiscardCount(interruptStore.selectedSkill) }"
                    :title="`确认弃牌 (${interruptStore.skillDiscardIndices.length}/${requiredDiscardCount(interruptStore.selectedSkill)})`"
                    :disabled="interruptStore.skillDiscardIndices.length < requiredDiscardCount(interruptStore.selectedSkill)"
                    @click="confirmSkillDiscard()"
                >
                    <img v-if="isMainActionImageReady('confirm')" class="action-image-btn-fill" :src="mainActionButtonImage('confirm')" alt="" @error="onMainActionImageError('confirm')" />
                    <span v-else class="action-image-fallback-text">确</span>
                    <span class="action-image-btn-label">确认弃牌 ({{ interruptStore.skillDiscardIndices.length }}/{{ requiredDiscardCount(interruptStore.selectedSkill) }})</span>
                </button>
                <button class="action-image-btn w-16 sm:w-20" title="取消" @click="interruptStore.clearSkillMode()">
                    <img v-if="isMainActionImageReady('cancel')" class="action-image-btn-fill" :src="mainActionButtonImage('cancel')" alt="" @error="onMainActionImageError('cancel')" />
                    <span v-else class="action-image-fallback-text">消</span>
                    <span class="action-image-btn-label">取消</span>
                </button>
            </div>
            </template>
        </div>

        <!-- 技能发动流程：选择目标 -->
        <div v-else-if="interruptStore.skillMode === 'choosing_target' && interruptStore.selectedSkill" class="space-y-3 skill-target-panel">
            <div class="flex items-center justify-between">
                <span class="text-amber-400 text-sm font-bold">{{ interruptStore.selectedSkill.title }}</span>
                <span class="step-indicator">
          {{ interruptStore.skillTargetIds.length }}/{{ (interruptStore.selectedSkill.max_targets > 0 ? interruptStore.selectedSkill.max_targets : 1) }}
        </span>
            </div>
            <p v-if="interruptStore.selectedSkill.description && !isBoardGuidedSkillFlow" class="text-xs text-gray-400 whitespace-pre-wrap break-words">{{ interruptStore.selectedSkill.description }}</p>
            <div class="text-xs text-gray-400">
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
            <div v-if="interruptStore.skillTargetIds.length > 0" class="text-[11px] text-amber-300">
                已选目标：
                {{ interruptStore.skillTargetIds.map((id) => snapshotStore.players[id]?.name || id).join('、') }}
            </div>
            <div class="flex gap-3 justify-center mt-2">
                <button
                    v-if="isManualTargetConfirmSkillFlow"
                    class="action-image-btn w-16 sm:w-20"
                    :class="{ 'opacity-50 cursor-not-allowed': !canConfirmSkill }"
                    title="确认发动"
                    :disabled="!canConfirmSkill"
                    @click="confirmSkill()"
                >
                    <img v-if="isMainActionImageReady('confirm')" class="action-image-btn-fill" :src="mainActionButtonImage('confirm')" alt="" @error="onMainActionImageError('confirm')" />
                    <span v-else class="action-image-fallback-text">确</span>
                    <span class="action-image-btn-label">确认发动</span>
                </button>
                <button class="action-image-btn w-16 sm:w-20" title="取消" @click="interruptStore.clearSkillMode()">
                    <img v-if="isMainActionImageReady('cancel')" class="action-image-btn-fill" :src="mainActionButtonImage('cancel')" alt="" @error="onMainActionImageError('cancel')" />
                    <span v-else class="action-image-fallback-text">消</span>
                    <span class="action-image-btn-label">取消</span>
                </button>
            </div>
        </div>

        <!-- 等待提示 -->
        <div v-else-if="interruptStore.waitingFor && !prompt" class="text-center py-2 sm:py-3 text-gray-400 text-sm">
            <div class="animate-pulse">等待 {{ waitingName || interruptStore.waitingFor }} 操作...</div>
        </div>

        <!-- 非行动类 Prompt：在行动区内直接显示按钮操作 -->
        <div v-else-if="prompt && isPromptForMe && !isActionHubContext" class="prompt-inline-host">
            <PromptDialog />
        </div>

        <!-- 行动区域 -->
        <div v-else-if="isActionHubContext" class="action-hub-desktop">
            <div class="action-hub-desktop-main">
                <button
                    v-if="hasActionPromptOption('attack')"
                    class="action-hub-desktop-btn action-image-btn action-image-btn--attack"
                    :title="actionPromptLabel('attack', '攻击')"
                    :aria-label="actionPromptLabel('attack', '攻击')"
                    @click="invokeActionHubOption('attack')"
                >
                    <img
                        v-if="isMainActionImageReady('attack')"
                        class="action-image-btn-fill"
                        :src="mainActionButtonImage('attack')"
                        alt=""
                        @error="onMainActionImageError('attack')"
                    />
                    <span v-else class="action-image-fallback-text">攻</span>
                    <span class="action-image-btn-label">{{ actionPromptLabel('attack', '攻击') }}</span>
                </button>
                <button
                    v-if="hasActionPromptOption('magic')"
                    class="action-hub-desktop-btn action-image-btn action-image-btn--magic"
                    :title="actionPromptLabel('magic', '法术')"
                    :aria-label="actionPromptLabel('magic', '法术')"
                    @click="invokeActionHubOption('magic')"
                >
                    <img
                        v-if="isMainActionImageReady('magic')"
                        class="action-image-btn-fill"
                        :src="mainActionButtonImage('magic')"
                        alt=""
                        @error="onMainActionImageError('magic')"
                    />
                    <span v-else class="action-image-fallback-text">术</span>
                    <span class="action-image-btn-label">{{ actionPromptLabel('magic', '法术') }}</span>
                </button>
                <button
                    v-if="showSpecialHubEntry"
                    class="action-hub-desktop-btn action-image-btn action-image-btn--special"
                    :class="{ 'action-image-btn--muted': !hasHubSpecialActions || isStartupSpecialLocked }"
                    :title="isStartupSpecialLocked ? '本回合已执行启动技能，特殊行动已禁用' : actionPromptLabel('special', '特殊')"
                    :aria-label="actionPromptLabel('special', '特殊')"
                    :disabled="!hasHubSpecialActions || isStartupSpecialLocked"
                    @click="openSpecialActionModal"
                >
                    <img
                        v-if="isMainActionImageReady('special')"
                        class="action-image-btn-fill"
                        :src="mainActionButtonImage('special')"
                        alt=""
                        @error="onMainActionImageError('special')"
                    />
                    <span v-else class="action-image-fallback-text">特</span>
                    <span class="action-image-btn-label">{{ actionPromptLabel('special', '特殊') }}</span>
                </button>
                <template v-if="isActionSelectionPrompt">
                    <button
                        v-if="hasActionPromptOption('cannot_act')"
                        class="action-hub-desktop-btn action-image-btn action-image-btn--cannot-act"
                        :title="isExtraActionPrompt ? '跳过' : cannotActButtonLabel"
                        :aria-label="isExtraActionPrompt ? '跳过' : cannotActButtonLabel"
                        @click="invokeActionHubOption('cannot_act')"
                    >
                        <img
                            v-if="isMainActionImageReady('cannot_act')"
                            class="action-image-btn-fill"
                            :src="mainActionButtonImage('cannot_act')"
                            alt=""
                            @error="onMainActionImageError('cannot_act')"
                        />
                        <span v-else class="action-image-fallback-text">{{ isExtraActionPrompt ? '跳过' : '无法' }}</span>
                        <span class="action-image-btn-label">{{ isExtraActionPrompt ? '跳过' : cannotActButtonLabel }}</span>
                    </button>
                </template>
                <template v-else>
                    <button
                        v-if="effectiveAvailableSkills.length > 0"
                        class="action-hub-desktop-btn action-image-btn action-image-btn--skill"
                        title="发动技能"
                        aria-label="发动技能"
                        @click="invokeActionHubOption('skill')"
                    >
                        <img
                            v-if="isMainActionImageReady('skill')"
                            class="action-image-btn-fill"
                            :src="mainActionButtonImage('skill')"
                            alt=""
                            @error="onMainActionImageError('skill')"
                        />
                        <span v-else class="action-image-fallback-text">技</span>
                        <span class="action-image-btn-label">发动技能</span>
                    </button>
                    <button
                        class="action-hub-desktop-btn action-image-btn action-image-btn--pass"
                        title="结束回合"
                        aria-label="结束回合"
                        @click="invokeActionHubOption('pass')"
                    >
                        <img
                            v-if="isMainActionImageReady('pass')"
                            class="action-image-btn-fill"
                            :src="mainActionButtonImage('pass')"
                            alt=""
                            @error="onMainActionImageError('pass')"
                        />
                        <span v-else class="action-image-fallback-text">过</span>
                        <span class="action-image-btn-label">结束回合</span>
                    </button>
                </template>
            </div>

            <div
                v-if="isActionSelectionPrompt && !hasActionPromptOption('attack') && !hasActionPromptOption('magic') && !hasHubSpecialActions"
                class="action-hub-desktop-empty"
            >
                当前无可执行行动，请等待下一步结算
            </div>
        </div>

        <!-- 非我的回合 -->
        <div v-else class="text-center py-4 text-gray-400">
            等待其他玩家回合...
        </div>

        <button
            v-if="debugAvailable"
            class="debug-toggle-btn"
            type="button"
            @click="openDebugPanel"
        >
            调试
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
.action-panel-root {
    position: relative;
    width: 100%;
    min-height: 0;
    border: none !important;
    box-shadow: none !important;
}

.action-panel-root--panel {
    padding: 10px;
    border-radius: 12px;
    background:
        linear-gradient(180deg, rgba(10, 24, 39, 0.9), rgba(6, 17, 28, 0.94)),
        url('/assets/ui/panel-ornament.svg') center/cover no-repeat;
    border: none !important;
    box-shadow:
        inset 0 1px 0 rgba(244, 250, 255, 0.08),
        0 12px 26px rgba(2, 8, 23, 0.42);
    min-height: 176px;
    max-height: min(52vh, 460px);
    overflow-y: auto;
    overflow-x: hidden;
    scrollbar-width: thin;
}

.action-panel-root--panel::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 0.75rem;
    pointer-events: none;
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0) 46%);
}

.action-panel-root--hub {
    min-height: 0;
    overflow: visible;
    background: transparent !important;
    border: none !important;
    box-shadow: none !important;
    backdrop-filter: none !important;
    z-index: 1202;
}

.action-panel-root--prompt-inline {
    padding: 6px;
    min-height: 0;
    max-height: none;
    overflow-y: visible;
}

.action-hub-desktop .btn-danger {
    background: linear-gradient(180deg, #cb5f5f, #ad3d3d) !important;
    border: 1px solid #f1a1a1 !important;
    color: #fff !important;
}

.action-hub-desktop .btn-primary {
    background: linear-gradient(180deg, #5ba4de, #356da5) !important;
    border: 1px solid #9dd3ff !important;
    color: #fff !important;
}

.action-hub-desktop .btn-skill,
.action-hub-desktop .btn-economy {
    background: linear-gradient(180deg, #c8a86d, #91713f) !important;
    border: 1px solid #efd7a3 !important;
    color: #fff !important;
}

.action-hub-desktop .btn-secondary {
    background: linear-gradient(180deg, #5a6577, #3b4554) !important;
    border: 1px solid #b8c3d8 !important;
    color: #eef4ff !important;
}

.skill-btn-disabled {
    opacity: 0.45;
    filter: grayscale(0.25);
    cursor: not-allowed;
}

.action-hub-desktop {
    width: 100%;
    max-width: 100%;
    border-radius: 14px;
    padding: 8px;
    background:
        linear-gradient(180deg, rgba(12, 28, 43, 0.96), rgba(7, 19, 31, 0.98)),
        url('/assets/ui/panel-ornament.svg') center/cover no-repeat;
    border: none !important;
    box-shadow:
        inset 0 1px 0 rgba(242, 250, 255, 0.08),
        0 10px 24px rgba(3, 12, 22, 0.42);
}

.action-hub-desktop-main {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 6px;
}

.action-hub-desktop-btn {
    min-height: 33px;
    font-size: 12px;
    line-height: 1.1;
}

.action-hub-desktop-btn--cannot-act {
    min-height: 33px;
    padding: 0.52rem 0.65rem !important;
    display: flex;
    justify-content: center;
    align-items: center;
    border-radius: 10px;
    border: 1px solid #b8c3d8 !important;
    background: linear-gradient(180deg, #5a6577, #3b4554) !important;
    box-shadow: 0 3px 8px rgba(7, 15, 25, 0.28) !important;
    color: #eef4ff !important;
    text-shadow: 0 1px 1px rgba(0, 0, 0, 0.2);
    transition:
        transform 0.16s ease,
        filter 0.16s ease,
        box-shadow 0.18s ease,
        border-color 0.18s ease;
}

.action-hub-desktop-btn--cannot-act:hover {
    transform: translateY(-1px);
    filter: brightness(1.03);
    border-color: #c6d2e7 !important;
    box-shadow: 0 5px 10px rgba(7, 15, 25, 0.34) !important;
}

.action-hub-desktop-btn--cannot-act:active {
    transform: translateY(1px);
    filter: brightness(0.96);
    box-shadow: 0 2px 6px rgba(7, 15, 25, 0.24) !important;
}

.action-hub-desktop-btn--cannot-act:focus-visible {
    outline: none;
    border-color: #cde4ff !important;
    box-shadow:
        0 0 0 2px rgba(85, 170, 255, 0.38),
        0 5px 10px rgba(7, 15, 25, 0.32) !important;
}

.action-hub-desktop-btn--cannot-act-extra {
    border-color: #c6dfff !important;
    background: linear-gradient(180deg, #60799a, #3f5c7d) !important;
    color: #f2f8ff !important;
}

.debug-toggle-btn {
    position: absolute;
    top: 8px;
    right: 8px;
    z-index: 5;
    padding: 4px 10px;
    font-size: 12px;
    border-radius: 999px;
    border: 1px solid rgba(148, 163, 184, 0.5);
    background: rgba(15, 23, 42, 0.8);
    color: #e2e8f0;
}

.debug-toggle-btn:hover {
    background: rgba(30, 41, 59, 0.9);
}

.debug-modal-mask {
    position: fixed;
    inset: 0;
    background: rgba(4, 10, 20, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2400;
}

.debug-modal-card {
    width: min(980px, 94vw);
    max-height: 88vh;
    background: linear-gradient(180deg, rgba(14, 24, 39, 0.97), rgba(8, 16, 28, 0.96));
    border-radius: 16px;
    border: 1px solid rgba(148, 163, 184, 0.3);
    box-shadow: 0 20px 40px rgba(2, 6, 23, 0.45);
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
    border-bottom: 1px solid rgba(148, 163, 184, 0.2);
}

.debug-modal-title {
    font-size: 16px;
    font-weight: 700;
    color: #f8fafc;
}

.debug-modal-subtitle {
    font-size: 12px;
    color: rgba(226, 232, 240, 0.7);
}

.debug-modal-close {
    padding: 6px 12px;
    border-radius: 10px;
    border: 1px solid rgba(148, 163, 184, 0.4);
    background: rgba(30, 41, 59, 0.8);
    color: #e2e8f0;
    font-size: 12px;
}

.debug-modal-controls {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    padding: 12px 18px;
    border-bottom: 1px solid rgba(148, 163, 184, 0.18);
}

.debug-input,
.debug-select {
    background: rgba(15, 23, 42, 0.85);
    border: 1px solid rgba(148, 163, 184, 0.35);
    color: #e2e8f0;
    padding: 6px 10px;
    border-radius: 10px;
    font-size: 12px;
    min-width: 140px;
}

.debug-status {
    margin-left: auto;
    font-size: 12px;
    color: rgba(148, 163, 184, 0.85);
}

.debug-modal-body {
    padding: 12px 18px 18px;
    overflow: auto;
}

.debug-skill-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 12px;
}

.debug-skill-item {
    background: rgba(15, 23, 42, 0.6);
    border: 1px solid rgba(148, 163, 184, 0.2);
    border-radius: 12px;
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.debug-skill-head {
    display: flex;
    justify-content: space-between;
    gap: 10px;
}

.debug-skill-title {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
}

.debug-skill-role {
    font-size: 11px;
    color: rgba(148, 163, 184, 0.85);
}

.debug-skill-name {
    font-size: 13px;
    font-weight: 600;
    color: #f8fafc;
}

.debug-skill-type {
    font-size: 11px;
    color: rgba(251, 191, 36, 0.9);
}

.debug-skill-cost {
    font-size: 11px;
    color: rgba(226, 232, 240, 0.8);
    white-space: nowrap;
}

.debug-skill-desc {
    font-size: 11px;
    color: rgba(226, 232, 240, 0.65);
    line-height: 1.4;
}

.debug-skill-btn {
    align-self: flex-start;
    padding: 6px 10px;
    border-radius: 10px;
    background: linear-gradient(180deg, #d0b36f, #8d6a2e);
    color: #fff;
    font-size: 12px;
    border: 1px solid rgba(251, 191, 36, 0.6);
}

.debug-manual {
    margin-top: 16px;
    padding-top: 12px;
    border-top: 1px solid rgba(148, 163, 184, 0.2);
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.debug-manual-title {
    font-size: 12px;
    font-weight: 600;
    color: rgba(226, 232, 240, 0.9);
}

.debug-manual-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
}

.debug-mini-btn {
    padding: 6px 10px;
    border-radius: 10px;
    border: 1px solid rgba(148, 163, 184, 0.35);
    background: rgba(30, 41, 59, 0.85);
    color: #e2e8f0;
    font-size: 12px;
}

.action-image-btn {
    -webkit-appearance: none !important;
    appearance: none !important;
    border: none !important;
    background: transparent !important;
    border-radius: 12px !important;
    aspect-ratio: 1 / 1;
    padding: 0 !important;
    box-shadow: none !important;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    position: relative;
    overflow: hidden;
    transition: transform 0.14s ease, filter 0.14s ease;
    flex-shrink: 0 !important;
}

.action-hub-desktop-btn.action-image-btn {
    width: 100%;
    max-width: 100%;
    min-height: 0;
}

.action-image-btn:hover {
    transform: translateY(-1px);
    filter: brightness(1.06);
}

.action-image-btn:active {
    transform: translateY(0);
    filter: brightness(0.98);
}

.action-image-btn:disabled {
    cursor: not-allowed;
    transform: none !important;
    filter: grayscale(0.62) saturate(0.66) brightness(0.76);
}

.action-image-btn:focus,
.action-image-btn:focus-visible {
    outline: none !important;
    box-shadow: none !important;
}

.action-image-btn--muted {
    filter: grayscale(0.55) saturate(0.7) brightness(0.82);
}

.action-image-btn-fill {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: contain;
    pointer-events: none;
    user-select: none;
}

.action-image-fallback-text {
    position: relative;
    z-index: 1;
    font-size: 14px;
    font-weight: 700;
    color: #f2f8ff;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.45);
}

.action-image-btn-label {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
}

.action-hub-desktop-empty {
    margin-top: 6px;
    text-align: center;
    font-size: 12px;
    color: rgba(223, 236, 248, 0.88);
}

.prompt-inline-host {
    width: 100%;
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

.action-mode-panel,
.skill-select-panel,
.skill-discard-panel,
.skill-target-panel {
    background: rgba(7, 20, 33, 0.58);
    border: 1px solid rgba(121, 156, 177, 0.24);
    border-radius: 12px;
    padding: 10px;
    box-shadow: inset 0 1px 0 rgba(236, 247, 254, 0.06);
}

.skill-discard-card-row {
    /* 与手牌区一致，给上移选中态留缓冲，避免顶部裁切。 */
    padding-top: 12px;
    margin-top: -6px;
}

.action-mode-panel {
    min-height: clamp(176px, 24vh, 248px);
}

.target-group-stack {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.target-group-card {
    border: 1px solid rgba(117, 154, 176, 0.3);
    border-radius: 10px;
    background: rgba(9, 22, 37, 0.66);
    box-shadow: inset 0 1px 0 rgba(230, 244, 255, 0.05);
    padding: 8px;
}

.target-group-title {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.05em;
    margin-bottom: 6px;
}

.target-group-title--enemy {
    color: rgba(251, 113, 133, 0.92);
}

.target-group-title--ally {
    color: rgba(125, 211, 252, 0.92);
}

.target-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
}

.target-grid-btn {
    min-width: min(160px, 100%);
    flex: 1 1 168px;
}

.target-grid-name {
    font-weight: 700;
    line-height: 1.2;
}

.target-grid-meta {
    margin-top: 2px;
    font-size: 11px;
    opacity: 0.82;
    line-height: 1.25;
}

.prompt-option-btn {
    min-height: 34px;
    line-height: 1.15;
}

.special-modal-mask {
    position: fixed;
    inset: 0;
    z-index: 2150;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    background: rgba(5, 10, 18, 0.64);
    backdrop-filter: blur(4px);
}

.special-modal-card {
    width: min(430px, calc(100vw - 28px));
    border-radius: 14px;
    border: 1px solid rgba(164, 196, 216, 0.5);
    background:
        linear-gradient(180deg, rgba(16, 33, 52, 0.96), rgba(8, 18, 31, 0.98)),
        url('/assets/ui/panel-ornament.svg') center/cover no-repeat;
    box-shadow:
        inset 0 1px 0 rgba(239, 248, 255, 0.12),
        0 18px 34px rgba(2, 10, 20, 0.52);
    padding: 14px 12px 12px;
}

.special-modal-title {
    font-size: 15px;
    font-weight: 700;
    color: rgba(237, 246, 253, 0.95);
    text-align: center;
}

.special-modal-subtitle {
    margin-top: 6px;
    margin-bottom: 8px;
    text-align: center;
    font-size: 12px;
    color: rgba(188, 213, 230, 0.88);
    line-height: 1.4;
}

.special-modal-group-title {
    font-size: 11px;
    letter-spacing: 0.08em;
    color: #f6d9a1;
    font-weight: 700;
    margin-bottom: 8px;
}

.special-modal-actions {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.special-action-card {
    border-radius: 10px;
    border: 1px solid rgba(117, 156, 182, 0.34);
    background: rgba(9, 22, 37, 0.72);
    padding: 8px;
    box-shadow: inset 0 1px 0 rgba(233, 245, 255, 0.06);
}

.special-action-card--disabled {
    border-color: rgba(104, 124, 140, 0.3);
    background: rgba(8, 16, 26, 0.6);
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
    background: linear-gradient(180deg, rgba(219, 186, 125, 0.28), rgba(150, 106, 49, 0.34));
    color: #ffe3b5;
    font-size: 14px;
    box-shadow: inset 0 1px 0 rgba(255, 246, 223, 0.22);
}

.special-action-meta {
    min-width: 0;
}

.special-action-label {
    font-size: 13px;
    font-weight: 700;
    color: #ecf5fc;
    line-height: 1.1;
}

.special-action-summary {
    margin-top: 2px;
    font-size: 11px;
    color: rgba(183, 210, 229, 0.86);
    line-height: 1.35;
}

.special-action-detail {
    margin-top: 6px;
    font-size: 11px;
    color: rgba(161, 192, 213, 0.84);
    line-height: 1.35;
}

.special-action-reason {
    margin-top: 5px;
    font-size: 11px;
    color: #f2bc9e;
    line-height: 1.35;
}

.special-modal-btn {
    min-height: 34px;
    width: 100%;
    font-size: 12px;
    margin-top: 6px;
}

.special-modal-btn--disabled {
    opacity: 0.55;
    cursor: not-allowed;
}

.special-modal-cancel {
    margin-top: 10px;
    width: 100%;
    min-height: 34px;
    font-size: 12px;
}

.special-modal-fade-enter-active,
.special-modal-fade-leave-active {
    transition: opacity 0.18s ease;
}

.special-modal-fade-enter-from,
.special-modal-fade-leave-to {
    opacity: 0;
}

@media (max-width: 900px) {
    .action-panel-root--panel {
        padding: 8px;
        min-height: 160px;
        max-height: min(44vh, 360px);
    }

    .action-panel-root--prompt-inline {
        padding: 6px;
        min-height: 0;
        max-height: none;
    }

    .action-mode-panel,
    .skill-select-panel,
    .skill-discard-panel,
    .skill-target-panel {
        border-radius: 10px;
    }

    .prompt-option-btn {
        min-height: 32px;
        font-size: 12px !important;
        padding: 0.35rem 0.62rem !important;
    }
}

@media (max-width: 640px) {
    .action-panel-root--panel {
        padding: 6px;
        min-height: 0;
        max-height: min(36vh, 300px);
    }

    .action-panel-root--prompt-inline {
        padding: 4px;
        min-height: 0;
        max-height: none;
    }

    .action-mode-panel .btn-primary,
    .action-mode-panel .btn-secondary,
    .action-mode-panel .btn-skill {
        flex: 1 1 calc(50% - 6px);
        min-width: 0;
    }

    .skill-select-panel button:not(.action-image-btn),
    .skill-target-panel button:not(.action-image-btn),
    .skill-discard-panel button:not(.action-image-btn) {
        min-height: 34px;
    }

    .action-panel-root--panel button:not(.action-image-btn) {
        padding-top: 0.4rem;
        padding-bottom: 0.4rem;
    }

    .skill-select-panel .action-image-btn,
    .skill-target-panel .action-image-btn,
    .skill-discard-panel .action-image-btn,
    .action-mode-panel .action-image-btn {
        flex: 0 0 auto !important;
    }

    .action-panel-root--panel .text-lg {
        font-size: 1rem;
    }

    .action-hub-desktop-btn {
        font-size: 11px;
        min-height: 29px;
        padding: 0.32rem 0.45rem !important;
    }

    .action-hub-desktop-btn--cannot-act {
        min-height: 29px;
        padding: 0.32rem 0.45rem !important;
    }

    .prompt-option-btn {
        min-height: 30px;
        font-size: 11px !important;
        padding: 0.28rem 0.52rem !important;
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

    .special-modal-btn,
    .special-modal-cancel {
        min-height: 31px;
        font-size: 11px;
    }
}
</style>
