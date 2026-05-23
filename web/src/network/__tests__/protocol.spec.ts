import { describe, expect, it } from 'vitest'
import { PLAYER_ACTION_TYPES, ROOM_ACTION_TYPES, WS_COMMANDS } from '../../types/generated'
import { normalizeWsMessage } from '../protocol'

describe('protocol boundary helpers', () => {
  it('exposes generated command and action literal sets used by the network layer', () => {
    expect(WS_COMMANDS).toContain('ProtocolError')
    expect(WS_COMMANDS).toContain('RoomAction')
    expect(ROOM_ACTION_TYPES).toEqual([
      'dissolve_room',
      'add_bot',
      'remove_bot',
      'takeover_player',
      'change_camp',
      'change_role',
      'start',
    ])
    expect(PLAYER_ACTION_TYPES).toContain('Select')
    expect(PLAYER_ACTION_TYPES).toContain('Respond')
  })

  it('normalizes known envelopes and labels unknown commands at the websocket boundary', () => {
    expect(normalizeWsMessage({
      Cmd: 'ProtocolError',
      Data: { code: 'unknown_cmd', message: '未知命令' },
    })).toEqual({
      Cmd: 'ProtocolError',
      Data: { code: 'unknown_cmd', message: '未知命令' },
    })

    expect(normalizeWsMessage({
      Cmd: 'FutureCommand',
      Data: { ok: true },
    })).toEqual({
      Known: false,
      Cmd: 'FutureCommand',
      Data: { ok: true },
    })

    expect(normalizeWsMessage({ Data: {} })).toBeNull()
  })
})
