// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: summary.sql

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const getMonthlySummary = `-- name: GetMonthlySummary :many
WITH monthly_totals AS (
    SELECT 
        i.currency::varchar(3) AS currency,
        COALESCE(SUM(i.amount), 0)::float8 as total_income,
        COALESCE(SUM(e.amount), 0)::float8 as total_expenses,
        COALESCE(SUM(i.amount) - SUM(e.amount), 0)::float8 as total_savings
    FROM income i
    FULL OUTER JOIN expenses e ON 
        e.user_id = i.user_id 
        AND e.currency = i.currency::varchar(3)
        AND DATE_TRUNC('month', e.date) = DATE_TRUNC('month', i.date)
        AND e.deleted_at IS NULL
    WHERE i.user_id = $1
        AND i.deleted_at IS NULL
        AND DATE_TRUNC('month', i.date) = DATE_TRUNC('month', $2::TIMESTAMPTZ)
    GROUP BY i.currency::varchar(3)
),
top_categories AS (
    SELECT 
        e.category_id,
        c.name as category_name,
        e.currency::varchar(3),
        COUNT(*) as usage_count,
        SUM(e.amount) as total_spent,
        ROW_NUMBER() OVER (PARTITION BY e.currency::varchar(3) ORDER BY SUM(e.amount) DESC) as rank
    FROM expenses e
    JOIN categories c ON e.category_id = c.id
    WHERE e.user_id = $1
        AND e.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND DATE_TRUNC('month', e.date) = DATE_TRUNC('month', $2::TIMESTAMPTZ)
    GROUP BY e.category_id, c.name, e.currency::varchar(3)
)
SELECT 
    mt.currency, mt.total_income, mt.total_expenses, mt.total_savings,
    json_agg(
        json_build_object(
            'category_id', tc.category_id,
            'category_name', tc.category_name,
            'usage_count', tc.usage_count,
            'total_spent', tc.total_spent
        )
    ) FILTER (WHERE tc.category_id IS NOT NULL) as top_categories
FROM monthly_totals mt
LEFT JOIN top_categories tc ON 
    tc.currency = mt.currency::varchar(3)
    AND tc.rank <= 5
GROUP BY 
    mt.currency::varchar(3), 
    mt.total_income, 
    mt.total_expenses, 
    mt.total_savings
`

type GetMonthlySummaryParams struct {
	UserID uuid.UUID `json:"user_id"`
	Date   time.Time `json:"date"`
}

type GetMonthlySummaryRow struct {
	Currency      string  `json:"currency"`
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
	TotalSavings  float64 `json:"total_savings"`
	TopCategories []byte  `json:"top_categories"`
}

func (q *Queries) GetMonthlySummary(ctx context.Context, arg GetMonthlySummaryParams) ([]GetMonthlySummaryRow, error) {
	rows, err := q.db.Query(ctx, getMonthlySummary, arg.UserID, arg.Date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetMonthlySummaryRow
	for rows.Next() {
		var i GetMonthlySummaryRow
		if err := rows.Scan(
			&i.Currency,
			&i.TotalIncome,
			&i.TotalExpenses,
			&i.TotalSavings,
			&i.TopCategories,
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

const getYearlySummary = `-- name: GetYearlySummary :many
WITH yearly_totals AS (
    SELECT 
        i.currency::varchar(3) AS currency,
        COALESCE(SUM(i.amount), 0)::float8 as total_income,
        COALESCE(SUM(e.amount), 0)::float8 as total_expenses,
        COALESCE(SUM(i.amount) - SUM(e.amount), 0)::float8 as total_savings
    FROM income i
    FULL OUTER JOIN expenses e ON 
        e.user_id = i.user_id 
        AND e.currency = i.currency::varchar(3)
        AND DATE_TRUNC('year', e.date) = DATE_TRUNC('year', i.date)
        AND e.deleted_at IS NULL
    WHERE i.user_id = $1
        AND i.deleted_at IS NULL
        AND DATE_TRUNC('year', i.date) = DATE_TRUNC('year', $2::TIMESTAMPTZ)
    GROUP BY i.currency
),
top_categories AS (
    SELECT 
        e.category_id,
        c.name as category_name,
        e.currency::varchar(3) AS currency,
        COUNT(*) as usage_count,
        SUM(e.amount) as total_spent,
        ROW_NUMBER() OVER (PARTITION BY e.currency::varchar(3) ORDER BY SUM(e.amount) DESC) as rank
    FROM expenses e
    JOIN categories c ON e.category_id = c.id
    WHERE e.user_id = $1
        AND e.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND DATE_TRUNC('year', e.date) = DATE_TRUNC('year', $2::TIMESTAMPTZ)
    GROUP BY e.category_id, c.name, e.currency::varchar(3)
),
monthly_trend AS (
    SELECT 
        DATE_TRUNC('month', e.date) as month,
        e.currency::varchar(3),
        c.name as category_name,
        SUM(e.amount) as monthly_expenses
    FROM expenses e
    JOIN categories c ON e.category_id = c.id
    WHERE e.user_id = $1
        AND e.deleted_at IS NULL
        AND c.deleted_at IS NULL
        AND DATE_TRUNC('year', e.date) = DATE_TRUNC('year', $2::TIMESTAMPTZ)
    GROUP BY DATE_TRUNC('month', e.date), e.currency::varchar(3), c.name
    ORDER BY month
)
SELECT 
    yt.currency, yt.total_income, yt.total_expenses, yt.total_savings,
    json_agg(
        json_build_object(
            'category_id', tc.category_id,
            'category_name', tc.category_name,
            'usage_count', tc.usage_count,
            'total_spent', tc.total_spent
        )
    ) FILTER (WHERE tc.category_id IS NOT NULL) as top_categories,
    json_agg(
        json_build_object(
            'month', mt.month,
            'category_name', mt.category_name,
            'amount', mt.monthly_expenses
        )
    ) FILTER (WHERE mt.month IS NOT NULL) as monthly_trend
FROM yearly_totals yt
LEFT JOIN top_categories tc ON 
    tc.currency = yt.currency::varchar(3)
    AND tc.rank <= 5
LEFT JOIN monthly_trend mt ON 
    mt.currency = yt.currency::varchar(3)
GROUP BY 
    yt.currency::varchar(3), 
    yt.total_income, 
    yt.total_expenses, 
    yt.total_savings
`

type GetYearlySummaryParams struct {
	UserID uuid.UUID `json:"user_id"`
	Date   time.Time `json:"date"`
}

type GetYearlySummaryRow struct {
	Currency      string  `json:"currency"`
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
	TotalSavings  float64 `json:"total_savings"`
	TopCategories []byte  `json:"top_categories"`
	MonthlyTrend  []byte  `json:"monthly_trend"`
}

func (q *Queries) GetYearlySummary(ctx context.Context, arg GetYearlySummaryParams) ([]GetYearlySummaryRow, error) {
	rows, err := q.db.Query(ctx, getYearlySummary, arg.UserID, arg.Date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetYearlySummaryRow
	for rows.Next() {
		var i GetYearlySummaryRow
		if err := rows.Scan(
			&i.Currency,
			&i.TotalIncome,
			&i.TotalExpenses,
			&i.TotalSavings,
			&i.TopCategories,
			&i.MonthlyTrend,
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
