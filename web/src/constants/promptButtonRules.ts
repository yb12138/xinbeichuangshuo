export type PromptImageButtonKind = 'take' | 'counter' | 'defend' | 'cancel' | 'confirm' | 'card' | 'action'

export function promptImageButtonKindByOption(option: {
  id: string
  presentationKind?: string
  cancelPolicy?: string
  hasDecline?: boolean
  declineIndex?: number
  optionIndex?: number
}): PromptImageButtonKind {
  const id = String(option.id || '').trim().toLowerCase()
  if (option.presentationKind === 'response') {
    if (id === 'take' || id === 'take_damage') return 'take'
    if (id === 'counter') return 'counter'
    if (id === 'defend') return 'defend'
  }
  if (option.hasDecline && option.optionIndex === option.declineIndex) return 'cancel'
  if (option.cancelPolicy && option.cancelPolicy !== 'deny' && (id === 'cancel' || id === 'skip' || id === 'decline' || id === 'pass' || id === '-1')) return 'cancel'
  if (option.presentationKind === 'card_picker') return 'card'
  if (option.presentationKind === 'numeric') return 'confirm'
  if (id === 'confirm' || id === 'yes') return 'confirm'
  return 'action'
}
