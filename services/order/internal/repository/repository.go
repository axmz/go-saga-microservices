package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/axmz/go-saga-microservices/lib/adapter/db"
	"github.com/axmz/go-saga-microservices/pkg/proto/events"
	"github.com/axmz/go-saga-microservices/services/order/internal/domain"
	"google.golang.org/protobuf/proto"
)

type Repository struct {
	DB *db.DB
}

func New(db *db.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) CreateOrder(ctx context.Context, o *domain.Order) error {
	tx, err := r.DB.GetConn().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	var itemIDs string
	for i, item := range o.Items {
		if i > 0 {
			itemIDs += ","
		}
		itemIDs += item.ProductID
	}
	q := `INSERT INTO orders (id, item_ids, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.ExecContext(ctx, q, o.ID, itemIDs, o.Status, o.CreatedAt, o.UpdatedAt); err != nil {
		_ = tx.Rollback()
		return err
	}

	evtItems := make([]*events.Item, len(o.Items))
	for i, it := range o.Items {
		evtItems[i] = &events.Item{Id: it.ProductID}
	}
	env := &events.OrderEventEnvelope{
		Event: &events.OrderEventEnvelope_OrderCreated{
			OrderCreated: &events.OrderCreatedEvent{
				Id:    o.ID,
				Items: evtItems,
			},
		},
	}
	payload, err := proto.Marshal(env)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := r.InsertOutbox(ctx, tx, OutboxMessage{
		AggregateType: "order",
		AggregateID:   o.ID,
		EventType:     "OrderCreated",
		Payload:       payload,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *Repository) CreateOrderTx(ctx context.Context, tx *sql.Tx, o *domain.Order) error {
	var itemIDs string
	for i, item := range o.Items {
		if i > 0 {
			itemIDs += ","
		}
		itemIDs += item.ProductID
	}
	q := `INSERT INTO orders (id, item_ids, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := tx.ExecContext(ctx, q, o.ID, itemIDs, o.Status, o.CreatedAt, o.UpdatedAt)
	return err
}

func (r *Repository) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	row := r.DB.GetConn().QueryRowContext(ctx, `
		SELECT id, status, item_ids, created_at, updated_at
		FROM orders
		WHERE id = $1
	`, id)

	var o domain.Order
	var itemIDs string
	err := row.Scan(&o.ID, &o.Status, &itemIDs, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewErrOrderNotFound(id)
		}
		return nil, fmt.Errorf("query order by id %s: %w", id, err)
	}
	o.Items = make([]domain.Item, 0)
	for _, itemID := range strings.Split(itemIDs, ",") {
		o.Items = append(o.Items, domain.Item{ProductID: itemID})
	}
	return &o, nil
}

func (r *Repository) UpdateOrder(ctx context.Context, o *domain.Order) error {
	_, err := r.DB.GetConn().ExecContext(ctx, `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`, o.Status, o.UpdatedAt, o.ID)
	return err
}
