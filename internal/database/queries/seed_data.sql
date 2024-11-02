-- name: SeedUsers :exec
INSERT INTO users (id, name, email, password_hash, created_at, role)
VALUES 
    (uuid_generate_v4(), 'John Doe', 'john.doe@example.com', 'hashed_password', CURRENT_TIMESTAMP, 'User');

-- name: SeedCategories :exec
INSERT INTO categories (id, user_id, name, created_at)
VALUES 
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 'Groceries', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 'Entertainment', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 'Utilities', CURRENT_TIMESTAMP);

-- name: SeedIncome :exec
INSERT INTO income (id, user_id, amount, currency, source, date, description, created_at)
VALUES 
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 5000.00, 'USD', 'Salary', '2023-10-01', 'Monthly Salary', CURRENT_TIMESTAMP);

-- name: SeedExpenses :exec
INSERT INTO expenses (id, user_id, amount, currency, category, date, description, created_at)
VALUES 
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 200.00, 'USD', 'Groceries', '2023-10-05', 'Weekly groceries', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 50.00, 'USD', 'Entertainment', '2023-10-10', 'Movie night', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 150.00, 'USD', 'Utilities', '2023-10-15', 'Monthly utility bill', CURRENT_TIMESTAMP);

-- name: SeedBudgets :exec
INSERT INTO budgets (id, user_id, amount, currency, category, type, start_date, end_date, created_at)
VALUES 
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 800.00, 'USD', 'Groceries', 'recurring', '2023-10-01', '2023-10-31', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 300.00, 'USD', 'Entertainment', 'recurring', '2023-10-01', '2023-10-31', CURRENT_TIMESTAMP),
    (uuid_generate_v4(), (SELECT id FROM users WHERE email = 'john.doe@example.com'), 200.00, 'USD', 'Utilities', 'recurring', '2023-10-01', '2023-10-31', CURRENT_TIMESTAMP);

-- name: DeleteSeedData :exec
DELETE FROM users WHERE email = 'john.doe@example.com';
