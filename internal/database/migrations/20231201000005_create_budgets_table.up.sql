CREATE TABLE budgets (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    currency VARCHAR(3) NOT NULL,
    category VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('recurring', 'one-time')),
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT valid_date_range CHECK (end_date >= start_date)
);

CREATE INDEX idx_budgets_user_id ON budgets (user_id);
CREATE INDEX idx_budgets_start_end_date ON budgets (start_date, end_date);
CREATE INDEX idx_budgets_user_date ON budgets (user_id, start_date, end_date);
