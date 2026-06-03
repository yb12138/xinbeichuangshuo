package protocol

import (
	"encoding/json"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/viewmodel"
)

const (
	CmdSyncState      WSCommand = "SyncState"
	CmdRequireAction  WSCommand = "RequireAction"
	CmdNotifyTimeline WSCommand = "NotifyTimeline"
	CmdSubmitAction   WSCommand = "SubmitAction"
	CmdRoomAction     WSCommand = "RoomAction"
	CmdRoomEvent      WSCommand = "RoomEvent"
	CmdChatMessage    WSCommand = "ChatMessage"
	CmdProtocolError  WSCommand = "ProtocolError"
)

type WSCommand string

// WSMessage is the standard websocket envelope used by both client and server.
type WSMessage struct {
	Cmd  WSCommand       `json:"Cmd"`
	Data json.RawMessage `json:"Data,omitempty"`
}

type ProtocolErrorPayload struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Cmd     WSCommand              `json:"cmd,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

type TargetNode struct {
	TargetUserID       string   `json:"target_user_id,omitempty"`
	SelectedFieldCards []string `json:"selected_field_cards,omitempty"`
	SelectedHandCards  []string `json:"selected_hand_cards,omitempty"`
	SelectedTokens     []string `json:"selected_tokens,omitempty"`
}

// ClientActionRequest is the submit-action protocol payload.
type ClientActionRequest struct {
	ActionType model.PlayerActionType `json:"action_type"`
	CardID     string                 `json:"card_id,omitempty"`
	CardIDs    []string               `json:"card_ids,omitempty"`
	Targets    []TargetNode           `json:"targets,omitempty"`
	SkillID    string                 `json:"skill_id,omitempty"`

	// OptionIndexes carries prompt option indexes for Select/Confirm actions.
	// It is distinct from CardID/CardIDs which carry card selection by UUID.
	OptionIndexes []int `json:"option_indexes,omitempty"`
	// ExtraArgs carries additional string arguments (Respond modes, Cheat subcommands).
	ExtraArgs []string `json:"extra_args,omitempty"`
}

type RoomActionRequest struct {
	Action   RoomActionType `json:"action"`
	Camp     string         `json:"camp,omitempty"`
	CharRole string         `json:"char_role,omitempty"`
	TargetID string         `json:"target_id,omitempty"`
	BotName  string         `json:"bot_name,omitempty"`
}

type RoomActionType string

const (
	RoomActionDissolveRoom   RoomActionType = "dissolve_room"
	RoomActionAddBot         RoomActionType = "add_bot"
	RoomActionRemoveBot      RoomActionType = "remove_bot"
	RoomActionTakeoverPlayer RoomActionType = "takeover_player"
	RoomActionChangeCamp     RoomActionType = "change_camp"
	RoomActionChangeRole     RoomActionType = "change_role"
	RoomActionStart          RoomActionType = "start"
)

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
	InterruptType string               `json:"interrupt_type"`
	TargetUserID  string               `json:"target_user_id"`
	Timeout       int                  `json:"timeout"`
	Msg           string               `json:"msg"`
	ValidActions  []WSCommand          `json:"valid_actions,omitempty"`
	RequireCount  int                  `json:"require_count,omitempty"`
	PromptType    string               `json:"prompt_type,omitempty"`
	Prompt        *viewmodel.PromptDTO `json:"prompt,omitempty"`
}

type TimelineDelta struct {
	Type          string           `json:"type"`
	Scope         string           `json:"scope,omitempty"`
	TargetUserID  string           `json:"target_user_id,omitempty"`
	Camp          string           `json:"camp,omitempty"`
	Field         string           `json:"field,omitempty"`
	Before        int              `json:"before,omitempty"`
	After         int              `json:"after,omitempty"`
	Value         int              `json:"value,omitempty"`
	Reason        string           `json:"reason,omitempty"`
	SourceEventID string           `json:"source_event_id,omitempty"`
	BeforeText    string           `json:"before_text,omitempty"`
	AfterText     string           `json:"after_text,omitempty"`
	FieldCard     *model.FieldCard `json:"field_card,omitempty"`
}

type TimelineEvent struct {
	EventID            int64            `json:"event_id"`
	TurnID             int              `json:"turn_id"`
	TurnStage          string           `json:"turn_stage,omitempty"`
	CombatStage        string           `json:"combat_stage,omitempty"`
	Subflow            string           `json:"subflow,omitempty"`
	Timing             string           `json:"timing,omitempty"`
	ChainID            string           `json:"chain_id"`
	ParentEventID      *int64           `json:"parent_event_id,omitempty"`
	SourceEventID      string           `json:"source_event_id,omitempty"`
	NarrativeWindowID  string           `json:"narrative_window_id,omitempty"`
	ActionID           string           `json:"action_id,omitempty"`
	CombatID           string           `json:"combat_id,omitempty"`
	NarrativeKind      string           `json:"narrative_kind,omitempty"`
	VisualKind         string           `json:"visual_kind,omitempty"`
	CardRole           string           `json:"card_role,omitempty"`
	SkillPhase         string           `json:"skill_phase,omitempty"`
	EffectType         string           `json:"effect_type,omitempty"`
	ExtraActionType    string           `json:"extra_action_type,omitempty"`
	ExtraActionElement string           `json:"extra_action_element,omitempty"`
	FieldCard          *model.FieldCard `json:"field_card,omitempty"`
	Type               string           `json:"type"`
	Outcome            string           `json:"outcome"`
	Visibility         string           `json:"visibility"`
	ActorUserID        string           `json:"actor_user_id,omitempty"`
	ActorName          string           `json:"actor_name,omitempty"`
	TargetUserIDs      []string         `json:"target_user_ids,omitempty"`
	TargetName         string           `json:"target_name,omitempty"`
	ActionType         string           `json:"action_type,omitempty"`
	SkillID            string           `json:"skill_id,omitempty"`
	SkillName          string           `json:"skill_name,omitempty"`
	EffectText         string           `json:"effect_text,omitempty"`
	Summary            string           `json:"summary,omitempty"`
	CardIDs            []string         `json:"card_ids,omitempty"`
	Cards              []model.Card     `json:"cards,omitempty"`
	Hidden             bool             `json:"hidden,omitempty"`
	Damage             int              `json:"damage,omitempty"`
	DamageType         string           `json:"damage_type,omitempty"`
	DetailKind         string           `json:"detail_kind,omitempty"`
	CuePhase           string           `json:"cue_phase,omitempty"`
	DrawCount          int              `json:"draw_count,omitempty"`
	Reason             string           `json:"reason,omitempty"`
	Deltas             []TimelineDelta  `json:"deltas,omitempty"`
	Message            string           `json:"message,omitempty"`
	GameplayType       string           `json:"gameplay_type,omitempty"`
}

type ActionFlowActorDTO struct {
	PlayerID string `json:"player_id"`
	Order    int    `json:"order"`
}

type ActionFlowNodeDTO struct {
	ID            string       `json:"id"`
	Order         int          `json:"order"`
	Kind          string       `json:"kind"`
	ActorUserID   string       `json:"actor_user_id,omitempty"`
	TargetUserIDs []string     `json:"target_user_ids,omitempty"`
	EventID       int64        `json:"event_id,omitempty"`
	AnchorEdgeID  string       `json:"anchor_edge_id,omitempty"`
	SkillID       string       `json:"skill_id,omitempty"`
	SkillName     string       `json:"skill_name,omitempty"`
	EffectText    string       `json:"effect_text,omitempty"`
	Cards         []model.Card `json:"cards,omitempty"`
	Outcome       string       `json:"outcome,omitempty"`
	Damage        int          `json:"damage,omitempty"`
	DamageType    string       `json:"damage_type,omitempty"`
	Label         string       `json:"label,omitempty"`
}

type ActionFlowEdgeDTO struct {
	ID          string       `json:"id"`
	Order       int          `json:"order"`
	FromUserID  string       `json:"from_user_id"`
	ToUserID    string       `json:"to_user_id"`
	Phase       string       `json:"phase"`
	CardEventID int64        `json:"card_event_id,omitempty"`
	Cards       []model.Card `json:"cards,omitempty"`
	NodeIDs     []string     `json:"node_ids,omitempty"`
	Outcome     string       `json:"outcome,omitempty"`
	Damage      int          `json:"damage,omitempty"`
	DamageType  string       `json:"damage_type,omitempty"`
	Label       string       `json:"label,omitempty"`
}

type ActionFlowLogDTO struct {
	Order int    `json:"order"`
	Text  string `json:"text"`
}

type ActionFlowDTO struct {
	FlowID            string               `json:"flow_id"`
	ActionID          string               `json:"action_id"`
	NarrativeWindowID string               `json:"narrative_window_id,omitempty"`
	ActorUserID       string               `json:"actor_user_id,omitempty"`
	ActionType        string               `json:"action_type"`
	Actors            []ActionFlowActorDTO `json:"actors"`
	Nodes             []ActionFlowNodeDTO  `json:"nodes,omitempty"`
	Edges             []ActionFlowEdgeDTO  `json:"edges,omitempty"`
	Logs              []ActionFlowLogDTO   `json:"logs,omitempty"`
}

type TimelineNotifyPayload struct {
	RoomID      string          `json:"room_id"`
	SeqStart    int64           `json:"seq_start"`
	SeqEnd      int64           `json:"seq_end"`
	IsReplay    bool            `json:"is_replay"`
	Events      []TimelineEvent `json:"events"`
	ActionFlows []ActionFlowDTO `json:"action_flows,omitempty"`
}
