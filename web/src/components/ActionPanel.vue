<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { useWebSocket } from '../composables/useWebSocket'
import type { AvailableSkill, PromptOption, PlayerView } from '../types/game'
import CardComponent from './CardComponent.vue'

const store = useGameStore()
const ws = useWebSocket()

const prompt = computed(() => store.currentPrompt)

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
const debugStatus = ref('')

const debugAvailable = computed(() => {
    if (typeof window === 'undefined') return false
    const query = new URLSearchParams(window.location.search)
    return import.meta.env.DEV || query.has('debug')
})

type MainActionIconId = 'attack' | 'magic' | 'special'
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
    attack: ['/assets/ui/action_attack_btn.png', '/assets/ui/action_attack.png', '/assets/ui/action_attack.svg'],
    magic: ['/assets/ui/action_magic_btn.png', '/assets/ui/action_magic.png', '/assets/ui/action_magic.svg'],
    special: ['/assets/ui/action_special_btn.png', '/assets/ui/action_special.png', '/assets/ui/action_special.svg'],
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
const isMyTurn = computed(() => store.isMyTurn)
const waitingName = computed(() => {
    if (!store.waitingFor) return ''
    return store.players[store.waitingFor]?.name || store.waitingFor
})
const specialActionModalVisible = ref(false)
const isIdleMainTurnPanel = computed(() =>
    isMyTurn.value &&
    !prompt.value &&
    store.actionMode === 'none' &&
    store.skillMode === 'none'
)

function isActionSelectionPromptMessage(message: string): boolean {
    return message.includes('行动类型')
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
    if (!p || !store.isPromptForMe) return false
    if (p.ui_mode === 'action_hub') return true
    if (p.type !== 'confirm') return false
    if (!isActionSelectionPromptMessage(p.message || '')) return false
    return (p.options || []).some((opt) => normalizeActionHubOptionId(opt) !== null)
})

const isActionHubContext = computed(() =>
    (isIdleMainTurnPanel.value || isActionSelectionPrompt.value) &&
    store.actionMode === 'none' &&
    store.skillMode === 'none'
)

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
    if (store.hasPerformedStartup) {
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
const isStartupSpecialLocked = computed(() => store.hasPerformedStartup)
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
    store.myCamp === 'Red'
        ? store.redGems + store.redCrystals
        : store.blueGems + store.blueCrystals
)
const personalEnergy = computed(() => {
    const me = store.myPlayer
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
    Object.values(store.players).sort((a, b) => a.name.localeCompare(b.name, 'zh-Hans-CN'))
)

const debugRoleList = computed(() =>
    Object.values(store.characters).sort((a, b) => a.name.localeCompare(b.name, 'zh-Hans-CN'))
)

const debugExclusiveSkillOptions = computed(() => {
    const role = store.characters[debugExclusiveRoleId.value]
    if (!role || !Array.isArray(role.skills)) return []
    return role.skills
})

const mainActionImageIndex = ref<Record<MainActionIconId, number>>({
    attack: 0,
    magic: 0,
    special: 0,
})
const mainActionImageFailed = ref<Record<MainActionIconId, boolean>>({
    attack: false,
    magic: false,
    special: false,
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

function triggerActionHubOption(optionId: string) {
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
            ws.sendAction({ player_id: store.myPlayerId, type: 'CannotAct' })
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
        store.setError('你本回合已执行启动技能，不能执行特殊行动')
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
        store.setError(chosen.disabledReason || '该特殊行动当前不可执行')
        return
    }
    specialActionModalVisible.value = false
    triggerActionHubOption(optionId)
}

function openActionHubAttack() {
    store.setActionModeForAttack('attack')
}

function openActionHubMagic() {
    store.setActionModeForAttack('magic')
}

function openSkillAction() {
    if (store.effectiveAvailableSkills.length === 0) {
        store.setError('当前没有可发动技能')
        return
    }
    store.setSkillMode('choosing_skill')
}

function openBuyAction() {
    ws.buy()
}

function openSynthesizeAction() {
    ws.sendAction({ player_id: store.myPlayerId, type: 'Synthesize' })
}

function openExtractAction() {
    ws.extract()
}

function openPassAction() {
    ws.pass()
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
            store.setError('你本回合已执行启动技能，不能执行特殊行动')
            return
        }
        openSpecialActionModal()
        return
    } else if (optionId === 'buy') {
        ws.buy()
    } else if (optionId === 'extract') {
        ws.extract()
    } else if (optionId === 'synthesize') {
        ws.sendAction({ player_id: store.myPlayerId, type: 'Synthesize' })
    } else if (optionId === 'attack') {
        store.setActionModeForAttack('attack')
    } else if (optionId === 'magic') {
        store.setActionModeForAttack('magic')
    } else if (optionId === 'cannot_act') {
        ws.sendAction({ player_id: store.myPlayerId, type: 'CannotAct' })
    } else if (optionId === 'skip' || optionId === 'cancel') {
        ws.cancel()
        return
    } else if (optionId === 'confirm') {
        ws.confirm()
    } else if (optionId === 'take') {
        ws.respond('take')
    } else if (optionId === 'counter') {
        if (store.selectedCards.length === 0) {
            store.setError(isMagicMissilePromptMessage() ? '请先选择一张【魔弹】再传递' : '请先选择一张应战牌')
            return
        }
        ws.respond('counter', store.selectedCards[0])
    } else if (optionId === 'defend') {
        if (store.selectedCards.length === 0) {
            store.setError('请先选择一张【圣光】进行防御（圣盾需提前放置）')
            return
        }
        ws.respond('defend', store.selectedCards[0])
    } else if (optionId === 'yes' || optionId === 'no') {
        // 魔弹融合等确认选项：yes=0, no=1
        ws.select([optionId === 'yes' ? 0 : 1])
    } else if (optionId === 'normal' || optionId === 'reverse') {
        // 魔弹掌控方向选择：normal=0, reverse=1
        ws.select([optionId === 'normal' ? 0 : 1])
    } else if (prompt.value.type === 'choose_skill') {
        const idx = prompt.value.options.findIndex((o: { id: string }) => o.id === optionId)
        if (idx >= 0) {
            ws.select([idx])
        } else {
            ws.cancel()
            return
        }
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
    // 不在此处清除 prompt：等待后端 state_update（成功）或新 prompt 到达后再清除
    // 若后端报错，prompt 保持显示，用户可重新选择
}

function backFromMagicCard() {
    store.setMagicSubChoice('none')
    store.setSelectedCardForAction(null)
}

function isMagicBulletCard(cardIdx: number): boolean {
    const card = store.myPlayableCards.find(item => item.index === cardIdx)?.card
    return !!card && card.type === 'Magic' && card.name === '魔弹'
}

function confirmTarget(playerId: string) {
    const cardIdx = store.selectedCardForAction
    if (cardIdx === null) return
    const selectedItem = store.myPlayableCards.find((item) => item.index === cardIdx)
    if (!selectedItem) {
        store.setSelectedCardForAction(null)
        store.setError('所选卡牌已变化，请重新选择')
        return
    }
    if (store.actionMode === 'attack') {
        if (selectedItem.card.type !== 'Attack') {
            store.setSelectedCardForAction(null)
            store.setError('所选卡牌不是攻击牌，请重新选择')
            return
        }
        ws.attack(playerId, cardIdx)
    } else if (store.actionMode === 'magic') {
        if (selectedItem.card.type !== 'Magic') {
            store.setSelectedCardForAction(null)
            store.setError('所选卡牌不是法术牌，请重新选择')
            return
        }
        if (isMagicBulletCard(cardIdx)) {
            ws.magic(undefined, cardIdx)
        } else {
            ws.magic(playerId, cardIdx)
        }
    }
}

const attackTargetCandidates = computed(() => {
    if (store.actionMode !== 'attack') return []
    return Object.values(store.players).filter((p) => p.camp !== store.myCamp)
})

function isAllyTarget(player: PlayerView): boolean {
    if (store.myCamp) return player.camp === store.myCamp
    const myCamp = store.players[store.myPlayerId]?.camp
    if (myCamp) return player.camp === myCamp
    return player.id === store.myPlayerId
}

function splitTargetsByCamp(targets: PlayerView[]): { enemies: PlayerView[]; allies: PlayerView[] } {
    const enemies: PlayerView[] = []
    const allies: PlayerView[] = []
    for (const target of targets) {
        if (isAllyTarget(target)) allies.push(target)
        else enemies.push(target)
    }
    return { enemies, allies }
}

const actionTargets = computed<PlayerView[]>(() => {
    if (store.actionMode === 'attack') return attackTargetCandidates.value
    if (store.actionMode === 'magic') return store.targetablePlayers
    return []
})

const hasActionTargets = computed(() => actionTargets.value.length > 0)
const groupedActionTargets = computed(() => splitTargetsByCamp(actionTargets.value))
const groupedSkillTargets = computed(() => splitTargetsByCamp(store.targetablePlayersForSkill))

function targetBasicInfo(targetId: string): string {
    const player = store.players[targetId]
    if (!player) return ''
    const handCount = Number.isFinite(player.hand_count) ? player.hand_count : (player.hand?.length ?? 0)
    const heal = Number.isFinite(player.heal) ? player.heal : 0
    return `手牌 ${handCount} · 治疗 ${heal}`
}

function canTargetPlayer(playerId: string): boolean {
    return store.targetablePlayers.some((p) => p.id === playerId)
}

function isStealthBlockedTarget(playerId: string): boolean {
    if (store.actionMode !== 'attack') return false
    if (canTargetPlayer(playerId)) return false
    const p = store.players[playerId]
    if (!p || !Array.isArray(p.field)) return false
    return p.field.some((fc) => fc.mode === 'Effect' && fc.effect === 'Stealth')
}

const hasStealthBlockedAttackTarget = computed(() =>
    attackTargetCandidates.value.some((p) => isStealthBlockedTarget(p.id))
)

function confirmSkill() {
    const skill = store.selectedSkill
    if (!skill) return
    // 发送时带上弃牌索引
    ws.useSkill(skill.id, store.skillTargetIds, store.skillDiscardIndices.length > 0 ? store.skillDiscardIndices : undefined)
}

function selectSkill(skill: AvailableSkill) {
    if (!canSelectSkill(skill)) {
        store.setError(skillDisabledReason(skill))
        return
    }
    store.setSelectedSkill(skill)
    // 如果技能需要弃牌，先进入弃牌选择模式
    if (skill.cost_discards > 0) {
        const required = requiredDiscardCount(skill)
        if (required <= 0) {
            proceedAfterDiscard(skill)
            return
        }
        store.setSkillMode('choosing_discard')
        return
    }
    // 无需弃牌，直接进入目标选择或发动
    proceedAfterDiscard(skill)
}

function cardMatchesSkillDiscard(card: { type: string; element: string; faction?: string; exclusive_char1?: string; exclusive_char2?: string; exclusive_skill1?: string; exclusive_skill2?: string }, skill: AvailableSkill): boolean {
    if (skill.require_exclusive) {
        const char = store.getCharacter(store.myCharRole)
        if (!char || !store.cardMatchesExclusive(card, char.name, skill.title)) return false
    }
    if (skill.discard_type && card.type !== skill.discard_type) return false
    if (skill.discard_element) return card.element === skill.discard_element
    if (skill.id === 'magic_bullet_fusion') return card.element === 'Fire' || card.element === 'Earth'
    return true
}

function countSkillDiscardCandidates(skill: AvailableSkill): number {
    if (!skill || skill.cost_discards <= 0) return 0
    return store.myHand.filter(card => cardMatchesSkillDiscard(card, skill)).length
}

function hasOnmyojiSameFactionPair(): boolean {
    const countByFaction = new Map<string, number>()
    for (const card of store.myHand) {
        if (!card.faction) continue
        const next = (countByFaction.get(card.faction) || 0) + 1
        if (next >= 2) return true
        countByFaction.set(card.faction, next)
    }
    return false
}

function hasAnyBasicEffectTarget(): boolean {
    const isBasicEffect = (effect?: string) => {
        return effect === 'Shield' || effect === 'Weak' || effect === 'Poison' ||
            effect === 'SealFire' || effect === 'SealWater' || effect === 'SealEarth' ||
            effect === 'SealWind' || effect === 'SealThunder' ||
            effect === 'PowerBlessing' || effect === 'SwiftBlessing'
    }
    return Object.values(store.players).some((p) => {
        if (!p || !Array.isArray(p.field)) return false
        return p.field.some((fc) =>
            fc.mode === 'Effect' && isBasicEffect(fc.effect)
        )
    })
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
    ms_shadow_meteor: [{ token: 'ms_shadow_form', min: 1, label: '暗影形态' }],
    bw_heavenfire_cleave: [{ token: 'bw_rebirth', min: 1, label: '重生' }],
}

function getMyTokenValue(token: string): number {
    return store.myPlayer?.tokens?.[token] ?? 0
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
        const form = getMyTokenValue('hb_form')
        const faith = getMyTokenValue('hb_faith')
        const heal = store.myPlayer?.heal ?? 0
        if (form > 0) return '已处于圣煌形态，无法再次发动。'
        if (heal < 2 && faith < 2) return '治疗与信仰均不足2，无法发动。'
    }
    if (skill.id === 'hb_light_burst') {
        const form = getMyTokenValue('hb_form')
        if (form <= 0) return '仅圣煌形态下可发动。'
    }
    if (skill.id === 'hb_radiant_cannon') {
        const form = getMyTokenValue('hb_form')
        const cannon = getMyTokenValue('hb_radiant_cannon')
        const faith = getMyTokenValue('hb_faith')
        const myMorale = store.myCamp === 'Red' ? store.redMorale : store.blueMorale
        const enemyMorale = store.myCamp === 'Red' ? store.blueMorale : store.redMorale
        const moraleGap = Math.max(0, enemyMorale - myMorale)
        const requiredFaith = 4 + moraleGap
        if (form <= 0) return '仅圣煌形态下可发动。'
        if (cannon <= 0) return '圣煌辉光炮指示物不足。'
        if (faith < requiredFaith) return `信仰不足（需要 ${requiredFaith}，当前 ${faith}）。`
    }
    return ''
}

function canPaySkillEnergy(skill: AvailableSkill): boolean {
    const me = store.myPlayer
    if (!me) return false
    const gemNeed = skill.cost_gem || 0
    const crystalNeed = skill.cost_crystal || 0
    if (me.gem < gemNeed) return false
    const usableCrystal = me.crystal + (me.gem - gemNeed)
    return usableCrystal >= crystalNeed
}

function canSelectSkill(skill: AvailableSkill): boolean {
    if (!skill) return false
    if (!canPaySkillEnergy(skill)) return false
    if (skillTokenDisabledReason(skill)) return false
    if (skill.id === 'prayer_radiant_faith' || skill.id === 'prayer_dark_curse') {
        const prayerForm = store.myPlayer?.tokens?.prayer_form ?? 0
        const prayerRune = store.myPlayer?.tokens?.prayer_rune ?? 0
        if (prayerForm <= 0 || prayerRune <= 0) return false
    }
    if (skill.id === 'elementalist_ignite') {
        const element = store.myPlayer?.tokens?.element ?? 0
        if (element < 3) return false
    }
    if (skill.id === 'onmyoji_shikigami_descend') {
        return hasOnmyojiSameFactionPair()
    }
    if (skill.id === 'angel_cleanse' && !hasAnyBasicEffectTarget()) {
        return false
    }
    if (skill.cost_discards > 0) {
        const required = requiredDiscardCount(skill)
        if (required > 0 && countSkillDiscardCandidates(skill) < required) {
            return false
        }
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
        const prayerForm = store.myPlayer?.tokens?.prayer_form ?? 0
        const prayerRune = store.myPlayer?.tokens?.prayer_rune ?? 0
        if (prayerForm <= 0) return '仅祈祷形态下可发动。'
        if (prayerRune <= 0) return '祈祷符文不足，无法发动。'
    }
    if (skill.id === 'elementalist_ignite') {
        return '元素不足3点，无法发动【元素点燃】。'
    }
    if (skill.id === 'angel_blessing') {
        return '手牌中没有水系牌，无法发动【天使祝福】。'
    }
    if (skill.id === 'angel_cleanse') {
        if (!hasAnyBasicEffectTarget()) {
            return '场上没有可移除的基础效果，无法发动【风之洁净】。'
        }
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
        ws.useSkill(skill.id, [], store.skillDiscardIndices.length > 0 ? store.skillDiscardIndices : undefined)
        return
    }
    // target_type=1 (Self): 自动选中自己并发动
    if (skill.target_type === 1) {
        ws.useSkill(skill.id, [store.myPlayerId], store.skillDiscardIndices.length > 0 ? store.skillDiscardIndices : undefined)
        return
    }
    store.setSkillMode('choosing_target')
}

function confirmSkillDiscard() {
    const skill = store.selectedSkill
    if (!skill) return
    const required = requiredDiscardCount(skill)
    if (store.skillDiscardIndices.length < required) {
        store.setError(`请选择 ${required} 张牌`)
        return
    }
    proceedAfterDiscard(skill)
}

function onSkillTargetClick(playerId: string) {
    store.toggleSkillTarget(playerId)
    const skill = store.selectedSkill
    if (!skill) return
    // 单目标技能选中后自动确认
    if (skill.max_targets === 1 && store.skillTargetIds.length === 1) {
        ws.useSkill(skill.id, store.skillTargetIds, store.skillDiscardIndices.length > 0 ? store.skillDiscardIndices : undefined)
    }
}

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
        debugTargetPlayerId.value = store.myPlayerId
    }
    if (!debugExclusiveRoleId.value) {
        debugExclusiveRoleId.value = store.myCharRole || debugRoleList.value[0]?.id || ''
    }
    if (!debugExclusiveSkillId.value) {
        debugExclusiveSkillId.value = debugExclusiveSkillOptions.value[0]?.id || ''
    }
}

function closeDebugPanel() {
    debugOpen.value = false
}

function ensureDebugTargetPlayerId(): string | null {
    const pid = debugTargetPlayerId.value || store.myPlayerId
    if (!pid || !store.players[pid]) {
        store.setError('请选择有效的目标角色')
        return null
    }
    return pid
}

function debugTargetName(pid: string): string {
    return store.players[pid]?.name || pid
}

function applyDebugEffect() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const count = Number(debugEffectCount.value)
    if (!Number.isFinite(count) || count < 0) {
        store.setError('基础效果数量需为 >= 0 的数字')
        return
    }
    ws.cheatEffect(pid, debugEffectType.value, Math.floor(count))
    debugStatus.value = `已设置 ${debugTargetName(pid)} 的基础效果 ${debugEffectType.value}=${Math.floor(count)}`
}

function applyDebugSet() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const value = Number(debugSetValue.value)
    if (!Number.isFinite(value)) {
        store.setError('请输入有效数字')
        return
    }
    ws.cheatSet(pid, debugSetField.value, Math.floor(value))
    debugStatus.value = `已设置 ${debugTargetName(pid)} 的 ${debugSetField.value}=${Math.floor(value)}`
}

function applyDebugToken() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const key = debugTokenKey.value.trim()
    const value = Number(debugTokenValue.value)
    if (!key) {
        store.setError('请输入指示物 key')
        return
    }
    if (!Number.isFinite(value)) {
        store.setError('请输入有效数字')
        return
    }
    ws.cheatToken(pid, key, Math.floor(value))
    debugStatus.value = `已设置 ${debugTargetName(pid)} 的指示物 ${key}=${Math.floor(value)}`
}

function applyDebugExclusiveCard() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    if (!debugExclusiveRoleId.value) {
        store.setError('请选择角色来源')
        return
    }
    if (!debugExclusiveSkillId.value) {
        store.setError('请选择独有技')
        return
    }
    const count = Number(debugExclusiveCount.value)
    if (!Number.isFinite(count) || count <= 0) {
        store.setError('独有牌数量需为 > 0 的数字')
        return
    }
    ws.cheatGiveExclusive(pid, debugExclusiveRoleId.value, debugExclusiveSkillId.value, Math.floor(count))
    debugStatus.value = `已给 ${debugTargetName(pid)} 添加 ${Math.floor(count)} 张独有技手牌`
}

function applyDebugElementCards() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const count = Number(debugElementCount.value)
    if (!Number.isFinite(count) || count <= 0) {
        store.setError('系别手牌数量需为 > 0 的数字')
        return
    }
    ws.cheatGiveByElement(pid, debugElement.value, Math.floor(count))
    debugStatus.value = `已给 ${debugTargetName(pid)} 添加 ${Math.floor(count)} 张 ${elementName(debugElement.value)}手牌`
}

function applyDebugFactionCards() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const faction = debugFaction.value.trim()
    if (!faction) {
        store.setError('请输入命格')
        return
    }
    const count = Number(debugFactionCount.value)
    if (!Number.isFinite(count) || count <= 0) {
        store.setError('命格手牌数量需为 > 0 的数字')
        return
    }
    ws.cheatGiveByFaction(pid, faction, Math.floor(count))
    debugStatus.value = `已给 ${debugTargetName(pid)} 添加 ${Math.floor(count)} 张 ${faction}命格手牌`
}

function applyDebugMagicCard() {
    const pid = ensureDebugTargetPlayerId()
    if (!pid) return
    const cardName = debugMagicCardName.value.trim()
    if (!cardName) {
        store.setError('请输入法术牌名称')
        return
    }
    const count = Number(debugMagicCardCount.value)
    if (!Number.isFinite(count) || count <= 0) {
        store.setError('法术牌数量需为 > 0 的数字')
        return
    }
    ws.cheatGiveMagicByName(pid, cardName, Math.floor(count))
    debugStatus.value = `已给 ${debugTargetName(pid)} 添加 ${Math.floor(count)} 张法术牌【${cardName}】`
}

function requiredDiscardCount(skill: AvailableSkill): number {
    if (!skill || skill.cost_discards <= 0) return 0
    // 神官-神圣领域：手牌不足2时，改为弃全部手牌。
    if (skill.id === 'priest_divine_domain') {
        return Math.min(skill.cost_discards, store.myHand.length)
    }
    // 神官-水之神力：若弃完水系后无剩余手牌，则仅需弃1张水系牌。
    if (skill.id === 'priest_water_power') {
        return Math.min(skill.cost_discards, store.myHand.length)
    }
    return skill.cost_discards
}

function isCardSelectableForSkillDiscard(card: { type: string; element: string; faction?: string; exclusive_char1?: string; exclusive_char2?: string; exclusive_skill1?: string; exclusive_skill2?: string }): boolean {
    const skill = store.selectedSkill
    if (!skill) return false
    if (skill.id === 'priest_water_power') {
        const selected = store.skillDiscardIndices
            .map((i) => store.myHand[i])
            .filter((c): c is NonNullable<typeof c> => !!c)
        if (selected.length === 0) {
            return card.element === 'Water'
        }
        // 第一张已是水系后，第二张可为任意手牌（但上限仍由 requiredDiscardCount 控制）。
        return selected[0]?.element === 'Water'
    }
    // 独有技：必须使用卡牌下标了该技能名的牌
    if (skill.require_exclusive) {
        const char = store.getCharacter(store.myCharRole)
        if (!char) return false
        if (!store.cardMatchesExclusive(card, char.name, skill.title)) return false
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
        const selected = store.skillDiscardIndices
            .map((i) => store.myHand[i])
            .filter((c): c is NonNullable<typeof c> => !!c)
        if (selected.length > 0) {
            const reqFaction = selected[0]?.faction
            if (reqFaction && card.faction !== reqFaction) return false
        }
    }
    return true
}

function toggleSkillDiscardCard(idx: number) {
    const skill = store.selectedSkill
    if (!skill) return
    const required = requiredDiscardCount(skill)
    const card = store.myHand[idx]
    if (!card) return
    // 独有技：必须使用卡牌下标了该技能名的牌
    if (skill.require_exclusive) {
        const char = store.getCharacter(store.myCharRole)
        if (!char || !store.cardMatchesExclusive(card, char.name, skill.title)) {
            store.setError('必须使用标有该技能名的独有牌')
            return
        }
    }
    // 检查元素要求
    if (skill.discard_element && card.element !== skill.discard_element) {
        store.setError(`需要弃置${elementName(skill.discard_element)}牌`)
        return
    }
    if (skill.discard_type && card.type !== skill.discard_type) {
        store.setError(`需要弃置${skill.discard_type === 'Magic' ? '法术' : '攻击'}牌`)
        return
    }
    if (skill.id === 'priest_water_power' && !store.skillDiscardIndices.includes(idx)) {
        const selected = store.skillDiscardIndices
            .map((i) => store.myHand[i])
            .filter((c): c is NonNullable<typeof c> => !!c)
        if (selected.length === 0 && card.element !== 'Water') {
            store.setError('水之神力第一张需弃置水系牌')
            return
        }
        if (selected.length > 0 && selected[0]?.element !== 'Water') {
            store.setError('水之神力第一张需弃置水系牌')
            return
        }
    }
    if (skill.id === 'magic_bullet_fusion' && card.element !== 'Fire' && card.element !== 'Earth') {
        store.setError('魔弹融合需要弃置1张火系或地系牌')
        return
    }
    // 阴阳师：式神降临必须弃置两张同命格手牌
    if (skill.id === 'onmyoji_shikigami_descend' && !store.skillDiscardIndices.includes(idx)) {
        if (!card.faction) {
            store.setError('式神降临需要弃置有命格的手牌')
            return
        }
        const selected = store.skillDiscardIndices
            .map((i) => store.myHand[i])
            .filter((c): c is NonNullable<typeof c> => !!c)
        if (selected.length > 0) {
            const reqFaction = selected[0]?.faction
            if (reqFaction && card.faction !== reqFaction) {
                store.setError('式神降临需要弃置2张命格相同的手牌')
                return
            }
        }
    }
    // 如果已选满且不是取消选择，不允许继续选
    if (!store.skillDiscardIndices.includes(idx) && store.skillDiscardIndices.length >= required) {
        store.setError(`最多选择 ${required} 张牌`)
        return
    }
    store.toggleSkillDiscard(idx)
}

function elementName(el: string): string {
    const map: Record<string, string> = { Water: '水系', Fire: '火系', Wind: '风系', Earth: '土系', Dark: '暗灭' }
    return map[el] || el
}

function basicEffectSummary(playerId: string): string {
    const player = store.players[playerId]
    if (!player) return ''
    const labels: string[] = []
    for (const fc of player.field) {
        if (fc.mode !== 'Effect') continue
        if (fc.effect === 'Shield') labels.push('圣盾')
        else if (fc.effect === 'Weak') labels.push('虚弱')
        else if (fc.effect === 'Poison') labels.push('中毒')
    }
    return labels.join('、')
}
</script>

<template>
    <div
        class="action-panel-root"
        :class="[isActionHubContext ? 'action-panel-root--hub' : 'action-panel-root--panel']"
    >
        <!-- 攻击/法术模式 -->
        <div v-if="store.actionMode !== 'none'" class="space-y-3 action-mode-panel">
            <!-- 法术行动：先选择 出牌 或 发动技能 -->
            <template v-if="store.actionMode === 'magic' && store.magicSubChoice === 'none'">
                <div class="text-amber-400 text-sm font-bold">✨ 法术行动</div>
                <div class="text-xs text-gray-400">发动法术有两种方式：打出法术牌或发动角色技能</div>
                <div class="flex gap-2 flex-wrap">
                    <button class="btn-primary px-4 py-2.5" @click="store.setMagicSubChoice('card')">
                        打出法术牌
                    </button>
                    <button
                        v-if="store.effectiveAvailableSkills.length > 0"
                        class="btn-skill px-4 py-2.5"
                        @click="store.setSkillMode('choosing_skill'); store.clearActionMode()"
                    >
                        发动技能
                    </button>
                    <button class="btn-secondary flex-1 py-2.5" @click="store.clearActionMode()">取消</button>
                </div>
            </template>
            <!-- 攻击模式 或 法术已选「出牌」：选牌 + 选目标 -->
            <template v-else>
                <div class="flex items-center justify-between">
          <span class="text-amber-400 text-sm font-bold">
            {{ store.actionMode === 'attack' ? '⚔️ 攻击模式' : '✨ 法术模式' }}
          </span>
                    <span class="step-indicator">
            {{ store.selectedCardForAction === null ? '步骤 1/2' : '步骤 2/2' }} · {{ store.selectedCardForAction === null ? '选牌' : '选目标' }}
          </span>
                </div>
                <div v-if="store.selectedCardForAction !== null" class="space-y-2">
                    <div
                        v-if="hasActionTargets"
                        class="text-xs text-gray-400"
                    >
                        点击目标玩家完成 {{ store.actionMode === 'attack' ? '攻击' : '施法' }}：
                    </div>
                    <div v-else class="text-xs text-gray-400">当前法术无需手动选目标，将按规则自动结算。</div>
                    <div
                        class="target-group-stack"
                        v-if="hasActionTargets"
                    >
                        <div v-if="groupedActionTargets.enemies.length > 0" class="target-group-card">
                            <div class="target-group-title target-group-title--enemy">敌方阵营</div>
                            <div class="target-grid">
                                <button
                                    v-for="target in groupedActionTargets.enemies"
                                    :key="target.id"
                                    class="btn-target target-grid-btn px-3 py-2.5 rounded-lg text-sm font-medium text-left"
                                    :class="[
                                        target.camp === 'Red'
                                          ? 'bg-red-900/60 hover:bg-red-800/80 border-2 border-red-500 text-red-200'
                                          : 'bg-blue-900/60 hover:bg-blue-800/80 border-2 border-blue-500 text-blue-200',
                                        canTargetPlayer(target.id) ? '' : 'opacity-50 grayscale cursor-not-allowed'
                                    ]"
                                    :disabled="!canTargetPlayer(target.id)"
                                    @click="canTargetPlayer(target.id) ? confirmTarget(target.id) : null"
                                >
                                    <div class="target-grid-name">{{ target.id === store.myPlayerId ? '自己' : target.name }}</div>
                                    <div class="target-grid-meta">{{ targetBasicInfo(target.id) }}</div>
                                </button>
                            </div>
                        </div>
                        <div v-if="groupedActionTargets.allies.length > 0" class="target-group-card">
                            <div class="target-group-title target-group-title--ally">我方阵营</div>
                            <div class="target-grid">
                                <button
                                    v-for="target in groupedActionTargets.allies"
                                    :key="target.id"
                                    class="btn-target target-grid-btn px-3 py-2.5 rounded-lg text-sm font-medium text-left"
                                    :class="[
                                        target.camp === 'Red'
                                          ? 'bg-red-900/60 hover:bg-red-800/80 border-2 border-red-500 text-red-200'
                                          : 'bg-blue-900/60 hover:bg-blue-800/80 border-2 border-blue-500 text-blue-200',
                                        canTargetPlayer(target.id) ? '' : 'opacity-50 grayscale cursor-not-allowed'
                                    ]"
                                    :disabled="!canTargetPlayer(target.id)"
                                    @click="canTargetPlayer(target.id) ? confirmTarget(target.id) : null"
                                >
                                    <div class="target-grid-name">{{ target.id === store.myPlayerId ? '自己' : target.name }}</div>
                                    <div class="target-grid-meta">{{ targetBasicInfo(target.id) }}</div>
                                </button>
                            </div>
                        </div>
                    </div>
                    <div v-if="hasStealthBlockedAttackTarget" class="text-[11px] text-gray-400">
                        潜行状态无法选中
                    </div>
                </div>
                <div v-else class="text-xs text-gray-400 py-1">
                    先在下方手牌选一张{{ store.actionMode === 'attack' ? '攻击' : '法术' }}牌
                </div>
                <div class="flex gap-2 flex-wrap">
                    <button class="btn-secondary text-sm flex-1 py-2" @click="store.actionMode === 'magic' ? backFromMagicCard() : store.clearActionMode()">
                        {{ store.actionMode === 'magic' ? '返回' : '取消' }}
                    </button>
                    <button
                        v-if="store.actionMode === 'magic' && store.effectiveAvailableSkills.length > 0"
                        class="btn-skill text-sm px-4 py-2"
                        @click="store.setSkillMode('choosing_skill'); store.clearActionMode()"
                    >
                        改用技能
                    </button>
                </div>
            </template>
        </div>

        <!-- 技能发动流程：选择技能 -->
        <div v-else-if="store.skillMode === 'choosing_skill'" class="space-y-3 skill-select-panel">
            <div class="text-amber-400 text-sm font-bold">选择要发动的技能</div>
            <div class="flex flex-col gap-2">
                <button
                    v-for="skill in store.effectiveAvailableSkills"
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
            <button class="btn-secondary text-sm w-full py-2" @click="store.clearSkillMode()">
                取消
            </button>
        </div>

        <!-- 技能发动流程：选择弃牌 -->
        <div v-else-if="store.skillMode === 'choosing_discard' && store.selectedSkill" class="space-y-3 skill-discard-panel">
            <div class="flex items-center justify-between">
                <span class="text-amber-400 text-sm font-bold">{{ store.selectedSkill.title }}</span>
                <span class="step-indicator">
          {{ store.skillDiscardIndices.length }}/{{ requiredDiscardCount(store.selectedSkill) }}
        </span>
            </div>
            <div class="text-xs text-gray-400">
                请选择要弃置的牌
                <span v-if="store.selectedSkill.require_exclusive" class="text-amber-300">
          （须为标有「{{ store.selectedSkill.title }}」的独有牌）
        </span>
                <span v-else-if="store.selectedSkill.discard_element" class="text-amber-300">
          （需要{{ elementName(store.selectedSkill.discard_element) }}牌）
                </span>
                <span v-else-if="store.selectedSkill.discard_type" class="text-amber-300">
          （需要{{ store.selectedSkill.discard_type === 'Magic' ? '法术牌' : '攻击牌' }}）
                </span>
                <span v-else-if="store.selectedSkill.id === 'priest_water_power'" class="text-amber-300">
          （第一张需水系；若仍有手牌，第二张将交给目标队友）
                </span>
                <span v-else-if="store.selectedSkill.id === 'magic_bullet_fusion'" class="text-amber-300">
          （需要火系或地系牌）
                </span>
                <span v-else-if="store.selectedSkill.id === 'onmyoji_shikigami_descend'" class="text-amber-300">
          （需要2张命格相同的手牌）
                </span>
            </div>
            <div class="flex gap-1 flex-wrap justify-center">
                <CardComponent
                    v-for="(card, idx) in store.myHand"
                    :key="idx"
                    :card="card"
                    :index="idx"
                    medium
                    :selectable="isCardSelectableForSkillDiscard(card)"
                    :selected="store.skillDiscardIndices.includes(idx)"
                    @click="toggleSkillDiscardCard(idx)"
                />
            </div>
            <div class="flex gap-2">
                <button
                    class="btn-success flex-1 py-2"
                    :class="{ 'opacity-50 cursor-not-allowed': store.skillDiscardIndices.length < requiredDiscardCount(store.selectedSkill) }"
                    :disabled="store.skillDiscardIndices.length < requiredDiscardCount(store.selectedSkill)"
                    @click="confirmSkillDiscard()"
                >
                    确认弃牌 ({{ store.skillDiscardIndices.length }}/{{ requiredDiscardCount(store.selectedSkill) }})
                </button>
                <button class="btn-secondary py-2 px-4" @click="store.clearSkillMode()">取消</button>
            </div>
        </div>

        <!-- 技能发动流程：选择目标 -->
        <div v-else-if="store.skillMode === 'choosing_target' && store.selectedSkill" class="space-y-3 skill-target-panel">
            <div class="flex items-center justify-between">
                <span class="text-amber-400 text-sm font-bold">{{ store.selectedSkill.title }}</span>
                <span class="step-indicator">
          {{ store.skillTargetIds.length }}/{{ (store.selectedSkill.max_targets > 0 ? store.selectedSkill.max_targets : 1) }}
        </span>
            </div>
            <p v-if="store.selectedSkill.description" class="text-xs text-gray-400 whitespace-pre-wrap break-words">{{ store.selectedSkill.description }}</p>
            <div class="text-xs text-gray-400">
                点击玩家头像或下方按钮选择目标
                <span v-if="store.selectedSkill.min_targets > 0">（至少 {{ store.selectedSkill.min_targets }} 个）</span>
                <span v-if="(store.selectedSkill.max_targets || 1) === 1"> · 选中后自动发动</span>
            </div>
            <div class="target-group-stack">
                <div v-if="groupedSkillTargets.enemies.length > 0" class="target-group-card">
                    <div class="target-group-title target-group-title--enemy">敌方阵营</div>
                    <div class="target-grid">
                        <button
                            v-for="target in groupedSkillTargets.enemies"
                            :key="target.id"
                            class="btn-target target-grid-btn px-3 py-2 rounded-lg text-sm font-medium text-left"
                            :class="[
                                store.skillTargetIds.includes(target.id)
                                  ? 'ring-2 ring-yellow-400 bg-amber-900/70'
                                  : 'bg-gray-700 hover:bg-gray-600',
                                target.camp === 'Red' ? 'border border-red-500/50' : 'border border-blue-500/50'
                            ]"
                            @click="onSkillTargetClick(target.id)"
                        >
                            <div class="target-grid-name">{{ target.id === store.myPlayerId ? '自己' : target.name }}</div>
                            <div class="target-grid-meta">{{ targetBasicInfo(target.id) }}</div>
                            <div v-if="store.selectedSkill?.id === 'angel_cleanse'" class="text-[11px] opacity-80 mt-0.5">
                                可移除：{{ basicEffectSummary(target.id) || '无' }}
                            </div>
                        </button>
                    </div>
                </div>
                <div v-if="groupedSkillTargets.allies.length > 0" class="target-group-card">
                    <div class="target-group-title target-group-title--ally">我方阵营</div>
                    <div class="target-grid">
                        <button
                            v-for="target in groupedSkillTargets.allies"
                            :key="target.id"
                            class="btn-target target-grid-btn px-3 py-2 rounded-lg text-sm font-medium text-left"
                            :class="[
                                store.skillTargetIds.includes(target.id)
                                  ? 'ring-2 ring-yellow-400 bg-amber-900/70'
                                  : 'bg-gray-700 hover:bg-gray-600',
                                target.camp === 'Red' ? 'border border-red-500/50' : 'border border-blue-500/50'
                            ]"
                            @click="onSkillTargetClick(target.id)"
                        >
                            <div class="target-grid-name">{{ target.id === store.myPlayerId ? '自己' : target.name }}</div>
                            <div class="target-grid-meta">{{ targetBasicInfo(target.id) }}</div>
                            <div v-if="store.selectedSkill?.id === 'angel_cleanse'" class="text-[11px] opacity-80 mt-0.5">
                                可移除：{{ basicEffectSummary(target.id) || '无' }}
                            </div>
                        </button>
                    </div>
                </div>
            </div>
            <div class="flex gap-2">
                <button
                    class="btn-success flex-1 py-2"
                    :class="{ 'opacity-50 cursor-not-allowed': !store.canConfirmSkill }"
                    :disabled="!store.canConfirmSkill"
                    @click="confirmSkill()"
                >
                    确认发动
                </button>
                <button class="btn-secondary py-2 px-4" @click="store.clearSkillMode()">取消</button>
            </div>
        </div>

        <!-- 等待提示 -->
        <div v-else-if="store.waitingFor && !prompt" class="text-center py-2 sm:py-3 text-gray-400 text-sm">
            <div class="animate-pulse">等待 {{ waitingName || store.waitingFor }} 操作...</div>
        </div>

        <!-- 非行动类 Prompt 统一交给中央 PromptDialog，避免右侧行动区重复出现选项 -->
        <div v-else-if="prompt && store.isPromptForMe && !isActionHubContext" class="text-center py-2 sm:py-3 text-slate-300 text-sm">
            <div>当前为中断/响应阶段，请在中央弹框中操作</div>
        </div>

        <!-- 行动区域 -->
        <div v-else-if="isActionHubContext" class="action-hub-desktop">
            <div class="action-hub-desktop-main">
                <button
                    v-if="hasActionPromptOption('attack')"
                    class="action-hub-desktop-btn action-image-btn action-image-btn--attack"
                    :title="actionPromptLabel('attack', '攻击')"
                    :aria-label="actionPromptLabel('attack', '攻击')"
                    @click="triggerActionHubOption('attack')"
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
                    @click="triggerActionHubOption('magic')"
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
                        class="btn-secondary action-hub-desktop-btn"
                        @click="triggerActionHubOption('cannot_act')"
                    >
                        {{ cannotActButtonLabel }}
                    </button>
                </template>
                <template v-else>
                    <button
                        v-if="store.effectiveAvailableSkills.length > 0"
                        class="btn-skill action-hub-desktop-btn"
                        @click="triggerActionHubOption('skill')"
                    >
                        发动技能
                    </button>
                    <button class="btn-secondary action-hub-desktop-btn" @click="triggerActionHubOption('pass')">
                        结束回合
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
                            <div class="debug-modal-subtitle">可给任意角色设置基础效果、资源/治疗/指示物，并按条件补牌。</div>
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
    width: 100%;
    max-width: 100%;
    aspect-ratio: 1 / 1;
    min-height: 0;
    padding: 0 !important;
    box-shadow: none !important;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    position: relative;
    overflow: hidden;
    transition: transform 0.14s ease, filter 0.14s ease;
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

    .action-mode-panel .btn-primary,
    .action-mode-panel .btn-secondary,
    .action-mode-panel .btn-skill {
        flex: 1 1 calc(50% - 6px);
        min-width: 0;
    }

    .skill-select-panel button,
    .skill-target-panel button,
    .skill-discard-panel button {
        min-height: 34px;
    }

    .action-panel-root--panel button {
        padding-top: 0.4rem !important;
        padding-bottom: 0.4rem !important;
    }

    .action-panel-root--panel .text-lg {
        font-size: 1rem;
    }

    .action-hub-desktop-btn {
        font-size: 11px;
        min-height: 29px;
        padding: 0.32rem 0.45rem !important;
    }

    .action-image-btn {
        min-height: 40px;
        padding: 0 !important;
    }

    .prompt-option-btn {
        min-height: 30px;
        font-size: 11px !important;
        padding: 0.28rem 0.52rem !important;
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
