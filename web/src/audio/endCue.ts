import type { SfxCue } from './sfx'

/**
 * 按结束消息和我的阵营选结尾音：我方胜利 → victory，我方失败 → defeat，
 * 观战或消息里看不出胜方时 → game-end。
 */
export function endCueFor(message: string, myCamp: string | undefined): SfxCue {
  const winner = message.includes('红方胜利') ? 'Red' : message.includes('蓝方胜利') ? 'Blue' : ''
  if (!winner || (myCamp !== 'Red' && myCamp !== 'Blue')) return 'game-end'
  return winner === myCamp ? 'victory' : 'defeat'
}
