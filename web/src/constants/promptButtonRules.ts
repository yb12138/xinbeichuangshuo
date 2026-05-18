export type ResponseOptionKind = 'take' | 'counter' | 'defend' | null
export type PromptImageButtonKind = 'take' | 'counter' | 'defend' | 'cancel' | 'confirm' | 'card' | 'action'

export const PLAIN_NO_HINT_BUTTONS = new Set(['发动', '确认', '确定', '是', '取消', '不弃牌', '应战', '防御', '命中', '顺序', '反向'])

export function responseOptionKind(option: { id?: string; label?: string; button_label?: string }): ResponseOptionKind {
  const id = String(option.id || '').trim().toLowerCase()
  if (id === 'take' || id === 'take_damage') return 'take'
  if (id === 'defend') return 'defend'
  if (id === 'counter') return 'counter'
  return null
}

export function isActivationCostText(raw: string): boolean {
  const text = String(raw || '').replace(/\s+/g, '')
  if (!text) return false
  return text.includes('发动') && /(弃|弃置|移除|消耗|支付)/.test(text)
}

export function isDeclineLabel(label: string): boolean {
  const text = String(label || '').trim()
  if (!text) return false
  const compact = text.replace(/\s+/g, '')
  const lower = compact.toLowerCase()
  if (compact.includes('不发动') || compact.includes('无法行动') || compact.includes('不弃牌') || compact.includes('不弃置')) return true
  if (compact === '放弃' || compact.startsWith('放弃本次') || compact.startsWith('放弃发动')) return true
  if (compact === '跳过' || compact.startsWith('跳过本次') || compact.startsWith('跳过此')) return true
  if (compact.startsWith('拒绝')) return true
  if (compact === '取消' || compact.startsWith('取消并') || compact.startsWith('取消本次') || compact.startsWith('取消行动')) return true
  if (lower === 'cancel' || lower === 'pass' || lower === 'skip' || lower === 'refuse') return true
  return false
}

export function isConfirmLikeLabel(label: string): boolean {
  const text = String(label || '').trim().replace(/\s+/g, '')
  if (!text) return false
  if (text === '是' || text === '发动' || text === '确认' || text === '确定') return true
  if (text.startsWith('发动') || text.startsWith('确认') || text.startsWith('确定')) return true
  return false
}

export function isCardSelectionLikeText(text: string): boolean {
  const normalized = String(text || '').trim()
  if (!normalized) return false
  if (/^\d+\s*[:：]/.test(normalized)) return true
  if (/第\d+张\s*[:：]/.test(normalized)) return true
  if (/^茧\[\d+\]\s*[:：]/.test(normalized)) return true
  if (/^移除茧\[\d+\]\s*[:：]/.test(normalized)) return true
  return false
}

export function promptImageButtonKindByOption(option: { id?: string; label?: string; buttonLabel?: string; hint?: string }): PromptImageButtonKind {
  const id = String(option.id || '').trim().toLowerCase()
  const label = String(option.label || '').trim()
  const buttonLabel = String(option.buttonLabel || '').trim()
  const responseKind = responseOptionKind({ id, label, button_label: buttonLabel })
  if (responseKind) return responseKind
  if (buttonLabel === '打出卡牌' || buttonLabel === '选择卡牌') return 'card'
  if (buttonLabel === '发动' || buttonLabel === '确认' || buttonLabel === '确定' || buttonLabel === '是') return 'confirm'
  if (id === 'confirm' || id === 'yes') return 'confirm'
  if (id === 'skip' || id === 'cancel' || id === 'no' || id === 'pass' || id === 'cannot_act') return 'cancel'
  if (buttonLabel === '取消' || buttonLabel === '跳过' || buttonLabel === '不发动') return 'cancel'
  return 'action'
}
