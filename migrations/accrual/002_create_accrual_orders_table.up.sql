CREATE TABLE IF NOT EXISTS accrual_orders (
    order_number VARCHAR(255) PRIMARY KEY,
    status VARCHAR(50) NOT NULL CHECK (status IN ('REGISTERED', 'INVALID', 'PROCESSING', 'PROCESSED')),
    accrual DECIMAL(10, 2),
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accrual_orders_status ON accrual_orders(status) WHERE status IN ('REGISTERED', 'PROCESSING');
