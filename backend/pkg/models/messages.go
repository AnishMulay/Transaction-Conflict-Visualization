package models

import "time"

// MessageType defines the types of WebSocket messages
type MessageType string

const (
	MessageTypeJoin      MessageType = "join"
	MessageTypeLeave     MessageType = "leave"
	MessageTypeMove      MessageType = "move"
	MessageTypeGameState MessageType = "gameState"
	MessageTypeError     MessageType = "error"
	MessageTypeConflict  MessageType = "conflict"
)

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data"`
	PlayerID  string      `json:"playerId,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// JoinRequest represents a player joining the game
type JoinRequest struct {
	PlayerName string `json:"playerName"`
}

// MoveRequest represents a move command with optimistic concurrency
type MoveRequest struct {
	Direction     string `json:"direction"`
	ObjectVersion int64  `json:"objectVersion"`
	RequestID     string `json:"requestId"`
}

// ErrorResponse represents error information
type ErrorResponse struct {
	Message   string `json:"message"`
	Code      string `json:"code"`
	RequestID string `json:"requestId,omitempty"`
}

// ConflictResponse represents a concurrency conflict
type ConflictResponse struct {
	Message         string    `json:"message"`
	ExpectedVersion int64     `json:"expectedVersion"`
	ActualVersion   int64     `json:"actualVersion"`
	RequestID       string    `json:"requestId"`
	Timestamp       time.Time `json:"timestamp"`
}
