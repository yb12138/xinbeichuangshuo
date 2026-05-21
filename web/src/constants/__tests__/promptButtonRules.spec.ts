import { describe, expect, it } from 'vitest'
import {
  promptImageButtonKindByOption,
} from '../promptButtonRules'

describe('promptButtonRules', () => {
  it('maps structured prompt semantics to stable image kinds', () => {
    expect(promptImageButtonKindByOption({ id: 'take', presentationKind: 'response' })).toBe('take')
    expect(promptImageButtonKindByOption({ id: 'counter', presentationKind: 'response' })).toBe('counter')
    expect(promptImageButtonKindByOption({ id: '0', presentationKind: 'card_picker' })).toBe('card')
    expect(promptImageButtonKindByOption({ id: '1', presentationKind: 'numeric' })).toBe('confirm')
    expect(promptImageButtonKindByOption({ id: 'decline', cancelPolicy: 'decline' })).toBe('cancel')
    expect(promptImageButtonKindByOption({ id: 'decline', presentationKind: 'branch_select', cancelPolicy: 'decline' })).toBe('action')
  })
})
