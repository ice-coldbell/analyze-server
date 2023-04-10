package cassandra

import (
	"encoding/json"

	"github.com/scylladb/gocqlx/v2/table"
)

var (
	metadataEvent = table.Metadata{
		Name: "event",
		Columns: []string{
			"id",
			"user_id",
			"identifier",
			"event_timestamp",
			"type",
		},
		PartKey: []string{"id"},
		SortKey: []string{"user_id", "identifier", "event_timestamp", "type"},
	}

	metadataEventData = table.Metadata{
		Name: "event_data",
		Columns: []string{
			"id",
			"data",
		},
		PartKey: []string{"id"},
	}

	metadataEventDate = table.Metadata{
		Name: "event_date",
		Columns: []string{
			"event_date",
			"event_timestamp",
			"id",
		},
		PartKey: []string{"event_date"},
		SortKey: []string{"event_timestamp", "id"},
	}

	metadataEventUserID = table.Metadata{
		Name: "event_user_id",
		Columns: []string{
			"user_id",
			"identifier",
			"id",
		},
		PartKey: []string{"user_id"},
		SortKey: []string{"identifier", "id"},
	}
)

var (
	tableEvent       = table.New(metadataEvent)
	tableEventData   = table.New(metadataEventData)
	tableEventDate   = table.New(metadataEventDate)
	tableEventUserID = table.New(metadataEventUserID)
)

type event struct {
	ID             [16]byte `json:"id"`
	EventTimestamp int64    `json:"event_timestamp"` // UnixMilli
	Type           int      `json:"type"`
	Identifier     string   `json:"identifier"`
	UserID         string   `json:"user_id"`
}

type eventData struct {
	ID   [16]byte        `json:"id"`
	Data json.RawMessage `json:"data"`
}

type eventDate struct {
	EventDate      string   `json:"event_date"`
	EventTimestamp int64    `json:"event_timestamp"` // UnixMilli
	ID             [16]byte `json:"id"`
}

type eventUserID struct {
	UserID     string   `json:"user_id"`
	Identifier string   `json:"identifier"`
	ID         [16]byte `json:"id"`
}
