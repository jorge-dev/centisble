// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: budgets.sql

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createBudget = `-- name: CreateBudget :one
INSERT INTO budgets (
    id, user_id, amount, currency, category, 
    type, start_date, end_date, created_at, updated_at
)
VALUES (
    $1, $2, $3, $4, $5, 
    $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING id, user_id, amount, currency, category, type, start_date, end_date, created_at, deleted_at
`

type CreateBudgetParams struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Category  string    `json:"category"`
	Type      string    `json:"type"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

func (q *Queries) CreateBudget(ctx context.Context, arg CreateBudgetParams) (Budget, error) {
	row := q.db.QueryRow(ctx, createBudget,
		arg.ID,
		arg.UserID,
		arg.Amount,
		arg.Currency,
		arg.Category,
		arg.Type,
		arg.StartDate,
		arg.EndDate,
	)
	var i Budget
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Amount,
		&i.Currency,
		&i.Category,
		&i.Type,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const deleteBudget = `-- name: DeleteBudget :exec
UPDATE budgets 
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type DeleteBudgetParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

func (q *Queries) DeleteBudget(ctx context.Context, arg DeleteBudgetParams) error {
	_, err := q.db.Exec(ctx, deleteBudget, arg.ID, arg.UserID)
	return err
}

const getActiveBudgets = `-- name: GetActiveBudgets :many
SELECT id, user_id, amount, currency, category, type, start_date, end_date, created_at, deleted_at FROM budgets
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND start_date <= CURRENT_DATE
    AND (end_date >= CURRENT_DATE OR end_date IS NULL)
ORDER BY start_date ASC
`

func (q *Queries) GetActiveBudgets(ctx context.Context, userID uuid.UUID) ([]Budget, error) {
	rows, err := q.db.Query(ctx, getActiveBudgets, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Budget
	for rows.Next() {
		var i Budget
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Amount,
			&i.Currency,
			&i.Category,
			&i.Type,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getBudgetByID = `-- name: GetBudgetByID :one
SELECT id, user_id, amount, currency, category, type, start_date, end_date, created_at, deleted_at FROM budgets
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type GetBudgetByIDParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

func (q *Queries) GetBudgetByID(ctx context.Context, arg GetBudgetByIDParams) (Budget, error) {
	row := q.db.QueryRow(ctx, getBudgetByID, arg.ID, arg.UserID)
	var i Budget
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Amount,
		&i.Currency,
		&i.Category,
		&i.Type,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const getBudgetUsage = `-- name: GetBudgetUsage :one
WITH budget_expenses AS (
    SELECT COALESCE(SUM(amount), 0) AS total_spent
    FROM expenses
    WHERE user_id = $2
      AND category = (SELECT category FROM budgets WHERE id = $1)
      AND date >= (SELECT start_date FROM budgets WHERE id = $1)
      AND date <= COALESCE((SELECT end_date FROM budgets WHERE id = $1), CURRENT_DATE)
      AND deleted_at IS NULL
)
SELECT 
    b.id, b.user_id, b.amount, b.currency, b.category, b.type, b.start_date, b.end_date, b.created_at, b.deleted_at, -- Embed the entire budget row
    CASE 
        WHEN b.amount > 0 THEN (e.total_spent / b.amount * 100)
        ELSE 0 
    END AS usage_percentage
FROM budgets b
CROSS JOIN budget_expenses e
WHERE b.id = $1 
  AND b.user_id = $2 
  AND b.deleted_at IS NULL
`

type GetBudgetUsageParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

type GetBudgetUsageRow struct {
	Budget          Budget `json:"budget"`
	UsagePercentage int32  `json:"usage_percentage"`
}

func (q *Queries) GetBudgetUsage(ctx context.Context, arg GetBudgetUsageParams) (GetBudgetUsageRow, error) {
	row := q.db.QueryRow(ctx, getBudgetUsage, arg.ID, arg.UserID)
	var i GetBudgetUsageRow
	err := row.Scan(
		&i.Budget.ID,
		&i.Budget.UserID,
		&i.Budget.Amount,
		&i.Budget.Currency,
		&i.Budget.Category,
		&i.Budget.Type,
		&i.Budget.StartDate,
		&i.Budget.EndDate,
		&i.Budget.CreatedAt,
		&i.Budget.DeletedAt,
		&i.UsagePercentage,
	)
	return i, err
}

const getBudgetsByCategory = `-- name: GetBudgetsByCategory :many
SELECT id, user_id, amount, currency, category, type, start_date, end_date, created_at, deleted_at FROM budgets
WHERE user_id = $1 
    AND category = $2 
    AND deleted_at IS NULL
ORDER BY start_date DESC
`

type GetBudgetsByCategoryParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Category string    `json:"category"`
}

func (q *Queries) GetBudgetsByCategory(ctx context.Context, arg GetBudgetsByCategoryParams) ([]Budget, error) {
	rows, err := q.db.Query(ctx, getBudgetsByCategory, arg.UserID, arg.Category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Budget
	for rows.Next() {
		var i Budget
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Amount,
			&i.Currency,
			&i.Category,
			&i.Type,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getBudgetsNearLimit = `-- name: GetBudgetsNearLimit :many
SELECT 
    b.id, b.user_id, b.amount, b.currency, b.category, b.type, b.start_date, b.end_date, b.created_at, b.deleted_at, -- Embed the entire budget row
    COALESCE(spent_data.spent_amount, 0) AS spent_amount,
    CASE 
        WHEN b.amount > 0 THEN (COALESCE(spent_data.spent_amount, 0) / b.amount * 100)
        ELSE 0 
    END AS usage_percentage
FROM budgets b
LEFT JOIN (
    SELECT 
        e.category,
        e.user_id,
        SUM(e.amount) AS spent_amount
    FROM expenses e
    WHERE e.deleted_at IS NULL
    GROUP BY e.category, e.user_id
) AS spent_data 
ON b.category = spent_data.category 
   AND b.user_id = spent_data.user_id
WHERE b.user_id = $1 
  AND b.deleted_at IS NULL
  AND b.start_date <= CURRENT_DATE
  AND (b.end_date >= CURRENT_DATE OR b.end_date IS NULL)
  AND CASE 
        WHEN b.amount > 0 THEN (COALESCE(spent_data.spent_amount, 0) / b.amount * 100)
        ELSE 0 
      END >= $2
ORDER BY usage_percentage DESC
`

type GetBudgetsNearLimitParams struct {
	UserID uuid.UUID `json:"user_id"`
	Amount float64   `json:"amount"`
}

type GetBudgetsNearLimitRow struct {
	Budget          Budget `json:"budget"`
	SpentAmount     int64  `json:"spent_amount"`
	UsagePercentage int32  `json:"usage_percentage"`
}

func (q *Queries) GetBudgetsNearLimit(ctx context.Context, arg GetBudgetsNearLimitParams) ([]GetBudgetsNearLimitRow, error) {
	rows, err := q.db.Query(ctx, getBudgetsNearLimit, arg.UserID, arg.Amount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetBudgetsNearLimitRow
	for rows.Next() {
		var i GetBudgetsNearLimitRow
		if err := rows.Scan(
			&i.Budget.ID,
			&i.Budget.UserID,
			&i.Budget.Amount,
			&i.Budget.Currency,
			&i.Budget.Category,
			&i.Budget.Type,
			&i.Budget.StartDate,
			&i.Budget.EndDate,
			&i.Budget.CreatedAt,
			&i.Budget.DeletedAt,
			&i.SpentAmount,
			&i.UsagePercentage,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRecurringBudgets = `-- name: GetRecurringBudgets :many
SELECT id, user_id, amount, currency, category, type, start_date, end_date, created_at, deleted_at FROM budgets
WHERE user_id = $1 
    AND type = 'recurring'
    AND deleted_at IS NULL
ORDER BY start_date ASC
`

func (q *Queries) GetRecurringBudgets(ctx context.Context, userID uuid.UUID) ([]Budget, error) {
	rows, err := q.db.Query(ctx, getRecurringBudgets, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Budget
	for rows.Next() {
		var i Budget
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Amount,
			&i.Currency,
			&i.Category,
			&i.Type,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBudgets = `-- name: ListBudgets :many
SELECT id, user_id, amount, currency, category, type, start_date, end_date, created_at, deleted_at FROM budgets
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
`

func (q *Queries) ListBudgets(ctx context.Context, userID uuid.UUID) ([]Budget, error) {
	rows, err := q.db.Query(ctx, listBudgets, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Budget
	for rows.Next() {
		var i Budget
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Amount,
			&i.Currency,
			&i.Category,
			&i.Type,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateBudget = `-- name: UpdateBudget :one
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
RETURNING id, user_id, amount, currency, category, type, start_date, end_date, created_at, deleted_at
`

type UpdateBudgetParams struct {
	ID        uuid.UUID `json:"id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Category  string    `json:"category"`
	Type      string    `json:"type"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	UserID    uuid.UUID `json:"user_id"`
}

func (q *Queries) UpdateBudget(ctx context.Context, arg UpdateBudgetParams) (Budget, error) {
	row := q.db.QueryRow(ctx, updateBudget,
		arg.ID,
		arg.Amount,
		arg.Currency,
		arg.Category,
		arg.Type,
		arg.StartDate,
		arg.EndDate,
		arg.UserID,
	)
	var i Budget
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Amount,
		&i.Currency,
		&i.Category,
		&i.Type,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedAt,
		&i.DeletedAt,
	)
	return i, err
}
