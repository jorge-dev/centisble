// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: income.sql

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createIncome = `-- name: CreateIncome :one
INSERT INTO income (
    id, user_id, amount, currency, source,
    date, description, created_at
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, CURRENT_TIMESTAMP
)
RETURNING id, user_id, amount, currency, source, date, description, created_at, updated_at, deleted_at
`

type CreateIncomeParams struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Source      string    `json:"source"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

func (q *Queries) CreateIncome(ctx context.Context, arg CreateIncomeParams) (Income, error) {
	row := q.db.QueryRow(ctx, createIncome,
		arg.ID,
		arg.UserID,
		arg.Amount,
		arg.Currency,
		arg.Source,
		arg.Date,
		arg.Description,
	)
	var i Income
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Amount,
		&i.Currency,
		&i.Source,
		&i.Date,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const deleteIncome = `-- name: DeleteIncome :exec
UPDATE income 
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type DeleteIncomeParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

func (q *Queries) DeleteIncome(ctx context.Context, arg DeleteIncomeParams) error {
	_, err := q.db.Exec(ctx, deleteIncome, arg.ID, arg.UserID)
	return err
}

const getIncomeByDateRange = `-- name: GetIncomeByDateRange :many
SELECT id, user_id, amount, currency, source, date, description, created_at, updated_at, deleted_at FROM income
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND date >= $2::TIMESTAMPTZ
    AND date <= $3::TIMESTAMPTZ
ORDER BY date DESC
`

type GetIncomeByDateRangeParams struct {
	UserID    uuid.UUID `json:"user_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

func (q *Queries) GetIncomeByDateRange(ctx context.Context, arg GetIncomeByDateRangeParams) ([]Income, error) {
	rows, err := q.db.Query(ctx, getIncomeByDateRange, arg.UserID, arg.StartDate, arg.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Income
	for rows.Next() {
		var i Income
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Amount,
			&i.Currency,
			&i.Source,
			&i.Date,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const getIncomeByID = `-- name: GetIncomeByID :one
SELECT id, user_id, amount, currency, source, date, description, created_at, updated_at, deleted_at FROM income
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type GetIncomeByIDParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

func (q *Queries) GetIncomeByID(ctx context.Context, arg GetIncomeByIDParams) (Income, error) {
	row := q.db.QueryRow(ctx, getIncomeByID, arg.ID, arg.UserID)
	var i Income
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Amount,
		&i.Currency,
		&i.Source,
		&i.Date,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const getIncomeBySource = `-- name: GetIncomeBySource :many
SELECT id, user_id, amount, currency, source, date, description, created_at, updated_at, deleted_at FROM income
WHERE user_id = $1 
    AND source = $2 
    AND deleted_at IS NULL
ORDER BY date DESC
`

type GetIncomeBySourceParams struct {
	UserID uuid.UUID `json:"user_id"`
	Source string    `json:"source"`
}

func (q *Queries) GetIncomeBySource(ctx context.Context, arg GetIncomeBySourceParams) ([]Income, error) {
	rows, err := q.db.Query(ctx, getIncomeBySource, arg.UserID, arg.Source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Income
	for rows.Next() {
		var i Income
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Amount,
			&i.Currency,
			&i.Source,
			&i.Date,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const getIncomeSummaryBySource = `-- name: GetIncomeSummaryBySource :many
SELECT 
    source,
    currency,
    COUNT(*) as transaction_count,
    SUM(amount) as total_amount,
    AVG(amount) as average_amount
FROM income
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND date >= $2::TIMESTAMPTZ
    AND date <= $3::TIMESTAMPTZ
GROUP BY source, currency
ORDER BY total_amount DESC
`

type GetIncomeSummaryBySourceParams struct {
	UserID    uuid.UUID `json:"user_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type GetIncomeSummaryBySourceRow struct {
	Source           string  `json:"source"`
	Currency         string  `json:"currency"`
	TransactionCount int64   `json:"transaction_count"`
	TotalAmount      int64   `json:"total_amount"`
	AverageAmount    float64 `json:"average_amount"`
}

func (q *Queries) GetIncomeSummaryBySource(ctx context.Context, arg GetIncomeSummaryBySourceParams) ([]GetIncomeSummaryBySourceRow, error) {
	rows, err := q.db.Query(ctx, getIncomeSummaryBySource, arg.UserID, arg.StartDate, arg.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetIncomeSummaryBySourceRow
	for rows.Next() {
		var i GetIncomeSummaryBySourceRow
		if err := rows.Scan(
			&i.Source,
			&i.Currency,
			&i.TransactionCount,
			&i.TotalAmount,
			&i.AverageAmount,
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

const getMonthlyIncomeTotal = `-- name: GetMonthlyIncomeTotal :one
SELECT 
    COALESCE(SUM(amount), 0) as total_amount,
    currency
FROM income
WHERE user_id = $1 
    AND deleted_at IS NULL
    AND DATE_TRUNC('month', date) = DATE_TRUNC('month',  $2::TIMESTAMPTZ)
GROUP BY currency
`

type GetMonthlyIncomeTotalParams struct {
	UserID uuid.UUID `json:"user_id"`
	Date   time.Time `json:"date"`
}

type GetMonthlyIncomeTotalRow struct {
	TotalAmount interface{} `json:"total_amount"`
	Currency    string      `json:"currency"`
}

func (q *Queries) GetMonthlyIncomeTotal(ctx context.Context, arg GetMonthlyIncomeTotalParams) (GetMonthlyIncomeTotalRow, error) {
	row := q.db.QueryRow(ctx, getMonthlyIncomeTotal, arg.UserID, arg.Date)
	var i GetMonthlyIncomeTotalRow
	err := row.Scan(&i.TotalAmount, &i.Currency)
	return i, err
}

const getRecentIncome = `-- name: GetRecentIncome :many
SELECT id, user_id, amount, currency, source, date, description, created_at, updated_at, deleted_at FROM income
WHERE user_id = $1 
    AND deleted_at IS NULL
ORDER BY date DESC
LIMIT $2
`

type GetRecentIncomeParams struct {
	UserID uuid.UUID `json:"user_id"`
	Limit  int32     `json:"limit"`
}

func (q *Queries) GetRecentIncome(ctx context.Context, arg GetRecentIncomeParams) ([]Income, error) {
	rows, err := q.db.Query(ctx, getRecentIncome, arg.UserID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Income
	for rows.Next() {
		var i Income
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Amount,
			&i.Currency,
			&i.Source,
			&i.Date,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const listIncome = `-- name: ListIncome :many
SELECT id, user_id, amount, currency, source, date, description, created_at, updated_at, deleted_at FROM income
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY date DESC
`

func (q *Queries) ListIncome(ctx context.Context, userID uuid.UUID) ([]Income, error) {
	rows, err := q.db.Query(ctx, listIncome, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Income
	for rows.Next() {
		var i Income
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Amount,
			&i.Currency,
			&i.Source,
			&i.Date,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const updateIncome = `-- name: UpdateIncome :one
UPDATE income 
SET 
    amount = $2,
    currency = $3,
    source = $4,
    date = $5,
    description = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $7 AND deleted_at IS NULL
RETURNING id, user_id, amount, currency, source, date, description, created_at, updated_at, deleted_at
`

type UpdateIncomeParams struct {
	ID          uuid.UUID `json:"id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Source      string    `json:"source"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	UserID      uuid.UUID `json:"user_id"`
}

func (q *Queries) UpdateIncome(ctx context.Context, arg UpdateIncomeParams) (Income, error) {
	row := q.db.QueryRow(ctx, updateIncome,
		arg.ID,
		arg.Amount,
		arg.Currency,
		arg.Source,
		arg.Date,
		arg.Description,
		arg.UserID,
	)
	var i Income
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Amount,
		&i.Currency,
		&i.Source,
		&i.Date,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}
