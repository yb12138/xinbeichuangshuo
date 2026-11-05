import type { useInterruptStore } from '../stores/interrupt.store'

/**
 * 后端 timeline `error` 只带文案、没有错误码（ProtocolError 才有 code）。
 * 需要按文案兜底识别的关键字集中在这里，别在各处散落 includes。
 */
export const ERROR_MESSAGE_KEYWORDS = {
  /** engine/action_submission_runtime.go：UseSkill 返回错误时的前缀 */
  skillFailed: '技能发动失败',
  /** 各角色中断处理里卡牌索引越界 */
  invalidCardIndex: '无效的卡牌索引',
} as const

type InterruptStore = ReturnType<typeof useInterruptStore>

/**
 * 服务器拒绝了一次提交：
 * 1. 有提交快照时恢复提交前的选择（技能、目标、弃牌、应战牌），玩家改一下就能重试；
 * 2. 应战目标一律清掉，否则恢复回来的目标会立刻触发再次提交，玩家也没法换目标；
 * 3. 没有快照（不是这次提交引起的错误）时，才按文案关键字兜底清理残留状态。
 * 不论哪种情况，"提交中"状态都会被解除。
 */
export function recoverFromRejectedSubmission(interruptStore: InterruptStore, message: string, code = '') {
  const restored = interruptStore.restoreSubmission()
  if (interruptStore.promptCounterTarget || interruptStore.counterFlowStage === 'target') {
    interruptStore.setPromptCounterTarget('')
  }
  if (restored) return
  if (code) return
  if (message.includes(ERROR_MESSAGE_KEYWORDS.skillFailed)) {
    interruptStore.clearSkillMode()
    interruptStore.clearActionMode()
  }
  if (message.includes(ERROR_MESSAGE_KEYWORDS.invalidCardIndex)) {
    interruptStore.clearActionMode()
  }
}
