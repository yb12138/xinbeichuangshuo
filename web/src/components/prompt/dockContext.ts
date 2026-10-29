import type { InjectionKey } from 'vue'

/**
 * True when prompts render inside the floating DecisionDock: overlay renderers then render in
 * place (no Teleport to <body>, no full-screen fixed root) and drop their own context line,
 * because the dock header already says who / which skill / why.
 */
export const DECISION_DOCKED: InjectionKey<boolean> = Symbol('decisionDocked')
