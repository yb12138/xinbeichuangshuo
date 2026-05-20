import type {
  AvailableSkill as GeneratedAvailableSkill,
  Buff as GeneratedBuff,
  Camp as GeneratedCamp,
  Card as GeneratedCard,
  CardType as GeneratedCardType,
  CharacterView as GeneratedCharacterView,
  Element as GeneratedElement,
  FieldCard as GeneratedFieldCard,
  GameplayEvent as GeneratedGameplayEvent,
  GameStateUpdate as GeneratedGameStateUpdate,
  PlayerInfo as GeneratedPlayerInfo,
  PlayerView as GeneratedPlayerView,
  PresentationKind as GeneratedPresentationKind,
  PromptDTO as GeneratedPromptDTO,
  PromptInteractionDTO as GeneratedPromptInteractionDTO,
  PromptOptionDTO as GeneratedPromptOptionDTO,
  PromptPresentation as GeneratedPromptPresentation,
  RoomEvent as GeneratedRoomEvent,
  SkillView as GeneratedSkillView,
  WSMessage as GeneratedWSMessage,
} from './generated'

export type Element = GeneratedElement
export type CardType = GeneratedCardType
export type Camp = GeneratedCamp
export type Card = GeneratedCard
export type AvailableSkill = GeneratedAvailableSkill
export type Buff = GeneratedBuff
export type FieldCard = GeneratedFieldCard
export type SkillView = GeneratedSkillView
export type CharacterView = GeneratedCharacterView
export type PlayerInfo = GeneratedPlayerInfo
export type PresentationKind = GeneratedPresentationKind
export type PromptPresentation = GeneratedPromptPresentation
export type PromptInteraction = GeneratedPromptInteractionDTO
export type PromptOption = GeneratedPromptOptionDTO
export type RoomEvent = GeneratedRoomEvent
export type WSMessage = GeneratedWSMessage

export type PlayerView = GeneratedPlayerView

// The frontend snapshot keeps players keyed by id after normalizing SyncState.
export type GameStateUpdate = Omit<GeneratedGameStateUpdate, 'players'> & {
  players: Record<string, PlayerView>
}

export type PromptType = 'choose_card' | 'choose_cards' | 'choose_target' | 'choose_skill' | 'confirm' | 'choose_extract'

export type Prompt = Omit<GeneratedPromptDTO, 'type' | 'options' | 'special_options'> & {
  type: PromptType
  options: PromptOption[]
  special_options?: PromptOption[]
}

export type PlayerActionType =
  | 'Start' | 'Quit' | 'Pass' | 'Help'
  | 'Attack' | 'Magic' | 'Buy' | 'Synthesize' | 'Extract' | 'Skill'
  | 'Confirm' | 'Cancel' | 'Select' | 'Respond'
  | 'CannotAct'
  | 'Cheat'

// Frontend-private intent model. The wire shape is generated as ClientActionRequest.
export interface PlayerAction {
  player_id: string
  type: PlayerActionType
  target_id?: string
  target_ids?: string[]
  card_id?: string
  card_ids?: string[]
  skill_id?: string
  selections?: number[]
  extra_args?: string[]
}

export type GameEvent =
  | Exclude<GeneratedGameplayEvent, { event_type: 'state_update' | 'prompt' }>
  | { event_type: 'state_update'; state: GameStateUpdate }
  | { event_type: 'prompt'; prompt: Prompt }
  | { event_type: 'waiting'; player_id: string }

export const ELEMENT_COLORS: Record<Element, string> = {
  Water: 'element-water',
  Fire: 'element-fire',
  Earth: 'element-earth',
  Wind: 'element-wind',
  Thunder: 'element-thunder',
  Light: 'element-light',
  Dark: 'element-dark',
}

export const ELEMENT_NAMES: Record<Element, string> = {
  Water: '水',
  Fire: '火',
  Earth: '地',
  Wind: '风',
  Thunder: '雷',
  Light: '光',
  Dark: '暗',
}
