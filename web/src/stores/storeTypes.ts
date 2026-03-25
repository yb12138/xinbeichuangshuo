export interface GameEndSnapshot {
  message: string
  triggerType: 'cups' | 'morale' | 'unknown'
  finalRedMorale: number
  finalBlueMorale: number
  finalRedCups: number
  finalBlueCups: number
  triggerCamp?: 'Red' | 'Blue'
  triggerDelta?: number
  triggerSource?: string
}

export interface SkillModalAnchor {
  x: number
  y: number
  width: number
  height: number
}
