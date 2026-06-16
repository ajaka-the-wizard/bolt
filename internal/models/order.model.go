package models

import "time"

type Order struct {
	ID              string      `db:"id" json:"id"`
	OrderNumber     string      `db:"order_number" json:"order_number"`
	CustomerEmail   string      `db:"customer_email" json:"customer_email"`
	CustomerName    string      `db:"customer_name" json:"customer_name"`
	ShippingAddress Address     `db:"shipping_address" json:"shipping_address"`
	Items           []OrderItem `db:"items" json:"items"`
	Subtotal        float64     `db:"sub_total" json:"sub_total"`
	ShippingCost    float64     `db:"shipping_cost" json:"shipping_cost"`
	Tax             float64     `db:"tax" json:"tax"`
	Total           float64     `db:"total" json:"total"`
	Discount        float64     `db:"discount" json:"discount"`
	PaymentMethod   string      `db:"payment_method" json:"payment_method"`
	Currency        string      `db:"currency" json:"currency"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
	Status          string      `db:"status"`
}

type OrderItem struct {
	Name       string  `db:"name" json:"name"`
	Quantity   int     `db:"quantity" json:"quantity"`
	UnitPrice  float64 `db:"unit_price" json:"unit_price"`
	TotalPrice float64 `db:"total_price" json:"total_price"`
}
