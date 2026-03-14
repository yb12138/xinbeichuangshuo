// 元素类型
export type Element = 'Water' | 'Fire' | 'Earth' | 'Wind' | 'Thunder' | 'Light' | 'Dark'

// 卡牌类型
export type CardType = 'Attack' | 'Magic'

// 阵营
export type Camp = 'Red' | 'Blue'

// 卡牌
export interface Card {
  id: string
  name: string
  type: CardType
  element: Element
  faction?: string
  damage: number
  description: string
  /** 独有技相关：角色/技能名，卡牌上可展示 */
  exclusive_char1?: string
  exclusive_char2?: string
  exclusive_skill1?: string
  exclusive_skill2?: string
}

// 当前可发动的主动技能（与后端 AvailableSkill 一致）
export interface AvailableSkill {
  id: string
  title: string
  description: string
  min_targets: number
  max_targets: number
  target_type: number  // 0=None, 1=Self, 2=Enemy, 3=Ally, 4=AllySelf, 5=Any, 6=Specific
  cost_gem: number
  cost_crystal: number
  cost_discards: number
  discard_type?: CardType
  discard_element?: string  // 弃牌元素要求（如 "Water"）
  require_exclusive?: boolean  // 是否必须使用独有牌（卡牌下标了技能名）
  place_card?: boolean
  place_effect?: string
}

// Buff/状态效果
export interface Buff {
  id: string
  name: string
  duration: number
  value: number
  source_id: string
}

// 场上卡牌
export interface FieldCard {
  card: Card
  owner_id: string
  source_id: string
  mode: 'Effect' | 'Cover'
  effect: string
  trigger: string
  locked: boolean
  duration: number
}

// 玩家视图（前端接收的数据格式）
export interface PlayerView {
  id: string
  name: string
  camp: string
  role: string
  hand_count: number
  exclusive_card_count: number
  hand?: Card[]  // 只有自己能看到
  blessings?: Card[] // 只有自己能看到（精灵射手祝福区）
  exclusive_cards?: Card[] // 只有自己能看到（专属技能卡区）
  field: FieldCard[]
  heal: number
  max_heal: number
  gem: number     // 个人能量：宝石
  crystal: number // 个人能量：水晶
  is_active: boolean
  buffs: Buff[]
  tokens?: Record<string, number>
}

// 技能摘要（后端 CharacterView.Skills）
export interface SkillView {
  id: string
  title: string
  description: string
  type?: number      // 0=Passive, 1=Startup, 2=Action, 3=Response
  min_targets?: number
  max_targets?: number
  target_type?: number
  cost_gem?: number
  cost_crystal?: number
  cost_discards?: number
  discard_type?: CardType
  discard_element?: string
  require_exclusive?: boolean  // 是否必须使用独有牌
}

// 角色摘要（后端 CharacterView，与 data.GetCharacters 一致）
export interface CharacterView {
  id: string
  name: string
  title: string
  faction: string
  skills: SkillView[]
}

// 游戏状态更新
export interface GameStateUpdate {
  phase: string
  current_player: string
  has_performed_startup?: boolean
  players: Record<string, PlayerView>
  red_morale: number
  blue_morale: number
  red_cups: number
  blue_cups: number
  red_gems: number
  blue_gems: number
  red_crystals: number
  blue_crystals: number
  deck_count: number
  discard_count?: number
  available_skills: AvailableSkill[]
  characters?: CharacterView[]
}

// Prompt 选项
export interface PromptOption {
  id: string
  label: string
}

// Prompt 类型
export type PromptType = 'choose_card' | 'choose_cards' | 'choose_target' | 'choose_skill' | 'confirm' | 'choose_extract'

// Prompt（请求玩家输入）
export interface Prompt {
  type: PromptType
  player_id: string
  message: string
  options: PromptOption[]
  /** 行动选择时“特殊”按钮对应的子选项（购买/合成/提炼） */
  special_options?: PromptOption[]
  /** 前端渲染提示：action_hub 表示用底部半球行动面板承载，不弹大面板 */
  ui_mode?: string
  effect_hints?: string[]
  min: number
  max: number
  /** 应战专用：发起攻击方ID */
  attacker_id?: string
  /** 应战专用：可选反弹目标（攻击方的队友，不含攻击者本人） */
  counter_target_ids?: string[]
  /** 应战专用：攻击牌元素（应战须同系或暗灭）Earth/Water/Fire/Wind/Thunder/Dark */
  attack_element?: string
}

// 玩家操作类型
export type PlayerActionType = 
  | 'Start' | 'Quit' | 'Pass' | 'Help'
  | 'Attack' | 'Magic' | 'Buy' | 'Synthesize' | 'Extract' | 'Skill'
  | 'Confirm' | 'Cancel' | 'Select' | 'Respond'
  | 'CannotAct'
  | 'Cheat'

// 玩家操作
export interface PlayerAction {
  player_id: string
  type: PlayerActionType
  target_id?: string
  target_ids?: string[]
  card_index?: number
  skill_id?: string
  selections?: number[]  // 选牌/弃牌索引，技能发动时用
  extra_args?: string[]
}

// WebSocket 消息
export interface WSMessage {
  type: 'action' | 'event' | 'room' | 'chat'
  payload: unknown
}

// 房间事件
export interface RoomEvent {
  action: 'joined' | 'left' | 'started' | 'player_list' | 'assigned' | 'error' | 'dissolved'
  room_code: string
  player_id?: string
  player_name?: string
  players?: PlayerInfo[]
  characters?: CharacterView[]
  message?: string
  camp?: string
  char_role?: string
  reconnect_token?: string
}

// 玩家信息（房间内）
export interface PlayerInfo {
  id: string
  name: string
  camp: string
  char_role: string
  ready: boolean
  is_online?: boolean
  is_bot?: boolean
  is_host?: boolean
  bot_mode?: string
}

// 游戏事件
export interface GameEvent {
  event_type: 'log' | 'state_update' | 'prompt' | 'waiting' | 'error' | 'game_end' | 'chat' | 'card_revealed' | 'damage_dealt' | 'action_step' | 'combat_cue' | 'draw_cards'
  message?: string
  state?: GameStateUpdate
  prompt?: Prompt
  player_id?: string
  player_name?: string
  /** 明牌展示（出牌/弃牌动画） */
  cards?: Card[]
  action_type?: 'attack' | 'magic' | 'discard' | 'defend' | 'counter'
  /** 是否为暗弃（隐藏牌面） */
  hidden?: boolean
  /** 伤害结算（暴血特效） */
  source_id?: string
  source_name?: string
  target_id?: string
  target_name?: string
  damage?: number
  damage_type?: string
  /** 行动步骤（桌面展示） */
  line?: string
  kind?: 'detail' | 'summary'
  /** 对战提示（战区动画） */
  attacker_id?: string
  phase?: 'attack' | 'defend' | 'take' | 'counter'
  /** 摸牌事件（公共牌堆 -> 玩家区域动画） */
  draw_count?: number
  reason?: string
}

// 元素颜色映射
export const ELEMENT_COLORS: Record<Element, string> = {
  Water: 'element-water',
  Fire: 'element-fire',
  Earth: 'element-earth',
  Wind: 'element-wind',
  Thunder: 'element-thunder',
  Light: 'element-light',
  Dark: 'element-dark'
}

// 元素中文名
export const ELEMENT_NAMES: Record<Element, string> = {
  Water: '水',
  Fire: '火',
  Earth: '地',
  Wind: '风',
  Thunder: '雷',
  Light: '光',
  Dark: '暗'
}
