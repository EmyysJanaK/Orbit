package repository

import (
	"context"
	"encoding/json"
	"errors"

	"employee-management-system/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func (r *EventRepository) Create(ctx context.Context, event models.Event) (models.Event, error) {
	metadata := event.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return models.Event{}, err
	}

	err = r.pool.QueryRow(
		ctx,
		`INSERT INTO events (user_id, event_type, metadata)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		event.UserID,
		event.EventType,
		metadataJSON,
	).Scan(&event.ID, &event.CreatedAt)
	if err != nil {
		return models.Event{}, err
	}

	return event, nil
}

func (r *EventRepository) EnsureCreated(ctx context.Context, event models.Event) (models.Event, error) {
	created, err := r.Create(ctx, event)
	if err != nil {
		return models.Event{}, err
	}

	return created, nil
}

var ErrEventRepositoryUnavailable = errors.New("event repository unavailable")
