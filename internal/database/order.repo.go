package database

import (
	"context"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/google/uuid"
)

func (r *Repo) SaveOrder(ctx context.Context, data *models.Order) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// var pgErr *pgconn.PgError
	var id uuid.UUID

	query := `
	INSERT INTO orders (order_number,customer_name,customer_email,shipping_address,items,sub_total,shipping_cost,tax,discount,total,payment_method)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	RETURNING id
	`
	if err := r.pool.QueryRow(ctx, query, data.OrderNumber, data.CustomerName, data.CustomerEmail, data.ShippingAddress, data.Items, data.Subtotal, data.ShippingCost, data.Tax, data.Discount, data.Total, data.PaymentMethod).Scan(&id); err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}
