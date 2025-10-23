CREATE TABLE IF NOT EXISTS goods_rewards (
    description VARCHAR(255) UNIQUE NOT NULL,
    reward_type VARCHAR(20) NOT NULL CHECK (reward_type IN ('percentage', 'fixed')),
    reward_value DECIMAL(10, 2) NOT NULL CHECK (reward_value > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_goods_rewards_description ON goods_rewards(description);
