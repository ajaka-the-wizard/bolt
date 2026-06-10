DROP INDEX IF EXISTS idx_orders_customer_email;
DROP INDEX IF EXISTS idx_orders_created_at;

DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS status_enum;
