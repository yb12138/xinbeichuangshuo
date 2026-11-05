import type { AtomicStep } from '../stores/stageQueue.store'
import type { SfxCue } from './sfx'

/**
 * 演出步骤 → 音效。只在步骤真正播放时调用（stageQueue.tick），
 * 快进结算和回放（holdMs = 0）都不出声，所以音效与画面始终同步。
 */
export function cueForStep(step: AtomicStep): SfxCue | null {
  if (step.holdMs === 0) return null
  switch (step.type) {
    case 'link':
      if (step.color === 'heal' || step.color === 'magic') return 'magic'
      if (step.color === 'skill') return 'skill'
      if (step.color === 'grey') return null
      return 'attack'
    case 'link-end':
      if (step.style === 'hit') return 'hit'
      if (step.style === 'blocked') return 'block'
      if (step.style === 'miss') return 'miss'
      // 治疗声由随后的 value-chip 给出，避免同一次治疗响两遍
      return null
    case 'skill-bubble':
      return step.opening === 'trigger' ? 'trigger' : 'skill'
    case 'value-chip':
      // 伤害声已经由 link-end 的命中给出；这里只负责治疗
      return step.color === 'heal' ? 'heal' : null
    case 'stone-flight':
      return 'gem'
    case 'deck-to-seat':
      return 'draw'
    case 'turn-banner':
      return 'turn'
    default:
      return null
  }
}
