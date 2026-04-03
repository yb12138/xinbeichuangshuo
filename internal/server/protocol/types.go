package protocol

import (
	"encoding/json"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/viewmodel"
)

const (
	CmdSyncState      = "SyncState"
	CmdRequireAction  = "RequireAction"
	CmdNotifyTimeline = "NotifyTimeline"
	CmdSubmitAction   = "SubmitAction"
	CmdRoomAction     = "RoomAction"
	CmdRoomEvent      = "RoomEvent"
	CmdChatMessage    = "ChatMessage"
)

// WSMessage is the standard websocket envelope used by both client and server.
type WSMessage struct {
	Cmd  string          `json:"Cmd"`
	Data json.RawMessage `json:"Data,omitempty"`
}

type TargetNode struct {
	TargetUserID       string   `json:"target_user_id,omitempty"`
	SelectedFieldCards []string `json:"selected_field_cards,omitempty"`
	SelectedHandCards  []string `json:"selected_hand_cards,omitempty"`
	SelectedTokens     []string `json:"selected_tokens,omitempty"`
}

// ClientActionRequest is the new submit-action protocol payload.
// Some transitional fields are kept so the current engine can still be adapted
// from the docs-shaped request without rewriting the whole interrupt system at once.
type ClientActionRequest struct {
	ActionType        string         `json:"action_type"`
	UsedCardUUIDs     []string       `json:"used_card_uuids,omitempty"`
	Targets           []TargetNode   `json:"targets,omitempty"`
	TargetAllocations map[string]int `json:"target_allocations,omitempty"`
	SkillID           string         `json:"skill_id,omitempty"`
	ActionRef         string         `json:"action_ref,omitempty"`
	StoneRef          string         `json:"stone_ref,omitempty"`
	ElementRef        string         `json:"element_ref,omitempty"`
	NamedValues       map[string]int `json:"named_values,omitempty"`

	// Transitional adapter fields.
	OptionIndexes []int    `json:"option_indexes,omitempty"`
	ResponseMode  string   `json:"response_mode,omitempty"`
	ExtraArgs     []string `json:"extra_args,omitempty"`
	TargetRef     string   `json:"target_ref,omitempty"`
}

type RoomActionRequest struct {
	Action   string `json:"action"`
	Camp     string `json:"camp,omitempty"`
	CharRole string `json:"char_role,omitempty"`
	TargetID string `json:"target_id,omitempty"`
	BotName  string `json:"bot_name,omitempty"`
}

type SyncStatePayload struct {
	RoomState           string                     `json:"room_state"`
	TurnStage           string                     `json:"turn_stage,omitempty"`
	CombatStage         string                     `json:"combat_stage,omitempty"`
	Subflow             string                     `json:"subflow,omitempty"`
	TurnPlayerID        string                     `json:"turn_player_id"`
	HasPerformedStartup bool                       `json:"has_performed_startup"`
	MoraleRed           int                        `json:"morale_red"`
	MoraleBlue          int                        `json:"morale_blue"`
	CupsRed             int                        `json:"cups_red"`
	CupsBlue            int                        `json:"cups_blue"`
	StonesRed           []int                      `json:"stones_red"`
	StonesBlue          []int                      `json:"stones_blue"`
	DeckCount           int                        `json:"deck_count"`
	DiscardCount        int                        `json:"discard_count"`
	AvailableSkills     []viewmodel.AvailableSkill `json:"available_skills,omitempty"`
	Characters          []viewmodel.CharacterView  `json:"characters,omitempty"`
	Players             []viewmodel.PlayerView     `json:"players"`
}

type RequireActionPayload struct {
	InterruptType string        `json:"interrupt_type"`
	TargetUserID  string        `json:"target_user_id"`
	Timeout       int           `json:"timeout"`
	Msg           string        `json:"msg"`
	ValidActions  []string      `json:"valid_actions,omitempty"`
	RequireCount  int           `json:"require_count,omitempty"`
	PromptType    string        `json:"prompt_type,omitempty"`
	Prompt        *model.Prompt `json:"prompt,omitempty"`
}

type TimelineDelta struct {
	Type         string `json:"type"`
	TargetUserID string `json:"target_user_id,omitempty"`
	Value        int    `json:"value,omitempty"`
}

type TimelineEvent struct {
	EventID       int64           `json:"event_id"`
	TurnID        int             `json:"turn_id"`
	TurnStage     string          `json:"turn_stage,omitempty"`
	CombatStage   string          `json:"combat_stage,omitempty"`
	Subflow       string          `json:"subflow,omitempty"`
	Timing        string          `json:"timing,omitempty"`
	ChainID       string          `json:"chain_id"`
	ParentEventID *int64          `json:"parent_event_id,omitempty"`
	Type          string          `json:"type"`
	Outcome       string          `json:"outcome"`
	Visibility    string          `json:"visibility"`
	ActorUserID   string          `json:"actor_user_id,omitempty"`
	ActorName     string          `json:"actor_name,omitempty"`
	TargetUserIDs []string        `json:"target_user_ids,omitempty"`
	TargetName    string          `json:"target_name,omitempty"`
	ActionType    string          `json:"action_type,omitempty"`
	SkillID       string          `json:"skill_id,omitempty"`
	CardIDs       []string        `json:"card_ids,omitempty"`
	Cards         []model.Card    `json:"cards,omitempty"`
	Hidden        bool            `json:"hidden,omitempty"`
	Damage        int             `json:"damage,omitempty"`
	DamageType    string          `json:"damage_type,omitempty"`
	DetailKind    string          `json:"detail_kind,omitempty"`
	CuePhase      string          `json:"cue_phase,omitempty"`
	DrawCount     int             `json:"draw_count,omitempty"`
	Reason        string          `json:"reason,omitempty"`
	Deltas        []TimelineDelta `json:"deltas,omitempty"`
	Message       string          `json:"message,omitempty"`
	GameplayType  string          `json:"gameplay_type,omitempty"`
}

type TimelineNotifyPayload struct {
	RoomID   string          `json:"room_id"`
	SeqStart int64           `json:"seq_start"`
	SeqEnd   int64           `json:"seq_end"`
	IsReplay bool            `json:"is_replay"`
	Events   []TimelineEvent `json:"events"`
}
