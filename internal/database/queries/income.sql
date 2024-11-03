-- name: CreateIncome :one
INSERT INTO income (
    id, user_id, amount, currency, source,
    date, description, created_at, updated_at
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: GetIncomeByID :one
SELECT * FROM income
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListIncome :many
SELECT * FROM income
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY date DESC;

-- name: UpdateIncome :one
UPDATE income 
SET 
    amount = $2,
    currency = $3,
    source = $4,
    date = $5,
    description = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $7 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteIncome :execrows
UPDATE income 
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetIncomeByDateRange :many
SELECT * FROM income
WHERE user_id = sqlc.arg(user_id) 
    AND deleted_at IS NULL
    AND date >= sqlc.arg(start_date)::TIMESTAMPTZ
    AND date <= sqlc.arg(end_date)::TIMESTAMPTZ
ORDER BY date DESC;

-- name: GetIncomeBySource :many
SELECT * FROM income
WHERE user_id = $1 
    AND source = $2 
    AND deleted_at IS NULL
ORDER BY date DESC;

-- name: GetMonthlyIncomeTotal :one
SELECT 
    COALESCE(SUM(amount), 0) as total_amount,
    currency
FROM income
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND DATE_TRUNC('month', date) = DATE_TRUNC('month',  sqlc.arg(date)::TIMESTAMPTZ)
GROUP BY currency;

-- name: GetIncomeSummaryBySource :many
SELECT 
    source,
    currency,
    COUNT(*) as transaction_count,
    SUM(amount) as total_amount,
    AVG(amount) as average_amount
FROM income
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND date >= sqlc.arg(start_date)::TIMESTAMPTZ
    AND date <= sqlc.arg(end_date)::TIMESTAMPTZ
GROUP BY source, currency
ORDER BY total_amount DESC;

-- name: GetRecentIncome :many
SELECT * FROM income
WHERE user_id = $1 
    AND deleted_at IS NULL
ORDER BY date DESC
LIMIT $2;
