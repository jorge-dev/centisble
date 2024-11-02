-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, name, email, created_at 
FROM users 
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users 
SET 
    name = $2,
    email = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users 
SET 
    password_hash = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id;

-- name: DeleteUser :exec
UPDATE users 
SET 
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: CheckEmailExists :one
SELECT EXISTS(
    SELECT 1 
    FROM users 
    WHERE email = $1 AND deleted_at IS NULL
);

-- name: GetUserStats :one
SELECT 
    u.id,
    u.name,
    COUNT(DISTINCT i.id) as total_income_records,
    COUNT(DISTINCT e.id) as total_expense_records,
    COUNT(DISTINCT b.id) as total_budgets
FROM users u
LEFT JOIN income i ON u.id = i.user_id AND i.deleted_at IS NULL
LEFT JOIN expenses e ON u.id = e.user_id AND e.deleted_at IS NULL
LEFT JOIN budgets b ON u.id = b.user_id AND b.deleted_at IS NULL
WHERE u.id = $1 AND u.deleted_at IS NULL
GROUP BY u.id, u.name;
