CREATE TABLE IF NOT EXISTS orders (
    id               UUID PRIMARY KEY DEFAULT uuidv7(),
    order_number     VARCHAR(50) UNIQUE NOT NULL,

    customer_name    VARCHAR(150) NOT NULL,
    customer_email   VARCHAR(150) NOT NULL,

    shipping_address JSONB NOT NULL,
    items            JSONB NOT NULL,

    subtotal         DECIMAL(12,2) NOT NULL DEFAULT 0,
    shipping_cost    DECIMAL(12,2) NOT NULL DEFAULT 0,
    tax              DECIMAL(12,2) NOT NULL DEFAULT 0,
    discount         DECIMAL(12,2) NOT NULL DEFAULT 0,
    total            DECIMAL(12,2) NOT NULL DEFAULT 0,

    payment_method   VARCHAR(50) NOT NULL,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_orders_customer_email ON orders(customer_email);
CREATE INDEX idx_orders_created_at ON orders(created_at);
