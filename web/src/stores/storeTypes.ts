export interface GameEndSnapshot {
  message: string
  endReasonKind: 'cups' | 'morale' | 'unknown'
  finalRedMorale: number
  finalBlueMorale: number
  finalRedCups: number
  finalBlueCups: number
  endMoraleCamp?: 'Red' | 'Blue'
  endMoraleLoss?: number
  endCauseSource?: string
}

export interface SkillModalAnchor {
  x: number
  y: number
  width: number
  height: number
}
