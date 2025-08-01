package repository

import (
	"context"
	"fmt"

	"github.com/axmz/go-saga-microservices/inventory-service/internal/domain"
	"github.com/axmz/go-saga-microservices/lib/adapter/db"
	"github.com/axmz/go-saga-microservices/pkg/events"
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
	rows, err := r.DB.Conn().QueryContext(ctx, query)
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

func (r *Repository) ReserveItems(ctx context.Context, items []*events.Item) error {
	skus := make([]string, len(items))
	for i, item := range items {
		skus[i] = item.GetId()
	}

	const reserveQ = `
		UPDATE products
		SET    status = 'reserved'
		WHERE  sku = ANY($1) AND status = 'available'
	`

	res, err := r.DB.Conn().ExecContext(ctx, reserveQ, pq.Array(skus))
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows != int64(len(skus)) {
		return fmt.Errorf("one or more SKUs were not available; reserved %d of %d", rows, len(skus))
	}

	return nil
}

func (r *Repository) ReleaseItems(orderID, productID string) {
	query := `UPDATE products
			  SET status = 'available', updated_at = CURRENT_TIMESTAMP
			  WHERE sku = $1 AND status = 'reserved'`
	_, err := r.DB.Conn().Exec(query, productID)
	if err != nil {
		// Handle error (e.g., log it)
	}
}

func (r *Repository) ReleaseReservedItems(ctx context.Context, orderID string) error {
	// Note: This is a basic implementation. In a real system, we would need
	// a reservation table to track which products were reserved for which order.
	// For now, this is a no-op since we don't have order-to-product tracking.
	// In production, this would need proper order-product reservation tracking.
	return nil
}
