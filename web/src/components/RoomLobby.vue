<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useSubmitAction } from '../composables/useSubmitAction'
import { useWebSocket } from '../composables/useWebSocket'
import { ROLE_NAME_MAP } from '../constants/roleNameMap'
import { autoJoinRoomFromUrl } from '../utils/autoJoinRoom'
import SkillDetailModal from './SkillDetailModal.vue'

const interruptStore = useInterruptStore()
const sessionStore = useSessionStore()
const snapshotStore = useSnapshotStore()
const actions = useSubmitAction()
const ws = useWebSocket()

const playerName = ref('')
const roomCodeInput = ref('')
const isJoining = ref(false)
const errorMsg = ref('')
const copyFeedback = ref(false)
const skillModalRoleId = ref('')
const portraitFallbackStep = ref<Record<string, number>>({})
const previewRoleId = ref('')
const factionFilter = ref('')
const searchKeyword = ref('')
const page = ref(0)
const inviteFeedback = ref(false)

const PAGE_SIZE = 12

const IMAGE_EXTS = ['png', 'webp', 'jpg', 'jpeg']

/**
 * 复制到剪贴板。navigator.clipboard 只在安全上下文（https / localhost）存在，
 * 局域网 IP 打开时退回到隐藏 textarea + execCommand('copy')。
 */
async function writeClipboard(text: string) {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text)
      return
    } catch {
      // 权限被拒或窗口未聚焦：走下面的兜底
    }
  }
  const area = document.createElement('textarea')
  area.value = text
  area.setAttribute('readonly', '')
  area.style.position = 'fixed'
  area.style.opacity = '0'
  document.body.appendChild(area)
  area.select()
  const ok = document.execCommand('copy')
  document.body.removeChild(area)
  if (!ok) throw new Error('copy failed')
}

async function copyRoomCode() {
  try {
    await writeClipboard(sessionStore.roomCode)
    copyFeedback.value = true
    setTimeout(() => {
      copyFeedback.value = false
    }, 1500)
  } catch {
    errorMsg.value = '复制失败'
  }
}

const roomPlayers = computed(() => sessionStore.roomPlayers)
const isHost = computed(() => roomPlayers.value.some(p => p.id === sessionStore.myPlayerId && p.is_host))
const botCount = computed(() => roomPlayers.value.filter(p => p.is_bot).length)

const redCount = computed(() => roomPlayers.value.filter(p => p.camp === 'Red').length)
const blueCount = computed(() => roomPlayers.value.filter(p => p.camp === 'Blue').length)
const allCampsSelected = computed(() => {
  if (roomPlayers.value.length < 2) return false
  return roomPlayers.value.every(p => p.camp && p.camp !== '')
})
const allRolesSelected = computed(() => {
  if (roomPlayers.value.length < 2) return false
  return roomPlayers.value.every(p => p.char_role && p.char_role !== '')
})
const allReadyToStart = computed(() => allCampsSelected.value && allRolesSelected.value)

const characterOptions = computed(() => Object.values(snapshotStore.characters))
const characterMap = computed(() => {
  const map = new Map<string, (typeof characterOptions.value)[number]>()
  for (const role of characterOptions.value) {
    map.set(role.id, role)
  }
  return map
})

const takenRoleMap = computed(() => {
  const map = new Map<string, (typeof roomPlayers.value)[number]>()
  for (const p of roomPlayers.value) {
    if (p.char_role) {
      map.set(p.char_role, p)
    }
  }
  return map
})

const bluePlayers = computed(() =>
  roomPlayers.value
    .filter(p => p.camp === 'Blue')
    .slice()
    .sort((a, b) => a.id.localeCompare(b.id))
)

const redPlayers = computed(() =>
  roomPlayers.value
    .filter(p => p.camp === 'Red')
    .slice()
    .sort((a, b) => a.id.localeCompare(b.id))
)

const unassignedPlayers = computed(() =>
  roomPlayers.value
    .filter(p => !p.camp)
    .slice()
    .sort((a, b) => a.id.localeCompare(b.id))
)

const blueSlots = computed(() => Array.from({ length: 3 }, (_, i) => bluePlayers.value[i] || null))
const redSlots = computed(() => Array.from({ length: 3 }, (_, i) => redPlayers.value[i] || null))

const teams = computed(() => [
  { key: 'blue', camp: 'Blue' as const, label: '蓝方', slots: blueSlots.value, otherKey: 'red', otherCamp: 'Red' as const },
  { key: 'red', camp: 'Red' as const, label: '红方', slots: redSlots.value, otherKey: 'blue', otherCamp: 'Blue' as const },
])

const skillModalCharacter = computed(() => {
  if (!skillModalRoleId.value) return null
  return characterMap.value.get(skillModalRoleId.value) ?? null
})
const displayError = computed(() => errorMsg.value || interruptStore.errorMessage)

const lobbyHint = computed(() => {
  if (roomPlayers.value.length < 2) return `至少需要 2 名玩家（当前 ${roomPlayers.value.length}/6）`
  if (!allCampsSelected.value) return '等待所有玩家选择阵营'
  if (!allRolesSelected.value) return '等待所有玩家锁定角色'
  return '所有玩家已完成阵营与角色选择，系统将自动开始对局...'
})

function createRoom() {
  if (!playerName.value.trim()) {
    errorMsg.value = '请输入玩家名称'
    return
  }
  isJoining.value = true
  errorMsg.value = ''
  ws.connect('', playerName.value.trim(), true)
}

function joinRoom() {
  if (!playerName.value.trim()) {
    errorMsg.value = '请输入玩家名称'
    return
  }
  if (!roomCodeInput.value.trim()) {
    errorMsg.value = '请输入房间码'
    return
  }
  isJoining.value = true
  errorMsg.value = ''
  ws.connect(roomCodeInput.value.trim().toUpperCase(), playerName.value.trim())
}

onMounted(() => {
  if (typeof window === 'undefined') return

  const didAutoJoin = autoJoinRoomFromUrl({
    search: window.location.search,
    isInRoom: () => sessionStore.isInRoom,
    connect: (roomCode, name) => {
      playerName.value = name
      roomCodeInput.value = roomCode
      isJoining.value = true
      errorMsg.value = ''
      ws.connect(roomCode, name)
    },
  })

  if (!didAutoJoin) {
    // 邀请链接只带房间码时，预填到加入输入框
    const room = new URLSearchParams(window.location.search).get('room')
    if (room) roomCodeInput.value = room.trim().toUpperCase()
  }
})

function getCharacterName(roleId: string) {
  if (!roleId) return '未选择角色'
  return snapshotStore.characters[roleId]?.name || ROLE_NAME_MAP[roleId] || '未知角色'
}

function portraitSrc(roleId: string) {
  const step = portraitFallbackStep.value[roleId] ?? 0
  const ext = IMAGE_EXTS[Math.min(step, IMAGE_EXTS.length - 1)]
  return `/characters/${roleId}.${ext}`
}

function onPortraitError(roleId: string, event: Event) {
  const current = portraitFallbackStep.value[roleId] ?? 0
  const next = current + 1
  if (next < IMAGE_EXTS.length) {
    portraitFallbackStep.value = { ...portraitFallbackStep.value, [roleId]: next }
    const img = event.target as HTMLImageElement | null
    if (img) {
      img.src = `/characters/${roleId}.${IMAGE_EXTS[next]}`
    }
    return
  }
  const img = event.target as HTMLImageElement | null
  if (img) {
    img.src = '/assets/card-back.webp'
  }
}

function roleTakenBy(roleId: string) {
  return takenRoleMap.value.get(roleId)
}

function isRoleTakenByOther(roleId: string) {
  const owner = roleTakenBy(roleId)
  return !!owner && owner.id !== sessionStore.myPlayerId
}

function canSelectRole(roleId: string) {
  if (sessionStore.gameStarted) return false
  return !isRoleTakenByOther(roleId)
}

function openSkillModal(roleId: string) {
  skillModalRoleId.value = roleId
}

function closeSkillModal() {
  skillModalRoleId.value = ''
}

function selectCamp(camp: string) {
  actions.changeCamp(camp)
}

function selectRole(role: string) {
  if (!role) return
  actions.changeRole(role)
}

function selectCampFor(playerId: string, camp: string) {
  actions.changeCamp(camp, playerId)
}

function selectRoleFor(playerId: string, role: string) {
  if (!role) return
  actions.changeRole(role, playerId)
}

function pickRole(roleId: string) {
  if (!canSelectRole(roleId)) return
  selectRole(roleId)
}

function canJoinCamp(camp: 'Red' | 'Blue') {
  if (sessionStore.myCamp === camp) return true
  if (camp === 'Red') return redCount.value < 3
  return blueCount.value < 3
}

function addBot() {
  actions.addBot(`机器人${botCount.value + 1}`)
}

function removeBot(playerId: string) {
  actions.removeBot(playerId)
}

function startGame() {
  actions.startRoom()
}

function dissolveRoom() {
  if (!isHost.value) return
  const confirmed = window.confirm('确认解散房间吗？所有玩家将被退出到大厅。')
  if (!confirmed) return
  actions.dissolveRoom()
}

// 五大命格固定顺序
const FACTION_ORDER = ['圣', '血', '幻', '技', '咏']

const factionOptions = computed(() => {
  const present = new Set(characterOptions.value.map(role => role.faction))
  return FACTION_ORDER.filter(faction => present.has(faction))
})

const filteredRoles = computed(() => {
  const keyword = searchKeyword.value.trim()
  return characterOptions.value.filter(role => {
    if (factionFilter.value && role.faction !== factionFilter.value) return false
    if (keyword && !role.name.includes(keyword) && !(role.title || '').includes(keyword)) return false
    return true
  })
})

const pageCount = computed(() => Math.max(1, Math.ceil(filteredRoles.value.length / PAGE_SIZE)))
const pagedRoles = computed(() => {
  const current = Math.min(page.value, pageCount.value - 1)
  return filteredRoles.value.slice(current * PAGE_SIZE, (current + 1) * PAGE_SIZE)
})

const previewCharacter = computed(() => {
  const id = previewRoleId.value || sessionStore.myCharRole || ''
  return (id && characterMap.value.get(id)) || null
})

const previewSkillLine = computed(() => {
  const skill = previewCharacter.value?.skills?.[0]
  if (!skill) return ''
  return `${skill.title}：${skill.description}`
})

const previewLockedByMe = computed(() => !!previewCharacter.value && previewCharacter.value.id === sessionStore.myCharRole)
const canConfirmPreview = computed(() => {
  const role = previewCharacter.value
  if (!role || previewLockedByMe.value) return false
  return canSelectRole(role.id)
})

function setFaction(faction: string) {
  factionFilter.value = faction
  page.value = 0
}

function onSearchInput() {
  page.value = 0
}

function changePage(delta: number) {
  page.value = Math.min(pageCount.value - 1, Math.max(0, page.value + delta))
}

function previewRole(roleId: string) {
  previewRoleId.value = roleId
}

function confirmPreview() {
  const role = previewCharacter.value
  if (!role || !canConfirmPreview.value) return
  pickRole(role.id)
}

async function copyInviteLink() {
  const url = `${window.location.origin}${window.location.pathname}?room=${encodeURIComponent(sessionStore.roomCode)}`
  try {
    await writeClipboard(url)
    inviteFeedback.value = true
    setTimeout(() => {
      inviteFeedback.value = false
    }, 1500)
  } catch {
    errorMsg.value = '复制邀请链接失败'
  }
}
</script>

<template>
  <div class="lobby-shell" :class="sessionStore.isInRoom ? 'lobby-shell-draft' : 'lobby-shell-login'">
    <!-- 登录 / 入口 -->
    <div v-if="!sessionStore.isInRoom" class="login-stage">
      <div class="login-brand">
        <img class="brand-emblem" src="/assets/ui/star-emblem.svg" alt="" aria-hidden="true" />
        <h1 class="brand-title">星杯传说</h1>
        <div class="brand-sub"><span class="brand-line"></span>STAR CUP LEGEND<span class="brand-line"></span></div>
      </div>

      <div class="login-card ornate-panel">
        <span class="panel-crest" aria-hidden="true">✦</span>
        <label class="field-label" for="lobby-player-name">玩家名称</label>
        <div class="field">
          <svg class="field-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 12a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9Zm0 2c-4.4 0-8 2.4-8 5.4V21h16v-1.6c0-3-3.6-5.4-8-5.4Z" /></svg>
          <input
            id="lobby-player-name"
            v-model="playerName"
            type="text"
            placeholder="输入你的名字"
            maxlength="20"
            @keyup.enter="createRoom"
          />
        </div>

        <button
          class="gold-btn login-create"
          data-testid="create-room-button"
          :disabled="isJoining"
          @click="createRoom"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3 2 11.5h3V21h5.5v-6h3v6H19v-9.5h3L12 3Z" /></svg>
          创建房间
        </button>

        <div class="login-divider"><span>或者</span></div>

        <div class="login-join">
          <div class="field">
            <svg class="field-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M3 6h18v4a2 2 0 0 0 0 4v4H3v-4a2 2 0 0 0 0-4V6Zm6 3v2h6V9H9Zm0 4v2h6v-2H9Z" /></svg>
            <input
              v-model="roomCodeInput"
              type="text"
              placeholder="请输入房间码"
              aria-label="房间码"
              class="uppercase"
              maxlength="4"
              @keyup.enter="joinRoom"
            />
          </div>
          <button class="blue-btn" :disabled="isJoining" @click="joinRoom">加入</button>
        </div>

        <div v-if="displayError" class="login-error" role="alert">{{ displayError }}</div>
        <div v-if="isJoining" class="login-pending">
          <span class="spinner" aria-hidden="true"></span>
          连接中...
        </div>
      </div>
    </div>

    <!-- 英雄选择 -->
    <div v-else class="draft-stage">
      <header class="draft-topbar">
        <div class="topbar-brand">
          <img class="topbar-emblem" src="/assets/ui/star-emblem.svg" alt="" aria-hidden="true" />
          <div>
            <div class="topbar-title">星杯传说</div>
            <div class="topbar-sub">STAR CUP LEGEND</div>
          </div>
        </div>
        <div class="topbar-center">
          <div class="topbar-heading">英雄选择</div>
          <div class="topbar-caption">3v3 卡牌对战游戏</div>
        </div>
        <div class="topbar-ready" :class="{ ready: allReadyToStart }">
          已锁定 {{ roomPlayers.filter(p => p.char_role).length }}/{{ roomPlayers.length }}
        </div>
      </header>

      <div class="draft-info-row">
        <div class="room-box ornate-panel">
          <div class="room-box-label">房间码</div>
          <!-- 房间码本身可选中复制；按钮只包“复制”字样（按钮内的文字浏览器不让选中） -->
          <div class="room-code-btn">
            <span class="room-code" data-testid="room-code">{{ sessionStore.roomCode }}</span>
            <button type="button" class="room-copy-hint" @click="copyRoomCode">{{ copyFeedback ? '✓ 已复制' : '复制' }}</button>
          </div>
        </div>

        <div class="status-box ornate-panel" :class="{ 'status-ready': allReadyToStart }">
          <div class="status-text">
            <div class="status-main">{{ lobbyHint }}</div>
            <div class="status-sub">蓝方 {{ blueCount }}/3 · 红方 {{ redCount }}/3 · 总人数 {{ roomPlayers.length }}/6</div>
          </div>
          <div class="status-actions">
            <button class="blue-btn" @click="copyInviteLink">{{ inviteFeedback ? '✓ 链接已复制' : '+ 邀请队友' }}</button>
            <button
              v-if="isHost && !sessionStore.gameStarted && roomPlayers.length < 6"
              class="blue-btn"
              data-testid="add-bot-button"
              @click="addBot"
            >
              + 添加机器人
            </button>
            <button
              v-if="isHost && allReadyToStart && !sessionStore.gameStarted"
              class="gold-btn compact"
              data-testid="start-game-button"
              @click="startGame"
            >
              手动开始
            </button>
            <button v-if="isHost" class="bronze-btn" @click="dissolveRoom">解散房间</button>
          </div>
        </div>
      </div>

      <div class="draft-main">
        <aside
          v-for="team in teams"
          :key="team.camp"
          class="team-panel ornate-panel"
          :class="`team-${team.key}`"
        >
          <div class="team-head">
            <span class="team-badge" aria-hidden="true">✦</span>
            <span class="team-title">{{ team.label }}</span>
            <button
              v-if="!sessionStore.gameStarted"
              class="team-join-btn"
              :data-testid="`join-camp-${team.key}`"
              :disabled="!canJoinCamp(team.camp)"
              @click="selectCamp(team.camp)"
            >
              {{ sessionStore.myCamp === team.camp ? '已加入' : '加入' }}
            </button>
          </div>

          <div class="team-slots">
            <div
              v-for="(player, idx) in team.slots"
              :key="`${team.key}-${idx}`"
              class="team-slot"
              :class="{ filled: !!player, me: player?.id === sessionStore.myPlayerId, locked: !!player?.char_role }"
            >
              <template v-if="player">
                <button
                  type="button"
                  class="slot-avatar"
                  :disabled="!player.char_role"
                  :aria-label="player.char_role ? `${getCharacterName(player.char_role)} 技能详情` : '未选择角色'"
                  @click="player.char_role && openSkillModal(player.char_role)"
                >
                  <img
                    v-if="player.char_role"
                    :src="portraitSrc(player.char_role)"
                    :alt="getCharacterName(player.char_role)"
                    @error="onPortraitError(player.char_role, $event)"
                  />
                  <svg v-else viewBox="0 0 24 24" aria-hidden="true"><path d="M12 12a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9Zm0 2c-4.4 0-8 2.4-8 5.4V21h16v-1.6c0-3-3.6-5.4-8-5.4Z" /></svg>
                </button>
                <div class="slot-info">
                  <div class="slot-name">{{ player.name }}</div>
                  <div class="slot-role">{{ player.char_role ? getCharacterName(player.char_role) : '选择中…' }}</div>
                  <div class="slot-tags">
                    <span v-if="player.id === sessionStore.myPlayerId" class="slot-tag me">你</span>
                    <span v-if="player.is_host" class="slot-tag host">房主</span>
                    <span v-if="player.is_bot" class="slot-tag bot">BOT</span>
                    <span v-if="!player.is_bot && player.is_online === false" class="slot-tag offline">离线</span>
                  </div>
                </div>

                <div v-if="player.is_bot && isHost && !sessionStore.gameStarted" class="slot-bot-controls">
                  <select
                    class="bot-role-select"
                    :data-testid="`bot-role-${player.id}`"
                    :value="player.char_role || ''"
                    aria-label="机器人角色"
                    @change="selectRoleFor(player.id, ($event.target as HTMLSelectElement).value)"
                  >
                    <option value="">选择角色</option>
                    <option
                      v-for="c in characterOptions"
                      :key="`${team.key}-bot-${player.id}-${c.id}`"
                      :value="c.id"
                      :disabled="roomPlayers.some(p => p.id !== player.id && p.char_role === c.id)"
                    >
                      {{ c.name }}（{{ c.title }}）
                    </option>
                  </select>
                  <div class="bot-btn-row">
                    <button
                      class="bot-mini-btn"
                      :data-testid="`bot-camp-${team.otherKey}-${player.id}`"
                      :disabled="team.otherKey === 'red' ? redCount >= 3 : blueCount >= 3"
                      @click="selectCampFor(player.id, team.otherCamp)"
                    >
                      改去{{ team.otherKey === 'red' ? '红方' : '蓝方' }}
                    </button>
                    <button class="bot-mini-btn danger" @click="removeBot(player.id)">移除</button>
                  </div>
                </div>
              </template>
              <template v-else>
                <span class="slot-avatar empty" aria-hidden="true">
                  <svg viewBox="0 0 24 24"><path d="M12 12a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9Zm0 2c-4.4 0-8 2.4-8 5.4V21h16v-1.6c0-3-3.6-5.4-8-5.4Z" /></svg>
                </span>
                <span class="slot-empty-text">空位</span>
              </template>
            </div>
          </div>
        </aside>

        <section class="pool-panel ornate-panel">
          <div v-if="unassignedPlayers.length > 0" class="unassigned-strip">
            <span class="unassigned-title">待分配阵营</span>
            <div v-for="player in unassignedPlayers" :key="`pending-${player.id}`" class="pending-chip">
              <span>{{ player.name }}</span>
              <span v-if="player.id === sessionStore.myPlayerId" class="pending-me">你</span>
              <template v-if="player.is_bot && isHost && !sessionStore.gameStarted">
                <button class="pending-btn blue" :data-testid="`bot-camp-blue-${player.id}`" :disabled="blueCount >= 3" @click="selectCampFor(player.id, 'Blue')">蓝方</button>
                <button class="pending-btn red" :data-testid="`bot-camp-red-${player.id}`" :disabled="redCount >= 3" @click="selectCampFor(player.id, 'Red')">红方</button>
              </template>
            </div>
          </div>

          <div class="pool-toolbar">
            <div class="faction-tabs" role="tablist" aria-label="派系筛选">
              <button
                type="button"
                role="tab"
                class="faction-tab"
                :class="{ active: !factionFilter }"
                :aria-selected="!factionFilter"
                @click="setFaction('')"
              >
                全部
              </button>
              <button
                v-for="faction in factionOptions"
                :key="faction"
                type="button"
                role="tab"
                class="faction-tab"
                :class="{ active: factionFilter === faction }"
                :aria-selected="factionFilter === faction"
                @click="setFaction(faction)"
              >
                {{ faction }}
              </button>
            </div>
            <label class="pool-search">
              <input
                v-model="searchKeyword"
                type="search"
                placeholder="搜索角色..."
                aria-label="搜索角色"
                data-testid="role-search"
                @input="onSearchInput"
              />
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M10.5 3a7.5 7.5 0 0 1 5.9 12.1l4.8 4.8-1.4 1.4-4.8-4.8A7.5 7.5 0 1 1 10.5 3Zm0 2a5.5 5.5 0 1 0 0 11 5.5 5.5 0 0 0 0-11Z" /></svg>
            </label>
          </div>

          <div class="role-grid">
            <button
              v-for="role in pagedRoles"
              :key="role.id"
              type="button"
              class="role-card"
              :class="{
                previewing: previewCharacter?.id === role.id,
                mine: sessionStore.myCharRole === role.id,
                taken: isRoleTakenByOther(role.id),
              }"
              :data-testid="`role-card-${role.id}`"
              :aria-pressed="previewCharacter?.id === role.id"
              @click="previewRole(role.id)"
              @dblclick="pickRole(role.id)"
            >
              <img
                class="role-portrait"
                :src="portraitSrc(role.id)"
                :alt="role.name"
                loading="lazy"
                @error="onPortraitError(role.id, $event)"
              />
              <span class="role-faction" aria-hidden="true">{{ role.faction }}</span>
              <span class="role-name">{{ role.name }}</span>
              <span v-if="roleTakenBy(role.id)" class="role-owner">
                {{ sessionStore.myCharRole === role.id ? '已锁定' : roleTakenBy(role.id)?.name }}
              </span>
            </button>
            <div v-if="pagedRoles.length === 0" class="role-empty">没有符合条件的角色</div>
          </div>

          <div class="pool-pager">
            <button type="button" class="pager-btn" aria-label="上一页" :disabled="page <= 0" @click="changePage(-1)">‹</button>
            <span class="pager-text">{{ Math.min(page, pageCount - 1) + 1 }} / {{ pageCount }}</span>
            <button type="button" class="pager-btn" aria-label="下一页" :disabled="page >= pageCount - 1" @click="changePage(1)">›</button>
          </div>
        </section>
      </div>

      <footer class="preview-bar ornate-panel">
        <template v-if="previewCharacter">
          <button type="button" class="preview-portrait" aria-label="查看技能详情" @click="openSkillModal(previewCharacter.id)">
            <img
              :src="portraitSrc(previewCharacter.id)"
              :alt="previewCharacter.name"
              @error="onPortraitError(previewCharacter.id, $event)"
            />
          </button>
          <div class="preview-info">
            <div class="preview-head">
              <span class="preview-name">{{ previewCharacter.name }}</span>
              <span class="preview-tag">{{ previewCharacter.faction }} / {{ previewCharacter.title }}</span>
              <button type="button" class="preview-detail" @click="openSkillModal(previewCharacter.id)">技能详情</button>
            </div>
            <p class="preview-desc">{{ previewSkillLine || '点击「技能详情」查看完整技能描述。' }}</p>
          </div>
        </template>
        <div v-else class="preview-empty">点击上方英雄卡查看介绍，确认后锁定。</div>

        <button
          class="gold-btn preview-confirm"
          data-testid="confirm-role-button"
          :disabled="!canConfirmPreview"
          @click="confirmPreview"
        >
          {{ previewLockedByMe ? '已锁定' : previewCharacter && isRoleTakenByOther(previewCharacter.id) ? '已被选择' : '确认选择' }}
        </button>
      </footer>

      <div v-if="displayError" class="draft-error" role="alert">{{ displayError }}</div>
    </div>

    <SkillDetailModal
      :character="skillModalCharacter"
      :visible="!!skillModalCharacter"
      @close="closeSkillModal"
    />
  </div>
</template>

<style scoped>
.lobby-shell {
  position: relative;
  min-height: calc(100 * var(--vh, 1vh));
  height: calc(100 * var(--vh, 1vh));
  overflow-y: auto;
  color: #e6ecf5;
  background-color: #070b1c;
  background-position: center;
  background-size: cover;
  padding: max(16px, var(--safe-top)) max(20px, var(--safe-right)) calc(16px + var(--safe-bottom)) max(20px, var(--safe-left));
}

.lobby-shell-login {
  background-image:
    radial-gradient(ellipse at 50% 45%, rgba(10, 14, 38, 0.15), rgba(4, 6, 18, 0.7) 85%),
    url('/assets/bg/arena_starlit_cathedral.webp');
}

.lobby-shell-draft {
  background-image:
    linear-gradient(180deg, rgba(6, 9, 26, 0.72), rgba(5, 7, 20, 0.88)),
    url('/assets/bg/arena_starlit_cathedral.webp');
}

/* ---------- 通用组件 ---------- */
.ornate-panel {
  position: relative;
  border: 1px solid rgba(201, 170, 113, 0.42);
  border-radius: 6px;
  background: linear-gradient(180deg, rgba(16, 20, 44, 0.86), rgba(9, 12, 30, 0.9));
  box-shadow:
    inset 0 0 0 1px rgba(255, 236, 196, 0.05),
    inset 0 0 24px rgba(80, 100, 190, 0.08),
    0 12px 30px rgba(0, 0, 0, 0.45);
}

/* 四角饰线 */
.ornate-panel::before,
.ornate-panel::after {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  pointer-events: none;
  border-color: rgba(232, 202, 140, 0.8);
  border-style: solid;
}
.ornate-panel::before {
  top: -1px;
  left: -1px;
  border-width: 2px 0 0 2px;
}
.ornate-panel::after {
  right: -1px;
  bottom: -1px;
  border-width: 0 2px 2px 0;
}

.gold-btn,
.blue-btn,
.bronze-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border-radius: 6px;
  font-weight: 600;
  letter-spacing: 0.08em;
  transition: filter 0.15s ease, transform 0.15s ease, box-shadow 0.15s ease;
  white-space: nowrap;
}

.gold-btn {
  color: #fff7e4;
  border: 1px solid #e3c27e;
  background: linear-gradient(180deg, #a2804a 0%, #6d522a 55%, #4f3a1d 100%);
  box-shadow: inset 0 1px 0 rgba(255, 236, 190, 0.45), 0 0 14px rgba(226, 186, 110, 0.25);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}
.gold-btn svg {
  width: 18px;
  height: 18px;
  fill: currentColor;
}
.gold-btn.compact {
  padding: 8px 14px;
}

.blue-btn {
  padding: 8px 16px;
  color: #e9f1ff;
  border: 1px solid rgba(124, 160, 230, 0.7);
  background: linear-gradient(180deg, #2b4a86 0%, #1a2f5c 100%);
  box-shadow: inset 0 1px 0 rgba(190, 215, 255, 0.25);
}

.bronze-btn {
  padding: 8px 16px;
  color: #fbe9cf;
  border: 1px solid rgba(214, 170, 112, 0.75);
  background: linear-gradient(180deg, #7a5634 0%, #4a321c 100%);
  box-shadow: inset 0 1px 0 rgba(255, 220, 170, 0.28);
}

.gold-btn:hover:not(:disabled),
.blue-btn:hover:not(:disabled),
.bronze-btn:hover:not(:disabled) {
  filter: brightness(1.15);
}

.gold-btn:disabled,
.blue-btn:disabled,
.bronze-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

button:focus-visible,
input:focus-visible,
select:focus-visible {
  outline: 2px solid #f0d08f;
  outline-offset: 2px;
}

/* ---------- 登录 ---------- */
.login-stage {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 28px;
  padding-bottom: calc(6 * var(--vh, 1vh));
}

.login-brand {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.brand-emblem {
  width: 110px;
  height: 110px;
  margin-bottom: -18px;
  filter: drop-shadow(0 0 14px rgba(255, 224, 160, 0.45));
}

.brand-title {
  font-family: "Songti SC", "STSong", "Noto Serif SC", serif;
  font-size: clamp(44px, calc(6 * var(--vw, 1vw)), 72px);
  font-weight: 700;
  letter-spacing: 0.12em;
  background: linear-gradient(180deg, #fff3d6 0%, #e7c98d 55%, #a98446 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  filter: drop-shadow(0 3px 10px rgba(0, 0, 0, 0.6));
}

.brand-sub {
  display: flex;
  white-space: nowrap;
  align-items: center;
  gap: 14px;
  margin-top: 2px;
  font-family: "Times New Roman", serif;
  font-size: 17px;
  letter-spacing: 0.22em;
  color: #e6d3a6;
}

.brand-line {
  width: clamp(24px, calc(10 * var(--vw, 1vw)), 90px);
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(230, 205, 150, 0.8), transparent);
}

.login-card {
  width: min(520px, 100%);
  padding: 34px 36px 30px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.panel-crest {
  position: absolute;
  top: -13px;
  left: 50%;
  transform: translateX(-50%);
  color: #f0d595;
  font-size: 22px;
  line-height: 1;
  text-shadow: 0 0 10px rgba(255, 220, 150, 0.8);
}

.field-label {
  font-size: 15px;
  color: #dfe5f2;
}

.field {
  position: relative;
  display: flex;
  align-items: center;
  flex: 1;
}

.field-icon {
  position: absolute;
  left: 16px;
  width: 20px;
  height: 20px;
  fill: #aab4cc;
  pointer-events: none;
}

.field input {
  width: 100%;
  height: 52px;
  padding: 0 16px 0 48px;
  border-radius: 6px;
  border: 1px solid rgba(150, 165, 205, 0.32);
  background: rgba(10, 13, 30, 0.78);
  color: #f2f5fb;
  font-size: 16px;
}

.field input::placeholder {
  color: #7f89a3;
  text-transform: none;
}

.field input:focus {
  border-color: rgba(230, 200, 135, 0.75);
  outline: none;
  box-shadow: 0 0 0 3px rgba(230, 200, 135, 0.15);
}

.login-create {
  height: 54px;
  font-size: 18px;
  margin-top: 8px;
}

.login-divider {
  display: flex;
  align-items: center;
  gap: 16px;
  color: #aab3c8;
  font-size: 14px;
}
.login-divider::before,
.login-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: rgba(170, 180, 210, 0.25);
}

.login-join {
  display: flex;
  gap: 12px;
}

.login-join .blue-btn {
  width: 104px;
  height: 52px;
  font-size: 17px;
}

.login-error {
  color: #ff9a8f;
  font-size: 14px;
  text-align: center;
}

.login-pending {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  color: #b9c3da;
  font-size: 14px;
}

.spinner {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid currentColor;
  border-top-color: transparent;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* ---------- 选将 ---------- */
.draft-stage {
  max-width: 1600px;
  min-height: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.draft-topbar {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
}

.topbar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.topbar-emblem {
  width: 54px;
  height: 54px;
}

.topbar-title {
  font-family: "Songti SC", "STSong", "Noto Serif SC", serif;
  font-size: 26px;
  font-weight: 700;
  letter-spacing: 0.12em;
  color: #f1dfb6;
}

.topbar-sub {
  font-family: "Times New Roman", serif;
  font-size: 11px;
  letter-spacing: 0.2em;
  color: #c9b68d;
}

.topbar-center {
  text-align: center;
}

.topbar-heading {
  font-family: "Songti SC", "STSong", "Noto Serif SC", serif;
  font-size: 28px;
  font-weight: 700;
  letter-spacing: 0.16em;
  color: #f6ead0;
}
.topbar-heading::before,
.topbar-heading::after {
  content: '✦';
  margin: 0 14px;
  font-size: 18px;
  color: #e3c47f;
  vertical-align: middle;
}

.topbar-caption {
  font-size: 13px;
  color: #b8c0d6;
}

.topbar-ready {
  justify-self: end;
  padding: 8px 18px;
  border-radius: 999px;
  border: 1px solid rgba(180, 190, 220, 0.3);
  background: rgba(10, 14, 32, 0.7);
  color: #d4dbea;
  font-size: 14px;
}
.topbar-ready.ready {
  color: #f6dc9b;
  border-color: rgba(230, 200, 135, 0.6);
}

.draft-info-row {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 14px;
}

.room-box {
  padding: 12px 18px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.room-box-label {
  font-size: 13px;
  color: #c7cfe0;
}

.room-code-btn {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid rgba(160, 170, 210, 0.3);
  background: rgba(6, 9, 24, 0.7);
}

.room-code {
  user-select: all;
  cursor: text;
  color: #f3d389;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: 0.18em;
}

.room-copy-hint {
  padding: 0;
  border: 0;
  background: none;
  cursor: pointer;
  font: inherit;
  font-size: 13px;
  color: #c3cbe0;
}
.room-copy-hint:hover,
.room-copy-hint:focus-visible {
  color: #f3d389;
}

.status-box {
  padding: 12px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.status-box.status-ready {
  border-color: rgba(240, 210, 140, 0.75);
}

.status-main {
  font-size: 16px;
  color: #eef1f8;
}

.status-sub {
  margin-top: 4px;
  font-size: 13px;
  color: #aeb7cc;
}

.status-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.draft-main {
  flex: 1;
  display: grid;
  grid-template-columns: 270px minmax(0, 1fr) 270px;
  gap: 14px;
  min-height: 0;
}

.team-blue {
  order: 1;
}
.pool-panel {
  order: 2;
}
.team-red {
  order: 3;
}

.team-panel {
  display: flex;
  flex-direction: column;
  padding: 0 12px 12px;
}

.team-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 -12px 12px;
  padding: 10px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.team-blue .team-head {
  background: linear-gradient(90deg, rgba(48, 86, 170, 0.75), rgba(25, 40, 90, 0.2));
}
.team-red .team-head {
  flex-direction: row-reverse;
  background: linear-gradient(270deg, rgba(150, 38, 52, 0.75), rgba(80, 20, 30, 0.2));
}

.team-badge {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  font-size: 20px;
}
.team-blue .team-badge {
  color: #cfe0ff;
  background: radial-gradient(circle, #3d66c4, #1a2c62);
  box-shadow: 0 0 12px rgba(90, 140, 255, 0.6);
}
.team-red .team-badge {
  color: #ffd6d6;
  background: radial-gradient(circle, #c13a48, #5e1520);
  box-shadow: 0 0 12px rgba(255, 90, 100, 0.55);
}

.team-title {
  flex: 1;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.12em;
}
.team-blue .team-title {
  color: #dbe6ff;
}
.team-red .team-title {
  color: #ffdede;
  text-align: right;
}

.team-join-btn {
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 13px;
  border: 1px solid rgba(255, 255, 255, 0.35);
  background: rgba(0, 0, 0, 0.25);
  color: #f2f4fa;
}
.team-join-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.team-slots {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.team-slot {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  min-height: 92px;
  padding: 10px 12px;
  border-radius: 6px;
  border: 1px solid rgba(160, 170, 210, 0.16);
  background: rgba(6, 9, 24, 0.55);
}
.team-red .team-slot {
  flex-direction: row-reverse;
  text-align: right;
}
.team-slot.filled {
  border-color: rgba(170, 185, 225, 0.32);
}
.team-blue .team-slot.locked {
  border-color: rgba(110, 160, 255, 0.65);
  box-shadow: inset 0 0 18px rgba(80, 130, 255, 0.15);
}
.team-red .team-slot.locked {
  border-color: rgba(255, 110, 120, 0.6);
  box-shadow: inset 0 0 18px rgba(255, 90, 100, 0.13);
}
.team-slot.me {
  outline: 1px solid rgba(240, 210, 140, 0.55);
}

.slot-avatar {
  flex-shrink: 0;
  display: grid;
  place-items: center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  overflow: hidden;
  border: 1px solid rgba(180, 190, 220, 0.35);
  background: rgba(40, 46, 72, 0.8);
}
.slot-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: top;
}
.slot-avatar svg {
  width: 38px;
  height: 38px;
  fill: #5a6382;
  margin-top: 8px;
}
.slot-avatar:disabled {
  cursor: default;
}

.slot-info {
  flex: 1;
  min-width: 0;
}

.slot-name {
  font-size: 15px;
  font-weight: 600;
  color: #eef1f8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.slot-role {
  margin-top: 2px;
  font-size: 13px;
  color: #f0d595;
}

.slot-tags {
  display: flex;
  gap: 4px;
  margin-top: 4px;
}
.team-red .slot-tags {
  justify-content: flex-end;
}

.slot-tag {
  padding: 0 6px;
  border-radius: 3px;
  font-size: 11px;
  line-height: 18px;
  background: rgba(255, 255, 255, 0.1);
  color: #d6dbe8;
}
.slot-tag.me {
  background: rgba(240, 210, 140, 0.25);
  color: #f6dfa5;
}
.slot-tag.host {
  background: rgba(120, 160, 255, 0.25);
  color: #cddcff;
}
.slot-tag.offline {
  background: rgba(255, 110, 110, 0.25);
  color: #ffc1c1;
}

.slot-empty-text {
  color: #8790aa;
  font-size: 15px;
}

.slot-bot-controls {
  flex-basis: 100%;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.bot-role-select {
  width: 100%;
  padding: 5px 8px;
  border-radius: 4px;
  border: 1px solid rgba(160, 170, 210, 0.3);
  background: rgba(6, 9, 24, 0.85);
  color: #e6ecf5;
  font-size: 12px;
}

.bot-btn-row {
  display: flex;
  gap: 6px;
}

.bot-mini-btn {
  flex: 1;
  padding: 4px 0;
  border-radius: 4px;
  font-size: 12px;
  border: 1px solid rgba(160, 170, 210, 0.3);
  background: rgba(30, 40, 75, 0.7);
  color: #dfe5f2;
}
.bot-mini-btn.danger {
  border-color: rgba(255, 120, 120, 0.4);
  color: #ffc4c4;
}
.bot-mini-btn:disabled {
  opacity: 0.45;
}

.pool-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 14px 18px;
  min-width: 0;
}

.unassigned-strip {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px dashed rgba(230, 200, 135, 0.4);
  background: rgba(30, 26, 14, 0.35);
}

.unassigned-title {
  font-size: 13px;
  color: #f0d595;
}

.pending-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 13px;
  background: rgba(255, 255, 255, 0.08);
}

.pending-me {
  color: #f6dfa5;
}

.pending-btn {
  padding: 1px 8px;
  border-radius: 3px;
  font-size: 12px;
}
.pending-btn.blue {
  background: #2b4a86;
}
.pending-btn.red {
  background: #8a2a36;
}
.pending-btn:disabled {
  opacity: 0.45;
}

.pool-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.faction-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.faction-tab {
  min-width: 64px;
  padding: 6px 16px;
  border-radius: 999px;
  border: 1px solid rgba(160, 170, 210, 0.25);
  background: rgba(14, 18, 40, 0.8);
  color: #c9d0e2;
  font-size: 14px;
}
.faction-tab.active {
  border-color: rgba(120, 165, 255, 0.8);
  background: linear-gradient(180deg, #2c4f96, #1b3266);
  color: #fff;
}

.pool-search {
  position: relative;
  width: min(290px, 100%);
}
.pool-search input {
  width: 100%;
  height: 38px;
  padding: 0 40px 0 16px;
  border-radius: 999px;
  border: 1px solid rgba(160, 170, 210, 0.3);
  background: rgba(6, 9, 24, 0.8);
  color: #eef1f8;
  font-size: 14px;
}
.pool-search svg {
  position: absolute;
  right: 14px;
  top: 50%;
  width: 18px;
  height: 18px;
  transform: translateY(-50%);
  fill: #aab4cc;
  pointer-events: none;
}

.role-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  grid-auto-rows: min-content;
  gap: 16px 18px;
  align-content: start;
}

.role-card {
  position: relative;
  aspect-ratio: 3 / 4;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid rgba(160, 140, 200, 0.45);
  background: #12152c;
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
}
.role-card:hover:not(.taken) {
  transform: translateY(-3px);
  border-color: rgba(240, 210, 140, 0.8);
}
.role-card.previewing {
  border: 2px solid #f0d08f;
  box-shadow: 0 0 0 2px rgba(240, 208, 143, 0.25), 0 0 20px rgba(240, 208, 143, 0.45);
}
.role-card.mine {
  border: 2px solid #f0d08f;
}
.role-card.taken .role-portrait {
  filter: grayscale(0.9) brightness(0.55);
}

.role-portrait {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: top;
}

.role-faction {
  position: absolute;
  top: 6px;
  left: 6px;
  display: grid;
  place-items: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 1px solid #e3c47f;
  background: rgba(20, 16, 40, 0.85);
  color: #f3d796;
  font-size: 12px;
  line-height: 1;
}

.role-name {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 18px 4px 6px;
  text-align: center;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(180deg, transparent, rgba(6, 8, 20, 0.92) 55%);
}

.role-owner {
  position: absolute;
  top: 40%;
  left: 0;
  right: 0;
  padding: 3px 4px;
  text-align: center;
  font-size: 12px;
  color: #fff;
  background: rgba(0, 0, 0, 0.6);
}
.role-card.mine .role-owner {
  background: rgba(150, 115, 50, 0.85);
}

.role-empty {
  grid-column: 1 / -1;
  padding: 40px 0;
  text-align: center;
  color: #8790aa;
}

.pool-pager {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 4px;
}

.pager-btn {
  width: 48px;
  height: 32px;
  font-size: 22px;
  line-height: 1;
  color: #fff;
  border: 1px solid rgba(120, 160, 230, 0.6);
  background: linear-gradient(180deg, #2b4a86, #1a2f5c);
}
.pager-btn:first-child {
  border-radius: 16px 0 0 16px;
}
.pager-btn:last-child {
  border-radius: 0 16px 16px 0;
}
.pager-btn:disabled {
  opacity: 0.4;
}

.pager-text {
  min-width: 140px;
  height: 32px;
  display: grid;
  place-items: center;
  color: #cfd6e6;
  background: rgba(6, 9, 24, 0.8);
  border-top: 1px solid rgba(160, 170, 210, 0.25);
  border-bottom: 1px solid rgba(160, 170, 210, 0.25);
}

.preview-bar {
  display: flex;
  align-items: center;
  gap: 20px;
  min-height: 132px;
  padding: 12px 22px 12px 14px;
}

.preview-portrait {
  flex-shrink: 0;
  width: 150px;
  height: 108px;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid #e3c47f;
}
.preview-portrait img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: top;
}

.preview-info {
  flex: 1;
  min-width: 0;
}

.preview-head {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.preview-name {
  font-family: "Songti SC", "STSong", "Noto Serif SC", serif;
  font-size: 26px;
  font-weight: 700;
  color: #f6e2b0;
}
.preview-name::before {
  content: '✦ ';
  color: #e3c47f;
  font-size: 18px;
}

.preview-tag {
  padding: 3px 12px;
  border-radius: 999px;
  border: 1px solid rgba(180, 190, 220, 0.35);
  color: #d4dbea;
  font-size: 13px;
}

.preview-detail {
  font-size: 13px;
  color: #9ec0ff;
  text-decoration: underline;
  text-underline-offset: 3px;
}

.preview-desc {
  margin-top: 8px;
  font-size: 14px;
  line-height: 1.6;
  color: #d2d8e6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.preview-empty {
  flex: 1;
  color: #9aa3bb;
}

.preview-confirm {
  flex-shrink: 0;
  width: 250px;
  height: 64px;
  font-size: 22px;
  letter-spacing: 0.2em;
  border-radius: 10px;
}

.draft-error {
  color: #ff9a8f;
  font-size: 14px;
  text-align: center;
}

@container stage (max-width: 1400px) {
  .role-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@container stage (max-width: 1100px) {
  .draft-main {
    grid-template-columns: 1fr 1fr;
  }
  .pool-panel {
    grid-column: 1 / -1;
    order: 0;
  }
  .draft-info-row {
    grid-template-columns: 1fr;
  }
}

@container stage (max-width: 640px) {
  .draft-topbar {
    grid-template-columns: 1fr auto;
  }
  .topbar-center {
    display: none;
  }
  .draft-main {
    grid-template-columns: 1fr;
  }
  .role-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 10px;
  }
  .preview-bar {
    flex-wrap: wrap;
  }
  .preview-portrait {
    width: 96px;
    height: 80px;
  }
  .preview-confirm {
    width: 100%;
    height: 52px;
  }
  .login-card {
    padding: 28px 20px 22px;
  }
}

/* 手机横屏等矮屏（大厅不缩放，逻辑高度即视口高度 ≈320–430）：
 * logo 与标题缩小、和登录卡左右并排，压缩卡片内边距，让输入框、「创建房间」「加入」都在首屏。 */
@container stage (max-height: 520px) {
  .lobby-shell {
    padding: max(8px, var(--safe-top)) max(12px, var(--safe-right)) calc(8px + var(--safe-bottom)) max(12px, var(--safe-left));
  }

  .login-stage {
    flex-direction: row;
    gap: clamp(12px, calc(4 * var(--vw, 1vw)), 40px);
    padding-bottom: 0;
  }

  .login-brand {
    flex: 0 1 auto;
    min-width: 0;
  }

  .brand-emblem {
    width: clamp(44px, calc(16 * var(--vh, 1vh)), 72px);
    height: clamp(44px, calc(16 * var(--vh, 1vh)), 72px);
    margin-bottom: -8px;
  }

  .brand-title {
    font-size: clamp(26px, min(calc(11 * var(--vh, 1vh)), calc(5.4 * var(--vw, 1vw))), 44px);
    letter-spacing: 0.06em;
    white-space: nowrap;
  }

  .brand-sub {
    gap: 8px;
    font-size: 11px;
    letter-spacing: 0.16em;
  }

  .brand-line {
    width: clamp(12px, calc(3 * var(--vw, 1vw)), 36px);
  }

  .login-card {
    flex: 0 1 380px;
    width: auto;
    min-width: 0;
    padding: 16px 18px 14px;
    gap: 8px;
  }

  .panel-crest {
    top: -11px;
    font-size: 18px;
  }

  .field-label {
    font-size: 14px;
  }

  .field input {
    height: 42px;
    font-size: 16px;
  }

  .login-create {
    height: 42px;
    margin-top: 2px;
    font-size: 16px;
  }

  .login-divider {
    font-size: 13px;
  }

  .login-join {
    gap: 8px;
  }

  .login-join .blue-btn {
    width: 84px;
    height: 42px;
    font-size: 16px;
  }
}

/* 特别窄的横屏（iPhone SE 568 宽）：去掉副标题，名称标签收进占位文字 */
@container stage (max-height: 520px) and (max-width: 640px) {
  .brand-sub {
    display: none;
  }

  .field-label {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
  }

  .login-card {
    padding-top: 14px;
  }
}
</style>
