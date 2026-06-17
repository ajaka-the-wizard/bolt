package database

import (
	"context"
	"errors"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) SaveOrder(ctx context.Context, data *models.Order) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// var pgErr *pgconn.PgError
	var id uuid.UUID

	query := `
	INSERT INTO orders (order_number,customer_name,customer_email,shipping_address,items,sub_total,shipping_cost,tax,discount,total,payment_method,currency)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	RETURNING id
	`
	if err := r.pool.QueryRow(ctx, query, data.OrderNumber, data.CustomerName, data.CustomerEmail, data.ShippingAddress, data.Items, data.Subtotal, data.ShippingCost, data.Tax, data.Discount, data.Total, data.PaymentMethod, data.Currency).Scan(&id); err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

func (r *Repo) FetchOrder(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
	SELECT * FROM orders
	WHERE id = $1
	`
	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	order, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.Order])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNoExists
		}
		return nil, err
	}

	return &order, nil
}
