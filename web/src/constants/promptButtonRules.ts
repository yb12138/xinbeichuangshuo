export type PromptImageButtonKind = 'take' | 'counter' | 'defend' | 'cancel' | 'confirm' | 'card' | 'action'

export function promptImageButtonKindByOption(option: { buttonLabel: string }): PromptImageButtonKind {
  const buttonLabel = String(option.buttonLabel || '').trim()
  if (buttonLabel === '打出卡牌' || buttonLabel === '选择卡牌') return 'card'
  if (buttonLabel === '发动' || buttonLabel === '确认' || buttonLabel === '确定' || buttonLabel === '是') return 'confirm'
  if (buttonLabel === '取消' || buttonLabel === '跳过' || buttonLabel === '不发动') return 'cancel'
  return 'action'
}
