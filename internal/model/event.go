package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeNone = 1 + iota
	EventTypeUser
)

func NewEvent(eventType int, identifier, userID string, data json.RawMessage) Event {
	return Event{
		ID:         uuid.New(),
		Timestamp:  time.Now().UnixMilli(),
		Type:       eventType,
		Identifier: identifier,
		UserID:     userID,
		Data:       data,
	}
}

type Event struct {
	ID         uuid.UUID       `json:"id"`
	Timestamp  int64           `json:"timestamp"` // UnixMilli
	Type       int             `json:"type"`
	Identifier string          `json:"identifier"`
	UserID     string          `json:"user_id"`
	Data       json.RawMessage `json:"data"`
}

func (Event) Name() string {
	return "event"
}
