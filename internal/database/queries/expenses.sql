-- name: CreateExpense :one
INSERT INTO expenses (
    id, user_id, amount, currency, category,
    date, description, created_at
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: GetExpenseByID :one
SELECT * FROM expenses
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListExpenses :many
SELECT * FROM expenses
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY date DESC;

-- name: UpdateExpense :one
UPDATE expenses 
SET 
    amount = $2,
    currency = $3,
    category = $4,
    date = $5,
    description = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $7 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteExpense :exec
UPDATE expenses 
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetExpensesByCategory :many
SELECT * FROM expenses
WHERE user_id = $1 
    AND category = $2 
    AND deleted_at IS NULL
    AND date >= sqlc.arg(start_date)::TIMESTAMPTZ
    AND date <= sqlc.arg(end_date)::TIMESTAMPTZ
ORDER BY date DESC;

-- name: GetExpensesByDateRange :many
SELECT * FROM expenses
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND date >= sqlc.arg(start_date)::TIMESTAMPTZ
    AND date <= sqlc.arg(end_date)::TIMESTAMPTZ
ORDER BY date DESC;

-- name: GetExpenseTotalsByCategory :many
SELECT 
    category,
    currency,
    COUNT(*) as transaction_count,
    SUM(amount) as total_amount
FROM expenses
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND date >= sqlc.arg(start_date)::TIMESTAMPTZ
    AND date <= sqlc.arg(end_date)::TIMESTAMPTZ
GROUP BY category, currency
ORDER BY total_amount DESC;

-- name: GetRecentExpenses :many
SELECT * FROM expenses
WHERE user_id = $1 
    AND deleted_at IS NULL
ORDER BY date DESC
LIMIT $2;

-- name: GetMonthlyExpenseTotal :one
SELECT 
    COALESCE(SUM(amount), 0) as total_amount,
    currency
FROM expenses
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND DATE_TRUNC('month', date) = DATE_TRUNC('month', sqlc.arg(date)::TIMESTAMPTZ)
GROUP BY currency;
