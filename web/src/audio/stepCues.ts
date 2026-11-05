import type { AtomicStep } from '../stores/stageQueue.store'
import type { SfxCue } from './sfx'

/**
 * 演出步骤 → 音效。只在步骤真正播放时调用（stageQueue.tick），
 * 快进结算和回放（holdMs = 0）都不出声，所以音效与画面始终同步。
 *
 * 刻意只保留出牌（攻击 / 法术连线出现）。命中、格挡、士气、摸牌、回合切换等都不出声，
 * 避免一条战斗链里声音叠成一片。
 */
export function cueForStep(step: AtomicStep): SfxCue | null {
  if (step.holdMs === 0) return null
  switch (step.type) {
    case 'link':
      if (step.color === 'attack') return 'attack'
      if (step.color === 'magic' || step.color === 'heal') return 'magic'
      return null
    default:
      return null
  }
}
