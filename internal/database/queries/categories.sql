-- name: CreateCategory :one
INSERT INTO categories (id, user_id, name, created_at, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListCategories :many
SELECT * FROM categories
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateCategory :one
UPDATE categories 
SET 
    name = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $3 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteCategory :execrows
UPDATE categories 
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: CheckCategoryExists :one
SELECT EXISTS(
    SELECT 1 FROM categories
    WHERE user_id = $1 AND name = $2 AND deleted_at IS NULL
);

-- name: GetCategoryUsage :one
SELECT 
    c.id,
    c.name,
    COUNT(DISTINCT e.id) as expense_count,
    COUNT(DISTINCT b.id) as budget_count,
    COALESCE(SUM(e.amount), 0) as total_expenses
FROM categories c
LEFT JOIN expenses e ON 
     e.category_id = c.id
    AND e.user_id = c.user_id 
    AND e.deleted_at IS NULL
LEFT JOIN budgets b ON 
     b.category_id = c.id 
    AND b.user_id = c.user_id 
    AND b.deleted_at IS NULL
WHERE c.id = $1 AND c.user_id = $2 AND c.deleted_at IS NULL
GROUP BY c.id, c.name;

-- name: GetMostUsedCategories :many
SELECT 
    c.name,
    COUNT(e.id) as usage_count,
    COALESCE(SUM(e.amount), 0) as total_amount
FROM categories c
LEFT JOIN expenses e ON 
    e.category_id = c.id 
    AND e.user_id = c.user_id 
    AND e.deleted_at IS NULL
WHERE c.user_id = sqlc.arg('id')::UUID
    AND c.deleted_at IS NULL
GROUP BY c.name
ORDER BY usage_count DESC
LIMIT sqlc.arg('limit')::int;
