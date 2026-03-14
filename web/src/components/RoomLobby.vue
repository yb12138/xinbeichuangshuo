<script setup lang="ts">
import { ref, computed } from 'vue'
import { useGameStore } from '../stores/gameStore'
import { useWebSocket } from '../composables/useWebSocket'
import SkillDetailModal from './SkillDetailModal.vue'

const store = useGameStore()
const ws = useWebSocket()

const playerName = ref('')
const roomCodeInput = ref('')
const isJoining = ref(false)
const errorMsg = ref('')
const copyFeedback = ref(false)
const skillModalRoleId = ref('')
const portraitFallbackStep = ref<Record<string, number>>({})

const IMAGE_EXTS = ['png', 'webp', 'jpg', 'jpeg']

async function copyRoomCode() {
  try {
    await navigator.clipboard.writeText(store.roomCode)
    copyFeedback.value = true
    setTimeout(() => {
      copyFeedback.value = false
    }, 1500)
  } catch {
    errorMsg.value = '复制失败'
  }
}

const roomPlayers = computed(() => store.roomPlayers)
const isHost = computed(() => roomPlayers.value.some(p => p.id === store.myPlayerId && p.is_host))
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

const characterOptions = computed(() => Object.values(store.characters))
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

const skillModalCharacter = computed(() => {
  if (!skillModalRoleId.value) return null
  return characterMap.value.get(skillModalRoleId.value) ?? null
})
const displayError = computed(() => errorMsg.value || store.errorMessage)

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

function getCharacterName(roleId: string) {
  if (!roleId) return '未选择角色'
  return store.getRoleDisplayName(roleId)
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
  return !!owner && owner.id !== store.myPlayerId
}

function canSelectRole(roleId: string) {
  if (store.gameStarted) return false
  return !isRoleTakenByOther(roleId)
}

function openSkillModal(roleId: string) {
  skillModalRoleId.value = roleId
}

function closeSkillModal() {
  skillModalRoleId.value = ''
}

function selectCamp(camp: string) {
  ws.sendRoomAction('change_camp', { camp })
}

function selectRole(role: string) {
  if (!role) return
  ws.sendRoomAction('change_role', { char_role: role })
}

function selectCampFor(playerId: string, camp: string) {
  ws.sendRoomAction('change_camp', { target_id: playerId, camp })
}

function selectRoleFor(playerId: string, role: string) {
  if (!role) return
  ws.sendRoomAction('change_role', { target_id: playerId, char_role: role })
}

function pickRole(roleId: string) {
  if (!canSelectRole(roleId)) return
  selectRole(roleId)
}

function canJoinCamp(camp: 'Red' | 'Blue') {
  if (store.myCamp === camp) return true
  if (camp === 'Red') return redCount.value < 3
  return blueCount.value < 3
}

function addBot() {
  ws.sendRoomAction('add_bot', { bot_name: `机器人${botCount.value + 1}` })
}

function removeBot(playerId: string) {
  ws.sendRoomAction('remove_bot', { target_id: playerId })
}

function startGame() {
  ws.sendRoomAction('start')
}

function dissolveRoom() {
  if (!isHost.value) return
  const confirmed = window.confirm('确认解散房间吗？所有玩家将被退出到大厅。')
  if (!confirmed) return
  ws.sendRoomAction('dissolve_room')
}
</script>

<template>
  <div class="lobby-shell min-h-[100dvh] max-h-[100dvh] flex flex-col overflow-hidden">
    <div class="lobby-scroll flex-1 overflow-y-auto py-6 px-4 sm:px-6">
      <div class="lobby-container w-full mx-auto">
        <div class="lobby-title-wrap text-center mb-6">
          <h1 class="text-4xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-yellow-400 to-orange-500 mb-2">
            星杯传说
          </h1>
          <p class="text-gray-400">3v3 卡牌对战游戏</p>
        </div>

        <div v-if="!store.isInRoom" class="lobby-card rounded-2xl p-6 shadow-xl max-w-md mx-auto">
          <div class="mb-6">
            <label class="block text-sm text-gray-400 mb-2">玩家名称</label>
            <input
              v-model="playerName"
              type="text"
              placeholder="输入你的名字"
              class="w-full px-4 py-3 bg-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-yellow-500"
              maxlength="20"
            />
          </div>

          <div class="space-y-4">
            <button
              class="w-full py-3 bg-gradient-to-r from-yellow-500 to-orange-500 text-white font-bold rounded-lg hover:from-yellow-400 hover:to-orange-400 transition-all transform hover:scale-105 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="isJoining"
              @click="createRoom"
            >
              🏠 创建房间
            </button>

            <div class="relative">
              <div class="absolute inset-0 flex items-center">
                <div class="w-full border-t border-gray-600"></div>
              </div>
              <div class="relative flex justify-center text-sm">
                <span class="px-4 bg-gray-800 text-gray-400">或者</span>
              </div>
            </div>

            <div class="flex gap-2">
              <input
                v-model="roomCodeInput"
                type="text"
                placeholder="房间码"
                class="flex-1 px-4 py-3 bg-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 uppercase"
                maxlength="4"
              />
              <button
                class="px-6 py-3 bg-blue-600 text-white font-bold rounded-lg hover:bg-blue-500 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                :disabled="isJoining"
                @click="joinRoom"
              >
                加入
              </button>
            </div>
          </div>

          <div v-if="displayError" class="mt-4 text-red-400 text-sm text-center">
            {{ displayError }}
          </div>

          <div v-if="isJoining && !store.isInRoom" class="mt-4 text-center text-gray-400">
            <div class="animate-spin inline-block w-6 h-6 border-2 border-current border-t-transparent rounded-full"></div>
            <span class="ml-2">连接中...</span>
          </div>
        </div>

        <div v-else class="draft-root lobby-card rounded-2xl p-3 sm:p-4 shadow-xl">
          <header class="draft-header">
            <div class="room-meta">
              <div class="room-meta-label">房间码</div>
              <button
                type="button"
                class="room-code-btn"
                @click="copyRoomCode"
              >
                <span class="room-code">{{ store.roomCode }}</span>
                <span class="room-copy-hint">{{ copyFeedback ? '✓ 已复制' : '📋 复制' }}</span>
              </button>
            </div>

            <div class="draft-status" :class="allReadyToStart ? 'draft-status-ready' : ''">
              <div class="status-main">{{ lobbyHint }}</div>
              <div class="status-sub">蓝方 {{ blueCount }}/3 · 红方 {{ redCount }}/3 · 总人数 {{ roomPlayers.length }}/6</div>
            </div>

            <div class="draft-actions">
              <button
                v-if="isHost && !store.gameStarted && roomPlayers.length < 6"
                class="draft-btn draft-btn-bot"
                @click="addBot"
              >
                + 添加机器人
              </button>
              <button
                v-if="isHost && allReadyToStart && !store.gameStarted"
                class="draft-btn draft-btn-start"
                @click="startGame"
              >
                手动开始（备用）
              </button>
              <button
                v-if="isHost"
                class="draft-btn draft-btn-danger"
                @click="dissolveRoom"
              >
                解散房间
              </button>
            </div>
          </header>

          <div class="draft-board">
            <aside class="team-panel team-panel-blue">
              <div class="team-head">
                <div class="team-title">蓝方阵营</div>
                <button
                  v-if="!store.gameStarted"
                  class="team-join-btn"
                  :disabled="!canJoinCamp('Blue')"
                  @click="selectCamp('Blue')"
                >
                  {{ store.myCamp === 'Blue' ? '已在蓝方' : '加入蓝方' }}
                </button>
              </div>

              <div class="team-slots">
                <div v-for="(player, idx) in blueSlots" :key="`blue-${idx}`" class="team-slot" :class="player ? 'team-slot-filled' : 'team-slot-empty'">
                  <template v-if="player">
                    <div class="slot-main">
                      <div class="slot-name-wrap">
                        <span class="slot-name">{{ player.name }}</span>
                        <span class="slot-id">{{ player.id }}</span>
                      </div>
                      <div class="slot-tags">
                        <span v-if="player.id === store.myPlayerId" class="slot-tag me">你</span>
                        <span v-if="!player.is_bot && player.is_online === false" class="slot-tag offline">离线</span>
                        <span v-if="player.is_bot" class="slot-tag bot">BOT</span>
                        <span v-if="player.is_host" class="slot-tag host">房主</span>
                      </div>
                    </div>

                    <div class="slot-role-row">
                      <img
                        v-if="player.char_role"
                        class="slot-role-portrait"
                        :src="portraitSrc(player.char_role)"
                        :alt="getCharacterName(player.char_role)"
                        @error="onPortraitError(player.char_role, $event)"
                      />
                      <div v-else class="slot-role-placeholder">?</div>
                      <div class="slot-role-meta">
                        <div class="slot-role">{{ getCharacterName(player.char_role) }}</div>
                        <button
                          v-if="player.char_role"
                          type="button"
                          class="slot-skill-btn"
                          @click="openSkillModal(player.char_role)"
                        >
                          技能详情
                        </button>
                      </div>
                    </div>

                    <div v-if="player.is_bot && isHost && !store.gameStarted" class="slot-bot-controls">
                      <select
                        class="bot-role-select"
                        :value="player.char_role || ''"
                        @change="selectRoleFor(player.id, ($event.target as HTMLSelectElement).value)"
                      >
                        <option value="">选择角色</option>
                        <option
                          v-for="c in characterOptions"
                          :key="`blue-bot-${player.id}-${c.id}`"
                          :value="c.id"
                          :disabled="roomPlayers.some(p => p.id !== player.id && p.char_role === c.id)"
                        >
                          {{ c.name }}（{{ c.title }}）
                        </option>
                      </select>
                      <div class="bot-camp-row">
                        <button class="bot-camp-btn" @click="selectCampFor(player.id, 'Red')" :disabled="redCount >= 3">改去红方</button>
                        <button class="bot-remove-btn" @click="removeBot(player.id)">移除</button>
                      </div>
                    </div>
                  </template>
                  <template v-else>
                    <div class="slot-empty-text">空位</div>
                  </template>
                </div>
              </div>
            </aside>

            <section class="role-draft-center">
              <div class="unassigned-strip" v-if="unassignedPlayers.length > 0">
                <div class="unassigned-title">待分配阵营</div>
                <div class="unassigned-list">
                  <div v-for="player in unassignedPlayers" :key="`pending-${player.id}`" class="pending-chip">
                    <span>{{ player.name }} ({{ player.id }})</span>
                    <span v-if="player.id === store.myPlayerId" class="pending-me">你</span>
                    <template v-if="player.is_bot && isHost && !store.gameStarted">
                      <button class="pending-btn pending-btn-blue" :disabled="blueCount >= 3" @click="selectCampFor(player.id, 'Blue')">蓝方</button>
                      <button class="pending-btn pending-btn-red" :disabled="redCount >= 3" @click="selectCampFor(player.id, 'Red')">红方</button>
                    </template>
                  </div>
                </div>
              </div>

              <div class="role-pool-grid">
                <div
                  v-for="role in characterOptions"
                  :key="role.id"
                  class="role-card"
                  :class="{
                    'role-card-selected': store.myCharRole === role.id,
                    'role-card-taken': roleTakenBy(role.id),
                    'role-card-disabled': !canSelectRole(role.id)
                  }"
                  @click="pickRole(role.id)"
                >
                  <img
                    class="role-portrait"
                    :src="portraitSrc(role.id)"
                    :alt="role.name"
                    @error="onPortraitError(role.id, $event)"
                  />
                  <div class="role-overlay"></div>
                  <div class="role-meta">
                    <div class="role-name">{{ role.name }}</div>
                    <div class="role-sub">{{ role.title }}</div>
                  </div>
                  <button
                    type="button"
                    class="role-skill-btn"
                    @click.stop="openSkillModal(role.id)"
                  >
                    技能详情
                  </button>
                  <div v-if="roleTakenBy(role.id)" class="role-owner-chip" :class="roleTakenBy(role.id)?.id === store.myPlayerId ? 'mine' : 'other'">
                    {{ roleTakenBy(role.id)?.id === store.myPlayerId ? '已锁定' : `已被 ${roleTakenBy(role.id)?.name} 选择` }}
                  </div>
                </div>
              </div>
            </section>

            <aside class="team-panel team-panel-red">
              <div class="team-head">
                <div class="team-title">红方阵营</div>
                <button
                  v-if="!store.gameStarted"
                  class="team-join-btn"
                  :disabled="!canJoinCamp('Red')"
                  @click="selectCamp('Red')"
                >
                  {{ store.myCamp === 'Red' ? '已在红方' : '加入红方' }}
                </button>
              </div>

              <div class="team-slots">
                <div v-for="(player, idx) in redSlots" :key="`red-${idx}`" class="team-slot" :class="player ? 'team-slot-filled' : 'team-slot-empty'">
                  <template v-if="player">
                    <div class="slot-main">
                      <div class="slot-name-wrap">
                        <span class="slot-name">{{ player.name }}</span>
                        <span class="slot-id">{{ player.id }}</span>
                      </div>
                      <div class="slot-tags">
                        <span v-if="player.id === store.myPlayerId" class="slot-tag me">你</span>
                        <span v-if="!player.is_bot && player.is_online === false" class="slot-tag offline">离线</span>
                        <span v-if="player.is_bot" class="slot-tag bot">BOT</span>
                        <span v-if="player.is_host" class="slot-tag host">房主</span>
                      </div>
                    </div>

                    <div class="slot-role-row">
                      <img
                        v-if="player.char_role"
                        class="slot-role-portrait"
                        :src="portraitSrc(player.char_role)"
                        :alt="getCharacterName(player.char_role)"
                        @error="onPortraitError(player.char_role, $event)"
                      />
                      <div v-else class="slot-role-placeholder">?</div>
                      <div class="slot-role-meta">
                        <div class="slot-role">{{ getCharacterName(player.char_role) }}</div>
                        <button
                          v-if="player.char_role"
                          type="button"
                          class="slot-skill-btn"
                          @click="openSkillModal(player.char_role)"
                        >
                          技能详情
                        </button>
                      </div>
                    </div>

                    <div v-if="player.is_bot && isHost && !store.gameStarted" class="slot-bot-controls">
                      <select
                        class="bot-role-select"
                        :value="player.char_role || ''"
                        @change="selectRoleFor(player.id, ($event.target as HTMLSelectElement).value)"
                      >
                        <option value="">选择角色</option>
                        <option
                          v-for="c in characterOptions"
                          :key="`red-bot-${player.id}-${c.id}`"
                          :value="c.id"
                          :disabled="roomPlayers.some(p => p.id !== player.id && p.char_role === c.id)"
                        >
                          {{ c.name }}（{{ c.title }}）
                        </option>
                      </select>
                      <div class="bot-camp-row">
                        <button class="bot-camp-btn" @click="selectCampFor(player.id, 'Blue')" :disabled="blueCount >= 3">改去蓝方</button>
                        <button class="bot-remove-btn" @click="removeBot(player.id)">移除</button>
                      </div>
                    </div>
                  </template>
                  <template v-else>
                    <div class="slot-empty-text">空位</div>
                  </template>
                </div>
              </div>
            </aside>
          </div>

          <div class="draft-footer">
            <div class="footer-text">选角说明：左侧蓝方，右侧红方。点击角色卡或槽位中的“技能详情”可弹窗查看完整技能描述，选择完成后系统自动开局。</div>
          </div>
        </div>

        <SkillDetailModal
          :character="skillModalCharacter"
          :visible="!!skillModalCharacter"
          @close="closeSkillModal"
        />

        <div v-if="displayError" class="mt-3 text-red-400 text-sm text-center">
          {{ displayError }}
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.lobby-shell {
  min-height: 100dvh;
  max-height: 100dvh;
  background:
    radial-gradient(900px 330px at 14% -18%, rgba(96, 171, 177, 0.22), transparent 72%),
    radial-gradient(760px 320px at 86% -12%, rgba(220, 167, 96, 0.2), transparent 75%),
    linear-gradient(165deg, rgba(5, 13, 23, 0.96), rgba(7, 16, 28, 0.95)),
    url('/assets/ui/board-veil.svg') center/cover no-repeat;
}

.lobby-scroll {
  padding-top: max(24px, var(--safe-top));
  padding-bottom: calc(20px + var(--safe-bottom));
}

.lobby-container {
  max-width: 1320px;
}

.lobby-title-wrap h1 {
  letter-spacing: 0.08em;
  text-shadow: 0 2px 10px rgba(3, 8, 15, 0.54);
}

.lobby-title-wrap p {
  color: rgba(177, 196, 214, 0.86);
}

.lobby-card {
  background:
    linear-gradient(180deg, rgba(9, 23, 38, 0.9), rgba(7, 16, 29, 0.92)),
    url('/assets/ui/panel-ornament.svg') center/cover no-repeat;
  border: 1px solid rgba(124, 159, 180, 0.36);
  box-shadow:
    inset 0 1px 0 rgba(238, 248, 254, 0.08),
    0 16px 30px rgba(2, 8, 20, 0.46);
}

.draft-root {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.draft-header {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr) 240px;
  gap: 10px;
  align-items: stretch;
}

.room-meta,
.draft-status,
.draft-actions {
  border-radius: 12px;
  border: 1px solid rgba(132, 166, 186, 0.34);
  background: rgba(8, 20, 35, 0.74);
  padding: 10px 12px;
}

.room-meta {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
}

.room-meta-label {
  color: #9eb9cb;
  font-size: 12px;
  letter-spacing: 0.08em;
}

.room-code-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: rgba(7, 17, 29, 0.7);
  border: 1px solid rgba(156, 187, 205, 0.3);
  border-radius: 9px;
  padding: 6px 10px;
}

.room-code {
  color: #f7d896;
  font-weight: 800;
  font-size: 23px;
  letter-spacing: 0.2em;
}

.room-copy-hint {
  font-size: 12px;
  color: #99b5c8;
}

.draft-status {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.status-main {
  color: #dbe9f4;
  font-size: 14px;
  font-weight: 600;
}

.status-sub {
  margin-top: 4px;
  color: #9eb8cb;
  font-size: 12px;
}

.draft-status-ready {
  border-color: rgba(106, 185, 150, 0.42);
  background: linear-gradient(180deg, rgba(12, 42, 34, 0.76), rgba(8, 24, 23, 0.8));
}

.draft-actions {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 8px;
}

.draft-btn {
  border-radius: 9px;
  border: 1px solid transparent;
  height: 34px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.02em;
  transition: filter 0.16s ease;
}

.draft-btn:hover {
  filter: brightness(1.06);
}

.draft-btn-bot {
  background: linear-gradient(130deg, rgba(67, 94, 176, 0.88), rgba(45, 64, 120, 0.92));
  border-color: rgba(132, 158, 228, 0.52);
  color: #e6edff;
}

.draft-btn-start {
  background: linear-gradient(130deg, rgba(58, 154, 102, 0.9), rgba(37, 115, 79, 0.92));
  border-color: rgba(118, 211, 154, 0.52);
  color: #e8fff1;
}

.draft-btn-danger {
  background: linear-gradient(130deg, rgba(167, 73, 73, 0.9), rgba(124, 44, 44, 0.92));
  border-color: rgba(221, 128, 128, 0.5);
  color: #ffe8e8;
}

.draft-board {
  display: grid;
  grid-template-columns: 254px minmax(0, 1fr) 254px;
  gap: 10px;
  min-height: 0;
}

.team-panel {
  border-radius: 12px;
  border: 1px solid rgba(120, 158, 179, 0.32);
  background: rgba(7, 18, 31, 0.78);
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.team-panel-blue {
  box-shadow: inset 0 0 0 1px rgba(78, 136, 182, 0.17);
}

.team-panel-red {
  box-shadow: inset 0 0 0 1px rgba(182, 95, 83, 0.18);
}

.team-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.team-title {
  font-size: 15px;
  font-weight: 800;
  color: #dbe8f4;
  letter-spacing: 0.06em;
}

.team-join-btn {
  height: 28px;
  border-radius: 999px;
  padding: 0 10px;
  font-size: 12px;
  font-weight: 700;
  color: #d7e8f7;
  border: 1px solid rgba(120, 161, 185, 0.42);
  background: rgba(11, 29, 47, 0.84);
}

.team-join-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.team-slots {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.team-slot {
  border-radius: 10px;
  border: 1px solid rgba(120, 156, 176, 0.26);
  background: rgba(8, 20, 34, 0.72);
  padding: 8px;
  min-height: 72px;
}

.team-slot-filled {
  box-shadow: inset 0 1px 0 rgba(227, 241, 250, 0.06);
}

.slot-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.slot-name-wrap {
  display: flex;
  align-items: baseline;
  gap: 6px;
  min-width: 0;
}

.slot-name {
  color: #e8f2fa;
  font-size: 13px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.slot-id {
  color: #8fb0c4;
  font-size: 11px;
}

.slot-tags {
  display: flex;
  align-items: center;
  gap: 4px;
}

.slot-tag {
  font-size: 10px;
  font-weight: 700;
  border-radius: 999px;
  padding: 1px 6px;
  border: 1px solid transparent;
}

.slot-tag.me {
  color: #f7ddb0;
  border-color: rgba(224, 184, 121, 0.48);
  background: rgba(96, 70, 31, 0.48);
}

.slot-tag.bot {
  color: #bde6ff;
  border-color: rgba(100, 174, 216, 0.48);
  background: rgba(27, 74, 101, 0.52);
}

.slot-tag.offline {
  color: #ffd6d1;
  border-color: rgba(208, 118, 108, 0.5);
  background: rgba(96, 35, 33, 0.54);
}

.slot-tag.host {
  color: #f8dba4;
  border-color: rgba(212, 165, 83, 0.44);
  background: rgba(111, 78, 31, 0.48);
}

.slot-role {
  color: #b9d2e3;
  font-size: 12px;
}

.slot-role-row {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.slot-role-portrait {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  border: 1px solid rgba(129, 163, 184, 0.4);
  object-fit: cover;
  background: rgba(7, 17, 30, 0.86);
}

.slot-role-placeholder {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  border: 1px dashed rgba(112, 146, 166, 0.34);
  color: #6f899d;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  background: rgba(7, 16, 28, 0.66);
}

.slot-role-meta {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.slot-skill-btn {
  align-self: flex-start;
  height: 22px;
  border-radius: 999px;
  border: 1px solid rgba(132, 167, 186, 0.44);
  background: rgba(10, 29, 46, 0.76);
  color: #cbe1ef;
  font-size: 10px;
  font-weight: 700;
  padding: 0 7px;
}

.slot-empty-text {
  color: #67849a;
  font-size: 12px;
}

.slot-bot-controls {
  margin-top: 7px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.bot-role-select {
  width: 100%;
  height: 28px;
  border-radius: 8px;
  border: 1px solid rgba(124, 156, 176, 0.36);
  background: rgba(8, 20, 33, 0.86);
  color: #dfeef8;
  font-size: 12px;
  padding: 0 8px;
}

.bot-camp-row {
  display: flex;
  gap: 6px;
}

.bot-camp-btn,
.bot-remove-btn {
  flex: 1;
  height: 26px;
  border-radius: 7px;
  font-size: 11px;
  font-weight: 700;
  border: 1px solid rgba(122, 154, 172, 0.4);
  color: #d7e5f0;
  background: rgba(11, 30, 47, 0.8);
}

.bot-camp-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.bot-remove-btn {
  border-color: rgba(196, 109, 96, 0.44);
  color: #ffd7d1;
  background: rgba(88, 35, 33, 0.76);
}

.role-draft-center {
  border-radius: 12px;
  border: 1px solid rgba(120, 158, 179, 0.32);
  background:
    radial-gradient(620px 240px at 50% 12%, rgba(94, 156, 166, 0.14), transparent 68%),
    radial-gradient(560px 230px at 50% 96%, rgba(209, 163, 92, 0.12), transparent 72%),
    rgba(8, 18, 32, 0.78);
  padding: 10px;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.unassigned-strip {
  border-radius: 10px;
  border: 1px dashed rgba(130, 163, 183, 0.35);
  background: rgba(8, 21, 36, 0.62);
  padding: 8px;
}

.unassigned-title {
  color: #aecaDC;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.07em;
}

.unassigned-list {
  margin-top: 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.pending-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  border-radius: 999px;
  border: 1px solid rgba(122, 158, 179, 0.35);
  background: rgba(10, 28, 45, 0.78);
  color: #d9e7f1;
  font-size: 11px;
  padding: 3px 8px;
}

.pending-me {
  color: #f7d6a0;
}

.pending-btn {
  border-radius: 999px;
  border: 1px solid transparent;
  font-size: 10px;
  font-weight: 700;
  padding: 1px 6px;
}

.pending-btn-blue {
  color: #cfe6ff;
  border-color: rgba(99, 155, 212, 0.5);
  background: rgba(28, 72, 112, 0.8);
}

.pending-btn-red {
  color: #ffd8d1;
  border-color: rgba(211, 108, 96, 0.5);
  background: rgba(117, 38, 32, 0.8);
}

.role-pool-grid {
  flex: 1 1 auto;
  min-height: 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(132px, 1fr));
  gap: 8px;
  overflow: auto;
  padding-right: 2px;
}

.role-card {
  position: relative;
  overflow: hidden;
  border-radius: 10px;
  border: 1px solid rgba(128, 163, 184, 0.33);
  background: rgba(8, 20, 33, 0.8);
  min-height: 166px;
  text-align: left;
  transition: transform 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
}

.role-card:hover {
  transform: translateY(-1px);
  border-color: rgba(222, 188, 119, 0.58);
  box-shadow: 0 10px 16px rgba(2, 8, 15, 0.44);
}

.role-card:disabled {
  cursor: not-allowed;
}

.role-card-disabled {
  cursor: not-allowed;
}

.role-card-selected {
  border-color: rgba(237, 199, 123, 0.65);
  box-shadow:
    inset 0 0 0 1px rgba(239, 207, 147, 0.35),
    0 10px 20px rgba(58, 38, 14, 0.35);
}

.role-card-taken {
  filter: grayscale(0.68) saturate(0.52) brightness(0.78);
}

.role-portrait {
  width: 100%;
  height: 100%;
  position: absolute;
  inset: 0;
  object-fit: cover;
}

.role-overlay {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgba(6, 16, 30, 0.08), rgba(4, 9, 16, 0.92) 84%);
}

.role-meta {
  position: absolute;
  left: 8px;
  right: 8px;
  bottom: 7px;
  z-index: 1;
}

.role-name {
  color: #f4e2bb;
  font-size: 14px;
  font-weight: 800;
}

.role-sub {
  color: #b5ccdb;
  font-size: 11px;
  margin-top: 1px;
}

.role-owner-chip {
  position: absolute;
  left: 8px;
  top: 8px;
  z-index: 1;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
}

.role-owner-chip.mine {
  color: #ffe2ad;
  border: 1px solid rgba(225, 182, 105, 0.52);
  background: rgba(100, 70, 30, 0.72);
}

.role-owner-chip.other {
  color: #d5deea;
  border: 1px solid rgba(130, 145, 164, 0.5);
  background: rgba(44, 52, 67, 0.74);
}

.role-skill-btn {
  position: absolute;
  right: 8px;
  top: 8px;
  z-index: 1;
  height: 22px;
  border-radius: 999px;
  border: 1px solid rgba(140, 175, 194, 0.45);
  background: rgba(8, 25, 40, 0.78);
  color: #d9ebf8;
  font-size: 10px;
  font-weight: 700;
  padding: 0 8px;
}

.draft-footer {
  border-radius: 10px;
  border: 1px solid rgba(121, 157, 178, 0.3);
  background: rgba(8, 20, 35, 0.64);
  padding: 8px 10px;
}

.footer-text {
  color: #a8c2d4;
  font-size: 12px;
}

@media (max-width: 1200px) {
  .draft-header {
    grid-template-columns: 1fr;
  }

  .draft-board {
    grid-template-columns: 1fr;
  }

  .team-panel {
    order: 2;
  }

  .role-draft-center {
    order: 1;
  }

}

@media (max-width: 640px) {
  .lobby-scroll {
    padding-top: max(16px, var(--safe-top));
    padding-bottom: calc(14px + var(--safe-bottom));
  }

  .room-code {
    font-size: 19px;
  }

  .role-pool-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .role-card {
    min-height: 150px;
  }
}
</style>
