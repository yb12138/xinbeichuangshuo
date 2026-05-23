export interface AutoJoinRoomOptions {
  search: string
  isInRoom: () => boolean
  connect: (roomCode: string, playerName: string) => void
}

export function autoJoinRoomFromUrl(options: AutoJoinRoomOptions): boolean {
  if (options.isInRoom()) return false

  const query = new URLSearchParams(options.search)
  const roomCode = (query.get('room') || '').trim().toUpperCase()
  const playerName = (query.get('name') || '').trim()

  if (!roomCode || !playerName) return false

  options.connect(roomCode, playerName)
  return true
}
