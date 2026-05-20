import { describe, expect, it } from 'vitest'
import { promptRendererUsesInlineSurface, selectPromptRenderer } from '../rendererRegistry'

describe('rendererRegistry', () => {
  it('prefers dedicated overlays before inline surfaces', () => {
    expect(selectPromptRenderer({
      visible: true,
      isActionHub: false,
      isExtractPrompt: false,
      extractOptionCount: 0,
      isSkillChoicePrompt: true,
      isMultiSkillNameChoiceMode: true,
      skillPromptButtonCount: 2,
      skillBranchCount: 2,
      showTargetSelectionHintRow: true,
      hasCounterOrDefend: true,
      responseOptionCount: 2,
      inlinePrimaryButtonCount: 3,
      promptNeedsCardConfirm: true,
      canCancelPrompt: true,
      showDecisionOverlay: true,
      isDirectionPrompt: false,
      directionOptionCount: 0,
      isFraudElementCardPickerPrompt: false,
      fraudElementOptionCount: 0,
      isSaintHealAllocatePrompt: false,
      isRuneReforgeAllocatePrompt: false,
    })).toBe('skill_overlay')
  })

  it('returns inline surface for normal prompt content', () => {
    expect(selectPromptRenderer({
      visible: true,
      isActionHub: false,
      isExtractPrompt: false,
      extractOptionCount: 0,
      isSkillChoicePrompt: false,
      isMultiSkillNameChoiceMode: false,
      skillPromptButtonCount: 0,
      skillBranchCount: 0,
      showTargetSelectionHintRow: false,
      hasCounterOrDefend: false,
      responseOptionCount: 0,
      inlinePrimaryButtonCount: 2,
      promptNeedsCardConfirm: false,
      canCancelPrompt: false,
      showDecisionOverlay: false,
      isDirectionPrompt: false,
      directionOptionCount: 0,
      isFraudElementCardPickerPrompt: false,
      fraudElementOptionCount: 0,
      isSaintHealAllocatePrompt: false,
      isRuneReforgeAllocatePrompt: false,
    })).toBe('inline_buttons')
  })

  it('treats overlay and non-inline renderers as non-inline surfaces', () => {
    expect(promptRendererUsesInlineSurface('decision_overlay')).toBe(false)
    expect(promptRendererUsesInlineSurface('response')).toBe(true)
  })
})
