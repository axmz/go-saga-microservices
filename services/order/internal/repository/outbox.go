package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxMessage struct {
	ID            uuid.UUID       `json:"id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Headers       map[string]any  `json:"headers"`
	CreatedAt     time.Time       `json:"created_at"`
}

func (r *Repository) InsertOutbox(ctx context.Context, tx *sql.Tx, msg OutboxMessage) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.Headers == nil {
		msg.Headers = map[string]any{}
	}
	b, _ := json.Marshal(msg.Headers)
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO outbox (id, aggregate_type, aggregate_id, event_type, payload, headers, created_at)
         VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)`,
		msg.ID, msg.AggregateType, msg.AggregateID, msg.EventType, []byte(msg.Payload), string(b), msg.CreatedAt,
	)
	return err
}
