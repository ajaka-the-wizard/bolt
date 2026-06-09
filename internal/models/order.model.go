package models

import "time"

type Order struct {
	ID              string      `db:"id" json:"id"`
	OrderNumber     string      `db:"order_number" json:"order_number"`
	CustomerEmail   string      `db:"customer_email" json:"customer_email"`
	CustomerName    string      `db:"customer_name" json:"customer_name"`
	ShippingAddress Address     `db:"address" json:"address"`
	Items           []OrderItem `db:"items" json:"items"`
	Subtotal        float64     `db:"subtotal" json:"subtotal"`
	ShippingCost    float64     `db:"shipping_cost" json:"shipping_cost"`
	Tax             float64     `db:"tax" json:"tax"`
	Total           float64     `db:"total" json:"total"`
	Discount        float64     `db:"discount" json:"discount"`
	PaymentMethod   string      `db:"payment_method" json:"payment_method"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
}

type OrderItem struct {
	Name       string  `db:"name" json:"name"`
	Quantity   int     `db:"quantity" json:"quantity"`
	UnitPrice  float64 `db:"unit_price" json:"unit_price"`
	TotalPrice float64 `db:"total_price" json:"total_price"`
}
