export type PromptRendererKey =
  | 'none'
  | 'extract'
  | 'skill_inline'
  | 'skill_overlay'
  | 'target_picker'
  | 'response'
  | 'inline_buttons'
  | 'card_picker'
  | 'cancel_only'
  | 'decision_overlay'
  | 'direction_overlay'
  | 'fraud_element_overlay'
  | 'allocation_heal_overlay'
  | 'allocation_rune_overlay'

export type PromptRendererSelectionContext = {
  visible: boolean
  isActionHub: boolean
  isExtractPrompt: boolean
  extractOptionCount: number
  isSkillChoicePrompt: boolean
  isMultiSkillNameChoiceMode: boolean
  skillPromptButtonCount: number
  skillBranchCount: number
  showTargetSelectionHintRow: boolean
  hasCounterOrDefend: boolean
  responseOptionCount: number
  inlinePrimaryButtonCount: number
  promptNeedsCardConfirm: boolean
  canCancelPrompt: boolean
  showDecisionOverlay: boolean
  isDirectionPrompt: boolean
  directionOptionCount: number
  isFraudElementCardPickerPrompt: boolean
  fraudElementOptionCount: number
  isSaintHealAllocatePrompt: boolean
  isRuneReforgeAllocatePrompt: boolean
}

export function selectPromptRenderer(ctx: PromptRendererSelectionContext): PromptRendererKey {
  if (!ctx.visible || ctx.isActionHub) return 'none'

  if (ctx.isSaintHealAllocatePrompt) return 'allocation_heal_overlay'
  if (ctx.isRuneReforgeAllocatePrompt) return 'allocation_rune_overlay'
  if (ctx.isDirectionPrompt && ctx.directionOptionCount > 0) return 'direction_overlay'
  if (ctx.isFraudElementCardPickerPrompt && ctx.fraudElementOptionCount > 0) return 'fraud_element_overlay'
  if (ctx.isMultiSkillNameChoiceMode && ctx.skillBranchCount > 0) return 'skill_overlay'
  if (ctx.showDecisionOverlay) return 'decision_overlay'

  if (ctx.isExtractPrompt && ctx.extractOptionCount > 0) return 'extract'
  if (ctx.isSkillChoicePrompt && ctx.skillPromptButtonCount > 0) return 'skill_inline'
  if (ctx.showTargetSelectionHintRow) return 'target_picker'
  if (ctx.hasCounterOrDefend && ctx.responseOptionCount > 0) return 'response'
  if (ctx.inlinePrimaryButtonCount > 0) return 'inline_buttons'
  if (ctx.promptNeedsCardConfirm) return 'card_picker'
  if (ctx.canCancelPrompt) return 'cancel_only'
  return 'none'
}

export function promptRendererUsesInlineSurface(renderer: PromptRendererKey): boolean {
  return renderer === 'extract' ||
    renderer === 'skill_inline' ||
    renderer === 'target_picker' ||
    renderer === 'response' ||
    renderer === 'inline_buttons' ||
    renderer === 'card_picker' ||
    renderer === 'cancel_only'
}
