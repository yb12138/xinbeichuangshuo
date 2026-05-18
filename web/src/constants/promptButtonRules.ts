export type ResponseOptionKind = 'take' | 'counter' | 'defend' | null
export type PromptImageButtonKind = 'take' | 'counter' | 'defend' | 'cancel' | 'confirm' | 'card' | 'action'

export const PROMPT_OPTION_BUTTON_LABELS: Record<string, string> = {
  confirm: '发动',
  yes: '发动',
  no: '取消',
  cancel: '取消',
  skip: '取消',
  take: '命中',
  counter: '应战',
  defend: '防御',
  normal: '顺时针',
  reverse: '逆时针',
  refuse: '不弃牌',
  cannot_act: '取消',
  pass: '取消',
}

export const PLAIN_NO_HINT_BUTTONS = new Set(['发动', '确认', '确定', '是', '取消', '不弃牌', '应战', '防御', '命中', '顺序', '反向'])

export function responseOptionKind(option: { id?: string; label?: string; button_label?: string }): ResponseOptionKind {
  const id = String(option.id || '').trim().toLowerCase()
  // 魔弹掌控方向选择不是战斗应答（take/defend/counter），避免被“传递”关键词误判为应战。
  if (id === 'normal' || id === 'reverse') return null
  const label = String(option.label || '').trim()
  const buttonLabel = String(option.button_label || '').trim()
  const text = `${label} ${buttonLabel}`.toLowerCase()
  if (id === 'take' || id === 'take_damage' || text.includes('承受') || text.includes('命中')) return 'take'
  if (id === 'defend' || text.includes('防御')) return 'defend'
  if (id === 'counter' || text.includes('应战')) return 'counter'
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
  const hint = String(option.hint || '').trim()
  const combinedText = `${label} ${buttonLabel}`
  const hasExplicitResponseText =
    combinedText.includes('命中') ||
    combinedText.includes('承受') ||
    combinedText.includes('防御') ||
    combinedText.includes('应战') ||
    combinedText.includes('传递')
  if (isCardSelectionLikeText(label) || isCardSelectionLikeText(buttonLabel)) {
    return 'card'
  }
  if ((isConfirmLikeLabel(buttonLabel) || isConfirmLikeLabel(label)) && !hasExplicitResponseText) {
    return 'confirm'
  }
  const responseKind = responseOptionKind({ id, label, button_label: buttonLabel })
  if (responseKind) return responseKind
  if (
    (isActivationCostText(hint) || isActivationCostText(label) || isActivationCostText(buttonLabel)) &&
    !isDeclineLabel(hint) &&
    !isDeclineLabel(label) &&
    !isDeclineLabel(buttonLabel)
  ) {
    return 'confirm'
  }
  if (id === 'confirm' || id === 'yes') return 'confirm'
  if (id === 'skip' || id === 'cancel' || id === 'no' || id === 'pass' || id === 'cannot_act') return 'cancel'
  if (buttonLabel === '取消' || isDeclineLabel(buttonLabel) || isDeclineLabel(label)) return 'cancel'
  return 'action'
}

export function normalizeButtonLabel(
  rawLabel: string,
  optionId: string,
  optionLabel: string,
  responseKind: ResponseOptionKind,
): string {
  const text = String(rawLabel || '').trim()
  const lowerId = String(optionId || '').trim().toLowerCase()
  if (responseKind === 'take' || lowerId === 'take' || lowerId === 'take_damage' || text.includes('承受') || text.includes('命中')) {
    return '命中'
  }
  if (responseKind === 'counter' || lowerId === 'counter' || text.includes('应战')) {
    return '应战'
  }
  if (responseKind === 'defend' || lowerId === 'defend' || text.includes('防御')) {
    return '防御'
  }
  if (
    lowerId === 'cancel' ||
    lowerId === 'skip' ||
    lowerId === 'refuse' ||
    lowerId === 'no' ||
    lowerId === 'pass' ||
    lowerId === 'cannot_act' ||
    isDeclineLabel(text) ||
    isDeclineLabel(optionLabel)
  ) {
    return '取消'
  }
  return text
}
