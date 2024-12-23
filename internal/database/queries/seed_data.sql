-- name: SeedUsers :exec
WITH role_ids AS (
    SELECT id, name FROM roles
)
INSERT INTO users (id, name, email, password_hash, created_at, updated_at, role_id)
VALUES 
    (uuid_generate_v4(), 'John Doe', 'john.doe@example.com', 
     '$2a$10$YZjEaHHtUBD/4RniGrx7ZO5TQShEBurJmc4Yz9Un.RFS4rP1W1hjm', 
     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 
     (SELECT id FROM role_ids WHERE name = 'Admin')),
    (uuid_generate_v4(), 'Jane Smith', 'jane.smith@example.com', 
     '$2a$10$YZjEaHHtUBD/4RniGrx7ZO5TQShEBurJmc4Yz9Un.RFS4rP1W1hjm', 
     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 
     (SELECT id FROM role_ids WHERE name = 'User')),
    (uuid_generate_v4(), 'Bob Wilson', 'bob.wilson@example.com', 
     '$2a$10$YZjEaHHtUBD/4RniGrx7ZO5TQShEBurJmc4Yz9Un.RFS4rP1W1hjm', 
     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 
     (SELECT id FROM role_ids WHERE name = 'User'));

-- name: SeedCategories :exec
WITH users AS (
    SELECT id, email FROM users WHERE email IN ('john.doe@example.com', 'jane.smith@example.com', 'bob.wilson@example.com')
)
INSERT INTO categories (id, user_id, name, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    users.id,
    category_name,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM users, unnest(ARRAY[
    'Groceries', 'Entertainment', 'Utilities', 'Rent', 'Transportation',
    'Healthcare', 'Shopping', 'Restaurants', 'Travel', 'Education'
]) AS category_name;

-- name: SeedIncome :exec
INSERT INTO income (id, user_id, amount, currency, source, date, description, created_at, updated_at)
SELECT
    uuid_generate_v4(),
    u.id,
    CASE 
        WHEN u.email = 'john.doe@example.com' THEN 7500.00
        WHEN u.email = 'jane.smith@example.com' THEN 5500.00
        ELSE 4500.00
    END,
    'USD',
    CASE (random() * 2)::int
        WHEN 0 THEN 'Salary'
        WHEN 1 THEN 'Freelance'
        ELSE 'Investments'
    END,
    date,
    'Monthly Income',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM users u
CROSS JOIN generate_series(
    '2023-07-01'::date,
    '2023-12-01'::date,
    '1 month'::interval
) AS date
WHERE u.email IN ('john.doe@example.com', 'jane.smith@example.com', 'bob.wilson@example.com');

-- name: SeedExpenses :exec
WITH users AS (
    SELECT id, email FROM users WHERE email IN ('john.doe@example.com', 'jane.smith@example.com', 'bob.wilson@example.com')
),
dates AS (
    SELECT generate_series(
        '2023-07-01'::date,
        '2023-12-31'::date,
        '3 days'::interval
    ) AS date
)
INSERT INTO expenses (id, user_id, category_id, amount, currency, date, description, created_at, updated_at)
SELECT
    uuid_generate_v4(),
    u.id,
    c.id,
    (random() * 300 + 20)::numeric(10,2),
    'USD',
    d.date,
    CASE (random() * 3)::int
        WHEN 0 THEN 'Regular expense'
        WHEN 1 THEN 'Monthly payment'
        WHEN 2 THEN 'One-time purchase'
        ELSE 'Miscellaneous'
    END,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM users u
CROSS JOIN dates d
JOIN categories c ON c.user_id = u.id
WHERE random() < 0.3;

-- name: SeedBudgets :exec
WITH users AS (
    SELECT id, email FROM users WHERE email IN ('john.doe@example.com', 'jane.smith@example.com', 'bob.wilson@example.com')
)
INSERT INTO budgets (id, user_id, category_id, amount, currency, type, start_date, end_date, name, created_at, updated_at)
SELECT
    uuid_generate_v4(),
    u.id,
    c.id,
    (random() * 1000 + 200)::numeric(10,2),
    'USD',
    CASE (random() * 1)::int
        WHEN 0 THEN 'recurring'
        ELSE 'one-time'
    END,
    date_trunc('month', CURRENT_DATE),
    date_trunc('month', CURRENT_DATE) + interval '1 month' - interval '1 day',
    c.name || ' Budget', -- Add budget name based on category
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM users u
JOIN categories c ON c.user_id = u.id;

-- name: DeleteSeedData :exec
-- First delete the users (cascading delete will handle related records)
DELETE FROM users WHERE email IN (
    'john.doe@example.com',
    'jane.smith@example.com',
    'bob.wilson@example.com'
);
