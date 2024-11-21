-- name: CreateBudget :one
INSERT INTO budgets (
    id, user_id, amount, currency, category_id, 
    type, start_date, end_date, created_at, updated_at
)
VALUES (
    $1, $2, $3, $4, $5, 
    $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: GetBudgetByID :one
SELECT * FROM budgets
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListBudgets :many
SELECT * FROM budgets
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateBudget :one
UPDATE budgets 
SET 
    amount = $2,
    currency = $3,
    category_id = $4,
    type = $5,
    start_date = $6,
    end_date = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $8 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteBudget :execrows
UPDATE budgets 
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetActiveBudgets :many
SELECT * FROM budgets
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND start_date <= CURRENT_DATE
    AND (end_date >= CURRENT_DATE OR end_date IS NULL)
ORDER BY start_date ASC;

-- name: GetBudgetsByCategory :many
SELECT * FROM budgets
WHERE user_id = $1 
    AND category_id = $2 
    AND deleted_at IS NULL
ORDER BY start_date DESC;

-- name: GetBudgetUsage :one
WITH budget_expenses AS (
    SELECT COALESCE(SUM(amount), 0)::float8 AS total_spent
    FROM expenses
    WHERE user_id = sqlc.arg('user_id')::uuid
      AND category_id = (SELECT category_id FROM budgets WHERE id = sqlc.arg('budget_id')::uuid)
      AND deleted_at IS NULL
)
SELECT 
    sqlc.embed(b), -- Embed the entire budget row
    e.total_spent::float8 AS spent_amount,
    CASE 
        WHEN b.amount > 0 THEN (e.total_spent / b.amount * 100)::float8
        ELSE 0.0
    END AS usage_percentage
FROM budgets b
CROSS JOIN budget_expenses e
WHERE b.id = sqlc.arg('budget_id')::uuid
  AND b.user_id = sqlc.arg('user_id')::uuid
  AND b.deleted_at IS NULL;

-- name: GetRecurringBudgets :many
SELECT * FROM budgets
WHERE user_id = $1 
    AND type = 'recurring'
    AND deleted_at IS NULL
ORDER BY start_date ASC;

-- name: GetOneTimeBudgets :many
SELECT * FROM budgets
WHERE user_id = $1 
    AND type = 'one-time'
    AND deleted_at IS NULL
ORDER BY start_date ASC;

-- name: GetBudgetsNearLimit :many
SELECT 
    sqlc.embed(b), -- Embed the entire budget row
    COALESCE(spent_data.spent_amount, 0)::float8 AS spent_amount,
    CASE 
        WHEN b.amount > 0 THEN (COALESCE(spent_data.spent_amount, 0) / b.amount * 100)::float8
        ELSE 0.0
    END AS usage_percentage
FROM budgets b
LEFT JOIN (
    SELECT 
        e.category_id,
        e.user_id,
        SUM(e.amount) AS spent_amount
    FROM expenses e
    WHERE e.deleted_at IS NULL
    GROUP BY e.category_id, e.user_id
) AS spent_data 
ON b.category_id = spent_data.category_id 
   AND b.user_id = spent_data.user_id
WHERE b.user_id = sqlc.arg('user_id')::uuid
  AND b.deleted_at IS NULL
  AND b.start_date <= CURRENT_DATE
  AND (b.end_date >= CURRENT_DATE OR b.end_date IS NULL)
  AND CASE 
        WHEN b.amount > 0 THEN (COALESCE(spent_data.spent_amount, 0) / b.amount * 100)
        ELSE 0 
      END >= sqlc.arg('threshold')::float8
ORDER BY usage_percentage DESC;

