-- name: CreateBudget :one
INSERT INTO budgets (
    id, user_id, amount, currency, category, 
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
    category = $4,
    type = $5,
    start_date = $6,
    end_date = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $8 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteBudget :exec
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
    AND category = $2 
    AND deleted_at IS NULL
ORDER BY start_date DESC;

-- name: GetBudgetUsage :one
WITH budget_expenses AS (
    SELECT COALESCE(SUM(amount), 0) as total_spent
    FROM expenses
    WHERE user_id = $2 
        AND category = (SELECT category FROM budgets WHERE id = $1)
        AND date >= (SELECT start_date FROM budgets WHERE id = $1)
        AND date <= COALESCE((SELECT end_date FROM budgets WHERE id = $1), CURRENT_DATE)
        AND deleted_at IS NULL
)
SELECT 
    b.*,
    e.total_spent,
    CASE 
        WHEN b.amount > 0 THEN (e.total_spent / b.amount * 100)
        ELSE 0 
    END as usage_percentage
FROM budgets b
CROSS JOIN budget_expenses e
WHERE b.id = $1 AND b.user_id = $2 AND b.deleted_at IS NULL;

-- name: GetBudgetsNearLimit :many
WITH budget_usage AS (
    SELECT 
        b.id,
        b.amount as budget_amount,
        COALESCE(SUM(e.amount), 0) as spent_amount,
        CASE 
            WHEN b.amount > 0 THEN (COALESCE(SUM(e.amount), 0) / b.amount * 100)
            ELSE 0 
        END as usage_percentage
    FROM budgets b
    LEFT JOIN expenses e ON 
        e.category = b.category 
        AND e.user_id = b.user_id
        AND e.date >= b.start_date
        AND e.date <= COALESCE(b.end_date, CURRENT_DATE)
        AND e.deleted_at IS NULL
    WHERE b.user_id = $1 
        AND b.deleted_at IS NULL
        AND b.start_date <= CURRENT_DATE
        AND (b.end_date >= CURRENT_DATE OR b.end_date IS NULL)
    GROUP BY b.id, b.amount
)
SELECT b.*, bu.spent_amount, bu.usage_percentage
FROM budgets b
JOIN budget_usage bu ON b.id = bu.id
WHERE bu.usage_percentage >= $2
ORDER BY bu.usage_percentage DESC;

-- name: GetRecurringBudgets :many
SELECT * FROM budgets
WHERE user_id = $1 
    AND type = 'recurring'
    AND deleted_at IS NULL
ORDER BY start_date ASC;
