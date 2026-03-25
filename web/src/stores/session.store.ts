import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { PlayerInfo } from '../types/game'

export const useSessionStore = defineStore('session', () => {
  const roomCode = ref('')
  const myPlayerId = ref('')
  const myName = ref('')
  const myCamp = ref('')
  const myCharRole = ref('')
  const reconnectToken = ref('')
  const roomPlayers = ref<PlayerInfo[]>([])
  const isInRoom = ref(false)
  const gameStarted = ref(false)
  const isConnected = ref(false)

  function setMyName(name: string) {
    myName.value = name
  }

  function setRoomInfo(code: string, playerId: string, camp: string, charRole: string) {
    roomCode.value = code
    myPlayerId.value = playerId
    myCamp.value = camp
    myCharRole.value = charRole
    isInRoom.value = true
  }

  function setSeat(camp: string, charRole: string) {
    if (camp) myCamp.value = camp
    if (charRole) myCharRole.value = charRole
  }

  function setReconnectToken(token: string) {
    reconnectToken.value = token || ''
  }

  function updateRoomPlayers(playerList: PlayerInfo[], myID?: string) {
    roomPlayers.value = playerList
    if (!myID) return
    const me = playerList.find(player => player.id === myID)
    if (!me) return
    if (me.camp) myCamp.value = me.camp
    if (me.char_role) myCharRole.value = me.char_role
  }

  function setGameStarted(started = true) {
    gameStarted.value = started
  }

  function setConnected(connected: boolean) {
    isConnected.value = connected
  }

  function reset() {
    roomCode.value = ''
    myPlayerId.value = ''
    myCamp.value = ''
    myCharRole.value = ''
    reconnectToken.value = ''
    roomPlayers.value = []
    isInRoom.value = false
    gameStarted.value = false
    isConnected.value = false
  }

  return {
    roomCode,
    myPlayerId,
    myName,
    myCamp,
    myCharRole,
    reconnectToken,
    roomPlayers,
    isInRoom,
    gameStarted,
    isConnected,
    setMyName,
    setRoomInfo,
    setSeat,
    setReconnectToken,
    updateRoomPlayers,
    setGameStarted,
    setConnected,
    reset,
  }
})
