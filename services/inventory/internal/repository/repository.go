package repository

import (
	"context"
	"fmt"

	"github.com/axmz/go-saga-microservices/inventory-service/internal/domain"
	"github.com/axmz/go-saga-microservices/lib/adapter/db"
	"github.com/axmz/go-saga-microservices/pkg/proto/events"
	"github.com/lib/pq"
)

type Repository struct {
	DB *db.DB
}

func New(db *db.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetProducts(ctx context.Context) ([]domain.Product, error) {
	query := `SELECT id, name, sku, status, price
			  FROM products
			  ORDER BY name`
	rows, err := r.DB.GetConn().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		err := rows.Scan(&product.ID, &product.Name, &product.SKU, &product.Status, &product.Price)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *Repository) ReserveItems(ctx context.Context, event *events.OrderCreatedEvent) error {
	items := event.GetItems()
	skus := make([]string, len(items))
	for i, item := range items {
		skus[i] = item.GetId()
	}

	const reserveQ = `
		UPDATE products
		SET status = $1,
		order_id = $2
		WHERE sku = ANY($3) AND status = $4
	`

	res, err := r.DB.GetConn().ExecContext(
		ctx,
		reserveQ,
		domain.StatusReserved,
		event.Id,
		pq.Array(skus),
		domain.StatusAvailable,
	)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows != int64(len(skus)) {
		return fmt.Errorf("one or more SKUs were not available; reserved %d of %d", rows, len(skus))
	}

	return nil
}

func (r *Repository) ReleaseItems(ctx context.Context, orderID, productID string) {
	query := `UPDATE products
			  SET status = $1, updated_at = CURRENT_TIMESTAMP
			  WHERE sku = $2 AND status = $3`
	_, err := r.DB.GetConn().ExecContext(
		ctx,
		query,
		domain.StatusAvailable,
		productID,
		domain.StatusReserved,
	)
	if err != nil {
		// Handle error (e.g., log it)
	}
}

func (r *Repository) MarkItemsSold(ctx context.Context, orderID string) error {
	const query = `
		UPDATE products
		SET status = 'sold'
		WHERE order_id = $1
	`

	_, err := r.DB.GetConn().ExecContext(ctx, query, orderID)
	return err
}

func (r *Repository) ReleaseReservedItems(ctx context.Context, orderID string) error {
	const query = `
		UPDATE products
		SET status = $1, order_id = NULL
		WHERE order_id = $2
	`

	_, err := r.DB.GetConn().ExecContext(ctx, query, domain.StatusAvailable, orderID)
	return err
}
