CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
    accrual DECIMAL(10, 2),
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для оптимизации
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_orders_status ON orders(status) WHERE status IN ('NEW', 'PROCESSING');

ALTER TABLE orders DROP CONSTRAINT orders_status_check;

-- Добавляем новое ограничение с поддержкой REGISTERED
ALTER TABLE orders ADD CONSTRAINT orders_status_check CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED', 'REGISTERED'));

- Удаляем старый индекс
DROP INDEX IF EXISTS idx_orders_status;

-- Создаём новый индекс, включающий REGISTERED
CREATE INDEX idx_orders_status ON orders(status) WHERE status IN ('NEW', 'PROCESSING', 'REGISTERED');