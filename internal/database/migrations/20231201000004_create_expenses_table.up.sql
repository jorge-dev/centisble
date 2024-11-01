CREATE TABLE expenses (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    currency VARCHAR(3) NOT NULL,
    category VARCHAR(255) NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    description VARCHAR(510) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_expenses_user_id ON expenses (user_id);
CREATE INDEX idx_expenses_category ON expenses (category);

