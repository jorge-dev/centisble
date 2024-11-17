-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, role_id
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

-- name: DeleteUser :execrows
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

-- name: CheckUserIsAdmin :one
SELECT EXISTS(
    SELECT 1 
    FROM users 
    WHERE id = sqlc.arg('user_id')::uuid AND role_id = (SELECT id FROM roles WHERE name = 'Admin') AND deleted_at IS NULL
) AS is_admin;

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

-- name: UpdateUserRole :one
UPDATE users 
SET 
    role_id = sqlc.arg('role_id')::uuid,
    updated_at = CURRENT_TIMESTAMP
WHERE users.id = sqlc.arg('user_id')::uuid AND deleted_at IS NULL
RETURNING (
    SELECT json_build_object(
        'user_id', u.id,
        'user_name', u.name,
        'role_id', r.id,
        'role_name', r.name
    )
    FROM users u
    JOIN roles r ON r.id = sqlc.arg('role_id')::uuid
    WHERE u.id = sqlc.arg('user_id')::uuid
);

-- name: GetUserRole :one
SELECT 
    u.id as user_id,
    u.name as user_name,
    r.id as role_id,
    r.name as role_name
FROM users u
JOIN roles r ON u.role_id = r.id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: ListUsersByRole :many
SELECT u.id::uuid, u.name::varchar(255), u.email::varchar(255), r.name::varchar(255) as role
FROM roles r
LEFT JOIN users u ON u.role_id = r.id AND u.deleted_at IS NULL
WHERE r.name = $1
ORDER BY u.created_at DESC;
